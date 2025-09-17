/*
Copyright 2023 Huawei Technologies Co., Ltd.

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

// Package v1 is used to define Job object and its initialization.
package v1

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job is the Schema for the Job API
type Job struct {
	// Standard Kubernetes type metadata.
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired state of the Job.
	// +optional
	Spec JobSpec `json:"spec,omitempty"`

	// Most recently observed status of the Job.
	// Populated by the system.
	// Read-only.
	// +optional
	Status commonv1.JobStatus `json:"status,omitempty"`
}

// JobSpec defines the desired state of Job
type JobSpec struct {
	// RunPolicy encapsulates various runtime policies of the distributed training
	// job, for example how to clean up resources and how long the job can stay
	// active.
	// +kubebuilder:validation:Optional
	RunPolicy commonv1.RunPolicy `json:"runPolicy"`

	// SuccessPolicy defines the policy to mark the Job as succeeded.
	// Default to "", using the default rules.
	// +optional
	SuccessPolicy *SuccessPolicy `json:"successPolicy,omitempty"`

	// SchedulerName defines the job scheduler with gang-scheduling enabled
	SchedulerName string `json:"schedulerName,omitempty"`

	/*	 A map of ReplicaType (type) to ReplicaSpec (value). Specifies the ML cluster configuration.
		 For example,
		   {
		     "Scheduler": ReplacaSpec,
		     "Worker": ReplicaSpec,
		   }
	*/
	ReplicaSpecs map[commonv1.ReplicaType]*commonv1.ReplicaSpec `json:"replicaSpecs"`
}

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}

// SuccessPolicy is the success policy.
type SuccessPolicy string

const (
	// SuccessPolicyDefault is the default policy of success
	SuccessPolicyDefault SuccessPolicy = ""
	// SuccessPolicyAllWorkers is the 'ALLWorkers' policy of success
	SuccessPolicyAllWorkers SuccessPolicy = "AllWorkers"
)
