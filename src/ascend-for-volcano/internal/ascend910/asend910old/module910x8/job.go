/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package module910x8 is using for HuaWei Ascend pin affinity schedule.
*/
package module910x8

import (
	"fmt"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (tp *module910x8) checkJobTrainMode() error {
	if tp.NPUTaskNum == 0 {
		klog.V(util.LogErrorLev).Infof("GetVTaskNumInVJob %s has no npu tasks.", tp.Name)
		return fmt.Errorf("%s no npu job", tp.Name)
	}
	if tp.NPUTaskNum == 1 {
		if err := tp.checkSingleTrainMode(); err != nil {
			klog.V(util.LogErrorLev).Infof("%s checkSingleTrainMode %s: %s", tp.GetPluginName(), tp.Name, err)
			return err
		}
		return nil
	}

	if err := tp.checkModuleDistributeTrainMode(); err != nil {
		klog.V(util.LogErrorLev).Infof("%s check distribute %s err: %s", tp.GetPluginName(), tp.Name, err)
		return err
	}

	return nil
}

// CheckSingleTrainMode Single Train job has only one task.
func (tp *module910x8) checkSingleTrainMode() error {
	klog.V(util.LogDebugLev).Infof("checkSingleTrainMode job(%s) has %d tasks.", tp.Name, len(tp.Tasks))

	jobNPU := tp.ReqNPUNum
	if jobNPU == 1 || jobNPU == util.NPUIndex2 || jobNPU == util.NPUIndex4 || jobNPU == util.NPUIndex8 {
		return nil
	}
	return fmt.Errorf("%s checkSingleTrainMode %s req npu not in [1,2,4,8]", tp.GetPluginName(), tp.Name)
}

// checkModuleDistributeTrainMode Distribute Train job has more than one task.
func (tp *module910x8) checkModuleDistributeTrainMode() error {
	klog.V(util.LogDebugLev).Infof("half DistributeTrainMode %s has %d tasks.", tp.Name, len(tp.Tasks))

	for _, task := range tp.Tasks {
		if !task.IsNPUTask() {
			continue
		}

		if task.ReqNPUNum == util.NPUIndex1 || task.ReqNPUNum == util.NPUIndex2 || task.ReqNPUNum == util.NPUIndex4 ||
			task.ReqNPUNum == tp.MaxNodeNPUNum {
			return nil
		}

		return fmt.Errorf("checkModuleDistributeTrainMode %s req %d not in [1,2,4,8]", task.Name, task.ReqNPUNum)
	}
	return nil
}
