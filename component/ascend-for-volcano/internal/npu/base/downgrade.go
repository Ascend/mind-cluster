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
Package base is using for HuaWei Ascend affinity schedule base.
*/
package base

import (
	"strconv"
	"time"

	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// IsDowngradeEnabled reports the huawei.com/scheduler.downgrade annotation,
// the feature is off by default.
func (tp *NPUHandler) IsDowngradeEnabled() bool {
	if tp == nil {
		return false
	}
	return tp.Annotation[util.SchedulerDowngradeAnnoKey] == "true"
}

// DowngradeConstraintIfTimeout is the common entry of the constraint downgrade, run
// by the policy's DowngradeConstraint hook after the job validation. It keeps the
// settled constraint while the round holds tasks scheduled under it, otherwise it
// settles the constraint of the level the wait of the session buys. The level counts
// the whole windows waited at the session start, so every invocation of one session
// settles the same constraint, a session never steps twice and a long wait crosses
// several steps at once. A policy plugs in with its own ladder: nextConfig maps the
// number of waited windows to the constraint of that level and to the level the
// constraint settled at (false when the policy holds no ladder for it, the bottom
// line included), applyConfig applies it and reports whether it took.
func (tp *NPUHandler) DowngradeConstraintIfTimeout(jobID api.JobID,
	nextConfig func(windows int) (string, int, bool), applyConfig func(config string) bool) {
	if tp == nil || jobID == "" || nextConfig == nil || applyConfig == nil || !tp.IsDowngradeEnabled() {
		return
	}
	tp.restoreDowngradedConstraint(jobID, applyConfig)
	tp.downgradeConstraintOnTimeout(jobID, nextConfig, applyConfig)
}

// restoreDowngradedConstraint re-applies the settled constraint while the round
// holds tasks scheduled under it: the tasks still waiting have to follow the
// topology of the ones already scheduled, rank arithmetic reads one constraint. A
// constraint the policy refuses is ignored, the level is never advanced here.
func (tp *NPUHandler) restoreDowngradedConstraint(jobID api.JobID, applyConfig func(string) bool) {
	if downgrade.SessionAllWaiting(jobID) {
		return
	}
	config, effectTime, ok := downgrade.DowngradeState(jobID)
	if !ok || !applyConfig(config) {
		return
	}
	tp.markDowngraded(config, effectTime)
}

// downgradeConstraintOnTimeout settles the constraint of the level the wait of the
// session buys: the policy ladder is asked for that level in one call, the level and
// the constraint it maps to settle the session at once.
func (tp *NPUHandler) downgradeConstraintOnTimeout(jobID api.JobID,
	nextConfig func(windows int) (string, int, bool), applyConfig func(config string) bool) {
	if !downgrade.SessionAllWaiting(jobID) {
		return
	}
	// the configuration layer validates and defaults the window
	window := time.Duration(tp.FrameAttr.SchedulerDowngradeTimeout) * time.Second
	windows, ok := tp.waitedWindows(jobID, window)
	if !ok {
		return
	}
	config, level, ok := nextConfig(windows)
	if !ok {
		return
	}
	// the effect time is derived before the application: a refused constraint
	// leaves no half written state
	effectTime := tp.downgradeEffectTime(jobID, window, level)
	if effectTime <= 0 || !applyConfig(config) {
		return
	}
	klog.V(util.LogInfoLev).Infof("job <%s> waited %d windows, downgrade to <%s>", jobID, windows, config)
	downgrade.RecordDowngrade(jobID, config, effectTime)
	tp.markDowngraded(config, effectTime)
}

// waitedWindows is the number of whole downgrade windows the job waited at the
// session start, the level its wait buys: the n-th level waits n full windows.
func (tp *NPUHandler) waitedWindows(jobID api.JobID, window time.Duration) (int, bool) {
	elapsed, ok := cache.SessionElapsed(jobID)
	if !ok || elapsed <= 0 || window <= 0 {
		return 0, false
	}
	windows := int(elapsed / int64(window.Seconds()))
	return windows, windows > 0
}

// downgradeEffectTime is the end of the window which bought the level, zero when
// the job holds no clock.
func (tp *NPUHandler) downgradeEffectTime(jobID api.JobID, window time.Duration, level int) int64 {
	waitStart, ok := cache.WaitStartTime(jobID)
	if !ok {
		return 0
	}
	return waitStart + int64(window*time.Duration(level)/time.Second)
}

// markDowngraded copies the settled effect time and the constraint snapshot to the
// session mark, so a restore session annotates what the downgrade session wrote. The
// caller hands over the pair it settled, the store is never read back here.
func (tp *NPUHandler) markDowngraded(config string, effectTime int64) {
	tp.downgradedConfig, tp.downgradedEffectTime = config, effectTime
}

// setDowngradedToPod writes the downgrade markers (effect time and constraint
// snapshot) for the ops observation, the volcano frame attaches them at bind time.
func (tp *NPUHandler) setDowngradedToPod(task *api.TaskInfo) {
	if tp == nil || tp.downgradedEffectTime <= 0 || tp.downgradedConfig == "" ||
		task == nil || task.Pod == nil || task.Pod.Annotations == nil {
		return
	}
	if _, ok := task.Pod.Annotations[util.SchedulerDowngradedAnnoKey]; ok {
		return
	}
	task.Pod.Annotations[util.SchedulerDowngradedAnnoKey] =
		strconv.FormatInt(tp.downgradedEffectTime, util.Base10)
	task.Pod.Annotations[util.SchedulerDowngradedLevelAnnoKey] = tp.downgradedConfig
	klog.V(util.LogInfoLev).Infof("job <%s> scheduled with downgraded constraint %s, "+
		"mark pod <%s/%s> at %d", task.Job, tp.downgradedConfig,
		task.Pod.Namespace, task.Pod.Name, tp.downgradedEffectTime)
}
