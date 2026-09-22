/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package internal

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

// TestControllerRestoreAnnotation verifies the module-level Controller delegates
// the Running-restore to its inner NPU policy handlers (which embed
// base.NPUHandler and implement plugin.RunningRestoreHook), so a Running task's
// chips are re-claimed on the node free-top from the pod's own annotation while
// the pod annotation stays untouched.
func TestControllerRestoreAnnotation(t *testing.T) {
	node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       "node1",
			Annotation: map[string]string{util.NPU910CardName: "Ascend910-4,Ascend910-3"},
		},
	}

	c := New().(*Controller)
	c.PolicyHandler = append(c.PolicyHandler, &base.NPUHandler{
		SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{ReqNPUName: util.NPU910CardName}},
		MaxNodeNPUNum:    64,
	})

	// nil task -> nil
	if got := c.RestoreAnnotation(nil, node); got != nil {
		t.Errorf("RestoreAnnotation(nil task) = %v, want nil", got)
	}
	// node without annotations -> nil
	if got := c.RestoreAnnotation(test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-4"),
		plugin.NPUNode{}); got != nil {
		t.Errorf("RestoreAnnotation(no annotation node) = %v, want nil", got)
	}

	// a Running pod holding chip 4 is re-claimed on the node free-top
	rt := test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-4")
	got := c.RestoreAnnotation(rt, node)
	if got == nil {
		t.Fatal("RestoreAnnotation = nil, want *NPUNode")
	}
	if left := got.Annotation[util.NPU910CardName]; left != "Ascend910-3" {
		t.Errorf("RestoreAnnotation node free-top = %q, want %q", left, "Ascend910-3")
	}
	if anno := rt.Pod.Annotations[util.NPU910CardName]; anno != "Ascend910-4" {
		t.Errorf("RestoreAnnotation must not touch pod annotation, got %q", anno)
	}
}

// TestControllerRestoreAnnotationNoHandler verifies the Controller tolerates
// being asked to restore with no inner policy handler (never nil-derefs).
func TestControllerRestoreAnnotationNoHandler(t *testing.T) {
	c := New().(*Controller)
	if got := c.RestoreAnnotation(test.BuildTestTaskWithAnnotation(util.NPU910CardName, "1", "Ascend910-4"),
		plugin.NPUNode{CommonNode: plugin.CommonNode{Annotation: map[string]string{util.NPU910CardName: "Ascend910-4"}}}); got == nil {
		t.Error("RestoreAnnotation with no inner handler = nil, want *NPUNode")
	}
}
