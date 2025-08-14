#include "bindings.hpp"

#include <algorithm>
#include <iterator>
#include <sstream>
#include <string>
#include <thread>
#include <tuple>

#include <nlohmann/json.hpp>
#include <pybind11/embed.h>
#include <pybind11/gil.h>
#include <pybind11/pybind11.h>
#include <pybind11/pytypes.h>
#include <pybind11/subinterpreter.h>
#include <pylifecycle.h>

#include <pymodule/pymodule.hpp>

namespace py = pybind11;

class [[gnu::visibility("hidden")]] MainInterpreter::Impl {
  py::scoped_interpreter m_maininterpreter;
  py::gil_scoped_release m_release;

public:
  Impl() = default;
  Impl(const Impl&) = delete;
  Impl(Impl&&) = delete;
  Impl& operator=(const Impl&) = delete;
  Impl& operator=(Impl&&) = delete;
  ~Impl() = default;
};

class [[gnu::visibility("hidden")]] SubInterpreter::Impl {
  py::subinterpreter m_subinterpreter;

public:
  Impl() {
    py::gil_scoped_acquire gil{};
    m_subinterpreter = py::subinterpreter::create();
  }
  Impl(const Impl&) = delete;
  Impl(Impl&&) = delete;
  Impl& operator=(const Impl&) = delete;
  Impl& operator=(Impl&&) = delete;
  ~Impl() = default;

  GoResult<std::vector<std::string>> RunScript(const std::string& scriptPath,
                                               long long callbackID, long long taskID,
                                               const std::string& operatorName);
  GoResult<std::string> RunAliasCallback(const std::string& scriptPath, long long taskID,
                                         const std::string& aliasName,
                                         const std::string& taskJson);
};

GoResult<std::vector<std::string>>
SubInterpreter::Impl::RunScript(const std::string& scriptPath, long long callbackID,
                                long long taskID, const std::string& operatorName) {
  py::subinterpreter_scoped_activate guard{m_subinterpreter};

  try {
    using namespace py::literals;
    namespace pymodule = forgescript::pymodule;

    std::set<std::string> registered{};

    pymodule::SharedState state{pymodule::RunScriptState{
      .operator_name = operatorName,
      .callback_id = callbackID,
      .task_id = taskID,
      .registered = registered,
    }};

    pymodule::set_shared_state(state);

    auto runpy = py::module_::import("runpy");
    runpy.attr("run_path")(scriptPath, "run_name"_a = "__main__");

    std::vector<std::string> result{};
    result.reserve(registered.size());
    std::ranges::transform(registered, std::back_inserter(result), [](const auto& v) {
      return v;
    });

    return {result, {}};
  } catch (py::error_already_set& exc) {
    exc.discard_as_unraisable(__func__);
    return {{}, exc.what()};
  } catch (std::exception& exc) {
    return {{}, exc.what()};
  }

  return {};
}

GoResult<std::string>
SubInterpreter::Impl::RunAliasCallback(const std::string& scriptPath, long long taskID,
                                       const std::string& aliasName,
                                       const std::string& taskJson) {
  py::subinterpreter_scoped_activate guard{m_subinterpreter};

  try {
    using namespace py::literals;
    namespace pymodule = forgescript::pymodule;

    auto deserialized_task = nlohmann::json::parse(taskJson);
    const auto& task_args = deserialized_task["args"];
    const auto& callback = deserialized_task["callback"];

    pymodule::Task task{};
    task.callback = pymodule::Callback{
      .last_checkin = callback["last_checkin"],
      .user = callback["user"],
      .host = callback["host"],
      .pid = callback["pid"],
      .ip = callback["ip"],
      .ips = callback["ips"],
      .external_ip = callback["external_ip"],
      .process_name = callback["process_name"],
      .description = callback["description"],
      .operator_username = callback["operator_username"],
      .active = callback["active"],
      .integrity_level = callback["integrity_level"],
      .locked = callback["locked"],
      .operation_name = callback["operation_name"],
      .os = callback["os"],
      .architecture = callback["architecture"],
      .domain = callback["domain"],
      .extra_info = callback["extra_info"],
      .sleep_info = callback["sleep_info"],
    };

    if (!task_args.empty()) {
      for (const auto& [key, value]: task_args.items()) {
        if (value.is_string()) {
          task.args[py::str(key)] = value.template get<std::string>();
        } else if (value.is_number_float()) {
          task.args[py::str(key)] = value.template get<float>();
        } else if (value.is_number()) {
          task.args[py::str(key)] = value.template get<int>();
        } else if (value.is_boolean()) {
          task.args[py::str(key)] = value.template get<bool>();
        } else if (value.is_array()) {
          py::list dictlist{};
          for (const auto& arrvalue: value.template get<std::vector<std::string>>()) {
            dictlist.append(arrvalue);
          }

          task.args[py::str(key)] = dictlist;
        }
      }
    }

    task.command_line = deserialized_task["command_line"];

    pymodule::SharedState state{pymodule::RunAliasState{
      .alias_name = aliasName,
      .task_id = taskID,
      .callback = {},
    }};

    pymodule::set_shared_state(state);

    auto runpy = py::module_::import("runpy");

    runpy.attr("run_path")(scriptPath);

    auto& runstate = std::get<pymodule::RunAliasState>(state);
    if (runstate.callback) {

      auto resp = (*runstate.callback)(task);
      py::print("Alias function response", resp);

      auto aliased = resp.cast<pymodule::AliasedCommand>();

      py::dict aliased_dict{};
      aliased_dict["name"] = aliased.name;
      aliased_dict["args"] = aliased.args;
      aliased_dict["command_line"] = aliased.display_params;

      auto pyjson = py::module_::import("json");
      auto pyserialized =
        pyjson.attr("dumps")(aliased_dict, "separators"_a = std::make_tuple(',', ':'));

      return {pyserialized.cast<std::string>(), {}};
    }

    throw std::runtime_error("could not find script registered alias callback function");
  } catch (py::error_already_set& exc) {
    exc.discard_as_unraisable(__func__);
    return {{}, exc.what()};
  } catch (std::exception& exc) {
    return {{}, exc.what()};
  }

  return {};
}

SubInterpreter::SubInterpreter(): pImpl(new Impl) {}
SubInterpreter::~SubInterpreter() = default;
GoResult<std::vector<std::string>>
SubInterpreter::RunScript(const std::string& scriptPath, long long callbackID,
                          long long taskID, const std::string& operatorName) {
  return pImpl->RunScript(scriptPath, callbackID, taskID, operatorName);
}

GoResult<std::string> SubInterpreter::RunAliasCallback(const std::string& scriptPath,
                                                       long long taskID,
                                                       const std::string& aliasName,
                                                       const std::string& taskJson) {
  return pImpl->RunAliasCallback(scriptPath, taskID, aliasName, taskJson);
}

MainInterpreter::MainInterpreter(): pImpl(new Impl) {}
MainInterpreter::~MainInterpreter() = default;

std::string OSThreadId() {
  std::stringstream tid;
  tid << std::this_thread::get_id();
  return tid.str();
}
