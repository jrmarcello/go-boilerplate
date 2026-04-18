package main

import (
	"os"
	"path/filepath"
	"testing"
)

// trendMetric is a test helper that builds a time-trend metric with common stats,
// matching k6's flat `--summary-export` shape.
func trendMetric(p50, p95, p99 float64) Metric {
	return Metric{
		Values: map[string]float64{
			"med":   p50,
			"p(95)": p95,
			"p(99)": p99,
			"avg":   p50,
			"min":   p50 * 0.5,
			"max":   p99 * 1.2,
		},
	}
}

// counterMetric is a test helper for non-time-trend metrics (no p(95)/p(99)).
func counterMetric(count, rate float64) Metric {
	return Metric{
		Values: map[string]float64{
			"count": count,
			"rate":  rate,
		},
	}
}

// summary wraps a metric map in K6Summary for terser tests.
func summary(metrics map[string]Metric) K6Summary {
	return K6Summary{Metrics: metrics}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		name                string
		baseline            K6Summary
		current             K6Summary
		threshold           float64
		wantPassed          bool
		wantRegressionsN    int
		wantImprovementsN   int
		wantNewMetricsN     int
		wantMissingMetricsN int
	}{
		{
			name:       "TC-UC-01 summary within threshold passes",
			baseline:   summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			current:    summary(map[string]Metric{"http_req_duration": trendMetric(11, 105, 160)}),
			threshold:  0.15,
			wantPassed: true,
		},
		{
			name:     "TC-UC-02 summary better than baseline is reported as improvement",
			baseline: summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			// p95: -20% (beyond 15% threshold) -> improvement.
			// p99: -40% (beyond 30% noise-adjusted threshold) -> improvement.
			current:           summary(map[string]Metric{"http_req_duration": trendMetric(8, 80, 90)}),
			threshold:         0.15,
			wantPassed:        true,
			wantImprovementsN: 2,
		},
		{
			name:       "TC-UC-03 p95 degrades exactly at threshold (inclusive upper bound)",
			baseline:   summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			current:    summary(map[string]Metric{"http_req_duration": trendMetric(10, 115, 150)}),
			threshold:  0.15,
			wantPassed: true,
		},
		{
			name:             "TC-UC-04 p95 degrades 15.01% -> fail",
			baseline:         summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			current:          summary(map[string]Metric{"http_req_duration": trendMetric(10, 115.01, 150)}),
			threshold:        0.15,
			wantPassed:       false,
			wantRegressionsN: 1,
		},
		{
			name:             "TC-UC-05 p99 degrades past its 2x-threshold while p95 is fine -> fail",
			baseline:         summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			current:          summary(map[string]Metric{"http_req_duration": trendMetric(10, 105, 250)}), // p95 +5%, p99 +66%
			threshold:        0.15,
			wantPassed:       false,
			wantRegressionsN: 1,
		},
		{
			name: "TC-UC-06 current introduces a new metric not in baseline -> pass with warning",
			baseline: summary(map[string]Metric{
				"http_req_duration": trendMetric(10, 100, 150),
			}),
			current: summary(map[string]Metric{
				"http_req_duration":     trendMetric(10, 105, 150),
				"new_endpoint_duration": trendMetric(5, 50, 80),
			}),
			threshold:       0.15,
			wantPassed:      true,
			wantNewMetricsN: 1,
		},
		{
			name: "TC-UC-07 baseline metric missing in current -> fail",
			baseline: summary(map[string]Metric{
				"http_req_duration":    trendMetric(10, 100, 150),
				"create_user_duration": trendMetric(20, 200, 300),
			}),
			current: summary(map[string]Metric{
				"http_req_duration": trendMetric(10, 100, 150),
			}),
			threshold:           0.15,
			wantPassed:          false,
			wantMissingMetricsN: 1,
		},
		{
			name:             "TC-UC-10 tighter custom threshold 5% catches 6% regression",
			baseline:         summary(map[string]Metric{"http_req_duration": trendMetric(10, 100, 150)}),
			current:          summary(map[string]Metric{"http_req_duration": trendMetric(10, 106, 150)}),
			threshold:        0.05,
			wantPassed:       false,
			wantRegressionsN: 1,
		},
		{
			name: "edge: non-time-trend metrics are ignored",
			baseline: summary(map[string]Metric{
				"http_reqs": counterMetric(100, 50),
			}),
			current: summary(map[string]Metric{
				"http_reqs": counterMetric(200, 100),
			}),
			threshold:  0.15,
			wantPassed: true,
		},
		{
			name: "edge: baseline with zero p95 is skipped (cannot compute delta)",
			baseline: summary(map[string]Metric{
				"http_req_duration": trendMetric(0, 0, 0),
			}),
			current: summary(map[string]Metric{
				"http_req_duration": trendMetric(10, 100, 150),
			}),
			threshold:  0.15,
			wantPassed: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare(tc.baseline, tc.current, tc.threshold)

			if got.Passed != tc.wantPassed {
				t.Errorf("Passed = %v, want %v\n  regressions: %+v\n  missing: %v\n  new: %v",
					got.Passed, tc.wantPassed, got.Regressions, got.MissingMetrics, got.NewMetrics)
			}
			if len(got.Regressions) != tc.wantRegressionsN {
				t.Errorf("Regressions count = %d, want %d: %+v",
					len(got.Regressions), tc.wantRegressionsN, got.Regressions)
			}
			if len(got.Improvements) != tc.wantImprovementsN {
				t.Errorf("Improvements count = %d, want %d: %+v",
					len(got.Improvements), tc.wantImprovementsN, got.Improvements)
			}
			if len(got.NewMetrics) != tc.wantNewMetricsN {
				t.Errorf("NewMetrics count = %d, want %d: %v",
					len(got.NewMetrics), tc.wantNewMetricsN, got.NewMetrics)
			}
			if len(got.MissingMetrics) != tc.wantMissingMetricsN {
				t.Errorf("MissingMetrics count = %d, want %d: %v",
					len(got.MissingMetrics), tc.wantMissingMetricsN, got.MissingMetrics)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	cases := []struct {
		name    string
		setupFn func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "TC-UC-08 malformed JSON returns clear error",
			setupFn: func(t *testing.T) string {
				p := filepath.Join(t.TempDir(), "bad.json")
				if writeErr := os.WriteFile(p, []byte("{not json"), 0o600); writeErr != nil {
					t.Fatal(writeErr)
				}
				return p
			},
			wantErr: true,
		},
		{
			name: "TC-UC-09 missing file returns clear error",
			setupFn: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist.json")
			},
			wantErr: true,
		},
		{
			name: "happy: valid minimal summary parses",
			setupFn: func(t *testing.T) string {
				p := filepath.Join(t.TempDir(), "ok.json")
				data := `{"metrics":{"http_req_duration":{"type":"trend","contains":"time","values":{"p(95)":100,"p(99)":150}}}}`
				if writeErr := os.WriteFile(p, []byte(data), 0o600); writeErr != nil {
					t.Fatal(writeErr)
				}
				return p
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setupFn(t)
			_, loadErr := Load(path)
			if (loadErr != nil) != tc.wantErr {
				t.Errorf("Load() err = %v, wantErr %v", loadErr, tc.wantErr)
			}
		})
	}
}

// TC-UC-11 is satisfied by Load being path-agnostic: callers can feed any
// baseline file (smoke.json, load.json, etc.). This test exercises the
// committed testdata fixtures to cover the end-to-end Load -> Compare flow
// without needing a running k6.
func TestCompareFromFixtures(t *testing.T) {
	baseline, err := Load(filepath.Join("testdata", "baseline_ok.json"))
	if err != nil {
		t.Fatalf("loading baseline fixture: %v", err)
	}
	regression, err := Load(filepath.Join("testdata", "summary_regression.json"))
	if err != nil {
		t.Fatalf("loading regression fixture: %v", err)
	}

	report := Compare(baseline, regression, 0.15)
	if report.Passed {
		t.Errorf("expected Passed=false for regression fixture, got true: %+v", report)
	}
	if len(report.Regressions) == 0 {
		t.Errorf("expected at least one regression, got none")
	}
}
