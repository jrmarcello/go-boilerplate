// Package bootstrap is the composition root for the application.
// It wires all dependencies (repos, use cases, handlers) into a typed Container.
// This is the only package allowed to import from all architecture layers.
package bootstrap

import (
	"github.com/jmoiron/sqlx"

	"github.com/jrmarcello/gopherplate/internal/infrastructure/db/postgres/repository"
	infratelemetry "github.com/jrmarcello/gopherplate/internal/infrastructure/telemetry"
	"github.com/jrmarcello/gopherplate/internal/infrastructure/web/handler"
	roleuc "github.com/jrmarcello/gopherplate/internal/usecases/role"
	useruc "github.com/jrmarcello/gopherplate/internal/usecases/user"
	"github.com/jrmarcello/gopherplate/pkg/cache"
)

// Container holds all application dependencies grouped by layer.
// Only Handlers is exported — it is the legitimate external surface.
// Intermediate layers (repos, use cases) are unexported so callers cannot
// reach into the composition root and bypass constructor invariants.
type Container struct {
	repos        Repos
	roleUseCases RoleUseCases
	userUseCases UserUseCases
	Handlers     Handlers
}

// Repos groups all repository implementations.
type Repos struct {
	Role *repository.RoleRepository
	User *repository.UserRepository
}

// RoleUseCases groups all role domain use cases.
type RoleUseCases struct {
	Create *roleuc.CreateUseCase
	List   *roleuc.ListUseCase
	Delete *roleuc.DeleteUseCase
}

// UserUseCases groups all user domain use cases.
type UserUseCases struct {
	Create *useruc.CreateUseCase
	Get    *useruc.GetUseCase
	List   *useruc.ListUseCase
	Update *useruc.UpdateUseCase
	Delete *useruc.DeleteUseCase
}

// Handlers groups all HTTP handlers.
type Handlers struct {
	Role *handler.RoleHandler
	User *handler.UserHandler
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
	c.repos = Repos{
		Role: repository.NewRoleRepository(writer, reader),
		User: repository.NewUserRepository(writer, reader),
	}
}

func (c *Container) buildUseCases(cacheClient cache.Cache) {
	flightGroup := cache.NewFlightGroup()
	_ = flightGroup // used by domains with cache support
	_ = cacheClient // used by domains with cache support

	c.roleUseCases = RoleUseCases{
		Create: roleuc.NewCreateUseCase(c.repos.Role),
		List:   roleuc.NewListUseCase(c.repos.Role),
		Delete: roleuc.NewDeleteUseCase(c.repos.Role),
	}
	c.userUseCases = UserUseCases{
		Create: useruc.NewCreateUseCase(c.repos.User),
		Get:    useruc.NewGetUseCase(c.repos.User),
		List:   useruc.NewListUseCase(c.repos.User),
		Update: useruc.NewUpdateUseCase(c.repos.User),
		Delete: useruc.NewDeleteUseCase(c.repos.User),
	}
}

func (c *Container) buildHandlers(metrics *infratelemetry.Metrics) {
	_ = metrics // used by domains with business metrics
	c.Handlers = Handlers{
		Role: handler.NewRoleHandler(
			c.roleUseCases.Create,
			c.roleUseCases.List,
			c.roleUseCases.Delete,
		),
		User: handler.NewUserHandler(
			c.userUseCases.Create,
			c.userUseCases.Get,
			c.userUseCases.List,
			c.userUseCases.Update,
			c.userUseCases.Delete,
			metrics,
		),
	}
}
