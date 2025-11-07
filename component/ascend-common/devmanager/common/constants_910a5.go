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

// Package common define common variable
package common

import "k8s.io/apimachinery/pkg/util/sets"

var (
	// a900A5SuperPodBoardIds for A900 A5 Super Pod Board IDs
	a900A5SuperPodBoardIds = sets.NewInt32(A900A5SuperPodBin1BoardId, A900A5SuperPodBin2BoardId,
		A900A5SuperPodBin3BoardId)
	// a800A5ServerBoardIds for A800 A5 Server Board Ids
	a800A5ServerBoardIds = sets.NewInt32(A800A5ServerBin21BoardId, A800A5ServerBin22BoardId,
		A800A5ServerMultiModalBin22BoardId)
	// standardCard300IA5BoardIds for 300I A5 Board IDs
	standardCard300IA5BoardIds = sets.NewInt32(A5300IBoardId, A5300IBoardId2)
)

const (
	// MaxUBEIDByteLen max eid bytes len for device ub port
	MaxUBEIDByteLen = 128 / 8

	// MaxUBCNAByteLen max cna bytes len for device ub port
	MaxUBCNAByteLen = 32 / 8

	// A900A5SuperPodBin1BoardId board id of A900 A5 SuperPod Bin0
	A900A5SuperPodBin1BoardId = 0x28
	// A900A5SuperPodBin2BoardId board id of A900 A5 SuperPod Bin1-1
	A900A5SuperPodBin2BoardId = 0x29
	// A900A5SuperPodBin3BoardId board id of A900 A5 SuperPod Bin1-2
	A900A5SuperPodBin3BoardId = 0x2a

	// A800A5ServerMultiModalBin22BoardId board id of A800 A5 Server for MultiModal Bin2-2
	A800A5ServerMultiModalBin22BoardId = 0x0c
	// A800A5ServerBin21BoardId board id of A800 A5 Server for MultiModal Bin2-1
	A800A5ServerBin21BoardId = 0x2b
	// A800A5ServerBin22BoardId board id of A800 A5 Server for MultiModal Bin2-2
	A800A5ServerBin22BoardId = 0x2c

	// A5300IBoardId board id of 300I A5
	A5300IBoardId = 0x1a

	// A5300IBoardId2 board id of 300I A5 specification 2
	A5300IBoardId2 = 0x1b
)
