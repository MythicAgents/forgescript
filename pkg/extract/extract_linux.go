package extract

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func detectMimeType(data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "foregescript")
	if err != nil {
		log.Fatalf("failed creating temporary directory %s", err.Error())
	}
	defer os.Remove(tmpFile.Name())

	dataSize := min(len(data), 512)
	if w, err := tmpFile.Write(data[:dataSize]); err != nil {
		return "", fmt.Errorf("could not write temporary file %s", err.Error())
	} else if w != dataSize {
		return "", fmt.Errorf("temporary file write truncated. wrote = %d, expected = %d", w, dataSize)
	}

	cmdPath, err := exec.LookPath("file")
	if err != nil {
		return "", fmt.Errorf("could not find 'file' binary")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(cmdPath, "-b", "--mime-type", tmpFile.Name())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("could not run 'file' command to check mimetype %s", err.Error())
	}

	return strings.TrimSuffix(stdout.String(), "\n"), nil
}
