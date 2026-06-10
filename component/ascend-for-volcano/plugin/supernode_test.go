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

package plugin

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

type npuNodeBuilder struct {
	name   string
	labels map[string]string
}

func newNPUNodeBuilder(name string) *npuNodeBuilder {
	return &npuNodeBuilder{name: name, labels: make(map[string]string)}
}

func (b *npuNodeBuilder) withTopoTree(tree string) *npuNodeBuilder {
	b.labels[util.TopoTreeLabel] = tree
	return b
}

func (b *npuNodeBuilder) withLabel(key, value string) *npuNodeBuilder {
	b.labels[key] = value
	return b
}

func (b *npuNodeBuilder) build() NPUNode {
	return NPUNode{CommonNode: CommonNode{Name: b.name, Label: b.labels}}
}

type resourceLevelBuilder struct {
	levels []util.ResourceTreeLevel
}

func newResourceLevels() *resourceLevelBuilder {
	return &resourceLevelBuilder{levels: make([]util.ResourceTreeLevel, 0)}
}

func (b *resourceLevelBuilder) withTree() *resourceLevelBuilder {
	b.levels = append(b.levels, util.ResourceTreeLevel{Type: util.LevelTypeTree, Label: util.TopoTreeLabel})
	return b
}

func (b *resourceLevelBuilder) withMiddle(label string) *resourceLevelBuilder {
	b.levels = append(b.levels, util.ResourceTreeLevel{Type: util.LevelTypeMiddle, Label: label})
	return b
}

func (b *resourceLevelBuilder) withNode() *resourceLevelBuilder {
	b.levels = append(b.levels, util.ResourceTreeLevel{Type: util.LevelTypeNode})
	return b
}

func (b *resourceLevelBuilder) build() []util.ResourceTreeLevel {
	return b.levels
}

type taskLevelBuilder struct {
	levels []util.TaskTreeLevel
}

func newTaskLevels() *taskLevelBuilder {
	return &taskLevelBuilder{levels: make([]util.TaskTreeLevel, 0)}
}

func (b *taskLevelBuilder) withLevel(name string, reqNode int) *taskLevelBuilder {
	b.levels = append(b.levels, util.TaskTreeLevel{Name: name, ReqNode: reqNode})
	return b
}

func (b *taskLevelBuilder) withNode(reqNode int) *taskLevelBuilder {
	return b.withLevel("node", reqNode)
}

func (b *taskLevelBuilder) build() []util.TaskTreeLevel {
	return b.levels
}

func newTaskNode(name string) *util.TaskNode {
	return &util.TaskNode{ResourceNodeName: name, Children: make(map[int]*util.TaskNode)}
}

func newTaskTree(rootName string, levels []util.TaskTreeLevel, resLevels []util.ResourceTreeLevel) *util.TaskTree {
	return &util.TaskTree{TaskNode: newTaskNode(rootName), ResourceLevels: resLevels, Levels: levels}
}

func newConverter(taskLevels []util.TaskTreeLevel, resLevels []util.ResourceTreeLevel, nodes map[string]NPUNode) *taskTreeConverter {
	return &taskTreeConverter{
		firstLevelLogicGroup: make(map[string][]SuperNode),
		taskLevels:           taskLevels,
		resourceLevels:       resLevels,
		nodes:                nodes}
}

type getResourceTreesCase struct {
	name              string
	npuNodes          map[string]NPUNode
	resourceLevelsMap map[string][]util.ResourceTreeLevel
	taskLevel         []util.TaskTreeLevel
	wantErr           bool
	wantTreeCount     int
}

func buildGetResourceTreesCases() []getResourceTreesCase {
	return []getResourceTreesCase{
		{"empty npuNodes", map[string]NPUNode{}, map[string][]util.ResourceTreeLevel{"tree": {}},
			newTaskLevels().withLevel("job", 1).withNode(1).build(), true, 0},
		{"nil resourceLevelsMap", map[string]NPUNode{"node1": {}}, nil,
			newTaskLevels().withLevel("job", 1).withNode(1).build(), true, 0},
		{"valid input", map[string]NPUNode{"node1": newNPUNodeBuilder("node1").withTopoTree("tree").
			withLabel("level1", "rack1").build()},
			map[string][]util.ResourceTreeLevel{"tree": newResourceLevels().withTree().withMiddle("level1").withNode().build()},
			newTaskLevels().withLevel("job", 1).withNode(1).build(), false, 1},
		{"multiple topoTrees", map[string]NPUNode{
			"node1": newNPUNodeBuilder("node1").withTopoTree("tree1").build(),
			"node2": newNPUNodeBuilder("node2").withTopoTree("tree2").build()},
			map[string][]util.ResourceTreeLevel{
				"tree1": newResourceLevels().withTree().withNode().build(),
				"tree2": newResourceLevels().withTree().withNode().build()},
			newTaskLevels().withLevel("job", 1).withNode(1).build(), false, 2},
		{"missing middle label", map[string]NPUNode{"node1": newNPUNodeBuilder("node1").withTopoTree("tree1").build()},
			map[string][]util.ResourceTreeLevel{"tree1": newResourceLevels().withTree().withMiddle("level1").withNode().build()},
			newTaskLevels().withLevel("job", 1).withLevel("level1", 1).withNode(1).build(), true, 0},
	}
}

func TestGetResourceTrees(t *testing.T) {
	for _, tt := range buildGetResourceTreesCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResourceTrees(tt.npuNodes, tt.resourceLevelsMap, tt.taskLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourceTrees() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantTreeCount {
				t.Errorf("GetResourceTrees() tree count = %v, want %v", len(got), tt.wantTreeCount)
			}
		})
	}
}

type getResourceTreeCase struct {
	name           string
	nodes          map[string]NPUNode
	resourceLevels []util.ResourceTreeLevel
	topoTreeName   string
	taskLevel      []util.TaskTreeLevel
	wantErr        bool
}

func buildGetResourceTreeCases() []getResourceTreeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 1).withNode(1).build()
	taskLevelsWithL1 := newTaskLevels().withLevel("job", 1).withLevel("level1", 1).withNode(1).build()
	return []getResourceTreeCase{
		{"valid nodes", map[string]NPUNode{"node1": {}}, simpleLevels, "tree", taskLevels, false},
		{"node missing label", map[string]NPUNode{"node1": {CommonNode: CommonNode{Name: "node1", Label: nil}}},
			newResourceLevels().withTree().withMiddle("level1").withNode().build(), "tree", taskLevelsWithL1, true},
		{"multiple nodes with middle level", map[string]NPUNode{
			"node1": newNPUNodeBuilder("node1").withLabel("level1", "rack1").build(),
			"node2": newNPUNodeBuilder("node2").withLabel("level1", "rack1").build()},
			newResourceLevels().withTree().withMiddle("level1").withNode().build(), "tree", taskLevelsWithL1, false},
		{"empty nodes", map[string]NPUNode{}, simpleLevels, "tree", taskLevels, false},
	}
}

func TestGetResourceTree(t *testing.T) {
	for _, tt := range buildGetResourceTreeCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getResourceTree(tt.nodes, tt.resourceLevels, tt.topoTreeName, tt.taskLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResourceTree() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("getResourceTree() should not be nil")
			}
		})
	}
}

type removeLayersCase struct {
	name      string
	origin    []util.ResourceTreeLevel
	targetLen int
	wantLen   int
}

func buildRemoveLayersCases() []removeLayersCase {
	return []removeLayersCase{
		{"same length", []util.ResourceTreeLevel{{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle}}, 2, 2},
		{"truncate", []util.ResourceTreeLevel{{Type: util.LevelTypeTree}, {Type: util.LevelTypeMiddle},
			{Type: util.LevelTypeNode}}, 2, 2},
	}
}

func TestRemoveRedundantLayers(t *testing.T) {
	for _, tt := range buildRemoveLayersCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeRedundantLayers(tt.origin, tt.targetLen); len(got) != tt.wantLen {
				t.Errorf("removeRedundantLayers() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

type groupNodesCase struct {
	name     string
	npuNodes map[string]NPUNode
	wantLen  int
}

func buildGroupNodesCases() []groupNodesCase {
	return []groupNodesCase{
		{"empty", map[string]NPUNode{}, 0},
		{"group by topoTree", map[string]NPUNode{
			"node1": newNPUNodeBuilder("node1").withTopoTree("tree1").build(),
			"node2": newNPUNodeBuilder("node2").withTopoTree("tree2").build()}, 2},
		{"missing topoTree label", map[string]NPUNode{
			"node1": {CommonNode: CommonNode{Name: "node1", Label: map[string]string{}}}}, 1},
	}
}

func TestGroupNodeByTopoTrees(t *testing.T) {
	for _, tt := range buildGroupNodesCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := groupNodeByTopoTrees(tt.npuNodes); len(got) != tt.wantLen {
				t.Errorf("groupNodeByTopoTrees() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

type healthyNodesCase struct {
	name     string
	npuNodes map[string]NPUNode
	nodes    []*api.NodeInfo
	wantLen  int
}

func buildHealthyNodesCases() []healthyNodesCase {
	return []healthyNodesCase{
		{"empty npuNodes", map[string]NPUNode{}, []*api.NodeInfo{{Name: "node1"}}, 0},
		{"matched", map[string]NPUNode{"node1": {}}, []*api.NodeInfo{{Name: "node1"}}, 1},
		{"no match", map[string]NPUNode{"node1": {}}, []*api.NodeInfo{{Name: "node2"}}, 0},
	}
}

func TestGetHealthyNPUNodes(t *testing.T) {
	for _, tt := range buildHealthyNodesCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHealthyNPUNodes(tt.npuNodes, tt.nodes); len(got) != tt.wantLen {
				t.Errorf("GetHealthyNPUNodes() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

type superNodeCase struct {
	name        string
	nodeTopoMap map[string]string
	wantName    string
	wantTree    string
}

func buildSuperNodeCases() []superNodeCase {
	return []superNodeCase{
		{"missing topoTree", map[string]string{util.NodeLevelName: "node1"}, "node1", util.DefaultTopoTree},
		{"topoTree from label", map[string]string{util.NodeLevelName: "node1", util.TopoTreeLabel: "tree1"},
			"node1", "tree1"},
		{"empty map", map[string]string{}, "", util.DefaultTopoTree},
	}
}

func TestGetSuperNode(t *testing.T) {
	for _, tt := range buildSuperNodeCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSuperNode(tt.nodeTopoMap)
			if err != nil || got.Name != tt.wantName || got.TopoTreeName != tt.wantTree {
				t.Errorf("getSuperNode() = %v, %v, want %v, %v", got.Name, got.TopoTreeName, tt.wantName, tt.wantTree)
			}
		})
	}
}

type resourceNodeNameCase struct {
	name        string
	level       util.ResourceTreeLevel
	nodeName    string
	nodeTopoMap map[string]string
	want        string
	wantErr     bool
}

func buildResourceNodeNameCases() []resourceNodeNameCase {
	return []resourceNodeNameCase{
		{"level type node", util.ResourceTreeLevel{Type: util.LevelTypeNode}, "node1", nil, "node1", false},
		{"level type tree no label", util.ResourceTreeLevel{Type: util.LevelTypeTree}, "", map[string]string{},
			util.DefaultTopoTree, false},
		{"level type tree with label", util.ResourceTreeLevel{Type: util.LevelTypeTree}, "",
			map[string]string{util.TopoTreeLabel: "tree1"}, "tree1", false},
		{"middle type nil map", util.ResourceTreeLevel{Type: util.LevelTypeMiddle, Label: "level1"},
			"node1", nil, "", true},
		{"middle type missing label", util.ResourceTreeLevel{Type: util.LevelTypeMiddle, Label: "level1"},
			"node1", map[string]string{}, "", true},
		{"middle type valid", util.ResourceTreeLevel{Type: util.LevelTypeMiddle, Label: "level1"}, "",
			map[string]string{"level1": "rack1"}, "rack1", false},
	}
}

func TestGetResourceNodeName(t *testing.T) {
	for _, tt := range buildResourceNodeNameCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getResourceNodeName(tt.level, tt.nodeName, tt.nodeTopoMap)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("getResourceNodeName() = %v, %v, want %v, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

type superNodeMapFromTaskTreeCase struct {
	name      string
	taskTree  *util.TaskTree
	wantTotal int
	wantErr   bool
}

func buildSuperNodeMapFromTaskTreeCases() []superNodeMapFromTaskTreeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 2).withNode(1).build()
	tt1 := newTaskTree("tree", taskLevels, simpleLevels)
	tt1.TaskNode.Children = map[int]*util.TaskNode{
		0: {ResourceNodeName: "node1", Index: 0},
		1: {ResourceNodeName: "node2", Index: 1}}
	resLevels := newResourceLevels().withTree().withMiddle("level1").withNode().build()
	taskLevelsWithL1 := newTaskLevels().withLevel("job", 2).withLevel(util.TopoLevelPrefix+"1", 2).withNode(1).build()
	child := newTaskNode("rack1")
	child.Children[0] = &util.TaskNode{ResourceNodeName: "node1", Index: 0}
	child.Children[1] = &util.TaskNode{ResourceNodeName: "node2", Index: 1}
	tt2 := newTaskTree("tree", taskLevelsWithL1, resLevels)
	tt2.TaskNode.Children = map[int]*util.TaskNode{0: child}
	return []superNodeMapFromTaskTreeCase{
		{"simple tree", tt1, 2, false},
		{"tree with level1", tt2, 2, false},
	}
}

func TestGetSuperNodeMapFromTaskTree(t *testing.T) {
	for _, tt := range buildSuperNodeMapFromTaskTreeCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSuperNodeMapFromTaskTree(tt.taskTree, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSuperNodeMapFromTaskTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			total := 0
			for _, nodes := range got {
				total += len(nodes)
			}
			if total != tt.wantTotal {
				t.Errorf("GetSuperNodeMapFromTaskTree() total = %v, want %v", total, tt.wantTotal)
			}
		})
	}
}

func TestSortNodesByCachedRank_NoCache(t *testing.T) {
	groups := map[string][]SuperNode{
		"0": {{Name: "node-c"}, {Name: "node-a"}, {Name: "node-b"}},
		"1": {{Name: "node-z"}, {Name: ""}, {Name: "node-m"}},
	}
	sortNodesByCachedRank(groups, nil)

	g0 := groups["0"]
	if g0[0].Name != "node-a" || g0[1].Name != "node-b" || g0[2].Name != "node-c" {
		t.Errorf("group 0 not sorted: got [%s,%s,%s], want [node-a,node-b,node-c]",
			g0[0].Name, g0[1].Name, g0[2].Name)
	}

	g1 := groups["1"]
	if g1[0].Name != "" || g1[1].Name != "node-m" || g1[2].Name != "node-z" {
		t.Errorf("group 1 not sorted: got [%s,%s,%s], want [\"\",node-m,node-z]",
			g1[0].Name, g1[1].Name, g1[2].Name)
	}
}

func TestSortNodesByCachedRank_WithCache(t *testing.T) {
	groups := map[string][]SuperNode{
		"0": {{Name: "node-c"}, {Name: "node-a"}, {Name: "node-b"}},
	}
	cached := map[string][]SuperNode{
		"0": {{Name: "node-a"}, {Name: "node-b"}, {Name: "node-c"}},
	}
	sortNodesByCachedRank(groups, cached)

	g0 := groups["0"]
	if g0[0].Name != "node-a" || g0[1].Name != "node-b" || g0[2].Name != "node-c" {
		t.Errorf("group 0 not reordered by cache: got [%s,%s,%s], want [node-a,node-b,node-c]",
			g0[0].Name, g0[1].Name, g0[2].Name)
	}
}

func TestSortNodesByCachedRank_WithCachePartial(t *testing.T) {
	groups := map[string][]SuperNode{
		"0": {{Name: "node-b"}, {Name: "node-d"}, {Name: "node-c"}},
	}
	cached := map[string][]SuperNode{
		"0": {{Name: "node-a"}, {Name: "node-b"}, {Name: "node-c"}},
	}
	sortNodesByCachedRank(groups, cached)

	g0 := groups["0"]
	if g0[0].Name != "node-d" || g0[1].Name != "node-b" || g0[2].Name != "node-c" {
		t.Errorf("group 0: got [%s,%s,%s], want [node-d,node-b,node-c]",
			g0[0].Name, g0[1].Name, g0[2].Name)
	}
}

type taskTreeFromSuperNodeCase struct {
	name       string
	superPods  map[string][]SuperNode
	taskLevels []util.TaskTreeLevel
	resLevels  []util.ResourceTreeLevel
	nodes      map[string]NPUNode
	wantErr    bool
}

func buildTaskTreeFromSuperNodeCases() []taskTreeFromSuperNodeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 1).withNode(1).build()
	resLevels := newResourceLevels().withTree().withMiddle("level1").withNode().build()
	taskLevelsWithL1 := newTaskLevels().withLevel("job", 2).withLevel(util.TopoLevelPrefix+"1", 2).withNode(1).build()
	return []taskTreeFromSuperNodeCase{
		{"invalid logic group", map[string][]SuperNode{"invalid": {}}, taskLevels, simpleLevels, map[string]NPUNode{}, true},
		{"node not exist", map[string][]SuperNode{"0": {{Name: "node1", TopoTreeName: "tree"}}},
			taskLevels, simpleLevels, map[string]NPUNode{}, true},
		{"valid input", map[string][]SuperNode{"0": {{Name: "node1", TopoTreeName: "tree"}}},
			taskLevels, simpleLevels, map[string]NPUNode{"node1": newNPUNodeBuilder("node1").withTopoTree("tree").build()}, false},
		{"topo mismatch", map[string][]SuperNode{"0": {{Name: "node1", TopoTreeName: "tree1"}}},
			taskLevels, simpleLevels, map[string]NPUNode{"node1": newNPUNodeBuilder("node1").withTopoTree("tree2").build()}, true},
		{"multiple nodes", map[string][]SuperNode{
			"0": {{Name: "node1", TopoTreeName: "tree"}},
			"1": {{Name: "node2", TopoTreeName: "tree"}}},
			taskLevels, simpleLevels, map[string]NPUNode{
				"node1": newNPUNodeBuilder("node1").withTopoTree("tree").build(),
				"node2": newNPUNodeBuilder("node2").withTopoTree("tree").build()}, false},
		{"with level1", map[string][]SuperNode{
			"0": {{Name: "node1", TopoTreeName: "tree"}, {Name: "node2", TopoTreeName: "tree"}}},
			taskLevelsWithL1, resLevels, map[string]NPUNode{
				"node1": newNPUNodeBuilder("node1").withTopoTree("tree").withLabel("level1", "rack1").build(),
				"node2": newNPUNodeBuilder("node2").withTopoTree("tree").withLabel("level1", "rack1").build()}, false},
	}
}

func TestGetTaskTreeFromSuperNodeMap(t *testing.T) {
	for _, tt := range buildTaskTreeFromSuperNodeCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTaskTreeFromSuperNodeMap(tt.superPods, tt.taskLevels, tt.resLevels, tt.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskTreeFromSuperNodeMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("GetTaskTreeFromSuperNodeMap() should not be nil")
			}
		})
	}
}

type buildSuperNodeMapCase struct {
	name        string
	taskNode    *util.TaskNode
	depth       int
	taskLevels  []util.TaskTreeLevel
	resLevels   []util.ResourceTreeLevel
	wantErr     bool
	wantEntries int
}

func buildSuperNodeMapCases() []buildSuperNodeMapCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 1).withNode(1).build()
	tn1 := newTaskNode("node1")
	tn2 := newTaskNode("tree")
	tn2.Children[0] = &util.TaskNode{ResourceNodeName: "node1", Index: 0}
	tn2.Children[1] = &util.TaskNode{ResourceNodeName: "node2", Index: 1}
	resLevels := newResourceLevels().withTree().withMiddle("level1").withNode().build()
	taskLevelsWithL1 := newTaskLevels().withLevel("job", 1).withLevel(util.TopoLevelPrefix+"1", 1).withNode(1).build()
	child := newTaskNode("rack1")
	child.Children[0] = &util.TaskNode{ResourceNodeName: "node1", Index: 0}
	tn3 := newTaskNode("tree")
	tn3.Children[0] = child
	tn4 := newTaskNode("node1")
	return []buildSuperNodeMapCase{
		{"depth out of range", tn1, 10, taskLevels, simpleLevels, true, 0},
		{"with children", tn2, 0, taskLevels, simpleLevels, false, 2},
		{"multiple levels", tn3, 0, taskLevelsWithL1, resLevels, false, 1},
		{"node level type", tn4, 0, newTaskLevels().withLevel("job", 1).build(),
			[]util.ResourceTreeLevel{{Type: util.LevelTypeNode}}, false, 1},
	}
}

func TestBuildSuperNodeMap(t *testing.T) {
	for _, tt := range buildSuperNodeMapCases() {
		t.Run(tt.name, func(t *testing.T) {
			converter := newConverter(tt.taskLevels, tt.resLevels, nil)
			err := converter.buildSuperNodeMap(tt.taskNode, tt.depth, 0, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildSuperNodeMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(converter.firstLevelLogicGroup) != tt.wantEntries {
				t.Errorf("buildSuperNodeMap() entries = %v, want %v",
					len(converter.firstLevelLogicGroup), tt.wantEntries)
			}
		})
	}
}

type logicGroupSizeCase struct {
	name       string
	taskLevels []util.TaskTreeLevel
	cachedSize int
	want       int
}

func buildLogicGroupSizeCases() []logicGroupSizeCase {
	return []logicGroupSizeCase{
		{"no level1", newTaskLevels().withLevel("job", 4).withNode(1).build(), 0, 1},
		{"level1 found", newTaskLevels().withLevel("job", 8).withLevel(util.TopoLevelPrefix+"1", 4).build(), 0, 4},
		{"cached value", newTaskLevels().withLevel("job", 4).build(), 8, 8},
	}
}

func TestGetLogicGroupSize(t *testing.T) {
	for _, tt := range buildLogicGroupSizeCases() {
		t.Run(tt.name, func(t *testing.T) {
			converter := &taskTreeConverter{taskLevels: tt.taskLevels, firstLevelLogicGroupSize: tt.cachedSize}
			if got := converter.getLogicGroupSize(); got != tt.want {
				t.Errorf("getLogicGroupSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

type rootResourceNodeCase struct {
	name      string
	node      SuperNode
	nodes     map[string]NPUNode
	resLevels []util.ResourceTreeLevel
	wantErr   bool
}

func buildRootResourceNodeCases() []rootResourceNodeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	return []rootResourceNodeCase{
		{"node not exist", SuperNode{Name: "node1"}, map[string]NPUNode{}, simpleLevels, true},
		{"valid", SuperNode{Name: "node1", TopoTreeName: "tree"},
			map[string]NPUNode{"node1": newNPUNodeBuilder("node1").withTopoTree("tree").build()}, simpleLevels, false},
	}
}

func TestGetRootResourceNodeName(t *testing.T) {
	for _, tt := range buildRootResourceNodeCases() {
		t.Run(tt.name, func(t *testing.T) {
			converter := &taskTreeConverter{resourceLevels: tt.resLevels, nodes: tt.nodes}
			_, err := converter.getRootResourceNodeName(tt.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRootResourceNodeName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type addTaskNodeCase struct {
	name         string
	superNode    SuperNode
	taskLevels   []util.TaskTreeLevel
	resLevels    []util.ResourceTreeLevel
	nodes        map[string]NPUNode
	existing     map[int]*util.TaskNode
	wantErr      bool
	wantChildren int
}

func buildAddTaskNodeCases() []addTaskNodeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 1).withNode(1).build()
	resLevels := newResourceLevels().withTree().withMiddle("level1").withNode().build()
	return []addTaskNodeCase{
		{"valid", SuperNode{Name: "node1"}, taskLevels, simpleLevels,
			map[string]NPUNode{"node1": {}}, nil, false, 1},
		{"middle level nil label", SuperNode{Name: "node1"},
			newTaskLevels().withLevel("job", 1).withLevel("level1", 1).withNode(1).build(), resLevels,
			map[string]NPUNode{"node1": {CommonNode: CommonNode{Name: "node1", Label: nil}}}, nil, true, 0},
		{"existing child continue", SuperNode{Name: "node1"}, taskLevels, simpleLevels,
			map[string]NPUNode{"node1": {}}, map[int]*util.TaskNode{0: {Index: 0, ResourceNodeName: "node1"}}, false, 1},
		{"multiple children", SuperNode{Name: "node1"}, taskLevels, simpleLevels,
			map[string]NPUNode{"node1": {}, "node2": {}}, nil, false, 1},
	}
}

func TestAddTaskNode(t *testing.T) {
	for _, tt := range buildAddTaskNodeCases() {
		t.Run(tt.name, func(t *testing.T) {
			converter := newConverter(tt.taskLevels, tt.resLevels, tt.nodes)
			rootNode := &util.TaskNode{Children: tt.existing}
			if rootNode.Children == nil {
				rootNode.Children = make(map[int]*util.TaskNode)
			}
			err := converter.addTaskNode(tt.superNode, rootNode, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("addTaskNode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(rootNode.Children) != tt.wantChildren {
				t.Errorf("addTaskNode() children = %v, want %v", len(rootNode.Children), tt.wantChildren)
			}
		})
	}
}

type buildTaskTreeCase struct {
	name       string
	superPods  map[string][]SuperNode
	nodes      map[string]NPUNode
	taskLevels []util.TaskTreeLevel
	resLevels  []util.ResourceTreeLevel
	groupSize  int
	wantErr    bool
	wantRoot   string
}

func buildTaskTreeCases() []buildTaskTreeCase {
	simpleLevels := newResourceLevels().withTree().withNode().build()
	taskLevels := newTaskLevels().withLevel("job", 1).withNode(1).build()
	return []buildTaskTreeCase{
		{"empty superPods", map[string][]SuperNode{}, map[string]NPUNode{}, taskLevels, simpleLevels, 2, false, ""},
		{"multiple nodes", map[string][]SuperNode{
			"0": {{Name: "node1", TopoTreeName: "tree"}, {Name: "node2", TopoTreeName: "tree"}}},
			map[string]NPUNode{
				"node1": newNPUNodeBuilder("node1").withTopoTree("tree").build(),
				"node2": newNPUNodeBuilder("node2").withTopoTree("tree").build()},
			taskLevels, simpleLevels, 2, false, "tree"},
	}
}

func TestBuildTaskTree(t *testing.T) {
	for _, tt := range buildTaskTreeCases() {
		t.Run(tt.name, func(t *testing.T) {
			converter := &taskTreeConverter{
				firstLevelLogicGroup:     tt.superPods,
				taskLevels:               tt.taskLevels,
				resourceLevels:           tt.resLevels,
				nodes:                    tt.nodes,
				firstLevelLogicGroupSize: tt.groupSize}
			got, err := converter.buildTaskTree()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTaskTree() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("buildTaskTree() should not be nil")
			}
			if !tt.wantErr && tt.wantRoot != "" && got.TaskNode.ResourceNodeName != tt.wantRoot {
				t.Errorf("buildTaskTree() root = %v, want %v", got.TaskNode.ResourceNodeName, tt.wantRoot)
			}
		})
	}
}
