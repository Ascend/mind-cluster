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

// Package jobdetectionmanager for node and cluster level detection interact interface
package jobdetectionmanager

import (
	"sync"

	"ascend-common/common-utils/hwlog"
)

/* 记录数据 */
var detectionDataRecorderCluster = make(map[string]JobDetectionFlag)

var dataRecorderSyncLockCluster sync.RWMutex

// CheckDetectionClusterJobExist 检测任务是否存在
func CheckDetectionClusterJobExist(jobName string) bool {
	dataRecorderSyncLockCluster.RLock()
	defer dataRecorderSyncLockCluster.RUnlock()
	if _, exist := detectionDataRecorderCluster[jobName]; exist {
		return true
	}
	return false
}

// AddDetectionClusterLevel 添加Job级检查(若已存在返回false)
func AddDetectionClusterLevel(jobName string) bool {
	dataRecorderSyncLockCluster.Lock()
	defer dataRecorderSyncLockCluster.Unlock()
	if _, exist := detectionDataRecorderCluster[jobName]; exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s detection already started or not exited!", jobName)
		return false
	}
	loopFlag := new(bool)
	*loopFlag = true
	errorFlag := new(bool)
	*errorFlag = false
	lock := new(sync.Mutex)
	cond := sync.NewCond(lock)
	detectionDataRecorderCluster[jobName] = JobDetectionFlag{
		DetectionExitedCond:  cond,
		DetectionCondLock:    lock,
		DetectionLoopFlag:    loopFlag,
		DetectionErrorExited: errorFlag,
	}
	hwlog.RunLog.Infof("[SLOWNODE ALGO]Add %s cluster level detection", jobName)
	return true
}

// DeleteDetectionClusterLevel 删除Job级检查（停止任务先设置为false之后再调用否则失败）
func DeleteDetectionClusterLevel(jobName string) {
	dataRecorderSyncLockCluster.Lock()
	defer dataRecorderSyncLockCluster.Unlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection info not exist!", jobName)
		return
	}
	if *(detectionDataRecorderCluster[jobName].DetectionLoopFlag) {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection not stopped!", jobName)
		return
	}
	delete(detectionDataRecorderCluster, jobName)
}

// GetDetectionLoopStatusClusterLevel 获取Job级检测flag
func GetDetectionLoopStatusClusterLevel(jobName string) bool {
	dataRecorderSyncLockCluster.RLock()
	defer dataRecorderSyncLockCluster.RUnlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection loop status not exist!", jobName)
		return false
	}
	flag := *(detectionDataRecorderCluster[jobName].DetectionLoopFlag)
	return flag
}

// SetDetectionLoopStatusClusterLevel 设置flag
func SetDetectionLoopStatusClusterLevel(jobName string, status bool) {
	dataRecorderSyncLockCluster.Lock()
	defer dataRecorderSyncLockCluster.Unlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection loop status not exist!", jobName)
		return
	}
	flag := detectionDataRecorderCluster[jobName].DetectionLoopFlag
	*flag = status
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%s loop status set flag:%v", jobName, status)
}

// GetDetectionCondClusterLevel 获取job条件变量
func GetDetectionCondClusterLevel(jobName string) *sync.Cond {
	dataRecorderSyncLockCluster.RLock()
	defer dataRecorderSyncLockCluster.RUnlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection cond not exist!", jobName)
		return nil
	}
	return detectionDataRecorderCluster[jobName].DetectionExitedCond
}

// GetDetectionCondLockClusterLevel 获取job条件变量锁
func GetDetectionCondLockClusterLevel(jobName string) *sync.Mutex {
	dataRecorderSyncLockCluster.RLock()
	defer dataRecorderSyncLockCluster.RUnlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection lock not exist!", jobName)
		return nil
	}
	return detectionDataRecorderCluster[jobName].DetectionCondLock
}

// GetDetectionExitedStatusClusterLevel 获取flag判断是否存在异常退出
func GetDetectionExitedStatusClusterLevel(jobName string) bool {
	dataRecorderSyncLockCluster.RLock()
	defer dataRecorderSyncLockCluster.RUnlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection exited status not exist!", jobName)
		return false
	}
	flag := *(detectionDataRecorderCluster[jobName].DetectionErrorExited)
	return flag
}

// SetDetectionExitedStatusClusterLevel 设置循环过程中异常退出flag
func SetDetectionExitedStatusClusterLevel(jobName string, status bool) {
	dataRecorderSyncLockCluster.Lock()
	defer dataRecorderSyncLockCluster.Unlock()
	if _, exist := detectionDataRecorderCluster[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection exited status not exist!", jobName)
		return
	}
	flag := detectionDataRecorderCluster[jobName].DetectionErrorExited
	*flag = status
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%s exit status set flag:%v", jobName, status)
}
