package logger

import (
	"fmt"
	"log"
)

type Logger interface {
	Info(message string, args ...any)
	Warn(message string, args ...any)
}

type logger struct{}

func (l logger) Info(message string, args ...any) {
	message = fmt.Sprintf(message, args...)
	log.Printf("%s\n", message)
}

func (l logger) Warn(message string, args ...any) {
	message = fmt.Sprintf(message, args...)
	log.Printf("%s\n", message)
}

func NewSimpleLogger() Logger {
	return logger{}
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
