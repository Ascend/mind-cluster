/*
Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

import (
	"testing"

	"ascend-common/api"
)

func TestGetMainBoardTopo(t *testing.T) {
	cases := []struct {
		name string
		id   uint32
		want string
	}{
		// A3 SuperPod / A800I A3: 16 chips grouped in pairs [0,1],[2,3],...
		{name: "hit", id: A900A3SuperPodMainBoardId1, want: "[[0,1],[2,3],[4,5],[6,7],[8,9],[10,11],[12,13],[14,15]]"},
		{name: "hit-superpod-a900-2", id: A900A3SuperPodMainBoardId2, want: "[[0,1],[2,3],[4,5],[6,7],[8,9],[10,11],[12,13],[14,15]]"},
		{name: "hit-superpod-a9000-1", id: A9000A3SuperPodMainBoardId1, want: Topo4x2Chip},
		{name: "hit-superpod-a9000-2", id: A9000A3SuperPodMainBoardId2, want: Topo4x2Chip},
		{name: "hit-a800i-a3", id: A800IA3MainBoardId, want: "[[0,1],[2,3],[4,5],[6,7],[8,9],[10,11],[12,13],[14,15]]"},
		// Atlas 850/950/3501P: flat 8 chips.
		{name: "hit-atlas-850", id: api.Atlas850MainBoardID, want: "[0,1,2,3,4,5,6,7]"},
		{name: "hit-atlas-850-2", id: api.Atlas850MainBoardID2, want: "[0,1,2,3,4,5,6,7]"},
		{name: "hit-atlas-850-3", id: api.Atlas850MainBoardID3, want: "[0,1,2,3,4,5,6,7]"},
		{name: "hit-atlas-950", id: api.Atlas950MainBoardID, want: "[0,1,2,3,4,5,6,7]"},
		{name: "hit-atlas-950-1d", id: api.Atlas9501DMainBoardID, want: "[0,1,2,3,4,5,6,7]"},
		// 0x68/0x6c also cover the device-plugin/pkg/common A5300I main board ids
		// (300I-Atlas350, same value as Atlas3501P/3504P).
		{name: "hit-atlas-350-1p", id: api.Atlas3501PMainBoardID, want: "[0,1,2,3,4,5,6,7]"},
		{name: "hit-atlas-350-2p", id: api.Atlas3502PMainBoardID, want: "[[0,1],[2,3],[4,5],[6,7]]"},
		// Atlas 350 4P: 8 chips in two groups of 4.
		{name: "hit-atlas-350-4p", id: api.Atlas3504PMainBoardID, want: "[[0,1,2,3],[4,5,6,7],[8,9,10,11],[12,13,14,15]]"},
		// A5 PC168P: 16 chips, two groups of 8.
		{name: "hit-a5-pc168p-1", id: api.Atlas950FlexMainBoardID1, want: "[[0,1,2,3,4,5,6,7],[8,9,10,11,12,13,14,15]]"},
		{name: "hit-a5-pc168p-2", id: api.Atlas950FlexMainBoardID2, want: "[[0,1,2,3,4,5,6,7],[8,9,10,11,12,13,14," +
			"15]]"},
		// A5 custom UBX/TX/DY cards (constants_v2.go) are intentionally absent: lookup
		// returns "" so the scheduler falls back. api.UbxMainBoardID == A5UBXMainBoardId (0x44).
		{name: "miss-a5-ubx", id: A5UBXMainBoardId, want: ""},
		{name: "miss-a5-ubx-alias-api", id: api.UbxMainBoardID, want: ""},
		{name: "miss-a5-tx", id: A5TXMainBoardId, want: ""},
		{name: "miss-a5-dy", id: A5DYMainBoardId, want: ""},
		{name: "miss", id: 0xDEAD, want: ""},
		{name: "invalid-id", id: InvalidID, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetMainBoardTopo(tc.id); got != tc.want {
				t.Fatalf("GetMainBoardTopo(%d) = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
}

func TestGetBoardTopo(t *testing.T) {
	cases := []struct {
		name string
		id   uint32
		want string
	}{
		{name: "hit-a300i-a2", id: A300IA2BoardId, want: TopoDefault8Chip},
		{name: "hit-a300i-a2-gb64", id: A300IA2GB64BoardId, want: TopoDefault8Chip},
		{name: "hit-a800i-a2-none-hccs", id: A800IA2NoneHccsBoardId, want: TopoDefault8Chip},
		{name: "hit-a800i-a2-none-hccs-old", id: A800IA2NoneHccsBoardIdOld, want: TopoDefault8Chip},
		// Atlas 200T A2: 16 chips, two groups of 8.
		{name: "hit-atlas-200t-a2-1", id: Atlas200TA2BoardId1, want: TopoDefault16Chip},
		{name: "hit-atlas-200t-a2-2", id: Atlas200TA2BoardId2, want: TopoDefault16Chip},
		{name: "hit-atlas-200t-a2-3", id: Atlas200TA2BoardId3, want: TopoDefault16Chip},
		// A3 board ids are keyed in the mainBoard table; the board table misses them
		{name: "miss-a3-bin-board", id: A900A3SuperPodBin1BoardId, want: ""},
		{name: "miss-unknown", id: 0xDEAD, want: ""},
		{name: "invalid-id", id: InvalidID, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetBoardTopo(tc.id); got != tc.want {
				t.Fatalf("GetBoardTopo(%d) = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
}
