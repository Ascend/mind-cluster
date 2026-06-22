/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package fault for fault check and fault report
package fault

import (
	"context"
	"errors"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	util "github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestGetHcaDeviceIDWithPrefix(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("0xa222"), nil
	})
	defer patches.Reset()

	convey.Convey("When device ID has 0x prefix", t, func() {
		result := GetHcaDeviceID("mlx5_0")
		convey.Convey("Then it should be returned as-is", func() {
			convey.So(result, convey.ShouldEqual, "0xa222")
		})
	})
}

func TestGetHcaDeviceIDWithoutPrefix(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("a222"), nil
	})
	defer patches.Reset()

	convey.Convey("When device ID does not have 0x prefix", t, func() {
		result := GetHcaDeviceID("mlx5_0")
		convey.Convey("Then 0x prefix should be added", func() {
			convey.So(result, convey.ShouldEqual, "0xa222")
		})
	})
}

func TestGetHcaDeviceIDReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("When reading device ID fails", t, func() {
		result := GetHcaDeviceID("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaVendorWithPrefix(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("0x19e5"), nil
	})
	defer patches.Reset()

	convey.Convey("When vendor ID has 0x prefix", t, func() {
		result := GetHcaVendor("mlx5_0")
		convey.Convey("Then it should be returned as-is", func() {
			convey.So(result, convey.ShouldEqual, "0x19e5")
		})
	})
}

func TestGetHcaVendorWithoutPrefix(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("19e5"), nil
	})
	defer patches.Reset()

	convey.Convey("When vendor ID does not have 0x prefix", t, func() {
		result := GetHcaVendor("mlx5_0")
		convey.Convey("Then 0x prefix should be added", func() {
			convey.So(result, convey.ShouldEqual, "0x19e5")
		})
	})
}

func TestGetHcaVendorReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("When reading vendor ID fails", t, func() {
		result := GetHcaVendor("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaEthNameFromInfiniband(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "infiniband") && strings.Contains(name, "device") && strings.Contains(name, "net") {
			return []os.DirEntry{&mockHcaDirEntry{name: "enp0s1"}}, nil
		}
		return nil, errors.New("not found")
	})
	defer patches.Reset()

	convey.Convey("When eth name is found from infiniband path", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then it should be returned", func() {
			convey.So(result, convey.ShouldEqual, "enp0s1")
		})
	})
}

func TestGetHcaEthNameFromUBBus(t *testing.T) {
	callCount := 0
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		callCount++
		if strings.Contains(name, "device/net") {
			return nil, errors.New("not found from infiniband")
		}
		if strings.Contains(name, "infiniband") {
			return []os.DirEntry{&mockHcaDirEntry{name: "mlx5_0"}}, nil
		}
		if strings.Contains(name, "net") {
			return []os.DirEntry{&mockHcaDirEntry{name: "enp0s1"}}, nil
		}
		if strings.Contains(name, common.SysBusUb) {
			return []os.DirEntry{&mockHcaDirEntry{name: "ub0"}}, nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When eth name is found from UB bus path", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then it should be returned", func() {
			convey.So(result, convey.ShouldEqual, "enp0s1")
		})
	})
}

func TestGetHcaEthNameNotFound(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("not found")
	})
	defer patches.Reset()

	convey.Convey("When eth name cannot be found", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaEthNameUBBusReadError(t *testing.T) {
	firstCall := true
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if firstCall {
			firstCall = false
			return nil, errors.New("infiniband net dir not found")
		}
		return nil, errors.New("ub bus read error")
	})
	defer patches.Reset()

	convey.Convey("When reading UB bus directory fails", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaIpAddrEmptyEthName(t *testing.T) {
	convey.Convey("When eth name is empty", t, func() {
		result := GetHcaIpAddr("")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaIpAddrInterfaceError(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return nil, errors.New("interface not found")
	})
	defer patches.Reset()

	convey.Convey("When getting interface fails", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaIpAddrAddrsError(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return &net.Interface{}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&net.Interface{}), "Addrs", func(_ *net.Interface) ([]net.Addr, error) {
		return nil, errors.New("addrs error")
	})
	defer patches.Reset()

	convey.Convey("When getting addresses fails", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaIpAddrIPv4(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return &net.Interface{}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&net.Interface{}), "Addrs", func(_ *net.Interface) ([]net.Addr, error) {
		ipNet := &net.IPNet{IP: net.ParseIP("10.0.0.1")}
		return []net.Addr{ipNet}, nil
	})
	defer patches.Reset()

	convey.Convey("When IPv4 address is available", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then IPv4 address should be returned", func() {
			convey.So(result, convey.ShouldEqual, "10.0.0.1")
		})
	})
}

func TestGetHcaIpAddrIPv6(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return &net.Interface{}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&net.Interface{}), "Addrs", func(_ *net.Interface) ([]net.Addr, error) {
		ipNet := &net.IPNet{IP: net.ParseIP("fe80::1")}
		return []net.Addr{ipNet}, nil
	})
	defer patches.Reset()

	convey.Convey("When only IPv6 address is available", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then IPv6 address should be returned", func() {
			convey.So(result, convey.ShouldEqual, "fe80::1")
		})
	})
}

func TestGetHcaIpAddrLoopbackOnly(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return &net.Interface{}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&net.Interface{}), "Addrs", func(_ *net.Interface) ([]net.Addr, error) {
		ipNet := &net.IPNet{IP: net.ParseIP("127.0.0.1")}
		return []net.Addr{ipNet}, nil
	})
	defer patches.Reset()

	convey.Convey("When only loopback address is available", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestBuildDPUInfoCfgNoFaults(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024"},
			Hca:     "mlx5_0",
			Result:  "false",
			Details: "port state: ACTIVE",
		},
	}

	patches := gomonkey.ApplyFunc(GetHcaDeviceID, func(hca string) string { return "0xa222" })
	patches.ApplyFunc(GetHcaVendor, func(hca string) string { return "0x19e5" })
	patches.ApplyFunc(GetHcaEthName, func(hca string) string { return "enp0s1" })
	patches.ApplyFunc(GetHcaIpAddr, func(ethName string) string { return "10.0.0.1" })
	patches.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When no faults are detected", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then DPU list should have one entry with empty fault list", func() {
			convey.So(len(cfg.DPUInfo.DPUList), convey.ShouldEqual, 1)
			convey.So(cfg.DPUInfo.DPUList[0].HcaName, convey.ShouldEqual, "mlx5_0")
			convey.So(cfg.DPUInfo.DPUList[0].EthName, convey.ShouldEqual, "enp0s1")
			convey.So(cfg.DPUInfo.DPUList[0].IpAddr, convey.ShouldEqual, "10.0.0.1")
			convey.So(cfg.DPUInfo.DPUList[0].DeviceID, convey.ShouldEqual, "0xa222")
			convey.So(cfg.DPUInfo.DPUList[0].VendorID, convey.ShouldEqual, "0x19e5")
			convey.So(len(cfg.DPUInfo.DPUList[0].FaultList), convey.ShouldEqual, 0)
		})
	})
}

func TestBuildDPUInfoCfgWithFaults(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024", FaultLevel: "SeparateDPU", Description: "port down"},
			Hca:     "mlx5_0",
			Result:  "true",
			Details: "port state: DOWN",
		},
	}

	patches := gomonkey.ApplyFunc(GetHcaDeviceID, func(hca string) string { return "0xa222" })
	patches.ApplyFunc(GetHcaVendor, func(hca string) string { return "0x19e5" })
	patches.ApplyFunc(GetHcaEthName, func(hca string) string { return "enp0s1" })
	patches.ApplyFunc(GetHcaIpAddr, func(ethName string) string { return "10.0.0.1" })
	patches.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When faults are detected", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then DPU list should have one entry with faults", func() {
			convey.So(len(cfg.DPUInfo.DPUList), convey.ShouldEqual, 1)
			convey.So(len(cfg.DPUInfo.DPUList[0].FaultList), convey.ShouldEqual, 1)
			convey.So(cfg.DPUInfo.DPUList[0].FaultList[0].FaultCode, convey.ShouldEqual, "21000024")
			convey.So(cfg.DPUInfo.DPUList[0].FaultList[0].FaultLevel, convey.ShouldEqual, "SeparateDPU")
		})
	})
}

func TestBuildDPUInfoCfgWithNodeEvent(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "dpu_card_drop", FaultCode: "22000026", FaultLevel: "SeparateDPU", Description: "card missing"},
			Hca:     "",
			Result:  "true",
			Details: "dpu card dropped",
		},
	}

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When node-level fault is detected", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then NodeEvent should be populated", func() {
			convey.So(cfg.DPUInfo.NodeEvent, convey.ShouldNotBeNil)
			convey.So(cfg.DPUInfo.NodeEvent.NodeName, convey.ShouldEqual, "test-node")
			convey.So(len(cfg.DPUInfo.NodeEvent.FaultList), convey.ShouldEqual, 1)
			convey.So(cfg.DPUInfo.NodeEvent.FaultList[0].FaultCode, convey.ShouldEqual, "22000026")
		})
	})
}

func TestBuildDPUInfoCfgNodeEventGetNodeNameError(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "dpu_card_drop", FaultCode: "22000026", FaultLevel: "SeparateDPU", Description: "card missing"},
			Hca:     "",
			Result:  "true",
			Details: "dpu card dropped",
		},
	}

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) { return "", errors.New("node name error") })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When GetNodeName fails for NodeEvent", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then NodeEvent should be nil", func() {
			convey.So(cfg.DPUInfo.NodeEvent, convey.ShouldBeNil)
		})
	})
}

func TestBuildDPUInfoCfgMultipleHCAs(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024"},
			Hca:     "mlx5_1",
			Result:  "false",
			Details: "port ok",
		},
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024"},
			Hca:     "mlx5_0",
			Result:  "false",
			Details: "port ok",
		},
	}

	patches := gomonkey.ApplyFunc(GetHcaDeviceID, func(hca string) string { return "0xa222" })
	patches.ApplyFunc(GetHcaVendor, func(hca string) string { return "0x19e5" })
	patches.ApplyFunc(GetHcaEthName, func(hca string) string { return "enp0s1" })
	patches.ApplyFunc(GetHcaIpAddr, func(ethName string) string { return "10.0.0.1" })
	patches.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When multiple HCAs exist", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then DPU list should be sorted by HCA name", func() {
			convey.So(len(cfg.DPUInfo.DPUList), convey.ShouldEqual, 2)
			convey.So(cfg.DPUInfo.DPUList[0].HcaName, convey.ShouldEqual, "mlx5_0")
			convey.So(cfg.DPUInfo.DPUList[1].HcaName, convey.ShouldEqual, "mlx5_1")
		})
	})
}

func TestBuildDPUInfoCfgFaultTimeCacheCleared(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024", FaultLevel: "SeparateDPU", Description: "port down"},
			Hca:     "mlx5_0",
			Result:  "true",
			Details: "port state: DOWN",
		},
	}

	patches := gomonkey.ApplyFunc(GetHcaDeviceID, func(hca string) string { return "0xa222" })
	patches.ApplyFunc(GetHcaVendor, func(hca string) string { return "0x19e5" })
	patches.ApplyFunc(GetHcaEthName, func(hca string) string { return "enp0s1" })
	patches.ApplyFunc(GetHcaIpAddr, func(ethName string) string { return "10.0.0.1" })
	patches.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCache["mlx5_0:21000022"] = 1000
	faultTimeCacheMu.Unlock()

	convey.Convey("When previously cached fault is no longer active", t, func() {
		_ = BuildDPUInfoCfg(results)
		faultTimeCacheMu.Lock()
		_, exists := faultTimeCache["mlx5_0:21000022"]
		faultTimeCacheMu.Unlock()
		convey.Convey("Then stale cache entry should be removed", func() {
			convey.So(exists, convey.ShouldBeFalse)
		})
	})
}

func TestBuildDPUInfoCfgExistingFaultTimePreserved(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "hca_port_down", FaultCode: "21000024", FaultLevel: "SeparateDPU", Description: "port down"},
			Hca:     "mlx5_0",
			Result:  "true",
			Details: "port state: DOWN",
		},
	}

	patches := gomonkey.ApplyFunc(GetHcaDeviceID, func(hca string) string { return "0xa222" })
	patches.ApplyFunc(GetHcaVendor, func(hca string) string { return "0x19e5" })
	patches.ApplyFunc(GetHcaEthName, func(hca string) string { return "enp0s1" })
	patches.ApplyFunc(GetHcaIpAddr, func(ethName string) string { return "10.0.0.1" })
	patches.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	cachedTime := int64(999999)
	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCache["mlx5_0:21000024"] = cachedTime
	faultTimeCacheMu.Unlock()

	convey.Convey("When fault was previously cached", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then cached time should be preserved", func() {
			convey.So(cfg.DPUInfo.DPUList[0].FaultList[0].Time, convey.ShouldEqual, cachedTime)
		})
	})
}

func TestBuildDPUInfoCfgEmptyResults(t *testing.T) {
	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When results are empty", t, func() {
		cfg := BuildDPUInfoCfg([]FaultResult{})
		convey.Convey("Then DPU list should be empty and NodeEvent should have no faults", func() {
			convey.So(len(cfg.DPUInfo.DPUList), convey.ShouldEqual, 0)
			convey.So(cfg.DPUInfo.NodeEvent, convey.ShouldNotBeNil)
			convey.So(len(cfg.DPUInfo.NodeEvent.FaultList), convey.ShouldEqual, 0)
		})
	})
}

func TestBuildDPUInfoCfgNodeFaultCacheCleared(t *testing.T) {
	results := []FaultResult{}

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCache["node:22000026"] = 1000
	faultTimeCacheMu.Unlock()

	convey.Convey("When previously cached node fault is no longer active", t, func() {
		_ = BuildDPUInfoCfg(results)
		faultTimeCacheMu.Lock()
		_, exists := faultTimeCache["node:22000026"]
		faultTimeCacheMu.Unlock()
		convey.Convey("Then stale node cache entry should be removed", func() {
			convey.So(exists, convey.ShouldBeFalse)
		})
	})
}

func TestReadFileSuccess(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("  test content  "), nil
	})
	defer patches.Reset()

	convey.Convey("When reading file succeeds", t, func() {
		result := ReadFile("/sys/class/infiniband/mlx5_0/device/vendor")
		convey.Convey("Then trimmed content should be returned", func() {
			convey.So(result, convey.ShouldEqual, "test content")
		})
	})
}

func TestReadFileError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("When reading file fails", t, func() {
		result := ReadFile("/sys/class/infiniband/mlx5_0/device/vendor")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetEthNameFromInfinibandReadDirError(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("read dir error")
	})
	defer patches.Reset()

	convey.Convey("When reading infiniband net directory fails", t, func() {
		result := getEthNameFromInfiniband("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetEthNameFromInfinibandEmptyDir(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	})
	defer patches.Reset()

	convey.Convey("When infiniband net directory is empty", t, func() {
		result := getEthNameFromInfiniband("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetEthNameFromInfinibandSuccess(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{&mockHcaDirEntry{name: "enp0s1"}}, nil
	})
	defer patches.Reset()

	convey.Convey("When infiniband net directory has entries", t, func() {
		result := getEthNameFromInfiniband("mlx5_0")
		convey.Convey("Then first entry name should be returned", func() {
			convey.So(result, convey.ShouldEqual, "enp0s1")
		})
	})
}

func TestGetHcaEthNameUBBusNetDirError(t *testing.T) {
	infinibandReadCalled := false
	ubReadCalled := false

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "device/net") {
			infinibandReadCalled = true
			return nil, errors.New("infiniband net not found")
		}
		if strings.Contains(name, "infiniband") {
			return []os.DirEntry{&mockHcaDirEntry{name: "mlx5_0"}}, nil
		}
		if strings.Contains(name, "net") {
			return nil, errors.New("net dir read error")
		}
		if strings.Contains(name, common.SysBusUb) {
			ubReadCalled = true
			return []os.DirEntry{&mockHcaDirEntry{name: "ub0"}}, nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When reading UB net directory fails", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
			convey.So(infinibandReadCalled, convey.ShouldBeTrue)
			convey.So(ubReadCalled, convey.ShouldBeTrue)
		})
	})
}

func TestGetHcaEthNameUBBusNoMatchingHca(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "device/net") {
			return nil, errors.New("not found")
		}
		if strings.Contains(name, "infiniband") {
			return []os.DirEntry{&mockHcaDirEntry{name: "mlx5_1"}}, nil
		}
		if strings.Contains(name, common.SysBusUb) {
			return []os.DirEntry{&mockHcaDirEntry{name: "ub0"}}, nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When no matching HCA is found in UB bus", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaEthNameUBBusInfinibandReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "device/net") {
			return nil, errors.New("not found")
		}
		if strings.Contains(name, "infiniband") {
			return nil, errors.New("infiniband read error")
		}
		if strings.Contains(name, common.SysBusUb) {
			return []os.DirEntry{&mockHcaDirEntry{name: "ub0"}}, nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When reading UB infiniband directory fails", t, func() {
		result := GetHcaEthName("mlx5_0")
		convey.Convey("Then empty string should be returned", func() {
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestGetHcaIpAddrIPv4PreferredOverIPv6(t *testing.T) {
	patches := gomonkey.ApplyFunc(net.InterfaceByName, func(name string) (*net.Interface, error) {
		return &net.Interface{}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(&net.Interface{}), "Addrs", func(_ *net.Interface) ([]net.Addr, error) {
		ipNet6 := &net.IPNet{IP: net.ParseIP("fe80::1")}
		ipNet4 := &net.IPNet{IP: net.ParseIP("10.0.0.1")}
		return []net.Addr{ipNet6, ipNet4}, nil
	})
	defer patches.Reset()

	convey.Convey("When both IPv6 and IPv4 are available", t, func() {
		result := GetHcaIpAddr("enp0s1")
		convey.Convey("Then IPv4 should be preferred", func() {
			convey.So(result, convey.ShouldEqual, "10.0.0.1")
		})
	})
}

func TestBuildDPUInfoCfgWithEmptyHcaResult(t *testing.T) {
	results := []FaultResult{
		{
			Fault:   FaultConfig{Name: "dpu_card_drop", FaultCode: "22000026"},
			Hca:     "",
			Result:  "false",
			Details: "no drop",
		},
	}

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) { return "test-node", nil })
	defer patches.Reset()

	faultTimeCacheMu.Lock()
	faultTimeCache = make(map[string]int64)
	faultTimeCacheMu.Unlock()

	convey.Convey("When results contain empty Hca entries with no faults", t, func() {
		cfg := BuildDPUInfoCfg(results)
		convey.Convey("Then DPU list should be empty", func() {
			convey.So(len(cfg.DPUInfo.DPUList), convey.ShouldEqual, 0)
		})
	})
}

type mockHcaDirEntry struct {
	name  string
	isDir bool
}

func (m *mockHcaDirEntry) Name() string               { return m.name }
func (m *mockHcaDirEntry) IsDir() bool                { return m.isDir }
func (m *mockHcaDirEntry) Type() os.FileMode          { return 0 }
func (m *mockHcaDirEntry) Info() (os.FileInfo, error) { return nil, nil }
