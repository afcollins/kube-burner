package main

// BuildBuildFarmProfile generates the build-farm-metrics.yml profile.
func BuildBuildFarmProfile(g *Generator) {
	// etcd DB Size
	g.AddQuery("etcdDBTotalSize",
		QRaw(MetricEtcdDBSize).Agg(AggAvg))

	g.AddQuery("etcdDBSizeInUse",
		QRaw(MetricEtcdDBSizeInUse).Agg(AggAvg))

	g.Add("etcdDBFragmentationBytes",
		"avg(etcd_mvcc_db_total_size_in_bytes) - avg(etcd_mvcc_db_total_size_in_use_in_bytes)")
}
