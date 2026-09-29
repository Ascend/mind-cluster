/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// ScalingStrategyKPA identifies the Knative-style pod autoscaling algorithm.
	ScalingStrategyKPA = "KPA"
)

// ScaleTargetRef references the InstanceSet controlled by a PodAutoscaler.
type ScaleTargetRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// MetricSource describes a Prometheus metric exposed by every target Pod.
type MetricSource struct {
	TargetName  string `json:"name"`
	TargetValue string `json:"value"`
	Port        string `json:"port"`
	Path        string `json:"path,omitempty"`
}

// PodAutoscalerSpec defines the desired state of PodAutoscaler.
type PodAutoscalerSpec struct {
	ScaleTargetRef ScaleTargetRef `json:"scaleTargetRef"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas"`

	// ScalingStrategy is retained as an algorithm extension point. Only KPA is supported now.
	// +kubebuilder:validation:Enum=KPA
	// +kubebuilder:default=KPA
	ScalingStrategy string `json:"scalingStrategy,omitempty"`

	// +kubebuilder:validation:MinItems=1
	MetricsSources []MetricSource `json:"metricsSources"`

	// +kubebuilder:validation:Minimum=1
	StableWindowSeconds *int32 `json:"stableWindowSeconds,omitempty"`

	// +kubebuilder:validation:Minimum=1
	PanicWindowSeconds *int32 `json:"panicWindowSeconds,omitempty"`
}

// PodAutoscalerReplicaStatus summarizes observed and recommended capacity.
type PodAutoscalerReplicaStatus struct {
	Current int32 `json:"current,omitempty"`
	Ready   int32 `json:"ready,omitempty"`
	Desired int32 `json:"desired,omitempty"`
}

// PodAutoscalerScaleStatus describes the most recent applied replica change.
type PodAutoscalerScaleStatus struct {
	Time             metav1.Time `json:"time"`
	PreviousReplicas int32       `json:"previousReplicas"`
	NewReplicas      int32       `json:"newReplicas"`
	Reason           string      `json:"reason,omitempty"`
}

// PodAutoscalerStatus defines the observed state of PodAutoscaler.
type PodAutoscalerStatus struct {
	ObservedGeneration int64                      `json:"observedGeneration,omitempty"`
	Replicas           PodAutoscalerReplicaStatus `json:"replicas,omitempty"`
	MetricValues       map[string]float64         `json:"metricValues,omitempty"`
	LastScale          *PodAutoscalerScaleStatus  `json:"lastScale,omitempty"`
	Conditions         []metav1.Condition         `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=pa
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=".spec.scaleTargetRef.name"
// +kubebuilder:printcolumn:name="Strategy",type=string,JSONPath=".spec.scalingStrategy"
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=".status.replicas.current"
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=".status.replicas.desired"

// PodAutoscaler is the Schema for the podautoscalers API.
type PodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodAutoscalerSpec   `json:"spec,omitempty"`
	Status PodAutoscalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodAutoscalerList contains a list of PodAutoscaler.
type PodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodAutoscaler{}, &PodAutoscalerList{})
}
