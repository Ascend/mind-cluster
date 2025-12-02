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
	"fmt"

	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

// TestIsAlphanumeric test for func isAlphanumeric
func TestIsAlphanumeric(t *testing.T) {
	convey.Convey("Test isAlphanumeric", t, func() {
		convey.Convey("should return false when request is invalid", func() {
			s := "abc123$#"
			ret := isAlphanumeric(s)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when request is valid", func() {
			s := "a1b2c3"
			ret := isAlphanumeric(s)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestContainsElement test for func containsElement
func TestContainsElement(t *testing.T) {
	convey.Convey("Test containsElement", t, func() {
		convey.Convey("should return false when request is invalid", func() {
			slice := []string{"abc123$#"}
			str := "1"
			ret := containsElement(slice, str)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when request is valid", func() {
			slice := []string{"abc123$#", "1"}
			str := "1"
			ret := containsElement(slice, str)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestContainsElement test for func containsElement
func TestCheckDiffConfig(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test checkDiffConfig", t, func() {
		convey.Convey("should return false when config file is not existed", func() {
			superPodFilePath := "1"
			ret := CheckDiffConfig(superPodFilePath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return valid value when config file is existed", func() {
			controllerflags.IsControllerStarted.SetState(true)
			superPodFilePath := "./"
			configFile, err := os.Create(superPodFilePath + "cathelper.conf")
			if err != nil {
				return
			}
			defer configFile.Close()
			defer configFile.Chmod(0600) //文件权限
			ret := CheckDiffConfig(superPodFilePath)
			expectReturnValue := map[string]interface{}(nil)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
			controllerflags.IsControllerStarted.SetState(false)
			err = os.Remove(superPodFilePath + "cathelper.conf")
			if err != nil {
				return
			}
		})
	})
}

func TestSpliceSuperPodFilePath(t *testing.T) {
	convey.Convey("Test spliceSuperPodFilePath", t, func() {
		expectedSuperPodPath := "/xx/xx/super-pod-0/super-pod-0.json"
		superPodPath := "/xx/xx/super-pod-0/"
		ret := spliceSuperPodFilePath(superPodPath)
		convey.So(ret, convey.ShouldResemble, expectedSuperPodPath)
	})
}

// TestGetCurrentSuperPodInfo test for func getCurrentSuperPodInfo
func TestGetCurrentSuperPodInfoWhenSuperPodPathInvalid(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test getCurrentSuperPodInfo", t, func() {
		convey.Convey("should return nil when superPodPath invalid", func() {
			superPodPath := ""
			controllerflags.IsControllerExited.SetState(false)
			actualReturnValue1, actualReturnValue2 := getCurrentSuperPodInfo(superPodPath, nil)
			convey.So(actualReturnValue1, convey.ShouldBeNil)
			convey.So(actualReturnValue2, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when empty superPodJsonFile cause configMap invalid", func() {
			superPodPath := "/a/b/super-pod-0/"
			patch := gomonkey.ApplyFunc(readConfigMap,
				func(configFilePath string) *SuperPodInfo {
					return nil
				})
			defer patch.Reset()
			controllerflags.IsControllerExited.SetState(false)
			actualReturnValue1, actualReturnValue2 := getCurrentSuperPodInfo(superPodPath, nil)
			convey.So(actualReturnValue1, convey.ShouldBeNil)
			convey.So(actualReturnValue2, convey.ShouldBeNil)
		})
	})
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

// TestGetCurrentSuperPodInfo test for func getCurrentSuperPodInfo
func TestGetCurrentSuperPodInfoWhenEmptySpliceAlgorithmInputCauseAlgoPingListInputInvalid(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test getCurrentSuperPodInfo", t, func() {
		convey.Convey("should return nil when empty algoPingListInput cause algoPingListInput invalid", func() {
			superPodPath := "/a/b/super-pod-0/"
			patch := gomonkey.ApplyFunc(readConfigMap,
				func(configFilePath string) *SuperPodInfo {
					output := &SuperPodInfo{
						SuperPodID: "1",
					}
					return output
				})
			defer patch.Reset()
			patch2 := gomonkey.ApplyFunc(spliceAlgorithmInput,
				func(npu2DFullMesh []string, npuOutOfRackPath map[string][]string) map[string]interface{} {
					return nil
				})
			defer patch2.Reset()
			controllerflags.IsControllerExited.SetState(false)
			actualReturnValue1, actualReturnValue2 := getCurrentSuperPodInfo(superPodPath, nil)
			convey.So(actualReturnValue1, convey.ShouldBeNil)
			convey.So(actualReturnValue2, convey.ShouldBeNil)
		})
	})
}

// TestGetCurrentSuperPodInfo test for func getCurrentSuperPodInfo
func TestGetCurrentSuperPodInfoWhenEmptyAlgoPingListInputCauseJsonPingListInvalid(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test getCurrentSuperPodInfo", t, func() {
		convey.Convey("should return valid value", func() {
			superPodPath := "/a/b/super-pod-0/"
			patch := gomonkey.ApplyFunc(readConfigMap,
				func(configFilePath string) *SuperPodInfo {
					return &SuperPodInfo{SuperPodID: "1"}
				})
			defer patch.Reset()
			patch2 := gomonkey.ApplyFunc(spliceAlgorithmInput,
				func(npu2DFullMesh []string, npuOutOfRackPath map[string][]string) map[string]interface{} {
					resultArgMap := make(map[string]interface{})
					npu2DFullMesh1 := []string{"1"}
					resultArgMap["npu_npu"] = npu2DFullMesh1
					npuNetPlanes := make(map[string]interface{})
					npuNetPlanes["netplane_0"] = []string{"1", "22"}
					resultArgMap["npu_netplane"] = npuNetPlanes
					return resultArgMap
				})
			defer patch2.Reset()
			detectObj := algo.NewNetDetect("0")
			patch3 := gomonkey.ApplyMethod(detectObj, "GenPingStrategy",
				func(nd *algo.NetDetect, input map[string]interface{}) map[string]interface{} {
					return nil
				})
			defer patch3.Reset()
			controllerflags.IsControllerExited.SetState(false)
			actualReturnValue1, actualReturnValue2 := getCurrentSuperPodInfo(superPodPath, detectObj)
			convey.So(actualReturnValue1, convey.ShouldBeNil)
			convey.So(actualReturnValue2, convey.ShouldBeNil)
		})
	})
}

func TestGetTargetSuperPodNpuMapWhenInvalid(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
	defer mockTime.Reset()
	controllerflags.IsControllerExited.SetState(false)

	convey.Convey("test GetTargetSuperPodNpuMap when invalid", t, func() {
		convey.Convey("should return false and nil when get superPodInfo err", func() {
			mockReadConfigMap := gomonkey.ApplyFunc(readConfigMap,
				func(_ string) *SuperPodInfo {
					return nil
				})
			defer mockReadConfigMap.Reset()

			flag, npuInfomap := GetTargetSuperPodNpuMap("", 0)
			convey.So(flag, convey.ShouldBeFalse)
			convey.So(npuInfomap == nil, convey.ShouldBeTrue)
		})
		convey.Convey("should return false when MAXretry", func() {
			configMap := &SuperPodInfo{Version: DiagVersionA5}
			mockReadTopoFromSuperPodFile := gomonkey.ApplyFuncReturn(readConfigMap, configMap)
			defer mockReadTopoFromSuperPodFile.Reset()
			flag, _ := GetTargetSuperPodNpuMap("", 0)
			convey.So(flag, convey.ShouldBeFalse)
		})

		convey.Convey("should return nil when get Veriosn Info err", func() {
			configMap := &SuperPodInfo{}
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap, configMap)
			defer mockReadConfigMap.Reset()
			flag, _ := GetTargetSuperPodNpuMap("", 0)
			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

func TestGetTargetSuperPodNpuMapWhenValid(t *testing.T) {
	controllerflags.IsControllerExited.SetState(false)
	convey.Convey("Test GetTargetSuperPodNpuMap When Valid", t, func() {
		convey.Convey("should return true and call right func when A3", func() {
			configMap := &SuperPodInfo{Version: DiagVersionA3}
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap, configMap)
			defer mockReadConfigMap.Reset()
			mockGetInfoA3 := gomonkey.ApplyFuncReturn(GetCurSuperPodInfoFromMapA3,
				[]string{"0"}, map[string][]string{"1": {"1"}})
			defer mockGetInfoA3.Reset()
			mockWait := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer mockWait.Reset()
			flag, _ := GetTargetSuperPodNpuMap("", 0)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("should return true and call right func when A5", func() {
			configMap := &SuperPodInfo{Version: DiagVersionA5}
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap, configMap)
			defer mockReadConfigMap.Reset()
			mockGetInfoA5 := gomonkey.ApplyFuncReturn(handleA5NpuMapInfo, nil, true)
			defer mockGetInfoA5.Reset()
			mockWait := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer mockWait.Reset()
			flag, _ := GetTargetSuperPodNpuMap("", 0)
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}

func TestSetCallAlgorithmParamInfo(t *testing.T) {
	convey.Convey("Test setCallAlgorithmParamInfo", t, func() {
		callAlgorithmParam := make(map[string]interface{})
		callAlgorithmParam["serverIdMap"] = make(map[string]string)
		convey.Convey("should return error when input nil", func() {
			err := SetCallAlgorithmParamInfo(0, "Path", nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error when read err", func() {
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap, nil)
			defer mockReadConfigMap.Reset()
			mockSleep := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer mockSleep.Reset()
			err := SetCallAlgorithmParamInfo(0, "Path", callAlgorithmParam)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error when version err", func() {
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap, &SuperPodInfo{Version: "A9"})
			defer mockReadConfigMap.Reset()
			err := SetCallAlgorithmParamInfo(0, "Path", callAlgorithmParam)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should call func and set version when valid", func() {
			mockReadConfigMap := gomonkey.ApplyFuncReturn(readConfigMap,
				&SuperPodInfo{Version: DiagVersionA3})
			defer mockReadConfigMap.Reset()
			mockGetWorkMapping := gomonkey.ApplyFuncReturn(getWorKMapping, nil)
			defer mockGetWorkMapping.Reset()
			mockWait := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer mockWait.Reset()
			err := SetCallAlgorithmParamInfo(0, "Path", callAlgorithmParam)
			convey.So(err, convey.ShouldBeNil)
			convey.So(callAlgorithmParam[NpuType], convey.ShouldEqual, DiagVersionA3)
		})
	})
}

func TestGetWorkMapping(t *testing.T) {
	convey.Convey("Test getWorKMapping", t, func() {
		callAlgorithmParam := make(map[string]interface{})
		callAlgorithmParam["serverIdMap"] = make(map[string]string)

		convey.Convey("should return err when input invalid", func() {
			err1 := getWorKMapping(nil, nil)
			convey.So(err1, convey.ShouldNotBeNil)
			err2 := getWorKMapping(map[string]interface{}{}, &SuperPodInfo{NodeDeviceMap: nil})
			convey.So(err2, convey.ShouldNotBeNil)
		})
		convey.Convey("should return err when get work INFo err", func() {
			superPodInfos := map[string]*SuperPodInfo{
				"test1": {Version: DiagVersionA3, NodeDeviceMap: nil},
				"test2": {Version: DiagVersionA3, NodeDeviceMap: map[string]*NodeDevice{
					"": {NodeName: "", ServerID: "1"}},
				},
				"test3": {Version: DiagVersionA3, NodeDeviceMap: map[string]*NodeDevice{
					"work1": {NodeName: "work1"}},
				},
				"test4": {Version: DiagVersionA3, NodeDeviceMap: map[string]*NodeDevice{
					"work1": {NodeName: "work1", ServerID: ""}},
				},
			}
			var err error
			for _, superPodInfo := range superPodInfos {
				err = getWorKMapping(callAlgorithmParam, superPodInfo)
				convey.So(err, convey.ShouldNotBeNil)
			}
		})
		convey.Convey("should set Mapping when valid", func() {
			superPodInfo := &SuperPodInfo{
				Version: DiagVersionA3,
				NodeDeviceMap: map[string]*NodeDevice{
					"1": {NodeName: "work1", ServerID: "1"},
					"2": {NodeName: "work2", ServerID: "2"},
				},
			}

			expectCallAlgoParam := make(map[string]interface{})
			expectCallAlgoParam["serverIdMap"] = map[string]string{"1": "work1", "2": "work2"}

			err := getWorKMapping(callAlgorithmParam, superPodInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(callAlgorithmParam, convey.ShouldResemble, expectCallAlgoParam)
		})
	})
}

// TestProcessSuperPodJsonWhenInvalid test for func processSuperPodJson when Invalid
func TestProcessSuperPodJsonWhenVersionInfoInvalid(t *testing.T) {
	convey.Convey("test processSuperPodJson when version valid", t, func() {
		mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
		defer mockSleep.Reset()
		convey.Convey("should return nil when read configMap MAX retry", func() {
			mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, nil)
			defer mockGetSuperpodInfo.Reset()
			patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer patch.Reset()
			configmap, fullMesh, linkPath := processSuperPodJson("Path", "")
			convey.So(configmap, convey.ShouldBeNil)
			convey.So(fullMesh, convey.ShouldBeNil)
			convey.So(linkPath, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when get Version info err", func() {
			testConfigmap := map[int]*SuperPodInfo{
				0: {Version: ""},
				1: {Version: "A8"},
			}
			for _, configmapRet := range testConfigmap {
				patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
				mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, configmapRet)
				configmap, fullMesh, linkPath := processSuperPodJson("Path", "")
				convey.So(configmap, convey.ShouldBeNil)
				convey.So(fullMesh, convey.ShouldBeNil)
				convey.So(linkPath, convey.ShouldBeNil)
				mockGetSuperpodInfo.Reset()
				patch.Reset()
			}
		})
	})
}

// TestProcessSuperPodJsonWhenVersionA5
func TestProcessSuperPodJsonWhenVersionA5(t *testing.T) {
	mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer mockSleep.Reset()
	convey.Convey("Test processSuperPodJson when Version A5 but not 1D 2D", t, func() {
		convey.Convey("should call correct func and return value", func() {
			mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, &SuperPodInfo{Version: DiagVersionA5})
			defer mockGetSuperpodInfo.Reset()
			mockGetType := gomonkey.ApplyFuncReturn(getNetWorkType, "0D00")
			defer mockGetType.Reset()
			patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer patch.Reset()
			configmap, MeshInfo, linkInfo := processSuperPodJson("Path", "")
			convey.So(configmap, convey.ShouldBeNil)
			convey.So(MeshInfo, convey.ShouldBeNil)
			convey.So(linkInfo, convey.ShouldBeNil)
		})

		convey.Convey("should return call Get1D2DInfo when 1D or 2D", func() {
			mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, &SuperPodInfo{Version: DiagVersionA5})
			defer mockGetSuperpodInfo.Reset()
			mockGetType := gomonkey.ApplyFuncReturn(getNetWorkType, "2D")
			defer mockGetType.Reset()
			patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer patch.Reset()
			call := false
			mockGet1D2DInfo := gomonkey.ApplyFunc(GetA5CurSuperPod1D2DNpuInfo, func(_ string, _ *SuperPodInfo) (
				[]string, map[string][]string, map[string]algo.NpuInfo) {
				call = true
				return nil, nil, nil
			})

			defer mockGet1D2DInfo.Reset()
			processSuperPodJson("Path", "")
			convey.So(call, convey.ShouldBeTrue)
		})
	})
}

func TestProcessSuperPodJsonWhenVersionA3(t *testing.T) {
	mockSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer mockSleep.Reset()
	convey.Convey("Test processSuperPodJson when Version A3", t, func() {
		convey.Convey("should call correct func and return value", func() {
			fullMesh := []string{"1"}
			linkPath := map[string][]string{"1": {"1"}}
			configMapA3 := &SuperPodInfo{Version: DiagVersionA3}
			mockGetSuperpodInfo := gomonkey.ApplyFuncReturn(readConfigMap, configMapA3)
			defer mockGetSuperpodInfo.Reset()
			patch := gomonkey.ApplyFuncReturn(loopWaitFile, true)
			defer patch.Reset()
			mockGetFullMeshInfo := gomonkey.ApplyFunc(GetCurSuperPodInfoFromMapA3,
				func(_ *SuperPodInfo) ([]string, map[string][]string) {
					return fullMesh, linkPath
				})
			defer mockGetFullMeshInfo.Reset()
			configmap, MeshInfo, linkInfo := processSuperPodJson("Path", "")
			convey.So(configmap != nil, convey.ShouldBeTrue)
			convey.So(MeshInfo != nil, convey.ShouldBeTrue)
			convey.So(linkInfo != nil, convey.ShouldBeTrue)
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
			convey.So(ret1, convey.ShouldResemble, map[string]algo.NpuInfo{
				"1": algo.NpuInfo{
					SuperPodName: "", RackName: "Rack-0", OsName: "1", SlotName: "NSlot-0",
					NpuNumber: 0, IP: "", NetPlaneId: "",
				},
			})
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
	convey.Convey("test func getOneTopoFilePath", t, func() {
		convey.Convey("should return nil when servermap nil", func() {
			superPodInfo := &SuperPodInfo{
				RackMap: map[string]*RackInfo{
					"rack1": {
						RackID:    "1",
						ServerMap: map[string]*ServerInfo{},
					},
				},
			}
			ret := getOneTopoFilePath("", superPodInfo)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("when normal", func() {
			superPodInfo := &SuperPodInfo{
				RackMap: map[string]*RackInfo{
					"rack1": {
						RackID: "1",
						ServerMap: map[string]*ServerInfo{
							"1": {ServerIndex: "1"},
						},
					},
				},
			}
			expectValue := filepath.Join("0", "rack-1", "topo_1.json")
			ret := getOneTopoFilePath("0", superPodInfo)
			convey.So(ret, convey.ShouldEqual, expectValue)
		})
	})
}

func TestGetCurSuperPod1DNpuInfo(t *testing.T) {
	convey.Convey("Test func GetCurSuperPod1DNpuInfo", t, func() {
		convey.Convey("should return nil when no rackNums", func() {
			s := &SuperPodInfo{}
			npuFmlink, npuNetPlanePaths, npuEidMap := GetA5CurSuperPod1D2DNpuInfo("", s)
			convey.So(npuFmlink, convey.ShouldBeNil)
			convey.So(npuNetPlanePaths, convey.ShouldBeNil)
			convey.So(npuEidMap, convey.ShouldBeNil)
		})
	})
}
func TestGetNpuLinkPath(t *testing.T) {
	convey.Convey("test func getNpuLinkPath", t, func() {
		convey.Convey("When level 1 not exist", func() {
			npuNetPlanePaths := make(map[string][]string)
			npu := &NpuInfo{
				PhyId: "1",
				LevelList: []LevelElement{
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
					{NetLayer: 0, RankAddrList: []RankAddrItem{{Addr: "addr1", PlaneId: "0"}, {Addr: "addr2", PlaneId: "1"}}},
				},
			}
			rackId := "1"

			getA51D2DNpuLinkPath(npuNetPlanePaths, npu, rackId, "1D")

			fmt.Println(npuNetPlanePaths)
			expectVal := map[string][]string{}
			convey.So(npuNetPlanePaths, convey.ShouldResemble, expectVal)
		})
	})
}
func TestCheckIfNew1D(t *testing.T) {
	convey.Convey("test func checkIfNew1D", t, func() {
		convey.Convey("should return false when no 1D", func() {
			convey.Convey("When rackInfo is nil", func() {
				result := checkIfNew1D(nil)
				convey.So(result, convey.ShouldBeFalse)
			})
			convey.Convey("When rackInfo is empty", func() {
				result := checkIfNew1D(map[string]*RackInfo{})
				convey.So(result, convey.ShouldBeFalse)
			})
			convey.Convey("When rackInfo contains rack with nil ServerMap", func() {
				rackInfo := map[string]*RackInfo{
					"rack1": {ServerMap: nil},
				}
				result := checkIfNew1D(rackInfo)
				convey.So(result, convey.ShouldBeFalse)
			})
			convey.Convey("When rack.ServerMap is empty", func() {
				rackInfo := map[string]*RackInfo{
					"rack1": {ServerMap: map[string]*ServerInfo{}},
				}
				result := checkIfNew1D(rackInfo)
				convey.So(result, convey.ShouldBeFalse)
			})
			convey.Convey("When server's NpuMap is nil", func() {
				rackInfo := map[string]*RackInfo{
					"rack1": {
						ServerMap: map[string]*ServerInfo{
							"server1": {NpuMap: nil},
						},
					},
				}
				result := checkIfNew1D(rackInfo)
				convey.So(result, convey.ShouldBeFalse)
			})
			convey.Convey("When npu's LevelList is nil", func() {
				rackInfo := map[string]*RackInfo{
					"rack1": {
						ServerMap: map[string]*ServerInfo{
							"server1": {NpuMap: map[string]*NpuInfo{"npu1": {LevelList: nil}}},
						},
					},
				}
				result := checkIfNew1D(rackInfo)
				convey.So(result, convey.ShouldBeFalse)
			})
		})
	})
}
func TestCheckIfNew1DTrue(t *testing.T) {
	convey.Convey("test func checkIfNew1D", t, func() {
		convey.Convey("When all structure are valid", func() {
			rackInfo := map[string]*RackInfo{
				"rack1": {
					ServerMap: map[string]*ServerInfo{
						"server1": {
							NpuMap: map[string]*NpuInfo{
								"npu1": {
									LevelList: []LevelElement{
										{NetLayer: 0},
									},
								},
							},
						},
					},
				},
			}
			ret := checkIfNew1D(rackInfo)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}
func TestGetNetWorkTypePart0(t *testing.T) {
	convey.Convey("test func getWorkType", t, func() {
		controllerflags.IsControllerExited.SetState(false)
		patch0 := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
		defer patch0.Reset()

		convey.Convey("should return nil when no workType", func() {
			ret := getNetWorkType("", nil)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("should return nil when topfile err", func() {
			racInfo := RackInfo{RackID: "1"}
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &racInfo}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "")
			defer patch.Reset()
			ret := getNetWorkType("", s)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("should return nil when read file err and err is File not exist", func() {
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &RackInfo{RackID: "1"}}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "path")
			defer patch.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, nil, os.ErrNotExist)
			defer patch1.Reset()
			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer patchSleep.Reset()
			ret := getNetWorkType("", s)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("should return nil when data nil", func() {
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &RackInfo{RackID: "1"}}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "path")
			defer patch.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, nil, nil)
			defer patch1.Reset()
			ret := getNetWorkType("", s)
			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer patchSleep.Reset()
			convey.So(ret, convey.ShouldBeEmpty)
		})
	})
}

var str2D = `{"hardwareType": "Atlas 950 SuperPod 2D"}`
var str1D = `{"hardwareType": "Atlas 950 SuperPod 1D"}`
var strErr = `{ "hardwareType": ""}`
var data2D = []byte(str2D)
var data1D = []byte(str1D)
var dataErr = []byte(strErr)
var dataTest = [][]byte{dataErr, data1D, data2D}

func TestGetNetWorkTypePart1(t *testing.T) {
	convey.Convey("test func getWorkType part 1", t, func() {
		controllerflags.IsControllerExited.SetState(false)
		patch0 := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
		defer patch0.Reset()
		convey.Convey("when read file err", func() {
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &RackInfo{RackID: "1"}}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "path")
			defer patch.Reset()
			controllerflags.IsControllerExited.SetState(false)
			patch0 := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, nil, errors.New("err"))
			defer patch1.Reset()
			ret := getNetWorkType("", s)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("when json marshal err", func() {
			var data []byte = []byte{1}
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &RackInfo{RackID: "1"}}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "path")
			defer patch.Reset()
			controllerflags.IsControllerExited.SetState(false)
			patch0 := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, data, nil)
			defer patch1.Reset()
			ret := getNetWorkType("", s)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("when data can marshal", func() {
			res := []string{"", "1D", "2D"}
			s := &SuperPodInfo{RackMap: map[string]*RackInfo{"1": &RackInfo{RackID: "1"}}}
			patch := gomonkey.ApplyFuncReturn(getOneTopoFilePath, "path")
			defer patch.Reset()
			controllerflags.IsControllerExited.SetState(false)
			patch0 := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer patch0.Reset()
			if len(res) < len(dataTest) {
				dataTest = dataTest[:len(res)]
			}
			for i, data := range dataTest {
				patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, data, nil)
				ret := getNetWorkType("", s)
				convey.So(ret, convey.ShouldEqual, res[i])
				patch1.Reset()
			}
		})
	})
}
