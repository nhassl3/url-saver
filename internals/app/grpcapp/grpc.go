package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	urlshortener "github.com/nhassl3/url-saver/internals/clients/urlshortener/http"
	"github.com/nhassl3/url-saver/internals/domain/services/urlsaver"
	urlSavergrpc "github.com/nhassl3/url-saver/internals/grpc/urlsaver"
	"google.golang.org/grpc"
)

const opStart = "grpcapp.MustStart"

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log *slog.Logger,
	gRPCPort int,
	urlSaverObj *urlsaver.UrlSaver,
	urlShortenerClient *urlshortener.Client) *App {
	gRPCServer := grpc.NewServer()

	urlSavergrpc.Register(gRPCServer, urlSaverObj, urlShortenerClient)

	return &App{
		gRPCServer: gRPCServer,
		port:       gRPCPort,
		log:        log,
	}
}

func (app *App) MustStart() {
	log := app.log.With(slog.String("op", opStart), slog.Int("port", app.port))

	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", app.port))
	if err != nil {
		panic(fmt.Errorf("%s: %w", opStart, err))
	}

	log.Info("Server started", slog.String("address", l.Addr().String()))

	if err := app.gRPCServer.Serve(l); err != nil {
		panic(fmt.Errorf("%s: %w", opStart, err))
	}
}

func (app *App) Stop() {
	app.gRPCServer.GracefulStop()
}
