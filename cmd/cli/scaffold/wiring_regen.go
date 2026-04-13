package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// DomainInfo holds the name variants and capabilities for a detected domain.
type DomainInfo struct {
	Name       string // snake_case: "order_item"
	Pascal     string // PascalCase: "OrderItem"
	Camel      string // camelCase: "orderItem"
	Plural     string // plural: "order_items"
	PluralPath string // plural for URL paths: "order-items" (kebab-case)
	// Use case capabilities — detected by file presence in internal/usecases/<name>/
	HasCreate  bool
	HasGet     bool
	HasList    bool
	HasUpdate  bool
	HasDelete  bool
	HasMetrics bool // true if handler constructor expects *telemetry.Metrics
}

// NewDomainInfo creates a DomainInfo from a snake_case domain name.
// Use DetectDomainInfo to populate capability fields from the filesystem.
func NewDomainInfo(snakeName string) DomainInfo {
	return DomainInfo{
		Name:       snakeName,
		Pascal:     ToPascalCase(snakeName),
		Camel:      ToCamelCase(snakeName),
		Plural:     ToPlural(snakeName),
		PluralPath: ToKebabCase(ToPlural(snakeName)),
		// Default to full CRUD for newly-scaffolded domains
		HasCreate: true,
		HasGet:    true,
		HasList:   true,
		HasUpdate: true,
		HasDelete: true,
	}
}

// DetectDomainInfo inspects projectDir for the given domain and populates
// capability fields (HasCreate/Get/List/Update/Delete, HasMetrics) based on
// which files actually exist.
func DetectDomainInfo(projectDir, snakeName string) DomainInfo {
	d := NewDomainInfo(snakeName)

	ucDir := filepath.Join(projectDir, "internal", "usecases", snakeName)
	d.HasCreate = fileExists(filepath.Join(ucDir, "create.go"))
	d.HasGet = fileExists(filepath.Join(ucDir, "get.go"))
	d.HasList = fileExists(filepath.Join(ucDir, "list.go"))
	d.HasUpdate = fileExists(filepath.Join(ucDir, "update.go"))
	d.HasDelete = fileExists(filepath.Join(ucDir, "delete.go"))

	// Detect whether handler takes *telemetry.Metrics by grepping the constructor signature
	handlerPath := filepath.Join(projectDir, "internal", "infrastructure", "web", "handler", snakeName+".go")
	if content, readErr := os.ReadFile(handlerPath); readErr == nil { //nolint:gosec // CLI reads project files
		if bytes.Contains(content, []byte("*telemetry.Metrics")) || bytes.Contains(content, []byte("*infratelemetry.Metrics")) {
			d.HasMetrics = true
		}
	}

	return d
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// regenData is the template data for N-domain wiring regeneration.
type regenData struct {
	ModulePath string
	Domains    []DomainInfo
}

// RegenerateFromDomains regenerates the 4 wiring files (server.go, router.go,
// container.go, test_helpers.go) based on the detected domains.
func RegenerateFromDomains(projectDir, modulePath string, domains []DomainInfo) error {
	data := regenData{
		ModulePath: modulePath,
		Domains:    domains,
	}

	files := []struct {
		relPath  string
		template string
		name     string
	}{
		{
			relPath:  filepath.Join("cmd", "api", "server.go"),
			template: serverGoNDomainsTemplate,
			name:     "server.go",
		},
		{
			relPath:  filepath.Join("internal", "infrastructure", "web", "router", "router.go"),
			template: routerGoNDomainsTemplate,
			name:     "router.go",
		},
		{
			relPath:  filepath.Join("internal", "bootstrap", "container.go"),
			template: containerGoNDomainsTemplate,
			name:     "container.go",
		},
		{
			relPath:  filepath.Join("internal", "bootstrap", "test_helpers.go"),
			template: testHelpersGoNDomainsTemplate,
			name:     "test_helpers.go",
		},
	}

	funcMap := template.FuncMap{
		"plural":     ToPlural,
		"singular":   ToSingular,
		"pascalCase": ToPascalCase,
		"camelCase":  ToCamelCase,
		"snakeCase":  ToSnakeCase,
		"kebabCase":  ToKebabCase,
		"lower":      ToLower,
	}

	for _, f := range files {
		tmpl, parseErr := template.New(f.name).Funcs(funcMap).Parse(f.template)
		if parseErr != nil {
			return fmt.Errorf("parsing template %s: %w", f.name, parseErr)
		}

		var buf bytes.Buffer
		if execErr := tmpl.Execute(&buf, data); execErr != nil {
			return fmt.Errorf("executing template %s: %w", f.name, execErr)
		}

		outPath := filepath.Join(projectDir, f.relPath)
		dirPath := filepath.Dir(outPath)
		if mkdirErr := os.MkdirAll(dirPath, 0o750); mkdirErr != nil {
			return fmt.Errorf("creating directory %s: %w", dirPath, mkdirErr)
		}

		if writeErr := os.WriteFile(outPath, buf.Bytes(), 0o600); writeErr != nil {
			return fmt.Errorf("writing %s: %w", f.name, writeErr)
		}
	}

	return nil
}

//nolint:lll
const serverGoNDomainsTemplate = `package main

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

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"{{.ModulePath}}/config"
{{- if gt (len .Domains) 0}}
	docs "{{.ModulePath}}/docs"
	"{{.ModulePath}}/internal/bootstrap"
	infratelemetry "{{.ModulePath}}/internal/infrastructure/telemetry"
{{- end}}
	"{{.ModulePath}}/internal/infrastructure/web/router"
	"{{.ModulePath}}/pkg/cache/redisclient"
	"{{.ModulePath}}/pkg/database"
	"{{.ModulePath}}/pkg/health"
	"{{.ModulePath}}/pkg/idempotency"
	"{{.ModulePath}}/pkg/idempotency/redisstore"
	"{{.ModulePath}}/pkg/logutil"
	pkgtelemetry "{{.ModulePath}}/pkg/telemetry"
	"{{.ModulePath}}/pkg/telemetry/otelgrpc"
	_ "github.com/lib/pq"
{{- if gt (len .Domains) 0}}
	"go.opentelemetry.io/otel"
{{- end}}
)

// Start initializes the application following the composition pattern:
// Config -> Logger -> Telemetry -> Database -> Dependencies -> Router -> Server
func Start(ctx context.Context, cfg *config.Config) error {
	// 0. Validate config
	if validateErr := cfg.Validate(); validateErr != nil {
		return fmt.Errorf("invalid configuration: %w", validateErr)
	}

	// 1. Logger
	logger := setupLogger()
	slog.SetDefault(logger)

	// Set Gin mode from config (avoid "Running in debug mode" warning in production)
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// 2. Telemetry (OpenTelemetry Traces + Metrics)
	// Graceful degradation: if OTel setup fails, app continues without telemetry.
	var exporterOpts []pkgtelemetry.Option
	if cfg.Otel.CollectorURL != "" {
		grpcOpts, exporterErr := otelgrpc.Exporters(ctx, otelgrpc.Config{
			CollectorURL: cfg.Otel.CollectorURL,
			Insecure:     cfg.Otel.Insecure,
		})
		if exporterErr != nil {
			slog.Warn("Telemetry exporter setup failed, continuing without observability", "error", exporterErr)
		} else {
			exporterOpts = grpcOpts
		}
	}

	tp, tpErr := pkgtelemetry.Setup(ctx, pkgtelemetry.Config{
		ServiceName: cfg.Otel.ServiceName,
		Enabled:     cfg.Otel.CollectorURL != "",
	}, exporterOpts...)
	if tpErr != nil {
		slog.Warn("Telemetry setup failed, continuing without observability", "error", tpErr)
	}
	if tp != nil {
		defer shutdownTelemetry(tp, logger)
	}

	// 3. Database (Writer/Reader Cluster)
	writerCfg := database.Config{
		Driver:          "postgres",
		DSN:             cfg.DB.GetWriterDSN(),
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.DB.ConnMaxIdleTime,
	}

	var readerCfg *database.Config
	if cfg.DB.ReplicaEnabled {
		readerCfg = &database.Config{
			Driver:          "postgres",
			DSN:             cfg.DB.GetReaderDSN(),
			MaxOpenConns:    cfg.DB.ReplicaMaxOpenConns,
			MaxIdleConns:    cfg.DB.ReplicaMaxIdleConns,
			ConnMaxLifetime: cfg.DB.ReplicaConnMaxLifetime,
			ConnMaxIdleTime: cfg.DB.ReplicaConnMaxIdleTime,
		}
	}

	cluster, clusterErr := database.NewDBCluster(writerCfg, readerCfg)
	if clusterErr != nil {
		return clusterErr
	}
	defer func() { _ = cluster.Close() }()

	// Wrap stdlib connections for sqlx-based repositories
	sqlxWriter := sqlx.NewDb(cluster.Writer(), "postgres")
	sqlxReader := sqlx.NewDb(cluster.Reader(), "postgres")

	// SSL mode warning for non-development environments
	if cfg.DB.SSLMode == "disable" && cfg.Server.Env != "development" {
		slog.Warn("database connection using sslmode=disable in non-development environment")
	}

	// 4. Register DB Pool Metrics
	if regErr := pkgtelemetry.RegisterDBPoolMetrics(ctx, cfg.Otel.ServiceName, cluster.Writer(), "writer"); regErr != nil {
		slog.Warn("Failed to register DB pool metrics", "error", regErr)
	}

	if cluster.HasSeparateReader() {
		if regErr := pkgtelemetry.RegisterDBPoolMetrics(ctx, cfg.Otel.ServiceName, cluster.Reader(), "reader"); regErr != nil {
			slog.Warn("Failed to register reader DB pool metrics", "error", regErr)
		}
	}
{{if gt (len .Domains) 0}}
	// 5. Business Metrics (injected into handlers, not global)
	businessMetrics, metricsErr := infratelemetry.NewMetrics(otel.Meter(cfg.Otel.ServiceName))
	if metricsErr != nil {
		slog.Warn("Failed to create business metrics", "error", metricsErr)
	}
{{end}}
	// 6. Dependencies (Dependency Injection)
	var httpMetrics *pkgtelemetry.HTTPMetrics
	if tp != nil {
		httpMetrics = tp.HTTPMetrics()
	}
	deps := buildDependencies(cluster, sqlxWriter, sqlxReader, cfg, httpMetrics{{if gt (len .Domains) 0}}, businessMetrics{{end}})
{{if gt (len .Domains) 0}}
	// Swagger Dynamic Config
	if cfg.Swagger.Host != "" {
		docs.SwaggerInfo.Host = cfg.Swagger.Host
	} else {
		docs.SwaggerInfo.Host = "localhost:" + cfg.Server.Port
	}
{{end}}
	// 7. Router
	r := router.Setup(deps)

	// 8. Server
	srv := newServer(cfg.Server.Port, r)

	// 9. Graceful Shutdown
	return runWithGracefulShutdown(srv, logger)
}

func setupLogger() *slog.Logger {
	stdout := slog.NewJSONHandler(os.Stdout, nil)
	masked := logutil.NewMaskingHandler(logutil.NewMasker(logutil.DefaultBRConfig()), stdout)
	return slog.New(logutil.NewFanoutHandler(masked))
}

func shutdownTelemetry(tp *pkgtelemetry.Provider, logger *slog.Logger) {
	if shutdownErr := tp.Shutdown(context.Background()); shutdownErr != nil {
		logger.Error("failed to shutdown telemetry", "error", shutdownErr)
	}
}

func buildDependencies(cluster *database.DBCluster, sqlxWriter, sqlxReader *sqlx.DB, cfg *config.Config, httpMetrics *pkgtelemetry.HTTPMetrics{{if gt (len .Domains) 0}}, businessMetrics *infratelemetry.Metrics{{end}}) router.Dependencies {
	// Cache (optional -- config-dependent, stays in server.go)
	redisClient, cacheErr := redisclient.NewRedisClient(redisclient.RedisConfig{
		URL:          cfg.Redis.URL,
		TTL:          cfg.Redis.TTL,
		Enabled:      cfg.Redis.Enabled,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if cacheErr != nil {
		slog.Warn("Redis cache disabled", "error", cacheErr)
	}

	// Health Checker (cross-cutting, stays in server.go)
	checker := health.New()
	checker.Register("database_writer", true, func(ctx context.Context) error {
		return cluster.Writer().PingContext(ctx)
	})
	if cluster.HasSeparateReader() {
		checker.Register("database_reader", false, func(ctx context.Context) error {
			return cluster.Reader().PingContext(ctx)
		})
	}
	if redisClient != nil && redisClient.UnderlyingClient() != nil {
		checker.Register("redis", false, func(ctx context.Context) error {
			return redisClient.Ping(ctx)
		})
	}
{{if gt (len .Domains) 0}}
	// Bootstrap container (repos, use cases, handlers for all domains)
	c := bootstrap.New(sqlxWriter, sqlxReader, redisClient, businessMetrics)
{{end}}
	// Idempotency Store (optional -- config-dependent, stays in server.go)
	var idempotencyStore idempotency.Store
	if cfg.Idempotency.Enabled {
		if rc := redisClient.UnderlyingClient(); rc != nil {
			ttl, _ := time.ParseDuration(cfg.Idempotency.TTL)
			lockTTL, _ := time.ParseDuration(cfg.Idempotency.LockTTL)
			idempotencyStore = redisstore.NewRedisStore(rc, ttl, lockTTL)
		}
	}

	return router.Dependencies{
		HealthChecker: checker,
{{- range .Domains}}
		{{.Pascal}}Handler: c.Handlers.{{.Pascal}},
{{- end}}
		HTTPMetrics:      httpMetrics,
		IdempotencyStore: idempotencyStore,
		Config: router.Config{
			ServiceName:        cfg.Otel.ServiceName,
			ServiceKeysEnabled: cfg.Auth.Enabled,
			ServiceKeys:        cfg.Auth.ServiceKeys,
			SwaggerEnabled:     cfg.Swagger.Enabled,
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
		MaxHeaderBytes:    1 << 20, // 1MB -- protects against oversized headers
	}
}

func runWithGracefulShutdown(srv *http.Server, logger *slog.Logger) error {
	// Error channel to capture server startup failures without os.Exit in goroutine
	errCh := make(chan error, 1)
	go func() {
		logger.Info("Starting server", "port", srv.Addr)
		if listenErr := srv.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			errCh <- listenErr
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case listenErr := <-errCh:
		return listenErr
	case <-quit:
		// proceed to graceful shutdown
	}

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
		return shutdownErr
	}

	logger.Info("server exited properly")
	return nil
}
`

//nolint:lll
const routerGoNDomainsTemplate = `package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
{{- if gt (len .Domains) 0}}
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
{{- end}}
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

{{- if gt (len .Domains) 0}}
	"{{.ModulePath}}/internal/infrastructure/web/handler"
{{- end}}
	"{{.ModulePath}}/internal/infrastructure/web/middleware"
	"{{.ModulePath}}/pkg/health"
	"{{.ModulePath}}/pkg/httputil/httpgin"
	"{{.ModulePath}}/pkg/idempotency"
	"{{.ModulePath}}/pkg/telemetry"
)

// Config holds router configuration.
type Config struct {
	ServiceName        string
	ServiceKeysEnabled bool   // fail-closed in HML/PRD if keys empty
	ServiceKeys        string // "service1:key1,service2:key2"
	SwaggerEnabled     bool
}

// Dependencies groups all dependencies required by the router.
type Dependencies struct {
	HealthChecker *health.Checker
{{- range .Domains}}
	{{.Pascal}}Handler *handler.{{.Pascal}}Handler
{{- end}}
	HTTPMetrics      *telemetry.HTTPMetrics
	IdempotencyStore idempotency.Store
	Config           Config
}

// Setup configures and returns the Gin engine with all middlewares and routes.
func Setup(deps Dependencies) *gin.Engine {
	r := gin.New()

	// Recovery middleware (panic recovery -- returns JSON 500, not HTML)
	r.Use(middleware.CustomRecovery())

	// OpenTelemetry (must be before Logger to populate trace_id)
	r.Use(otelgin.Middleware(deps.Config.ServiceName))

	// HTTP Metrics (count, duration, Apdex)
	r.Use(middleware.Metrics(deps.HTTPMetrics))

	// Custom structured logger
	r.Use(middleware.Logger())

	// Idempotency (optional -- only if store is provided)
	if deps.IdempotencyStore != nil {
		r.Use(middleware.Idempotency(deps.IdempotencyStore))
	}

	// Public routes (no auth required)
	if deps.Config.SwaggerEnabled {
		registerSwaggerRoutes(r)
	}
	registerHealthRoutes(r, deps)

	// Protected routes (auth required if SERVICE_KEYS is configured)
	authConfig := middleware.ServiceKeyConfig{
		Enabled: deps.Config.ServiceKeysEnabled,
		Keys:    middleware.ParseServiceKeys(deps.Config.ServiceKeys),
	}
	protected := r.Group("")
	protected.Use(middleware.ServiceKeyAuth(authConfig))
{{- range .Domains}}
	Register{{.Pascal}}Routes(protected, deps.{{.Pascal}}Handler)
{{- end}}

	return r
}

// registerSwaggerRoutes registers Swagger routes.
func registerSwaggerRoutes(r *gin.Engine) {
{{- if gt (len .Domains) 0}}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
{{- else}}
	// TODO: uncomment after running swag init
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	_ = r
{{- end}}
}

// registerHealthRoutes registers health check routes.
func registerHealthRoutes(r *gin.Engine, deps Dependencies) {
	// Liveness - always ok (K8s restart if process is dead)
	r.GET("/health", func(c *gin.Context) {
		httpgin.SendSuccess(c, http.StatusOK, gin.H{
			"status":  "ok",
			"service": deps.Config.ServiceName,
		})
	})

	// Readiness - checks all dependencies
	r.GET("/ready", func(c *gin.Context) {
		healthy, statuses := deps.HealthChecker.RunAll(c.Request.Context())

		result := gin.H{
			"status": "ready",
		}
		if !healthy {
			result["status"] = "not ready"
		}
		result["checks"] = statuses

		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}
		httpgin.SendSuccess(c, status, result)
	})
}
`

//nolint:lll
const containerGoNDomainsTemplate = `// Package bootstrap is the composition root for the application.
// It wires all dependencies (repos, use cases, handlers) into a typed Container.
// This is the only package allowed to import from all architecture layers.
package bootstrap

import (
	"github.com/jmoiron/sqlx"

	"{{.ModulePath}}/internal/infrastructure/db/postgres/repository"
	infratelemetry "{{.ModulePath}}/internal/infrastructure/telemetry"
	"{{.ModulePath}}/internal/infrastructure/web/handler"
{{- range .Domains}}
	{{.Camel}}uc "{{$.ModulePath}}/internal/usecases/{{.Name}}"
{{- end}}
	"{{.ModulePath}}/pkg/cache"
)

// Container holds all application dependencies grouped by layer.
type Container struct {
	Repos    Repos
{{- range .Domains}}
	{{.Pascal}}UseCases {{.Pascal}}UseCases
{{- end}}
	Handlers Handlers
}

// Repos groups all repository implementations.
type Repos struct {
{{- range .Domains}}
	{{.Pascal}} *repository.{{.Pascal}}Repository
{{- end}}
}
{{range .Domains}}
// {{.Pascal}}UseCases groups all {{.Name}} domain use cases.
type {{.Pascal}}UseCases struct {
{{- if .HasCreate}}
	Create *{{.Camel}}uc.CreateUseCase
{{- end}}
{{- if .HasGet}}
	Get    *{{.Camel}}uc.GetUseCase
{{- end}}
{{- if .HasList}}
	List   *{{.Camel}}uc.ListUseCase
{{- end}}
{{- if .HasUpdate}}
	Update *{{.Camel}}uc.UpdateUseCase
{{- end}}
{{- if .HasDelete}}
	Delete *{{.Camel}}uc.DeleteUseCase
{{- end}}
}
{{end}}
// Handlers groups all HTTP handlers.
type Handlers struct {
{{- range .Domains}}
	{{.Pascal}} *handler.{{.Pascal}}Handler
{{- end}}
}

// New creates a fully wired Container. The construction follows a strict phase order:
// repos -> use cases -> handlers, preventing circular dependencies.
// metrics may be nil (for tests or contexts without OTel).
func New(writer, reader *sqlx.DB, cacheClient cache.Cache, metrics *infratelemetry.Metrics) *Container {
	c := &Container{}
	c.buildRepos(writer, reader)
	c.buildUseCases(cacheClient)
	c.buildHandlers(metrics)
	return c
}

func (c *Container) buildRepos(writer, reader *sqlx.DB) {
	c.Repos = Repos{
{{- range .Domains}}
		{{.Pascal}}: repository.New{{.Pascal}}Repository(writer, reader),
{{- end}}
	}
}

func (c *Container) buildUseCases(cacheClient cache.Cache) {
	flightGroup := cache.NewFlightGroup()
	_ = flightGroup // used by domains with cache support
	_ = cacheClient // used by domains with cache support
{{range .Domains}}
	c.{{.Pascal}}UseCases = {{.Pascal}}UseCases{
{{- if .HasCreate}}
		Create: {{.Camel}}uc.NewCreateUseCase(c.Repos.{{.Pascal}}),
{{- end}}
{{- if .HasGet}}
		Get:    {{.Camel}}uc.NewGetUseCase(c.Repos.{{.Pascal}}),
{{- end}}
{{- if .HasList}}
		List:   {{.Camel}}uc.NewListUseCase(c.Repos.{{.Pascal}}),
{{- end}}
{{- if .HasUpdate}}
		Update: {{.Camel}}uc.NewUpdateUseCase(c.Repos.{{.Pascal}}),
{{- end}}
{{- if .HasDelete}}
		Delete: {{.Camel}}uc.NewDeleteUseCase(c.Repos.{{.Pascal}}),
{{- end}}
	}
{{- end}}
}

func (c *Container) buildHandlers(metrics *infratelemetry.Metrics) {
	_ = metrics // used by domains with business metrics
	c.Handlers = Handlers{
{{- range .Domains}}
		{{.Pascal}}: handler.New{{.Pascal}}Handler(
{{- if .HasCreate}}
			c.{{.Pascal}}UseCases.Create,
{{- end}}
{{- if .HasGet}}
			c.{{.Pascal}}UseCases.Get,
{{- end}}
{{- if .HasList}}
			c.{{.Pascal}}UseCases.List,
{{- end}}
{{- if .HasUpdate}}
			c.{{.Pascal}}UseCases.Update,
{{- end}}
{{- if .HasDelete}}
			c.{{.Pascal}}UseCases.Delete,
{{- end}}
{{- if .HasMetrics}}
			metrics,
{{- end}}
		),
{{- end}}
	}
}
`

//nolint:lll
const testHelpersGoNDomainsTemplate = `package bootstrap

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"{{.ModulePath}}/internal/infrastructure/web/middleware"
	"{{.ModulePath}}/internal/infrastructure/web/router"
	"{{.ModulePath}}/pkg/cache"
	"{{.ModulePath}}/pkg/httputil/httpgin"
)

// NewForTest creates a Container suitable for testing. It uses the same DB
// connection as both writer and reader, and passes nil metrics.
func NewForTest(t testing.TB, db *sqlx.DB, cacheClient cache.Cache) *Container {
	t.Helper()
	return New(db, db, cacheClient, nil)
}

// SetupTestRouter creates a gin.Engine in test mode with all routes
// registered, CustomRecovery middleware, health endpoints, and a panic-test
// route. No auth middleware is applied.
func SetupTestRouter(t testing.TB, db *sqlx.DB, cacheClient cache.Cache) *gin.Engine {
	t.Helper()

	c := NewForTest(t, db, cacheClient)
	r := newTestEngine()

	registerTestHealthRoutes(r, db)

	// Register all domain routes without auth
	group := r.Group("")
{{- range .Domains}}
	router.Register{{.Pascal}}Routes(group, c.Handlers.{{.Pascal}})
{{- end}}

	return r
}

// SetupTestRouterWithAuth creates a gin.Engine in test mode with all routes
// registered behind service key authentication middleware.
// serviceKeys uses the format "service1:key1,service2:key2".
func SetupTestRouterWithAuth(t testing.TB, db *sqlx.DB, cacheClient cache.Cache, serviceKeys string) *gin.Engine {
	t.Helper()

	c := NewForTest(t, db, cacheClient)
	r := newTestEngine()

	registerTestHealthRoutes(r, db)

	// Register all domain routes behind auth middleware
	authConfig := middleware.ServiceKeyConfig{
		Enabled: true,
		Keys:    middleware.ParseServiceKeys(serviceKeys),
	}
	protected := r.Group("")
	protected.Use(middleware.ServiceKeyAuth(authConfig))
{{- range .Domains}}
	router.Register{{.Pascal}}Routes(protected, c.Handlers.{{.Pascal}})
{{- end}}

	return r
}

// newTestEngine creates a minimal gin.Engine for testing with TestMode and
// CustomRecovery middleware. It also registers a panic-test route used by
// E2E recovery middleware tests.
func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CustomRecovery())

	// Panic test route (only for E2E testing)
	r.GET("/panic-test", func(_ *gin.Context) {
		panic("test panic for recovery middleware")
	})

	return r
}

// registerTestHealthRoutes registers simplified health/ready endpoints for tests.
// Uses health.New() with no checks -- always returns healthy (DB connectivity is
// already validated by the test container setup).
func registerTestHealthRoutes(r *gin.Engine, db *sqlx.DB) {
	r.GET("/health", func(c *gin.Context) {
		httpgin.SendSuccess(c, http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ready", func(c *gin.Context) {
		if pingErr := db.Ping(); pingErr != nil {
			httpgin.SendError(c, http.StatusServiceUnavailable, "database connection failed")
			return
		}
		httpgin.SendSuccess(c, http.StatusOK, gin.H{"status": "ready"})
	})
}
`
