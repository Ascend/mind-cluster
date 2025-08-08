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

// Package service provides some funcs relevant to the global rank
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/dealdb"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/enum"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

// QueryZpHdDurationFunc 查找zp host device 耗时的函数
type QueryZpHdDurationFunc func(dbCtx *db.SnpDbContext, startNsMin int64,
	startNsMax int64) (*model.HostDeviceDuration, error)

// buildQueryZpHdDurationFunc 构造查询zp算子耗时的函数
func buildQueryZpHdDurationFunc(groupKey int64, traffic int64) QueryZpHdDurationFunc {
	if groupKey != -1 { // tp算子存在
		return func(dbCtx *db.SnpDbContext, startNsMin int64,
			startNsMax int64) (*model.HostDeviceDuration, error) {
			opIdView, err := dealdb.QueryStepReduceTypeCnt(dbCtx, groupKey, startNsMin, startNsMax)
			if err != nil {
				return nil, err
			}
			if opIdView == nil {
				return nil, nil
			}
			queryParam := []any{groupKey, startNsMin, startNsMax, opIdView.Id, traffic}
			duration, err := dealdb.QueryZpDurationWhenTpEnable(dbCtx, queryParam)
			if err != nil {
				return nil, err
			}
			return duration, nil
		}
	}
	return func(dbCtx *db.SnpDbContext, startNsMin int64, startNsMax int64) (*model.HostDeviceDuration, error) {
		return dealdb.QueryZpDurationWhenTpDisable(dbCtx, startNsMin, startNsMax)
	}
}

// queryGroupInfo 查找tp并行域的key在表STRING_IDS中对应的ID
func queryGroupInfo(infoMap map[string]*model.OpGroupInfo, dbCtx *db.SnpDbContext, groupName string) (int64, error) {
	groupInfoKeys := make([]string, 0, len(infoMap))
	for key := range infoMap {
		groupInfoKeys = append(groupInfoKeys, key)
	}
	sort.Slice(groupInfoKeys, func(i, j int) bool {
		ni := extractNumber(groupInfoKeys[i])
		nj := extractNumber(groupInfoKeys[j])
		return ni < nj
	})
	// 从并行域中获取key作为group_name来源
	for _, key := range groupInfoKeys {
		opGroupInfo := infoMap[key]
		if matchGroupKey(opGroupInfo, groupName) {
			groupIdView, err := dealdb.QueryStringId(dbCtx, key)
			if err == nil && groupIdView != nil {
				return groupIdView.Id, nil
			}
		}
	}
	return -1, nil
}

func extractNumber(groupInfo string) int {
	parts := strings.Split(groupInfo, "_")
	if len(parts) == 0 {
		return 0
	}
	number, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return number
}

func matchGroupKey(opGroupInfo *model.OpGroupInfo, groupName string) bool {
	if groupName == string(enum.Pp) {
		return opGroupInfo.GroupName == string(enum.Pp)
	}
	return opGroupInfo.GroupName == string(enum.Tp) && len(opGroupInfo.GlobalRanks) > 1
}

// queryDataLoaderHostDuration 查找dataloader host耗时
func queryDataLoaderHostDuration(dbCtx *db.SnpDbContext, stepId int64) (*model.Duration, error) {
	timeDur, err := dealdb.QueryHostStartEndTime(dbCtx, stepId)
	if err != nil {
		return nil, err
	}
	if timeDur == nil {
		return nil, nil
	}
	return dealdb.QueryDataLoaderHost(dbCtx, timeDur.StartNs, timeDur.EndNs)
}

// 收集某个group在各个step中的时间信息
func collectGroupRank(dbCtx *db.SnpDbContext, groupName int64, startEndNsList []*model.StepStartEndNs,
	queryZpHdDurationFunc QueryZpHdDurationFunc) ([]*model.StepGlobalRank, error) {
	result := make([]*model.StepGlobalRank, 0)
	for _, ns := range startEndNsList {
		duration, err := queryZpHdDurationFunc(dbCtx, ns.StartNs, ns.EndNs)
		if err != nil {
			return nil, err
		}
		ppDeviceDur, err := dealdb.QueryPpDevDuration(dbCtx, groupName, ns.StartNs, ns.EndNs)
		if err != nil {
			return nil, err
		}
		ppHostDur, err := dealdb.QueryPpHostDuration(dbCtx, groupName, ns.StartNs, ns.EndNs)
		if err != nil {
			return nil, err
		}
		hostDuration, err := queryDataLoaderHostDuration(dbCtx, ns.Id)
		if err != nil {
			return nil, err
		}
		zPDevice := setHostDeviceDuration(duration, "device")
		zPHost := setHostDeviceDuration(duration, "host")
		pPDevice := setDuration(ppDeviceDur)
		pPHost := setDuration(ppHostDur)
		dataLoaderHost := setDuration(hostDuration)
		if utils.AllZero([]int64{zPDevice, zPHost, pPDevice, pPHost, dataLoaderHost}) {
			continue
		}
		var stepGlobalRank = &model.StepGlobalRank{
			StepIndex:      ns.Id,
			ZPDevice:       zPDevice,
			ZPHost:         zPHost,
			PPDevice:       pPDevice,
			PPHost:         pPHost,
			DataLoaderHost: dataLoaderHost,
		}
		result = append(result, stepGlobalRank)
	}
	return result, nil
}

func setHostDeviceDuration(dur *model.HostDeviceDuration, unit string) int64 {
	if dur == nil {
		return 0
	}
	if unit == "host" {
		return dur.HostDuration
	}
	return dur.DeviceDuration
}

func setDuration(dur *model.Duration) int64 {
	if dur == nil {
		return 0
	}
	return dur.Dur
}

// CollectGlobalRank 收集全量global rank数据
func CollectGlobalRank(ctxData *contextdata.SnpRankContextData,
	startEndNsList []*model.StepStartEndNs) ([]*model.StepGlobalRank, error) {
	// 从parallel_group.json中读取并行域信息
	parallelGroupInfoMap, err := readParallelGroupInfo(ctxData.Config.ParGroupJsonInputFilePath)
	if err != nil {
		return nil, err
	}

	tpGroupKey, err := queryGroupInfo(parallelGroupInfoMap, ctxData.DbCtx, string(enum.Tp))
	if err != nil {
		return nil, err
	}
	queryZpHdDurationFunc := buildQueryZpHdDurationFunc(tpGroupKey, ctxData.Config.Traffic)

	ppGroupKey, err := queryGroupInfo(parallelGroupInfoMap, ctxData.DbCtx, string(enum.Pp))
	if err != nil {
		return nil, err
	}

	// 收集所有的rank信息
	globalRank, err := collectGroupRank(ctxData.DbCtx, ppGroupKey, startEndNsList, queryZpHdDurationFunc)
	if err != nil {
		return nil, err
	}

	return globalRank, nil
}

// readParallelGroupInfo 从json文件中读取并行域信息
func readParallelGroupInfo(inputPath string) (map[string]*model.OpGroupInfo, error) {
	parallelGroupInfo, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	var infoMap map[string]*model.OpGroupInfo
	if err = json.Unmarshal(parallelGroupInfo, &infoMap); err != nil {
		return nil, fmt.Errorf("the path of the parallel group informateion file is as follows: '%s', "+
			"error unmarshalling JSON: %v", inputPath, err)
	}
	return infoMap, nil
}
