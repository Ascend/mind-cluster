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

// Package utils provides some DTs
package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToFloat64(t *testing.T) {
	type testCase struct {
		origin   any
		expected float64
	}

	var testCases = []testCase{
		{
			origin:   float32(123),
			expected: 123,
		},
		{
			origin:   float64(123),
			expected: 123,
		},
		{
			origin:   123,
			expected: 123,
		},
		{
			origin:   "123.3333",
			expected: 123.3333,
		},
	}
	var defaultValue = float64(123)
	for _, tc := range testCases {
		if got := ToFloat64(tc.origin, defaultValue); got != tc.expected {
			t.Errorf("ToFloat64(%v, %v) = %v, expected %v", tc.origin, defaultValue, got, tc.expected)
		}
	}

	// test default value
	varaibles := []any{"test", map[string]string{"test": "test"}, []string{"test"}}
	for _, v := range varaibles {
		if got := ToFloat64(v, defaultValue); got != defaultValue {
			t.Errorf("ToFloat64(%v, %v) = %v, expected %v", v, defaultValue, got, defaultValue)
		}
	}
}

func TestToString(t *testing.T) {

	type testCase struct {
		origin   any
		expected string
	}
	var testCases = []testCase{
		{
			origin:   "123",
			expected: "123",
		},
		{
			origin:   "",
			expected: "",
		},
		{
			origin:   123.00,
			expected: "",
		},
		{
			origin:   []string{"1", "2", "3"},
			expected: "",
		},
		{
			origin:   111,
			expected: "",
		},
	}

	for _, tc := range testCases {
		if got := ToString(tc.origin); got != tc.expected {
			t.Errorf("ToString(%v) = %v, expected %v", tc.origin, got, tc.expected)
		}
	}
}

func TestCopyInstance(t *testing.T) {
	type testCase struct {
		origin   any
		expected string
	}
	// src为nil
	var valueNil interface{}
	instanceNil, errNil := CopyInstance(valueNil)
	assert.Error(t, errNil)
	assert.Equal(t, errNil.Error(), "src cannot be nil")
	assert.Nil(t, instanceNil)

	// src不是指针对象
	testCases := testCase{
		origin:   "name",
		expected: "123",
	}
	instance, err := CopyInstance(testCases)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "copy instance src is not ptr")
	assert.Nil(t, instance)

	// src是指针对象，但为空
	var testCasesOne *testCase
	instanceOne, errOne := CopyInstance(testCasesOne)
	assert.Error(t, errOne)
	assert.Equal(t, errOne.Error(), "src ptr cannot be nil")
	assert.Nil(t, instanceOne)

	// 正常复制实例
	testCasesTwo := &testCases
	instanceTwo, notErr := CopyInstance(testCasesTwo)
	assert.Nil(t, notErr)
	assert.Equal(t, instanceTwo, testCasesTwo)
}
