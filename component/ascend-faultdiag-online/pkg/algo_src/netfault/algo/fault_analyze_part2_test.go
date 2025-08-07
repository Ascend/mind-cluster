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

func TestSetFuzzyFaultAlarm(t *testing.T) {
	convey.Convey("Given a rootCauseAlarm map", t, func() {
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseAlarm is nil", func() {
			setFuzzyFaultAlarm("srcId"+fuzzyAlarmFlagChar+"dstId", nil)
			convey.Convey("Then it should return without any error", func() {
				// No error to check
			})
		})

		convey.Convey("When rootCauseObj does not have the correct format", func() {
			setFuzzyFaultAlarm("incorrectFormat", rootCauseAlarm)
			convey.Convey("Then rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When rootCauseObj has the correct format", func() {
			rootCauseObj := "srcId" + fuzzyAlarmFlagChar + "dstId"
			setFuzzyFaultAlarm(rootCauseObj, rootCauseAlarm)

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "srcId")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "dstId")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, majorType)
			})
		})
	})
}

func TestSetNpuFaultAlarm(t *testing.T) {
	convey.Convey("Given a rootCauseAlarm map", t, func() {
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseAlarm is nil", func() {
			setNpuFaultAlarm("testObj", nil)
			convey.Convey("Then it should return without any error", func() {
				// No error to check
			})
		})

		convey.Convey("When rootCauseAlarm is not nil", func() {
			rootCauseObj := "testObj"
			setNpuFaultAlarm(rootCauseObj, rootCauseAlarm)

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, majorType)
			})
		})
	})
}

func TestSetRackFaultAlarm(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a rootCauseAlarm map", t, func() {
		nd := NewNetDetect("testSuperPod1")
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseAlarm is nil", func() {
			nd.setRackFaultAlarm("testObj", nil, "netplane_0")
			convey.Convey("Then it should return without any error", func() {
				// No error to check
			})
		})

		convey.Convey("When rootCauseObj indicates a port failure", func() {
			rootCauseObj := "rack1:2"
			nd.setRackFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "rack1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, rackNetplaneType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, minorType)
			})
		})

		convey.Convey("When rootCauseObj indicates a rack failure", func() {
			rootCauseObj := "rack1"
			nd.setRackFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, rackNetplaneType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, rackNetplaneType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
			})
		})

		convey.Convey("When rootCauseObj format is incorrect", func() {
			rootCauseObj := "rack1:2:extraSegment"
			nd.setRackFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When rootCauseObj contains a non-integer NPU number", func() {
			rootCauseObj := "rack1:invalidNumber"
			nd.setRackFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestSetNodeFaultAlarm(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a rootCauseAlarm map", t, func() {
		nd := NewNetDetect("testSuperPod1")
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseAlarm is nil", func() {
			nd.setNodeFaultAlarm("testObj", nil, "netplane_0")
			convey.Convey("Then it should return without any error", func() {
				// No error to check
			})
		})

		convey.Convey("When rootCauseObj indicates a port failure", func() {
			rootCauseObj := "Node-1.L1:2"
			nd.setNodeFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, npuType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "Node-1.L1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, workNodeType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, minorType)
			})
		})

		convey.Convey("When rootCauseObj indicates a rack failure", func() {
			rootCauseObj := "rack1"
			nd.setNodeFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should be populated correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, workNodeType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, workNodeType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
			})
		})

		convey.Convey("When rootCauseObj format is incorrect", func() {
			rootCauseObj := "rack1:2:extraSegment"
			nd.setNodeFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When rootCauseObj contains a non-integer NPU number", func() {
			rootCauseObj := "rack1:invalidNumber"
			nd.setNodeFaultAlarm(rootCauseObj, rootCauseAlarm, "netplane_0")

			convey.Convey("Then rootCauseAlarm should remain empty", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestClassifyByLayer(t *testing.T) {
	convey.Convey("Given a list of items with layer information", t, func() {
		input := []any{
			map[string]any{informationConstant: "layer_1:someInfo"},
			map[string]any{informationConstant: "layer_2:otherInfo"},
			map[string]any{informationConstant: "layer_3:additionalInfo"},
			map[string]any{informationConstant: "Invalid:info"},
			map[string]any{informationConstant: "layer_1:moreInfo"},
		}

		convey.Convey("When classifyByLayer is called", func() {
			output := classifyByLayer(input)

			convey.Convey("Then the output should categorize items correctly", func() {
				convey.So(len(output[layer1Constant]), convey.ShouldEqual, 2) // 预期长度
				convey.So(len(output[layer2Constant]), convey.ShouldEqual, 1)
				convey.So(len(output[layer3Constant]), convey.ShouldEqual, 1)

				convey.So(output[layer1Constant][0][informationConstant], convey.ShouldEqual, "layer_1:someInfo")
				convey.So(output[layer1Constant][1][informationConstant], convey.ShouldEqual, "layer_1:moreInfo")

				convey.So(output[layer2Constant][0][informationConstant], convey.ShouldEqual, "layer_2:otherInfo")
				convey.So(output[layer3Constant][0][informationConstant], convey.ShouldEqual, "layer_3:additionalInfo")
			})
		})

		convey.Convey("When input contains invalid items", func() {
			inputWithInvalid := []any{
				map[string]any{informationConstant: "layer_1:someInfo"},
				map[string]any{informationConstant: 123}, // Invalid type
				map[string]any{informationConstant: "layer_2:otherInfo"},
			}

			output := classifyByLayer(inputWithInvalid)

			convey.Convey("Then the output should ignore invalid items", func() {
				convey.So(len(output[layer1Constant]), convey.ShouldEqual, 1)
				convey.So(len(output[layer2Constant]), convey.ShouldEqual, 1)
			})
		})
	})
}

func TestClassifyByRack(t *testing.T) {
	convey.Convey("Given a list of items with rack information", t, func() {
		input := []map[string]any{
			{fromLayerConstant: "Layer1#Rack1:Port1"},
			{fromLayerConstant: "Layer1#Rack2:Port1"},
			{fromLayerConstant: "Layer2#Rack1:Port2"},
			{fromLayerConstant: "Layer3#Rack1:Port3"},
			{fromLayerConstant: "InvalidLayer"},
		}

		convey.Convey("When classifyByRack is called", func() {
			output := classifyByRack(input)

			convey.Convey("Then the output should categorize items correctly by rack name", func() {
				expectedLen := 2                                        // 预期长度
				convey.So(len(output), convey.ShouldEqual, expectedLen) // Rack1, Rack2

				convey.So(len(output["Rack1"].([]map[string]any)), convey.ShouldEqual, 3) // 预期长度
				convey.So(len(output["Rack2"].([]map[string]any)), convey.ShouldEqual, 1) // 预期长度

				convey.So(output["Rack1"].([]map[string]any)[0][fromLayerConstant], convey.ShouldEqual,
					"Layer1#Rack1:Port1")
				convey.So(output["Rack2"].([]map[string]any)[0][fromLayerConstant], convey.ShouldEqual,
					"Layer1#Rack2:Port1")
			})
		})

		convey.Convey("When input contains invalid items", func() {
			inputWithInvalid := []map[string]any{
				{fromLayerConstant: "Layer1#Rack1:Port1"},
				{fromLayerConstant: "Layer2#Rack1:Port2"},
				{fromLayerConstant: "InvalidLayer"}, // Invalid item
			}

			output := classifyByRack(inputWithInvalid)

			convey.Convey("Then the output should ignore invalid items", func() {
				convey.So(len(output), convey.ShouldEqual, 1)
				convey.So(len(output["Rack1"].([]map[string]any)), convey.ShouldEqual, baseSegmentNum)
			})
		})

		convey.Convey("When input has multiple items for the same rack", func() {
			inputWithDuplicates := []map[string]any{
				{fromLayerConstant: "Layer1#Rack1:Port1"},
				{fromLayerConstant: "Layer1#Rack1:Port2"},
				{fromLayerConstant: "Layer1#Rack1:Port3"},
			}

			output := classifyByRack(inputWithDuplicates)

			convey.Convey("Then the output should group all items under the same rack", func() {
				expectedLen := 3 // 预期长度
				convey.So(len(output), convey.ShouldEqual, 1)
				convey.So(len(output["Rack1"].([]map[string]any)), convey.ShouldEqual, expectedLen)
			})
		})
	})
}

func TestExcludeLastPaths(t *testing.T) {
	convey.Convey("Given a current path and a last path", t, func() {
		curPath := []map[string]any{
			{fromLayerConstant: "Layer1#Rack1", toLayerConstant: "Layer2#Rack2"},
			{fromLayerConstant: "Layer1#Rack2", toLayerConstant: "Layer2#Rack3"},
			{fromLayerConstant: "Layer1#Rack3", toLayerConstant: "Layer2#Rack1"},
			{fromLayerConstant: "Layer1#Rack4", toLayerConstant: "Layer2#Rack5"},
		}

		lastPath := []map[string]any{
			{fromLayerConstant: "Layer1#Rack1", toLayerConstant: "Layer2#Rack2"},
			{fromLayerConstant: "Layer1#Rack3", toLayerConstant: "Layer2#Rack1"},
		}

		convey.Convey("When excludeLastPaths is called", func() {
			excludeLastPaths(&curPath, lastPath)

			convey.Convey("Then the current path should exclude paths that match the last path racks", func() {
				convey.So(len(curPath), convey.ShouldEqual, 1)
				convey.So(curPath[0][fromLayerConstant], convey.ShouldEqual, "Layer1#Rack4")
				convey.So(curPath[0][toLayerConstant], convey.ShouldEqual, "Layer2#Rack5")
			})
		})

		convey.Convey("When last path has no matching racks", func() {
			lastPathNoMatch := []map[string]any{
				{fromLayerConstant: "Layer1#Rack5", toLayerConstant: "Layer2#Rack6"},
			}

			excludeLastPaths(&curPath, lastPathNoMatch)

			expectedLen := 3 // 预期长度
			convey.Convey("Then the current path should remain unchanged", func() {
				convey.So(len(curPath), convey.ShouldEqual, expectedLen)
			})
		})

		convey.Convey("When current path is empty", func() {
			var curPathEmpty []map[string]any

			excludeLastPaths(&curPathEmpty, lastPath)

			convey.Convey("Then the current path should remain empty", func() {
				convey.So(len(curPathEmpty), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestSetFaultPathList(t *testing.T) {
	convey.Convey("Given a list of alarms with descriptions", t, func() {
		eachChildDfAlarmAll := []map[string]any{
			{descriptionConstant: "S-1 TO S-2#S-3"},
			{descriptionConstant: "D-4 TO D-5#D-4"},
			{descriptionConstant: "S-6 TO S-7#D-8"},
			{descriptionConstant: "S-1 TO S-2#S-2"},
		}

		faultPathList := make([][]string, len(eachChildDfAlarmAll))
		swFaultPathList := make([][]string, len(eachChildDfAlarmAll))

		convey.Convey("When setFaultPathList is called", func() {
			setFaultPathList(eachChildDfAlarmAll, &faultPathList, &swFaultPathList)

			convey.Convey("Then the faultPathList should contain unique elements", func() {
				convey.So(len(faultPathList), convey.ShouldEqual, 4) // 预期长度
			})

			convey.Convey("Then the swFaultPathList should contain unique switches", func() {
				convey.So(len(swFaultPathList), convey.ShouldEqual, 4) // 预期长度
			})
		})

		convey.Convey("When some descriptions are missing", func() {
			eachChildDfAlarmAllWithMissing := []map[string]any{
				{descriptionConstant: "S-1 TO S-2#S-3"},
				{},
				{descriptionConstant: "S-6 TO S-7#D-8"},
			}

			faultPathListMissing := make([][]string, len(eachChildDfAlarmAllWithMissing))
			swFaultPathListMissing := make([][]string, len(eachChildDfAlarmAllWithMissing))

			setFaultPathList(eachChildDfAlarmAllWithMissing, &faultPathListMissing, &swFaultPathListMissing)

			convey.Convey("Then the faultPathList should only include valid descriptions", func() {
				convey.So(len(faultPathListMissing), convey.ShouldEqual, 3) // 预期长度
			})

			convey.Convey("Then the swFaultPathList should only include valid switches", func() {
				convey.So(len(swFaultPathListMissing), convey.ShouldEqual, 3) // 预期长度
			})
		})
	})
}

func TestSetRootCauseList(t *testing.T) {
	convey.Convey("Given fault path and software fault path lists", t, func() {
		nd := NewNetDetect("")

		convey.Convey("When valid paths are provided", func() {
			faultPathList := [][]string{{"192.168.1.1"}, {"192.168.1.2"}}
			swFaultPathList := [][]string{{"192.168.1.1"}, {"192.168.1.3"}}
			var rootCauseList []string
			nd.setRootCauseList(&rootCauseList, faultPathList, swFaultPathList)
			convey.So(rootCauseList, convey.ShouldContain, "192.168.1.1")
		})

		convey.Convey("When both lists are empty", func() {
			var rootCauseList []string
			nd.setRootCauseList(&rootCauseList, [][]string{}, [][]string{})
			convey.So(rootCauseList, convey.ShouldBeEmpty)
		})

		convey.Convey("When lists have different lengths", func() {
			faultPathList := [][]string{{"192.168.1.1"}}
			swFaultPathList := [][]string{{"192.168.1.2"}, {"192.168.1.3"}}
			var rootCauseList []string
			nd.setRootCauseList(&rootCauseList, faultPathList, swFaultPathList)
			convey.So(rootCauseList, convey.ShouldBeEmpty)
		})

		convey.Convey("When all paths are identical", func() {
			faultPathList := [][]string{{"192.168.1.1"}, {"192.168.1.1"}}
			swFaultPathList := [][]string{{"192.168.1.1"}}
			rootCauseList := make([]string, 0)
			nd.setRootCauseList(&rootCauseList, faultPathList, swFaultPathList)
			convey.So(len(rootCauseList), convey.ShouldEqual, 0)
		})
	})
}

func TestGetLastLayer(t *testing.T) {
	convey.Convey("Given a string with numbers", t, func() {
		convey.Convey("When the input is 'Layer 3 and Layer 5'", func() {
			result := getLastLayer("Layer 3 and Layer 5")
			convey.So(result, convey.ShouldEqual, "Layer 2 and Layer 4")
		})

		convey.Convey("When the input is 'Version 10.0.2'", func() {
			result := getLastLayer("Version 10")
			convey.So(result, convey.ShouldEqual, "Version 9")
		})

		convey.Convey("When the input is 'No numbers here'", func() {
			result := getLastLayer("No numbers here")
			convey.So(result, convey.ShouldEqual, "No numbers here")
		})

		convey.Convey("When the input is 'Layer -1'", func() {
			result := getLastLayer("Layer -1")
			convey.So(result, convey.ShouldEqual, "Layer -0")
		})

		convey.Convey("When the input is 'Layer 2, Layer 2'", func() {
			result := getLastLayer("Layer 2, Layer 2")
			convey.So(result, convey.ShouldEqual, "Layer 1, Layer 1")
		})
	})
}

func TestGetCurRootCauseList(t *testing.T) {
	nd := NewNetDetect("testSuperPod1")

	convey.Convey("Given a root cause list and a root cause alarm map", t, func() {
		rootCauseAlarm := map[string][]string{
			"Layer 2": {"Alarm1"},
			"Layer 3": {"Alarm2"},
		}

		convey.Convey("When the layer exists in the root cause alarm", func() {
			rootCauseList := []string{"192.168.1.1", "192.168.1.2", "Alarm3"}
			layer := "Layer 2"
			result := nd.getCurRootCauseList(rootCauseList, &rootCauseAlarm, layer)

			convey.Convey("Then the result should contain unique items from the root cause list", func() {
				convey.So(result, convey.ShouldResemble, []string{"192.168.1.1", "192.168.1.2", "Alarm3"})
			})
		})

		convey.Convey("When the layer does not exist in the root cause alarm", func() {
			rootCauseList := []string{"192.168.1.1", "192.168.1.2", "npu"}
			layer := "Layer 4" // 不存在的层
			result := nd.getCurRootCauseList(rootCauseList, &rootCauseAlarm, layer)

			convey.Convey("Then the result should include all items from the root cause list", func() {
				convey.So(result, convey.ShouldResemble, []string{"192.168.1.1", "192.168.1.2", "npu"})
			})
		})
	})
}

func TestGetRootCauseEvent(t *testing.T) {
	tests := []struct {
		rootCauseList  []string
		rootCauseEvent map[string]string
		expected       string
	}{
		{
			rootCauseList:  []string{"192.168.1.1:some issue", "NSlot:other issue"},
			rootCauseEvent: map[string]string{"event1": npuConstant},
			expected:       "192.168.1.1:some issue",
		},
		{
			rootCauseList:  []string{"10.0.0.1:another issue", "NSlot:port:issue"},
			rootCauseEvent: map[string]string{"event2": npuConstant},
			expected:       "10.0.0.1:another issue",
		},
		{
			rootCauseList:  []string{"192.168.1.1:some issue", "192.168.1.2:npu issue"},
			rootCauseEvent: map[string]string{"event3": "192.168.1.2"},
			expected:       "192.168.1.2:npu issue",
		},
		{
			rootCauseList:  []string{"invalid:entry"},
			rootCauseEvent: map[string]string{"event4": npuConstant},
			expected:       "",
		},
	}

	for _, test := range tests {
		got := getRootCauseEvent(test.rootCauseList, test.rootCauseEvent)
		if got != test.expected {
			t.Errorf("got %q, want %q", got, test.expected)
		}
	}
}

func TestGetFinalAlarm(t *testing.T) {
	convey.Convey("Test getFinalAlarm", t, func() {
		nd := NewNetDetect("testSuperPod1")
		nd.curSlideWindowsMaxTs = 150000
		nd.curSlideWindows = []map[string]any{
			{
				avgLoseRateConstant: float64(0), minLoseRateConstant: float64(0), maxLoseRateConstant: float64(0),
				avgDelayConstant: float64(0), minDelayConstant: float64(0), maxDelayConstant: float64(0),
				srcAddrConstant: "192.168.0.1", dstAddrConstant: "192.168.0.2", timestampConstant: int64(2000),
				pingTaskIDConstant: "aa",
			},
			{
				avgLoseRateConstant: float64(0), minLoseRateConstant: float64(0), maxLoseRateConstant: float64(0),
				avgDelayConstant: float64(0), minDelayConstant: float64(0), maxDelayConstant: float64(0),
				srcAddrConstant: "192.168.0.1", dstAddrConstant: "192.168.0.2", timestampConstant: int64(100000),
				pingTaskIDConstant: "aa",
			},
		}

		convey.Convey("Case 1: Have samePath", func() {
			input := []map[string]any{
				{
					avgLoseRateConstant: float64(100), minLoseRateConstant: float64(100),
					maxLoseRateConstant: float64(100), avgDelayConstant: float64(500), minDelayConstant: float64(100),
					maxDelayConstant: float64(1000), srcAddrConstant: "192.168.0.1", dstAddrConstant: "192.168.0.2",
					timestampConstant: int64(1000), pingTaskIDConstant: "aa",
				},
			}

			expectedLen := 2
			result := nd.getFinalAlarm(input)
			convey.So(len(result), convey.ShouldResemble, expectedLen)
		})

		convey.Convey("Case 2: No samePath", func() {
			input := []map[string]any{
				{
					avgLoseRateConstant: float64(100), minLoseRateConstant: float64(100),
					maxLoseRateConstant: float64(100), avgDelayConstant: float64(500), minDelayConstant: float64(100),
					maxDelayConstant: float64(1000), srcAddrConstant: "192.168.0.3", dstAddrConstant: "192.168.0.4",
					timestampConstant: int64(1000),
				},
			}

			expected := make([]any, 0)
			result := nd.getFinalAlarm(input)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestSetL2FaultAlarm(t *testing.T) {
	convey.Convey("Given a NetDetect instance with an empty curServerIdMap", t, func() {
		nd := NewNetDetect("testSuperPod1")
		nd.curServerIdMap = make(map[string]string)

		convey.Convey("When setL2FaultAlarm is called with a valid rootCauseObj "+
			"containing portIntervalChar", func() {
			rootCauseObj := "L2" + portIntervalChar + "server1"
			rootCauseAlarm := make(map[string]any)
			nd.curServerIdMap["server1"] = "work1"

			nd.setA3L2FaultAlarm(rootCauseObj, rootCauseAlarm)

			convey.Convey("Then rootCauseAlarm should be populated with the correct values", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "work1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, workNodeType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "L2")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, l2NetplaneType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, majorType)
			})
		})

		convey.Convey("When setL2FaultAlarm is called with a valid rootCauseObj not "+
			"containing portIntervalChar", func() {
			rootCauseObj := "switch1"
			rootCauseAlarm := make(map[string]any)

			nd.setA3L2FaultAlarm(rootCauseObj, rootCauseAlarm)

			convey.Convey("Then rootCauseAlarm should be populated with the correct "+
				"values for switch failure", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "switch1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, l2NetplaneType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "switch1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, l2NetplaneType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
			})
		})

		convey.Convey("When setL2FaultAlarm is called with a nil rootCauseAlarm", func() {
			rootCauseObj := "rack1" + portIntervalChar + "server1"
			var rootCauseAlarm map[string]any = nil

			convey.So(func() { nd.setA3L2FaultAlarm(rootCauseObj, rootCauseAlarm) }, convey.ShouldNotPanic)
		})
	})
}

var npuOtherAlarmList = []map[string]any{
	{
		"description": "S-L2:1#Work-1:0#work-1.NSlot-0:0#4194304 TO D-L2:3#Work-3:0#work-3.NSlot-0:0#12582912",
		"dstAddr":     "12582912",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "L2:1#Work-1:0#work-1.NSlot-0:0#4194304",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "4194304",
		"srcType":     "0",
		"toLayer":     "L2:3#Work-3:0#work-3.NSlot-0:0#12582912",
	},
	{
		"description": "S-L2:3#Work-3:0#work-3.NSlot-0:0#12582912 TO D-L2:2#Work-2:0#work-2.NSlot-0:0#8388608",
		"dstAddr":     "8388608",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "L2:3#Work-3:0#work-3.NSlot-0:0#12582912",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "12582912",
		"srcType":     "0",
		"toLayer":     "L2:2#Work-2:0#work-2.NSlot-0:0#8388608",
	},
	{
		"description": "S-L2:1#Work-1:1#work-1.NSlot-0:0#4259841 TO D-L2:3#Work-3:1#work-3.NSlot-0:0#12648449",
		"dstAddr":     "12648449",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "L2:1#Work-1:1#work-1.NSlot-0:0#4259841",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "4259841",
		"srcType":     "0",
		"toLayer":     "L2:3#Work-3:1#work-3.NSlot-0:0#12648449",
	},
	{
		"description": "S-L2:3#Work-3:1#work-3.NSlot-0:0#12648449 TO D-L2:2#Work-2:1#work-2.NSlot-0:0#8454145",
		"dstAddr":     "8454145",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "L2:3#Work-3:1#work-3.NSlot-0:0#12648449",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "12648449",
		"srcType":     "0",
		"toLayer":     "L2:2#Work-2:1#work-2.NSlot-0:0#8454145",
	},
}

func TestGetRootCauseAlarm(t *testing.T) {
	expected := map[string][]string{
		"layer_3": {"L2:3"},
	}

	nd := NewNetDetect("testSuperPod1")

	// 创建一个空的 []any 切片
	var convertedList []any

	// 遍历原始切片并转换
	for _, alarm := range npuOtherAlarmList {
		convertedList = append(convertedList, alarm)
	}

	convey.Convey("Given a NetDetect instance with an empty curServerIdMap", t, func() {
		convey.Convey("With root cause alarms", func() {
			result := nd.getRootCauseAlarm(convertedList)

			convey.So(len(result), convey.ShouldEqual, len(expected))
			convey.So(result["layer_3"], convey.ShouldResemble, []string{"L2:3"})
		})
	})
}

func TestGetOtherFaultAlarm(t *testing.T) {
	nd := NewNetDetect("testSuperPod1")

	// 创建一个空的 []any 切片
	var convertedList []any

	// 遍历原始切片并转换
	for _, alarm := range npuOtherAlarmList {
		convertedList = append(convertedList, alarm)
	}

	convey.Convey("Given a NetDetect instance with an empty curServerIdMap", t, func() {
		convey.Convey("With root cause alarms", func() {
			rootCauseAlarmAll := make([]any, 0)
			detectType := "delay"

			// 调用 getOtherFaultAlarm
			nd.getOtherFaultAlarm(convertedList, &rootCauseAlarmAll, detectType)

			// 验证 rootCauseAlarmAll 是否正确更新
			expectedLength := 1 // 根因对象数量
			convey.So(len(rootCauseAlarmAll), convey.ShouldEqual, expectedLength)
		})
	})
}

func TestSetSuperPodFaultAlarm(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a root cause alarm map", t, func() {
		nd := &NetDetect{
			curNpuInfo:     map[string]NpuInfo{"000031": NpuInfo{SuperPodName: "superPod-1", NpuNumber: 1}},
			curPingObjType: EidType,
		}
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseObj has one element", func() {
			rootCauseObj := "superPod-1"
			nd.setSuperPodFaultAlarm(rootCauseObj, rootCauseAlarm, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, rootCauseObj)
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, superPodType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, roceSwitchConstant)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, roceSwitchType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
				convey.So(rootCauseAlarm[descriptionConstant], convey.ShouldEqual, rootCauseObj+" fault ips: []")
			})
		})

		convey.Convey("When rootCauseObj has two elements", func() {
			rootCauseObj := "superPod-1:1"
			nd.setSuperPodFaultAlarm(rootCauseObj, rootCauseAlarm, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "000031")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, EidType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "superPod-1")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, superPodType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, minorType)
			})
		})

		convey.Convey("When rootCauseObj has four elements", func() {
			rootCauseObj := "superPod-1:superPod-2:[000031, 000032]:[000033, 000034]"
			nd.setSuperPodFaultAlarm(rootCauseObj, rootCauseAlarm, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "superPod-1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, superPodType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, "superPod-2")
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, superPodType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
			})
		})
	})
}

func TestSetRoceSwitchFaultAlarm(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a root cause alarm map", t, func() {
		nd := &NetDetect{
			curNpuInfo:     map[string]NpuInfo{"000031": NpuInfo{SuperPodName: "superPod-1", NpuNumber: 1}},
			curPingObjType: EidType,
		}
		rootCauseAlarm := make(map[string]any)

		convey.Convey("When rootCauseObj has one element", func() {
			rootCauseObj := "ROCESwitch:1"
			nd.setRoceSwitchFaultAlarm(rootCauseObj, rootCauseAlarm, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, "SuperPod-1")
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, superPodType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, roceSwitchConstant)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, roceSwitchType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
				convey.So(rootCauseAlarm[descriptionConstant], convey.ShouldEqual, "SuperPod-1 fault ips: []")
			})
		})

		convey.Convey("When rootCauseObj has two elements", func() {
			rootCauseObj := "ROCESwitch"
			nd.setRoceSwitchFaultAlarm(rootCauseObj, rootCauseAlarm, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(rootCauseAlarm[srcIdConstant], convey.ShouldEqual, roceSwitchConstant)
				convey.So(rootCauseAlarm[srcTypeConstant], convey.ShouldEqual, roceSwitchType)
				convey.So(rootCauseAlarm[dstIdConstant], convey.ShouldEqual, roceSwitchConstant)
				convey.So(rootCauseAlarm[dstTypeConstant], convey.ShouldEqual, roceSwitchType)
				convey.So(rootCauseAlarm[levelConstant], convey.ShouldEqual, criticalType)
			})
		})

		convey.Convey("When rootCauseObj has four elements", func() {
			nd.setSuperPodFaultAlarm("rootCauseObj", nil, nil)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(len(rootCauseAlarm), convey.ShouldEqual, 0)
			})
		})
	})
}

var npuOtherAlarmList2 = []map[string]any{
	{
		"description": "S-NA.L2:1#NA.Work-1:0#Work-1.NSlot-0:0#NPU-0.001 TO " +
			"D-NA.L2:3#NA.Work-3:1#Work-3.NSlot-1:0#NPU-1.002",
		"dstAddr":     "002",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "NA.L2:1#NA.Work-1:0#Work-1.NSlot-0:0#NPU-0.001",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "001",
		"srcType":     "0",
		"toLayer":     "NA.L2:3#NA.Work-3:1#Work-3.NSlot-1:0#NPU-1.002",
	},
	{
		"description": "S-NA.L2:3#NA.Work-3:1#Work-3.NSlot-1:0#NPU-1.002 TO " +
			"D-NA.L2:2#NA.Work-2:0#Work-2.NSlot-0:0#NPU-4.004",
		"dstAddr":     "004",
		"dstType":     "0",
		"faultType":   "avgLossRate",
		"fromLayer":   "NA.L2:3#NA.Work-3:1#Work-3.NSlot-1:0#NPU-1.002",
		"information": "layer_3:L2",
		"pingTaskId":  "077c779f-6e37-c7cc-7d79-b99cfe8cd564",
		"srcAddr":     "002",
		"srcType":     "0",
		"toLayer":     "NA.L2:2#NA.Work-2:0#Work-2.NSlot-0:0#NPU-4.004",
	},
}

func TestGetRoceSwitchDesc(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a root cause alarm map", t, func() {
		nd := &NetDetect{
			curNpuInfo: map[string]NpuInfo{"000031": NpuInfo{SuperPodName: "superPod-1", NpuNumber: 1}},
		}

		// 创建一个空的 []any 切片
		var convertedList []any

		// 遍历原始切片并转换
		for _, alarm := range npuOtherAlarmList {
			convertedList = append(convertedList, alarm)
		}

		convey.Convey("When rootCauseObj has one element", func() {
			desc := nd.getRoceSwitchDesc(convertedList, "SuperPod-1")

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(desc, convey.ShouldEqual, " fault ips: []")
			})
		})
	})
}

func TestGetSuperPodRootCauseList(t *testing.T) {
	convey.Convey("Given a NetDetect instance and a root cause alarm map", t, func() {
		nd := &NetDetect{
			curNpuInfo: map[string]NpuInfo{
				"192.168.0.1": NpuInfo{SuperPodName: "SuperPod-1", NpuNumber: 1},
				"192.168.0.2": NpuInfo{SuperPodName: "SuperPod-2", NpuNumber: 2},
			},
			curSuperPodArr: []string{"SuperPod-1", "SuperPod-2"},
		}

		convey.Convey("When rootCauseObj has one element", func() {
			input := []string{"ROCESwitch:1"}
			list := nd.getSuperPodRootCauseList(input)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(list[0], convey.ShouldEqual, "ROCESwitch:1")
			})
		})

		convey.Convey("When rootCauseObj has two element", func() {
			input := []string{"SuperPod-1:1", "192.168.0.1"}
			list := nd.getSuperPodRootCauseList(input)

			convey.Convey("Then the alarm should be set correctly", func() {
				convey.So(list[0], convey.ShouldEqual, "SuperPod-1:1")
			})
		})

		convey.Convey("When rootCauseObj has four element", func() {
			input := []string{"SuperPod-1", "192.168.0.1", "SuperPod-2", "192.168.0.2"}
			list := nd.getSuperPodRootCauseList(input)

			convey.Convey("Then the alarm should be set correctly", func() {
				expectedLen := 2
				convey.So(len(list), convey.ShouldEqual, expectedLen)
				convey.So(list[0], convey.ShouldEqual, "SuperPod-1:SuperPod-2:[192.168.0.1]:[192.168.0.2]")
				convey.So(list[1], convey.ShouldEqual, "SuperPod-2:SuperPod-1:[192.168.0.2]:[192.168.0.1]")
			})
		})
	})
}
