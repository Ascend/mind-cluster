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
// Package externalbridge for node and cluster level detection interact interface
package externalbridge

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/clusterlevel"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
	"ascend-faultdiag-online/pkg/algo_src/slownode/nodelevel"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

func errorStationHandle(jobName string, level string) {
	if level == string(enum.Cluster) {
		jobdetectionmanager.SetDetectionLoopStatusClusterLevel(jobName, false)
		jobdetectionmanager.DeleteDetectionClusterLevel(jobName)
	} else if level == string(enum.Node) {
		jobdetectionmanager.SetDetectionLoopStatusNodeLevel(jobName, false)
		jobdetectionmanager.DeleteDetectionNodeLevel(jobName)
	}
}

func curDetectionCondSignal(jobName string, level string) {
	var lock *sync.Mutex
	var cond *sync.Cond
	if level == string(enum.Cluster) {
		lock = jobdetectionmanager.GetDetectionCondLockClusterLevel(jobName)
		cond = jobdetectionmanager.GetDetectionCondClusterLevel(jobName)
		if lock == nil || cond == nil {
			jobdetectionmanager.DeleteDetectionClusterLevel(jobName)
			hwlog.RunLog.Infof("[SLOWNODE ALGO]%s %s slow node detection exit!", jobName, level)
			return
		}

	} else if level == string(enum.Node) {
		lock = jobdetectionmanager.GetDetectionCondLockNodeLevel(jobName)
		cond = jobdetectionmanager.GetDetectionCondNodeLevel(jobName)
		if lock == nil || cond == nil {
			jobdetectionmanager.DeleteDetectionNodeLevel(jobName)
			hwlog.RunLog.Infof("[SLOWNODE ALGO]%s %s slow node detection exit!", jobName, level)
			return
		}
	}
	if lock != nil && cond != nil {
		lock.Lock()
		cond.Signal()
		lock.Unlock()
	}
}

func curDetectionCondWait(jobName string, target enum.DeployMode) {
	var lock *sync.Mutex
	var cond *sync.Cond
	if target == enum.Cluster {
		/* 若已异常退出,则不调用wait等待 */
		if jobdetectionmanager.GetDetectionExitedStatusClusterLevel(jobName) {
			return
		}
		lock = jobdetectionmanager.GetDetectionCondLockClusterLevel(jobName)
		cond = jobdetectionmanager.GetDetectionCondClusterLevel(jobName)
		if lock == nil || cond == nil {
			jobdetectionmanager.DeleteDetectionClusterLevel(jobName)
			hwlog.RunLog.Infof("[SLOWNODE ALGO]%s %v slow node detection exit!", jobName, target)
			return
		}
	} else if target == enum.Node {
		/* 若已异常退出,则不调用wait等待 */
		if jobdetectionmanager.GetDetectionExitedStatusNodeLevel(jobName) {
			return
		}
		lock = jobdetectionmanager.GetDetectionCondLockNodeLevel(jobName)
		cond = jobdetectionmanager.GetDetectionCondNodeLevel(jobName)
		if lock == nil || cond == nil {
			jobdetectionmanager.DeleteDetectionNodeLevel(jobName)
			hwlog.RunLog.Infof("[SLOWNODE ALGO]%s %v slow node detection exit!", jobName, target)
			return
		}
	}
	if lock != nil && cond != nil {
		lock.Lock()
		cond.Wait()
		lock.Unlock()
	}
}

// StartAlgo slow node detection by config
func StartAlgo(conf config.AlgoInputConfig) {
	if conf.DetectionLevel == string(enum.Cluster) && !jobdetectionmanager.AddDetectionClusterLevel(conf.JobName) {
		return
	}
	if conf.DetectionLevel == string(enum.Node) && !jobdetectionmanager.AddDetectionNodeLevel(conf.JobName) {
		return
	}
	/* check validation */
	if !checkConfigValid(conf) {
		errorStationHandle(conf.JobName, conf.DetectionLevel)
		return
	}
	/* 检测当前任务是否开启了CP或EP, 若开启了EP或CP则不进行检测 */
	if conf.DetectionLevel == string(enum.Node) && config.CheckCurJobOpenCpOrEp(conf, string(enum.Node)) {
		errorStationHandle(conf.JobName, conf.DetectionLevel)
		return
	}
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%s start %s detection", conf.JobName, conf.DetectionLevel)
	if conf.DetectionLevel == string(enum.Node) {
		/* node级检测 */
		nodelevel.NodeJobLevelDetectionLoopA3(conf)
	} else {
		/* cluster级检测 */
		clusterlevel.ClusterJobLevelDetectionLoopA3(conf)
	}
	/* exit */
	curDetectionCondSignal(conf.JobName, conf.DetectionLevel)
	hwlog.RunLog.Info("[SLOWNODE ALGO]slow node detection completed")
}

// StopAlgo detection
func StopAlgo(jobName string, target enum.DeployMode) {
	/* 修改flag */
	if target == enum.Cluster {
		jobdetectionmanager.SetDetectionLoopStatusClusterLevel(jobName, false)
		curDetectionCondWait(jobName, target)
		jobdetectionmanager.DeleteDetectionClusterLevel(jobName)
		hwlog.RunLog.Infof("[SLOWNODE ALGO]%v target %s slownode detection stopped", target, jobName)
	} else if target == enum.Node {
		jobdetectionmanager.SetDetectionLoopStatusNodeLevel(jobName, false)
		curDetectionCondWait(jobName, target)
		jobdetectionmanager.DeleteDetectionNodeLevel(jobName)
		hwlog.RunLog.Infof("[SLOWNODE ALGO]%v target %s slownode detection stopped", target, jobName)
	}
}

// ReloadAlgo to update detection algorithm config
func ReloadAlgo(configData config.AlgoInputConfig) {
	if !checkConfigValid(configData) {
		return
	}
	/* 若reload之前未启动，则reload不启动 */
	if configData.DetectionLevel == string(enum.Cluster) &&
		!jobdetectionmanager.CheckDetectionClusterJobExist(configData.JobName) {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]cluster level slownode detection",
			configData.JobName, " not started before(forbidden reload)")
		return
	} else if configData.DetectionLevel == string(enum.Node) &&
		!jobdetectionmanager.CheckDetectionNodeJobExist(configData.JobName) {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]node level slownode detection",
			configData.JobName, " not started before(forbidden reload)")
		return
	}
	StopAlgo(configData.JobName, enum.DeployMode(configData.DetectionLevel))
	StartAlgo(configData)
}

// RegisterAlgo function address by uintptr
func RegisterAlgo(externalFunc model.CallbackFunc, target enum.DeployMode) {
	/* register cluster and node data callback */
	hwlog.RunLog.Infof("[SLOWNODE ALGO]%v slownode detection register callback: %v", target, externalFunc)
	switch target {
	case enum.Cluster:
		clusterlevel.RegisterClusterLevelCallback(externalFunc)
	case enum.Node:
		nodelevel.RegisterNodeLevelCallback(externalFunc)
	default:
		hwlog.RunLog.Infof("[SLOWNODE ALGO] unsupported target: %v", target)
		return
	}
}

var startOrReloadAlgo = map[enum.Command]func(config.AlgoInputConfig){
	enum.Start:  StartAlgo,
	enum.Reload: ReloadAlgo,
}

var startOrReloadDataParse = map[enum.Command]func(config.DataParseModel){
	enum.Start:  StartDataParse,
	enum.Reload: ReloadDataParse,
}

func callAlgoStart(inputData *model.Input) bool {
	cg := config.AlgoInputConfig{}
	if err := transformJsonToStruct(inputData, &cg); err != nil {
		hwlog.RunLog.Errorf("model parse error: %v", err)
		return false
	}
	cg.DetectionLevel = string(inputData.Target)
	go startOrReloadAlgo[inputData.Command](cg)
	return true
}

func callAlgoStop(inputData *model.Input) bool {
	cg := config.AlgoInputConfig{}
	if err := transformJsonToStruct(inputData, &cg); err != nil {
		hwlog.RunLog.Errorf("model parse error: %v", err)
		return false
	}
	cg.DetectionLevel = string(inputData.Target)
	go StopAlgo(cg.JobName, inputData.Target)
	return true
}

// Execute for uniform interface
func Execute(inputData *model.Input) int {
	if !checkInvalidInput(inputData) {
		return -1
	}

	// after the data check, all the data are valid
	switch inputData.Command {
	case enum.Start, enum.Reload:
		return startAndReload(inputData)
	case enum.Stop:
		if inputData.EventType == enum.SlowNodeAlgo {
			if !callAlgoStop(inputData) {
				return -1
			}
		} else if inputData.EventType == enum.DataParse {
			cg := config.DataParseModel{}
			if err := transformJsonToStruct(inputData, &cg); err != nil {
				hwlog.RunLog.Errorf("model parse error: %v", err)
				return -1
			}
			go StopDataParse(cg)
		}
	case enum.Register:
		if inputData.EventType == enum.SlowNodeAlgo {
			go RegisterAlgo(inputData.Func, inputData.Target)
		} else if inputData.EventType == enum.DataParse {
			if inputData.Target == enum.Cluster {
				go RegisterMergeParGroup(inputData.Func)
			} else {
				go RegisterDataParse(inputData.Func)
			}
		}
	default:
		hwlog.RunLog.Error("command is not in input or not a string or invalid.")
		return -1
	}
	return 0
}

func startAndReload(inputData *model.Input) int {
	if inputData.EventType == enum.SlowNodeAlgo {
		if !callAlgoStart(inputData) {
			return -1
		}
	} else {
		cg := config.DataParseModel{}
		if err := transformJsonToStruct(inputData, &cg); err != nil {
			hwlog.RunLog.Errorf("model parse error: %v", err)
			return -1
		}
		if inputData.Target == enum.Cluster {
			go StartMergeParGroupInfo(cg)
		} else {
			cg.JobStartTime = time.Now().Unix()
			go startOrReloadDataParse[inputData.Command](cg)
		}
	}
	return 0
}
