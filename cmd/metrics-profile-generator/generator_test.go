package main

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func parseYAML(t *testing.T, data []byte) []metricDefinition {
	t.Helper()
	var metrics []metricDefinition
	if err := yaml.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}
	return metrics
}

func TestGeneratorAdd(t *testing.T) {
	g := &Generator{}
	g.Add("testMetric", "up{}")
	g.AddInstant("testInstant", "up{}", true)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	if metrics[0].MetricName != "testMetric" || metrics[0].Query != "up{}" {
		t.Errorf("unexpected metric[0]: %+v", metrics[0])
	}
	if metrics[0].Instant || metrics[0].CaptureStart {
		t.Errorf("metric[0] should not be instant")
	}

	if metrics[1].MetricName != "testInstant" || !metrics[1].Instant || !metrics[1].CaptureStart {
		t.Errorf("unexpected metric[1]: %+v", metrics[1])
	}
}

func TestGeneratorAddQuery(t *testing.T) {
	g := &Generator{}
	g.AddQuery("test", Q(MetricNodeCPU, "").IRate(Rate2m).Agg(AggSum, GroupByInstance))
	g.AddQueryInstant("testInstant", Q(MetricEtcdDBSize, "").Agg(AggAvg), false)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Query != "sum(irate(node_cpu_seconds_total{}[2m])) by (instance)" {
		t.Errorf("unexpected query: %s", metrics[0].Query)
	}
	if !metrics[1].Instant {
		t.Error("expected instant query")
	}
}

func TestHistogramQuantiles(t *testing.T) {
	g := &Generator{}
	g.HistogramQuantiles("etcdCommit", MetricEtcdDiskCommitDuration,
		[]Percentile{P50, P99}, "", []GroupBy{GroupByInstance})

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	if metrics[0].MetricName != "etcdCommitP50" {
		t.Errorf("expected name etcdCommitP50, got %s", metrics[0].MetricName)
	}
	if !strings.Contains(metrics[0].Query, "0.50") {
		t.Errorf("expected P50 quantile in query: %s", metrics[0].Query)
	}
	if !strings.Contains(metrics[0].Query, "instance,le") {
		t.Errorf("expected instance,le in group by: %s", metrics[0].Query)
	}

	if metrics[1].MetricName != "etcdCommitP99" {
		t.Errorf("expected name etcdCommitP99, got %s", metrics[1].MetricName)
	}
	if !strings.Contains(metrics[1].Query, "0.99") {
		t.Errorf("expected P99 quantile in query: %s", metrics[1].Query)
	}
}

func TestHistogramQuantilesIRate(t *testing.T) {
	g := &Generator{}
	g.HistogramQuantilesIRate("apiLatency", MetricAPIServerRequestDuration,
		[]Percentile{P99}, Rate2m,
		`apiserver="kube-apiserver"`, []GroupBy{GroupByVerb})

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if !strings.Contains(metrics[0].Query, "irate(") {
		t.Errorf("expected irate in query: %s", metrics[0].Query)
	}
}

func TestRateForMetrics(t *testing.T) {
	g := &Generator{}
	g.RateForMetrics(
		[]Metric{MetricNodeNetworkRx, MetricNodeNetworkTx},
		Rate2m, "", []GroupBy{GroupByInstance},
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	// Verify names are auto-generated with Rate suffix
	if !strings.HasSuffix(metrics[0].MetricName, "Rate") {
		t.Errorf("expected Rate suffix, got %s", metrics[0].MetricName)
	}
	if !strings.HasSuffix(metrics[1].MetricName, "Rate") {
		t.Errorf("expected Rate suffix, got %s", metrics[1].MetricName)
	}
}

func TestIRateForMetrics(t *testing.T) {
	g := &Generator{}
	g.IRateForMetrics(
		[]Metric{MetricNodeCPU},
		Rate2m, "", []GroupBy{GroupByMode, GroupByInstance},
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if !strings.Contains(metrics[0].Query, "irate(") {
		t.Errorf("expected irate in query: %s", metrics[0].Query)
	}
	if !strings.HasSuffix(metrics[0].MetricName, "IRate") {
		t.Errorf("expected IRate suffix, got %s", metrics[0].MetricName)
	}
}

func TestAggForMetrics(t *testing.T) {
	g := &Generator{}
	g.AggForMetrics(AggAvg,
		[]Metric{MetricNodeMemoryAvailable, MetricNodeMemoryActive},
		"", []GroupBy{GroupByInstance},
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	for _, m := range metrics {
		if !strings.HasPrefix(m.Query, "avg(") {
			t.Errorf("expected avg() query: %s", m.Query)
		}
	}
}

func TestForNodeRoles(t *testing.T) {
	g := &Generator{}
	g.ForNodeRoles("containerCPU", []NodeRole{RoleMaster, RoleWorker, RoleInfra},
		func(role NodeRole) *Query {
			return Q(MetricContainerCPU, "").Agg(AggSum, GroupByNode).
				AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(role))
		},
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(metrics))
	}

	expectedNames := []string{"containerCPUMaster", "containerCPUWorker", "containerCPUInfra"}
	expectedRoles := []string{"master", "worker", "infra"}
	for i, m := range metrics {
		if m.MetricName != expectedNames[i] {
			t.Errorf("metric %d: expected name %s, got %s", i, expectedNames[i], m.MetricName)
		}
		if !strings.Contains(m.Query, `role="`+expectedRoles[i]+`"`) {
			t.Errorf("metric %d: expected role %s in query: %s", i, expectedRoles[i], m.Query)
		}
	}
}

func TestForNodeRolesInstant(t *testing.T) {
	g := &Generator{}
	g.ForNodeRolesInstant("cpuSeconds", []NodeRole{RoleMaster, RoleWorker}, true,
		func(role NodeRole) *Query {
			return Q(MetricProcessCPU, "").Agg(AggSum, GroupByInstance).
				AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(role))
		},
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	for _, m := range metrics {
		if !m.Instant {
			t.Errorf("expected instant: %+v", m)
		}
		if !m.CaptureStart {
			t.Errorf("expected captureStart: %+v", m)
		}
	}
}

func TestCustomTemplate(t *testing.T) {
	g := &Generator{}
	g.CustomTemplate(
		`sum(rate({{.metric}}_total{job="{{.job}}"}[2m]))`,
		`{{.name}}Rate`,
		[]map[string]string{
			{"metric": "process_cpu_seconds", "job": "kubelet", "name": "kubeletCPU"},
			{"metric": "process_cpu_seconds", "job": "crio", "name": "crioCPU"},
		},
		false,
	)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	if metrics[0].MetricName != "kubeletCPURate" {
		t.Errorf("expected kubeletCPURate, got %s", metrics[0].MetricName)
	}
	if !strings.Contains(metrics[0].Query, `job="kubelet"`) {
		t.Errorf("expected kubelet job in query: %s", metrics[0].Query)
	}

	if metrics[1].MetricName != "crioCPURate" {
		t.Errorf("expected crioCPURate, got %s", metrics[1].MetricName)
	}
}

func TestCustomTemplateInstant(t *testing.T) {
	g := &Generator{}
	g.CustomTemplate(`up{}`, `testUp`, []map[string]string{{"_": ""}}, true)

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	metrics := parseYAML(t, out)
	if len(metrics) != 1 || !metrics[0].Instant {
		t.Errorf("expected 1 instant metric, got %+v", metrics)
	}
}

func TestCount(t *testing.T) {
	g := &Generator{}
	if g.Count() != 0 {
		t.Error("expected 0")
	}
	g.Add("a", "up{}")
	g.Add("b", "up{}")
	if g.Count() != 2 {
		t.Errorf("expected 2, got %d", g.Count())
	}
}

func TestGenerateValidYAML(t *testing.T) {
	g := &Generator{}
	g.Add("test1", "up{}")
	g.AddInstant("test2", "up{}", true)
	g.AddQuery("test3", Q(MetricNodeCPU, "").IRate(Rate2m))

	out, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's valid YAML that roundtrips
	var metrics []metricDefinition
	if err := yaml.Unmarshal(out, &metrics); err != nil {
		t.Fatalf("generated invalid YAML: %v", err)
	}
	if len(metrics) != 3 {
		t.Fatalf("expected 3 metrics after roundtrip, got %d", len(metrics))
	}

	// Verify instant/captureStart are omitted when false
	yamlStr := string(out)
	if strings.Contains(yamlStr, "instant: false") {
		t.Error("YAML should omit instant when false (omitempty)")
	}
	if strings.Contains(yamlStr, "captureStart: false") {
		t.Error("YAML should omit captureStart when false (omitempty)")
	}
}

func TestMetricToName(t *testing.T) {
	tests := []struct {
		metric   Metric
		expected string
	}{
		{MetricNodeNetworkRx, "nodeNetworkReceive"},
		{MetricContainerCPU, "containerCpuUsage"},
		{MetricEtcdDBSize, "etcdMvccDbTotalSizeIn"},
		{MetricProcessCPU, "processCpu"},
	}
	for _, tt := range tests {
		t.Run(string(tt.metric), func(t *testing.T) {
			got := metricToName(tt.metric)
			if got != tt.expected {
				t.Errorf("metricToName(%q) = %q, want %q", tt.metric, got, tt.expected)
			}
		})
	}
}
