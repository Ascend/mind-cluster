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

package chip

import (
	"testing"
)

// TestPreemptOrReclaimSelectNetUnhealthy pins the parameter-plane network
// unhealthy tolerance on the selective eviction path (preemptOrReclaimSelect,
// used by Reclaimable and by the pre-v1.15 Preemptable): a tolerant job sees
// net-unhealthy chips as usable (zero eviction), a non-tolerant job abstains
// because none of the chips fit under its tolerance.
func TestPreemptOrReclaimSelectNetUnhealthy(t *testing.T) {
	node := newTestNode(allNetUnhealthyNode(8))

	tol := newTestHandler()
	tol.NPUJob.ParameterPlaneUnhealthyTolerance = true
	sel, ok := tol.preemptOrReclaimSelect(newPreemptorTask(2, ""), nil, node)
	if !ok || sel != nil {
		t.Fatalf("tolerant select on net-unhealthy: want (nil,true), got %v ok=%v", newNames(sel), ok)
	}

	dist := newTestHandler() // tolerance=false: distributed job without opt-in
	sel, ok = dist.preemptOrReclaimSelect(newPreemptorTask(2, ""), nil, node)
	if ok || sel != nil {
		t.Fatalf("non-tolerant select on net-unhealthy: want (nil,false), got %v ok=%v", newNames(sel), ok)
	}
}

// TestReclaimableNetUnhealthy reaches preemptOrReclaimSelect through Reclaimable,
// which is always routed selectively regardless of TopologyAwarePreemptActive.
func TestReclaimableNetUnhealthy(t *testing.T) {
	node := newTestNode(allNetUnhealthyNode(8))

	tp := newTestHandler()
	tp.NPUJob.ParameterPlaneUnhealthyTolerance = true
	sel, ok := tp.Reclaimable(newPreemptorTask(2, ""), nil, node)
	if !ok || sel != nil {
		t.Fatalf("tolerant reclaim on net-unhealthy: want (nil,true), got %v ok=%v", newNames(sel), ok)
	}

	dist := newTestHandler()
	sel, ok = dist.Reclaimable(newPreemptorTask(2, ""), nil, node)
	if ok || sel != nil {
		t.Fatalf("non-tolerant reclaim on net-unhealthy: want (nil,false), got %v ok=%v", newNames(sel), ok)
	}
}

// TestPreemptOrReclaimSelectNilGuards pins the defensive nil-args handling of the
// selective path in the untagged build, where the v1.15 aware-path tests are
// absent and the nil guards live purely in preemptOrReclaimSelect.
func TestPreemptOrReclaimSelectNilGuards(t *testing.T) {
	tp := newTestHandler()
	task := newPreemptorTask(2, "")
	node := newTestNode(allNetUnhealthyNode(8))

	var nilHandler *chipHandler
	if sel, ok := nilHandler.preemptOrReclaimSelect(task, nil, node); ok || sel != nil {
		t.Fatalf("nil handler: want (nil,false), got %v ok=%v", newNames(sel), ok)
	}
	if sel, ok := tp.preemptOrReclaimSelect(nil, nil, node); ok || sel != nil {
		t.Fatalf("nil preemptor: want (nil,false), got %v ok=%v", newNames(sel), ok)
	}
	if sel, ok := tp.preemptOrReclaimSelect(task, nil, nil); ok || sel != nil {
		t.Fatalf("nil node: want (nil,false), got %v ok=%v", newNames(sel), ok)
	}
}
