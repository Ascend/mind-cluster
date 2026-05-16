/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// package v1 contains API Schema definitions for the mindcluster v1 API group
package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// WorkloadType defines the type of workload
type WorkloadType struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

// ObjectMeta defines the metadata of the workload
type ObjectMeta struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	Name        string             `json:"name,omitempty"`
	ServiceMeta ObjectMeta         `json:"metadata,omitempty"`
	Spec        corev1.ServiceSpec `json:"spec,omitempty"`
}

// InstanceSetSpec defines the desired state of InstanceSet
type InstanceSetSpec struct {
	Name               string               `json:"name,omitempty"`
	Replicas           *int32               `json:"replicas,omitempty"`
	Services           []ServiceSpec        `json:"services,omitempty"`
	WorkloadTypeMeta   WorkloadType         `json:"workload,omitempty"`
	WorkloadObjectMeta ObjectMeta           `json:"metadata,omitempty"`
	InstanceSpec       runtime.RawExtension `json:"spec,omitempty"`
	Priority           *int32               `json:"priority,omitempty"`
}

// InstanceSetStatus defines the observed state of InstanceSet
type InstanceSetStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	Replicas           int32              `json:"replicas,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// InstanceSet is the Schema for the instancesets API
type InstanceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSetSpec   `json:"spec,omitempty"`
	Status InstanceSetStatus `json:"status,omitempty"`
}

// InstanceSetList contains a list of InstanceSet
type InstanceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstanceSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InstanceSet{}, &InstanceSetList{})
}
