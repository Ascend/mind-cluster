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

// Package chip4nodex is using for 300I server affinity schedule.
package chip4nodex

import (
	"fmt"
	"math"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// judgeNodeAndTaskNPU Determine whether the task required NPU number can be met
func (tp *chip4nodex) judgeNodeAndTaskNPU(taskNPU int, nodeTop []int) error {
	// need 4Pmesh affinity
	if is4PmeshAffinity(taskNPU) {
		if judgeNodeAndTaskNpuIn4Pmesh(taskNPU, nodeTop) {
			return nil
		}
		return fmt.Errorf("node topo does not meet task in 4Pmesh affinity, require(%d), available %v", taskNPU, nodeTop)
	}
	// do not need 4Pmesh affinity
	if taskNPU <= len(nodeTop) {
		return nil
	}
	return fmt.Errorf("node topo does not meet task require(%d), available %v", taskNPU, nodeTop)
}

// validNPUJob check the job req npu num and mode
func (tp *chip4nodex) validNPUJob() *api.ValidateResult {
	vResult := &api.ValidateResult{}
	var vErr error
	defer func() {
		if vErr != nil {
			vResult.Pass = false
			vResult.Reason = vErr.Error()
			vResult.Message = vErr.Error()
		}
	}()
	// check parameter.
	if tp == nil {
		vErr = fmt.Errorf("nil plugin")
		klog.V(util.LogErrorLev).Infof("ValidNPUJob err: %s.", vErr)
		return vResult
	}
	// There is no need to set sp-block and tp-block in standard cluster server
	if tp.SpBlockNPUNum != 0 {
		klog.V(util.LogWarningLev).Infof("There is no need to set sp-block in standard cluster server.")
	}
	if tp.TpBlockNPUNum != util.LeastTpBlock {
		klog.V(util.LogWarningLev).Infof("There is no need to set tp-block in standard cluster server.")
	}
	// check job mode:distribute and single.
	if vErr = tp.checkJobMode(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobTrainMode: %s.", vErr)
		return vResult
	}

	return nil
}

// checkJobMode to check job mode:distribute and single.
func (tp *chip4nodex) checkJobMode() error {
	if tp.NPUTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", tp.Name)
		return fmt.Errorf("%s no npu job", tp.Name)
	}
	klog.V(util.LogDebugLev).Infof("checkJobMode job(%s) has %d tasks.", tp.Name, len(tp.Tasks))
	nTaskReqNpuNum := tp.ReqNPUNum / tp.NPUTaskNum
	if nTaskReqNpuNum <= tp.MaxNodeNPUNum {
		return nil
	}
	return fmt.Errorf("%s checkJobMode %s req npu is invalid", tp.GetPluginName(), tp.Name)
}

// getNodeBestScore Based on the number of NPUs requested by the task and the number of NPUs actually available on the node,
// look up the corresponding score from affScoreList.
func (tp *chip4nodex) getNodeBestScore(taskNPUNum int, npuTop []int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	npuNum := len(npuTop)
	if npuNum < 1 || npuNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid, npu topo is: %v", npuNum, npuTop)
	}
	return tp.affScoreList[taskNPUNum-1][npuNum-1], nil
}

// judgeNodeAndTaskNpuIn4Pmesh Determine whether a node can meet 4Pmesh scheduling
func judgeNodeAndTaskNpuIn4Pmesh(taskNPU int, nodeTop []int) bool {
	meshCardCount, fullMeshes := getNodeMeshInfo(nodeTop)
	// The required number of cards must be a multiple of 4,
	//and the number of available meshes must be greater than or equal to this.
	if taskNPU >= cardsNumPerMesh {
		return fullMeshes >= taskNPU/cardsNumPerMesh
	}
	// The number of required cards is less than 4, and as long as any mesh meets the conditions, it will be scheduled.
	for i := 0; i < cardsNumPerMesh; i++ {
		if meshCardCount[i] >= taskNPU {
			return true
		}
	}
	return false
}

// getNodeMeshInfo Obtain the mesh information composed of available cards within a node
func getNodeMeshInfo(nodeTop []int) ([]int, int) {
	meshCardCount := make([]int, cardsNumPerMesh)
	for _, n := range nodeTop {
		meshId := n / cardsNumPerMesh
		if meshId >= cardsNumPerMesh {
			klog.V(util.LogErrorLev).Infof("Invalid card id: %v,out of range. nodeTop is %v", n, nodeTop)
			break
		}
		meshCardCount[meshId]++
	}
	fullMeshes := 0
	for _, mesh := range meshCardCount {
		if mesh == cardsNumPerMesh {
			fullMeshes++
		}
	}
	return meshCardCount, fullMeshes
}

// getAvailableMeshInNode Obtain the number of available meshes and the required number of meshes under the current demand card count taskNPUNum
func getAvailableMeshInNode(taskNPUNum int, nodeTop []int) (int, int) {
	meshCardCount, fullMeshes := getNodeMeshInfo(nodeTop)
	availableMesh := 0
	needMesh := 0

	// After the previous filtering, the taskNPUNum here must be less than 4 or divisible by 4.
	if taskNPUNum >= cardsNumPerMesh {
		needMesh = taskNPUNum / cardsNumPerMesh
		return needMesh, fullMeshes
	}

	for _, mesh := range meshCardCount {
		if mesh >= taskNPUNum {
			availableMesh++
		}
	}
	needMesh = 1
	return needMesh, availableMesh
}

// getNodeBestScoreIn4Pmesh Based on the number of NPUs requested by the task and the number of NPUs actually available on the node,
// and considering the 4Pmesh affinity requirements, look up the corresponding score value in the affScoreList.
func (tp *chip4nodex) getNodeBestScoreIn4Pmesh(taskNPUNum int, npuTop []int) (int, error) {
	if taskNPUNum < 1 || taskNPUNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("task req npu num<%d> is invalid", taskNPUNum)
	}
	npuNum := len(npuTop)
	if npuNum < 1 || npuNum > tp.MaxNodeNPUNum {
		return 0, fmt.Errorf("node npu num<%d> is invalid", npuNum)
	}

	needMesh, availableMesh := getAvailableMeshInNode(taskNPUNum, npuTop)
	// Priority is given to fewer meshes, followed by fewer available cards.
	// The available card check table result is always less than tp.MaxNodeNPUNum,
	//so the meshScore needs to be multiplied by tp.MaxNodeNPUNum to make the weight greater than the remaining number of cards.
	meshScore := tp.MaxNodeNPUNum * (availableMesh - needMesh)
	cardScore := tp.affScoreList[taskNPUNum-1][npuNum-1]
	return meshScore + cardScore, nil
}

// scoreNodeFor4Pmesh Scoring logic in scenarios where 4Pmesh && need Affinity
func (tp *chip4nodex) scoreNodeFor4Pmesh(taskNPUNum int, npuTop []int) float64 {
	klog.V(util.LogInfoLev).Info("begin to score node in 4Pmesh scenario")
	bestScore, err := tp.getNodeBestScoreIn4Pmesh(taskNPUNum, npuTop)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
		return 0
	}
	if tp.MaxNodeNPUNum == maxNodeNPUNumX8 {
		return float64(scoreWeightX16 * (scoreWeightX16 - bestScore))
	}
	if tp.MaxNodeNPUNum == maxNodeNPUNumX16 {
		return float64(scoreWeightX64 * (scoreWeightX64 - bestScore))
	}
	klog.V(util.LogErrorLev).Infof("error: MaxNodeNPUNum is not 8 or 16, value is %v", tp.MaxNodeNPUNum)
	return 0
}

// scoreNodeForGeneral Scoring logic for scenarios where not 4Pmesh or affinity does not need to be ensured
func (tp *chip4nodex) scoreNodeForGeneral(taskNPUNum int, npuTop []int) float64 {
	klog.V(util.LogInfoLev).Info("begin to score node in general scenario")
	bestScore, err := tp.getNodeBestScore(taskNPUNum, npuTop)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s ScoreBestNPUNodes getErr: %s", tp.GetPluginName(), err)
		return 0
	}
	return float64(tp.MaxNodeNPUNum * (tp.MaxNodeNPUNum - bestScore))
}

// selectNPUIn4Pmesh NPU selection logic in 4Pmesh scenarios
func (tp *chip4nodex) selectNPUIn4Pmesh(taskNPUNum int, nodeTop []int) []int {
	// select single mesh
	if taskNPUNum < cardsNumPerMesh {
		return selectNPUinSingleMesh(taskNPUNum, nodeTop)
	}
	// select multiple meshes
	return selectNPUMultiMesh(taskNPUNum, nodeTop)
}

// selectNPUMultiMesh The required number of cards is a multiple of 4, select multiple meshes.
func selectNPUMultiMesh(taskNPUNum int, nodeTop []int) []int {
	if taskNPUNum < 0 {
		klog.V(util.LogErrorLev).Infof("invalid taskNPUNum:%d", taskNPUNum)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("begin to select npu in multi mesh:taskNPUNum=%d,nodeTop=%v",
		taskNPUNum, nodeTop)
	meshCardCount, _ := getNodeMeshInfo(nodeTop)

	ret := make([]int, 0, taskNPUNum)

	for id, count := range meshCardCount {
		if count != cardsNumPerMesh {
			continue
		}
		for card := id * cardsNumPerMesh; card < (id+1)*cardsNumPerMesh &&
			len(ret) < taskNPUNum; card++ {
			ret = append(ret, card)
		}
	}
	if len(ret) < taskNPUNum {
		klog.V(util.LogErrorLev).Infof("nodeTop %v do not satisify taskNPUNum %d", nodeTop, taskNPUNum)
		return nil
	}
	return ret
}

// selectNPUinSingleMesh If fewer than 4 cards are needed, select only one mesh.
func selectNPUinSingleMesh(taskNPUNum int, nodeTop []int) []int {
	if taskNPUNum < 0 {
		klog.V(util.LogErrorLev).Infof("invalid taskNPUNum:%d", taskNPUNum)
		return nil
	}
	klog.V(util.LogInfoLev).Infof("begin to select npu in single mesh:taskNPUNum=%d,nodeTop=%v",
		taskNPUNum, nodeTop)
	meshCardCount, _ := getNodeMeshInfo(nodeTop)

	const uninitialized = -1
	const positiveInfinity = math.MaxInt32
	chosenMesh := uninitialized
	minExtra := positiveInfinity

	for id, count := range meshCardCount {
		extra := count - taskNPUNum
		if count >= taskNPUNum && extra < minExtra {
			minExtra = extra
			chosenMesh = id
		}
		if minExtra == 0 {
			break
		}
	}

	// chosenMesh is still the initial value, and no available mesh can be selected.
	if chosenMesh == -1 {
		klog.V(util.LogErrorLev).Infof("there is no available mesh in %v when choose %d", nodeTop, taskNPUNum)
		return nil
	}

	ret := make([]int, 0, taskNPUNum)
	for _, card := range nodeTop {
		if card/cardsNumPerMesh == chosenMesh {
			ret = append(ret, card)
			if len(ret) == taskNPUNum {
				break
			}
		}
	}

	if len(ret) < taskNPUNum {
		klog.V(util.LogErrorLev).Infof("nodeTop %v do not satisify taskNPUNum %d", nodeTop, taskNPUNum)
		return nil
	}
	return ret
}
