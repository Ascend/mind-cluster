/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
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

func TestGetSuperPodType(t *testing.T) {
	convey.Convey("Test GetSuperPodType", t, func() {
		oldValue := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = api.Ascend910A5
		defer func() {
			common.ParamOption.RealCardType = oldValue
		}()
		devMgr := device.NewHwAscend910Manager()
		hdm := HwDevManager{
			manager: devMgr,
		}
		convey.So(hdm.GetSuperPodType(), convey.ShouldEqual, common.SuperPodTypeAbnormal)
	})
}

const (
	testConfigPath    = "/tmp/test-npu-nic-mapping.json"
	testIPv4Address   = "192.168.1.100"
	testIPv6Address   = "2001:db8::1"
	testLoopbackIPv4  = "127.0.0.1"
	testLoopbackIPv6  = "::1"
	testLinkLocalIPv4 = "169.254.1.1"
	testLinkLocalIPv6 = "fe80::1"
	testNicName       = "eth0"
	testNicName2      = "eth1"
)

func createTestConfig(content string) error {
	return os.WriteFile(testConfigPath, []byte(content), 0644)
}

func removeTestConfig() {
	os.Remove(testConfigPath)
	resetConfigCache()
}

func resetConfigCache() {
	npuNicMappingCache = nil
	npuNicMappingErr = nil
}

func TestGetIPAddressType_ShouldReturnIPv4_WhenIPv4Address(t *testing.T) {
	convey.Convey("TestGetIPAddressType should return IPv4", t, func() {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"standard IPv4", testIPv4Address, addrTypeIPV4},
			{"loopback IPv4", testLoopbackIPv4, addrTypeIPV4},
			{"link-local IPv4", testLinkLocalIPv4, addrTypeIPV4},
			{"invalid IP defaults to IPv4", "invalid-ip", addrTypeIPV4},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				result := getIPAddressType(tt.input)
				convey.So(result, convey.ShouldEqual, tt.expected)
			})
		}
	})
}

func TestGetIPAddressType_ShouldReturnIPv6_WhenIPv6Address(t *testing.T) {
	convey.Convey("TestGetIPAddressType should return IPv6", t, func() {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"standard IPv6", testIPv6Address, addrTypeIPV6},
			{"loopback IPv6", testLoopbackIPv6, addrTypeIPV6},
			{"link-local IPv6", testLinkLocalIPv6, addrTypeIPV6},
			{"full IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", addrTypeIPV6},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				result := getIPAddressType(tt.input)
				convey.So(result, convey.ShouldEqual, tt.expected)
			})
		}
	})
}

func TestGetInterfaceIPs_ShouldReturnEmpty_WhenInterfaceNotFound(t *testing.T) {
	convey.Convey("TestGetInterfaceIPs should return empty", t, func() {
		tests := []struct {
			name    string
			nicName string
		}{
			{"non-existent interface", "nonexistent0"},
			{"empty interface name", ""},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				result := getInterfaceIPs(tt.nicName)
				convey.So(result, convey.ShouldBeEmpty)
			})
		}
	})
}

func TestGetInterfaceIPs_ShouldReturnValidIPs_WhenInterfaceExists(t *testing.T) {
	convey.Convey("TestGetInterfaceIPs should return valid IPs", t, func() {
		tests := []struct {
			name      string
			ipAddrs   []net.Addr
			wantCount int
		}{
			{"only IPv4", []net.Addr{&net.IPNet{IP: net.ParseIP(testIPv4Address)}}, 1},
			{"only IPv6", []net.Addr{&net.IPNet{IP: net.ParseIP(testIPv6Address)}}, 1},
			{"skip loopback", []net.Addr{
				&net.IPNet{IP: net.ParseIP(testLoopbackIPv4)},
				&net.IPNet{IP: net.ParseIP(testIPv4Address)},
			}, 1},
			{"skip link-local", []net.Addr{
				&net.IPNet{IP: net.ParseIP(testLinkLocalIPv4)},
				&net.IPNet{IP: net.ParseIP(testIPv4Address)},
			}, 1},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				iface := &net.Interface{Index: 1, Name: testNicName}
				patches := gomonkey.NewPatches()
				patches.ApplyFuncReturn(net.InterfaceByName, iface, nil)
				patches.ApplyMethodReturn(iface, "Addrs", tt.ipAddrs, nil)
				defer patches.Reset()

				result := getInterfaceIPs(testNicName)
				convey.So(len(result), convey.ShouldEqual, tt.wantCount)
			})
		}
	})
}

func TestGetInterfaceIPsByPriority_ShouldReturnFirstValidIP_WhenMultipleInterfaces(t *testing.T) {
	convey.Convey("TestGetInterfaceIPsByPriority should return first valid IP", t, func() {
		tests := []struct {
			name     string
			nicNames []string
			ipMap    map[string][]string
			wantIP   string
			wantErr  bool
		}{
			{"first nic has IP", []string{testNicName, testNicName2},
				map[string][]string{testNicName: {testIPv4Address}}, testIPv4Address, false},
			{"second nic has IP when first empty", []string{testNicName, testNicName2},
				map[string][]string{testNicName2: {testIPv6Address}}, testIPv6Address, false},
			{"all nics empty", []string{testNicName, testNicName2},
				map[string][]string{}, "", true},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				patches := gomonkey.ApplyFunc(getInterfaceIPs, func(nicName string) []string {
					return tt.ipMap[nicName]
				})
				defer patches.Reset()

				result, err := getInterfaceIPsByPriority(tt.nicNames)
				convey.So(err != nil, convey.ShouldEqual, tt.wantErr)
				convey.So(result, convey.ShouldEqual, tt.wantIP)
			})
		}
	})
}

func TestGetNpuToNicNames_ShouldReturnNicList_WhenNpuIdExists(t *testing.T) {
	convey.Convey("TestGetNpuToNicNames should return NIC list", t, func() {
		tests := []struct {
			name     string
			mapping  *NpuNicMapping
			npuId    int
			wantNics []string
			wantErr  bool
		}{
			{name: "npuId 0 exists",
				mapping: &NpuNicMapping{
					NpuNics: []NpuNicItem{{NpuId: 0, NicNames: []string{"eth0", "eth1"}}},
				},
				npuId:    0,
				wantNics: []string{"eth0", "eth1"},
				wantErr:  false,
			},
			{name: "npuId 1 exists",
				mapping: &NpuNicMapping{
					NpuNics: []NpuNicItem{{NpuId: 1, NicNames: []string{"eth2"}}},
				},
				npuId:    1,
				wantNics: []string{"eth2"},
				wantErr:  false,
			},
			{name: "npuId not found",
				mapping: &NpuNicMapping{
					NpuNics: []NpuNicItem{{NpuId: 0, NicNames: []string{"eth0"}}},
				},
				npuId:    2,
				wantNics: nil,
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				npuNicMappingCache = tt.mapping
				npuNicMappingErr = nil
				result, err := getNpuToNicNames(tt.npuId)
				convey.So(err != nil, convey.ShouldEqual, tt.wantErr)
				if !tt.wantErr {
					convey.So(result, convey.ShouldResemble, tt.wantNics)
				}
			})
		}
	})
}

func TestGetNpuToNicNames_ShouldReturnNil_WhenConfigNotExists(t *testing.T) {
	convey.Convey("TestGetNpuToNicNames should return nil when config not exists", t, func() {
		// 直接设置缓存变量为 nil
		npuNicMappingCache = nil
		npuNicMappingErr = nil
		result, err := getNpuToNicNames(0)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestGetNpuToNicNames_ShouldReturnCachedResult_WhenConfigAlreadyLoaded(t *testing.T) {
	convey.Convey("TestGetNpuToNicNames should return cached result", t, func() {
		npuNicMappingCache = &NpuNicMapping{
			NpuNics: []NpuNicItem{{NpuId: 0, NicNames: []string{"cached-eth"}}},
		}
		npuNicMappingErr = nil

		result, _ := getNpuToNicNames(0)
		convey.So(result, convey.ShouldResemble, []string{"cached-eth"})
	})
}

func TestGetROCEAddrList_ShouldReturnEmpty_WhenDeviceNil(t *testing.T) {
	convey.Convey("TestGetROCEAddrList should return empty when device is nil", t, func() {
		result := (&HwDevManager{}).getROCEAddrList(nil, 8)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestGetROCEAddrList_ShouldReturnAddr_WhenConfigAndInterfaceValid(t *testing.T) {
	convey.Convey("TestGetROCEAddrList should return addr when config valid", t, func() {
		dev := &common.NpuDevice{PhyID: 0}
		nicNames := []string{testNicName}

		patches := gomonkey.NewPatches()
		patches.ApplyFuncReturn(getNpuToNicNames, nicNames, nil)
		patches.ApplyFuncReturn(getInterfaceIPsByPriority, testIPv6Address, nil)
		defer patches.Reset()

		result := (&HwDevManager{}).getROCEAddrList(dev, 8)
		convey.So(len(result), convey.ShouldEqual, 1)
		convey.So(result[0].AddrType, convey.ShouldEqual, addrTypeIPV6)
		convey.So(result[0].Addr, convey.ShouldEqual, testIPv6Address)
	})
}
