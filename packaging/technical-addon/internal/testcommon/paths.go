package testcommon

import (
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
			return "", fmt.Errorf("no git repository found")
		}
		dir = parentDir
	}
}

func GetSourceDir() (string, error) {
	var err error
	sourceDir := os.Getenv("SOURCE_DIR")
	if sourceDir == "" {
		fmt.Println("SOURCE_DIR not set, searching for make root")
		sourceDir, err = findSourceRoot(".")
		if err != nil || sourceDir == "" {
			return "", fmt.Errorf("could not makefile to use as SOURCE_DIR and SOURCE_DIR was not specified as an env var")
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
			return "", fmt.Errorf("no Makefile found")
		}
		dir = parentDir
	}
}
