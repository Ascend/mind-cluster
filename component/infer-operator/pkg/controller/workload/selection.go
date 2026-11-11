/*
Copyright(C) 2026. Huawei Technologies CO.,Ltd. All rights reserved.

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

package workload

import (
	"sort"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// selectionCandidate describes an instance when selecting which instances to delete.
// It is kept in memory only and never persisted.
type selectionCandidate struct {
	index int  // parsed instance index
	ready bool // readiness, from IsWorkLoadReady
}

// parseInstanceIndex parses the instance index from the object meta labels.
// It logs at debug level and returns ok=false when the label is missing or not an integer,
// in which case the caller must skip (keep) the object. The debug level keeps a long-lived
// abnormal object from flooding the log on every reconcile round.
func parseInstanceIndex(meta v1.ObjectMeta) (index int, ok bool) {
	instanceIndexStr, ok := meta.Labels[common.InstanceIndexLabelKey]
	if !ok {
		hwlog.RunLog.Debugf("workload<%s> has no instance index label, skip it", meta.Name)
		return 0, false
	}
	instanceIndex, err := strconv.Atoi(instanceIndexStr)
	if err != nil {
		hwlog.RunLog.Debugf("workload<%s> failed to convert instance index<%s> to int: %v",
			meta.Name, instanceIndexStr, err)
		return 0, false
	}
	return instanceIndex, true
}

// selectIndicesToDelete returns the indices of the instances to delete, in deletion order:
// illegal instances first, then unready ones, then ready ones, each group by descending index.
// It converts the listed workloads into selection candidates and delegates the decision to
// selectIndicesToDeleteFromCandidates, so the policy stays free of business objects. Workloads
// whose index label is missing or illegal are skipped here (they are always kept) but they are
// still present instances, which is why len(workloads) rather than len(candidates) is passed as
// the current count.
func selectIndicesToDelete(workloads []WorkLoadInterface, indexLimit int) []int {
	candidates := make([]selectionCandidate, 0, len(workloads))
	for _, workLoad := range workloads {
		index, ok := parseInstanceIndex(workLoad.GetWorkLoadObjMeta())
		if !ok {
			continue
		}
		candidates = append(candidates, selectionCandidate{index: index, ready: workLoad.IsWorkLoadReady()})
	}
	return selectIndicesToDeleteFromCandidates(candidates, len(workloads), indexLimit)
}

// Selection priorities of a candidate when choosing which instances to delete: a lower value is
// deleted first. Illegal instances (negative index) rank above unready ones, which rank above
// ready ones. Every selected candidate consumes one slot of the shrink quota.
const (
	selectionPriorityIllegal = iota
	selectionPriorityUnready
	selectionPriorityReady
)

// selectionPriority returns the deletion priority of a candidate.
func selectionPriority(candidate selectionCandidate) int {
	switch {
	case candidate.index < 0:
		return selectionPriorityIllegal
	case candidate.ready:
		return selectionPriorityReady
	default:
		return selectionPriorityUnready
	}
}

// selectIndicesToDeleteFromCandidates decides which instances to delete, in deletion order, from
// plain in-memory candidates. currentCount is the number of instances present, including the ones
// whose index could not be parsed: they are never deleted but still consume the shrink quota. The
// candidates are ordered by priority (illegal, unready, ready) and by descending index inside each
// priority, so the deletion set is simply the first (currentCount - indexLimit) candidates of that
// order. Nothing is deleted while the current count does not exceed indexLimit, not even illegal
// candidates. When every instance is ready this degrades to deleting the highest indices, which is
// the same behavior as before. The candidates slice is reordered, so callers must not rely on its
// order afterwards.
func selectIndicesToDeleteFromCandidates(candidates []selectionCandidate, currentCount, indexLimit int) []int {
	sort.Slice(candidates, func(i, j int) bool {
		priorityI, priorityJ := selectionPriority(candidates[i]), selectionPriority(candidates[j])
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}
		return candidates[i].index > candidates[j].index
	})

	shrinkCount := min(max(currentCount-indexLimit, 0), len(candidates))
	deletedIndices := make([]int, 0, shrinkCount)
	for _, candidate := range candidates[:shrinkCount] {
		deletedIndices = append(deletedIndices, candidate.index)
	}
	return deletedIndices
}

// SelectIndicesToCreate returns the indices of the instances to create, in ascending order.
// Instances are created only when the current count is below the target; the difference is the
// target minus the deduplicated current count. Gaps inside the existing index range are filled
// first (ascending) and the rest are appended after the largest existing index.
// currentIndices must contain the indices of Terminating instances: their index counts towards
// the current count and is never treated as a gap. Instances with a negative index are illegal
// (purged) objects, they are neither counted nor treated as a gap.
func SelectIndicesToCreate(currentIndices []int, target int) []int {
	existing := make(map[int]struct{}, len(currentIndices))
	for _, index := range currentIndices {
		if index < 0 {
			continue
		}
		existing[index] = struct{}{}
	}

	diff := target - len(existing)
	if diff <= 0 {
		// The count has reached the target, create nothing to preserve the shrink result.
		return nil
	}

	// Scan the index space from 0 upwards and take the first diff absent indices: the gaps inside
	// the existing range come first (ascending), then the tail. No upper bound is needed, because
	// every index above the largest existing one is absent, so the same loop appends them.
	toCreate := make([]int, 0, diff)
	for index := 0; len(toCreate) < diff; index++ {
		if _, ok := existing[index]; !ok {
			toCreate = append(toCreate, index)
		}
	}
	return toCreate
}
