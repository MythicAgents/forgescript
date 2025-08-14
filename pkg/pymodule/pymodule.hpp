#pragma once

#include <functional>
#include <memory>
#include <mutex>
#include <optional>
#include <set>
#include <string_view>
#include <utility>
#include <variant>
#include <vector>

#include <pybind11/cast.h>
#include <pybind11/detail/internals.h>
#include <pybind11/pybind11.h>
#include <pybind11/pytypes.h>
#include <pybind11/typing.h>

namespace forgescript::pymodule {

  namespace details {
    constexpr std::string_view shared_state_key = "forgescript_state";

    // `pybind11::(get/set)_shared_data` only accept a `const std::string&` for the key
    // even though it gets cloned and stored as a `std::string`.
    // This will create a lazily initialized `std::string` with the key so that it does
    // not need to be constantly re-created during each access.
    static inline const std::string& get_shared_state_key() {
      static std::once_flag flag;
      static std::string keystring;
      std::call_once(flag, []() { keystring = shared_state_key; });
      return keystring;
    }

  }; // namespace details

  struct Callback {
    std::string last_checkin;
    std::string user;
    std::string host;
    long long pid;
    std::string ip;
    std::vector<std::string> ips;
    std::string external_ip;
    std::string process_name;
    std::string description;
    std::string operator_username;
    bool active;
    long long integrity_level;
    bool locked;
    std::string operation_name;
    std::string os;
    std::string architecture;
    std::string domain;
    std::string extra_info;
    std::string sleep_info;
  };

  struct [[gnu::visibility("hidden")]] Task {
    Callback callback;
    pybind11::dict args;
    std::string command_line;
  };

  struct [[gnu::visibility("hidden")]] AliasedCommand {
    std::string name;
    pybind11::dict args;
    std::string display_params;
  };

  enum class AliasParameterType : unsigned char {
    String = 0,
    Boolean,
    Number,
    ChooseOne,
    Array,
  };

  constexpr std::string_view alias_parameter_type_str(const AliasParameterType& v) {
    switch (v) {
    case AliasParameterType::String:
      return "String";
    case AliasParameterType::Boolean:
      return "Boolean";
    case AliasParameterType::Number:
      return "Number";
    case AliasParameterType::ChooseOne:
      return "ChooseOne";
    case AliasParameterType::Array:
      return "Array";
    }

    std::unreachable();
  }

  struct [[gnu::visibility("hidden")]] AliasParameter {
    using DefaultValueType =
      std::variant<std::string, bool, float, int, std::vector<std::string>>;

    std::string name;
    std::string display_name;
    std::string cli_name;
    AliasParameterType type;
    std::string description;
    std::vector<std::string> choices;
    std::optional<DefaultValueType> default_value;
  };

  struct AliasAttributes {
    std::vector<std::string> supported_os;
  };

  using AliasCallbackParam = Task;
  using AliasCallbackReturn = AliasedCommand;
  using AliasCallback =
    pybind11::typing::Callable<AliasCallbackReturn(AliasCallbackParam)>;

  struct [[gnu::visibility("hidden")]] RunAliasState {
    std::string_view alias_name;
    long long task_id;
    std::optional<AliasCallback> callback;
  };

  struct [[gnu::visibility("hidden")]] RunScriptState {
    std::string_view operator_name;
    long long callback_id;
    long long task_id;
    std::reference_wrapper<std::set<std::string>> registered;
  };

  using SharedState = std::variant<std::monostate, RunAliasState, RunScriptState>;
  using SharedStateRef = std::reference_wrapper<SharedState>;

  static inline void set_shared_state(SharedState& state) {
    pybind11::set_shared_data(details::get_shared_state_key(), std::addressof(state));
  }

  static inline std::optional<SharedStateRef> get_shared_state() {
    if (auto *ptr = pybind11::get_shared_data(details::get_shared_state_key())) {
      // pybind11 returns shared data as a `void *`
      // NOLINTNEXTLINE(cppcoreguidelines-pro-type-reinterpret-cast)
      return {*reinterpret_cast<SharedState *>(ptr)};
    }

    return {};
  }

}; // namespace forgescript::pymodule
