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

/*
Package nslb is using for HuaWei Ascend pin tor affinity.
*/
package nslb

import (
	"container/heap"
	"strconv"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

func allocate(superPods []*superPodTors, vPodNum, vPodSize int) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	// init queue
	for _, c := range superPods {
		if c.remainFull+c.remainPart >= vPodSize {
			item := &item{
				spTors:   c,
				taskSize: vPodSize,
			}
			heap.Push(pq, item)
		}
	}
	// update each super pod currentGain and futureGain
	updateQueue(pq, vPodSize, vPodNum)

	// allocate resource one vPod by one vPod
	allocateResource(pq, vPodNum, vPodSize)
}

func initSuperPodTors(torCount int, id int32, tors []*plugin.Tor) *superPodTors {
	fullTors, partTors := getFullAndPartialTors(tors, torCount)
	return &superPodTors{
		name:       strconv.FormatInt(int64(id), util.NPUIndex10),
		superPodId: id,
		full:       len(fullTors) * torCount,
		partial:    0,
		remainFull: len(fullTors) * torCount,
		remainPart: 0,
		torCount:   torCount,
		fullTors:   fullTors,
		partialTors: [util.NPUIndex3][]*plugin.Tor{
			getNotShareAndFreeTorServer(partTors, descOrder),
			getSharedTorServer(partTors, descOrder),
			getNotShareTorServer(partTors, descOrder),
		},
	}
}

func getFullAndPartialTors(tors []*plugin.Tor, torCount int) ([]*plugin.Tor, []*plugin.Tor) {
	var fullTors []*plugin.Tor
	var partialTors []*plugin.Tor
	for _, tor := range tors {
		if tor.FreeServerCount == 0 {
			continue
		}
		if tor.FreeServerCount == torCount {
			fullTors = append(fullTors, tor)
			continue
		}
		partialTors = append(partialTors, tor)
	}
	return fullTors, partialTors
}

func allocateResource(pq *PriorityQueue, vPodNum, vPodSize int) {
	allocated := 0
	for allocated < vPodNum && pq.Len() > 0 {

		item, ok := heap.Pop(pq).(*item)
		if !ok {
			return
		}
		c := item.spTors

		allocFull := util.Min(c.remainFull, vPodSize)
		allocPart := util.Min(c.remainPart, vPodSize-allocFull)
		if allocFull+allocPart < vPodSize {
			continue
		}
		// update allocated info
		c.usedFull += allocFull
		c.usedPartial += allocPart
		c.remainFull -= allocFull
		c.remainPart -= allocPart
		allocated++
		klog.V(util.LogWarningLev).Infof("allocate to %v,allocate resource full: %v, part: %v",
			c.superPodId, allocFull, allocPart)
		// push back to pq
		if c.remainFull+c.remainPart >= vPodSize {
			heap.Push(pq, item)
		}
		// update pq current gain and future gain
		updateQueue(pq, vPodSize, vPodNum-allocated)
	}
}

func (pq PriorityQueue) Len() int { return len(pq.items) }

func (pq PriorityQueue) Less(i, j int) bool {
	a, b := pq.items[i], pq.items[j]
	// 1. compare current gains
	if a.currentGain != b.currentGain {
		return a.currentGain > b.currentGain
	}
	// 2. if current gains are equal, compare future gains
	if a.futureGain != b.futureGain {
		// 2.1. if task future gains is not equal, compare whether the task size is a multiple of the number of tors
		// 2.2. compare future gains
		return a.futureGain > b.futureGain
	}

	// 3. if future gains is equal and used full tor num is equal, use more use num
	if a.spTors.full-a.spTors.remainFull == b.spTors.full-b.spTors.remainFull {
		return a.spTors.full < b.spTors.full
	}
	// 4. if full used tor num is not equal, compare whether the number of used full tors is greater
	return a.spTors.full-a.spTors.remainFull > b.spTors.full-b.spTors.remainFull
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item, ok := x.(*item)
	if !ok {
		return
	}
	item.index = n
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old.items)
	item := old.items[n-1]
	item.index = -1
	pq.items = old.items[:n-1]
	return item
}

func calcCombinations(used, torCount int) int {
	return used / torCount
}

func calcFutureGain(c *superPodTors, vPodSize, remainVPodNum int) int {
	// cal remain max vPod num
	maxVPodNum := (c.remainFull + c.remainPart) / vPodSize
	if maxVPodNum > remainVPodNum {
		maxVPodNum = remainVPodNum
	}

	totalGain := 0
	tempFull := c.remainFull
	tempPart := c.remainPart
	currentUsed := c.usedFull + c.usedPartial

	// Simulate the allocation of logical super nodes one by one in the current super node
	for i := 0; i < maxVPodNum; i++ {
		allocFull := util.Min(tempFull, vPodSize)
		allocPart := vPodSize - allocFull
		if allocPart > tempPart {
			break
		}

		prevComb := calcCombinations(currentUsed, c.torCount)
		currentUsed += vPodSize
		newComb := calcCombinations(currentUsed, c.torCount)
		totalGain += newComb - prevComb

		tempFull -= allocFull
		tempPart -= allocPart
	}
	// finally calculate the number of Tors used as the future score.
	return totalGain
}

func updateQueue(pq *PriorityQueue, vPodSize, remainVPodNum int) {
	for i, item := range pq.items {
		c := item.spTors

		// update current Gain
		// can allocate full tor num, if remain full tor is not enough for vPod size, use remain full tor
		allocFull := util.Min(c.remainFull, vPodSize)
		// can allocate Part  tor num
		allocPart := util.Min(c.remainPart, vPodSize-allocFull)
		// calculate full tor num before allocate
		prevComb := calcCombinations(c.usedFull+c.usedPartial, c.torCount)
		// calculate full tor num after allocate
		newComb := calcCombinations(c.usedFull+allocFull+c.usedPartial+allocPart, c.torCount)
		// calculate current gain
		pq.items[i].currentGain = newComb - prevComb

		// update future Gain
		pq.items[i].futureGain = calcFutureGain(c, vPodSize, remainVPodNum)
	}
	heap.Init(pq)
}

func getTorsFreeServerNum(tors []*plugin.Tor) int {
	nodeNum := 0
	for _, tor := range tors {
		nodeNum += tor.FreeServerCount
	}
	return nodeNum
}
