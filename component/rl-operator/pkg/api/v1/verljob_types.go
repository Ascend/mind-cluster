/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package v1

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VerlJobSpec defines the desired state of VerlJob
type VerlJobSpec struct {
	// RunPolicy encapsulates various runtime policies of the distributed training
	// job, for example how to clean up resources and how long the job can stay
	// active.
	// +kubebuilder:validation:Optional
	RunPolicy commonv1.RunPolicy `json:"runPolicy"`

	// SchedulerName defines the job scheduler with gang-scheduling enabled
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	ExtraConfig map[string]string `json:"extraConfig,omitempty"`

	// +required
	Template v1.PodTemplateSpec `json:"template,omitempty"`

	/*	 A map of ReplicaType (type) to ReplicaSpec (value). Specifies the ML cluster configuration.
		 For example,
		   {
		     "Scheduler": ReplacaSpec,
		     "Worker": ReplicaSpec,
		   }
	*/
	ReplicaSpecs map[commonv1.ReplicaType]*commonv1.ReplicaSpec `json:"replicaSpecs"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VerlJob is the Schema for the verljobs API
type VerlJob struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of VerlJob
	// +required
	Spec VerlJobSpec `json:"spec"`

	// status defines the observed state of VerlJob
	// +optional
	Status commonv1.JobStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// VerlJobList contains a list of VerlJob
type VerlJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []VerlJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VerlJob{}, &VerlJobList{})
	SchemeBuilder.SchemeBuilder.Register(RegisterDefaults)
}
