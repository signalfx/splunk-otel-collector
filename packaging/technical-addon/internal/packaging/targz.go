package packaging

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func PackageAddon(sourceDir, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	parent := filepath.Dir(sourceDir)

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(parent, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// Directory names need to end with path delim
		if info.IsDir() {
			header.Name += "/"
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil // directories only need the header
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(tarWriter, srcFile)
		return err
	})

	return err
}
