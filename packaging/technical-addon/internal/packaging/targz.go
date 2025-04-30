// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packaging

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
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

		if err = tarWriter.WriteHeader(header); err != nil {
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

// ExtractAddon extracts a .tar.gz file from sourcePath to destinationPath
func ExtractAddon(sourcePath, destinationPath string) error {
	// Open the tar.gz file
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		// Create the full path for the file
		target := filepath.Join(destinationPath, header.Name)

		// Check for illegal paths (path traversal attack prevention)
		if !isInDirectory(target, destinationPath) {
			return fmt.Errorf("illegal path traversal attempt: %s", header.Name)
		}

		// Handle different types of files
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directories with proper permissions
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create containing directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", target, err)
			}

			// Create the file
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}

			// Copy the file data
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}
			file.Close()

		default:
			// Skip other types of files (symlinks, devices, etc.)
			fmt.Printf("Skipping unsupported file type: %c for %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// isInDirectory checks if the path is inside the specified directory (prevents path traversal)
func isInDirectory(path, directory string) bool {
	// Convert to absolute paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return false
	}

	// Check if the path is within the directory
	return filepath.HasPrefix(absPath, absDir)
}
