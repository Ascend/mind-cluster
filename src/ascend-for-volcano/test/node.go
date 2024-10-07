/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package test is using for HuaWei Ascend pin scheduling test.
*/
package test

import (
	"strconv"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// BuildNPUNode built NPU node object
func BuildNPUNode(node NPUNode) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        node.Name,
			Labels:      node.Labels,
			Annotations: node.Annotation,
		},
		Status: v1.NodeStatus{
			Capacity:    node.Capacity,
			Allocatable: node.Allocatable,
		},
	}
}

// FakeNormalTestNode fake normal test node.
func FakeNormalTestNode(name string) *api.NodeInfo {
	node := NPUNode{
		Name:        name,
		Capacity:    make(v1.ResourceList, npuIndex3),
		Allocatable: make(v1.ResourceList, npuIndex3),
		Labels:      make(map[string]string, npuIndex3),
		Selector:    make(map[string]string, npuIndex3),
		Annotation:  make(map[string]string, npuIndex3),
		Other:       make(map[string]interface{}, npuIndex3),
	}
	nodeInfo := api.NewNodeInfo(BuildNPUNode(node))
	return nodeInfo
}

// FakeNormalTestNodes fake normal test nodes.
func FakeNormalTestNodes(num int) []*api.NodeInfo {
	var nodes []*api.NodeInfo

	for i := 0; i < num; i++ {
		strNum := strconv.Itoa(i)
		nodeInfo := FakeNormalTestNode("node" + strNum)
		nodes = append(nodes, nodeInfo)
	}

	return nodes
}

// SetFakeNodeSource Set fake node the idle, Capability, Allocatable source.
func SetFakeNodeSource(nodeInf *api.NodeInfo, name string, value int) {
	idle := api.Resource{ScalarResources: map[v1.ResourceName]float64{
		v1.ResourceName(name): float64(value) * NPUHexKilo}}
	nodeInf.Idle = &idle
	Capability := api.Resource{ScalarResources: map[v1.ResourceName]float64{
		v1.ResourceName(name): float64(value) * NPUHexKilo}}
	nodeInf.Capability = &Capability
	nodeInf.Allocatable = &Capability
}
