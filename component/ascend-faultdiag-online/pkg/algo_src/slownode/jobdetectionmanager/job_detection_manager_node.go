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

// JobDetectionFlag 控制job检测启动与退出标志
type JobDetectionFlag struct {
	// DetectionExitCond 退出条件变量
	DetectionExitedCond *sync.Cond
	// DetectionCondLock 条件变量锁
	DetectionCondLock *sync.Mutex
	// DetectionLoopFlag 循环检测flag
	DetectionLoopFlag *bool
	// DetectionExited 循环检测过程中错误退出
	DetectionErrorExited *bool
}

var detectionRecorderNode = make(map[string]JobDetectionFlag)

var recorderSyncLockNode sync.RWMutex

// CheckDetectionNodeJobExist 检测job级任务是否存在
func CheckDetectionNodeJobExist(jobName string) bool {
	recorderSyncLockNode.RLock()
	defer recorderSyncLockNode.RUnlock()
	if _, exist := detectionRecorderNode[jobName]; exist {
		return true
	}
	return false
}

// AddDetectionNodeLevel 添加Job级检查(若已存在返回false)
func AddDetectionNodeLevel(jobName string) bool {
	recorderSyncLockNode.Lock()
	defer recorderSyncLockNode.Unlock()
	if _, exist := detectionRecorderNode[jobName]; exist {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s detection already started or not exited!", jobName)
		return false
	}
	loopFlag := new(bool)
	*loopFlag = true
	errorFlag := new(bool)
	*errorFlag = false
	lock := new(sync.Mutex)
	cond := sync.NewCond(lock)
	detectionRecorderNode[jobName] = JobDetectionFlag{
		DetectionExitedCond:  cond,
		DetectionCondLock:    lock,
		DetectionLoopFlag:    loopFlag,
		DetectionErrorExited: errorFlag,
	}
	hwlog.RunLog.Infof("[SLOWNODE ALGO]Add %s node level detection", jobName)
	return true
}

// DeleteDetectionNodeLevel 删除Job级检查（停止任务先设置为false之后再调用否则失败）
func DeleteDetectionNodeLevel(jobName string) {
	recorderSyncLockNode.Lock()
	defer recorderSyncLockNode.Unlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection info not exist!", jobName)
		return
	}
	if *(detectionRecorderNode[jobName].DetectionLoopFlag) {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection not stopped!", jobName)
		return
	}
	delete(detectionRecorderNode, jobName)
}

// GetDetectionLoopStatusNodeLevel 获取Job级检测flag
func GetDetectionLoopStatusNodeLevel(jobName string) bool {
	recorderSyncLockNode.RLock()
	defer recorderSyncLockNode.RUnlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection loop status not exist!", jobName)
		return false
	}
	flag := *(detectionRecorderNode[jobName].DetectionLoopFlag)
	return flag
}

// SetDetectionLoopStatusNodeLevel 设置flag
func SetDetectionLoopStatusNodeLevel(jobName string, status bool) {
	recorderSyncLockNode.Lock()
	defer recorderSyncLockNode.Unlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection loop status not exist!", jobName)
		return
	}
	flag := detectionRecorderNode[jobName].DetectionLoopFlag
	*flag = status
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%s loop status set flag:%v", jobName, status)
}

// GetDetectionCondNodeLevel 获取job条件变量
func GetDetectionCondNodeLevel(jobName string) *sync.Cond {
	recorderSyncLockNode.RLock()
	defer recorderSyncLockNode.RUnlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection cond not exist!", jobName)
		return nil
	}
	return detectionRecorderNode[jobName].DetectionExitedCond
}

// GetDetectionCondLockNodeLevel 获取job条件变量锁
func GetDetectionCondLockNodeLevel(jobName string) *sync.Mutex {
	recorderSyncLockNode.RLock()
	defer recorderSyncLockNode.RUnlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection lock not exist!", jobName)
		return nil
	}
	return detectionRecorderNode[jobName].DetectionCondLock
}

// GetDetectionExitedStatusNodeLevel 获取flag判断是否存在异常退出
func GetDetectionExitedStatusNodeLevel(jobName string) bool {
	recorderSyncLockNode.RLock()
	defer recorderSyncLockNode.RUnlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection exited status not exist!", jobName)
		return false
	}
	flag := *(detectionRecorderNode[jobName].DetectionErrorExited)
	return flag
}

// SetDetectionExitedStatusNodeLevel 设置循环过程中异常退出flag
func SetDetectionExitedStatusNodeLevel(jobName string, status bool) {
	recorderSyncLockNode.Lock()
	defer recorderSyncLockNode.Unlock()
	if _, exist := detectionRecorderNode[jobName]; !exist {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s detection exited status not exist!", jobName)
		return
	}
	flag := detectionRecorderNode[jobName].DetectionErrorExited
	*flag = status
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%s exit status set flag:%v", jobName, status)
}
