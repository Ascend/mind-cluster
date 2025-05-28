/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package nslb is using for HuaWei Ascend pin tor affinity.
*/
package nslb

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// PtrInit return base type ptr
func ptrInit[T any](v T) *T { return &v }

type InitDPPolicyHandlerTest struct {
	name string
	attr util.SchedulerJobAttr
	env  plugin.ScheduleEnv
	want plugin.SchedulerPluginNeed
	ok   bool
}

func buildInitDPPolicyHandlerTestCases() []InitDPPolicyHandlerTest {
	defaultJob := util.SchedulerJobAttr{ComJob: util.ComJob{Name: "test-job"}}
	defaultJob.NPUJob = &util.NPUJob{SpBlockNPUNum: util.NPUIndex16}
	validTor := &plugin.TorList{
		Tors: []*plugin.Tor{{IP: "test-ip", Servers: []*plugin.Server{{Name: "test-node"}}}},
	}
	validNodes := map[string]plugin.NPUNode{"test-node": {CommonNode: plugin.CommonNode{SuperPodID: 1}}}
	validEnv := plugin.ScheduleEnv{ClusterCache: plugin.ClusterCache{Tors: validTor, Nodes: validNodes,
		Jobs: map[api.JobID]plugin.SchedulerJob{"test-job": {SchedulerJobAttr: defaultJob}}}}
	return []InitDPPolicyHandlerTest{
		{
			name: "01 will return nil when env.Tors is nil",
			attr: defaultJob,
			env:  plugin.ScheduleEnv{ClusterCache: plugin.ClusterCache{Tors: nil}},
			want: nil, ok: false,
		},
		{
			name: "02 will return nil when job not exist in env",
			attr: util.SchedulerJobAttr{ComJob: util.ComJob{Name: "not-exist-job"}},
			env:  validEnv,
			want: nil, ok: false,
		},
		{
			name: "03 will return TorHandlerDP when initialization success",
			attr: defaultJob,
			env:  validEnv,
			want: &TorHandlerDP{TorHandler: TorHandler{pluginName: pluginName, globalTorEnv: validTor,
				Job: &plugin.SchedulerJob{SchedulerJobAttr: defaultJob}},
				vPodSize:     1,
				superPodTors: map[int32][]*plugin.Tor{1: validTor.Tors},
			}, ok: true,
		},
	}
}

func TestInitDPPolicyHandler(t *testing.T) {
	for _, tt := range buildInitDPPolicyHandlerTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := InitDPPolicyHandler(tt.attr, tt.env)
			if !reflect.DeepEqual(got, tt.want) || ok != tt.ok {
				t.Errorf("InitDPPolicyHandler() got = %v, %v, want %v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

type setJobServerListTest struct {
	name    string
	Tors    []*plugin.Tor
	taskNum int
	want    []*plugin.Tor
}

func buildSetJobServerListTestCases() []setJobServerListTest {
	tmpJobId := api.JobID("test-job")
	return []setJobServerListTest{
		{
			name:    "01 taskNum is 0 not change ServerList",
			Tors:    []*plugin.Tor{{IP: "tor1", Servers: []*plugin.Server{{CurrentJob: ptrInit(tmpJobId)}}}},
			taskNum: 0,
			want:    nil,
		},
		{
			name: "02 normal allocate ServerList",
			Tors: []*plugin.Tor{{IP: "tor1", FreeServerCount: util.NPUIndex2,
				Servers: []*plugin.Server{
					{Name: "server1", CurrentJob: ptrInit(tmpJobId)},
					{Name: "server2", CurrentJob: ptrInit(tmpJobId)}}}},
			taskNum: util.NPUIndex2,
			want: []*plugin.Tor{{IP: "tor1", FreeServerCount: util.NPUIndex2,
				Servers: []*plugin.Server{
					{Name: "server1", CurrentJob: ptrInit(tmpJobId)},
					{Name: "server2", CurrentJob: ptrInit(tmpJobId)}}}},
		},
	}
}

func TestSetJobServerList(t *testing.T) {
	for _, tt := range buildSetJobServerListTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			th := &TorHandlerDP{TorHandler: TorHandler{Job: &plugin.SchedulerJob{
				SchedulerJobAttr: util.SchedulerJobAttr{ComJob: util.ComJob{Name: "test-job"}}}},
			}
			th.setJobServerList(tt.Tors, tt.taskNum)
			if !reflect.DeepEqual(th.ServerList, tt.want) {
				t.Errorf("setJobServerList() got = %v, want %v", th.ServerList, tt.want)
			}
		})
	}
}

type setPartialTorsTest struct {
	name            string
	useSuperPodTors []*superPodTors
	jobNPUTaskNum   int
	vPodSize        int
	wantPartial     int
}

func buildSetPartialTorsTestCases() []setPartialTorsTest {
	return []setPartialTorsTest{
		{
			name: "01-When full tors are enough, should not set partial tors",
			useSuperPodTors: []*superPodTors{
				{
					full: util.NPUIndex16, partial: 0, torCount: util.NPUIndex8,
				},
			},
			jobNPUTaskNum: util.NPUIndex8, vPodSize: util.NPUIndex1,
			wantPartial: 0,
		},
		{
			name: "02-When full tors are not enough, should set partial tors",
			useSuperPodTors: []*superPodTors{
				{
					full: util.NPUIndex8, partial: 0, torCount: util.NPUIndex8,
					partialTors: [util.NPUIndex3][]*plugin.Tor{
						{{FreeServerCount: util.NPUIndex4}},
					},
				},
			},
			jobNPUTaskNum: util.NPUIndex16, vPodSize: 1,
			wantPartial: util.NPUIndex4,
		},
		{
			name: "03-Should accumulate partial tors across stages",
			useSuperPodTors: []*superPodTors{
				{
					full: util.NPUIndex8, partial: 0, torCount: util.NPUIndex8,
					partialTors: [util.NPUIndex3][]*plugin.Tor{
						{{FreeServerCount: util.NPUIndex2}},
						{{FreeServerCount: util.NPUIndex2}},
					},
				},
			},
			jobNPUTaskNum: util.NPUIndex16, vPodSize: util.NPUIndex1,
			wantPartial: util.NPUIndex4,
		},
	}
}

func TestSetPartialTors(t *testing.T) {
	for _, tt := range buildSetPartialTorsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			th := &TorHandlerDP{
				useSuperPodTors: tt.useSuperPodTors,
				TorHandler: TorHandler{Job: &plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{
					NPUJob: &util.NPUJob{NPUTaskNum: tt.jobNPUTaskNum}}}},
				vPodSize: tt.vPodSize,
			}
			th.setPartialTors()
			for _, spt := range th.useSuperPodTors {
				if spt.partial != tt.wantPartial {
					t.Errorf("setPartialTors() got partial = %d, want %d", spt.partial, tt.wantPartial)
				}
			}
		})
	}
}

type setTorAffinityJobNodesScoreTest struct {
	name     string
	th       *TorHandlerDP
	task     *api.TaskInfo
	nodeMaps map[string]*api.NodeInfo
	scoreMap map[string]float64
	wantErr  bool
}

func buildSetTorAffinityJobNodesScoreTestCases() []setTorAffinityJobNodesScoreTest {
	return []setTorAffinityJobNodesScoreTest{
		{
			name:     "01-will return nil when job not ready",
			th:       &TorHandlerDP{TorHandler: TorHandler{Job: &plugin.SchedulerJob{JobReadyTag: ptrInit(false)}}},
			task:     &api.TaskInfo{},
			nodeMaps: make(map[string]*api.NodeInfo),
			scoreMap: make(map[string]float64),
			wantErr:  false,
		},
		{
			name: "02-will return nil when server list exists",
			th: &TorHandlerDP{TorHandler: TorHandler{Job: &plugin.SchedulerJob{JobReadyTag: ptrInit(true)},
				ServerList: []*plugin.Tor{{}}}},
			task:     &api.TaskInfo{},
			nodeMaps: make(map[string]*api.NodeInfo),
			scoreMap: make(map[string]float64),
			wantErr:  false,
		},
		{
			name: "03-will return nil when job score success",
			th: &TorHandlerDP{
				TorHandler: TorHandler{globalTorEnv: &plugin.TorList{},
					Job: &plugin.SchedulerJob{JobReadyTag: ptrInit(true), SchedulerJobAttr: util.SchedulerJobAttr{
						ComJob: util.ComJob{Name: "test-job"}, NPUJob: &util.NPUJob{NPUTaskNum: util.NPUIndex2}}}},
				superPodTors: map[int32][]*plugin.Tor{util.NPUIndex1: {{IP: "test-ip"}}},
				vPodSize:     util.NPUIndex1},
			task:     &api.TaskInfo{},
			nodeMaps: make(map[string]*api.NodeInfo),
			scoreMap: make(map[string]float64),
			wantErr:  false,
		},
	}
}

func TestSetTorAffinityJobNodesScore(t *testing.T) {
	for _, tt := range buildSetTorAffinityJobNodesScoreTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.th.setTorAffinityJobNodesScore(tt.task, tt.nodeMaps, tt.scoreMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("setTorAffinityJobNodesScore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTorHandlerDPUseAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		th       *TorHandlerDP
		task     *api.TaskInfo
		node     plugin.NPUNode
		wantNode *plugin.NPUNode
	}{
		{
			name: "01 set default annotations when globalTorEnv is nil",
			th:   &TorHandlerDP{TorHandler: TorHandler{globalTorEnv: nil}},
			task: &api.TaskInfo{
				Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string)}}},
			node:     plugin.NPUNode{},
			wantNode: &plugin.NPUNode{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.th.UseAnnotation(tt.task, tt.node)
			if !reflect.DeepEqual(got, tt.wantNode) {
				t.Errorf("UseAnnotation() = %v, want %v", got, tt.wantNode)
			}
		})
	}
}
