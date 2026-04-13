package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/jrmarcello/gopherplate/cmd/cli/scaffold"
	domaintmpl "github.com/jrmarcello/gopherplate/cmd/cli/templates/domain"
)

var addEndpointCmd = &cobra.Command{
	Use:   "endpoint [domain] [name]",
	Short: "Add a custom endpoint to an existing domain",
	Long: `Scaffolds a custom endpoint (use case + DTO + test) inside an existing domain.
Does NOT modify the handler or router — prints manual wiring instructions.

Example:
  gopherplate add endpoint order cancel
  gopherplate add endpoint user activate`,
	Args: cobra.ExactArgs(2),
	RunE: runAddEndpoint,
}

// endpointNamePattern validates endpoint names: must be snake_case and start with a letter.
var endpointNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// crudEndpoints are protected names (they already exist as standard CRUD operations).
var crudEndpoints = map[string]bool{
	"create": true,
	"get":    true,
	"update": true,
	"delete": true,
	"list":   true,
}

func runAddEndpoint(_ *cobra.Command, args []string) error {
	domainName := args[0]
	endpointName := args[1]

	projectRoot, getwdErr := os.Getwd()
	if getwdErr != nil {
		return fmt.Errorf("getting working directory: %w", getwdErr)
	}

	// 1. Detect module path
	modulePath, detectErr := detectModulePath()
	if detectErr != nil {
		return fmt.Errorf("detecting module path: %w (are you in a Go project directory?)", detectErr)
	}

	// 2. Validate endpoint name format
	if !endpointNamePattern.MatchString(endpointName) {
		return fmt.Errorf("invalid endpoint name '%s' (must be snake_case, start with a letter, contain only lowercase letters, digits, and underscores)", endpointName)
	}

	// 3. Check it is not a reserved CRUD name
	if crudEndpoints[endpointName] {
		return fmt.Errorf("endpoint '%s' is a standard CRUD operation — it already exists for each domain", endpointName)
	}

	// 4. Validate domain exists
	snakeDomain := scaffold.ToSnakeCase(domainName)
	domainDir := filepath.Join(projectRoot, "internal", "domain", snakeDomain)
	if _, statErr := os.Stat(domainDir); os.IsNotExist(statErr) {
		return fmt.Errorf("domain '%s' not found at %s", snakeDomain, domainDir)
	}

	// 5. Ensure the endpoint use-case file does not already exist
	endpointFile := filepath.Join(projectRoot, "internal", "usecases", snakeDomain, endpointName+".go")
	if _, statErr := os.Stat(endpointFile); statErr == nil {
		return fmt.Errorf("endpoint '%s' already exists in domain '%s' (%s)", endpointName, snakeDomain, endpointFile)
	}

	// 6. Build template data (reuses the domain TemplateData plus endpoint fields)
	cfg := scaffold.Config{ModulePath: modulePath}
	data := scaffold.NewTemplateData(domainName, cfg).WithEndpoint(endpointName)

	// 7. Render the three templates
	mappings := map[string]string{
		"endpoint_usecase.go.tmpl":      filepath.Join(projectRoot, "internal", "usecases", snakeDomain, endpointName+".go"),
		"endpoint_dto.go.tmpl":          filepath.Join(projectRoot, "internal", "usecases", snakeDomain, "dto", endpointName+".go"),
		"endpoint_usecase_test.go.tmpl": filepath.Join(projectRoot, "internal", "usecases", snakeDomain, endpointName+"_test.go"),
	}

	fmt.Printf("\nScaffolding endpoint '%s' in domain '%s'...\n\n", endpointName, snakeDomain)

	createdFiles := make([]string, 0, len(mappings))
	for tmplName, outputPath := range mappings {
		tmplContent, readErr := fs.ReadFile(domaintmpl.Templates, tmplName)
		if readErr != nil {
			return fmt.Errorf("reading template %s: %w", tmplName, readErr)
		}

		rendered, renderErr := scaffold.RenderTemplate(string(tmplContent), data)
		if renderErr != nil {
			return fmt.Errorf("rendering %s: %w", tmplName, renderErr)
		}

		dirPath := filepath.Dir(outputPath)
		if mkdirErr := os.MkdirAll(dirPath, 0o750); mkdirErr != nil {
			return fmt.Errorf("creating directory %s: %w", dirPath, mkdirErr)
		}

		if writeErr := os.WriteFile(outputPath, []byte(rendered), 0o600); writeErr != nil {
			return fmt.Errorf("writing file %s: %w", outputPath, writeErr)
		}

		// Print a project-relative path for clarity.
		relPath, relErr := filepath.Rel(projectRoot, outputPath)
		if relErr != nil {
			relPath = outputPath
		}
		createdFiles = append(createdFiles, relPath)
		fmt.Printf("  \u2713 %s\n", relPath)
	}

	fmt.Printf("\n%d files created.\n\n", len(createdFiles))
	printEndpointWiringInstructions(data)

	return nil
}

// printEndpointWiringInstructions prints the manual wiring steps for the new endpoint.
func printEndpointWiringInstructions(data scaffold.TemplateData) {
	fmt.Println("Manual wiring needed:")
	fmt.Println()
	fmt.Printf("  1. Add a handler method in internal/infrastructure/web/handler/%s.go:\n", data.DomainNameSnake)
	fmt.Printf("     func (h *%sHandler) %s(c *gin.Context) { ... }\n", data.DomainNamePascal, data.EndpointNamePascal)
	fmt.Println()
	fmt.Printf("  2. Add the route in internal/infrastructure/web/router/%s.go:\n", data.DomainNameSnake)
	fmt.Printf("     rg.POST(\"/%s/:id/%s\", h.%s)\n", data.DomainNamePlural, data.EndpointNameSnake, data.EndpointNamePascal)
	fmt.Println()
	fmt.Printf("  3. Add the use-case field to %sHandler and wire it in cmd/api/server.go:\n", data.DomainNamePascal)
	fmt.Printf("     %sUC *%suc.%sUseCase\n", data.EndpointNamePascal, data.DomainNameCamel, data.EndpointNamePascal)
}
