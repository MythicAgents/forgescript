#pragma once

#include <memory>
#include <string>
#include <utility>
#include <vector>

template<typename T>
using GoResult = std::pair<T, std::string>;

class SubInterpreter {
public:
  SubInterpreter();
  SubInterpreter(const SubInterpreter&) = delete;
  SubInterpreter(SubInterpreter&&) = delete;
  SubInterpreter& operator=(const SubInterpreter&) = delete;
  SubInterpreter& operator=(SubInterpreter&&) = delete;
  ~SubInterpreter();

  /**
   * Runs the script at the specified path.
   * This equates to running the following python code
   *     import runpy
   *     runpy.run_path(scriptPath, run_name="__main__")
   *
   * @param scriptPath The script path to run.
   * @param callbackID The callback ID.
   * @param taskID The task ID.
   * @return GoResult<std::vector<std::string>> List of aliases registered
   */
  GoResult<std::vector<std::string>> RunScript(const std::string& scriptPath,
                                               long long callbackID, long long taskID,
                                               const std::string& operatorName);

  /**
   * Runs an alias callback function.
   * @param scriptPath The script path with the callback function.
   * @param aliasName The name of the callback function to run.
   * @param taskJson The serialized task JSON.
   * @return GoResult<std::string> TODO
   */
  GoResult<std::string> RunAliasCallback(const std::string& scriptPath, long long taskID,
                                         const std::string& aliasName,
                                         const std::string& taskJson);

private:
  class Impl;
  std::unique_ptr<Impl> pImpl;
};

class MainInterpreter {
public:
  MainInterpreter();
  MainInterpreter(const MainInterpreter&) = delete;
  MainInterpreter(MainInterpreter&&) = delete;
  MainInterpreter& operator=(const MainInterpreter&) = delete;
  MainInterpreter& operator=(MainInterpreter&&) = delete;
  ~MainInterpreter();

private:
  class Impl;
  std::unique_ptr<Impl> pImpl;
};

std::string OSThreadId();
