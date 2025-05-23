// Copyright 2024 The Kube-burner Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package measurements

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-bulldozer/go-commons/v2/indexers"
	"github.com/kube-burner/kube-burner/pkg/config"
	"github.com/kube-burner/kube-burner/pkg/measurements/metrics"
	"github.com/kube-burner/kube-burner/pkg/measurements/types"
	kutil "github.com/kube-burner/kube-burner/pkg/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
)

var (
	supportedLatencyMetricsMap = map[string]struct{}{
		"P99": {},
		"P95": {},
		"P50": {},
		"Avg": {},
		"Max": {},
	}
)

type BaseMeasurementFactory struct {
	Config   types.Measurement
	Uuid     string
	Runid    string
	Metadata map[string]interface{}
}

type BaseMeasurement struct {
	Config     types.Measurement
	Uuid       string
	Runid      string
	JobConfig  *config.Job
	ClientSet  kubernetes.Interface
	RestConfig *rest.Config
	Metadata   map[string]interface{}
}

func NewBaseMeasurementFactory(configSpec config.Spec, measurement types.Measurement, metadata map[string]interface{}) BaseMeasurementFactory {
	return BaseMeasurementFactory{
		Config:   measurement,
		Uuid:     configSpec.GlobalConfig.UUID,
		Runid:    configSpec.GlobalConfig.RUNID,
		Metadata: metadata,
	}
}

func (bmf BaseMeasurementFactory) NewBaseLatency(jobConfig *config.Job, clientSet kubernetes.Interface, restConfig *rest.Config) BaseMeasurement {
	return BaseMeasurement{
		Config:     bmf.Config,
		Uuid:       bmf.Uuid,
		Runid:      bmf.Runid,
		JobConfig:  jobConfig,
		ClientSet:  clientSet,
		RestConfig: restConfig,
		Metadata:   bmf.Metadata,
	}
}

func VerifyMeasurementConfig(config types.Measurement, supportedConditions map[string]struct{}) error {
	for _, th := range config.LatencyThresholds {
		if _, supported := supportedConditions[th.ConditionType]; !supported {
			return fmt.Errorf("unsupported condition type in measurement: %s", th.ConditionType)
		}
		if _, supportedLatency := supportedLatencyMetricsMap[th.Metric]; !supportedLatency {
			return fmt.Errorf("unsupported metric %s in measurement, supported are: %s", th.Metric, strings.Join(maps.Keys(supportedLatencyMetricsMap), ", "))
		}
	}
	return nil
}

func IndexLatencyMeasurement(config types.Measurement, jobName string, metricMap map[string][]interface{}, indexerList map[string]indexers.Indexer) {
	indexDocuments := func(indexer indexers.Indexer, metricName string, data []interface{}) {
		log.Infof("Indexing metric %s", metricName)
		indexingOpts := indexers.IndexingOpts{
			MetricName: fmt.Sprintf("%s-%s", metricName, jobName),
		}
		log.Debugf("Indexing [%d] documents: %s", len(data), metricName)
		resp, err := indexer.Index(data, indexingOpts)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(resp)
		}
	}
	for metricName, data := range metricMap {
		// Use the configured TimeseriesIndexer or QuantilesIndexer when specified or else use all indexers
		if config.TimeseriesIndexer != "" && (metricName == podLatencyMeasurement || metricName == svcLatencyMeasurement || metricName == nodeLatencyMeasurement || metricName == pvcLatencyMeasurement) {
			indexer := indexerList[config.TimeseriesIndexer]
			indexDocuments(indexer, metricName, data)
		} else if config.QuantilesIndexer != "" && (metricName == podLatencyQuantilesMeasurement || metricName == svcLatencyQuantilesMeasurement || metricName == nodeLatencyQuantilesMeasurement || metricName == pvcLatencyQuantilesMeasurement) {
			indexer := indexerList[config.QuantilesIndexer]
			indexDocuments(indexer, metricName, data)
		} else {
			for _, indexer := range indexerList {
				indexDocuments(indexer, metricName, data)
			}
		}
	}

}

// Common function to calculate quantiles for both node and pod latencies
// Receives a list of normalized latencies and a function to get the latencies for each condition
func CalculateQuantiles(uuid, jobName string, metadata map[string]interface{}, normLatencies []interface{}, getLatency func(interface{}) map[string]float64, metricName string) []interface{} {
	quantileMap := map[string][]float64{}
	for _, normLatency := range normLatencies {
		for condition, latency := range getLatency(normLatency) {
			quantileMap[condition] = append(quantileMap[condition], latency)
		}
	}
	calcSummary := func(name string, inputLatencies []float64) metrics.LatencyQuantiles {
		latencySummary := metrics.NewLatencySummary(inputLatencies, name)
		latencySummary.UUID = uuid
		latencySummary.Metadata = metadata
		latencySummary.MetricName = metricName
		latencySummary.JobName = jobName
		return latencySummary
	}

	var latencyQuantiles []interface{}
	for condition, latencies := range quantileMap {
		latencyQuantiles = append(latencyQuantiles, calcSummary(condition, latencies))
	}

	return latencyQuantiles
}

func getIntFromLabels(labels map[string]string, key string) int {
	strVal, ok := labels[key]
	if ok {
		val, err := strconv.Atoi(strVal)
		if err == nil {
			return val
		}
	}
	return 0
}

func deployPodInNamespace(clientSet kubernetes.Interface, namespace, podName, image string, command []string) error {
	var podObj = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: ptr.To[int64](0),
			Containers: []corev1.Container{
				{
					Image:           image,
					Command:         command,
					Name:            podName,
					ImagePullPolicy: corev1.PullAlways,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: ptr.To[bool](false),
						Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						RunAsNonRoot:             ptr.To[bool](true),
						SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
						RunAsUser:                ptr.To[int64](1000),
					},
				},
			},
		},
	}

	var err error
	if err = kutil.CreateNamespace(clientSet, namespace, nil, nil); err != nil {
		return err
	}
	if _, err = clientSet.CoreV1().Pods(namespace).Create(context.TODO(), podObj, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Warn(err)
		} else {
			return err
		}
	}
	err = wait.PollUntilContextCancel(context.TODO(), 100*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		pod, err := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return true, err
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		return true, nil
	})
	return err
}

func getGroupVersionClient(restConfig *rest.Config, gv schema.GroupVersion, types ...runtime.Object) *rest.RESTClient {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	// Add CDI objects to the scheme
	metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)
	scheme.AddKnownTypes(gv, types...)

	shallowCopy := *restConfig
	shallowCopy.ContentConfig.GroupVersion = &gv

	shallowCopy.APIPath = "/apis"
	shallowCopy.NegotiatedSerializer = codecs.WithoutConversion()
	shallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()

	cdiClient, err := rest.RESTClientFor(&shallowCopy)
	if err != nil {
		log.Fatalf("failed to create CDI Client - %v", err)
	}
	return cdiClient
}
