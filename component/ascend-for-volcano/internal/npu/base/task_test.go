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
Package base is using for HuaWei Ascend pin affinity schedule.
*/
package base

import (
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type setRankIndexTestCase struct {
	Name              string
	Jobs              map[api.JobID]plugin.SchedulerJob
	Annotation        map[string]string
	Task              *api.TaskInfo
	WantRankIndex     string
	WantAnnotationSet bool
}

func buildSetRankIndexTestCases() []setRankIndexTestCase {
	taskUID := api.TaskID("task-uid-1")
	jobID := api.JobID("test-job")
	taskName := "test-pod"

	newTask := func(job api.JobID, uid api.TaskID, name string, annotations map[string]string) *api.TaskInfo {
		if annotations == nil {
			annotations = make(map[string]string)
		}
		return &api.TaskInfo{
			UID:  uid,
			Name: name,
			Job:  job,
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
			},
		}
	}

	newSchedulerJob := func(kind string, tasks map[api.TaskID]util.NPUTask) plugin.SchedulerJob {
		return plugin.SchedulerJob{
			Owner: plugin.OwnerInfo{
				OwnerReference: metav1.OwnerReference{
					Kind: kind,
				},
			},
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: &util.NPUJob{
					Tasks: tasks,
				},
			},
		}
	}

	taskWithIndex := map[api.TaskID]util.NPUTask{
		taskUID: {Index: 3},
	}

	return []setRankIndexTestCase{
		{
			Name:              "01-job not found in tp.Jobs",
			Jobs:              map[api.JobID]plugin.SchedulerJob{},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "",
			WantAnnotationSet: false,
		},
		{
			Name: "02-pod already has rankIndex set",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob(plugin.ReplicaSetType, taskWithIndex),
			},
			Task: newTask(jobID, taskUID, taskName, map[string]string{
				plugin.PodRankIndexKey: "5",
			}),
			WantRankIndex:     "5",
			WantAnnotationSet: true,
		},
		{
			Name: "03-ReplicaSet owner sets rankIndex to task index",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob(plugin.ReplicaSetType, taskWithIndex),
			},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "3",
			WantAnnotationSet: true,
		},
		{
			Name: "04-MinAvailableKey present, no ReplicaSet owner",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob("Deployment", taskWithIndex),
			},
			Annotation: map[string]string{
				util.MinAvailableKey: "true",
			},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "3",
			WantAnnotationSet: true,
		},
		{
			Name: "05-both ReplicaSet and MinAvailableKey, ReplicaSet path wins",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob(plugin.ReplicaSetType, taskWithIndex),
			},
			Annotation: map[string]string{
				util.MinAvailableKey: "true",
			},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "3",
			WantAnnotationSet: true,
		},
		{
			Name: "06-neither ReplicaSet nor MinAvailableKey",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob("StatefulSet", taskWithIndex),
			},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "",
			WantAnnotationSet: false,
		},
		{
			Name: "07-MinAvailableKey present but rankIndex already empty-string set",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob("Deployment", taskWithIndex),
			},
			Annotation: map[string]string{
				util.MinAvailableKey: "true",
			},
			Task: newTask(jobID, taskUID, taskName, map[string]string{
				plugin.PodRankIndexKey: "",
			}),
			WantRankIndex:     "3",
			WantAnnotationSet: true,
		},
		{
			Name: "08-task UID not in job.Tasks, zero Index used",
			Jobs: map[api.JobID]plugin.SchedulerJob{
				jobID: newSchedulerJob(plugin.ReplicaSetType, map[api.TaskID]util.NPUTask{}),
			},
			Task:              newTask(jobID, taskUID, taskName, map[string]string{}),
			WantRankIndex:     "0",
			WantAnnotationSet: true,
		},
	}
}

// TestSetRankIndex tests all branches of the setRankIndex method.
func TestSetRankIndex(t *testing.T) {
	testCases := buildSetRankIndexTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			tp := &NPUHandler{
				ScheduleEnv: plugin.ScheduleEnv{
					ClusterCache: plugin.ClusterCache{
						Jobs: tt.Jobs,
					},
				},
				SchedulerJobAttr: util.SchedulerJobAttr{
					ComJob: util.ComJob{
						Annotation: tt.Annotation,
					},
				},
			}

			tp.setRankIndex(tt.Task)

			got, exists := tt.Task.Pod.Annotations[plugin.PodRankIndexKey]
			if tt.WantAnnotationSet {
				if !exists {
					t.Errorf("setRankIndex() want annotation %s to be set, but it was not present",
						plugin.PodRankIndexKey)
					return
				}
				if got != tt.WantRankIndex {
					t.Errorf("setRankIndex() rankIndex = %q, want %q", got, tt.WantRankIndex)
				}
			} else {
				if exists && got != "" {
					t.Errorf("setRankIndex() rankIndex = %q, want not set", got)
				}
			}
		})
	}
}
