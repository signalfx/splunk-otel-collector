package modularinput

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func SetupAddonLogger(logFilePath string) (closer func(), err error) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error: failed to open log file %s: %v", logFilePath, err)
		return func() {}, fmt.Errorf("could not open log file %s", logFilePath)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return func() { _ = logFile.Close() }, nil
}

func GetDefaultLogFilePath(schemaName string) (string, error) {
	splunkHome, ok := os.LookupEnv("SPLUNK_HOME")
	if !ok {
		ex, err := os.Executable()
		if err != nil {
			// Don't attempt to further figure out path
			return "", err
		}
		// $SPLUNK_HOME/etc/apps/Sample_Addon/${platform}_x86_64/bin/executable
		splunkHome = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(ex)))))
	}

	return filepath.Join(splunkHome, "var", "log", "splunk", schemaName+".log"), nil
}
