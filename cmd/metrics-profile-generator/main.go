package main

import (
	"fmt"
	"os"
)

// ---------------------------------------------------------------------------
// Define your metrics profile below.
// Replace this example with your actual metric definitions.
// ---------------------------------------------------------------------------
func main() {
	g := &Generator{}

	// --- API Server latency histograms ---
	g.HistogramQuantiles(
		"apiRequestLatency",
		MetricAPIServerRequestDuration,
		[]Percentile{P99},
		`apiserver="kube-apiserver",verb!~"WATCH",subresource!="log"`,
		[]GroupBy{GroupByResource, GroupByVerb, GroupByScope},
	)

	// --- API request rates (read vs write) ---
	g.AddQuery("readOnlyAPIRequestRate",
		Q(MetricAPIServerRequestTotal, `apiserver="kube-apiserver",verb=~"LIST|GET",subresource!~"log|exec|portforward|attach|proxy"`).
			IRate(Rate2m).
			Agg(AggSum, GroupByVerb, GroupByResource, GroupByCode).
			Gt("0"),
	)
	g.AddQuery("mutatingAPIRequestRate",
		Q(MetricAPIServerRequestTotal, `apiserver="kube-apiserver",verb=~"POST|PUT|PATCH|DELETE",subresource!~"log|exec|portforward|attach|proxy"`).
			IRate(Rate2m).
			Agg(AggSum, GroupByVerb, GroupByResource, GroupByCode).
			Gt("0"),
	)

	// --- Inflight requests ---
	g.AddQuery("apiInflightRequests",
		Q(MetricAPIServerInflightReqs, "").Agg(AggSum, GroupByRequestKind).Gt("0"),
	)

	// --- Container CPU/Memory per node role ---
	allRoles := []NodeRole{RoleMaster, RoleWorker, RoleInfra}

	g.ForNodeRoles("containerCPU", allRoles, func(role NodeRole) *Query {
		return Q(MetricContainerCPU, `name!="",container!="POD",namespace=~"openshift-.*"`).
			IRate(Rate2m).Multiply("100").
			Agg(AggSum, GroupByContainer, GroupByPod, GroupByNamespace, GroupByNode).
			Paren().
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(role)).
			Gt("0")
	})

	g.ForNodeRoles("containerMemory", allRoles, func(role NodeRole) *Query {
		return Q(MetricContainerMemoryRSS, `name!="",container!="POD",namespace=~"openshift-.*"`).
			Agg(AggSum, GroupByContainer, GroupByPod, GroupByNamespace, GroupByNode).
			AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(role)).
			Gt("0")
	})

	// --- Node CPU by mode per role ---
	g.ForNodeRoles("nodeCPU", allRoles, func(role NodeRole) *Query {
		return Q(MetricNodeCPU, "").
			IRate(Rate2m).
			Agg(AggSum, GroupByMode, GroupByInstance).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(role))
	})

	// --- Node memory utilization (total - available) per role ---
	g.ForNodeRoles("nodeMemoryUtilization", allRoles, func(role NodeRole) *Query {
		return Q(MetricNodeMemoryTotal, "").
			Sub(Q(MetricNodeMemoryAvailable, "")).
			Paren().
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(role))
	})

	// --- Node network per role ---
	g.ForNodeRoles("rxNetworkBytes", allRoles, func(role NodeRole) *Query {
		return Q(MetricNodeNetworkRx, `device=~"^(ens|eth|bond|team).*"`).
			IRate(Rate2m).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(role))
	})
	g.ForNodeRoles("txNetworkBytes", allRoles, func(role NodeRole) *Query {
		return Q(MetricNodeNetworkTx, `device=~"^(ens|eth|bond|team).*"`).
			IRate(Rate2m).
			AndOn([]GroupBy{GroupByInstance}, NodeRoleLabelReplace(role))
	})

	// --- etcd histograms ---
	g.HistogramQuantiles("etcdDiskCommit", MetricEtcdDiskCommitDuration,
		[]Percentile{P99}, "", []GroupBy{GroupByInstance})
	g.HistogramQuantiles("etcdDiskWALSync", MetricEtcdDiskWALSyncDuration,
		[]Percentile{P99}, "", []GroupBy{GroupByInstance})
	g.HistogramQuantiles("etcdPeerRoundTrip", MetricEtcdNetworkPeerRoundTrip,
		[]Percentile{P99}, "", []GroupBy{GroupByInstance})

	// --- etcd DB size ---
	g.AddQuery("etcdDBSize", Q(MetricEtcdDBSize, "").Agg(AggAvg))
	g.AddQuery("etcdDBSizeInUse", Q(MetricEtcdDBSizeInUse, "").Agg(AggAvg))

	// --- etcd cluster version (instant) ---
	g.AddQueryInstant("etcdVersion",
		Q(MetricEtcdClusterVersion, "").Agg(AggSum, GroupByClusterVersion), false)

	// --- Kube-state counts (instant) ---
	g.AddQueryInstant("namespaceCount", Q(MetricKubeNamespacePhase, "").Agg(AggSum, GroupByPhase).Gt("0"), false)
	g.AddQueryInstant("secretCount", Q(MetricKubeSecretInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("deploymentCount", Q(MetricKubeDeploymentReplicas, "").Agg(AggCount), false)
	g.AddQueryInstant("configmapCount", Q(MetricKubeConfigmapInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("serviceCount", Q(MetricKubeServiceInfo, "").Agg(AggCount), false)
	g.AddQueryInstant("routeCount", Q(MetricOCPRouteCreated, "").Agg(AggCount), false)
	g.AddQueryInstant("nodeRoles", Q(MetricKubeNodeRole, ""), false)

	// --- Report-style: max over elapsed time (instant) ---
	g.AddQueryInstant("99thEtcdDiskCommit",
		Q(MetricEtcdDiskCommitDuration, "").Rate(Rate2m).
			Agg(AggSum, GroupByLE).
			HistogramQuantile(P99).
			OverTime(TimeAggMax).
			Gt("0"),
		false,
	)

	// --- Scheduler ---
	g.AddQuery("schedulerRate",
		Q(MetricSchedulerAttempts, `result="scheduled"`).Rate(Rate2m).Agg(AggSum),
	)
	g.HistogramQuantiles("schedulerLatency", MetricSchedulerDuration,
		[]Percentile{P99}, "", nil)

	// --- KubeVirt health (instant) ---
	g.AddQueryInstant("kubevirtHealth",
		Q(MetricKubevirtHCOHealth, ""), false)
	g.AddQueryInstant("kubevirtVMCount",
		Q(MetricKubevirtVMCount, "").Agg(AggSum), false)

	// --- CNV storage IOPS avg/max over time (instant) ---
	g.AddQueryInstant("storageIOPSReadAvg",
		Q(MetricKubevirtStorageIOPSRead, "").Rate(Rate60s).
			OverTime(TimeAggAvg).Agg(AggAvg, GroupByName), false)
	g.AddQueryInstant("storageIOPSReadMax",
		Q(MetricKubevirtStorageIOPSRead, "").Rate(Rate60s).
			OverTime(TimeAggMax).Agg(AggMax, GroupByName), false)

	// --- Process CPU with topk and role filtering ---
	g.AddQuery("kubeletCPU",
		Q(MetricProcessCPU, `service="kubelet",job="kubelet"`).
			IRate(Rate2m).Multiply("100").
			AndOn([]GroupBy{GroupByNode},
				Q(MetricProcessCPU, `service="kubelet",job="kubelet"`).
					IRate(Rate2m).Multiply("100").
					OverTime(TimeAggAvg).TopK(3).
					AndOn([]GroupBy{GroupByNode}, NodeRoleFilter(RoleWorker)),
			),
	)

	// --- Or with default zero ---
	g.AddQueryInstant("virtAPIPods",
		Q(Metric("up"), `namespace="openshift-cnv",pod=~"virt-api-.*"`).
			Agg(AggSum).
			Or(VectorZero()),
		false,
	)

	// --- Output ---
	out, err := g.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		if err := os.WriteFile(os.Args[1], out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Wrote %d metrics to %s\n", g.Count(), os.Args[1])
	} else {
		fmt.Print(string(out))
	}
}
