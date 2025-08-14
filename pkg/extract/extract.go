package extract

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
)


const (
	mimeTypeTar = "application/x-tar"
	mimeTypeXz = "application/x-xz"
	mimeTypeBzip2 = "application/x-bzip2"
	mimeTypeGzip = "application/gzip"
	mimeTypeZip = "application/zip"
)


type BundleExtractor interface {
	ExtractTo(*os.Root) error
	ListFilePaths() ([]string, error)
}

func decompressXz(data []byte) ([]byte, error) {
	return []byte{}, errors.New("xz decompression not implemented")
}

func decompressBzip2(data []byte) ([]byte, error) {
	return io.ReadAll(bzip2.NewReader(bytes.NewReader(data)))
}

func decompressGzip(data []byte) ([]byte, error) {
	rd, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}

	return io.ReadAll(rd)
}


func extractorForMimeType(mimeType string, bundle []byte) (BundleExtractor, error) {
	if mimeType == mimeTypeZip {
		return NewZipExtractor(bundle), nil
	}

	decompressMimeTypes := map[string]func([]byte) ([]byte, error){
		mimeTypeXz: decompressXz,
		mimeTypeBzip2: decompressBzip2,
		mimeTypeGzip: decompressGzip,
	}

	if decompressor, ok := decompressMimeTypes[mimeType]; ok {
		bundle, err := decompressor(bundle)
		if err != nil {
			return nil, err
		}

		mimeType, err = detectMimeType(bundle)
		if err != nil {
			return nil, err
		}

		return extractorForMimeType(mimeType, bundle)
	} else if mimeType == mimeTypeTar {
		return NewTarExtractor(bundle), nil
	}

	return nil, fmt.Errorf("unknown file type %s", mimeType)
}

func NewBundleExtractor(bundle []byte) (BundleExtractor, error) {
	mimeType, err := detectMimeType(bundle)
	if err != nil {
		return nil, err
	}

	return extractorForMimeType(mimeType, bundle)
}

