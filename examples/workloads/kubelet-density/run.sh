#!/bin/bash
KUBE_BURNER=/root/git/kube-burner/bin/amd64/kube-burner
$KUBE_BURNER init  -c kubelet-density.yml -e mep.yml \
	--set jobs.0.incrementalLoad.scrapeMetricsPerStep=true \
	--set jobs.0.incrementalLoad.pattern.type=linear \
	--set jobs.0.incrementalLoad.pattern.linear.stepSize=10 \
	--set jobs.0.incrementalLoad.startIterations=10 \
	--set jobs.0.incrementalLoad.totalIterations=40
