/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common this for util method
package common

import "ascend-common/api"

// Default topology values. They are placeholders keyed by chip count only; refine per
// mainBoardId/boardId once the real interconnect topology of each model is confirmed.
const (
	// TopoDefault8Chip is the default flat 8-chip topology.
	TopoDefault8Chip  = "[0,1,2,3,4,5,6,7]"
	Topo4x2Chip       = "[[0,1],[2,3],[4,5],[6,7]]"
	TopoDefault16Chip = "[[0,1,2,3,4,5,6,7],[8,9,10,11,12,13,14,15]]"
	TopoDefaultA3Chip = "[[0,1],[2,3],[4,5],[6,7],[8,9],[10,11],[12,13],[14,15]]"
)

var mainBoardTopoTable = map[uint32]string{
	A900A3SuperPodMainBoardId1:  TopoDefaultA3Chip,
	A900A3SuperPodMainBoardId2:  TopoDefaultA3Chip,
	A9000A3SuperPodMainBoardId1: Topo4x2Chip,
	A9000A3SuperPodMainBoardId2: Topo4x2Chip,
	A800IA3MainBoardId:          TopoDefaultA3Chip,

	api.Atlas850MainBoardID:  TopoDefault8Chip,
	api.Atlas850MainBoardID2: TopoDefault8Chip,
	api.Atlas850MainBoardID3: TopoDefault8Chip,

	api.Atlas950MainBoardID:   TopoDefault8Chip,
	api.Atlas9501DMainBoardID: TopoDefault8Chip,

	api.Atlas3501PMainBoardID: TopoDefault8Chip,
	api.Atlas3502PMainBoardID: Topo4x2Chip,
	api.Atlas3504PMainBoardID: "[[0,1,2,3],[4,5,6,7],[8,9,10,11],[12,13,14,15]]",

	api.Atlas950FlexMainBoardID1: TopoDefault16Chip,
	api.Atlas950FlexMainBoardID2: TopoDefault16Chip,
}

// GetMainBoardTopo returns the topology mapped to mainBoardId, or "" when not found.
func GetMainBoardTopo(mainBoardId uint32) string {
	if !IsValidMainBoardInfo(mainBoardId) {
		return ""
	}
	return mainBoardTopoTable[mainBoardId]
}

var boardTopoTable = map[uint32]string{
	// A300I A2 / 910proB
	A300IA2BoardId: TopoDefault8Chip,
	// A300I A2 64GB
	A300IA2GB64BoardId: TopoDefault8Chip,
	// Atlas 800I A2 server without HCCS (legacy boardId 0x33)
	A800IA2NoneHccsBoardIdOld: TopoDefault8Chip,
	// Atlas 800I A2 server without HCCS (boardId 0x3c)
	A800IA2NoneHccsBoardId: TopoDefault8Chip,
	Atlas200TA2BoardId1:    TopoDefault16Chip,
	Atlas200TA2BoardId2:    TopoDefault16Chip,
	Atlas200TA2BoardId3:    TopoDefault16Chip,
}

// GetBoardTopo returns the topology mapped to boardId for A2 models, or "" when not found.
func GetBoardTopo(boardId uint32) string {
	if boardId == InvalidID {
		return ""
	}
	return boardTopoTable[boardId]
}
