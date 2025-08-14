package extract

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
)

type zipExtractor struct {
	buffer []byte
}

func NewZipExtractor(data []byte) zipExtractor {
	return zipExtractor{buffer: data}
}

func (extractor zipExtractor) ListFilePaths() ([]string, error) {
	byteRd := bytes.NewReader(extractor.buffer)

	rd, err := zip.NewReader(byteRd, byteRd.Size())
	if err != nil {
		return []string{}, err
	}

	filePaths := []string{}
	for _, file := range rd.File {
		if file.FileInfo().IsDir() {
			continue
		}

		filePaths = append(filePaths, file.Name)
	}

	return filePaths, nil
}

func (extractor zipExtractor) ExtractTo(root *os.Root) error {
	return errors.New("zip ExtractTo unimplemented")
}
