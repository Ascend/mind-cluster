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

// Package slicetool provides some DTs
package slicetool

import (
	"fmt"
	"testing"
)

func toPtr[T int | float32 | float64](i T) *T {
	return &i
}

func TestIn(t *testing.T) {
	type testCase struct {
		value    int
		slice    []*int
		expected error
	}

	var testCases = []testCase{
		{
			value:    42,
			slice:    []*int{toPtr(0), toPtr(42), toPtr(3)},
			expected: nil,
		},
		{
			value:    42,
			slice:    []*int{toPtr(0), toPtr(2), toPtr(3)},
			expected: fmt.Errorf("the parameter 42 is not in the list: [1 2 3]"),
		},
		{
			value:    42,
			slice:    []*int{},
			expected: fmt.Errorf("the parameter 42 is not in the list: []"),
		},
	}

	for _, tc := range testCases {
		err := In(&tc.value, tc.slice)
		if (err == nil && tc.expected != nil) || (err != nil && tc.expected == nil) {
			t.Errorf("In(%v, %v) = %v, expected %v", tc.value, tc.slice, err, tc.expected)
		}
	}
}

func TestValueIn(t *testing.T) {
	type testCase struct {
		value    int
		slice    []int
		expected error
	}

	var testCases = []testCase{
		{
			value:    42,
			slice:    []int{1, 42, 3},
			expected: nil,
		},
		{
			value:    42,
			slice:    []int{1, 2, 3},
			expected: fmt.Errorf("the parameter 42 is not in the list: [1 2 3]"),
		},
		{
			value:    42,
			slice:    []int{},
			expected: fmt.Errorf("the parameter 42 is not in the list: []"),
		},
		{
			value:    42,
			slice:    []int{42, 42, 42},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		err := ValueIn(tc.value, tc.slice)
		if (err == nil && tc.expected != nil) || (err != nil && tc.expected == nil) {
			t.Errorf("ValueIn(%v, %v) = %v, expected %v", tc.value, tc.slice, err, tc.expected)
		}
	}
}

func TestMap(t *testing.T) {
	type testCase struct {
		origin   []*int
		mapper   func(*int) *int
		expected []*int
	}

	var testCases = []testCase{
		{
			origin: []*int{toPtr(1), toPtr(2), toPtr(3)},
			mapper: func(i *int) *int {
				val := *i * 2
				return &val
			},
			expected: []*int{toPtr(2), toPtr(4), toPtr(6)},
		},
		{
			origin: []*int{toPtr(1), nil, toPtr(3)},
			mapper: func(i *int) *int {
				if i == nil {
					return nil
				}
				val := *i * 2
				return &val
			},
			expected: []*int{toPtr(2), nil, toPtr(6)},
		},
		{
			origin: []*int{},
			mapper: func(i *int) *int {
				val := *i * 2
				return &val
			},
			expected: []*int{},
		},
		{
			origin: nil,
			mapper: func(i *int) *int {
				val := *i * 2
				return &val
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := Map(tc.origin, tc.mapper)
		if len(result) != len(tc.expected) {
			t.Errorf("Map(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
		for i, v := range result {
			if v == nil && tc.expected[i] == nil {
				continue
			}
			if v == nil || tc.expected[i] == nil || *v != *tc.expected[i] {
				t.Errorf("Map(%v) = %v, expected %v", tc.origin, result, tc.expected)
			}
		}
	}
}

func TestFilter(t *testing.T) {
	type testCase struct {
		origin    []*int
		predicate func(*int) bool
		expected  []*int
	}

	var testCases = []testCase{
		{
			origin: []*int{toPtr(1), toPtr(2), toPtr(3)},
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: []*int{toPtr(2), toPtr(3)},
		},
		{
			origin: []*int{toPtr(1), toPtr(3)},
			predicate: func(i *int) bool {
				return i != nil && *i > 1
			},
			expected: []*int{toPtr(3)},
		},
		{
			origin: []*int{},
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: []*int{},
		},
		{
			origin: nil,
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := Filter(tc.origin, tc.predicate)
		if len(result) != len(tc.expected) {
			t.Errorf("Filter(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
		for i, v := range result {
			if *v != *tc.expected[i] {
				t.Errorf("Filter(%v) = %v, expected %v", tc.origin, result, tc.expected)
			}
		}
	}
}

func TestMapToValue(t *testing.T) {
	type testCase struct {
		origin   []*int
		mapper   func(*int) int
		expected []int
	}

	var testCases = []testCase{
		{
			origin: []*int{toPtr(1), toPtr(2), toPtr(3)},
			mapper: func(i *int) int {
				return *i * 2
			},
			expected: []int{2, 4, 6},
		},
		{
			origin: []*int{toPtr(1), nil, toPtr(3)},
			mapper: func(i *int) int {
				if i == nil {
					return 0
				}
				return *i * 2
			},
			expected: []int{2, 0, 6},
		},
		{
			origin: []*int{},
			mapper: func(i *int) int {
				return *i * 2
			},
			expected: []int{},
		},
		{
			origin: nil,
			mapper: func(i *int) int {
				return *i * 2
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := MapToValue(tc.origin, tc.mapper)
		if len(result) != len(tc.expected) {
			t.Errorf("MapToValue(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
		for i, v := range result {
			if v != tc.expected[i] {
				t.Errorf("MapToValue(%v) = %v, expected %v", tc.origin, result, tc.expected)
			}
		}
	}
}

func TestAny(t *testing.T) {
	type testCase struct {
		origin    []*int
		predicate func(*int) bool
		expected  bool
	}

	var testCases = []testCase{
		{
			origin: []*int{toPtr(1), toPtr(2), toPtr(3)},
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: true,
		},
		{
			origin: []*int{toPtr(1), toPtr(3)},
			predicate: func(i *int) bool {
				return i != nil && *i > 1
			},
			expected: true,
		},
		{
			origin: []*int{},
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: false,
		},
		{
			origin: nil,
			predicate: func(i *int) bool {
				return *i > 1
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		result := Any(tc.origin, tc.predicate)
		if result != tc.expected {
			t.Errorf("Any(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
	}
}

func TestAll(t *testing.T) {
	type testCase struct {
		origin    []*int
		predicate func(*int) bool
		expected  bool
	}

	var testCases = []testCase{
		{
			origin: []*int{toPtr(2), toPtr(4), toPtr(6)},
			predicate: func(i *int) bool {
				return *i%2 == 0 // Check if all numbers are even
			},
			expected: true,
		},
		{
			origin: []*int{toPtr(2), toPtr(3), toPtr(6)},
			predicate: func(i *int) bool {
				return *i%2 == 0 // One odd number, should return false
			},
			expected: false,
		},
		{
			origin: []*int{toPtr(1), nil, toPtr(3)},
			predicate: func(i *int) bool {
				return i != nil && *i > 0 // Check if all items are non-nil and positive
			},
			expected: false, // Because of the nil value
		},
		{
			origin: []*int{},
			predicate: func(i *int) bool {
				return *i > 0 // Empty slice should return true (vacuously true)
			},
			expected: true,
		},
		{
			origin: nil,
			predicate: func(i *int) bool {
				return *i > 0 // Nil slice should return true (vacuously true)
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		result := All(tc.origin, tc.predicate)
		if result != tc.expected {
			t.Errorf("All(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
	}
}

func TestChain(t *testing.T) {
	type testCase struct {
		origin   [][]*int
		expected []*int
	}

	var testCases = []testCase{
		{
			origin: [][]*int{
				{toPtr(1), toPtr(2)},
				{toPtr(3), toPtr(4)},
			},
			expected: []*int{toPtr(1), toPtr(2), toPtr(3), toPtr(4)},
		},
		{
			origin: [][]*int{
				{toPtr(1), nil},
				{toPtr(3), toPtr(4)},
			},
			expected: []*int{toPtr(1), nil, toPtr(3), toPtr(4)},
		},
		{
			origin: [][]*int{
				{},
				{toPtr(3), toPtr(4)},
			},
			expected: []*int{toPtr(3), toPtr(4)},
		},
		{
			origin: [][]*int{
				{toPtr(1), toPtr(2)},
				{},
			},
			expected: []*int{toPtr(1), toPtr(2)},
		},
		{
			origin: [][]*int{
				{},
				{},
			},
			expected: []*int{},
		},
		{
			origin:   nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := Chain(tc.origin)
		if len(result) != len(tc.expected) {
			t.Errorf("Chain(%v) = %v, expected %v", tc.origin, result, tc.expected)
		}
		for i, v := range result {
			if v == nil && tc.expected[i] == nil {
				continue
			}
			if v == nil || tc.expected[i] == nil || *v != *tc.expected[i] {
				t.Errorf("Chain(%v) = %v, expected %v", tc.origin, result, tc.expected)
			}
		}
	}
}
