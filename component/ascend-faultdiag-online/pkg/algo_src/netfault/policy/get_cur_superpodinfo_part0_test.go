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

// Package policy is used for processing super pod information
package policy

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

func TestFormatDifferentByProtocol(t *testing.T) {
	convey.Convey("Test formatDifferentByProtocol", t, func() {
		// 模拟数据
		npuNetPlanePaths := make(map[string][]string)
		param := superPodParam{
			protocol:   "ROCE",
			superPodId: "1",
			npu: &NpuInfo{
				PhyId: "123",
			},
			rack: &RackInfo{
				RackID: "1",
			},
			server: &ServerInfo{
				ServerIndex: "1",
			},
			protocolLevels: &LevelElement{
				NetLayer: 3, RankAddrList: []RankAddrItem{{Addr: "1"}, {Addr: "2"}},
			},
		}
		patches := gomonkey.ApplyFunc(strconv.Itoa, func(i int) string {
			return fmt.Sprintf("%d", i)
		})
		defer patches.Reset()
		convey.Convey("npuNetPlanePaths is nil", func() {
			result := formatDifferentByProtocol(nil, param, 0)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("ROCE protocol", func() {
			result := formatDifferentByProtocol(npuNetPlanePaths, param, 0)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(npuNetPlanePaths["1-1"], convey.ShouldResemble, []string{
				"NA.ROCESwitch:0#NA.SuperPod-1:0#NA.NSlot-0:0#NPU-123.1:0",
			})
		})

		param.protocol = "undefined"
		convey.Convey("undefined protocol", func() {
			result := formatDifferentByProtocol(npuNetPlanePaths, param, 0)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestGetRockNpuLinkPathFromSuperPodJson(t *testing.T) {
	convey.Convey("Test getRoceNpuLinkPathFromSuperPodJson", t, func() {
		convey.Convey("Given nil SuperPodInfo", func() {
			result, _ := getRoceNpuLinkPathFromSuperPodJson(nil, "ROCE")
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("normal", func() {
			s := &SuperPodInfo{
				SuperPodID: "1",
				RackMap: map[string]*RackInfo{
					"rack1": &RackInfo{
						ServerMap: map[string]*ServerInfo{
							"server1": {NpuMap: map[string]*NpuInfo{
								"npu1": {PhyId: "123",
									LevelList: []LevelElement{
										{NetLayer: 3, RankAddrList: []RankAddrItem{{Addr: "1"}, {Addr: "2"}}}},
								}}},
						},
					},
				},
			}
			expectVal := map[string][]string{"-": {"NA.ROCESwitch:0#NA.SuperPod-1:0#NA.NSlot-0:0#NPU-123.1:0"}}
			result, _ := getRoceNpuLinkPathFromSuperPodJson(s, "ROCE")
			convey.So(result, convey.ShouldResemble, expectVal)
		})
	})
}

func TestMergeNpuEidMap(t *testing.T) {
	convey.Convey("TestMergeNpuEidMap", t, func() {
		convey.Convey("when_npuEidMapOutRack_is_empty", func() {
			npuEidMapOutRack := map[string]algo.NpuInfo{}
			npuEidMapFromTopo := map[string]algo.NpuInfo{
				"key1": {},
				"key2": {},
			}
			result := mergeNpuEidMap(npuEidMapOutRack, npuEidMapFromTopo)
			expected := map[string]algo.NpuInfo{
				"key1": {},
				"key2": {},
			}
			convey.So(len(result), convey.ShouldEqual, len(expected))
			for key, _ := range expected {
				convey.So(result[key], convey.ShouldNotBeNil)
			}
		})

		// 测试第二个map为空
		convey.Convey("when_npuEidMapFromTopo_is_empty", func() {
			npuEidMapOutRack := map[string]algo.NpuInfo{
				"key1": {},
				"key2": {},
			}
			npuEidMapFromTopo := map[string]algo.NpuInfo{}
			result := mergeNpuEidMap(npuEidMapOutRack, npuEidMapFromTopo)
			expected := map[string]algo.NpuInfo{
				"key1": {},
				"key2": {},
			}
			convey.So(len(result), convey.ShouldEqual, len(expected))
			for key, _ := range expected {
				convey.So(result[key], convey.ShouldNotBeNil)
			}
		})
	})
}

func TestFuncGetNpuEidMapInfo(t *testing.T) {
	convey.Convey("test func get NpuEidMapInfo", t, func() {
		convey.Convey("", func() {
			npuEidMap := map[string]algo.NpuInfo{
				"11": algo.NpuInfo{}}
			param := superPodParam{
				rack: &RackInfo{
					RackID: "1"},
				server: &ServerInfo{ServerIndex: "4",
					NpuMap: map[string]*NpuInfo{
						"1": &NpuInfo{PhyId: "1", LevelList: []LevelElement{
							{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "11"}, {Addr: "2", PlaneId: "1"}}}}},
					},
				},

				superPodId: "1",
			}
			getNpuEidMapInfo(npuEidMap, param)
			expect := map[string]algo.NpuInfo{
				"2": {SuperPodName: "SuperPod-1", RackName: "Rack-1", OsName: "4", SlotName: "NSlot-0", NpuNumber: 1,
					NetPlaneId: "netplane_1"},
				"11": algo.NpuInfo{},
			}
			convey.So(npuEidMap, convey.ShouldResemble, expect)
		})
	})
}

func TestFuncGetServerNpuEidMapInfo(t *testing.T) {
	convey.Convey("test func getServerNpuEidMapInfo", t, func() {
		convey.Convey("", func() {
			npuEidMap := map[string]algo.NpuInfo{
				"11": algo.NpuInfo{}}
			param := superPodParam{
				rack: &RackInfo{
					RackID: "1",
					ServerMap: map[string]*ServerInfo{"4": &ServerInfo{ServerIndex: "4",
						NpuMap: map[string]*NpuInfo{
							"1": &NpuInfo{PhyId: "1", LevelList: []LevelElement{
								{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "11"}, {Addr: "2", PlaneId: "1"}}}}},
						}},
						"0": {},
					}},

				superPodId: "1",
			}
			getServerNpuEidMapInfo(npuEidMap, param)
			expect := map[string]algo.NpuInfo{
				"2": {SuperPodName: "SuperPod-1", RackName: "Rack-1", OsName: "4", SlotName: "NSlot-0", NpuNumber: 1,
					NetPlaneId: "netplane_1"},
				"11": algo.NpuInfo{},
			}
			convey.So(npuEidMap, convey.ShouldResemble, expect)
		})
	})
}

func TestLoopWaitFile(t *testing.T) {
	convey.Convey("test func loopWaitFile", t, func() {
		convey.Convey("return false when not exist", func() {
			mockCheck := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer mockCheck.Reset()
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist)
			defer mockStat.Reset()
			mockSleep := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer mockSleep.Reset()
			ret := loopWaitFile("filePath", "DirPath")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("return false when controllered exitd", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer mockStat.Reset()
			mockCheck := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer mockCheck.Reset()
			controllerflags.IsControllerExited.SetState(true)
			ret := loopWaitFile("filePath", "DirPath")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

func TestGetSuperPodsRoceNpuInfo(t *testing.T) {
	convey.Convey("Test GetSuperPodsRoceNpuInfo", t, func() {

		convey.Convey("return nil When paths is empty", func() {
			ret1, ret2 := GetSuperPodsRoceNpuInfo([]string{})
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})

		convey.Convey("return nil When readConfigMap", func() {
			patch0 := gomonkey.ApplyFuncReturn(readConfigMap, nil)
			defer patch0.Reset()
			ret1, ret2 := GetSuperPodsRoceNpuInfo([]string{})
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})
		convey.Convey("return nil when get Roce err", func() {
			patch0 := gomonkey.ApplyFuncReturn(readConfigMap, &SuperPodInfo{})
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFuncReturn(getRoceNpuLinkPathFromSuperPodJson, nil, nil)
			defer patch1.Reset()
			ret1, ret2 := GetSuperPodsRoceNpuInfo([]string{})
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})

	})
}

func TestGetNpuServerIdFromRackMap(t *testing.T) {
	convey.Convey("test func getNpuServerIdFromRackMap", t, func() {
		convey.Convey("return nil when ServerMap nil", func() {
			ret := getNpuServerIdFromRackMap(0, &RackInfo{})
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("return nil when NpuMap nil", func() {
			r := &RackInfo{ServerMap: map[string]*ServerInfo{"1": {}}}
			ret := getNpuServerIdFromRackMap(0, r)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("return ServerId when ServerMap normal", func() {
			r := &RackInfo{ServerMap: map[string]*ServerInfo{
				"1": {ServerIndex: "1", NpuMap: map[string]*NpuInfo{"1": {PhyId: "1"}}}}}
			ret := getNpuServerIdFromRackMap(1, r)
			convey.So(ret, convey.ShouldEqual, "1")
		})
		convey.Convey("return ServerId when phyId Itoa failed", func() {
			r := &RackInfo{ServerMap: map[string]*ServerInfo{
				"1": {NpuMap: map[string]*NpuInfo{"1": {PhyId: "S"}}}}}
			ret := getNpuServerIdFromRackMap(0, r)
			convey.So(ret, convey.ShouldBeEmpty)
		})
	})
}

func TestGetRoceLimitNpuNumPerSuperPod(t *testing.T) {
	convey.Convey("test func getRoceLimitNpuNumsPerSuperPod", t, func() {
		convey.Convey("return false when getRoceLimitNpuNumsPerSuperPod", func() {
			_, _, ret := getRoceLimitNpuNumsPerSuperPod(map[string][]string{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("return false when len valid", func() {
			s := map[string][]string{
				"1": {"a", "b"},
				"2": {"a", "b"},
				"3": {"a", "b"},
				"4": {"a", "b"},
			}
			_, _, ret := getRoceLimitNpuNumsPerSuperPod(s)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("return true when normal", func() {
			s := map[string][]string{
				"1": {"a", "b"},
				"2": {"a", "b"},
				"3": {"a", "b"},
				"4": {"a", "b"},
				"5": {"a", "b"},
			}
			_, _, ret := getRoceLimitNpuNumsPerSuperPod(s)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestGetEidInfoOrIpFromPorts(t *testing.T) {
	convey.Convey("TestGetEidInfoOrIpFromPorts", t, func() {
		convey.Convey("valid param", func() {
			npuMap := make(map[string]algo.NpuInfo)
			param := superPodParam{
				superPodId: "0", protocol: "test", rack: &RackInfo{RackID: "0"},
				npu: &NpuInfo{PhyId: "0", LevelList: []LevelElement{
					{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				}},
				server: &ServerInfo{
					ServerIndex: "0",
				},
			}
			getEidInfoOrIpFromPorts(npuMap, param)
			convey.So(len(npuMap) == 1, convey.ShouldBeTrue)
		})
		convey.Convey("invalid param", func() {
			npuMap := make(map[string]algo.NpuInfo)
			param := superPodParam{
				superPodId: "0", protocol: "test", rack: &RackInfo{RackID: "0"},
				npu: &NpuInfo{PhyId: "0", LevelList: []LevelElement{
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				}},
				server: &ServerInfo{
					ServerIndex: "0",
				},
			}
			getEidInfoOrIpFromPorts(npuMap, param)
			convey.So(len(npuMap) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("invalid digit", func() {
			npuMap := make(map[string]algo.NpuInfo)
			param := superPodParam{
				superPodId: "0", protocol: "test", rack: &RackInfo{RackID: "0"},
				npu: &NpuInfo{PhyId: "invalid", LevelList: []LevelElement{
					{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				}},
				server: &ServerInfo{
					ServerIndex: "0",
				},
			}
			getEidInfoOrIpFromPorts(npuMap, param)
			convey.So(len(npuMap) == 0, convey.ShouldBeTrue)
		})
	})
}

func TestGenRoceSuperPodLevelPingList(t *testing.T) {
	convey.Convey("test func GenRoceSuperPodLevelPingList", t, func() {
		convey.Convey("return false when gen PingList err", func() {
			detectObj := algo.NewNetDetect("0")
			patch := gomonkey.ApplyMethod(detectObj, "GenPingStrategy",
				func(nd *algo.NetDetect, input map[string]interface{}) map[string]interface{} {
					return nil
				})
			defer patch.Reset()
			ret := GenRoceSuperPodLevelPingList("", nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

func TestIsPureLetter(t *testing.T) {
	convey.Convey("TestIsPureLetter", t, func() {
		// 测试纯字母字符串
		convey.Convey("when_str_is_pure_letter", func() {
			convey.So(isPureLetter("HelloWorld"), convey.ShouldBeTrue)
			convey.So(isPureLetter("abc"), convey.ShouldBeTrue)
			convey.So(isPureLetter("ABC"), convey.ShouldBeTrue)
		})

		// 测试包含数字的字符串
		convey.Convey("when_str_contains_digit", func() {
			convey.So(isPureLetter("Hello1"), convey.ShouldBeFalse)
			convey.So(isPureLetter("a1b2c3"), convey.ShouldBeFalse)
		})

		// 测试包含特殊字符的字符串
		convey.Convey("when_str_contains_special_char", func() {
			convey.So(isPureLetter("Hello@World"), convey.ShouldBeFalse)
			convey.So(isPureLetter("abc!"), convey.ShouldBeFalse)
		})
	})
}

func TestIsPureNumber(t *testing.T) {
	convey.Convey("TestIsPureNumber", t, func() {
		// 测试纯数字字符串
		convey.Convey("when_str_is_pure_number", func() {
			convey.So(isPureNumber("12345"), convey.ShouldEqual, true)
			convey.So(isPureNumber("0"), convey.ShouldEqual, true)
		})

		// 测试包含字母的字符串
		convey.Convey("when_str_contains_letter", func() {
			convey.So(isPureNumber("123abc"), convey.ShouldBeFalse)
			convey.So(isPureNumber("abc123"), convey.ShouldBeFalse)
		})

		// 测试包含特殊字符的字符串
		convey.Convey("when_str_contains_special_char", func() {
			convey.So(isPureNumber("123@45"), convey.ShouldBeFalse)
			convey.So(isPureNumber("123!45"), convey.ShouldBeFalse)
		})
	})
}

func TestGetSuperPodRackLevelNpuMap(t *testing.T) {
	convey.Convey("Test func getSuperPodRackLevelNpuMap", t, func() {
		convey.Convey("invalid super pod info", func() {
			ret := getSuperPodRackLevelNpuMap(nil)
			convey.So(ret == nil, convey.ShouldBeTrue)
		})
		convey.Convey("invalid server numbers", func() {
			superPodInfo := &SuperPodInfo{
				Version:    "A4",
				SuperPodID: "0",
				RackMap: map[string]*RackInfo{
					"0": {RackID: "0", ServerMap: map[string]*ServerInfo{}},
				},
			}
			ret := getSuperPodRackLevelNpuMap(superPodInfo)
			convey.So(ret == nil, convey.ShouldBeTrue)
		})
		convey.Convey("valid param", func() {
			superPodInfo := &SuperPodInfo{
				Version:    "A4",
				SuperPodID: "0",
				RackMap: map[string]*RackInfo{
					"0": {RackID: "0", ServerMap: map[string]*ServerInfo{
						"0": {ServerIndex: "0", NpuMap: map[string]*NpuInfo{
							"0": {PhyId: "0"},
						}},
					}},
				},
			}
			ret := getSuperPodRackLevelNpuMap(superPodInfo)
			convey.So(ret != nil, convey.ShouldBeTrue)
		})
	})
}

func TestGetNpuMapValueInfoUnit(t *testing.T) {
	convey.Convey("TestGetNpuMapValueInfoUnit", t, func() {
		convey.Convey("when rackAndServerIds is empty", func() {
			rackAndServerIds := make([][]string, 0)
			result := getNpuMapValueInfoUnit(rackAndServerIds, 0, "1", 1, "server1")
			expected := algo.NpuInfo{
				RackName:   "",
				SlotName:   "NSlot-1",
				NpuNumber:  1,
				NetPlaneId: "",
				OsName:     "server1",
			}
			convey.So(result, convey.ShouldResemble, expected)
		})
		convey.Convey("when index is out of range", func() {
			rackAndServerIds := [][]string{{"rack1", "rack2"}}
			result := getNpuMapValueInfoUnit(rackAndServerIds, 1, "1", 1, "server1")
			expected := algo.NpuInfo{
				RackName:   "Rack-rack2",
				SlotName:   "NSlot-1",
				NpuNumber:  1,
				NetPlaneId: "",
				OsName:     "server1",
			}
			convey.So(result, convey.ShouldResemble, expected)
		})
		convey.Convey("when all inputs are valid", func() {
			rackAndServerIds := [][]string{{"rack1", "rack2"}}
			result := getNpuMapValueInfoUnit(rackAndServerIds, 0, "1", 1, "server1")
			expected := algo.NpuInfo{
				RackName:   "Rack-rack1",
				SlotName:   "NSlot-1",
				NpuNumber:  1,
				NetPlaneId: "",
				OsName:     "server1",
			}
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestStoreA51D2DNpuFmLink(t *testing.T) {
	convey.Convey("test func storeA51D2DNpuFmLink", t, func() {
		convey.Convey("nil param", func() {
			var fmLink []string
			storeA51D2DNpuFmLink(nil, &fmLink, "", "", "")
			convey.So(len(fmLink) == 0, convey.ShouldEqual, true)
		})
		convey.Convey("correct param", func() {
			fmLink := make([]string, 0)
			param := npuMapParam{
				rackNpuMap: make(map[string]bool),
			}
			storeA51D2DNpuFmLink(&param, &fmLink, "a", "b", "000")
			convey.So(len(fmLink) == 0, convey.ShouldEqual, true)
		})
	})
}

func TestProcessSuperPodJsonWhenReasoningServer(t *testing.T) {
	mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer mockSleep.Reset()
	convey.Convey("Test processSuperPodJson when version 800I-SuperPod-A5-8", t, func() {
		convey.Convey("800I-SuperPod-A5-8", func() {
			mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, &SuperPodInfo{Version: DiagVersionServer})
			defer mockGetSuperpodInfo.Reset()
			patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer patch.Reset()
			configmap, MeshInfo, linkInfo := processSuperPodJson("Path", "")
			convey.So(configmap != nil, convey.ShouldBeTrue)
			convey.So(MeshInfo, convey.ShouldBeNil)
			convey.So(linkInfo, convey.ShouldBeNil)
		})
	})
}

func TestGetReasoningServerNpuInfo(t *testing.T) {
	convey.Convey("TestGetReasoningServerNpuInfo", t, func() {
		convey.Convey("test level != 1", func() {
			npuMap := make(map[string]algo.NpuInfo)
			npu := &NpuInfo{
				PhyId: "1",
				LevelList: []LevelElement{
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				},
			}
			getReasoningServerNpuInfo(npuMap, 0, npu, "0", "0")
			convey.So(len(npuMap) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("test level == 1, phyid != number", func() {
			npuMap := make(map[string]algo.NpuInfo)
			npu := &NpuInfo{
				PhyId: "null",
				LevelList: []LevelElement{
					{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				},
			}
			getReasoningServerNpuInfo(npuMap, 0, npu, "0", "0")
			convey.So(len(npuMap) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("test level == 1, phyid == number", func() {
			npuMap := make(map[string]algo.NpuInfo)
			npu := &NpuInfo{
				PhyId: "0",
				LevelList: []LevelElement{
					{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "192.168.1.1"}}},
				},
			}
			getReasoningServerNpuInfo(npuMap, 0, npu, "0", "0")
			convey.So(len(npuMap) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestGetReasoningServerSuperPodNpuMap(t *testing.T) {
	convey.Convey("TestGetReasoningServerSuperPodNpuMap", t, func() {
		convey.Convey("test empty rack", func() {
			superPodInfo := &SuperPodInfo{RackMap: make(map[string]*RackInfo)}
			ret := getReasoningServerSuperPodNpuMap(superPodInfo)
			convey.So(ret == nil, convey.ShouldBeTrue)
		})
		convey.Convey("test empty server map", func() {
			superPodInfo := &SuperPodInfo{RackMap: map[string]*RackInfo{
				"0": {ServerMap: make(map[string]*ServerInfo)},
			}}
			ret := getReasoningServerSuperPodNpuMap(superPodInfo)
			convey.So(ret == nil, convey.ShouldBeTrue)
		})
		convey.Convey("test invalid server id", func() {
			superPodInfo := &SuperPodInfo{RackMap: map[string]*RackInfo{
				"0": {ServerMap: map[string]*ServerInfo{
					"a": {ServerIndex: "a"},
					"b": {ServerIndex: "b"},
				}},
			}}
			ret := getReasoningServerSuperPodNpuMap(superPodInfo)
			convey.So(len(ret) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("test empty npu data", func() {
			superPodInfo := &SuperPodInfo{RackMap: map[string]*RackInfo{
				"0": {ServerMap: map[string]*ServerInfo{
					"1": {ServerIndex: "1"},
					"2": {ServerIndex: "2"},
				}},
			}}
			patch := gomonkey.ApplyFunc(getReasoningServerNpuInfo, func(npuEidMap map[string]algo.NpuInfo,
				serverIndex int, npu *NpuInfo, serverId string, superPodId string) {
				return
			})
			defer patch.Reset()
			ret := getReasoningServerSuperPodNpuMap(superPodInfo)
			convey.So(len(ret) == 0, convey.ShouldBeTrue)
		})
	})
}

func TestHandleReasoningServer(t *testing.T) {

}

func TestGetA51D2DServerLevelInfo(t *testing.T) {
	convey.Convey("test func getA51D2DServerLevelInfo", t, func() {
		mu1 := sync.Mutex{}
		mu2 := sync.Mutex{}
		var called1, called2 bool
		patch := gomonkey.ApplyFunc(getReasoningServerNpuLinkPath,
			func(npuNetPlanePaths map[string][]string, serverIds []int, serverMap map[string]*ServerInfo) {
				mu1.Lock()
				called1 = true
				mu1.Unlock()
			})
		defer patch.Reset()
		patch2 := gomonkey.ApplyFunc(getA51D2DNpuLinkPath,
			func(npuNetPlanePaths map[string][]string, npu *NpuInfo, rackId string, typeStr string) {
				mu2.Lock()
				called2 = true
				mu2.Unlock()
			})
		defer patch2.Reset()
		convey.Convey("reasoningServer", func() {
			paths := make(map[string][]string)
			rack := &RackInfo{
				RackID: "1",
				ServerMap: map[string]*ServerInfo{
					"test": {ServerIndex: "test"},
					"4":    {ServerIndex: "4"},
				},
			}
			getA51D2DServerLevelInfo(paths, rack, "reasoningServer")
			convey.So(called1, convey.ShouldBeTrue)
		})
		convey.Convey("1D2D", func() {
			paths := make(map[string][]string)
			rack := &RackInfo{
				RackID: "1",
				ServerMap: map[string]*ServerInfo{
					"4": {ServerIndex: "4", NpuMap: map[string]*NpuInfo{"0": {LevelList: nil},
						"1": {LevelList: []LevelElement{{NetLayer: 1}}}}},
					"0": nil,
				},
			}
			getA51D2DServerLevelInfo(paths, rack, "")
			convey.So(called2, convey.ShouldBeTrue)
		})
	})
}

func TestGetReasoningServerNpuLinkPath(t *testing.T) {
	convey.Convey("TestGetReasoningServerNpuLinkPath", t, func() {
		paths := make(map[string][]string)
		serverIds := []int{1}
		serverMap := map[string]*ServerInfo{
			"1": {ServerIndex: "1",
				NpuMap: map[string]*NpuInfo{
					"1": {PhyId: "1", LevelList: []LevelElement{
						{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "addr1"}, {Addr: "addr2"}}},
					}},
				}},
		}
		convey.Convey("empty npu info", func() {
			serverMap := map[string]*ServerInfo{
				"1": {ServerIndex: "1",
					NpuMap: map[string]*NpuInfo{}},
			}
			getReasoningServerNpuLinkPath(paths, serverIds, serverMap)
			convey.So(len(paths) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("valid npu info", func() {
			patch := gomonkey.ApplyFunc(getReasoningServerNpuLinkPathStr,
				func(npuNetPlanePaths map[string][]string, npu *NpuInfo, serverIndex int) { return })
			defer patch.Reset()
			getReasoningServerNpuLinkPath(paths, serverIds, serverMap)
			convey.So(len(paths) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("invalid level", func() {
			serverMap2 := map[string]*ServerInfo{
				"1": {ServerIndex: "1",
					NpuMap: map[string]*NpuInfo{
						"1": {PhyId: "1", LevelList: []LevelElement{
							{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1"}, {Addr: "addr2"}}},
						}},
					}},
			}
			getReasoningServerNpuLinkPath(paths, serverIds, serverMap2)
			convey.So(len(paths) == 0, convey.ShouldBeTrue)
		})
		convey.Convey("valid level", func() {
			getReasoningServerNpuLinkPath(paths, serverIds, serverMap)
			convey.So(len(paths) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestGetReasoningServerNpuLinkPathStr(t *testing.T) {

}

func TestGetReasoningServerNpuLinkPathStrPartTwo(t *testing.T) {

}
