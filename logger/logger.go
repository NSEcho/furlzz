package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	infoColor  = color.New(color.FgGreen)
	errorColor = color.New(color.FgRed)
)

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, infoColor.Sprintf("%s ", "⚡"), 0),
		errorLogger: log.New(os.Stderr, errorColor.Sprintf("%s ", "❗️"), 0),
		lock:        &sync.Mutex{},
	}
}

type Logger struct {
	previousLineLength int
	infoLogger         *log.Logger
	errorLogger        *log.Logger
	lock               *sync.Mutex
}

func (l *Logger) Infof(format string, args ...any) {
	l.infoLogger.Printf(format, args...)
}

func (l *Logger) Scriptf(format string, args ...any) {
	l.lock.Lock()
	defer l.lock.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("\r%s", strings.Repeat(" ", l.previousLineLength))
	fmt.Printf("\r%s", msg)
	l.previousLineLength = len(msg)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.errorLogger.Printf(format, args...)
}
