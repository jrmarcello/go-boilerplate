# Performance regression gate

The project has a **performance fitness function**: each scenario (smoke, load, stress, spike)
may have a committed baseline under `tests/load/baselines/<scenario>.json`. CI and the
`make load-regression` target re-run the scenario and fail if p95 or p99 latency regresses
beyond a configurable threshold.

This guide covers the day-to-day flow: running it, updating the baseline, and debugging
failures. For the full spec see [.specs/k6-regression-gate.md](../../.specs/k6-regression-gate.md).

## How the gate works

1. k6 runs the scenario with `--summary-export=<file>.json`.
2. `tests/load/cmd/perfcompare` loads the committed baseline and the fresh summary.
3. For every time-trend metric present in both, it compares p95 and p99:
   - **p95 threshold** = base threshold (default `0.35` = 35%).
   - **p99 threshold** = base threshold × 2 (tail latency is noisier by nature).
4. Exit code:
   - `0` — no regression (improvements and new metrics are informational).
   - `1` — regression: some metric exceeded its threshold, or a baseline metric is missing
     from the current run.
   - `2` — usage error or I/O failure (missing baseline file, malformed JSON).

The default is deliberately loose: it catches **egregious** regressions (+50% p95, +100% p99),
not fine-grained drift. This is a conscious template trade-off — see "Tuning the threshold"
below for when and how to tighten it.

## Local commands

```bash
# Regenerate the load baseline (overwrites tests/load/baselines/load.json).
# Run this deliberately, from a representative environment, after a release you want to pin.
# Takes ~3.5 min.
make load-baseline SCENARIO=load

# Run the scenario and fail on regression vs. committed baseline.
make load-regression SCENARIO=load
```

`SCENARIO` maps 1:1 to the k6 scenarios declared in [tests/load/main.js](../../tests/load/main.js):
`smoke`, `load`, `stress`, `spike`. Each gets its own baseline file — no baseline, no gate.

### Why not smoke

The `smoke` scenario runs with `vus: 1, iterations: 1`. With a single sample, every
percentile (p90, p95, p99) collapses to the same value — the one measurement — and two
back-to-back runs against identical code routinely diverge by > 30% due to normal jitter. Using
smoke for the gate produces false positives on unchanged code.

**Use smoke for functional/assertion validation** (`k6 check` passes/fails), and **use `load`
or `stress` for the regression gate**. The CI workflow is configured with `SCENARIO: load` for
this reason.

### summaryTrendStats

`tests/load/main.js` explicitly sets `summaryTrendStats: ['min','med','avg','max','p(90)','p(95)','p(99)']`.
k6's default stops at `p(95)` — without this option the `p(99)` gate would silently be a no-op
because the stat is never emitted. If you fork `main.js` or write a new script, keep this
option.

### Tuning the threshold

```bash
# CI default: 35% on p95, 70% on p99. Override per run:
PERF_REGRESSION_THRESHOLD=0.15 make load-regression
```

The 35% default is empirically calibrated against this template's `load` scenario — back-to-back
identical runs show 17–23% p95 variance and up to 70% p99 variance (GC pauses, singleflight
leader/waiter timing, connection pool warm-up). A tighter threshold (10–15%) is appropriate when
you run in a **stable environment** with more samples (longer scenarios, dedicated hardware,
steady infra) and deliberately tuned a hot path. If your env is noisier (shared CI runners), run
the scenario longer or generate the baseline as the median of 3 runs to get stable percentiles.

## Updating the baseline

A PR that legitimately slows p95 (new feature, additional work per request) will fail the gate.
The intended workflow:

1. Run `make load-baseline SCENARIO=load` locally, against a dev environment matching prod
   shape (not a laptop running 30 Chrome tabs).
2. `git diff tests/load/baselines/load.json` — inspect what changed. If the new numbers look
   reasonable, commit.
3. The PR now contains both the code change AND the new baseline. Reviewers see both diffs and
   approve together.

No "two-step approval" process is needed — the baseline diff is the control. If the new
baseline is 3× worse than the old one, reviewers push back.

## CI wiring

[.github/workflows/perf-regression.yml](../../.github/workflows/perf-regression.yml) runs on:

- Every push to `main` — the canonical source of truth.
- PRs **explicitly labeled** `perf` — performance is not on the default PR critical path, but
  you can opt in per-change.

Runs require a committed baseline. If `tests/load/baselines/<scenario>.json` is missing, the
job emits a warning and skips the comparison (rather than failing) — this prevents the first
PR ever from being blocked on infra it doesn't yet have.

The workflow:

1. Boots Postgres + Redis via `make docker-up`.
2. Runs migrations.
3. Builds and starts the API.
4. Waits for `/health`.
5. Runs `make load-regression SCENARIO=load` (~3.5 min).
6. Uploads the raw summary as a workflow artifact (14-day retention) for debugging.

## Debugging a failure

When CI fails with a regression:

1. Download the `perf-summary-*` artifact from the workflow run.
2. Compare locally: `go run ./tests/load/cmd/perfcompare --baseline tests/load/baselines/load.json --summary <downloaded>.json`
3. The report lists each regressed metric with baseline/current ms and the delta %.
4. Decide: **fix the regression** (the common case) or **accept and rebaseline** (intentional
   trade-off).

## References

- Fowler, ["Harness Engineering for Coding Agents"](https://martinfowler.com/articles/harness-engineering.html)
  — performance tests are called out as one of the canonical fitness-function sensors.
- [docs/harness.md](../harness.md) — this gate is listed as a sensor in the inventory.
- [.specs/k6-regression-gate.md](../../.specs/k6-regression-gate.md) — full spec with test plan.
