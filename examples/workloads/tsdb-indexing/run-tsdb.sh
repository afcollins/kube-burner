#!/bin/bash -x
git diff --stat

git log -1 --oneline

export KUBE_BURNER=/Users/ancollin/go/src/github.com/kube-burner/kube-burner/bin/arm64/kube-burner

#echo init
#$KUBE_BURNER init -c kubelet-density.yml -e metrics-endpoints.yaml

# A more interesting 2h time range, some kind of test with churn:
# schedulableControlPlanes.. few nodes, but 3 with 576vCPUs per machine (very large number of node_cpu per machine)
# I think it is this 7.9 GB chunk: 01KBEQHAVJHN2SS5S78EXCK7ZD.tar
# /Users/ancollin/Downloads/schedCP_cloud17/prom-db-raw-for-cardinality-analysis/prometheus-k8s-db
# % gdate +"%s" --date='2025-12-01 17:57:35Z'
# 1764611855
# % gdate -u +"%F-%H%M%S" --date=@1764611855
# 2025-12-01-175735
# % gdate +"%s" --date='2025-12-01 19:15:21Z'
# 1764616521
# % gdate -u +"%F-%H%M%S" --date=@1764616521
# 2025-12-01-191521
#echo index
#$KUBE_BURNER index -e metrics-endpoints-tsdb.yaml --start 1764611855 --end 1764616521
#echo index-kbo
#$KUBE_BURNER index -e metrics-endpoints-tsdb-kbo.yaml --start 1764611855 --end 1764616521
#$KUBE_BURNER index -e metrics-endpoints-tsdb.yaml --start 1764611855 --end 1764616521

# TODO do a large prom chunk (500 nodes) and check the time range (10m, 30m, 20h)


# A test range of 500x 16vCPU workers.
# gdate +"%s" --date='2026-02-10 11:36:00Z'
# 1770723360
# gdate +"%s" --date='2026-02-10 14:56:00Z'
# 1770735360


# A test range 1200s = 20 minutes
#$KUBE_BURNER index -e metrics-endpoints-tsdb-kbo-reprocessed.yaml --start 1770723360 --end 1770724560
# The full range (two tests, actually)
#$KUBE_BURNER index -e metrics-endpoints-tsdb-kbo-reprocessed.yaml --start 1770723360 --end 1770735360


# a prow 252 node run so I can use the raw dataset - sample size
echo 5m block
$KUBE_BURNER index -e metrics-endpoints-tsdb.yaml --start 1774229400 --end 1774230000
# 2 hour block:
# (performance-dashboards) {26-03-24 13:04}ancollin-mac:~/go/src/github.com/kube-burner/kube-burner/examples/workloads/tsdb-indexing@prom-exp✗✗ ancollin% gdate +"%s" --date='2026-03-23 01:45:00Z'
# 1774230300
# (performance-dashboards) {26-03-24 13:05}ancollin-mac:~/go/src/github.com/kube-burner/kube-burner/examples/workloads/tsdb-indexing@prom-exp✗✗ ancollin% gdate +"%s" --date='2026-03-23 03:45:00Z'
# 1774237500
echo 2h block
$KUBE_BURNER index -e metrics-endpoints-tsdb.yaml --start 1774230300 --end 1774237500
# Do the largest block we can from this run 01:45 to 06:30 (4h45m)
#echo 3h15m block
#$KUBE_BURNER index -e metrics-endpoints-tsdb.yaml --start 1774230300 --end 1774238400
