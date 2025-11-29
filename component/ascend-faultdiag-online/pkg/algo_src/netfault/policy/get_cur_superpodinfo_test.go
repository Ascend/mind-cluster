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
	"errors"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

func TestIsAlphanumeric(t *testing.T) {

}

func TestContainsElement(t *testing.T) {

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

func TestCheckDiffConfig(t *testing.T) {

}

func TestSpliceSuperPodFilePath(t *testing.T) {

}

func TestGetCurrentSuperPodInfoWhenSuperPodPathInvalid(t *testing.T) {

}

// TestGetCurrentSuperPodInfo test for func getCurrentSuperPodInfo
func TestGetCurrentSuperPodInfoWhenEmptyConfigMapCauseAlgoPingListInputInvalid(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test getCurrentSuperPodInfo", t, func() {
		convey.Convey("should return nil when empty configMap cause algoPingListInput invalid", func() {
			superPodPath := "/a/b/super-pod-0/"
			patch := gomonkey.ApplyFunc(readConfigMap,
				func(configFilePath string) *SuperPodInfo {
					output := &SuperPodInfo{
						SuperPodID: "1",
					}
					return output
				})
			defer patch.Reset()
			controllerflags.IsControllerExited.SetState(false)
			actualReturnValue1, actualReturnValue2 := getCurrentSuperPodInfo(superPodPath, nil)
			convey.So(actualReturnValue1, convey.ShouldBeNil)
			convey.So(actualReturnValue2, convey.ShouldBeNil)
		})
	})
}

func TestGetCurrentSuperPodInfoWhenEmptySpliceAlgorithmInputCauseAlgoPingListInputInvalid(t *testing.T) {

}

func TestGetCurrentSuperPodInfoWhenEmptyAlgoPingListInputCauseJsonPingListInvalid(t *testing.T) {

}

func TestGetTargetSuperPodNpuMapWhenInvalid(t *testing.T) {

}

func TestGetTargetSuperPodNpuMapWhenValid(t *testing.T) {

}

func TestSetCallAlgorithmParamInfo(t *testing.T) {

}

func TestGetWorkMapping(t *testing.T) {

}

func TestProcessSuperPodJsonWhenVersionInfoInvalid(t *testing.T) {

}

func TestProcessSuperPodJsonWhenVersionA5(t *testing.T) {

}

func TestProcessSuperPodJsonWhenVersionA3(t *testing.T) {

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

func TestParseA5SeverLevelTopologyFile(t *testing.T) {
	convey.Convey("test func parseA5ServerLevelTopologyFile", t, func() {
		convey.Convey("should return nil when allFiles nil", func() {
			allFile := make([]string, 0)
			param := parseTopoParam{topoServerDirPath: allFile, rackAndServerInfo: make([][]string, 0),
				superPodInfo: nil, superPodRackNpuMap: nil, typeStr: "1D", superPodPath: ""}
			ret1, ret2 := parseA5ServerLevelTopologyFile(&param)
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
			ret1, ret2 := parseA5ServerLevelTopologyFile(&param)
			convey.So(ret1, convey.ShouldBeEmpty)
			convey.So(ret2, convey.ShouldBeEmpty)
		})
		convey.Convey("should return nil when read file err", func() {
			allFile := make([]string, 1)
			mockRead := gomonkey.ApplyFuncReturn(os.ReadFile, []byte{}, errors.New("err"))
			defer mockRead.Reset()
			param := parseTopoParam{topoServerDirPath: allFile, rackAndServerInfo: make([][]string, 0),
				superPodInfo: nil, superPodRackNpuMap: nil, typeStr: "1D", superPodPath: ""}
			ret1, ret2 := parseA5ServerLevelTopologyFile(&param)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
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

func TestGetOneTopoFilePath(t *testing.T) {

}

func TestGetCurSuperPod1DNpuInfo(t *testing.T) {

}

func TestGetNpuLinkPath(t *testing.T) {

}

func TestCheckIfNew1DTrue(t *testing.T) {

}

func TestGetNetWorkTypePart0(t *testing.T) {

}

var str2D = `{"hardwareType": "Atlas 950 SuperPod 2D"}`
var str1D = `{"hardwareType": "Atlas 950 SuperPod 1D"}`
var strErr = `{ "hardwareType": ""}`
var data2D = []byte(str2D)
var data1D = []byte(str1D)
var dataErr = []byte(strErr)
var dataTest = [][]byte{dataErr, data1D, data2D}

func TestGetNetWorkTypePart1(t *testing.T) {

}
