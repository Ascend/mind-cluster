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
	"math"
	"sort"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func getBestScore(usedResourceMap map[int]softShareDevResource, cardIds []int,
	reqResource softShareDevResource, maxHbm int) int {
	bestScore := 0
	if len(usedResourceMap) == 0 {
		return bestScore
	}
	for _, cardId := range cardIds {
		usedResourceQuota, ok := usedResourceMap[cardId]
		if ok && usedResourceQuota.aicoreQuota+reqResource.aicoreQuota <= util.MaxAicoreQuota &&
			usedResourceQuota.hbmQuota+reqResource.hbmQuota <= maxHbm &&
			usedResourceQuota.schedulingPolicy == usedResourceQuota.schedulingPolicy {
			curScore := util.MaxNodeScoreForSoftShareDev - (util.MaxAicoreQuota - usedResourceQuota.aicoreQuota -
				reqResource.aicoreQuota)
			bestScore = int(math.Max(float64(bestScore), float64(curScore)))
		}
	}
	if bestScore == 0 {
		bestScore = len(usedResourceMap)
	}
	return bestScore
}

func npuPrioritySort(nodeTop []int, usedMap map[int]softShareDevResource, requestResource softShareDevResource,
	maxHbm int) []int {
	notUsedCardIDs := make([]int, 0, len(nodeTop))
	bestCard := -1
	bestScore := -1
	for _, npuID := range nodeTop {
		usedRes, exists := usedMap[npuID]
		if !exists {
			notUsedCardIDs = append(notUsedCardIDs, npuID)
			continue
		}
		if usedRes.schedulingPolicy != requestResource.schedulingPolicy ||
			usedRes.aicoreQuota+requestResource.aicoreQuota > util.MaxAicoreQuota ||
			usedRes.hbmQuota+requestResource.hbmQuota > maxHbm {
			continue
		}
		curScore := requestResource.aicoreQuota + usedRes.aicoreQuota
		if curScore > bestScore {
			bestScore = curScore
			bestCard = npuID
		}
	}
	if bestCard != -1 {
		return []int{bestCard}
	}
	if len(notUsedCardIDs) > 0 {
		sort.Ints(notUsedCardIDs)
		return notUsedCardIDs[:1]
	}
	return nil
}
