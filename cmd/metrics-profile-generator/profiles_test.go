package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// normalizeQuery strips all cosmetic whitespace so semantically equivalent
// PromQL expressions compare equal. PromQL is whitespace-insensitive.
func normalizeQuery(q string) string {
	// Remove all spaces/tabs
	q = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, q)
	q = strings.TrimSpace(q)
	return q
}

// loadGoldenMetrics reads a golden YAML file and returns parsed metrics.
func loadGoldenMetrics(t *testing.T, path string) []metricDefinition {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var metrics []metricDefinition
	if err := yaml.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("failed to parse golden file %s: %v", path, err)
	}
	return metrics
}

// compareProfiles compares a generated profile against a golden YAML file.
func compareProfiles(t *testing.T, goldenPath string, buildFn func(g *Generator)) {
	t.Helper()

	golden := loadGoldenMetrics(t, goldenPath)

	g := &Generator{}
	buildFn(g)
	generated := g.metrics

	// Write generated YAML to temp file for inspection
	testName := filepath.Base(goldenPath)
	testName = strings.TrimSuffix(testName, filepath.Ext(testName))
	if generatedYAML, err := yaml.Marshal(generated); err == nil {
		generatedPath := filepath.Join(os.TempDir(), "generated-"+testName+".yml")
		if err := os.WriteFile(generatedPath, generatedYAML, 0644); err == nil {
			t.Logf("Generated YAML written to: %s", generatedPath)
		}
	}

	if len(generated) != len(golden) {
		t.Errorf("metric count mismatch: generated %d, golden %d", len(generated), len(golden))
		// Print names for debugging
		t.Log("Generated names:")
		for i, m := range generated {
			t.Logf("  [%d] %s", i, m.MetricName)
		}
		t.Log("Golden names:")
		for i, m := range golden {
			t.Logf("  [%d] %s", i, m.MetricName)
		}
		// Continue to show all differences
	}

	// Build map of golden metrics by name for lookup
	goldenByName := make(map[string]metricDefinition)
	for _, m := range golden {
		goldenByName[m.MetricName] = m
	}

	// Check each generated metric exists in golden and matches
	for i, gen := range generated {
		gol, ok := goldenByName[gen.MetricName]
		if !ok {
			t.Errorf("[%d] generated metric %q not found in golden file", i, gen.MetricName)
			continue
		}

		if gen.Instant != gol.Instant {
			t.Errorf("[%s] instant mismatch: generated=%v, golden=%v", gen.MetricName, gen.Instant, gol.Instant)
		}
		if gen.CaptureStart != gol.CaptureStart {
			t.Errorf("[%s] captureStart mismatch: generated=%v, golden=%v", gen.MetricName, gen.CaptureStart, gol.CaptureStart)
		}

		genQ := normalizeQuery(gen.Query)
		golQ := normalizeQuery(gol.Query)
		if genQ != golQ {
			t.Errorf("[%s] query mismatch:\n  generated: %s\n  golden:    %s", gen.MetricName, genQ, golQ)
		}
	}

	// Check for golden metrics missing from generated
	generatedByName := make(map[string]bool)
	for _, m := range generated {
		generatedByName[m.MetricName] = true
	}
	for _, gol := range golden {
		if !generatedByName[gol.MetricName] {
			t.Errorf("golden metric %q not in generated output", gol.MetricName)
		}
	}
}

func TestBuildFarmProfile(t *testing.T) {
	compareProfiles(t, "testdata/build-farm-metrics.yml", BuildBuildFarmProfile)
}

func TestMetricsReportStackroxProfile(t *testing.T) {
	compareProfiles(t, "testdata/metrics-report-stackrox.yml", BuildMetricsReportStackroxProfile)
}

func TestKueueProfile(t *testing.T) {
	compareProfiles(t, "testdata/kueue-metrics.yml", BuildKueueProfile)
}

func TestMetricsProfile(t *testing.T) {
	compareProfiles(t, "testdata/metrics.yml", BuildMetricsProfile)
}
