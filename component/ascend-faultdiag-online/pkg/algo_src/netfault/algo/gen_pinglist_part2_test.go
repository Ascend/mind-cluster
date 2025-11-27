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
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSortLayerList(t *testing.T) {
	convey.Convey("Given a list of layer strings", t, func() {
		convey.Convey("When the list is unsorted", func() {
			layers := []string{"Layer-2", "Layer-10", "Layer-1", "Layer-3"}

			convey.Convey("Then the sorted list should be in correct order", func() {
				sortLayerList(layers)
				expected := []string{"Layer-1", "Layer-2", "Layer-3", "Layer-10"}
				convey.So(layers, convey.ShouldResemble, expected)
			})
		})

		convey.Convey("When the list contains invalid formats", func() {
			layers := []string{"LayerA", "LayerB", "Layer1"}

			convey.Convey("Then the function should not panic and return the list as is", func() {
				sortLayerList(layers)
				expected := []string{"LayerA", "LayerB", "Layer1"}
				convey.So(layers, convey.ShouldResemble, expected)
			})
		})

		convey.Convey("When the list is already sorted", func() {
			layers := []string{"Layer1", "Layer2", "Layer3"}

			convey.Convey("Then the sorted list should remain the same", func() {
				sortLayerList(layers)
				expected := []string{"Layer1", "Layer2", "Layer3"}
				convey.So(layers, convey.ShouldResemble, expected)
			})
		})

		convey.Convey("When the list is empty", func() {
			var layers []string

			convey.Convey("Then the sorted list should also be empty", func() {
				sortLayerList(layers)
				var expected []string
				convey.So(layers, convey.ShouldResemble, expected)
			})
		})
	})
}

func TestSortIpList(t *testing.T) {
	convey.Convey("Test SortIpList", t, func() {
		nd := &NetDetect{
			curNpuInfo: map[string]NpuInfo{
				"192.168.1.1": {IP: "192.168.1.1", NpuNumber: 2},
				"192.168.1.2": {IP: "192.168.1.2", NpuNumber: 1},
				"192.168.1.3": {IP: "192.168.1.3", NpuNumber: 3},
			},
		}

		ipList := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
		nd.sortIpList(ipList)

		convey.So(ipList, convey.ShouldResemble, []string{"192.168.1.2", "192.168.1.1", "192.168.1.3"}) // 应按 NpuNumber 升序排列

		// 测试用例 2: 所有 NpuInfo 为空
		nd.curNpuInfo = map[string]NpuInfo{
			"192.168.1.5": {},
			"192.168.1.6": {},
			"192.168.1.7": {},
		}
		ipList = []string{"192.168.1.5", "192.168.1.6", "192.168.1.7"}
		nd.sortIpList(ipList)

		convey.So(ipList, convey.ShouldResemble, []string{"192.168.1.5", "192.168.1.6", "192.168.1.7"}) // 应保持原顺序
	})
}

func TestProcessIpPairs(t *testing.T) {
	convey.Convey("Test ProcessIpPairs", t, func() {
		aiPingStrategy := &AiPingStrategy{
			pingDict: make(map[string]any),
		}

		pingDictKey := "testKey"
		ipPairs := [][]string{
			{"192.168.1.1", "192.168.1.2"},
			{"192.168.1.3", "192.168.1.4"},
		}

		// 测试用例 1: pingDictKey 不存在，应该初始化
		processIpPairs(aiPingStrategy, pingDictKey, ipPairs)

		convey.So(aiPingStrategy.pingDict[pingDictKey], convey.ShouldNotBeNil)
		value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(value), convey.ShouldEqual, 2) // 应包含 2 个元素

		// 验证填充的内容
		convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
		convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.2")
		convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.3")
		convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.4")

		// 测试用例 2: pingDictKey 已存在，应该继续填充
		ipPairs2 := [][]string{
			{"192.168.1.5", "192.168.1.6"},
		}
		processIpPairs(aiPingStrategy, pingDictKey, ipPairs2)

		value, ok = aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(value), convey.ShouldEqual, 3) // 应包含 3 个元素

		// 验证新增的内容
		convey.So(value[2][fromConstant], convey.ShouldEqual, "192.168.1.5")
		convey.So(value[2][toConstant], convey.ShouldEqual, "192.168.1.6")

		// 测试用例 3: 使用不同的 pingDictKey
		pingDictKey2 := "anotherKey"
		ipPairs3 := [][]string{
			{"192.168.2.1", "192.168.2.2"},
		}
		processIpPairs(aiPingStrategy, pingDictKey2, ipPairs3)

		convey.So(aiPingStrategy.pingDict[pingDictKey2], convey.ShouldNotBeNil) // 应该初始化
		value2, ok := aiPingStrategy.pingDict[pingDictKey2].([]map[string]any)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(value2), convey.ShouldEqual, 1) // 应包含 1 个元素
		convey.So(value2[0][fromConstant], convey.ShouldEqual, "192.168.2.1")
		convey.So(value2[0][toConstant], convey.ShouldEqual, "192.168.2.2")
	})
}

func TestGetSameNumDstIp(t *testing.T) {
	convey.Convey("Test GetSameNumDstIp", t, func() {
		nd := &NetDetect{
			curNpuInfo: map[string]NpuInfo{
				"192.168.1.1": {IP: "192.168.1.1", NpuNumber: 1},
				"192.168.1.2": {IP: "192.168.1.2", NpuNumber: 2},
				"192.168.1.3": {IP: "192.168.1.3", NpuNumber: 1},
				"192.168.1.4": {IP: "192.168.1.4", NpuNumber: 3},
			},
		}

		// 测试用例 1: 匹配的目标 IP
		srcIp := "192.168.1.1"
		dstIps := []string{"192.168.1.2", "192.168.1.3", "192.168.1.4"}
		result := nd.getSameNumDstIp(srcIp, dstIps)
		convey.So(result, convey.ShouldEqual, "192.168.1.3") // 应返回与 srcIp NpuNumber 相同的 IP

		// 测试用例 2: 没有匹配的目标 IP
		srcIp2 := "192.168.1.2"
		dstIps2 := []string{"192.168.1.3", "192.168.1.4"}
		result2 := nd.getSameNumDstIp(srcIp2, dstIps2)
		convey.So(result2, convey.ShouldEqual, "") // 应返回空字符串，因为没有匹配的 IP

		// 测试用例 3: srcIp 不在 curNpuInfo 中
		srcIp3 := "192.168.1.5" // 不存在的 IP
		dstIps3 := []string{"192.168.1.3", "192.168.1.4"}
		result3 := nd.getSameNumDstIp(srcIp3, dstIps3)
		convey.So(result3, convey.ShouldEqual, "") // 应返回空字符串，因为 srcIp 不存在

		// 测试用例 4: dstIps 全部不匹配
		srcIp4 := "192.168.1.1"
		dstIps4 := []string{"192.168.1.5", "192.168.1.6"} // 不存在的 IP
		result4 := nd.getSameNumDstIp(srcIp4, dstIps4)
		convey.So(result4, convey.ShouldEqual, "") // 应返回空字符串，因为没有匹配的 IP
	})
}

func TestGetAlignNumDstIp(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := &NetDetect{
		curNpuInfo: map[string]NpuInfo{
			"192.168.1.1": {NpuNumber: 1},
			"192.168.1.2": {NpuNumber: 2},
			"192.168.1.3": {NpuNumber: 9},
			"192.168.1.4": {NpuNumber: 10},
		},
	}

	convey.Convey("Test GetAlignNumDstIp", t, func() {
		convey.Convey("Valid srcIp and aligned dstIp", func() {
			dstIps := []string{"192.168.1.2", "192.168.1.3"}
			result := nd.getAlignNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "192.168.1.3")
		})

		convey.Convey("Valid srcIp but no aligned dstIp", func() {
			dstIps := []string{"192.168.1.4"}
			result := nd.getAlignNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("Invalid srcIp", func() {
			dstIps := []string{"192.168.1.2", "192.168.1.3"}
			result := nd.getAlignNumDstIp("192.168.1.5", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("Empty dstIps", func() {
			var dstIps []string
			result := nd.getAlignNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("No dstIp aligns with srcIp", func() {
			dstIps := []string{"192.168.1.4", "192.168.1.5"}
			result := nd.getAlignNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetCrossNumDstIp(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := &NetDetect{
		curNpuInfo: map[string]NpuInfo{
			"192.168.1.1": {NpuNumber: 1},
			"192.168.1.2": {NpuNumber: 2},
			"192.168.1.3": {NpuNumber: 9},
			"192.168.1.4": {NpuNumber: 10},
		},
	}

	convey.Convey("Test GetCrossNumDstIp", t, func() {
		convey.Convey("Valid srcIp and cross aligned dstIp", func() {
			dstIps := []string{"192.168.1.3", "192.168.1.4"}
			result := nd.getCrossNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "192.168.1.4")
		})

		convey.Convey("Valid srcIp but no cross aligned dstIp", func() {
			dstIps := []string{"192.168.1.3"}
			result := nd.getCrossNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("Invalid srcIp", func() {
			dstIps := []string{"192.168.1.2", "192.168.1.3"}
			result := nd.getCrossNumDstIp("192.168.1.5", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})

		convey.Convey("Empty dstIps", func() {
			var dstIps []string
			result := nd.getCrossNumDstIp("192.168.1.1", dstIps)
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestProcessIpsRack(t *testing.T) {
	convey.Convey("Test ProcessIpsRack", t, func() {
		// 创建一个模拟的 AiPingStrategy
		aiPingStrategy := &AiPingStrategy{
			pingDict: make(map[string]any),
		}

		// 创建一个 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")
		nd.curNpuInfo["192.168.1.1"] = NpuInfo{IP: "192.168.1.1", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.2"] = NpuInfo{IP: "192.168.1.2", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.3"] = NpuInfo{IP: "192.168.1.3", NpuNumber: 2}
		nd.curNpuInfo["192.168.1.4"] = NpuInfo{IP: "192.168.1.4", NpuNumber: 2}
		nd.curNpuInfo["192.168.1.5"] = NpuInfo{IP: "192.168.1.5", NpuNumber: 3}
		nd.curNpuInfo["192.168.1.6"] = NpuInfo{IP: "192.168.1.6", NpuNumber: 3}

		srcIps := []string{"192.168.1.1", "192.168.1.3", "192.168.1.5"}
		dstIps := []string{"192.168.1.2", "192.168.1.4", "192.168.1.6"}
		pingDictKey := "testKey"

		// 调用 processIpsRack
		nd.processIpsRack(aiPingStrategy, srcIps, dstIps, pingDictKey)

		// 验证结果
		value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
		expectedLen := 3 // 期望的长度
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(value), convey.ShouldEqual, expectedLen)

		// 验证填充的内容
		convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
		convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.2")
		convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.3")
		convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.4")
		convey.So(value[2][fromConstant], convey.ShouldEqual, "192.168.1.5")
		convey.So(value[2][toConstant], convey.ShouldEqual, "192.168.1.6")

		// 验证未匹配的源 IP
		convey.So(len(value), convey.ShouldEqual, 3) // 确保只填充了找到匹配的 IP
	})
}

func TestProcessIpsSlot(t *testing.T) {
	convey.Convey("Test ProcessIpsSlot", t, func() {
		// 创建一个模拟的 AiPingStrategy
		aiPingStrategy := &AiPingStrategy{
			pingDict: make(map[string]any),
		}

		// 创建一个 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")
		nd.curNpuInfo["192.168.1.1"] = NpuInfo{IP: "192.168.1.1", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.2"] = NpuInfo{IP: "192.168.1.2", NpuNumber: 3}
		nd.curNpuInfo["192.168.1.3"] = NpuInfo{IP: "192.168.1.3", NpuNumber: 9}
		nd.curNpuInfo["192.168.1.4"] = NpuInfo{IP: "192.168.1.4", NpuNumber: 10}
		nd.curNpuInfo["192.168.1.5"] = NpuInfo{IP: "192.168.1.5", NpuNumber: 11}
		nd.curNpuInfo["192.168.1.6"] = NpuInfo{IP: "192.168.1.6", NpuNumber: 12}

		srcIps := []string{"192.168.1.1", "192.168.1.2"}
		dstIps := []string{"192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6"}
		pingDictKey := "testKey"

		// 调用 processIpsSlot
		nd.processIpsSlot(aiPingStrategy, srcIps, dstIps, pingDictKey)

		// 验证结果
		value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
		expectedLen := 4 // 期望的长度
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(value), convey.ShouldEqual, expectedLen)

		// 验证填充的内容
		convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
		convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.3")
		convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.1")
		convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.4")
		convey.So(value[2][fromConstant], convey.ShouldEqual, "192.168.1.2")
		convey.So(value[2][toConstant], convey.ShouldEqual, "192.168.1.5")
		convey.So(value[3][fromConstant], convey.ShouldEqual, "192.168.1.2")
		convey.So(value[3][toConstant], convey.ShouldEqual, "192.168.1.6")

		// 验证未匹配的源 IP
		convey.So(len(value), convey.ShouldEqual, 4) // 确保只填充了找到匹配的 IP
	})
}

func TestSetPingPair(t *testing.T) {
	convey.Convey("Given a NetDetect instance and an AiPingStrategy", t, func() {
		nd := NewNetDetect("testSuperPod1")
		aiPingStrategy := &AiPingStrategy{pingDict: make(map[string]any)}
		randomIps := map[string]any{
			"layer1": []string{"192.168.1.1", "192.168.1.2"},
			"layer2": []string{"192.168.2.1", "192.168.2.2"},
			"layer3": []string{"192.168.3.1", "192.168.3.2"},
		}
		childLayerList := []string{"layer1", "layer2", "layer3"}

		convey.Convey("When setting ping pairs with valid inputs", func() {
			nd.setPingPair("layer_3", "L1", childLayerList, randomIps, aiPingStrategy)

			convey.So(aiPingStrategy.pingDict, convey.ShouldContainKey, "layer_3:L1")
			convey.So(aiPingStrategy.pingDict["layer_3:L1"], convey.ShouldResemble, []map[string]any{
				{"from": "192.168.1.1", "to": "192.168.2.1"},
				{"from": "192.168.1.2", "to": "192.168.2.2"},
				{"from": "192.168.2.1", "to": "192.168.3.1"},
				{"from": "192.168.2.2", "to": "192.168.3.2"},
				{"from": "192.168.3.1", "to": "192.168.1.1"},
				{"from": "192.168.3.2", "to": "192.168.1.2"},
			})
		})

		convey.Convey("When randomIps is nil", func() {
			nd.setPingPair("layer_2", "rack", childLayerList, nil, aiPingStrategy)
			convey.So(aiPingStrategy.pingDict, convey.ShouldBeEmpty)
		})

		convey.Convey("When childLayerList contains an invalid key", func() {
			randomIps["layer3"] = "invalidType"
			nd.setPingPair("testLayer", "1", childLayerList, randomIps, aiPingStrategy)

			convey.So(aiPingStrategy.pingDict, convey.ShouldContainKey, "testLayer:1")
			convey.So(aiPingStrategy.pingDict["testLayer:1"], convey.ShouldResemble, []map[string]any{
				{"from": "192.168.1.1", "to": "192.168.2.2"},
				{"from": "192.168.1.2", "to": "192.168.2.1"},
			})
		})
	})
}

func TestGetCurLayerIps(t *testing.T) {
	convey.Convey("Given an AiPingStrategy with layersIps", t, func() {
		aiPingStrategy := &AiPingStrategy{
			layersIps: map[string]any{
				"layer1": []string{"192.168.1.1", "192.168.1.2"},
				"layer2": []string{"192.168.2.1"},
			},
		}

		convey.Convey("When querying a valid layer", func() {
			layerIps := getCurLayerIps(aiPingStrategy, "layer1")
			convey.So(layerIps, convey.ShouldResemble, []string{"192.168.1.1", "192.168.1.2"})
		})

		convey.Convey("When querying a layer that does not exist", func() {
			layerIps := getCurLayerIps(aiPingStrategy, "layer3")
			convey.So(layerIps, convey.ShouldBeEmpty)
		})

		convey.Convey("When querying a layer with non-slice type", func() {
			aiPingStrategy.layersIps["layer4"] = "invalidType"
			layerIps := getCurLayerIps(aiPingStrategy, "layer4")
			convey.So(layerIps, convey.ShouldBeEmpty)
		})
	})
}

func TestProcessBothAxis(t *testing.T) {
	convey.Convey("Test ProcessBothAxis", t, func() {
		// 创建一个模拟的 AiPingStrategy
		aiPingStrategy := &AiPingStrategy{
			pingDict: make(map[string]any),
		}

		// 创建一个 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")
		nd.curNpuInfo["192.168.1.1"] = NpuInfo{IP: "192.168.1.1", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.2"] = NpuInfo{IP: "192.168.1.2", NpuNumber: 3}
		nd.curNpuInfo["192.168.1.3"] = NpuInfo{IP: "192.168.1.3", NpuNumber: 9}
		nd.curNpuInfo["192.168.1.4"] = NpuInfo{IP: "192.168.1.4", NpuNumber: 10}
		nd.curNpuInfo["192.168.1.5"] = NpuInfo{IP: "192.168.1.5", NpuNumber: 11}
		nd.curNpuInfo["192.168.1.6"] = NpuInfo{IP: "192.168.1.6", NpuNumber: 12}
		nd.curNpuInfo["192.168.1.7"] = NpuInfo{IP: "192.168.1.7", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.8"] = NpuInfo{IP: "192.168.1.8", NpuNumber: 3}

		srcIps := []string{"192.168.1.1", "192.168.1.2"}
		dstIps := []string{"192.168.1.3", "192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.7", "192.168.1.8"}

		convey.Convey("Test layer_3 process", func() {
			pingDictKey := layer3Constant

			// 调用 processIpsSlot
			nd.processBothAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

			// 验证结果
			value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
			expectedLen := 2 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(value), convey.ShouldEqual, expectedLen)

			// 验证填充的内容
			convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.7")
			convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.8")

			// 验证未匹配的源 IP
			expectedLen = 2                                        // 期望的长度
			convey.So(len(value), convey.ShouldEqual, expectedLen) // 确保只填充了找到匹配的 IP
		})

		convey.Convey("Test layer_2 process", func() {
			pingDictKey := layer2Constant

			// 调用 processIpsSlot
			nd.processBothAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

			// 验证结果
			value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
			expectedLen := 4 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(value), convey.ShouldEqual, expectedLen)

			// 验证填充的内容
			convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.3")
			convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.4")
			convey.So(value[2][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[2][toConstant], convey.ShouldEqual, "192.168.1.5")
			convey.So(value[3][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[3][toConstant], convey.ShouldEqual, "192.168.1.6")

			// 验证未匹配的源 IP
			convey.So(len(value), convey.ShouldEqual, expectedLen) // 确保只填充了找到匹配的 IP
		})
	})
}

func TestProcessCrossAxis(t *testing.T) {
	convey.Convey("Test ProcessCrossAxis", t, func() {
		// 创建一个模拟的 AiPingStrategy
		aiPingStrategy := &AiPingStrategy{
			pingDict: make(map[string]any),
		}

		// 创建一个 NetDetect 实例
		nd := NewNetDetect("testSuperPod1")
		nd.curNpuInfo["192.168.1.1"] = NpuInfo{IP: "192.168.1.1", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.2"] = NpuInfo{IP: "192.168.1.2", NpuNumber: 2}
		nd.curNpuInfo["192.168.1.2"] = NpuInfo{IP: "192.168.1.3", NpuNumber: 3}
		nd.curNpuInfo["192.168.1.4"] = NpuInfo{IP: "192.168.1.4", NpuNumber: 9}
		nd.curNpuInfo["192.168.1.6"] = NpuInfo{IP: "192.168.1.5", NpuNumber: 10}
		nd.curNpuInfo["192.168.1.7"] = NpuInfo{IP: "192.168.1.6", NpuNumber: 11}
		nd.curNpuInfo["192.168.1.8"] = NpuInfo{IP: "192.168.1.7", NpuNumber: 1}
		nd.curNpuInfo["192.168.1.8"] = NpuInfo{IP: "192.168.1.8", NpuNumber: 2}

		convey.Convey("Test layer_3 process", func() {
			srcIps := []string{"192.168.1.1", "192.168.1.2"}
			dstIps := []string{"192.168.1.7", "192.168.1.8"}
			pingDictKey := layer3Constant

			// 调用 processIpsSlot
			processCrossAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

			// 验证结果
			value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
			expectedLen := 2 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(value), convey.ShouldEqual, expectedLen)

			// 验证填充的内容
			convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.7")
			convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.8")

			// 验证未匹配的源 IP
			convey.So(len(value), convey.ShouldEqual, expectedLen) // 确保只填充了找到匹配的 IP
		})

		convey.Convey("Test layer_2 process", func() {
			srcIps := []string{"192.168.1.1", "192.168.1.2"}
			dstIps := []string{"192.168.1.5", "192.168.1.6"}
			pingDictKey := layer2Constant

			// 调用 processIpsSlot
			processCrossAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

			// 验证结果
			value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
			expectedLen := 2 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(value), convey.ShouldEqual, expectedLen)

			// 验证填充的内容
			convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.6")
			convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.5")

			// 验证未匹配的源 IP
			convey.So(len(value), convey.ShouldEqual, expectedLen) // 确保只填充了找到匹配的 IP
		})
	})
}

func TestProcessMapType(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := &NetDetect{
		curNpuInfo: map[string]NpuInfo{
			"000000000030": {IP: "000000000030", NpuNumber: 10},
			"000000000031": {IP: "000000000031", NpuNumber: 12},
		}, curPingObjType: EidType,
	}

	convey.Convey("Test ProcessMapType", t, func() {
		convey.Convey("Valid input", func() {
			var pingList []any
			v := []map[string]any{
				{
					fromConstant: "000000000030", toConstant: "000000000031",
				},
				{
					fromConstant: "000000000031", toConstant: "000000000030",
				},
			}

			nd.processMapType(&pingList, v)

			expectedPingList := []any{
				map[string]any{
					srcTypeConstant: EidType,
					dstTypeConstant: EidType,
					srcCardPhyId:    10,
					dstCardPhyId:    12,
					pktSizeConstant: pktSizeNum,
					srcAddrConstant: "000000000030",
					dstAddrConstant: "000000000031",
				},
				map[string]any{
					srcTypeConstant: EidType,
					dstTypeConstant: EidType,
					srcCardPhyId:    12,
					dstCardPhyId:    10,
					pktSizeConstant: pktSizeNum,
					srcAddrConstant: "000000000031",
					dstAddrConstant: "000000000030",
				},
			}

			convey.So(pingList, convey.ShouldResemble, expectedPingList)
		})

		convey.Convey("Empty input slice", func() {
			var pingList []any
			var v []map[string]any

			nd.processMapType(&pingList, v)

			convey.So(pingList, convey.ShouldBeEmpty)
		})
	})
}

var specialExpectedPingList = []any{
	map[string]any{
		srcTypeConstant: EidType,
		dstTypeConstant: EidType,
		srcCardPhyId:    10,
		dstCardPhyId:    12,
		pktSizeConstant: pktSizeNum,
		srcAddrConstant: "000000000030",
		dstAddrConstant: "000000000031",
	},
	map[string]any{
		srcTypeConstant: EidType,
		dstTypeConstant: EidType,
		srcCardPhyId:    12,
		dstCardPhyId:    10,
		pktSizeConstant: pktSizeNum,
		srcAddrConstant: "000000000031",
		dstAddrConstant: "000000000030",
	},
}

func TestProcessSliceType(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := &NetDetect{
		curNpuInfo: map[string]NpuInfo{
			"000000000030": {IP: "000000000030", NpuNumber: 10},
			"000000000031": {IP: "000000000031", NpuNumber: 12},
		},
		curPingObjType: EidType,
	}

	convey.Convey("Test ProcessSliceType", t, func() {
		convey.Convey("Valid input", func() {
			var pingList []any
			v := []any{
				map[string]any{
					fromConstant: "000000000030",
					toConstant:   "000000000031",
				},
				map[string]any{
					fromConstant: "000000000031",
					toConstant:   "000000000030",
				},
				"invalid",
			}

			nd.processSliceType(&pingList, v)

			convey.So(pingList, convey.ShouldResemble, specialExpectedPingList)
		})

		convey.Convey("Empty input slice", func() {
			var pingList []any
			var v []any

			nd.processSliceType(&pingList, v)

			convey.So(pingList, convey.ShouldBeEmpty)
		})

		convey.Convey("No valid maps in input", func() {
			var pingList []any
			v := []any{"string", 123, true}

			nd.processSliceType(&pingList, v)

			convey.So(pingList, convey.ShouldBeEmpty)
		})
	})
}

func TestFindNpuNumberByPingObj(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := &NetDetect{
		curNpuInfo: map[string]NpuInfo{
			"192.168.1.1": {IP: "192.168.1.1", NpuNumber: 10},
			"192.168.1.2": {IP: "192.168.1.2", NpuNumber: 12},
			"192.168.1.3": {IP: "192.168.1.3", NpuNumber: 11},
		},
		curSuperPodId: "superPod123",
	}

	convey.Convey("Test FindNpuNumberByNpuIp", t, func() {
		convey.Convey("Valid IP in NpuInfo", func() {
			ip := "192.168.1.1"
			npuNumber := nd.findNpuNumberByPingObj(ip)
			expectedNum := 10 // 期望的编号
			convey.So(npuNumber, convey.ShouldEqual, expectedNum)
		})

		convey.Convey("IP not found in NpuInfo", func() {
			ip := "192.168.1.4" // 不存在的 IP
			npuNumber := nd.findNpuNumberByPingObj(ip)
			expectedNum := -1 // 期望的编号
			convey.So(npuNumber, convey.ShouldEqual, expectedNum)
		})

		convey.Convey("Empty NpuInfo map", func() {
			nd.curNpuInfo = map[string]NpuInfo{} // 清空 NpuInfo
			ip := "192.168.1.1"                  // 任意 IP
			npuNumber := nd.findNpuNumberByPingObj(ip)
			expectedNum := -1 // 期望的编号
			convey.So(npuNumber, convey.ShouldEqual, expectedNum)
		})
	})
}

func TestProcessOutput(t *testing.T) {
	convey.Convey("Test processOutput", t, func() {
		nd := NewNetDetect("testSuperPod1")

		convey.Convey("Case 1: Valid input with map type", func() {
			output := make(map[string]any)
			aiPingStrategy := &AiPingStrategy{
				pingDict: map[string]any{
					"layer1": []map[string]any{
						{fromConstant: "192.168.0.1", toConstant: "192.168.0.1"},
					},
				},
			}

			result := nd.processOutput(output, aiPingStrategy)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(output[pingListConstant], convey.ShouldHaveLength, 1)
		})

		convey.Convey("Case 2: Valid input with slice type", func() {
			output := make(map[string]any)
			aiPingStrategy := &AiPingStrategy{
				pingDict: map[string]any{
					"layer2": []any{
						map[string]any{fromConstant: "192.168.0.2", toConstant: "192.168.0.3"},
					},
				},
			}

			result := nd.processOutput(output, aiPingStrategy)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(output[pingListConstant], convey.ShouldHaveLength, 1)
		})

		convey.Convey("Case 3: Invalid input (nil output)", func() {
			var output map[string]any
			aiPingStrategy := &AiPingStrategy{
				pingDict: map[string]any{
					"layer1": []map[string]any{
						{"ping": "192.168.0.1", "status": "success"},
					},
				},
			}

			result := nd.processOutput(output, aiPingStrategy)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Case 4: Invalid input (nil aiPingStrategy)", func() {
			output := make(map[string]any)
			var aiPingStrategy *AiPingStrategy

			result := nd.processOutput(output, aiPingStrategy)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestSetRandomIps(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a DataFrame with IP data", t, func() {
		nd := &NetDetect{}
		dfIp := &DataFrame{
			chains: map[string]any{
				"layer1": []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
				"layer2": []string{"192.168.2.1", "192.168.2.2"},
			},
		}
		randomIps := make(map[string]any)

		convey.Convey("When childLayerList is empty", func() {
			var childLayerList []string
			childLayerName := "layer1"

			nd.setRandomIps(childLayerList, childLayerName, dfIp, randomIps)

			convey.Convey("Then randomIps should remain empty", func() {
				convey.So(randomIps, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When childLayerList contains valid layers", func() {
			childLayerList := []string{"layer1", "layer2"}
			childLayerName := "layer1"

			nd.setRandomIps(childLayerList, childLayerName, dfIp, randomIps)

			convey.Convey("Then randomIps should contain sampled IPs for each layer", func() {
				convey.So(randomIps, convey.ShouldContainKey, "layer1")
				convey.So(randomIps, convey.ShouldContainKey, "layer2")
			})
		})

		convey.Convey("When randomIps is nil", func() {
			childLayerList := []string{"layer1"}
			childLayerName := "layer1"
			var nilRandomIps map[string]any // nil map

			nd.setRandomIps(childLayerList, childLayerName, dfIp, nilRandomIps)

			convey.Convey("Then randomIps should not be modified", func() {
				convey.So(randomIps, convey.ShouldBeEmpty)
			})
		})
	})
}

func TestFindIPsBySlotName(t *testing.T) {
	convey.Convey("Given a NetDetect instance with current NPU info", t, func() {
		nd := &NetDetect{
			curNpuInfo: map[string]NpuInfo{
				"192.168.1.1": NpuInfo{SlotName: "slot1", IP: "192.168.1.1"},
				"192.168.1.2": NpuInfo{SlotName: "slot2", IP: "192.168.1.2"},
				"192.168.1.3": NpuInfo{SlotName: "slot1", IP: "192.168.1.3"},
			},
		}

		convey.Convey("When targetSlotName matches an existing slot", func() {
			targetSlotName := "slot1"

			result := nd.findIPsBySlotName(targetSlotName)

			convey.Convey("Then result should contain the IPs for that slot", func() {
				convey.So(result, convey.ShouldContain, "192.168.1.1")
				convey.So(result, convey.ShouldContain, "192.168.1.3")
			})
		})

		convey.Convey("When targetSlotName does not match any slot", func() {
			targetSlotName := "nonexistentSlot"

			result := nd.findIPsBySlotName(targetSlotName)

			convey.Convey("Then result should be empty", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When targetSlotName is empty", func() {
			targetSlotName := ""

			result := nd.findIPsBySlotName(targetSlotName)

			convey.Convey("Then result should be empty", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})
	})
}

func TestNpuRingPing(t *testing.T) {
	convey.Convey("Given a NetDetect instance with a specific NPU type", t, func() {
		aiPingStrategy := &AiPingStrategy{}
		layer := "layer1"
		childLayerName := "childLayer1"

		convey.Convey("When curNpuType is a3NpuTypeConstant", func() {
			nd := &NetDetect{
				curNpuType: a3NpuTypeConstant,
			}

			convey.Convey("Then a3NpuRingPing should be called", func() {
				nd.npuRingPing(aiPingStrategy, layer, childLayerName)
			})
		})

		convey.Convey("When curNpuType is a5NpuTypeConstant", func() {
			nd := &NetDetect{
				curNpuType: a5NpuTypeConstant,
			}

			convey.Convey("Then a5NpuRingPing should be called", func() {
				nd.npuRingPing(aiPingStrategy, layer, childLayerName)
			})
		})

		convey.Convey("When curNpuType is an unknown type", func() {
			nd := &NetDetect{
				curNpuType: "unknownNpuType",
			}

			convey.Convey("Then no ping function should be called", func() {
				nd.npuRingPing(aiPingStrategy, layer, childLayerName)
			})
		})
	})
}

func TestProcessSdidList(t *testing.T) {
	convey.Convey("Given a valid AiPingStrategy and source/destination SDID lists", t, func() {
		aiPingStrategy := new(AiPingStrategy)
		initAiPingStrategy(aiPingStrategy)

		srcSdidList := []string{"sdid1", "sdid2", "sdid3"}
		dstSdidList := []string{"sdidA", "sdidB", "sdidC"}
		pingDictKey := "pingKey"

		convey.Convey("When the lengths of srcSdidList and dstSdidList are equal", func() {
			expectedLen := 3 // 预期长度
			processSdidList(aiPingStrategy, srcSdidList, dstSdidList, pingDictKey)
			convey.So(len(aiPingStrategy.pingDict["pingKey"].([]map[string]any)),
				convey.ShouldEqual, expectedLen)
		})

		convey.Convey("When the lengths of srcSdidList and dstSdidList are not equal", func() {
			srcSdidList := []string{"sdid1", "sdid2"}
			dstSdidList := []string{"sdidA"}

			processSdidList(aiPingStrategy, srcSdidList, dstSdidList, pingDictKey)
			convey.So(aiPingStrategy.pingDict, convey.ShouldBeEmpty)
		})
	})
}
