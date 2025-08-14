package xdg

import (
	"log"
	"os"
)


func RuntimeDir() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if len(runtimeDir) == 0 {
		runtimeDir = os.TempDir()
	}

	return runtimeDir
}


func CacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("Cache directory could not be found: %s\n", err.Error())
	}

	return userCacheDir
}
