package scaffold

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaintmpl "github.com/jrmarcello/gopherplate/cmd/cli/templates/domain"
)

// TC-UC-80 / TC-UC-81 / TC-UC-84: the scaffold-rendered use-case errors
// file MUST emit []ucshared.ExpectedError referencing the shared AttrKey*
// constants — not []error and not raw "app.result" string literals. Guarantees
// every service created via `gopherplate new` / `gopherplate add domain`
// starts aligned with the observability posture in
// docs/guides/observability.md without the operator having to remember to
// edit the generated files.
func TestObservabilityTemplate_UsecaseErrors_EmitsExpectedErrorStruct(t *testing.T) {
	cfg := Config{ModulePath: "github.com/test/my-service"}
	data := NewTemplateData("product", cfg)

	tmplBytes, readErr := fs.ReadFile(domaintmpl.Templates, "usecase_errors.go.tmpl")
	require.NoError(t, readErr)

	rendered, renderErr := RenderTemplate(string(tmplBytes), data)
	require.NoError(t, renderErr)

	// Must declare []ucshared.ExpectedError, not []error
	assert.Contains(t, rendered, "[]ucshared.ExpectedError",
		"usecase_errors.go must declare ExpectedError slices, not []error; see REQ-2/REQ-6")
	assert.NotContains(t, rendered, "= []error{",
		"legacy []error{...} pattern must not appear in generated code")

	// Must reference shared constants, not raw string literals (REQ-7)
	assert.Contains(t, rendered, "ucshared.AttrKeyAppResult",
		"generated errors.go must reference ucshared.AttrKeyAppResult constant")
	assert.NotContains(t, rendered, `"app.result"`,
		"raw \"app.result\" literal must not appear — use ucshared.AttrKeyAppResult")
	assert.NotContains(t, rendered, `"app.validation_error"`,
		"raw \"app.validation_error\" literal must not appear — use ucshared.AttrKeyAppValidationError")

	// Must import the shared package as ucshared
	assert.Contains(t, rendered, `ucshared "github.com/test/my-service/internal/usecases/shared"`,
		"generated errors.go must import internal/usecases/shared as ucshared")

	// TC-UC-84: the list-has-no-expected-errors comment must be preserved
	assert.Contains(t, rendered, "listExpectedErrors is intentionally nil",
		"the documenting comment about list use case must survive the refactor")

	// All four expected-error slices must be present, each mapping a sentinel
	// to AttrKeyAppResult. The AttrValue strings are implementation details
	// (not in the REQ's wording) — assert the structural shape is correct.
	for _, name := range []string{"createExpectedErrors", "getExpectedErrors", "updateExpectedErrors", "deleteExpectedErrors"} {
		require.True(t, strings.Contains(rendered, name),
			"missing expected-errors slice: %s", name)
	}
}

// TC-UC-81: the rendered create/get/update/delete/list use cases still
// invoke ucshared.ClassifyError with the same call shape; only the slice
// element type changed. Guarantees no drift between the errors.go slices
// and the call sites.
func TestObservabilityTemplate_UseCases_CallClassifyErrorWithExpectedErrorsSlice(t *testing.T) {
	cfg := Config{ModulePath: "github.com/test/my-service"}
	data := NewTemplateData("product", cfg)

	useCaseTemplates := []string{
		"create_usecase.go.tmpl",
		"get_usecase.go.tmpl",
		"update_usecase.go.tmpl",
		"delete_usecase.go.tmpl",
	}
	for _, name := range useCaseTemplates {
		t.Run(name, func(t *testing.T) {
			tmplBytes, readErr := fs.ReadFile(domaintmpl.Templates, name)
			require.NoError(t, readErr)
			rendered, renderErr := RenderTemplate(string(tmplBytes), data)
			require.NoError(t, renderErr)

			assert.Contains(t, rendered, "ucshared.ClassifyError(span,",
				"%s must call ucshared.ClassifyError", name)
			assert.NotContains(t, rendered, "[]error{",
				"%s must not inline a []error literal — expected-errors come from errors.go", name)
		})
	}
}
