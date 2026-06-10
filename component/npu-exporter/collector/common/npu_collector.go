/* Copyright(C) 2021-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for general collector
package common

import (
	"context"
	"strconv"
	"sync"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/cache"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"

	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	npuContainerInfoInit sync.Once
	npuChipInfoInit      sync.Once
	// Collector base collector for prometheus and telegraf
	Collector *NpuCollector

	// ChainForSingleGoroutine a list of collectors for single goroutine
	ChainForSingleGoroutine []MetricsCollector

	// ChainForMultiGoroutine a list of collectors for multi goroutine
	ChainForMultiGoroutine []MetricsCollector

	// ChainForCustomPlugin a list of collectors for plugin
	ChainForCustomPlugin []MetricsCollector

	updateTimeForCardIds = time.Minute

	npuInfoCache sync.Map

	chainsMu sync.RWMutex
)

const (
	maxCollectTimeout              = 10 * time.Second
	tickerIntervalForContainerInfo = 5 * time.Second
	pcieDomain                     = "PcieDomain"
	fetchTimeout                   = "FetchTimeoutError"
)

// fetchPcieOptions for control pcie error logs num
var fetchPcieOptions = logger.LogOptions{
	Domain: pcieDomain,
	ID:     fetchTimeout,
}

// NpuCollector for collect metrics
type NpuCollector struct {
	cache         *cache.ConcurrencyLRUCache
	devicesParser *container.DevicesParser
	updateTime    time.Duration
	cacheTime     time.Duration
	Dmgr          devmanager.DeviceInterface
}

// GetUpdateTime returns the updateTime duration configured in NpuCollector.
// updateTime > 0 indicates legacy compatibility mode where all collectors use a unified interval;
// updateTime == 0 indicates using per-group configured intervalSeconds.
func (n *NpuCollector) GetUpdateTime() time.Duration {
	return n.updateTime
}

// SetChains atomically replaces the three collector chains (single, multi, plugin).
func SetChains(single, multi, plugin []MetricsCollector) {
	chainsMu.Lock()
	defer chainsMu.Unlock()

	ChainForSingleGoroutine = single
	ChainForMultiGoroutine = multi
	ChainForCustomPlugin = plugin
}

// GetChainsSnapshot returns a shallow copy of the current three collector chains.
func GetChainsSnapshot() (single, multi, plugin []MetricsCollector) {
	chainsMu.RLock()
	defer chainsMu.RUnlock()

	single = append([]MetricsCollector(nil), ChainForSingleGoroutine...)
	multi = append([]MetricsCollector(nil), ChainForMultiGoroutine...)
	plugin = append([]MetricsCollector(nil), ChainForCustomPlugin...)
	return single, multi, plugin
}

// NewNpuCollector creates and initializes an NpuCollector instance.
func NewNpuCollector(cacheTime time.Duration, updateTime time.Duration,
	deviceParser *container.DevicesParser, dmgr devmanager.DeviceInterface) *NpuCollector {
	CommonCollector := &NpuCollector{
		cache:         cache.New(cacheSize),
		cacheTime:     cacheTime,
		updateTime:    updateTime,
		devicesParser: deviceParser,
		Dmgr:          dmgr,
	}
	InitNpuDevNetPortInfos(CommonCollector)
	return CommonCollector
}

// StartCollect starts all collector goroutines. First initializes the chip list cache,
func StartCollect(group *sync.WaitGroup, ctx context.Context, n *NpuCollector) {
	npuChipInfoInitAtFirstTime(n)
	startCollectSingleGoroutine(group, ctx, n)
	startCollectForMultiGoroutine(group, ctx, n)
	startCollectForPluginGoroutine(group, ctx, n)
}

func startCollectForPluginGoroutine(group *sync.WaitGroup, ctx context.Context, n *NpuCollector) {
	group.Add(1)
	go func() {
		defer group.Done()
		runCollectorLoop(ctx, n, collectorLoopOptions{
			loadChain: func() []MetricsCollector {
				_, _, pluginChain := GetChainsSnapshot()
				return pluginChain
			},
			onStop: func() {
				logger.Info("received the stop signal, stop plugin collect")
			},
			runDueCollectors: func(dueCollectors []MetricsCollector) {
				logger.Debugf("start to collect plugin info, collectors: %v", getCollectorNames(dueCollectors))
				begin := time.Now()
				collectPluginMetricsFor(n, dueCollectors)
				logger.Infof("end to collect plugin info, collectors: %v, time cost: %v",
					getCollectorNames(dueCollectors), time.Since(begin))
			},
		})
	}()
}

func collectPluginMetricsFor(n *NpuCollector, collectors []MetricsCollector) {
	chipList := getChipListCache(n)
	for _, c := range collectors {
		resultChan := make(chan struct{}, 1)
		go func(cur MetricsCollector) {
			cur.CollectToCache(n, chipList)
			resultChan <- struct{}{}
		}(c)
		select {
		case <-resultChan:
			continue
		case <-time.After(maxCollectTimeout):
			logger.Errorf("collect timeout for %v", GetCacheKey(c))
			continue
		}

	}
}

func startCollectForMultiGoroutine(group *sync.WaitGroup, ctx context.Context, n *NpuCollector) {
	chips := getChipListCache(n)

	group.Add(len(chips))
	for _, chip := range chips {
		go func(chip HuaWeiAIChip) {
			defer group.Done()
			runCollectorLoop(ctx, n, collectorLoopOptions{
				loadChain: func() []MetricsCollector {
					_, multiChain, _ := GetChainsSnapshot()
					return multiChain
				},
				onStop: func() {
					logger.Infof("received the stop signal, stop collect network info of npu(%d)", chip.LogicID)
				},
				runDueCollectors: func(dueCollectors []MetricsCollector) {
					logger.Debug("start to collect npu info by hccn_tool, phyID: %v, collectors: %v",
						chip.PhyId, getCollectorNames(dueCollectors))
					begin := time.Now()
					singleChipSlice := []HuaWeiAIChip{chip}
					for _, c := range dueCollectors {
						c.CollectToCache(n, singleChipSlice)
					}
					logger.Infof("end to collect npu info by hccn_tool, phyID: %v, collectors: %v, time cost: %v",
						chip.PhyId, getCollectorNames(dueCollectors), time.Since(begin))
				},
			})
		}(chip)
	}
}

func goroutinePreCollect(collectors []MetricsCollector, n *NpuCollector) {
	chipList := getChipListCache(n)
	for _, c := range collectors {
		c.PreCollect(n, chipList)
	}
}

func goroutinePostCollect(collectors []MetricsCollector, n *NpuCollector) {
	for _, c := range collectors {
		c.PostCollect(n)
	}
}

func startCollectSingleGoroutine(group *sync.WaitGroup, ctx context.Context, n *NpuCollector) {
	group.Add(1)
	go func() {
		defer group.Done()
		runCollectorLoop(ctx, n, collectorLoopOptions{
			loadChain: func() []MetricsCollector {
				singleChain, _, _ := GetChainsSnapshot()
				return singleChain
			},
			onStop: func() {
				logger.Info("received the stop signal, stop npu base info collect")
			},
			runDueCollectors: func(dueCollectors []MetricsCollector) {
				logger.Debugf("start to collect npu info by dcmi, collectors: %v", getCollectorNames(dueCollectors))
				begin := time.Now()
				chipList := getChipListCache(n)
				for _, c := range dueCollectors {
					c.CollectToCache(n, chipList)
				}
				logger.Infof("end to collect npu info by dcmi, collectors: %v, time cost: %v",
					getCollectorNames(dueCollectors), time.Since(begin))
			},
		})
	}()
}

func getCollectorNames(collectors []MetricsCollector) interface{} {
	var names []string
	for _, c := range collectors {
		names = append(names, GetCacheKey(c))
	}
	return names
}

// collectorLoopOptions configures the behavior of the collector loop.
type collectorLoopOptions struct {
	loadChain        func() []MetricsCollector
	onStop           func()
	runDueCollectors func([]MetricsCollector)
}

// runCollectorLoop is the core event loop for all collector goroutines.
func runCollectorLoop(ctx context.Context, n *NpuCollector, opts collectorLoopOptions) {
	currentChain := opts.loadChain()
	goroutinePreCollect(currentChain, n)
	defer func() {
		goroutinePostCollect(currentChain, n)
	}()

	schedule := buildSchedule(currentChain)
	schedule.markAllDue()

	reloadCh := subscribeConfigReload()
	defer unsubscribeConfigReload(reloadCh)

	for {
		result := waitForNextSignal(ctx, schedule.nextWaitDuration(), reloadCh)
		if result == wakeByContext {
			if opts.onStop != nil {
				opts.onStop()
			}
			return
		}
		if result == wakeByConfigReload {
			handleConfigReload(&currentChain, &schedule, n, opts)
			drainReloadSignal(reloadCh)
			continue
		}
		// wakeByTimer: timer expired, execute due collectors
		now := time.Now()
		dueCollectors := schedule.popDue(now)
		if len(dueCollectors) == 0 {
			continue
		}
		opts.runDueCollectors(dueCollectors)
		schedule.updateNext(dueCollectors, now)
	}
}

func handleConfigReload(currentChain *[]MetricsCollector, schedule *collectorSchedule,
	n *NpuCollector, opts collectorLoopOptions) {
	goroutinePostCollect(*currentChain, n)
	*currentChain = opts.loadChain()
	goroutinePreCollect(*currentChain, n)
	*schedule = buildSchedule(*currentChain)
	schedule.markAllDue()
}

// npuChipInfoInitAtFirstTime When first enter, the cache data is empty,
// need to get the data from the device, and build the cache
func npuChipInfoInitAtFirstTime(n *NpuCollector) {
	npuChipInfoInit.Do(func() {
		_, err := n.cache.Get(npuListCacheKey)
		if err != nil {
			logger.Debug("no cache in first time, start to collect chip list and rebuild cache")
			begin := time.Now()
			npuInfo := getNPUChipList(n.Dmgr)
			if err := n.cache.Set(npuListCacheKey, npuInfo, -1); err != nil {
				logger.Error(err)
			} else {
				logger.Infof(UpdateCachePattern+", collect time cost: %v", npuListCacheKey, time.Since(begin))
			}
			logger.Debug("rebuild cache successfully")
		}
	})
}

// InitCardInfo init card info
func InitCardInfo(group *sync.WaitGroup, ctx context.Context, n *NpuCollector) {

	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(updateTimeForCardIds)
		defer ticker.Stop()
		for {
			logger.Info("start to collect npu chip list info")
			select {
			case <-ctx.Done():
				logger.Info("received the stop signal,stop card info collect")
				return
			default:
				begin := time.Now()
				npuInfo := getNPUChipList(n.Dmgr)
				if err := n.cache.Set(npuListCacheKey, npuInfo, -1); err != nil {
					logger.Error(err)
				} else {
					logger.Infof(UpdateCachePattern+", collect time cost: %v", npuListCacheKey, time.Since(begin))
				}
				if _, ok := <-ticker.C; !ok {
					logger.Errorf(tickerFailedPattern, npuListCacheKey)
					return
				}
			}
		}
	}()
}

func getNPUChipList(dmgr devmanager.DeviceInterface) (npuInfo []HuaWeiAIChip) {
	chipList := make([]HuaWeiAIChip, 0)

	devNum, devList, err := dmgr.GetDeviceList()
	if err != nil || devNum == 0 {
		logger.Errorf("failed to get npu info, error is: %v", err)
		return chipList
	}

	chipListIDs := make([]int32, 0)

	for _, logicID := range devList {
		var chip HuaWeiAIChip
		chip.LogicID = logicID
		chip.MainBoardId = dmgr.GetMainBoardId()
		cardID, deviceID, _ := dmgr.GetCardIDDeviceID(logicID)
		chip.CardId = cardID

		setPhyId(&chip, dmgr, deviceID)
		setChipInfo(&chip, dmgr, deviceID)
		setBoardInfo(&chip, dmgr, deviceID)
		setVdieID(&chip, dmgr)
		assemblevNPUInfo(dmgr, logicID, &chip)
		setPCIeBusInfo(logicID, dmgr, &chip)
		setElabelInfo(&chip, dmgr)
		setProductType(&chip, dmgr)

		chipList = append(chipList, chip)
		chipListIDs = append(chipListIDs, logicID)
	}

	logger.Debugf("flush chip info list succeeded, chip num is : %v, chipLogicIDs: %v",
		len(chipList), chipListIDs)
	return chipList
}

func logSetError(domain string, chip *HuaWeiAIChip, deviceID int32, err error, msg string) {
	if chip.CardId == -1 {
		logger.Errorf("%s of logicID: %v failed: %v", msg, chip.LogicID, err)
	} else {
		logger.Errorf("%s of card: %v device:%v failed: %v", msg, chip.CardId, deviceID, err)
	}
}

func setBoardInfo(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface, deviceID int32) {
	boardInfo, err := dmgr.GetBoardInfo(chip.LogicID)
	if err != nil {
		logSetError("board info", chip, deviceID, err, "get board info")
		boardInfo = common.BoardInfo{}
	}
	chip.BoardInfo = &boardInfo
}

func setVdieID(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface) {
	vdieID, err := dmgr.GetDieID(chip.LogicID, dcmi.VDIE)
	if err != nil {
		logger.Debugf("get vdie ID of logicID: %v failed: %v", chip.LogicID, err)
		return
	}
	chip.VDieID = vdieID
}

func setPhyId(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface, deviceID int32) {
	phyID, err := dmgr.GetPhysicIDFromLogicID(chip.LogicID)
	if err != nil {
		logSetError("phy ID", chip, deviceID, err, "get phy ID")
		return
	}
	chip.PhyId = phyID

	if chip.CardId == -1 {
		chip.DeviceID = chip.LogicID
	} else {
		chip.DeviceID = phyID
	}
}

func setChipInfo(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface, deviceID int32) {
	chipInfo, err := dmgr.GetChipInfo(chip.LogicID)
	if err != nil {
		logSetError("chip info", chip, deviceID, err, "get chip info")
		chipInfo = &common.ChipInfo{}
	}
	chip.ChipInfo = chipInfo
}

func setPCIeBusInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	productTypes := dmgr.GetProductTypeArray()
	pcieInfo, err := dmgr.GetPCIeBusInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			logger.Debugf("pcie bus info is not supported on %s", common.Atlas200ISoc)
			hwChip.PCIeBusInfo = ""
			return
		}
		logger.LogfWithOptions(logger.ErrorLevel, fetchPcieOptions, err.Error())
		pcieInfo = ""
	} else {
		hwlog.ResetErrCnt(pcieDomain, fetchTimeout)
	}
	hwChip.PCIeBusInfo = pcieInfo
}

func setElabelInfo(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface) {
	var elabelInfo common.ElabelInfo
	var err error

	if chip.CardId != -1 {
		elabelInfo, err = dmgr.GetCardElabelV2(chip.CardId)
		if err != nil {
			logger.Errorf("get elabel info of cardID: %v failed: %v", chip.CardId, err)
			chip.ElabelInfo = &common.ElabelInfo{SerialNumber: "NA"}
			return
		}
	} else {
		elabelInfo, err = dmgr.GetCardElabelV2(chip.LogicID)
		if err != nil {
			logger.Errorf("get elabel info of logicID: %v failed: %v", chip.LogicID, err)
			chip.ElabelInfo = &common.ElabelInfo{SerialNumber: "NA"}
			return
		}
	}

	chip.ElabelInfo = &common.ElabelInfo{
		SerialNumber: elabelInfo.SerialNumber,
	}
}

func setProductType(chip *HuaWeiAIChip, dmgr devmanager.DeviceInterface) {
	if dmgr.GetDevType() != api.Ascend310P {
		logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{
			Domain:    "setProductType",
			ID:        dmgr.GetDevType(),
			MaxCounts: 1,
		},
			"%v does not support product type info", utils.MaskDevType(dmgr.GetDevType()))
		return
	}

	if productType, ok := getFromCache(DomainForProductType, chip.LogicID).(string); ok {
		chip.ProductType = productType
		return
	}

	productType, err := dmgr.GetProductType(chip.LogicID)
	if err != nil {
		if chip.CardId == -1 {
			logger.LogfWithOptions(logger.ErrorLevel, logger.LogOptions{
				Domain: DomainForProductType,
				ID:     chip.LogicID},
				"get product type info of logicID: %v failed: %v", chip.LogicID, err)
		} else {
			logger.LogfWithOptions(logger.ErrorLevel, logger.LogOptions{
				Domain: DomainForProductType,
				ID:     chip.CardId},
				"get product type info of card: %v failed: %v", chip.CardId, err)
		}
		return
	}
	chip.ProductType = productType
	saveToCache(DomainForProductType, chip.LogicID, productType)
	hwlog.ResetErrCnt(DomainForProductType, chip.LogicID)
}

func assemblevNPUInfo(dmgr devmanager.DeviceInterface, logicID int32, baseChipInfo *HuaWeiAIChip) {
	devType := dmgr.GetDevType()
	if devType != api.Ascend310P {
		return
	}
	vDevInfos, err := dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		logger.Warnf("failed to get virtual device info,logicID(%d),err: %v", logicID, err)
		baseChipInfo.VDevInfos = nil
	}
	if vDevInfos.TotalResource.VDevNum == 0 {
		baseChipInfo.VDevInfos = &common.VirtualDevInfo{}
	}
	baseChipInfo.VDevInfos = &vDevInfos
}

// GetChipListWithVNPU get chip list with vnpu
func GetChipListWithVNPU(n *NpuCollector) []HuaWeiAIChip {
	result := make([]HuaWeiAIChip, 0)
	chips := getChipListCache(n)
	devType := n.Dmgr.GetDevType()

	for _, chipInfo := range chips {
		isNeedHandleVnpu := devType == api.Ascend310P && chipInfo.VDevInfos != nil &&
			len(chipInfo.VDevInfos.VDevActivityInfo) > 0

		if !isNeedHandleVnpu {
			result = append(result, chipInfo)
			continue
		}

		for _, activityVDev := range chipInfo.VDevInfos.VDevActivityInfo {
			vDevInfo := chipInfo
			activityVDevCopy := activityVDev
			vDevInfo.VDevActivityInfo = &activityVDevCopy
			result = append(result, vDevInfo)
		}
	}

	return result

}
func getChipListCache(n *NpuCollector) []HuaWeiAIChip {
	obj, err := n.cache.Get(npuListCacheKey)
	if err != nil {
		logger.Errorf("get npu chip list from cache failed,err is : %v", err)
		return make([]HuaWeiAIChip, 0)
	}
	if obj == nil {
		logger.LogfWithOptions(logger.ErrorLevel, logger.LogOptions{Domain: "getChipListCache"},
			"there is no chip list info in cache,please check collect logs")
		return make([]HuaWeiAIChip, 0)
	}

	chipList, ok := obj.([]HuaWeiAIChip)
	if !ok {
		logger.Errorf("error npu chip info cache and convert failed,real type is (%T)", obj)
		n.cache.Delete(npuListCacheKey)
		return make([]HuaWeiAIChip, 0)
	}
	// if cache is empty or nil, return empty list
	if len(chipList) == 0 {
		return make([]HuaWeiAIChip, 0)
	}
	return chipList
}

func saveToCache(domain string, logicID int32, value interface{}) {
	key := domain + strconv.Itoa(int(logicID))
	npuInfoCache.Store(key, value)
}

func getFromCache(domain string, logicID int32) interface{} {
	key := domain + strconv.Itoa(int(logicID))
	value, ok := npuInfoCache.Load(key)
	if !ok {
		return nil
	}
	return value
}
