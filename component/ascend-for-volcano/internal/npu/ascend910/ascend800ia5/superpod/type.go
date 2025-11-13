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
Package superpod is using for HuaWei ascend 800I A5 SuperPod affinity schedule.
*/
package superpod

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend800ia5"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type module800SuperPod struct {
	ascend800ia5.Base800ia5
	netUnhealthyKey        string
	spBlock                int
	nodeVPodId             map[string]string
	isSoftSuperPodAffinity bool
}

const (
	// AcceleratorType super pod accelerator type
	AcceleratorType = "800I-SuperPod-A5-8"
	// AcceleratorTypeTrain super pod accelerator type for train server
	AcceleratorTypeTrain = "800T-SuperPod-A5-8"
	// InferSchedulerName name of infer server scheduler
	InferSchedulerName = "huawei.com/Ascend910" + AcceleratorType
	// TrainSchedulerName name of train scheduler
	TrainSchedulerName  = "huawei.com/Ascend910" + AcceleratorTypeTrain
	nodeNPUNumber       = 8
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"

	jobCheckFailedReason       = "npu num is invalid"
	spBlockInvalidReason       = "sp-block is invalid"
	getNPUFromPodFailedPattern = "%s getUsableTopFromNode err: %s"
	scoreForNode               = 100000000
	delayingTime               = 10
	superPodRankKey            = "super-pod-rank"
	superPodIdKey              = "super-pod-id"
	superPodAffinity           = "super-pod-affinity"
	softRequire                = "soft"
)

type superPodInfo struct {
	firstLevel     remainderTop
	countVSuperPod int
}

type vPodIdRecorder struct {
	unReadyId  []string
	leftIndex  int
	rightIndex int
}

type remainderTop = [][][]superPod

type superPod = map[string]plugin.NPUNode
