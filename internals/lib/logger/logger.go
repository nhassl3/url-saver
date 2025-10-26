package logger

import (
	"log/slog"
	"os"

	"github.com/nhassl3/url-saver/internals/lib/logger/handler/slogpretty"
)

func MustLoad(envLevel uint8) *slog.Logger {
	var level slog.Level
	switch envLevel {
	case 1:
		level = slog.LevelDebug
	case 2:
		level = slog.LevelDebug
	case 3:
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: level,
		},
	}

	return slog.New(opts.NewPrettyLogger(os.Stdout))
}
