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
Package nslb is using for HuaWei Ascend pin tor affinity.
*/
package nslb

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type allocateTest struct {
	name      string
	superPods []*superPodTors
	vPodNum   int
	vPodSize  int
	wantErr   bool
}

func buildAllocateTest01() allocateTest {
	return allocateTest{
		name: "01-Allocate with enough full tors",
		superPods: []*superPodTors{
			{
				name:       "sp0",
				superPodId: 0,
				full:       util.NPUIndex16,
				remainFull: util.NPUIndex16,
				torCount:   util.NPUIndex8,
			},
			{
				name:       "sp1",
				superPodId: 0,
				full:       util.NPUIndex16,
				remainFull: util.NPUIndex16,
				torCount:   util.NPUIndex8,
			},
		},
		vPodNum:  util.NPUIndex2,
		vPodSize: util.NPUIndex8,
		wantErr:  false,
	}
}

func buildAllocateTest02() allocateTest {
	return allocateTest{
		name: "02-Allocate with not enough tors",
		superPods: []*superPodTors{
			{
				name:       "sp0",
				superPodId: 0,
				full:       util.NPUIndex8,
				remainFull: util.NPUIndex8,
				torCount:   util.NPUIndex8,
			},
		},
		vPodNum:  util.NPUIndex2,
		vPodSize: util.NPUIndex8,
		wantErr:  true,
	}
}

func buildAllocateTest03() allocateTest {
	return allocateTest{
		name: "03-Allocate with mixed full and partial tors",
		superPods: []*superPodTors{
			{
				name:       "sp0",
				superPodId: 0,
				full:       util.NPUIndex8,
				remainFull: util.NPUIndex8,
				partial:    util.NPUIndex4,
				remainPart: util.NPUIndex4,
				torCount:   util.NPUIndex8,
			},
		},
		vPodNum:  util.NPUIndex2,
		vPodSize: util.NPUIndex6,
		wantErr:  false,
	}
}

func buildAllocateTestCases() []allocateTest {
	return []allocateTest{
		buildAllocateTest01(),
		buildAllocateTest02(),
		buildAllocateTest03(),
	}
}

func TestAllocate(t *testing.T) {
	for _, tt := range buildAllocateTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			allocate(tt.superPods, tt.vPodNum, tt.vPodSize)
			allocated := 0
			for _, sp := range tt.superPods {
				allocated += (sp.usedFull + sp.usedPartial) / tt.vPodSize
			}
			if (allocated < tt.vPodNum) != tt.wantErr {
				t.Errorf("Allocate() got allocated = %d, wantErr %v", allocated, tt.wantErr)
			}
		})
	}
}

type initSuperPodTorsTest struct {
	name     string
	torCount int
	id       int32
	tors     []*plugin.Tor
	want     *superPodTors
}

func buildInitSuperPodTorsTest01() initSuperPodTorsTest {
	return initSuperPodTorsTest{
		name:     "01-Init with full TORs only",
		torCount: util.NPUIndex8,
		id:       util.NPUIndex1,
		tors: []*plugin.Tor{
			{FreeServerCount: util.NPUIndex8},
			{FreeServerCount: util.NPUIndex8},
		},
		want: &superPodTors{
			name:       "1",
			superPodId: util.NPUIndex1,
			full:       util.NPUIndex16,
			remainFull: util.NPUIndex16,
			torCount:   util.NPUIndex8,
		},
	}
}

func buildInitSuperPodTorsTest02() initSuperPodTorsTest {
	return initSuperPodTorsTest{
		name:     "02-Init with partial TORs only",
		torCount: util.NPUIndex8,
		id:       util.NPUIndex2,
		tors: []*plugin.Tor{
			{FreeServerCount: util.NPUIndex4},
			{FreeServerCount: util.NPUIndex2},
		},
		want: &superPodTors{
			name:       "2",
			superPodId: util.NPUIndex2,
			full:       0,
			remainFull: 0,
			torCount:   util.NPUIndex8,
		},
	}
}

func buildInitSuperPodTorsTest03() initSuperPodTorsTest {
	return initSuperPodTorsTest{
		name:     "03-Init with mixed full and partial TORs",
		torCount: util.NPUIndex8,
		id:       util.NPUIndex3,
		tors: []*plugin.Tor{
			{FreeServerCount: util.NPUIndex8},
			{FreeServerCount: util.NPUIndex4},
		},
		want: &superPodTors{
			name:       "3",
			superPodId: util.NPUIndex3,
			full:       util.NPUIndex8,
			remainFull: util.NPUIndex8,
			torCount:   util.NPUIndex8,
		},
	}
}

func buildInitSuperPodTorsTestCases() []initSuperPodTorsTest {
	return []initSuperPodTorsTest{
		buildInitSuperPodTorsTest01(),
		buildInitSuperPodTorsTest02(),
		buildInitSuperPodTorsTest03(),
	}
}

func TestInitSuperPodTors(t *testing.T) {
	for _, tt := range buildInitSuperPodTorsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			got := initSuperPodTors(tt.torCount, tt.id, tt.tors)
			if got.name != tt.want.name || got.superPodId != tt.want.superPodId ||
				got.full != tt.want.full || got.remainFull != tt.want.remainFull ||
				got.torCount != tt.want.torCount {
				t.Errorf("initSuperPodTors() = %v, want %v", got, tt.want)
			}
		})
	}
}
