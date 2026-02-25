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

// Package chip4nodex is using for HuaWei 300I A5 affinity schedule.
package chip4nodex

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	singleChip = 1 // require single card
	twoChips   = 2 // require two cards
	threeChips = 3 // require three cards
	fiveChips  = 5
	sixChips   = 6
	npuHexKilo = 1000
)

// createValidTask Construct a valid task
func createValidTask(name string) *api.TaskInfo {
	return &api.TaskInfo{
		Name: name,
	}
}

func TestNew(t *testing.T) {
	handler := New(SchedulePolicy4Px8)
	if handler.GetPluginName() != SchedulePolicy4Px8 {
		t.Errorf("expected plugin name '4p-8', got %v", handler.GetPluginName())
	}
	if handler.GetAnnoName(util.NPU910CardName) != util.NPU910CardName {
		t.Errorf("expected anno name '%v', got %v", util.NPU910CardName, handler.GetAnnoName(util.NPU910CardName))
	}
	if handler.GetAnnoPreVal(util.NPU910CardName) != util.NPU910CardNamePre {
		t.Errorf("expected anno pre value '%v', got %v", util.NPU910CardNamePre, handler.GetAnnoPreVal(util.NPU910CardName))
	}
}

func TestReleaseAnnotation(t *testing.T) {
	handler := New(SchedulePolicy4Px8)
	task := createValidTask("task1")
	node := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       "node1",
			Annotation: map[string]string{"dummy": "value"},
		},
	}
	newNode := handler.ReleaseAnnotation(task, node)
	if newNode == nil {
		t.Error("ReleaseAnnotation returned nil")
	}
	if newNode.Annotation["dummy"] != "value" {
		t.Error("ReleaseAnnotation modified node annotation unexpectedly")
	}
}

// makeTaskWithNPUCount Construct a TaskInfo containing Pod Annotations
func makeTaskWithNPUCount(name, count string) *api.TaskInfo {
	ti := &api.TaskInfo{
		Name: name,
		Pod:  &v1.Pod{},
	}
	ti.Pod.Annotations = map[string]string{
		util.NPU910CardName: count,
	}
	return ti
}

// makeNodeWithKChips Construct a node with k NPU Annotations
func makeNodeWithKChips(k int) plugin.NPUNode {
	anno := map[string]string{}
	for i := 0; i < k; i++ {
		key := util.NPU910CardNamePre + strconv.Itoa(i)
		anno[key] = "healthy"
	}
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       "node-x",
			Annotation: anno,
			Label:      map[string]string{util.AcceleratorType: util.NPU910CardName},
			Allocate: map[v1.ResourceName]float64{
				util.NPU910CardName: float64(k * npuHexKilo),
			},
		},
	}
}

// TestCheckNodeNPUByTaskInvalidArgs Parameter validation: an error should be returned if task is nil or node.Annotation is empty.
func TestCheckNodeNPUByTaskInvalidArgs(t *testing.T) {
	h := New("p")
	// task=nil
	if err := h.CheckNodeNPUByTask(nil, makeNodeWithKChips(singleChip)); err == nil {
		t.Error("expected error when task is nil")
	}
	// node.Annotation is null
	bad := plugin.NPUNode{CommonNode: plugin.CommonNode{Annotation: map[string]string{}}}
	if err := h.CheckNodeNPUByTask(&api.TaskInfo{}, bad); err == nil {
		t.Error("expected error when node.Annotation is empty")
	}
}

// TestCheckNodeNPUByTaskInsufficientTopology Insufficient topology: 2 NPUs requested, but the node has only 1 annotation.
// CheckNodeNPUByTask should report an error.
func TestCheckNodeNPUByTaskInsufficientTopology(t *testing.T) {
	h := New("p")
	task := makeTaskWithNPUCount("t2", "2")
	node := makeNodeWithKChips(singleChip)
	if err := h.CheckNodeNPUByTask(task, node); err == nil {
		t.Error("expected error when node has fewer NPU annotations than requested")
	}
}

// TestScoreBestNPUNodesInvalidArgs ScoreBestNPUNodes Parameter check
func TestScoreBestNPUNodesInvalidArgs(t *testing.T) {
	h := &chip4nodex{}
	if err := h.ScoreBestNPUNodes(nil, nil, nil); err == nil ||
		!strings.Contains(err.Error(), util.ArgumentError) {
		t.Errorf("invalid args → expected ArgumentError, got %v", err)
	}
}

// TestUseAnnotationInvalidArgs UseAnnotation Parameter check
func TestUseAnnotationInvalidArgs(t *testing.T) {
	h := &chip4nodex{}
	// nil task
	if out := h.UseAnnotation(nil, makeNodeWithKChips(singleChip)); out != nil {
		t.Errorf("nil task → expected nil, got %v", out)
	}
	// empty annotation
	t2 := makeTaskWithNPUCount("t2", "1")
	bad := plugin.NPUNode{CommonNode: plugin.CommonNode{Annotation: map[string]string{}}}
	if out := h.UseAnnotation(t2, bad); out != nil {
		t.Errorf("empty annotation → expected nil, got %v", out)
	}
}

// TestSelectNPUFromNodeError selectNPUFromNode Wrong path
func TestSelectNPUFromNodeError(t *testing.T) {
	h := &chip4nodex{
		NPUHandler: base.NPUHandler{
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{ReqNPUName: util.NPU910CardName}},
			MaxNodeNPUNum: 5,
		},
	}
	// Pod.Annotations is null → error
	task := &api.TaskInfo{Name: "t3", Pod: &v1.Pod{}}
	_, err := h.selectNPUFromNode(task, makeNodeWithKChips(twoChips))
	if err == nil {
		t.Error("empty Pod.Annotations → expected error, got nil")
	}
}

// makeTask returns a TaskInfo whose Pod.Annotations[util.NPU910CardName]
// is set to reqCount.
func makeTask(name string, reqCount int) *api.TaskInfo {
	return &api.TaskInfo{
		Name: name,
		UID:  api.TaskID(name + "-uid"),
		Job:  api.JobID(name + "-job"),
		Pod: &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					util.NPU910CardName: strconv.Itoa(reqCount)}}},
	}
}

// newHandler creates a chip4nodex and injects a util.NPUJob with
// SpBlockNPUNum=1 so that multi-card topology splits into k blocks.
func newHandler(pluginName string, task *api.TaskInfo, reqCount int) *chip4nodex {
	baseObj := New(pluginName)
	h, ok := baseObj.(*chip4nodex)
	if !ok {
		klog.Error("Type assertion failed: expected *chip4nodex, got ", reflect.TypeOf(baseObj))
		return &chip4nodex{}
	}
	if h.NPUHandler.Jobs == nil {
		h.NPUHandler.Jobs = make(map[api.JobID]plugin.SchedulerJob)
	}
	if h.NPUHandler.ScheduleEnv.ClusterCache.Jobs == nil {
		h.NPUHandler.ScheduleEnv.ClusterCache.Jobs = h.NPUHandler.Jobs
	}
	if h.NPUHandler.ScheduleEnv.ClusterCache.Nodes == nil {
		h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = make(map[string]plugin.NPUNode)
	}
	if h.Nodes == nil {
		h.Nodes = make(map[string]plugin.NPUNode)
	}
	nJob := &util.NPUJob{
		ReqNPUName:    util.NPU910CardName,
		ReqNPUNum:     reqCount,
		SpBlockNPUNum: 1,
		TpBlockNPUNum: util.LeastTpBlock,
		Tasks:         make(map[api.TaskID]util.NPUTask),
	}
	nJob.Tasks[task.UID] = util.NPUTask{
		ReqNPUName: util.NPU910CardName,
		ReqNPUNum:  reqCount,
	}
	h.SchedulerJobAttr.NPUJob = nJob
	h.NPUTaskNum = reqCount
	sj := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{Name: task.Job},
			NPUJob: nJob},
	}
	h.NPUHandler.Jobs[task.Job] = sj
	h.NPUHandler.ScheduleEnv.ClusterCache.Jobs[task.Job] = sj
	return h
}

func TestGetTaskReqNPUNumSuccess(t *testing.T) {
	task := makeTask("t1", threeChips)
	h := newHandler(SchedulePolicy4Px8, task, threeChips)
	got, err := h.GetTaskReqNPUNum(task)
	if err != nil || got != threeChips {
		t.Fatalf("want (3,nil); got (%d,%v)", got, err)
	}
}

func TestGetTaskReqNPUNumNoJob(t *testing.T) {
	task := makeTask("t2", singleChip)
	baseObj := New("p")
	h, ok := baseObj.(*chip4nodex) // no jobs injected
	if !ok {
		t.Fatalf("Type assertion failed: expected *chip4nodex, got %T", baseObj)
	}
	_, err := h.GetTaskReqNPUNum(task)
	if err == nil || !strings.Contains(err.Error(), "is not npu job") {
		t.Fatalf("want 'is not npu job'; got %v", err)
	}
}

func TestGetTaskReqNPUNumNoTask(t *testing.T) {
	task := makeTask("t3", twoChips)
	h := newHandler(SchedulePolicy4Px8, task, twoChips)
	// remove the registered UID to trigger “is not npu task”
	delete(h.NPUHandler.Jobs[task.Job].SchedulerJobAttr.NPUJob.Tasks, task.UID)
	_, err := h.GetTaskReqNPUNum(task)
	if err == nil || !strings.Contains(err.Error(), "is not npu task") {
		t.Fatalf("want 'is not npu task'; got %v", err)
	}
}

func TestValidNPUJobDoesNotPanic(t *testing.T) {
	task := makeTask("t8", singleChip)
	h := newHandler(SchedulePolicy4Px8, task, singleChip)
	_ = h.ValidNPUJob()
}

func prepareNode(k int) plugin.NPUNode {
	node := makeNodeWithKChips(k)
	node.Name = fmt.Sprintf("node-%d", k)
	totalDigits := 0
	for i := 0; i < k; i++ {
		totalDigits += len(strconv.Itoa(i))
	}
	estimateSize := totalDigits + (k - 1)
	var sb strings.Builder
	sb.Grow(estimateSize)
	for i := 0; i < k; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(i))
	}
	if node.Annotation == nil {
		node.Annotation = make(map[string]string, 1)
	}
	node.Annotation[util.NPU910CardName] = sb.String()
	return node
}

func TestCheckNodeNPUByTaskSuccess(t *testing.T) {
	task := makeTask("task-ok", twoChips)
	h := newHandler(SchedulePolicy4Px8, task, twoChips)
	node := prepareNode(twoChips)
	// Mapping of the two nodes registered in the handler
	h.Nodes = map[string]plugin.NPUNode{node.Name: node}
	h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = map[string]plugin.NPUNode{node.Name: node}
	if err := h.CheckNodeNPUByTask(task, node); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScoreBestNPUNodesSuccess(t *testing.T) {
	task := makeTask("task-score", 1)
	h := newHandler(SchedulePolicy4Px8, task, singleChip)
	nodeA := prepareNode(singleChip)
	nodeA.Name = "node-A"
	nodeB := prepareNode(singleChip)
	nodeB.Name = "node-B"
	h.Nodes = map[string]plugin.NPUNode{
		nodeA.Name: nodeA,
		nodeB.Name: nodeB,
	}
	h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = h.Nodes
	nodes := []*api.NodeInfo{
		{Name: nodeA.Name},
		{Name: nodeB.Name},
	}
	// Insert a dummy so that len(sMap) > 0
	scores := map[string]float64{"dummy": 0}
	if err := h.ScoreBestNPUNodes(task, nodes, scores); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ni := range nodes {
		if _, ok := scores[ni.Name]; !ok {
			t.Errorf("missing score for node %q", ni.Name)
		}
	}
}

func TestScoreBestNPUNodesSuccessNoMesh(t *testing.T) {
	task := makeTask("task-score", 1)
	h := newHandler(SchedulePolicy4Px8, task, fiveChips)
	nodeA := prepareNode(fiveChips)
	nodeA.Name = "node-A"
	nodeB := prepareNode(sixChips)
	nodeB.Name = "node-B"
	h.Nodes = map[string]plugin.NPUNode{
		nodeA.Name: nodeA,
		nodeB.Name: nodeB,
	}
	h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = h.Nodes
	nodes := []*api.NodeInfo{
		{Name: nodeA.Name},
		{Name: nodeB.Name},
	}
	// Insert a dummy so that len(sMap) > 0
	scores := map[string]float64{"dummy": 0}
	var expectedA float64 = 64
	var expectedB float64 = 56
	if err := h.ScoreBestNPUNodes(task, nodes, scores); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ni := range nodes {
		if _, ok := scores[ni.Name]; !ok {
			t.Errorf("missing score for node %q", ni.Name)
		}
	}
	if scores[nodeA.Name] != expectedA {
		t.Errorf("expect score for node is %f, not %f", expectedA, scores[nodeA.Name])
	}
	if scores[nodeB.Name] != expectedB {
		t.Errorf("expect score for node is %f, not %f", expectedB, scores[nodeB.Name])
	}
}

func TestSelectNPUFromNodeSuccess(t *testing.T) {
	task := makeTask("task-sel", threeChips)
	h := newHandler(SchedulePolicy4Px8, task, threeChips)
	node := prepareNode(threeChips)
	h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = map[string]plugin.NPUNode{node.Name: node}
	picked, err := h.selectNPUFromNode(task, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1, 2}
	for i, id := range want {
		if picked[i] != id {
			t.Errorf("picked[%d]=%d, want=%d", i, picked[i], id)
		}
	}
}

func TestUseAnnotationDefault(t *testing.T) {
	task := makeTask("task-use", twoChips)
	h := newHandler(SchedulePolicy4Px8, task, twoChips)
	node := prepareNode(twoChips)
	node.Name = "node-use"
	h.Nodes = map[string]plugin.NPUNode{node.Name: node}
	h.NPUHandler.ScheduleEnv.ClusterCache.Nodes = h.Nodes
	// Ensure that the Pod and its Annotations map are initialized
	task.Pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
	out := h.UseAnnotation(task, node)
	if out == nil {
		t.Fatal("UseAnnotation returned nil, expected non-nil")
	}
	// Check that the Pod's NPU annotation has been written
	annoKey := util.NPU910CardName
	if _, ok := task.Pod.Annotations[annoKey]; !ok {
		t.Errorf("pod annotation %q not found", annoKey)
	}
	// Check that the returned node.Annotation also includes the NPU annotation
	if _, ok := out.Annotation[annoKey]; !ok {
		t.Errorf("node annotation %q not found", annoKey)
	}
}
