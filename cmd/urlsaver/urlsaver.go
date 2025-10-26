package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nhassl3/url-saver/internals/app"
	"github.com/nhassl3/url-saver/internals/config"
	"github.com/nhassl3/url-saver/internals/lib/logger"
)

var (
	cfg *config.Config
	log *slog.Logger
)

func init() {
	cfg = config.MustLoad()

	log = logger.MustLoad(cfg.EnvLevel)
	slog.SetDefault(log)
}

func main() {
	log.Info("Starting URL Saver", slog.Int("port", cfg.GRPC.Port))

	// loading configuration for this service and other too
	application := app.NewApp(
		log,
		cfg.GRPC.Port,
		cfg.HTTP.UrlShortener.MaxRetires,
		cfg.StoragePath,
		cfg.HTTP.UrlShortener.BaseUrl,
		cfg.HTTP.UrlShortener.Timeout,
	)

	go application.GRPCServer.MustStart()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	log.Info("URL Server stopped", slog.String("signal", (<-sig).String()))
}
