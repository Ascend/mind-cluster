/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package metrics

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev1 "k8s.io/api/core/v1"

	apiv1 "infer-operator/pkg/api/v1"
)

const (
	maxMetricsResponseSize = 4 * 1024 * 1024
	maxErrorResponseSize   = 1024
)

// FetcherConfig controls HTTP timeouts, retries, backoff, and TLS verification.
type FetcherConfig struct {
	Timeout     time.Duration
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	InsecureTLS bool
}

// DefaultMetricsFetcherConfig returns conservative defaults for Pod-local metric requests.
func DefaultMetricsFetcherConfig() FetcherConfig {
	return FetcherConfig{Timeout: 2 * time.Second, MaxRetries: 1, BaseDelay: 200 * time.Millisecond,
		MaxDelay: time.Second, InsecureTLS: false}
}

// Fetcher retrieves Prometheus text-format metrics from Pod endpoints.
type Fetcher struct {
	client *http.Client
	config FetcherConfig
}

// NewMetricsFetcher creates a fetcher with default settings.
func NewMetricsFetcher() *Fetcher { return NewMetricsFetcherWithConfig(DefaultMetricsFetcherConfig()) }

// NewMetricsFetcherWithConfig creates a fetcher with explicit settings.
func NewMetricsFetcherWithConfig(config FetcherConfig) *Fetcher {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureTLS}}
	return &Fetcher{client: &http.Client{Timeout: config.Timeout, Transport: transport}, config: config}
}

// NewMetricsFetcherWithHTTPClient creates a fetcher with an injected HTTP client.
func NewMetricsFetcherWithHTTPClient(client *http.Client, config FetcherConfig) *Fetcher {
	if client == nil {
		return NewMetricsFetcherWithConfig(config)
	}
	return &Fetcher{client: client, config: config}
}

// FetchMetrics fetches one named metric from http://<pod-ip>:<port><path>.
// Prometheus counters and gauges return their value; histograms and summaries
// return sample_sum because the API currently represents each source as one scalar.
func (fetcher *Fetcher) FetchMetrics(ctx context.Context, pod corev1.Pod,
	source apiv1.MetricSource) (float64, error) {
	if pod.Status.PodIP == "" {
		return 0, fmt.Errorf("pod %s/%s has no IP address", pod.Namespace, pod.Name)
	}
	path := source.Path
	if path == "" {
		path = "/metrics"
	}
	if source.Port == "" {
		return 0, fmt.Errorf("metric source port is required")
	}
	endpoint := fmt.Sprintf("http://%s:%s%s", pod.Status.PodIP, source.Port, path)
	var lastErr error
	for attempt := 0; attempt <= fetcher.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := fetcher.config.BaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			if delay > fetcher.config.MaxDelay {
				delay = fetcher.config.MaxDelay
			}
			if err := waitForRetry(ctx, delay); err != nil {
				return 0, err
			}
		}
		families, err := fetcher.fetchAll(ctx, endpoint)
		if err != nil {
			lastErr = err
			continue
		}
		family, found := families[source.TargetName]
		if !found || len(family.Metric) == 0 {
			return 0, fmt.Errorf("metric %q not found at %s", source.TargetName, endpoint)
		}
		return metricValue(family, family.Metric[0])
	}
	return 0, fmt.Errorf("failed to fetch metric %q from %s: %w", source.TargetName, endpoint, lastErr)
}

func (fetcher *Fetcher) fetchAll(ctx context.Context, endpoint string) (map[string]*dto.MetricFamily, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxErrorResponseSize))
		if readErr != nil {
			return nil, fmt.Errorf("metrics endpoint returned %d and response body could not be read: %w",
				response.StatusCode, readErr)
		}
		return nil, fmt.Errorf("metrics endpoint returned %d: %s", response.StatusCode, body)
	}
	limitedBody := &io.LimitedReader{R: response.Body, N: maxMetricsResponseSize + 1}
	var parser expfmt.TextParser
	families, parseErr := parser.TextToMetricFamilies(limitedBody)
	if limitedBody.N <= 0 {
		return nil, fmt.Errorf("metrics response exceeds %d bytes", maxMetricsResponseSize)
	}
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse metrics response: %w", parseErr)
	}
	return families, nil
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func metricValue(family *dto.MetricFamily, metric *dto.Metric) (float64, error) {
	switch family.GetType() {
	case dto.MetricType_COUNTER:
		return metric.GetCounter().GetValue(), nil
	case dto.MetricType_GAUGE:
		return metric.GetGauge().GetValue(), nil
	case dto.MetricType_HISTOGRAM:
		return metric.GetHistogram().GetSampleSum(), nil
	case dto.MetricType_SUMMARY:
		return metric.GetSummary().GetSampleSum(), nil
	default:
		return 0, fmt.Errorf("metric has unsupported type %s", family.GetType())
	}
}
