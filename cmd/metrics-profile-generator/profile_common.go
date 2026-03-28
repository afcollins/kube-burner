package main

// addAPILatencyMetrics adds the standard API server latency and request rate metrics.
func addAPILatencyMetrics(g *Generator) {
	g.AddQuery("readOnlyAPICallsLatency",
		Q(MetricAPIServerRequestDuration, FilterAPIServer+`, verb=~"LIST|GET", `+FilterSubresource).
			BucketIRate(Rate2m).
			Agg(AggSum, GroupByLE, GroupByResource, GroupByVerb, GroupByScope).
			HistogramQuantile(P99).Gt("0"))

	g.AddQuery("mutatingAPICallsLatency",
		Q(MetricAPIServerRequestDuration, FilterAPIServer+`, verb=~"POST|PUT|DELETE|PATCH", `+FilterSubresource).
			BucketIRate(Rate2m).
			Agg(AggSum, GroupByLE, GroupByResource, GroupByVerb, GroupByScope).
			HistogramQuantile(P99).Gt("0"))

	g.AddQuery("APIRequestRate",
		Q(MetricAPIServerRequestTotal, `apiserver="kube-apiserver",verb!="WATCH"`).
			IRate(Rate2m).
			Agg(AggSum, GroupByVerb, GroupByResource, GroupByCode).Gt("0"))
}

// addEtcdMetrics adds standard etcd histogram and version metrics.
func addEtcdMetrics(g *Generator) {
	g.AddQuery("99thEtcdDiskBackendCommitDurationSeconds",
		QRaw(MetricEtcdDiskCommitDuration).BucketRate(Rate2m).HistogramQuantile(P99))
	g.AddQuery("99thEtcdDiskWalFsyncDurationSeconds",
		QRaw(MetricEtcdDiskWALSyncDuration).BucketRate(Rate2m).HistogramQuantile(P99))
	g.AddQuery("99thEtcdRoundTripTimeSeconds",
		QRaw(MetricEtcdNetworkPeerRoundTrip).BucketRate(Rate5m).HistogramQuantile(P99))

	g.AddQueryInstant("etcdVersion",
		QRaw(MetricEtcdClusterVersion).AggBy(AggSum, GroupByClusterVersion), false)
}

// addClusterMetrics adds standard cluster state metrics.
func addClusterMetrics(g *Generator) {
	g.AddQuery("namespaceCount",
		QRaw(MetricKubeNamespacePhase).Agg(AggSum, GroupByPhase).Gt("0"))
	g.AddQuery("podStatusCount",
		Q(MetricKubePodStatusPhase, "").Agg(AggSum, GroupByPhase))
	g.AddQueryInstant("secretCount", Q(MetricKubeSecretInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("deploymentCount", Q(MetricKubeDeploymentReplicas, "").Agg(AggCount), false)
	g.AddQueryInstant("configmapCount", Q(MetricKubeConfigmapInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("serviceCount", Q(MetricKubeServiceInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("routeCount", Q(MetricOCPRouteCreated, "").Agg(AggCount), false)
	g.AddQuery("nodeRoles", QRaw(MetricKubeNodeRole))
	g.AddQuery("nodeStatus",
		Q(MetricKubeNodeStatusCondition, `status="true"`).Agg(AggSum, GroupByCondition))
}

// addPrometheusMetrics adds TSDB head series and ingestion rate metrics.
func addPrometheusMetrics(g *Generator) {
	g.AddQuery("prometheus-timeseriestotal",
		Q(MetricOCPTSDBHeadSeries, `job="prometheus-k8s"`))
	g.AddQuery("prometheus-ingestionrate",
		Q(MetricOCPTSDBHeadSamples, `job="prometheus-k8s"`))
}

// addNodeCPUSecondsCapture adds raw CPU seconds totals per role (instant + captureStart).
func addNodeCPUSecondsCapture(g *Generator) {
	g.AddQueryInstant("nodeCPUSeconds-Workers",
		QRaw(MetricNodeCPU).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplaceExclude(RoleWorker, RoleInfra)).
			Agg(AggSum, GroupByMode),
		true)
	g.AddQueryInstant("nodeCPUSeconds-Masters",
		QRaw(MetricNodeCPU).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(RoleMaster)).
			Agg(AggSum, GroupByMode),
		true)
	g.AddQueryInstant("nodeCPUSeconds-Infra",
		QRaw(MetricNodeCPU).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(RoleInfra)).
			Agg(AggSum, GroupByMode),
		true)
}

// addCgroupCapture adds cgroup CPU/memory capture metrics per role (instant + captureStart).
func addCgroupCapture(g *Generator) {
	cgroupFilter := FilterCgroupIDs

	type roleSpec struct {
		suffix     string
		roleFilter string
	}
	roles := []roleSpec{
		{"Workers", `role="worker",role!="infra"`},
		{"Masters", `role="master"`},
		{"Infra", `role="infra"`},
	}

	for _, r := range roles {
		g.AddQueryInstant("cgroupCPUSeconds-"+r.suffix,
			Q(MetricContainerCPU, cgroupFilter).
				AndOn([]GroupBy{GroupByNode}, Q(MetricKubeNodeRole, r.roleFilter)).
				Agg(AggSum, GroupByID),
			true)
		g.AddQueryInstant("cgroupMemoryRSS-"+r.suffix,
			Q(MetricContainerMemoryRSS, cgroupFilter).
				AndOn([]GroupBy{GroupByNode}, Q(MetricKubeNodeRole, r.roleFilter)).
				Agg(AggSum, GroupByID),
			true)
	}

	g.AddQueryInstant("cgroupCPUSeconds-namespaces",
		Q(MetricContainerCPU, `container!~"POD|",namespace=~"openshift-.*"`).
			Agg(AggSum, GroupByNamespace),
		true)
	g.AddQueryInstant("cgroupMemoryRSS-namespaces",
		Q(MetricContainerMemoryRSS, `container!~"POD|",namespace=~"openshift-.*"`).
			Agg(AggSum, GroupByNamespace),
		true)
}

// Common filter strings shared across profiles.
const (
	FilterAPIServer   = `apiserver="kube-apiserver"`
	FilterSubresource = `subresource!~"log|exec|portforward|attach|proxy"`
	FilterCgroupIDs   = `id=~"/system.slice|/system.slice/kubelet.service|.*/ovs-vswitchd.service|/system.slice/crio.service|/kubepods.slice"`
)
