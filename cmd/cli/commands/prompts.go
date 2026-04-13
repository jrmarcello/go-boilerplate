package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptConfirm prompts user with y/N question. Default returned if empty input.
// Accepts "y/yes/s/sim" as true, "n/no/nao/não" as false. Unknown values return the default.
func PromptConfirm(reader *bufio.Reader, label string, defaultVal bool) (bool, error) {
	defaultHint := "Y/n"
	if !defaultVal {
		defaultHint = "y/N"
	}

	fmt.Printf("  %s [%s]: ", label, defaultHint)

	line, readErr := reader.ReadString('\n')
	if readErr != nil {
		return false, fmt.Errorf("reading input: %w", readErr)
	}

	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultVal, nil
	}

	switch line {
	case "y", "yes", "s", "sim":
		return true, nil
	case "n", "no", "não", "nao":
		return false, nil
	default:
		return defaultVal, nil
	}
}

// IsInteractive returns true if stdin is a terminal (TTY).
// When not interactive, prompts are skipped and defaults are used.
func IsInteractive() bool {
	fi, statErr := os.Stdin.Stat()
	if statErr != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
