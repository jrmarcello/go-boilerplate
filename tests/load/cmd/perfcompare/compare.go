// Package main implements perfcompare, a CLI that compares a k6 summary JSON
// against a committed baseline and exits non-zero on p95/p99 regressions.
//
// See .specs/k6-regression-gate.md for the contract this implements.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// K6Summary mirrors the shape of `k6 run --summary-export=...` output.
// Only the metrics map is modeled; root_group, state, etc. are ignored.
type K6Summary struct {
	Metrics map[string]Metric `json:"metrics"`
}

// Metric holds the numeric stats of one k6 metric. k6's export shape varies by
// metric type (trend has p(95)/avg/min/max, counter has count/rate, check has
// passes/fails/value, etc.), so we flatten every numeric field into Values and
// let callers probe by key.
type Metric struct {
	Values map[string]float64
}

// UnmarshalJSON flattens every numeric field of a k6 metric into Values.
// Nested objects (e.g., `thresholds`) and non-numeric values are dropped.
func (m *Metric) UnmarshalJSON(data []byte) error {
	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Values = make(map[string]float64, len(raw))
	for key, value := range raw {
		if num, ok := value.(float64); ok {
			m.Values[key] = num
		}
	}
	return nil
}

// MarshalJSON mirrors UnmarshalJSON so round-tripping test fixtures works.
func (m Metric) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Values)
}

// IsTimeTrend reports whether a metric carries a p(95) — the canonical marker
// of a time-trend metric in k6 summaries (counters, rates, and checks don't).
func (m Metric) IsTimeTrend() bool {
	_, ok := m.Values["p(95)"]
	return ok
}

// Report is the structured output of Compare. Passed is false when any
// gated statistic crossed its threshold or a baseline metric is missing.
type Report struct {
	Passed         bool
	Regressions    []Delta
	Improvements   []Delta
	NewMetrics     []string
	MissingMetrics []string
	ThresholdUsed  float64
	Statistics     []string
}

// Delta describes a per-statistic movement of one metric between baseline and
// current. Used for both regressions (Passed=false) and improvements.
type Delta struct {
	Metric    string
	Statistic string
	Baseline  float64
	Current   float64
	DeltaPct  float64 // signed: negative = faster, positive = slower
}

// gatedStatistics is the fixed set of k6 trend stats that act as a gate.
// Baselines/currents that omit these keys are skipped silently.
var gatedStatistics = []string{"p(95)", "p(99)"}

// statThresholdMultiplier scales the base threshold per statistic. p99 is
// allowed to move twice as much as p95 before being called a regression,
// because tail latency is inherently noisier.
var statThresholdMultiplier = map[string]float64{
	"p(95)": 1.0,
	"p(99)": 2.0,
}

// Compare evaluates current against baseline and returns a Report.
// threshold is the base fraction (e.g. 0.15 = 15% on p95, 30% on p99).
// A metric is gated only when both baseline and current expose it as a
// time trend (p(95) present); other metric types are ignored.
func Compare(baseline, current K6Summary, threshold float64) Report {
	report := Report{
		Passed:        true,
		ThresholdUsed: threshold,
		Statistics:    append([]string(nil), gatedStatistics...),
	}

	for _, name := range sortedKeys(baseline.Metrics) {
		baseMetric := baseline.Metrics[name]
		if !baseMetric.IsTimeTrend() {
			continue
		}

		curMetric, present := current.Metrics[name]
		if !present {
			report.MissingMetrics = append(report.MissingMetrics, name)
			report.Passed = false
			continue
		}
		if !curMetric.IsTimeTrend() {
			continue
		}

		for _, stat := range gatedStatistics {
			baseVal, baseOK := baseMetric.Values[stat]
			curVal, curOK := curMetric.Values[stat]
			if !baseOK || !curOK {
				continue
			}
			if baseVal == 0 {
				continue
			}

			delta := (curVal - baseVal) / baseVal
			statThreshold := threshold * statThresholdMultiplier[stat]
			d := Delta{
				Metric:    name,
				Statistic: stat,
				Baseline:  baseVal,
				Current:   curVal,
				DeltaPct:  delta * 100,
			}

			switch {
			case delta > statThreshold:
				report.Regressions = append(report.Regressions, d)
				report.Passed = false
			case delta < -statThreshold:
				report.Improvements = append(report.Improvements, d)
			}
		}
	}

	for _, name := range sortedKeys(current.Metrics) {
		m := current.Metrics[name]
		if !m.IsTimeTrend() {
			continue
		}
		if _, ok := baseline.Metrics[name]; !ok {
			report.NewMetrics = append(report.NewMetrics, name)
		}
	}

	return report
}

func sortedKeys(m map[string]Metric) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Load reads a k6 summary JSON from disk and decodes it. It returns a clear
// error on missing file or malformed JSON — the caller renders the message
// to the user.
func Load(path string) (K6Summary, error) {
	// Path is intentionally user-supplied (CLI flag): this is the entire
	// purpose of the tool, so G304 does not apply.
	data, readErr := os.ReadFile(path) //nolint:gosec // G304: file path is the CLI contract
	if readErr != nil {
		return K6Summary{}, fmt.Errorf("reading %s: %w", path, readErr)
	}
	var s K6Summary
	if unmarshalErr := json.Unmarshal(data, &s); unmarshalErr != nil {
		return K6Summary{}, fmt.Errorf("parsing %s: %w", path, unmarshalErr)
	}
	return s, nil
}
