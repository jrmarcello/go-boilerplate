package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRemoveDomainCmdForTest builds a fresh removeDomainCmd-like command
// so flags can be set per-test without leaking across tests.
func runRemoveDomainForTest(t *testing.T, args []string, yes bool) error { //nolint:unparam // yes kept for test flexibility
	t.Helper()
	cmdCopy := *removeDomainCmd
	cmd := &cmdCopy
	// Reset flags by creating a new flag set
	cmd.ResetFlags()
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	if yes {
		require.NoError(t, cmd.Flags().Set("yes", "true"))
	}
	return runRemoveDomain(cmd, args)
}

// withChdir switches into dir for the duration of the test.
func withChdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

// seedDomainTree creates a fake project tree for domain 'order' inside root.
// Returns the list of absolute paths that should be removed.
func seedDomainTree(t *testing.T, root, domain string) []string {
	t.Helper()

	dirs := []string{
		filepath.Join(root, "internal", "domain", domain),
		filepath.Join(root, "internal", "usecases", domain, "interfaces"),
		filepath.Join(root, "internal", "usecases", domain, "dto"),
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "repository"),
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "migration"),
		filepath.Join(root, "internal", "infrastructure", "web", "handler"),
		filepath.Join(root, "internal", "infrastructure", "web", "router"),
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, 0o750))
	}

	files := map[string]string{
		filepath.Join(root, "internal", "domain", domain, "entity.go"):                                       "package " + domain,
		filepath.Join(root, "internal", "usecases", domain, "create.go"):                                     "package " + domain,
		filepath.Join(root, "internal", "usecases", domain, "interfaces", "repository.go"):                   "package interfaces",
		filepath.Join(root, "internal", "usecases", domain, "dto", "create.go"):                              "package dto",
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "repository", domain+".go"):      "package repository",
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "repository", domain+"_test.go"): "package repository",
		filepath.Join(root, "internal", "infrastructure", "web", "handler", domain+".go"):                    "package handler",
		filepath.Join(root, "internal", "infrastructure", "web", "router", domain+".go"):                     "package router",
	}
	for path, content := range files {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}

	// Expected removed items (directories are reported as single entries).
	return []string{
		filepath.Join(root, "internal", "domain", domain),
		filepath.Join(root, "internal", "usecases", domain),
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "repository", domain+".go"),
		filepath.Join(root, "internal", "infrastructure", "db", "postgres", "repository", domain+"_test.go"),
		filepath.Join(root, "internal", "infrastructure", "web", "handler", domain+".go"),
		filepath.Join(root, "internal", "infrastructure", "web", "router", domain+".go"),
	}
}

func TestRunRemoveDomain_Success(t *testing.T) {
	root := t.TempDir()
	domain := "order"
	expected := seedDomainTree(t, root, domain)

	withChdir(t, root)

	err := runRemoveDomainForTest(t, []string{domain}, true)
	require.NoError(t, err)

	for _, p := range expected {
		_, statErr := os.Stat(p)
		assert.True(t, os.IsNotExist(statErr), "expected %s to be deleted, stat err=%v", p, statErr)
	}
}

func TestRunRemoveDomain_NotFound(t *testing.T) {
	root := t.TempDir()
	withChdir(t, root)

	err := runRemoveDomainForTest(t, []string{"nonexistent"}, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunRemoveDomain_PreservesMigrations(t *testing.T) {
	root := t.TempDir()
	domain := "order"
	seedDomainTree(t, root, domain)

	// Create migration files (plural = orders).
	migrationDir := filepath.Join(root, "internal", "infrastructure", "db", "postgres", "migration")
	migrationPath := filepath.Join(migrationDir, "20240101000000_create_orders.sql")
	require.NoError(t, os.WriteFile(migrationPath, []byte("-- +goose Up"), 0o600))

	withChdir(t, root)

	err := runRemoveDomainForTest(t, []string{domain}, true)
	require.NoError(t, err)

	// Migration preserved
	_, statErr := os.Stat(migrationPath)
	assert.NoError(t, statErr, "migration should be preserved")

	// Domain dir removed
	_, domStat := os.Stat(filepath.Join(root, "internal", "domain", domain))
	assert.True(t, os.IsNotExist(domStat))
}

func TestRunRemoveDomain_InvalidName(t *testing.T) {
	root := t.TempDir()
	withChdir(t, root)

	err := runRemoveDomainForTest(t, []string{""}, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestCollectDomainFiles(t *testing.T) {
	root := t.TempDir()
	domain := "order"
	expected := seedDomainTree(t, root, domain)

	got := collectDomainFiles(root, domain)
	assert.ElementsMatch(t, expected, got)
}

func TestCollectDomainFiles_MissingPathsSkipped(t *testing.T) {
	root := t.TempDir()

	// Only create domain dir, nothing else.
	domain := "order"
	require.NoError(t, os.MkdirAll(filepath.Join(root, "internal", "domain", domain), 0o750))

	got := collectDomainFiles(root, domain)
	require.Len(t, got, 1)
	assert.True(t, strings.HasSuffix(got[0], filepath.Join("internal", "domain", domain)))
}

func TestCollectMigrationFiles(t *testing.T) {
	root := t.TempDir()
	migrationDir := filepath.Join(root, "internal", "infrastructure", "db", "postgres", "migration")
	require.NoError(t, os.MkdirAll(migrationDir, 0o750))

	p1 := filepath.Join(migrationDir, "20240101000000_create_orders.sql")
	p2 := filepath.Join(migrationDir, "20240201000000_create_orders.sql")
	unrelated := filepath.Join(migrationDir, "20240301000000_create_users.sql")
	for _, p := range []string{p1, p2, unrelated} {
		require.NoError(t, os.WriteFile(p, []byte("-- +goose Up"), 0o600))
	}

	got := collectMigrationFiles(root, "order")
	assert.ElementsMatch(t, []string{p1, p2}, got)
}
