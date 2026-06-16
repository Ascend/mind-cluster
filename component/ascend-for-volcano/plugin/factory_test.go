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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type fields struct {
	NPUPlugins  sets.String
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
	return gomonkey.ApplyFunc(k8s.GetConfigMap, func(client kubernetes.Interface, namespace, cmName string) (
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

func buildBatchNodeOrderFn() []batchNodeOrderFnTest {
	return []batchNodeOrderFnTest{
		buildBatchNodeOrderFn01(),
		buildBatchNodeOrderFn02(),
		buildBatchNodeOrderFn03(),
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
				handle.Tors.TorLevel = SingleLayer
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
			fields: fields{NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{}}},
		},
		{
			name: "02-BeforeCloseHandler save cache test",
			fields: fields{NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					OutputCache: ScheduleCache{Names: map[string]string{"fault": "test"},
						Namespaces: map[string]string{"fault": "hahaNameSpace"},
						Data:       map[string]map[string]string{"fault": {"test1": "testData"}}}}},
		},
		{
			name: "03-BeforeCloseHandler save reset cm and tor infos",
			fields: fields{NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: newDefaultsHandlerByFakeSsn().ScheduleEnv},
		},
	}
	return tests
}

func TestBeforeCloseHandler(t *testing.T) {
	tests := buildBeforeCloseHandler()
	tmpPatche := gomonkey.ApplyFunc(k8s.CreateOrUpdateConfigMap,
		func(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, cmNameSpace string) error {
			return nil
		})
	tmpPatche2 := gomonkey.ApplyFunc(k8s.GetConfigMapWithRetry, func(
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
			if err := tt.sHandler.InitNPUSession(tt.args.ssn); (err != nil) != tt.wantErr {
				t.Errorf("InitNPUSession() error = %v, wantErr %v", err, tt.wantErr)
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
			sHandle.preStartPlugin(tt.args.ssn)
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   40,
					ReservePodSize: 2,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   40,
					ReservePodSize: 2,
				}}},
		},
	}
}

func TestGetForceEnqueueConfig(t *testing.T) {
	tests := []struct {
		name     string
		conf     map[string]string
		expected bool
	}{
		{
			name:     "01-when config is empty should return true",
			conf:     map[string]string{},
			expected: true,
		},
		{
			name:     "02-when forceEnqueue key not in config should return true",
			conf:     map[string]string{"other_key": "value"},
			expected: true,
		},
		{
			name:     "03-when the value of forceEnqueue key in config is true should return true",
			conf:     map[string]string{util.ForceEnqueue: "true"},
			expected: true,
		},
		{
			name:     "04-when the value of forceEnqueue key in config is false should return false",
			conf:     map[string]string{util.ForceEnqueue: "false"},
			expected: false,
		},
		{
			name:     "05-when the value of forceEnqueue key in config is not true should return false",
			conf:     map[string]string{util.ForceEnqueue: "1"},
			expected: false,
		},
		{
			name:     "06-when the value of forceEnqueue key in config is empty string should return false",
			conf:     map[string]string{util.ForceEnqueue: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getForceEnqueueConfig(tt.conf)
			if result != tt.expected {
				t.Errorf("getForceEnqueueConfig() = %v, expected %v", result, tt.expected)
			}
		})
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   8,
					ReservePodSize: 2,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   2,
					ReservePodSize: 0,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{

					SuperPodSize:   defaultSuperPodSize,
					ReservePodSize: defaultReserveNodes,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   defaultSuperPodSize,
					ReservePodSize: defaultReserveNodes,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   defaultSuperPodSize,
					ReservePodSize: 3,
				}}},
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
				ConfigParameters: ConfigParameters{DynamicParameters: DynamicParameters{
					SuperPodSize:   defaultSuperPodSize,
					ReservePodSize: 4,
				}}},
		},
	}
}

func TestInitVolcanoFrameFromSsn(t *testing.T) {
	ssn := &framework.Session{}
	sHandle := newDefaultHandler()
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
	scheduleHandler := &ScheduleHandler{
		NPUPlugins: sets.String{},
		ScheduleEnv: ScheduleEnv{
			ClusterCache:            NewClusterCache(),
			FrameAttr:               NewVolcanoFrame(),
			JobScheduleInfoRecorder: NewJobScheduleInfoRecorder(),
		},
	}

	scheduleHandler.FrameAttr.OnceInit = &sync.Once{}
	scheduleHandler.PolicyBuilder = func() SchedulerPluginNeed {
		return New(util.NPU910CardName)
	}
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

func fakeDeploymentJobInfo() *api.JobInfo {
	trueTag := true
	fakeJob := &api.JobInfo{
		Name:      "fakeJob",
		Namespace: "default",
		PodGroup:  &api.PodGroup{},
	}
	fakeJob.PodGroup.OwnerReferences = []metav1.OwnerReference{
		{
			Kind:       ReplicaSetType,
			Controller: &trueTag,
			Name:       "fakePG",
		},
	}
	return fakeJob
}

func fakeInformerFactory() informers.SharedInformerFactory {
	return informers.NewSharedInformerFactory(fake.NewSimpleClientset(), 0)
}

type getOwnerInfoTest struct {
	name    string
	vf      VolcanoFrame
	jobInfo *api.JobInfo
	wantErr bool
}

func buildGetOwnerInfoTest() []getOwnerInfoTest {
	return []getOwnerInfoTest{
		{
			name:    "01 will return nil when job is not deployment",
			vf:      VolcanoFrame{},
			jobInfo: &api.JobInfo{PodGroup: &api.PodGroup{}},
			wantErr: false,
		},
		{
			name:    "02 will return err when job is not exist",
			vf:      VolcanoFrame{KubeClient: fake.NewSimpleClientset(), informerFactory: fakeInformerFactory()},
			jobInfo: fakeDeploymentJobInfo(),
			wantErr: true,
		},
	}
}

func TestGetOwnerInfo(t *testing.T) {
	for _, tt := range buildGetOwnerInfoTest() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getOwnerInfo(tt.jobInfo, tt.vf)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOwnerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type getGraceDeleteTimeTest struct {
	name string
	conf map[string]string
	want int64
}

func buildGetGraceDeleteTimeTest() []getGraceDeleteTimeTest {
	return []getGraceDeleteTimeTest{
		{
			name: "01 will return default time when value is not int64",
			conf: map[string]string{GraceOverTimeKey: "test"},
			want: DefaultGraceOverTime,
		},
		{
			name: "01 will return default time when value is lower then 0",
			conf: map[string]string{GraceOverTimeKey: "1"},
			want: DefaultGraceOverTime,
		},
	}
}

func TestGetGraceDeleteTime(t *testing.T) {
	for _, tt := range buildGetGraceDeleteTimeTest() {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGraceDeleteTime(tt.conf); got != tt.want {
				t.Errorf("getGraceDeleteTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

type taskOrderFnArgs struct {
	l interface{}
	r interface{}
}
type taskOrderFnTest struct {
	name   string
	fields fields
	args   taskOrderFnArgs
	want   int
}

func createTaskOrderFnTask(jobId string, taskNum int) []*api.TaskInfo {
	tTask := test.FakeNormalTestTasks(taskNum)
	for index, task := range tTask {
		task.Job = api.JobID(jobId)
		task.Pod.Annotations[PodRankIndexKey] = strconv.Itoa(index)
	}
	return tTask
}

func createTaskOrderFnCasesOfPreferPreviousNode(
	jobIdStr string, task1, task2 *api.TaskInfo) []taskOrderFnTest {

	jobId := api.JobID(jobIdStr)
	schedulerJob := fakeSchedulerJobEmptyTask(jobIdStr, "")
	schedulerJob.Owner = OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}}

	bothHaveNodeCache := cache.NewPodNodeAffinityCache()
	bothHaveNodeCache.RecordAssignment(types.UID("owner-uid"), "0", "node-a")
	bothHaveNodeCache.RecordAssignment(types.UID("owner-uid"), "1", "node-b")

	onlyFirstHasNodeCache := cache.NewPodNodeAffinityCache()
	onlyFirstHasNodeCache.RecordAssignment(types.UID("owner-uid"), "0", "node-a")

	onlySecondHasNodeCache := cache.NewPodNodeAffinityCache()
	onlySecondHasNodeCache.RecordAssignment(types.UID("owner-uid"), "1", "node-b")

	neitherHasNodeCache := cache.NewPodNodeAffinityCache()

	sameNodeCache := cache.NewPodNodeAffinityCache()
	sameNodeCache.RecordAssignment(types.UID("owner-uid"), "0", "node-a")
	sameNodeCache.RecordAssignment(types.UID("owner-uid"), "1", "node-a")

	return []taskOrderFnTest{
		{
			name: "07-TaskOrderFn PreferPreviousNode: both have preferred nodes, l goes first",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs:          map[api.JobID]SchedulerJob{jobId: schedulerJob},
						AffinityCache: bothHaveNodeCache,
					},
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
					},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: -1,
		},
		{
			name: "08-TaskOrderFn PreferPreviousNode: task with node goes first (l has node)",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs:          map[api.JobID]SchedulerJob{jobId: schedulerJob},
						AffinityCache: onlyFirstHasNodeCache,
					},
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
					},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: -1,
		},
		{
			name: "09-TaskOrderFn PreferPreviousNode: task with node goes first (r has node)",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs:          map[api.JobID]SchedulerJob{jobId: schedulerJob},
						AffinityCache: onlySecondHasNodeCache,
					},
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
					},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: 1,
		},
		{
			name: "10-TaskOrderFn PreferPreviousNode: neither has node, same priority",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs:          map[api.JobID]SchedulerJob{jobId: schedulerJob},
						AffinityCache: neitherHasNodeCache,
					},
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
					},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: 0,
		},
		{
			name: "11-TaskOrderFn PreferPreviousNode: both have same node, same priority",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs:          map[api.JobID]SchedulerJob{jobId: schedulerJob},
						AffinityCache: sameNodeCache,
					},
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
					},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: 0,
		},
	}
}

func createTaskOrderFnCasesOfResolveRankIndexFallback(
	jobIdStr string, taskNoAnno, taskWithAnno *api.TaskInfo) []taskOrderFnTest {

	jobId := api.JobID(jobIdStr)
	schedulerJob := fakeSchedulerJobEmptyTask(jobIdStr, "")
	schedulerJob.Label[PodGroupScheduleKey] = PodGroupScheduleValue
	schedulerJob.NPUJob.Tasks = map[api.TaskID]util.NPUTask{
		taskNoAnno.UID: {Index: 3},
	}

	return []taskOrderFnTest{
		{
			name: "12-TaskOrderFn resolveRankIndex: fallback to vcJob.Tasks (rank 3 vs rank 1)",
			fields: fields{
				ScheduleEnv: ScheduleEnv{
					ClusterCache: ClusterCache{
						Jobs: map[api.JobID]SchedulerJob{jobId: schedulerJob},
					},
					FrameAttr: VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: taskNoAnno, r: taskWithAnno},
			want: 1,
		},
	}
}

func buildTaskOrderFnTest() []taskOrderFnTest {
	jobIdStr := "job1"
	tTask := createTaskOrderFnTask(jobIdStr, util.NPUIndex2)
	var tests []taskOrderFnTest
	tests = append(tests, createTaskOrderFnCasesOfTaskNil()...)
	tests = append(tests, createTaskOrderFnCasesOfJobNotExist(tTask[util.NPUIndex0], tTask[util.NPUIndex1])...)
	tests = append(tests, createTaskOrderFnCasesOfPodGroupLabelNotExist(jobIdStr,
		tTask[util.NPUIndex0], tTask[util.NPUIndex1])...)
	tests = append(tests, createTaskOrderFnCasesOfCompareTask(jobIdStr,
		tTask[util.NPUIndex0], tTask[util.NPUIndex1])...)
	tests = append(tests, createTaskOrderFnCasesOfPreferPreviousNode(jobIdStr,
		tTask[util.NPUIndex0], tTask[util.NPUIndex1])...)
	noAnnoTask := test.FakeNormalTestTask("no-anno-pod", "node1", jobIdStr)
	noAnnoTask.UID = "t-no-anno"
	noAnnoTask.Job = api.JobID(jobIdStr)
	tests = append(tests, createTaskOrderFnCasesOfResolveRankIndexFallback(jobIdStr,
		noAnnoTask, tTask[util.NPUIndex1])...)
	return tests
}

func createTaskOrderFnCasesOfTaskNil() []taskOrderFnTest {
	return []taskOrderFnTest{
		{
			name: "01-TaskOrderFn nil task",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: nil, r: nil},
			want: 0,
		},
		{
			name: "02-TaskOrderFn right value nil task",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: &api.TaskInfo{}, r: nil},
			want: 0,
		},
	}
}

func createTaskOrderFnCasesOfJobNotExist(task1, task2 *api.TaskInfo) []taskOrderFnTest {
	return []taskOrderFnTest{
		{
			name: "03-TaskOrderFn job not exist",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: NewClusterCache(),
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: 0,
		},
	}
}

func createTaskOrderFnCasesOfPodGroupLabelNotExist(jobIdStr string, task1, task2 *api.TaskInfo) []taskOrderFnTest {
	jobId := api.JobID(jobIdStr)
	schedulerJob := fakeSchedulerJobEmptyTask(jobIdStr, "")
	clusterCache := NewClusterCache()
	clusterCache.Jobs = map[api.JobID]SchedulerJob{jobId: schedulerJob}

	return []taskOrderFnTest{
		{
			name: "04-TaskOrderFn job podgroup label not exist",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: clusterCache,
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: 0,
		},
	}
}

func createTaskOrderFnCasesOfCompareTask(jobIdStr string, task1, task2 *api.TaskInfo) []taskOrderFnTest {
	jobId := api.JobID(jobIdStr)
	schedulerJob := fakeSchedulerJobEmptyTask(jobIdStr, "")
	schedulerJob.Label[PodGroupScheduleKey] = PodGroupScheduleValue
	clusterCache := NewClusterCache()
	clusterCache.Jobs = map[api.JobID]SchedulerJob{jobId: schedulerJob}
	return []taskOrderFnTest{
		{
			name: "05-TaskOrderFn job l < r",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: clusterCache,
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: task1, r: task2},
			want: -1,
		},
		{
			name: "06-TaskOrderFn job l > r",
			fields: fields{
				NPUPlugins: map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{
					ClusterCache: clusterCache,
					FrameAttr:    VolcanoFrame{},
				},
			},
			args: taskOrderFnArgs{l: task2, r: task1},
			want: 1,
		},
	}
}

// TestTaskOrderFn test of TaskOrderFn
func TestTaskOrderFn(t *testing.T) {
	tests := buildTaskOrderFnTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:    tt.fields.NPUPlugins,
				ScheduleEnv:   tt.fields.ScheduleEnv,
				AffinityCache: tt.fields.ScheduleEnv.ClusterCache.AffinityCache,
			}
			if got := sHandle.TaskOrderFn(tt.args.l, tt.args.r); got != tt.want {
				t.Errorf("TaskOrderFn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeSchedulerJobEmptyTask(jobName, namespace string) SchedulerJob {
	job := SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			ComJob: util.ComJob{
				Name:      api.JobID(jobName),
				NameSpace: namespace,
				Selector:  map[string]string{},
				Label:     map[string]string{},
			},
			NPUJob: &util.NPUJob{
				ReqNPUName: util.NPU910CardName,
				ReqNPUNum:  0,
				Tasks:      make(map[api.TaskID]util.NPUTask),
			},
		},
	}
	return job
}

type taskRankIDFnArgs struct {
	task *api.TaskInfo
}
type taskRankIdFnTest struct {
	name    string
	fields  fields
	args    taskRankIDFnArgs
	want    int
	wantErr bool
}

func buildObtainTaskRankIdCases() []taskRankIdFnTest {
	task1 := test.FakeNormalTestTask("pod1", "node1", "acjob")
	delete(task1.Pod.Annotations, PodRankIndexKey)
	task2 := test.FakeNormalTestTask("pod1", "node1", "acjob")
	task2.Pod.Annotations[PodRankIndexKey] = ""

	return []taskRankIdFnTest{
		{
			name:    "01-obtainTaskRankId, task is nil",
			args:    taskRankIDFnArgs{task: nil},
			want:    0,
			wantErr: true,
		},
		{
			name:    "02-obtainTaskRankId, pod annotation not exist",
			args:    taskRankIDFnArgs{task: task1},
			want:    0,
			wantErr: true,
		},
		{
			name:    "03-obtainTaskRankId, pod annotation not int",
			args:    taskRankIDFnArgs{task: task2},
			want:    0,
			wantErr: true,
		},
	}
}

// TestObtainTaskRankId test obtainTaskRankId
func TestObtainTaskRankId(t *testing.T) {
	tests := buildObtainTaskRankIdCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  map[string]sets.Empty{},
				ScheduleEnv: ScheduleEnv{ClusterCache: NewClusterCache(), FrameAttr: VolcanoFrame{}},
			}
			res, err := sHandle.obtainTaskRankId(tt.args.task)
			if res != tt.want || (err != nil) != tt.wantErr {
				t.Errorf("TaskOrderFn() = %v, want %v", res, tt.want)
			}
		})
	}
}

type resourceLevelsCase struct {
	name       string
	configs    map[string]string
	wantErr    bool
	wantLen    int
	wantSubLen int
}

func buildResourceLevelsCases() []resourceLevelsCase {
	validConfig := `{"tree1":{"level1":{"Label":"level1"}}}`
	return []resourceLevelsCase{
		{name: "01-no config", configs: map[string]string{}, wantErr: true},
		{name: "02-invalid json", configs: map[string]string{configResourceLevelConfig: "invalid"}, wantErr: true},
		{name: "03-empty config", configs: map[string]string{configResourceLevelConfig: "{}"}, wantErr: true},
		{
			name:       "04-valid config",
			configs:    map[string]string{configResourceLevelConfig: validConfig},
			wantErr:    false,
			wantLen:    1,
			wantSubLen: 3,
		},
	}
}

func TestInitResourceLevels(t *testing.T) {
	for _, tt := range buildResourceLevelsCases() {
		t.Run(tt.name, func(t *testing.T) {
			got := initResourceLevels(tt.configs)
			if (len(got) == 0) != tt.wantErr {
				t.Errorf("initResourceLevels() empty = %v, wantErr %v", len(got) == 0, tt.wantErr)
			}
		})
	}
}

func TestGetConfigLevels(t *testing.T) {
	for _, tt := range buildResourceLevelsCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigLevels(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigLevels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("getConfigLevels() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

type configLevelCase struct {
	name        string
	levelConfig map[string]util.ResourceTreeLevel
	wantErr     bool
	wantLen     int
}

func buildConfigLevelCases() []configLevelCase {
	return []configLevelCase{
		{name: "01-empty config", levelConfig: map[string]util.ResourceTreeLevel{}, wantErr: false, wantLen: 2},
		{
			name: "02-missing level1",
			levelConfig: map[string]util.ResourceTreeLevel{
				"level2": {Label: "level2"},
			},
			wantErr: true,
		},
		{
			name: "03-valid single level",
			levelConfig: map[string]util.ResourceTreeLevel{
				"level1": {Label: "level1", ReservedNode: 1},
			},
			wantErr: false,
			wantLen: 3,
		},
		{
			name: "04-valid multi level",
			levelConfig: map[string]util.ResourceTreeLevel{
				"level1": {Label: "level1", ReservedNode: 1},
				"level2": {Label: "level2"},
			},
			wantErr: false,
			wantLen: 4,
		},
	}
}

func TestGetConfigLevel(t *testing.T) {
	for _, tt := range buildConfigLevelCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigLevel(tt.levelConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("getConfigLevel() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

type initJobsFromSsnTest struct {
	name      string
	ssn       *framework.Session
	jobsInSsn []*api.JobInfo
	wantJobIn bool
}

func buildInitJobsFromSsnTestCases() []initJobsFromSsnTest {
	return []initJobsFromSsnTest{
		{
			name: "01-NPU job with MinResource NPU>0 is kept.",
			ssn:  test.FakeNormalSSN(test.FakeConfigurations()),
			jobsInSsn: func() []*api.JobInfo {
				j := test.FakeJobInfoByName("npuJob", 1)
				return []*api.JobInfo{j}
			}(),
			wantJobIn: true,
		},
		{
			name: "02-NPU job with annotation-based NPU (ReqNPUNum=0) is kept.",
			ssn:  test.FakeNormalSSN(test.FakeConfigurations()),
			jobsInSsn: func() []*api.JobInfo {
				j := test.FakeNormalTestJob("annoNPUJob", 1)
				j.PodGroup.Annotations = map[string]string{
					util.SchedulePluginAnno: util.Ascend910,
				}
				j.TotalRequest.ScalarResources = map[v1.ResourceName]float64{}
				j.PodGroup.Spec.MinResources = &v1.ResourceList{}
				return []*api.JobInfo{j}
			}(),
			wantJobIn: true,
		},
		{
			name: "03-Job with annotation-based NPU (ReqNPUNum=0) gets policy handler.",
			ssn:  test.FakeNormalSSN(test.FakeConfigurations()),
			jobsInSsn: func() []*api.JobInfo {
				j := test.FakeNormalTestJob("annoNPUJob2", 1)
				j.PodGroup.Annotations = map[string]string{
					util.SchedulePluginAnno: util.Ascend310P,
				}
				j.TotalRequest.ScalarResources = map[v1.ResourceName]float64{}
				j.PodGroup.Spec.MinResources = &v1.ResourceList{}
				return []*api.JobInfo{j}
			}(),
			wantJobIn: true,
		},
	}
}

func TestInitJobsFromSsn(t *testing.T) {
	patch1 := PatchGetCm(TorNodeCMName, "kube-system", test.FakeTorNodeData())
	defer patch1.Reset()
	for _, tt := range buildInitJobsFromSsnTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			for _, job := range tt.jobsInSsn {
				test.AddJobInfoIntoSsn(tt.ssn, job)
			}
			handler := newDefaultHandler()
			initNormalsHandlerBySsnFunc(tt.ssn, handler.InitVolcanoFrameFromSsn,
				handler.InitNodesFromSsn, handler.InitJobsFromSsn)
			for _, job := range tt.jobsInSsn {
				_, exists := handler.Jobs[job.UID]
				if tt.wantJobIn && !exists {
					t.Errorf("InitJobsFromSsn() job %s should be in handler.Jobs", job.UID)
				}
				if !tt.wantJobIn && exists {
					t.Errorf("InitJobsFromSsn() job %s should NOT be in handler.Jobs", job.UID)
				}
			}
		})
	}
}

// fakeFaultHandle implements FaultHandler for testing.
type fakeFaultHandle struct {
	isFaultTaskFn func(api.JobID, string) bool
	isNodeFaultFn func(string) bool
}

func (f *fakeFaultHandle) Execute(*ScheduleEnv, *framework.Session) error      { return nil }
func (f *fakeFaultHandle) CheckNodeNPUByTask(*api.TaskInfo, *NPUNode) error    { return nil }
func (f *fakeFaultHandle) ScoreBestNPUNodes(*api.TaskInfo, map[string]float64) {}
func (f *fakeFaultHandle) UseAnnotation(*api.TaskInfo)                         {}
func (f *fakeFaultHandle) PreStopAction(*ScheduleEnv) error                    { return nil }
func (f *fakeFaultHandle) IsNodeFault(nodeName string) bool {
	if f.isNodeFaultFn != nil {
		return f.isNodeFaultFn(nodeName)
	}
	return false
}
func (f *fakeFaultHandle) IsFaultTaskByRank(jobID api.JobID, rankIndex string) bool {
	if f.isFaultTaskFn != nil {
		return f.isFaultTaskFn(jobID, rankIndex)
	}
	return false
}

func TestAddPreferPreviousNodeScore(t *testing.T) {
	prefMap := map[int]string{0: "node1", 1: "node2"}
	baseScoreMap := func() map[string]float64 {
		return map[string]float64{"node1": 8.0, "node2": 7.5, "node3": 7.0, "node4": 6.0}
	}
	baseTask := &api.TaskInfo{
		Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}},
		Job: "job1",
	}
	vcJobWithPrefMap := func(pref map[int]string) SchedulerJob {
		return SchedulerJob{
			Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
			},
			PrefNodeMap: pref,
		}
	}
	enabledFrame := func() *ScheduleHandler {
		return &ScheduleHandler{
			ScheduleEnv: ScheduleEnv{
				FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
					DynamicParameters: DynamicParameters{PreferPreviousNode: true}},
				},
			},
		}
	}

	tests := []struct {
		name      string
		sHandle   *ScheduleHandler
		task      *api.TaskInfo
		vcJob     SchedulerJob
		wantNode1 float64
		wantNode2 float64
		scoreMap  map[string]float64
	}{
		{
			name:    "01-superpod job returns early",
			sHandle: enabledFrame(),
			task:    baseTask,
			vcJob: SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Annotation: map[string]string{util.SchedulePolicyAnnoKey: util.Chip8Node8Sp}},
					NPUJob: &util.NPUJob{NPUTaskNum: 2},
				},
			},
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name: "02-prefer-previous-node disabled returns early",
			sHandle: &ScheduleHandler{
				ScheduleEnv: ScheduleEnv{
					FrameAttr: VolcanoFrame{ConfigParameters: ConfigParameters{
						DynamicParameters: DynamicParameters{PreferPreviousNode: false}},
					},
				},
			},
			task:      baseTask,
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:    "03-multilevel job returns early",
			sHandle: enabledFrame(),
			task:    baseTask,
			vcJob: SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{Annotation: map[string]string{util.SchedulePolicyAnnoKey: util.MultiLevel}},
					NPUJob: &util.NPUJob{NPUTaskNum: 2},
				},
			},
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:      "04-nil task returns early",
			sHandle:   enabledFrame(),
			task:      nil,
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:      "05-empty scoreMap returns early",
			sHandle:   enabledFrame(),
			task:      baseTask,
			vcJob:     vcJobWithPrefMap(prefMap),
			scoreMap:  map[string]float64{},
			wantNode1: 0,
			wantNode2: 0,
		},
		{
			name:    "06-empty rank index returns early",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}},
				Job: "job1",
			},
			vcJob: SchedulerJob{
				Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
				SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: map[api.TaskID]util.NPUTask{}},
				},
			},
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:    "07-no PrefNodeMap returns early",
			sHandle: enabledFrame(),
			task:    baseTask,
			vcJob: SchedulerJob{
				Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
				SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
				},
			},
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:      "08-empty PrefNodeMap returns early",
			sHandle:   enabledFrame(),
			task:      baseTask,
			vcJob:     vcJobWithPrefMap(map[int]string{}),
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:    "09-invalid rank string returns early",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "abc"}}},
				Job: "job1",
			},
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			wantNode2: 7.5,
		},
		{
			name:      "10-nonfault: selfNode in scoreMap gets boosted",
			sHandle:   enabledFrame(),
			task:      baseTask,
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0 + defaultPreferPreviousScore,
			wantNode2: 7.5,
		},
		{
			name:    "11-nonfault: selfNode below max gets boosted",
			sHandle: enabledFrame(),
			task:    baseTask,
			vcJob:   vcJobWithPrefMap(prefMap),
			scoreMap: func() map[string]float64 {
				return map[string]float64{"node1": 6.0, "node2": 7.5, "node3": 7.0, "node4": 6.0}
			}(),
			wantNode1: 7.5 + defaultPreferPreviousScore,
			wantNode2: 7.5,
		},
		{
			name:    "12-nonfault: rank1 selfNode=node2 gets boosted",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "1"}}},
				Job: "job1",
			},
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			wantNode2: 8.0 + defaultPreferPreviousScore,
			scoreMap:  map[string]float64{"node1": 8.0, "node2": 6.0, "node3": 7.0, "node4": 6.0},
		},
		{
			name:    "13-nonfault: selfNode not in map, boost best otherNode",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "1"}}},
				Job: "job1",
			},
			vcJob:     vcJobWithPrefMap(prefMap),
			scoreMap:  map[string]float64{"node1": 8.0, "node3": 7.0, "node4": 6.0},
			wantNode1: 8.0,
			wantNode2: 8.0 + defaultPreferPreviousScore,
		},
		{
			name:    "14-nonfault: rank not in prefMap, boost best otherNode",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "5"}}},
				Job: "job1",
			},
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			scoreMap:  map[string]float64{"node1": 8.0, "node3": 7.0, "node4": 6.0},
			wantNode2: 8.0 + defaultPreferPreviousScore,
		},
		{
			name:    "15-nonfault: no selfNode/otherNodes, boost best peerNode",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "5"}}},
				Job: "job1",
			},
			vcJob:     vcJobWithPrefMap(prefMap),
			scoreMap:  map[string]float64{"node1": 8.0, "node2": 7.0},
			wantNode1: 8.0 + defaultPreferPreviousScore,
			wantNode2: 7.0,
		},
		{
			name: "16-fault: boost best otherNode",
			sHandle: func() *ScheduleHandler {
				h := enabledFrame()
				h.FaultHandle = &mockFaultHandler{faultByRank: true}
				return h
			}(),
			task:      baseTask,
			scoreMap:  map[string]float64{"node1": 8.0, "node3": 7.0, "node4": 6.0},
			vcJob:     vcJobWithPrefMap(prefMap),
			wantNode1: 8.0,
			wantNode2: 8.0 + defaultPreferPreviousScore,
		},
		{
			name: "17-fault: no otherNodes, fallback to selfNode",
			sHandle: func() *ScheduleHandler {
				h := enabledFrame()
				h.FaultHandle = &mockFaultHandler{faultByRank: true}
				return h
			}(),
			task:  baseTask,
			vcJob: vcJobWithPrefMap(prefMap),
			scoreMap: func() map[string]float64 {
				return map[string]float64{"node1": 8.0, "node2": 7.5}
			}(),
			wantNode1: 8.0 + defaultPreferPreviousScore,
			wantNode2: 7.5,
		},
		{
			name:    "18-rank resolved from task index fallback gets boosted",
			sHandle: enabledFrame(),
			task: &api.TaskInfo{
				Pod:  &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}},
				Job:  "job1",
				UID:  "t1",
				Name: "task1",
			},
			vcJob: SchedulerJob{
				Owner: OwnerInfo{OwnerReference: metav1.OwnerReference{UID: "owner-uid"}},
				SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{NPUTaskNum: 2, Tasks: fakeTasksForRank()},
				},
				PrefNodeMap: prefMap,
			},
			scoreMap:  map[string]float64{"node1": 8.0, "node2": 6.0, "node3": 7.0, "node4": 6.0},
			wantNode1: 8.0,
			wantNode2: 8.0 + defaultPreferPreviousScore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := baseScoreMap()
			if tt.scoreMap != nil {
				sm = tt.scoreMap
			}

			tt.sHandle.addPreferPreviousNodeScore(tt.task, sm, tt.vcJob)
			if sm["node1"] != tt.wantNode1 {
				t.Errorf("node1 = %v, want %v", sm["node1"], tt.wantNode1)
			}
			key2 := "node2"
			if _, ok := sm[key2]; !ok {
				key2 = "node3"
			}
			if tt.wantNode2 != 0 && sm[key2] != tt.wantNode2 {
				t.Errorf("%s = %v, want %v", key2, sm[key2], tt.wantNode2)
			}
		})
	}
}
func TestIsFaultPod(t *testing.T) {
	baseJob := SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: fakeTasksForRank()}}}
	tests := []struct {
		name    string
		sHandle *ScheduleHandler
		task    *api.TaskInfo
		want    bool
	}{
		{
			name:    "01-nil FaultHandle returns false",
			sHandle: &ScheduleHandler{},
			task:    &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}}, Job: "job1"},
			want:    false,
		},
		{
			name: "02-fault task rank returns true",
			sHandle: &ScheduleHandler{
				FaultHandle: &fakeFaultHandle{isFaultTaskFn: func(j api.JobID, r string) bool { return j == "job1" && r == "0" }},
			},
			task: &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}}, Job: "job1"},
			want: true,
		},
		{
			name: "03-not fault task returns false",
			sHandle: &ScheduleHandler{
				FaultHandle: &fakeFaultHandle{isFaultTaskFn: func(j api.JobID, r string) bool { return false }},
			},
			task: &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "0"}}}, Job: "job1"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sHandle.isFaultPod(tt.task, baseJob); got != tt.want {
				t.Errorf("isFaultPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveRankIndex(t *testing.T) {
	tests := []struct {
		name string
		task *api.TaskInfo
		job  SchedulerJob
		want string
	}{
		{
			name: "01-annotation present returns rank",
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: "5"}}},
				UID: "uid-1",
			},
			job:  SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"uid-1": {Index: 3}}}}},
			want: "5",
		},
		{
			name: "02-no annotation falls back to Index",
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}},
				UID: "uid-1",
			},
			job:  SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"uid-1": {Index: 3}}}}},
			want: "3",
		},
		{
			name: "03-task not found returns empty",
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}},
				UID: "uid-missing",
			},
			job:  SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{}}}},
			want: "",
		},
		{
			name: "04-empty annotation falls back to Index",
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{PodRankIndexKey: ""}}},
				UID: "uid-1",
			},
			job:  SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{Tasks: map[api.TaskID]util.NPUTask{"uid-1": {Index: 7}}}}},
			want: "7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (&ScheduleHandler{}).resolveRankIndex(tt.task, tt.job); got != tt.want {
				t.Errorf("resolveRankIndex() = %q, want %q", got, tt.want)
			}
		})
	}
}

func fakeTasksForRank() map[api.TaskID]util.NPUTask {
	return map[api.TaskID]util.NPUTask{
		"t0": {Annotation: map[string]string{PodRankIndexKey: "0"}, Index: 0},
		"t1": {Annotation: map[string]string{PodRankIndexKey: "1"}, Index: 1},
		"t2": {Annotation: map[string]string{PodRankIndexKey: "2"}, Index: 2},
	}
}

// mockFaultHandler implements FaultHandler for testing, allowing control over
// IsFaultTaskByRank return value without importing the rescheduling package.
type mockFaultHandler struct {
	faultByRank bool
}

func (m *mockFaultHandler) Execute(*ScheduleEnv, *framework.Session) error      { return nil }
func (m *mockFaultHandler) CheckNodeNPUByTask(*api.TaskInfo, *NPUNode) error    { return nil }
func (m *mockFaultHandler) ScoreBestNPUNodes(*api.TaskInfo, map[string]float64) {}
func (m *mockFaultHandler) UseAnnotation(*api.TaskInfo)                         {}
func (m *mockFaultHandler) PreStopAction(*ScheduleEnv) error                    { return nil }
func (m *mockFaultHandler) IsNodeFault(nodeName string) bool                    { return false }
func (m *mockFaultHandler) IsFaultTaskByRank(jobID api.JobID, rankIndex string) bool {
	return m.faultByRank
}
