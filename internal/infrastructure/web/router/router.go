package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrmarcello/gopherplate/internal/infrastructure/web/handler"
	"github.com/jrmarcello/gopherplate/internal/infrastructure/web/middleware"
	"github.com/jrmarcello/gopherplate/pkg/health"
	"github.com/jrmarcello/gopherplate/pkg/httputil/httpgin"
	"github.com/jrmarcello/gopherplate/pkg/idempotency"
	"github.com/jrmarcello/gopherplate/pkg/telemetry"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Config holds router configuration.
type Config struct {
	ServiceName        string
	ServiceKeysEnabled bool   // fail-closed in HML/PRD if keys empty
	ServiceKeys        string // "service1:key1,service2:key2"
	SwaggerEnabled     bool
	MaxBodySize        int64 // Max request body in bytes (0 disables the cap)
}

// Dependencies groups all dependencies required by the router.
type Dependencies struct {
	HealthChecker    *health.Checker
	RoleHandler      *handler.RoleHandler
	UserHandler      *handler.UserHandler
	HTTPMetrics      *telemetry.HTTPMetrics
	IdempotencyStore idempotency.Store
	Config           Config
}

// Setup configures and returns the Gin engine with all middlewares and routes.
func Setup(deps Dependencies) *gin.Engine {
	r := gin.New()

	// Recovery middleware (panic recovery -- returns JSON 500, not HTML)
	r.Use(middleware.CustomRecovery())

	// Body size cap (must run before any middleware that reads the body,
	// e.g. Idempotency and handlers doing ShouldBindJSON).
	r.Use(middleware.BodyLimit(deps.Config.MaxBodySize))

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
	RegisterRoleRoutes(protected, deps.RoleHandler)
	RegisterUserRoutes(protected, deps.UserHandler)

	return r
}

// registerSwaggerRoutes registers Swagger routes.
func registerSwaggerRoutes(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
