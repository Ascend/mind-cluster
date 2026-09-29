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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	apiv1 "infer-operator/pkg/api/v1"
)

type errorReadCloser struct{ err error }

func (reader errorReadCloser) Read([]byte) (int, error) { return 0, reader.err }
func (errorReadCloser) Close() error                    { return nil }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestFetcherReadsPrometheusGauge(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
			"# TYPE waiting gauge\nwaiting 12\n")), Header: make(http.Header)}, nil
	})}
	fetcher := NewMetricsFetcherWithHTTPClient(client, FetcherConfig{Timeout: time.Second})
	value, err := fetcher.FetchMetrics(context.Background(), corev1.Pod{Status: corev1.PodStatus{PodIP: "10.0.0.1"}},
		apiv1.MetricSource{TargetName: "waiting", Port: "8000", Path: "/metrics"})
	if err != nil {
		t.Fatal(err)
	}
	if value != 12 {
		t.Fatalf("value=%v, want 12", value)
	}
}

func TestFetcherRejectsOversizedResponse(t *testing.T) {
	body := strings.Repeat("# padding\n", maxMetricsResponseSize/len("# padding\n")+2)
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)),
			Header: make(http.Header)}, nil
	})}
	fetcher := NewMetricsFetcherWithHTTPClient(client, FetcherConfig{})
	if _, err := fetcher.fetchAll(context.Background(), "http://10.0.0.1:8000/metrics"); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized response error=%v", err)
	}
}

func TestFetcherHandlesErrorResponseReadFailure(t *testing.T) {
	wantErr := errors.New("read failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusInternalServerError,
			Body: errorReadCloser{err: wantErr}, Header: make(http.Header)}, nil
	})}
	fetcher := NewMetricsFetcherWithHTTPClient(client, FetcherConfig{})
	if _, err := fetcher.fetchAll(context.Background(), "http://10.0.0.1:8000/metrics"); err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("response read error=%v", err)
	}
}

func TestWaitForRetryStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForRetry(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForRetry error=%v", err)
	}
}
