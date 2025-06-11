package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/splunk/splunk-technical-addon/internal/modularinput"
)

func main() {
	os.Exit(run())
}

func run() int {
	config := GetDefaultSplunkTAOtelLinuxAutoinstrumentationModularInputs()

	schemeFlag := flag.Bool("scheme", false, "Print the scheme and exit")
	validateFlag := flag.Bool("validate-arguments", false, "Validate the arguments and exit")
	flag.Parse()
	if *schemeFlag {
		return 0
	}
	defaultLogFilepath, err := modularinput.GetDefaultLogFilePath(config.SchemaName)
	if err != nil {
		panic(err)
	}
	logCloseFunc, err := modularinput.SetupAddonLogger(defaultLogFilepath)
	if err != nil {
		panic(err)
	}
	defer logCloseFunc()

	// Create a new modular input processor with the embedded configuration
	mip := modularinput.NewModinputProcessor(config.SchemaName, config.ModularInputs)

	xmlInput, err := modularinput.ReadXML(os.Stdin)
	if err != nil {
		panic(err)
	}
	err = mip.ProcessXML(xmlInput)

	if *validateFlag {
		if err != nil {
			return -1
		} else {
			return 0
		}
	}
	if err != nil {
		log.Errorf("Error parsing modinput: %+v", err)
		panic(err)
	}
	modInputs := GetSplunkTAOtelLinuxAutoinstrumentationModularInputs(mip)
	err = Run(modInputs)
	if err != nil {
		log.Errorf("error running splunk linux autoinstrumentation addon: %+v", err)
		panic(err)
	}
	log.Println("Successfully autoinstrumented opentelemetry via the splunk addon for such")
	// set up trap
	return 0
}

// Helper function to check if a string exists in a file
func grepFile(search string, filepath string) (bool, error) {
	file, err := os.Open(filepath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			return true, nil
		}
	}
	return false, nil
}

func Run(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	lowerModInput := strings.ToLower(modInputs.Remove.Value)
	if "false" == lowerModInput {
		return Instrument(modInputs)
	}
	if "true" == lowerModInput {
		return Uninstrument(modInputs)
	}

	return fmt.Errorf("unknown value for 'remove' modular input, expected (true|false) given %v", strings.ToLower(modInputs.Remove.Value))
}

func Instrument(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if err := InstrumentJava(modInputs); err != nil {
		return err
	}
	if err := AutoinstrumentLdPreload(modInputs); err != nil {
		return err
	}
	return nil
}

func AutoinstrumentLdPreload(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	found, err := grepFile(modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	if err != nil {
		return err
	}
	if !found {
		// todo check modinputs for backup of file
		if err = BackupFile(modInputs.AutoinstrumentationPreloadPath.Value); err != nil {
			return err
		}
		f, err2 := os.OpenFile(modInputs.AutoinstrumentationPreloadPath.Value, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err2 != nil {
			return fmt.Errorf("Error opening %s: %w\n", modInputs.AutoinstrumentationPreloadPath.Value, err2)
		}
		defer f.Close()
		if _, err2 = f.WriteString(modInputs.AutoinstrumentationPath.Value + "\n"); err2 != nil {
			return fmt.Errorf("Error writing to %s: %w\n", modInputs.AutoinstrumentationPreloadPath.Value, err2)
		}
		log.Printf("Successfully autoinstrumented preload at %v with %v \n", modInputs.AutoinstrumentationPreloadPath.Value, modInputs.AutoinstrumentationPath.Value)
	} else {
		log.Printf("Preload already autoinstrumented with %v at %v  \n", modInputs.AutoinstrumentationPath.Value, modInputs.AutoinstrumentationPreloadPath.Value)
	}
	return nil
}

func Uninstrument(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	return fmt.Errorf("uninstrument modular inputs not yet implemented")
}

func BackupFile(currPath string) error {
	if _, err := os.Stat(currPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	dir := filepath.Dir(currPath)
	ogFilename := filepath.Base(currPath)
	newPathName := filepath.Join(dir, fmt.Sprintf("%s.%v", ogFilename, time.Now().UnixNano()))
	if _, err := fileutils.CopyFile(currPath, newPathName); err != nil {
		return err
	}
	return nil
}
