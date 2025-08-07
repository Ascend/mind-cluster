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

// Package clusterlevel is used for file reading and writing, as well as data processing.
package clusterlevel

import (
	"path/filepath"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
)

// ClusterJobLevelDetectionLoopA3 A3集群侧任务级检测
func ClusterJobLevelDetectionLoopA3(conf config.AlgoInputConfig) {
	/* 记录节点侧检测结果文件是否有更新 */
	recorder := make(map[string]int64)
	detectionCount := 0
	for {
		if !jobdetectionmanager.GetDetectionLoopStatusClusterLevel(conf.JobName) ||
			detectionCount > maxLoopDetection {
			break
		}
		startTime := time.Now().Unix()
		/* job路径 */
		jobPath := filepath.Join(conf.FilePath, conf.JobName)
		/* 检查路径 */
		if !config.CheckExistDirectoryOrFile(jobPath, true, "cluster", conf.JobName) {
			jobdetectionmanager.SetDetectionExitedStatusClusterLevel(conf.JobName, true)
			return
		}
		/* 获取当前job topology中pp并行域信息 */
		flag, ppInfo := getJobLevelPpParallelDomain(jobPath)
		if !flag {
			jobdetectionmanager.SetDetectionExitedStatusClusterLevel(conf.JobName, true)
			return
		}
		/* tp并行域可能跨节点，集群侧获取完整tp并行域整合节点侧慢tp并行域结果 */
		flag, tpInfo := getJobLevelTpParallelDomain(jobPath)
		if !flag {
			jobdetectionmanager.SetDetectionExitedStatusClusterLevel(conf.JobName, true)
			return
		}
		hwlog.RunLog.Infof("[SLOWNODE ALGO]cluster %s TP:%v PP:%v", conf.JobName, tpInfo, ppInfo)
		jobLevelDetectionA3(ppInfo, tpInfo, conf, recorder)
		endTime := time.Now().Unix()
		config.LoopDetectionIntervalCheckSwitch(endTime-startTime, conf.NsecondsOneDetection,
			conf.JobName, conf.DetectionLevel)
		detectionCount++
	}
}
