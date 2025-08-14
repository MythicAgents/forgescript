package extract

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testsRelDir string = "../../tests"
)

func TestDetectMimeType(t *testing.T) {
	mimeType, err := detectMimeType([]byte("Hello World"))
	assert.Nil(t, err, "detectMimeType returned an error")
	assert.Equal(t, "text/plain", mimeType)
}

func TestTarListFiles(t *testing.T) {
	var tarBuf bytes.Buffer
	tarw := tar.NewWriter(&tarBuf)
	tarw.WriteHeader(&tar.Header{
		Name: "file.py",
		Size: 0,
	})

	assert.Nil(t, tarw.Close(), "failed creating test tar file")

	tarx := NewTarExtractor(tarBuf.Bytes())
	fileList, err := tarx.ListFilePaths()
	assert.Nil(t, err, "ListFiles() returned an error")
	assert.Contains(t, fileList, "file.py", "tar extractor could not list test file")
}

func TestTarListFilesSubdir(t *testing.T) {
	var tarBuf bytes.Buffer
	tarw := tar.NewWriter(&tarBuf)
	tarw.WriteHeader(&tar.Header{
		Name: "subdir/file.py",
		Size: 0,
	})

	assert.Nil(t, tarw.Close(), "failed creating test tar file")

	tarx := NewTarExtractor(tarBuf.Bytes())
	fileList, err := tarx.ListFilePaths()
	assert.Nil(t, err, "ListFiles() returned an error")
	assert.Contains(t, fileList, "subdir/file.py", "tar extractor could not list test file")
}

func TestZipListFiles(t *testing.T) {
	var zipBuf bytes.Buffer
	zipw := zip.NewWriter(&zipBuf)
	_, err := zipw.Create("file.py")
	assert.Nil(t, err, "failed creating file.py in test zip file")
	assert.Nil(t, zipw.Close(), "failed creating test zip file")

	assert.NotEqual(t, 0, len(zipBuf.Bytes()), "zip writer did not write any data")

	zipx := NewZipExtractor(zipBuf.Bytes())
	fileList, err := zipx.ListFilePaths()
	assert.Nil(t, err, "ListFiles() returned an error")
	assert.Contains(t, fileList, "file.py", "zip extractor could not list test file")
}
