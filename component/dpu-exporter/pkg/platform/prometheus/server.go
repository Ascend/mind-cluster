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

// Package prometheus implements the Prometheus platform adapter for dpu-exporter.
package prometheus

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"huawei.com/dpu-exporter/pkg/collector/dpucollector"
	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/utils/logger"
)

// PrometheusCollector implements prometheus.Collector.
// It aggregates metrics from all registered dpu-exporter MetricsCollectors
// and exposes them to Prometheus on each scrape.
type PrometheusCollector struct {
	mu         sync.Mutex
	collectors []metricscollector.MetricsCollector
}

// NewPrometheusCollector creates a new PrometheusCollector
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{}
}

// Register adds a dpu-exporter MetricsCollector to be scraped
func (p *PrometheusCollector) Register(c metricscollector.MetricsCollector) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.collectors = append(p.collectors, c)
}

// Describe implements prometheus.Collector
func (p *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.collectors {
		c.Describe(ch)
	}
}

// Collect implements prometheus.Collector.
// On each scrape, it calls UpdatePrometheus on all registered collectors,
// which read from the DpuCache and emit Prometheus metrics.
func (p *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := dpucollector.Collector
	if n == nil {
		logger.Warn("dpu collector not initialized, no metrics to expose")
		return
	}

	for _, c := range p.collectors {
		begin := time.Now()
		c.UpdatePrometheus(ch, n)
		logger.Debugf("prometheus collect %v, time cost: %v",
			metricscollector.GetCacheKey(c), time.Since(begin))
	}
}

// StartPrometheus starts the HTTP server that serves /metrics.
func StartPrometheus(port string, collector *PrometheusCollector) error {
	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":" + port
	logger.Infof("starting prometheus metrics server on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return srv.ListenAndServe()
}
