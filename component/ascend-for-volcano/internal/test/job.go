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
Package test is using for HuaWei Ascend testing.
*/
package test

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	fakeNodeName        = "node0"
	annoCards           = "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	networkUnhealthyNPU = "huawei.com/Ascend910-NetworkUnhealthy"
	unhealthyNPU        = "huawei.com/Ascend910-Unhealthy"
)

func fakeDeviceInfoCMDataByNode(nodeName string, deviceList map[string]string) *v1.ConfigMap {
	cmName := util.DevInfoPreName + nodeName
	const testTime = 1657527526
	cmData := plugin.NodeDeviceInfoWithDevPlugin{
		DeviceInfo: plugin.NodeDeviceInfo{
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
	tmpSwitchInfo := plugin.SwitchFaultInfo{
		NodeStatus: util.NodeHealthyByNodeD,
	}
	if bytes, err := json.Marshal(tmpSwitchInfo); err == nil {
		return string(bytes)
	}
	return ""
}

// FakeSchedulerJobAttrByJob fake scheduler attr by job
func FakeSchedulerJobAttrByJob(job *api.JobInfo) util.SchedulerJobAttr {
	attr := util.SchedulerJobAttr{
		ComJob: util.ComJob{
			Name:      job.UID,
			NameSpace: job.Namespace,
			Selector:  nil,
			Label:     nil,
		},
	}
	name, num, err := plugin.GetVCJobReqNPUTypeFromJobInfo(job)
	if err != nil {
		return attr
	}
	NPUJob := &util.NPUJob{
		ReqNPUName: name,
		ReqNPUNum:  num,
		Tasks:      plugin.GetJobNPUTasks(job),
	}
	NPUJob.NPUTaskNum = NPUJob.GetNPUTaskNumInJob()
	attr.NPUJob = NPUJob
	return attr
}

// NewDefaultHandler new default handler
func NewDefaultHandler() *plugin.ScheduleHandler {
	isFirstSession := true
	scheduleHandler := &plugin.ScheduleHandler{
		NPUPlugins: map[string]plugin.NPUBuilder{},
		ScheduleEnv: plugin.ScheduleEnv{
			IsFirstSession:   &isFirstSession,
			Jobs:             map[api.JobID]plugin.SchedulerJob{},
			JobSeverInfos:    map[api.JobID]struct{}{},
			JobDeleteFlag:    map[api.JobID]struct{}{},
			JobSinglePodFlag: map[api.JobID]bool{},
			Nodes:            map[string]plugin.NPUNode{},
			DeviceInfos: &plugin.DeviceInfosWithMutex{
				Mutex:   sync.Mutex{},
				Devices: map[string]plugin.NodeDeviceInfoWithID{},
			},
			NodeInfosFromCm: &plugin.NodeInfosFromCmWithMutex{
				Mutex: sync.Mutex{},
				Nodes: map[string]plugin.NodeDNodeInfo{},
			},
			SwitchInfosFromCm: &plugin.SwitchInfosFromCmWithMutex{
				Mutex:    sync.Mutex{},
				Switches: map[string]plugin.SwitchFaultInfo{},
			},
			FrameAttr: plugin.VolcanoFrame{},
			NslbAttr:  &plugin.NslbParameters{},
			SuperPodInfo: &plugin.SuperPodInfo{
				SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
				SuperPodFaultTaskNodes:    map[api.JobID][]string{},
				SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{},
			},
			JobPendingMessage: map[api.JobID]map[string]map[string]struct{}{},
		},
	}
	scheduleHandler.NPUPlugins[util.NPU910CardName] = func(string2 string) plugin.ISchedulerPlugin {
		return nil
	}
	return scheduleHandler
}

// FakeScheduleEnv fake normal schedule env
func FakeScheduleEnv() *plugin.ScheduleEnv {
	sHandle := NewDefaultHandler()
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	for _, jobInfo := range ssn.Jobs {
		test.SetJobStatusRunning(jobInfo)
	}
	for _, node := range ssn.Nodes {
		sHandle.UpdateConfigMap(fakeDeviceInfoCMDataByNode(node.Name, fakeDeviceList()), util.AddOperator)
	}
	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)
	sHandle.ScheduleEnv.FrameAttr.KubeClient = fake.NewSimpleClientset()
	node0 := sHandle.ScheduleEnv.Nodes[fakeNodeName]
	node0.Annotation[unhealthyNPU] = annoCards
	sHandle.ScheduleEnv.Nodes[fakeNodeName] = node0
	return &sHandle.ScheduleEnv
}

// PatchGetCm go monkey patch get cm
func PatchGetCm(name, nameSpace string, data map[string]string) *gomonkey.Patches {
	return gomonkey.ApplyFunc(k8s.GetConfigMap, func(client kubernetes.Interface, namespace, cmName string) (
		*v1.ConfigMap, error) {
		return test.FakeConfigmap(name, nameSpace, data), nil
	})
}

// InitNormalsHandlerBySsnFunc init normal sHandler by ssn func
func InitNormalsHandlerBySsnFunc(ssn *framework.Session, initSsnFunc ...func(ssn *framework.Session)) {
	if ssn == nil {
		return
	}
	for _, initFunc := range initSsnFunc {
		initFunc(ssn)
	}
}

// FakeTorNodeData Fake tor node date for
func FakeTorNodeData() map[string]string {
	torNodeDataPath := "../../testdata/tor/tor-node.json"
	torInfoCMKey := "tor_info"
	data, err := os.ReadFile(torNodeDataPath)
	if err != nil {
		return nil
	}
	return map[string]string{
		torInfoCMKey: string(data),
	}
}
