/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package annotation defines NPU node annotator methods
package annotation

import (
	"strings"

	"k8s.io/api/core/v1"
)

// GetNodeAnnotation reads annotation value from node, trying keys in priority order (front to back).
// Returns the annotation value and a bool indicating whether the value was found.
func GetNodeAnnotation(node *v1.Node, keys ...string) (string, bool) {
	if node == nil {
		return "", false
	}
	return GetAnnotationValue(node.Annotations, keys...)
}

// GetAnnotationValue reads annotation value from map, trying keys in priority order (front to back).
// Returns the annotation value and a bool indicating whether the value was found.
func GetAnnotationValue(annotations map[string]string, keys ...string) (string, bool) {
	if annotations == nil {
		return "", false
	}
	for _, key := range keys {
		if val, ok := annotations[key]; ok && val != "" {
			return strings.Trim(val, " "), true
		}
	}
	return "", false
}
