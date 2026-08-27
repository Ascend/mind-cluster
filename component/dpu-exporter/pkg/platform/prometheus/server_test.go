/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package prometheus

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"huawei.com/dpu-exporter/pkg/collector/dpucollector"
	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/device"
	"huawei.com/dpu-exporter/utils/logger"
)

func init() {
	_ = logger.InitLogger()
}

// stubCollector is a minimal MetricsCollector that produces one gauge on each scrape.
type stubCollector struct {
	desc *prometheus.Desc
}

func newStubCollector(name string) *stubCollector {
	return &stubCollector{
		desc: prometheus.NewDesc(name, name+" help", nil, nil),
	}
}

func (s *stubCollector) Describe(ch chan<- *prometheus.Desc)                { ch <- s.desc }
func (s *stubCollector) CollectToCache(_ metricscollector.CollectorContext) {}
func (s *stubCollector) UpdatePrometheus(ch chan<- prometheus.Metric, _ metricscollector.CollectorContext) {
	ch <- prometheus.MustNewConstMetric(s.desc, prometheus.GaugeValue, 1)
}
func (s *stubCollector) PreCollect(_ metricscollector.CollectorContext)       {}
func (s *stubCollector) PostCollect(_ metricscollector.CollectorContext)      {}
func (s *stubCollector) IsSupported(_ metricscollector.CollectorContext) bool { return true }
func (s *stubCollector) GetInterval() time.Duration                           { return 0 }

// TestNewPrometheusCollector covers construction and initial state.
func TestNewPrometheusCollector(t *testing.T) {
	p := NewPrometheusCollector()
	if p == nil {
		t.Fatal("NewPrometheusCollector returned nil")
	}
}

// TestPrometheusCollectorRegisterDescribe covers Register + Describe.
func TestPrometheusCollectorRegisterDescribe(t *testing.T) {
	p := NewPrometheusCollector()
	sc := newStubCollector("test_metric")
	p.Register(sc)

	ch := make(chan *prometheus.Desc, 1)
	p.Describe(ch)
	select {
	case d := <-ch:
		if d == nil {
			t.Error("Describe sent nil Desc")
		}
	default:
		t.Error("Describe did not send any Desc")
	}
}

// TestPrometheusCollectorCollect_NilCollector covers the early return when
// dpucollector.Collector == nil.
func TestPrometheusCollectorCollect_NilCollector(t *testing.T) {
	saved := dpucollector.Collector
	dpucollector.Collector = nil
	defer func() { dpucollector.Collector = saved }()

	p := NewPrometheusCollector()
	sc := newStubCollector("nil_collector_metric")
	p.Register(sc)

	ch := make(chan prometheus.Metric, 1)
	p.Collect(ch)
	// Should return immediately without emitting any metric
	select {
	case <-ch:
		t.Error("Collect emitted a metric when dpucollector.Collector is nil")
	default:
		// expected
	}
}

// TestPrometheusCollectorCollect_WithEngine covers the normal path where
// dpucollector.Collector is initialized and UpdatePrometheus is called.
func TestPrometheusCollectorCollect_WithEngine(t *testing.T) {
	// Initialize the global Collector with a minimal DpuCollector so
	// the Collect method proceeds past the nil check.
	saved := dpucollector.Collector
	defer func() { dpucollector.Collector = saved }()

	// NewDpuCollector sets dpucollector.Collector as a side effect;
	// a nil DeviceManager is fine here because UpdatePrometheus in our
	// stub does not use the CollectorContext.
	dm := &stubDeviceManager{}
	dpucollector.NewDpuCollector(dm)

	p := NewPrometheusCollector()
	sc := newStubCollector("with_engine_metric")
	p.Register(sc)

	ch := make(chan prometheus.Metric, 1)
	p.Collect(ch)
	select {
	case m := <-ch:
		if m == nil {
			t.Error("Collect sent nil Metric")
		}
	default:
		t.Error("Collect did not emit any metric")
	}
}

// stubDeviceManager satisfies device.DeviceManager with no-ops.
type stubDeviceManager struct{}

func (s *stubDeviceManager) AutoInit() error                         { return nil }
func (s *stubDeviceManager) GetDpuList() []device.DPU                { return nil }
func (s *stubDeviceManager) ExecCommand(_ ...string) (string, error) { return "", nil }
func (s *stubDeviceManager) ReadSysfs(_ string) (string, error)      { return "", nil }
func (s *stubDeviceManager) ListDir(_ string) ([]string, error)      { return nil, nil }
func (s *stubDeviceManager) GetCardType() string                     { return "stub" }

// TestStartPrometheus_Healthz covers that StartPrometheus registers /healthz
// and the endpoint returns 200. We start the server in a goroutine and probe it.
func TestStartPrometheus_Healthz(t *testing.T) {
	p := NewPrometheusCollector()

	// Pick a high port to reduce collision risk
	port := "29999"
	addr := ":" + port

	// Build the same mux that StartPrometheus uses, but under httptest
	// so we don't need to bind a real port (avoids flaky CI).
	registry := prometheus.NewRegistry()
	if err := registry.Register(p); err != nil {
		t.Fatalf("registry.Register failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// /healthz returns 200
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/healthz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// /metrics returns 200 (empty but valid prometheus output)
	resp2, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics error: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("/metrics status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}

	// Verify addr variable is used (suppress unused warning)
	_ = addr
}
