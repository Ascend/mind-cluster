/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package devmanager for device driver manager
package devmanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
)

type unsupportedDeviceTypeCacheTestCase struct {
	name              string
	deviceTypeCode    int32
	setup             func(*unsupportedDeviceTypeCache)
	expectUnsupported bool
}

func buildUnsupportedDeviceTypeCacheTestCases() []unsupportedDeviceTypeCacheTestCase {
	return []unsupportedDeviceTypeCacheTestCase{
		{
			name:              "should return false when cache is nil",
			deviceTypeCode:    1,
			expectUnsupported: false,
		},
		{
			name:              "should return false when device type not in cache",
			deviceTypeCode:    1,
			setup:             func(c *unsupportedDeviceTypeCache) { c.unsupported = map[int32]bool{2: true} },
			expectUnsupported: false,
		},
		{
			name:              "should return true when device type in cache",
			deviceTypeCode:    1,
			setup:             func(c *unsupportedDeviceTypeCache) { c.unsupported = map[int32]bool{1: true} },
			expectUnsupported: true,
		},
	}
}

func TestUnsupportedDeviceTypeCache(t *testing.T) {
	for _, tt := range buildUnsupportedDeviceTypeCacheTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache := &unsupportedDeviceTypeCache{}
			if tt.setup != nil {
				tt.setup(cache)
			}
			result := cache.isUnsupported(tt.deviceTypeCode)
			convey.Convey("", t, func() {
				convey.So(result, convey.ShouldEqual, tt.expectUnsupported)
			})
		})
	}
}

func TestUnsupportedDeviceTypeCacheMark(t *testing.T) {
	t.Run("should mark device type as unsupported when cache is nil", func(t *testing.T) {
		cache := &unsupportedDeviceTypeCache{}
		cache.markAsUnsupported(1)
		convey.Convey("", t, func() {
			convey.So(cache.unsupported, convey.ShouldNotBeNil)
			convey.So(cache.unsupported[1], convey.ShouldBeTrue)
		})
	})

	t.Run("should mark device type as unsupported when cache exists", func(t *testing.T) {
		cache := &unsupportedDeviceTypeCache{
			unsupported: map[int32]bool{1: true},
		}
		cache.markAsUnsupported(2)
		convey.Convey("", t, func() {
			convey.So(cache.unsupported[2], convey.ShouldBeTrue)
		})
	})
}

type utilizationFuncCacheTestCase struct {
	name      string
	setup     func(*utilizationFuncCache)
	expectNil bool
}

func buildUtilizationFuncCacheTestCases() []utilizationFuncCacheTestCase {
	return []utilizationFuncCacheTestCase{
		{
			name:      "should return nil when cache is nil",
			setup:     func(c *utilizationFuncCache) {},
			expectNil: true,
		},
		{
			name: "should return cached function when cache is set",
			setup: func(c *utilizationFuncCache) {
				c.fn = func(int32) (common.DcmiMultiUtilizationInfo, error) {
					return common.DcmiMultiUtilizationInfo{}, nil
				}
			},
			expectNil: false,
		},
	}
}

func TestUtilizationFuncCacheGet(t *testing.T) {
	for _, tt := range buildUtilizationFuncCacheTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			cache := &utilizationFuncCache{}
			if tt.setup != nil {
				tt.setup(cache)
			}
			result := cache.get()
			if tt.expectNil {
				convey.Convey("", t, func() {
					convey.So(result, convey.ShouldBeNil)
				})
			} else {
				convey.Convey("", t, func() {
					convey.So(result, convey.ShouldNotBeNil)
				})
			}
		})
	}
}

func TestUtilizationFuncCacheSet(t *testing.T) {
	t.Run("should set function when function is not nil", func(t *testing.T) {
		cache := &utilizationFuncCache{}
		fn := func(int32) (common.DcmiMultiUtilizationInfo, error) {
			return common.DcmiMultiUtilizationInfo{}, nil
		}
		cache.set(fn)
		convey.Convey("", t, func() {
			convey.So(cache.fn, convey.ShouldNotBeNil)
		})
	})

	t.Run("should not set function when function is nil", func(t *testing.T) {
		cache := &utilizationFuncCache{}
		cache.set(nil)
		convey.Convey("", t, func() {
			convey.So(cache.fn, convey.ShouldBeNil)
		})
	})
}
