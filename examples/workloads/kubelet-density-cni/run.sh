#!/bin/bash
KUBE_BURNER=/root/go/bin/kube-burner
$KUBE_BURNER init -e mep.yml -c kubelet-density-cni.yml \
	--set jobs.0.incrementalLoad.scrapeMetricsPerStep=true \
	--set jobs.0.incrementalLoad.startIterations=25 \
	--set jobs.0.incrementalLoad.totalIterations=75 \
	--set jobs.0.incrementalLoad.pattern.type=linear \
	--set jobs.0.incrementalLoad.pattern.linear.stepSize=50
