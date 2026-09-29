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

package types

import (
	"fmt"
	"time"
)

// ConditionType identifies a PodAutoscaler status condition.
type ConditionType string

const (
	// ConditionReady reports whether the PodAutoscaler is evaluating its target.
	ConditionReady ConditionType = "Ready"
	// ConditionValidSpec reports whether the PodAutoscaler configuration is valid.
	ConditionValidSpec ConditionType = "ValidSpec"
	// ConditionConflict reports whether another autoscaler owns the target.
	ConditionConflict ConditionType = "AutoscalingConflict"
	// ConditionMetrics reports whether usable metrics are available.
	ConditionMetrics ConditionType = "MetricsAvailable"
	// ConditionScalingActive reports whether a recommendation was computed.
	ConditionScalingActive ConditionType = "ScalingActive"

	// DefaultReconcileInterval is the periodic evaluation interval for PodAutoscalers.
	DefaultReconcileInterval = 10 * time.Second
)

// MetricKey uniquely identifies one metric window owned by one PodAutoscaler.
type MetricKey struct {
	Namespace   string
	Name        string
	MetricName  string
	PANamespace string
	PAName      string
}

// String returns the stable key used by the in-memory metrics store.
func (key MetricKey) String() string {
	return fmt.Sprintf("%s/%s/%s", key.PANamespace, key.PAName, key.MetricName)
}

// ScaleTarget identifies the workload and metric being evaluated.
type ScaleTarget struct {
	Namespace  string
	Name       string
	Kind       string
	APIVersion string
	MetricKey  MetricKey
}

// MetricObservation is the data boundary between Pod collection and replica computation.
type MetricObservation struct {
	Name      string
	Values    []float64
	Timestamp time.Time
}

// AggregatedMetrics contains the values produced by stable and panic windows.
type AggregatedMetrics struct {
	MetricKey    MetricKey
	CurrentValue float64
	StableValue  float64
	PanicValue   float64
	Confidence   float64
	LastUpdated  time.Time
}

// MetricPoint is a timestamped scalar metric sample.
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}
