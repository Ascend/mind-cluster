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
	"ascend-common/devmanager"
)

func (df *DpuFilter) SaveDpuConfToNode(dmgr devmanager.DeviceInterface) error {
	if !utils.IsExist(DpuConfigPath) {
		hwlog.RunLog.Infof("%s file %s not found", api.DpuLogPrefix, DpuConfigPath)
		return nil
	}
	err := df.loadDpuConfigFromFile()
	if err != nil {
		return fmt.Errorf("dpu devices find is not enable,err:%v", err)
	}
	hwlog.RunLog.Infof("%s start find bustype %s dpu", api.DpuLogPrefix, df.UserConfig.BusType)
	switch df.UserConfig.BusType {
	case busTypeUb:
		entries, err := os.ReadDir(netPath)
		if err != nil {
			return fmt.Errorf("read dpu file dir err:%v", err)
		}
		df.entries = entries
		dpuInfos, err := df.filterDpu()
		if err != nil {
			return fmt.Errorf("filter dpu err:%v", err)
		}
		df.dpuInfos = dpuInfos
		err = df.getNpuCorrespDpuInfo()
		if err != nil {
			return fmt.Errorf("build npu correspond dpu infos err:%v", err)
		}
	case busTypePcie:
		err = df.getDpuWithNpuPcieSwitch(dmgr)
		if err != nil {
			return fmt.Errorf("get dpu by npu error: %v", err)
		}
	default:
		return fmt.Errorf("unsupported busType: %s", df.UserConfig.BusType)
	}

	if len(df.NpuWithDpuInfos) == 0 {
		return errors.New("filter dpu infos result is nil")
	}
	hwlog.RunLog.Infof("%s successfully get DPU infos: %v", api.DpuLogPrefix, df.NpuWithDpuInfos)
	return nil
}

func (df *DpuFilter) getDpuWithNpuPcieSwitch(dcMgr devmanager.DeviceInterface) error {
	cardNum, cardIDList, err := dcMgr.GetCardList()
	if err != nil || cardNum == 0 {
		return fmt.Errorf("get card list error: %v", err)
	}
	pcieSwIds := make(map[string]struct{})
	for _, cardID := range cardIDList {
		pcieBusInfo, err := dcMgr.GetPCIeBusInfo(cardID)
		hwlog.RunLog.Infof("%s pcie bus info:%v", api.DpuLogPrefix, pcieBusInfo)
		if err != nil {
			return err
		}
		pcieSwId, dpuInfo, err := df.getDpuByPcieBusInfo(pcieBusInfo)
		if err != nil {
			hwlog.RunLog.Errorf("%s npu %v get dpu by busId err :%v", api.DpuLogPrefix, cardID, err)
			continue
		}
		if _, ok := pcieSwIds[pcieSwId]; !ok {
			pcieSwIds[pcieSwId] = struct{}{}
			df.addDpuByNpuId(cardID, dpuIndexFir, dpuInfo)
			continue
		}
		df.addDpuByNpuId(cardID, dpuIndexSec, dpuInfo)
	}
	return nil
}

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
	configVendors := df.UserConfig.Selectors.Vendor
	configDeviceIDs := df.UserConfig.Selectors.DeviceIds
	configDeviceNames := df.UserConfig.Selectors.DeviceNames
	if len(df.entries) == 0 || len(df.entries) > math.MaxInt32 {
		return []BaseDpuInfo{}, fmt.Errorf("the lengh of df.entries is invalid: %v", len(df.entries))
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
	jsonContent, err := utils.LoadFile(DpuConfigPath)
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
	hwlog.RunLog.Infof("%s UserConfig busType: %s, selectors: %v", api.DpuLogPrefix, busType, selectors)
	df.UserConfig = userConfig
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
	if df.UserConfig.BusType == busTypeUb {
		slotID, err := readFileContent(filepath.Join(dpuDeviceDirPath, slotIdFile))
		if err != nil {
			return "", fmt.Errorf("dpu %s read slot_id error:%v", ifaceName, err)
		}
		return slotID, nil
	}
	return "", fmt.Errorf("busType is %s not ub", df.UserConfig.BusType)
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
