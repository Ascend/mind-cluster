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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
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

type frameFields struct {
	UID        types.UID
	Conf       []config.Configuration
	KubeClient kubernetes.Interface
}

type AddDefaultSchedulerSelectorConfigTest struct {
	name   string
	fields frameFields
}

func buildAddDefaultSchedulerSelectorConfigTest() []AddDefaultSchedulerSelectorConfigTest {
	tests := []AddDefaultSchedulerSelectorConfigTest{
		{
			name:   "01-AddDefaultSchedulerSelectorConfig ok test.",
			fields: frameFields{Conf: nil},
		},
	}
	return tests
}

func TestVolcanoFrameAddDefaultSchedulerSelectorConfig(t *testing.T) {

	tests := buildAddDefaultSchedulerSelectorConfigTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vf := &VolcanoFrame{
				UID:        tt.fields.UID,
				Confs:      tt.fields.Conf,
				KubeClient: tt.fields.KubeClient,
			}
			vf.AddDefaultSchedulerSelectorConfig()
		})
	}
}
