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
// Package externalbridge for node parse interface
package externalbridge

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/handlejson"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/sdk"
	"ascend-faultdiag-online/pkg/core/model"
)

// StartDataParse start the data parse
func StartDataParse(cg config.DataParseModel) {
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Start parsing data, job start time is: %d", cg.JobStartTime)
	sdk.StartParse(cg)
}

// StopDataParse stop data parse
func StopDataParse(cg config.DataParseModel) {
	hwlog.RunLog.Info("[SLOWNODE PARSE]Receive parsing stop message")
	sdk.StopParse(cg)
}

// ReloadDataParse reload data parse
func ReloadDataParse(cg config.DataParseModel) {
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Reload parsing data, job reload time is: %d", cg.JobStartTime)
	sdk.ReloadParse(cg)
}

// RegisterDataParse register the data-parse by uintptr
func RegisterDataParse(callback model.CallbackFunc) {
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Callback for data parsing: %+v", callback)
	sdk.RegisterParseCallback(callback)
}

// StartMergeParGroupInfo 集群侧合并并行域信息
func StartMergeParGroupInfo(cg config.DataParseModel) {
	hwlog.RunLog.Info("[SLOWNODE PARSE]Start merging parallel group information")
	if err := handlejson.MergeParallelGroupInfo(cg); err != nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]Failed to merge parallel group information:", err)
		return
	}
	hwlog.RunLog.Info("[SLOWNODE PARSE]Succeeded in merging parallel group information")
}

// RegisterMergeParGroup register the parallel group info merge by uintptr
func RegisterMergeParGroup(callback model.CallbackFunc) {
	hwlog.RunLog.Infof("[SLOWNODE PARSE]Callback for parallel group information merging: %+v", callback)
	handlejson.RegisterParGroupCallback(callback)
}
