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
	"fmt"
	"strings"
	"sync"
	"time"

	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

type windows struct {
	stable *autoscalingtypes.TimeWindow
	panic  *autoscalingtypes.TimeWindow
}

// MetricsClient stores stable and panic metric windows in process memory.
type MetricsClient struct {
	mu          sync.RWMutex
	granularity time.Duration
	windows     map[string]*windows
}

// NewMetricsClient creates an empty metrics store with the specified bucket granularity.
func NewMetricsClient(granularity time.Duration) *MetricsClient {
	return &MetricsClient{granularity: granularity, windows: make(map[string]*windows)}
}

// UpdateMetrics averages the current Pod values and records that average in
// both windows. Averaging here preserves the per-Pod semantics expected by KPA.
func (client *MetricsClient) UpdateMetrics(now time.Time, key autoscalingtypes.MetricKey,
	stableDuration, panicDuration time.Duration, values ...float64) error {
	if len(values) == 0 {
		return fmt.Errorf("metric values are empty")
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	entry := client.windows[key.String()]
	if entry == nil {
		entry = &windows{stable: autoscalingtypes.NewTimeWindow(stableDuration, client.granularity),
			panic: autoscalingtypes.NewTimeWindow(panicDuration, client.granularity)}
		client.windows[key.String()] = entry
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	average := sum / float64(len(values))
	entry.stable.Record(now, average)
	entry.panic.Record(now, average)
	return nil
}

// GetMetricValues returns the current stable-window and panic-window averages.
func (client *MetricsClient) GetMetricValues(key autoscalingtypes.MetricKey) (float64, float64, error) {
	client.mu.RLock()
	defer client.mu.RUnlock()
	entry := client.windows[key.String()]
	if entry == nil {
		return 0, 0, fmt.Errorf("metric window %q does not exist", key.String())
	}

	return entry.stable.Avg(), entry.panic.Avg(), nil
}

// DeleteForPodAutoscaler removes all metric windows owned by one PodAutoscaler.
func (client *MetricsClient) DeleteForPodAutoscaler(namespace, name string) {
	client.mu.Lock()
	defer client.mu.Unlock()
	prefix := namespace + "/" + name + "/"

	for key := range client.windows {
		if strings.HasPrefix(key, prefix) {
			delete(client.windows, key)
		}
	}
}
