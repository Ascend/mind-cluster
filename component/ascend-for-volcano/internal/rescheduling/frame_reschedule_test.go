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
Package rescheduling is using for HuaWei Ascend pin fault
*/
package rescheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

const (
	sliceIndexZero   = 0
	sliceIndexOne    = 1
	sliceIndexTwo    = 2
	sliceIndexThree  = 3
	sliceIndexFour   = 4
	hbInterval       = 5
	fakeHBTime       = 12345
	ascend910FakeID0 = "Ascend910-0"
)

// NPUHandler base npu handler
type NPUHandler struct {
	plugin.SchedulerPlugin
	util.SchedulerJobAttr
	plugin.ScheduleEnv
	ReHandle      *ReScheduler
	MaxNodeNPUNum int
	MaxCardNPUNum int
}

// PreStartAction test for whole reScheduler pre-start operations
func (tp *module910x8Fields) PreStartAction(ssn *framework.Session) error {
	moduleFullName := util.NPU910CardName + util.ModuleAcceleratorType
	if tp == nil {
		return fmt.Errorf("%s handler not enabled: %s", moduleFullName, util.ArgumentError)
	}
	if ssn == nil {
		return fmt.Errorf("%s session is nil: %s", moduleFullName, util.ArgumentError)
	}
	reschEnable, ok := tp.baseHandler.SchedulerJobAttr.Label[JobRescheduleLabelKey]
	if !ok {
		return nil
	}
	if reschEnable == JobOffRescheduleLabelValue {
		return nil
	}
	tp.reHandle = New(&tp.baseHandler.ScheduleEnv, CmFaultJob)
	if tp.reHandle == nil {
		return fmt.Errorf("%s reSchedule not enabled: %s", moduleFullName, util.ArgumentError)
	}
	tp.reHandle.NewCommonReScheduler(CmFaultJob910x8Kind)
	tp.reHandle.SynCacheFaultNodeWithSession()
	tp.reHandle.AddFaultNodeWithSession()
	tp.reHandle.SynCacheFaultJobWithSession(ssn)
	tp.reHandle.SynCacheNodeRankOccMapWithSession(ssn)
	// 1. restart Fault Jobs that are recorded in cache
	if restartErr := tp.reHandle.RestartNeedForceDeleteJobs(ssn, tp.baseHandler.ScheduleEnv); restartErr != nil {
		klog.V(util.LogInfoLev).Infof("%s RestartNeedForceDeleteJobs: %s", moduleFullName, restartErr.Error())
	}
	// 2. get all the new 910x8 jobs in session
	runningJobs910x8, getRunErr := tp.reHandle.GetRunningJobs(ssn)
	if getRunErr != nil {
		klog.V(util.LogInfoLev).Infof("%s GetRunningJobs: %s", moduleFullName, getRunErr.Error())
	}
	// 3. get nodes of session and fault jobs of 910x8
	if err := tp.reHandle.AddFaultJobWithSession(runningJobs910x8, tp.baseHandler.ScheduleEnv); err != nil {
		klog.V(util.LogErrorLev).Infof("%s AddFaultJobWithSession", moduleFullName)
	}
	// 4. restart the fault jobs
	if restartErr := tp.reHandle.RestartFaultJobs(ssn, tp.baseHandler.ScheduleEnv); restartErr != nil {
		return restartErr
	}
	// 5. save structure for later allocation process
	tp.reHandle.GenerateNodeRankIndexTaskMap()
	return nil
}

type module910x8Fields struct {
	reHandle        *ReScheduler
	baseHandler     NPUHandler
	netUnhealthyKey string
	affScoreList    [][]int
}

type module910x8PreStartActionArgs struct {
	ssn              *framework.Session
	cacheFuncBefore1 func()
	cacheFuncBefore2 func()
	cacheFuncAfter1  func()
	cacheFuncAfter2  func()
	cacheFuncBefore3 func()
	cacheFuncAfter3  func()
	cacheFuncBefore4 func()
	cacheFuncAfter4  func()
}

type module910x8PreStartActionTests struct {
	fakeCm  *DealReSchedulerConfigmap
	name    string
	fields  module910x8Fields
	args    module910x8PreStartActionArgs
	wantErr bool
}

// buildModule910x8PreStartActionTest1 initial: no contents in reCache; 4 nodes, node0 card fault; 2 jobs,
// job0-pod0 on fault node
func buildModule910x8PreStartActionTest1() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	var tmpPatche1 *gomonkey.Patches = nil
	var tmpPatche2 *gomonkey.Patches = nil
	var tmpPatche3 *gomonkey.Patches = nil
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, nil)
	myArgs.ssn = ssn1
	test1 := module910x8PreStartActionTests{
		name: "01-PreStartAction()-no fault initially, add card fault and corresponding job fault in session",
		fields: module910x8Fields{
			baseHandler: NPUHandler{
				SchedulerPlugin:  plugin.SchedulerPlugin{},
				SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
				ScheduleEnv:      env,
				MaxNodeNPUNum:    0,
				MaxCardNPUNum:    0,
			},
			reHandle: &ReScheduler{
				GraceDeleteTime:      0,
				Level:                "",
				Jobs:                 nil,
				Nodes:                nil,
				DealReSchedulerCache: nil,
			},
		},
		args:    myArgs,
		wantErr: false,
	}
	test1.args.cacheFuncBefore2 = func() {
		tmpPatche2 = gomonkey.ApplyFunc(k8s.GetConfigMapWithRetry, func(
			_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
			return nil, errors.New("")
		})
	}
	return test1
}

func buildModule910x8PreStartActionTestCacheArgs(tmpPatche1 *gomonkey.Patches,
	tmpPatche2 *gomonkey.Patches, tmpPatche3 *gomonkey.Patches,
	faultCM *DealReSchedulerConfigmap) module910x8PreStartActionArgs {
	args := module910x8PreStartActionArgs{
		cacheFuncBefore1: func() {
			tmpPatche1 = gomonkey.ApplyMethod(reflect.TypeOf(&framework.Session{}), "Evict",
				func(_ *framework.Session, _ *api.TaskInfo, _ string) error { return nil })
		},
		cacheFuncBefore2: func() {
			tmpPatche2 = gomonkey.ApplyFunc(newReSchedulerCM, func() *DealReSchedulerConfigmap {
				return faultCM
			})
		},
		cacheFuncBefore3: func() {
			tmpPatche3 = gomonkey.ApplyMethod(reflect.TypeOf(&util.NPUTask{}), "DeleteRealPodByTask",
				func(_ *util.NPUTask, _ *framework.Session, _ int64) error { return nil })
		},
		cacheFuncAfter1: func() {
			if tmpPatche1 != nil {
				tmpPatche1.Reset()
			}
		},
		cacheFuncAfter2: func() {
			if tmpPatche2 != nil {
				tmpPatche2.Reset()
			}
		},
		cacheFuncAfter3: func() {
			if tmpPatche3 != nil {
				tmpPatche3.Reset()
			}
		},
	}
	return args
}

func buildModule910x8PreStartActionTest2() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	fakeEnvAddCacheFaultNodeToEnv(&env)
	fakeEnvAddCacheFaultJobToEnv(&env, []string{"job0", "node0", "node1"}, time.Now().Unix(), 0)

	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	reHandle := fakeReSchedulerNew(env)
	faultCM := fakeFaultCM(env)
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, faultCM)
	myArgs.ssn = ssn1
	test6 := module910x8PreStartActionTests{
		name: "02-PreStartAction()-with fault node and job in cm",
		fields: module910x8Fields{
			baseHandler: NPUHandler{
				SchedulerPlugin:  plugin.SchedulerPlugin{},
				SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
				ScheduleEnv:      env,
				MaxNodeNPUNum:    0,
				MaxCardNPUNum:    0,
			},
			reHandle: &reHandle,
		},
		fakeCm:  faultCM,
		args:    myArgs,
		wantErr: false,
	}
	return test6
}

func buildModule910x8PreStartActionTest4() module910x8PreStartActionTests {
	ssn1 := test.FakeSSNReSchedule()
	env := fakeEnvEmpty()
	fakeEnvAddJobsAndNodesToEnv(&env)
	fakeEnvAddCacheFaultNodeToEnv(&env)
	fakeEnvAddCacheFaultJobToEnv(&env, []string{"job0", "node0", "node1"}, 0, time.Now().Unix()-1)
	var tmpPatche1 *gomonkey.Patches
	var tmpPatche2 *gomonkey.Patches
	var tmpPatche3 *gomonkey.Patches
	reHandle := ReScheduler{
		GraceDeleteTime: DefaultGraceOverTime,
		Level:           "",
		Jobs:            env.Jobs,
		Nodes:           env.Nodes,
		DealReSchedulerCache: &DealReSchedulerCache{
			FaultNodes: nil,
			FaultJobs:  nil,
			DealReSchedulerConfigmap: &DealReSchedulerConfigmap{
				CMName:      CmName,
				CMNameSpace: CmNameSpace,
				CMData: map[string]string{CmFaultNodeKind: env.Cache.Data[RePropertyName][CmFaultNodeKind],
					CmFaultJob: env.Cache.Data[RePropertyName][CmFaultJob]},
			},
		},
	}
	faultCM := fakeFaultCM(env)
	myArgs := buildModule910x8PreStartActionTestCacheArgs(tmpPatche1, tmpPatche2, tmpPatche3, faultCM)
	myArgs.ssn = ssn1
	test7 := module910x8PreStartActionTests{
		name: "04-PreStartAction()-with fault node and job in cm and faultJob not in session",
		fields: module910x8Fields{
			baseHandler: fakeBaseHandlerEmpty(env),
			reHandle:    &reHandle,
		},
		fakeCm:  faultCM,
		args:    myArgs,
		wantErr: false,
	}
	return test7
}

func buildModule910x8PreStartActionTests() []module910x8PreStartActionTests {
	return []module910x8PreStartActionTests{
		buildModule910x8PreStartActionTest1(),
		buildModule910x8PreStartActionTest2(),
		buildModule910x8PreStartActionTest4(),
	}
}

// TestModule910x8PreStartAction test for 910x8 re-scheduling preStartAction
func TestModule910x8PreStartAction(t *testing.T) {
	tests := buildModule910x8PreStartActionTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reSchedulerConfigmap = tt.fakeCm
			tt.args.cacheFuncBefore1()
			tt.args.cacheFuncBefore2()
			if tt.args.cacheFuncBefore3 != nil {
				tt.args.cacheFuncBefore3()
			}
			tp := &module910x8Fields{
				baseHandler:     tt.fields.baseHandler,
				netUnhealthyKey: tt.fields.netUnhealthyKey,
				affScoreList:    tt.fields.affScoreList,
				reHandle:        tt.fields.reHandle,
			}
			if err := tp.PreStartAction(tt.args.ssn); (err != nil) != tt.wantErr {
				t.Errorf("PreStartAction() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.cacheFuncAfter1()
			tt.args.cacheFuncAfter2()
			if tt.args.cacheFuncAfter3 != nil {
				tt.args.cacheFuncAfter3()
			}
		})
	}
}

func fakeEnvEmpty() plugin.ScheduleEnv {
	schedulerCache := plugin.ScheduleCache{
		Names:      map[string]string{util.RePropertyCacheName: CmName},
		Namespaces: map[string]string{util.RePropertyCacheName: CmNameSpace},
		Data: map[string]map[string]string{
			util.RePropertyCacheName: {CmFaultNodeKind: "", CmFaultJob: ""},
		},
	}
	frameAttr := plugin.VolcanoFrame{
		Confs: []config.Configuration{
			{
				Name:      util.CMInitParamKey,
				Arguments: map[string]string{GraceOverTimeKey: "800"},
			},
		},
	}
	isFirstSession := false
	env := plugin.ScheduleEnv{
		IsFirstSession: &isFirstSession,
		Jobs:           make(map[api.JobID]plugin.SchedulerJob, util.NPUIndex2),
		Nodes:          make(map[string]plugin.NPUNode, util.NPUIndex4),
		FrameAttr:      frameAttr,
		Cache:          schedulerCache,
		SuperPodInfo: &plugin.SuperPodInfo{
			SuperPodReschdInfo:        map[api.JobID]map[string][]plugin.SuperNode{},
			SuperPodFaultTaskNodes:    map[api.JobID][]string{},
			SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{}},
	}
	return env
}

func fakeEnvAddJobsAndNodesToEnv(env *plugin.ScheduleEnv) {
	job0 := fakeSchedulerJobEmptyTask("job0", "vcjob")
	fakeSchedulerJobAddTask(&job0, "pod0", "vcjob", "node0", test.NPUIndex8)
	fakeSchedulerJobAddTask(&job0, "pod1", "vcjob", "node1", test.NPUIndex8)
	job1 := fakeSchedulerJobEmptyTask("job1", "vcjob")
	fakeSchedulerJobAddTask(&job1, "pod0", "vcjob", "node2", test.NPUIndex8)
	fakeSchedulerJobAddTask(&job1, "pod1", "vcjob", "node3", test.NPUIndex8)
	node0 := fakeNPUNodeUnhealthy("node0", []string{ascend910FakeID0}, []string{})
	node1 := fakeNPUNodeUnhealthy("node1", []string{}, []string{})
	node2 := fakeNPUNodeUnhealthy("node2", []string{}, []string{})
	node3 := fakeNPUNodeUnhealthy("node3", []string{}, []string{})
	env.Nodes = map[string]plugin.NPUNode{
		"node0": node0,
		"node1": node1,
		"node2": node2,
		"node3": node3,
	}
	env.Jobs = map[api.JobID]plugin.SchedulerJob{
		"vcjob/job0": job0,
		"vcjob/job1": job1,
	}
}

func fakeEnvAddCacheFaultJobToEnv(env *plugin.ScheduleEnv, paras []string, rankIdCreateTime int64,
	podCreateTime int64) {
	if len(paras) < util.NPUIndex3 {
		return
	}
	jobName := paras[sliceIndexZero]
	node0 := paras[sliceIndexOne]
	node1 := paras[sliceIndexTwo]
	faultTask1 := fakeReSchedulerFaultTask(true, []string{"pod0", "vcjob", node0, jobName, "0"}, podCreateTime,
		"ppppppppppppp")
	faultTask2 := fakeReSchedulerFaultTask(false, []string{"pod1", "vcjob", node1, jobName, "1"}, podCreateTime,
		"ppppppppppppp")
	faultJob := fakeReSchedulerFaultJobEmptyTask(jobName, "vcjob", time.Now().Unix()-1,
		rankIdCreateTime, true)
	fakeReSchedulerFaultJobAddTask(&faultJob, faultTask1)
	fakeReSchedulerFaultJobAddTask(&faultJob, faultTask2)
	faultJobString := dealFrameMarshal([]FaultJob{faultJob})
	env.Cache.Data[RePropertyName][CmFaultJob] = faultJobString
}

func fakeEnvAddCacheFaultNodeToEnv(env *plugin.ScheduleEnv) {
	now := time.Now().Unix()
	fCard1 := fakeReSchedulerFaultCard("Ascend910-0", "node0", true, CardUnhealthy)
	fCard2 := fakeReSchedulerFaultCard("Ascend910-1", "node0", false, CardHealthy)
	fCard3 := fakeReSchedulerFaultCard("Ascend910-2", "node0", false, CardHealthy)
	fCard4 := fakeReSchedulerFaultCard("Ascend910-3", "node0", false, CardHealthy)
	fCard5 := fakeReSchedulerFaultCard("Ascend910-4", "node0", false, CardHealthy)
	fCard6 := fakeReSchedulerFaultCard("Ascend910-5", "node0", false, CardHealthy)
	fCard7 := fakeReSchedulerFaultCard("Ascend910-6", "node0", false, CardHealthy)
	fCard8 := fakeReSchedulerFaultCard("Ascend910-7", "node0", false, CardHealthy)
	fNode := fakeReSchedulerFaultNodeEmptyCard("node0", []string{ascend910FakeID0}, []string{},
		hbInterval, []int64{now - 1, fakeHBTime, now - 1})
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard1)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard2)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard3)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard4)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard5)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard6)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard7)
	fakeReSchedulerFaultNodeAddFaultCards(&fNode, fCard8)
	faultNodes := []FaultNode{fNode}
	faultNodesMarshaled := dealFrameMarshal(faultNodes)
	env.Cache.Data[RePropertyName][CmFaultNodeKind] = faultNodesMarshaled
}

func fakeReSchedulerFaultNodeEmptyCard(nodeName string, unhealthyNPU []string, netUnhealthyNPU []string,
	HBinterval int, updateTimes []int64) FaultNode {
	var isFault bool
	if len(updateTimes) < util.NPUIndex3 {
		return FaultNode{}
	}
	updateTime := updateTimes[sliceIndexZero]
	hState := NodeHealthy
	if len(netUnhealthyNPU) > 0 {
		isFault = true
		hState = NodeCardNetworkUnhealthy
	}
	if len(unhealthyNPU) > 0 {
		isFault = true
		hState = NodeCardUnhealthy
	}

	faultNode := FaultNode{
		NodeName:            nodeName,
		UpdateTime:          updateTime,
		UnhealthyNPU:        unhealthyNPU,
		NetworkUnhealthyNPU: netUnhealthyNPU,
		IsFaultNode:         isFault,
		NodeDEnable:         true,
		NodeHealthState:     hState,
		AllCards: []string{"Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5," +
			"Ascend910-6,Ascend910-7"},
		FaultCards: make([]FaultCard, util.MapInitNum),
	}
	return faultNode
}

func fakeReSchedulerFaultCard(name, nodeName string, isFault bool, faultType string) FaultCard {
	faultCard := FaultCard{
		IsFaultCard: isFault,
		NPUName:     name,
		NodeName:    nodeName,
		FaultType:   faultType,
	}
	return faultCard
}

func fakeReSchedulerFaultNodeAddFaultCards(fNode *FaultNode, fCard FaultCard) {
	fNode.FaultCards = append(fNode.FaultCards, fCard)
}

func fakeReSchedulerFaultJobEmptyTask(jobName, jobNamespace string, updateTime int64,
	rankCreateTime int64, isFault bool) FaultJob {
	faultJob := FaultJob{
		ReScheduleKey:       "grace",
		IsFaultJob:          isFault,
		IsInSession:         true,
		JobName:             jobName,
		JobUID:              api.JobID(jobNamespace + "/" + jobName),
		JobNamespace:        jobNamespace,
		JobRankIds:          []string{},
		NodeNames:           []string{"node0", "node1"},
		FaultTasks:          []FaultTask{},
		UpdateTime:          updateTime,
		JobRankIdCreateTime: rankCreateTime,
	}
	return faultJob
}

func fakeReSchedulerFaultJobAddTask(fJob *FaultJob, fTask FaultTask) {
	fJob.FaultTasks = append(fJob.FaultTasks, fTask)
}

func fakeReSchedulerFaultTask(isFault bool, paras []string,
	podCreateTime int64, podUID types.UID) FaultTask {
	if len(paras) < test.NPUIndex5 {
		return FaultTask{}
	}
	name := paras[sliceIndexZero]
	ns := paras[sliceIndexOne]
	nodeName := paras[sliceIndexTwo]
	jobName := paras[sliceIndexThree]
	rankIndex := paras[sliceIndexFour]
	faultTask := FaultTask{
		IsFaultTask:   isFault,
		TaskUID:       api.TaskID(`"` + ns + `"-"` + name + `"`),
		TaskName:      name,
		TaskNamespace: ns,
		NodeName:      nodeName,
		JobName:       jobName,
		NodeRankIndex: rankIndex,
		UseCardName:   []string{"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"},
		PodCreateTime: podCreateTime,
		PodUID:        podUID,
	}
	return faultTask
}

func fakeSchedulerJobEmptyTask(jobName, namespace string) plugin.SchedulerJob {
	job0 := plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Name:      api.JobID(jobName),
				NameSpace: namespace,
				Selector:  map[string]string{util.AcceleratorType: util.ModuleAcceleratorType},
				Label: map[string]string{
					JobRescheduleLabelKey: JobGraceRescheduleLabelValue,
				},
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  0,
				Tasks:      make(map[api.TaskID]util.NPUTask, util.NPUIndex2),
			},
		},
	}
	return job0
}

func fakeSchedulerJobAddTask(sJob *plugin.SchedulerJob, taskName, ns, node string, reqNPUNum int) {
	task := util.NPUTask{
		Name:       taskName,
		ReqNPUName: util.NPU910CardName,
		ReqNPUNum:  reqNPUNum,
		Selector:   nil,
		VTask: &util.VTask{
			Allocated: util.TaskAllocated{
				NodeName: node,
			},
		},
	}
	sJob.Tasks[api.TaskID(`"`+ns+`"`+"-"+`"`+taskName+`"`)] = task
	sJob.ReqNPUNum += reqNPUNum
}

func fakeNPUNodeUnhealthy(nodeName string, unHealthyCard []string, networkUnhealthyCard []string) plugin.NPUNode {
	allCard := []string{ascend910FakeID0, "Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6," +
		"Ascend910-7"}
	var faultCards []string
	faultCards = append(faultCards, unHealthyCard...)
	faultCards = append(faultCards, networkUnhealthyCard...)
	healthyCards := make([]string, 0)
	var flag bool
	for _, card := range allCard {
		flag = false
		for _, fCardName := range faultCards {
			if card == fCardName {
				flag = true
				break
			}
		}
		if !flag {
			healthyCards = append(healthyCards, card)
		}
	}

	node0 := plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name:       nodeName,
			Capability: nil,
			Allocate:   nil,
			Idle:       nil,
			Annotation: map[string]string{
				util.NPU910CardName:                              strings.Join(healthyCards, ","),
				util.NPU910CardName + "-" + CardUnhealthy:        strings.Join(unHealthyCard, ","),
				util.NPU910CardName + "-" + CardNetworkUnhealthy: strings.Join(networkUnhealthyCard, ","),
			},
			Label: map[string]string{
				util.Accelerator: "huawei-Ascend910",
			},
		},
	}
	return node0
}

func dealFrameMarshal(data interface{}) string {
	dataString, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(dataString)
}

func fakeBaseHandlerEmpty(env plugin.ScheduleEnv) NPUHandler {
	return NPUHandler{
		SchedulerPlugin:  plugin.SchedulerPlugin{},
		SchedulerJobAttr: env.Jobs["vcjob/job0"].SchedulerJobAttr,
		ScheduleEnv:      env,
		MaxNodeNPUNum:    0,
		MaxCardNPUNum:    0,
	}
}

func fakeFaultCM(env plugin.ScheduleEnv) *DealReSchedulerConfigmap {
	cmData := make(map[string]string, util.MapInitNum)
	cmData[CmFaultNodeKind] = env.Cache.Data[RePropertyName][CmFaultNodeKind]
	cmData[CmFaultJob] =
		env.Cache.Data[RePropertyName][CmFaultJob]
	cmData[CmNodeRankTimeMapKind] = ""
	cmData[CmCheckCode] = util.MakeDataHash(cmData)
	return &DealReSchedulerConfigmap{
		CMName:      CmName,
		CMNameSpace: CmNameSpace,
		CMData:      cmData,
	}
}

func fakeReSchedulerNew(env plugin.ScheduleEnv) ReScheduler {
	reScheduler := ReScheduler{
		GraceDeleteTime: 0,
		Level:           "",
		Jobs:            env.Jobs,
		Nodes:           env.Nodes,
		DealReSchedulerCache: &DealReSchedulerCache{
			FaultNodes:               nil,
			FaultJobs:                nil,
			DealReSchedulerConfigmap: nil,
		},
	}
	return reScheduler
}
