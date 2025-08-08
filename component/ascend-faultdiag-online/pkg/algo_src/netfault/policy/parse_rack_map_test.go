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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetCurRackInfo(t *testing.T) {
	convey.Convey("Test getCurRackInfo", t, func() {
		convey.Convey("should return nil when RackID is not exist", func() {
			rackInfo := &RackInfo{
				ServerMap: nil,
			}
			ret := getCurRackInfo(nil, rackInfo)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when RackID is not number", func() {
			rackInfo := &RackInfo{
				RackID:    "rackid",
				ServerMap: nil,
			}
			ret := getCurRackInfo(nil, rackInfo)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when ServerMap is nil", func() {
			rackInfo := &RackInfo{
				RackID:    "1",
				ServerMap: nil,
			}
			ret := getCurRackInfo(nil, rackInfo)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when osInfo is nil", func() {
			rackInfo := &RackInfo{
				RackID: "1",
				ServerMap: map[string]*ServerInfo{
					"1": nil,
				},
			}
			ret := getCurRackInfo(nil, rackInfo)
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}

// TestCheck2DFullMeshLinkNotExist test for func check2DFullMeshLinkNotExist
func TestCheck2DFullMeshLinkNotExist(t *testing.T) {
	convey.Convey("Test check2DFullMeshLinkNotExist", t, func() {
		convey.Convey("should return false when array exist aToB or bToA", func() {
			array := []string{"aToB", "bToA", "111"}
			aToB := "aToB"
			bToA := "bToA"
			ret := check2DFullMeshLinkNotExist(array, aToB, bToA)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when array don't exist aToB or bToA", func() {
			array := []string{"111"}
			aToB := "aToB"
			bToA := "bToA"
			ret := check2DFullMeshLinkNotExist(array, aToB, bToA)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestStoreRack2DFullMeshLink test for func storeRack2DFullMeshLink
func TestStoreRack2DFullMeshLink(t *testing.T) {
	convey.Convey("Test storeRack2DFullMeshLink", t, func() {
		convey.Convey("should return nil when current Rack is invalid", func() {
			npu2DXFullMesh := map[int]string{}
			npu2DYFullMesh := map[int]string{}
			serverNpusMap := map[string]int{}
			ret := storeRack2DFullMeshLink(npu2DXFullMesh, npu2DYFullMesh, serverNpusMap)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return false when array exist aToB or bToA", func() {
			npu2DXFullMesh := map[int]string{0: "192.168.10.0", 1: "192.168.10.1", 2: "192.168.10.2", 3: "192.168.10.3",
				4: "192.168.10.4", 5: "192.168.10.5", 6: "192.168.10.6", 7: "192.168.10.7",
				8: "192.168.10.8", 9: "192.168.10.9", 10: "192.168.10.10", 11: "192.168.10.11", 12: "192.168.10.12",
				13: "192.168.10.13", 14: "192.168.10.14", 15: "192.168.10.15"}
			npu2DYFullMesh := map[int]string{0: "192.168.10.0", 1: "192.168.10.1", 2: "192.168.10.2", 3: "192.168.10.3",
				4: "192.168.10.4", 5: "192.168.10.5", 6: "192.168.10.6", 7: "192.168.10.7",
				8: "192.168.10.8", 9: "192.168.10.9", 10: "192.168.10.10", 11: "192.168.10.11", 12: "192.168.10.12",
				13: "192.168.10.13", 14: "192.168.10.14", 15: "192.168.10.15"}
			serverNpusMap := map[string]int{"server1": 8, "server2": 8}

			ret := storeRack2DFullMeshLink(npu2DXFullMesh, npu2DYFullMesh, serverNpusMap)
			expectReturnValue := []string{"192.168.10.0:0#192.168.10.1:0", "192.168.10.0:0#192.168.10.2:0",
				"192.168.10.0:0#192.168.10.3:0", "192.168.10.0:0#192.168.10.4:0", "192.168.10.0:0#192.168.10.5:0",
				"192.168.10.0:0#192.168.10.6:0", "192.168.10.0:0#192.168.10.7:0", "192.168.10.1:0#192.168.10.2:0",
				"192.168.10.1:0#192.168.10.3:0", "192.168.10.1:0#192.168.10.4:0", "192.168.10.1:0#192.168.10.5:0",
				"192.168.10.1:0#192.168.10.6:0", "192.168.10.1:0#192.168.10.7:0", "192.168.10.2:0#192.168.10.3:0",
				"192.168.10.2:0#192.168.10.4:0", "192.168.10.2:0#192.168.10.5:0", "192.168.10.2:0#192.168.10.6:0",
				"192.168.10.2:0#192.168.10.7:0", "192.168.10.3:0#192.168.10.4:0", "192.168.10.3:0#192.168.10.5:0",
				"192.168.10.3:0#192.168.10.6:0", "192.168.10.3:0#192.168.10.7:0", "192.168.10.4:0#192.168.10.5:0",
				"192.168.10.4:0#192.168.10.6:0", "192.168.10.4:0#192.168.10.7:0", "192.168.10.5:0#192.168.10.6:0",
				"192.168.10.5:0#192.168.10.7:0", "192.168.10.6:0#192.168.10.7:0", "192.168.10.8:0#192.168.10.9:0",
				"192.168.10.8:0#192.168.10.10:0", "192.168.10.8:0#192.168.10.11:0", "192.168.10.8:0#192.168.10.12:0",
				"192.168.10.8:0#192.168.10.13:0", "192.168.10.8:0#192.168.10.14:0", "192.168.10.8:0#192.168.10.15:0",
				"192.168.10.9:0#192.168.10.10:0", "192.168.10.9:0#192.168.10.11:0", "192.168.10.9:0#192.168.10.12:0",
				"192.168.10.9:0#192.168.10.13:0", "192.168.10.9:0#192.168.10.14:0", "192.168.10.9:0#192.168.10.15:0",
				"192.168.10.10:0#192.168.10.11:0", "192.168.10.10:0#192.168.10.12:0", "192.168.10.10:0#192.168.10.13:0",
				"192.168.10.10:0#192.168.10.14:0", "192.168.10.10:0#192.168.10.15:0", "192.168.10.11:0#192.168.10.12:0",
				"192.168.10.11:0#192.168.10.13:0", "192.168.10.11:0#192.168.10.14:0", "192.168.10.11:0#192.168.10.15:0",
				"192.168.10.12:0#192.168.10.13:0", "192.168.10.12:0#192.168.10.14:0", "192.168.10.12:0#192.168.10.15:0",
				"192.168.10.13:0#192.168.10.14:0", "192.168.10.13:0#192.168.10.15:0", "192.168.10.14:0#192.168.10.15:0",
				"192.168.10.0:0#192.168.10.8:0", "192.168.10.1:0#192.168.10.9:0", "192.168.10.2:0#192.168.10.10:0",
				"192.168.10.3:0#192.168.10.11:0", "192.168.10.4:0#192.168.10.12:0", "192.168.10.5:0#192.168.10.13:0",
				"192.168.10.6:0#192.168.10.14:0", "192.168.10.7:0#192.168.10.15:0"}
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})
	})
}

func TestStoreNpuBoard2DFullMeshIp(t *testing.T) {
	convey.Convey("test storeNpuBoard2DFullMeshIp", t, func() {

		convey.Convey("should return nil when portInfo nil", func() {
			portInfo := &VnicInfo{}
			var npuFullMeshIp map[int]string
			storeNpuBoard2DFullMeshIp(npuFullMeshIp, portInfo, 0)
		})

		convey.Convey("should return nil when portInfo no VnicIp", func() {
			portInfo := &VnicInfo{}
			storeNpuBoard2DFullMeshIp(map[int]string{}, portInfo, 0)
		})

		convey.Convey("should store when portInfo vaild", func() {
			portInfo := &VnicInfo{
				VnicIp: "0.0.0.0",
			}
			npuFullMeshInfo := map[int]string{}
			npuId := 0
			storeNpuBoard2DFullMeshIp(npuFullMeshInfo, portInfo, npuId)
			convey.So(npuFullMeshInfo[npuId], convey.ShouldEqual, portInfo.VnicIp)
		})
	})
}

func TestStoreNpuNetPlaneLink(t *testing.T) {
	convey.Convey("Test storeNpuNetplaneInfo", t, func() {
		convey.Convey("should return when portInfo no phyId", func() {
			portInfo := &VnicInfo{}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, map[string][]string{})
		})
		convey.Convey("should return when portInfo no VnicIp", func() {
			portInfo := &VnicInfo{
				PortId: "1",
			}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, map[string][]string{})
		})
		convey.Convey("should store when switch vaild", func() {
			portInfo := &VnicInfo{
				PortId: "1",
				VnicIp: "0.0.0.0",
			}
			npuNetplaneInfo := make(map[string][]string)

			portInfo.PortId = npuFirstPort
			storeNpuNetPlaneLink(0, 0, 0, portInfo, npuNetplaneInfo)
			convey.So(npuNetplaneInfo[npuFirstPort], convey.ShouldNotBeNil)
		})
	})
}

func TestStoreNpuNetPlaneLinkPartTwo(t *testing.T) {
	convey.Convey("test start", t, func() {
		npuNetplaneInfo := make(map[string][]string)
		convey.Convey("test first port", func() {
			portInfo := &VnicInfo{
				PortId: npuFirstPort,
				VnicIp: "127.0.0.1",
			}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, npuNetplaneInfo)
			convey.So(len(npuNetplaneInfo[npuFirstPort]) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("test second port", func() {
			portInfo := &VnicInfo{
				PortId: npuSecondPort,
				VnicIp: "127.0.0.1",
			}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, npuNetplaneInfo)
			convey.So(len(npuNetplaneInfo[npuSecondPort]) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("test third port", func() {
			portInfo := &VnicInfo{
				PortId: npuThirdPort,
				VnicIp: "127.0.0.1",
			}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, npuNetplaneInfo)
			convey.So(len(npuNetplaneInfo[npuThirdPort]) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("test fourth port", func() {
			portInfo := &VnicInfo{
				PortId: npuFourthPort,
				VnicIp: "127.0.0.1",
			}
			storeNpuNetPlaneLink(0, 0, 0, portInfo, npuNetplaneInfo)
			convey.So(len(npuNetplaneInfo[npuFourthPort]) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestGetCurOsInfo(t *testing.T) {
	convey.Convey("test getCurRackInfo", t, func() {
		convey.Convey("should return false when input osMap no NpuMap ", func() {
			osMap := &ServerInfo{}
			ret, _ := getCurOsInfo(0, osMap, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when input npuDetails no map[string]any", func() {
			osMap := &ServerInfo{
				NpuMap: map[string]*NpuInfo{
					"1": {PhyId: "1"},
				},
			}
			ret, _ := getCurOsInfo(0, osMap, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("should return false when get curNpusInfo false", func() {
			mockCreate := gomonkey.ApplyFunc(getCurNpusInfo, func(_ int, _ *NpuInfo,
				_ map[int]string, _ map[int]string, _ map[string][]string) bool {
				return false
			})
			defer mockCreate.Reset()
			osMap := &ServerInfo{
				NpuMap: map[string]*NpuInfo{
					"1": {PhyId: "1"},
				},
			}
			ret, _ := getCurOsInfo(0, osMap, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("should return true when get curNpusInfo true", func() {
			mockCreate := gomonkey.ApplyFunc(getCurNpusInfo, func(_ int, _ *NpuInfo,
				_ map[int]string, _ map[int]string, _ map[string][]string) bool {
				return true
			})
			defer mockCreate.Reset()
			osMap := &ServerInfo{
				NpuMap: map[string]*NpuInfo{
					"1": {PhyId: "1"},
				},
			}
			ret, _ := getCurOsInfo(0, osMap, nil, nil, nil)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestGetCurNpusInfo(t *testing.T) {
	convey.Convey("test getCurNpusInfo", t, func() {
		convey.Convey("should return false when input npuInfo no PhyId", func() {
			npuInfo := &NpuInfo{}
			ret := getCurNpusInfo(0, npuInfo, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when npuInfo PhyId atoi err", func() {
			npuInfo := &NpuInfo{
				PhyId: "a",
			}
			ret := getCurNpusInfo(0, npuInfo, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when K VnicLpMap formatted err", func() {
			npuInfo := &NpuInfo{
				PhyId:     "1",
				VnicIpMap: map[string]*VnicInfo{},
			}
			ret := getCurNpusInfo(0, npuInfo, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when  VnicLpMap V formatted err", func() {
			npuInfo := &NpuInfo{
				PhyId:     "1",
				VnicIpMap: map[string]*VnicInfo{"0": nil},
			}
			ret := getCurNpusInfo(0, npuInfo, nil, nil, nil)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when store success", func() {
			testCase := []string{npuInsideBoardPort, npuAcrossBoardPort, npuFirstPort}
			for _, portId := range testCase {
				npuInfo := &NpuInfo{
					PhyId: "1",
					VnicIpMap: map[string]*VnicInfo{
						portId: {PortId: portId, VnicIp: ""},
					},
				}
				ret := getCurNpusInfo(0, npuInfo, nil, nil, nil)
				convey.So(ret, convey.ShouldBeTrue)
			}

		})
	})
}

func TestRemoveTail(t *testing.T) {
	convey.Convey("Test removeTail", t, func() {
		convey.Convey("should return nil when input is nil", func() {
			ret := removeTail("a:b:c")
			convey.So(ret, convey.ShouldEqual, "a")
		})
	})
}

func TestGetNPUNum(t *testing.T) {
	convey.Convey("test getNPUNum", t, func() {
		convey.Convey("should return NPUNum when input valid", func() {
			npuNumber := getNPUNum("NPU0-0.0.0:1:0")
			convey.So(npuNumber, convey.ShouldEqual, 0)
		})

		convey.Convey("should return -1 when input invalid (no -)", func() {
			npuNumber := getNPUNum("NPU0:0.0.0.1:0")
			convey.So(npuNumber, convey.ShouldEqual, -1)
		})

		convey.Convey("should return -1 when input invalid (npu string err)", func() {
			npuNumber := getNPUNum("nPU0-0.0.0.1:0")
			convey.So(npuNumber, convey.ShouldEqual, -1)
		})

		convey.Convey("should return -1 when input invalid (atoi err)", func() {
			npuNumber := getNPUNum("NPUA-0.0.0.1:0")
			convey.So(npuNumber, convey.ShouldEqual, -1)
		})

	})
}

func TestParseRackInfo(t *testing.T) {
	convey.Convey("Test parseRackMap", t, func() {
		convey.Convey("should return nil when racksinfo is nil", func() {
			res, _ := parseRackMap(nil)
			convey.So(res, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when rackMap formatted err", func() {
			racksInfo := map[string]*RackInfo{}
			res, _ := parseRackMap(racksInfo)
			convey.So(res, convey.ShouldBeEmpty)
		})

		convey.Convey("should return nil when get CurRackInfo return nil", func() {
			racksInfo := map[string]*RackInfo{
				"0": {},
			}
			mockInfo := gomonkey.ApplyFunc(getCurRackInfo, func(_ map[string][]string, _ *RackInfo) []string {
				return nil
			})
			defer mockInfo.Reset()
			res, _ := parseRackMap(racksInfo)
			convey.So(res, convey.ShouldBeEmpty)
		})
	})
}
