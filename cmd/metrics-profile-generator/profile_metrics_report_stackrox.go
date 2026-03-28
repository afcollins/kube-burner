package main

// BuildMetricsReportStackroxProfile generates the metrics-report-stackrox.yml profile.
func BuildMetricsReportStackroxProfile(g *Generator) {
	nsFilter := `name!="", namespace="stackrox", container!="POD"`

	// avg CPU
	g.AddQueryInstant("cpu-stackrox",
		Q(MetricContainerCPU, nsFilter).
			IRate(Rate2m).
			OverTime(TimeAggAvg).
			Agg(AggAvg, GroupByContainer),
		false)

	// max CPU
	g.AddQueryInstant("max-cpu-stackrox",
		Q(MetricContainerCPU, nsFilter).
			IRate(Rate2m).
			OverTime(TimeAggMax).
			Agg(AggMax, GroupByContainer),
		false)

	// avg memory
	g.AddQueryInstant("memory-stackrox",
		Q(MetricContainerMemoryRSS, nsFilter).
			OverTime(TimeAggAvg).
			Agg(AggAvg, GroupByContainer),
		false)

	// max memory
	g.AddQueryInstant("max-memory-stackrox",
		Q(MetricContainerMemoryRSS, nsFilter).
			OverTime(TimeAggAvg).
			Agg(AggMax, GroupByContainer),
		false)
}
