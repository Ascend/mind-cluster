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
Package inferservice provides the task-to-task affinity scheduling logic shared by
SuperPod policy handlers: jobs carrying the same inferServiceID label are softly
pinned to the super pods that already host the service.
*/
package inferservice

import (
	"container/heap"
	"fmt"
	"strconv"

	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// IDLabelKey is the job label key carrying the infer service id.
	IDLabelKey = "inferserviceid"
	// GroupSameSP marks super pods already hosting the same infer service.
	GroupSameSP = 1
	// GroupOtherSP marks super pods without the infer service.
	GroupOtherSP = 2
)

// SPInfo records one super pod that already hosts the same infer service.
type SPInfo struct {
	SuperPodID  int32
	FreeNodeNum int
}

// PQItem is one candidate super pod in the priority queue.
type PQItem struct {
	SuperPodID int32
	FreeNodes  int
	Group      int
	index      int
}

// PQ is the two-level priority queue: same-service super pods first, then others.
type PQ []*PQItem

// Len returns the number of items in the priority queue.
func (pq PQ) Len() int { return len(pq) }

// Less reports whether the item at index i should sort before the one at index j:
// lower group first, and within the same group more free nodes first.
func (pq PQ) Less(i, j int) bool {
	a, b := pq[i], pq[j]
	if a.Group != b.Group {
		return a.Group < b.Group
	}
	return a.FreeNodes > b.FreeNodes
}

// Swap swaps the items at index i and j, keeping their heap indexes up to date.
func (pq PQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds x to the priority queue. Non-PQItem items are ignored.
func (pq *PQ) Push(x interface{}) {
	item, ok := x.(*PQItem)
	if !ok {
		return
	}
	item.index = len(*pq)
	*pq = append(*pq, item)
}

// Pop removes and returns the last item of the priority queue,
// or nil when the queue is empty.
func (pq *PQ) Pop() interface{} {
	old := *pq
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	item.index = -1
	*pq = old[:n-1]
	return item
}

// GetInferServiceID returns the infer service id carried in the job label,
// or "" when the label is absent or the key is missing.
func GetInferServiceID(label map[string]string) string {
	if label == nil {
		return ""
	}
	return label[IDLabelKey]
}

// CollectScheduledSPs collects the super pods that already have jobs with the same
// inferServiceID scheduled, excluding the current job itself.
func CollectScheduledSPs(jobs map[api.JobID]plugin.SchedulerJob, currentJob api.JobID,
	inferServiceID string) map[int32]*SPInfo {
	sameSPs := make(map[int32]*SPInfo)
	if inferServiceID == "" || jobs == nil {
		return sameSPs
	}
	for jobID, job := range jobs {
		if job.Label == nil {
			continue
		}
		jobInferID, ok := job.Label[IDLabelKey]
		if !ok || jobInferID != inferServiceID {
			continue
		}
		if jobID == currentJob {
			continue
		}
		if len(job.SuperPods) == 0 {
			continue
		}
		for _, spNodes := range job.SuperPods {
			for _, sn := range spNodes {
				if _, exist := sameSPs[sn.SuperPodID]; exist {
					continue
				}
				sameSPs[sn.SuperPodID] = &SPInfo{SuperPodID: sn.SuperPodID}
			}
		}
	}
	return sameSPs
}

// EnrichSPInfo fills the FreeNodeNum of each scheduled super pod from the current
// super pod topology.
func EnrichSPInfo(superPodTop map[int32]plugin.SuperPod, sameSPs map[int32]*SPInfo) {
	for spID, sp := range superPodTop {
		if info, ok := sameSPs[spID]; ok {
			info.FreeNodeNum = len(sp)
		}
	}
}

// InferServiceReq records the request of selecting nodes for one infer service job.
type InferServiceReq struct {
	// Jobs holds all jobs in the current scheduling environment.
	Jobs map[api.JobID]plugin.SchedulerJob
	// JobName is the id of the job being scheduled.
	JobName api.JobID
	// InferServiceID is the infer service id carried in the job label.
	InferServiceID string
	// SpBlock is the node number of one sp-block.
	SpBlock int
	// ReqNPUNum is the total NPU number required by the job.
	ReqNPUNum int
	// SpBlockNPUNum is the NPU number of one sp-block.
	SpBlockNPUNum int
	// SuperPodTop is the current super pod topology keyed by super pod id.
	SuperPodTop map[int32]plugin.SuperPod
}

// SelectNodesForInferService selects nodes for an infer service job with
// same-super-pod soft affinity: super pods already hosting the same inferServiceID
// are preferred, others are fallbacks. It falls back to other super pods when the
// preferred one lacks resources.
func SelectNodesForInferService(req InferServiceReq) (map[string][]plugin.SuperNode, error) {
	if req.SpBlock <= 0 {
		return nil, fmt.Errorf("invalid spBlock %d for infer service job", req.SpBlock)
	}
	if req.SpBlockNPUNum <= 0 {
		return nil, fmt.Errorf("invalid spBlockNPUNum %d for infer service job", req.SpBlockNPUNum)
	}
	sameSPs := CollectScheduledSPs(req.Jobs, req.JobName, req.InferServiceID)
	EnrichSPInfo(req.SuperPodTop, sameSPs)
	spBlockCount := req.ReqNPUNum / req.SpBlockNPUNum
	selectedNodes := make(map[string][]plugin.SuperNode)
	pq := buildPriorityQueue(req.SuperPodTop, sameSPs, req.SpBlock)
	for i := 0; i < spBlockCount; i++ {
		item := popValidSP(pq, req.SuperPodTop, req.SpBlock)
		if item == nil {
			break
		}
		selectNodesFromSP(req.SuperPodTop[item.SuperPodID], i, req.SpBlock, selectedNodes)
		sameSPs[item.SuperPodID] = &SPInfo{SuperPodID: item.SuperPodID}
		EnrichSPInfo(req.SuperPodTop, sameSPs)
		pq = buildPriorityQueue(req.SuperPodTop, sameSPs, req.SpBlock)
	}
	if len(selectedNodes) < spBlockCount {
		return nil, fmt.Errorf("infer service schedule failed, required %d sp-block, got %d",
			spBlockCount, len(selectedNodes))
	}
	klog.V(util.LogInfoLev).Infof("infer service schedule success, job %s, inferServiceID %s, selectedNodes %v",
		req.JobName, req.InferServiceID, selectedNodes)
	return selectedNodes, nil
}

// popValidSP pops the highest priority super pod that still has enough free nodes
// for one sp-block. Returns nil when no super pod in the queue satisfies it.
func popValidSP(pq *PQ, superPodTop map[int32]plugin.SuperPod, spBlock int) *PQItem {
	for pq.Len() > 0 {
		item, ok := heap.Pop(pq).(*PQItem)
		if !ok {
			continue
		}
		sp, exist := superPodTop[item.SuperPodID]
		if !exist || len(sp) < spBlock {
			continue
		}
		return item
	}
	return nil
}

// selectNodesFromSP picks up to spBlock nodes from the given super pod and records
// them under the spIndex key. Selected nodes are removed from the super pod topology.
func selectNodesFromSP(sp plugin.SuperPod, spIndex, spBlock int, selectedNodes map[string][]plugin.SuperNode) {
	spIndexKey := strconv.Itoa(spIndex)
	selectedNodes[spIndexKey] = make([]plugin.SuperNode, 0, spBlock)
	nodeCount := 0
	for nodeName, nNode := range sp {
		if nodeCount >= spBlock {
			break
		}
		selectedNodes[spIndexKey] = append(selectedNodes[spIndexKey], plugin.SuperNode{
			Name:       nodeName,
			SuperPodID: nNode.SuperPodID,
		})
		delete(sp, nodeName)
		nodeCount++
	}
}

// buildPriorityQueue builds a two-level priority queue: same-super-pod items
// (Group=GroupSameSP, already hosting the same inferServiceID) come first,
// other-super-pod items (Group=GroupOtherSP) come second. Within the same group,
// items with more free nodes come first.
func buildPriorityQueue(superPodTop map[int32]plugin.SuperPod, sameSPs map[int32]*SPInfo,
	spBlock int) *PQ {
	pq := make(PQ, 0)
	heap.Init(&pq)
	for spID, info := range sameSPs {
		sp, ok := superPodTop[spID]
		if !ok || len(sp) < spBlock {
			continue
		}
		heap.Push(&pq, &PQItem{
			SuperPodID: spID,
			FreeNodes:  info.FreeNodeNum,
			Group:      GroupSameSP,
		})
	}
	for spID, sp := range superPodTop {
		if _, ok := sameSPs[spID]; ok {
			continue
		}
		if len(sp) < spBlock {
			continue
		}
		heap.Push(&pq, &PQItem{
			SuperPodID: spID,
			FreeNodes:  len(sp),
			Group:      GroupOtherSP,
		})
	}
	return &pq
}
