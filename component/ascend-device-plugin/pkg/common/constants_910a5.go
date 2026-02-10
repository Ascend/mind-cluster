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

// Package common a series of common function
package common

const (
	// DefaultRackID default rack id -1
	DefaultRackID = -1
	// Ascend910A5 ascend 910A5 chip
	Ascend910A5 = "Ascend910A5"
	// RackIdAbnormal represents rack id abnormal value
	RackIdAbnormal = -2
	// SuperPodTypeAbnormal represents super pod type abnormal value for A5
	SuperPodTypeAbnormal = 0
	// DevA5 is A5 device type
	DevA5 = "A5"
)

// for DPU subHealth report
const (
	// exist one dpu fault for one npu
	OneDpuFault = 1
)

const (
	// NpuIdxCorrespDpuRangeMiddle npu index middle value
	NpuIdxCorrespDpuRangeMiddle = 4
	// DpuSlotIdx1 dpusolt index
	DpuSlotIdx1 = "1"
	// DpuSlotIdx2 dpusolt index
	DpuSlotIdx2 = "2"
	// DpuSlotIdx9 dpusolt index
	DpuSlotIdx9 = "9"
	// DpuSlotIdx10 dpusolt index
	DpuSlotIdx10 = "10"
	// DpuIpAddrsLen dpu ipaddrs num
	DpuIpAddrsLen = 2
	// NpuNum one node have 8 npu
	NpuNum = 8
)

const (
	// A5300IBoardId board id of 300I A5 with chip specification 1
	A5300IBoardId = 0x1a
	// A5300IBoardId2 board id of 300I A5 with chip specification 2
	A5300IBoardId2 = 0x1b
	// A5300IMainBoardId board id of 300I A5 MainBoard
	A5300IMainBoardId = 0x68
	// A5300I4PMainBoardId board id of 300I A5 4P MainBoard
	A5300I4PMainBoardId = 0x6c
	// A5300ICardName 300I A5 card name
	A5300ICardName = "300I-A5"
	// A54P300ICardName 300I A5 4P card name
	A54P300ICardName = "300I-A5-4p"
)

const (
	// TopologyRefreshTime per 30 second refresh
	TopologyRefreshTime = 30
	// PingMeshConfigCm the pingmesh config map name
	PingMeshConfigCm = "pingmesh-config"
	// RasGlobalKey one of the key of pingmesh config
	RasGlobalKey = "global"
	// RasNetDetectOnStr the string of detect on
	RasNetDetectOnStr = "on"
	// RasNetDetectOffStr the string of detect off
	RasNetDetectOffStr = "off"
	// Server8PTopoPath 8p server topo path
	Server8PTopoPath = "/usr/local/Ascend/driver/topo/a5/server_8p.json"
	// Pod1DTopoPath 1d pod topo path
	Pod1DTopoPath = "/usr/local/Ascend/driver/topo/a5/superpod_1d.json"
	// Pod2DTopoPath 2d pod topo path
	Pod2DTopoPath = "/usr/local/Ascend/driver/topo/a5/superpod_2d.json"
	// Server16PTopoPath 16p server topo path
	Server16PTopoPath = "/usr/local/Ascend/driver/topo/a5/server_16p.json"
	// Server32PTopoPath 32p server topo path
	Server32PTopoPath = "/usr/local/Ascend/driver/topo/a5/server_32p.json"
	// Card1PTopoPath 1p card topo path
	Card1PTopoPath = "/usr/local/Ascend/driver/topo/a5/card_1p.json"
	// Card4PTopoPath 4p card topo path
	Card4PTopoPath = "/usr/local/Ascend/driver/topo/a5/card_4p_mesh.json"
	// HcclTopoFilePathKey HcclTopoFilePath key
	HcclTopoFilePathKey = "HCCL_TOPO_FILE_PATH"
	// PhyLimit phy limit
	PhyLimit = 144
	// PhyLowerLimit phy lower limit
	PhyLowerLimit = 0
	// LogicUpperLimit logic upper limit
	LogicUpperLimit = 212
	// LogicLowerLimit logic lower limit
	LogicLowerLimit = 181
	// PhyPortNumPerDie phy port num per die
	PhyPortNumPerDie = 9
	// LogicPortNumPerDie logic port num per die
	LogicPortNumPerDie = 2
	// DieNumPerDev die num per dev
	DieNumPerDev = 2
	// Peer2Net peer 2 net
	Peer2Net = "PEER2NET"
)

const (
	// ProductTypeServer server
	ProductTypeServer = 0
	// ProductType1D 1d type
	ProductType1D = 1
	// ProductType2D 2d type
	ProductType2D = 2
	// ProductType16PServer 16p
	ProductType16PServer = 3
	// ProductType32PServer 32p
	ProductType32PServer = 4
	// ProductTypeCard1p standard 1p
	ProductType1PCard = 5
	// ProductTypeCard4p standard 4p
	ProductType4PCard = 6
	// UrmaFeId0 for level 1
	UrmaFeId0 = 0
	// UrmaFeId1 for level 0
	UrmaFeId1 = 1
	// UrmaFeId3 for level 2 ubg
	UrmaFeId3 = 3
	// UrmaFeId8 for level 2 uboe
	UrmaFeId8 = 8
	// UrmaFeId9 for level 2 uboe
	UrmaFeId9 = 9
	// FeIdIndexBit bit position in byte start
	FeIdIndexBit = 5
	// FeIdIndexByte byte position in eid byte array start with 1
	FeIdIndexByte = 8
	// FeIdMask fe id value mask
	FeIdMask = 0x1F
	// InvalidSuperPodID all f -> super pod / not super pod
	InvalidSuperPodID = 0xffffffff
	// InvalidSuperPodSize all f -> super pod / not super pod
	InvalidSuperPodSize = 0xffffffff
)

const (
	// PortIDSuffixLen the suffix length of portId
	PortIDSuffixLen = 2
	// DieIDOffset dieId's offset
	DieIDOffset = 3
	// PortIdLimit is the portId limit
	PortIdLimit = 8
)
