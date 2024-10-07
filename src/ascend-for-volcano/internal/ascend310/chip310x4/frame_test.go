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
Package chip310x4 is using for HuaWei Ascend pin affinity schedule.
*/
package chip310x4

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

const (
	npuNum2 = 2
)

type validNPUJobTestCase struct {
	name    string
	attr    util.SchedulerJobAttr
	wantErr *api.ValidateResult
}

func buildValidNPUJobTestCase01() []validNPUJobTestCase {
	job01 := test.FakeNormalTestJob("job01", 1)
	test.SetFakeJobResRequest(job01, util.NPU310CardName, "1")
	attr1 := itest.FakeSchedulerJobAttrByJob(job01)
	for _, task := range attr1.Tasks {
		fmt.Println(task.Name)
	}
	job02 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job02, util.NPU310CardName, "65")
	attr2 := itest.FakeSchedulerJobAttrByJob(job02)
	for _, task := range attr2.Tasks {
		fmt.Println(task.Name)
	}
	job03 := test.FakeNormalTestJob("job02", 1)
	test.SetFakeJobResRequest(job03, util.NPU310CardName, "2")
	attr3 := itest.FakeSchedulerJobAttrByJob(job03)
	return []validNPUJobTestCase{
		{
			name:    "01-ValidNPUJob should return nil when job request no npu",
			attr:    attr1,
			wantErr: nil,
		},
		{
			name: "02-ValidNPUJob should return error when tasks request npu more than 64",
			attr: attr2,
			wantErr: &api.ValidateResult{
				Pass:    false,
				Reason:  "task req npu num is invalid",
				Message: "task<vcjob/job02-pod0> req npu num<65> is invalid",
			},
		},
		{
			name:    "03-ValidNPUJob should return nil when tasks request is valid",
			attr:    attr3,
			wantErr: nil,
		},
	}
}

// TestValidNPUJob
func TestValidNPUJob(t *testing.T) {
	npu := New(SchedulerName)
	testCases := buildValidNPUJobTestCase01()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			npu.SetSchedulerAttr(tt.attr)
			if err := npu.ValidNPUJob(); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidNPUJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func buildUseAnnotationTestCase1() itest.UseAnnotationTestCase {
	return itest.UseAnnotationTestCase{
		Name: "01-UseAnnotation task will select the npu which is on the card that has 1 npu other than 2",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-4,Ascend310-5"},
			},
		},
		PodAnno: "Ascend310-0",
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Allocate: map[v1.ResourceName]float64{util.NPU310CardName: npuNum2 * util.NPUHexKilo},
			},
		},
	}
}

func buildUseAnnotationTestCase2() itest.UseAnnotationTestCase {
	return itest.UseAnnotationTestCase{
		Name: "02-UseAnnotation task will select the npu which is on the card that has 2 npu other than 3",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-4," +
					"Ascend310-5,Ascend310-6"},
			},
		},
		PodAnno: "Ascend310-0",
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-1,Ascend310-4,Ascend310-5,Ascend310-6"},
			},
		},
	}
}

func buildUseAnnotationTestCase3() itest.UseAnnotationTestCase {
	return itest.UseAnnotationTestCase{
		Name: "03-UseAnnotation task will select the npu which is on the card that has 3 npu other than 4",
		Task: test.FakeTaskWithResReq("pod0", util.NPU310CardName, 1),
		Node: plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-0,Ascend310-1,Ascend310-2,Ascend310-4," +
					"Ascend310-5,Ascend310-6,Ascend310-7"},
			},
		},
		PodAnno: "Ascend310-0",
		WantNode: &plugin.NPUNode{
			CommonNode: plugin.CommonNode{
				Annotation: map[string]string{util.NPU310CardName: "Ascend310-1,Ascend310-2,Ascend310-4,Ascend310-5," +
					"Ascend310-6,Ascend310-7"},
			},
		},
	}
}

func buildUseAnnotationTestCases() []itest.UseAnnotationTestCase {
	return []itest.UseAnnotationTestCase{
		buildUseAnnotationTestCase1(),
		buildUseAnnotationTestCase2(),
		buildUseAnnotationTestCase3(),
	}
}

// TestUseAnnotation
func TestUseAnnotation(t *testing.T) {
	npu := New(SchedulerName)
	job := test.FakeNormalTestJob("job", 1)
	test.SetFakeJobResRequest(job, util.NPU310CardName, "1")
	attr := itest.FakeSchedulerJobAttrByJob(job)
	env := plugin.ScheduleEnv{
		Jobs: map[api.JobID]plugin.SchedulerJob{
			test.FakeJobName: {SchedulerJobAttr: attr},
		},
	}
	npu.SetSchedulerAttr(attr)
	npu.SetSchedulerEnv(env)
	testCases := buildUseAnnotationTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			node := npu.UseAnnotation(tt.Task, tt.Node)
			if !reflect.DeepEqual(node.Annotation, tt.Node.Annotation) ||
				!reflect.DeepEqual(tt.Task.Pod.Annotations[util.NPU310CardName], tt.PodAnno) {
				t.Errorf("UseAnnotation() node: %v, wantNode: %v, anno %v, wantAnno %v",
					node, tt.WantNode, tt.Task.Pod.Annotations, tt.PodAnno)
			}
		})
	}
}
