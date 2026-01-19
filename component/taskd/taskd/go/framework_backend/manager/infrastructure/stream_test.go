/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package infrastructure for taskd manager backend infrastructure
package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var priorityTests = []struct {
	name        string
	ownerList   map[string]string
	requestList []string
	expected    []string
	description string
}{
	{
		name:        "all plugins have priority",
		ownerList:   map[string]string{"old_owner": "old_owner"},
		requestList: []string{"med_priority", "high_priority", "low_priority"},
		expected:    []string{"high_priority", "med_priority", "low_priority"},
		description: "sort in ascending order of priority",
	},
	{
		name:        "some plugins have priority",
		ownerList:   nil,
		requestList: []string{"no_priority1", "high_priority", "no_priority2"},
		expected:    []string{"high_priority", "no_priority1", "no_priority2"},
		description: "priority items are ranked first, while the rest are arranged in lexicographic order",
	},
	{
		name:        "only plugins with priority",
		ownerList:   nil,
		requestList: []string{"low_priority", "med_priority", "high_priority"},
		expected:    []string{"high_priority", "med_priority", "low_priority"},
		description: "sort completely by priority",
	},
	{
		name:        "only plugins without priority",
		ownerList:   nil,
		requestList: []string{"zebra", "apple", "banana"},
		expected:    []string{"apple", "banana", "zebra"},
		description: "sort by dictionary order",
	},
	{
		name:        "filter allocated plugins",
		ownerList:   map[string]string{"high_priority": "high_priority", "med_priority": "med_priority"},
		requestList: []string{"high_priority", "med_priority", "low_priority"},
		expected:    []string{"low_priority"},
		description: "filter allocated plugins and only retain unallocated ones",
	},
	{
		name:        "empty request list",
		ownerList:   nil,
		requestList: []string{},
		expected:    []string{},
		description: "empty request returns an empty list",
	},
	{
		name:        "all requests have been assigned",
		ownerList:   map[string]string{"p1": "p1", "p2": "p2"},
		requestList: []string{"p1", "p2"},
		expected:    []string{},
		description: "all requests are filtered",
	},
	{
		name:        "mixed priority and lexicographic order",
		ownerList:   nil,
		requestList: []string{"beta", "alpha", "high_priority", "gamma"},
		expected:    []string{"high_priority", "alpha", "beta", "gamma"},
		description: "prioritized items are ranked first, while others are arranged in lexicographic order",
	},
	{
		name:        "same priority",
		ownerList:   nil,
		requestList: []string{"pluginB", "pluginA"},
		expected:    []string{"pluginA", "pluginB"},
		description: "sort by dictionary order with the same priority",
	},
}

// TestNewStream test the function of NewStream
func TestNewStream(t *testing.T) {
	t.Run("create a new stream", func(t *testing.T) {
		name := "fault_stream"
		priorityConf := map[string]int{"pluginA": 1, "pluginB": 2}

		stream := NewStream(name, priorityConf)

		assert.Equal(t, name, stream.Name)
		assert.Equal(t, priorityConf, stream.PluginPriority)
		assert.Empty(t, stream.TokenOwner)
		assert.Empty(t, stream.OwnerMap)
	})

	t.Run("empty priority configuration", func(t *testing.T) {
		stream := NewStream("fault_stream", nil)

		assert.Nil(t, stream.PluginPriority)
		assert.NotNil(t, stream.OwnerMap)
	})
}

// TestGetName test the function of GetName
func TestGetName(t *testing.T) {
	stream := &Stream{Name: "test_stream"}
	assert.Equal(t, "test_stream", stream.GetName())
}

// TestBind test the function of Bind
func TestBind(t *testing.T) {
	t.Run("successfully bound idle stream", func(t *testing.T) {
		stream := &Stream{TokenOwner: "", OwnerMap: make(map[string]string, 0)}

		err := stream.Bind("plugin1")

		assert.NoError(t, err)
		assert.Equal(t, "plugin1", stream.TokenOwner)
		assert.Contains(t, stream.OwnerMap, "plugin1")
	})

	t.Run("bind an occupied stream", func(t *testing.T) {
		stream := &Stream{TokenOwner: "pluginX", OwnerMap: map[string]string{"pluginX": ""}}

		err := stream.Bind("pluginY")

		assert.Error(t, err)
		assert.EqualError(t, err, "tokenOwer is occupied")
		assert.Equal(t, "pluginX", stream.TokenOwner)
	})

	t.Run("after binding, the OwnerList includes the owners", func(t *testing.T) {
		stream := NewStream("stream1", nil)

		_ = stream.Bind("pluginA")
		_ = stream.Release("pluginA")
		_ = stream.Bind("pluginB")

		assert.Contains(t, stream.OwnerMap, "pluginA")
		assert.Contains(t, stream.OwnerMap, "pluginB")
	})
}

// TestRelease test the function of Release
func TestRelease(t *testing.T) {
	t.Run("release occupied streams", func(t *testing.T) {
		stream := &Stream{TokenOwner: "plugin1"}

		err := stream.Release("plugin1")

		assert.NoError(t, err)
		assert.Empty(t, stream.TokenOwner)
	})

	t.Run("release free streams", func(t *testing.T) {
		stream := &Stream{TokenOwner: ""}

		err := stream.Release("")

		assert.NotNil(t, err)
		assert.Empty(t, stream.TokenOwner)
	})
}

// TestReset test the function of reset
func TestReset(t *testing.T) {
	t.Run("reset stream status", func(t *testing.T) {
		stream := &Stream{
			TokenOwner: "pluginX",
			OwnerMap:   map[string]string{"pluginX": "pluginX", "pluginY": "pluginY"},
		}

		err := stream.Reset()

		assert.NoError(t, err)
		assert.Empty(t, stream.TokenOwner)
		assert.Empty(t, stream.OwnerMap)
	})

	t.Run("reset empty stream", func(t *testing.T) {
		stream := NewStream("empty_stream", nil)

		err := stream.Reset()

		assert.NoError(t, err)
		assert.Empty(t, stream.TokenOwner)
		assert.Empty(t, stream.OwnerMap)
	})
}

// TestGetTokenOwner test the function of GetTokenOwner
func TestGetTokenOwner(t *testing.T) {
	t.Run("get the current owner", func(t *testing.T) {
		stream := &Stream{TokenOwner: "current_owner"}
		assert.Equal(t, "current_owner", stream.GetTokenOwner())
	})

	t.Run("get empty owner", func(t *testing.T) {
		stream := &Stream{TokenOwner: ""}
		assert.Empty(t, stream.GetTokenOwner())
	})
}

// TestPrioritize test the function of Prioritize
func TestPrioritize(t *testing.T) {
	priorityConf := map[string]int{
		"high_priority": 1,
		"med_priority":  2,
		"low_priority":  3,
	}

	for _, tt := range priorityTests {
		t.Run(tt.description, func(t *testing.T) {
			stream := &Stream{
				OwnerMap:       tt.ownerList,
				PluginPriority: priorityConf,
			}

			result := stream.Prioritize(tt.requestList)
			assert.Equal(t, tt.expected, result, "the sorting result does not meet expectations")
		})
	}
}

// TestFilterAllocated test the function of filterAllocated
func TestFilterAllocated(t *testing.T) {
	tests := []struct {
		name        string
		ownerList   map[string]string
		requestList []string
		expected    []string
	}{
		{
			name:        "no owner list",
			ownerList:   nil,
			requestList: []string{"p1", "p2", "p3"},
			expected:    []string{"p1", "p2", "p3"},
		},
		{
			name:        "empty owner list",
			ownerList:   map[string]string{},
			requestList: []string{"p1", "p2"},
			expected:    []string{"p1", "p2"},
		},
		{
			name:        "filter partial requests",
			ownerList:   map[string]string{"p1": "p1", "p3": "p3"},
			requestList: []string{"p1", "p2", "p3", "p4"},
			expected:    []string{"p2", "p4"},
		},
		{
			name:        "filter all requests",
			ownerList:   map[string]string{"p1": "p1", "p2": "p2"},
			requestList: []string{"p1", "p2"},
			expected:    []string{},
		},
		{
			name:        "no filtering",
			ownerList:   map[string]string{"p5": "p5"},
			requestList: []string{"p1", "p2"},
			expected:    []string{"p1", "p2"},
		},
		{
			name:        "empty request list",
			ownerList:   map[string]string{"p1": "p1"},
			requestList: []string{},
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := &Stream{OwnerMap: tt.ownerList}
			result := stream.filterAllocated(tt.requestList)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPrioritizeStability test sorting stability
func TestPrioritizeStability(t *testing.T) {
	// create a plugin list without priority
	plugins := []string{"pluginC", "pluginA", "pluginB"}
	stream := NewStream("stable_stream", nil)

	// the results of multiple sorting should be consistent
	for i := 0; i < 5; i++ {
		result := stream.Prioritize(plugins)
		assert.Equal(t, []string{"pluginA", "pluginB", "pluginC"}, result)
	}
}

// TestEdgeCases test boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("empty plugin name", func(t *testing.T) {
		stream := NewStream("stream", nil)

		// bind an empty name
		err := stream.Bind("")
		assert.NoError(t, err)
		assert.Equal(t, "", stream.TokenOwner)
		assert.Contains(t, stream.OwnerMap, "")

		// include empty names in priority sorting
		result := stream.Prioritize([]string{"", "pluginA"})
		assert.Equal(t, []string{"pluginA"}, result) // the smallest empty string in lexicographic order
	})
}
