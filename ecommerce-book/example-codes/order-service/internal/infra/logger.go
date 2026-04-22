package infra

import "fmt"

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(layer string, format string, args ...any) {
	fmt.Printf("[%s] %s\n", layer, fmt.Sprintf(format, args...))
}
