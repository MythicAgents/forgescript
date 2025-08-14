package python

import (
	"errors"
	"os"
	"runtime"

	"github.com/MythicAgents/forgescript/pkg/python/bindings"
	"github.com/MythicMeta/MythicContainer/logging"
)

var eventQueue = make(chan func())

func StartExecutorLoop() {
	runtime.LockOSThread()

	interpreter := bindings.NewMainInterpreter()
	defer bindings.DeleteMainInterpreter(interpreter)

	mainTid := bindings.OSThreadId()
	logging.LogInfo("Started python executor event loop", "thread_id", mainTid)

	for f := range eventQueue {
		f()
	}
}

func do[T any](f func() T) chan T {
	done := make(chan T, 1)
	eventQueue <- func() {
		done <- f()
	}

	return done
}

func withSubInterpreter[T any](f func(sub bindings.SubInterpreter) T) T {
	subinterpreter := <-do(func() bindings.SubInterpreter {
		subinterpreter := bindings.NewSubInterpreter()
		logging.LogDebug("Creating new python subinterpreter", "thread_id", bindings.OSThreadId(), "pointer", subinterpreter.Swigcptr())
		return subinterpreter
	})

	runtime.LockOSThread()
	result := f(subinterpreter)
	runtime.UnlockOSThread()

	go do(func() interface{} {
		logging.LogDebug("Deleting python subinterpreter", "thread_id", bindings.OSThreadId(), "pointer", subinterpreter.Swigcptr())
		bindings.DeleteSubInterpreter(subinterpreter)
		return nil
	})

	return result
}

// Runs the script at the specified path.
// Returns a map with the commands registered and list of payload types that registered
// the command
func RunScript(scriptPath string, callbackID int, taskID int, operatorName string) ([]string, error) {
	if scriptStat, err := os.Stat(scriptPath); err != nil {
		return []string{}, err
	} else if scriptStat.IsDir() {
		return []string{}, errors.New("script path is a directory")
	}

	result := withSubInterpreter(func(subinterpreter bindings.SubInterpreter) bindings.GoVecStringResult {
		logging.LogDebug("Running python.RunScript", "thread_id", bindings.OSThreadId())
		return subinterpreter.RunScript(scriptPath, int64(callbackID), int64(taskID), operatorName)
	})
	defer bindings.DeleteGoVecStringResult(result)

	errv := result.GetSecond()
	if len(errv) > 0 {
		err := errors.New(errv)
		logging.LogError(err, "RunScript returned an error")
		return []string{}, err
	}

	registerResult := result.GetFirst()
	registeredAliases := make([]string, registerResult.Size())
	for i := range registeredAliases {
		registeredAliases[i] = registerResult.Get(i)
	}

	return registeredAliases, nil
}

func RunAliasCallback(scriptPath string, taskID int, aliasName string, taskJson string) (string, error) {
	if scriptStat, err := os.Stat(scriptPath); err != nil {
		return "", err
	} else if scriptStat.IsDir() {
		return "", errors.New("script path is a directory")
	}

	result := withSubInterpreter(func(subinterpreter bindings.SubInterpreter) bindings.GoStringResult {
		logging.LogDebug("Running python.RunAliasCallback", "thread_id", bindings.OSThreadId())
		return subinterpreter.RunAliasCallback(scriptPath, int64(taskID), aliasName, taskJson)
	})
	defer bindings.DeleteGoStringResult(result)

	errv := result.GetSecond()
	if len(errv) > 0 {
		return "", errors.New(errv)
	}

	return result.GetFirst(), nil
}
