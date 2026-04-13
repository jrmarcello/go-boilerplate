package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runRemoveEndpointForTest builds a fresh removeEndpointCmd-like command
// so flags can be set per-test without leaking across tests.
func runRemoveEndpointForTest(t *testing.T, args []string, yes bool) error {
	t.Helper()
	cmdCopy := *removeEndpointCmd
	cmd := &cmdCopy
	cmd.ResetFlags()
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	if yes {
		require.NoError(t, cmd.Flags().Set("yes", "true"))
	}
	return runRemoveEndpoint(cmd, args)
}

// seedEndpointFiles creates a minimal project structure with a domain and
// endpoint files for testing. Returns the list of files that should be removed.
func seedEndpointFiles(t *testing.T, root, domain, endpoint string) []string {
	t.Helper()

	// Create go.mod
	goModContent := "module github.com/test/my-service\n\ngo 1.25.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goModContent), 0o600))

	// Create domain dir (proves domain exists)
	domainDir := filepath.Join(root, "internal", "domain", domain)
	require.NoError(t, os.MkdirAll(domainDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, "entity.go"), []byte("package "+domain), 0o600))

	// Create use cases dir with CRUD files + custom endpoint
	usecasesDir := filepath.Join(root, "internal", "usecases", domain)
	require.NoError(t, os.MkdirAll(usecasesDir, 0o750))
	for _, crud := range []string{"create", "get", "update", "delete", "list"} {
		require.NoError(t, os.WriteFile(filepath.Join(usecasesDir, crud+".go"), []byte("package "+domain), 0o600))
	}

	// Create custom endpoint files
	require.NoError(t, os.WriteFile(filepath.Join(usecasesDir, endpoint+".go"), []byte("package "+domain), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(usecasesDir, endpoint+"_test.go"), []byte("package "+domain), 0o600))

	// Create DTO dir with custom endpoint DTO
	dtoDir := filepath.Join(usecasesDir, "dto")
	require.NoError(t, os.MkdirAll(dtoDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dtoDir, endpoint+".go"), []byte("package dto"), 0o600))

	return []string{
		filepath.Join(usecasesDir, endpoint+".go"),
		filepath.Join(dtoDir, endpoint+".go"),
		filepath.Join(usecasesDir, endpoint+"_test.go"),
	}
}

// TC-U-21: happy path - remove endpoint with --yes removes 3 files
func TestRunRemoveEndpoint_Success(t *testing.T) {
	root := t.TempDir()
	domain := "order"
	endpoint := "cancel"

	expectedFiles := seedEndpointFiles(t, root, domain, endpoint)

	// Verify files exist before removal
	for _, f := range expectedFiles {
		_, statErr := os.Stat(f)
		require.NoError(t, statErr, "expected %s to exist before removal", f)
	}

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	removeErr := runRemoveEndpointForTest(t, []string{domain, endpoint}, true)
	require.NoError(t, removeErr)

	// All 3 files should be deleted
	for _, f := range expectedFiles {
		_, statErr := os.Stat(f)
		assert.True(t, os.IsNotExist(statErr), "expected %s to be deleted", f)
	}

	// CRUD files should still exist
	usecasesDir := filepath.Join(root, "internal", "usecases", domain)
	for _, crud := range []string{"create", "get", "update", "delete", "list"} {
		_, statErr := os.Stat(filepath.Join(usecasesDir, crud+".go"))
		assert.NoError(t, statErr, "CRUD file %s.go should NOT be deleted", crud)
	}
}

// TC-U-22: CRUD protection - each of the 5 CRUD names returns error
func TestRunRemoveEndpoint_CRUDProtection(t *testing.T) {
	crudNames := []string{"create", "get", "update", "delete", "list"}

	for _, name := range crudNames {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			domain := "order"

			// Seed minimal project
			goModContent := "module github.com/test/my-service\n\ngo 1.25.0\n"
			require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goModContent), 0o600))
			domainDir := filepath.Join(root, "internal", "domain", domain)
			require.NoError(t, os.MkdirAll(domainDir, 0o750))
			require.NoError(t, os.WriteFile(filepath.Join(domainDir, "entity.go"), []byte("package "+domain), 0o600))

			prev, cwdErr := os.Getwd()
			require.NoError(t, cwdErr)
			require.NoError(t, os.Chdir(root))
			t.Cleanup(func() { _ = os.Chdir(prev) })

			removeErr := runRemoveEndpointForTest(t, []string{domain, name}, true)
			require.Error(t, removeErr)
			assert.Contains(t, removeErr.Error(), "cannot remove standard CRUD endpoint")
			assert.Contains(t, removeErr.Error(), name)
			assert.Contains(t, removeErr.Error(), "gopherplate remove domain")
		})
	}
}

// TC-U-23: user answers N in confirmation - nothing deleted
func TestRunRemoveEndpoint_ConfirmationAbort(t *testing.T) {
	root := t.TempDir()
	domain := "order"
	endpoint := "cancel"

	expectedFiles := seedEndpointFiles(t, root, domain, endpoint)

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	// Mock stdin with "n\n"
	origStdin := os.Stdin
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	_, writeErr := w.WriteString("n\n")
	require.NoError(t, writeErr)
	require.NoError(t, w.Close())
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	// Run WITHOUT --yes, so confirmation prompt is triggered
	removeErr := runRemoveEndpointForTest(t, []string{domain, endpoint}, false)
	require.NoError(t, removeErr) // "Aborted." is not an error, just a message

	// All files should still exist (nothing was deleted)
	for _, f := range expectedFiles {
		_, statErr := os.Stat(f)
		assert.NoError(t, statErr, "expected %s to still exist after abort", f)
	}
}

func TestRunRemoveEndpoint_DomainNotFound(t *testing.T) {
	root := t.TempDir()

	goModContent := "module github.com/test/my-service\n\ngo 1.25.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goModContent), 0o600))

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	removeErr := runRemoveEndpointForTest(t, []string{"nonexistent", "cancel"}, true)
	require.Error(t, removeErr)
	assert.Contains(t, removeErr.Error(), "not found")
}

func TestRunRemoveEndpoint_EndpointNotFound(t *testing.T) {
	root := t.TempDir()

	goModContent := "module github.com/test/my-service\n\ngo 1.25.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goModContent), 0o600))
	domainDir := filepath.Join(root, "internal", "domain", "order")
	require.NoError(t, os.MkdirAll(domainDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, "entity.go"), []byte("package order"), 0o600))
	usecasesDir := filepath.Join(root, "internal", "usecases", "order")
	require.NoError(t, os.MkdirAll(usecasesDir, 0o750))

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	removeErr := runRemoveEndpointForTest(t, []string{"order", "cancel"}, true)
	require.Error(t, removeErr)
	assert.Contains(t, removeErr.Error(), "not found")
	assert.Contains(t, removeErr.Error(), "cancel")
}

func TestRunRemoveEndpoint_MissingFilesGraceful(t *testing.T) {
	root := t.TempDir()

	goModContent := "module github.com/test/my-service\n\ngo 1.25.0\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goModContent), 0o600))
	domainDir := filepath.Join(root, "internal", "domain", "order")
	require.NoError(t, os.MkdirAll(domainDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, "entity.go"), []byte("package order"), 0o600))

	// Only create the use case file (not the DTO or test file)
	usecasesDir := filepath.Join(root, "internal", "usecases", "order")
	require.NoError(t, os.MkdirAll(usecasesDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(usecasesDir, "cancel.go"), []byte("package order"), 0o600))

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	removeErr := runRemoveEndpointForTest(t, []string{"order", "cancel"}, true)
	require.NoError(t, removeErr)

	// The use case file should be removed
	_, statErr := os.Stat(filepath.Join(usecasesDir, "cancel.go"))
	assert.True(t, os.IsNotExist(statErr), "cancel.go should be deleted")
}

func TestRunRemoveEndpoint_NoGoMod(t *testing.T) {
	root := t.TempDir()

	prev, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	removeErr := runRemoveEndpointForTest(t, []string{"order", "cancel"}, true)
	require.Error(t, removeErr)
	assert.Contains(t, removeErr.Error(), "detecting module path")
}
