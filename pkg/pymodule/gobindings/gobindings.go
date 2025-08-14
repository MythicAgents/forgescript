package gobindings

/*
#include <stdint.h>

typedef struct CGoReturnedString {
	char *ptr;
	long long size;
} CGoReturnedString;

typedef struct CGoReturnedError {
	uintptr_t ptr;
} CGoReturnedError;
*/
import "C"

import (
	"encoding/json"
	"errors"
	"runtime/cgo"

	"github.com/MythicAgents/forgescript/pkg/agentfunctions"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func NewCGOReturnedString(s string) C.CGoReturnedString {
	return C.CGoReturnedString{
		ptr: C.CString(s),
		size: (C.longlong)(len(s)),
	}
}

func NewCGOEmptyString() C.CGoReturnedString {
	return C.CGoReturnedString{
		ptr: nil,
		size: (C.longlong)(0),
	}
}

func NewCGOReturnedError(err error) C.CGoReturnedError {
	if err == nil {
		return C.CGoReturnedError{
			ptr: C.uintptr_t(0),
		}
	}

	return C.CGoReturnedError{
		ptr: C.uintptr_t(cgo.NewHandle(err)),
	}
}

//export ForgescriptPyModuleCreateCommandCGo
func ForgescriptPyModuleCreateCommandCGo(scriptPath string, callbackID int, taskID int, commandJson string) C.CGoReturnedError {
	commandSpec := agentstructs.Command{}
	if err := json.Unmarshal([]byte(commandJson), &commandSpec); err != nil {
		return NewCGOReturnedError(err)
	}


	if err := agentfunctions.AddAliasCommand(scriptPath, callbackID, taskID, commandSpec); err != nil {
		return NewCGOReturnedError(err)
	}

	return NewCGOReturnedError(nil)
}

//export ForgescriptPyModuleRegisterFileCGo
func ForgescriptPyModuleRegisterFileCGo(taskID int, contents []byte, fileName string, deleteAfterFetch bool) (C.CGoReturnedString, C.CGoReturnedError) {
	rpcResult, err := mythicrpc.SendMythicRPCFileCreate(mythicrpc.MythicRPCFileCreateMessage{
		TaskID: taskID,
		FileContents: contents,
		Filename: fileName,
		DeleteAfterFetch: deleteAfterFetch,
	})

	if err != nil {
		return NewCGOEmptyString(), NewCGOReturnedError(err)
	}

	if !rpcResult.Success {
		return NewCGOEmptyString(), NewCGOReturnedError(errors.New(rpcResult.Error))
	}

	return NewCGOReturnedString(rpcResult.AgentFileId), NewCGOReturnedError(nil)
}

//export ForgescriptUtilErrorToStringCGo
func ForgescriptUtilErrorToStringCGo(err C.CGoReturnedError) C.CGoReturnedString {
	errv := cgo.Handle(err.ptr)
	val := errv.Value().(error)
	return NewCGOReturnedString(val.Error())
}

//export ForgescriptUtilErrorDelete
func ForgescriptUtilErrorDelete(err C.CGoReturnedError) {
	h := cgo.Handle(err.ptr)
	h.Delete()
}
