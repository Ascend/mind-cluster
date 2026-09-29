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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InferServiceSetSpec defines the desired state of InferServiceSet
type InferServiceSetSpec struct {
	Replicas             *int32           `json:"replicas,omitempty"`
	InferServiceTemplate InferServiceSpec `json:"template,omitempty"`
}

// InferServiceSetStatus defines the observed state of InferServiceSet
type InferServiceSetStatus struct {
	ObservedGeneration int64          `json:"observedGeneration,omitempty"`
	ReadyReplicas      int32          `json:"readyReplicas,omitempty"`
	Replicas           int32          `json:"replicas,omitempty"`
	Conditions         []v1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// InferServiceSet is the Schema for the inferservicesets API
type InferServiceSet struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferServiceSetSpec   `json:"spec,omitempty"`
	Status InferServiceSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InferServiceSetList contains a list of InferServiceSet
type InferServiceSetList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata,omitempty"`
	Items       []InferServiceSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferServiceSet{}, &InferServiceSetList{})
}
