package scaffold

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateData(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModulePath = "github.com/org/my-service"

	t.Run("simple domain name", func(t *testing.T) {
		data := NewTemplateData("order", cfg)

		assert.Equal(t, "order", data.DomainName)
		assert.Equal(t, "orders", data.DomainNamePlural)
		assert.Equal(t, "Order", data.DomainNamePascal)
		assert.Equal(t, "order", data.DomainNameCamel)
		assert.Equal(t, "order", data.DomainNameSnake)
		assert.Equal(t, "github.com/org/my-service", data.ModulePath)
	})

	t.Run("multi-word domain name", func(t *testing.T) {
		data := NewTemplateData("order_item", cfg)

		assert.Equal(t, "order_item", data.DomainName)
		assert.Equal(t, "order_items", data.DomainNamePlural)
		assert.Equal(t, "OrderItem", data.DomainNamePascal)
		assert.Equal(t, "orderItem", data.DomainNameCamel)
		assert.Equal(t, "order_item", data.DomainNameSnake)
	})

	t.Run("PascalCase input is normalized", func(t *testing.T) {
		data := NewTemplateData("OrderItem", cfg)

		assert.Equal(t, "order_item", data.DomainName)
		assert.Equal(t, "order_items", data.DomainNamePlural)
		assert.Equal(t, "OrderItem", data.DomainNamePascal)
		assert.Equal(t, "orderItem", data.DomainNameCamel)
		assert.Equal(t, "order_item", data.DomainNameSnake)
	})
}

func TestRenderTemplate(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModulePath = "github.com/org/my-service"
	data := NewTemplateData("order", cfg)

	t.Run("renders simple template", func(t *testing.T) {
		tmpl := `package {{.DomainName}}`
		result, renderErr := RenderTemplate(tmpl, data)
		require.NoError(t, renderErr)
		assert.Equal(t, "package order", result)
	})

	t.Run("renders with domain name variants", func(t *testing.T) {
		tmpl := `type {{.DomainNamePascal}} struct {}
func new{{.DomainNamePascal}}() *{{.DomainNamePascal}} { return nil }
`
		result, renderErr := RenderTemplate(tmpl, data)
		require.NoError(t, renderErr)
		assert.Contains(t, result, "type Order struct")
		assert.Contains(t, result, "func newOrder()")
	})

	t.Run("renders with template functions", func(t *testing.T) {
		tmpl := `table: {{plural .DomainName}}`
		result, renderErr := RenderTemplate(tmpl, data)
		require.NoError(t, renderErr)
		assert.Equal(t, "table: orders", result)
	})

	t.Run("renders with module path", func(t *testing.T) {
		tmpl := `import "{{.ModulePath}}/internal/domain/{{.DomainName}}"`
		result, renderErr := RenderTemplate(tmpl, data)
		require.NoError(t, renderErr)
		assert.Equal(t, `import "github.com/org/my-service/internal/domain/order"`, result)
	})

	t.Run("returns error for invalid template", func(t *testing.T) {
		tmpl := `{{.Invalid`
		_, renderErr := RenderTemplate(tmpl, data)
		assert.Error(t, renderErr)
	})

	t.Run("returns error for missing field", func(t *testing.T) {
		tmpl := `{{.NonExistentField}}`
		_, renderErr := RenderTemplate(tmpl, data)
		assert.Error(t, renderErr)
	})
}

func TestRenderTemplateFile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModulePath = "github.com/org/my-service"
	data := NewTemplateData("order", cfg)

	t.Run("writes rendered content to file", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "domain", "order", "entity.go")

		tmpl := `package {{.DomainName}}

type {{.DomainNamePascal}} struct {
	ID string
}
`
		renderErr := RenderTemplateFile(tmpl, data, outputPath)
		require.NoError(t, renderErr)

		content, readErr := os.ReadFile(outputPath)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "package order")
		assert.Contains(t, string(content), "type Order struct")
	})

	t.Run("creates parent directories", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "deep", "nested", "dir", "file.go")

		renderErr := RenderTemplateFile("package x", data, outputPath)
		require.NoError(t, renderErr)

		_, statErr := os.Stat(outputPath)
		assert.NoError(t, statErr)
	})
}

func TestRenderFS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModulePath = "github.com/org/my-service"
	data := NewTemplateData("order", cfg)

	t.Run("renders all tmpl files and strips extension", func(t *testing.T) {
		templates := fstest.MapFS{
			"entity.go.tmpl": &fstest.MapFile{
				Data: []byte(`package {{.DomainName}}

type {{.DomainNamePascal}} struct {}
`),
			},
			"errors.go.tmpl": &fstest.MapFile{
				Data: []byte(`package {{.DomainName}}

var ErrNotFound = fmt.Errorf("{{.DomainName}} not found")
`),
			},
		}

		outputDir := t.TempDir()
		renderErr := RenderFS(templates, data, outputDir)
		require.NoError(t, renderErr)

		// Check entity.go (not entity.go.tmpl)
		entityContent, readErr := os.ReadFile(filepath.Join(outputDir, "entity.go"))
		require.NoError(t, readErr)
		assert.Contains(t, string(entityContent), "package order")
		assert.Contains(t, string(entityContent), "type Order struct")

		// Check errors.go
		errorsContent, readErr2 := os.ReadFile(filepath.Join(outputDir, "errors.go"))
		require.NoError(t, readErr2)
		assert.Contains(t, string(errorsContent), "order not found")
	})

	t.Run("handles nested template directories", func(t *testing.T) {
		templates := fstest.MapFS{
			"internal/domain/entity.go.tmpl": &fstest.MapFile{
				Data: []byte(`package {{.DomainName}}`),
			},
		}

		outputDir := t.TempDir()
		renderErr := RenderFS(templates, data, outputDir)
		require.NoError(t, renderErr)

		content, readErr := os.ReadFile(filepath.Join(outputDir, "internal", "domain", "entity.go"))
		require.NoError(t, readErr)
		assert.Equal(t, "package order", string(content))
	})

	t.Run("passes through non-tmpl files", func(t *testing.T) {
		templates := fstest.MapFS{
			"README.md": &fstest.MapFile{
				Data: []byte(`# {{.DomainNamePascal}} Service`),
			},
		}

		outputDir := t.TempDir()
		renderErr := RenderFS(templates, data, outputDir)
		require.NoError(t, renderErr)

		content, readErr := os.ReadFile(filepath.Join(outputDir, "README.md"))
		require.NoError(t, readErr)
		assert.Equal(t, "# Order Service", string(content))
	})
}
