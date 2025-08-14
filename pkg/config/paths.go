package config

import (
	"log"
	"os"
	"path"

	"github.com/MythicAgents/forgescript/pkg/xdg"
)

const forgePathSuffix = "forgescript"

var forgeRuntimePath = ""
var forgeCachePath = ""

func createNeededDir(val string) {
	if err := os.MkdirAll(val, 0700); err != nil {
		log.Fatalf("Could not create directory %s: %s", val, err.Error())
	}
}

func appendForgeSuffix(val string) string {
	_, fileName := path.Split(val)
	if fileName != forgePathSuffix {
		return path.Join(val, forgePathSuffix)
	}

	return val
}

func getOrSetForgePath(pathVal *string, defaultPath string) string {
	if pathVal == nil {
		return appendForgeSuffix(defaultPath)
	} else {
		if len(*pathVal) == 0 {
			*pathVal = appendForgeSuffix(defaultPath)
		}

		return *pathVal
	}
}

func SetForgeScriptRuntimePath(val string) {
	forgeRuntimePath = appendForgeSuffix(val)
}

func GetForgeScriptRuntimePath() string {
	return getOrSetForgePath(&forgeRuntimePath, xdg.RuntimeDir())
}

func GetForgeScriptBundleExtractPath() string {
	return path.Join(GetForgeScriptRuntimePath(), "extracted")
}

func GetAndCreateForgeScriptRuntimePath() string {
	runtimePath := GetForgeScriptRuntimePath()
	createNeededDir(runtimePath)
	return runtimePath
}

func SetForgeScriptCachePath(val string) {
	forgeCachePath = appendForgeSuffix(val)
}

func GetForgeScriptCachePath() string {
	return getOrSetForgePath(&forgeCachePath, xdg.CacheDir())
}

func GetAndCreateForgeScriptCachePath() string {
	cachePath := GetForgeScriptCachePath()
	createNeededDir(cachePath)
	return cachePath
}
