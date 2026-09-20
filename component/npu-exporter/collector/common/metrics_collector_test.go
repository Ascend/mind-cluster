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
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

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

func TestPreCollect(t *testing.T) {
	tests := []struct {
		name       string
		deviceType string
		expected   bool
	}{
		{name: "TestPreCollect_" + api.Ascend910,
			deviceType: api.Ascend910,
			expected:   true,
		},
		{name: "TestPreCollect_" + api.Ascend310,
			deviceType: api.Ascend310,
			expected:   false,
		},
	}
	convey.Convey("TestPreCollect", t, func() {
		n := mockNewNpuCollector()
		adapter := MetricsCollectorAdapter{
			Is910Series:  false,
			ContainerMap: nil,
			Chips:        nil,
		}
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				patches := gomonkey.NewPatches()
				defer patches.Reset()
				patches.ApplyMethodReturn(n.Dmgr, "GetDevType", tt.deviceType)
				adapter.PreCollect(n, nil)
				convey.So(adapter.Is910Series, convey.ShouldEqual, tt.expected)
			})
		}
	})
}

type cacheCase struct {
	name           string
	cacheKey       string
	preHandle      func()
	localEntries   map[int32]string
	invalidEntries map[int32]int
	expected       int
}

func buildTestsForUpdateCache() []cacheCase {
	const (
		testKey10         = int32(10)
		testKey20         = int32(20)
		testKey2          = int32(2)
		testInvalidVal1   = 123
		testInvalidVal2   = 456
		testExpectedTwo   = 2
		testExpectedThree = 3
	)
	return []cacheCase{
		{name: "should save info to cache when no old cache",
			cacheKey:     "mockKey1",
			localEntries: map[int32]string{0: "mockValue"},
			expected:     1,
		},
		{name: "should update old cache and skip log when noNeedToPrintUpdateLog set",
			cacheKey: "mockKey2",
			preHandle: func() {
				noNeedToPrintUpdateLog["mockKey2"] = true
			},
			localEntries: map[int32]string{testKey10: "mockValue"},
			expected:     testExpectedTwo,
		},
		{name: "should reset when old cache is in incorrect type",
			cacheKey:     "mockKey3",
			localEntries: map[int32]string{testKey20: "mockValue"},
			expected:     1,
		},
		{name: "should merge multiple entries from localCache",
			cacheKey:     "mockKey4",
			localEntries: map[int32]string{0: "v0", 1: "v1", testKey2: "v2"},
			expected:     testExpectedThree,
		},
		{name: "should skip invalid type entries in localCache",
			cacheKey:       "mockKey5",
			localEntries:   map[int32]string{0: "v0"},
			invalidEntries: map[int32]int{1: testInvalidVal1, testKey2: testInvalidVal2},
			expected:       1,
		},
		{name: "should handle empty localCache with no old cache",
			cacheKey: "mockKey6",
			expected: 0,
		},
		{name: "should overwrite old cache entry with new data",
			cacheKey:     "mockKey7",
			localEntries: map[int32]string{0: "newValue"},
			expected:     1,
		},
	}
}

func TestUpdateCache(t *testing.T) {
	const oldKey = int32(0)

	n := mockNewNpuCollector()
	n.cache.Set("mockKey2", map[int32]string{oldKey: "oldValue"}, n.cacheTime)
	n.cache.Set("mockKey3", map[int32]int{oldKey: 0}, n.cacheTime)
	n.cache.Set("mockKey7", map[int32]string{oldKey: "oldValue"}, n.cacheTime)

	tests := buildTestsForUpdateCache()

	convey.Convey("TestUpdateCache", t, func() {

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				localCache := sync.Map{}
				for k, v := range tt.localEntries {
					localCache.Store(k, v)
				}
				for k, v := range tt.invalidEntries {
					localCache.Store(k, v)
				}
				if tt.preHandle != nil {
					tt.preHandle()
				}
				UpdateCache[string](n, tt.cacheKey, &localCache)

				data, err := n.cache.Get(tt.cacheKey)
				convey.So(err, convey.ShouldBeNil)
				map2, ok := data.(map[int32]string)
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(len(map2), convey.ShouldEqual, tt.expected)
			})
		}

	})
}

func TestUpdateCacheSetError(t *testing.T) {
	const key = int32(0)
	n := mockNewNpuCollector()

	convey.Convey("should not panic when cache Set fails", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethodReturn(n.cache, "Set", errors.New("mock set error"))
		localCache := sync.Map{}
		localCache.Store(key, "mockValue")
		UpdateCache[string](n, "mockKeySetErr", &localCache)
	})
}

func TestGetInfoFromCache(t *testing.T) {
	const key = int32(0)
	tests := []struct {
		name     string
		cacheKey string
		expected int
	}{
		{name: "TestGetInfoFromCache_no info in cache",
			cacheKey: "mockKey1",
			expected: 0,
		},
		{name: "TestGetInfoFromCache_correct",
			cacheKey: "mockKey2",
			expected: 1,
		},
		{name: "TestGetInfoFromCache_info in cache is in incorrect type",
			cacheKey: "mockKey3",
			expected: 0,
		},
	}
	n := mockNewNpuCollector()
	// data init
	n.cache.Set("mockKey2", map[int32]string{key: "mockValue"}, n.cacheTime)
	n.cache.Set("mockKey3", map[int32]int{key: 0}, n.cacheTime)
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			cache := GetInfoFromCache[string](n, tt.cacheKey)
			convey.So(len(cache), convey.ShouldEqual, tt.expected)
		})
	}
}

func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		args     interface{}
		expected string
	}{
		{name: "TestGetCacheKey_ptr",
			args:     &MetricsCollectorAdapter{},
			expected: "MetricsCollectorAdapter",
		},
		{name: "TestGetCacheKey_int",
			args:     0,
			expected: "",
		},
		{name: "TestGetCacheKey_struct",
			args:     MetricsCollectorAdapter{},
			expected: "",
		},
	}

	convey.Convey("TestGetCacheKey", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				convey.So(GetCacheKey(tt.args), convey.ShouldEqual, tt.expected)
			})
		}
	})
}
