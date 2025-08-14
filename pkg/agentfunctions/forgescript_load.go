package agentfunctions

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/MythicAgents/forgescript/pkg/config"
	"github.com/MythicAgents/forgescript/pkg/extract"
	"github.com/MythicAgents/forgescript/pkg/python"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
)

func init() {
	agentstructs.AllPayloadData.Get(payloadName).AddCommand(agentstructs.Command{
		Name:        fmt.Sprintf("%s_load", payloadName),
		HelpString:  fmt.Sprintf("%s_load [popup]", payloadName),
		Description: "Load a forgescript bundle and evaluate the specified script",
		Version:     1,
		SupportedUIFeatures: []string{
			fmt.Sprintf("%s:load", payloadName),
		},
		Author:            "@M_alphaaa",
		ScriptOnlyCommand: true,
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:      supportedOSList,
			CommandIsBuiltin: true,
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "bundle",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "The script bundle to load",
				ModalDisplayName: "Script bundle (.tar.gz, .zip)",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "script",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "The path to the script inside the bundle to load",
				DefaultValue:     "forgescript_alias.py",
				ModalDisplayName: "Load script",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				TaskID: taskData.Task.ID,
			}

			fileId, err := taskData.Args.GetFileArg("bundle")
			if err != nil {
				logging.LogError(err, "failed to get loaded bundle")
				response.Error = err.Error()
				return response
			}

			fileSearchResp, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				TaskID:          taskData.Task.ID,
				CallbackID:      taskData.Callback.ID,
				LimitByCallback: true,
				AgentFileID:     fileId,
				MaxResults:      1,
			})
			if err != nil {
				logging.LogError(err, "failed getting file information for file ID", "file_id", fileId)
				response.Error = err.Error()
				return response
			} else if !fileSearchResp.Success {
				logging.LogError(errors.New(fileSearchResp.Error), "file search RPC call returned an error", "file_id", fileId)
				response.Error = fileSearchResp.Error
				return response
			}

			scriptName, _ := taskData.Args.GetStringArg("script")

			originalFileName := fileSearchResp.Files[0].Filename
			if response.DisplayParams == nil {
				response.DisplayParams = new(string)
			}

			*response.DisplayParams = fmt.Sprintf("-bundle %s -script %s", originalFileName, scriptName)

			fileContentResponse, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
				AgentFileID: fileId,
			})
			if err != nil {
				logging.LogError(err, "failed getting bundle content for file ID", "file_id", fileId)
				response.Error = err.Error()
				return response
			} else if !fileContentResponse.Success {
				logging.LogError(errors.New(fileContentResponse.Error), "file get content response for bundle returned an error", "file_id", fileId)
				response.Error = fileContentResponse.Error
				return response
			}

			fileExtractor, err := extract.NewBundleExtractor(fileContentResponse.Content)
			if err != nil {
				logging.LogError(err, "could not create bundle extractor")
				response.Error = fmt.Sprintf("could not create bundle extractor %s", err.Error())
				return response
			}

			bundleFiles, err := fileExtractor.ListFilePaths()
			if err != nil {
				logging.LogError(err, "could not list files in bundle")
				response.Error = fmt.Sprintf("could not list files in bundle %s", err.Error())
				return response
			}

			if !slices.Contains(bundleFiles, scriptName) {
				response.Error = fmt.Sprintf("script '%s' not found in bundle", scriptName)
				return response
			}

			extractPath := path.Join(config.GetForgeScriptRuntimePath(), fileId)

			if err := os.MkdirAll(extractPath, 0700); err != nil {
				logging.LogError(err, "could not create path for extracted bundle")
				response.Error = fmt.Sprintf("could not create path for extracted bundle %s", err.Error())
				return response
			}

			extractDir, err := os.OpenRoot(extractPath)
			if err != nil {
				logging.LogError(err, "could not open extract path")
				response.Error = fmt.Sprintf("could not open path for extracting the bundle %s", err.Error())
				return response
			}

			if err := fileExtractor.ExtractTo(extractDir); err != nil {
				logging.LogError(err, "could not extract bundle")
				response.Error = fmt.Sprintf("could not extract bundle %s", err.Error())
				return response
			}

			scriptFullPath := path.Join(extractPath, scriptName)
			if scriptStat, err := os.Stat(scriptFullPath); err != nil {
				logging.LogError(err, fmt.Sprintf("could not find %s in extracted bundle", scriptFullPath))
				response.Error = fmt.Sprintf("could not find %s in extracted bundle", scriptFullPath)
				return response
			} else if scriptStat.IsDir() {
				response.Error = "specified path for bundle script is a directory"
				return response
			}

			registered, err := python.RunScript(scriptFullPath, taskData.Callback.ID, taskData.Task.ID, taskData.Task.OperatorUsername)
			if err != nil {
				response.Error = fmt.Sprintf("failed loading script %s", err.Error())
				mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
					TaskID:   taskData.Task.ID,
					Response: []byte(err.Error()),
				})
				return response
			}

			outputResponse := fmt.Sprintf("Extracted bundle to %s", extractPath)
			for _, aliasName := range registered {
				outputResponse += fmt.Sprintf("\nRegistered alias %s", aliasName)
			}

			mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
				TaskID:   taskData.Task.ID,
				Response: []byte(outputResponse),
			})

			rabbitmq.SyncPayloadData(&payloadDefinition.Name, false)

			response.Success = true
			return response
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if len(input) > 0 {
				return args.LoadArgsFromJSONString(input)
			}
			return nil
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
	})
}
