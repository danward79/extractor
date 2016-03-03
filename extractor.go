package extractor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//ParseTarReader loops through a tar.Reader and creates the file or folder.
func ParseTarReader(tarReader *tar.Reader, destination string) error {

	fmt.Println("ParseTarReader")

	for {
		fmt.Println("For")
		hdr, err := tarReader.Next()
		fmt.Println(hdr)
		if err == io.EOF {
			fmt.Println("io.EOF")
			break
		} else if err != nil {
			return fmt.Errorf("Tar Header Error: %v", err)
		}

		fmt.Println(hdr.Name)
		path := filepath.Join(destination, hdr.Name)
		info := hdr.FileInfo()

		switch {

		case info.IsDir():
			fmt.Println("IsDir")
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return fmt.Errorf("Error generating folder: %v", err)
			}
			continue

		case strings.HasSuffix(info.Name(), ".gz") && !strings.HasPrefix(info.Name(), "."):
			fmt.Println(".gz")
			err = gzDecompress(tarReader, hdr, path)
			if err != nil {
				return err
			}

		case strings.HasSuffix(info.Name(), ".tar"):
			fmt.Println(".tar")
			err := copyReaderFile(tarReader, hdr, path)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

//gzDecompress handler to expand .gz files
func gzDecompress(r io.Reader, hdr *tar.Header, path string) error {

	path = strings.TrimSuffix(path, ".gz")

	info := hdr.FileInfo()

	content := make([]byte, info.Size())

	read, err := io.ReadFull(r, content)
	if err != nil {
		return fmt.Errorf("Read Error: %v ", err)
	}

	if int64(read) != info.Size() {
		return fmt.Errorf("Error: Size read error")
	}

	b := bytes.NewReader(content)

	gzReader, err := gzip.NewReader(b)
	defer gzReader.Close()
	if err != nil {
		return fmt.Errorf("Gzip reader: %v", err)
	}

	return copyReaderFile(gzReader, hdr, path)

}

//copyReaderFile handles file generation and copying reader content to file.
func copyReaderFile(r io.Reader, hdr *tar.Header, path string) error {

	info := hdr.FileInfo()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return fmt.Errorf("Open file error: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("Copy file error: %v", err)
	}

	return nil
}
