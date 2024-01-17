package logger

import (
	"log"
	"os"
)

var (
	logger  = log.New(os.Stdout, "", log.LstdFlags)
	verbose = false
)

func SetVerbose(val bool) {
	verbose = val
}

func Debug(message ...interface{}) {
	if verbose {
		logger.Println(message...)
	}
}

func Info(message ...interface{}) {
	logger.Println(message...)
}

func Infof(format string, message ...interface{}) {
	logger.Printf(format, message...)
}
