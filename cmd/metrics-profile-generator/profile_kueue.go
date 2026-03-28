package main

// BuildKueueProfile generates the kueue-metrics.yml profile.
func BuildKueueProfile(g *Generator) {
	// Jobs
	g.AddQuery("totalJobs",
		Q(MetricKubeJobInfo, `namespace=~"kueue-scale-.+"`).Agg(AggCount))

	g.AddQuery("podStatusCount",
		Q(MetricKubePodStatusPhase, `namespace=~"kueue-scale-.+"`).
			Agg(AggSum, GroupByPhase))

	// Pod CPU avg per namespace (instant)
	type nsEntry struct {
		name      string
		namespace string
		isRegex   bool
	}
	cpuNamespaces := []nsEntry{
		{"cpu-kube-apiserver-avg", "openshift-kube-apiserver", false},
		{"cpu-kueue-avg", "openshift-kueue-operator|kueue-system", true},
		{"cpu-etcd-avg", "openshift-etcd", false},
	}
	for _, ns := range cpuNamespaces {
		nsFilter := `namespace="` + ns.namespace + `"`
		if ns.isRegex {
			nsFilter = `namespace=~"` + ns.namespace + `"`
		}
		g.AddQueryInstant(ns.name,
			Q(MetricContainerCPU, `name!="", `+nsFilter).
				IRate(Rate2m).Agg(AggSum).
				OverTime(TimeAggAvg),
			false)
	}

	// Max memory per namespace (instant)
	g.AddQueryInstant("max-memory-kube-apiserver-aggregated",
		Q(MetricContainerMemoryWS, `name!="", namespace="openshift-kube-apiserver"`).
			Agg(AggSum).OverTime(TimeAggMax).Agg(AggMax),
		false)

	// kueue controller memory (not instant in golden)
	g.AddQuery("max-memory-kueue-aggregated",
		Q(MetricContainerMemoryWS, `pod=~"kueue-controller-manager-.*", name="", namespace=~"openshift-kueue-operator|kueue-system"`).
			Agg(AggSum).OverTime(TimeAggMax).Agg(AggMax))

	g.AddQueryInstant("max-memory-etcd-aggregated",
		Q(MetricContainerMemoryRSS, `name!="", namespace="openshift-etcd"`).
			Agg(AggSum).OverTime(TimeAggMax).Agg(AggMax),
		false)

	// Kueue-specific
	g.AddQuery("P99KueueAdmissionWaitTime",
		QRaw(MetricKueueAdmissionWaitTime).BucketRate(Rate2m).HistogramQuantile(P99))

	g.AddQueryInstant("KueueBuildInfo",
		QRaw(MetricKueueBuildInfo), false)
}
