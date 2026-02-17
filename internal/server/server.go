// Package server provides the HTTP server for awsim.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sivchari/awsim/internal/service"
)

// Config holds the server configuration.
type Config struct {
	Host     string
	Port     int
	LogLevel slog.Level
}

// DefaultConfig returns the default server configuration.
func DefaultConfig() Config {
	return Config{
		Host:     "0.0.0.0",
		Port:     4566,
		LogLevel: slog.LevelInfo,
	}
}

// Server is the main HTTP server for awsim.
type Server struct {
	config          Config
	router          *Router
	registry        *service.Registry
	jsonDispatcher  *JSONProtocolDispatcher
	queryDispatcher *QueryProtocolDispatcher
	cborDispatcher  *CBORProtocolDispatcher
	logger          *slog.Logger
	server          *http.Server
}

// New creates a new server with the given configuration.
// Services registered via init() are automatically loaded.
func New(config Config) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))

	registry := service.NewRegistry()
	router := NewRouter(logger)
	jsonDispatcher := NewJSONProtocolDispatcher()
	queryDispatcher := NewQueryProtocolDispatcher()
	cborDispatcher := NewCBORProtocolDispatcher()

	srv := &Server{
		config:          config,
		router:          router,
		registry:        registry,
		jsonDispatcher:  jsonDispatcher,
		queryDispatcher: queryDispatcher,
		cborDispatcher:  cborDispatcher,
		logger:          logger,
	}

	// Auto-register services from global registry
	for _, svc := range service.Services() {
		srv.RegisterService(svc)
	}

	// Register unified protocol dispatcher for POST /
	hasJSONServices := len(jsonDispatcher.handlers) > 0
	hasQueryServices := len(queryDispatcher.handlers) > 0

	if hasJSONServices || hasQueryServices {
		router.HandleFunc("POST", "/", srv.unifiedDispatcher)
		logger.Debug("registered unified protocol dispatcher for POST /")
	}

	// Register CBOR protocol dispatcher for /service/{serviceName}/operation/{operationName}
	hasCBORServices := len(cborDispatcher.handlers) > 0
	if hasCBORServices {
		router.HandleFunc("POST", "/service/{serviceName}/operation/{operationName}", srv.cborDispatcher.ServeHTTP)
		logger.Debug("registered CBOR protocol dispatcher for POST /service/*/operation/*")
	}

	return srv
}

// unifiedDispatcher routes requests to JSON or Query protocol handlers based on Content-Type.
func (s *Server) unifiedDispatcher(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	// Query protocol uses form-urlencoded.
	if contentType == "application/x-www-form-urlencoded" {
		s.queryDispatcher.ServeHTTP(w, r)

		return
	}

	// Default to JSON protocol.
	s.jsonDispatcher.ServeHTTP(w, r)
}

// Registry returns the service registry.
func (s *Server) Registry() *service.Registry {
	return s.registry
}

// Router returns the router.
func (s *Server) Router() *Router {
	return s.router
}

// RegisterService registers a service with the server.
func (s *Server) RegisterService(svc service.Service) {
	s.registry.Register(svc)
	svc.RegisterRoutes(s.router)

	// Check if service implements JSON protocol.
	if jsonSvc, ok := svc.(service.JSONProtocolService); ok {
		s.jsonDispatcher.Register(jsonSvc.TargetPrefix(), jsonSvc.DispatchAction)
		s.logger.Debug("registered JSON protocol service", "name", svc.Name(), "prefix", jsonSvc.TargetPrefix())
	}

	// Check if service implements Query protocol.
	if querySvc, ok := svc.(service.QueryProtocolService); ok {
		s.queryDispatcher.Register(querySvc.TargetPrefix(), querySvc.DispatchAction)

		// Register each action for proper routing.
		for _, action := range querySvc.Actions() {
			s.queryDispatcher.RegisterAction(action, querySvc.TargetPrefix(), querySvc.DispatchAction)
		}

		s.logger.Debug("registered Query protocol service", "name", svc.Name(), "prefix", querySvc.TargetPrefix())
	}

	// Check if service implements CBOR protocol.
	if cborSvc, ok := svc.(service.CBORProtocolService); ok {
		s.cborDispatcher.Register(cborSvc.ServiceName(), cborSvc.DispatchCBORAction)
		s.logger.Debug("registered CBOR protocol service", "name", svc.Name(), "serviceName", cborSvc.ServiceName())
	}

	s.logger.Info("registered service", "name", svc.Name())
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:              s.Addr(),
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.logger.Info("starting awsim server", "addr", s.Addr())

	// List registered services
	for _, name := range s.registry.Names() {
		s.logger.Info("service available", "name", name)
	}

	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// Run starts the server and handles graceful shutdown.
func (s *Server) Run() error {
	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		s.logger.Info("received signal", "signal", sig)
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
