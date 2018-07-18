package util

import (
	log "github.com/Sirupsen/logrus"
	"strings"
)

// Max size for a log's entry tag
const maxTagSize = 12

/*
Builds a tag for the log entries with indentation in order to align the log entries.
*/
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

func LogFormatter(disableColors bool, disableTimestamp bool) log.Formatter {
	logOutputFormatter := &log.TextFormatter{}
	logOutputFormatter.DisableColors = disableColors
	logOutputFormatter.DisableTimestamp = disableTimestamp
	return logOutputFormatter
}

func LogLevel(logLevel string) log.Level {
	switch logLevel {
	case "info":
		return log.InfoLevel
	case "debug":
		return log.DebugLevel
	case "warning":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	case "panic":
		return log.PanicLevel
	}
	return log.DebugLevel
}
