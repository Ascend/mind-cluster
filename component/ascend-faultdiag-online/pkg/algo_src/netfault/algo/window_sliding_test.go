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

// Package algo 网络连通性检测算法
package algo

import (
	"math"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetCurSlideWindow(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := NewNetDetect("testSuperPod1")

		convey.Convey("When curSlideWindows is nil", func() {
			nd.curSlideWindows = nil
			result := nd.getCurSlideWindow()
			convey.Convey("Then the result should be nil", func() {
				convey.So(result, convey.ShouldBeNil)
			})
		})

		convey.Convey("When curSlideWindows is empty", func() {
			nd.curSlideWindows = []map[string]any{}
			result := nd.getCurSlideWindow()
			convey.Convey("Then the result should be an empty slice", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When curSlideWindows contains data", func() {
			nd.curSlideWindows = []map[string]any{
				{"key1": "value1"},
				{"key2": "value2"},
			}
			result := nd.getCurSlideWindow()
			convey.Convey("Then the result should match the data", func() {
				convey.So(result, convey.ShouldResemble, []map[string]any{
					{"key1": "value1"},
					{"key2": "value2"},
				})
			})
		})
	})
}

func TestUpdateCurSlideWindows(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := NewNetDetect("testSuperPod1")
		nd.curPingPeriod = 1

		convey.Convey("When curSlideWindow is empty and newData is added", func() {
			nd.curSlideWindows = []map[string]any{}
			newData := []map[string]any{
				{"timestamp": int64(1000), "data": "value1"},
			}
			nd.updateCurSlideWindows(newData)
			convey.Convey("Then curSlideWindow should contain the new data", func() {
				convey.So(nd.curSlideWindows, convey.ShouldResemble, []map[string]any{
					{"timestamp": int64(1000), "data": "value1"},
				})
			})
		})

		convey.Convey("When curSlideWindow has data and newData is added", func() {
			nd.curSlideWindows = []map[string]any{
				{"timestamp": int64(1000), "data": "value1"},
			}
			newData := []map[string]any{
				{"timestamp": int64(1001), "data": "value2"},
			}
			nd.updateCurSlideWindows(newData)
			convey.Convey("Then curSlideWindow should contain merged and deduplicated data", func() {
				convey.So(nd.curSlideWindows, convey.ShouldResemble, []map[string]any{
					{"timestamp": int64(1000), "data": "value1"},
					{"timestamp": int64(1001), "data": "value2"},
				})
			})
		})

		convey.Convey("When curSlideWindow has old data that exceeds the time window", func() {
			nd.curSlideWindows = []map[string]any{
				{"timestamp": int64(1000), "data": "value1"},
				{"timestamp": int64(1001), "data": "value2"},
			}
			newData := []map[string]any{
				{"timestamp": int64(12000), "data": "value3"},
			}
			nd.updateCurSlideWindows(newData)
			convey.Convey("Then curSlideWindow should remove old data", func() {
				convey.So(nd.curSlideWindows, convey.ShouldResemble, []map[string]any{
					{"timestamp": int64(12000), "data": "value3"},
				})
			})
		})
	})
}

func TestConsumeQueueData(t *testing.T) {
	convey.Convey("Given a NetDetect instance with initial data", t, func() {
		nd := &NetDetect{
			curPingPeriod: 1,
			curSuperPodId: "testSuperPod",
			curSlideWindows: []map[string]any{
				{timestampConstant: int64(1000)},
				{timestampConstant: int64(1001)},
			},
			curConsumedQueue: []map[string]any{
				{timestampConstant: int64(1500)},
				{timestampConstant: int64(2500)},
				{timestampConstant: int64(3000)},
			},
			curSlideWindowsMaxTs: 1001,
		}

		convey.Convey("When consumeQueueData is called", func() {
			nd.consumeQueueData()

			convey.Convey("Then curSlideWindows should contain the correct items", func() {
				expectedSlideWindows := []map[string]any{
					{timestampConstant: int64(1000)},
					{timestampConstant: int64(1001)},
					{timestampConstant: int64(1500)},
				}
				convey.So(nd.curSlideWindows, convey.ShouldResemble, expectedSlideWindows)
			})

			convey.Convey("Then curConsumedQueue should contain the correct items", func() {
				expectedConsumedQueue := []map[string]any{
					{timestampConstant: int64(2500)},
					{timestampConstant: int64(3000)},
				}
				convey.So(nd.curConsumedQueue, convey.ShouldResemble, expectedConsumedQueue)
			})

			convey.Convey("Then curSlideWindowsMaxTs should be updated correctly", func() {
				expectedMaxTs := int64(1001 + 1*millisecondNum) // 预期最大时间戳
				convey.So(nd.curSlideWindowsMaxTs, convey.ShouldEqual, expectedMaxTs)
			})
		})
	})
}

func TestGetWindowData(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := NewNetDetect("testSuperPod1")
		nd.curPingPeriod = 1
		nd.curSlideWindowsMaxTs = 11000

		windows := []map[string]any{
			{"timestamp": int64(9000), "data": "value1"},
			{"timestamp": int64(9500), "data": "value2"},
			{"timestamp": int64(10000), "data": "value3"},
			{"timestamp": int64(11000), "data": "value4"},
		}

		convey.Convey("When startPeriod >= endPeriod", func() {
			startPeriod := 2 // 起始窗口数
			endPeriod := 1   // 结束窗口数
			result := nd.getWindowData(windows, startPeriod, endPeriod)
			convey.Convey("Then the result should be empty", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When data falls within the time window", func() {
			startPeriod := 1 // 起始窗口数
			endPeriod := 2   // 结束窗口数
			result := nd.getWindowData(windows, startPeriod, endPeriod)
			convey.Convey("Then the result should contain the correct data", func() {
				convey.So(result, convey.ShouldResemble, []map[string]any{
					{"timestamp": int64(9500), "data": "value2"},
					{"timestamp": int64(10000), "data": "value3"},
				})
			})
		})

		convey.Convey("When no data falls within the time window", func() {
			startPeriod := 3 // 起始窗口数
			endPeriod := 4   // 结束窗口数
			result := nd.getWindowData(windows, startPeriod, endPeriod)
			convey.Convey("Then the result should be empty", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})
	})
}

var specialWindows = []map[string]any{
	{
		pingTaskIDConstant: "task1",
		srcTypeConstant:    0,
		srcAddrConstant:    "192.168.1.1",
		dstTypeConstant:    0,
		dstAddrConstant:    "192.168.1.2",
		avgDelayConstant:   float64(10),
	},
	{
		pingTaskIDConstant: "task1",
		srcTypeConstant:    0,
		srcAddrConstant:    "192.168.1.3",
		dstTypeConstant:    0,
		dstAddrConstant:    "192.168.1.4",
		avgDelayConstant:   float64(20),
	},
	{
		pingTaskIDConstant: "task1",
		srcTypeConstant:    0,
		srcAddrConstant:    "192.168.1.1",
		dstTypeConstant:    0,
		dstAddrConstant:    "192.168.1.2",
		avgDelayConstant:   float64(15),
	},
	{
		pingTaskIDConstant: "task1",
		srcTypeConstant:    0,
		srcAddrConstant:    "192.168.1.1",
		dstTypeConstant:    0,
		dstAddrConstant:    "192.168.1.2",
		avgDelayConstant:   float64(25),
	},
	{
		pingTaskIDConstant: "task1",
		srcTypeConstant:    0,
		srcAddrConstant:    "192.168.1.5",
		dstTypeConstant:    0,
		dstAddrConstant:    "192.168.1.6",
		avgDelayConstant:   float64(30),
	},
}

// 寻找相同故障路径
func findSamePath(windows []map[string]any,
	path map[string]any, curPingPeriod int) []map[string]any {
	// 预分配结果切片（最多nd.curPingPeriod个相同路径）
	res := make([]map[string]any, curPingPeriod)

	for _, item := range windows {
		if isSamePath(item, path) {
			res = append(res, item)
		}
	}

	return res
}

func TestCalDynamicThresholds(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := NewNetDetect("testSuperPod1")

	convey.Convey("Given a set of windows and a path", t, func() {
		path := map[string]any{
			pingTaskIDConstant: "task1",
			srcTypeConstant:    0,
			srcAddrConstant:    "192.168.1.1",
			dstTypeConstant:    0,
			dstAddrConstant:    "192.168.1.2",
		}
		faultType := avgDelayConstant

		convey.Convey("When calculating dynamic thresholds", func() {
			samePaths := findSamePath(specialWindows, path, nd.curPingPeriod)
			result := calDynamicThresholds(samePaths, faultType)

			convey.Convey("Then the result should be correct", func() {
				// 计算均值和标准差
				value1 := 10.0  // 第一个值
				value2 := 15.0  // 第二个值
				value3 := 25.0  // 第三个值
				totalNum := 3.0 // 值的个数
				powNum := 2.0   // 平方数
				avg := (value1 + value2 + value3) / totalNum
				variance := (math.Pow(value1-avg, powNum) + math.Pow(value2-avg, powNum) + math.Pow(value3-avg, powNum)) /
					totalNum
				standardDeviation := math.Sqrt(variance)
				expected := avg + coefficientNum*math.Max(standardDeviation, 1.0)

				convey.So(result, convey.ShouldEqual, expected)
			})
		})
	})
}

var pathA = map[string]any{
	pingTaskIDConstant: "task1",
	srcTypeConstant:    "pod",
	srcAddrConstant:    "10.0.0.1",
	dstTypeConstant:    "service",
	dstAddrConstant:    "10.0.0.2",
}
var pathB = map[string]any{
	pingTaskIDConstant: "task1",
	srcTypeConstant:    "pod",
	srcAddrConstant:    "10.0.0.1",
	dstTypeConstant:    "service",
	dstAddrConstant:    "10.0.0.2",
}
var pathC = map[string]any{
	pingTaskIDConstant: "task2",
	srcTypeConstant:    "pod",
	srcAddrConstant:    "10.0.0.1",
	dstTypeConstant:    "service",
	dstAddrConstant:    "10.0.0.2",
}
var pathD = map[string]any{
	pingTaskIDConstant: "task1",
	srcTypeConstant:    "node",
	srcAddrConstant:    "10.0.0.1",
	dstTypeConstant:    "service",
	dstAddrConstant:    "10.0.0.2",
}

func TestIsSamePath(t *testing.T) {
	convey.Convey("Given two paths", t, func() {

		convey.Convey("When paths are identical", func() {
			result := isSamePath(pathA, pathB)

			convey.Convey("Then the result should be true", func() {
				convey.So(result, convey.ShouldBeTrue)
			})
		})

		convey.Convey("When paths have different pingTaskID", func() {
			result := isSamePath(pathA, pathC)

			convey.Convey("Then the result should be false", func() {
				convey.So(result, convey.ShouldBeFalse)
			})
		})

		convey.Convey("When paths have different srcType", func() {
			result := isSamePath(pathA, pathD)

			convey.Convey("Then the result should be false", func() {
				convey.So(result, convey.ShouldBeFalse)
			})
		})

		convey.Convey("When one path is missing a key", func() {
			incompletePath := map[string]any{
				pingTaskIDConstant: "task1",
				srcTypeConstant:    "pod",
				srcAddrConstant:    "10.0.0.1",
				dstTypeConstant:    "service",
			}
			result := isSamePath(pathA, incompletePath)

			convey.Convey("Then the result should be false", func() {
				convey.So(result, convey.ShouldBeFalse)
			})
		})
	})
}

func TestGetPeriodNum(t *testing.T) {
	convey.Convey("Given a NetDetect instance and windows data", t, func() {
		nd := NewNetDetect("testSuperPod1")
		nd.curPingPeriod = 10
		nd.curSlideWindowsMaxTs = 12000
		windows := []map[string]any{
			{timestampConstant: int64(11999)},
			{timestampConstant: int64(10999)},
			{timestampConstant: int64(9999)},
			{timestampConstant: "invalid"},
		}

		convey.Convey("When calculating the period number", func() {
			result := nd.getPeriodNum(windows)

			convey.Convey("Then the result should be correct", func() {
				// 计算最小时间戳
				minTimestamp := int64(9999)
				// 计算周期数
				periodCount := float64(nd.curSlideWindowsMaxTs-minTimestamp) / float64(millisecondNum*nd.curPingPeriod)
				expected := int(math.Ceil(periodCount))

				convey.So(result, convey.ShouldEqual, expected)
			})
		})

		convey.Convey("When windows is empty", func() {
			var emptyWindows []map[string]any
			result := nd.getPeriodNum(emptyWindows)

			convey.Convey("Then the result should be 0", func() {
				convey.So(result, convey.ShouldEqual, 0)
			})
		})
	})
}

func TestGetMaxTimeStamp(t *testing.T) {
	convey.Convey("Given a list of windows", t, func() {
		windows := []map[string]any{
			{timestampConstant: int64(1699999999000)},
			{timestampConstant: int64(1700000000000)},
			{timestampConstant: int64(1699999999500)},
			{timestampConstant: "invalid"},
		}

		convey.Convey("When calculating the maximum timestamp", func() {
			result := getMaxTimeStamp(windows)

			convey.Convey("Then the result should be the maximum timestamp", func() {
				convey.So(result, convey.ShouldEqual, int64(1700000000000))
			})
		})

		convey.Convey("When windows contain invalid timestamp", func() {
			invalidWindows := []map[string]any{
				{timestampConstant: "invalid"},
				{timestampConstant: "invalid"},
			}
			result := getMaxTimeStamp(invalidWindows)

			convey.Convey("Then the result should be math.MinInt64", func() {
				convey.So(result, convey.ShouldEqual, int64(math.MinInt64))
			})
		})

		convey.Convey("When windows is empty", func() {
			var emptyWindows []map[string]any
			result := getMaxTimeStamp(emptyWindows)

			convey.Convey("Then the result should be math.MinInt64", func() {
				convey.So(result, convey.ShouldEqual, int64(math.MinInt64))
			})
		})
	})
}
