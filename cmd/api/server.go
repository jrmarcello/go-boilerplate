package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/appmax-space/go-boilerplate/config"
	docs "bitbucket.org/appmax-space/go-boilerplate/docs"
	"bitbucket.org/appmax-space/go-boilerplate/internal/infrastructure/db/postgres/repository"
	"bitbucket.org/appmax-space/go-boilerplate/internal/infrastructure/web/handler"
	"bitbucket.org/appmax-space/go-boilerplate/internal/infrastructure/web/router"
	entityuc "bitbucket.org/appmax-space/go-boilerplate/internal/usecases/entity_example"
	pkgcache "bitbucket.org/appmax-space/go-boilerplate/pkg/cache"
	"bitbucket.org/appmax-space/go-boilerplate/pkg/database"
	pkgtelemetry "bitbucket.org/appmax-space/go-boilerplate/pkg/telemetry"
)

// Start initializes the application following the composition pattern:
// Config → Logger → Telemetry → Database → Dependencies → Router → Server
func Start(ctx context.Context, cfg *config.Config) error {
	// 1. Logger
	logger := setupLogger()
	slog.SetDefault(logger)

	// 2. Telemetry (OpenTelemetry Traces + Metrics)
	tp, tpErr := pkgtelemetry.Setup(ctx, pkgtelemetry.Config{
		ServiceName:  cfg.Otel.ServiceName,
		CollectorURL: cfg.Otel.CollectorURL,
		Enabled:      cfg.Otel.CollectorURL != "",
	})
	if tpErr != nil {
		return tpErr
	}
	defer shutdownTelemetry(tp, logger)

	// 3. Database (Writer/Reader Cluster)
	cluster, clusterErr := database.NewDBCluster(
		database.Config{
			DSN:             cfg.DB.DSN,
			MaxOpenConns:    cfg.DB.MaxOpenConns,
			MaxIdleConns:    cfg.DB.MaxIdleConns,
			ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		},
		cfg.DB.ReaderConfig(),
	)
	if clusterErr != nil {
		return clusterErr
	}
	defer cluster.Close()

	// 4. Register DB Pool Metrics
	if regErr := pkgtelemetry.RegisterDBPoolMetrics(ctx, cfg.Otel.ServiceName, cluster.Writer().DB, "writer"); regErr != nil {
		slog.Warn("Failed to register DB pool metrics", "error", regErr)
	}

	// 5. Dependencies (Dependency Injection)
	deps := buildDependencies(cluster, cfg, tp.HTTPMetrics())

	// Swagger Dynamic Config
	docs.SwaggerInfo.Host = "localhost:" + cfg.Server.Port

	// 6. Router
	r := router.Setup(deps)

	// 7. Server
	srv := newServer(cfg.Server.Port, r)

	// 8. Graceful Shutdown
	return runWithGracefulShutdown(srv, logger)
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func shutdownTelemetry(tp *pkgtelemetry.Provider, logger *slog.Logger) {
	if shutdownErr := tp.Shutdown(context.Background()); shutdownErr != nil {
		logger.Error("failed to shutdown telemetry", "error", shutdownErr)
	}
}

func buildDependencies(cluster *database.DBCluster, cfg *config.Config, httpMetrics *pkgtelemetry.HTTPMetrics) router.Dependencies {
	// Repositories (use writer for mutations, reader for queries)
	repo := &repository.EntityRepository{DB: cluster.Writer()}

	// Cache (optional)
	redisClient, cacheErr := pkgcache.NewRedisClient(pkgcache.RedisConfig{
		URL:     cfg.Redis.URL,
		TTL:     cfg.Redis.TTL,
		Enabled: cfg.Redis.Enabled,
	})
	if cacheErr != nil {
		slog.Warn("Redis cache disabled", "error", cacheErr)
	}

	// Use Cases (with optional cache via builder pattern)
	createUC := entityuc.NewCreateUseCase(repo)
	getUC := entityuc.NewGetUseCase(repo).WithCache(redisClient)
	listUC := entityuc.NewListUseCase(repo)
	updateUC := entityuc.NewUpdateUseCase(repo).WithCache(redisClient)
	deleteUC := entityuc.NewDeleteUseCase(repo).WithCache(redisClient)

	// Handlers
	entityHandler := handler.NewEntityHandler(createUC, getUC, listUC, updateUC, deleteUC)

	return router.Dependencies{
		DB:            cluster.Writer(),
		EntityHandler: entityHandler,
		HTTPMetrics:   httpMetrics,
		Config: router.Config{
			ServiceName:    cfg.Otel.ServiceName,
			ServiceKeys:    cfg.Auth.ServiceKeys,
			SwaggerEnabled: cfg.Swagger.Enabled,
		},
	}
}

func newServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func runWithGracefulShutdown(srv *http.Server, logger *slog.Logger) error {
	// Start server in goroutine
	go func() {
		logger.Info("Starting server", "port", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server exited properly")
	return nil
}
