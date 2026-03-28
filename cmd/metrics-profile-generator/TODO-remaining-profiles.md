# Metrics Profile Generator — Remaining Work

## Status: 4 of 8 profiles complete and tested

### Completed profiles (tests passing)
- `build-farm-metrics.yml` — 3 metrics (profile_build_farm.go)
- `metrics-report-stackrox.yml` — 4 metrics (profile_metrics_report_stackrox.go)
- `kueue-metrics.yml` — 10 metrics (profile_kueue.go)
- `metrics.yml` — 56 metrics (profile_metrics.go)

### Remaining profiles to implement with builder API

1. **metrics-aggregated.yml** — 47 metrics (profile_metrics_aggregated.go)
   - File exists but uses raw strings; needs rewrite with builder API
   - Golden file: testdata/metrics-aggregated.yml

2. **metrics-egressip.yml** — small profile
   - Needs new profile_metrics_egressip.go using builder API
   - Golden file: testdata/metrics-egressip.yml
   - Enums already exist: MetricEgressIPStartupLatency, MetricEgressIPRecoveryLatency, MetricStartupNonEIPTotal

3. **metrics-report.yml** — 99 metrics (largest profile)
   - Needs new profile_metrics_report.go using builder API
   - Golden file: testdata/metrics-report.yml

4. **cnv-metrics.yml** — 30 metrics
   - Needs new profile_cnv.go using builder API
   - Golden file: testdata/cnv-metrics.yml
   - Enums already exist for all KubeVirt/CNV metrics

### For each remaining profile
1. Write the `Build<Name>Profile(g *Generator)` function using the builder API
2. Add a `TestXxxProfile` entry in profiles_test.go
3. Run `make unit` to verify

### Builder notes
- Use `QRaw(metric)` for bare metric names (no `{}`) in golden files
- Use `Q(metric, "")` when golden files have `metric{}`
- Use `Q(metric, filters)` when there are label selectors
- Filter ordering matters — match the golden file's filter order exactly
- `AggBy`/`AggBySpaced`/`SpacedBy` are legacy whitespace variants (TODO: remove after all profiles migrated)
