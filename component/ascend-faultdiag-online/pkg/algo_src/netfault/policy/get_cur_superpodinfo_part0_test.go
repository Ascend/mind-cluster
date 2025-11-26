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

// Package policy is used for processing superpod information
package policy

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

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

func TestStoreA51D2DNpuFmLinkAndNpuEidMapInfo1(t *testing.T) {
	convey.Convey("test func storeA51D2DNpuFmLinkAndNpuEidMapInfo1", t, func() {
		infoMap := make(map[string]algo.NpuInfo)
		link := make([]string, 0)
		ids := [][]string{{"0"}}
		convey.Convey("should return empty when local id <0", func() {
			param := npuMapParam{serverTopology: &RackTopology{EdgeList: []Edge{{LocalA: -1}}}}
			ret1, ret2 := storeA51D2DNpuFmLinkAndNpuEidMapInfo(0, make([][]string, 0), &param)
			convey.So(ret1, convey.ShouldResemble, infoMap)
			convey.So(ret2, convey.ShouldResemble, link)
		})
		convey.Convey("should return empty when level !=0", func() {
			param := npuMapParam{serverTopology: &RackTopology{EdgeList: []Edge{{LocalA: 0, NetLayer: 1}}}}
			ret1, ret2 := storeA51D2DNpuFmLinkAndNpuEidMapInfo(0, make([][]string, 0), &param)
			convey.So(ret1, convey.ShouldResemble, infoMap)
			convey.So(ret2, convey.ShouldResemble, link)
		})
		convey.Convey("should return empty when server id is empty", func() {
			param := npuMapParam{serverTopology: &RackTopology{EdgeList: []Edge{{LocalA: 0, NetLayer: 0,
				LinkType: "PEER2PEER"}}},
				superPodInfo: &SuperPodInfo{RackMap: map[string]*RackInfo{"0": {RackID: "0"}}}}
			patch1 := gomonkey.ApplyFuncReturn(getNpuServerIdFromRackMap, "")
			defer patch1.Reset()
			ret1, ret2 := storeA51D2DNpuFmLinkAndNpuEidMapInfo(0, ids, &param)
			convey.So(ret1, convey.ShouldResemble, infoMap)
			convey.So(ret2, convey.ShouldResemble, link)
		})
		convey.Convey("should return empty when eid is empty", func() {
			param := npuMapParam{serverTopology: &RackTopology{EdgeList: []Edge{{LocalA: 0, NetLayer: 0,
				LinkType: "PEER2PEER"}}},
				superPodInfo: &SuperPodInfo{RackMap: map[string]*RackInfo{"0": {RackID: "0"}}}}
			patch1 := gomonkey.ApplyFuncReturn(getNpuServerIdFromRackMap, "1")
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFuncReturn(findEid, "")
			defer patch2.Reset()
			ret1, ret2 := storeA51D2DNpuFmLinkAndNpuEidMapInfo(0, ids, &param)
			convey.So(ret1, convey.ShouldResemble, infoMap)
			convey.So(ret2, convey.ShouldResemble, link)
		})
	})
}

func TestStoreA51D2DNpuFmLinkAndNpuEidMapInfo2(t *testing.T) {
	convey.Convey("test func storeA51D2DNpuFmLinkAndNpuEidMapInfo2", t, func() {
		link := make([]string, 0)
		ids := [][]string{{"0"}}
		convey.Convey("should return normal", func() {
			param := npuMapParam{serverTopology: &RackTopology{EdgeList: []Edge{{LocalA: 0, NetLayer: 0,
				LinkType: "PEER2PEER"}}},
				superPodInfo: &SuperPodInfo{RackMap: map[string]*RackInfo{"0": {RackID: "0"}}}}
			patch1 := gomonkey.ApplyFuncReturn(getNpuServerIdFromRackMap, "1")
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFuncReturn(findEid, "1")
			defer patch2.Reset()
			patch3 := gomonkey.ApplyFuncReturn(getNpuMapValueInfoUnit, algo.NpuInfo{})
			defer patch3.Reset()
			ret1, ret2 := storeA51D2DNpuFmLinkAndNpuEidMapInfo(0, ids, &param)
			convey.So(ret1, convey.ShouldResemble, map[string]algo.NpuInfo{"1": {}})
			convey.So(ret2, convey.ShouldResemble, link)
		})
	})
}

func TestFindEid(t *testing.T) {
	convey.Convey("test func findEid", t, func() {
		rackMap := map[string]*RackInfo{
			"rack1": {
				RackID: "1",
				ServerMap: map[string]*ServerInfo{
					"1": {ServerIndex: "1", NpuMap: map[string]*NpuInfo{"1": {
						PhyId: "1",
						LevelList: []LevelElement{
							{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", Ports: []string{"0/1"}}}},
						},
					}}},
				},
			},
		}
		convey.Convey("when find eid success", func() {
			eid := findEid("1", 1, []string{"0/1"}, rackMap["rack1"])
			convey.So(eid, convey.ShouldResemble, "addr1")
		})
		convey.Convey("when find eid failed", func() {
			eid := findEid("1", 1, []string{"0/2"}, rackMap["rack1"])
			convey.So(eid, convey.ShouldResemble, "")
		})
	})
}

func TestParseA5SeverLevelTopologyFile(t *testing.T) {
	convey.Convey("test func parseA5SeverLevelTopologyFile", t, func() {
		convey.Convey("should return nil when allFiles nil", func() {
			allFile := make([]string, 0)
			param := parseTopoParam{topoServerDirPath: allFile, rackAndServerInfo: make([][]string, 0),
				superPodInfo: nil, superPodRackNpuMap: nil, typeStr: "1D", superPodPath: ""}
			ret1, ret2 := parseA5SeverLevelTopologyFile(&param)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})
		convey.Convey("should retry when read file err", func() {
			allFile := make([]string, 1)
			mockRead := gomonkey.ApplyFuncReturn(os.ReadFile, []byte{}, os.ErrNotExist)
			defer mockRead.Reset()
			mockWait := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer mockWait.Reset()
			param := parseTopoParam{topoServerDirPath: allFile, rackAndServerInfo: make([][]string, 0),
				superPodInfo: nil, superPodRackNpuMap: nil, typeStr: "1D", superPodPath: ""}
			ret1, ret2 := parseA5SeverLevelTopologyFile(&param)
			convey.So(ret1, convey.ShouldBeEmpty)
			convey.So(ret2, convey.ShouldBeEmpty)
		})
		convey.Convey("should return nil when read file err", func() {
			allFile := make([]string, 1)
			mockRead := gomonkey.ApplyFuncReturn(os.ReadFile, []byte{}, errors.New("err"))
			defer mockRead.Reset()
			param := parseTopoParam{topoServerDirPath: allFile, rackAndServerInfo: make([][]string, 0),
				superPodInfo: nil, superPodRackNpuMap: nil, typeStr: "1D", superPodPath: ""}
			ret1, ret2 := parseA5SeverLevelTopologyFile(&param)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})
	})
}

func TestGetA51D2DNpuLinkPath(t *testing.T) {
	convey.Convey("test func getA51D2DNpuLinkPath", t, func() {
		convey.Convey("when level 1 exist", func() {
			npuNetPlanePaths := make(map[string][]string)
			npu := &NpuInfo{
				PhyId: "1",
				LevelList: []LevelElement{
					{NetLayer: 1, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
				},
			}
			getA51D2DNpuLinkPath(npuNetPlanePaths, npu, "1", "1D")
			expectVal := map[string][]string{
				"1": {"NA.L2-LogicPort0:0#Rack-1.L1-LogicPort0:0#Rack-1.NSlot-0:0#NPU-1.addr1:0"},
				"2": {"NA.L2-LogicPort1:0#Rack-1.L1-LogicPort1:0#Rack-1.NSlot-0:0#NPU-1.addr2:0"},
			}
			convey.So(npuNetPlanePaths, convey.ShouldResemble, expectVal)
		})
		convey.Convey("when level 1 not exist", func() {
			npuNetPlanePaths := make(map[string][]string)
			npu := &NpuInfo{
				PhyId: "1",
				LevelList: []LevelElement{
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
				},
			}
			getA51D2DNpuLinkPath(npuNetPlanePaths, npu, "1", "1D")
			expectVal := map[string][]string{}
			convey.So(npuNetPlanePaths, convey.ShouldResemble, expectVal)
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

func TestGetA51D2DSuperPodNpuLinkPath(t *testing.T) {
	convey.Convey("Test func getA51D2DSuperPodNpuLinkPath", t, func() {
		convey.Convey("should return nil when superPodInfo nil", func() {
			ret := getA51D2DSuperPodNpuLinkPath(&SuperPodInfo{}, "1D")
			convey.So(ret, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when rackMap nil", func() {
			rackMap := make(map[string]*RackInfo)
			ret := getA51D2DSuperPodNpuLinkPath(&SuperPodInfo{RackMap: rackMap}, "1D")
			convey.So(ret, convey.ShouldBeNil)
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

func TestGetA5CurSuperPod1D2DNpuInfo(t *testing.T) {
	convey.Convey("Test func GetA5CurSuperPod1D2DNpuInfo", t, func() {
		convey.Convey("should return nil when no rackNums", func() {
			s := &SuperPodInfo{}
			ret1, ret2, ret3 := GetA5CurSuperPod1D2DNpuInfo("", s)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
			convey.So(ret3, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when get npuMap err", func() {
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": {RackID: "1"}}}
			ret1, ret2, ret3 := GetA5CurSuperPod1D2DNpuInfo("", s)
			patch := gomonkey.ApplyFuncReturn(getSuperPodRackLevelNpuMap, 0)
			defer patch.Reset()
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
			convey.So(ret3, convey.ShouldBeNil)
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

func TestReadConfigFromFile(t *testing.T) {
	convey.Convey("TestReadConfigFromFile", t, func() {
		fileContent := []byte(`
supperssedPeriod=0
networkType=1
pingType=0
pingTimes=5
pingInterval=1
period=10
netFault=on
`)
		targetKeys := []string{"networkType", "pingType", "pingTimes", "pingInterval", "suppressedPeriod", "period"}
		result := ReadConfigFromFile(fileContent, targetKeys)

		convey.So(result, convey.ShouldNotBeEmpty)
	})
}

func TestCheckCurSuperPodConfigSwitch(t *testing.T) {
	convey.Convey("test CheckCurSuperPodConfigSwitch", t, func() {
		res := CheckCurSuperPodConfigSwitch(".")
		convey.So(res, convey.ShouldBeFalse)
		err := createTmpConfigFile()
		convey.So(err, convey.ShouldBeNil)
		defer removeTmpConfigFile()
		res = CheckCurSuperPodConfigSwitch(".")
		convey.So(res, convey.ShouldBeTrue)
	})
}

func createTmpConfigFile() error {
	configPath := "./cathelper.conf"
	fileContent := `
supperssedPeriod=0
networkType=1
pingType=0
pingTimes=5
pingInterval=1
period=10
netFault=on
`
	var fileMode0644 os.FileMode = 0644
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, fileMode0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fileContent)
	return err
}

func removeTmpConfigFile() {
	configPath := "./cathelper.conf"
	err := os.Remove(configPath)
	if err != nil {
		hwlog.RunLog.Errorf("remove temp config file %s failed: %v", configPath, err)
	}
}
