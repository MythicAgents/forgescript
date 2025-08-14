package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MythicAgents/forgescript/pkg/agentfunctions"
	"github.com/MythicAgents/forgescript/pkg/config"
	_ "github.com/MythicAgents/forgescript/pkg/pymodule"
	"github.com/MythicAgents/forgescript/pkg/python"
	"github.com/MythicMeta/MythicContainer"
)

func main() {
	subcommand := ""
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		subcommand = os.Args[1]
	}

	runtimeDir := flag.String("runtime-dir", "", "Set the runtime path")
	flag.Parse()
	if runtimeDir != nil && len(*runtimeDir) > 0 {
		config.SetForgeScriptRuntimePath(*runtimeDir)
	}

	if subcommand == "clean" {
		exitCode := 0

		configRuntimeDir := config.GetForgeScriptRuntimePath()
		fmt.Printf("Removing %s\n", configRuntimeDir)
		if err := os.RemoveAll(configRuntimeDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed removing %s (%s)", configRuntimeDir, err.Error())
			exitCode = 1
		}

		configCacheDir := config.GetForgeScriptCachePath()
		fmt.Printf("Removing %s\n", configCacheDir)
		if err := os.RemoveAll(configCacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed removing %s (%s)", configCacheDir, err.Error())
			exitCode = 1
		}

		os.Exit(exitCode)
	}

	agentfunctions.Initialize()

	go func() {
		MythicContainer.StartAndRunForever([]MythicContainer.MythicServices{
			MythicContainer.MythicServicePayload,
		})
	}()

	python.StartExecutorLoop()
}
