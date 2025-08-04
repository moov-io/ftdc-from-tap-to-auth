package issuer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	cardpersonalizer "github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/client"
	"github.com/moov-io/ftdc-from-tap-to-auth/internal/middleware"
	issuer8583 "github.com/moov-io/ftdc-from-tap-to-auth/issuer/iso8583"
)

// App is the main application, it contains all the components of the issuer service
// and is responsible for starting and stopping them.
type App struct {
	srv               *http.Server
	wg                *sync.WaitGroup
	Addr              string
	ISO8583ServerAddr string
	logger            *slog.Logger
	iso8583Server     io.Closer
	config            *Config
}

func NewApp(logger *slog.Logger, config *Config) *App {
	logger = logger.With(slog.String("app", "issuer"))

	if config == nil {
		config = DefaultConfig()
	}

	return &App{
		wg:     &sync.WaitGroup{},
		logger: logger,
		config: config,
	}
}

func (a *App) Start() error {
	a.logger.Info("starting app...")

	// setup the issuer
	router := chi.NewRouter()
	router.Use(middleware.NewStructuredLogger(a.logger))
	repository := NewRepository()

	err := repository.LoadFromFile()
	if err != nil {
		return fmt.Errorf("loading repository from file: %w", err)
	}

	cp := cardpersonalizer.New(a.config.CardPersonalizerURL)
	iss := NewService(a.logger, repository, cp)

	iso8583Server := issuer8583.NewServer(a.logger, a.config.ISO8583Addr, iss)
	err = iso8583Server.Start()
	if err != nil {
		return fmt.Errorf("starting iso8583 server: %w", err)
	}
	a.ISO8583ServerAddr = iso8583Server.Addr
	a.iso8583Server = iso8583Server

	api := NewAPI(a.logger, iss)
	api.AppendRoutes(router)

	l, err := net.Listen("tcp", a.config.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listening tcp port: %w", err)
	}

	a.Addr = l.Addr().String()

	a.srv = &http.Server{
		Handler: router,
	}

	a.wg.Add(1)
	go func() {
		a.logger.Info("http server started", slog.String("addr", a.Addr))

		if err := a.srv.Serve(l); err != nil {
			if err != http.ErrServerClosed {
				a.logger.Error("starting http server", "err", err)
			}

			a.logger.Info("http server stopped")
		}

		repository.SaveToFile()
		a.wg.Done()
	}()

	return nil
}

func (a *App) Shutdown() {
	a.logger.Info("shutting down app...")

	a.srv.Shutdown(context.Background())

	err := a.iso8583Server.Close()
	if err != nil {
		a.logger.Error("closing iso8583 server", "err", err)
	}

	a.wg.Wait()

	a.logger.Info("app stopped")
}
