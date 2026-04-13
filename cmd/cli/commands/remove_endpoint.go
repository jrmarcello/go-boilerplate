package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jrmarcello/gopherplate/cmd/cli/scaffold"
)

var removeEndpointCmd = &cobra.Command{
	Use:   "endpoint [domain] [endpoint-name]",
	Short: "Remove a custom endpoint from a domain",
	Long: `Removes a custom (non-CRUD) endpoint from an existing domain.

Deletes the use case file, DTO file, and test file for the endpoint.
Standard CRUD endpoints (create, get, update, delete, list) cannot be
removed individually — use 'gopherplate remove domain' instead.`,
	Args: cobra.ExactArgs(2),
	RunE: runRemoveEndpoint,
}

func init() {
	removeEndpointCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	removeCmd.AddCommand(removeEndpointCmd)
}

func runRemoveEndpoint(cmd *cobra.Command, args []string) error {
	domainName := args[0]
	endpointName := args[1]

	// 1. Detect module path via go.mod (validates we're in a Go project)
	if _, detectErr := detectModulePath(); detectErr != nil {
		return fmt.Errorf("detecting module path: %w (are you in a Go project directory?)", detectErr)
	}

	// 2. Validate domain name and normalize to snake_case
	if validateErr := validateDomainName(domainName); validateErr != nil {
		return validateErr
	}
	snakeDomain := scaffold.ToSnakeCase(domainName)

	// 3. Validate domain exists
	domainDir := filepath.Join("internal", "domain", snakeDomain)
	if _, statErr := os.Stat(domainDir); os.IsNotExist(statErr) {
		return fmt.Errorf("domain '%s' not found at %s", snakeDomain, domainDir)
	}

	// 4. Normalize endpoint name
	snakeEndpoint := scaffold.ToSnakeCase(endpointName)

	// 5. CRUD protection
	if crudEndpoints[snakeEndpoint] {
		return fmt.Errorf("cannot remove standard CRUD endpoint '%s'. Use 'gopherplate remove domain %s' to remove the entire domain", snakeEndpoint, snakeDomain)
	}

	// 6. Validate endpoint use case file exists
	usecaseFile := filepath.Join("internal", "usecases", snakeDomain, snakeEndpoint+".go")
	if _, statErr := os.Stat(usecaseFile); os.IsNotExist(statErr) {
		return fmt.Errorf("endpoint '%s' not found: %s does not exist", snakeEndpoint, usecaseFile)
	}

	// 7. Build file list (use case, DTO, test)
	filesToRemove := []string{
		filepath.Join("internal", "usecases", snakeDomain, snakeEndpoint+".go"),
		filepath.Join("internal", "usecases", snakeDomain, "dto", snakeEndpoint+".go"),
		filepath.Join("internal", "usecases", snakeDomain, snakeEndpoint+"_test.go"),
	}

	// 8. Confirmation (unless --yes)
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Println("The following files will be removed:")
		for _, f := range filesToRemove {
			fmt.Printf("  - %s\n", f)
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

	// 9. Remove each file (ignore missing files gracefully)
	removed := make([]string, 0, len(filesToRemove))
	for _, f := range filesToRemove {
		removeErr := os.Remove(f)
		if removeErr != nil {
			if os.IsNotExist(removeErr) {
				continue // skip missing files gracefully
			}
			return fmt.Errorf("removing %s: %w", f, removeErr)
		}
		removed = append(removed, f)
	}

	// 10. Print success message
	fmt.Printf("\nEndpoint '%s' removed from domain '%s' (%d file(s)):\n", snakeEndpoint, snakeDomain, len(removed))
	for _, f := range removed {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}
