/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package v1 contains API Schema definitions for the mindxdl v1 API group
package v1

import (
	"github.com/kubeflow/common/pkg/apis/common/v1"
)

const (
	// MempoolFrameworkName is the name of Mempool framework
	MempoolFrameworkName = "mempool"
	// MempoolReplicaTypeMaster is the type for Scheduler of distribute ML
	MempoolReplicaTypeMaster v1.ReplicaType = "Master"
)

const (
	// ScaleOutTypeLabel label task parameter surface large mesh type, e.g. RoCE, UBoE ignore case
	ScaleOutTypeLabel = "scaleout-type"
	// ScaleOutTypeRoCE label value of task for RoCE, which must be kept in uppercase format
	ScaleOutTypeRoCE = "ROCE"
	// ScaleOutTypeUBoE label value of task for UBoE, which must be kept in uppercase format
	ScaleOutTypeUBoE = "UBOE"

	// PortAddrTypeRoCE the value must be kept in uppercase format to be same with device-plugin
	PortAddrTypeRoCE = "ROCE"
	// PortAddrTypeUBoE the value must be kept in uppercase format to be same with device-plugin
	PortAddrTypeUBoE = "UBOE"
	// PortAddrTypeUBG the value must be kept in uppercase format to be same with device-plugin
	PortAddrTypeUBG = "UBG"
	// PortAddrTypeUBC the value must be kept in uppercase format to be same with device-plugin
	PortAddrTypeUBC = "UBC"
	// PortAddrTypeUB the value must be kept in uppercase format to be same with device-plugin
	PortAddrTypeUB = "UB"

	// RankAddrTypeEID the value must be kept in uppercase format
	RankAddrTypeEID = "EID"
	// RankAddrTypeIP the value must be kept in uppercase format
	RankAddrTypeIP = "IP"
)
