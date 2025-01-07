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
Package test is using for HuaWei Ascend pin scheduling test.
*/
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	testPluginName = "testPlugin"
	// testCardName test card
	testCardName = "huawei.com/AscendTest"
	// testCardNamePre for getting test card number.
	testCardNamePre     = "AscendTest-"
	annoCards           = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	unhealthyNPU        = "huawei.com/Ascend910-Unhealthy"
)

type ascendTest struct {
	// need plugin
	SchedulerPlugin
	// env
	ScheduleEnv
	// job's attribute
	util.SchedulerJobAttr
}

// New return npu plugin.
func New(npuName string) ISchedulerPlugin {
	var npuPlugin = &ascendTest{}
	npuPlugin.SetPluginName(npuName)
	npuPlugin.SetAnnoName(testCardName)
	npuPlugin.SetAnnoPreVal(testCardNamePre)
	npuPlugin.SetDefaultJobSchedulerConfig(nil)

	return npuPlugin
}

// Name This need by frame init plugin.
func (tp *ascendTest) Name() string {
	return PluginName
}

func (tp *ascendTest) InitMyJobPlugin(attr util.SchedulerJobAttr, env ScheduleEnv) error {
	fmt.Printf("enter %s InitMyJobPlugin", util.NPU910CardName)
	if tp == nil {
		mgs := fmt.Errorf("nil plugin %s", PluginName)
		fmt.Printf("InitMyJobPlugin %s.", util.SafePrint(mgs))
		return mgs
	}
	tp.SchedulerJobAttr = attr
	tp.ScheduleEnv = env

	fmt.Printf("leave %s InitMyJobPlugin", util.NPU910CardName)
	return nil
}

func (tp *ascendTest) ValidNPUJob() *api.ValidateResult {
	if tp == nil {
		err := errors.New(util.ArgumentError)
		return &api.ValidateResult{
			Pass:    false,
			Reason:  err.Error(),
			Message: err.Error(),
		}
	}
	return nil
}

func (tp *ascendTest) GetReHandle() interface{} {
	return nil
}

func (tp *ascendTest) CheckNodeNPUByTask(task *api.TaskInfo, node NPUNode) error {
	if tp == nil || task == nil || len(node.Annotation) == 0 {
		return errors.New(util.ArgumentError)
	}
	return nil
}

func (tp *ascendTest) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, scoreMap map[string]float64) error {
	return nil
}

func (tp *ascendTest) UseAnnotation(task *api.TaskInfo, node NPUNode) *NPUNode {
	return nil
}

func (tp *ascendTest) ReleaseAnnotation(task *api.TaskInfo, node NPUNode) *NPUNode {
	return nil
}

func (tp *ascendTest) PreStartAction(i interface{}, ssn *framework.Session) error {
	if tp == nil {
		return fmt.Errorf(util.ArgumentError)
	}

	return nil
}

func (tp *ascendTest) PreStopAction(env *ScheduleEnv) error {
	if tp == nil {
		return fmt.Errorf(util.ArgumentError)
	}

	return nil
}

func fakeDeviceInfoCMDataByNode(nodeName string, deviceList map[string]string) *v1.ConfigMap {
	cmName := util.DevInfoPreName + nodeName
	const testTime = 1657527526
	cmData := NodeDeviceInfoWithDevPlugin{
		DeviceInfo: NodeDeviceInfo{
			DeviceList: deviceList,
			UpdateTime: testTime,
		},
		CheckCode: "6b8de396fd9945be231d24720ca66ed950baf0a5972717f335aad7571cb6457a",
	}
	var data = make(map[string]string, 1)
	cmDataStr, err := json.Marshal(cmData)
	if err != nil {
		return nil
	}
	data["DeviceInfoCfg"] = string(cmDataStr)
	data[util.SwitchInfoCmKey] = fakeSwitchInfos()
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "kube-system",
		},
		Data: data,
	}
	return faultNPUConfigMap
}

func fakeDeviceList() map[string]string {
	return map[string]string{
		util.NPU910CardName: annoCards,
		networkUnhealthyNPU: "",
		unhealthyNPU:        "",
	}
}

func fakeSwitchInfos() string {
	tmpSwitchInfo := SwitchFaultInfo{
		NodeStatus: util.NodeHealthyByNodeD,
	}
	if bytes, err := json.Marshal(tmpSwitchInfo); err == nil {
		return string(bytes)
	}
	return ""
}

func fakeNodeInfos() map[string]string {
	nodeInfos := NodeDNodeInfo{
		NodeStatus: util.NodeHealthyByNodeD,
	}
	tmpData := NodeInfoWithNodeD{
		NodeInfo:  nodeInfos,
		CheckCode: util.MakeDataHash(nodeInfos),
	}

	nodeInfoBytes, err := json.Marshal(tmpData)
	if err != nil {
		return nil
	}
	return map[string]string{
		util.NodeInfoCMKey: string(nodeInfoBytes),
	}
}

func fakeResetCmInfos() map[string]string {
	resetInfos := TaskResetInfo{}

	resetInfosBytes, err := json.Marshal(resetInfos)
	if err != nil {
		return nil
	}
	return map[string]string{
		ResetInfoCMDataKey: string(resetInfosBytes),
	}
}
