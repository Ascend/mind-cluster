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

/*
Package downgrade is using for the scheduler constraint downgrade cross-session
state.
*/
package downgrade

import (
	"encoding/json"
	"sync"

	"volcano.sh/volcano/pkg/scheduler/api"
)

// downgradeStore is the process-wide constraint downgrade state, reached only
// through the package functions. It owns no clock: the effect time of a record
// is derived by the consumer from the wait clock of common/cache.
var downgradeStore = newDowngradeCache()

// downgradeCache is the state behind the package functions, guarded by a mutex
// for the day volcano parallelizes more scheduling stages.
type downgradeCache struct {
	mu         sync.Mutex
	downgrades map[api.JobID]downgradeState
	// session is the guard snapshot stamped at the session start
	session map[api.JobID]sessionGuard
	// podEvictions holds the jobs whose placed tasks were partly evicted by a
	// pod-level reschedule: such a job keeps its settled constraint even when
	// the eviction leaves the whole scope waiting again
	podEvictions map[api.JobID]bool
}

// downgradeState is the downgrade progress of one job: the config is the
// constraint snapshot of the handler, the store never reads it. The level the
// config belongs to is not carried here: it is derived from the wait clock of the
// session, the store keeps only the settled snapshot and its effect time.
type downgradeState struct {
	config     string
	effectTime int64
	effective  bool
}

// sessionGuard is the per-job decision snapshot of one session: it separates the
// round which restores the settled constraint from the one which derives the level
// of the wait. The derived level is a pure function of the snapshot, so no mark of
// an already taken step is needed to keep a session at one level.
type sessionGuard struct {
	// allWaiting reports whether the whole scope was waiting at the session start
	allWaiting bool
	// window is the effective downgrade window seconds of the session: the job
	// level annotation wins over the global configuration, a non positive value
	// means the feature is off for the job this session
	window int
}

// newDowngradeCache builds an empty state.
func newDowngradeCache() *downgradeCache {
	return &downgradeCache{
		downgrades:   map[api.JobID]downgradeState{},
		podEvictions: map[api.JobID]bool{},
	}
}

// RecordDowngrade records the settled downgrade snapshot and its effect time. The
// config is the handler's own bytes, only an empty one is refused.
func RecordDowngrade(jobID api.JobID, config string, effectTime int64) {
	if jobID == "" || config == "" {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	downgradeStore.downgrades[jobID] = downgradeState{config: config, effectTime: effectTime}
}

// SeedDowngrade fills the state of a job from its pod markers after the in-memory
// state is gone, taking only a JSON object: the annotations are the trust boundary.
func SeedDowngrade(jobID api.JobID, config string, effectTime int64) {
	if jobID == "" || effectTime <= 0 || !isJSONObject(config) {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	if _, ok := downgradeStore.downgrades[jobID]; ok {
		return
	}
	downgradeStore.downgrades[jobID] = downgradeState{config: config, effectTime: effectTime, effective: true}
}

// isJSONObject reports whether the config is a non-null JSON object, the only
// shape the store accepts from the pod markers.
func isJSONObject(config string) bool {
	var obj map[string]json.RawMessage
	return json.Unmarshal([]byte(config), &obj) == nil && obj != nil
}

// MarkEffective freezes the current constraint snapshot once a task is
// scheduled under it. A task placed again also ends the pod eviction guard:
// the constraint took effect on the new round, the guard has done its job.
func MarkEffective(jobID api.JobID) {
	if jobID == "" {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	delete(downgradeStore.podEvictions, jobID)
	if state, ok := downgradeStore.downgrades[jobID]; ok && !state.effective {
		state.effective = true
		downgradeStore.downgrades[jobID] = state
	}
}

// MarkPodEviction guards the constraint of a job whose placed tasks were partly
// evicted: the pod-level reschedule keeps the settled constraint, so the round
// which sees the whole scope waiting again after the eviction must not mistake
// it for a job-level reschedule.
func MarkPodEviction(jobID api.JobID) {
	if jobID == "" {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	downgradeStore.podEvictions[jobID] = true
}

// ClearPodEviction drops the guard of a job whose whole scope was evicted at
// once, the shape of a job-level reschedule: the next all-waiting round resets.
func ClearPodEviction(jobID api.JobID) {
	if jobID == "" {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	delete(downgradeStore.podEvictions, jobID)
}

// PodEvictionProtected reports whether a pod-level eviction guards the settled
// constraint of the job, the job-level reset steps over such a job.
func PodEvictionProtected(jobID api.JobID) bool {
	if jobID == "" {
		return false
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	return downgradeStore.podEvictions[jobID]
}

// ResetConstraint drops the constraint of a job rescheduled at job level. The
// wait clock of common/cache is restarted by the caller for the new round.
func ResetConstraint(jobID api.JobID) {
	if jobID == "" {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	delete(downgradeStore.downgrades, jobID)
}

// DowngradeState returns both halves of the settled downgrade of a job in one
// read. The two are written together, so a caller which needs both takes them
// from one version of the state instead of two. False means the job has never
// been downgraded.
func DowngradeState(jobID api.JobID) (string, int64, bool) {
	if jobID == "" {
		return "", 0, false
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	state, ok := downgradeStore.downgrades[jobID]
	return state.config, state.effectTime, ok
}

// DowngradeConfig returns the current downgraded constraint snapshot of a job,
// opaque to the store, false means the job has never been downgraded.
func DowngradeConfig(jobID api.JobID) (string, bool) {
	config, _, ok := DowngradeState(jobID)
	return config, ok
}

// DowngradeEffectTime returns the effect time of the last downgrade of a job,
// the caller derives the pod marker annotation from it.
func DowngradeEffectTime(jobID api.JobID) (int64, bool) {
	_, effectTime, ok := DowngradeState(jobID)
	return effectTime, ok
}

// IsEffective returns true when a task completed scheduling under the constraint,
// such a job never deepens and only resets on a job-level reschedule.
func IsEffective(jobID api.JobID) bool {
	if jobID == "" {
		return false
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	state, ok := downgradeStore.downgrades[jobID]
	return ok && state.effective
}

// HasState returns true when the store holds a downgrade record of the job, it
// gates the lazy read of the pod marker annotations.
func HasState(jobID api.JobID) bool {
	if jobID == "" {
		return false
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	_, ok := downgradeStore.downgrades[jobID]
	return ok
}

// StampSession rebuilds the guard snapshot once per session from the view the
// caller holds. The snapshot is the only freeze point of a session: the
// level is a pure function of it. Only the downgrade enabled jobs carry a
// guard, the disabled ones read a zero window and every downgrade path of
// them short circuits.
func StampSession(allWaitingOf map[api.JobID]bool, windows map[api.JobID]int) {
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	downgradeStore.session = make(map[api.JobID]sessionGuard, len(allWaitingOf))
	for jobID, allWaiting := range allWaitingOf {
		downgradeStore.session[jobID] = sessionGuard{allWaiting: allWaiting, window: windows[jobID]}
	}
}

// SessionAllWaiting returns whether every in-scope task of the job was waiting at
// the session start: the round derives the level of the wait only then, otherwise
// it keeps the settled constraint of the tasks already scheduled.
func SessionAllWaiting(jobID api.JobID) bool {
	if jobID == "" {
		return false
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	return downgradeStore.session[jobID].allWaiting
}

// SessionWindow returns the effective downgrade window seconds of the job of
// the session: a non positive value means the feature is off, the entry of
// the downgrade path and the eager freeze check it alike.
func SessionWindow(jobID api.JobID) int {
	if jobID == "" {
		return -1
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	return downgradeStore.session[jobID].window
}

// Sweep removes the records of the jobs which are no longer alive. The key
// snapshot and the removal hold the store lock, the alive predicate runs between
// the two windows outside it: it walks the job table of the schedule env and must
// never hold this lock.
func Sweep(alive func(api.JobID) bool) {
	downgradeStore.mu.Lock()
	jobIDs := make([]api.JobID, 0, len(downgradeStore.downgrades))
	for jobID := range downgradeStore.downgrades {
		jobIDs = append(jobIDs, jobID)
	}
	downgradeStore.mu.Unlock()

	var dead []api.JobID
	for _, jobID := range jobIDs {
		if !alive(jobID) {
			dead = append(dead, jobID)
		}
	}
	if len(dead) == 0 {
		return
	}
	downgradeStore.mu.Lock()
	defer downgradeStore.mu.Unlock()
	for _, jobID := range dead {
		delete(downgradeStore.downgrades, jobID)
		delete(downgradeStore.podEvictions, jobID)
	}
}

// ResetDowngradeCacheForTest drops every record of the downgrade state, unit
// tests only.
func ResetDowngradeCacheForTest() {
	downgradeStore = newDowngradeCache()
}
