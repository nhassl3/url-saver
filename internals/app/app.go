package app

import (
	"log/slog"
	"time"

	"github.com/nhassl3/url-saver/internals/app/grpcapp"
	urlshortener "github.com/nhassl3/url-saver/internals/clients/urlshortener/http"
	"github.com/nhassl3/url-saver/internals/domain/services/urlsaver"
	"github.com/nhassl3/url-saver/internals/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
}

func NewApp(
	log *slog.Logger,
	gRPCPort,
	maxRetries int,
	storagePath,
	baseUrlShortenerUrl string,
	timeout time.Duration,
) *App {
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	urlShortenerObject := urlshortener.NewClient(log, timeout, maxRetries, baseUrlShortenerUrl)

	urlSaverObj := urlsaver.NewUrlSaver(log, storage, storage, storage)

	return &App{
		GRPCServer: grpcapp.NewApp(log, gRPCPort, urlSaverObj, urlShortenerObject),
	}
}
