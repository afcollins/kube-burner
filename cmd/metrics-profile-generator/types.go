package main

// ---------------------------------------------------------------------------
// Typed enums for type-safe metrics profile generation
// ---------------------------------------------------------------------------

// Metric represents a Prometheus metric name.
type Metric string

const (
	// API Server
	MetricAPIServerRequestDuration Metric = "apiserver_request_duration_seconds"
	MetricAPIServerRequestTotal    Metric = "apiserver_request_total"
	MetricAPIServerInflightReqs    Metric = "apiserver_current_inflight_requests"

	// Container
	MetricContainerCPU         Metric = "container_cpu_usage_seconds_total"
	MetricContainerMemoryRSS   Metric = "container_memory_rss"
	MetricContainerMemoryWS    Metric = "container_memory_working_set_bytes"
	MetricContainerFSWrites    Metric = "container_fs_writes_bytes_total"
	MetricContainerCRIOLatency Metric = "container_runtime_crio_operations_latency_microseconds"

	// Process (kubelet, crio, etc.)
	MetricProcessCPU    Metric = "process_cpu_seconds_total"
	MetricProcessMemory Metric = "process_resident_memory_bytes"

	// Node
	MetricNodeCPU              Metric = "node_cpu_seconds_total"
	MetricNodeMemoryTotal      Metric = "node_memory_MemTotal_bytes"
	MetricNodeMemoryAvailable  Metric = "node_memory_MemAvailable_bytes"
	MetricNodeMemoryActive     Metric = "node_memory_Active_bytes"
	MetricNodeMemoryCached     Metric = "node_memory_Cached_bytes"
	MetricNodeMemoryBuffers    Metric = "node_memory_Buffers_bytes"
	MetricNodeNetworkRx        Metric = "node_network_receive_bytes_total"
	MetricNodeNetworkTx        Metric = "node_network_transmit_bytes_total"
	MetricNodeNetworkRxErrs    Metric = "node_network_receive_errs_total"
	MetricNodeNetworkTxErrs    Metric = "node_network_transmit_errs_total"
	MetricNodeNetworkTxDrop    Metric = "node_network_transmit_drop"
	MetricNodeNetworkRxDrop    Metric = "node_network_receive_drop_total"
	MetricNodeDiskWritten      Metric = "node_disk_written_bytes_total"
	MetricNodeDiskRead         Metric = "node_disk_read_bytes_total"
	MetricNodeVmstatPgmajfault Metric = "node_vmstat_pgmajfault"

	// Etcd
	MetricEtcdLeaderChanges        Metric = "etcd_server_leader_changes_seen_total"
	MetricEtcdDBSize               Metric = "etcd_mvcc_db_total_size_in_bytes"
	MetricEtcdDBSizeInUse          Metric = "etcd_mvcc_db_total_size_in_use_in_bytes"
	MetricEtcdDiskCommitDuration   Metric = "etcd_disk_backend_commit_duration_seconds"
	MetricEtcdDiskWALSyncDuration  Metric = "etcd_disk_wal_fsync_duration_seconds"
	MetricEtcdNetworkPeerRoundTrip Metric = "etcd_network_peer_round_trip_time_seconds"
	MetricEtcdClusterVersion       Metric = "etcd_cluster_version"
	MetricEtcdCompactionDuration   Metric = "etcd_debugging_mvcc_db_compaction_total_duration_milliseconds_sum"
	MetricEtcdDefragDuration       Metric = "etcd_disk_backend_defrag_duration_seconds_sum"

	// Kube-state
	MetricKubeNamespacePhase      Metric = "kube_namespace_status_phase"
	MetricKubePodStatusPhase      Metric = "kube_pod_status_phase"
	MetricKubePodInfo             Metric = "kube_pod_info"
	MetricKubeSecretInfo          Metric = "kube_secret_info"
	MetricKubeDeploymentReplicas  Metric = "kube_deployment_spec_replicas"
	MetricKubeReplicaSetReplicas  Metric = "kube_replicaset_spec_replicas"
	MetricKubeConfigmapInfo       Metric = "kube_configmap_info"
	MetricKubeServiceInfo         Metric = "kube_service_info"
	MetricKubeNodeRole            Metric = "kube_node_role"
	MetricKubeNodeStatusCondition Metric = "kube_node_status_condition"
	MetricKubeJobInfo             Metric = "kube_job_info"

	// OpenShift
	MetricOCPRouteCreated    Metric = "openshift_route_created"
	MetricOCPTSDBHeadSeries  Metric = "openshift:prometheus_tsdb_head_series:sum"
	MetricOCPTSDBHeadSamples Metric = "openshift:prometheus_tsdb_head_samples_appended_total:sum"

	// Scheduler
	MetricSchedulerAttempts Metric = "scheduler_schedule_attempts_total"
	MetricSchedulerDuration Metric = "scheduler_e2e_scheduling_duration_seconds"

	// KubeVirt / CNV
	MetricKubevirtHCOHealth          Metric = "kubevirt_hyperconverged_operator_health_status"
	MetricKubevirtSystemHealth       Metric = "kubevirt_hco_system_health_status"
	MetricKubevirtVMCount            Metric = "kubevirt_number_of_vms"
	MetricKubevirtVMIRunning         Metric = "cnv:vmi_status_running:count"
	MetricKubevirtVMICPUCores        Metric = "cluster:vmi_request_cpu_cores:sum"
	MetricKubevirtVMResourceRequests Metric = "kubevirt_vm_resource_requests"
	MetricKubevirtMemoryDelta        Metric = "kubevirt_memory_delta_from_requested_bytes"
	MetricKubevirtLauncherOverhead   Metric = "kubevirt_vmi_launcher_memory_overhead_bytes"
	MetricKubevirtStorageIOPSRead    Metric = "kubevirt_vmi_storage_iops_read_total"
	MetricKubevirtStorageIOPSWrite   Metric = "kubevirt_vmi_storage_iops_write_total"
	MetricKubevirtStorageReadBytes   Metric = "kubevirt_vmi_storage_read_traffic_bytes_total"
	MetricKubevirtStorageWriteBytes  Metric = "kubevirt_vmi_storage_write_traffic_bytes_total"
	MetricKubevirtNetworkRx          Metric = "kubevirt_vmi_network_receive_bytes_total"
	MetricKubevirtNetworkTx          Metric = "kubevirt_vmi_network_transmit_bytes_total"
	MetricKubevirtMigrationSucceeded Metric = "kubevirt_vmi_migration_succeeded"
	MetricKubevirtMigrationsSchedule Metric = "kubevirt_vmi_migrations_in_scheduling_phase"
	MetricKubevirtMigrationsRunning  Metric = "kubevirt_vmi_migrations_in_running_phase"
	MetricKubevirtMigrationsPending  Metric = "kubevirt_vmi_migrations_in_pending_phase"

	// OVS
	MetricOVSBuildInfo Metric = "ovs_build_info"

	// Kueue
	MetricKueueAdmissionWaitTime Metric = "kueue_admission_wait_time_seconds"
	MetricKueueBuildInfo         Metric = "kueue_build_info"

	// Cluster-level recording rules
	MetricClusterMemoryUsageRatio Metric = "cluster:memory_usage:ratio"
	MetricClusterNodeCPURatio     Metric = "cluster:node_cpu:ratio"

	// EgressIP
	MetricEgressIPStartupLatency  Metric = "scale_eip_startup_latency_total"
	MetricEgressIPRecoveryLatency Metric = "scale_eip_recovery_latency"
	MetricStartupNonEIPTotal      Metric = "scale_startup_non_eip_total"
)

// Percentile represents a histogram quantile value.
type Percentile struct {
	Value string // e.g. "0.99"
	Label string // e.g. "P99"
}

var (
	P50  = Percentile{"0.50", "P50"}
	P90  = Percentile{"0.90", "P90"}
	P95  = Percentile{"0.95", "P95"}
	P99  = Percentile{"0.99", "P99"}
	P999 = Percentile{"0.999", "P999"}
)

// GroupBy represents a Prometheus label to group by.
type GroupBy string

const (
	GroupByPod            GroupBy = "pod"
	GroupByNamespace      GroupBy = "namespace"
	GroupByNode           GroupBy = "node"
	GroupByInstance        GroupBy = "instance"
	GroupByVerb           GroupBy = "verb"
	GroupByResource       GroupBy = "resource"
	GroupByCode           GroupBy = "code"
	GroupByScope          GroupBy = "scope"
	GroupByContainer      GroupBy = "container"
	GroupByDevice         GroupBy = "device"
	GroupByPhase          GroupBy = "phase"
	GroupByCondition      GroupBy = "condition"
	GroupByMode           GroupBy = "mode"
	GroupByID             GroupBy = "id"
	GroupByJob            GroupBy = "job"
	GroupByName           GroupBy = "name"
	GroupByLE             GroupBy = "le"
	GroupBySubresource    GroupBy = "subresource"
	GroupByClusterVersion GroupBy = "cluster_version"
	GroupByResult         GroupBy = "result"
	GroupByRequestKind    GroupBy = "request_kind"
)

// RateInterval represents a Prometheus rate/irate window.
type RateInterval string

const (
	Rate1m  RateInterval = "1m"
	Rate2m  RateInterval = "2m"
	Rate5m  RateInterval = "5m"
	Rate10m RateInterval = "10m"
	Rate60s RateInterval = "60s"
)

// NodeRole represents a Kubernetes node role for filtering.
type NodeRole string

const (
	RoleMaster NodeRole = "master"
	RoleWorker NodeRole = "worker"
	RoleInfra  NodeRole = "infra"
)

// AggFunc represents a PromQL aggregation function.
type AggFunc string

const (
	AggSum   AggFunc = "sum"
	AggAvg   AggFunc = "avg"
	AggMax   AggFunc = "max"
	AggMin   AggFunc = "min"
	AggCount AggFunc = "count"
)

// TimeAggFunc represents a PromQL over-time aggregation function.
type TimeAggFunc string

const (
	TimeAggAvg TimeAggFunc = "avg_over_time"
	TimeAggMax TimeAggFunc = "max_over_time"
	TimeAggMin TimeAggFunc = "min_over_time"
)

// metricDefinition is the output YAML structure matching kube-burner's format.
type metricDefinition struct {
	Query        string `yaml:"query"`
	MetricName   string `yaml:"metricName"`
	Instant      bool   `yaml:"instant,omitempty"`
	CaptureStart bool   `yaml:"captureStart,omitempty"`
}
