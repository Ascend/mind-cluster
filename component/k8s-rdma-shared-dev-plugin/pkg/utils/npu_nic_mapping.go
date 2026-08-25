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

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
	ascutils "ascend-common/common-utils/utils"
)

const (
	npuNicMappingConfigPath = "/etc/rdma-plugin/npu-nic-mapping.json"
	// sysClassNetPath is the sysfs directory exposing network class devices.
	sysClassNetPath = "/sys/class/net"
	// npuNicMappingMaxSize is the max allowed size (in MB) of npu-nic-mapping config.
	npuNicMappingMaxSize = 1
)

// NpuNicMapping is npu nic mapping config
type NpuNicMapping struct {
	NpuNics []NpuNicItem `json:"npuNics"`
}

// NpuNicItem is npu nic item
type NpuNicItem struct {
	NpuId    int      `json:"npuId"`
	NicNames []string `json:"nicNames"`
}

// ProductMapping is a product form with its npu nic mapping.
type ProductMapping struct {
	ProductType string       `json:"productType"`
	NpuNics     []NpuNicItem `json:"npuNics"`
}

var (
	dpuToNpuOnce sync.Once
	dpuToNpuIDs  map[string][]int
)

// InitNpuNicMapping loads npu-nic-mapping.json once and builds reverse index.
func InitNpuNicMapping() {
	dpuToNpuOnce.Do(func() {
		realPath, err := ascutils.RealFileChecker(npuNicMappingConfigPath, false, false, npuNicMappingMaxSize)
		if err != nil {
			hwlog.RunLog.Warnf("check npu-nic-mapping config failed: %v, AffectedNPU will be empty", err)
			return
		}
		data, err := os.ReadFile(realPath)
		if err != nil {
			hwlog.RunLog.Warnf("read npu-nic-mapping config failed: %v, AffectedNPU will be empty", err)
			return
		}

		var products []ProductMapping
		if err = json.Unmarshal(data, &products); err == nil {
			index, form, ferr := selectByForm(products)
			if ferr != nil {
				hwlog.RunLog.Errorf("select npu-nic-mapping by form failed: %v", ferr)
				return
			}
			dpuToNpuIDs = index
			hwlog.RunLog.Infof("npu-nic-mapping loaded for form %q, primary-DPU reverse index size: %d",
				form, len(index))
			return
		}

		var mapping NpuNicMapping
		if err = json.Unmarshal(data, &mapping); err != nil {
			hwlog.RunLog.Errorf("parse npu-nic-mapping config failed: %v", err)
			return
		}

		dpuToNpuIDs = buildReverseIndex(mapping.NpuNics)
		hwlog.RunLog.Infof("npu-nic-mapping loaded as customize, "+
			"primary-DPU reverse index size: %d", len(dpuToNpuIDs))
	})
}

// GetAffectedNPU returns affected NPU ids by HCA name from reverse index.
func GetAffectedNPU(ethName string) []int {
	if ethName == "" {
		return []int{}
	}
	if npuIds, ok := dpuToNpuIDs[ethName]; ok {
		result := make([]int, len(npuIds))
		copy(result, npuIds)
		return result
	}
	return []int{}
}

func machineType() (string, error) {
	dirs, err := os.ReadDir(sysClassNetPath)
	if err != nil {
		return "", fmt.Errorf("read net class dir failed: %v", err)
	}
	var others []os.DirEntry
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Name(), "ens") {
			if mt, ok := cardTypeOf(dir.Name()); ok {
				return mt, nil
			}
		}
		others = append(others, dir)
	}
	for _, dir := range others {
		if mt, ok := cardTypeOf(dir.Name()); ok {
			return mt, nil
		}
	}
	return "", fmt.Errorf("no card_type found under %s", sysClassNetPath)
}

func cardTypeOf(ifName string) (string, bool) {
	cardType, ok := readCardType(ifName)
	if !ok {
		return "", false
	}
	switch {
	case cardType == "A5Server":
		return "Server", true
	case strings.HasPrefix(cardType, "A5Pod"):
		return "PoD", true
	default:
		hwlog.RunLog.Debugf("skip unknown card_type: %q on %s", cardType, ifName)
	}
	return "", false
}

// validateSysfsPath ensures the resolved path stays under /sys.
func validateSysfsPath(resolvedPath string) bool {
	return strings.HasPrefix(resolvedPath, "/sys/")
}

func readCardType(ifName string) (string, bool) {
	for _, path := range []string{
		fmt.Sprintf("%s/%s/device/card_type", sysClassNetPath, ifName),
		fmt.Sprintf("%s/%s/card_type", sysClassNetPath, ifName),
	} {
		data, err := ascutils.ReadLimitBytesWithSymlink(path, 1024, validateSysfsPath)
		if err != nil {
			hwlog.RunLog.Warnf("read card_type failed for %s: %v", path, err)
			continue
		}
		return strings.TrimSpace(string(data)), true
	}
	return "", false
}

// selectByForm finds the product type and builds the reverse index from its nics
func selectByForm(products []ProductMapping) (map[string][]int, string, error) {
	form, err := machineType()
	if err != nil {
		return nil, "", fmt.Errorf("detect machine form failed: %w", err)
	}
	for _, p := range products {
		if strings.EqualFold(p.ProductType, form) {
			return buildReverseIndex(p.NpuNics), form, nil
		}
	}
	return nil, form, fmt.Errorf("machine form %q not found in npu-nic-mapping", form)
}

// buildReverseIndex maps primary DPU (first NIC in NicNames) to NPU ids.
func buildReverseIndex(items []NpuNicItem) map[string][]int {
	index := make(map[string][]int)
	for _, item := range items {
		if len(item.NicNames) == 0 {
			continue
		}
		primaryDpu := item.NicNames[0]
		index[primaryDpu] = append(index[primaryDpu], item.NpuId)
	}
	return index
}
