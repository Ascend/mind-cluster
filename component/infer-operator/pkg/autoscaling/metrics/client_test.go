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
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ascend-common/common-utils/hwlog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestMetricsClientLifecycle(t *testing.T) {
	client := NewMetricsClient(time.Second)
	key := autoscalingtypes.MetricKey{PANamespace: "default", PAName: "pa", MetricName: "qps"}
	if err := client.UpdateMetrics(time.Now(), key, time.Minute, 30*time.Second); err == nil {
		t.Fatal("empty metrics accepted")
	}
	if err := client.UpdateMetrics(time.Now(), key, time.Minute, 30*time.Second, 2, 4); err != nil {
		t.Fatal(err)
	}
	stable, panicValue, err := client.GetMetricValues(key)
	if err != nil || stable != 3 || panicValue != 3 {
		t.Fatalf("stable=%v panic=%v err=%v", stable, panicValue, err)
	}
	client.DeleteForPodAutoscaler("default", "pa")
	if _, _, err := client.GetMetricValues(key); err == nil {
		t.Fatal("deleted metric window remained available")
	}
}

func TestCollectPartialAndTotalFailure(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Hostname() == "10.0.0.2" {
			return nil, context.DeadlineExceeded
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("# TYPE qps gauge\nqps 8\n")), Header: make(http.Header)}, nil
	})}
	collector := NewCollector(NewMetricsFetcherWithHTTPClient(httpClient, FetcherConfig{}))
	pods := []corev1.Pod{readyPod("one", "10.0.0.1"), readyPod("two", "10.0.0.2")}
	observation, err := collector.Observe(context.Background(), PodMetricQuery{
		Source: apiv1.MetricSource{TargetName: "qps", Port: "8000"}, Pods: pods, Timestamp: time.Now()})
	if err != nil || len(observation.Values) != 1 || observation.Values[0] != 8 {
		t.Fatalf("observation=%+v err=%v", observation, err)
	}
	_, err = collector.Observe(context.Background(), PodMetricQuery{
		Source: apiv1.MetricSource{TargetName: "qps", Port: "8000"},
		Pods:   []corev1.Pod{readyPod("two", "10.0.0.2")}})
	if err == nil {
		t.Fatal("total fetch failure was accepted")
	}
	if _, err := (*Collector)(nil).Observe(context.Background(), PodMetricQuery{}); err == nil {
		t.Fatal("nil collector was accepted")
	}
}

func TestCollectorFetchesPodsConcurrently(t *testing.T) {
	var active int32
	var maximum int32
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		current := atomic.AddInt32(&active, 1)
		for {
			previous := atomic.LoadInt32(&maximum)
			if current <= previous || atomic.CompareAndSwapInt32(&maximum, previous, current) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&active, -1)
		return &http.Response{StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader("# TYPE qps gauge\nqps 8\n")), Header: make(http.Header)}, nil
	})}
	collector := NewCollector(NewMetricsFetcherWithHTTPClient(httpClient, FetcherConfig{}))
	collector.maxConcurrency = 4
	pods := []corev1.Pod{readyPod("one", "10.0.0.1"), readyPod("two", "10.0.0.2"),
		readyPod("three", "10.0.0.3"), readyPod("four", "10.0.0.4")}
	observation, err := collector.Observe(context.Background(), PodMetricQuery{
		Source: apiv1.MetricSource{TargetName: "qps", Port: "8000"}, Pods: pods})
	if err != nil {
		t.Fatal(err)
	}
	if len(observation.Values) != len(pods) {
		t.Fatalf("collected %d values, want %d", len(observation.Values), len(pods))
	}
	if atomic.LoadInt32(&maximum) < 2 {
		t.Fatalf("maximum concurrent requests=%d, want at least 2", maximum)
	}
}

func readyPod(name, ip string) corev1.Pod {
	return corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"}, Status: corev1.PodStatus{
		Phase: corev1.PodRunning, PodIP: ip, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
}
