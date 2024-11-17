/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package util is using for the total variable.
package util

const (
	// LogErrorLev for log error.
	LogErrorLev = 1
	// LogWarningLev for log warning.
	LogWarningLev = 2
	// LogInfoLev for log information.
	LogInfoLev = 3
	// LogDebugLev for log debug.
	LogDebugLev = 4
	// ErrorInt return -1 when get error for int
	ErrorInt = -1
	// NPUIndex2 the 2 index.
	NPUIndex2 = 2
	// NPUIndex3 the 3 index.
	NPUIndex3 = 3
	// NPUIndex8 the 8 index.
	NPUIndex8 = 8
	// NPUIndex16 the 16 index.
	NPUIndex16 = 16
	// NPUIndex7 the 7 index.
	NPUIndex7 = 7
	// NPUIndex4 the 4 index.
	NPUIndex4 = 4
	// NPUIndex5 the 5 index.
	NPUIndex5 = 5
	// NPUIndex6 the 6 index.
	NPUIndex6 = 6
	// NPUIndex1 the 1 index.
	NPUIndex1 = 1
	// NPUIndex9 the 9 index.
	NPUIndex9 = 9
	// NPUIndex10 the 10 index.
	NPUIndex10 = 10
	// NPUIndex11 the 11 index.
	NPUIndex11 = 11
	// NPUIndex12 the 12 index.
	NPUIndex12 = 12
	// NPUIndex13 the 13 index.
	NPUIndex13 = 13
	// NPUIndex14 the 14 index.
	NPUIndex14 = 14
	// NPUIndex15 the 15 index.
	NPUIndex15 = 15
	// CoreNum32 32 core 910
	CoreNum32 = 32
	// CoreNum30 30 core 910
	CoreNum30 = 30
	// CpuNum14 14 cpu 910
	CpuNum14 = 14
	// MapInitNum for map init length.
	MapInitNum = 3
	// Base10 for const 10.
	Base10 = 10
	// BitSize64 for const 64
	BitSize64 = 64
	// MaxSliceNum max slice number
	MaxSliceNum = 128
	// NPUHexKilo for const 1000,volcano frame used.
	NPUHexKilo = 1000
	// HwPreName pre name
	HwPreName = "huawei.com/"
	// NPUCardPreName for NPU card pre-Name.
	NPUCardPreName = "huawei.com/Ascend"
	// ArchSelector MindX-dl arch selector.
	ArchSelector = "host-arch"
	// HuaweiArchArm for arm.
	HuaweiArchArm = "huawei-arm"
	// HuaweiArchX86 for x86.
	HuaweiArchX86 = "huawei-x86"

	// Accelerator for custom tag.
	Accelerator = "accelerator"
	// Accelerator910Value Ascend910 tag.
	Accelerator910Value = "huawei-Ascend910"
	// Accelerator310Value Ascend310 tag.
	Accelerator310Value = "huawei-Ascend310"

	// CMSelectorKey selector key in scheduler configmap.
	CMSelectorKey = "selector"
	// CMInitParamKey init param key in scheduler configmap
	CMInitParamKey = "init-params"
	// AcceleratorType for selector.
	AcceleratorType = "accelerator-type"
	// CardAcceleratorType for card mode.
	CardAcceleratorType = "card"
	// Module910bx16AcceleratorType for module mode.
	Module910bx16AcceleratorType = "module-910b-16"
	// Module910bx8AcceleratorType for module mode.
	Module910bx8AcceleratorType = "module-910b-8"
	// Card910bx2AcceleratorType for module mode.
	Card910bx2AcceleratorType = "card-910b-2"
	// Card910bx2InferAcceleratorType for infer mode.
	Card910bx2InferAcceleratorType = "card-910b-infer"
	// ModuleAcceleratorType for module mode.
	ModuleAcceleratorType = "module"
	// ChipAcceleratorType for chip mode.
	ChipAcceleratorType = "chip"
	// HalfAcceleratorType for half mode
	HalfAcceleratorType = "half"
	// ServerType server type value takes Ascend310P-10-dual/Ascend910-32...
	ServerType = "servertype"
	// ServerTypeDual dual card
	ServerTypeDual = "dual"

	// NPU910CardName for judge 910 npu resource.
	NPU910CardName = "huawei.com/Ascend910"
	// NPU910CardNamePre for getting card number.
	NPU910CardNamePre = "Ascend910-"
	// NPU310PCardName for judge 310P npu resource.
	NPU310PCardName = "huawei.com/Ascend310P"
	// NPU310CardName for judge 310 npu resource.
	NPU310CardName = "huawei.com/Ascend310"
	// NPU310CardNamePre for getting card number.
	NPU310CardNamePre = "Ascend310-"
	// NPU310PCardNamePre for getting card number.
	NPU310PCardNamePre = "Ascend310P-"
	// AscendNPUPodRealUse for NPU pod real use cards.
	AscendNPUPodRealUse = "huawei.com/AscendReal"
	// AscendNPUCore for NPU core num, like 56; Records the chip name that the scheduler assigns to the pod.
	AscendNPUCore = "huawei.com/npu-core"
	// Ascend910bName for judge Ascend910b npu resource.
	Ascend910bName = "huawei.com/Ascend910b"

	// SegmentEnable for VNPU segment enable flag. Default is "false".
	SegmentEnable = "presetVirtualDevice"

	// DevInfoNameSpace device-plugin install Namespace
	DevInfoNameSpace = "kube-system"
	// MindXDlNameSpace mindx dl Namespace
	MindXDlNameSpace = "mindx-dl"
	// DevInfoPreName like "mindx-dl-deviceinfo-ubuntu"
	DevInfoPreName = "mindx-dl-deviceinfo-"
	// NodeDCmInfoNamePrefix is for noded to report node healthy state
	NodeDCmInfoNamePrefix = "mindx-dl-nodeinfo-"
	// NodedHeartbeatTimeKey is the key of heartbeat time from configmap data of noded
	NodedHeartbeatTimeKey = "nodedHeartbeatTime"
	// NodeDNodeHeartbeatIntervalKey is key of node heartbeat interval from configmap data of noded
	NodeDNodeHeartbeatIntervalKey = "NodeDNodeHeartbeatInterval"

	// DevInfoCMKey mindx-dl-deviceinfo configmap key
	DevInfoCMKey = "DeviceInfoCfg"
	// NodeInfoCMKey node info configmap key
	NodeInfoCMKey = "NodeInfo"
	// RePropertyCacheName rescheduling keyword in init env.cache
	RePropertyCacheName = "re-scheduling"
	// CmCheckCode Check code key
	CmCheckCode = "checkCode"
	// CmName Name of ReSchedulerConfigmap
	CmName = "vcjob-fault-npu-cm"
	// JobRecovery keywords for retain
	JobRecovery = "job-recovery"

	// DeleteOperator informer delete operator
	DeleteOperator = "delete"
	// AddOperator informer add operator
	AddOperator = "add"
	// UpdateOperator informer update operator
	UpdateOperator = "update"

	// PodPredicateTime set pod PodPredicateTime for using by device-plugin.
	PodPredicateTime = "predicate-time"
	// NodeNotMeetTopologyWarning node not satisfy the schedulable topology warning.
	NodeNotMeetTopologyWarning = "the npus on this node don't satisfy the schedulable topology"
	// ArgumentError argument nil error.
	ArgumentError = "invalid argument"
	// JobKindKey for define the Job kind:ascend-310P, ascend-910
	JobKindKey = "ring-controller.atlas"
	// JobKind910Value in ring-controller.atlas.
	JobKind910Value = "ascend-910"
	// JobKind310Value in ring-controller.atlas.
	JobKind310Value = "ascend-310"
	// JobKind310PValue 310p ring controller name
	JobKind310PValue = "ascend-310P"
	// JobKind910BValue 910B ring controller name
	JobKind910BValue = "ascend-910b"
	// DistributedJobKey flag for distributed job
	DistributedJobKey = "distributed-job"
	// DistributedJobValue indicate distributed job
	DistributedJobValue = "true"
	// StandaloneJobValue indicate standalone job
	StandaloneJobValue = "false"
)

const (
	// AffScore0 value 0 for scored.
	AffScore0 = iota
	// AffScore1 value 1 for scored.
	AffScore1
	// AffScore2 value 2 for scored.
	AffScore2
	// AffScore3 value 3 for scored.
	AffScore3
	// AffScore4 value 4 for scored.
	AffScore4
	// AffScore5 value 4 for scored.
	AffScore5
	// AffScore6 value 4 for scored.
	AffScore6
	// AffScore7 value 4 for scored.
	AffScore7
	// AffScore8 value 4 for scored.
	AffScore8
)

const (
	// JobNotEnqueue job enqueue failed
	JobNotEnqueue = -1
	// JobEnqueue job enqueue success
	JobEnqueue = 1
	// JobEnqueueSkip skip the judgement of ascend-volcano-plugin in the job enqueue phase
	JobEnqueueSkip = 0
	// PodGroupInqueue the pg Inqueue status
	PodGroupInqueue = "Inqueue"
	// PodGroupPending the pg Pending status
	PodGroupPending = "Pending"
	// PodGroupRunning the pg Running status
	PodGroupRunning = "Running"
	// PodGroupUnknown the pg Unknown status
	PodGroupUnknown = "Unknown"
	// PodGroupUnschedulableType the pg Unschedulable Condition
	PodGroupUnschedulableType = "Unschedulable"
)

// VTemplate for vNode resource
type VTemplate struct {
	// ChipKind Ascend910/Ascend310P
	ChipKind   string
	AICore     int
	AICPU      int
	DVPPEnable string
}

// VResource resource dimensions
type VResource struct {
	Aicore int
	Aicpu  int
	DVPP   string
}
