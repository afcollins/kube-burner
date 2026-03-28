package main

// BuildMetricsAggregatedProfile generates the metrics-aggregated.yml profile (47 metrics).
func BuildMetricsAggregatedProfile(g *Generator) {
	// API server (same as metrics.yml)
	addAPILatencyMetrics(g)

	// Container CPU per role - Masters & Infra are sum-by, Workers are avg-by
	g.Add("containerCPU-Masters",
		`(sum(irate(container_cpu_usage_seconds_total{name!="",container!="POD",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"}[2m]) * 100) by (container, pod, namespace, node) and on (node) kube_node_role{role="master"}) > 0`)
	g.Add("containerCPU-AggregatedWorkers",
		`(avg(irate(container_cpu_usage_seconds_total{name!="",container!="POD",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"}[2m]) * 100 and on (node) kube_node_role{role="worker"}) by (namespace, container)) > 0`)
	g.Add("containerCPU-Infra",
		`(sum(irate(container_cpu_usage_seconds_total{name!="",container!="POD",namespace=~"openshift-(monitoring|sdn|ovn-kubernetes|multus|ingress)|stackrox"}[2m]) * 100) by (container, pod, namespace, node) and on (node) kube_node_role{role="infra"}) > 0`)

	// Container memory per role
	g.Add("containerMemory-Masters",
		`(sum(container_memory_rss{name!="",container!="POD",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"}) by (container, pod, namespace, node) and on (node) kube_node_role{role="master"}) > 0`)
	g.Add("containerMemory-AggregatedWorkers",
		`avg(container_memory_rss{name!="",container!="POD",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"} and on (node) kube_node_role{role="worker"}) by (container, namespace)`)
	g.Add("containerMemory-Infra",
		`(sum(container_memory_rss{name!="",container!="POD",namespace=~"openshift-.*|cilium|stackrox|calico.*|tigera.*"}) by (container, pod, namespace, node) and on (node) kube_node_role{role="infra"}) > 0`)

	// Node CPU per role
	g.Add("nodeCPU-Masters",
		`(sum(irate(node_cpu_seconds_total[2m])) by (mode,instance) and on (instance) label_replace(kube_node_role{role="master"}, "instance", "$1", "node", "(.+)")) > 0`)
	g.Add("nodeCPU-AggregatedWorkers",
		`(avg((sum(irate(node_cpu_seconds_total[2m])) by (mode,instance) and on (instance) label_replace(kube_node_role{role="worker"}, "instance", "$1", "node", "(.+)"))) by (mode)) > 0`)
	g.Add("nodeCPU-Infra",
		`(sum(irate(node_cpu_seconds_total[2m])) by (mode,instance) and on (instance) label_replace(kube_node_role{role="infra"}, "instance", "$1", "node", "(.+)")) > 0`)

	// Node memory utilization
	g.Add("nodeMemoryUtilization-AggregatedWorkers",
		`avg((node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) and on (instance) label_replace(kube_node_role{role="worker"}, "instance", "$1", "node", "(.+)"))`)
	g.Add("nodeMemoryUtilization-Masters",
		`(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) and on (instance) label_replace(kube_node_role{role="master"}, "instance", "$1", "node", "(.+)")`)
	g.Add("nodeMemoryUtilization-Infra",
		`(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) and on (instance) label_replace(kube_node_role{role="infra"}, "instance", "$1", "node", "(.+)")`)

	// Kubelet & CRI-O (topk pattern)
	g.Add("kubeletCPU",
		`irate(process_cpu_seconds_total{service="kubelet",job="kubelet"}[2m]) * 100 and on (node) topk(3,avg_over_time(irate(process_cpu_seconds_total{service="kubelet",job="kubelet"}[2m])[{{ .elapsed }}:]) and on (node) kube_node_role{role="worker"})`)
	g.Add("kubeletMemory",
		`process_resident_memory_bytes{service="kubelet",job="kubelet"} and on (node) topk(3,max_over_time(irate(process_resident_memory_bytes{service="kubelet",job="kubelet"}[2m])[{{ .elapsed }}:]) and on (node) kube_node_role{role="worker"})`)
	g.Add("crioCPU",
		`irate(process_cpu_seconds_total{service="kubelet",job="crio"}[2m]) * 100 and on (node) topk(3,avg_over_time(irate(process_cpu_seconds_total{service="kubelet",job="crio"}[2m])[{{ .elapsed }}:]) and on (node) kube_node_role{role="worker"})`)
	g.Add("crioMemory",
		`process_resident_memory_bytes{service="kubelet",job="crio"} and on (node) topk(3,max_over_time(irate(process_resident_memory_bytes{service="kubelet",job="crio"}[2m])[{{ .elapsed }}:]) and on (node) kube_node_role{role="worker"})`)

	// Etcd (no compaction/defrag in this profile)
	addEtcdMetrics(g)

	// Cluster
	addClusterMetrics(g)

	// Prometheus
	addPrometheusMetrics(g)

	// Retain raw CPU seconds
	addNodeCPUSecondsCapture(g)

	// Cgroup captures
	addCgroupCapture(g)

	// Cgroup CPU irate
	g.Add("cgroupCPU",
		`sum (  irate( container_cpu_usage_seconds_total { id =~ "/system.slice|/system.slice/kubelet.service|.*/ovs-vswitchd.service|/system.slice/crio.service|/kubepods.slice" }[2m])  )  by   (   id , node )`)
	g.Add("cgroupMemoryRSS",
		`sum( container_memory_rss { id =~ "/system.slice|/system.slice/kubelet.service|.*/ovs-vswitchd.service|/system.slice/crio.service|/kubepods.slice"}) by (id, node)`)
}
