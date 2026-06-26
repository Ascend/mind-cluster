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

package utils

import (
	"k8s.io/api/core/v1"
)

// CalcMinResources calculates the minimal resources required by a pod group
// that has `replicas` identical pods described by `podSpec`.
//
// For each container in the pod template, its resource requests are summed up;
// if Requests is omitted for a container, it defaults to Limits (following the
// Kubernetes convention). The per-pod result is then multiplied by `replicas`.
//
// Returns nil if replicas <= 0 or no resource requests are declared, so the
// caller can leave PodGroup.Spec.MinResources unset.
func CalcMinResources(replicas int32, podSpec v1.PodSpec) *v1.ResourceList {
	if replicas <= 0 {
		return nil
	}
	// calculate resources for a single pod
	singlePodRes := v1.ResourceList{}
	for _, container := range podSpec.Containers {
		AddResourceList(singlePodRes, container.Resources.Requests, container.Resources.Limits)
	}
	if len(singlePodRes) == 0 {
		return nil
	}
	// multiply by replicas via repeated addition
	minResources := v1.ResourceList{}
	for name, quantity := range singlePodRes {
		total := quantity
		for i := int32(1); i < replicas; i++ {
			total.Add(quantity)
		}
		minResources[name] = total
	}
	return &minResources
}

// AddResourceList adds resources into list with per-resource fallback.
// For each resource present in req, the req value is used. For resources
// only in limit (absent from req), the limit value is used as fallback,
// following the Kubernetes convention that if Requests is omitted for a
// resource, it defaults to Limits if explicitly specified.
func AddResourceList(list, req, limit v1.ResourceList) {
	for name, quantity := range req {
		if value, ok := list[name]; ok {
			value.Add(quantity)
			list[name] = value
		} else {
			list[name] = quantity
		}
	}
	for name, quantity := range limit {
		if _, ok := req[name]; ok {
			continue
		}
		if value, ok := list[name]; ok {
			value.Add(quantity)
			list[name] = value
		} else {
			list[name] = quantity
		}
	}
}
