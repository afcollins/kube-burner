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

package prometheus

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloud-bulldozer/go-commons/v2/indexers"
	"github.com/kube-burner/kube-burner/v2/pkg/config"
	"github.com/prometheus/common/model"
)

func TestPrometheus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus Suite")
}

var _ = Describe("createMetric groupId tagging", func() {
	var (
		p          Prometheus
		now        time.Time
		groupStart time.Time
		groupEnd   time.Time
		labels     model.Metric
	)

	BeforeEach(func() {
		p = Prometheus{
			UUID:     "test-uuid",
			metadata: map[string]any{},
		}
		now = time.Now().UTC()
		groupStart = now.Add(-10 * time.Minute)
		groupEnd = now.Add(10 * time.Minute)
		labels = model.Metric{"__name__": "test_metric"}
	})

	Context("when the job has no group windows", func() {
		It("should not tag the datapoint with a groupId", func() {
			job := Job{JobConfig: config.Job{Name: "job-1"}}
			m := p.createMetric("query", "test_metric", job, labels, 1.0, now, false)
			meta, ok := m.Metadata.(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(meta).ToNot(HaveKey("groupId"))
		})
	})

	Context("when the datapoint falls within a group window", func() {
		It("should tag the datapoint with the matching groupId", func() {
			windows := []config.GroupWindow{{ID: 2, Start: groupStart, End: groupEnd}}
			job := Job{JobConfig: config.Job{Name: "job-1", GroupWindows: &windows}}
			m := p.createMetric("query", "test_metric", job, labels, 1.0, now, false)
			meta, ok := m.Metadata.(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(meta["groupId"]).To(Equal(2))
		})
	})

	Context("when the datapoint falls outside all group windows", func() {
		It("should not tag the datapoint with a groupId", func() {
			windows := []config.GroupWindow{{ID: 1, Start: groupStart, End: groupEnd}}
			job := Job{JobConfig: config.Job{Name: "job-1", GroupWindows: &windows}}
			m := p.createMetric("query", "test_metric", job, labels, 1.0, now.Add(20*time.Minute), false)
			meta, ok := m.Metadata.(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(meta).ToNot(HaveKey("groupId"))
		})
	})

	Context("with multiple group windows", func() {
		It("should tag each datapoint with the groupId of the window it falls into", func() {
			g1Start := now.Add(-30 * time.Minute)
			g1End := now.Add(-20 * time.Minute)
			g2Start := now.Add(-10 * time.Minute)
			g2End := now.Add(10 * time.Minute)
			windows := []config.GroupWindow{
				{ID: 1, Start: g1Start, End: g1End},
				{ID: 2, Start: g2Start, End: g2End},
			}
			job := Job{JobConfig: config.Job{Name: "job-1", GroupWindows: &windows}}

			m1 := p.createMetric("query", "test_metric", job, labels, 1.0, now.Add(-25*time.Minute), false)
			meta1 := m1.Metadata.(map[string]any)
			Expect(meta1["groupId"]).To(Equal(1))

			m2 := p.createMetric("query", "test_metric", job, labels, 1.0, now, false)
			meta2 := m2.Metadata.(map[string]any)
			Expect(meta2["groupId"]).To(Equal(2))
		})
	})

	Context("when the query is instant", func() {
		It("should not tag the datapoint with a groupId", func() {
			windows := []config.GroupWindow{{ID: 2, Start: groupStart, End: groupEnd}}
			job := Job{JobConfig: config.Job{Name: "job-1", GroupWindows: &windows}}
			m := p.createMetric("query", "test_metric", job, labels, 1.0, now, true)
			meta, ok := m.Metadata.(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(meta).ToNot(HaveKey("groupId"))
		})
	})
})

var _ indexers.TSDBSampleConvertible = metric{}

var _ = Describe("metric.ToTSDBSamples", func() {
	It("produces correct labels, timestamp, and value", func() {
		ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		m := metric{
			Timestamp:  ts,
			Labels:     map[string]string{"instance": "node1", "namespace": "default"},
			Value:      42.5,
			UUID:       "test-uuid",
			MetricName: "cpuUsage",
			JobName:    "test-job",
		}

		samples := m.ToTSDBSamples("cpuUsage")
		Expect(samples).To(HaveLen(1))

		s := samples[0]
		Expect(s.Value).To(Equal(42.5))
		Expect(s.Timestamp).To(Equal(ts.UnixMilli()))
		Expect(s.Labels.Get("__name__")).To(Equal("cpuUsage"))
		Expect(s.Labels.Get("instance")).To(Equal("node1"))
		Expect(s.Labels.Get("namespace")).To(Equal("default"))
		Expect(s.Labels.Get("uuid")).To(Equal("test-uuid"))
		Expect(s.Labels.Get("job_name")).To(Equal("test-job"))
	})

	It("omits uuid and job_name labels when empty", func() {
		ts := time.Now().UTC()
		m := metric{
			Timestamp: ts,
			Labels:    map[string]string{"foo": "bar"},
			Value:     1.0,
		}

		samples := m.ToTSDBSamples("myMetric")
		Expect(samples).To(HaveLen(1))
		Expect(samples[0].Labels.Get("uuid")).To(Equal(""))
		Expect(samples[0].Labels.Get("job_name")).To(Equal(""))
		Expect(samples[0].Labels.Get("foo")).To(Equal("bar"))
	})

	It("uses metricName argument, not struct field", func() {
		m := metric{
			Timestamp:  time.Now().UTC(),
			Value:      1.0,
			MetricName: "structName",
		}

		samples := m.ToTSDBSamples("overrideName")
		Expect(samples[0].Labels.Get("__name__")).To(Equal("overrideName"))
	})
})
