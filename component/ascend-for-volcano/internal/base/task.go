/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetTaskReqNPUNum get task require npu num
func (tp *NPUHandler) GetTaskReqNPUNum(task *api.TaskInfo) (int, error) {
	if tp == nil || task == nil {
		return 0, errors.New(util.ArgumentError)
	}
	nJob, jOK := tp.Jobs[task.Job]
	if !jOK {
		err := fmt.Errorf("%s is not npu job", task.Job)
		klog.V(util.LogErrorLev).Infof("GetTaskReqNPUNum err: %s,%s,%#v", err, util.SafePrint(task.Job), tp.Jobs)
		return 0, err
	}
	nTask, tOK := nJob.Tasks[task.UID]
	if !tOK {
		err := fmt.Errorf("task<%s> is not npu task", task.Name)
		klog.V(util.LogErrorLev).Infof("GetTaskReqNPUNum err: %s,%s,%#v", err, util.SafePrint(task.UID), tp.Tasks)
		return 0, err
	}
	klog.V(util.LogDebugLev).Infof("GetTaskReqNPUNum task req npu<%s>-<%d> ", nTask.ReqNPUName, nTask.ReqNPUNum)
	return nTask.ReqNPUNum, nil
}

// SetNPUTopologyToPodFn set task select npu to pod annotation
func (tp *NPUHandler) SetNPUTopologyToPodFn(task *api.TaskInfo, top []int) {
	if tp == nil || task == nil || task.Pod == nil || task.Pod.Annotations == nil || len(top) == 0 {
		return
	}
	topologyStr := util.ChangeIntArrToStr(top, tp.GetAnnoPreVal())
	task.Pod.Annotations[tp.GetAnnoName()] = topologyStr
	// to device-plugin judge pending pod.
	tmp := strconv.FormatInt(time.Now().UnixNano(), util.Base10)
	task.Pod.Annotations[util.PodPredicateTime] = tmp
	klog.V(util.LogInfoLev).Infof("%s setNPUTopologyToPod %s==%v top:%s.", tp.GetPluginName(),
		task.Name, tmp, topologyStr)
}
