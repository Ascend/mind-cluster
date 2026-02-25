/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package schedulingexception is for collecting scheduling exception

package schedulingexception

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestCollector_Start(t *testing.T) {
	c := &Collector{
		jobExceptions: map[string]*jobExceptionInfo{},
		checkInterval: defaultCheckInterval,
	}
	const (
		timeout   = 100 * time.Millisecond
		totalTime = 150 * time.Millisecond
	)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go c.Start(ctx)
	time.Sleep(totalTime)

	if len(c.jobExceptions) != 0 {
		t.Errorf("expected empty jobs, got %d", len(c.jobExceptions))
	}
}

func TestCollector_ProcessPodGroupCreated(t *testing.T) {
	c := &Collector{}
	cond := c.processPodGroupCreated()

	if cond.Status != podGroupCreated {
		t.Errorf("expected status %s, got %s", podGroupCreated, cond.Status)
	}
	if cond.Reason != "PgNotInitialized" {
		t.Errorf("expected reason PgNotInitialized, got %s", cond.Reason)
	}
}

func TestCollector_ProcessPodGroupUnknown(t *testing.T) {
	c := &Collector{}

	tests := []struct {
		name     string
		pods     map[string]corev1.Pod
		expected jobStatus
	}{
		{
			name: "pending pod",
			pods: map[string]corev1.Pod{
				"pod1": {
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			expected: podGroupUnknown,
		},
		{
			name: "failed pod",
			pods: map[string]corev1.Pod{
				"pod1": {
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				},
			},
			expected: podGroupUnknown,
		},
		{
			name:     "no pods",
			pods:     map[string]corev1.Pod{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processPodGroupUnknown(tt.pods)
			if cond == nil && tt.expected != "" {
				t.Errorf("expected condition, got nil")
			}
			if cond != nil && cond.Status != tt.expected {
				t.Errorf("expected status %s, got %s", tt.expected, cond.Status)
			}
		})
	}
}

func TestCollector_ProcessPodGroupPending(t *testing.T) {
	c := &Collector{}
	tests := []struct {
		name     string
		pg       *v1beta1.PodGroup
		indices  conditionIndices
		expected string
	}{
		{
			name: "enqueue failed",
			pg: &v1beta1.PodGroup{
				Spec: v1beta1.PodGroupSpec{
					Queue: "test-queue",
				},
				Status: v1beta1.PodGroupStatus{
					Conditions: []v1beta1.PodGroupCondition{
						{
							Reason:  jobEnqueueFailedReason,
							Message: "enqueue failed message",
						},
					},
				},
			},
			indices: conditionIndices{
				jobEnqueueFailedIndex: 0,
			},
			expected: "enqueue failed message",
		},
		{
			name: "no enqueue failed",
			pg: &v1beta1.PodGroup{
				Spec: v1beta1.PodGroupSpec{
					Queue: "test-queue",
				},
			},
			indices: conditionIndices{
				jobEnqueueFailedIndex: -1,
			},
			expected: "the resources such as cpu, memory is not enough in Queue",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processPodGroupPending(tt.pg, tt.indices)
			if cond.Reason != jobEnqueueFailedReason {
				t.Errorf("expected reason %s, got %s", jobEnqueueFailedReason, cond.Reason)
			}
		})
	}
}

type processPodGroupRunningTestCase struct {
	name     string
	pods     map[string]corev1.Pod
	expected jobStatus
}

func buildProcessPodGroupRunningTestCases() []processPodGroupRunningTestCase {
	return []processPodGroupRunningTestCase{
		{
			name: "pending pod",
			pods: map[string]corev1.Pod{
				"pod1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			expected: podGroupRunning,
		},
		{
			name: "failed pod",
			pods: map[string]corev1.Pod{
				"pod1": {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
					Status: corev1.PodStatus{
						Phase:   corev1.PodFailed,
						Reason:  "UnexpectedAdmissionError",
						Message: "not get valid pod",
					},
				},
			},
			expected: podGroupRunning,
		},
		{
			name:     "no pods",
			pods:     map[string]corev1.Pod{},
			expected: "",
		},
	}
}

func TestCollector_ProcessPodGroupRunning(t *testing.T) {
	c := &Collector{}
	for _, tt := range buildProcessPodGroupRunningTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processPodGroupRunning(tt.pods)
			if cond == nil && tt.expected != "" {
				t.Errorf("expected condition, got nil")
			}
			if cond != nil && cond.Status != tt.expected {
				t.Errorf("expected status %s, got %s", tt.expected, cond.Status)
			}
		})
	}
}

func TestCollector_ProcessPodFailed(t *testing.T) {
	c := &Collector{}
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "unexpected admission error",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase:   corev1.PodFailed,
					Reason:  "UnexpectedAdmissionError",
					Message: "not get valid pod",
				},
			},
			expected: "PodFailed",
		},
		{
			name: "generic failed",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase:   corev1.PodFailed,
					Reason:  "GenericError",
					Message: "some error",
				},
			},
			expected: "PodFailed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processPodFailed(tt.pod)
			if cond.Reason != tt.expected {
				t.Errorf("expected reason %s, got %s", tt.expected, cond.Reason)
			}
		})
	}
}

type analyzePodGroupConditionsTestCase struct {
	name     string
	pg       *v1beta1.PodGroup
	expected conditionIndices
}

func buildAnalyzePodGroupConditionsTestCases() []analyzePodGroupConditionsTestCase {
	return []analyzePodGroupConditionsTestCase{
		{
			name: "multiple conditions",
			pg: &v1beta1.PodGroup{
				Status: v1beta1.PodGroupStatus{
					Conditions: []v1beta1.PodGroupCondition{
						{
							Type:    v1beta1.PodGroupUnschedulableType,
							Status:  corev1.ConditionTrue,
							Reason:  jobEnqueueFailedReason,
							Message: "enqueue failed",
						},
						{
							Type:    v1beta1.PodGroupUnschedulableType,
							Status:  corev1.ConditionTrue,
							Reason:  nodePredicateFailedReason,
							Message: "predicate failed",
						},
					},
				},
			},
			expected: conditionIndices{
				jobEnqueueFailedIndex:     0,
				jobValidFailedIndex:       invalidIndex,
				predicatedNodesErrorIndex: 1,
				batchOrderFailedIndex:     invalidIndex,
				notEnoughResourcesIndex:   invalidIndex,
			},
		},
		{
			name: "no conditions",
			pg: &v1beta1.PodGroup{
				Status: v1beta1.PodGroupStatus{
					Conditions: []v1beta1.PodGroupCondition{},
				},
			},
			expected: conditionIndices{
				jobEnqueueFailedIndex:     invalidIndex,
				jobValidFailedIndex:       invalidIndex,
				predicatedNodesErrorIndex: invalidIndex,
				batchOrderFailedIndex:     invalidIndex,
				notEnoughResourcesIndex:   invalidIndex,
			},
		},
	}
}

func TestCollector_AnalyzePodGroupConditions(t *testing.T) {
	c := &Collector{}
	for _, tt := range buildAnalyzePodGroupConditionsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			indices := c.analyzePodGroupConditions(tt.pg)
			if indices.jobEnqueueFailedIndex != tt.expected.jobEnqueueFailedIndex {
				t.Errorf("expected jobEnqueueFailedIndex %d, got %d", tt.expected.jobEnqueueFailedIndex, indices.jobEnqueueFailedIndex)
			}
			if indices.predicatedNodesErrorIndex != tt.expected.predicatedNodesErrorIndex {
				t.Errorf("expected predicatedNodesErrorIndex %d, got %d", tt.expected.predicatedNodesErrorIndex, indices.predicatedNodesErrorIndex)
			}
		})
	}
}

func TestCollector_CleanupJobs(t *testing.T) {
	c := &Collector{
		jobExceptions: map[string]*jobExceptionInfo{
			"job1": {
				JobName: "job1",
			},
			"job2": {
				JobName: "job2",
			},
		},
	}

	allJobs := map[string]constant.JobInfo{
		"job1": {
			Name: "job1",
		},
	}

	allMetaObjs := map[string]metav1.Object{}

	c.cleanupJobs(allJobs, allMetaObjs)

	if _, exists := c.jobExceptions["job2"]; exists {
		t.Error("expected job2 to be deleted")
	}
	if _, exists := c.jobExceptions["job1"]; !exists {
		t.Error("expected job1 to exist")
	}
}

type processPodGroupInqueueTestCase struct {
	name     string
	pg       *v1beta1.PodGroup
	pods     map[string]corev1.Pod
	indices  conditionIndices
	jobInfo  constant.JobInfo
	expected *conditionDetail
}

func buildProcessPodGroupInqueueTestCases() []processPodGroupInqueueTestCase {
	return []processPodGroupInqueueTestCase{
		{
			name: "batch order failed",
			pg: &v1beta1.PodGroup{
				Status: v1beta1.PodGroupStatus{
					Conditions: []v1beta1.PodGroupCondition{
						{Reason: batchOrderFailedReason, Message: "batch order failed"},
						{Reason: nodePredicateFailedReason, Message: "predicate error"},
					},
				},
			},
			indices: conditionIndices{batchOrderFailedIndex: 0, predicatedNodesErrorIndex: 1},
			expected: &conditionDetail{
				Status:  podGroupInqueue,
				Reason:  batchOrderFailedReason,
				Message: "batch order failed; predicate error",
			},
		},
		{
			name: "not enough resources",
			pg: &v1beta1.PodGroup{
				Spec: v1beta1.PodGroupSpec{MinMember: 2},
				Status: v1beta1.PodGroupStatus{
					Conditions: []v1beta1.PodGroupCondition{
						{Reason: notEnoughResourcesReason, Message: "not enough resources"},
					},
				},
			},
			pods:    map[string]corev1.Pod{"pod1": {}},
			indices: conditionIndices{notEnoughResourcesIndex: 0},
			jobInfo: constant.JobInfo{NameSpace: "ns", Name: "job"},
			expected: &conditionDetail{
				Status:  podGroupInqueue,
				Reason:  notEnoughResourcesReason,
				Message: "not enough resources the number of pods is less than minMember",
			},
		},
	}
}

func TestCollector_ProcessPodGroupInqueue(t *testing.T) {
	c := &Collector{}
	for _, tt := range buildProcessPodGroupInqueueTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processPodGroupInqueue(tt.pg, tt.pods, tt.indices, tt.jobInfo)
			if tt.expected == nil {
				if cond != nil {
					t.Errorf("expected nil, got %+v", cond)
				}
				return
			}
			if cond == nil {
				t.Errorf("expected condition, got nil")
				return
			}
			if cond.Status != tt.expected.Status {
				t.Errorf("expected status %s, got %s", tt.expected.Status, cond.Status)
			}
			if cond.Reason != tt.expected.Reason {
				t.Errorf("expected reason %s, got %s", tt.expected.Reason, cond.Reason)
			}
		})
	}
}

func TestCollector_AnalyzePodGroupPhase(t *testing.T) {
	c := &Collector{}
	tests := []struct {
		name     string
		pg       *v1beta1.PodGroup
		pods     map[string]corev1.Pod
		indices  conditionIndices
		jobInfo  constant.JobInfo
		expected jobStatus
	}{
		{
			name:     "empty phase",
			pg:       &v1beta1.PodGroup{Status: v1beta1.PodGroupStatus{Phase: ""}},
			expected: podGroupCreated,
		},
		{
			name:     "unknown phase with pending pod",
			pg:       &v1beta1.PodGroup{Status: v1beta1.PodGroupStatus{Phase: v1beta1.PodGroupUnknown}},
			pods:     map[string]corev1.Pod{"pod1": {Status: corev1.PodStatus{Phase: corev1.PodPending}}},
			expected: podGroupUnknown,
		},
		{
			name:     "running phase with no pods",
			pg:       &v1beta1.PodGroup{Status: v1beta1.PodGroupStatus{Phase: v1beta1.PodGroupRunning}},
			pods:     map[string]corev1.Pod{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.analyzePodGroupPhase(tt.pg, tt.pods, tt.indices, tt.jobInfo)
			if tt.expected == "" {
				if cond != nil {
					t.Errorf("expected nil, got %+v", cond)
				}
				return
			}
			if cond == nil {
				t.Errorf("expected condition, got nil")
				return
			}
			if cond.Status != tt.expected {
				t.Errorf("expected status %s, got %s", tt.expected, cond.Status)
			}
		})
	}
}

func TestCollector_ProcessVcJob(t *testing.T) {
	c := &Collector{}
	tests := []struct {
		name     string
		vcJob    *v1alpha1.Job
		expected *conditionDetail
	}{
		{
			name: "empty phase",
			vcJob: &v1alpha1.Job{
				Status: v1alpha1.JobStatus{
					State: v1alpha1.JobState{Phase: ""},
				},
			},
			expected: &conditionDetail{
				Status: jobStatusEmpty,
				Reason: "JobNoInitialized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := c.processVcJob(tt.vcJob)
			if tt.expected == nil {
				if cond != nil {
					t.Errorf("expected nil, got %+v", cond)
				}
				return
			}
			if cond == nil {
				t.Errorf("expected condition, got nil")
				return
			}
			if cond.Status != tt.expected.Status {
				t.Errorf("expected status %s, got %s", tt.expected.Status, cond.Status)
			}
			if cond.Reason != tt.expected.Reason {
				t.Errorf("expected reason %s, got %s", tt.expected.Reason, cond.Reason)
			}
		})
	}
}

func TestCollector_ProcessPodGroupInqueue_PredicateNodes(t *testing.T) {
	c := &Collector{}
	pg := &v1beta1.PodGroup{
		Status: v1beta1.PodGroupStatus{
			Conditions: []v1beta1.PodGroupCondition{
				{
					Reason:  nodePredicateFailedReason,
					Message: "predicate nodes failed",
				},
			},
		},
	}
	indices := conditionIndices{predicatedNodesErrorIndex: 0}

	cond := c.processPodGroupInqueue(pg, nil, indices, constant.JobInfo{})

	if cond == nil {
		t.Errorf("expected condition, got nil")
		return
	}
	if cond.Reason != nodePredicateFailedReason {
		t.Errorf("expected reason %s, got %s", nodePredicateFailedReason, cond.Reason)
	}
}

func TestCollector_ProcessPodGroupInqueue_JobValidFailed(t *testing.T) {
	c := &Collector{}
	pg := &v1beta1.PodGroup{
		Status: v1beta1.PodGroupStatus{
			Conditions: []v1beta1.PodGroupCondition{
				{
					Reason:  jobValidateFailedReason,
					Message: "job validation failed",
				},
			},
		},
	}
	indices := conditionIndices{jobValidFailedIndex: 0}

	cond := c.processPodGroupInqueue(pg, nil, indices, constant.JobInfo{})

	if cond == nil {
		t.Errorf("expected condition, got nil")
		return
	}
	if cond.Reason != jobValidateFailedReason {
		t.Errorf("expected reason %s, got %s", jobValidateFailedReason, cond.Reason)
	}
}
