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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	"ascend-common/devmanager/npu_nic"
)

const (
	npuNicMappingConfigPath = "/etc/rdma-plugin/npu-nic-mapping.json"
	// sysClassNetPath is the sysfs directory exposing network class devices.
	sysClassNetPath = "/sys/class/net"
	// npuNicMappingMaxSize is the max allowed size (in MB) of npu-nic-mapping config.
	npuNicMappingMaxSize = 1
	// npuNicPollInterval is the fixed interval of the nic count polling, not configurable.
	npuNicPollInterval = 5 * time.Minute
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
	dpuToNpuMu  sync.RWMutex
	dpuToNpuIDs map[string][]int

	npuNicMappingMu       sync.RWMutex
	npuNicMappingSnapshot *NpuNicMapping
	npuNicMappingLoaded   bool

	// deviceManager is initialized when the plugin starts, it is used to query the npu-nic affinity.
	deviceManager devmanager.DeviceInterface

	// affinityFromDriver records whether the current mapping comes from the driver.
	affinityFromDriver atomic.Bool
)

// InitNpuNicMapping records the device manager initialized by the plugin, loads the npu-nic
// mapping synchronously at startup, then polls the nic count.
// A failed refresh keeps the previous snapshot, so consumers always read the latest valid mapping.
func InitNpuNicMapping(ctx context.Context, mgr devmanager.DeviceInterface) {
	deviceManager = mgr
	refreshNpuNicMapping()
	nics, _ := npu_nic.ListNics()
	go watchNicCount(ctx, len(nics))
}

// queryNpuNicAffinity queries the npu-nic affinity from dcmi, the logic id returned by npu_nic.Query
// is converted to the npu physic id which is the npu id used by the device annotation.
func queryNpuNicAffinity() (map[int32][]string, error) {
	affinity, err := npu_nic.Query(deviceManager)
	if err != nil {
		return nil, err
	}
	phyAffinity := make(map[int32][]string, len(affinity))
	for logicID, nicNames := range affinity {
		phyID, err := deviceManager.GetPhysicIDFromLogicID(logicID)
		if err != nil {
			return nil, fmt.Errorf("get physic id of logicID(%d) failed: %w", logicID, err)
		}
		phyAffinity[phyID] = nicNames
	}
	return phyAffinity, nil
}

// watchNicCount refreshes the mapping only when the nic count changes.
func watchNicCount(ctx context.Context, lastCount int) {
	ticker := time.NewTicker(npuNicPollInterval)
	defer ticker.Stop()
	hwlog.RunLog.Infof("npu-nic-mapping poll task started, interval: %v, nic count: %d", npuNicPollInterval, lastCount)

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("npu-nic-mapping refresh stopped: %v", ctx.Err())
			return
		case <-ticker.C:
			lastCount = checkNicCountAndRefresh(lastCount)
		}
	}
}

// checkNicCountAndRefresh refreshes the mapping when the nic count changes and returns the new count.
func checkNicCountAndRefresh(lastCount int) int {
	nics, err := npu_nic.ListNics()
	if err != nil {
		hwlog.RunLog.Debugf("poll nic devices failed: %v, keep the npu-nic mapping", err)
		return lastCount
	}
	if len(nics) == lastCount {
		return lastCount
	}
	hwlog.RunLog.Infof("nic count changed from %d to %d, refresh npu-nic mapping", lastCount, len(nics))
	if !refreshNpuNicMapping() {
		return lastCount
	}
	return len(nics)
}

// refreshNpuNicMapping loads the mapping on the first call, after that only the driver is
// queried again, returning whether the refresh succeeded. When the driver is unavailable the
// existing snapshot is kept, the config file is used as fallback until a mapping is loaded.
func refreshNpuNicMapping() bool {
	if isNpuNicMappingLoaded() {
		return refreshNpuNicMappingFromDriver()
	}

	mapping, err := loadNpuNicMapping()
	switch {
	case err != nil:
		hwlog.RunLog.Warnf("load npu-nic-mapping failed: %v, will retry in the next refresh", err)
		return false
	case mapping == nil:
		hwlog.RunLog.Warnf("npu-nic-mapping config not found, will retry in the next refresh")
		return false
	default:
		setNpuNicMappingLoaded()
		indexSize := updateNpuNicMapping(mapping)
		hwlog.RunLog.Infof("npu-nic-mapping loaded from %s, primary-DPU reverse index size: %d",
			mappingSource(), indexSize)
		return true
	}
}

// refreshNpuNicMappingFromDriver refreshes an already loaded snapshot with driver data, returning
// whether the refresh succeeded. A failed refresh keeps the previous snapshot.
func refreshNpuNicMappingFromDriver() bool {
	affinity, err := queryNpuNicAffinity()
	if err != nil {
		hwlog.RunLog.Debugf("refresh npu-nic-mapping from driver failed: %v, use the memory snapshot", err)
		return false
	}
	if len(affinity) == 0 {
		hwlog.RunLog.Debugf("refresh npu-nic-mapping from driver returned empty result, use the memory snapshot")
		return false
	}

	affinityFromDriver.Store(true)
	indexSize := updateNpuNicMapping(newNpuNicMapping(affinity))
	hwlog.RunLog.Infof("npu-nic-mapping refreshed from driver, primary-DPU reverse index size: %d", indexSize)
	return true
}

// isNpuNicMappingLoaded reports whether a valid mapping has been loaded.
func isNpuNicMappingLoaded() bool {
	npuNicMappingMu.RLock()
	defer npuNicMappingMu.RUnlock()
	return npuNicMappingLoaded
}

// setNpuNicMappingLoaded marks the first load as attempted.
func setNpuNicMappingLoaded() {
	npuNicMappingMu.Lock()
	npuNicMappingLoaded = true
	npuNicMappingMu.Unlock()
}

// updateNpuNicMapping replaces the snapshot and the reverse index built from it,
// returning the size of the reverse index.
func updateNpuNicMapping(mapping *NpuNicMapping) int {
	index := buildReverseIndex(mapping.NpuNics)

	npuNicMappingMu.Lock()
	npuNicMappingSnapshot = mapping
	npuNicMappingMu.Unlock()

	dpuToNpuMu.Lock()
	dpuToNpuIDs = index
	dpuToNpuMu.Unlock()

	return len(index)
}

// getNpuNicMappingSnapshot returns the current mapping snapshot.
func getNpuNicMappingSnapshot() *NpuNicMapping {
	npuNicMappingMu.RLock()
	defer npuNicMappingMu.RUnlock()
	return npuNicMappingSnapshot
}

// GetAffectedNPU returns affected NPU ids by HCA name from reverse index.
func GetAffectedNPU(ethName string) []int {
	if ethName == "" {
		return []int{}
	}
	dpuToNpuMu.RLock()
	defer dpuToNpuMu.RUnlock()
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
			hwlog.RunLog.Debugf("Found net device %s", dir.Name())
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
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if cardType := strings.TrimSpace(string(data)); cardType != "" {
			return cardType, true
		}
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

// loadNpuNicMapping queries the npu-nic affinity from dcmi first, then falls back to
// the config file baked into the image.
func loadNpuNicMapping() (*NpuNicMapping, error) {
	affinity, err := queryNpuNicAffinity()
	if err == nil && len(affinity) > 0 {
		affinityFromDriver.Store(true)
		return newNpuNicMapping(affinity), nil
	}
	affinityFromDriver.Store(false)
	if err != nil {
		hwlog.RunLog.Warnf("query npu nic affinity from driver failed: %v, fallback to config file", err)
	} else {
		hwlog.RunLog.Warnf("query npu nic affinity from driver returned empty result, fallback to config file")
	}
	return loadNpuNicMappingFile()
}

// newNpuNicMapping converts the affinity queried from driver to NpuNicMapping. The primary nic
// returned by npu_nic.Query is already the first nic of each npu.
func newNpuNicMapping(affinity map[int32][]string) *NpuNicMapping {
	npuIds := make([]int, 0, len(affinity))
	for npuId := range affinity {
		npuIds = append(npuIds, int(npuId))
	}
	sort.Ints(npuIds)

	mapping := &NpuNicMapping{NpuNics: make([]NpuNicItem, 0, len(npuIds))}
	for _, npuId := range npuIds {
		mapping.NpuNics = append(mapping.NpuNics, NpuNicItem{NpuId: npuId, NicNames: affinity[int32(npuId)]})
	}
	return mapping
}

// mappingSource returns a readable description of where the current mapping comes from.
func mappingSource() string {
	if affinityFromDriver.Load() {
		return "driver"
	}
	return "config file"
}

// loadNpuNicMappingFile reads the NPU-NIC mapping config baked into the image.
func loadNpuNicMappingFile() (*NpuNicMapping, error) {
	data, err := utils.LoadFile(npuNicMappingConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %v", err)
	}
	// LoadFile returns empty data when the file does not exist.
	if len(data) == 0 {
		hwlog.RunLog.Warnf("npu-nic-mapping config file not found: %s", npuNicMappingConfigPath)
		return nil, nil
	}

	// The config may group mappings by machine form (server/pod), try it first.
	var products []ProductMapping
	if err = json.Unmarshal(data, &products); err == nil {
		form, err := machineType()
		if err != nil {
			return nil, fmt.Errorf("detect machine form failed: %w", err)
		}
		for _, p := range products {
			if strings.EqualFold(p.ProductType, form) {
				return &NpuNicMapping{NpuNics: p.NpuNics}, nil
			}
		}
		return nil, fmt.Errorf("machine form %q not found in npu-nic-mapping", form)
	}

	var mapping NpuNicMapping
	if err = json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("parse config file error: %v", err)
	}

	return &mapping, nil
}

// GetNicNames returns the NIC names mapped to the given NPU ID
func GetNicNames(npuId int) ([]string, error) {
	mapping := getNpuNicMappingSnapshot()
	if mapping == nil {
		return nil, fmt.Errorf("npu-nic-mapping config not found")
	}

	for _, item := range mapping.NpuNics {
		if item.NpuId == npuId {
			return item.NicNames, nil
		}
	}

	return nil, fmt.Errorf("npuId %d not found in mapping", npuId)
}
