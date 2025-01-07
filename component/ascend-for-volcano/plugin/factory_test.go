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
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type fields struct {
	NPUPlugins  map[string]NPUBuilder
	ScheduleEnv ScheduleEnv
}

type batchNodeOrderFnArgs struct {
	nodes []*api.NodeInfo
	ssn   *framework.Session
}

type batchNodeOrderFnTest struct {
	name    string
	args    batchNodeOrderFnArgs
	want    map[string]float64
	wantErr bool
}

// PatchGetCm go monkey patch get cm
func PatchGetCm(name, nameSpace string, data map[string]string) *gomonkey.Patches {
	return gomonkey.ApplyFunc(util.GetConfigMap, func(client kubernetes.Interface, namespace, cmName string) (
		*v1.ConfigMap, error) {
		return test.FakeConfigmap(name, nameSpace, data), nil
	})
}

func buildBatchNodeOrderFn01() batchNodeOrderFnTest {
	return batchNodeOrderFnTest{
		name:    "01-BatchNodeOrderFn nil Test",
		args:    batchNodeOrderFnArgs{nodes: nil, ssn: nil},
		wantErr: true,
	}
}

func buildBatchNodeOrderFn02() batchNodeOrderFnTest {
	tNodes := test.FakeNormalTestNodes(util.NPUIndex2)
	return batchNodeOrderFnTest{
		name:    "02-BatchNodeOrderFn ScoreBestNPUNodes ok Test",
		args:    batchNodeOrderFnArgs{nodes: tNodes, ssn: nil},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn03() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	handler := newDefaultHandler()
	initNormalsHandlerBySsnFunc(ssn, handler.InitVolcanoFrameFromSsn, handler.InitNodesFromSsn, handler.InitJobsFromSsn)
	return batchNodeOrderFnTest{
		name:    "03-BatchNodeOrderFn ScoreBestNPUNodes score node ok test",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn04() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex13)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "04-BatchNodeOrderFn nslb 1.0 test full tor is not enough, use logic tor",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn05() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex12)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "05-BatchNodeOrderFn nslb 1.0 test full tor is enough, use physic tor",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn06() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex2)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, LargeModelTag)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "06-BatchNodeOrderFn nslb 1.0 test, score node for fill job",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn07() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex15)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "07-BatchNodeOrderFn nslb 1.0 test full tor check filed by not enough logic tor",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: true,
	}
}

func buildBatchNodeOrderFn08() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex12)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "08-BatchNodeOrderFn nslb 2.0 test ,use 2 physic tor and 2 shared tor",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn09() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex2)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, LargeModelTag)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "09-BatchNodeOrderFn nslb 2.0 test, fill job use shared tor",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn010() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex15)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	handle := newDefaultHandler()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	return batchNodeOrderFnTest{
		name:    "10-BatchNodeOrderFn nslb 2.0 test, tor node num is not enough for normal job",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: true,
	}
}

func buildBatchNodeOrderFn011() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex2)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	return batchNodeOrderFnTest{
		name:    "11-BatchNodeOrderFn nslb test, score node for single single_layer job",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn012() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex8)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoAnnotations(fakeJob, JobDeleteFlag, fakeUsedNodeInfosByNodeNum(util.NPUIndex8))

	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	return batchNodeOrderFnTest{
		name:    "12-BatchNodeOrderFn nslb 2.0 test, score node for reschedule job when used tor node is enough",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn013() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	ssn.NodeList = deleteNodeByNodeName(ssn.NodeList, "node0")
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex8)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoAnnotations(fakeJob, JobDeleteFlag, fakeUsedNodeInfosByNodeNum(util.NPUIndex8))

	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	return batchNodeOrderFnTest{
		name:    "13-BatchNodeOrderFn nslb 2.0 test, score node for reschedule job when used tor node is not enough",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn014() batchNodeOrderFnTest {
	ssn := test.FakeNormalSSN(nil)
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex8)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoAnnotations(fakeJob, JobDeleteFlag, fakeUsedNodeInfosByNodeNum(util.NPUIndex8))

	test.AddJobInfoIntoSsn(ssn, fakeJob)
	for _, task := range fakeJob.Tasks {
		task.NodeName = ""
	}
	return batchNodeOrderFnTest{
		name:    "14-BatchNodeOrderFn nslb 1.0 test, score node for reschedule job when used tor node is enough",
		args:    batchNodeOrderFnArgs{nodes: ssn.NodeList, ssn: ssn},
		wantErr: false,
	}
}

func buildBatchNodeOrderFn() []batchNodeOrderFnTest {
	return []batchNodeOrderFnTest{
		buildBatchNodeOrderFn01(),
		buildBatchNodeOrderFn02(),
		buildBatchNodeOrderFn03(),
		buildBatchNodeOrderFn04(),
		buildBatchNodeOrderFn05(),
		buildBatchNodeOrderFn06(),
		buildBatchNodeOrderFn07(),
		buildBatchNodeOrderFn08(),
		buildBatchNodeOrderFn09(),
		buildBatchNodeOrderFn010(),
		buildBatchNodeOrderFn011(),
		buildBatchNodeOrderFn012(),
		buildBatchNodeOrderFn013(),
		buildBatchNodeOrderFn014(),
	}
}

func TestBatchNodeOrderFn(t *testing.T) {
	tests := buildBatchNodeOrderFn()
	tTask := test.FakeNormalTestTasks(1)[0]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := newDefaultHandler()
			patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
			defer patch1.Reset()
			initNormalsHandlerBySsnFunc(tt.args.ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
				handle.InitJobsFromSsn, handle.InitTorNodeInfo)
			if strings.Contains(tt.name, SingleLayer) {
				handle.Tors.torLevel = SingleLayer
			}
			_, err := handle.BatchNodeOrderFn(tTask, tt.args.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchNodeOrderFn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type beforeCloseHandlerTest struct {
	name   string
	fields fields
}

func buildBeforeCloseHandler() []beforeCloseHandlerTest {
	tests := []beforeCloseHandlerTest{
		{
			name: "01-BeforeCloseHandler no cache test",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
		},
		{
			name: "02-BeforeCloseHandler save cache test",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Cache: ScheduleCache{Names: map[string]string{"fault": "test"},
						Namespaces: map[string]string{"fault": "hahaNameSpace"},
						Data:       map[string]map[string]string{"fault": {"test1": "testData"}}}}},
		},
		{
			name: "03-BeforeCloseHandler save reset cm and tor infos",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: newDefaultsHandlerByFakeSsn().ScheduleEnv},
		},
	}
	return tests
}

func TestBeforeCloseHandler(t *testing.T) {
	tests := buildBeforeCloseHandler()
	tmpPatche := gomonkey.ApplyFunc(util.CreateOrUpdateConfigMap,
		func(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, cmNameSpace string) error {
			return nil
		})
	tmpPatche2 := gomonkey.ApplyFunc(util.GetConfigMapWithRetry, func(
		_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
		return test.FakeConfigmap(ResetInfoCMNamePrefix, "default", fakeResetCmInfos()), nil
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.BeforeCloseHandler()
		})
	}
	tmpPatche.Reset()
	tmpPatche2.Reset()
}

type getNPUSchedulerArgs struct {
	name string
}

type getNPUSchedulerTest struct {
	name   string
	fields fields
	args   getNPUSchedulerArgs
	want   ISchedulerPlugin
	want1  bool
}

func buildGetNPUSchedulerTest() []getNPUSchedulerTest {
	tests := []getNPUSchedulerTest{
		{
			name: "01-GetNPUScheduler not found test",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:  getNPUSchedulerArgs{name: "testPlugin"},
			want:  nil,
			want1: false,
		},
		{
			name: "02-GetNPUScheduler found test",
			fields: fields{NPUPlugins: map[string]NPUBuilder{"testPlugin": nil},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:  getNPUSchedulerArgs{name: "testPlugin"},
			want:  nil,
			want1: true,
		},
	}
	return tests
}

func TestGetNPUScheduler(t *testing.T) {
	tests := buildGetNPUSchedulerTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			got, got1 := sHandle.GetNPUScheduler(tt.args.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNPUScheduler() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetNPUScheduler() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type initNPUSessionArgs struct {
	ssn *framework.Session
}

type initNPUSessionTest struct {
	name     string
	sHandler *ScheduleHandler
	args     initNPUSessionArgs
	wantErr  bool
}

func buildInitNPUSessionTest() []initNPUSessionTest {
	tests := []initNPUSessionTest{
		{
			name:     "01-InitNPUSession nil ssn test",
			sHandler: &ScheduleHandler{},
			args:     initNPUSessionArgs{ssn: nil},
			wantErr:  true,
		},
		{
			name:     "02-InitNPUSession success test",
			sHandler: newDefaultHandler(),
			args:     initNPUSessionArgs{ssn: test.FakeNormalSSN(test.FakeConfigurations())},
			wantErr:  false,
		},
	}
	return tests
}

func TestInitNPUSession(t *testing.T) {
	tests := buildInitNPUSessionTest()
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addsHandlerCmInfosBySsn(tt.args.ssn, tt.sHandler)
			if err := tt.sHandler.InitNPUSession(tt.args.ssn); (err != nil) != tt.wantErr {
				t.Errorf("InitNPUSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type isPluginRegisteredArgs struct {
	name string
}

type isPluginRegisteredTest struct {
	name   string
	fields fields
	args   isPluginRegisteredArgs
	want   bool
}

func buildIsPluginRegisteredTest() []isPluginRegisteredTest {
	tests := []isPluginRegisteredTest{
		{
			name: "01-IsPluginRegistered not registered test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args: isPluginRegisteredArgs{name: "haha"},
			want: false,
		},
		{
			name:   "02-IsPluginRegistered registered test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{"haha": nil}},
			args:   isPluginRegisteredArgs{name: "haha"},
			want:   true,
		},
	}
	return tests
}

func TestIsPluginRegistered(t *testing.T) {
	tests := buildIsPluginRegisteredTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if got := sHandle.IsPluginRegistered(tt.args.name); got != tt.want {
				t.Errorf("IsPluginRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

type preStartPluginArgs struct {
	ssn *framework.Session
}

type preStartPluginTest struct {
	name   string
	fields fields
	args   preStartPluginArgs
}

func buildPreStartPluginTest() []preStartPluginTest {
	tests := []preStartPluginTest{
		{
			name:   "01-PreStartPlugin ok test",
			fields: fields{NPUPlugins: nil},
			args:   preStartPluginArgs{ssn: nil},
		},
	}
	return tests
}

func TestScheduleHandlerPreStartPlugin(t *testing.T) {
	tests := buildPreStartPluginTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.PreStartPlugin(tt.args.ssn)
		})
	}
}

type registerNPUSchedulerArgs struct {
	name string
	pc   NPUBuilder
}

type registerNPUSchedulerTest struct {
	name   string
	fields fields
	args   registerNPUSchedulerArgs
}

func buildRegisterNPUSchedulerTest() []registerNPUSchedulerTest {
	tests := []registerNPUSchedulerTest{
		{
			name:   "01-RegisterNPUScheduler not exist before test.",
			fields: fields{NPUPlugins: nil},
			args: registerNPUSchedulerArgs{
				name: "haha", pc: nil},
		},
		{
			name:   "02-RegisterNPUScheduler exist before test.",
			fields: fields{NPUPlugins: map[string]NPUBuilder{"haha": nil}},
			args: registerNPUSchedulerArgs{
				name: "haha", pc: nil},
		},
	}
	return tests
}

func TestRegisterNPUScheduler(t *testing.T) {
	tests := buildRegisterNPUSchedulerTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			sHandle.RegisterNPUScheduler(tt.args.name, tt.args.pc)
		})
	}
}

type unRegisterNPUSchedulerArgs struct {
	name string
}

type unRegisterNPUSchedulerTest struct {
	name    string
	fields  fields
	args    unRegisterNPUSchedulerArgs
	wantErr bool
}

func buildUnRegisterNPUSchedulerTest() []unRegisterNPUSchedulerTest {
	tests := []unRegisterNPUSchedulerTest{
		{
			name:    "01-UnRegisterNPUScheduler not exist before test.",
			fields:  fields{NPUPlugins: map[string]NPUBuilder{"hehe": nil}},
			args:    unRegisterNPUSchedulerArgs{name: "haha"},
			wantErr: false,
		},
		{
			name:    "02-UnRegisterNPUScheduler exist test.",
			fields:  fields{NPUPlugins: map[string]NPUBuilder{"haha": nil}},
			args:    unRegisterNPUSchedulerArgs{name: "haha"},
			wantErr: false,
		},
	}
	return tests
}

func TestUnRegisterNPUScheduler(t *testing.T) {
	tests := buildUnRegisterNPUSchedulerTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if err := sHandle.UnRegisterNPUScheduler(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("UnRegisterNPUScheduler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type initVolcanoFrameFromSsnTestCase struct {
	name    string
	configs []conf.Configuration
	want    VolcanoFrame
}

func buildInitVolcanoFrameFromSsnTestCases() []initVolcanoFrameFromSsnTestCase {
	superPodSizeKey := "super-pod-size"
	reserveNodesKey := "reserve-nodes"
	var testCases []initVolcanoFrameFromSsnTestCase
	testCases = append(testCases,
		getDefaultVolcanoFrameCasesOfSuperPodSizeFormatError(superPodSizeKey, reserveNodesKey)...)
	testCases = append(testCases,
		getDefaultVolcanoFrameCasesOfSuperPodSizeValueError(superPodSizeKey, reserveNodesKey)...)
	testCases = append(testCases,
		getDefaultVolcanoFrameCasesOfReserveNodesSelfValueError(superPodSizeKey, reserveNodesKey)...)
	testCases = append(testCases,
		getDefaultVolcanoFrameCasesOfReserveNodesValueMoreError(superPodSizeKey, reserveNodesKey)...)
	return testCases
}

func getDefaultVolcanoFrameCasesOfReserveNodesSelfValueError(superPodSizeKey,
	reserveNodesKey string) []initVolcanoFrameFromSsnTestCase {
	return []initVolcanoFrameFromSsnTestCase{
		{
			name: "05-GetReserveNodes failed, set default reserve-nodes: 2",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "40",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   40,
				ReservePodSize: 2,
			},
		},
		{
			name: "06-GetReserveNodes failed, set default reserve-nodes: 2",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "40",
						reserveNodesKey: "-1",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   40,
				ReservePodSize: 2,
			},
		},
	}
}

func getDefaultVolcanoFrameCasesOfReserveNodesValueMoreError(superPodSizeKey,
	reserveNodesKey string) []initVolcanoFrameFromSsnTestCase {
	return []initVolcanoFrameFromSsnTestCase{
		{
			name: "07-reserve-nodes is bigger than super-pod-size, set default reserve-nodes: 2",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "8",
						reserveNodesKey: "10",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   8,
				ReservePodSize: 2,
			},
		},
		{
			name: "08-reserve-nodes is bigger than super-pod-size, set default reserve-nodes: 1",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "2",
						reserveNodesKey: "90",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   2,
				ReservePodSize: 0,
			},
		},
	}
}

func getDefaultVolcanoFrameCasesOfSuperPodSizeFormatError(superPodSizeKey,
	reserveNodesKey string) []initVolcanoFrameFromSsnTestCase {
	return []initVolcanoFrameFromSsnTestCase{
		{
			name: "01-GetSizeOfSuperPod and GetReserveNodes failed, set default super-pod-size: 48, " +
				"default reserve-nodes: 2",
			configs: []conf.Configuration{
				{
					Name:      util.CMInitParamKey,
					Arguments: map[string]interface{}{},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   defaultSuperPodSize,
				ReservePodSize: defaultReserveNodes,
			},
		},
		{
			name: "02-GetSizeOfSuperPod failed, set default super-pod-size: 48",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "****",
						reserveNodesKey: "2",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   defaultSuperPodSize,
				ReservePodSize: defaultReserveNodes,
			},
		},
	}
}

func getDefaultVolcanoFrameCasesOfSuperPodSizeValueError(superPodSizeKey,
	reserveNodesKey string) []initVolcanoFrameFromSsnTestCase {
	return []initVolcanoFrameFromSsnTestCase{
		{
			name: "03-GetSizeOfSuperPod failed, set default super-pod-size: 48",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "-1",
						reserveNodesKey: "3",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   defaultSuperPodSize,
				ReservePodSize: 3,
			},
		},
		{
			name: "04-GetSizeOfSuperPod failed, set default super-pod-size: 48",
			configs: []conf.Configuration{
				{
					Name: util.CMInitParamKey,
					Arguments: map[string]interface{}{
						superPodSizeKey: "0",
						reserveNodesKey: "4",
					},
				},
			},
			want: VolcanoFrame{
				SuperPodSize:   defaultSuperPodSize,
				ReservePodSize: 4,
			},
		},
	}
}

func TestInitVolcanoFrameFromSsn(t *testing.T) {
	ssn := &framework.Session{}
	sHandle := &ScheduleHandler{}
	for _, tt := range buildInitVolcanoFrameFromSsnTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			ssn.Configurations = tt.configs
			sHandle.InitVolcanoFrameFromSsn(ssn)
			if !reflect.DeepEqual(sHandle.FrameAttr.SuperPodSize, tt.want.SuperPodSize) {
				t.Errorf("InitVolcanoFrameFromSsn() = %v, want %v", sHandle.FrameAttr.SuperPodSize, tt.want.SuperPodSize)
			}
			if !reflect.DeepEqual(sHandle.FrameAttr.ReservePodSize, tt.want.ReservePodSize) {
				t.Errorf("InitVolcanoFrameFromSsn() = %v, want %v", sHandle.FrameAttr.ReservePodSize, tt.want.ReservePodSize)
			}
		})
	}
}

// TestGetPodGroupOwnerRef test of getPodGroupOwnerRef
func TestGetPodGroupOwnerRef(t *testing.T) {
	t.Run("pg without ownerRef", func(t *testing.T) {
		pg := scheduling.PodGroup{}
		expectedOwner := metav1.OwnerReference{}
		owner := getPodGroupOwnerRef(pg)
		if !reflect.DeepEqual(expectedOwner, owner) {
			t.Errorf("getPodGroupOwnerRef = %v, want %v", owner, expectedOwner)
		}
	})
	t.Run("pg with ownerRef", func(t *testing.T) {
		controller := true
		pg := scheduling.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: &controller,
					},
				},
			},
		}
		expectedOwner := metav1.OwnerReference{
			Controller: &controller,
		}
		owner := getPodGroupOwnerRef(pg)
		if !reflect.DeepEqual(expectedOwner, owner) {
			t.Errorf("getPodGroupOwnerRef = %v, want %v", owner, expectedOwner)
		}
	})
}

// HandlerStart HuaWei NPU plugin start by frame.
func newDefaultHandler() *ScheduleHandler {
	isFirstSession := true
	scheduleHandler := &ScheduleHandler{
		NPUPlugins: map[string]NPUBuilder{},
		ScheduleEnv: ScheduleEnv{
			IsFirstSession:   &isFirstSession,
			Jobs:             map[api.JobID]SchedulerJob{},
			JobSeverInfos:    map[api.JobID]struct{}{},
			JobDeleteFlag:    map[api.JobID]struct{}{},
			JobSinglePodFlag: map[api.JobID]bool{},
			Nodes:            map[string]NPUNode{},
			DeleteJobInfos:   map[api.JobID]*api.JobInfo{},
			DeviceInfos: &DeviceInfosWithMutex{
				Mutex:   sync.Mutex{},
				Devices: map[string]NodeDeviceInfoWithID{},
			},
			NodeInfosFromCm: &NodeInfosFromCmWithMutex{
				Mutex: sync.Mutex{},
				Nodes: map[string]NodeDNodeInfo{},
			},
			SwitchInfosFromCm: &SwitchInfosFromCmWithMutex{
				Mutex:    sync.Mutex{},
				Switches: map[string]SwitchFaultInfo{},
			},
			FrameAttr: VolcanoFrame{},
			NslbAttr:  &NslbParameters{},
			SuperPodInfo: &SuperPodInfo{
				SuperPodReschdInfo:        map[api.JobID]map[string][]SuperNode{},
				SuperPodFaultTaskNodes:    map[api.JobID][]string{},
				SuperPodMapFaultTaskNodes: map[api.JobID]map[string]string{},
			},
			JobPendingMessage: map[api.JobID]map[string]map[string]struct{}{},
		},
	}

	scheduleHandler.RegisterNPUScheduler(util.NPU910CardName, New)
	return scheduleHandler
}

func newDefaultsHandlerByFakeSsn() *ScheduleHandler {
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	fakeJob := test.FakeJobInfoByName("pg0", util.NPUIndex8)
	test.AddJobInfoLabel(fakeJob, TorAffinityKey, NormalSchema)
	test.AddJobInfoLabel(fakeJob, util.SinglePodTag, util.EnableFunc)
	test.AddJobInfoIntoSsn(ssn, fakeJob)
	handle := newDefaultHandler()
	initNormalsHandlerBySsnFunc(ssn, handle.InitVolcanoFrameFromSsn, handle.InitNodesFromSsn,
		handle.InitJobsFromSsn, handle.InitTorNodeInfo)
	handle.FrameAttr.KubeClient = fake.NewSimpleClientset()
	return handle
}

func initNormalsHandlerBySsnFunc(ssn *framework.Session, initSsnFunc ...func(ssn *framework.Session)) {
	if ssn == nil {
		return
	}
	for _, initFunc := range initSsnFunc {
		initFunc(ssn)
	}
}

func initNormalsHandlerByNormalFunc(initFuncs ...func()) {
	for _, initFunc := range initFuncs {
		initFunc()
	}
}

func addsHandlerCmInfosBySsn(ssn *framework.Session, sHandler *ScheduleHandler) {
	if ssn == nil {
		return
	}
	for nodeName := range ssn.Nodes {
		sHandler.UpdateConfigMap(fakeDeviceInfoCMDataByNode(nodeName, fakeDeviceList()), util.DeleteOperator)
		sHandler.UpdateConfigMap(fakeDeviceInfoCMDataByNode(nodeName, fakeDeviceList()), util.AddOperator)
		sHandler.UpdateConfigMap(test.FakeConfigmap(util.NodeDCmInfoNamePrefix+nodeName, util.MindXDlNameSpace,
			fakeNodeInfos()), util.DeleteOperator)
		sHandler.UpdateConfigMap(test.FakeConfigmap(util.NodeDCmInfoNamePrefix+nodeName, util.MindXDlNameSpace,
			fakeNodeInfos()), util.AddOperator)
	}
}

// fakeUsedNodeInfosByNodeNum fake used node infos by node num, first node is fault
func fakeUsedNodeInfosByNodeNum(nodeNum int) string {
	var a []AllocNodeRankOccurrence
	for i := 0; i < nodeNum; i++ {
		tmpBool := false
		if i == 0 {
			tmpBool = true
		}
		tmp := AllocNodeRankOccurrence{
			NodeName:   "node" + strconv.Itoa(i),
			RankIndex:  strconv.Itoa(i),
			IsFault:    tmpBool,
			Occurrence: 0,
		}
		a = append(a, tmp)
	}
	if bytes, err := json.Marshal(a); err == nil {
		return string(bytes)
	}
	return ""
}

func deleteNodeByNodeName(nodes []*api.NodeInfo, nodeName string) []*api.NodeInfo {
	tmpNodes := make([]*api.NodeInfo, 0)
	for _, node := range nodes {
		if node.Name == nodeName {
			continue
		}
		tmpNodes = append(tmpNodes, node)
	}
	return tmpNodes
}

type initCmInformerTest struct {
	name    string
	ssn     *framework.Session
	sHandle *ScheduleHandler
}

func buildInitCmInformerTest01() initCmInformerTest {
	return initCmInformerTest{
		name:    "01-initCmInformerTest will return when kube client is nil",
		sHandle: &ScheduleHandler{},
	}
}

func buildInitCmInformerTest02() initCmInformerTest {
	ssn := test.FakeNormalSSN(test.FakeConfigurations())
	sHandler := newDefaultHandler()
	initNormalsHandlerBySsnFunc(ssn, sHandler.InitVolcanoFrameFromSsn)
	sHandler.FrameAttr.KubeClient = fake.NewSimpleClientset()
	return initCmInformerTest{
		name:    "02-initCmInformerTest will init cm by clusterd cm when conf is used cluster info manager",
		sHandle: sHandler,
	}
}

func buildInitCmInformerTest03() initCmInformerTest {
	tmpConf := test.FakeConfigurations()
	tmpConf[0].Arguments[util.UseClusterInfoManager] = "false"
	ssn := test.FakeNormalSSN(tmpConf)
	sHandler := newDefaultHandler()
	initNormalsHandlerBySsnFunc(ssn, sHandler.InitVolcanoFrameFromSsn)
	sHandler.FrameAttr.KubeClient = fake.NewSimpleClientset()
	return initCmInformerTest{
		name:    "03-initCmInformerTest will init cm by device info cm when conf is not use cluster info manager",
		sHandle: sHandler,
	}
}

func buildInitCmInformerTestCases() []initCmInformerTest {
	return []initCmInformerTest{
		buildInitCmInformerTest01(),
		buildInitCmInformerTest02(),
		buildInitCmInformerTest03(),
	}
}

func TestScheduleHandlerInitCmInformer(t *testing.T) {
	tests := buildInitCmInformerTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sHandle.initCmInformer()
		})
	}
}

type CheckVNPUSegmentEnableByConfigTest struct {
	name    string
	ssn     *framework.Session
	sHandle *ScheduleHandler
	want    bool
}

func TestVolcanoFrameCheckVNPUSegmentEnableByConfig(t *testing.T) {
	tmpConf := test.FakeConfigurations()
	tmpConf[0].Arguments[util.SegmentEnable] = "true"
	tests := []CheckVNPUSegmentEnableByConfigTest{
		{
			name:    "01-CheckVNPUSegmentEnableByConfigTest will return true when presetVirtualDevice is true",
			ssn:     test.FakeNormalSSN(tmpConf),
			sHandle: newDefaultHandler(),
			want:    true,
		},
		{
			name:    "02-CheckVNPUSegmentEnableByConfigTest will return true when conf is empty",
			ssn:     test.FakeNormalSSN(nil),
			sHandle: newDefaultHandler(),
			want:    false,
		},
		{
			name:    "03-CheckVNPUSegmentEnableByConfigTest will return true when presetVirtualDevice is false",
			ssn:     test.FakeNormalSSN(test.FakeConfigurations()),
			sHandle: newDefaultHandler(),
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sHandle.InitVolcanoFrameFromSsn(tt.ssn)
			if got := tt.sHandle.FrameAttr.CheckVNPUSegmentEnableByConfig(); got != tt.want {
				t.Errorf("CheckVNPUSegmentEnableByConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
