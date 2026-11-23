/* Copyright 2025 Huawei Technologies Co., Ltd.

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

// Package sdk provides node parse
package sdk

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
	coreModel "ascend-faultdiag-online/pkg/core/model"
)

var callbackFunc coreModel.CallbackFunc = nil

// RegisterParseCallback register a function to get node parse result
func RegisterParseCallback(callback coreModel.CallbackFunc) {
	callbackFunc = callback
}

// DealParseCallback 处理清洗回调
func DealParseCallback(snpCtxSlice []*context.SnpRankContext, parseJobInfo *model.ParseJobInfo) {
	if parseJobInfo == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil parseJobInfo")
		return
	}
	callbackPollerFunc := func() (bool, error) {
		stopCallback, err := parseCallback(snpCtxSlice, parseJobInfo.JobName, parseJobInfo.JobId)
		if err != nil {
			hwlog.RunLog.Error("[SLOWNODE PARSE]Failed to process parse callback result:", err)
		}
		return stopCallback, nil
	}
	parseJobInfo.JobWg.Add(1)
	go func(jobId string) {
		defer parseJobInfo.JobWg.Done()
		err := utils.Poller(callbackPollerFunc, constants.CallbackTime, 0, parseJobInfo.StopParseFlag)
		if err != nil {
			hwlog.RunLog.Error("[SLOWNODE PARSE]Failed to rotate callback:", err)
		} else {
			hwlog.RunLog.Info("[SLOWNODE PARSE]Succeeded in rotating callback and exited")
		}
	}(parseJobInfo.JobId)
}

func parseCallback(snpCtxSlice []*context.SnpRankContext, jobName string, jobId string) (bool, error) {
	minStepCount := int64(math.MaxInt64)
	rankIds := make([]string, 0)
	for _, snpCtx := range snpCtxSlice {
		if snpCtx == nil || snpCtx.ContextData == nil {
			continue
		}
		stepCount := snpCtx.ContextData.StepCount
		if stepCount < minStepCount {
			minStepCount = stepCount
		}
		rankIds = append(rankIds, snpCtx.RankId)
	}
	stopCallback := minStepCount >= constants.ClosStep
	if err := processParseResult(jobName, jobId, minStepCount, rankIds); err != nil {
		return stopCallback, err
	}
	return stopCallback, nil
}

func processParseResult(jobName string, jobId string, stepCount int64, rankIds []string) error {
	isFinished := true
	stepCountRes := stepCount
	if stepCount == int64(math.MaxInt64) {
		isFinished = false
		stepCountRes = 0
	}
	result := &model.NodeDataParseResult{
		JobName:      jobName,
		JobId:        jobId,
		IsFinished:   isFinished,
		FinishedTime: time.Now().UnixMilli(),
		StepCount:    stepCountRes,
		RankIds:      rankIds,
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent error: %v", err)
	}
	if callbackFunc != nil {
		go callbackFunc(string(jsonStr))
	}
	return nil
}
