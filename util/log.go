package util

import (
	logger "github.com/Sirupsen/logrus"
	"strings"
)

var Log *logger.Logger = nil

func Init(logger *logger.Logger) {
	if Log == nil {
		Log = logger
	}
}

// Max size for a log's entry tag
const maxTagSize = 12

// Builds a tag for the log entries with indentation in order to align the log entries.
func LogTag(tag string) string {
	tag = "[" + tag + "]"
	sizeOfIndentation := maxTagSize - len(tag)
	if sizeOfIndentation <= 0 {
		tag = tag[:maxTagSize-1]
		sizeOfIndentation = maxTagSize - len(tag)
	}
	indentation := strings.Repeat(" ", sizeOfIndentation)
	return tag + indentation
}

func LogFormatter(disableColors bool, disableTimestamp bool) logger.Formatter {
	logOutputFormatter := &logger.TextFormatter{}
	logOutputFormatter.DisableColors = disableColors
	logOutputFormatter.DisableTimestamp = disableTimestamp
	return logOutputFormatter
}

func LogLevel(logLevel string) logger.Level {
	switch logLevel {
	case "info":
		return logger.InfoLevel
	case "debug":
		return logger.DebugLevel
	case "warning":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	case "fatal":
		return logger.FatalLevel
	case "panic":
		return logger.PanicLevel
	default:
		return logger.DebugLevel
	}
}
