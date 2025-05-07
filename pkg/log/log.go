package log

import (
	"fmt"
	"io"
	baselog "log"
	"os"
	"runtime"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

var logLevelName = map[LogLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARN",
	ERROR:   "ERROR",
}

var (
	currentLevel = INFO
	logger       = baselog.New(os.Stdout, "", 0)
)

func SetLevel(level LogLevel) {
	currentLevel = level
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

func logMessage(level LogLevel, msg string) {
	if currentLevel > level {
		return
	}

	pc, file, line, ok := runtime.Caller(3)
	location := "unknown"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileParts := strings.Split(file, "/")
		location = fmt.Sprintf("%s:%d (%s)", fileParts[len(fileParts)-1], line, funcName)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logger.Printf("[%s] [%s] %s | %s\n", timestamp, logLevelName[level], msg, location)
}

func log(level LogLevel, v ...any) {
	logMessage(level, fmt.Sprint(v...))
}

func logf(level LogLevel, format string, v ...any) {
	logMessage(level, fmt.Sprintf(format, v...))
}

func Debug(v ...any) { log(DEBUG, v...) }
func Info(v ...any)  { log(INFO, v...) }
func Warn(v ...any)  { log(WARNING, v...) }
func Error(v ...any) { log(ERROR, v...) }

func Debugf(f string, v ...any) { logf(DEBUG, f, v...) }
func Infof(f string, v ...any)  { logf(INFO, f, v...) }
func Warnf(f string, v ...any)  { logf(WARNING, f, v...) }
func Errorf(f string, v ...any) { logf(ERROR, f, v...) }
