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

// Package label defines NPU node labeler interfaces.
package label

import (
	"strings"

	"k8s.io/api/core/v1"
)

// NodeContext provides shared context for labelers and annotators during node marking.
type NodeContext struct {
	Node            *v1.Node
	IsHeterogeneous bool
}

// NodeLabeler is the interface for writing node labels.
// Implementations encapsulate the logic to fetch values from device managers.
type NodeLabeler interface {
	Write(labels map[string]string, ctx *NodeContext) error
}

// Group orchestrates multiple NodeLabelers in order.
type Group struct {
	labelers []NodeLabeler
}

// NewLabelGroup creates a new LabelGroup with the given labelers.
func NewLabelGroup(labelers ...NodeLabeler) *Group {
	return &Group{labelers: labelers}
}

// WriteAll executes all labelers in order and returns the collected labels.
func (g *Group) WriteAll(ctx *NodeContext) (map[string]string, error) {
	labels := make(map[string]string)
	for _, l := range g.labelers {
		if err := l.Write(labels, ctx); err != nil {
			return nil, err
		}
	}
	return labels, nil
}

// GetNodeLabel reads label value from node, trying keys in priority order (front to back).
func GetNodeLabel(node *v1.Node, keys ...string) (string, bool) {
	if node == nil {
		return "", false
	}
	return GetLabelValue(node.Labels, keys...)
}

// GetLabelValue reads label value from map, trying keys in priority order (front to back).
func GetLabelValue(labels map[string]string, keys ...string) (string, bool) {
	if labels == nil {
		return "", false
	}
	for _, key := range keys {
		if val, ok := labels[key]; ok && val != "" {
			return strings.Trim(val, " "), true
		}
	}
	return "", false
}
