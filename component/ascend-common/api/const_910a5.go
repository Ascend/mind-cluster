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

// Package api common const
package api

// for DPU
const (
	// DpuLogPrefix  log prefix for dpu feature
	DpuLogPrefix = "[DPU controller]"
	// DpuStatusDown dpu status down
	DpuStatusDown = "down"
	// DpuInfoCMDataKey mindcluster-dpuinfo-cm data key, record dpu list
	DpuInfoCMDataKey = "DPUList"
	// DpuInfoCMBusTypeKey mindcluster-dpuinfo-cm data key, record dpu busType
	DpuInfoCMBusTypeKey = "BusType"
	// DpuInfoCMNpuToDpusMapKey mindcluster-dpuinfo-cm data key, record relationship between npu and dpu
	DpuInfoCMNpuToDpusMapKey = "NpuToDpusMap"
	// DpuInfoCMNamePrefix dpu info configmap name prefix
	DpuInfoCMNamePrefix = "mindcluster-dpuinfo-"
)

// Device Type
const (
	// A5PodType a5 pod type
	A5PodType = "900SuperPod-A5-8"
	// Ascend800ia5x8 normal inference server label
	Ascend800ia5x8 = "800I-A5-8"
	// Ascend800ta5x8 normal training server label
	Ascend800ta5x8 = "800T-A5-8"
	// Ascend800ia5Stacking stacking server label
	Ascend800ia5Stacking = "800I-Stacking-A5-8"
	// Ascend800ia5SuperPod superpod inference server label
	Ascend800ia5SuperPod = "800I-SuperPod-A5-8"
	// Ascend800ta5SuperPod superpod training server label
	Ascend800ta5SuperPod = "800T-SuperPod-A5-8"
)

const (
	// NpuCountPerNode NPU count per node
	NpuCountPerNode = 8
)

const (
	// RackIDKey rack id in a super pod
	RackIDKey = "rackID"
	// VersionA5 Type for RAS Net Fault Detection in A5
	VersionA5 = "A5"
	// VersionA3 Type for RAS Net Fault Detection in A3
	VersionA3 = "A3"
)

const (
	// LevelInfoTypeUB rank table UB for A5
	LevelInfoTypeUB = "UB"
	// LevelInfoTypeUBG rank table UBG for A5
	LevelInfoTypeUBG = "UBG"
	// LevelInfoTypeUBoE rank table UBOE for A5
	LevelInfoTypeUBoE = "UBOE"
	// LevelInfoTypeRoCE rank table ROCE for A5
	LevelInfoTypeRoCE = "ROCE"
	// LevelInfoTypeIgnore rank table for A5
	LevelInfoTypeIgnore = ""

	// RankLevel0 rank table levelList level0
	RankLevel0 = 0
	// RankLevel1 rank table levelList level1
	RankLevel1 = 1
	// RankLevel2 rank table levelList level2
	RankLevel2 = 2
	// RankLevel3 rank table levelList level3
	RankLevel3 = 3
	// RankLevelCnt rank table levelList count
	RankLevelCnt = 4

	// NetAttrEmpty is empty
	NetAttrEmpty = ""
	// NetTypeTopo TOPO_FILE_DESC
	NetTypeTopo = "TOPO_FILE_DESC"
	// NetTypeCLOS CLOS
	NetTypeCLOS = "CLOS"

	// DefaultClusterName default NetInstanceID value is CLUSTER1
	DefaultClusterName = "CLUSTER1"
	// DefaultRandAddrPlaneID default planeId
	DefaultRandAddrPlaneID = "CLUSTER"
)
