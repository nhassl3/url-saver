package slogpretty

import (
	"log"
	"log/slog"
)

type PrettyLogger struct {
	slog.Handler
	l     *log.Logger
	attrs slog.Attr
}

func NewPrettyLogger(l *log.Logger, attrs slog.Attr) *PrettyLogger {
	return &PrettyLogger{
		l:     l,
		attrs: attrs,
	}
}
