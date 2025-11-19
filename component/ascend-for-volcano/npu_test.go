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
Package main is using for HuaWei Ascend pin affinity schedule.
*/
package main

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type handlerStartTest struct {
	name string
	want *plugin.ScheduleHandler
}

func buildTestHandlerStartTestCases() []handlerStartTest {
	testCases := []handlerStartTest{
		{
			name: "HandlerStart ok test",
			want: &plugin.ScheduleHandler{},
		},
	}
	return testCases
}

func TestHandlerStart(t *testing.T) {
	tests := buildTestHandlerStartTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandlerStart(); got == nil {
				t.Errorf("HandlerStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name string
		tp   *huaweiNPUPlugin
		want string
	}{
		{
			name: "01-Name ok test",
			tp:   &huaweiNPUPlugin{},
			want: PluginName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tp.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		arguments framework.Arguments
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "New ok test",
			args: args{arguments: framework.Arguments{PluginName: "haha"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.arguments); got == nil {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fields struct {
	Scheduler *plugin.ScheduleHandler
	Arguments framework.Arguments
}

type args struct {
	ssn            *framework.Session
	cacheFunBefore func()
	cacheFunAfter  func()
}

type onSessionOpenTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionOpenTestCases() []onSessionOpenTest {
	tests := []onSessionOpenTest{
		{
			name:   "OnSessionOpen test ssn nil ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: nil, cacheFunBefore: func() {}, cacheFunAfter: func() {}},
		},
	}
	return tests
}

func TestOnSessionOpen(t *testing.T) {
	tests := buildOnSessionOpenTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tt.args.cacheFunBefore()
			tp.OnSessionOpen(tt.args.ssn)
			tt.args.cacheFunAfter()
		})
	}
}

type onSessionCloseTest struct {
	name   string
	fields fields
	args   args
}

func buildOnSessionCloseTestCases() []onSessionCloseTest {
	testSsn := test.FakeNormalSSN(nil)
	tests := []onSessionCloseTest{
		{
			name:   "OnSessionCloseTestCases test ok",
			fields: fields{Scheduler: HandlerStart()},
			args:   args{ssn: testSsn},
		},
	}
	return tests
}

type jobEnqueueableTestCase struct {
	name           string
	mocks          func() *gomonkey.Patches
	job            interface{}
	plugin         *huaweiNPUPlugin
	expectedResult int
}

func buildJobEnqueueableTestCases01() []jobEnqueueableTestCase {
	mockJobInfo := &api.JobInfo{
		Name: "test-job",
	}
	mockPlugin := &huaweiNPUPlugin{
		Scheduler: &plugin.ScheduleHandler{
			NPUPlugins: sets.NewString(),
		},
	}
	tests := []jobEnqueueableTestCase{
		{
			name: "01-when npu-plusgins is bil should return skip",
			job:  mockJobInfo,
			plugin: &huaweiNPUPlugin{
				Scheduler: &plugin.ScheduleHandler{},
			},
			expectedResult: util.JobEnqueueSkip,
		},
		{
			name:           "02-when job is not api.JobInfo should return skip",
			job:            "not-a-job-info",
			plugin:         mockPlugin,
			expectedResult: util.JobEnqueueSkip,
		},
		{
			name: "03-when req npu name is not in supported should return skip",
			mocks: func() *gomonkey.Patches {
				return gomonkey.ApplyFunc(plugin.GetVCJobReqNPUTypeFromJobInfo,
					func(job *api.JobInfo) (string, int, error) {
						return "non-existent-npu", 2, nil
					})
			},
			job:            mockJobInfo,
			plugin:         mockPlugin,
			expectedResult: util.JobEnqueueSkip,
		},
	}
	return tests
}

func patchWithArgs(jobNpu, clusterNpu int) *gomonkey.Patches {
	return gomonkey.ApplyFunc(plugin.GetVCJobReqNPUTypeFromJobInfo, func(job *api.JobInfo) (string,
		int, error) {
		return "available-npu", jobNpu, nil
	}).ApplyFunc(getNpuNum, func(ssn *framework.Session, tp *huaweiNPUPlugin, npuName string) int {
		return clusterNpu
	})
}

func buildJobEnqueueableTestCases02() []jobEnqueueableTestCase {
	mockJobInfo := &api.JobInfo{
		Name: "test-job",
	}

	tests := []jobEnqueueableTestCase{
		{
			name:  "04-when req npu num is not satisfied should return not-enqueue",
			mocks: func() *gomonkey.Patches { return patchWithArgs(5, 2) },
			job:   mockJobInfo,
			plugin: &huaweiNPUPlugin{
				Scheduler: &plugin.ScheduleHandler{
					NPUPlugins: sets.NewString("available-npu"),
				},
			},
			expectedResult: util.JobNotEnqueue,
		},
		{
			name:  "05-when req npu num is satisfied and enqueue is force should return enqueue",
			mocks: func() *gomonkey.Patches { return patchWithArgs(2, 5) },
			job:   mockJobInfo,
			plugin: func() *huaweiNPUPlugin {
				plg := &huaweiNPUPlugin{
					Scheduler: &plugin.ScheduleHandler{
						NPUPlugins: sets.NewString("available-npu"),
					},
				}
				plg.Scheduler.FrameAttr.ForceEnqueue = true
				return plg
			}(),
			expectedResult: util.JobEnqueue,
		},
		{
			name:  "06-when req npu num is satisfied and enqueue is not force should return skip",
			mocks: func() *gomonkey.Patches { return patchWithArgs(2, 5) },
			job:   mockJobInfo,
			plugin: func() *huaweiNPUPlugin {
				plg := &huaweiNPUPlugin{
					Scheduler: &plugin.ScheduleHandler{
						NPUPlugins: sets.NewString("available-npu"),
					},
				}
				plg.Scheduler.FrameAttr.ForceEnqueue = false
				return plg
			}(),
			expectedResult: util.JobEnqueueSkip,
		},
	}
	return tests
}

func TestJobEnqueueable(t *testing.T) {
	tests := append(buildJobEnqueueableTestCases01(), buildJobEnqueueableTestCases02()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mocks != nil {
				patch := tt.mocks()
				defer patch.Reset()
			}
			if result := jobEnqueueable(tt.job, &framework.Session{}, tt.plugin); result != tt.expectedResult {
				t.Errorf("jobEnqueueable() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestOnSessionClose(t *testing.T) {
	tests := buildOnSessionCloseTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &huaweiNPUPlugin{
				Scheduler: tt.fields.Scheduler,
				Arguments: tt.fields.Arguments,
			}
			tp.OnSessionClose(tt.args.ssn)
		})
	}
}

func newNPUNode(name string, anno, idle int) plugin.NPUNode {
	return plugin.NPUNode{
		CommonNode: plugin.CommonNode{
			Name: name,
			Idle: map[v1.ResourceName]float64{
				"huawei.com/Ascend910": float64(idle * 1000),
			},
			Capability: map[v1.ResourceName]float64{
				"huawei.com/Ascend910": float64(idle * 1000),
			},
			Annotation: map[string]string{
				"huawei.com/Ascend910": func() string {
					annoStr := ""
					for i := 0; i < anno; i++ {
						annoStr += strconv.Itoa(i)
						if i < anno-1 {
							annoStr += ","
						}
					}
					return annoStr
				}(),
			},
		},
	}
}

// TestGetNpuNum tests the getNpuNum function
func TestGetNpuNum(t *testing.T) {
	ssn := &framework.Session{
		Nodes: map[string]*api.NodeInfo{
			"node1": {Name: "node1"},
			"node2": {Name: "node2"},
			"node3": {Name: "node3"},
			"node4": {Name: "node4"},
		},
	}
	node1 := newNPUNode("node1", 4, 4)
	node3 := newNPUNode("node3", 1, 2)
	node4 := newNPUNode("node4", 0, 0)
	plg := &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{}}
	plg.Scheduler.Nodes = map[string]plugin.NPUNode{
		"node1": node1,
		"node3": node3,
		"node4": node4,
	}
	expectedResult := 5
	t.Run("test GetNpuNum", func(t *testing.T) {
		if result := getNpuNum(ssn, plg, "huawei.com/Ascend910"); result != expectedResult {
			t.Errorf("getNpuNum() = %v, want %v", result, expectedResult)
		}
	})
}

func boolPointer(b bool) *bool {
	return &b
}
func TestJobPipelined(t *testing.T) {
	testCases := []struct {
		name string
		job  interface{}
		plg  *huaweiNPUPlugin
		want int
	}{
		{
			name: "01-job is not *api.JobInfo should return reject",
			job:  &api.TaskInfo{},
			plg:  &huaweiNPUPlugin{},
			want: util.Reject,
		},
		{
			name: "02-job is not in cache should return abstain",
			job:  &api.JobInfo{UID: "test-job-uid"},
			plg:  &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{}},
			want: util.Abstain,
		},
		{
			name: "03-job is not ready return reject",
			job:  &api.JobInfo{UID: "test-job-uid"},
			plg: func() *huaweiNPUPlugin {
				p := &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{}}
				p.Scheduler.Jobs = map[api.JobID]plugin.SchedulerJob{
					"test-job-uid": {JobReadyTag: boolPointer(false)},
				}
				return p
			}(),
			want: util.Reject,
		},
		{
			name: "04-job is ready should return abstain",
			job:  &api.JobInfo{UID: "test-job-uid"},
			plg: func() *huaweiNPUPlugin {
				p := &huaweiNPUPlugin{Scheduler: &plugin.ScheduleHandler{}}
				p.Scheduler.Jobs = map[api.JobID]plugin.SchedulerJob{
					"test-job-uid": {JobReadyTag: boolPointer(true)},
				}
				return p
			}(),
			want: util.Abstain,
		},
	}
	for _, tc := range testCases {
		got := jobPipelined(tc.job, tc.plg)
		if got != tc.want {
			t.Errorf("jobEnqueueable(%v, %v) = %v, want %v", tc.job, tc.plg, got, tc.want)
		}
	}
}

func mockJobWithAnnotation(anno map[string]string, uid, phase string) *api.JobInfo {
	return &api.JobInfo{
		Queue: "mockQueue",
		Name:  uid,
		UID:   api.JobID(uid),
		PodGroup: &api.PodGroup{PodGroup: scheduling.PodGroup{
			ObjectMeta: metav1.ObjectMeta{Annotations: anno},
			Status:     scheduling.PodGroupStatus{Phase: scheduling.PodGroupPhase(phase)},
		}}}
}

type testJobOrderType struct {
	name string
	job1 interface{}
	job2 interface{}
	want int
}

func mockJobOrderTestCase() []testJobOrderType {
	return []testJobOrderType{
		{
			name: "job1 is not JobInfo Type, order same",
			job1: &api.TaskInfo{},
			job2: &api.JobInfo{},
			want: util.JobOrderSamePriority,
		},
		{
			name: "job2 is not JobInfo Type, order same",
			job1: &api.JobInfo{},
			job2: &api.TaskInfo{},
			want: util.JobOrderSamePriority,
		},
		{
			name: "job1 dequeueTimes anno is invalid, order same",
			job1: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1.1"}, "job1", ""),
			job2: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job2", ""),
			want: util.JobOrderSamePriority,
		},
		{
			name: "job2 dequeueTimes anno is invalid, order same",
			job1: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job1", ""),
			job2: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "test"}, "job2", ""),
			want: util.JobOrderSamePriority,
		},
		{
			name: "job1 dequeueTimes anno value is lower than job2's, order high",
			job1: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job1", ""),
			job2: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "2"}, "job2", ""),
			want: util.JobOrderHighPriority,
		},
		{
			name: "job1 dequeueTimes anno value is higher than job2's, order low",
			job1: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "2"}, "job1", ""),
			job2: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job2", ""),
			want: util.JobOrderLowPriority,
		},
		{
			name: "job1 dequeueTimes anno value is equal to job2's, order same",
			job1: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job1", ""),
			job2: mockJobWithAnnotation(map[string]string{util.DequeueFrequencyAnnoKey: "1"}, "job2", ""),
			want: util.JobOrderSamePriority,
		},
	}
}

func TestJobOrderFn(t *testing.T) {
	tests := mockJobOrderTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jobOrderFn(tt.job1, tt.job2)
			if got != tt.want {
				t.Errorf("jobOrderFn got %v, want %v", got, tt.want)
			}
		})
	}
}

type testUpdatePgAnnotationType struct {
	name  string
	job   *api.JobInfo
	mocks func() *gomonkey.Patches
	want  map[string]string
}

func mockUpdatePgAnnotationTestCase() []testUpdatePgAnnotationType {
	return []testUpdatePgAnnotationType{
		{
			name: "Inqueue job with annotation, should not update annotation",
			job: mockJobWithAnnotation(map[string]string{
				util.EnqueueTimeAnnoKey: "1234",
			}, "job1", util.PodGroupInqueue),
			want: map[string]string{util.EnqueueTimeAnnoKey: "1234"},
		},
		{
			name: "Inqueue job without annotation, should add annotation",
			job:  mockJobWithAnnotation(nil, "job2", util.PodGroupInqueue),
			mocks: func() *gomonkey.Patches {
				return gomonkey.ApplyFunc(strconv.FormatInt,
					func(i int64, base int) string {
						return "1234"
					})
			},
			want: map[string]string{util.EnqueueTimeAnnoKey: "1234"},
		},
		{
			name: "running job with annotations, should delete annotations",
			job: mockJobWithAnnotation(map[string]string{
				util.EnqueueTimeAnnoKey:      "1234",
				util.DequeueFrequencyAnnoKey: "11",
			}, "job3", util.PodGroupRunning),
			want: map[string]string{},
		},
	}
}

func TestUpdatePgAnnotation(t *testing.T) {
	tests := mockUpdatePgAnnotationTestCase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var patch *gomonkey.Patches = nil
			if tt.mocks != nil {
				patch = tt.mocks()
			}
			ssn := &framework.Session{Jobs: map[api.JobID]*api.JobInfo{tt.job.UID: tt.job}}
			updatePgAnnotation(ssn)
			if patch != nil {
				patch.Reset()
			}
			got := tt.job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey]
			if !reflect.DeepEqual(tt.job.PodGroup.Annotations, tt.want) {
				t.Errorf("updatePgAnnotation got %v, want %v", got, tt.job.PodGroup.Annotations)
			}
		})
	}
}

type testJobDequeueForTimeoutType struct {
	name  string
	vcJob *api.JobInfo
	job   *api.JobInfo
	got   string
}

func mockJobDequeueForTimeoutTestCase() []testJobDequeueForTimeoutType {
	return []testJobDequeueForTimeoutType{
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: &api.JobInfo{Queue: "fakeQueue-1"},
			job:   mockJobWithAnnotation(map[string]string{}, "job2", util.PodGroupInqueue),
			got:   util.PodGroupInqueue,
		},
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupPending),
			job:   mockJobWithAnnotation(map[string]string{}, "job2", util.PodGroupPending),
			got:   util.PodGroupPending,
		},
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupPending),
			job:   mockJobWithAnnotation(map[string]string{}, "job2", util.PodGroupInqueue),
			got:   util.PodGroupInqueue,
		},
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupPending),
			job: mockJobWithAnnotation(map[string]string{
				util.EnableDequeueAnnoKey: util.EnableDequeueOnVal,
			}, "job2", util.PodGroupInqueue),
			got: util.PodGroupInqueue,
		},
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupPending),
			job: mockJobWithAnnotation(map[string]string{
				util.EnableDequeueAnnoKey: util.EnableDequeueOnVal,
				util.EnqueueTimeAnnoKey:   "test",
			}, "job2", util.PodGroupInqueue),
			got: util.PodGroupInqueue,
		},
		{
			name:  "job without dequeue annotation, should set to 1",
			vcJob: mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupPending),
			job: mockJobWithAnnotation(map[string]string{
				util.EnableDequeueAnnoKey: util.EnableDequeueOnVal,
				util.EnqueueTimeAnnoKey:   "1234",
			}, "job2", util.PodGroupInqueue),
			got: "",
		},
	}
}

func TestJobDequeueForTimeout(t *testing.T) {
	tests := mockJobDequeueForTimeoutTestCase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssn := &framework.Session{Jobs: map[api.JobID]*api.JobInfo{tt.job.UID: tt.job}}
			// Execute
			jobDequeueForTimeout(tt.vcJob, ssn)
			if tt.got != string(tt.job.PodGroup.Status.Phase) {
				t.Errorf("jobDequeueForTimeout got %v, want %v", tt.got, string(tt.job.PodGroup.Status.Phase))
			}
		})
	}
}

func TestExecJobDequeue(t *testing.T) {
	tests := []struct {
		name                 string
		job                  *api.JobInfo
		expectedDequeueTimes string
	}{
		{
			name:                 "job without dequeue annotation, should set to 1",
			job:                  mockJobWithAnnotation(map[string]string{}, "job1", util.PodGroupInqueue),
			expectedDequeueTimes: "1",
		},
		{
			name: "job with existing dequeue annotation should increment",
			job: mockJobWithAnnotation(map[string]string{
				util.DequeueFrequencyAnnoKey: "5",
			}, "job2", util.PodGroupInqueue),
			expectedDequeueTimes: "6",
		},
		{
			name: "job with invalid dequeue annotation should reset to 1",
			job: mockJobWithAnnotation(map[string]string{
				util.DequeueFrequencyAnnoKey: "5.5",
			}, "job3", util.PodGroupInqueue),
			expectedDequeueTimes: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssn := &framework.Session{Jobs: map[api.JobID]*api.JobInfo{}}
			execJobDequeue(ssn, tt.job)
			got := tt.job.PodGroup.Annotations[util.DequeueFrequencyAnnoKey]
			if got != tt.expectedDequeueTimes {
				t.Errorf("jobEnqueueable got %v, want %v", got, tt.expectedDequeueTimes)
			}
		})
	}
}
