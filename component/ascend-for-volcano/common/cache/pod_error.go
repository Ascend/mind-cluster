/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package cache

import (
	"sync"
	"time"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// podErrorCache is the process-wide pod-error backoff state, reached only
// through the package functions. A pod-error record is card-level on the node:
// once a card errored a pod, every job backs off from that card until the
// backoff window since the card's failure time has elapsed.
var podErrorCache = newPodErrorCache()

// podErrorStore is the state behind the package functions, guarded by a mutex
// because the fault handler and the node predicates of a task can run in
// parallel across goroutines.
type podErrorStore struct {
	mu sync.Mutex
	// node name -> card name -> fault unix second (first error)
	records map[string]map[string]int64
}

func newPodErrorCache() *podErrorStore {
	return &podErrorStore{
		records: map[string]map[string]int64{},
	}
}

// RecordPodError records the cards that errored a pod on the given node, at the
// pod's actual failure time. A later (larger) failure time overwrites an earlier
// one: re-recording the same failure (e.g. the fault handler seeing the same
// fault task across sessions) leaves the timestamp unchanged, while a new error
// on the same card refreshes it so the backoff window restarts.
func RecordPodError(nodeName string, cards []string, faultTime int64) {
	if nodeName == "" || len(cards) == 0 {
		return
	}
	if faultTime <= 0 {
		faultTime = time.Now().Unix()
	}
	podErrorCache.mu.Lock()
	defer podErrorCache.mu.Unlock()
	cardTime, ok := podErrorCache.records[nodeName]
	if !ok {
		cardTime = make(map[string]int64)
		podErrorCache.records[nodeName] = cardTime
	}
	for _, card := range cards {
		if card == "" {
			continue
		}
		if existing, ok := cardTime[card]; !ok || faultTime > existing {
			cardTime[card] = faultTime
		}
	}
	expirePodErrorLocked(time.Now().Unix())
}

// expirePodErrorLocked drops pod-error records older than the maximum tolerance
// window. It must be called with the mutex held.
func expirePodErrorLocked(now int64) {
	expired := now - util.MaxPodErrorToleranceAfterTime
	for nodeName, cardTime := range podErrorCache.records {
		for card, faultTime := range cardTime {
			if faultTime < expired {
				delete(cardTime, card)
			}
		}
		if len(cardTime) == 0 {
			delete(podErrorCache.records, nodeName)
		}
	}
}

// PodErrorCards returns the pod-error card names of the given node, mapped to
// their failure unix second. An empty map means the node has no pod-error record.
func PodErrorCards(nodeName string) map[string]int64 {
	out := make(map[string]int64)
	if nodeName == "" {
		return out
	}
	podErrorCache.mu.Lock()
	defer podErrorCache.mu.Unlock()
	expirePodErrorLocked(time.Now().Unix())
	cardTime, ok := podErrorCache.records[nodeName]
	if !ok {
		return out
	}
	for card, t := range cardTime {
		out[card] = t
	}
	return out
}

// ResetPodError clears all pod-error records. It is intended for tests, where
// the process-wide cache is shared across test functions.
func ResetPodError() {
	podErrorCache.mu.Lock()
	defer podErrorCache.mu.Unlock()
	podErrorCache.records = map[string]map[string]int64{}
}
