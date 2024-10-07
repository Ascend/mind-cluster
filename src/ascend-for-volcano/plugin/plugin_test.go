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
	"strings"

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
	testCardNamePre = "AscendTest-"
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

func fakeDeviceInfoCMDataByNode(testNode *api.NodeInfo, cmName string) *v1.ConfigMap {
	const testTime = 1657527526
	cmData := NodeDeviceInfoWithDevPlugin{
		DeviceInfo: NodeDeviceInfo{
			DeviceList: testNode.Node.Annotations,
			UpdateTime: testTime,
		},
		CheckCode: "964c4c25b14a59cccf88df7ea62797acba0c8ff608f31c8989ac01f5d2e75c6c",
	}
	var data = make(map[string]string, 1)
	cmDataStr, err := json.Marshal(cmData)
	if err != nil {
		return nil
	}
	data["DeviceInfoCfg"] = string(cmDataStr)
	var faultNPUConfigMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "kube-system",
		},
		Data: data,
	}
	return faultNPUConfigMap
}

func fakeDeviceInfoCMDataByMap(testNodeMap map[string]*api.NodeInfo, cmName string) *v1.ConfigMap {
	nodeName := strings.TrimPrefix(cmName, util.DevInfoPreName)
	testNode, ok := testNodeMap[nodeName]
	if !ok {
		fmt.Printf("%s no %s in nodeMap\n", util.SafePrint(cmName), util.SafePrint(nodeName))
		return nil
	}
	return fakeDeviceInfoCMDataByNode(testNode, cmName)
}

// FakeDeviceInfoCM fake the DeviceInfo from device-plugin configMap.
func FakeDeviceInfoCM(testNode interface{}, cmName string) (*v1.ConfigMap, error) {
	var deviceInfoConfigMap *v1.ConfigMap
	switch value := testNode.(type) {
	case *api.NodeInfo:
		deviceInfoConfigMap = fakeDeviceInfoCMDataByNode(value, cmName)
	case map[string]*api.NodeInfo:
		deviceInfoConfigMap = fakeDeviceInfoCMDataByMap(value, cmName)
	default:
		fmt.Printf("not support: %#v", value)
	}

	if deviceInfoConfigMap == nil {
		return nil, fmt.Errorf("no deviceInfoConfigMap")
	}
	return deviceInfoConfigMap, nil
}
