package extract

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

type tarEntry struct {
	hdr *tar.Header
}

func (e tarEntry) FilePath() string {
	return e.hdr.Name
}

func (e tarEntry) FileInfo() fs.FileInfo {
	return e.FileInfo()
}

type tarExtractor struct {
	buffer []byte
}

func NewTarExtractor(data []byte) tarExtractor {
	return tarExtractor{buffer: data}
}


func (extractor tarExtractor) ListFilePaths() ([]string, error) {
	rd := tar.NewReader(bytes.NewReader(extractor.buffer))

	filePaths := []string{}
	for {
		hdr, err := rd.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, hdr.Name)
	}

	return filePaths, nil
}


func (extractor tarExtractor) ExtractTo(root *os.Root) error {
	rd := tar.NewReader(bytes.NewReader(extractor.buffer))

	for {
		hdr, err := rd.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		pathDir := path.Dir(hdr.Name)
		if pathDir != "." {
			if _, err := root.Stat(pathDir); err != nil {
				if err := root.Mkdir(pathDir, 0700); err != nil {
					return err
				}
			}
		}

		outFile, err := root.OpenFile(hdr.Name, os.O_CREATE | os.O_RDWR, hdr.FileInfo().Mode())
		if err != nil {
			return err
		}

		if w, err := io.Copy(outFile, rd); err != nil {
			return err
		} else if w != hdr.Size {
			return fmt.Errorf("file write for %s truncated", hdr.Name)
		}

		outFile.Sync()
	}

	return nil
}
