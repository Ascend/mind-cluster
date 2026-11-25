/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package rescheduling

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
)

func TestGetPodFailedTime(t *testing.T) {
	earlier := time.Unix(1700000000, 0)
	later := time.Unix(1700000100, 0)

	tests := []struct {
		name    string
		task    *api.TaskInfo
		want    int64
		wantErr bool
	}{
		{
			name:    "nil task",
			task:    nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "nil pod",
			task:    &api.TaskInfo{},
			want:    0,
			wantErr: true,
		},
		{
			name: "latest terminated finish time wins",
			task: &api.TaskInfo{Pod: &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.Time{Time: earlier}}}},
				{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.Time{Time: later}}}},
			}}}},
			want:    later.Unix(),
			wantErr: false,
		},
		{
			name: "last termination state used",
			task: &api.TaskInfo{Pod: &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.Time{Time: earlier}}}},
			}}}},
			want:    earlier.Unix(),
			wantErr: false,
		},
		{
			name: "no terminated time falls back to now",
			task: &api.TaskInfo{Pod: &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
			}}}},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPodFailedTime(tt.task)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getPodFailedTime() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.name == "no terminated time falls back to now" {
				if got <= 0 {
					t.Fatalf("getPodFailedTime() = %d, want a positive fallback timestamp", got)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("getPodFailedTime() = %d, want %d", got, tt.want)
			}
		})
	}
}
