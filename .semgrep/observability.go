// Semgrep test fixture for observability.yml — consumed by `semgrep --test .semgrep/`.
// Marker comments:
//   ruleid: <rule-id>   → the next line MUST match the rule
//   ok:     <rule-id>   → the next line MUST NOT match the rule
//
// Scenarios covered:
//   TC-UC-90  flow slog.* call inside a use-case path → rule fires
//   TC-UC-91  logutil.LogInfo on the replay (business-flow) path → rule fires
//   TC-UC-92  logutil.LogWarn in the fail-open infra-unreachable branch
//             (annotated with nosemgrep) → rule does NOT fire
//   TC-UC-93  slog.Info in cmd/api startup path is out of scope (different file
//             path; asserted by path exclusion — this fixture covers the
//             in-scope cases only).

//go:build semgrep_fixture

package semgrep_fixture_observability

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jrmarcello/gopherplate/pkg/logutil"
)

// TC-UC-90: plain slog call in a use-case file — rule must fire.
func flowWithSlog(ctx context.Context, key string) {
	// ruleid: gopherplate-usecase-no-slog-in-flow
	slog.Debug("cache hit", "key", key)
	// ruleid: gopherplate-usecase-no-slog-in-flow
	slog.Warn("failed to cache user", "key", key)
}

// TC-UC-91: logutil on the replay path (business flow, not infra failure) — fires.
func idempotencyReplay(ctx context.Context, key string, statusCode int) {
	// ruleid: gopherplate-usecase-no-slog-in-flow
	logutil.LogInfo(ctx, "idempotency replay",
		"idempotency_key", key, "status_code", statusCode)
}

// TC-UC-92: fail-open infra-unreachable branch with nosemgrep pragma —
// semgrep honors `// nosemgrep: <rule-id>` natively, so the rule does NOT fire.
func idempotencyStoreUnavailable(ctx context.Context, key string) {
	lockErr := errors.New("redis unreachable")

	// Fail-open per spec REQ-4 (logs-vs-traces posture): infra branch keeps
	// the log alongside the span event for operator visibility.
	// ok: gopherplate-usecase-no-slog-in-flow
	// nosemgrep: gopherplate-usecase-no-slog-in-flow
	logutil.LogWarn(ctx, "idempotency store unavailable, proceeding without",
		"error", lockErr.Error(), "idempotency_key", key)
}
