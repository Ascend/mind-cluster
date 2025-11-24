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

// Package common constants
package common

const (
	// Hex hexadecimal
	Hex = 16
)

// board id
const (
	// A300IA2BoardId board id of A300I A2 32GB
	A300IA2BoardId = 0x28
	// A300IA2GB64BoardId board id of A300I A2 64GB
	A300IA2GB64BoardId = 0x29
	// A800IA2NoneHccsBoardIdOld is the boardid of a800i a2 device, 0x33 is server without hccs
	A800IA2NoneHccsBoardIdOld = 0x33
	// A800IA2NoneHccsBoardId 0x33 changed to 0x3c, and compatible with the old boardId, since 2024.9.4
	A800IA2NoneHccsBoardId = 0x3c
	// EmptyBoardId is the boardid of device before initialized
	EmptyBoardId = 0x00
)

// device usage
const (
	// Infer means device for inference
	Infer = "infer"
	// Train means device for training
	Train = "train"
)

// ring related
const (
	// Ascend910RingsNum indicates the number of devices in a ring
	Ascend910RingsNum = 4
	// Ascend910BRingsNumTrain indicates the number of devices in a ring
	Ascend910BRingsNumTrain = 8
	// Ascend910BRingsNumInfer indicates the number of devices in a ring
	Ascend910BRingsNumInfer = 1
	// A200TA2RingsNum indicates the number of devices in a ring
	A200TA2RingsNum = 16
)

// fault related
const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// RestartRequest restart request
	RestartRequest = "RestartRequest"
	// RestartBusiness restart business
	RestartBusiness = "RestartBusiness"
	// FreeRestartNPU wait free and restart NPU
	FreeRestartNPU = "FreeRestartNPU"
	// RestartNPU restart NPU
	RestartNPU = "RestartNPU"
	// SeparateNPU separate NPU
	SeparateNPU = "SeparateNPU"
	// NormalNPU normal NPU
	NormalNPU = "NormalNPU"
	// UnknownLevel unknown level
	UnknownLevel = "Unknown"

	// FaultRecover device fault recover
	FaultRecover = int8(0)
	// FaultOccur device fault occur
	FaultOccur = int8(1)
	// FaultOnce once device fault
	FaultOnce = int8(2)
)

// StatusInfoFile status info file
const StatusInfoFile = "/tmp/status.json"

// container pause and resume strategy
const (
	// NeverStrategy never deal
	NeverStrategy = "never"
	// SingleStrategy deal single container and device
	SingleStrategy = "singleRecover"
	// RingStrategy deal container and device on ring
	RingStrategy = "ringRecover"
)

// container status
const (
	// StatusRunning container is running
	StatusRunning = "running"
	// StatusPausing container is pausing
	StatusPausing = "pausing"
	// StatusPaused container is paused
	StatusPaused = "paused"
	// StatusResuming container is resuming
	StatusResuming = "resuming"
)

// display description
const (
	// DescNormal container is runnig, normal description
	DescNormal = "normal"
	// DescUnknown container status is not in cache, unknown description
	DescUnknown = "unknown"
)

// device status
const (
	// StatusIgnorePause device ignore pause
	StatusIgnorePause = "ignore"
	// StatusNeedPause device need pause
	StatusNeedPause = "needPause"
)
