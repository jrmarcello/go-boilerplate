package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupFakeProject creates a temporary project directory with a go.mod and an
// existing domain layout, chdirs into it, and registers a cleanup that restores
// the original working directory.
func setupFakeProject(t *testing.T, domainName string) string { //nolint:unparam // domainName kept for test flexibility
	t.Helper()

	dir := t.TempDir()

	goMod := "module github.com/test/my-service\n\ngo 1.26.0\n"
	writeErr := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600)
	require.NoError(t, writeErr)

	// Create the domain directory and the usecases/dto subdirectories so they
	// look like a scaffolded domain.
	domainDir := filepath.Join(dir, "internal", "domain", domainName)
	require.NoError(t, os.MkdirAll(domainDir, 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "usecases", domainName, "dto"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "usecases", domainName, "interfaces"), 0o750))

	origDir, getErr := os.Getwd()
	require.NoError(t, getErr)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	return dir
}

func TestRunAddEndpoint_Success(t *testing.T) {
	projectRoot := setupFakeProject(t, "order")

	runErr := runAddEndpoint(nil, []string{"order", "cancel"})
	require.NoError(t, runErr)

	// Verify all 3 files were created.
	expected := []string{
		filepath.Join(projectRoot, "internal", "usecases", "order", "cancel.go"),
		filepath.Join(projectRoot, "internal", "usecases", "order", "dto", "cancel.go"),
		filepath.Join(projectRoot, "internal", "usecases", "order", "cancel_test.go"),
	}
	for _, path := range expected {
		_, statErr := os.Stat(path)
		assert.NoErrorf(t, statErr, "expected file to exist: %s", path)
	}

	// Verify that the rendered use-case file contains the expected identifier.
	content, readErr := os.ReadFile(filepath.Join(projectRoot, "internal", "usecases", "order", "cancel.go"))
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "type CancelUseCase struct")
	assert.Contains(t, string(content), "func NewCancelUseCase")
	assert.Contains(t, string(content), "package order")

	dtoContent, readDTOErr := os.ReadFile(filepath.Join(projectRoot, "internal", "usecases", "order", "dto", "cancel.go"))
	require.NoError(t, readDTOErr)
	assert.Contains(t, string(dtoContent), "CancelInput")
	assert.Contains(t, string(dtoContent), "CancelOutput")
	assert.Contains(t, string(dtoContent), "package dto")
}

func TestRunAddEndpoint_DomainNotFound(t *testing.T) {
	setupFakeProject(t, "order")

	runErr := runAddEndpoint(nil, []string{"nonexistent", "activate"})
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "domain 'nonexistent' not found")
}

func TestRunAddEndpoint_CrudProtected(t *testing.T) {
	setupFakeProject(t, "order")

	crudNames := []string{"create", "get", "update", "delete", "list"}
	for _, name := range crudNames {
		t.Run(name, func(t *testing.T) {
			runErr := runAddEndpoint(nil, []string{"order", name})
			require.Error(t, runErr)
			assert.Contains(t, runErr.Error(), "standard CRUD operation")
		})
	}
}

func TestRunAddEndpoint_InvalidName(t *testing.T) {
	setupFakeProject(t, "order")

	invalid := []string{"123abc", "CamelCase", "with-hyphen", "with space", ""}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			runErr := runAddEndpoint(nil, []string{"order", name})
			require.Error(t, runErr)
			assert.Contains(t, runErr.Error(), "invalid endpoint name")
		})
	}
}

func TestRunAddEndpoint_AlreadyExists(t *testing.T) {
	projectRoot := setupFakeProject(t, "order")

	// Pre-create the file to trigger the duplicate check.
	existing := filepath.Join(projectRoot, "internal", "usecases", "order", "cancel.go")
	require.NoError(t, os.WriteFile(existing, []byte("package order\n"), 0o600))

	runErr := runAddEndpoint(nil, []string{"order", "cancel"})
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "already exists")
}

func TestRunAddEndpoint_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	origDir, getErr := os.Getwd()
	require.NoError(t, getErr)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	runErr := runAddEndpoint(nil, []string{"order", "cancel"})
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "detecting module path")
}
