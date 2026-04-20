package shared

// Shared semantic attribute keys recorded on spans by use cases via
// ClassifyError. Centralizing the strings prevents drift between domains
// (e.g. user vs role) and keeps the trace vocabulary consistent for
// downstream analysis tools.
const (
	// AttrKeyAppResult labels expected business outcomes such as "not_found"
	// or "duplicate_email". The span status remains Unset/Ok — the attribute
	// is the signal, not the status.
	AttrKeyAppResult = "app.result"

	// AttrKeyAppValidationError labels expected validation failures (e.g.
	// invalid email, missing field). The value is typically the underlying
	// error message and the span status remains Unset/Ok.
	AttrKeyAppValidationError = "app.validation_error"
)
