/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
+   Licensed under the Apache License, Version 2.0 (the "License");
+   you may not use this file except in compliance with the License.
+   You may obtain a copy of the License at
+
+   http://www.apache.org/licenses/LICENSE-2.0
+
+   Unless required by applicable law or agreed to in writing, software
+   distributed under the License is distributed on an "AS IS" BASIS,
+   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
+   See the License for the specific language governing permissions and
+   limitations under the License.
+*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

// TestGetCardType for test getCardType
func TestGetCardType(t *testing.T) {
	hdm := &HwDevManager{
		manager: device.NewHwAscend910Manager(),
		allInfo: common.NpuAllInfo{
			AllDevs: []common.NpuDevice{{LogicID: 0}},
		},
	}
	mockGetDmgr := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend910Manager)), "GetDmgr",
		func(_ *device.HwAscend910Manager) devmanager.DeviceInterface { return &devmanager.DeviceManagerMock{} })
	defer mockGetDmgr.Reset()
	convey.Convey("test getCardType when get board info error", t, func() {
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{}, fmt.Errorf("get board info error")
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldBeEmpty)
	})
	convey.Convey("test getCardType success", t, func() {
		mockGetMainBoardId := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetMainBoardId", func(_ *devmanager.DeviceManagerMock) uint32 {
				return common.A5300IMainBoardId
			})
		defer mockGetMainBoardId.Reset()
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{BoardId: npuCommon.A5300IBoardId}, nil
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldEqual, common.A5300ICardName)
	})
	convey.Convey("test getCardType failed", t, func() {
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{BoardId: common.A300IA2BoardId}, nil
			})
		defer mockGetBoardInfo.Reset()
		cardType, _ := hdm.getCardType()
		convey.So(cardType, convey.ShouldBeEmpty)
	})
}

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// stubDevMgr replaces device.DevManager and only implements
type stubDevMgr struct {
	device.DevManager
	eidAddrs []string
	eidErr   error
	uboeIP   string
	uboeErr  error
}

// TestGetDpuInfo test get npu info
func TestGetDpuInfo(t *testing.T) {
	convey.Convey("Test getNpuCorrespDpuInfo", t, func() {
		hdm := &HwDevManager{
			dpuManager: &dpucontrol.DpuFilter{},
		}

		testNpuDevice := &common.NpuDevice{
			PhyID: 5, // Test physical ID
		}

		convey.Convey("When NPU with DPU infos is nil", func() {
			hdm.dpuManager.NpuWithDpuInfos = nil

			ipAddrs, err := hdm.getNpuCorrespDpuInfo(testNpuDevice)
			convey.So(ipAddrs, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "dpu infos is empty")
		})

		convey.Convey("When NPU with DPU infos exists", func() {
			npuId := testNpuDevice.PhyID % common.NpuNum
			expectedIpAddrs := []string{"10.0.0.1", "10.0.0.2"}

			hdm.dpuManager.NpuWithDpuInfos = []dpucontrol.NpuWithDpuInfo{
				{
					NpuId: npuId,
					DpuInfo: []dpucontrol.BaseDpuInfo{
						{DpuIP: expectedIpAddrs[0]},
						{DpuIP: expectedIpAddrs[1]},
					},
				},
				{
					NpuId: npuId + 1, // Different NPU ID
					DpuInfo: []dpucontrol.BaseDpuInfo{
						{DpuIP: "10.0.1.1"},
					},
				},
			}

			convey.Convey("Should return correct DPU IPs for matching NPU ID", func() {
				ipAddrs, err := hdm.getNpuCorrespDpuInfo(testNpuDevice)
				convey.So(err, convey.ShouldBeNil)
				convey.So(ipAddrs, convey.ShouldResemble, expectedIpAddrs)
			})
		})
	})
}

// TestGetDpuInfo2 test get dpu info part2
func TestGetDpuInfo2(t *testing.T) {
	convey.Convey("Test getNpuInfo", t, func() {
		hdm := &HwDevManager{
			dpuManager: &dpucontrol.DpuFilter{},
		}
		testNpuDevice := &common.NpuDevice{
			PhyID: 5, // Test physical ID
		}
		convey.Convey("When no matching NPU ID found", func() {
			hdm.dpuManager.NpuWithDpuInfos = []dpucontrol.NpuWithDpuInfo{
				{
					NpuId: 99, // Non-matching ID
					DpuInfo: []dpucontrol.BaseDpuInfo{
						{DpuIP: "10.0.1.1"},
					},
				},
			}

			expectedNpuId := testNpuDevice.PhyID % common.NpuNum
			expectedErr := fmt.Errorf("get npu %d correspond dpuinfos error", expectedNpuId)

			ipAddrs, err := hdm.getNpuCorrespDpuInfo(testNpuDevice)
			convey.So(ipAddrs, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, expectedErr.Error())
		})
	})
}

// TestHwDevManagerMethodGetDevManager test hwdev get dev manager
func TestHwDevManagerMethodGetDevManager(t *testing.T) {
	convey.Convey("test HwDevManager method GetDevManager", t, func() {
		convey.Convey("01-should return devManager instance when called", func() {
			devMgr := device.NewHwAscend910Manager()
			hdm := HwDevManager{
				manager: devMgr,
			}
			ret := hdm.GetDevManager()
			convey.So(ret, convey.ShouldEqual, devMgr)
		})
	})
}

// TestHwDevManagerMethodSetSuperPodInfo test set super pod info
func TestHwDevManagerMethodSetSuperPodInfo(t *testing.T) {
	convey.Convey("test HwDevManager method SetSuperPodInfo", t, func() {
		convey.Convey("01-should success when set super pod info is called when card type is A5", func() {
			oldValue := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A5
			defer func() {
				common.ParamOption.RealCardType = oldValue
			}()
			devMgr := device.NewHwAscend910Manager()
			hdm := HwDevManager{
				manager: devMgr,
			}

			theSuperPodSize := int32(8192)
			theSuperPodId := int32(1)
			theServerIndex := int32(1)
			theRackId := int32(1)
			patch := gomonkey.ApplyPrivateMethod(&hdm, "getSuperPodInfo", func() common.SuperPodInfo {
				return common.SuperPodInfo{
					ScaleType:  theSuperPodSize,
					SuperPodId: theSuperPodId,
					ServerId:   theServerIndex,
					RackId:     theRackId,
				}
			})
			defer patch.Reset()
			hdm.setSuperPodInfo()
			convey.So(hdm.GetRackID(), convey.ShouldEqual, theRackId)
			convey.So(hdm.GetSuperPodID(), convey.ShouldEqual, theSuperPodId)
		})
	})
}

// TestHwDevManagerMethodSetNodeInternalIPInK8s test set node internal IP in k8s
func TestHwDevManagerMethodSetNodeInternalIPInK8s(t *testing.T) {
	convey.Convey("test HwDevManager method SetNodeInternalIPInK8s", t, func() {
		dstAddr := "192.168.0.1"
		node := &v1.Node{Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: dstAddr},
			},
		}}
		convey.Convey("01-should failed when card type is not A5", func() {
			oldValue := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A3
			defer func() {
				common.ParamOption.RealCardType = oldValue
			}()
			devMgr := device.NewHwAscend910Manager()
			hdm := HwDevManager{
				manager: devMgr,
			}
			hdm.SetNodeInternalIPInK8s(node)
			convey.So(hdm.manager.GetNodeInternalIPInK8s(), convey.ShouldBeEmpty)
		})

		convey.Convey("02-should failed when node is nil", func() {
			oldValue := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A5
			defer func() {
				common.ParamOption.RealCardType = oldValue
			}()
			devMgr := device.NewHwAscend910Manager()
			hdm := HwDevManager{
				manager: devMgr,
			}
			hdm.SetNodeInternalIPInK8s(nil)
			convey.So(hdm.manager.GetNodeInternalIPInK8s(), convey.ShouldBeEmpty)
		})
		convey.Convey("03-should success when node is valid", func() {
			oldValue := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910A5
			defer func() {
				common.ParamOption.RealCardType = oldValue
			}()
			devMgr := device.NewHwAscend910Manager()
			hdm := HwDevManager{
				manager: devMgr,
			}
			hdm.SetNodeInternalIPInK8s(node)
			convey.So(hdm.manager.GetNodeInternalIPInK8s(), convey.ShouldEqual, dstAddr)
		})
	})
}
