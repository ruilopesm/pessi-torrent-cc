package logger

import (
	"fmt"
	"sync"
)

type Logger interface {
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
}

type logger struct {
	sync.Mutex
}

func NewSimpleLogger() Logger {
	return &logger{}
}

func SetLogger(l Logger) {
	CurrentLogger = l
}

var CurrentLogger = NewSimpleLogger()

func Info(message string, args ...any) {
	CurrentLogger.Info(message, args...)
}

func Warn(message string, args ...any) {
	CurrentLogger.Warn(message, args...)
}

func Error(message string, args ...any) {
	CurrentLogger.Error(message, args...)
}

func (l *logger) Info(message string, args ...any) {
	l.Lock()
	defer l.Unlock()
	message = fmt.Sprintf(message, args...)
	fmt.Printf("%s\n", message)
}

func (l *logger) Warn(message string, args ...any) {
	l.Lock()
	defer l.Unlock()
	message = fmt.Sprintf(message, args...)
	fmt.Printf("%s\n", message)
}

func (l *logger) Error(message string, args ...any) {
	l.Lock()
	defer l.Unlock()
	message = fmt.Sprintf(message, args...)
	fmt.Printf("%s\n", message)
}
