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

// Package nodelevel is used for file reading and writing, as well as data processing.
package nodelevel

import (
	"path/filepath"
	"time"

	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
	"ascend-faultdiag-online/pkg/algo_src/slownode/nodeleveldatarecorder"
)

// NodeJobLevelDetectionLoopA3 A3节点侧任务级检测
func NodeJobLevelDetectionLoopA3(conf config.AlgoInputConfig) {
	detectionCount := 0
	for {
		if !jobdetectionmanager.GetDetectionLoopStatusNodeLevel(conf.JobName) ||
			detectionCount > maxDetectionLoop {
			break
		}
		startTime := time.Now().Unix()
		/* job路径 */
		jobPath := filepath.Join(conf.FilePath, conf.JobName)
		/* 检查路径不存在 */
		if !config.CheckExistDirectoryOrFile(jobPath, true, "node", conf.JobName) {
			jobdetectionmanager.SetDetectionExitedStatusNodeLevel(conf.JobName, true)
			nodeleveldatarecorder.DeleteJobDetectionRecorder(conf.JobName)
			return
		}
		/* 获取当前job相关卡并行域信息, valideRanks为当前任务当前节点上使用的ranks */
		tpParallel, validRanks := isCurJobEnableDetectionA3(jobPath, conf.RankIds)
		if tpParallel == nil || validRanks == nil {
			jobdetectionmanager.SetDetectionExitedStatusNodeLevel(conf.JobName, true)
			nodeleveldatarecorder.DeleteJobDetectionRecorder(conf.JobName)
			return
		}
		/* topo更新情况，npu卡变动需要另加else逻辑,此处若检测过程中rank目录有变化无法感知和更新 */
		if detectionCount == 0 {
			nodeleveldatarecorder.InitCurJobDetectionRecorder(conf.JobName, validRanks)
		}
		jobLevelDetectionA3(tpParallel, validRanks, conf)
		endTime := time.Now().Unix()
		config.LoopDetectionIntervalCheckSwitch(endTime-startTime, conf.NsecondsOneDetection,
			conf.JobName, conf.DetectionLevel)
		detectionCount++
	}
	/* loop break */
	nodeleveldatarecorder.DeleteJobDetectionRecorder(conf.JobName)
}
