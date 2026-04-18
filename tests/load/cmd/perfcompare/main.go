package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	// defaultThreshold is set at 35% because realistic k6 load runs against
	// local Postgres+Redis show ~17-23% p95 variance and up to 70% p99
	// variance between back-to-back identical runs (GC pauses, singleflight
	// leader/waiter timing, connection pool warm-up). A tighter default
	// produces false positives. The gate is calibrated as an "egregious
	// regression detector" (+50% p95, +100% p99) — users tune via
	// --threshold or PERF_REGRESSION_THRESHOLD when they run longer
	// scenarios or have steadier infra.
	defaultThreshold = 0.35
	envThresholdKey  = "PERF_REGRESSION_THRESHOLD"
)

func main() {
	baselinePath := flag.String("baseline", "", "path to committed k6 baseline summary JSON")
	summaryPath := flag.String("summary", "", "path to current k6 summary export JSON")
	thresholdFlag := flag.Float64("threshold", -1, "max tolerated p95 regression as fraction (e.g. 0.15 = 15%)")
	flag.Parse()

	if *baselinePath == "" || *summaryPath == "" {
		fmt.Fprintln(os.Stderr, "usage: perfcompare --baseline <path> --summary <path> [--threshold 0.15]")
		os.Exit(2)
	}

	threshold := resolveThreshold(*thresholdFlag, os.Getenv(envThresholdKey))

	baseline, baseErr := Load(*baselinePath)
	if baseErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", baseErr)
		os.Exit(2)
	}
	current, curErr := Load(*summaryPath)
	if curErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", curErr)
		os.Exit(2)
	}

	report := Compare(baseline, current, threshold)
	printReport(os.Stdout, report)

	if !report.Passed {
		os.Exit(1)
	}
}

// resolveThreshold picks the first explicit source (flag > env > default).
// flagValue == -1 means "unset" (users pass a real fraction).
func resolveThreshold(flagValue float64, envValue string) float64 {
	if flagValue >= 0 {
		return flagValue
	}
	if envValue != "" {
		if parsed, parseErr := strconv.ParseFloat(envValue, 64); parseErr == nil && parsed >= 0 {
			return parsed
		}
	}
	return defaultThreshold
}

// printf/writeln wrap fmt.Fprintf/Fprintln discarding errors — writes to
// stdout from a CLI cannot meaningfully fail, and errcheck wants the
// ignore to be explicit.
func printf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeln(w io.Writer, line string) {
	_, _ = fmt.Fprintln(w, line)
}

func printReport(w io.Writer, r Report) {
	printf(w, "perfcompare: threshold=%.2f%% (p99 gate=%.2f%%)\n",
		r.ThresholdUsed*100, r.ThresholdUsed*200)

	if len(r.Regressions) > 0 {
		writeln(w, "\n[FAIL] regressions:")
		for _, d := range r.Regressions {
			printf(w, "  %s %s: %.2fms -> %.2fms (%+.2f%%)\n",
				d.Metric, d.Statistic, d.Baseline, d.Current, d.DeltaPct)
		}
	}

	if len(r.MissingMetrics) > 0 {
		writeln(w, "\n[FAIL] baseline metrics missing in current summary:")
		for _, name := range r.MissingMetrics {
			printf(w, "  %s\n", name)
		}
	}

	if len(r.Improvements) > 0 {
		writeln(w, "\n[OK] improvements:")
		for _, d := range r.Improvements {
			printf(w, "  %s %s: %.2fms -> %.2fms (%+.2f%%)\n",
				d.Metric, d.Statistic, d.Baseline, d.Current, d.DeltaPct)
		}
	}

	if len(r.NewMetrics) > 0 {
		writeln(w, "\n[INFO] new metrics (not in baseline):")
		for _, name := range r.NewMetrics {
			printf(w, "  %s\n", name)
		}
	}

	if r.Passed {
		writeln(w, "\nOK -- no regressions detected")
	} else {
		writeln(w, "\nFAIL -- see regressions above")
	}
}
