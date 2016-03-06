package extractor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//Extract ...
func Extract(source, dest string) error {

	//Chk the destination path is present and create if not
	err := os.MkdirAll(dest, os.ModeDir|os.ModePerm)
	if err != nil {
		return fmt.Errorf("Destination path: , %v", err)
	}

	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Open File: %v", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("File info: %v", err)
	}

	switch {
	case strings.HasSuffix(fi.Name(), ".gz"):

		ar, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("Gzip reader: %v", err)
		}
		defer ar.Close()

		if strings.HasSuffix(ar.Header.Name, ".tar") { //|| ar.Header.Name == "" {

			tr := tar.NewReader(ar)
			return tarParse(tr, dest)
		}

		path := filepath.Join(dest, strings.TrimSuffix(fi.Name(), ".gz"))
		return toFile(ar, path)

	case strings.HasSuffix(fi.Name(), ".tar"):

		tr := tar.NewReader(f)

		return tarParse(tr, dest)

	default:
		return fmt.Errorf("Extract - Unknown file type: %v", fi.Name())
	}
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

	return toFile(gzReader, path)
}

//toFile ...
func toFile(r io.Reader, dest string) error {

	w, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Open file error: %v", err)
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("Copy file error: %v", err)
	}

	return nil
}

//tarParse ...
func tarParse(tr *tar.Reader, dest string) error {

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fi := hdr.FileInfo()
		//path := filepath.Join(dest, fi.Name())
		path := filepath.Join(dest, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:

			//if hdr.Name != fi.Name()+"/" {

			if err = os.MkdirAll(path, fi.Mode()); err != nil {
				return err
			}
			//}

			continue

		case tar.TypeReg:

			if !strings.HasPrefix(fi.Name(), ".") {

				err = gzDecompress(tr, hdr, path)
				if err != nil {
					return err
				}
			}

		}

	}
	return nil
}

//ExtractTarGz extract a source .tar.gz file to a dest
func ExtractTarGz(source, dest string) error {
	cmd := exec.Command("tar", "-xf", source, "-C", dest)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Extract error:\n%v", cmd.Stderr)
	}

	return nil
}
