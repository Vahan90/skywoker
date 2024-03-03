package logger

import (
	"log"
	"os"
)

var (
	LogLevel int
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

func Debugf(format string, v ...interface{}) {
	if LogLevel <= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if LogLevel <= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

func Warningf(format string, v ...interface{}) {
	if LogLevel <= LogLevelWarning {
		log.Printf("[WARNING] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if LogLevel <= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
		os.Exit(1)
	}
}
