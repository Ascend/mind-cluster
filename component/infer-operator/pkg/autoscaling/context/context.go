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

package context

import (
	"fmt"
	"strconv"
	"time"

	apiv1 "infer-operator/pkg/api/v1"
	autoscalingtypes "infer-operator/pkg/autoscaling/types"
)

const (
	defaultScaleRate       = 2
	defaultTolerance       = 0.1
	defaultPanicThreshold  = 2
	defaultScaleDownWindow = 300 * time.Second
)

// ScalingPolicy is the validated configuration used for one evaluation cycle.
// It is rebuilt on every reconciliation so removed annotations use defaults.
type ScalingPolicy struct {
	MinReplicas             int32
	MaxReplicas             int32
	MaxScaleUpRate          float64
	MaxScaleDownRate        float64
	ScaleUpTolerance        float64
	ScaleDownTolerance      float64
	ScaleUpCooldownWindow   time.Duration
	ScaleDownCooldownWindow time.Duration
	PanicThreshold          float64
	MetricTargets           map[string]float64
}

// RuntimeState contains observations that survive between evaluations.
type RuntimeState struct {
	StableValue      float64
	PanicValue       float64
	PanicMode        bool
	MaxPanicReplicas int32
}

// EvaluationContext keeps validated configuration separate from mutable state.
type EvaluationContext struct {
	Policy       ScalingPolicy
	MetricStates map[string]*RuntimeState
}

// NewEvaluationContext returns a context initialized with controller defaults.
func NewEvaluationContext() *EvaluationContext {
	return &EvaluationContext{Policy: defaultPolicy(), MetricStates: make(map[string]*RuntimeState)}
}

// Refresh validates the complete policy before replacing the active one.
func (evaluation *EvaluationContext) Refresh(pa *apiv1.PodAutoscaler) error {
	if pa == nil {
		return fmt.Errorf("PodAutoscaler is required")
	}

	policy := defaultPolicy()
	policy.MinReplicas = pa.Spec.MinReplicas
	policy.MaxReplicas = pa.Spec.MaxReplicas
	policy.MetricTargets = make(map[string]float64, len(pa.Spec.MetricsSources))
	for _, source := range pa.Spec.MetricsSources {
		value, err := strconv.ParseFloat(source.TargetValue, 64)
		if err != nil || value <= 0 {
			return fmt.Errorf("metric %q target value must be positive", source.TargetName)
		}
		policy.MetricTargets[source.TargetName] = value
	}

	for key, value := range pa.Annotations {
		if err := applyAnnotation(&policy, key, value); err != nil {
			return fmt.Errorf("invalid annotation %q value %q: %w", key, value, err)
		}
	}
	if err := validatePolicy(policy); err != nil {
		return err
	}

	retainedStates := make(map[string]*RuntimeState, len(policy.MetricTargets))
	for metricName := range policy.MetricTargets {
		if previous, found := evaluation.MetricStates[metricName]; found {
			retainedStates[metricName] = previous
			continue
		}
		retainedStates[metricName] = &RuntimeState{}
	}
	evaluation.Policy = policy
	evaluation.MetricStates = retainedStates
	return nil
}

// StateFor returns the runtime observations owned by one metric source.
func (evaluation *EvaluationContext) StateFor(metricName string) *RuntimeState {
	if state, found := evaluation.MetricStates[metricName]; found {
		return state
	}
	state := &RuntimeState{}
	evaluation.MetricStates[metricName] = state
	return state
}

// TargetValue returns the configured target for one metric.
func (policy ScalingPolicy) TargetValue(metricName string) (float64, bool) {
	value, found := policy.MetricTargets[metricName]
	return value, found
}

func defaultPolicy() ScalingPolicy {
	return ScalingPolicy{
		MaxScaleUpRate: defaultScaleRate, MaxScaleDownRate: defaultScaleRate,
		ScaleUpTolerance: defaultTolerance, ScaleDownTolerance: defaultTolerance,
		ScaleDownCooldownWindow: defaultScaleDownWindow, PanicThreshold: defaultPanicThreshold,
		MetricTargets: make(map[string]float64),
	}
}

func applyAnnotation(policy *ScalingPolicy, key, value string) error {
	var err error
	switch key {
	case autoscalingtypes.MaxScaleUpRateLabel:
		policy.MaxScaleUpRate, err = strconv.ParseFloat(value, 64)
	case autoscalingtypes.MaxScaleDownRateLabel:
		policy.MaxScaleDownRate, err = strconv.ParseFloat(value, 64)
	case autoscalingtypes.ScaleUpToleranceLabel:
		policy.ScaleUpTolerance, err = strconv.ParseFloat(value, 64)
	case autoscalingtypes.ScaleDownToleranceLabel:
		policy.ScaleDownTolerance, err = strconv.ParseFloat(value, 64)
	case autoscalingtypes.PanicThresholdLabel:
		policy.PanicThreshold, err = strconv.ParseFloat(value, 64)
	case autoscalingtypes.ScaleUpCooldownWindowLabel:
		policy.ScaleUpCooldownWindow, err = time.ParseDuration(value)
	case autoscalingtypes.ScaleDownCooldownWindowLabel:
		policy.ScaleDownCooldownWindow, err = time.ParseDuration(value)
	}
	return err
}

func validatePolicy(policy ScalingPolicy) error {
	if policy.MinReplicas < 1 || policy.MaxReplicas < policy.MinReplicas {
		return fmt.Errorf("replica bounds are invalid: minReplicas=%d maxReplicas=%d",
			policy.MinReplicas, policy.MaxReplicas)
	}
	if policy.MaxScaleUpRate <= 0 {
		return fmt.Errorf("max scale-up rate must be positive: maxScaleUpRate=%g", policy.MaxScaleUpRate)
	}
	if policy.MaxScaleDownRate <= 0 {
		return fmt.Errorf("max scale-down rate must be positive: maxScaleDownRate=%g", policy.MaxScaleDownRate)
	}
	if policy.PanicThreshold <= 0 {
		return fmt.Errorf("panic threshold must be positive: panicThreshold=%g", policy.PanicThreshold)
	}
	if policy.ScaleUpTolerance < 0 {
		return fmt.Errorf("scale-up tolerance must not be negative: scaleUpTolerance=%g", policy.ScaleUpTolerance)
	}
	if policy.ScaleDownTolerance < 0 {
		return fmt.Errorf("scale-down tolerance must not be negative: scaleDownTolerance=%g", policy.ScaleDownTolerance)
	}
	if policy.ScaleUpCooldownWindow < 0 {
		return fmt.Errorf("scale-up cooldown must not be negative: scaleUpCooldown=%s", policy.ScaleUpCooldownWindow)
	}
	if policy.ScaleDownCooldownWindow < 0 {
		return fmt.Errorf("scale-down cooldown must not be negative: scaleDownCooldown=%s", policy.ScaleDownCooldownWindow)
	}
	return nil
}
