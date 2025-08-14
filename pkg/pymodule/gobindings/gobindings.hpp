#pragma once

#include "gobindings.h"

#include <cassert>
#include <cstddef>
#include <expected>
#include <span>
#include <string>
#include <string_view>
#include <type_traits>

namespace gobindings {

  namespace details {

    template<class T>
    concept GoResultPair = requires(T a) {
      a.r0;
      std::is_same_v<decltype(a.r1), CGoReturnedError>;
    };

    constexpr auto to_gostring(std::string_view s) {
      return GoString{.p = s.data(), .n = static_cast<long>(s.size())};
    }

    constexpr auto from_gostringsv(GoString s) {
      return std::string_view{s.p, static_cast<unsigned long>(s.n)};
    }

    constexpr auto from_gostring(CGoReturnedString s) {
      std::string ret{s.ptr, static_cast<unsigned long>(s.size)};
      std::free(s.ptr);
      return ret;
    }

    template<typename T>
    constexpr auto to_goslice(std::span<T> data) {
      return GoSlice{.data = reinterpret_cast<void *>(data.data()),
                     .len = static_cast<long>(data.size()),
                     .cap = static_cast<long>(data.size())};
    }

    static inline std::expected<void, std::string>
    goerr_to_expected(CGoReturnedError err) {
      if (err.ptr != 0) {
        auto ret = from_gostring(ForgescriptUtilErrorToStringCGo(err));
        ForgescriptUtilErrorDelete(err);
        return std::unexpected(ret);
      }

      return {};
    }

    static inline auto goresult_to_expected(const GoResultPair auto& result)
      -> std::expected<decltype(result.r0), std::string> {
      if (result.r1.ptr != 0) {
        auto ret = from_gostring(ForgescriptUtilErrorToStringCGo(result.r1));
        ForgescriptUtilErrorDelete(result.r1);
        return std::unexpected(ret);
      }

      return result.r0;
    }
  }; // namespace details

  static inline std::expected<void, std::string>
  create_command(std::string_view script_path, long long callback_id, long long task_id,
                 std::string_view command_json) {

    return details::goerr_to_expected(
      ForgescriptPyModuleCreateCommandCGo(details::to_gostring(script_path),
                                          callback_id,
                                          task_id,
                                          details::to_gostring(command_json)));
  }

  static inline std::expected<std::string, std::string>
  register_file(long long task_id, std::span<std::byte> contents,
                std::string_view file_name, bool delete_after_fetch) {
    return details::goresult_to_expected(
             ForgescriptPyModuleRegisterFileCGo(
               task_id,
               details::to_goslice(contents),
               details::to_gostring(file_name),
               static_cast<unsigned char>(delete_after_fetch)))
      .transform(details::from_gostring);
  }
}; // namespace gobindings
