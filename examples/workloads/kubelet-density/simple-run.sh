#!/bin/bash
kube-burner init --set global.gcMetrics=true --set jobs.0.jobIterations=700 -c kubelet-density.yml -e mep.yml
