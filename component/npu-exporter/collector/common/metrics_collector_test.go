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

// Package common for general collector
package common

import (
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestCopyMap test copyMap
func TestCopyMap(t *testing.T) {
	type testStruct struct {
		name string
		age  int
	}

	mockString := "mock"
	tests := []struct {
		name     string
		input    map[int32]testStruct
		validate func(*testing.T, interface{})
	}{
		{name: "NilInput", input: (map[int32]testStruct)(nil),
			validate: func(t *testing.T, got interface{}) {
				g, ok := got.(map[int32]testStruct)
				if !ok || g == nil || len(g) != 0 {
					t.Errorf("should return empty map for nil input")
				}
			}},
		{name: "EmptyMap", input: map[int32]testStruct{},
			validate: func(t *testing.T, got interface{}) {
				if len(got.(map[int32]testStruct)) != 0 {
					t.Errorf("expected empty map")
				}
			}},
		{name: "SingleElement", input: map[int32]testStruct{1: {name: mockString, age: 1}},
			validate: func(t *testing.T, got interface{}) {
				g, ok := got.(map[int32]testStruct)
				if !ok || g[1].name != mockString || g[1].age != 1 || len(g) != 1 {
					t.Errorf("element mismatch")
				}
			}},
		{name: "MultipleElements", input: map[int32]testStruct{1: {name: mockString, age: 1}, 2: {name: mockString, age: 1}},
			validate: func(t *testing.T, got interface{}) {
				expected := map[int32]testStruct{1: {name: mockString, age: 1}, 2: {name: mockString, age: 1}}
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("deepEqual failed")
				}
			}},
	}

	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			got := copyMap[testStruct](tt.input)
			tt.validate(t, got)
		})
	}
}
