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

// Package utils includes some DT for the common utils.
package utils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToStringList(t *testing.T) {
	var testCases = []struct {
		Params []any
		Expect []string
	}{
		{
			[]any{"string", 123, 123.01, []int{1, 2, 3}, map[int]int{1: 2}, true, nil},
			[]string{"string", "123", "123.01", "[1 2 3]", "map[1:2]", "true", "<nil>"},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Expect, ToStringList(tc.Params))
	}
}

func TestUniqueSlice(t *testing.T) {
	var testCases = []struct {
		Params []string
		Expect []string
	}{
		{
			[]string{"1", "2", "3", "1", "2"},
			[]string{"1", "2", "3"},
		},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.Expect, UniqueSlice(tc.Params))
	}
}

func TestInSlice(t *testing.T) {
	testcasesSlice := []string{"1", "2", "3"}
	assert.True(t, InSlice(testcasesSlice, "1"))
	assert.False(t, InSlice(testcasesSlice, "5"))
}

func TestStrContains(t *testing.T) {
	// 测试nil切片
	if StrContains(nil, "test") {
		t.Error("Expected false, got true")
	}

	// 测试空切片
	if StrContains([]string{}, "test") {
		t.Error("Expected false, got true")
	}

	// 测试切片中没有包含txt的情况
	if StrContains([]string{"hello", "world"}, "test") {
		t.Error("Expected false, got true")
	}

	// 测试切片中包含txt的情况
	if !StrContains([]string{"hello", "world", "test"}, "test") {
		t.Error("Expected true, got false")
	}

	// 测试txt为空字符串的情况
	if !StrContains([]string{"hello", "world", "test"}, "") {
		t.Error("Expected true, got false")
	}

	// 测试txt在切片中的某个字符串中
	if !StrContains([]string{"hello", "world", "test"}, "est") {
		t.Error("Expected true, got false")
	}
}

func TestReadLinesFromOffset(t *testing.T) {
	// 创建临时文件
	file, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// 写入测试数据
	_, err = file.WriteString("line1\nline2\nline3\n")
	if err != nil {
		t.Fatal(err)
	}

	// 测试从文件开头读取
	lines, nextOffset, err := ReadLinesFromOffset(file.Name(), 0)
	if err != nil {
		t.Error(err)
	}
	const line3 = 3
	if len(lines) != line3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	const offset18 = 18
	const offset6 = 6
	const offset14 = 14
	if nextOffset != offset18 {
		t.Errorf("Expected nextOffset to be 14, got %d", nextOffset)
	}

	// 测试从文件中间读取
	lines, nextOffset, err = ReadLinesFromOffset(file.Name(), offset6)
	if err != nil {
		t.Error(err)
	}
	const line2 = 2
	if len(lines) != line2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
	if nextOffset != offset18 {
		t.Errorf("Expected nextOffset to be 14, got %d", nextOffset)
	}

	// 测试从文件末尾读取
	lines, nextOffset, err = ReadLinesFromOffset(file.Name(), offset14)
	if err != nil {
		t.Error(err)
	}
	const line1 = 1
	if len(lines) != line1 {
		t.Errorf("Expected 0 lines, got %d", len(lines))
	}
	if nextOffset != offset18 {
		t.Errorf("Expected nextOffset to be 14, got %d", nextOffset)
	}

	// 测试文件不存在的情况
	_, _, err = ReadLinesFromOffset("nonexistent", 0)
	if err == nil {
		t.Error("Expected error when file does not exist")
	}
}

func TestIsFileAble(t *testing.T) {
	// 创建临时文件用于测试
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // 测试结束后删除临时文件

	// 测试文件可读可写
	if !isFileAble(tmpfile.Name(), os.O_RDWR) {
		t.Errorf("Expected file to be readable and writable")
	}

	// 测试文件只读
	if !isFileAble(tmpfile.Name(), os.O_RDONLY) {
		t.Errorf("Expected file to be readable")
	}

	// 测试文件只写
	if !isFileAble(tmpfile.Name(), os.O_WRONLY) {
		t.Errorf("Expected file to be writable")
	}

	// 测试不存在的文件
	if isFileAble("nonexistentfile", os.O_RDWR) {
		t.Errorf("Expected non-existent file to be unopenable")
	}
}

func TestPoller(t *testing.T) {
	const twoSecond = 2
	t.Run("Should return nil when conditionFunc returns true", func(t *testing.T) {
		conditionFunc := func() (bool, error) {
			return true, nil
		}
		err := Poller(conditionFunc, time.Second, time.Second*twoSecond, nil)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("Should return error when conditionFunc returns error", func(t *testing.T) {
		conditionFunc := func() (bool, error) {
			return false, fmt.Errorf("error")
		}
		err := Poller(conditionFunc, time.Second, time.Second*twoSecond, nil)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Should return timeout error when timeout", func(t *testing.T) {
		conditionFunc := func() (bool, error) {
			return false, nil
		}
		err := Poller(conditionFunc, time.Second, time.Second*twoSecond, nil)
		if err == nil || err.Error() != "timeout: 2s" {
			t.Errorf("Expected timeout error, got %v", err)
		}
	})

	t.Run("Should return nil when stopChan is closed", func(t *testing.T) {
		conditionFunc := func() (bool, error) {
			return false, nil
		}
		stopChan := make(chan struct{})
		go func() {
			time.Sleep(time.Second)
			close(stopChan)
		}()
		err := Poller(conditionFunc, time.Second, time.Second*twoSecond, stopChan)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})
}

func TestSubtractAndDedupe(t *testing.T) {
	var testCases = []struct {
		Params1 []int
		Params2 []int
		Expect  []int
	}{
		{
			[]int{1, 2, 3, 4, 5, 5},
			[]int{3, 4, 6},
			[]int{1, 2, 5},
		},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.Expect, SubtractAndDedupe(tc.Params1, tc.Params2))
	}
}
