package extractor

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//Extract ...
func Extract(source, destination string) error {

	//Chk the destination path is present and creat if not
	err := os.MkdirAll(destination, os.ModeDir|os.ModePerm)
	if err != nil {
		return fmt.Errorf("Destination path: , %v", err)
	}

	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Open File: %v", err)
	}
	defer f.Close()

	info, _ := f.Stat()

	switch {
	case strings.HasSuffix(info.Name(), ".gz"):

		archiveReader, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("Gzip reader: %v", err)
		}
		defer archiveReader.Close()

		handleType(archiveReader.Header.Name, destination, info.Mode(), archiveReader)

	case strings.HasSuffix(info.Name(), ".tar"):

	default:
		log.Println("Extract - Unknown file type: ", info.Name())
	}

	return nil
}

func handleType(name, dest string, mode os.FileMode, r io.Reader) error {

	switch {
	case strings.HasSuffix(name, ".tar"):

		fmt.Println(name)

		handleTar(r, dest)

	case strings.HasSuffix(name, ".txt"):

		dest = dest + "/" + name

		toFile(r, dest, mode)

	}
	return nil
}

func toFile(r io.Reader, dest string, mode os.FileMode) error {

	file, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
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

func handleTar(r io.Reader, dest string) error {

	tarReader := tar.NewReader(r)
	fmt.Println(tarReader)
	for {

		hdr, err := tarReader.Next()
		fmt.Println(hdr)
		if err == io.EOF {
			fmt.Println("io.EOF")
			break
		} else if err != nil {
			return fmt.Errorf("Tar Header Error: %v", err)
		}

		fmt.Println(hdr.Name)
		path := filepath.Join(dest, hdr.Name)
		info := hdr.FileInfo()
		fmt.Println(path)
		fmt.Println(info)
	}

	return nil
}
