package agentfunctions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/MythicAgents/forgescript/pkg/config"
	"github.com/MythicAgents/forgescript/pkg/python"
	"github.com/MythicAgents/forgescript/pkg/versioninfo"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/utils/sharedStructs"
)

type AliasCommand struct {
	Name string `json:"name"`
	Args map[string]any `json:"args"`
	DisplayParams string `json:"display_params"`
}

type AliasTask struct {
	Callback agentstructs.PTTaskMessageCallbackData `json:"callback"`
	Args map[string]any `json:"args"`
	CommandLine string `json:"command_line"`
}

const (
	payloadName = "forgescript"
	supportedPayloadsKey = "forgescript_payloads"
)

var (
	supportedOSList = []string{
		agentstructs.SUPPORTED_OS_MACOS,
		agentstructs.SUPPORTED_OS_WINDOWS,
		agentstructs.SUPPORTED_OS_LINUX,
		agentstructs.SUPPORTED_OS_CHROME,
		agentstructs.SUPPORTED_OS_WEBSHELL,
	}
)

func formatDescriptionMetadata() string {
	metadataFields := []string{}

	moduleVersion := versioninfo.ModuleVersion()
	if len(moduleVersion) == 0 {
		moduleVersion = "development"
	}

	metadataFields = append(metadataFields, fmt.Sprintf("Version %s", moduleVersion))

	gitRevision := versioninfo.GitRevision()
	if len(gitRevision) == 0 {
		gitRevision = "unknown"
	}

	metadataFields = append(metadataFields, fmt.Sprintf("Git Commit %s", gitRevision))

	requiredMythic := versioninfo.RequiredMythicVersion()
	if len(requiredMythic) == 0 {
		requiredMythic = "unknown"
	}

	metadataFields = append(metadataFields, fmt.Sprintf("Requires Mythic %s", requiredMythic))
	return strings.Join(metadataFields, "\n")
}

var payloadDefinition = agentstructs.PayloadType{
	Name:                   payloadName,
	FileExtension:          "bin",
	Author:                 "@M_alphaaa",
	SupportsDynamicLoading: true,
	Description:            fmt.Sprintf("Dynamic scriptable aliases with Python.\n%s", formatDescriptionMetadata()),
	MythicEncryptsData:     true,
	AgentType:              agentstructs.AgentTypeCommandAugment,
	OnContainerStartFunction: func(message sharedStructs.ContainerOnStartMessage) sharedStructs.ContainerOnStartMessageResponse {
		return sharedStructs.ContainerOnStartMessageResponse{
			ContainerName: message.ContainerName,
		}
	},
	CheckIfCallbacksAliveFunction: func(message agentstructs.PTCheckIfCallbacksAliveMessage) agentstructs.PTCheckIfCallbacksAliveMessageResponse {
		return agentstructs.PTCheckIfCallbacksAliveMessageResponse{
			Success: true,
		}
	},
}

func Initialize() {
	runtimeDir := config.GetForgeScriptRuntimePath()

	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		logging.LogFatalError(err, fmt.Sprintf("could not create forgescript runtime path %s", runtimeDir))
	}

	logging.LogInfo("Configured runtime directory", "runtimePath", runtimeDir)

	payloadData := agentstructs.AllPayloadData.Get(payloadName)
	payloadData.AddPayloadDefinition(payloadDefinition)
}

func AddAliasCommand(scriptPath string, callbackID int, taskID int, command agentstructs.Command) error {
	logging.LogDebug("Adding alias command", "command", command)

	command.TaskFunctionCreateTasking = func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
		response := agentstructs.PTTaskCreateTaskingMessageResponse{
			TaskID: taskData.Task.ID,
			Success: false,
		}

		aliasTask := AliasTask{
			Callback: taskData.Callback,
			CommandLine: taskData.Args.GetRawCommandLine(),
			Args: map[string]any{},
		}

		for _, commandParamSpec := range command.CommandParameters {
			switch commandParamSpec.ParameterType {
			case agentstructs.COMMAND_PARAMETER_TYPE_STRING:
				arg, err := taskData.Args.GetStringArg(commandParamSpec.Name)
				if err != nil {
					logging.LogError(err, "Could not parse task parameter", "name", commandParamSpec.Name)
					response.Error = fmt.Sprintf("Could not parse task parameter %s", commandParamSpec.Name)
					return response
				}

				aliasTask.Args[commandParamSpec.Name] = arg
			case agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN:
				arg, err := taskData.Args.GetBooleanArg(commandParamSpec.Name)
				if err != nil {
					logging.LogError(err, "Could not parse task parameter", "name", commandParamSpec.Name)
					response.Error = fmt.Sprintf("Could not parse task parameter %s", commandParamSpec.Name)
					return response
				}

				aliasTask.Args[commandParamSpec.Name] = arg
			case agentstructs.COMMAND_PARAMETER_TYPE_NUMBER:
				arg, err := taskData.Args.GetNumberArg(commandParamSpec.Name)
				if err != nil {
					logging.LogError(err, "Could not parse task parameter", "name", commandParamSpec.Name)
					response.Error = fmt.Sprintf("Could not parse task parameter %s", commandParamSpec.Name)
					return response
				}

				aliasTask.Args[commandParamSpec.Name] = arg
			case agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE:
				arg, err := taskData.Args.GetChooseOneArg(commandParamSpec.Name)
				if err != nil {
					logging.LogError(err, "Could not parse task parameter", "name", commandParamSpec.Name)
					response.Error = fmt.Sprintf("Could not parse task parameter %s", commandParamSpec.Name)
					return response
				}

				aliasTask.Args[commandParamSpec.Name] = arg
			case agentstructs.COMMAND_PARAMETER_TYPE_ARRAY:
				arg, err := taskData.Args.GetArrayArg(commandParamSpec.Name)
				if err != nil {
					logging.LogError(err, "Could not parse task parameter", "name", commandParamSpec.Name)
					response.Error = fmt.Sprintf("Could not parse task parameter %s", commandParamSpec.Name)
					return response
				}

				if len(arg) == 1 && len(arg[0]) == 0 {
					aliasTask.Args[commandParamSpec.Name] = []string{}
				} else {
					aliasTask.Args[commandParamSpec.Name] = arg
				}
			}

			taskData.Args.RemoveArg(commandParamSpec.Name)
		}

		serializedTask, err := json.Marshal(aliasTask)
		if err != nil {
			logging.LogError(err, "Could not serialize task data")
			response.Error = err.Error()
			return response
		}

		aliasCallbackResult, err := python.RunAliasCallback(scriptPath, taskData.Task.ID, command.Name, string(serializedTask))
		if err != nil {
			logging.LogError(err, "Could not run alias callback")
			response.Error = err.Error()
			return response
		}

		logging.LogDebug("Received new alias command", "commandJson", aliasCallbackResult)
		aliasCommand := AliasCommand{}
		if err := json.Unmarshal([]byte(aliasCallbackResult), &aliasCommand); err != nil {
			logging.LogError(err, "Could not deserialize alias callback result")
			response.Error = err.Error()
			return response
		}

		logging.LogDebug("Received new alias command", "command", aliasCommand)

		callbackAgent := taskData.PayloadType
		logging.LogDebug("Callback payload", "agent", callbackAgent)

		response.CommandName = &aliasCommand.Name
		response.ReprocessAtNewCommandPayloadType = callbackAgent
		if len(aliasCommand.DisplayParams) > 0 {
			response.DisplayParams = &aliasCommand.DisplayParams
		}

		newParameters := []agentstructs.CommandParameter{}

		if len(aliasCommand.Args) > 0 {
			logging.LogDebug("Reprocessed alias arguments", "args", aliasCommand.Args)

			jsonArgs, _ := json.Marshal(aliasCommand.Args)

			// This gets forwarded to the commandline during GenerateArgsData. Some agents
			// require the command line to be a valid argument JSON string.
			// https://github.com/MythicAgents/Apollo/blob/ed19409215b62521593fed314ed21a1dcc6aa2fa/Payload_Type/apollo/apollo/mythic/agent_functions/execute_coff.py#L161
			taskData.Task.Params = string(jsonArgs)
			taskData.Task.OriginalParams = string(jsonArgs)

			for key, val := range aliasCommand.Args {
				argtype := ""
				switch val.(type) {
				case string:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_STRING
				case bool:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN
				case []string:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_ARRAY
				case [][]string:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_TYPED_ARRAY
				case int:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_NUMBER
				case float64:
					argtype = agentstructs.COMMAND_PARAMETER_TYPE_NUMBER
				}

				newParameters = append(newParameters, agentstructs.CommandParameter{
					Name: key,
					ParameterType: argtype,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							GroupName: "Default",
						},
					},
				})
			}
		}

		newTaskArgs, err := agentstructs.GenerateArgsData(newParameters, *taskData)
		if err != nil {
			logging.LogError(err, "Could not create new arg data")
			response.Error = err.Error()
			return response
		}

		if err := newTaskArgs.LoadArgsFromDictionary(aliasCommand.Args); err != nil {
			logging.LogError(err, "Could not load new task args")
			response.Error = err.Error()
			return response
		}

		taskData.Args = newTaskArgs
		response.Success = true
		return response
	}

	command.TaskFunctionParseArgString = func(args *agentstructs.PTTaskMessageArgsData, input string) error {
		if len(input) > 0 {
			return args.LoadArgsFromJSONString(input)
		}

		return nil
	}

	command.TaskFunctionParseArgDictionary = func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
		return args.LoadArgsFromDictionary(input)
	}

	command.CommandAttributes.CommandIsSuggested = true
	agentstructs.AllPayloadData.Get(payloadName).AddCommand(command)
	logging.LogDebug("Registered alias", "name", command.Name)
	return nil
}
