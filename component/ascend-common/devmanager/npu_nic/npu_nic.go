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

// Package npu_nic provides the NPU-NIC affinity query based on the dcmi topology information.
package npu_nic

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
)

// nicDevice is an rdma device and the ethernet interface bound to it, ibName is the dcmi nic name and
// ethName is the ethernet interface name used by the network configuration.
type nicDevice struct {
	ibName  string
	ethName string
}

// ListNics lists the ethernet interface names of the rdma devices, sorted in lexicographic order,
// which is also the shared nic order used by matchPrimaryNics.
func ListNics() ([]string, error) {
	devices, err := listNicDevices()
	if err != nil {
		return nil, err
	}
	ethNames := make([]string, 0, len(devices))
	for _, device := range devices {
		ethNames = append(ethNames, device.ethName)
	}
	sort.Strings(ethNames)
	return ethNames, nil
}

// listNicDevices returns the rdma devices with their ethernet interfaces, sorted by the rdma device name.
func listNicDevices() ([]nicDevice, error) {
	entries, err := os.ReadDir(common.InfinibandClassPath)
	if err != nil {
		return nil, fmt.Errorf("read %s failed: %w", common.InfinibandClassPath, err)
	}
	devices := make([]nicDevice, 0, len(entries))
	for _, entry := range entries {
		ethName, err := ethNameOf(entry.Name())
		if err != nil {
			hwlog.RunLog.Debugf("get ethernet interface of nic(%s) failed: %v", entry.Name(), err)
			continue
		}
		devices = append(devices, nicDevice{ibName: entry.Name(), ethName: ethName})
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no rdma device with ethernet interface found in %s", common.InfinibandClassPath)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ibName < devices[j].ibName })
	return devices, nil
}

// ethNameOf returns the ethernet interface name of the rdma device.
func ethNameOf(ibName string) (string, error) {
	netDir := filepath.Join(common.InfinibandClassPath, ibName, common.InfinibandDeviceDirName,
		common.InfinibandNetDirName)
	entries, err := os.ReadDir(netDir)
	if err != nil {
		return "", fmt.Errorf("read %s failed: %w", netDir, err)
	}
	ethNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		ethNames = append(ethNames, entry.Name())
	}
	if len(ethNames) == 0 {
		return "", fmt.Errorf("no ethernet interface found in %s", netDir)
	}
	// A rdma device is bound to one ethernet interface, the names returned by ReadDir are sorted
	// so the first name keeps the result stable when the driver reports more than one interface.
	if len(ethNames) > 1 {
		hwlog.RunLog.Warnf("rdma device(%s) is bound to multiple ethernet interfaces: %v, use %s",
			ibName, ethNames, ethNames[0])
	}
	return ethNames[0], nil
}

// Query returns the affinity ethernet interface names of each npu, the key is the logic id and the
// primary nic is moved to the front of each nic list. Only the nics whose topo type is UB are kept.
func Query(dcMgr devmanager.DeviceInterface) (map[int32][]string, error) {
	if dcMgr == nil {
		return nil, errors.New("device manager is nil")
	}
	devices, err := listNicDevices()
	if err != nil {
		return nil, err
	}
	_, logicIDs, err := dcMgr.GetDeviceList()
	if err != nil {
		return nil, fmt.Errorf("get device list failed: %w", err)
	}
	affinity := make(map[int32][]string, len(logicIDs))
	for _, logicID := range logicIDs {
		nicNames, err := queryNpuNics(dcMgr, logicID, devices)
		if err != nil {
			return nil, err
		}
		hwlog.RunLog.Infof("npu nic affinity queried, logicID(%d), nic: %v", logicID, nicNames)
		if len(nicNames) > 0 {
			affinity[logicID] = nicNames
		}
	}
	if len(affinity) == 0 {
		return nil, fmt.Errorf("no npu nic affinity found, npu count: %d, nic count: %d",
			len(logicIDs), len(devices))
	}
	return sortWithPrimaryNics(affinity), nil
}

// matchPrimaryNics selects the primary nic of each npu with the greedy round-robin algorithm, the nic
// list wraps around so that both plugins get the same primary nic.
func matchPrimaryNics(affinity map[int32][]string) map[int32]string {
	primary := make(map[int32]string, len(affinity))
	nicNames := collectNics(affinity)
	if len(nicNames) == 0 {
		return primary
	}

	npuIDs := make([]int32, 0, len(affinity))
	for npuID := range affinity {
		npuIDs = append(npuIDs, npuID)
	}
	sort.Slice(npuIDs, func(i, j int) bool { return npuIDs[i] < npuIDs[j] })

	cursor := 0
	for _, npuID := range npuIDs {
		connected := make(map[string]struct{}, len(affinity[npuID]))
		for _, nicName := range affinity[npuID] {
			connected[nicName] = struct{}{}
		}
		for j := cursor; j < len(nicNames)+cursor; j++ {
			nicName := nicNames[j%len(nicNames)]
			if _, ok := connected[nicName]; ok {
				primary[npuID] = nicName
				cursor = j + 1
				break
			}
		}
	}
	return primary
}

// sortWithPrimaryNics returns a copy of the affinity with the matched primary nic moved to the front.
func sortWithPrimaryNics(affinity map[int32][]string) map[int32][]string {
	primary := matchPrimaryNics(affinity)
	ordered := make(map[int32][]string, len(affinity))
	for npuID, nicNames := range affinity {
		ordered[npuID] = movePrimaryFirst(nicNames, primary[npuID])
	}
	return ordered
}

// collectNics returns the deduplicated nic names of the affinity in lexicographic order.
func collectNics(affinity map[int32][]string) []string {
	seen := make(map[string]struct{})
	nics := make([]string, 0)
	for _, nicNames := range affinity {
		for _, nicName := range nicNames {
			if _, ok := seen[nicName]; ok {
				continue
			}
			seen[nicName] = struct{}{}
			nics = append(nics, nicName)
		}
	}
	sort.Strings(nics)
	return nics
}

// movePrimaryFirst moves the primary nic to the front, the order of the other nics is kept.
func movePrimaryFirst(nicNames []string, primary string) []string {
	if primary == "" {
		return nicNames
	}
	ordered := make([]string, 0, len(nicNames))
	ordered = append(ordered, primary)
	for _, nicName := range nicNames {
		if nicName != primary {
			ordered = append(ordered, nicName)
		}
	}
	return ordered
}

// queryNpuNics returns the affinity ethernet interface names of one npu.
func queryNpuNics(dcMgr devmanager.DeviceInterface, logicID int32, nics []nicDevice) ([]string, error) {
	ethNames := make([]string, 0, len(nics))
	for _, nic := range nics {
		topoType, err := dcMgr.GetTopoInfoByDeviceIDAndNicName(logicID, nic.ibName)
		if err != nil {
			if isNotSupportErr(err) {
				return nil, fmt.Errorf("topo query is not supported by driver, logicID(%d), nic(%s): %v",
					logicID, nic.ibName, err)
			}
			return nil, fmt.Errorf("get topo type of logicID(%d) and nic(%s) failed: %w",
				logicID, nic.ibName, err)
		}
		if topoType == common.TopoTypeButt || topoType == common.TopoTypeMax {
			hwlog.RunLog.Warnf("get abnormal topo type(%d) of logicID(%d) and nic(%s)",
				topoType, logicID, nic.ibName)
			continue
		}
		if topoType.IsUBAffinity() {
			ethNames = append(ethNames, nic.ethName)
		}
	}
	sort.Strings(ethNames)
	return ethNames, nil
}

// isNotSupportErr checks whether the error means the driver does not support the interface.
func isNotSupportErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), common.NotSupportErrorCode)
}
