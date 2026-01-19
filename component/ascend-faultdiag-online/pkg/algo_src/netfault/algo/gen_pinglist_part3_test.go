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

func TestSortBySuffix(t *testing.T) {
	convey.Convey("Testing sortBySuffix function", t, func() {
		convey.Convey("When both strings are valid", func() {
			convey.So(sortBySuffix("item-1", "item-2"), convey.ShouldBeTrue)   // 1 < 2
			convey.So(sortBySuffix("item-2", "item-1"), convey.ShouldBeFalse)  // 2 > 1
			convey.So(sortBySuffix("item-10", "item-2"), convey.ShouldBeFalse) // 10 > 2
			convey.So(sortBySuffix("item-3", "item-3"), convey.ShouldBeFalse)  // 3 == 3
		})

		convey.Convey("When one or both strings have invalid formats", func() {
			convey.So(sortBySuffix("item-abc", "item-2"), convey.ShouldBeFalse) // 非法格式
			convey.So(sortBySuffix("item-2", "item-xyz"), convey.ShouldBeFalse) // 非法格式
			convey.So(sortBySuffix("item-", "item-1"), convey.ShouldBeFalse)    // 非法格式
			convey.So(sortBySuffix("item-1", "item-"), convey.ShouldBeFalse)    // 非法格式
		})

		convey.Convey("When one string is missing the suffix", func() {
			convey.So(sortBySuffix("item-", "item-1"), convey.ShouldBeFalse) // 非法格式
			convey.So(sortBySuffix("item-1", "item-"), convey.ShouldBeFalse) // 非法格式
		})
	})
}

func TestSortLayerList2(t *testing.T) {
	convey.Convey("Testing sortLayerList2 function", t, func() {
		convey.Convey("When the input is a valid slice of strings", func() {
			childLayerList := []string{"1.2", "1.1", "2.1", "1.10", "2.2"}
			sortLayerList2(childLayerList)
			expectedLen := 5 // 预期长度
			convey.So(len(childLayerList), convey.ShouldResemble, expectedLen)

			childLayerList = []string{"3.1", "3.2", "2.1", "2.10", "2.2"}
			sortLayerList2(childLayerList)
			convey.So(len(childLayerList), convey.ShouldResemble, expectedLen)
		})

		convey.Convey("When the input contains invalid formats", func() {
			childLayerList := []string{"1.2", "1.invalid", "2.1"}
			sortLayerList2(childLayerList)
			// 由于 "1.invalid" 无法转换为整数，排序可能不确定
			// 这里我们只检查排序后的结果是否包含原始元素
			convey.So(childLayerList, convey.ShouldContain, "1.2")
			convey.So(childLayerList, convey.ShouldContain, "1.invalid")
			convey.So(childLayerList, convey.ShouldContain, "2.1")
		})

		convey.Convey("When the input is empty", func() {
			childLayerList := make([]string, 0)
			sortLayerList2(childLayerList)
			convey.So(childLayerList, convey.ShouldBeEmpty)
		})

		convey.Convey("When the input has one element", func() {
			childLayerList := []string{"42.1"}
			sortLayerList2(childLayerList)
			convey.So(childLayerList, convey.ShouldResemble, []string{"42.1"})
		})
	})
}

func TestSortSdidList(t *testing.T) {
	convey.Convey("Testing sortSdidList function", t, func() {
		convey.Convey("When the input is a valid slice of strings representing integers", func() {
			numStrings := []string{"3", "1", "2"}
			sortSdidList(numStrings)
			convey.So(numStrings, convey.ShouldResemble, []string{"1", "2", "3"})

			numStrings = []string{"10", "2", "1", "20"}
			sortSdidList(numStrings)
			convey.So(numStrings, convey.ShouldResemble, []string{"1", "2", "10", "20"})
		})

		convey.Convey("When the input contains invalid strings", func() {
			numStrings := []string{"3", "invalid", "2"}
			sortSdidList(numStrings)
			// 由于 "invalid" 无法转换为整数，排序可能不确定
			// 这里我们只检查排序后的结果是否包含原始元素
			convey.So(numStrings, convey.ShouldContain, "3")
			convey.So(numStrings, convey.ShouldContain, "2")
			convey.So(numStrings, convey.ShouldContain, "invalid")
		})

		convey.Convey("When the input is empty", func() {
			numStrings := make([]string, 0)
			sortSdidList(numStrings)
			convey.So(numStrings, convey.ShouldBeEmpty)
		})

		convey.Convey("When the input has one element", func() {
			numStrings := []string{"42"}
			sortSdidList(numStrings)
			convey.So(numStrings, convey.ShouldResemble, []string{"42"})
		})
	})
}

func TestSetNecessaryParams(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a valid paramsMap", t, func() {
		nd := &NetDetect{}

		paramsMap := map[string]any{
			argsPeriod:      10,
			argsSPeriod:     5,
			argsPingObjType: 1,
			argsServerIdMap: map[string]string{"server1": "id1", "server2": "id2"},
		}

		npuInfoMap := map[string]NpuInfo{
			"npu1": {RackName: "rack1", NpuNumber: 0, NetPlaneId: "netplane1"},
		}

		convey.Convey("When setNecessaryParams is called with valid parameters", func() {
			result := nd.setNecessaryParams(paramsMap, npuInfoMap)

			convey.Convey("Then it should return true and set parameters correctly", func() {
				convey.So(result, convey.ShouldBeTrue)
				convey.So(nd.curPingPeriod, convey.ShouldEqual, 10)      // 预期的值
				convey.So(nd.curSuppressedPeriod, convey.ShouldEqual, 5) // 预期的值
				convey.So(nd.curPingObjType, convey.ShouldEqual, 1)
				convey.So(nd.curServerIdMap, convey.ShouldResemble,
					map[string]string{"server1": "id1", "server2": "id2"})
				convey.So(nd.curNpuInfo, convey.ShouldResemble, npuInfoMap)
			})
		})
	})
}

func TestSetDefaultParams(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := &NetDetect{}
		convey.Convey("When valid parameters are provided", func() {
			params := map[string]any{argsAxisStrategy: "both_axis", argsNpuType: "A3",
				argsSuperPodJobFlag: true}
			result := nd.setDefaultParams(params)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(nd.curAxisStrategy, convey.ShouldEqual, "both_axis")
			convey.So(nd.curNpuType, convey.ShouldEqual, "A3")
			convey.So(nd.curSuperPodJobFlag, convey.ShouldBeTrue)
		})

		convey.Convey("When invalid parameters are provided", func() {
			params := map[string]any{argsAxisStrategy: "aaa_axis", argsNpuType: "A33",
				argsSuperPodJobFlag: "true", "superPodArr": "aaa"}
			result := nd.setDefaultParams(params)
			convey.So(result, convey.ShouldBeFalse)
			convey.So(nd.curAxisStrategy, convey.ShouldEqual, "")
			convey.So(nd.curNpuType, convey.ShouldEqual, "")
			convey.So(nd.curSuperPodJobFlag, convey.ShouldBeFalse)
		})
	})
}

func TestSetPingDict(t *testing.T) {
	convey.Convey("Given a NetDetect instance", t, func() {
		nd := &NetDetect{}
		aiPingStrategy := new(AiPingStrategy)

		convey.Convey("When valid parameters are provided", func() {
			res := nd.setPingDict(aiPingStrategy, nil)
			convey.So(res, convey.ShouldBeFalse)
		})

		convey.Convey("When no parameters are provided", func() {
			dfChainsMap := make(map[string]*DataFrame)
			dfChainsMap["aa"] = &DataFrame{}
			res := nd.setPingDict(aiPingStrategy, dfChainsMap)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestProcessSameAxis(t *testing.T) {
	convey.Convey("Test TestProcessSameAxis", t, func() {
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

			processSameAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

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

			processSameAxis(aiPingStrategy, srcIps, dstIps, pingDictKey)

			// 验证结果
			value, ok := aiPingStrategy.pingDict[pingDictKey].([]map[string]any)
			expectedLen := 2 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(value), convey.ShouldEqual, expectedLen)

			// 验证填充的内容
			convey.So(value[0][fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(value[0][toConstant], convey.ShouldEqual, "192.168.1.5")
			convey.So(value[1][fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(value[1][toConstant], convey.ShouldEqual, "192.168.1.6")

			// 验证未匹配的源 IP
			convey.So(len(value), convey.ShouldEqual, expectedLen) // 确保只填充了找到匹配的 IP
		})
	})
}
