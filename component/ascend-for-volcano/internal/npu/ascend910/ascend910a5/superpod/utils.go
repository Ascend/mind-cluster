/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package superpod for utils function
package superpod

import (
	"errors"
	"sort"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// transferSuperPodToRackIdMap get a map with rackID key, and get all nodes by rackID in this superPod
func transferSuperPodToRackIdMap(superPod map[string]nodeBaseInfo) map[int32][]nodeBaseInfo {
	superPodWithRack := make(map[int32][]nodeBaseInfo)

	for _, nNode := range superPod {
		_, ok := superPodWithRack[nNode.rackID]
		if !ok {
			superPodWithRack[nNode.rackID] = make([]nodeBaseInfo, 0)
		}
		superPodWithRack[nNode.rackID] = append(superPodWithRack[nNode.rackID], nNode)
	}

	klog.V(util.LogInfoLev).Infof("the superpod with rackID info is %v", superPodWithRack)
	return superPodWithRack
}

// the param superPodWithRack contain all rackID and nodes slice with the same rackID
// return value is a slice contains many of rackIDs sorted by the len of rack nodes ascending order.
func sortRackIdByLengthInOneSuperPod(superPodWithRack map[int32][]nodeBaseInfo) []int32 {
	var rackIdOrder []int32

	for rackId := range superPodWithRack {
		rackIdOrder = append(rackIdOrder, rackId)
	}
	sort.Slice(rackIdOrder, func(i, j int) bool {
		iLen := len(superPodWithRack[rackIdOrder[i]])
		jLen := len(superPodWithRack[rackIdOrder[j]])
		return iLen < jLen
	})
	return rackIdOrder
}

// the param superPodWithRack contain all rackID and nodes slice with the same rackID
// return value is a map which key is 1-8(=n) means how many nodes are available
// and which value is a slice contains all rackID whose nodes len is equal to n
func getRackLengthInOneSuperPod(superPodWithRack map[int32][]nodeBaseInfo) map[int][]int32 {
	// key is 1-8,value is a slice contain all rackID
	nodesLenWithRackIDMap := make(map[int][]int32, rackNodeNum)

	for rackId, nodes := range superPodWithRack {
		if nodesLenWithRackIDMap[len(nodes)] == nil {
			nodesLenWithRackIDMap[len(nodes)] = make([]int32, 0)
		}
		nodesLenWithRackIDMap[len(nodes)] = append(nodesLenWithRackIDMap[len(nodes)], rackId)
	}
	return nodesLenWithRackIDMap
}

// the input length is in range [0,8], contains value in range [0,63]
// eg: input1 []int{0,1,2,3,4,5,6,7} means 8 cards can be used
// eg: input2 []int{8,9,10,11,15} means 5 cards can be used
func getUsableNPUIndex(input []int) [nodeNPUNum]bool {
	res := [nodeNPUNum]bool{}
	for _, num := range input {
		if num < rackNodeNum*nodeNPUNum {
			res[num%nodeNPUNum] = true
		}
	}
	return res
}

func getPositionInTop(podNum int, spBlock int) (int, int) {
	if spBlock == 0 {
		klog.V(util.LogWarningLev).Infof("getPositionInTop spBlock=%d is error value, will change to 1", spBlock)
		spBlock = 1
	}
	row := (podNum - 1) % spBlock
	col := (podNum - 1) / spBlock
	return row, col
}

func getSuperPodMap(npuNodes map[string]plugin.NPUNode, nodes []*api.NodeInfo, pluginName string) map[int32]superPod {
	totalNodes := make(map[int32]superPod)
	for _, node := range nodes {
		nNode, ok := npuNodes[node.Name]
		if !ok {
			klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes %s is not npu node",
				pluginName, node.Name)
			continue
		}
		_, exist := totalNodes[nNode.SuperPodID]
		if !exist {
			totalNodes[nNode.SuperPodID] = superPod{}
		}

		totalNodes[nNode.SuperPodID][nNode.Name] = nodeBaseInfo{
			name:       nNode.Name,
			superPodID: nNode.SuperPodID,
			rackID:     nNode.RackID,
		}
	}
	return totalNodes
}

func initTable(superPodSize int, spBlock int) [][][]superPod {
	maxRow := superPodSize/spBlock + 1
	table := make([][][]superPod, spBlock)
	for i := range table {
		table[i] = make([][]superPod, maxRow)
	}
	return table
}

func getSuperPodsInfo(totalNodes map[int32]superPod, superPodSize int, spBlock int) (superPodsInfo, error) {
	superPodTable := initTable(superPodSize, spBlock)
	countVSuperPod := 0
	column := 0
	row := 0
	for index, sp := range totalNodes {
		klog.V(util.LogDebugLev).Infof("super-pod: %d, len: %d", index, len(sp))
		if superPodSize < len(sp) {
			klog.V(util.LogErrorLev).Infof("please adjust super-pod-size, now super-pod-size=(%d) "+
				"but someone superPod's node size=(%d)", superPodSize, len(sp))
			return superPodsInfo{
				superPodTable: superPodTable,
				spCount:       countVSuperPod,
			}, errors.New("super-pod-size is smaller than superPod's node size")
		}

		if spBlock != 0 {
			countVSuperPod += len(sp) / spBlock
			if len(sp) > 0 {
				column = (len(sp) - nodeNum1) / spBlock
				row = (len(sp) - nodeNum1) % spBlock
			} else {
				continue
			}
		}

		klog.V(util.LogInfoLev).Infof("super-pod-id<%d> in table index: column[%d], row[%d]", index, column,
			row)

		if len(superPodTable[row][column]) == 0 {
			superPodTable[row][column] = make([]superPod, 0, 1)
		}
		superPodTable[row][column] = append(superPodTable[row][column], sp)
	}

	return superPodsInfo{
		superPodTable: superPodTable,
		spCount:       countVSuperPod,
	}, nil
}

// filterRackIdByTpBlock filter RackIds if their len is less than tpBlock
func filterRackIdByTpBlock(superPodWithRack map[int32][]nodeBaseInfo, tpBlock int) {
	for rackId, nodes := range superPodWithRack {
		if len(nodes) < tpBlock {
			klog.V(util.LogInfoLev).Infof("the usable nodes %v in rack %v are unreachable because of tp-block",
				nodes, rackId)
			delete(superPodWithRack, rackId)
		}
	}
}

func (tp *module910a5SuperPod) getOriginRackId(superPodWithRackId map[int32][]nodeBaseInfo,
	faultNodeNameMap map[string]struct{}, vSuperPod []plugin.SuperNode) int32 {
	for _, nodeOfFJob := range vSuperPod {
		if _, ok := faultNodeNameMap[nodeOfFJob.Name]; !ok {
			continue
		}
		for rackId, _ := range superPodWithRackId {
			if rackId == nodeOfFJob.RackID {
				klog.V(util.LogInfoLev).Infof("choose origin rackId=%d", rackId)
				return rackId
			}
		}
	}
	return UninitializedRestRackLenMapId
}
