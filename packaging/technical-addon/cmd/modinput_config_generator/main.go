package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"
)

func main() {
	sourceDir := flag.String("source-dir", "", "Source directory ($(SOURCE_DIR))")
	schemaName := flag.String("schema-name", "", "Modular input schema name")
	buildDir := flag.String("build-dir", "", "Build directory ($(BUILD_DIR))")
	flag.Parse()

	if sourceDir == nil || *sourceDir == "" ||
		schemaName == nil || *schemaName == "" ||
		buildDir == nil || *buildDir == "" {
		log.Fatal("source-dir, schema-name, and build-dir must be provided")
	}
	yamlPath := filepath.Join(*sourceDir, "pkg", strings.ToLower(*schemaName), "runner", "modular-inputs.yaml")
	config, err := loadYaml(yamlPath, *schemaName)
	if err != nil {
		log.Fatalf("Error loading modinput yaml in %s: %v", *sourceDir, err)
	}

	if err := generateModinputConfig(config, filepath.Dir(yamlPath)); err != nil {
		log.Fatalf("Error processing %v: %v\n", config, err)
	}
	if err := generateTaModInputConfs(config, filepath.Dir(filepath.Dir(yamlPath)), filepath.Join(*buildDir, *schemaName)); err != nil {
		log.Fatalf("Error processing %v: %v\n", config, err)
	}
}
