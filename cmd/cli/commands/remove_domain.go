package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jrmarcello/gopherplate/cmd/cli/scaffold"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove project components (domain, endpoint)",
	Long:  "Remove domains or other components from an existing project.",
}

var removeDomainCmd = &cobra.Command{
	Use:   "domain [name]",
	Short: "Remove a domain from the project",
	Long: `Removes all files of a domain: domain entities, use cases, repository, handler, router.

Migration files are NOT deleted (data loss risk) — they are listed but preserved.
Wiring in cmd/api/server.go and internal/infrastructure/web/router/router.go
is not removed automatically; the CLI prints manual cleanup steps.`,
	Args: cobra.ExactArgs(1),
	RunE: runRemoveDomain,
}

func init() {
	removeDomainCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	removeCmd.AddCommand(removeDomainCmd)
}

func runRemoveDomain(cmd *cobra.Command, args []string) error {
	domainName := args[0]

	// 1. Validate domain name and normalize to snake_case
	if validateErr := validateDomainName(domainName); validateErr != nil {
		return validateErr
	}
	snakeName := scaffold.ToSnakeCase(domainName)

	// 2. Resolve project root from cwd
	projectRoot, cwdErr := os.Getwd()
	if cwdErr != nil {
		return fmt.Errorf("resolving working directory: %w", cwdErr)
	}

	// 3. Validate domain exists
	domainDir := filepath.Join(projectRoot, "internal", "domain", snakeName)
	if _, statErr := os.Stat(domainDir); os.IsNotExist(statErr) {
		return fmt.Errorf("domain '%s' not found at %s", snakeName, domainDir)
	}

	// 4. Collect files
	toRemove := collectDomainFiles(projectRoot, snakeName)
	migrationFiles := collectMigrationFiles(projectRoot, snakeName)

	// 5. Confirmation (unless --yes)
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Println("The following files/directories will be DELETED:")
		for _, f := range toRemove {
			fmt.Printf("  - %s\n", f)
		}
		if len(migrationFiles) > 0 {
			fmt.Println("\nThe following migration files will be PRESERVED (data loss risk):")
			for _, f := range migrationFiles {
				fmt.Printf("  - %s\n", f)
			}
		}
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		confirmed, confirmErr := PromptConfirm(reader, "Are you sure?", false)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// 6. Remove files
	for _, f := range toRemove {
		if rmErr := os.RemoveAll(f); rmErr != nil {
			return fmt.Errorf("removing %s: %w", f, rmErr)
		}
	}

	// 7. Summary + manual cleanup instructions
	fmt.Printf("\nDomain '%s' removed (%d items).\n\n", snakeName, len(toRemove))
	fmt.Println("Manual cleanup needed:")
	fmt.Printf("  1. Remove %s wiring from cmd/api/server.go\n", snakeName)
	fmt.Printf("  2. Remove Register%sRoutes from internal/infrastructure/web/router/router.go\n", scaffold.ToPascalCase(snakeName))
	if len(migrationFiles) > 0 {
		fmt.Printf("  3. Consider reverting migration: make migrate-down (preserved: %d file(s))\n", len(migrationFiles))
	}

	return nil
}

// collectDomainFiles lists all files/dirs belonging to a domain (excluding migrations).
// Only existing paths are returned.
func collectDomainFiles(projectRoot, domainName string) []string {
	candidates := []string{
		filepath.Join(projectRoot, "internal", "domain", domainName),
		filepath.Join(projectRoot, "internal", "usecases", domainName),
		filepath.Join(projectRoot, "internal", "infrastructure", "db", "postgres", "repository", domainName+".go"),
		filepath.Join(projectRoot, "internal", "infrastructure", "db", "postgres", "repository", domainName+"_test.go"),
		filepath.Join(projectRoot, "internal", "infrastructure", "web", "handler", domainName+".go"),
		filepath.Join(projectRoot, "internal", "infrastructure", "web", "router", domainName+".go"),
	}

	existing := make([]string, 0, len(candidates))
	for _, p := range candidates {
		if _, statErr := os.Stat(p); statErr == nil {
			existing = append(existing, p)
		}
	}
	return existing
}

// collectMigrationFiles lists migration SQL files matching the domain plural form.
// These are returned for reporting only; the caller preserves them.
func collectMigrationFiles(projectRoot, domainName string) []string {
	migrationDir := filepath.Join(projectRoot, "internal", "infrastructure", "db", "postgres", "migration")
	plural := scaffold.ToPlural(domainName)
	pattern := filepath.Join(migrationDir, "*_create_"+plural+".sql")
	matches, _ := filepath.Glob(pattern)
	return matches
}
