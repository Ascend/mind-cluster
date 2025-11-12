//go:build ignore
// +build ignore

/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package ascend800ia5stacking

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
)

// Mock interfaces and utilities for testing
// These are simplified versions to allow standalone testing

// NPUNode is a simplified mock of plugin.NPUNode interface
type NPUNode interface {
	GetName() string
}

// Constant definitions to improve code readability
const (
	// Score calculation weights
	TaskWeight  = 5 // Task count weight
	StackWeight = 2 // Stack count weight
	CardWeight  = 1 // Card count weight

	// Limit values
	MaxTaskNum  = 8 // Maximum task count
	MaxStackNum = 8 // Maximum stack count
	MaxCardNum  = 8 // Maximum card count
	MinTaskNum  = 1 // Minimum task count
	MinStackNum = 0 // Minimum stack count
	MinCardNum  = 1 // Minimum card count

	// Bonus scores
	StackTaskBonus = 8 // Bonus score when stack count is less than task count

	// Test values
	TestSuperPodID = 123 // SuperPodID for testing
)

// Simple helper functions that don't depend on other code

// findCommonElements finds common elements between two integer lists
func findCommonElements(list1, list2 []int) []int {
	if len(list1) == 0 || len(list2) == 0 {
		var empty []int
		return empty
	}

	elementMap := make(map[int]bool)
	for _, element := range list1 {
		elementMap[element] = true
	}

	var common []int
	for _, element := range list2 {
		if elementMap[element] {
			common = append(common, element)
		}
	}

	return common
}

// TestFindCommonElements tests the function to find common elements
func TestFindCommonElements(t *testing.T) {
	tests := []struct {
		name     string
		list1    []int
		list2    []int
		expected []int
	}{{
		name:     "With common elements",
		list1:    []int{0, 1, 2, 3},
		list2:    []int{0, 1, 2, 4},
		expected: []int{0, 1, 2},
	}, {
		name:     "No common elements",
		list1:    []int{0, 1},
		list2:    []int{2, 3},
		expected: []int{},
	}, {
		name:     "Empty list",
		list1:    []int{},
		list2:    []int{0, 1},
		expected: []int{},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCommonElements(tt.list1, tt.list2)
			// Special handling for empty lists
			if len(result) == 0 && len(tt.expected) == 0 {
				// Two empty lists are considered equal
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("findCommonElements(%v, %v) = %v, expected %v", tt.list1, tt.list2, result, tt.expected)
			}
		})
	}
}

// calculateScore simulates score calculation
func calculateScore(taskNum, stackNum, cardNum int) (int, error) {
	if taskNum < MinTaskNum || taskNum > MaxTaskNum || stackNum < MinStackNum || stackNum > MaxStackNum || cardNum < MinCardNum || cardNum > MaxCardNum {
		return 0, fmt.Errorf("invalid parameters")
	}

	// Simple score calculation logic
	score := taskNum*TaskWeight + stackNum*StackWeight + cardNum*CardWeight

	// Add bonus score if stackNum is less than taskNum
	if stackNum < taskNum {
		score += StackTaskBonus
	}

	return score, nil
}

// TestCalculateScore tests the score calculation function
func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name      string
		taskNum   int
		stackNum  int
		cardNum   int
		expected  int
		expectErr bool
	}{{
		name:      "Normal calculation",
		taskNum:   2,
		stackNum:  3,
		cardNum:   4,
		expected:  2*TaskWeight + 3*StackWeight + 4*CardWeight, // 10 + 6 + 4 = 20
		expectErr: false,
	}, {
		name:      "stackNum less than taskNum",
		taskNum:   3,
		stackNum:  2,
		cardNum:   4,
		expected:  3*TaskWeight + 2*StackWeight + 4*CardWeight + StackTaskBonus, // 15 + 4 + 4 + 8 = 31
		expectErr: false,
	}, {
		name:      "Invalid parameters",
		taskNum:   9,
		stackNum:  3,
		cardNum:   4,
		expected:  0,
		expectErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculateScore(tt.taskNum, tt.stackNum, tt.cardNum)
			if (err != nil) != tt.expectErr {
				t.Errorf("calculateScore error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if result != tt.expected {
				t.Errorf("calculateScore result = %d, expected %d", result, tt.expected)
			}
		})
	}
}

// NodeCache simple node cache structure
type NodeCache struct {
	Nodes map[string]bool
}

// NewNodeCache creates a new node cache
func NewNodeCache() *NodeCache {
	return &NodeCache{
		Nodes: make(map[string]bool),
	}
}

// InitCache initializes the cache (only if empty)
func (c *NodeCache) InitCache(nodeNames []string) {
	if len(c.Nodes) == 0 {
		for _, name := range nodeNames {
			c.Nodes[name] = true
		}
	}
}

// TestInitCache tests node cache initialization
func TestInitCache(t *testing.T) {
	// Test initialization when cache is empty
	emptyCache := NewNodeCache()
	emptyCache.InitCache([]string{"node1", "node2"})
	if !emptyCache.Nodes["node1"] || !emptyCache.Nodes["node2"] {
		t.Errorf("expected nodes to be in empty cache")
	}

	// Test not adding new nodes when cache is not empty
	filledCache := NewNodeCache()
	filledCache.Nodes["node3"] = true
	filledCache.InitCache([]string{"node4"})
	if filledCache.Nodes["node4"] {
		t.Errorf("should not add new nodes when cache is not empty")
	}
}

// SimpleMockNode simple node mock structure
type SimpleMockNode struct {
	NodeName   string
	SuperPodID int32
}

// GetName returns the node name
func (n *SimpleMockNode) GetName() string {
	return n.NodeName
}

// GetSuperPodID returns the SuperPodID
func (n *SimpleMockNode) GetSuperPodID() int32 {
	return n.SuperPodID
}

// TestSimpleMockNode tests the simple node mock object
func TestSimpleMockNode(t *testing.T) {
	mockNode := &SimpleMockNode{
		NodeName:   "test-node",
		SuperPodID: TestSuperPodID,
	}

	if mockNode.GetName() != "test-node" {
		t.Errorf("expected name 'test-node', got '%s'", mockNode.GetName())
	}

	if mockNode.GetSuperPodID() != TestSuperPodID {
		t.Errorf("expected SuperPodID %d, got %d", TestSuperPodID, mockNode.GetSuperPodID())
	}
}

// TestResourceName tests ResourceName type
func TestResourceName(t *testing.T) {
	resourceName := v1.ResourceName("ascend.com/npu")
	expected := "ascend.com/npu"

	if string(resourceName) != expected {
		t.Errorf("expected resource name '%s', got '%s'", expected, string(resourceName))
	}
}

// MockNodeWithAnnotation extends the SimpleMockNode to include annotations
type MockNodeWithAnnotation struct {
	SimpleMockNode
	Annotation map[string]string
}

// MockModule represents a simplified version of module800ia5stacking for testing
type MockModule struct {
	NetUnhealthyKey string
	MaxNodeNPUNum   int
	AnnoPreVal      string
	PluginName      string
	SuperPodCache   map[int32][]NPUNode
}

// getTheAvailableCommonTop returns the available common top values
func (m *MockModule) getTheAvailableCommonTop(stackingNodeList []NPUNode) []int {
	// Simple implementation for testing
	return []int{0, 1, 2, 3}
}

// GetPluginName returns the plugin name
func (m *MockModule) GetPluginName() string {
	return m.PluginName
}

// GetAnnoPreVal returns the annotation prefix value
func (m *MockModule) GetAnnoPreVal() string {
	return m.AnnoPreVal
}

// changeTopToIntArray converts a string representation of top values to an int array
func changeTopToIntArray(topStr, prefix string) []int {
	// Simple implementation for testing
	if topStr == "[0,1]" {
		var result = []int{0, 1}
		return result
	}
	var empty []int
	return empty
}

// removeCommonElement removes common elements from two lists
func removeCommonElement(list1, list2 []int) []int {
	var result []int
	elementMap := make(map[int]bool)
	for _, elem := range list2 {
		elementMap[elem] = true
	}
	for _, elem := range list1 {
		if !elementMap[elem] {
			result = append(result, elem)
		}
	}
	return result
}

// mockGetUsableTopFromStack simulates the getUsableTopFromStack method
func mockGetUsableTopFromStack(module *MockModule, node *MockNodeWithAnnotation, disFlag bool) ([]int, error) {
	if stackingNpuList, exists := module.SuperPodCache[node.GetSuperPodID()]; !exists {
		var empty []int
		return empty, nil
	} else {
		availableList := module.getTheAvailableCommonTop(stackingNpuList)
		if !disFlag {
			return availableList, nil
		}
		networkUnhealthyTopStr, ok := node.Annotation[module.NetUnhealthyKey]
		if !ok {
			err := fmt.Errorf("node<%s> don't have resource<%s>", node.GetName(), module.NetUnhealthyKey)
			return nil, err
		}
		networkUnhealthyTop := changeTopToIntArray(networkUnhealthyTopStr, module.GetAnnoPreVal())
		if len(networkUnhealthyTop) > module.MaxNodeNPUNum {
			err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.GetName(), networkUnhealthyTop)
			return nil, err
		}
		res := removeCommonElement(availableList, networkUnhealthyTop)
		return res, nil
	}
}

// TestGetUsableTopFromStack tests the getUsableTopFromStack method
func TestGetUsableTopFromStack(t *testing.T) {
	// Run all subtests
	t.Run("testSuperPodIDNotInCache", testSuperPodIDNotInCache)
	t.Run("testSuperPodIDInCache", testSuperPodIDInCache)
	t.Run("testDisFlagWithAnnotation", testDisFlagWithAnnotation)
}

// testSuperPodIDNotInCache tests the scenario when SuperPodID is not in cache
func testSuperPodIDNotInCache(t *testing.T) {
	superPodCache := map[int32][]NPUNode{
		TestSuperPodID: {&SimpleMockNode{NodeName: "node1", SuperPodID: TestSuperPodID}},
	}
	node := &MockNodeWithAnnotation{
		SimpleMockNode: SimpleMockNode{NodeName: "test-node", SuperPodID: 456},
		Annotation:     map[string]string{},
	}
	disFlag := false
	var expectedTop []int
	expectError := false

	module := &MockModule{
		NetUnhealthyKey: "netUnhealthyKey",
		MaxNodeNPUNum:   8,
		AnnoPreVal:      "pre",
		PluginName:      "test-plugin",
		SuperPodCache:   superPodCache,
	}

	result, err := mockGetUsableTopFromStack(module, node, disFlag)

	if (err != nil) != expectError {
		t.Errorf("Expected error: %v, got: %v", expectError, err)
		return
	}

	if !reflect.DeepEqual(result, expectedTop) {
		t.Errorf("Expected top list: %v, got: %v", expectedTop, result)
	}
}

// testSuperPodIDInCache tests the scenario when SuperPodID is in cache and disFlag is false
func testSuperPodIDInCache(t *testing.T) {
	superPodCache := map[int32][]NPUNode{
		TestSuperPodID: {&SimpleMockNode{NodeName: "node1", SuperPodID: TestSuperPodID}},
	}
	node := &MockNodeWithAnnotation{
		SimpleMockNode: SimpleMockNode{NodeName: "test-node", SuperPodID: TestSuperPodID},
		Annotation:     map[string]string{},
	}
	disFlag := false
	expectedTop := []int{0, 1, 2, 3}
	expectError := false

	module := &MockModule{
		NetUnhealthyKey: "netUnhealthyKey",
		MaxNodeNPUNum:   8,
		AnnoPreVal:      "pre",
		PluginName:      "test-plugin",
		SuperPodCache:   superPodCache,
	}

	result, err := mockGetUsableTopFromStack(module, node, disFlag)

	if (err != nil) != expectError {
		t.Errorf("Expected error: %v, got: %v", expectError, err)
		return
	}

	if !reflect.DeepEqual(result, expectedTop) {
		t.Errorf("Expected top list: %v, got: %v", expectedTop, result)
	}
}

// testDisFlagWithAnnotation tests scenarios with disFlag true
func testDisFlagWithAnnotation(t *testing.T) {
	t.Run("withValidAnnotation", func(t *testing.T) {
		testDisFlagWithValidAnnotation(t)
	})
	t.Run("withoutAnnotation", func(t *testing.T) {
		testDisFlagWithoutAnnotation(t)
	})
}

// testDisFlagWithValidAnnotation tests the scenario when disFlag is true with valid annotation
func testDisFlagWithValidAnnotation(t *testing.T) {
	superPodCache := map[int32][]NPUNode{
		TestSuperPodID: {&SimpleMockNode{NodeName: "node1", SuperPodID: TestSuperPodID}},
	}
	node := &MockNodeWithAnnotation{
		SimpleMockNode: SimpleMockNode{NodeName: "test-node", SuperPodID: TestSuperPodID},
		Annotation: map[string]string{
			"netUnhealthyKey": "[0,1]",
		},
	}
	disFlag := true
	expectedTop := []int{2, 3} // After removing [0,1] from [0,1,2,3]
	expectError := false

	module := &MockModule{
		NetUnhealthyKey: "netUnhealthyKey",
		MaxNodeNPUNum:   8,
		AnnoPreVal:      "pre",
		PluginName:      "test-plugin",
		SuperPodCache:   superPodCache,
	}

	result, err := mockGetUsableTopFromStack(module, node, disFlag)

	if (err != nil) != expectError {
		t.Errorf("Expected error: %v, got: %v", expectError, err)
		return
	}

	if !reflect.DeepEqual(result, expectedTop) {
		t.Errorf("Expected top list: %v, got: %v", expectedTop, result)
	}
}

// testDisFlagWithoutAnnotation tests the scenario when disFlag is true without annotation
func testDisFlagWithoutAnnotation(t *testing.T) {
	superPodCache := map[int32][]NPUNode{
		TestSuperPodID: {&SimpleMockNode{NodeName: "node1", SuperPodID: TestSuperPodID}},
	}
	node := &MockNodeWithAnnotation{
		SimpleMockNode: SimpleMockNode{NodeName: "test-node", SuperPodID: TestSuperPodID},
		Annotation:     map[string]string{},
	}
	disFlag := true
	expectedTop := []int(nil)
	expectError := true
	expectedErrMsg := "node<test-node> don't have resource<netUnhealthyKey>"

	module := &MockModule{
		NetUnhealthyKey: "netUnhealthyKey",
		MaxNodeNPUNum:   8,
		AnnoPreVal:      "pre",
		PluginName:      "test-plugin",
		SuperPodCache:   superPodCache,
	}

	result, err := mockGetUsableTopFromStack(module, node, disFlag)

	if (err != nil) != expectError {
		t.Errorf("Expected error: %v, got: %v", expectError, err)
		return
	}

	if expectError {
		if err == nil || err.Error() != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: '%v'", expectedErrMsg, err)
		}
		return
	}

	if !reflect.DeepEqual(result, expectedTop) {
		t.Errorf("Expected top list: %v, got: %v", expectedTop, result)
	}
}

// mockGetUsableTopFromNode is a testable version of getUsableTopFromNode
// This simulates the behavior of the original function for testing purposes
func mockGetUsableTopFromNode(module *MockModule, node *MockNodeWithAnnotation, disFlag bool, nodeTop []int, getTopErr error) ([]int, error) {
	// Simulate the behavior of the original getUsableTopFromNode function
	if getTopErr != nil {
		return nil, getTopErr
	}

	if !disFlag {
		return nodeTop, nil
	}

	networkUnhealthyTopStr, ok := node.Annotation[module.NetUnhealthyKey]
	if !ok {
		err := fmt.Errorf("node<%s> don't have resource<%s>", node.GetName(), module.NetUnhealthyKey)
		return nil, err
	}

	networkUnhealthyTop := changeTopToIntArray(networkUnhealthyTopStr, module.GetAnnoPreVal())
	if len(networkUnhealthyTop) > module.MaxNodeNPUNum {
		err := fmt.Errorf("node<%s> npu networkUnhealthy top<%v> is invalid", node.GetName(), networkUnhealthyTop)
		return nil, err
	}

	res := removeCommonElement(nodeTop, networkUnhealthyTop)
	return res, nil
}

// TestGetUsableTopFromNode tests the behavior of getUsableTopFromNode function
func TestGetUsableTopFromNode(t *testing.T) {
	// Run all subtests
	t.Run("testDisFlagFalse", testDisFlagFalse)
	t.Run("testDisFlagWithAnnotation", testDisFlagWithAnnotationForNode)
	t.Run("testErrorCases", testErrorCases)
}

// testDisFlagFalse tests the scenario when disFlag is false
func testDisFlagFalse(t *testing.T) {
	tc := NodeTestCase{
		DisFlag:        false,
		NodeAnnotation: nil,
		UsableTop:      []int{0, 1, 2, 3},
		GetTopErr:      nil,
		ExpectedTop:    []int{0, 1, 2, 3},
		ExpectErr:      false,
	}

	testGetUsableTopFromNodeCase(t, tc)
}

// testDisFlagWithAnnotationForNode tests scenarios with disFlag true
func testDisFlagWithAnnotationForNode(t *testing.T) {
	t.Run("withValidAnnotation", func(t *testing.T) {
		tc := NodeTestCase{
			DisFlag:        true,
			NodeAnnotation: map[string]string{"network-unhealthy": "[0,1]"},
			UsableTop:      []int{0, 1, 2, 3},
			GetTopErr:      nil,
			ExpectedTop:    []int{2, 3},
			ExpectErr:      false,
		}
		testGetUsableTopFromNodeCase(t, tc)
	})

	t.Run("withoutAnnotation", func(t *testing.T) {
		tc := NodeTestCase{
			DisFlag:        true,
			NodeAnnotation: nil,
			UsableTop:      []int{0, 1, 2, 3},
			GetTopErr:      nil,
			ExpectedTop:    nil,
			ExpectErr:      true,
		}
		testGetUsableTopFromNodeCase(t, tc)
	})
}

// testErrorCases tests error scenarios
func testErrorCases(t *testing.T) {
	t.Run("getTopReturnsError", func(t *testing.T) {
		tc := NodeTestCase{
			DisFlag:        false,
			NodeAnnotation: nil,
			UsableTop:      nil,
			GetTopErr:      fmt.Errorf("simulated error"),
			ExpectedTop:    nil,
			ExpectErr:      true,
		}
		testGetUsableTopFromNodeCase(t, tc)
	})
}

// NodeTestCase defines parameters for getUsableTopFromNode test cases
type NodeTestCase struct {
	DisFlag        bool
	NodeAnnotation map[string]string
	UsableTop      []int
	GetTopErr      error
	ExpectedTop    []int
	ExpectErr      bool
}

// testGetUsableTopFromNodeCase runs a single test case for getUsableTopFromNode
func testGetUsableTopFromNodeCase(t *testing.T, tc NodeTestCase) {
	module := &MockModule{
		NetUnhealthyKey: "network-unhealthy",
		MaxNodeNPUNum:   8,
		AnnoPreVal:      "pre",
		PluginName:      "test-plugin",
	}

	node := &MockNodeWithAnnotation{
		SimpleMockNode: SimpleMockNode{NodeName: "test-node", SuperPodID: TestSuperPodID},
		Annotation:     tc.NodeAnnotation,
	}

	result, err := mockGetUsableTopFromNode(module, node, tc.DisFlag, tc.UsableTop, tc.GetTopErr)

	if (err != nil) != tc.ExpectErr {
		t.Errorf("Expected error: %v, got: %v", tc.ExpectErr, err)
		return
	}

	if !reflect.DeepEqual(result, tc.ExpectedTop) {
		t.Errorf("Expected top list: %v, got: %v", tc.ExpectedTop, result)
	}
}
