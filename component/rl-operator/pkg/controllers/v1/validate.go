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
	"fmt"
	"strconv"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "rl-operator/pkg/api/v1"
	"rl-operator/pkg/common"
)

func (r *RayClusterReconciler) ValidateJob(resource mindxdlv1.RayCluster) *common.ValidateError {
	if r == nil {
		return &common.ValidateError{
			Reason:  common.ArgumentValidateReason,
			Message: "nil pointer",
		}
	}

	var err *common.ValidateError
	defer func() {
		if err != nil {
			r.Recorder.Event(&resource, corev1.EventTypeWarning, err.Reason, err.Message)
		}
	}()

	if err = validateRayClusterMetaInfo(resource); err != nil {
		return err
	}
	err = r.validateSpec(resource)
	return err
}

func (r *VerlJobReconciler) ValidateJob(resource mindxdlv1.VerlJob) *common.ValidateError {
	if r == nil {
		return &common.ValidateError{
			Reason:  common.ArgumentValidateReason,
			Message: "nil pointer",
		}
	}

	var err *common.ValidateError
	defer func() {
		if err != nil {
			r.Recorder.Event(&resource, corev1.EventTypeWarning, err.Reason, err.Message)
		}
	}()

	if err = validateVerlJobMetaInfo(resource); err != nil {
		return err
	}
	err = validateContainer(common.VerlReplicaType, &resource.Spec.Template)
	return err
}

func validateRayClusterMetaInfo(resource mindxdlv1.RayCluster) *common.ValidateError {
	for _, portKey := range common.ServicePortKeys {
		port, isExists := resource.Labels[portKey]
		if isExists && !isValidPort(port) {
			return &common.ValidateError{
				Reason:  common.ArgumentValidateReason,
				Message: fmt.Sprintf("valid %s argument", portKey),
			}
		}
	}
	return nil
}

func validateVerlJobMetaInfo(resource mindxdlv1.VerlJob) *common.ValidateError {
	annos := resource.GetAnnotations()
	autoSubmit, ok := annos[common.AutoSubmitLabelKey]
	if !ok {
		autoSubmit = "false"
	}
	_, hasConfigPath := annos[common.VerlPathLabelKey]
	if autoSubmit == "true" && !hasConfigPath {
		return &common.ValidateError{
			Reason: common.ArgumentValidateReason,
			Message: fmt.Sprintf("%s is %s, %s must be configed, but not found",
				common.AutoSubmitLabelKey, autoSubmit, common.VerlPathLabelKey),
		}
	}
	return nil
}

func (r *RayClusterReconciler) validateSpec(resource mindxdlv1.RayCluster) *common.ValidateError {
	spec := &resource.Spec

	if r.EnableGangScheduling && spec.RunPolicy.SchedulingPolicy != nil {
		queueName := spec.RunPolicy.SchedulingPolicy.Queue
		if _, err := r.VolcanoControl.GetQueue(queueName); err != nil {
			return &common.ValidateError{
				Reason:  common.ArgumentValidateReason,
				Message: fmt.Sprintf("query queueName %s error: %s", queueName, err.Error()),
			}
		}
	}

	return checkReplicaSpecs(spec.ReplicaSpecs)
}

func checkReplicaSpecs(specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *common.ValidateError {
	hasHeadReplica := false
	for rType, value := range specs {
		if value == nil {
			return &common.ValidateError{
				Reason:  common.ArgumentValidateReason,
				Message: fmt.Sprintf("value of replicaSpec %v is not valid: nil", rType),
			}
		}
		if string(rType) == common.RayHeadType {
			hasHeadReplica = true
		}
		if ve := validateReplicas(rType, value); ve != nil {
			return ve
		}
		if err := validateContainer(rType, &value.Template); err != nil {
			return err
		}
	}

	if hasHeadReplica {
		return nil
	}
	return &common.ValidateError{
		Reason:  common.ArgumentValidateReason,
		Message: fmt.Sprintf("head replica must be included, but not fount"),
	}
}

func validateReplicas(rType commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *common.ValidateError {
	if spec.Replicas == nil {
		return nil
	}
	if rType == common.RayHeadType && *spec.Replicas != 1 {
		return &common.ValidateError{
			Reason: "ReplicaTypeError",
			Message: fmt.Sprintf("replicaSpec %v is not valid: replicas must be 1, but got %d",
				rType, *spec.Replicas),
		}
	}
	if *spec.Replicas < 0 {
		return &common.ValidateError{
			Reason: "ReplicaTypeError",
			Message: fmt.Sprintf("replicaSpec %v is not valid: replicas can not be negative num, but got %d",
				rType, *spec.Replicas),
		}
	}
	if *spec.Replicas > common.MaxReplicas {
		return &common.ValidateError{
			Reason: "ReplicaTypeError",
			Message: fmt.Sprintf("jobSpec is not valid: replicas can not be larger than %d, but got %d",
				common.MaxReplicas, *spec.Replicas),
		}
	}
	return nil
}

func validateContainer(rType commonv1.ReplicaType, template *corev1.PodTemplateSpec) *common.ValidateError {
	if template == nil || len(template.Spec.Containers) == 0 {
		return &common.ValidateError{
			Reason:  common.ArgumentValidateReason,
			Message: fmt.Sprintf("replicaSpecs is not valid: containers definition expected in %v", rType),
		}
	}

	hasDefaultContainer := false
	for _, container := range template.Spec.Containers {
		if container.Image == "" {
			return &common.ValidateError{
				Reason: common.ArgumentValidateReason,
				Message: fmt.Sprintf("replicaType %v is not valid: Image is undefined in the container of %v",
					rType, container.Name),
			}
		}
		if container.Name != common.DefaultContainerName {
			continue
		}
		hasDefaultContainer = true
	}
	if hasDefaultContainer {
		return nil
	}
	return &common.ValidateError{
		Reason: common.ArgumentValidateReason,
		Message: fmt.Sprintf("replicaType %v is not valid: There is no container named %s",
			rType, common.DefaultContainerName),
	}
}

func isValidPort(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	if port <= common.MinPort || port > common.MaxPort {
		return false
	}
	return true
}
