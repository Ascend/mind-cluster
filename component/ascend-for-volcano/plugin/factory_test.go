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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	task  *api.TaskInfo
	nodes []*api.NodeInfo
	ssn   *framework.Session
}

type batchNodeOrderFnTest struct {
	name    string
	fields  fields
	args    batchNodeOrderFnArgs
	want    map[string]float64
	wantErr bool
}

func buildBatchNodeOrderFn() []batchNodeOrderFnTest {
	tTask := test.FakeNormalTestTasks(1)[0]
	tNodes := test.FakeNormalTestNodes(util.NPUIndex2)
	tests := []batchNodeOrderFnTest{
		{
			name:    "01-BatchNodeOrderFn nil Test",
			fields:  fields{},
			args:    batchNodeOrderFnArgs{task: nil, nodes: nil, ssn: nil},
			want:    nil,
			wantErr: true,
		},
		{
			name: "02-BatchNodeOrderFn ScoreBestNPUNodes ok Test",
			fields: fields{NPUPlugins: map[string]NPUBuilder{},
				ScheduleEnv: ScheduleEnv{
					Jobs:      map[api.JobID]SchedulerJob{},
					Nodes:     map[string]NPUNode{},
					FrameAttr: VolcanoFrame{}}},
			args:    batchNodeOrderFnArgs{task: tTask, nodes: tNodes, ssn: nil},
			want:    nil,
			wantErr: false,
		},
	}
	return tests
}

func TestBatchNodeOrderFn(t *testing.T) {
	tests := buildBatchNodeOrderFn()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			got, err := sHandle.BatchNodeOrderFn(tt.args.task, tt.args.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchNodeOrderFn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchNodeOrderFn() got = %v, want %v", got, tt.want)
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
		return nil, nil
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
	name    string
	fields  fields
	args    initNPUSessionArgs
	wantErr bool
}

func buildInitNPUSessionTest() []initNPUSessionTest {
	tests := []initNPUSessionTest{
		{
			name:    "01-InitNPUSession nil ssn test",
			fields:  fields{},
			args:    initNPUSessionArgs{ssn: nil},
			wantErr: true,
		},
	}
	return tests
}

func TestInitNPUSession(t *testing.T) {
	tests := buildInitNPUSessionTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			if err := sHandle.InitNPUSession(tt.args.ssn); (err != nil) != tt.wantErr {
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

// TestUpdatePodGroupOfDeploy test of updatePodGroupOfDeploy
func TestUpdatePodGroupOfDeploy(t *testing.T) {
	t.Run("updatePodGroupOfDeploy", func(t *testing.T) {
		job := &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{},
			},
		}
		rs := &appv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{"xxx": "yyy"},
			},
		}
		expectedAnno := map[string]string{"xxx": "yyy"}
		updatePodGroupOfDeploy(job, rs)
		if !reflect.DeepEqual(expectedAnno, job.PodGroup.Annotations) {
			t.Errorf("updatePodGroupOfDeploy = %v, want %v", job.PodGroup.Annotations, expectedAnno)
		}
	})
}

// TestUpdatePodOfDeploy test of updatePodOfDeploy
func TestUpdatePodOfDeploy(t *testing.T) {
	t.Run("updatePodOfDeploy", func(t *testing.T) {
		job := &api.JobInfo{
			Tasks: map[api.TaskID]*api.TaskInfo{
				"task1": {
					Pod: &v1.Pod{},
				},
				"task2": {
					Pod: &v1.Pod{},
				},
			},
		}
		expectedRankIndexes := map[string]struct{}{"0": {}, "1": {}}
		updatePodOfDeploy(job)
		indexes := getJobRankIndexes(job)
		if !reflect.DeepEqual(expectedRankIndexes, indexes) {
			t.Errorf("updatePodOfDeploy = %v, want %v", indexes, expectedRankIndexes)
		}
	})
}

func getJobRankIndexes(job *api.JobInfo) map[string]struct{} {
	indexes := make(map[string]struct{}, 0)
	for _, task := range job.Tasks {
		if task.Pod == nil {
			continue
		}
		idx, ok := task.Pod.Annotations[podRankIndex]
		if !ok {
			continue
		}
		indexes[idx] = struct{}{}
	}
	return indexes
}

// TestRefreshJobWithIndex test of refreshJobWithIndex
func TestRefreshJobWithIndex(t *testing.T) {
	t.Run("refreshJobWithIndex", func(t *testing.T) {
		handler := &ScheduleHandler{
			ScheduleEnv: ScheduleEnv{
				JobWithIndex: map[api.JobID]struct{}{
					"job1": {},
					"job2": {},
				},
			},
		}
		ssn := &framework.Session{
			Jobs: map[api.JobID]*api.JobInfo{
				"job2": {},
			},
		}
		expectedJobWithIndex := map[api.JobID]struct{}{"job2": {}}
		handler.refreshJobWithIndex(ssn)
		if !reflect.DeepEqual(expectedJobWithIndex, handler.JobWithIndex) {
			t.Errorf("refreshJobWithIndex = %v, want %v", handler.JobWithIndex, expectedJobWithIndex)
		}
	})
}
