package expUtils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func CatIPPort(ip string, port int) string {
	res := ip + ":" + strconv.Itoa(port)
	return res
}

func CompressDirToTarGz(dir, outputDir string) error {
	// Get the base name of the directory to use as the root directory in the tar archive
	rootDirName := filepath.Base(dir)

	// Create a file to store the .tar.gz archive
	tarGzFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s.tar.gz", rootDirName)))
	if err != nil {
		return fmt.Errorf("error creating tar.gz file: %v", err)
	}
	defer tarGzFile.Close()

	// Create a new gzip writer
	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	// Create a new tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk the directory and add files to the tar archive
	err = filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the tar header for files and directories
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update the header name to reflect the relative path in the archive,
		// including the root directory name
		relativePath, err := filepath.Rel(dir, filePath)
		if err != nil {
			return err
		}
		header.Name = filepath.Join(rootDirName, relativePath)

		// Write the header to the tar writer
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		// If it's a directory, we don't need to copy any content (empty directory)
		if info.IsDir() {
			return nil
		}

		// Copy the file content to the tar archive
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
	if err != nil {
		return fmt.Errorf("error walking through directory: %v", err)
	}

	return nil
}
