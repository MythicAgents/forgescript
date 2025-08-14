package versioninfo

import (
	"runtime/debug"
)

var embedVersion = ""
var embedGitRevision = ""

// Returns the current version of the main module if found
// This will use the latest 'v' prefixed SemVer compatible Git tag (ex. v0.1.0) discovered during `go build`
func ModuleVersion() string {
	if len(embedVersion) > 0 {
		return embedVersion
	}

	if buildinfo, ok := debug.ReadBuildInfo(); ok && buildinfo != nil {
		return buildinfo.Main.Version
	}

	return ""
}

// Returns the current Git commit hash of the main module if found
func GitRevision() string {
	if len(embedGitRevision) > 0 {
		return embedGitRevision
	}

	// Handy feature for getting VCS info from Go: https://icinga.com/blog/embedding-git-commit-information-in-go-binaries/
	if buildInfo, ok := debug.ReadBuildInfo(); ok && buildInfo != nil {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}

// Returns the Mythic version required for running this service
func RequiredMythicVersion() string {
	return "3.3.0+"
}
