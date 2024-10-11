package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const monitorMetadataFile = "metadata.yaml"

// metricMetadata contains a metric's metadata.
type metricMetadata struct {
	Alias       string  `json:"alias,omitempty"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Group       *string `json:"group"`
	Default     bool    `json:"default" default:"false"`
}

// propMetadata contains a property's metadata.
type propMetadata struct {
	Dimension   string `json:"dimension"`
	Description string `json:"description"`
}

// groupMetadata contains a group's metadata.
type groupMetadata struct {
	Description string   `json:"description"`
	Metrics     []string `json:"metrics"`
}

// monitorMetadata contains a monitor's metadata.
type monitorMetadata struct {
	MonitorType  string                    `json:"monitorType" yaml:"monitorType"`
	SendAll      bool                      `json:"sendAll" yaml:"sendAll"`
	SendUnknown  bool                      `json:"sendUnknown" yaml:"sendUnknown"`
	NoneIncluded bool                      `json:"noneIncluded" yaml:"noneIncluded"`
	Dimensions   map[string]dimMetadata    `json:"dimensions"`
	Doc          string                    `json:"doc"`
	Groups       map[string]*groupMetadata `json:"groups"`
	Metrics      map[string]metricMetadata `json:"metrics"`
	Properties   map[string]propMetadata   `json:"properties"`
}

// packageMetadata describes a package directory that may have one or more monitors.
type packageMetadata struct {
	// Common is a section to allow multiple monitors to place shared data.
	Common map[string]interface{}
	// PackageDir is the directory to output the generated code if not the same directory as the metadata.yaml.
	PackageDir string `json:"packageDir" yaml:"packageDir"`
	Monitors   []monitorMetadata
	// Name of the package in go. If not set defaults to the directory name.
	GoPackage *string `json:"goPackage" yaml:"goPackage"`
	// Filesystem path to the package directory.
	PackagePath string `json:"-" yaml:"-"`
	Path        string `json:"-" yaml:"-"`
}

// dimMetadata contains a dimension's metadata.
type dimMetadata struct {
	Description string `json:"description"`
}

// collectMetadata loads metadata for all monitors located in root as well as any subdirectories.
func collectMetadata(root string) ([]packageMetadata, error) {
	var packages []packageMetadata

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != monitorMetadataFile {
			return nil
		}

		var pkg packageMetadata

		if bytes, err := os.ReadFile(path); err != nil {
			return fmt.Errorf("unable to read metadata file %s: %w", path, err)
		} else if err := yaml.UnmarshalStrict(bytes, &pkg); err != nil {
			return fmt.Errorf("unable to unmarshal file %s: %w", path, err)
		}

		pkg.PackagePath = filepath.Dir(path)
		pkg.Path = path

		packages = append(packages, pkg)

		return nil
	}); err != nil {
		return nil, err
	}

	return packages, nil
}
