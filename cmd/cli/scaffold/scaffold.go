package scaffold

import "fmt"

// Scaffold is the main scaffold engine.
type Scaffold struct {
	cfg Config
}

// New creates a new Scaffold engine with the given config.
func New(cfg Config) *Scaffold {
	return &Scaffold{cfg: cfg}
}

// Validate checks that the config is valid before executing.
func (s *Scaffold) Validate() error {
	if s.cfg.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if s.cfg.ModulePath == "" {
		return fmt.Errorf("module path is required")
	}
	if s.cfg.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if s.cfg.Idempotency && !s.cfg.Redis {
		return fmt.Errorf("idempotency requires Redis to be enabled")
	}
	if s.cfg.Protocol != ProtocolHTTP {
		return fmt.Errorf("protocol %q is not yet supported (only 'http' is available)", s.cfg.Protocol)
	}
	if s.cfg.DI != DIManual {
		return fmt.Errorf("DI strategy %q is not yet supported (only 'manual' is available)", s.cfg.DI)
	}
	return nil
}
