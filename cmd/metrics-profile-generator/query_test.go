package main

import (
	"testing"
)

func TestQ(t *testing.T) {
	tests := []struct {
		name     string
		query    *Query
		expected string
	}{
		{
			name:     "metric with no filters",
			query:    Q(MetricNodeCPU, ""),
			expected: `node_cpu_seconds_total{}`,
		},
		{
			name:     "metric with filters",
			query:    Q(MetricAPIServerRequestTotal, `apiserver="kube-apiserver",verb!="WATCH"`),
			expected: `apiserver_request_total{apiserver="kube-apiserver",verb!="WATCH"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRate(t *testing.T) {
	got := Q(MetricContainerCPU, `name!=""`).Rate(Rate2m).String()
	want := `rate(container_cpu_usage_seconds_total{name!=""}[2m])`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIRate(t *testing.T) {
	got := Q(MetricContainerCPU, `name!=""`).IRate(Rate5m).String()
	want := `irate(container_cpu_usage_seconds_total{name!=""}[5m])`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAgg(t *testing.T) {
	tests := []struct {
		name     string
		query    *Query
		expected string
	}{
		{
			name:     "sum with group by",
			query:    Q(MetricContainerMemoryRSS, "").Agg(AggSum, GroupByPod, GroupByNamespace),
			expected: `sum(container_memory_rss{}) by (pod,namespace)`,
		},
		{
			name:     "avg no group by",
			query:    Q(MetricEtcdDBSize, "").Agg(AggAvg),
			expected: `avg(etcd_mvcc_db_total_size_in_bytes{})`,
		},
		{
			name:     "count",
			query:    Q(MetricKubeSecretInfo, "").Agg(AggCount),
			expected: `count(kube_secret_info{})`,
		},
		{
			name:     "max with group by",
			query:    Q(MetricNodeMemoryAvailable, "").Agg(AggMax, GroupByInstance),
			expected: `max(node_memory_MemAvailable_bytes{}) by (instance)`,
		},
		{
			name:     "min",
			query:    Q(MetricNodeMemoryAvailable, "").Agg(AggMin, GroupByNode),
			expected: `min(node_memory_MemAvailable_bytes{}) by (node)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTopK(t *testing.T) {
	got := Q(MetricProcessCPU, `job="kubelet"`).IRate(Rate2m).TopK(3).String()
	want := `topk(3, irate(process_cpu_seconds_total{job="kubelet"}[2m]))`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHistogramQuantile(t *testing.T) {
	got := Q(MetricAPIServerRequestDuration, `apiserver="kube-apiserver"`).
		Rate(Rate2m).
		Agg(AggSum, GroupByVerb, GroupByLE).
		HistogramQuantile(P99).
		String()
	want := `histogram_quantile(0.99, sum(rate(apiserver_request_duration_seconds{apiserver="kube-apiserver"}[2m])) by (verb,le))`
	if got != want {
		t.Errorf("got:\n  %q\nwant:\n  %q", got, want)
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		query    *Query
		expected string
	}{
		{
			name:     "multiply",
			query:    Q(MetricProcessCPU, "").IRate(Rate2m).Multiply("100"),
			expected: `irate(process_cpu_seconds_total{}[2m]) * 100`,
		},
		{
			name:     "subtract",
			query:    Q(MetricNodeMemoryTotal, "").Sub(Q(MetricNodeMemoryAvailable, "")),
			expected: `node_memory_MemTotal_bytes{} - node_memory_MemAvailable_bytes{}`,
		},
		{
			name:     "divide",
			query:    Q(MetricKubevirtLauncherOverhead, "").Agg(AggSum).Div(Q(MetricKubevirtLauncherOverhead, "").Agg(AggCount)),
			expected: `sum(kubevirt_vmi_launcher_memory_overhead_bytes{}) / count(kubevirt_vmi_launcher_memory_overhead_bytes{})`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestComparisons(t *testing.T) {
	got := Q(MetricContainerMemoryRSS, "").Agg(AggSum, GroupByPod).Gt("0").String()
	want := `sum(container_memory_rss{}) by (pod) > 0`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = Q(MetricContainerMemoryRSS, "").Gte("100").String()
	want = `container_memory_rss{} >= 100`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBinaryJoins(t *testing.T) {
	tests := []struct {
		name     string
		query    *Query
		expected string
	}{
		{
			name: "and on with node role",
			query: Q(MetricContainerCPU, "").Agg(AggSum, GroupByNode).
				AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)),
			expected: `sum(container_cpu_usage_seconds_total{}) by (node) and on (node) kube_node_role{role="worker"}`,
		},
		{
			name: "or with vector zero",
			query: Q(Metric("up"), `pod=~"virt-api-.*"`).Agg(AggSum).
				Or(VectorZero()),
			expected: `sum(up{pod=~"virt-api-.*"}) or vector(0)`,
		},
		{
			name: "and (no label matching)",
			query: Q(MetricNodeCPU, "").IRate(Rate2m).
				And(Q(MetricKubeNodeRole, `role="master"`)),
			expected: `irate(node_cpu_seconds_total{}[2m]) and kube_node_role{role="master"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestOverTime(t *testing.T) {
	tests := []struct {
		name     string
		query    *Query
		expected string
	}{
		{
			name: "avg_over_time with elapsed",
			query: Q(MetricClusterMemoryUsageRatio, "").
				OverTime(TimeAggAvg),
			expected: `avg_over_time(cluster:memory_usage:ratio{}[{{.elapsed}}:])`,
		},
		{
			name: "max_over_time with step",
			query: Q(MetricEtcdCompactionDuration, "").
				OverTimeStep(TimeAggMax, "30s"),
			expected: `max_over_time(etcd_debugging_mvcc_db_compaction_total_duration_milliseconds_sum{}[{{.elapsed}}:30s])`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.String(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDelta(t *testing.T) {
	got := Q(MetricEtcdDefragDuration, "").Delta("1m", "30s").String()
	want := `delta(etcd_disk_backend_defrag_duration_seconds_sum{}[1m:30s])`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabelReplace(t *testing.T) {
	got := NodeRoleLabelReplace(RoleMaster).String()
	want := `label_replace(kube_node_role{role="master"}, "instance", "$1", "node", "(.+)")`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParen(t *testing.T) {
	got := Q(MetricNodeMemoryTotal, "").Sub(Q(MetricNodeMemoryAvailable, "")).Paren().String()
	want := `(node_memory_MemTotal_bytes{} - node_memory_MemAvailable_bytes{})`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestComplexCompoundQuery(t *testing.T) {
	// Matches a real-world pattern from metrics.yml:
	// Container CPU per node role with irate, multiply, sum, paren, and on
	got := Q(MetricContainerCPU, `name!="",container!="POD",namespace=~"openshift-.*"`).
		IRate(Rate2m).Multiply("100").
		Agg(AggSum, GroupByContainer, GroupByPod, GroupByNamespace, GroupByNode).
		Paren().
		AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleMaster)).
		Gt("0").
		String()

	want := `(sum(irate(container_cpu_usage_seconds_total{name!="",container!="POD",namespace=~"openshift-.*"}[2m]) * 100) by (container,pod,namespace,node)) and on (node) kube_node_role{role="master"} > 0`
	if got != want {
		t.Errorf("got:\n  %s\nwant:\n  %s", got, want)
	}
}

func TestReportStyleQuery(t *testing.T) {
	// Pattern: histogram_quantile over time (from metrics-report.yml)
	got := Q(MetricEtcdDiskCommitDuration, "").
		Rate(Rate2m).
		Agg(AggSum, GroupByLE).
		HistogramQuantile(P99).
		OverTime(TimeAggMax).
		Gt("0").
		String()

	want := `max_over_time(histogram_quantile(0.99, sum(rate(etcd_disk_backend_commit_duration_seconds{}[2m])) by (le))[{{.elapsed}}:]) > 0`
	if got != want {
		t.Errorf("got:\n  %s\nwant:\n  %s", got, want)
	}
}
