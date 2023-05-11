package cgroups

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func withOpenFile(path string, handler func(fd *os.File)) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer fd.Close()
	handler(fd)

	return nil
}

func parseSingleInt(fileReader io.Reader) (int64, error) {
	bytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return 0, err
	}

	out, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func walkControllerHierarchy(root string, cb func(cgroupName string, files []string)) {
	dirs := []string{root}

	var currentDir string
	for len(dirs) > 0 {
		currentDir, dirs = dirs[0], dirs[1:]
		f, err := os.Open(currentDir)
		if err != nil {
			continue
		}
		defer f.Close()

		contents, err := f.Readdir(0)
		if err != nil {
			continue
		}

		var files []string

		for i := range contents {
			name := contents[i].Name()
			if name == "." || name == ".." {
				continue
			}
			if contents[i].IsDir() {
				dirs = append(dirs, filepath.Join(currentDir, name))
			} else {
				files = append(files, filepath.Join(currentDir, name))
			}
		}

		cgroupName, err := filepath.Rel(root, currentDir)
		if err != nil {
			continue
		}

		cb("/"+strings.TrimPrefix(strings.TrimPrefix(cgroupName, "."), "/"), files)
	}
}
