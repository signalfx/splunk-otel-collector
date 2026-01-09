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
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func GetBuildDir() string {
	var err error
	buildDir := os.Getenv("BUILD_DIR")
	if buildDir == "" {
		fmt.Println("$BUILD_DIR not set, searching for git root")
		buildDir, err = findGitRoot(".")
		if err != nil || buildDir == "" {
			fmt.Println("no git dir found, defaulting to cwd")
			buildDir, err = os.Getwd()
			if err != nil {
				fmt.Println("couldn't get cwd, defaulting to ./")
				buildDir = "."
			}
		}
		buildDir = filepath.Join(buildDir, "build")
	}
	return buildDir
}

func findGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		_, err := os.Stat(gitPath)
		if err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for .git: %w", err)
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", errors.New("no git repository found")
		}
		dir = parentDir
	}
}

func GetSourceDir() (string, error) {
	var err error
	sourceDir := os.Getenv("ADDONS_SOURCE_DIR")
	if sourceDir == "" {
		fmt.Println("ADDONS_SOURCE_DIR not set, searching for make root")
		sourceDir, err = findSourceRoot(".")
		if err != nil || sourceDir == "" {
			return "", errors.New("could not makefile to use as ADDONS_SOURCE_DIR and ADDONS_SOURCE_DIR was not specified as an env var")
		}
	}
	return sourceDir, nil
}

func findSourceRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		gitPath := filepath.Join(dir, "Makefile")
		_, err := os.Stat(gitPath)
		if err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for Makefile: %w", err)
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", errors.New("no Makefile found")
		}
		dir = parentDir
	}
}
