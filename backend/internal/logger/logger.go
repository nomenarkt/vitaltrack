package logger

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Logger defines structured logging methods.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, kv ...any)
}

// StdLogger logs to the standard library logger.
type StdLogger struct{}

// NewStdLogger creates a StdLogger.
func NewStdLogger() *StdLogger { return &StdLogger{} }

func (l *StdLogger) Info(ctx context.Context, msg string, kv ...any) {
	log.Println(format(msg, kv...))
}

func (l *StdLogger) Error(ctx context.Context, msg string, kv ...any) {
	log.Println(format(msg, kv...))
}

func format(msg string, kv ...any) string {
	if len(kv) == 0 {
		return msg
	}
	pairs := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprint(kv[i])
		val := ""
		if i+1 < len(kv) {
			val = fmt.Sprint(kv[i+1])
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, val))
	}
	return msg + " " + strings.Join(pairs, " ")
}
