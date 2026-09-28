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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	defaultMaxHBM = 32 * util.MBPerGB
	hbm16GB       = 16 * util.MBPerGB
	hbm10GB       = 10 * util.MBPerGB
	hbm20GB       = 20 * util.MBPerGB
	hbm30GB       = 30 * util.MBPerGB
	hbm5GB        = 5 * util.MBPerGB

	aicoreQuota20 = 20
	aicoreQuota30 = 30
	aicoreQuota40 = 40
	aicoreQuota50 = 50
	aicoreQuota90 = 90

	defaultScore        = 0
	resourceExceedScore = 1
	num2                = 2
	num3                = 3
)

func TestGetBestScore(t *testing.T) {
	tests := []struct {
		name         string
		usedResource map[int]softShareDevResource
		cardIds      []int
		reqResource  softShareDevResource
		maxHbm       int
		want         int
	}{
		{
			name:         "empty used resource map",
			usedResource: map[int]softShareDevResource{},
			cardIds:      []int{0, 1},
			reqResource:  softShareDevResource{},
			maxHbm:       defaultMaxHBM,
			want:         defaultScore,
		},
		{
			name: "valid resource match",
			usedResource: map[int]softShareDevResource{
				0: {aicoreQuota: aicoreQuota20, hbmQuota: hbm10GB, schedulingPolicy: util.SoftShareDevPolicyFixedShare},
			},
			cardIds: []int{0, 1},
			reqResource: softShareDevResource{aicoreQuota: aicoreQuota30, hbmQuota: hbm20GB,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare},
			maxHbm: defaultMaxHBM,
			want:   util.MaxNodeScoreForSoftShareDev - (util.SoftShareAicoreBudget() - aicoreQuota20 - aicoreQuota30),
		},
		{
			name: "resource exceed limit",
			usedResource: map[int]softShareDevResource{
				0: {aicoreQuota: aicoreQuota90, hbmQuota: hbm30GB, schedulingPolicy: util.SoftShareDevPolicyFixedShare},
			},
			cardIds: []int{0},
			reqResource: softShareDevResource{
				aicoreQuota: aicoreQuota20, hbmQuota: hbm5GB, schedulingPolicy: util.SoftShareDevPolicyFixedShare},
			maxHbm: defaultMaxHBM,
			want:   resourceExceedScore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBestScore(tt.usedResource, tt.cardIds, tt.reqResource, tt.maxHbm)
			if got != tt.want {
				t.Errorf("getBestScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

type npuPrioritySortTestCase struct {
	name            string
	nodeTop         []int
	usedMap         map[int]softShareDevResource
	requestResource softShareDevResource
	maxHbm          int
	want            []int
}

func buildNpuPrioritySortTestCases() []npuPrioritySortTestCase {
	return []npuPrioritySortTestCase{
		{
			name:    "no used cards, return sorted first",
			nodeTop: []int{num3, 1, num2},
			usedMap: map[int]softShareDevResource{},
			requestResource: softShareDevResource{
				aicoreQuota:      aicoreQuota50,
				hbmQuota:         hbm16GB,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			maxHbm: defaultMaxHBM,
			want:   []int{1},
		},
		{
			name:    "best card match",
			nodeTop: []int{0, 1, num2},
			usedMap: map[int]softShareDevResource{
				0: {aicoreQuota: aicoreQuota40, hbmQuota: hbm10GB, schedulingPolicy: util.SoftShareDevPolicyFixedShare},
				1: {aicoreQuota: aicoreQuota30, hbmQuota: hbm10GB, schedulingPolicy: util.SoftShareDevPolicyFixedShare},
			},
			requestResource: softShareDevResource{
				aicoreQuota:      aicoreQuota20,
				hbmQuota:         hbm10GB,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			maxHbm: defaultMaxHBM,
			want:   []int{0},
		},
		{
			name:    "policy mismatch",
			nodeTop: []int{0, 1},
			usedMap: map[int]softShareDevResource{
				0: {aicoreQuota: aicoreQuota20, hbmQuota: hbm10GB, schedulingPolicy: util.SoftShareDevPolicyElastic},
			},
			requestResource: softShareDevResource{
				aicoreQuota:      aicoreQuota20,
				hbmQuota:         hbm10GB,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			maxHbm: defaultMaxHBM,
			want:   []int{1},
		},
	}
}

func TestNpuPrioritySort(t *testing.T) {
	tests := buildNpuPrioritySortTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := npuPrioritySort(tt.nodeTop, tt.usedMap, tt.requestResource, tt.maxHbm)
			if len(got) != len(tt.want) {
				t.Fatalf("npuPrioritySort() len = %v, want len %v", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("npuPrioritySort()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetBestScoreWithOversell(t *testing.T) {
	defer util.SetSoftShareCoreScaling(1)

	used := map[int]softShareDevResource{
		0: {aicoreQuota: aicoreQuota40 + aicoreQuota40, hbmQuota: hbm20GB,
			schedulingPolicy: util.SoftShareDevPolicyFixedShare},
	}
	req := softShareDevResource{aicoreQuota: aicoreQuota40, hbmQuota: hbm10GB,
		schedulingPolicy: util.SoftShareDevPolicyFixedShare}

	util.SetSoftShareCoreScaling(1)
	if got := getBestScore(used, []int{0}, req, defaultMaxHBM); got != resourceExceedScore {
		t.Fatalf("scaling=1 getBestScore() = %v, want exceed %v", got, resourceExceedScore)
	}

	util.SetSoftShareCoreScaling(1.5)
	want := util.MaxNodeScoreForSoftShareDev - (util.SoftShareAicoreBudget() -
		(aicoreQuota40+aicoreQuota40) - aicoreQuota40)
	if got := getBestScore(used, []int{0}, req, defaultMaxHBM); got != want {
		t.Fatalf("scaling=1.5 getBestScore() = %v, want %v", got, want)
	}
}
