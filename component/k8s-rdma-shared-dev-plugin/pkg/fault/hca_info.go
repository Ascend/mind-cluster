/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	util "github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

var (
	faultTimeCache   = make(map[string]int64)
	faultTimeCacheMu sync.Mutex
)

// HcaInfo represents basic information about an HCA device
type HcaInfo struct {
	Hca       string
	State     string
	PhysState string
}

// FaultDetail represents detailed information about a fault
type FaultDetail struct {
	FaultCode   string `json:"FaultCode"`
	Time        int64  `json:"Time"`
	Description string `json:"Description"`
	FaultLevel  string `json:"FaultLevel"`
}

// DPUItem represents a single DPU device information
type DPUItem struct {
	HcaName   string        `json:"HcaName"`
	EthName   string        `json:"EthName"`
	IpAddr    string        `json:"IpAddr,omitempty"`
	DeviceID  string        `json:"DeviceID"`
	VendorID  string        `json:"VendorID"`
	FaultList []FaultDetail `json:"FaultList"`
}

// NodeEvent represents node-level fault events (e.g. DPU card missing)
type NodeEvent struct {
	NodeName  string        `json:"NodeName"`
	FaultList []FaultDetail `json:"FaultList"`
}

type hcaBasicInfo struct {
	DeviceID string
	VendorID string
	EthName  string
	IpAddr   string
}

// DpuInfoCfg represents the DPU information configuration structure
type DpuInfoCfg struct {
	DPUInfo struct {
		DPUList   []DPUItem  `json:"DPUList"`
		NodeEvent *NodeEvent `json:"NodeEvent"`
	} `json:"DPUInfo"`
	UpdateTime int64 `json:"UpdateTime"`
}

// GetHcaDeviceID retrieves the device ID of the specified HCA device
func GetHcaDeviceID(hca string) string {
	deviceID := ReadFile(filepath.Join(common.SysClassInfiniband, hca, "device/device"))
	if deviceID != "" && !strings.HasPrefix(deviceID, "0x") {
		deviceID = "0x" + deviceID
	}
	return deviceID
}

// GetHcaVendor retrieves the vendor ID of the specified HCA device
func GetHcaVendor(hca string) string {
	vendor := ReadFile(filepath.Join(common.SysClassInfiniband, hca, "device/vendor"))
	if vendor != "" && !strings.HasPrefix(vendor, "0x") {
		vendor = "0x" + vendor
	}
	return vendor
}

func getCurrentTimeMs() int64 {
	return time.Now().UnixMilli()
}

// GetHcaEthName retrieves the Ethernet interface name associated with the specified HCA device
func GetHcaEthName(hca string) string {
	if ethName := getEthNameFromInfiniband(hca); ethName != "" {
		return ethName
	}

	entries, err := os.ReadDir(common.SysBusUb)
	if err != nil {
		hwlog.RunLog.Errorf("Error reading UB devices directory %s: %v", common.SysBusUb, err)
		return ""
	}

	for _, entry := range entries {
		ubID := entry.Name()
		infinibandDir := filepath.Join(common.SysBusUb, ubID, "infiniband")

		ibEntries, err := os.ReadDir(infinibandDir)
		if err != nil {
			continue
		}

		for _, ibEntry := range ibEntries {
			if ibEntry.Name() == hca {
				netDir := filepath.Join(common.SysBusUb, ubID, "net")
				netEntries, err := os.ReadDir(netDir)
				if err != nil {
					hwlog.RunLog.Errorf("Error reading net directory %s: %v", netDir, err)
					return ""
				}

				if len(netEntries) > 0 {
					return netEntries[0].Name()
				}
				return ""
			}
		}
	}
	return ""
}

func getEthNameFromInfiniband(hca string) string {
	netDir := filepath.Join(common.SysClassInfiniband, hca, "device", "net")
	entries, err := os.ReadDir(netDir)
	if err != nil {
		hwlog.RunLog.Errorf("Error reading infiniband net directory %s: %v", netDir, err)
		return ""
	}

	if len(entries) > 0 {
		return entries[0].Name()
	}
	return ""
}

// GetHcaIpAddr retrieves the IP address associated with the specified Ethernet interface
func GetHcaIpAddr(ethName string) string {
	if ethName == "" {
		return ""
	}

	iface, err := net.InterfaceByName(ethName)
	if err != nil {
		hwlog.RunLog.Errorf("Error getting interface %s: %v", ethName, err)
		return ""
	}

	addrs, err := iface.Addrs()
	if err != nil {
		hwlog.RunLog.Errorf("Error getting addresses for interface %s: %v", ethName, err)
		return ""
	}

	var ipv6Addr string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
			if ipv6Addr == "" {
				ipv6Addr = ipNet.IP.String()
			}
		}
	}

	return ipv6Addr
}

// BuildDPUInfoCfg builds the DPU information configuration from fault check results
func BuildDPUInfoCfg(results []FaultResult) DpuInfoCfg {
	var cfg DpuInfoCfg

	hcaFaultMap := buildHcaFaultMap(results)
	hcaBasicMap := buildHcaBasicMap(results)

	hcaNames := make([]string, 0, len(hcaBasicMap))
	for hcaName := range hcaBasicMap {
		hcaNames = append(hcaNames, hcaName)
	}
	sort.Strings(hcaNames)

	for _, hcaName := range hcaNames {
		faults := getHcaFaults(hcaName, hcaFaultMap)
		cfg.DPUInfo.DPUList = append(cfg.DPUInfo.DPUList, buildDPUItem(hcaName, hcaBasicMap[hcaName], faults))
	}

	cfg.DPUInfo.NodeEvent = buildNodeEvent(results)
	cfg.UpdateTime = time.Now().UnixMilli()
	return cfg
}

func buildHcaFaultMap(results []FaultResult) map[string][]FaultDetail {
	hcaFaultMap := make(map[string][]FaultDetail)
	activeKeys := make(map[string]bool)

	faultTimeCacheMu.Lock()
	defer faultTimeCacheMu.Unlock()

	for _, fr := range results {
		if fr.Hca == "" || fr.Result != "true" {
			continue
		}

		cacheKey := fmt.Sprintf("%s:%s", fr.Hca, fr.Fault.FaultCode)
		activeKeys[cacheKey] = true

		detectTime, exists := faultTimeCache[cacheKey]
		if !exists {
			detectTime = getCurrentTimeMs()
			faultTimeCache[cacheKey] = detectTime
			hwlog.RunLog.Errorf("New fault detected: Hca=%s, Code=%s, Level=%s, Description=%s, Time=%d",
				fr.Hca, fr.Fault.FaultCode, fr.Fault.FaultLevel, fr.Fault.Description, detectTime)
			hwlog.RunLog.Errorf("Details: %s", fr.Details)
		}

		faultDetail := FaultDetail{
			FaultCode:   fr.Fault.FaultCode,
			Time:        detectTime,
			Description: fr.Fault.Description,
			FaultLevel:  fr.Fault.FaultLevel,
		}
		hcaFaultMap[fr.Hca] = append(hcaFaultMap[fr.Hca], faultDetail)
	}

	for cacheKey := range faultTimeCache {
		if strings.HasPrefix(cacheKey, "node:") {
			continue
		}
		if !activeKeys[cacheKey] {
			delete(faultTimeCache, cacheKey)
		}
	}

	return hcaFaultMap
}

func buildHcaBasicMap(results []FaultResult) map[string]hcaBasicInfo {
	hcaBasicMap := make(map[string]hcaBasicInfo)

	for _, fr := range results {
		if fr.Hca == "" {
			continue
		}

		if _, exists := hcaBasicMap[fr.Hca]; exists {
			continue
		}

		hcaBasicMap[fr.Hca] = buildHcaBasicInfo(fr.Hca)
	}

	return hcaBasicMap
}

func buildHcaBasicInfo(hca string) hcaBasicInfo {
	ethName := GetHcaEthName(hca)
	return hcaBasicInfo{
		DeviceID: GetHcaDeviceID(hca),
		VendorID: GetHcaVendor(hca),
		EthName:  ethName,
		IpAddr:   GetHcaIpAddr(ethName),
	}
}

func getHcaFaults(hcaName string, hcaFaultMap map[string][]FaultDetail) []FaultDetail {
	if faults, hasFaults := hcaFaultMap[hcaName]; hasFaults {
		return faults
	}
	return []FaultDetail{}
}

func buildNodeEvent(results []FaultResult) *NodeEvent {
	faultTimeCacheMu.Lock()
	defer faultTimeCacheMu.Unlock()

	nodeFaults := make([]FaultDetail, 0)
	activeKeys := make(map[string]bool)
	for _, fr := range results {
		if fr.Hca != "" || fr.Result != "true" {
			continue
		}

		cacheKey := fmt.Sprintf("node:%s", fr.Fault.FaultCode)
		activeKeys[cacheKey] = true

		detectTime, exists := faultTimeCache[cacheKey]
		if !exists {
			detectTime = getCurrentTimeMs()
			faultTimeCache[cacheKey] = detectTime
			hwlog.RunLog.Errorf("Node fault detected: Code=%s, Level=%s, Description=%s, Time=%d",
				fr.Fault.FaultCode, fr.Fault.FaultLevel, fr.Fault.Description, detectTime)
			hwlog.RunLog.Errorf("Fault result details: %s", fr.Details)
		}

		nodeFaults = append(nodeFaults, FaultDetail{
			FaultCode:   fr.Fault.FaultCode,
			Time:        detectTime,
			Description: fr.Fault.Description,
			FaultLevel:  fr.Fault.FaultLevel,
		})
	}

	for cacheKey := range faultTimeCache {
		if !strings.HasPrefix(cacheKey, "node:") {
			continue
		}
		if !activeKeys[cacheKey] {
			delete(faultTimeCache, cacheKey)
		}
	}

	nodeName, err := util.GetNodeName()
	if err != nil {
		hwlog.RunLog.Errorf("Failed to get node name for NodeEvent: %v", err)
		return nil
	}
	return &NodeEvent{
		NodeName:  nodeName,
		FaultList: nodeFaults,
	}
}

func buildDPUItem(hcaName string, basicInfo hcaBasicInfo, faults []FaultDetail) DPUItem {
	return DPUItem{
		HcaName:   hcaName,
		EthName:   basicInfo.EthName,
		IpAddr:    basicInfo.IpAddr,
		DeviceID:  basicInfo.DeviceID,
		VendorID:  basicInfo.VendorID,
		FaultList: faults,
	}
}

// ReadFile reads a file and returns its content as a trimmed string
func ReadFile(path string) string {
	data, err := utils.ReadLimitBytesWithSymlink(path, 1024, validateSysfsPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
