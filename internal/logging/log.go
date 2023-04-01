package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func SetupLogging(logDirPath string) {
	setupLogfile(logDirPath)
}

// setupLogfile creates a log file and calls the SetOutput function of the standard logger
func setupLogfile(logDirPath string) {
	logfileName := "geizhalsbot.log"
	logDir := filepath.Clean(logDirPath)

	mkdirErr := os.MkdirAll(logDir, 0o666)
	if mkdirErr != nil {
		log.Fatal("error creating folders:", mkdirErr)
	}

	logFilePath := filepath.Join(logDir, logfileName)

	logFile, openErr := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if openErr != nil {
		log.Fatalf("error opening file: %v", openErr)
	}

	// create a MultiWriter which can write to multiple destinations. In this case stdout and the given log file.
	w := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(w)
}
