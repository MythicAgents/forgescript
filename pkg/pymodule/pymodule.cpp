#include <cassert>
#include <cstdint>
#include <filesystem>
#include <fstream>
#include <iostream>
#include <optional>
#include <string>
#include <string_view>
#include <variant>
#include <vector>

#include <nlohmann/json.hpp>
#include <pybind11/cast.h>
#include <pybind11/embed.h>
#include <pybind11/native_enum.h>
#include <pybind11/pybind11.h>
#include <pybind11/pytypes.h>
#include <pybind11/stl.h>
#include <pybind11/stl/filesystem.h>
#include <pybind11/subinterpreter.h>
#include <pybind11/typing.h>
#include <pyerrors.h>

#include "gobindings/gobindings.hpp"
#include "pybind11/detail/common.h"
#include "pymodule.hpp"

namespace py = pybind11;
namespace pymodule = forgescript::pymodule;

namespace {

  void register_alias(std::string_view name, const pymodule::AliasCallback& callback,
                      const std::vector<pymodule::AliasParameter>& parameters = {},
                      std::string_view description = {},
                      std::string_view help_string = {}, const std::uint32_t version = 1,
                      std::string_view author = {},
                      const std::optional<pymodule::AliasAttributes>& attributes = {}) {

    if (name.empty()) {
      py::set_error(PyExc_ValueError, "name is an empty string");
      return;
    }

    auto state = pymodule::get_shared_state();
    assert(state);

    if (auto *run_alias = std::get_if<pymodule::RunAliasState>(&state->get())) {
      if (!run_alias->callback && run_alias->alias_name == name) {
        run_alias->callback = callback;
      }

      return;
    }

    // This should be an exclusive globals access since there is only one interpreter
    // per thread.
    py::dict gbls = py::globals();
    auto script_path = gbls["__file__"].cast<std::string_view>();

    auto& run_script_state = std::get<pymodule::RunScriptState>(state->get());

    if (!author.empty()) {
      author = run_script_state.operator_name;
    }

    nlohmann::json command{
      {"name", name},
      {"needs_admin_permissions", false},
      {"help_string", help_string},
      {"description", description},
      {"version", version},
      {"suported_ui_features", nlohmann::json::array()},
      {"author", author},
      {"attack", nlohmann::json::array()},
      {"script_only", false},
      {"parameters", nlohmann::json::array()},
    };

    if (attributes) {
      command["attributes"] = nlohmann::json::object();
      command["attributes"]["supported_os"] = attributes->supported_os;
    }

    for (const auto& [idx, parameter]: std::ranges::views::enumerate(parameters)) {
      nlohmann::json jsonparam{
        {"name", parameter.name},
        {"display_name", parameter.display_name},
        {"cli_name", parameter.cli_name},
        {"parameter_type", pymodule::alias_parameter_type_str(parameter.type)},
        {"description", parameter.description},
        {"choices", parameter.choices},
        {"parameter_group_info",
         nlohmann::json::array({nlohmann::json::object({
           {"group_name", "Default"},
           {"ui_position", idx + 1},
           {"required", true},
         })})}};

      if (parameter.default_value) {
        if (parameter.type == pymodule::AliasParameterType::String) {
          const auto default_value = std::get<std::string>(*parameter.default_value);
          if (default_value.empty()) {
            jsonparam["parameter_group_info"][0]["required"] = false;
          }

          jsonparam["default_value"] = default_value;
        } else if (parameter.type == pymodule::AliasParameterType::Boolean) {
          jsonparam["default_value"] = std::get<bool>(*parameter.default_value);
        } else if (parameter.type == pymodule::AliasParameterType::Number) {
          if (const auto *float_val = std::get_if<float>(&*parameter.default_value)) {
            jsonparam["default_value"] = *float_val;
          } else {
            jsonparam["default_value"] = std::get<int>(*parameter.default_value);
          }
        } else if (parameter.type == pymodule::AliasParameterType::ChooseOne) {
          jsonparam["default_value"] = std::get<std::string>(*parameter.default_value);
        } else if (parameter.type == pymodule::AliasParameterType::Array) {
          jsonparam["default_value"] =
            std::get<std::vector<std::string>>(*parameter.default_value);
        } else {
          throw py::value_error{"Invalid default value for parameter type"};
        }
      }

      command["parameters"].push_back(jsonparam);
    }

    std::string command_json{command.dump()};

    const auto result = gobindings::create_command(script_path,
                                                   run_script_state.callback_id,
                                                   run_script_state.task_id,
                                                   command_json);

    if (!result.has_value()) {
      py::set_error(PyExc_Exception, result.error().c_str());
      return;
    }

    auto& registered_aliases = run_script_state.registered.get();
    registered_aliases.insert(std::string{name});
  }

  std::string register_file(const std::filesystem::path& path) {
    std::filesystem::path full_path{};

    if (path.is_relative()) {
      auto script_path = pybind11::globals()["__file__"].cast<std::filesystem::path>();
      full_path = script_path.parent_path() / path;
    } else {
      full_path = path;
    }

    auto file_size = std::filesystem::file_size(full_path);
    std::vector<std::byte> file_data(file_size);
    std::ifstream filestream{full_path, std::ios::binary};
    filestream.read(reinterpret_cast<char *>(file_data.data()),
                    static_cast<long>(file_size));

    auto state = pymodule::get_shared_state();
    auto task_id = 0L;
    auto delete_after_fetch = false;
    if (auto *run_alias = std::get_if<pymodule::RunAliasState>(&state->get())) {
      task_id = run_alias->task_id;
      delete_after_fetch = true;
    } else if (auto *run_script = std::get_if<pymodule::RunScriptState>(&state->get())) {
      task_id = run_script->task_id;
    } else {
      std::unreachable();
    }

    auto file_name = full_path.filename().string();

    auto register_file_result =
      gobindings::register_file(task_id, file_data, file_name, delete_after_fetch);
    if (!register_file_result) {
      throw std::runtime_error(register_file_result.error());
    }

    return *register_file_result;
  }

}; // namespace

// NOLINTBEGIN
PYBIND11_EMBEDDED_MODULE(forgescript, mod,
                         py::multiple_interpreters::per_interpreter_gil()) {
  mod.doc() = "builtin forgescript module",

  mod.def("register_alias",
          &register_alias,
          "Register a command alias with the specified payload types",
          py::arg("name"),
          py::arg("callback"),
          py::kw_only(),
          py::arg("parameters") = std::vector<pymodule::AliasParameter>{},
          py::arg("description") = std::string_view{},
          py::arg("help_string") = std::string_view{},
          py::arg("version") = 1,
          py::arg("author") = std::string_view{},
          py::arg("attributes") = std::nullopt);

  mod.def("register_file",
          &register_file,
          "Registers a file with Mythic and returns the file uuid",
          py::arg("path"));

  py::class_<pymodule::Callback>(mod, "Callback")
    .def_readonly("last_checkin", &pymodule::Callback::last_checkin)
    .def_readonly("user", &pymodule::Callback::user)
    .def_readonly("host", &pymodule::Callback::host)
    .def_readonly("pid", &pymodule::Callback::pid)
    .def_readonly("ips", &pymodule::Callback::ips)
    .def_readonly("external_ip", &pymodule::Callback::external_ip)
    .def_readonly("process_name", &pymodule::Callback::process_name)
    .def_readonly("description", &pymodule::Callback::description)
    .def_readonly("operator_username", &pymodule::Callback::operator_username)
    .def_readonly("active", &pymodule::Callback::active)
    .def_readonly("integrity_level", &pymodule::Callback::integrity_level)
    .def_readonly("locked", &pymodule::Callback::locked)
    .def_readonly("operation_name", &pymodule::Callback::operation_name)
    .def_readonly("os", &pymodule::Callback::os)
    .def_readonly("architecture", &pymodule::Callback::architecture)
    .def_readonly("domain", &pymodule::Callback::domain)
    .def_readonly("extra_info", &pymodule::Callback::extra_info)
    .def_readonly("sleep_info", &pymodule::Callback::sleep_info);

  py::class_<pymodule::Task>(mod, "Task")
    .def_readonly("callback", &pymodule::Task::callback)
    .def_readonly("args", &pymodule::Task::args)
    .def_readonly("command_line", &pymodule::Task::command_line);

  py::class_<pymodule::AliasedCommand>(mod, "AliasedCommand")
    .def(py::init<std::string>())
    .def(py::init([](std::string name, py::dict args, std::string display_params) {
           return pymodule::AliasedCommand{name, args, display_params};
         }),
         py::arg("name"),
         py::kw_only(),
         py::arg("args") = py::dict(),
         py::arg("display_params") = std::string{});

  py::class_<pymodule::AliasParameter>(mod, "AliasParameter")
    .def(py::init([](std::string name,
                     std::string display_name,
                     std::string cli_name,
                     pymodule::AliasParameterType type,
                     std::string description,
                     std::vector<std::string>
                       choices,
                     std::optional<pymodule::AliasParameter::DefaultValueType>
                       default_value) {
           return pymodule::AliasParameter{.name = name,
                                           .display_name = display_name,
                                           .cli_name = cli_name,
                                           .type = type,
                                           .description = description,
                                           .choices = choices,
                                           .default_value = default_value};
         }),
         py::arg("name"),
         py::kw_only(),
         py::arg("display_name") = std::string{},
         py::arg("cli_name") = std::string{},
         py::arg("type"),
         py::arg("description") = std::string{},
         py::arg("choices") = std::vector<std::string>{},
         py::arg("default_value") = std::nullopt);

  py::class_<pymodule::AliasAttributes>(mod, "AliasAttributes")
    .def(py::init<>())
    .def(py::init([](std::vector<std::string> supported_os) {
           return pymodule::AliasAttributes{.supported_os = supported_os};
         }),
         py::kw_only(),
         py::arg("supported_os"));

  py::native_enum<pymodule::AliasParameterType>(mod, "AliasParameterType", "enum.Enum")
    .value("String", pymodule::AliasParameterType::String)
    .value("Boolean", pymodule::AliasParameterType::Boolean)
    .value("Number", pymodule::AliasParameterType::Number)
    .value("ChooseOne", pymodule::AliasParameterType::ChooseOne)
    .value("Array", pymodule::AliasParameterType::Array)
    .export_values()
    .finalize();
}
// NOLINTEND
