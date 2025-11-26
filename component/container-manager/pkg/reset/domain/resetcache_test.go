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

// Package domain unit tests for reset cache
package domain

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	if err != nil {
		panic(err)
	}
	m.Run()
}

// TestNpuInResetCache_New tests the GetNpuInResetCache function
func TestNpuInResetCache_New(t *testing.T) {
	convey.Convey("Test GetNpuInResetCache", t, func() {
		cache := GetNpuInResetCache()
		convey.So(cache, convey.ShouldNotBeNil)
		convey.So(cache.npuInResetCache, convey.ShouldNotBeNil)
		convey.So(len(cache.npuInResetCache), convey.ShouldEqual, 0) // Verify initial state is empty
	})
}

func getTestResetCache() *NpuInResetCache {
	cache := GetNpuInResetCache()
	testNpus := []int32{1, 2, 3}
	cache.SetNpuInReset(testNpus...)
	return cache
}

// TestNpuInResetCache_Set tests the SetNpuInReset method
func TestNpuInResetCache_Set(t *testing.T) {
	convey.Convey("Test SetNpuInReset", t, func() {
		cache := getTestResetCache()
		convey.So(cache.npuInResetCache, convey.ShouldContainKey, int32(1))
		convey.So(cache.npuInResetCache, convey.ShouldContainKey, int32(2))
		convey.So(cache.npuInResetCache, convey.ShouldContainKey, int32(3))
	})
}

// TestNpuInResetCache_Get tests the ISNpuInReset method
func TestNpuInResetCache_Get(t *testing.T) {
	convey.Convey("Test ISNpuInReset", t, func() {
		cache := getTestResetCache()
		convey.So(cache.IsNpuInReset(1), convey.ShouldBeTrue)
	})
}

// TestNpuInResetCache_Clear tests the ClearNpuInReset method
func TestNpuInResetCache_Clear(t *testing.T) {
	convey.Convey("Test ClearNpuInReset", t, func() {
		cache := getTestResetCache()
		testToClear := []int32{1, 3}
		cache.ClearNpuInReset(testToClear...)
		convey.So(cache.npuInResetCache, convey.ShouldNotContainKey, int32(1))
		convey.So(cache.npuInResetCache, convey.ShouldContainKey, int32(2))
		convey.So(cache.npuInResetCache, convey.ShouldNotContainKey, int32(3))
		convey.So(len(cache.npuInResetCache), convey.ShouldEqual, 1)
	})
}

// TestNpuInResetCache_DeepCopy tests the DeepCopy method
func TestNpuInResetCache_DeepCopy(t *testing.T) {
	convey.Convey("Test DeepCopy", t, func() {
		cache := getTestResetCache()
		copyCache := cache.DeepCopy()
		convey.So(copyCache, convey.ShouldResemble, map[int32]struct{}{1: {}, 2: {}, 3: {}}) // Verify copy content
		// Modifying copy should not affect original cache
		delete(copyCache, 1)
		convey.So(cache.DeepCopy(), convey.ShouldContainKey, int32(1)) // Original cache should remain unchanged
	})
}

// TestFailedResetCountCache_New tests the NewFailedResetCountCache function
func TestFailedResetCountCache_New(t *testing.T) {
	convey.Convey("Test NewFailedResetCountCache", t, func() {
		cache := NewFailedResetCountCache()
		convey.So(cache, convey.ShouldNotBeNil)
		convey.So(cache.failedResetCountCache, convey.ShouldNotBeNil)
		convey.So(len(cache.failedResetCountCache), convey.ShouldEqual, 0)
	})
}

// TestFailedResetCountCache_Set tests the SetFailedResetCount method
func TestFailedResetCountCache_Set(t *testing.T) {
	convey.Convey("Test SetFailedResetCount", t, func() {
		cache := NewFailedResetCountCache()
		testId := int32(1)
		testCount := 5
		cache.SetFailedResetCount(testId, testCount)
		convey.So(cache.failedResetCountCache[testId], convey.ShouldEqual, testCount)
	})
}

// TestFailedResetCountCache_Get tests the GetFailedResetCount method
func TestFailedResetCountCache_Get(t *testing.T) {
	convey.Convey("Test GetFailedResetCount", t, func() {
		convey.Convey("Should return count for existing phyId", func() {
			cache := NewFailedResetCountCache()
			testId := int32(1)
			testCount := 3
			cache.SetFailedResetCount(testId, testCount)
			count := cache.GetFailedResetCount(testId)
			convey.So(count, convey.ShouldEqual, testCount)
		})

		convey.Convey("Should return 0 for non-existing phyId", func() {
			cache := NewFailedResetCountCache()
			count := cache.GetFailedResetCount(999) // Non-existing phyId
			convey.So(count, convey.ShouldEqual, 0)
		})
	})
}

// TestFailedResetCountCache_GetAll tests the GetAllFailedResetCountNpuId method
func TestFailedResetCountCache_GetAll(t *testing.T) {
	convey.Convey("Test GetAllFailedResetCountNpuId", t, func() {
		convey.Convey("Should return all phyIds", func() {
			cache := NewFailedResetCountCache()
			testDate := map[int32]int{1: 1, 2: 2}
			for id, count := range testDate {
				cache.SetFailedResetCount(id, count)
			}
			ids := cache.GetAllFailedResetCountNpuId()
			convey.So(ids, convey.ShouldContain, int32(1))
			convey.So(ids, convey.ShouldContain, int32(2))
			convey.So(len(ids), convey.ShouldEqual, len(testDate))
		})

		convey.Convey("Should return empty slice for empty cache", func() {
			cache := NewFailedResetCountCache()
			ids := cache.GetAllFailedResetCountNpuId()
			convey.So(ids, convey.ShouldBeEmpty)
		})
	})
}

// TestFailedResetCountCache_Clear tests the ClearFailedResetCount method
func TestFailedResetCountCache_Clear(t *testing.T) {
	convey.Convey("Test ClearFailedResetCount", t, func() {
		cache := NewFailedResetCountCache()
		testId := int32(1)
		testCount := 5
		cache.SetFailedResetCount(testId, testCount)
		cache.ClearFailedResetCount(1)
		convey.So(cache.failedResetCountCache, convey.ShouldNotContainKey, int32(1))
		convey.So(len(cache.failedResetCountCache), convey.ShouldEqual, 0)
	})
}
