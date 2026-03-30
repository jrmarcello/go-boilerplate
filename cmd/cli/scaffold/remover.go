package scaffold

import (
	"os"
	"path/filepath"
)

// FeatureFiles maps features to their file/directory paths (relative to project root).
// When a feature is disabled, these paths are removed.
var FeatureFiles = map[string][]string{
	"redis": {
		"pkg/cache",
		"pkg/idempotency",
	},
	"idempotency": {
		"pkg/idempotency",
		"internal/infrastructure/web/middleware/idempotency.go",
	},
	"auth": {
		"internal/infrastructure/web/middleware/service_key.go",
	},
	"examples": {
		"internal/domain/user",
		"internal/domain/role",
		"internal/usecases/user",
		"internal/usecases/role",
		"internal/infrastructure/db/postgres/repository/user.go",
		"internal/infrastructure/db/postgres/repository/role.go",
		"internal/infrastructure/web/handler/user.go",
		"internal/infrastructure/web/handler/role.go",
		"internal/infrastructure/web/router/user.go",
		"internal/infrastructure/web/router/role.go",
		"internal/infrastructure/telemetry",
	},
}

// RemoveDisabledFeatures removes files and directories for disabled features.
func RemoveDisabledFeatures(projectDir string, cfg Config) error {
	var toRemove []string

	if !cfg.Redis {
		toRemove = append(toRemove, FeatureFiles["redis"]...)
	}
	if !cfg.Idempotency {
		toRemove = append(toRemove, FeatureFiles["idempotency"]...)
	}
	if !cfg.Auth {
		toRemove = append(toRemove, FeatureFiles["auth"]...)
	}
	if !cfg.KeepExamples {
		toRemove = append(toRemove, FeatureFiles["examples"]...)
	}

	for _, rel := range toRemove {
		abs := filepath.Join(projectDir, rel)
		removeErr := os.RemoveAll(abs)
		if removeErr != nil && !os.IsNotExist(removeErr) {
			return removeErr
		}
	}

	return nil
}
