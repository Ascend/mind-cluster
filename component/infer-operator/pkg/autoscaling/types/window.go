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

package types

import (
	"sort"
	"sync"
	"time"
)

// TimeWindow stores one averaged value per granularity bucket.
type TimeWindow struct {
	mu          sync.RWMutex
	duration    time.Duration
	granularity time.Duration
	buckets     map[int64]float64
	values      []float64
}

// NewTimeWindow creates a duration-bounded metric window.
func NewTimeWindow(duration, granularity time.Duration) *TimeWindow {
	return &TimeWindow{duration: duration, granularity: granularity, buckets: make(map[int64]float64)}
}

// Record stores or replaces the value in its granularity bucket and removes expired buckets.
func (window *TimeWindow) Record(timestamp time.Time, value float64) {
	window.mu.Lock()
	defer window.mu.Unlock()
	bucket := timestamp.UnixNano() / int64(window.granularity)
	window.buckets[bucket] = value
	cutoff := timestamp.Add(-window.duration).UnixNano() / int64(window.granularity)
	for key := range window.buckets {
		if key < cutoff {
			delete(window.buckets, key)
		}
	}
	keys := make([]int64, 0, len(window.buckets))
	for key := range window.buckets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	window.values = window.values[:0]
	for _, key := range keys {
		window.values = append(window.values, window.buckets[key])
	}
}

// Avg returns the arithmetic mean of all retained bucket values.
func (window *TimeWindow) Avg() float64 {
	window.mu.RLock()
	defer window.mu.RUnlock()
	if len(window.values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range window.values {
		sum += value
	}
	return sum / float64(len(window.values))
}

// Size returns the number of retained buckets.
func (window *TimeWindow) Size() int {
	window.mu.RLock()
	defer window.mu.RUnlock()
	return len(window.values)
}
