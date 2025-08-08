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
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

const logLineLength = 256

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestSetFaultDetectParam(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := NewNetDetect("testSuperPod1")

	convey.Convey("Test SetFaultDetectParam", t, func() {
		convey.Convey("Empty paramsMap", func() {
			paramsMap := make(map[string]any)
			npuInfoMap := make(map[string]NpuInfo)
			result := nd.SetFaultDetectParam(paramsMap, npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Invalid paramsMap", func() {
			paramsMap := map[string]any{
				"invalidKey": "invalidValue",
			}
			npuInfoMap := make(map[string]NpuInfo)
			result := nd.SetFaultDetectParam(paramsMap, npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		paramsMap := map[string]any{
			argsPeriod:       10,
			argsSPeriod:      20,
			argsAxisStrategy: "crossAxis",
			argsPingObjType:  EidType,
		}

		convey.Convey("Invalid npuInfoMap", func() {
			npuInfoMap := map[string]NpuInfo{}
			result := nd.SetFaultDetectParam(paramsMap, npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Valid params", func() {
			npuInfoMap := map[string]NpuInfo{
				"validKey": {},
			}
			result := nd.SetFaultDetectParam(paramsMap, npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Default axis strategy", func() {
			paramsMap = map[string]any{
				argsPeriod:      10,
				argsSPeriod:     20,
				argsServerIdMap: map[string]string{"0": "node1"},
			}
			result := nd.SetFaultDetectParam(paramsMap, nil)
			convey.So(result, convey.ShouldBeFalse)
			convey.So(nd.curAxisStrategy, convey.ShouldEqual, crossAxisConstant)
		})
	})
}

func TestCheckParams(t *testing.T) {
	convey.Convey("Test CheckParams", t, func() {
		convey.Convey("Valid paramsMap with argsPeriod and argsSPeriod", func() {
			paramsMap := map[string]any{
				argsPeriod:  10,
				argsSPeriod: 5,
			}
			result := checkParamsMap(paramsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Missing argsSPeriod", func() {
			paramsMap := map[string]any{
				argsPeriod: 10,
			}
			result := checkParamsMap(paramsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Missing argsPeriod", func() {
			paramsMap := map[string]any{
				argsSPeriod: 5,
			}
			result := checkParamsMap(paramsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Empty paramsMap", func() {
			paramsMap := map[string]any{}
			result := checkParamsMap(paramsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Invalid key in paramsMap", func() {
			paramsMap := map[string]any{
				argsPeriod: 10,
				"otherKey": 100,
			}
			result := checkParamsMap(paramsMap)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestCheckNpuAInfoMap(t *testing.T) {
	convey.Convey("Test CheckNpuAInfoMap", t, func() {
		convey.Convey("Valid npuInfoMap", func() {
			npuInfoMap := map[string]NpuInfo{
				"npu1": {
					RackName:   "Rack1",
					SlotName:   "Slot1",
					NpuNumber:  1,
					IP:         "192.168.1.1",
					NetPlaneId: "net_plane0",
				},
			}
			result := checkNpuAInfoMap(npuInfoMap)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Empty npuInfoMap", func() {
			npuInfoMap := map[string]NpuInfo{}
			result := checkNpuAInfoMap(npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Nil npuInfoMap", func() {
			var npuInfoMap map[string]NpuInfo = nil
			result := checkNpuAInfoMap(npuInfoMap)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestInitAiPingStrategy(t *testing.T) {
	convey.Convey("Test InitAiPingStrategy", t, func() {
		// 创建一个 AiPingStrategy 实例
		aiPingStrategy := &AiPingStrategy{}

		// 调用初始化函数
		initAiPingStrategy(aiPingStrategy)

		convey.Convey("Verify npuNpuList is empty", func() {
			convey.So(len(aiPingStrategy.npuNpuList), convey.ShouldEqual, 0)
		})

		convey.Convey("Verify chainList is empty", func() {
			convey.So(len(aiPingStrategy.chainList), convey.ShouldEqual, 0)
		})

		convey.Convey("Verify pingList is empty", func() {
			convey.So(len(aiPingStrategy.pingList), convey.ShouldEqual, 0)
		})

		convey.Convey("Verify layersIps is empty", func() {
			convey.So(len(aiPingStrategy.layersIps), convey.ShouldEqual, 0)
		})

		convey.Convey("Verify dfGrouped is initialized", func() {
			convey.So(aiPingStrategy.dfGrouped, convey.ShouldNotBeNil)
		})

		convey.Convey("Verify pingDict is empty", func() {
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})
	})
}

func TestProcessInput(t *testing.T) {
	// 初始化 NetDetect 实例
	nd := NewNetDetect("testSuperPod1")

	convey.Convey("Given a NetDetect instance", t, func() {
		convey.Convey("When aiPingStrategy is nil", func() {
			input := map[string]any{
				argsNpu2NetPlane: map[string]any{},
				argsNpu2Npu:      []string{},
			}
			result := nd.processInput(input, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When input parameter type assertion fails", func() {
			input := map[string]any{
				argsNpu2NetPlane: "invalidType", // 错误的类型
				argsNpu2Npu:      []string{},
			}
			aiPingStrategy := &AiPingStrategy{}
			result := nd.processInput(input, aiPingStrategy)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When npu2NetPlane internal type assertion fails", func() {
			input := map[string]any{
				argsNpu2NetPlane: map[string]any{
					"key": "invalidType", // 错误的类型
				},
				argsNpu2Npu: []string{},
			}
			aiPingStrategy := &AiPingStrategy{}
			result := nd.processInput(input, aiPingStrategy)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When input is valid", func() {
			input := map[string]any{
				argsNpu2NetPlane: map[string]any{
					"plane1": []string{"path1", "path2"},
					"plane2": []string{"path3", "path4"},
				},
				argsNpu2SuperPod: map[string]any{},
				argsNpu2Npu:      []string{"npu1", "npu2"},
			}
			aiPingStrategy := &AiPingStrategy{}
			result := nd.processInput(input, aiPingStrategy)
			convey.So(result, convey.ShouldBeTrue)
		})
	})
}

func TestSetNpu2NetPlane(t *testing.T) {
	nd := NewNetDetect("testSuperPod1")

	aiPingStrategy := new(AiPingStrategy)
	initAiPingStrategy(aiPingStrategy)

	convey.Convey("Given a NetDetect instance and an AiPingStrategy", t, func() {
		// 测试 net_plane0
		convey.Convey("When planeName is net_plane0", func() {
			fullPaths := []string{"L2.NA:0#Rack-2.NA:0#Rack-2.NSlot-0:0#NPU-0.0.0.0.1:0",
				"L2.NA:0#Rack-2.NA:0#Rack-2.NSlot-0:0#NPU-1.0.0.0.2:0"}
			expected := []string{"L2:2#Rack-2:0#Rack-2.NSlot-0:0#0.0.0.1", "L2:2#Rack-2:1#Rack-2.NSlot-0:0#0.0.0.2"}

			nd.setNpuFullPath("net_plane0", fullPaths, aiPingStrategy)

			convey.So(len(aiPingStrategy.chainList["net_plane0"]), convey.ShouldEqual, len(expected))
			for i, exp := range expected {
				convey.So(aiPingStrategy.chainList["net_plane0"][i], convey.ShouldEqual, exp)
			}
		})

		// 测试 net_plane1
		convey.Convey("When planeName is net_plane1", func() {
			fullPaths := []string{"L2.NA:0#Rack-2.NA:0#Rack-2.NSlot-0:0#NPU-0.0.0.1.0:0",
				"L2.NA:0#Rack-2.NA:0#Rack-2.NSlot-0:0#NPU-1.0.0.2.0:0"}
			expected := []string{"L2:2#Rack-2:0#Rack-2.NSlot-0:0#0.0.1.0", "L2:2#Rack-2:1#Rack-2.NSlot-0:0#0.0.2.0"}

			nd.setNpuFullPath("net_plane1", fullPaths, aiPingStrategy)

			convey.So(len(aiPingStrategy.chainList["net_plane1"]), convey.ShouldEqual, len(expected))
			for i, exp := range expected {
				convey.So(aiPingStrategy.chainList["net_plane1"][i], convey.ShouldEqual, exp)
			}
		})

		// 测试 net_plane2
		convey.Convey("When planeName is net_plane2", func() {
			fullPaths := []string{"InvalidPath", "AnotherInvalidPath"}
			var expected []string

			nd.setNpuFullPath("net_plane2", fullPaths, aiPingStrategy)

			convey.So(len(aiPingStrategy.chainList["net_plane2"]), convey.ShouldEqual, len(expected))
			for i, exp := range expected {
				convey.So(aiPingStrategy.chainList["net_plane2"][i], convey.ShouldEqual, exp)
			}
		})

		// 检查 curTopo
		convey.Convey("Checking curTopo length", func() {
			expectedLen := 4 // 期望的长度
			convey.So(len(nd.curTopo), convey.ShouldEqual, expectedLen)
		})
	})
}

func TestSetNpu2Npu(t *testing.T) {
	convey.Convey("Given a NetDetect instance and an AiPingStrategy", t, func() {
		convey.Convey("When npu_npu is not null", func() {
			fullPaths := []string{"0.0.0.1:0#0.0.0.2:0", "0.0.1.0:0#0.0.2.0:0"}
			expected := []string{"0.0.0.1:0#0.0.0.2:0", "0.0.1.0:0#0.0.2.0:0"}

			aiPingStrategy := new(AiPingStrategy)
			initAiPingStrategy(aiPingStrategy)

			setNpu2Npu(fullPaths, aiPingStrategy)

			convey.So(len(aiPingStrategy.npuNpuList), convey.ShouldEqual, len(expected))
			for i, exp := range expected {
				convey.So(aiPingStrategy.npuNpuList[i], convey.ShouldEqual, exp)
			}
		})

		convey.Convey("When npu_npu is null", func() {
			var fullPaths []string
			var expected []string

			aiPingStrategy := new(AiPingStrategy)
			initAiPingStrategy(aiPingStrategy)

			setNpu2Npu(fullPaths, aiPingStrategy)

			convey.So(len(aiPingStrategy.npuNpuList), convey.ShouldEqual, len(expected))
			for i, exp := range expected {
				convey.So(aiPingStrategy.npuNpuList[i], convey.ShouldEqual, exp)
			}
		})
	})
}

func TestSetDfChainMap(t *testing.T) {
	convey.Convey("Given a chainLists and dfChainsMap", t, func() {
		convey.Convey("When dfChainsMap is nil", func() {
			chainLists := map[string][]string{
				"plane1": {"row1", "row2"},
			}
			result := setDfChainMap(chainLists, nil)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When colNum is less than minimumColNum", func() {
			chainLists := map[string][]string{
				"plane1": {"row1"}, // 列数不足
			}
			dfChainsMap := make(map[string]*DataFrame)
			result := setDfChainMap(chainLists, dfChainsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When colNum is multiple of baseEvenNum", func() {
			chainLists := map[string][]string{
				"plane1": {"row1-row2-row3-row4"}, // 列数为 4
			}
			dfChainsMap := make(map[string]*DataFrame)
			result := setDfChainMap(chainLists, dfChainsMap)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("When input is valid", func() {
			chainLists := map[string][]string{
				"plane1": {"L1:0#Rack-2:0#rack-2.NSlot-0:0#NPU0-0.0.0.1"},
				"plane2": {"L1:0#Rack-2:0#rack-2.NSlot-0:0#NPU0-0.0.0.2"},
			}
			dfChainsMap := make(map[string]*DataFrame)
			result := setDfChainMap(chainLists, dfChainsMap)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(len(dfChainsMap), convey.ShouldEqual, len(chainLists))
		})
	})
}

func TestInitDfChain(t *testing.T) {
	convey.Convey("Given the initDfChain function", t, func() {
		dfChains := initDfChain()

		convey.Convey("When dfChains is initialized", func() {
			convey.So(dfChains, convey.ShouldNotBeNil)

			convey.Convey("Then columnNames should be empty", func() {
				convey.So(len(dfChains.columnNames), convey.ShouldEqual, 0)
			})

			convey.Convey("Then chains should be initialized", func() {
				convey.So(dfChains.chains, convey.ShouldNotBeNil)
				convey.So(len(dfChains.chains), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestSetDfChainColumn(t *testing.T) {
	convey.Convey("Given colNum 5", t, func() {
		colNum := 5
		expectedCols := []string{"layer_2", "port_2", "layer_1", "port_1", "ip"}

		dfChains := initDfChain()
		setDfChainColumn(colNum, dfChains)

		convey.Convey("Then the number of columns should be 5", func() {
			convey.So(len(dfChains.columnNames), convey.ShouldEqual, len(expectedCols))
		})

		convey.Convey("And the columns should match the expected values", func() {
			convey.So(dfChains.columnNames[0], convey.ShouldEqual, expectedCols[0])
			convey.So(dfChains.columnNames[1], convey.ShouldEqual, expectedCols[1])
			convey.So(dfChains.columnNames[2], convey.ShouldEqual, expectedCols[2])
			convey.So(dfChains.columnNames[3], convey.ShouldEqual, expectedCols[3])
			convey.So(dfChains.columnNames[4], convey.ShouldEqual, expectedCols[4])
		})
	})

	convey.Convey("Given colNum 3", t, func() {
		colNum := 3
		expectedCols := []string{"layer_1", "port_1", "ip"}

		dfChains := initDfChain()
		setDfChainColumn(colNum, dfChains)

		convey.Convey("Then the number of columns should be 3", func() {
			convey.So(len(dfChains.columnNames), convey.ShouldEqual, len(expectedCols))
		})

		convey.Convey("And the columns should match the expected values", func() {
			convey.So(dfChains.columnNames[0], convey.ShouldEqual, expectedCols[0])
			convey.So(dfChains.columnNames[1], convey.ShouldEqual, expectedCols[1])
			convey.So(dfChains.columnNames[2], convey.ShouldEqual, expectedCols[2])
		})
	})

	convey.Convey("Given colNum 1", t, func() {
		colNum := 1
		expectedCols := []string{"ip"}

		dfChains := initDfChain()
		setDfChainColumn(colNum, dfChains)

		convey.Convey("Then the number of columns should be 1", func() {
			convey.So(len(dfChains.columnNames), convey.ShouldEqual, len(expectedCols))
		})

		convey.Convey("And the column should match the expected value", func() {
			convey.So(dfChains.columnNames[0], convey.ShouldEqual, expectedCols[0])
		})
	})
}

func TestSetDfChainRow(t *testing.T) {
	dfChains := &DataFrame{
		columnNames: []string{"layer_2", "port_2", "layer_1"},
		chains:      make(map[string]any),
	}

	for _, col := range dfChains.columnNames {
		dfChains.chains[col] = []string{}
	}

	chainList := []string{"slot2:0#NPU0-0.0.0.1", "slot2:0#NPU0-0.0.0.2"}

	setDfChainRow(chainList, dfChains)

	expectedChains := map[string][]string{
		"layer_2": {"slot2", "slot2"},
		"port_2":  {"0", "0"},
		"layer_1": {"NPU0-0.0.0.1", "NPU0-0.0.0.2"},
	}

	convey.Convey("Given a DataFrame with initialized chains", t, func() {
		convey.Convey("Then rowNum should match the length of chainList", func() {
			convey.So(dfChains.rowNum, convey.ShouldEqual, len(chainList))
		})

		convey.Convey("And each chain should match the expected values", func() {
			convey.Convey("For layer_2", func() {
				chain, chainOK := dfChains.chains["layer_2"].([]string)
				convey.So(chainOK, convey.ShouldBeTrue)
				convey.So(chain[0], convey.ShouldEqual, expectedChains["layer_2"][0])
				convey.So(chain[1], convey.ShouldEqual, expectedChains["layer_2"][1])
			})

			convey.Convey("For port_2", func() {
				chain, chainOK := dfChains.chains["port_2"].([]string)
				convey.So(chainOK, convey.ShouldBeTrue)
				convey.So(chain[0], convey.ShouldEqual, expectedChains["port_2"][0])
				convey.So(chain[1], convey.ShouldEqual, expectedChains["port_2"][1])
			})

			convey.Convey("For layer_1", func() {
				chain, chainOK := dfChains.chains["layer_1"].([]string)
				convey.So(chainOK, convey.ShouldBeTrue)
				convey.So(chain[0], convey.ShouldEqual, expectedChains["layer_1"][0])
				convey.So(chain[1], convey.ShouldEqual, expectedChains["layer_1"][1])
			})
		})
	})
}

func TestExtractLayers(t *testing.T) {
	convey.Convey("Given a DataFrame with column names", t, func() {
		dfChains := &DataFrame{
			columnNames: []string{"layer1_data", "non_layer_data", "layer2_data", "layer3_data"},
		}

		convey.Convey("When no column contains the layer constant", func() {
			dfChains.columnNames = []string{"data1", "data2", "data3"}
			result := extractLayers(dfChains)
			convey.Convey("Then the result should be an empty slice", func() {
				convey.So(result, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When one column contains the layer constant", func() {
			dfChains.columnNames = []string{"data1", "layer1_data", "data2"}
			result := extractLayers(dfChains)
			convey.Convey("Then the result should contain that column", func() {
				convey.So(result, convey.ShouldResemble, []string{"layer1_data"})
			})
		})

		convey.Convey("When multiple columns contain the layer constant", func() {
			dfChains.columnNames = []string{"layer1_data", "data1", "layer2_data", "layer3_data"}
			result := extractLayers(dfChains)
			convey.Convey("Then the result should contain all matching columns", func() {
				convey.So(result, convey.ShouldResemble, []string{"layer1_data", "layer2_data", "layer3_data"})
			})
		})
	})
}

func TestGroupBy(t *testing.T) {
	convey.Convey("Given a DataFrame and empty DataFrameGroupBy", t, func() {
		df := &DataFrame{
			columnNames: []string{"category", "value"},
			chains: map[string]any{
				"category": []string{"A", "B", "A", "C"},
				"value":    []string{"1", "2", "3", "4"},
			},
			rowNum: 4,
		}
		dfGrouped := &DataFrameGroupBy{}

		convey.Convey("When grouping by 'category' column", func() {
			groupBy(df, "category", dfGrouped)

			convey.Convey("Should create correct number of groups", func() {
				convey.So(dfGrouped.groupNums, convey.ShouldEqual, 3) // 期望的分组数
			})

			convey.Convey("Group 'A' should contain 2 rows", func() {
				group := findGroupByKey(dfGrouped, "A")
				convey.So(group, convey.ShouldNotBeNil)
				convey.So(group.groupData.rowNum, convey.ShouldEqual, 2) // 期望的分组数
				convey.So(group.groupData.chains["value"].([]string), convey.ShouldResemble, []string{"1", "3"})
			})

			convey.Convey("Group 'B' should contain 1 row", func() {
				group := findGroupByKey(dfGrouped, "B")
				convey.So(group, convey.ShouldNotBeNil)
				convey.So(group.groupData.rowNum, convey.ShouldEqual, 1) // 期望的分组数
				convey.So(group.groupData.chains["value"].([]string), convey.ShouldResemble, []string{"2"})
			})

			convey.Convey("Group 'C' should contain 1 row", func() {
				group := findGroupByKey(dfGrouped, "C")
				convey.So(group, convey.ShouldNotBeNil)
				convey.So(group.groupData.rowNum, convey.ShouldEqual, 1) // 期望的分组数
				convey.So(group.groupData.chains["value"].([]string), convey.ShouldResemble, []string{"4"})
			})
		})
	})

	convey.Convey("Given empty DataFrame", t, func() {
		df := &DataFrame{
			columnNames: []string{"category", "value"},
			chains:      map[string]any{},
			rowNum:      0,
		}
		dfGrouped := &DataFrameGroupBy{}

		convey.Convey("When grouping empty data", func() {
			groupBy(df, "category", dfGrouped)

			convey.Convey("Should create zero groups", func() {
				convey.So(dfGrouped.groupNums, convey.ShouldEqual, 0)
			})
		})
	})
}

func TestGetGroup(t *testing.T) {
	// 创建一个 DataFrameGroupBy 实例
	dfGrouped := &DataFrameGroupBy{
		groupNums: 2,
		groups: []*Group{
			{key: "group1", groupData: &DataFrame{}},
			{key: "group2", groupData: &DataFrame{}},
		},
	}

	convey.Convey("Test GetGroup", t, func() {
		convey.Convey("Group exists", func() {
			result := getGroup(dfGrouped, "group1")
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldEqual, dfGrouped.groups[0].groupData)
		})

		convey.Convey("Group does not exist", func() {
			result := getGroup(dfGrouped, "group3")
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Empty group key", func() {
			result := getGroup(dfGrouped, "")
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("Nil dfGrouped", func() {
			result := getGroup(nil, "group1")
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

// Helper function to find group by key
func findGroupByKey(dfGrouped *DataFrameGroupBy, key string) *Group {
	for _, group := range dfGrouped.groups {
		if group.key == key {
			return group
		}
	}
	return nil
}

func TestNpuFullPing(t *testing.T) {
	convey.Convey("Given an AiPingStrategy instance", t, func() {
		// 创建一个 AiPingStrategy 实例
		aiPingStrategy := &AiPingStrategy{}

		// 调用初始化函数
		initAiPingStrategy(aiPingStrategy)

		convey.Convey("When npuNpuList contains 2 pairs", func() {
			aiPingStrategy.npuNpuList = []string{"0.0.0.1:0#0.0.0.2:0", "0.0.0.1:0#0.0.0.3:0"}

			convey.Convey("Should generate 4 ping directions (bidirectional)", func() {
				npuFullPing(aiPingStrategy, "layer_1")
				expectedLen := 4 // 期望的长度
				convey.So(len(aiPingStrategy.pingDict[argsNpu2Npu].([]any)), convey.ShouldEqual, expectedLen)
			})
		})

		convey.Convey("When npuNpuList contains 3 pairs", func() {
			aiPingStrategy.npuNpuList = []string{"0.0.0.1:0#0.0.0.2:0", "0.0.0.1:0#0.0.0.3:0", "0.0.0.1:0#0.0.0.4:0"}

			convey.Convey("Should generate 6 ping directions (bidirectional)", func() {
				npuFullPing(aiPingStrategy, "layer_1")
				expectedLen := 6 // 期望的长度
				convey.So(len(aiPingStrategy.pingDict[argsNpu2Npu].([]any)), convey.ShouldEqual, expectedLen)
			})
		})

		convey.Convey("When npuNpuList is empty", func() {
			aiPingStrategy.npuNpuList = []string{}

			convey.Convey("Should generate 0 ping directions", func() {
				npuFullPing(aiPingStrategy, "layer_1")
				convey.So(aiPingStrategy.pingDict[argsNpu2Npu], convey.ShouldEqual, nil)
			})
		})
	})
}

func TestAddPingPair(t *testing.T) {
	// 创建一个 AiPingStrategy 实例
	aiPingStrategy := &AiPingStrategy{}

	// 调用初始化函数
	initAiPingStrategy(aiPingStrategy)

	convey.Convey("Test AddPingPair", t, func() {
		convey.Convey("Nil aiPingStrategy", func() {
			addPingPair(nil, "192.168.1.1", "192.168.1.2", "testKey")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Empty fromIp", func() {
			addPingPair(aiPingStrategy, "", "192.168.1.2", "testKey")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Empty toIp", func() {
			addPingPair(aiPingStrategy, "192.168.1.1", "", "testKey")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Empty pingDictKey", func() {
			addPingPair(aiPingStrategy, "192.168.1.1", "192.168.1.2", "")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Valid parameters", func() {
			addPingPair(aiPingStrategy, "192.168.1.1", "192.168.1.2", "testKey")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 1)

			// 验证 pingDict 中的内容
			value, exists := aiPingStrategy.pingDict["testKey"]
			convey.So(exists, convey.ShouldBeTrue)

			layerKey, layerKeyOk := value.([]any)
			convey.So(layerKeyOk, convey.ShouldBeTrue)
			convey.So(len(layerKey), convey.ShouldEqual, 1)

			pingPair, pingPairOk := layerKey[0].(map[string]any)
			convey.So(pingPairOk, convey.ShouldBeTrue)
			convey.So(pingPair[fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(pingPair[toConstant], convey.ShouldEqual, "192.168.1.2")
		})
	})
}

func TestGeneratePermutations(t *testing.T) {
	// 创建一个 AiPingStrategy 实例
	aiPingStrategy := &AiPingStrategy{}

	// 调用初始化函数
	initAiPingStrategy(aiPingStrategy)

	convey.Convey("Test GeneratePermutations", t, func() {
		convey.Convey("Empty ipList", func() {
			generatePermutations([]string{}, aiPingStrategy, "layer1", "192.168.1.1")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Nil aiPingStrategy", func() {
			generatePermutations([]string{"192.168.1.1"}, nil, "layer1", "192.168.1.1")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Empty layer", func() {
			generatePermutations([]string{"192.168.1.1"}, aiPingStrategy, "", "192.168.1.1")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Empty layerIp", func() {
			generatePermutations([]string{"192.168.1.1"}, aiPingStrategy, "layer1", "")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Less than two IPs", func() {
			generatePermutations([]string{"192.168.1.1"}, aiPingStrategy, "layer1", "192.168.1.1")
			convey.So(len(aiPingStrategy.pingDict), convey.ShouldEqual, 0)
		})

		convey.Convey("Valid parameters", func() {
			ipList := []string{"192.168.1.1", "192.168.1.2"}
			generatePermutations(ipList, aiPingStrategy, "layer1", "192.168.1.1")
			expectedLen := 2 // 期望的长度
			convey.So(len(aiPingStrategy.pingDict["layer1:192.168.1.1"].([]any)),
				convey.ShouldEqual, expectedLen)

			// 验证 pingDict 中的内容
			pingDictKey := "layer1:192.168.1.1"
			value, exists := aiPingStrategy.pingDict[pingDictKey]
			convey.So(exists, convey.ShouldBeTrue)

			layerKey, ok := value.([]any)
			expectedLen = 2 // 期望的长度
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(layerKey), convey.ShouldEqual, expectedLen)

			// 验证生成的 IP 对
			pingPair1, pingPair1Ok := layerKey[0].(map[string]any)
			convey.So(pingPair1Ok, convey.ShouldBeTrue)
			convey.So(pingPair1[fromConstant], convey.ShouldEqual, "192.168.1.1")
			convey.So(pingPair1[toConstant], convey.ShouldEqual, "192.168.1.2")

			pingPair2, pingPair2Ok := layerKey[1].(map[string]any)
			convey.So(pingPair2Ok, convey.ShouldBeTrue)
			convey.So(pingPair2[fromConstant], convey.ShouldEqual, "192.168.1.2")
			convey.So(pingPair2[toConstant], convey.ShouldEqual, "192.168.1.1")
		})
	})
}

func TestGetChildColUniqueList(t *testing.T) {
	// 创建一个 DataFrame 实例
	dfIp := &DataFrame{
		chains: map[string]any{
			"layer1": []string{"ip1", "ip2", "ip1"}, // 包含重复值
			"layer2": []string{"ip3", "ip4"},
		},
	}

	convey.Convey("Test GetChildColUniqueList", t, func() {
		convey.Convey("Valid childLayerName with duplicates", func() {
			result := getChildColUniqueList("layer1", dfIp)
			convey.So(result, convey.ShouldResemble, []string{"ip1", "ip2"}) // 期望唯一值
		})

		convey.Convey("Valid childLayerName without duplicates", func() {
			result := getChildColUniqueList("layer2", dfIp)
			convey.So(result, convey.ShouldResemble, []string{"ip3", "ip4"}) // 期望唯一值
		})

		convey.Convey("Non-existent childLayerName", func() {
			result := getChildColUniqueList("layer3", dfIp)
			convey.So(result, convey.ShouldResemble, []string{}) // 期望返回空切片
		})

		convey.Convey("ChildLayerName is empty", func() {
			result := getChildColUniqueList("", dfIp)
			convey.So(result, convey.ShouldResemble, []string{}) // 期望返回空切片
		})

		convey.Convey("Nil dfIp", func() {
			result := getChildColUniqueList("layer1", nil)
			convey.So(result, convey.ShouldResemble, []string{}) // 期望返回空切片
		})
	})
}

func TestFilterAndExtractIps(t *testing.T) {
	convey.Convey("Test FilterAndExtractIps", t, func() {
		// 创建一个 DataFrame 实例，包含示例数据
		dfIp := &DataFrame{
			rowNum: 5,
			chains: map[string]any{
				"child_layer_name": []string{"layer1", "layer2", "layer1", "layer3", "layer1"},
				"ip":               []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"},
			},
		}

		// 测试用例 1: 正常情况
		childLayerName := "child_layer_name"
		childLayer := "layer1"
		result := filterAndExtractIps(dfIp, childLayerName, childLayer)

		convey.So(len(result), convey.ShouldEqual, 3) // 应返回3个IP地址
		convey.So(result, convey.ShouldResemble, []string{"192.168.1.1", "192.168.1.3", "192.168.1.5"})

		// 测试用例 2: childLayer 不存在
		childLayer = "layer4"
		result = filterAndExtractIps(dfIp, childLayerName, childLayer)

		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片

		// 测试用例 3: childLayerName 列不存在
		childLayerName = "nonexistent_layer_name"
		result = filterAndExtractIps(dfIp, childLayerName, "layer1")

		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片

		// 测试用例 4: IP 列不存在
		dfIp.chains[ipConstant] = "not_a_slice" // 修改为非切片类型
		result = filterAndExtractIps(dfIp, childLayerName, childLayer)

		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片

		// 测试用例 5: 行数不匹配
		dfIp.rowNum = 3 // 修改行数
		result = filterAndExtractIps(dfIp, childLayerName, childLayer)

		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片
	})
}

func TestRandomSample(t *testing.T) {
	convey.Convey("Test RandomSample", t, func() {
		// 测试用例 1: 正常情况
		ipList := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
		n := 3
		result := randomSample(ipList, n)

		convey.So(len(result), convey.ShouldEqual, n)
		convey.So(containsAll(result, ipList), convey.ShouldBeTrue)

		// 测试用例 2: n 大于 ipList 长度
		n = 10
		result = randomSample(ipList, n)
		convey.So(len(result), convey.ShouldEqual, len(ipList))
		convey.So(containsAll(result, ipList), convey.ShouldBeTrue)

		// 测试用例 3: n 为 0
		n = 0
		result = randomSample(ipList, n)
		convey.So(len(result), convey.ShouldEqual, 0)

		// 测试用例 4: ipList 为空
		ipList = []string{}
		n = 3
		result = randomSample(ipList, n)
		convey.So(len(result), convey.ShouldEqual, 0)
	})

}

func TestSampleIPs(t *testing.T) {
	convey.Convey("Test SampleIPs", t, func() {
		// 测试用例 1: 正常情况
		ipList := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
		sampleSize := 3
		result := sampleIPs(ipList, sampleSize)

		convey.So(len(result), convey.ShouldEqual, sampleSize)      // 应返回 sampleSize 个元素
		convey.So(containsAll(result, ipList), convey.ShouldBeTrue) // 结果应包含在原列表中

		// 测试用例 2: sampleSize 大于 ipList 长度
		sampleSize = 10
		result = sampleIPs(ipList, sampleSize)
		convey.So(len(result), convey.ShouldEqual, len(ipList))     // 应返回 ipList 的长度
		convey.So(containsAll(result, ipList), convey.ShouldBeTrue) // 结果应包含在原列表中

		// 测试用例 3: sampleSize 为 0
		sampleSize = 0
		result = sampleIPs(ipList, sampleSize)
		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片

		// 测试用例 4: ipList 为空
		ipList = []string{}
		sampleSize = 3
		result = sampleIPs(ipList, sampleSize)
		convey.So(len(result), convey.ShouldEqual, 0) // 应返回空切片
	})

}

// 辅助函数：检查切片是否包含所有元素
func containsAll(subset, set []string) bool {
	setMap := make(map[string]struct{})
	for _, item := range set {
		setMap[item] = struct{}{}
	}

	for _, item := range subset {
		if _, found := setMap[item]; !found {
			return false
		}
	}

	return true
}

func TestIsEmptyNpuInfo(t *testing.T) {
	convey.Convey("Test IsEmptyNpuInfo", t, func() {
		// 测试用例 1: 所有字段均为空
		info1 := NpuInfo{
			IP:         "",
			NetPlaneId: "",
			SlotName:   "",
			RackName:   "",
			NpuNumber:  0,
		}
		convey.So(isEmptyNpuInfo(info1), convey.ShouldBeTrue)

		// 测试用例 2: 部分字段填充
		info2 := NpuInfo{
			IP:         "192.168.1.1",
			NetPlaneId: "",
			SlotName:   "",
			RackName:   "",
			NpuNumber:  0,
		}
		convey.So(isEmptyNpuInfo(info2), convey.ShouldBeFalse)

		// 测试用例 3: 所有字段均填充
		info3 := NpuInfo{
			IP:         "192.168.1.1",
			NetPlaneId: "net-001",
			SlotName:   "slot-1",
			RackName:   "rack-1",
			NpuNumber:  1,
		}
		convey.So(isEmptyNpuInfo(info3), convey.ShouldBeFalse)

		// 测试用例 4: 仅 NpuNumber 填充
		info4 := NpuInfo{
			IP:         "",
			NetPlaneId: "",
			SlotName:   "",
			RackName:   "",
			NpuNumber:  1,
		}
		convey.So(isEmptyNpuInfo(info4), convey.ShouldBeFalse)
	})
}
