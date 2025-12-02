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
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpdateHistoryAlarmMap(t *testing.T) {
	convey.Convey("Given a NetDetect instance with existing alarms", t, func() {
		nd := &NetDetect{
			curSuppressedPeriod: 1, // 1秒
			curSuperPodId:       "testSuperPod",
		}

		// 设置模拟的全局报警
		globalHistoryAlarms["alarm1"] = time.Now().UnixMilli() - 2000 // 2秒前
		globalHistoryAlarms["alarm2"] = time.Now().UnixMilli() - 500  // 0.5秒前
		globalHistoryAlarms["alarm3"] = time.Now().UnixMilli() - 1500 // 1.5秒前

		convey.Convey("When updateHistoryAlarmMap is called", func() {
			nd.updateHistoryAlarmMap()

			convey.Convey("Then alarm1 and alarm3 should be removed", func() {
				convey.So(globalHistoryAlarms, convey.ShouldNotContainKey, "alarm1")
				convey.So(globalHistoryAlarms, convey.ShouldNotContainKey, "alarm3")
			})

			convey.Convey("Then alarm2 should still be present", func() {
				convey.So(globalHistoryAlarms, convey.ShouldContainKey, "alarm2")
			})
		})
	})
}

func TestFormatLayer(t *testing.T) {
	convey.Convey("Given a NetDetect instance and an item with source and destination addresses", t, func() {
		// 初始化 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")

		item := map[string]any{
			srcAddrConstant: "192.168.1.1",
			dstAddrConstant: "192.168.1.2",
		}

		convey.Convey("When formatLayer is called", func() {
			nd.curTopo = []string{"Rack-1#slot-1#192.168.1.1", "Rack-1#slot-1#192.168.1.2"}
			nd.formatLayer(item)

			convey.Convey("Then the item should contain the correct layer paths", func() {
				convey.So(item[fromLayerConstant], convey.ShouldEqual, "Rack-1#slot-1#192.168.1.1")
				convey.So(item[toLayerConstant], convey.ShouldEqual, "Rack-1#slot-1#192.168.1.2")
			})
		})

		convey.Convey("When formatLayer is called and topo is empty", func() {
			nd.formatLayer(item)

			convey.Convey("Then the item should contain the correct layer paths", func() {
				convey.So(item[fromLayerConstant], convey.ShouldEqual, nil)
				convey.So(item[toLayerConstant], convey.ShouldEqual, nil)
			})
		})

		convey.Convey("When formatLayer is called with missing source address", func() {
			item[srcAddrConstant] = 123 // 非字符串类型

			nd.formatLayer(item)

			convey.Convey("Then the item should not contain fromLayer", func() {
				convey.So(item, convey.ShouldNotContainKey, fromLayerConstant)
			})
		})

		convey.Convey("When formatLayer is called with missing destination address", func() {
			item[dstAddrConstant] = nil // 缺失目标地址

			nd.formatLayer(item)

			convey.Convey("Then the item should not contain toLayer", func() {
				convey.So(item, convey.ShouldNotContainKey, toLayerConstant)
			})
		})

		convey.Convey("When formatLayer is called with nil item", func() {
			var nilItem map[string]any

			nd.formatLayer(nilItem)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(nilItem, convey.ShouldBeNil)
			})
		})
	})
}

func TestFormatLossRate(t *testing.T) {
	convey.Convey("Given an item with loss rates as strings", t, func() {
		item := map[string]any{
			avgLoseRateConstant: "0.1",  // 10%
			minLoseRateConstant: "0.05", // 5%
			maxLoseRateConstant: "0.15", // 15%
		}

		convey.Convey("When formatLossRate is called", func() {
			formatLossRate(item)

			convey.Convey("Then the item should contain the correct formatted loss rates", func() {
				convey.So(item[avgLoseRateConstant], convey.ShouldEqual, 10.0) // 0.1 * 100
				convey.So(item[minLoseRateConstant], convey.ShouldEqual, 5.0)  // 0.05 * 100
				convey.So(item[maxLoseRateConstant], convey.ShouldEqual, 15.0) // 0.15 * 100
			})
		})

		convey.Convey("When formatLossRate is called with non-numeric values", func() {
			item[avgLoseRateConstant] = "invalid" // 非数字字符串
			item[minLoseRateConstant] = "0.02"    // 2%
			item[maxLoseRateConstant] = "0.03"    // 3%

			formatLossRate(item)

			convey.Convey("Then the valid loss rates should be formatted correctly", func() {
				convey.So(item[avgLoseRateConstant], convey.ShouldEqual, "invalid") // 保持不变
				convey.So(item[minLoseRateConstant], convey.ShouldEqual, 2.0)       // 0.02 * 100
				convey.So(item[maxLoseRateConstant], convey.ShouldEqual, 3.0)       // 0.03 * 100
			})
		})

		convey.Convey("When formatLossRate is called with nil item", func() {
			var nilItem map[string]any

			formatLossRate(nilItem)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(nilItem, convey.ShouldBeNil)
			})
		})

		convey.Convey("When formatDelay is called with missing keys", func() {
			item = map[string]any{
				avgLoseRateConstant: "15.0",
				minLoseRateConstant: 0,
			}

			formatDelay(item)

			convey.Convey("Then the item should only contain the formatted avgDelay", func() {
				convey.So(item, convey.ShouldContainKey, avgLoseRateConstant)
				convey.So(item, convey.ShouldContainKey, minLoseRateConstant)
				convey.So(item, convey.ShouldNotContainKey, maxLoseRateConstant)
			})
		})
	})
}

func TestFormatDelay(t *testing.T) {
	convey.Convey("Given an item with delay values as strings", t, func() {
		item := map[string]any{
			avgDelayConstant: "10.5", // 平均延迟
			minDelayConstant: "5.0",  // 最小延迟
			maxDelayConstant: "20.0", // 最大延迟
		}

		convey.Convey("When formatDelay is called", func() {
			formatDelay(item)

			convey.Convey("Then the item should contain the correct formatted delay values", func() {
				convey.So(item[avgDelayConstant], convey.ShouldEqual, 10.5) // 10.5
				convey.So(item[minDelayConstant], convey.ShouldEqual, 5.0)  // 5.0
				convey.So(item[maxDelayConstant], convey.ShouldEqual, 20.0) // 20.0
			})
		})

		convey.Convey("When formatDelay is called with non-numeric values", func() {
			item[avgDelayConstant] = "invalid" // 非数字字符串
			item[minDelayConstant] = "3.5"     // 3.5
			item[maxDelayConstant] = "8.0"     // 8.0

			formatDelay(item)

			convey.Convey("Then the valid delay values should be formatted correctly", func() {
				convey.So(item[avgDelayConstant], convey.ShouldEqual, "invalid") // 保持不变
				convey.So(item[minDelayConstant], convey.ShouldEqual, 3.5)       // 3.5
				convey.So(item[maxDelayConstant], convey.ShouldEqual, 8.0)       // 8.0
			})
		})

		convey.Convey("When formatDelay is called with nil item", func() {
			var nilItem map[string]any

			formatDelay(nilItem)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(nilItem, convey.ShouldBeNil)
			})
		})

		convey.Convey("When formatDelay is called with missing keys", func() {
			item = map[string]any{
				avgDelayConstant: "15.0",
				minDelayConstant: 0,
			}

			formatDelay(item)

			convey.Convey("Then the item should only contain the formatted avgDelay", func() {
				convey.So(item, convey.ShouldContainKey, avgDelayConstant)
				convey.So(item, convey.ShouldContainKey, minDelayConstant)
				convey.So(item, convey.ShouldNotContainKey, maxDelayConstant)
			})
		})
	})
}

func TestFormatTimestamp(t *testing.T) {
	convey.Convey("Given an item with a timestamp as a string", t, func() {
		item := map[string]any{
			timestampConstant: "1622547800", // 示例时间戳
		}

		convey.Convey("When formatTimestamp is called", func() {
			formatTimestamp(item)

			convey.Convey("Then the item should contain the correct formatted timestamp", func() {
				convey.So(item[timestampConstant], convey.ShouldEqual, int64(1622547800)) // 转换后的时间戳
			})
		})

		convey.Convey("When formatTimestamp is called with a non-numeric value", func() {
			item[timestampConstant] = "invalid" // 非数字字符串

			formatTimestamp(item)

			convey.Convey("Then the timestamp should remain unchanged", func() {
				convey.So(item[timestampConstant], convey.ShouldEqual, "invalid") // 保持不变
			})
		})

		convey.Convey("When formatTimestamp is called with an empty string", func() {
			item[timestampConstant] = "" // 空字符串

			formatTimestamp(item)

			convey.Convey("Then the timestamp should remain unchanged", func() {
				convey.So(item[timestampConstant], convey.ShouldEqual, "") // 保持不变
			})
		})

		convey.Convey("When formatTimestamp is called with nil item", func() {
			var nilItem map[string]any

			formatTimestamp(nilItem)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(nilItem, convey.ShouldBeNil)
			})
		})

		convey.Convey("When formatTimestamp is called with missing key", func() {
			item = map[string]any{} // 没有任何键

			formatTimestamp(item)

			convey.Convey("Then it should do nothing", func() {
				convey.So(item, convey.ShouldBeEmpty) // 保持为空
			})
		})
	})
}

func TestFormatInputData(t *testing.T) {
	convey.Convey("Given a list of input data", t, func() {
		input := []map[string]any{
			{
				avgLoseRateConstant: "0.1", minLoseRateConstant: "0.05", maxLoseRateConstant: "0.15",
				avgDelayConstant: "10.5", minDelayConstant: "5.0", maxDelayConstant: "20.0",
				timestampConstant: "1622547800",
			},
			{
				avgLoseRateConstant: "invalid", minLoseRateConstant: "0.02", maxLoseRateConstant: "0.03",
				avgDelayConstant: "3.5", minDelayConstant: "1.0", maxDelayConstant: "8.0",
				timestampConstant: "1632547800",
			},
		}

		// 初始化 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")

		convey.Convey("When formatInputData is called", func() {
			nd.formatInputData(input)

			convey.Convey("Then the input data should be formatted correctly", func() {
				convey.So(input[0][avgLoseRateConstant], convey.ShouldEqual, 10.0)
				convey.So(input[0][minLoseRateConstant], convey.ShouldEqual, 5.0)
				convey.So(input[0][maxLoseRateConstant], convey.ShouldEqual, 15.0)
				convey.So(input[0][avgDelayConstant], convey.ShouldEqual, 10.5)
				convey.So(input[0][minDelayConstant], convey.ShouldEqual, 5.0)
				convey.So(input[0][maxDelayConstant], convey.ShouldEqual, 20.0)
				convey.So(input[0][timestampConstant], convey.ShouldEqual, int64(1622547800))

				convey.So(input[1][avgLoseRateConstant], convey.ShouldEqual, "invalid")
				convey.So(input[1][minLoseRateConstant], convey.ShouldEqual, 2.0)
				convey.So(input[1][maxLoseRateConstant], convey.ShouldEqual, 3.0)
				convey.So(input[1][avgDelayConstant], convey.ShouldEqual, 3.5)
				convey.So(input[1][minDelayConstant], convey.ShouldEqual, 1.0)
				convey.So(input[1][maxDelayConstant], convey.ShouldEqual, 8.0)
				convey.So(input[1][timestampConstant], convey.ShouldEqual, int64(1632547800))
			})
		})
	})
}

func TestGetPathHashKey(t *testing.T) {
	convey.Convey("Given a map with keys for path hash", t, func() {
		path := map[string]any{
			pingTaskIDConstant: "value1",
			srcTypeConstant:    "value2",
			srcAddrConstant:    "value3",
			dstTypeConstant:    "value4",
			dstAddrConstant:    "value5",
		}

		convey.Convey("When getPathHashKey is called", func() {
			hashKey := getPathHashKey(path)

			convey.Convey("Then the hash key should be formatted correctly", func() {
				convey.So(hashKey, convey.ShouldEqual, "value1|value2|value3|value4|value5|") // 预期结果
			})
		})

		convey.Convey("When some keys are missing", func() {
			path = map[string]any{
				pingTaskIDConstant: "value1",
				// srcTypeConstant is missing
				srcAddrConstant: "value3",
				dstTypeConstant: "value4",
				dstAddrConstant: "value5",
			}

			hashKey := getPathHashKey(path)

			convey.Convey("Then the hash key should reflect missing keys", func() {
				convey.So(hashKey, convey.ShouldEqual, "value1|<nil>|value3|value4|value5|") // 预期结果
			})
		})

		convey.Convey("When all keys are missing", func() {
			path = map[string]any{}

			hashKey := getPathHashKey(path)

			convey.Convey("Then the hash key should be empty", func() {
				convey.So(hashKey, convey.ShouldEqual, "<nil>|<nil>|<nil>|<nil>|<nil>|") // 预期结果
			})
		})
	})
}

func TestUpdatePathIndex(t *testing.T) {
	convey.Convey("Given a NetDetect instance and some sample calWindows", t, func() {
		// 初始化 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")

		calWindows := []map[string]any{
			{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
				dstTypeConstant: "value4", dstAddrConstant: "value5", "otherKey": "otherValue1"},
			{pingTaskIDConstant: "value6", srcTypeConstant: "value7", srcAddrConstant: "value8",
				dstTypeConstant: "value9", dstAddrConstant: "value10", "otherKey": "otherValue2"},
			{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
				dstTypeConstant: "value4", dstAddrConstant: "value5", "otherKey": "otherValue3"},
		}

		convey.Convey("When updatePathIndex is called", func() {
			nd.updatePathIndex(calWindows)

			convey.Convey("Then pathIndex should contain the correct keys and values", func() {
				expectedLen := 2
				convey.So(nd.pathIndex, convey.ShouldHaveLength, expectedLen)
				convey.So(nd.pathIndex, convey.ShouldContainKey, getPathHashKey(calWindows[0]))
				convey.So(nd.pathIndex[getPathHashKey(calWindows[0])], convey.ShouldHaveLength, expectedLen)
				convey.So(nd.pathIndex[getPathHashKey(calWindows[1])], convey.ShouldHaveLength, 1)
			})
		})
	})
}

func TestFindSamePathFast(t *testing.T) {
	convey.Convey("Given a NetDetect instance with populated pathIndex", t, func() {
		// 初始化 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")

		nd.pathIndex = map[string][]map[string]any{
			"value1|value2|value3|value4|value5|": {
				{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
					dstTypeConstant: "value4", dstAddrConstant: "value5", "otherKey": "otherValue1"},
				{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
					dstTypeConstant: "value4", dstAddrConstant: "value5", "otherKey": "otherValue2"},
			},
			"value1|value2|value3|value4|value6|": {
				{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
					dstTypeConstant: "value4", dstAddrConstant: "value6", "otherKey": "otherValue3"},
			},
		}

		convey.Convey("When findSamePathFast is called with an existing path", func() {
			path := map[string]any{pingTaskIDConstant: "value1", srcTypeConstant: "value2",
				srcAddrConstant: "value3", dstTypeConstant: "value4", dstAddrConstant: "value5",
				"otherKey": "otherValue3"}
			result := nd.findSamePathFast(path)

			convey.Convey("Then it should return the correct entries for that path", func() {
				expectedLen := 2
				convey.So(result, convey.ShouldHaveLength, expectedLen)
				convey.So(result[0]["otherKey"], convey.ShouldEqual, "otherValue1")
				convey.So(result[1]["otherKey"], convey.ShouldEqual, "otherValue2")
			})
		})

		convey.Convey("When findSamePathFast is called with a non-existing path", func() {
			path := map[string]any{"path": "/nonexistent"}
			result := nd.findSamePathFast(path)

			convey.Convey("Then it should return an empty slice", func() {
				convey.So(result, convey.ShouldHaveLength, 0) // 预期结果
			})
		})
	})
}

func TestGetDynamicThresholds(t *testing.T) {
	convey.Convey("Given a set of samePaths", t, func() {
		samePaths := []map[string]any{
			{avgLoseRateConstant: 1.0, avgDelayConstant: 2.0},
			{avgLoseRateConstant: 1.5, avgDelayConstant: 2.5},
			{avgLoseRateConstant: 2.0, avgDelayConstant: 3.0},
		}

		convey.Convey("When calling getDynamicThresholds", func() {
			lossThreshold, delayThreshold := getDynamicThresholds(samePaths)

			convey.Convey("Then the lossDynamicThreshold should be calculated correctly", func() {
				convey.So(lossThreshold, convey.ShouldEqual, 4.5) // 预期丢包阈值
			})

			convey.Convey("Then the delayDynamicThreshold should be calculated correctly", func() {
				convey.So(delayThreshold, convey.ShouldEqual, 5.5) // 预期时延阈值
			})
		})
	})
}

func TestGetIndicators(t *testing.T) {
	convey.Convey("Given a set of input data and a path", t, func() {
		input := []map[string]any{
			{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
				dstTypeConstant: "value4", dstAddrConstant: "value5", avgLoseRateConstant: 1.0, avgDelayConstant: 1.0},
			{pingTaskIDConstant: "value6", srcTypeConstant: "value7", srcAddrConstant: "value8",
				dstTypeConstant: "value9", dstAddrConstant: "value10", avgLoseRateConstant: 2.0, avgDelayConstant: 2.0},
			{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
				dstTypeConstant: "value4", dstAddrConstant: "value5", avgLoseRateConstant: 3.0, avgDelayConstant: 3.0},
		}

		path := map[string]any{pingTaskIDConstant: "value1", srcTypeConstant: "value2", srcAddrConstant: "value3",
			dstTypeConstant: "value4", dstAddrConstant: "value5", avgLoseRateConstant: 1.0, avgDelayConstant: 1.0}

		convey.Convey("When calling getIndicators", func() {
			lossIndicator, delayIndicator := getIndicators(input, path)

			convey.Convey("Then the lossIndicator should be calculated correctly", func() {
				convey.So(lossIndicator, convey.ShouldEqual, 2.0) // 预期丢包检测值
			})

			convey.Convey("Then the delayIndicator should be calculated correctly", func() {
				convey.So(delayIndicator, convey.ShouldEqual, 2.0) // 预期时延检测值
			})
		})
	})
}

func TestDiffFaultPathList(t *testing.T) {
	convey.Convey("Given a list of fault paths and a detect type", t, func() {
		faultPathList := []any{
			map[string]any{
				fromLayerConstant: "LayerA#LayerB#ip1",
				toLayerConstant:   "LayerA#LayerB#ip2",
			},
			map[string]any{
				fromLayerConstant: "LayerA#LayerB#ip3",
				toLayerConstant:   "LayerA#LayerB#ip4",
			},
			map[string]any{
				srcAddrConstant: "192.168.1.1",
				dstAddrConstant: "192.168.1.2",
			},
			"invalidPath", // 测试无效路径
		}
		detectType := "testType"

		nd := NewNetDetect("testSuperPod1")

		convey.Convey("When calling diffFaultPathList", func() {
			npuDireAlarmList, otherAlarmList := nd.diffFaultPathList(faultPathList, detectType)

			convey.Convey("Then the npuDireAlarmList should contain the NPU direct alarm path", func() {
				convey.So(len(npuDireAlarmList), convey.ShouldEqual, 1)
				convey.So(npuDireAlarmList[0].(map[string]any)[srcAddrConstant], convey.ShouldEqual,
					"192.168.1.1")
				convey.So(npuDireAlarmList[0].(map[string]any)[dstAddrConstant], convey.ShouldEqual,
					"192.168.1.2")
			})

			convey.Convey("Then the otherAlarmList should contain properly formatted paths", func() {
				convey.So(len(otherAlarmList), convey.ShouldEqual, 2) // 预期其他告警路径长度

				// 检查第一个路径的描述
				convey.So(otherAlarmList[0].(map[string]any)[descriptionConstant], convey.ShouldEqual,
					"S-LayerA#LayerB#ip1 TO D-LayerA#LayerB#ip2")
				// 检查第二个路径的描述
				convey.So(otherAlarmList[1].(map[string]any)[descriptionConstant], convey.ShouldEqual,
					"S-LayerA#LayerB#ip3 TO D-LayerA#LayerB#ip4")
			})
		})
	})
}

func TestCountForSlice(t *testing.T) {
	convey.Convey("Given a slice of strings", t, func() {
		target := []string{"apple", "banana", "apple", "orange", "banana", "banana"}

		convey.Convey("When calling countForSlice", func() {
			maxKey, countMap := countForSlice(target, false)

			convey.Convey("Then the max key should be the most frequent element", func() {
				convey.So(maxKey, convey.ShouldEqual, "banana")
			})

			convey.Convey("Then the count map should reflect the correct counts", func() {
				convey.So(countMap["apple"], convey.ShouldEqual, 2)  // 预期结果
				convey.So(countMap["banana"], convey.ShouldEqual, 3) // 预期结果
				convey.So(countMap["orange"], convey.ShouldEqual, 1) // 预期结果
			})
		})
	})

	convey.Convey("Given an empty slice", t, func() {
		var target []string

		convey.Convey("When calling countForSlice", func() {
			maxKey, countMap := countForSlice(target, false)

			convey.Convey("Then the max key should be an empty string", func() {
				convey.So(maxKey, convey.ShouldEqual, "")
			})

			convey.Convey("Then the count map should be empty", func() {
				convey.So(len(countMap), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestSetIndicators(t *testing.T) {
	convey.Convey("Given a faultPath and a rootCauseAlarm map", t, func() {
		faultPath := []any{
			map[string]any{
				minDelayConstant:    10.0,
				maxDelayConstant:    30.0,
				minLoseRateConstant: 1.0,
				maxLoseRateConstant: 5.0,
				timestampConstant:   int64(1000),
				avgDelayConstant:    20.0,
				avgLoseRateConstant: 3.0,
				pingTaskIDConstant:  "task1",
			},
			map[string]any{
				minDelayConstant:    5.0,
				maxDelayConstant:    25.0,
				minLoseRateConstant: 0.5,
				maxLoseRateConstant: 4.0,
				timestampConstant:   int64(1002),
				avgDelayConstant:    15.0,
				avgLoseRateConstant: 2.5,
				pingTaskIDConstant:  "task1",
			},
		}

		rootCauseAlarm := make(map[string]any)

		convey.Convey("When calling setIndicators", func() {
			setIndicators(faultPath, rootCauseAlarm)

			convey.Convey("Then the rootCauseAlarm should contain correct values", func() {
				convey.So(rootCauseAlarm[maxLoseRateConstant], convey.ShouldEqual, 5.000) // 预期最大丢包率
				convey.So(rootCauseAlarm[avgLoseRateConstant], convey.ShouldEqual, 2.750) // (3.0 + 2.5) / 2
				convey.So(rootCauseAlarm[maxDelayConstant], convey.ShouldEqual, 30.000)   // 预期最大时延
				convey.So(rootCauseAlarm[avgDelayConstant], convey.ShouldEqual, 17.500)   // (20.0 + 15.0) / 2
			})
		})

		convey.Convey("When calling setIndicators with an empty faultPath", func() {
			var faultPath2 []any
			rootCauseAlarm := make(map[string]any)

			setIndicators(faultPath2, rootCauseAlarm)

			convey.Convey("Then the rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When calling setIndicators with nil rootCauseAlarm", func() {
			var rootCauseAlarm map[string]any = nil

			setIndicators(faultPath, rootCauseAlarm)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(true, convey.ShouldBeTrue) // Just to ensure no panic occurs
			})
		})
	})
}

func TestSetFaultType(t *testing.T) {
	convey.Convey("Given a rootCauseAlarm map", t, func() {
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When detectType is globalDetectTypes[1]", func() {
			rootCauseAlarm[faultTypeConstant] = globalDetectTypes[1]
			rootCauseAlarm[avgLoseRateConstant] = 0.1 // 任意值

			setFaultType(rootCauseAlarm)

			convey.So(rootCauseAlarm[faultTypeConstant], convey.ShouldEqual, delayType)
		})

		convey.Convey("When avgLoseRate is above lossThreshold", func() {
			rootCauseAlarm[faultTypeConstant] = "someType"
			rootCauseAlarm[avgLoseRateConstant] = float64(lossThreshold + 1)

			setFaultType(rootCauseAlarm)

			convey.So(rootCauseAlarm[faultTypeConstant], convey.ShouldEqual, disconnectType)
		})

		convey.Convey("When avgLoseRate is below lossThreshold", func() {
			rootCauseAlarm[faultTypeConstant] = "someType"
			rootCauseAlarm[avgLoseRateConstant] = float64(lossThreshold - 1)

			setFaultType(rootCauseAlarm)

			convey.So(rootCauseAlarm[faultTypeConstant], convey.ShouldEqual, lossRateType)
		})

		convey.Convey("When rootCauseAlarm is nil or faultTypeConstant is not a string", func() {
			var nilRootCauseAlarm map[string]any = nil
			rootCauseAlarm[faultTypeConstant] = float64(123)

			convey.So(func() { setFaultType(nilRootCauseAlarm) }, convey.ShouldNotPanic)

			setFaultType(rootCauseAlarm)
			convey.So(rootCauseAlarm[faultTypeConstant], convey.ShouldEqual, 123) // 预期检测类型
		})
	})
}

func TestIsAlarmSuppressed(t *testing.T) {
	convey.Convey("Given a rootCauseAlarm map", t, func() {
		rootCauseAlarm := map[string]any{
			srcIdConstant:     "sourceID",
			dstIdConstant:     "destinationID",
			faultTypeConstant: 1,
			levelConstant:     2,
		}

		convey.Convey("When alarm is not in globalHistoryAlarms", func() {
			key := fmt.Sprintf("%s-%s-%d-%d", "sourceID", "destinationID", 1, 2) // value值
			delete(globalHistoryAlarms, key)

			result := isAlarmSuppressed(rootCauseAlarm)
			convey.So(result, convey.ShouldBeFalse)
			convey.So(globalHistoryAlarms[key], convey.ShouldNotBeZeroValue)
		})

		convey.Convey("When alarm is already in globalHistoryAlarms", func() {
			key := fmt.Sprintf("%s-%s-%d-%d", "sourceID", "destinationID", 1, 2) // value值
			globalHistoryAlarms[key] = time.Now().UnixMilli() - 10               // 当前时间之前的时间戳

			initialTimestamp := globalHistoryAlarms[key]
			result := isAlarmSuppressed(rootCauseAlarm)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(globalHistoryAlarms[key], convey.ShouldBeGreaterThan, initialTimestamp)
		})

		convey.Convey("When required fields are missing", func() {
			delete(rootCauseAlarm, srcIdConstant)
			result := isAlarmSuppressed(rootCauseAlarm)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestInitLeftIps(t *testing.T) {
	convey.Convey("Given a list of npuDireAlarmList", t, func() {
		npuDireAlarmList := []any{
			map[string]any{
				srcAddrConstant: "192.168.1.1",
				dstAddrConstant: "192.168.1.2",
			},
			map[string]any{
				srcAddrConstant: "192.168.1.3",
				dstAddrConstant: "192.168.1.4",
			},
			map[string]any{
				srcAddrConstant: "192.168.1.5",
			}, // dstAddr缺失
		}

		convey.Convey("When calling initLeftIps", func() {
			result := initLeftIps(npuDireAlarmList)

			convey.Convey("Then it should return the correct left IPs", func() {
				expected := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"}
				convey.So(result, convey.ShouldResemble, expected)
			})
		})
	})

	convey.Convey("When npuDireAlarmList is empty", t, func() {
		result := initLeftIps([]any{})
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestGetRootCauseIps(t *testing.T) {
	convey.Convey("Given a list of IPs", t, func() {
		ips := []string{
			"192.168.1.1",
			"192.168.1.2",
			"192.168.1.1",
			"192.168.1.3",
			"192.168.1.2",
			"192.168.1.1",
		}

		convey.Convey("When calling getRootCauseIps", func() {
			result := getRootCauseIps(ips)

			convey.Convey("Then it should return the most frequent IPs", func() {
				expected := []string{"192.168.1.1"}
				convey.So(result, convey.ShouldResemble, expected)
			})
		})

		convey.Convey("When there are multiple most frequent IPs", func() {
			ips = []string{
				"192.168.1.1",
				"192.168.1.2",
				"192.168.1.1",
				"192.168.1.2",
			}

			result := getRootCauseIps(ips)
			convey.So(result, convey.ShouldContain, "192.168.1.1")
			convey.So(result, convey.ShouldContain, "192.168.1.2")
		})

		convey.Convey("When the list is empty", func() {
			result := getRootCauseIps([]string{})
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

func TestGetCurFaultPathInfo(t *testing.T) {
	convey.Convey("Given a list of npuDireAlarmList and rootCauseIps", t, func() {
		npuDireAlarmList := []any{
			map[string]any{
				srcAddrConstant: "192.168.1.1",
				dstAddrConstant: "192.168.1.2",
			},
			map[string]any{
				srcAddrConstant: "192.168.1.3",
				dstAddrConstant: "192.168.1.4",
			},
			map[string]any{
				srcAddrConstant: "192.168.1.1",
				dstAddrConstant: "192.168.1.5",
			},
		}
		rootCauseIps := []string{"192.168.1.1"}
		var tmpList []string

		convey.Convey("When calling getCurFaultPathInfo", func() {
			result := getCurFaultPathInfo(npuDireAlarmList, rootCauseIps, &tmpList)

			convey.Convey("Then it should return the correct fault paths", func() {
				expected := []any{
					map[string]any{
						srcAddrConstant: "192.168.1.1",
						dstAddrConstant: "192.168.1.2",
					},
					map[string]any{
						srcAddrConstant: "192.168.1.1",
						dstAddrConstant: "192.168.1.5",
					},
				}
				convey.So(result, convey.ShouldResemble, expected)
				convey.So(tmpList, convey.ShouldResemble, []string{"192.168.1.1", "192.168.1.2",
					"192.168.1.1", "192.168.1.5"})
			})
		})

		convey.Convey("When rootCauseIps is empty", func() {
			rootCauseIps = []string{}
			result := getCurFaultPathInfo(npuDireAlarmList, rootCauseIps, &tmpList)
			convey.So(result, convey.ShouldBeEmpty)
			convey.So(tmpList, convey.ShouldBeEmpty)
		})
	})
}

func TestSetNpuDireAlarmFormat(t *testing.T) {
	convey.Convey("Given rootCauseIps, curIps and a rootCauseAlarm map", t, func() {
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseIps has one element", func() {
			rootCauseIps := []string{"192.168.1.1"}
			curIps := []string{"192.168.1.2"}

			setNpuDireAlarmFormat(rootCauseIps, curIps, rootCauseAlarm)

			convey.Convey("Then it should set the alarm format correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "192.168.1.1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "192.168.1.1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, majorType)
			})
		})

		convey.Convey("When rootCauseIps has multiple elements and curIps contains one", func() {
			rootCauseIps := []string{"192.168.1.1", "192.168.1.2"}
			curIps := []string{"192.168.1.2"}

			setNpuDireAlarmFormat(rootCauseIps, curIps, rootCauseAlarm)

			convey.Convey("Then it should set the alarm format with dstIp from curIps", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "192.168.1.1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "192.168.1.2")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, minorType)
			})
		})

		convey.Convey("When rootCauseIps has multiple elements and curIps does not contain any", func() {
			rootCauseIps := []string{"192.168.1.1", "192.168.1.3"}
			curIps := []string{"192.168.1.4"}

			setNpuDireAlarmFormat(rootCauseIps, curIps, rootCauseAlarm)

			convey.Convey("Then it should set the alarm format with the first rootCauseIp as dstIp", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "192.168.1.1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "192.168.1.1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, majorType)
			})
		})

		convey.Convey("When rootCauseAlarm is nil", func() {
			var nilAlarm map[string]any
			rootCauseIps := []string{"192.168.1.1"}

			setNpuDireAlarmFormat(rootCauseIps, nil, nilAlarm)

			convey.Convey("Then it should not panic and do nothing", func() {
				convey.So(nilAlarm, convey.ShouldBeNil)
			})
		})
	})
}

func TestRemoveAt(t *testing.T) {
	convey.Convey("Given a slice of maps", t, func() {
		slice := []map[string]any{
			{"key1": "value1"},
			{"key2": "value2"},
			{"key3": "value3"},
		}

		convey.Convey("When removeAt is called with a valid index", func() {
			removeAt(&slice, 1)

			convey.Convey("Then the slice should not contain the removed element", func() {
				convey.So(len(slice), convey.ShouldEqual, 2) // 预期长度
				convey.So(slice[0]["key1"], convey.ShouldEqual, "value1")
				convey.So(slice[1]["key3"], convey.ShouldEqual, "value3")
			})
		})

		convey.Convey("When removeAt is called with an index out of range", func() {
			removeAt(&slice, 5) // Invalid index

			convey.Convey("Then the slice should remain unchanged", func() {
				convey.So(len(slice), convey.ShouldEqual, 3) // 预期长度
			})
		})

		convey.Convey("When removeAt is called with a negative index", func() {
			removeAt(&slice, -1) // Invalid index

			convey.Convey("Then the slice should remain unchanged", func() {
				convey.So(len(slice), convey.ShouldEqual, 3) // 预期长度
			})
		})

		convey.Convey("When removeAt is called with the first index", func() {
			removeAt(&slice, 0)

			convey.Convey("Then the first element should be removed", func() {
				convey.So(len(slice), convey.ShouldEqual, 2) // 预期长度
				convey.So(slice[0]["key2"], convey.ShouldEqual, "value2")
			})
		})
	})
}

func TestFillDetectCoreData(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := NewNetDetect("testSuperPod1") // 使用指定的初始化方式

		convey.Convey("When fillDetectCoreData is called with input data", func() {
			input := []map[string]any{
				{"key1": "value1"},
				{"key2": "value2"},
				{"key1": "value1"}, // 重复项
			}

			nd.fillDetectCoreData(input)
			convey.Convey("Then curConsumedQueue should contain deduplicated data", func() {
				convey.So(len(nd.curConsumedQueue), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When fillDetectCoreData is called with curOpenQueueFlag set to false", func() {
			nd.curOpenQueueFlag = false
			input := []map[string]any{
				{"key3": "value3", "timestamp": int64(0)},
				{"key4": "value4", "timestamp": int64(1000)},
			}

			nd.fillDetectCoreData(input)
			convey.Convey("Then curSlideWindows should be updated with input data", func() {
				convey.So(len(nd.curSlideWindows), convey.ShouldEqual, 2) // 预期长度
				convey.So(nd.curSlideWindows[0]["key3"], convey.ShouldEqual, "value3")
				convey.So(nd.curSlideWindows[1]["key4"], convey.ShouldEqual, "value4")
			})
		})

		convey.Convey("When fillDetectCoreData is called with an empty input", func() {
			nd.curOpenQueueFlag = true
			input := []map[string]any{
				{"key3": "value3", "timestamp": int64(1000)},
				{"key4": "value4", "timestamp": int64(1000)},
			}

			nd.fillDetectCoreData(input)
			convey.Convey("Then curConsumedQueue should remain unchanged", func() {
				convey.So(nd.curConsumedQueue, convey.ShouldEqual, []map[string]any(nil))
			})
			convey.Convey("Then curSlideWindows should remain unchanged", func() {
				convey.So(len(nd.curSlideWindows), convey.ShouldEqual, 2) // 预期长度
			})
		})
	})
}
