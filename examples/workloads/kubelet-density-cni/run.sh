#!/bin/bash
KUBE_BURNER=/root/git/kube-burner/bin/amd64/kube-burner
$KUBE_BURNER init --set jobs.0.incrementalLoad.scrapeMetricsPerStep=true --set jobs.0.incrementalLoad.pattern.type=linear --set jobs.0.incrementalLoad.pattern.linear.stepSize=50 --set jobs.0.incrementalLoad.startIterations=600 --set jobs.0.incrementalLoad.totalIterations=900 -c kubelet-density-cni.yml -e mep.yml
