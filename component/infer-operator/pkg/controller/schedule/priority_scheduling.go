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

// Package schedule contains the scheduling logic
package schedule

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

const (
	AllReadyPriority = int32(common.MaxRoleTypeCount + 1)
)

// ShouldSchedule check if instanceset should be scheduled
func ShouldSchedule(ctx context.Context, k8sClient client.Client, instanceSet *v1.InstanceSet) (bool, error) {
	schedulingStrategy := instanceSet.Labels[common.PrioritySchedulingStrategyLabelKey]
	switch schedulingStrategy {
	case common.SchedulingStrategyPriority:
		return shouldScheduleWithPriority(ctx, k8sClient, instanceSet)
	case common.SchedulingStrategySequential:
		return true, nil
	default:
		return false, fmt.Errorf("unknown scheduling strategy %s", schedulingStrategy)
	}
}

func shouldScheduleWithPriority(ctx context.Context, k8sClient client.Client, instanceSet *v1.InstanceSet) (bool, error) {
	inferService, err := getInferService(ctx, k8sClient, instanceSet)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get InferService for InstanceSet %s/%s: %v",
			instanceSet.Namespace, instanceSet.Name, err)
		return false, err
	}

	currentRoleName := instanceSet.Labels[common.InstanceSetNameLabelKey]
	currentRoleSpec, currentRoleFound := findRoleSpec(inferService, currentRoleName)
	if !currentRoleFound {
		hwlog.RunLog.Errorf("InstanceSet %s/%s role %s not found in InferService roles, maybe InferService is modified",
			instanceSet.Namespace, instanceSet.Name, currentRoleName)
		return false, fmt.Errorf("current role %s not found in InferService roles", currentRoleName)
	}

	if common.IsInstanceSetReady(instanceSet) {
		// current role is ready, reject schedule
		return false, nil
	}

	siblingInstanceSets, err := listSiblingInstanceSets(ctx, k8sClient, instanceSet)
	if err != nil {
		return false, err
	}
	instanceSetMap := make(map[string]*v1.InstanceSet)
	instanceSetMap[currentRoleName] = instanceSet
	for _, instanceSet := range siblingInstanceSets {
		roleName := instanceSet.Labels[common.InstanceSetNameLabelKey]
		instanceSetMap[roleName] = instanceSet
	}

	schedulingPriority := findHighestPriorityOfNotReadyRole(inferService, instanceSetMap)
	if schedulingPriority == AllReadyPriority {
		// all roles include current role are ready, reject schedule, normally, this case should not happen
		return false, nil
	}

	currentPriority := getPriority(currentRoleSpec)
	if currentPriority == schedulingPriority {
		// current role is the highest priority not ready role, accept schedule
		hwlog.RunLog.Infof("InstanceSet %s/%s is scheduled with priority scheduling, ",
			instanceSet.Namespace, instanceSet.Name)
		return true, nil
	}

	// there is a higher priority role not ready, requeue to wait for ready
	hwlog.RunLog.Infof("InstanceSet %s/%s is not scheduled due to priority scheduling policy, "+
		"will requeue this request", instanceSet.Namespace, instanceSet.Name)
	return false, common.NewRequeueError("wait for higher priority instanceSet to be ready")
}

func findRoleSpec(inferService *v1.InferService, roleName string) (*v1.InstanceSetSpec, bool) {
	for i := range inferService.Spec.Roles {
		if inferService.Spec.Roles[i].Name == roleName {
			return &inferService.Spec.Roles[i], true
		}
	}
	return nil, false
}

func findHighestPriorityOfNotReadyRole(inferService *v1.InferService, instanceSetMap map[string]*v1.InstanceSet) int32 {
	highestPriority := AllReadyPriority

	for i := range inferService.Spec.Roles {
		role := &inferService.Spec.Roles[i]
		instanceSet, exists := instanceSetMap[role.Name]

		if !exists || !common.IsInstanceSetReady(instanceSet) {
			priority := getPriority(role)
			// lower value means higher priority
			if priority < highestPriority {
				highestPriority = priority
			}
		}
	}

	return highestPriority
}

func getPriority(role *v1.InstanceSetSpec) int32 {
	if role.Priority == nil {
		return common.DefaultPriority
	}
	return *role.Priority
}

func getInferService(ctx context.Context, k8sClient client.Client, instanceSet *v1.InstanceSet) (*v1.InferService, error) {
	serviceName, ok := instanceSet.Labels[common.InferServiceNameLabelKey]
	if !ok {
		return nil, fmt.Errorf("InstanceSet %s/%s missing label %s",
			instanceSet.Namespace, instanceSet.Name, common.InferServiceNameLabelKey)
	}

	inferService := &v1.InferService{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: instanceSet.Namespace,
	}, inferService); err != nil {
		return nil, err
	}

	return inferService, nil
}

func listSiblingInstanceSets(ctx context.Context, k8sClient client.Client, instanceSet *v1.InstanceSet) ([]*v1.InstanceSet, error) {
	serviceName, ok := instanceSet.Labels[common.InferServiceNameLabelKey]
	if !ok {
		return nil, fmt.Errorf("InstanceSet %s/%s missing label %s",
			instanceSet.Namespace, instanceSet.Name, common.InferServiceNameLabelKey)
	}

	instanceSetList := &v1.InstanceSetList{}
	selector := labels.SelectorFromSet(labels.Set{
		common.InferServiceNameLabelKey: serviceName,
	})

	if err := k8sClient.List(ctx, instanceSetList,
		client.InNamespace(instanceSet.Namespace),
		client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, err
	}

	var siblingInstanceSets []*v1.InstanceSet
	for i := range instanceSetList.Items {
		siblingInstanceSet := &instanceSetList.Items[i]
		if siblingInstanceSet.UID != instanceSet.UID {
			siblingInstanceSets = append(siblingInstanceSets, siblingInstanceSet)
		}
	}

	return siblingInstanceSets, nil
}
