package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jrmarcello/gopherplate/cmd/cli/scaffold"
)

var wiringYes bool

var wiringCmd = &cobra.Command{
	Use:   "wiring",
	Short: "Regenerate wiring files (server.go, router.go, container.go, test_helpers.go)",
	Long: `Detects all domains in internal/domain/ and regenerates the 4 wiring files
to reflect the current set of domains:
  - cmd/api/server.go
  - internal/infrastructure/web/router/router.go
  - internal/bootstrap/container.go
  - internal/bootstrap/test_helpers.go

Each domain must have a handler, router, and repository file to be included.`,
	Args: cobra.NoArgs,
	RunE: runWiring,
}

func init() {
	wiringCmd.Flags().BoolVarP(&wiringYes, "yes", "y", false, "Skip confirmation prompt")
}

func runWiring(_ *cobra.Command, _ []string) error {
	// 1. Detect module path from go.mod
	modulePath, detectErr := detectModulePath()
	if detectErr != nil {
		return fmt.Errorf("detecting module path: %w (are you in a Go project directory?)", detectErr)
	}

	// 2. Scan internal/domain/ for subdirectories
	domains, scanErr := scanDomains(".")
	if scanErr != nil {
		return fmt.Errorf("scanning domains: %w", scanErr)
	}

	// 3. Print plan
	fmt.Println()
	if len(domains) == 0 {
		fmt.Println("Detected 0 domains.")
	} else {
		names := make([]string, 0, len(domains))
		for _, d := range domains {
			names = append(names, d.Name)
		}
		fmt.Printf("Detected %d domains: [%s]\n", len(domains), strings.Join(names, ", "))
	}
	fmt.Println("Will regenerate: server.go, router.go, container.go, test_helpers.go")
	fmt.Println()

	// 4. Confirm (default YES)
	if !wiringYes {
		reader := bufio.NewReader(os.Stdin)
		confirmed, confirmErr := PromptConfirm(reader, "Continue?", true)
		if confirmErr != nil {
			return fmt.Errorf("reading confirmation: %w", confirmErr)
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// 5. Regenerate files
	projectDir, wdErr := os.Getwd()
	if wdErr != nil {
		return fmt.Errorf("getting working directory: %w", wdErr)
	}

	if regenErr := scaffold.RegenerateFromDomains(projectDir, modulePath, domains); regenErr != nil {
		return fmt.Errorf("regenerating wiring files: %w", regenErr)
	}

	// 6. Print success
	fmt.Println()
	fmt.Println("Wiring files regenerated successfully:")
	fmt.Println("  - cmd/api/server.go")
	fmt.Println("  - internal/infrastructure/web/router/router.go")
	fmt.Println("  - internal/bootstrap/container.go")
	fmt.Println("  - internal/bootstrap/test_helpers.go")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: go mod tidy")
	fmt.Println("  2. Run: go build ./...")
	fmt.Println()

	return nil
}

// scanDomains scans internal/domain/ for subdirectories that have matching
// handler, router, and repository files. Returns sorted DomainInfo list.
func scanDomains(projectDir string) ([]scaffold.DomainInfo, error) {
	domainDir := filepath.Join(projectDir, "internal", "domain")
	entries, readErr := os.ReadDir(domainDir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading internal/domain: %w", readErr)
	}

	var domains []scaffold.DomainInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Verify the domain has all required infrastructure files
		if !hasDomainInfrastructure(projectDir, name) {
			continue
		}

		domains = append(domains, scaffold.DetectDomainInfo(projectDir, name))
	}

	// Sort alphabetically for deterministic output
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Name < domains[j].Name
	})

	return domains, nil
}

// hasDomainInfrastructure checks that a domain has handler, router, and repository files.
func hasDomainInfrastructure(projectDir, domainName string) bool {
	handlerPath := filepath.Join(projectDir, "internal", "infrastructure", "web", "handler", domainName+".go")
	routerPath := filepath.Join(projectDir, "internal", "infrastructure", "web", "router", domainName+".go")
	repoPath := filepath.Join(projectDir, "internal", "infrastructure", "db", "postgres", "repository", domainName+".go")

	if _, statErr := os.Stat(handlerPath); statErr != nil {
		return false
	}
	if _, statErr := os.Stat(routerPath); statErr != nil {
		return false
	}
	if _, statErr := os.Stat(repoPath); statErr != nil {
		return false
	}

	return true
}
