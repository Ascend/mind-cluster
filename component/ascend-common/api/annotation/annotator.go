/*Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package annotation defines NPU node annotators.
package annotation

import "ascend-common/api/label"

// NodeAnnotator is the interface for writing node annotations.
// Implementations encapsulate the logic to fetch values from device managers.
type NodeAnnotator interface {
	Write(annotations map[string]string, ctx *label.NodeContext) error
}

// Group orchestrates multiple NodeAnnotators in order.
type Group struct {
	annotators []NodeAnnotator
}

// NewAnnotationGroup creates a new AnnotationGroup with the given annotators.
func NewAnnotationGroup(annotators ...NodeAnnotator) *Group {
	return &Group{annotators: annotators}
}

// WriteAll executes all annotators in order and returns the collected annotations.
func (g *Group) WriteAll(ctx *label.NodeContext) (map[string]string, error) {
	annotations := make(map[string]string)
	for _, a := range g.annotators {
		if err := a.Write(annotations, ctx); err != nil {
			return nil, err
		}
	}
	return annotations, nil
}
