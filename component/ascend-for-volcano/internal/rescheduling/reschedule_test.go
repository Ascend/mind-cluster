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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	fakeTime        = 123455
	fakeTime2       = 11111
	createTime      = 10000
	graceDeleteTime = 900
	zero            = 0
	one             = 1
	two             = 2
	three           = 3
	mockJobName1    = "jobName1"
	mockJobName2    = "jobName2"
)

func fakeTestFaultCardUnhealthy(name string, nodeName string, faultType string) *FaultCard {
	return &FaultCard{
		IsFaultCard: true,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
}

func fakeTestFaultCardHealthy(name string, nodeName string) *FaultCard {
	return &FaultCard{
		IsFaultCard: false,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   CardHealthy,
	}
}

func fakeTestFaultCardsUnhealthy(nodeName string, isUnHealth bool) []FaultCard {
	var card0 *FaultCard
	if !isUnHealth {
		card0 = fakeTestFaultCardUnhealthy("Ascend910-0", nodeName, CardUnhealthy)
	} else {
		card0 = fakeTestFaultCardHealthy("Ascend910-0", nodeName)
	}
	cards := []FaultCard{
		*card0,
		*fakeTestFaultCardHealthy("Ascend910-1", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-2", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-3", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-4", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-5", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-6", nodeName),
		*fakeTestFaultCardHealthy("Ascend910-7", nodeName),
	}
	return cards
}

func fakeTestFaultNodeCardUnhealthy(nodeName string, allCard []string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, true)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        []string{"Ascend910-0"},
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeCardUnhealthy,
		AllCards:            allCard,
		FaultCards:          faultCards,
	}
}

func fakeTestFaultNodeNodeUnhealthy(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, false)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeUnhealthy,
		AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
			"Ascend910-5", "Ascend910-6", "Ascend910-7"},
		FaultCards: faultCards,
	}
}

func fakeTestFaultNodeNodeHealthy(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, false)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         true,
		NodeHealthState:     NodeHealthy,
		AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
			"Ascend910-5", "Ascend910-6", "Ascend910-7"},
		FaultCards: faultCards,
	}
}

func fakeTestFaultNodeNodeHealthyOneCard(nodeName string) *FaultNode {
	updateTime := int64(fakeTime2)
	faultCards := []FaultCard{*fakeTestFaultCardHealthy("Ascend910-0", nodeName)}
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: nil,
		IsFaultNode:         false,
		NodeDEnable:         true,
		NodeHealthState:     NodeHealthy,
		AllCards:            []string{"Ascend910-0"},
		FaultCards:          faultCards,
	}
}

func fakeTestFaultTaskFault(name string, namespace string,
	nodeName string, nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   true,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(createTime),
		PodUID:        podUID,
	}
}

func fakeTestFaultTaskHealth(name string, namespace string, nodeName string,
	nodeRankIndex string, podUID types.UID) *FaultTask {
	return &FaultTask{
		IsFaultTask:   false,
		TaskName:      name,
		TaskNamespace: namespace,
		NodeName:      nodeName,
		NodeRankIndex: nodeRankIndex,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3"},
		PodCreateTime: int64(createTime),
		PodUID:        podUID,
	}
}

func fakeTestFaultJob(
	nodeNames []string, jobRankIds []string, faultTasks []FaultTask, jobName string, nameSpace string) *FaultJob {
	updateTime := int64(fakeTime2)
	return &FaultJob{
		ReScheduleKey:       JobGraceRescheduleLabelValue,
		IsFaultJob:          true,
		JobName:             jobName,
		JobUID:              api.JobID(nameSpace + `/` + jobName),
		JobNamespace:        nameSpace,
		JobRankIds:          jobRankIds,
		NodeNames:           nodeNames,
		FaultTasks:          faultTasks,
		UpdateTime:          updateTime,
		JobRankIdCreateTime: int64(createTime),
	}
}

func fakeReSchedulerCache() *DealReSchedulerCache {
	nodeNames := []string{"ubuntu1", "ubuntu2", "ubuntu3", "ubuntu4"}
	nodeRankIds := []string{"0", "1", "2", "3"}
	taskNames := []string{"task1", "task2", "task3", "task4"}
	jobRankIds := []string{"0", "10", "18", "27"}
	allCard := []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
		"Ascend910-5", "Ascend910-6", "Ascend910-7"}
	nameSpace := "vcjob"
	jobName := "job1"
	faultTasks := []FaultTask{
		*fakeTestFaultTaskHealth(taskNames[zero], nameSpace, nodeNames[zero], nodeRankIds[zero], "pod1"),
		*fakeTestFaultTaskFault(taskNames[one], nameSpace, nodeNames[one], nodeRankIds[one], "pod2"),
		*fakeTestFaultTaskHealth(taskNames[two], nameSpace, nodeNames[two], nodeRankIds[two], "pod3"),
		*fakeTestFaultTaskFault(taskNames[three], nameSpace, nodeNames[three], nodeRankIds[three], "pod4"),
	}
	return &DealReSchedulerCache{
		FaultNodes: []FaultNode{
			*fakeTestFaultNodeNodeHealthy(nodeNames[zero]),
			*fakeTestFaultNodeCardUnhealthy(nodeNames[one], allCard),
			*fakeTestFaultNodeNodeHealthy(nodeNames[two]),
			*fakeTestFaultNodeNodeUnhealthy(nodeNames[three]),
		},
		FaultJobs: []FaultJob{
			*fakeTestFaultJob(nodeNames, jobRankIds, faultTasks, jobName, nameSpace),
		},
		DealReSchedulerConfigmap: nil,
	}
}

func fakeNPUNodeNilDeviceInfo(name string) *plugin.NPUNode {
	nodeInfo := test.FakeNormalTestNode(name)
	return &plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       name,
			Capability: nodeInfo.Capability.ScalarResources,
			Allocate:   nodeInfo.Allocatable.ScalarResources,
			Idle:       nodeInfo.Idle.ScalarResources,
			Annotation: nodeInfo.Node.Annotations,
			Label:      nodeInfo.Node.Labels,
		},
	}
}

func fakeNPUNodeWithDeviceInfo(name string) *plugin.NPUNode {
	anno := map[string]string{
		util.NPU910CardName:                              "Ascend910-0,Ascend910-1,Ascend910-2",
		util.NPU910CardName + "-" + CardUnhealthy:        "Ascend910-1",
		util.NPU910CardName + "-" + CardNetworkUnhealthy: "Ascend910-2",
	}
	nodeInfo := test.FakeNormalTestNode(name)
	npuNode := &plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       name,
			Capability: nodeInfo.Capability.ScalarResources,
			Allocate:   nodeInfo.Allocatable.ScalarResources,
			Idle:       nodeInfo.Idle.ScalarResources,
			Annotation: anno,
			Label:      nodeInfo.Node.Labels,
		},
	}
	return npuNode
}

type FaultReSchedulerGetGraceDeleteTimeFields struct {
	DealReSchedulerCache *DealReSchedulerCache
	GraceDeleteTime      int64
}

type FaultReSchedulerGetGraceDeleteTimeArgs struct {
	conf           []config.Configuration
	cacheFunBefore func()
	cacheFunAfter  func()
}

type FaultReSchedulerGetGraceDeleteTimeTests struct {
	name    string
	fields  FaultReSchedulerGetGraceDeleteTimeFields
	args    FaultReSchedulerGetGraceDeleteTimeArgs
	want    int64
	wantErr bool
}

func fakeSchedulerConfiguration(_ string, _ []conf.Configuration) *config.Configuration {
	schedulerConf := &config.Configuration{
		Name:      "util.CMInitParamKey",
		Arguments: map[string]string{GraceOverTimeKey: "800"},
	}
	return schedulerConf
}

func fakeSchedulerConfGraceOverTime() []config.Configuration {
	schedulerConf := []config.Configuration{
		{
			Name:      "util.CMInitParamKey",
			Arguments: map[string]string{GraceOverTimeKey: "800"},
		},
	}
	return schedulerConf
}

func buildFaultReSchedulerGetGraceDeleteArgs(conf2 []config.Configuration,
	tmpPatche *gomonkey.Patches) FaultReSchedulerGetGraceDeleteTimeArgs {
	args := FaultReSchedulerGetGraceDeleteTimeArgs{
		conf: conf2,
		cacheFunBefore: func() {
			tmpPatche = gomonkey.ApplyFunc(util.GetConfigFromSchedulerConfigMap, fakeSchedulerConfiguration)
		},
		cacheFunAfter: func() {
			if tmpPatche != nil {
				tmpPatche.Reset()
			}
		},
	}
	return args
}

func buildFaultReSchedulerGetGraceDeleteTestCases() []FaultReSchedulerGetGraceDeleteTimeTests {
	var tmpPatche *gomonkey.Patches
	conf2 := fakeSchedulerConfGraceOverTime()
	testCases := []FaultReSchedulerGetGraceDeleteTimeTests{
		{
			name: "01-test FaultReSchedulerGetGraceDelete-no config",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      graceDeleteTime,
				DealReSchedulerCache: fakeReSchedulerCache(),
			},
			args:    buildFaultReSchedulerGetGraceDeleteArgs(nil, tmpPatche),
			want:    graceDeleteTime,
			wantErr: true,
		},
		{
			name: "02-test FaultReSchedulerGetGraceDelete-succeed",
			fields: FaultReSchedulerGetGraceDeleteTimeFields{
				GraceDeleteTime:      graceDeleteTime,
				DealReSchedulerCache: fakeReSchedulerCache(),
			},
			args:    buildFaultReSchedulerGetGraceDeleteArgs(conf2, tmpPatche),
			want:    graceDeleteTime,
			wantErr: false,
		},
	}
	return testCases
}

// TestFaultReSchedulerGetGraceDeleteTime test for grace delete time
func TestFaultReSchedulerGetGraceDeleteTime(t *testing.T) {
	tests := buildFaultReSchedulerGetGraceDeleteTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := &ReScheduler{
				GraceDeleteTime:      tt.fields.GraceDeleteTime,
				DealReSchedulerCache: tt.fields.DealReSchedulerCache,
			}
			got, err := reScheduler.GetGraceDeleteTime(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGraceDeleteTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetGraceDeleteTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type ReSchedulerAddFaultJobWithSessionArgs struct {
	jobs        map[api.JobID]*api.JobInfo
	cardName    string
	cardPreName string
}

type ReSchedulerAddFaultJobWithSessionTests struct {
	fields  *ReScheduler
	name    string
	args    ReSchedulerAddFaultJobWithSessionArgs
	wantErr bool
}

func fakeCacheNoneFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := DealReSchedulerCache{
		FaultNodes: []FaultNode{
			{
				NodeName:            "node0",
				UpdateTime:          fakeTime,
				UnhealthyNPU:        []string{"Ascend910-0"},
				NetworkUnhealthyNPU: nil,
				IsFaultNode:         true,
				NodeDEnable:         true,
				NodeHealthState:     NodeCardUnhealthy,
				AllCards: []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3",
					"Ascend910-4", "Ascend910-5", "Ascend910-6", "Ascend910-7"},
				FaultCards: []FaultCard{
					*fakeTestFaultCardUnhealthy("Ascend910-0", "node0", NodeCardUnhealthy),
					*fakeTestFaultCardHealthy("Ascend910-1", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-2", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-3", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-4", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-5", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-6", "node0"),
					*fakeTestFaultCardHealthy("Ascend910-7", "node0"),
				},
			},
		},
		FaultJobs:                nil,
		DealReSchedulerConfigmap: nil,
	}
	return &reCache
}

func fakeFaultTask2P(ns string, name string, node string, job string, index string) FaultTask {
	fTask := FaultTask{
		IsFaultTask:   true,
		TaskUID:       api.TaskID(`"` + ns + `"` + `"` + name + `"`),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      node,
		JobName:       job,
		NodeRankIndex: index,
		UseCardName:   []string{"Ascend910-0", "Ascend910-1"},
		PodCreateTime: fakeTime,
		PodUID:        types.UID(`"` + ns + `"` + `"` + name + `"`),
	}
	return fTask
}

func fakeFaultJob() FaultJob {
	return FaultJob{
		ReScheduleKey: "grace",
		IsFaultJob:    true,
		IsInSession:   true,
		JobName:       "job0",
		JobUID:        "vcjob/job0",
		JobNamespace:  "test",
		JobRankIds:    []string{"0", "1", "8", "9"},
		NodeNames:     []string{"node0", "node1"},
		FaultTasks: []FaultTask{
			fakeFaultTask2P("vcjob", "pod0", "node0", "job0", "0"),
			fakeFaultTask2P("vcjob", "pod1", "node1", "job0", "1"),
		},
		UpdateTime:          test.FakeUpdateTime + 1,
		JobRankIdCreateTime: test.FakeUpdateTime,
	}
}

func fakeCacheWithFJobReSchedulerAddFaultJobWithSession() *DealReSchedulerCache {
	reCache := fakeCacheNoneFJobReSchedulerAddFaultJobWithSession()
	reCache.FaultJobs = []FaultJob{fakeFaultJob()}
	return reCache
}

func reAddFaultJobWithSessionModifyJobInfo(jobInfos map[api.JobID]*api.JobInfo) map[api.JobID]*api.JobInfo {
	jobInfos["vcjob/job0"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job1"].PodGroup.Labels = map[string]string{JobRescheduleLabelKey: "grace"}
	jobInfos["vcjob/job0"].Tasks[test.FakeTaskName1].Pod.Annotations =
		map[string]string{podRankIndex: "0", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job0"].Tasks[test.FakeTaskName1].Pod.Annotations =
		map[string]string{podRankIndex: "1", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[test.FakeTaskName1].Pod.Annotations =
		map[string]string{podRankIndex: "2", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[test.FakeTaskName1].Pod.Annotations =
		map[string]string{podRankIndex: "3", util.NPU910CardName: "Ascend910-0,Ascend910-1"}
	jobInfos["vcjob/job1"].Tasks[test.FakeTaskName0].NodeName = "node3"
	jobInfos["vcjob/job1"].Tasks[test.FakeTaskName1].NodeName = "node4"
	return jobInfos
}

func reCreateNPUTask910(name, namespace string, reqResourceNum int) util.NPUTask {
	return util.NPUTask{
		Name:       namespace + "/" + name,
		ReqNPUName: "huawei.com/Ascend910",
		ReqNPUNum:  reqResourceNum,
		Selector:   nil,
	}
}

func addNPUTaskToNPUJob(npuJob plugin.SchedulerJob, taskName, taskNamespace string, reqNPUNum int) plugin.SchedulerJob {
	task := reCreateNPUTask910(taskName, taskNamespace, reqNPUNum)
	npuJob.Tasks[api.TaskID(taskName)] = task
	npuJob.ReqNPUNum += reqNPUNum
	return npuJob
}

func reCreateSchedulerJob910(namespace string, UID api.JobID) plugin.SchedulerJob {
	sJob := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Name:      UID,
				NameSpace: namespace,
				Selector:  nil,
				Label:     nil,
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: "huawei.com/Ascend910",
				ReqNPUNum:  zero,
				Tasks:      make(map[api.TaskID]util.NPUTask, util.MapInitNum),
			},
		},
	}
	return sJob
}

func reNewReScheduler(graceTime int64) *ReScheduler {
	fakeReScheduler := ReScheduler{
		GraceDeleteTime:      graceTime,
		Level:                "",
		Jobs:                 nil,
		Nodes:                nil,
		kubeClient:           nil,
		DealReSchedulerCache: nil,
	}
	return &fakeReScheduler
}

func addReCacheToReScheduler(reScheduler *ReScheduler, reCache *DealReSchedulerCache) {
	reScheduler.DealReSchedulerCache = reCache
}

func addSchedulerJobToReScheduler(reScheduler *ReScheduler, sJob map[api.JobID]plugin.SchedulerJob) {
	reScheduler.Jobs = sJob
}

func buildReSchedulerAddFaultJobWithSession() []ReSchedulerAddFaultJobWithSessionTests {
	jobInfos1 := map[api.JobID]*api.JobInfo{
		"vcjob/job0": test.FakeNormalTestJob("job0", util.NPUIndex2),
		"vcjob/job1": test.FakeNormalTestJob("job1", util.NPUIndex2),
	}
	jobInfos1 = reAddFaultJobWithSessionModifyJobInfo(jobInfos1)
	jobs11 := reCreateSchedulerJob910("vcjob", "job0")
	jobs11 = addNPUTaskToNPUJob(jobs11, "task0", "vcjob", util.NPUIndex4)
	jobs11 = addNPUTaskToNPUJob(jobs11, "task1", "vcjob", util.NPUIndex4)
	jobs12 := reCreateSchedulerJob910("vcjob", "job1")
	jobs12 = addNPUTaskToNPUJob(jobs12, "task3", "vcjob", util.NPUIndex4)
	jobs12 = addNPUTaskToNPUJob(jobs12, "task4", "vcjob", util.NPUIndex4)
	jobs1 := map[api.JobID]plugin.SchedulerJob{
		"vcjob/job0": jobs11,
		"vcjob/job1": jobs12,
	}

	reCache1 := fakeCacheNoneFJobReSchedulerAddFaultJobWithSession()
	reScheduler1 := reNewReScheduler(0)
	addSchedulerJobToReScheduler(reScheduler1, jobs1)
	addReCacheToReScheduler(reScheduler1, reCache1)

	reCache2 := fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	reScheduler2 := reNewReScheduler(DefaultGraceOverTime)
	addSchedulerJobToReScheduler(reScheduler2, jobs1)
	addReCacheToReScheduler(reScheduler2, reCache2)
	test1 := ReSchedulerAddFaultJobWithSessionTests{
		name:   "01-AddFaultJobWithSession()-GraceDeleteTime 0",
		fields: reScheduler1,
		args: ReSchedulerAddFaultJobWithSessionArgs{
			jobs:        jobInfos1,
			cardName:    "huawei.com/Ascend910",
			cardPreName: "Ascend910-",
		},
		wantErr: false,
	}
	test2 := ReSchedulerAddFaultJobWithSessionTests{
		name:   "02-AddFaultJobWithSession()-GraceDeleteTime 900",
		fields: reScheduler2,
		args: ReSchedulerAddFaultJobWithSessionArgs{
			jobs:        jobInfos1,
			cardName:    "",
			cardPreName: "",
		},
		wantErr: false,
	}
	tests := []ReSchedulerAddFaultJobWithSessionTests{
		test1,
		test2,
	}
	return tests
}

// TestReSchedulerAddFaultJobWithSession test for add fault job
func TestReSchedulerAddFaultJobWithSession(t *testing.T) {
	env := plugin.ScheduleEnv{
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	tests := buildReSchedulerAddFaultJobWithSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := tt.fields
			if err := reScheduler.AddFaultJobWithSession(tt.args.jobs, env); (err != nil) != tt.wantErr {
				t.Errorf("AddFaultJobWithSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FaultNodeGetUnhealthyCardsFromDeviceInfoArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type FaultNodeGetUnhealthyCardsFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetUnhealthyCardsFromDeviceInfoArgs
	want    []string
	wantErr bool
}

func buildFaultNodeGetUnhealthyCardsFromDeviceInfoTests() []FaultNodeGetUnhealthyCardsFromDeviceInfoTests {
	node2 := fakeNPUNodeWithDeviceInfo("node0")
	test1 := FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		name:   "01-FaultNodeGetUnhealthyCardsFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetUnhealthyCardsFromDeviceInfoArgs{
			node:     fakeNPUNodeNilDeviceInfo("node0"),
			cardName: util.NPU910CardName,
		},
		want:    nil,
		wantErr: true,
	}
	test2 := FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		name:   "02-FaultNodeGetUnhealthyCardsFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetUnhealthyCardsFromDeviceInfoArgs{
			node:     node2,
			cardName: util.NPU910CardName,
		},
		want:    []string{"Ascend910-1"},
		wantErr: false,
	}
	tests := []FaultNodeGetUnhealthyCardsFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetUnhealthyCardsFromDeviceInfo test for get unhealthy card
func TestFaultNodeGetUnhealthyCardsFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetUnhealthyCardsFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getUnhealthyCardsFromDeviceInfo(tt.args.node, tt.args.cardName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUnhealthyCardsFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUnhealthyCardsFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type GetNetworkUnhealthyCardsFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs
	want    []string
	wantErr bool
}

func buildFaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests() []GetNetworkUnhealthyCardsFromDeviceInfoTests {
	node2 := fakeNPUNodeWithDeviceInfo("node0")
	test1 := GetNetworkUnhealthyCardsFromDeviceInfoTests{
		name:   "01-FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs{
			node:     fakeNPUNodeNilDeviceInfo("node0"),
			cardName: util.NPU910CardName,
		},
		want:    nil,
		wantErr: true,
	}
	test2 := GetNetworkUnhealthyCardsFromDeviceInfoTests{
		name:   "02-FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNetworkUnhealthyCardsFromDeviceInfoArgs{
			node:     node2,
			cardName: util.NPU910CardName,
		},
		want:    []string{"Ascend910-2"},
		wantErr: false,
	}
	tests := []GetNetworkUnhealthyCardsFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetNetworkUnhealthyCardsFromDeviceInfo test for get network unhealthy card
func TestFaultNodeGetNetworkUnhealthyCardsFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetNetworkUnhealthyCardsFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getNetworkUnhealthyCardsFromDeviceInfo(tt.args.node, tt.args.cardName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNetworkUnhealthyCardsFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNetworkUnhealthyCardsFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeTestFaultNodeCardNetworkUnhealthyOneCard(nodeName string) *FaultNode {
	updateTime := test.FakeUpdateTime + 1
	faultCards := fakeTestFaultCardsUnhealthy(nodeName, true)
	return &FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        nil,
		NetworkUnhealthyNPU: []string{"Ascend910-0"},
		IsFaultNode:         true,
		NodeDEnable:         true,
		NodeHealthState:     NodeCardNetworkUnhealthy,
		AllCards:            []string{"Ascend910-0"},
		FaultCards:          faultCards,
	}
}

type FaultNodeNewFaultCardHandlersArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type FaultNodeNewFaultCardHandlersTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeNewFaultCardHandlersArgs
	want    []FaultCard
	wantErr bool
}

func faultNodeNewFaultCardHandlersFaultCard(isFault bool, name, nodeName, faultType string) FaultCard {
	return FaultCard{
		IsFaultCard: isFault,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
}

func buildFaultNodeNewFaultCardHandlers() []FaultNodeNewFaultCardHandlersTests {
	test1 := FaultNodeNewFaultCardHandlersTests{
		name:   "01-newFaultCardHandlers()-NodeUnhealthy",
		fields: fakeTestFaultNodeNodeHealthyOneCard("node0"),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(false, "Ascend910-0", "node0", "Healthy"),
		},
		wantErr: false,
	}
	test2 := FaultNodeNewFaultCardHandlersTests{
		name:   "02-newFaultCardHandlers()-CardUnhealthy",
		fields: fakeTestFaultNodeCardUnhealthy("node0", []string{"Ascend910-0"}),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(
				true, "Ascend910-0", "node0", "Unhealthy"),
		},
		wantErr: false,
	}
	test3 := FaultNodeNewFaultCardHandlersTests{
		name:   "03-newFaultCardHandlers()-CardNetworkUnhealthy",
		fields: fakeTestFaultNodeCardNetworkUnhealthyOneCard("node0"),
		args: FaultNodeNewFaultCardHandlersArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want: []FaultCard{
			faultNodeNewFaultCardHandlersFaultCard(true, "Ascend910-0", "node0", "NetworkUnhealthy"),
		},
		wantErr: false,
	}
	tests := []FaultNodeNewFaultCardHandlersTests{
		test1,
		test2,
		test3,
	}
	return tests
}

// TestFaultNodeNewFaultCardHandlers test for new fault card handlers
func TestFaultNodeNewFaultCardHandlers(t *testing.T) {
	tests := buildFaultNodeNewFaultCardHandlers()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.createFaultCardHandlers(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFaultCardHandlers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFaultCardHandlers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeUpdateFaultNodesAttrArgs struct {
	node *plugin.NPUNode
}

type FaultNodeUpdateFaultNodesAttrTests struct {
	name    string
	fields  FaultNode
	args    FaultNodeUpdateFaultNodesAttrArgs
	wantErr bool
}

func buildFaultNodeUpdateFaultNodesAttrTestCases() []FaultNodeUpdateFaultNodesAttrTests {
	allCard := []string{"Ascend910-0", "Ascend910-1", "Ascend910-2", "Ascend910-3", "Ascend910-4",
		"Ascend910-5", "Ascend910-6", "Ascend910-7"}
	test1 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node0",
		fields: *fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node0"),
		},
	}
	test2 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node1",
		fields: *fakeTestFaultNodeNodeUnhealthy("node1"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node1"),
		},
		wantErr: false,
	}
	test3 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node2",
		fields: *fakeTestFaultNodeCardUnhealthy("node2", allCard),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node2"),
		},
		wantErr: false,
	}
	test4 := FaultNodeUpdateFaultNodesAttrTests{
		name:   "node3",
		fields: *fakeTestFaultNodeNodeUnhealthy("node3"),
		args: FaultNodeUpdateFaultNodesAttrArgs{
			fakeNPUNodeNilDeviceInfo("node3"),
		},
		wantErr: false,
	}
	tests := []FaultNodeUpdateFaultNodesAttrTests{
		test1,
		test2,
		test3,
		test4,
	}
	return tests
}

// TestFaultNodeUpdateFaultNodesAttr test fault node attribute
func TestFaultNodeUpdateFaultNodesAttr(t *testing.T) {
	tests := buildFaultNodeUpdateFaultNodesAttrTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := *fakeTestFaultNodeNodeHealthy("node0")
			if err := fNode.updateFaultNodesAttr(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("updateFaultNodesAttr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type TestReScheduler struct {
	DealReSchedulerCache *DealReSchedulerCache
	GraceDeleteTime      int64
	Level                string
	Jobs                 map[api.JobID]plugin.SchedulerJob
	Nodes                map[string]plugin.NPUNode
	kubeClient           kubernetes.Interface
}

type ReSchedulerCheckNodeNPUByTaskArgs struct {
	task   *api.TaskInfo
	vcNode plugin.NPUNode
}

type ReSchedulerCheckNodeNPUByTaskTests struct {
	name    string
	npuName string
	fields  TestReScheduler
	args    ReSchedulerCheckNodeNPUByTaskArgs
	wantErr error
}

func buildReSchedulerCheckNodeNPUByTaskTests() []ReSchedulerCheckNodeNPUByTaskTests {
	faultNode := fakeTestFaultNodeNodeUnhealthy("node0")
	faultTask00 := fakeTestFaultTaskFault("pod0", "vcjob", "node0", "0", "pppp")
	faultTask01 := fakeTestFaultTaskHealth("pod1", "vcjob", "node1", "1", "oooo")
	faultJob0 := fakeTestFaultJob([]string{"node0", "node1"}, []string{"0", "1"}, []FaultTask{*faultTask00,
		*faultTask01}, "job0", "vcjob")
	field1 := TestReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{
			DealReSchedulerConfigmap:   nil,
			FaultNodes:                 []FaultNode{*faultNode},
			FaultJobs:                  []FaultJob{*faultJob0},
			AllocNodeRankOccurrenceMap: nil,
		},
		GraceDeleteTime: 0,
		Level:           "",
		Jobs:            nil,
		Nodes:           nil,
		kubeClient:      nil,
	}
	arg1 := ReSchedulerCheckNodeNPUByTaskArgs{
		task: test.FakeNormalTestTask("pod1", "node1", "job0"),
		vcNode: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Name: "node2",
			},
		},
	}
	test1 := ReSchedulerCheckNodeNPUByTaskTests{
		name:    "01-CheckNodeNPUByTaskTests()-old task bind to new pod should be abandoned",
		npuName: util.NPU910CardName,
		fields:  field1,
		args:    arg1,
		wantErr: errors.New("task corresponding job not in session"),
	}
	tests := []ReSchedulerCheckNodeNPUByTaskTests{
		test1,
	}
	return tests
}

// TestReSchedulerCheckNodeNPUByTask test for re-scheduler check node NPU
func TestReSchedulerCheckNodeNPUByTask(t *testing.T) {
	tests := buildReSchedulerCheckNodeNPUByTaskTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(tt.fields)
			if err := reScheduler.CheckNodeNPUByTask(tt.args.task, tt.args.vcNode, tt.npuName); !reflect.DeepEqual(err,
				tt.wantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ReSchedulerCheckNodeCurNodeIsFaultArgs struct {
	curFJob *FaultJob
	task    *api.TaskInfo
	vcNode  plugin.NPUNode
}

type ReSchedulerCheckNodeCurNodeIsFaultTests struct {
	name    string
	fields  TestReScheduler
	args    ReSchedulerCheckNodeCurNodeIsFaultArgs
	wantErr bool
}

func buildReSchedulerCheckNodeCurNodeIsFaultTests() []ReSchedulerCheckNodeCurNodeIsFaultTests {
	test1 := ReSchedulerCheckNodeCurNodeIsFaultTests{
		name: "01-checkNodeCurNodeIsFault()-succeed",
		fields: TestReScheduler{
			DealReSchedulerCache: &DealReSchedulerCache{
				FaultNodes: []FaultNode{*fakeTestFaultNodeNodeUnhealthy("node0")},
				FaultJobs:  nil,
			},
		},
		args: ReSchedulerCheckNodeCurNodeIsFaultArgs{
			curFJob: &FaultJob{
				FaultTasks: []FaultTask{
					{
						TaskName: "pod0",
					},
				},
			},
			vcNode: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node0",
				},
			},
			task: &api.TaskInfo{
				Name: "pod0",
			},
		},
		wantErr: true,
	}
	tests := []ReSchedulerCheckNodeCurNodeIsFaultTests{
		test1,
	}
	return tests
}

// TestReSchedulerCheckNodeCurNodeIsFault test for check current node is fault node
func TestReSchedulerCheckNodeCurNodeIsFault(t *testing.T) {
	tests := buildReSchedulerCheckNodeCurNodeIsFaultTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(tt.fields)
			if err := reScheduler.checkNodeCurNodeIsFault(tt.args.vcNode, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("checkNodeCurNodeIsFault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func fakeTestTTReScheduler(fields TestReScheduler) *ReScheduler {
	return &ReScheduler{
		DealReSchedulerCache: fields.DealReSchedulerCache,
		GraceDeleteTime:      fields.GraceDeleteTime,
		Level:                fields.Level,
		Jobs:                 fields.Jobs,
		Nodes:                fields.Nodes,
		kubeClient:           fields.kubeClient,
	}
}

func TestCheckGraceDeleteTimeValid(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	t.Run("01-checkGraceDeleteTimeValid() return true when overTime in [2,3600]", func(t *testing.T) {
		res := reScheduler.checkGraceDeleteTimeValid(maxGraceOverTime)
		if !res {
			t.Errorf("checkGraceDeleteTimeValid() res = %v, wantRes is true", res)
		}
	})
	t.Run("02-checkGraceDeleteTimeValid() return false when overTime not in [2,3600]", func(t *testing.T) {
		res := reScheduler.checkGraceDeleteTimeValid(0)
		if res {
			t.Errorf("checkGraceDeleteTimeValid() res = %v, wantRes is false", res)
		}
	})
}

func mockJobInfo(jobName string, taskNum int) *api.JobInfo {
	jobInfo := test.FakeNormalTestJob(jobName, taskNum)
	var minRes = make(v1.ResourceList, taskNum)
	for _, task := range jobInfo.Tasks {
		for k, v := range task.Resreq.ScalarResources {
			minRes[k] = resource.MustParse(fmt.Sprintf("%f", v))
		}
	}
	jobInfo.PodGroup.Spec.MinResources = &minRes
	return jobInfo
}

func TestGetRunningJobs(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	ssn := test.FakeNormalSSN(nil)
	jobInfo1 := mockJobInfo(mockJobName1, test.NPUIndex4)
	jobInfo1.PodGroup.Status.Phase = util.PodGroupPending
	ssn.Jobs[jobInfo1.UID] = jobInfo1
	jobInfo2 := mockJobInfo(mockJobName2, test.NPUIndex4)
	ssn.Jobs[jobInfo2.UID] = jobInfo2
	reScheduler.Jobs = map[api.JobID]plugin.SchedulerJob{
		jobInfo2.UID: {SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}},
	}
	t.Run("01-GetRunningJobs() return error when running jobs is nil", func(t *testing.T) {
		if _, err := reScheduler.GetRunningJobs(ssn); err == nil {
			t.Errorf("GetRunningJobs() error = %v， wantErr is not nil", err)
		}
	})
}

func TestGetTaskRestartReason(t *testing.T) {
	t.Run("01-GetTaskRestartReason() return non-empty string when json marshal success",
		func(t *testing.T) {
			if res := GetTaskRestartReason([]FaultReasonList{}); res == "" {
				t.Errorf("GetTaskRestartReason() res = %v, wantRes is non-empty string", res)
			}
		})
	t.Run("02-GetTaskRestartReason() return empty string when json marshal failed",
		func(t *testing.T) {
			patch := gomonkey.ApplyFunc(json.Marshal, func(_ interface{}) ([]byte, error) {
				return nil, errors.New("json marshal failed")
			})
			defer patch.Reset()
			if res := GetTaskRestartReason(nil); res != "" {
				t.Errorf("GetTaskRestartReason() res = %v, wantRes is empty string", res)
			}
		})
}

func TestGetGraceDeleteFaultJobs(t *testing.T) {
	t.Run("01-getGraceDeleteFaultJobs() return not nil when fault jobs with grace label",
		func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(TestReScheduler{})
			reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
			reScheduler.DealReSchedulerCache.FaultJobs = append(reScheduler.DealReSchedulerCache.FaultJobs,
				[]FaultJob{{}}...)
			if res := reScheduler.getGraceDeleteFaultJobs(); res == nil {
				t.Errorf("getGraceDeleteFaultJobs() res = %v, wantRes is not nil", res)
			}
		})
}

func TestGetNeedForceDeleteDelayingNPUJobs(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	t.Run("01-GetNeedForceDeleteDelayingNPUJobs() return error when schedulerJobs or ssn is nil",
		func(t *testing.T) {
			_, err := reScheduler.GetNeedForceDeleteDelayingNPUJobs(nil, nil)
			if err == nil {
				t.Errorf("GetNeedForceDeleteDelayingNPUJobs() err = %v, wantErr is not nil", err)
			}
		})
	t.Run("02-GetNeedForceDeleteDelayingNPUJobs() return error when get none jobs", func(t *testing.T) {
		reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
		ssn := test.FakeNormalSSN(nil)
		jobInfo := mockJobInfo("job0", test.NPUIndex4)
		ssn.Jobs[jobInfo.UID] = jobInfo
		schedulerJobs := map[api.JobID]plugin.SchedulerJob{
			jobInfo.UID: {SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}},
		}
		_, err := reScheduler.GetNeedForceDeleteDelayingNPUJobs(schedulerJobs, ssn)
		if err == nil {
			t.Errorf("GetNeedForceDeleteDelayingNPUJobs() err = %v, wantErr is not nil", err)
		}
	})
}

func TestIsDelayingJobTimeout(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	reScheduler.GraceDeleteTime = graceDeleteTime
	t.Run("01-isDelayingJobTimeout() return true when job is timeout", func(t *testing.T) {
		fJob := &FaultJob{JobRankIdCreateTime: test.FakeUpdateTime}
		if res := reScheduler.isDelayingJobTimeout(fJob); !res {
			t.Errorf("isDelayingJobTimeout() res = %v, wantRes is true", res)
		}
	})
	t.Run("02-isDelayingJobTimeout() return true when job is not timeout", func(t *testing.T) {
		fJob := &FaultJob{JobRankIdCreateTime: time.Now().Unix()}
		if res := reScheduler.isDelayingJobTimeout(fJob); res {
			t.Errorf("isDelayingJobTimeout() res = %v, wantRes is false", res)
		}
	})
}

func TestRestartNeedForceDeleteJobs(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	ssn := test.FakeNormalSSN(nil)
	t.Run("01-RestartNeedForceDeleteJobs() return error when ssn is nil",
		func(t *testing.T) {
			err := reScheduler.RestartNeedForceDeleteJobs(nil, plugin.ScheduleEnv{})
			if err == nil {
				t.Errorf("RestartNeedForceDeleteJobs() err = %v, wantErr is not nil", err)
			}
		})
	t.Run("02-RestartNeedForceDeleteJobs() return error when get none jobs", func(t *testing.T) {
		err := reScheduler.RestartNeedForceDeleteJobs(ssn, plugin.ScheduleEnv{})
		if err == nil {
			t.Errorf("RestartNeedForceDeleteJobs() err = %v, wantErr is not nil", err)
		}
	})
	t.Run("03-RestartNeedForceDeleteJobs() return nil when jobs restart success", func(t *testing.T) {
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(reScheduler), "GetNeedForceDeleteDelayingNPUJobs",
			func(_ *ReScheduler, _ map[api.JobID]plugin.SchedulerJob,
				_ *framework.Session) ([]plugin.SchedulerJob, error) {
				jobs := reCreateSchedulerJob910("", mockJobUID)
				return []plugin.SchedulerJob{jobs}, nil
			})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(&FaultJob{}), "ForceDeleteJob",
			func(_ *FaultJob, _ *framework.Session, _ *plugin.SchedulerJob, _ plugin.ScheduleEnv) error {
				return errors.New("ForceDeleteJob error")
			})
		defer patch2.Reset()
		reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
		reScheduler.FaultJobs = append(reScheduler.FaultJobs, FaultJob{FaultTasks: []FaultTask{{}}})
		err := reScheduler.RestartNeedForceDeleteJobs(ssn, plugin.ScheduleEnv{})
		if err != nil {
			t.Errorf("RestartNeedForceDeleteJobs() err = %v, wantErr is not nil", err)
		}
	})
}

func TestRestartFaultJobs(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	ssn := test.FakeNormalSSN(nil)
	t.Run("01-RestartFaultJobs() return error when ssn is nil",
		func(t *testing.T) {
			err := reScheduler.RestartFaultJobs(nil, plugin.ScheduleEnv{})
			if err == nil {
				t.Errorf("RestartFaultJobs() err = %v, wantErr is not nil", err)
			}
		})
	t.Run("02-RestartFaultJobs() return nil when ssn is not nil and fault jobs restart success",
		func(t *testing.T) {
			reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
			err := reScheduler.RestartFaultJobs(ssn, plugin.ScheduleEnv{})
			if err != nil {
				t.Errorf("RestartFaultJobs() err = %v, wantErr is nil", err)
			}
		})
}

func TestUpdateRescheduleReason(t *testing.T) {
	t.Run("01-updateRescheduleReason() return nil when fJob is nil",
		func(t *testing.T) {
			res := updateRescheduleReason(nil, nil)
			if res != nil {
				t.Errorf("updateRescheduleReason() res = %v, wantRes is nil", res)
			}
		})
	t.Run("02-updateRescheduleReason() return not nil when fJob is not nil",
		func(t *testing.T) {
			fJob := &FaultJob{JobRankIdCreateTime: test.FakeUpdateTime}
			res := updateRescheduleReason(nil, fJob)
			if res == nil {
				t.Errorf("updateRescheduleReason() res = %v, wantRes is not nil", res)
			}
		})
}

func createFaultTasks(count int) []FaultTask {
	faultTasks := make([]FaultTask, count)
	for i := range faultTasks {
		faultTasks[i].IsFaultTask = true
	}
	return faultTasks
}

func TestConvertFaultTaskToRecords(t *testing.T) {
	t.Run("01-convertFaultTaskToRecords() return empty slice when fJob is nil",
		func(t *testing.T) {
			res := convertFaultTaskToRecords(nil)
			if len(res) != 0 {
				t.Errorf("convertFaultTaskToRecords() res = %v, wantRes is empty slice", res)
			}
		})
	t.Run("02-convertFaultTaskToRecords() return slice when fJob is not nil",
		func(t *testing.T) {
			fJob := fakeFaultJob()
			fJob.FaultTasks = append(fJob.FaultTasks, FaultTask{})
			fJob.FaultTasks = append(fJob.FaultTasks, createFaultTasks(10)...)
			res := convertFaultTaskToRecords(&fJob)
			if len(res) == 0 {
				t.Errorf("convertFaultTaskToRecords() res = %v, wantRes is non-empty slice", res)
			}
		})
}

func TestGetNewCacheJobs(t *testing.T) {
	t.Run("01-getNewCacheJobs() return slice when reScheduler.DealReSchedulerCache is not nil",
		func(t *testing.T) {
			reScheduler := fakeTestTTReScheduler(TestReScheduler{})
			reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
			if res := reScheduler.getNewCacheJobs([]FaultJob{}); res == nil {
				t.Errorf("getNewCacheJobs() res = %v, wantRes is non-empty slice", res)
			}
		})
}

func TestScoreBestNPUNodes(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	taskInfo := &api.TaskInfo{}
	scoreMap := map[string]float64{"test": 1}
	t.Run("01-ScoreBestNPUNodes() return error when task is nil",
		func(t *testing.T) {
			err := reScheduler.ScoreBestNPUNodes(nil, scoreMap)
			if err == nil {
				t.Errorf("ScoreBestNPUNodes() err = %v, wantErr is not nil", err)
			}
		})
	t.Run("02-ScoreBestNPUNodes() return nil when fJob is nil",
		func(t *testing.T) {
			err := reScheduler.ScoreBestNPUNodes(taskInfo, scoreMap)
			if err != nil {
				t.Errorf("ScoreBestNPUNodes() err = %v, wantErr is nil", err)
			}
		})
	taskInfo.Job = mockJobUID
	t.Run("03-ScoreBestNPUNodes() return nil when add scores on scoreMap success",
		func(t *testing.T) {
			err := reScheduler.ScoreBestNPUNodes(taskInfo, scoreMap)
			if err != nil {
				t.Errorf("ScoreBestNPUNodes() err = %v, wantErr is nil", err)
			}

		})
	t.Run("04-ScoreBestNPUNodes() return error when fJob.IsFaultJob is false",
		func(t *testing.T) {
			reScheduler.DealReSchedulerCache.FaultJobs = []FaultJob{
				{JobUID: mockJobUID, IsFaultJob: false},
			}
			err := reScheduler.ScoreBestNPUNodes(taskInfo, scoreMap)
			if err == nil {
				t.Errorf("ScoreBestNPUNodes() err = %v, wantErr is not nil", err)
			}
		})
}

func fakeSchedulerJobAttrByJob(job *api.JobInfo) util.SchedulerJobAttr {
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

func TestValidJobByReschedule(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	jobInfo1 := mockJobInfo(mockJobName1, test.NPUIndex4)
	curSchedulerJob := fakeSchedulerJobAttrByJob(jobInfo1)
	t.Run("01-ValidJobByReschedule() return nil when IsJobSinglePodDelete return true",
		func(t *testing.T) {
			res := reScheduler.ValidJobByReschedule(curSchedulerJob)
			if res != nil {
				t.Errorf("ValidJobByReschedule() res = %v, wantRes is nil", res)
			}
		})
	curSchedulerJob.SchedulingTaskNum = test.NPUIndex4
	t.Run("02-ValidJobByReschedule() return nil when fJob.IsFaultJob is false",
		func(t *testing.T) {
			res := reScheduler.ValidJobByReschedule(curSchedulerJob)
			if res != nil {
				t.Errorf("ValidJobByReschedule() res = %v, wantRes is nil", res)
			}
		})
	t.Run("03-ValidJobByReschedule() return nil when checkFJobUsedNormNodeRelease success",
		func(t *testing.T) {
			reScheduler.DealReSchedulerCache.FaultJobs = []FaultJob{{IsFaultJob: true}}
			res := reScheduler.ValidJobByReschedule(curSchedulerJob)
			if res != nil {
				t.Errorf("ValidJobByReschedule() res = %v, wantRes is nil", res)
			}
		})
	t.Run("04-ValidJobByReschedule() return not nil when checkFJobUsedNormNodeRelease failed",
		func(t *testing.T) {
			fJob := fakeFaultJob()
			fJob.JobNamespace = ""
			reScheduler.DealReSchedulerCache.FaultJobs = []FaultJob{fJob}
			reScheduler.DealReSchedulerCache.FaultNodeMaps = map[string]SimpleFNodeInfo{
				"node1": {NodeName: "node1"},
			}
			res := reScheduler.ValidJobByReschedule(curSchedulerJob)
			if res == nil {
				t.Errorf("ValidJobByReschedule() res = %v, wantRes is nil", res)
			}
		})
}

func TestCheckFJobUsedNormNodeRelease(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	reScheduler.DealReSchedulerCache.FaultNodeMaps = map[string]SimpleFNodeInfo{
		"node1": {NodeName: "node1"},
	}
	curFJob := fakeFaultJob()
	jobInfo1 := mockJobInfo(mockJobName1, test.NPUIndex4)
	curSchedulerJob := fakeSchedulerJobAttrByJob(jobInfo1)
	t.Run("01-checkFJobUsedNormNodeRelease() return error when node in faultJob hasn't been released",
		func(t *testing.T) {
			err := reScheduler.checkFJobUsedNormNodeRelease(curFJob, curSchedulerJob)
			if err == nil {
				t.Errorf("checkFJobUsedNormNodeRelease() res = %v, wantErr is not nil", err)
			}
		})
	reScheduler.Nodes = map[string]plugin.NPUNode{
		"node1": {
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: nil,
			}},
	}
	t.Run("02-checkFJobUsedNormNodeRelease() return error when check resource stabilize failed",
		func(t *testing.T) {
			err := reScheduler.checkFJobUsedNormNodeRelease(curFJob, curSchedulerJob)
			if err == nil {
				t.Errorf("checkFJobUsedNormNodeRelease() res = %v, wantErr is not nil", err)
			}
		})
}

func TestGetFNodeByNodeName(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	reScheduler.DealReSchedulerCache = fakeCacheWithFJobReSchedulerAddFaultJobWithSession()
	t.Run("01-getFNodeByNodeName() return nil when reScheduler.FaultNodeMaps is empty",
		func(t *testing.T) {
			res := reScheduler.getFNodeByNodeName("")
			if res != nil {
				t.Errorf("getFNodeByNodeName() res = %v, wantRes is nil", res)
			}
		})
	reScheduler.DealReSchedulerCache.FaultNodeMaps = map[string]SimpleFNodeInfo{
		"node1": {NodeName: "node1"},
	}
	t.Run("02-getFNodeByNodeName() return nil when get node failed",
		func(t *testing.T) {
			res := reScheduler.getFNodeByNodeName("")
			if res != nil {
				t.Errorf("getFNodeByNodeName() res = %v, wantRes is nil", res)
			}
		})
	t.Run("03-getFNodeByNodeName() return not nil when get node success",
		func(t *testing.T) {
			res := reScheduler.getFNodeByNodeName("node1")
			if res == nil {
				t.Errorf("getFNodeByNodeName() res = %v, wantRes is not nil", res)
			}
		})
}

func TestGetNPUNodeOfGiveNodeNameFromReScheduler(t *testing.T) {
	reScheduler := fakeTestTTReScheduler(TestReScheduler{})
	t.Run("01-getNPUNodeOfGiveNodeNameFromReScheduler() return nil when reScheduler.Nodes is empty",
		func(t *testing.T) {
			res := reScheduler.getNPUNodeOfGiveNodeNameFromReScheduler("")
			if res != nil {
				t.Errorf("getNPUNodeOfGiveNodeNameFromReScheduler() res = %v, wantRes is nil", res)
			}
		})

	reScheduler.Nodes = map[string]plugin.NPUNode{
		"node1": {
			CommonNode: plugin.CommonNode{
				Name:       "node1",
				Annotation: nil,
			}},
	}
	t.Run("02-getNPUNodeOfGiveNodeNameFromReScheduler() return nil when get node failed",
		func(t *testing.T) {
			res := reScheduler.getNPUNodeOfGiveNodeNameFromReScheduler("")
			if res != nil {
				t.Errorf("getNPUNodeOfGiveNodeNameFromReScheduler() res = %v, wantRes is nil", res)
			}
		})
	t.Run("03-getNPUNodeOfGiveNodeNameFromReScheduler() return not nil  when get node success",
		func(t *testing.T) {
			res := reScheduler.getNPUNodeOfGiveNodeNameFromReScheduler("node1")
			if res == nil {
				t.Errorf("getNPUNodeOfGiveNodeNameFromReScheduler() res = %v, wantRes is not nil", res)
			}
		})
}
