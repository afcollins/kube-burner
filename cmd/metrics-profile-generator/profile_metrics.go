package main

// BuildMetricsProfile generates the metrics.yml profile (56 metrics).
func BuildMetricsProfile(g *Generator) {
	ocpNS := `name!="",container!~"POD|",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"`
	netDevice := `device!~"lo|ovs-system"`

	// API server
	addAPILatencyMetrics(g)

	// Containers & pod metrics (no per-role split)
	g.AddQuery("containerCPU",
		Q(MetricContainerCPU, ocpNS).
			IRate(Rate2m).Multiply("100").
			Agg(AggSum, GroupByContainer, GroupByPod, GroupByNamespace, GroupByNode).
			Paren().Gt("0"))

	g.AddQuery("containerMemory",
		Q(MetricContainerMemoryRSS, ocpNS).
			Agg(AggSum, GroupByContainer, GroupByPod, GroupByNamespace, GroupByNode))

	// Kubelet & CRI-O
	g.AddQuery("kubeletCPU",
		Q(MetricProcessCPU, FilterKubeletCPU).
			IRate(Rate2m).Multiply("100").
			Agg(AggSum, GroupByNode).
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)))

	g.AddQuery("kubeletMemory",
		Q(MetricProcessMemory, FilterKubeletCPU).
			Agg(AggSum, GroupByNode).
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)))

	g.AddQuery("crioCPU",
		Q(MetricProcessCPU, FilterCrioCPU).
			IRate(Rate2m).Multiply("100").
			Agg(AggSum, GroupByNode).
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)))

	g.AddQuery("crioMemory",
		Q(MetricProcessMemory, FilterCrioCPU).
			Agg(AggSum, GroupByNode).
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)))

	g.AddQuery("containerNetworkSetupLatency",
		Q(MetricContainerCRIOLatency, `operation_type="network_setup_pod"`).
			IRate(Rate2m).Gt("0"))

	// Node network
	g.AddQuery("nodeRxNetwork",
		Q(MetricNodeNetworkRx, netDevice).IRate(Rate2m).Gt("0").
			AggBySpaced(AggSum, GroupByInstance, GroupByDevice))

	g.AddQuery("nodeTxNetwork",
		Q(MetricNodeNetworkTx, netDevice).IRate(Rate2m).Gt("0").
			AggBySpaced(AggSum, GroupByInstance, GroupByDevice))

	g.AddQuery("nodeNetworkErrRXTotal",
		Q(MetricNodeNetworkRxErrs, netDevice).IRate(Rate2m).Gt("0").
			AggBySpaced(AggSum, GroupByInstance, GroupByDevice))

	g.AddQuery("nodeNetworkErrTXTotal",
		Q(MetricNodeNetworkTxErrs, netDevice).IRate(Rate2m).Gt("0").
			AggBySpaced(AggSum, GroupByInstance, GroupByDevice))

	g.AddQuery("nodeNetworkDeviceDropTXTotal",
		Q(MetricNodeNetworkTxDrop, netDevice).IRate(Rate2m).Gt("0").
			AggBy(AggSum, GroupByInstance, GroupByDevice))

	g.AddQuery("nodeNetworkDeviceDropRXTotal",
		Q(MetricNodeNetworkRxDrop, netDevice).IRate(Rate2m).Gt("0").
			AggBySpaced(AggSum, GroupByInstance, GroupByDevice))

	// Node CPU per role
	addNodeCPUPerRole(g)

	// Node memory utilization per role
	addNodeMemoryUtilPerRole(g)

	// Etcd
	addEtcdMetrics(g)
	g.AddQuery("99thEtcdCompaction",
		QRaw(MetricEtcdCompactionDuration).Delta("1m", "30s").DivConst("2").Gt("0"))
	g.AddQuery("99thEtcdDefrag",
		QRaw(MetricEtcdDefragDuration).Delta("1m", "30s").DivConst("2").Gt("0"))

	// Scheduler
	g.AddQuery("schedulerThroughput",
		Q(MetricSchedulerAttempts, `result="scheduled"`).Rate(Rate2m).Agg(AggSum))
	g.AddQuery("99thSchedulerE2ELatency",
		QRaw(MetricSchedulerDuration).BucketRate(Rate2m).
			Agg(AggSum, GroupByLE).HistogramQuantile(P99))

	// Cluster
	addClusterMetrics(g)
	g.AddQueryInstant("ovsBuildInfo",
		QRaw(MetricOVSBuildInfo).TopK(1), false)

	// Prometheus
	addPrometheusMetrics(g)

	// Retain raw CPU seconds
	addNodeCPUSecondsCapture(g)

	// Cgroup captures
	addCgroupCapture(g)

	// Major faults
	g.AddQuery("nodeMajorFaults",
		QRaw(MetricNodeVmstatPgmajfault).Rate(Rate1m))

	// Cgroup CPU irate
	g.AddQuery("cgroupCPU",
		Q(MetricContainerCPU, FilterCgroupIDs).
			IRate(Rate2m).
			Agg(AggSum, GroupByID, GroupByNode))

	// Cgroup RSS memory
	g.AddQuery("cgroupMemoryRSS",
		Q(MetricContainerMemoryRSS, FilterCgroupIDs).
			Agg(AggSum, GroupByID, GroupByNode))
}

// addNodeCPUPerRole adds node CPU per role metrics (used by metrics.yml).
func addNodeCPUPerRole(g *Generator) {
	roles := []struct {
		name string
		role NodeRole
	}{
		{"nodeCPU-Workers", RoleWorker},
		{"nodeCPU-Masters", RoleMaster},
		{"nodeCPU-Infra", RoleInfra},
	}
	for _, r := range roles {
		g.AddQuery(r.name,
			QRaw(MetricNodeCPU).
				IRate(Rate2m).
				Agg(AggSum, GroupByMode, GroupByInstance).
				AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(r.role)).
				Paren().Gt("0"))
	}
}

// addNodeMemoryUtilPerRole adds node memory utilization per role.
func addNodeMemoryUtilPerRole(g *Generator) {
	roles := []struct {
		name string
		role NodeRole
	}{
		{"nodeMemoryUtilization-Masters", RoleMaster},
		{"nodeMemoryUtilization-Workers", RoleWorker},
		{"nodeMemoryUtilization-Infra", RoleInfra},
	}
	for _, r := range roles {
		g.AddQuery(r.name,
			QRaw(MetricNodeMemoryTotal).
				Sub(QRaw(MetricNodeMemoryAvailable)).
				Paren().
				AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(r.role)))
	}
}

// Common process filter strings.
const (
	FilterKubeletCPU = `service="kubelet",job="kubelet"`
	FilterCrioCPU    = `service="kubelet",job="crio"`
)
