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

package podautoscaler

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingcontext "infer-operator/pkg/autoscaling/context"
)

type recommendation struct {
	at       time.Time
	replicas int32
}

type runtimeState struct {
	evaluation      *autoscalingcontext.EvaluationContext
	recommendations []recommendation
	configSignature string
}

type runtimeStateStore struct {
	mu     sync.Mutex
	states map[types.NamespacedName]*runtimeState
}

func newRuntimeStateStore() *runtimeStateStore {
	return &runtimeStateStore{states: make(map[types.NamespacedName]*runtimeState)}
}

func (store *runtimeStateStore) getOrCreate(pa *apiv1.PodAutoscaler) (*runtimeState, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	key := types.NamespacedName{Namespace: pa.Namespace, Name: pa.Name}
	state := store.states[key]
	if state == nil {
		state = &runtimeState{evaluation: autoscalingcontext.NewEvaluationContext()}
		store.states[key] = state
	}
	if err := state.evaluation.Refresh(pa); err != nil {
		return nil, err
	}
	return state, nil
}

func (store *runtimeStateStore) delete(key types.NamespacedName) {
	store.mu.Lock()
	delete(store.states, key)
	store.mu.Unlock()
}

func (store *runtimeStateStore) updateConfigSignature(state *runtimeState, signature string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	if state.configSignature == signature {
		return false
	}
	state.configSignature = signature
	return true
}

// stabilize applies the configured cooldown window. Scale-up chooses the
// lowest recent recommendation, while scale-down chooses the highest.
func (store *runtimeStateStore) stabilize(state *runtimeState,
	current, desired int32, now time.Time) int32 {
	store.mu.Lock()
	defer store.mu.Unlock()
	window := state.evaluation.Policy.ScaleDownCooldownWindow
	if desired > current {
		window = state.evaluation.Policy.ScaleUpCooldownWindow
	}
	// A zero cooldown disables stabilization. Do not retain prior
	// recommendations, otherwise an old value can block scaling forever.
	if window <= 0 {
		state.recommendations = nil
		return desired
	}
	state.recommendations = append(state.recommendations, recommendation{at: now, replicas: desired})
	cutoff := now.Add(-window)
	kept := state.recommendations[:0]
	for _, entry := range state.recommendations {
		if !entry.at.Before(cutoff) {
			kept = append(kept, entry)
		}
	}
	state.recommendations = kept
	result := desired
	for _, entry := range kept {
		if desired > current && entry.replicas < result {
			result = entry.replicas
		}
		if desired < current && entry.replicas > result {
			result = entry.replicas
		}
	}
	return result
}
