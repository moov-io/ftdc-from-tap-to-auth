package cardpersonalizer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/alovak/cardflow-playground/internal/middleware"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

type App struct {
	srv     *http.Server
	wg      *sync.WaitGroup
	Addr    string
	logger  *slog.Logger
	config  *Config
	service *Service
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewApp(logger *slog.Logger, config *Config) *App {
	logger = logger.With(slog.String("app", "card-personalizer"))

	if config == nil {
		config = DefaultConfig()
	}

	return &App{
		logger: logger,
		wg:     &sync.WaitGroup{},
		config: config,
	}
}

func (a *App) Start() error {
	a.logger.Info("starting app...")

	// setup the acquirer
	router := chi.NewRouter()
	router.Use(middleware.NewStructuredLogger(a.logger))

	cp, err := NewService(a.logger, a.config.CardReader)
	if err != nil {
		return fmt.Errorf("creating card personalizer service: %w", err)
	}

	a.service = cp
	a.ctx, a.cancel = context.WithCancel(context.Background())

	go cp.Run(a.ctx)
	api := NewAPI(a.logger, cp)
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
				a.logger.Error("Error starting acquirer http server", "err", err)
			}

			a.logger.Info("http server stopped")
		}

		a.wg.Done()
	}()

	return nil
}

func (a *App) Shutdown() {
	a.logger.Info("shutting down app...")

	// Stop the service
	a.cancel()

	a.srv.Shutdown(context.Background())

	a.wg.Wait()

	a.logger.Info("app stopped")
}
