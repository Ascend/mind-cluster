/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for Huawei Ascend pin affinity scheduling utilities.
*/

package util

// Scoring-framework bit contract (bit-lexicographic strict priority: high bit wins,
// low bits never compensate). Composite formula =
// Σ(dimension segment value << registered weight) × ScheduleHandler.ScoreWeight; weight is the
// dimension's bit offset (segment start; assigned by initScorePlugins: topology=ScoreTopoShift /
// previousNode=ScoreOriginalShift / subHealth=ScoreHealthShift / chipCount=ScoreAvailShift).
// A one-bit lead is decisive; ×ScoreWeight is a uniform positive scaling that preserves order.
// Per-dimension segment values:
//
//	topology {1 FitNormal, 0 evict-only} (scoreTopology normalizes the absolute score);
//	previousNode {2,1,0} tier (tier 2 → 2<<11=bit12 P1 last-non-fault, tier 1 → bit11 P2
//	  back-to-original; tier must stay integral and only 2 encodes P1). With
//	  PreferPreviousNode=false the value collapses to the {1,0} binary: current job's
//	  previous-fault task landing (IsFaultTaskByRank×PrefNodeMap, same source as isFaultPod)
//	  writes 0, other normal candidates hold 1;
//	subHealth {0..3} (3 healthy, 2 switch-only, 1 card-only, 0 card+switch coexisting;
//	  FaultHandle.ScoreSubHealthGrade writes 0..2 in place for sub-healthy nodes, dimension
//	  predate defaults to 3);
//	chipCount (req/free)×255 rounded with math.Round inside the dimension.
//
// fault-pod: the fault-node still enters candidates this round, all otherNodes get P1 (leave
// the fault first), selfNode gets P2 as fallback, peerNode stays 0. Non-fault-pod: selfNode P1,
// all otherNodes P2, peerNode stays 0. No arbitrary best-node pick inside a category — the
// fixed topology tie-breaks among equals.
// Registration order topology→previousNode→subHealth→chipCount is fixed: reordering changes the
// accumulated bits and breaks the per-bit agreement with the independent frameworkBitFormula;
// the fault dimension was removed from the registry.
// prev/subHealth are independent of topo: the topo frame is carried only by the topology
// dimension; prev uses a uniform base-value frame as a read-only partition, subHealth grades on
// the predate default frame.
// Legacy path (non-chip policy, plugin/factory.go): scoreMap accumulates in place then
// ×= ScoreWeight; it does not reference these constants.
const (
	// Each Shift is both the segment start and the dimension's registered weight: the
	// synthesizer does bits |= uint16(segment value) << weight (see initScorePlugins).
	// bit12 (P1) has no dedicated constant — it is the derived tier of prev tier 2 shifted by
	// ScoreOriginalShift (2<<11); an explicit constant would double-source the tier encoding
	// (tests express it as 2<<ScoreOriginalShift).
	// ScoreOriginalShift previousNode segment start (bit11), also its registered weight:
	// tier 1 → bit11 P2 (back-to-original), tier 2 → 2<<11=bit12 P1 (last-non-fault).
	ScoreOriginalShift = 11
	// ScoreTopoShift topology segment start (bit10), also its registered weight;
	// segment {1 FitNormal, 0 evict-only}.
	ScoreTopoShift = 10
	// ScoreHealthShift subHealth segment start (bit8, occupying bit9-8, 2 bits),
	// also its registered weight; segment 0..3 (3 healthy, 2 switch-only, 1 card-only, 0 both),
	// 3<<8=0x300 does not overflow into topo bit10.
	ScoreHealthShift = 8
	// ScoreAvailShift chipCount segment start (bit7-0, 8 bits), also its registered weight
	// (=0, segment integers OR directly into the low 8 bits). Value range (0,255] is guaranteed
	// by the FitNormal-predicate free≥req; segment width is implicitly bounded by
	// ScoreHealthShift (bit8).
	ScoreAvailShift = 0
)
