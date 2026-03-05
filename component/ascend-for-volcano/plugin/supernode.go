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

/*
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// GetResourceTrees gets resource tree
func GetResourceTrees(npuNodes map[string]NPUNode, resourceLevelsMap map[string][]util.ResourceTreeLevel,
	taskLevel []util.TaskTreeLevel) ([]*util.ResourceTree, error) {
	groupedNodes := groupNodeByTopoTrees(npuNodes)
	var trees []*util.ResourceTree
	// generate scheduling trees based on current nodes
	for topoTreeName, resourceLevels := range resourceLevelsMap {
		// filter out trees with insufficient network levels
		if len(resourceLevels) < len(taskLevel) {
			continue
		}
		nodes := groupedNodes[topoTreeName]
		tree, err := getResourceTree(nodes, resourceLevels, topoTreeName, taskLevel)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("get resource tree failed, %v", err)
			continue
		}
		klog.V(util.LogDebugLev).Infof("get resource tree %v success", topoTreeName)
		trees = append(trees, tree)
	}
	if len(trees) == 0 {
		return nil, fmt.Errorf("no resource tree found for npu nodes")
	}
	return trees, nil
}

func getResourceTree(nodes map[string]NPUNode, resourceLevels []util.ResourceTreeLevel,
	topoTreeName string, taskLevel []util.TaskTreeLevel) (*util.ResourceTree, error) {
	resourceLevels = removeRedundantLayers(resourceLevels, len(taskLevel))
	// trees in the same configuration are definitely interconnected, use tree name as root node name
	rootNode := util.ResourceNode{
		Name: topoTreeName,
	}
	for _, node := range nodes {
		currentNode := &rootNode
		for index, resourceLevel := range resourceLevels {
			if index == 0 {
				continue
			}
			childName, err := getResourceNodeName(resourceLevel, node.Name, node.Label)
			if err != nil {
				return nil, fmt.Errorf("get resource node name failed, %v", err)
			}
			var nextNode *util.ResourceNode
			if child, ok := currentNode.Children[childName]; ok {
				nextNode = child
			} else {
				nextNode = &util.ResourceNode{Name: childName}
				currentNode.AddOrUpdateChild(nextNode)
			}
			currentNode = nextNode
		}
	}
	tree := util.ResourceTree{
		ResourceNode: &rootNode,
		Name:         topoTreeName,
		Levels:       resourceLevels,
	}
	return &tree, nil
}

// removeRedundantLayers removes redundant layers for resource node
func removeRedundantLayers(origin []util.ResourceTreeLevel, targetLen int) []util.ResourceTreeLevel {
	if len(origin) == targetLen {
		return origin
	}
	usingResourceLevels := make([]util.ResourceTreeLevel, targetLen)
	for i := 0; i < targetLen; i++ {
		usingResourceLevels[targetLen-1-i] = origin[len(origin)-1-i]
	}
	usingResourceLevels[0] = util.ResourceTreeLevel{Type: util.LevelTypeTree, Label: util.TopoTreeLabel}
	return usingResourceLevels
}

func groupNodeByTopoTrees(npuNodes map[string]NPUNode) map[string]map[string]NPUNode {
	var groupedNodes = make(map[string]map[string]NPUNode)
	for nodeName, npuNode := range npuNodes {
		topoTree, exist := npuNode.Label[util.TopoTreeLabel]
		if !exist {
			topoTree = util.DefaultTopoTree
		}
		_, ok := groupedNodes[topoTree]
		if !ok {
			groupedNodes[topoTree] = make(map[string]NPUNode)
		}
		groupedNodes[topoTree][nodeName] = npuNode
	}
	return groupedNodes
}

// GetHealthyNPUNodes get healthy npu nodes
func GetHealthyNPUNodes(npuNodes map[string]NPUNode, nodes []*api.NodeInfo) map[string]NPUNode {
	healthyNPUNodes := make(map[string]NPUNode)
	for _, node := range nodes {
		if npuNode, ok := npuNodes[node.Name]; ok {
			healthyNPUNodes[npuNode.Name] = npuNode
		}
	}
	return healthyNPUNodes
}

// GetSuperNodeMapFromTaskTree gets level1 logic group map from task tree
func GetSuperNodeMapFromTaskTree(taskTree *util.TaskTree) (map[string][]SuperNode, error) {
	converter := taskTreeConverter{
		firstLevelLogicGroup: make(map[string][]SuperNode),
		taskLevels:           taskTree.Levels,
		resourceLevels:       taskTree.ResourceLevels,
	}

	var startRank int
	err := converter.buildSuperNodeMap(taskTree.TaskNode, 0, startRank, make(map[string]string, len(converter.resourceLevels)))
	for _, nodes := range converter.firstLevelLogicGroup {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})
	}
	return converter.firstLevelLogicGroup, err
}

// GetTaskTreeFromSuperNodeMap gets task tree from super pod map
func GetTaskTreeFromSuperNodeMap(
	superPods map[string][]SuperNode, taskLevels []util.TaskTreeLevel,
	resourceLevels []util.ResourceTreeLevel, nodes map[string]NPUNode) (*util.TaskTree, error) {
	converter := taskTreeConverter{
		firstLevelLogicGroup: superPods,
		taskLevels:           taskLevels,
		resourceLevels:       removeRedundantLayers(resourceLevels, len(taskLevels)),
		nodes:                nodes,
	}
	return converter.buildTaskTree()
}

type taskTreeConverter struct {
	firstLevelLogicGroup     map[string][]SuperNode
	taskLevels               []util.TaskTreeLevel
	resourceLevels           []util.ResourceTreeLevel
	nodes                    map[string]NPUNode
	firstLevelLogicGroupSize int
}

func (ttc *taskTreeConverter) buildSuperNodeMap(
	taskNode *util.TaskNode, depth int, startRank int, nodeTopoMap map[string]string) error {
	if depth >= len(ttc.resourceLevels) {
		return fmt.Errorf("depth out of resource levels range %v", len(ttc.resourceLevels))
	}
	var topoKey string
	resourceTreeLevel := ttc.resourceLevels[depth]
	if resourceTreeLevel.Type != util.LevelTypeNode {
		topoKey = resourceTreeLevel.Label
	} else {
		topoKey = util.NodeLevelName
	}

	nodeTopoMap[topoKey] = taskNode.ResourceNodeName
	if depth+1 >= len(ttc.taskLevels) {
		logicFirstGroupStr := strconv.Itoa(startRank / ttc.getLogicGroupSize())
		logicFirstGroup, ok := ttc.firstLevelLogicGroup[logicFirstGroupStr]
		if !ok {
			logicFirstGroup = make([]SuperNode, 0, ttc.getLogicGroupSize())
		}
		superNode, err := getSuperNode(nodeTopoMap)
		if err != nil {
			return err
		}

		ttc.firstLevelLogicGroup[logicFirstGroupStr] = append(logicFirstGroup, superNode)
		return nil
	}

	for rank, child := range taskNode.Children {
		nodeRank := startRank + rank*ttc.taskLevels[depth+1].ReqNode
		if err := ttc.buildSuperNodeMap(child, depth+1, nodeRank, nodeTopoMap); err != nil {
			return err
		}
	}

	return nil
}

func (ttc *taskTreeConverter) buildTaskTree() (*util.TaskTree, error) {
	rootNode := util.TaskNode{}
	for logicGroupStr, nodes := range ttc.firstLevelLogicGroup {
		logicGroupID, err := strconv.Atoi(logicGroupStr)
		if err != nil {
			return nil, err
		}
		for nodeIndex, node := range nodes {
			nodeRank := logicGroupID*ttc.getLogicGroupSize() + nodeIndex
			if addErr := ttc.addTaskNode(node, &rootNode, nodeRank); addErr != nil {
				return nil, addErr
			}
			if rootNode.ResourceNodeName != "" {
				continue
			}
			topotreeName, getErr := ttc.getRootResourceNodeName(node)
			if getErr != nil {
				return nil, fmt.Errorf("get root node resource node name failed, %v", getErr)
			}
			if node.TopoTreeName != topotreeName {
				return nil, fmt.Errorf("job topotree %v don't not match resource level %v",
					node.TopoTreeName, ttc.resourceLevels)
			}
		}
	}
	return &util.TaskTree{
		TaskNode:       &rootNode,
		ResourceLevels: ttc.resourceLevels,
		Levels:         ttc.taskLevels,
	}, nil
}

func (ttc *taskTreeConverter) getRootResourceNodeName(node SuperNode) (string, error) {
	if len(ttc.resourceLevels) == 0 {
		return "", errors.New("resourceLevels is empty")
	}
	npuNode, nodeExist := ttc.nodes[node.Name]
	if !nodeExist {
		return "", errors.New("npuNode is not exist")
	}
	resourceNodeName, err := getResourceNodeName(ttc.resourceLevels[0], node.Name, npuNode.Label)
	if err != nil {
		return "", err
	}
	return resourceNodeName, nil
}

func (ttc *taskTreeConverter) addTaskNode(superNode SuperNode, rootNode *util.TaskNode, nodeRank int) error {
	currentNode := rootNode
	for levelIndex, level := range ttc.taskLevels {
		if levelIndex == 0 {
			continue
		}
		groupRank := nodeRank / level.ReqNode
		nodeRank = nodeRank % level.ReqNode
		nextNode, ok := currentNode.Children[groupRank]
		if ok {
			currentNode = nextNode
			continue
		}
		resourceNodeName, err := getResourceNodeName(ttc.resourceLevels[levelIndex], superNode.Name,
			ttc.nodes[superNode.Name].Label)
		if err != nil {
			return err
		}

		nextNode = &util.TaskNode{
			Index:            groupRank,
			ResourceNodeName: resourceNodeName,
		}
		currentNode.AddOrUpdateChild(nextNode)
		currentNode = nextNode
	}
	return nil
}

func (ttc *taskTreeConverter) getLogicGroupSize() int {
	if ttc.firstLevelLogicGroupSize > 0 {
		return ttc.firstLevelLogicGroupSize
	}
	for _, level := range ttc.taskLevels {
		if level.Name == util.TopoLevelPrefix+strconv.Itoa(util.Level1Number) {
			ttc.firstLevelLogicGroupSize = level.ReqNode
			return ttc.firstLevelLogicGroupSize
		}
	}
	ttc.firstLevelLogicGroupSize = 1
	return ttc.firstLevelLogicGroupSize
}

func getSuperNode(nodeTopoMap map[string]string) (SuperNode, error) {
	treeName, ok := nodeTopoMap[util.TopoTreeLabel]
	if !ok {
		treeName = util.DefaultTopoTree
	}
	return SuperNode{
		Name:         nodeTopoMap[util.NodeLevelName],
		TopoTreeName: treeName,
	}, nil
}

func getResourceNodeName(level util.ResourceTreeLevel, nodeName string, nodeTopoMap map[string]string) (string, error) {
	klog.V(util.LogDebugLev).Infof("getResourceNodeName level: %v, nodeName: %v, nodeTopoMap: %v", level, nodeName, nodeTopoMap)
	if level.Type == util.LevelTypeNode {
		return nodeName, nil
	}
	if level.Type == util.LevelTypeTree {
		topoTreeName, ok := nodeTopoMap[util.TopoTreeLabel]
		if !ok {
			topoTreeName = util.DefaultTopoTree
		}
		return topoTreeName, nil
	}

	if nodeTopoMap == nil {
		return "", fmt.Errorf("node %s has no labels", nodeName)
	}
	id, ok := nodeTopoMap[level.Label]
	if !ok {
		return "", fmt.Errorf("node %s has no label %s", nodeName, level.Label)
	}
	return id, nil
}
