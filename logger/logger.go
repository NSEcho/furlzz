package logger

import (
	"github.com/fatih/color"
	"log"
	"os"
)

var (
	infoColor  = color.New(color.FgGreen)
	errorColor = color.New(color.FgRed)
)

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, infoColor.Sprintf("%s ", "⚡"), 0),
		errorLogger: log.New(os.Stderr, errorColor.Sprintf("%s ", "❗️"), 0),
	}
}

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func (l *Logger) Infof(format string, args ...any) {
	l.infoLogger.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.errorLogger.Printf(format, args...)
}
