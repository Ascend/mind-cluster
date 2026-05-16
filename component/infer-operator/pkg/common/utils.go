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

package common

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
)

// DeepCopyLabelsMap deep copies the labels map.
func DeepCopyLabelsMap(m map[string]string) map[string]string {
	newMap := make(map[string]string, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

// WorkLoadTypeToGVK converts workload type to GVK.
func WorkLoadTypeToGVK(tm v1.WorkloadType) (schema.GroupVersionKind, error) {
	gv, err := schema.ParseGroupVersion(tm.APIVersion)
	if err != nil {
		hwlog.RunLog.Warnf(`Could not parse GroupVersion "%s", error: %v`, tm.APIVersion, err)
		return schema.GroupVersionKind{}, err
	}
	return gv.WithKind(tm.Kind), nil
}

// AddLabelsFromIndexer adds labels from instance indexer.
func AddLabelsFromIndexer(labels map[string]string, indexer InstanceIndexer) map[string]string {
	newLabels := DeepCopyLabelsMap(labels)
	newLabels[InferServiceNameLabelKey] = indexer.ServiceName
	newLabels[InstanceSetNameLabelKey] = indexer.InstanceSetKey
	newLabels[InstanceIndexLabelKey] = indexer.InstanceIndex
	newLabels[OperatorNameKey] = TrueBool
	return newLabels
}

// GetInstanceSetNameFromLabels gets instance set name from labels.
func GetInstanceSetNameFromLabels(labels map[string]string) string {
	serviceName := labels[InferServiceNameLabelKey]
	instanceSetKey := labels[InstanceSetNameLabelKey]
	return fmt.Sprintf("%s-%s", serviceName, instanceSetKey)
}

// GetWorkLoadNameFromIndexer gets workload name from instance indexer.
func GetWorkLoadNameFromIndexer(indexer InstanceIndexer) string {
	return fmt.Sprintf("%s-%s-%s", indexer.ServiceName, indexer.InstanceSetKey, indexer.InstanceIndex)
}

// GetServiceNameFromIndexer gets service name from instance indexer.
func GetServiceNameFromIndexer(indexer InstanceIndexer) string {
	return fmt.Sprintf("service-%s-%s-%s", indexer.ServiceName, indexer.InstanceSetKey, indexer.InstanceIndex)
}

// GetPGNameFromIndexer gets pg name from instance indexer.
func GetPGNameFromIndexer(indexer InstanceIndexer) string {
	return fmt.Sprintf("pg-%s-%s-%s", indexer.ServiceName, indexer.InstanceSetKey, indexer.InstanceIndex)
}

// AddEnvToPodTemplate adds environment variables to pod template.
func AddEnvToPodTemplate(pod *corev1.PodTemplateSpec, indexer InstanceIndexer) {
	for index := range pod.Spec.Containers {
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
			Name:  InstanceIndexEnvKey,
			Value: indexer.InstanceIndex,
		})
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
			Name:  InstanceRoleEnvKey,
			Value: indexer.InstanceSetKey,
		})
		splitResult := strings.Split(indexer.ServiceName, "-")
		if len(splitResult) < InferServiceNameSplitNum {
			hwlog.RunLog.Warnf("Service name '%s' has insufficient segments after splitting, "+
				"cannot extract infer service index", indexer.ServiceName)
			return
		}
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
			Name:  InferServiceIndexEnvKey,
			Value: splitResult[len(splitResult)-1],
		})
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
			Name:  InferServiceNameEnvKey,
			Value: strings.Join(splitResult[:len(splitResult)-1], "-"),
		})
	}
}

// IsRequeueError checks if the error is a requeue error.
func IsRequeueError(err error) bool {
	var reQueueError *RequeueError
	ok := errors.As(err, &reQueueError)
	return ok
}

// IsInstanceSetReady check if instanceset is ready
func IsInstanceSetReady(instanceSet *v1.InstanceSet) bool {
	condition := meta.FindStatusCondition(instanceSet.Status.Conditions, string(InstanceSetReady))
	if condition == nil {
		return false
	}
	return condition.Status == metav1.ConditionTrue
}
