package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jrmarcello/gopherplate/cmd/cli/scaffold"
)

// createFakeDomain creates the minimal directory structure for a domain:
// internal/domain/<name>/, handler, router, and repository files.
func createFakeDomain(t *testing.T, projectDir, name string) {
	t.Helper()

	dirs := []string{
		filepath.Join(projectDir, "internal", "domain", name),
		filepath.Join(projectDir, "internal", "infrastructure", "web", "handler"),
		filepath.Join(projectDir, "internal", "infrastructure", "web", "router"),
		filepath.Join(projectDir, "internal", "infrastructure", "db", "postgres", "repository"),
		filepath.Join(projectDir, "internal", "usecases", name),
	}

	for _, dir := range dirs {
		mkdirErr := os.MkdirAll(dir, 0o750)
		require.NoError(t, mkdirErr)
	}

	// Create the 3 infrastructure files that scanDomains checks
	files := []string{
		filepath.Join(projectDir, "internal", "infrastructure", "web", "handler", name+".go"),
		filepath.Join(projectDir, "internal", "infrastructure", "web", "router", name+".go"),
		filepath.Join(projectDir, "internal", "infrastructure", "db", "postgres", "repository", name+".go"),
	}

	for _, f := range files {
		writeErr := os.WriteFile(f, []byte("package placeholder\n"), 0o600)
		require.NoError(t, writeErr)
	}
}

// createFakeProject creates a minimal project structure with go.mod.
func createFakeProject(t *testing.T, projectDir, modulePath string) {
	t.Helper()

	goModContent := "module " + modulePath + "\n\ngo 1.25.0\n"
	writeErr := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0o600)
	require.NoError(t, writeErr)

	// Create the necessary directories
	dirs := []string{
		filepath.Join(projectDir, "cmd", "api"),
		filepath.Join(projectDir, "internal", "bootstrap"),
		filepath.Join(projectDir, "internal", "infrastructure", "web", "router"),
		filepath.Join(projectDir, "internal", "domain"),
	}
	for _, dir := range dirs {
		mkdirErr := os.MkdirAll(dir, 0o750)
		require.NoError(t, mkdirErr)
	}
}

func TestScanDomains(t *testing.T) {
	t.Run("detects domains with full infrastructure", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "user")
		createFakeDomain(t, dir, "role")

		domains, scanErr := scanDomains(dir)
		require.NoError(t, scanErr)
		require.Len(t, domains, 2)

		// Sorted alphabetically
		assert.Equal(t, "role", domains[0].Name)
		assert.Equal(t, "Role", domains[0].Pascal)
		assert.Equal(t, "role", domains[0].Camel)

		assert.Equal(t, "user", domains[1].Name)
		assert.Equal(t, "User", domains[1].Pascal)
		assert.Equal(t, "user", domains[1].Camel)
	})

	t.Run("skips domains without handler", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "user")

		// Create a domain dir without infrastructure
		domainOnly := filepath.Join(dir, "internal", "domain", "incomplete")
		mkdirErr := os.MkdirAll(domainOnly, 0o750)
		require.NoError(t, mkdirErr)

		domains, scanErr := scanDomains(dir)
		require.NoError(t, scanErr)
		require.Len(t, domains, 1)
		assert.Equal(t, "user", domains[0].Name)
	})

	t.Run("returns empty for no domains", func(t *testing.T) {
		dir := t.TempDir()
		domainDir := filepath.Join(dir, "internal", "domain")
		mkdirErr := os.MkdirAll(domainDir, 0o750)
		require.NoError(t, mkdirErr)

		domains, scanErr := scanDomains(dir)
		require.NoError(t, scanErr)
		assert.Empty(t, domains)
	})

	t.Run("returns nil when internal/domain does not exist", func(t *testing.T) {
		dir := t.TempDir()

		domains, scanErr := scanDomains(dir)
		require.NoError(t, scanErr)
		assert.Nil(t, domains)
	})
}

func TestHasDomainInfrastructure(t *testing.T) {
	t.Run("returns true when all 3 files exist", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "order")

		assert.True(t, hasDomainInfrastructure(dir, "order"))
	})

	t.Run("returns false when handler is missing", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "order")

		// Remove handler file
		removeErr := os.Remove(filepath.Join(dir, "internal", "infrastructure", "web", "handler", "order.go"))
		require.NoError(t, removeErr)

		assert.False(t, hasDomainInfrastructure(dir, "order"))
	})

	t.Run("returns false when router is missing", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "order")

		removeErr := os.Remove(filepath.Join(dir, "internal", "infrastructure", "web", "router", "order.go"))
		require.NoError(t, removeErr)

		assert.False(t, hasDomainInfrastructure(dir, "order"))
	})

	t.Run("returns false when repository is missing", func(t *testing.T) {
		dir := t.TempDir()
		createFakeDomain(t, dir, "order")

		removeErr := os.Remove(filepath.Join(dir, "internal", "infrastructure", "db", "postgres", "repository", "order.go"))
		require.NoError(t, removeErr)

		assert.False(t, hasDomainInfrastructure(dir, "order"))
	})
}

func TestRegenerateFromDomains_ThreeDomains(t *testing.T) {
	dir := t.TempDir()
	modulePath := "github.com/test/myservice"
	createFakeProject(t, dir, modulePath)

	domains := []scaffold.DomainInfo{
		scaffold.NewDomainInfo("order"),
		scaffold.NewDomainInfo("role"),
		scaffold.NewDomainInfo("user"),
	}

	regenErr := scaffold.RegenerateFromDomains(dir, modulePath, domains)
	require.NoError(t, regenErr)

	// Verify server.go was generated with all 3 domain imports
	serverContent, readErr := os.ReadFile(filepath.Join(dir, "cmd", "api", "server.go"))
	require.NoError(t, readErr)
	serverStr := string(serverContent)

	assert.Contains(t, serverStr, "package main")
	assert.Contains(t, serverStr, modulePath+"/internal/bootstrap")
	assert.Contains(t, serverStr, "OrderHandler:")
	assert.Contains(t, serverStr, "RoleHandler:")
	assert.Contains(t, serverStr, "UserHandler:")

	// Verify router.go was generated with all 3 domains
	routerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "infrastructure", "web", "router", "router.go"))
	require.NoError(t, readErr)
	routerStr := string(routerContent)

	assert.Contains(t, routerStr, "package router")
	assert.Contains(t, routerStr, "RegisterOrderRoutes(protected, deps.OrderHandler)")
	assert.Contains(t, routerStr, "RegisterRoleRoutes(protected, deps.RoleHandler)")
	assert.Contains(t, routerStr, "RegisterUserRoutes(protected, deps.UserHandler)")
	assert.Contains(t, routerStr, "OrderHandler *handler.OrderHandler")
	assert.Contains(t, routerStr, "RoleHandler *handler.RoleHandler")
	assert.Contains(t, routerStr, "UserHandler *handler.UserHandler")

	// Verify container.go was generated with all 3 domains
	containerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "bootstrap", "container.go"))
	require.NoError(t, readErr)
	containerStr := string(containerContent)

	assert.Contains(t, containerStr, "package bootstrap")
	assert.Contains(t, containerStr, modulePath+"/internal/usecases/order")
	assert.Contains(t, containerStr, modulePath+"/internal/usecases/role")
	assert.Contains(t, containerStr, modulePath+"/internal/usecases/user")
	assert.Contains(t, containerStr, "OrderUseCases")
	assert.Contains(t, containerStr, "RoleUseCases")
	assert.Contains(t, containerStr, "UserUseCases")

	// Verify test_helpers.go was generated with all 3 domains
	helpersContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "bootstrap", "test_helpers.go"))
	require.NoError(t, readErr)
	helpersStr := string(helpersContent)

	assert.Contains(t, helpersStr, "package bootstrap")
	assert.Contains(t, helpersStr, "RegisterOrderRoutes(group, c.Handlers.Order)")
	assert.Contains(t, helpersStr, "RegisterRoleRoutes(group, c.Handlers.Role)")
	assert.Contains(t, helpersStr, "RegisterUserRoutes(group, c.Handlers.User)")
}

func TestRegenerateFromDomains_TwoDomains(t *testing.T) {
	dir := t.TempDir()
	modulePath := "github.com/test/myservice"
	createFakeProject(t, dir, modulePath)

	domains := []scaffold.DomainInfo{
		scaffold.NewDomainInfo("role"),
		scaffold.NewDomainInfo("user"),
	}

	regenErr := scaffold.RegenerateFromDomains(dir, modulePath, domains)
	require.NoError(t, regenErr)

	// Verify server.go was generated correctly
	serverContent, readErr := os.ReadFile(filepath.Join(dir, "cmd", "api", "server.go"))
	require.NoError(t, readErr)
	serverStr := string(serverContent)

	assert.Contains(t, serverStr, "package main")
	assert.Contains(t, serverStr, "RoleHandler:")
	assert.Contains(t, serverStr, "UserHandler:")
	// Should NOT contain order
	assert.NotContains(t, serverStr, "OrderHandler")

	// Verify container.go
	containerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "bootstrap", "container.go"))
	require.NoError(t, readErr)
	containerStr := string(containerContent)

	assert.Contains(t, containerStr, "RoleUseCases")
	assert.Contains(t, containerStr, "UserUseCases")
	assert.NotContains(t, containerStr, "OrderUseCases")

	// Verify router.go
	routerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "infrastructure", "web", "router", "router.go"))
	require.NoError(t, readErr)
	routerStr := string(routerContent)

	assert.Contains(t, routerStr, "RegisterRoleRoutes")
	assert.Contains(t, routerStr, "RegisterUserRoutes")
	assert.NotContains(t, routerStr, "RegisterOrderRoutes")
}

func TestRegenerateFromDomains_ZeroDomains(t *testing.T) {
	dir := t.TempDir()
	modulePath := "github.com/test/myservice"
	createFakeProject(t, dir, modulePath)

	var domains []scaffold.DomainInfo

	regenErr := scaffold.RegenerateFromDomains(dir, modulePath, domains)
	require.NoError(t, regenErr)

	// Verify server.go was generated as minimal (no domain imports)
	serverContent, readErr := os.ReadFile(filepath.Join(dir, "cmd", "api", "server.go"))
	require.NoError(t, readErr)
	serverStr := string(serverContent)

	assert.Contains(t, serverStr, "package main")
	assert.Contains(t, serverStr, "func Start(ctx context.Context")
	// Should NOT import bootstrap or infratelemetry when 0 domains
	assert.NotContains(t, serverStr, "internal/bootstrap")
	assert.NotContains(t, serverStr, "infratelemetry")

	// Verify router.go (no domain handlers)
	routerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "infrastructure", "web", "router", "router.go"))
	require.NoError(t, readErr)
	routerStr := string(routerContent)

	assert.Contains(t, routerStr, "package router")
	assert.NotContains(t, routerStr, "Handler *handler.")
	// Should have the swagger TODO comment
	assert.Contains(t, routerStr, "TODO: uncomment after running swag init")

	// Verify container.go exists and is minimal
	containerContent, readErr := os.ReadFile(filepath.Join(dir, "internal", "bootstrap", "container.go"))
	require.NoError(t, readErr)
	containerStr := string(containerContent)

	assert.Contains(t, containerStr, "package bootstrap")
	// No domain-specific use case types should be present
	assert.NotContains(t, containerStr, "UserUseCases")
	assert.NotContains(t, containerStr, "RoleUseCases")
}

func TestNewDomainInfo(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantPascal string
		wantCamel  string
		wantPlural string
	}{
		{
			name:       "simple name",
			input:      "user",
			wantName:   "user",
			wantPascal: "User",
			wantCamel:  "user",
			wantPlural: "users",
		},
		{
			name:       "snake_case name",
			input:      "order_item",
			wantName:   "order_item",
			wantPascal: "OrderItem",
			wantCamel:  "orderItem",
			wantPlural: "order_items",
		},
		{
			name:       "role",
			input:      "role",
			wantName:   "role",
			wantPascal: "Role",
			wantCamel:  "role",
			wantPlural: "roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := scaffold.NewDomainInfo(tt.input)
			assert.Equal(t, tt.wantName, info.Name)
			assert.Equal(t, tt.wantPascal, info.Pascal)
			assert.Equal(t, tt.wantCamel, info.Camel)
			assert.Equal(t, tt.wantPlural, info.Plural)
		})
	}
}

func TestRegenerateFromDomains_AllFourFilesCreated(t *testing.T) {
	dir := t.TempDir()
	modulePath := "github.com/test/myservice"
	createFakeProject(t, dir, modulePath)

	domains := []scaffold.DomainInfo{
		scaffold.NewDomainInfo("user"),
	}

	regenErr := scaffold.RegenerateFromDomains(dir, modulePath, domains)
	require.NoError(t, regenErr)

	expectedFiles := []string{
		filepath.Join(dir, "cmd", "api", "server.go"),
		filepath.Join(dir, "internal", "infrastructure", "web", "router", "router.go"),
		filepath.Join(dir, "internal", "bootstrap", "container.go"),
		filepath.Join(dir, "internal", "bootstrap", "test_helpers.go"),
	}

	for _, f := range expectedFiles {
		_, statErr := os.Stat(f)
		assert.NoError(t, statErr, "expected file to exist: %s", f)
	}
}

func TestScanDomains_IgnoresFiles(t *testing.T) {
	dir := t.TempDir()
	domainDir := filepath.Join(dir, "internal", "domain")
	mkdirErr := os.MkdirAll(domainDir, 0o750)
	require.NoError(t, mkdirErr)

	// Create a file (not a directory) in internal/domain/
	writeErr := os.WriteFile(filepath.Join(domainDir, "shared.go"), []byte("package domain\n"), 0o600)
	require.NoError(t, writeErr)

	domains, scanErr := scanDomains(dir)
	require.NoError(t, scanErr)
	assert.Empty(t, domains)
}

func TestRegenerateFromDomains_ServerContainsAllImports(t *testing.T) {
	dir := t.TempDir()
	modulePath := "github.com/test/svc"
	createFakeProject(t, dir, modulePath)

	domains := []scaffold.DomainInfo{
		scaffold.NewDomainInfo("order"),
		scaffold.NewDomainInfo("role"),
		scaffold.NewDomainInfo("user"),
	}

	regenErr := scaffold.RegenerateFromDomains(dir, modulePath, domains)
	require.NoError(t, regenErr)

	serverContent, readErr := os.ReadFile(filepath.Join(dir, "cmd", "api", "server.go"))
	require.NoError(t, readErr)
	serverStr := string(serverContent)

	// Count occurrences of each domain handler in the Dependencies return
	for _, d := range domains {
		handlerField := d.Pascal + "Handler:"
		count := strings.Count(serverStr, handlerField)
		assert.GreaterOrEqual(t, count, 1, "expected at least one occurrence of %s", handlerField)
	}
}
