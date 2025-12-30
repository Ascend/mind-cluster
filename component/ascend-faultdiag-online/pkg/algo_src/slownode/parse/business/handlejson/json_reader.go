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

/*
Package handlejson converts JSON to DB statements
*/
package handlejson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

var regexpFile = regexp.MustCompile(`^\d{10}$`)

const pollWaitTime = 2 * time.Second
const maxFilesCount = 100000

// StartReadJson 读取json数据
func StartReadJson(
	snpRankCtx *context.SnpRankContext,
	parseCtx *ParseFileContext,
	stopFlag chan struct{},
	jobStartTime int64,
) {
	if snpRankCtx == nil || snpRankCtx.ContextData == nil || snpRankCtx.ContextData.Config == nil || parseCtx == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil snpRankCtx, snpRankCtx.ContextData, " +
			"snpRankCtx.ContextData.Config or parseCtx")
		return
	}
	go func() {
		for {
			select {
			case _, ok := <-stopFlag:
				if !ok {
					return
				}
			default:
				isUpdated, err := updateCurFile(snpRankCtx.ContextData.Config.RankDir, parseCtx.CurFile, jobStartTime)
				if err != nil || !isUpdated {
					time.Sleep(pollWaitTime)
					continue
				}
				jsonDataSlice, err := readJsonData(snpRankCtx.ContextData, parseCtx.CurFile)
				if err != nil {
					time.Sleep(pollWaitTime)
					continue
				}
				snpRankCtx.JsonDataQue <- jsonDataSlice
			}
		}
	}()
}

func readJsonData(ctxData *contextdata.SnpRankContextData, curFile *TimeStampFile) ([]*model.JsonData, error) {
	lines, newOffset, err := utils.ReadLinesFromOffset(filepath.Join(ctxData.Config.RankDir, curFile.Name), curFile.Offset)
	if err != nil {
		return nil, fmt.Errorf("read profile %s failed: %v", curFile.Name, err)
	}
	curFile.Offset = newOffset
	var res []*model.JsonData = nil
	for _, line := range lines {
		var data *model.JsonData
		if err := json.Unmarshal([]byte(line), &data); err != nil { // 转json失败，说明这里还没完整写完一行，回退游标
			curFile.Offset -= int64(len(line))
			break
		}
		if data == nil {
			continue
		}
		res = append(res, data)
		updateStepCount(data.Name, ctxData)
	}
	return res, nil
}

func updateStepCount(name string, ctxData *contextdata.SnpRankContextData) {
	if !strings.Contains(strings.ToLower(name), constants.StepWord) {
		return
	}
	if ctxData.StepCount >= constants.ClosStep {
		return
	}
	stepId, err := utils.SplitNum(name)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse the 'Name' field, error: %v", err)
		return
	}
	if ctxData.StartStep == 0 {
		ctxData.StartStep = stepId
		hwlog.RunLog.Infof("[SLOWNODE PARSE]Start step: %d, rank num is: %s",
			ctxData.StartStep, filepath.Base(ctxData.Config.RankDir))
	}
	ctxData.StepCount = stepId - ctxData.StartStep
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Read step count: %d, step id: %d, rank num is: %s",
		ctxData.StepCount, stepId, filepath.Base(ctxData.Config.RankDir))
}

func updateCurFile(rankDir string, curFile *TimeStampFile, jobStartTime int64) (bool, error) {
	profiles, err := matchAndSortProfiles(rankDir)
	if err != nil {
		return false, err
	}
	if len(curFile.Name) == 0 { // 初始化
		return checkFileTime(curFile, profiles, jobStartTime), nil
	}
	// 判断文件是否更新
	fileInfo, err := os.Stat(filepath.Join(rankDir, curFile.Name))
	if err != nil {
		return false, err
	}
	if curFile.Offset < fileInfo.Size()-1 { // 当前文件还没有读取完
		return true, nil
	}
	if curFile.Name != profiles[len(profiles)-1] { // 旧文件已经完成读取完成，且有新文件，跳到新文件
		nextIdx := utils.IndexOf(profiles, curFile.Name) + 1
		curFile.Name = profiles[nextIdx]
		curFile.Offset = 0
		return true, nil
	}
	return false, nil
}

func checkFileTime(curFile *TimeStampFile, profiles []string, jobStartTime int64) bool {
	for _, profile := range profiles {
		if profile < fmt.Sprintf("%d", jobStartTime) {
			continue
		}
		curFile.Name = profile
		curFile.Offset = 0
		return true
	}
	return false
}

func matchAndSortProfiles(rankDir string) ([]string, error) {
	files, err := os.ReadDir(rankDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %v", err)
	}

	var count = 0
	var matchedFiles []string = nil
	for _, file := range files {
		count += 1
		if count >= maxFilesCount {
			return matchedFiles, fmt.Errorf("reached the max files count: %d", maxFilesCount)
		}
		// 跳过内层目录遍历
		if file.IsDir() {
			continue
		}
		if regexpFile.MatchString(file.Name()) {
			if err := utils.CheckFilePerm(filepath.Join(rankDir, file.Name()), true, false); err != nil {
				return matchedFiles, err
			}
			matchedFiles = append(matchedFiles, file.Name())
		}
	}

	if len(matchedFiles) != 0 {
		sort.Strings(matchedFiles)
		return matchedFiles, nil
	}
	return matchedFiles, fmt.Errorf("no matching file found")
}
