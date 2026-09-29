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
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

const defaultCollectionConcurrency = 8

// PodMetricQuery identifies the Pods and endpoint used for one observation.
type PodMetricQuery struct {
	Namespace  string
	TargetName string
	Source     apiv1.MetricSource
	Pods       []corev1.Pod
	Timestamp  time.Time
}

// ObservationCollector collects one metric observation from a set of Pods.
type ObservationCollector interface {
	Observe(context.Context, PodMetricQuery) (autoscalingtypes.MetricObservation, error)
}

// Collector owns Pod endpoint access. Its output contains data only and can be
// passed to the recommendation engine without exposing HTTP dependencies.
type Collector struct {
	fetcher        *Fetcher
	maxConcurrency int
}

var _ ObservationCollector = (*Collector)(nil)

// NewDefaultCollector creates a Pod collector with the default HTTP settings.
func NewDefaultCollector() *Collector {
	return &Collector{fetcher: NewMetricsFetcher(), maxConcurrency: defaultCollectionConcurrency}
}

// NewCollector creates a Pod collector around an injected fetcher.
func NewCollector(fetcher *Fetcher) *Collector {
	return &Collector{fetcher: fetcher, maxConcurrency: defaultCollectionConcurrency}
}

type podMetricResult struct {
	value float64
	err   error
}

// Observe collects one metric from every eligible Pod. Pods that are
// terminating, not running, not ready, or missing an IP are skipped. Individual
// request failures are tolerated when at least one Pod returns a valid value.
func (collector *Collector) Observe(ctx context.Context,
	query PodMetricQuery) (autoscalingtypes.MetricObservation, error) {
	if collector == nil || collector.fetcher == nil {
		return autoscalingtypes.MetricObservation{}, fmt.Errorf("metric fetcher is not configured")
	}
	results := make([]podMetricResult, len(query.Pods))
	readyPodIndexes := make([]int, 0, len(query.Pods))
	for i := range query.Pods {
		if ready(&query.Pods[i]) {
			readyPodIndexes = append(readyPodIndexes, i)
		}
	}

	workerCount := collector.maxConcurrency
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(readyPodIndexes) {
		workerCount = len(readyPodIndexes)
	}
	jobs := make(chan int)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer workers.Done()
			for podIndex := range jobs {
				results[podIndex].value, results[podIndex].err =
					collector.fetcher.FetchMetrics(ctx, query.Pods[podIndex], query.Source)
			}
		}()
	}
	for _, podIndex := range readyPodIndexes {
		jobs <- podIndex
	}
	close(jobs)
	workers.Wait()

	values := make([]float64, 0, len(readyPodIndexes))
	errors := make([]string, 0)
	skipped := len(query.Pods) - len(readyPodIndexes)
	for i, pod := range query.Pods {
		if !ready(&pod) {
			continue
		}
		if results[i].err != nil {
			errors = append(errors, fmt.Sprintf("%s/%s: %v", pod.Namespace, pod.Name, results[i].err))
			continue
		}
		values = append(values, results[i].value)
	}
	if len(values) == 0 {
		if len(errors) == 0 {
			return autoscalingtypes.MetricObservation{},
				fmt.Errorf("no ready Pods available for metric %q", query.Source.TargetName)
		}
		return autoscalingtypes.MetricObservation{},
			fmt.Errorf("all Pod metric requests failed: %s", strings.Join(errors, "; "))
	}
	hwlog.RunLog.Infof("collected Pod metric namespace=%s target=%s metric=%s podValues=%v successfulPods=%d skippedPods=%d failedPods=%d",
		query.Namespace, query.TargetName, query.Source.TargetName, values, len(values), skipped, len(errors))
	if len(errors) > 0 {
		hwlog.RunLog.Warnf("partially failed to collect Pod metric namespace=%s target=%s metric=%s errors=%s",
			query.Namespace, query.TargetName, query.Source.TargetName, strings.Join(errors, "; "))
	}
	return autoscalingtypes.MetricObservation{
		Name: query.Source.TargetName, Values: values, Timestamp: query.Timestamp}, nil
}

func ready(pod *corev1.Pod) bool {
	if pod.DeletionTimestamp != nil || pod.Status.Phase != corev1.PodRunning || pod.Status.PodIP == "" {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
