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

// Package dpucontrol is used for find dpu.
package dpucontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

func (df *DpuFilter) addDpuByNpuId(cardID int32, dpuIndex int, dpuInfo []BaseDpuInfo) {
	if len(dpuInfo) == onlyOneDpu {
		df.NpuWithDpuInfos = append(df.NpuWithDpuInfos, NpuWithDpuInfo{
			NpuId:   cardID,
			DpuInfo: dpuInfo,
		})
		return
	}
	if dpuIndex >= 0 && dpuIndex < len(dpuInfo) {
		df.NpuWithDpuInfos = append(df.NpuWithDpuInfos, NpuWithDpuInfo{
			NpuId:   cardID,
			DpuInfo: []BaseDpuInfo{dpuInfo[dpuIndex]},
		})
	} else {
		hwlog.RunLog.Errorf("dpuIndex %d out of range, dpuInfo: %v", dpuIndex, dpuInfo)
	}
}

func (df *DpuFilter) getDpuByPcieBusInfo(pcieBusInfo string) (string, []BaseDpuInfo, error) {
	pcieSw, err := df.getPcieswByBusId(pcieBusInfo)
	if err != nil {
		return "", []BaseDpuInfo{}, err
	}
	nics, err := df.getNicsByPcieSw(pcieSw)
	if err != nil {
		return pcieSw, []BaseDpuInfo{}, err
	}
	df.entries = nics
	dpuInfos, err := df.filterDpu()
	if err != nil {
		return pcieSw, []BaseDpuInfo{}, err
	}
	if len(dpuInfos) == 0 {
		return pcieSw, []BaseDpuInfo{}, fmt.Errorf("filter dpu infos is nil")
	}
	return pcieSw, dpuInfos, nil
}

func (df *DpuFilter) getPcieswByBusId(busId string) (string, error) {
	targetPath, err := os.Readlink(filepath.Join(pcieSwitchDir, busId))
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	absPath := filepath.Join("/sys", strings.TrimPrefix(abs, "/"))
	parts := strings.Split(absPath, "/")
	if len(parts) < pcieDirLen {
		return "", fmt.Errorf("%s get pcieswitch by bus id parts: %s have err", api.DpuLogPrefix, parts)
	}
	return strings.Join(parts[:pcieDirLen], "/"), nil
}

func (df *DpuFilter) getNicsByPcieSw(busId string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(netPath)
	if err != nil {
		return nil, err
	}
	var nics []os.DirEntry
	for _, entry := range entries {
		sysFs := filepath.Join(netPath, entry.Name())
		targetPath, err := os.Readlink(sysFs)
		if err != nil {
			hwlog.RunLog.Errorf("get error when Readlink [%v] err: %v", sysFs, err.Error())
			continue
		}
		absPath := filepath.Join(filepath.Dir(sysFs), targetPath)
		hwlog.RunLog.Infof("%s %s %s %s", api.DpuLogPrefix, entry.Name(), targetPath, busId)
		if strings.Contains(absPath, busId) {
			nics = append(nics, entry)
		}
	}
	return nics, nil
}

func (df *DpuFilter) filterDpu() ([]BaseDpuInfo, error) {
	configVendors := df.userConfig.Selectors.Vendor
	configDeviceIDs := df.userConfig.Selectors.DeviceIds
	configDeviceNames := df.userConfig.Selectors.DeviceNames
	if len(df.entries) == 0 || len(df.entries) > math.MaxInt32 {
		return []BaseDpuInfo{}, fmt.Errorf("dpu entries number have err")
	}
	var dpuInfos []BaseDpuInfo
	for _, entry := range df.entries {
		ifaceName := entry.Name()
		dpuPath := filepath.Join(netPath, ifaceName)
		dpuDir, err := os.Readlink(dpuPath)
		if err != nil {
			hwlog.RunLog.Errorf("dpu path [%s] Readlink failed, err: %v", dpuPath, err.Error())
			continue
		}
		ifacePath := filepath.Join(filepath.Dir(dpuPath), dpuDir)
		dpuDeviceDirPath := filepath.Join(ifacePath, deviceDir)

		// check dpu-config selectors filter
		isVendorFiltered, vendorValue := df.shouldFilterByVendor(dpuDeviceDirPath, configVendors)
		isDeviceIDFiltered, deviceIDValue := df.shouldFilterByDeviceID(dpuDeviceDirPath, configDeviceIDs)
		isDeviceNameFiltered := df.shouldFilterByDeviceName(ifaceName, configDeviceNames)
		if isVendorFiltered || isDeviceIDFiltered || isDeviceNameFiltered {
			continue
		}
		ips := getInterfaceIPs(ifaceName)
		dpuInfos = append(dpuInfos, BaseDpuInfo{
			DeviceName: ifaceName,
			DpuIP:      ips,
			Vendor:     vendorValue,
			DeviceId:   deviceIDValue,
			Operstate:  api.DpuStatusDown,
		})
	}
	return dpuInfos, nil
}

func (df *DpuFilter) shouldFilterByField(basePath, fileName string, allowed []string) (bool, string) {
	value, err := readFileContent(filepath.Join(basePath, fileName))
	if err != nil {
		hwlog.RunLog.Errorf("read [%v] [%v] FileContent err: %v", basePath, fileName, err)
		return true, ""
	}
	if len(allowed) > 0 && !slices.Contains(allowed, value) {
		return true, ""
	}
	return false, value
}

func (df *DpuFilter) shouldFilterByVendor(dpuDeviceDirPath string, vendors []string) (bool, string) {
	return df.shouldFilterByField(dpuDeviceDirPath, vendorFile, vendors)
}

func (df *DpuFilter) shouldFilterByDeviceID(dpuDeviceDirPath string, deviceIDs []string) (bool, string) {
	return df.shouldFilterByField(dpuDeviceDirPath, deviceFile, deviceIDs)
}

func (df *DpuFilter) shouldFilterByDeviceName(ifaceName string, deviceNames []string) bool {
	return len(deviceNames) > 0 && !slices.Contains(deviceNames, ifaceName)
}

// getNpuCorrespDpuInfo get npu correspond dpu info
func (df *DpuFilter) getNpuCorrespDpuInfo() error {
	for npuId := 0; npuId < api.NpuCountPerNode; npuId++ {
		if npuId < npuIdxCorrespDpuRangeMiddle {
			dpuInfos := df.getDpuPair(dpuSlotIdx1, dpuSlotIdx9)
			if len(dpuInfos) != dpuIpAddrsLen {
				return fmt.Errorf("get npu %d correspond dpuinfos error", npuId)
			}
			df.NpuWithDpuInfos = append(df.NpuWithDpuInfos, NpuWithDpuInfo{
				NpuId:   int32(npuId),
				DpuInfo: dpuInfos,
			})
		}
		if npuId >= npuIdxCorrespDpuRangeMiddle {
			dpuInfos := df.getDpuPair(dpuSlotIdx2, dpuSlotIdx10)
			if len(dpuInfos) != dpuIpAddrsLen {
				return fmt.Errorf("get npu %d correspond dpuinfos error", npuId)
			}
			df.NpuWithDpuInfos = append(df.NpuWithDpuInfos, NpuWithDpuInfo{
				NpuId:   int32(npuId),
				DpuInfo: dpuInfos,
			})
		}
	}
	return nil
}

// getDpuPair get npu pair dpu
func (df *DpuFilter) getDpuPair(slot1 string, slot2 string) []BaseDpuInfo {
	if df.dpuInfos == nil {
		hwlog.RunLog.Errorf("dpuInfos is nil")
		return nil
	}
	var dpus []BaseDpuInfo
	for _, dpuinfo := range df.dpuInfos {
		slotId, err := df.getSlotId(dpuinfo.DeviceName)
		if err != nil {
			hwlog.RunLog.Errorf("get dpu %s slot_id error: %v", dpuinfo.DeviceName, err)
			continue
		}
		if slotId == slot1 || slotId == slot2 {
			dpus = append(dpus, dpuinfo)
		}
	}
	return dpus
}

func (df *DpuFilter) loadDpuConfigFromFile() error {
	jsonContent, err := utils.LoadFile(dpuConfigPath)
	if err != nil {
		return fmt.Errorf("load config from file error:%v", err)
	}
	var configList ConfigList
	if err = json.Unmarshal(jsonContent, &configList); err != nil {
		return fmt.Errorf("parse config from file error:%v", err)
	}
	userConfigList := configList.UserDpuConfigList
	if len(userConfigList) == 0 || userConfigList[0].Selectors == nil || (userConfigList[0].BusType == "") {
		return errors.New("config missing parameter, dpu devices find is not enable")
	}
	userConfig := userConfigList[0]
	busType := userConfig.BusType
	if busType != busTypeUb && busType != busTypePcie {
		return fmt.Errorf("invalid busType: %s", busType)
	}
	selectors := userConfig.Selectors
	if len(selectors.Vendor) == 0 && len(selectors.DeviceIds) == 0 {
		return errors.New("no vendor and deviceIds found, dpu devices find is not enable")
	}
	hwlog.RunLog.Infof("%s userConfig busType: %s, selectors: %v", api.DpuLogPrefix, busType, selectors)
	df.userConfig = userConfig
	return nil
}

// getSlotId get dpu slot id
func (df *DpuFilter) getSlotId(ifaceName string) (string, error) {
	dpuPath := filepath.Join(netPath, ifaceName)
	dpuDir, err := os.Readlink(dpuPath)
	if err != nil {
		return "", fmt.Errorf("readlink %s error:%v", dpuPath, err)
	}
	ifacePath := filepath.Join(filepath.Dir(dpuPath), dpuDir)
	dpuDeviceDirPath := filepath.Join(ifacePath, deviceDir)
	if df.userConfig.BusType == busTypeUb {
		slotID, err := readFileContent(filepath.Join(dpuDeviceDirPath, slotIdFile))
		if err != nil {
			return "", fmt.Errorf("dpu %s read slot_id error:%v", ifaceName, err)
		}
		return slotID, nil
	}
	return "", fmt.Errorf("busType is %s not ub", df.userConfig.BusType)
}

func readFileContent(path string) (string, error) {
	data, err := utils.LoadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("file %s is empty", path)
	}
	return strings.TrimSpace(string(data)), nil
}

func getInterfaceIPs(ifaceName string) string {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		hwlog.RunLog.Errorf("get interface %s error:%v", ifaceName, err)
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		hwlog.RunLog.Errorf("get interface %s addrs error:%v", ifaceName, err)
		return ""
	}
	ipNet, ok := addrs[0].(*net.IPNet)
	if !ok {
		hwlog.RunLog.Errorf("get interface %s addr net error:%v", ifaceName, addrs[0])
		return ""
	}
	return ipNet.IP.String()
}
