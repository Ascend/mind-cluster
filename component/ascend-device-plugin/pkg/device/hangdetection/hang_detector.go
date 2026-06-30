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

// Package hangdetection implements NPU hang detection logic
package hangdetection

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
)

const (
	utimeIndex   = 13
	stimeIndex   = 14
	procStatPath = "/proc/%d/stat"

	// If you want to set this variable to a different value, you need to modify the isHangConditionMet detection metric,
	// as these metrics are calculated based on a 60 second span
	hangDetectInterval = 60

	roceTxAllPktNum = "roce_tx_all_pkt_num"
	roceRxAllPktNum = "roce_rx_all_pkt_num"
	txBusiFlitNum   = "tx_busi_flit_num"
	rxBusiFlitNum   = "rx_busi_flit_num"

	procAuxvPath = "/proc/%d/auxv"

	rightParenthesis   = ")"
	postCommUtimeIndex = utimeIndex - 2
	postCommStimeIndex = stimeIndex - 2
)

// HangDetector implements NPU hang detection
type HangDetector struct {
	dmgr            devmanager.DeviceInterface
	npuDevPortInfos map[int][]int
}

var (
	hangDetector   = &HangDetector{npuDevPortInfos: make(map[int][]int)}
	hangStateMap   = make(map[int32]*HangState)
	hangStateMapMu sync.Mutex
	clkTck         = int64(100)

	npuFaultCacheMu sync.Mutex
	npuFaultCache   = make([]*npuCommon.DevFaultInfo, 0)

	logicIdMap = sync.Map{}
)

// StartHangDetectionProducer starts a background goroutine that periodically collects hang detection metrics for every registered logicID
func StartHangDetectionProducer(ctx context.Context, dmgr devmanager.DeviceInterface) {
	if dmgr == nil {
		hwlog.RunLog.Error("hang detection producer start failed: dmgr is nil")
		return
	}
	hangDetector.dmgr = dmgr
	npuDevPortInfos, err := hangDetector.getNpuDevNetPortInfos()
	if err != nil {
		hwlog.RunLog.Errorf("hang detection get NPU port info failed: %v", err)
	}
	if npuDevPortInfos != nil {
		hangDetector.npuDevPortInfos = npuDevPortInfos
	}

	setClkTck()
	common.LoadHangDetectionConfigFromFile()

	runHangDetectionProducer(ctx)
}

func runHangDetectionProducer(ctx context.Context) {
	hwlog.RunLog.Info("hang detection producer start")
	ticker := time.NewTicker(hangDetectInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("hang detection producer stop")
			return
		case <-ticker.C:
			for _, logicID := range snapshotRegisteredLogicIDs() {
				hwlog.RunLog.Debugf("hang detection start, logicID=%d", logicID)
				hangDetector.detectNPU(logicID)
				hwlog.RunLog.Debugf("hang detection end, logicID=%d", logicID)
			}
		}
	}
}

func (hd *HangDetector) getNpuDevNetPortInfos() (map[int][]int, error) {
	if common.ParamOption.RealCardType != api.Ascend910A5 {
		return nil, nil
	}
	_, npuList, err := hd.dmgr.GetDeviceList()
	if err != nil {
		return nil, err
	}
	var err1 error = nil
	var netPortInfos map[int][]int = nil
	for _, logicID := range npuList {
		netPortInfos, err1 = hccn.GetNpuDevNetPortInfo(logicID)
		if err1 != nil {
			continue
		}
		return netPortInfos, nil
	}
	return nil, err1
}

// DetectNPU performs one round of hang detection for a single NPU
func (hd *HangDetector) detectNPU(logicID int32) {
	if !common.IsHangDetectionEnabled() {
		hwlog.RunLog.Debugf("hang detection disabled, logicID=%d", logicID)
		hd.npuHangEventDisappear(logicID, nil)
		return
	}
	procInfo, err := hd.dmgr.GetDevProcessInfo(logicID)
	if err != nil || procInfo == nil {
		hwlog.RunLog.Errorf("hang detection get process info failed, logicID=%d: %v", logicID, err)
		hd.npuHangEventDisappear(logicID, nil)
		return
	}
	hwlog.RunLog.Debugf("hang detection process num, logicID=%d, procInfo=%v", logicID, procInfo)
	if procInfo.ProcNum == 0 {
		hd.npuHangEventDisappear(logicID, nil)
		return
	}

	hd.refreHangStateIfProcessChanged(logicID, extractAndSortPids(procInfo))

	var metrics HangMetrics = HangMetrics{}
	hd.collectMetrics(logicID, procInfo, &metrics)
	hwlog.RunLog.Debugf("hang detection metrics, logicID=%d, metrics=%v", logicID, metrics)

	if hd.isHangConditionMet(logicID, &metrics) {
		hd.npuHangEventOccur(logicID, &metrics)
	} else {
		hd.npuHangEventDisappear(logicID, &metrics)
	}
}

func (hd *HangDetector) refreHangStateIfProcessChanged(logicID int32, curPIDs []int32) {
	hangStateMapMu.Lock()
	defer hangStateMapMu.Unlock()
	state, ok := hangStateMap[logicID]
	if !ok {
		state = &HangState{LogicID: logicID}
		hangStateMap[logicID] = state
	}
	if state.PIDs == nil {
		state.PIDs = curPIDs
		return
	}
	if common.SliceEqual[int32](state.PIDs, curPIDs) {
		return
	}
	hwlog.RunLog.Infof("process set changed, reset hang baseline, logicID=%d", logicID)
	state.PIDs = curPIDs
	state.Metrics = nil
}

func extractAndSortPids(procInfo *npuCommon.DevProcessInfo) []int32 {
	pids := make([]int32, 0, len(procInfo.DevProcArray))
	for _, proc := range procInfo.DevProcArray {
		pids = append(pids, proc.Pid)
	}
	sort.Slice(pids, func(i, j int) bool { return pids[i] < pids[j] })
	return pids
}

func (hd *HangDetector) collectMetrics(logicID int32, procInfo *npuCommon.DevProcessInfo, metrics *HangMetrics) {
	if metrics == nil {
		return
	}
	metrics.ProcessNum = int8(procInfo.ProcNum)
	hd.collectUtilization(logicID, metrics)
	hd.collectMemoryUsage(logicID, metrics)
	hd.collectTraffic(logicID, metrics)
	hd.collectCPUTime(procInfo, metrics)
}

func (hd *HangDetector) collectUtilization(logicID int32, metrics *HangMetrics) {
	utilizationRate, err := hd.dmgr.GetDeviceUtilizationRate(logicID, npuCommon.AICore)
	if err != nil {
		hwlog.RunLog.Errorf("logicID=%d get utilization failed: %v", logicID, err)
		return
	}
	metrics.AICoreUsage = int32(utilizationRate)
}

func (hd *HangDetector) collectMemoryUsage(logicID int32, metrics *HangMetrics) {
	utilizationRate, err := hd.dmgr.GetDeviceUtilizationRate(logicID, npuCommon.HbmUtilization)
	if err != nil {
		hwlog.RunLog.Errorf("logicID=%d get hbm info failed: %v", logicID, err)
		return
	}
	metrics.MemoryUsage = int32(utilizationRate)
}

func (hd *HangDetector) collectTraffic(logicID int32, metrics *HangMetrics) {
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		hd.collectUBTraffic(logicID, metrics)
		return
	}
	hd.collectRoCETraffic(logicID, metrics)
}

func (hd *HangDetector) collectRoCETraffic(logicID int32, metrics *HangMetrics) {
	phyID, err := hd.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("hang detection get phyID failed, logicID=%d: %v", logicID, err)
		return
	}
	statInfo, err := hccn.GetNPUStatInfo(phyID)
	if err != nil {
		hwlog.RunLog.Errorf("hang detection get roce stat failed, logicID=%d: %v", logicID, err)
		return
	}
	metrics.RoceTxPkts = uint64(statInfo[roceTxAllPktNum])
	metrics.RoceRxPkts = uint64(statInfo[roceRxAllPktNum])
}

func (hd *HangDetector) collectUBTraffic(logicID int32, metrics *HangMetrics) {
	if len(hd.npuDevPortInfos) == 0 {
		hwlog.RunLog.Debugf("empty UB port info, logicID=%d", logicID)
		return
	}
	var currentTotalTx uint64
	var currentTotalRx uint64
	for udieID, portIDs := range hd.npuDevPortInfos {
		for _, portID := range portIDs {
			ubStat, err := hccn.GetNPUUbStatInfo(logicID, int32(udieID), int32(portID))
			if err != nil {
				hwlog.RunLog.Debugf("hang detection get UB stat failed, logicID=%d, udieID=%d, portID=%d: %v",
					logicID, udieID, portID, err)
				continue
			}
			if txVal := hccn.GetIntDataFromStr(ubStat[txBusiFlitNum], txBusiFlitNum); txVal > 0 {
				currentTotalTx += uint64(txVal)
			}
			if rxVal := hccn.GetIntDataFromStr(ubStat[rxBusiFlitNum], rxBusiFlitNum); rxVal > 0 {
				currentTotalRx += uint64(rxVal)
			}
		}
	}
	metrics.UBTxFlits = currentTotalTx
	metrics.UBRxFlits = currentTotalRx
}

func (hd *HangDetector) collectCPUTime(procInfo *npuCommon.DevProcessInfo, metrics *HangMetrics) {
	var totalCPUTime int64
	for _, proc := range procInfo.DevProcArray {
		cpuTime, err := getProcessCPUTime(proc.Pid)
		if err != nil {
			hwlog.RunLog.Errorf("hang detection get CPU time failed, pid=%d: %v", proc.Pid, err)
			continue
		}
		totalCPUTime += cpuTime
	}
	metrics.CPUTime = totalCPUTime
}

func (hd *HangDetector) isHangConditionMet(logicID int32, curMetrics *HangMetrics) bool {
	state := hd.getOrCreateHangState(logicID)
	lastMetric := state.Metrics
	isHangConditionMet := false
	if lastMetric != nil {
		threshold := common.GetHangDetectionThreshold()

		memoryDelta := curMetrics.MemoryUsage - lastMetric.MemoryUsage

		trafficDelta := uint64(0)
		if common.ParamOption.RealCardType == api.Ascend910A5 {
			trafficDelta = (curMetrics.UBTxFlits - lastMetric.UBTxFlits) + (curMetrics.UBRxFlits - lastMetric.UBRxFlits)
		} else {
			trafficDelta = (curMetrics.RoceTxPkts - lastMetric.RoceTxPkts) + (curMetrics.RoceRxPkts - lastMetric.RoceRxPkts)
		}

		cpuTimeDelta := curMetrics.CPUTime - lastMetric.CPUTime

		hwlog.RunLog.Debugf("logicID=%d, memoryDelta=%d(%%), trafficDelta=%d(pkt/min), cpuTimeDelta=%d(s/min), aiCoreUsage=%d(%%)",
			logicID, memoryDelta, trafficDelta, cpuTimeDelta, int32(curMetrics.AICoreUsage))

		isHangConditionMet = curMetrics.ProcessNum > 0 &&
			curMetrics.AICoreUsage < int32(threshold.AICoreUtilization) &&
			memoryDelta <= int32(threshold.HbmMemoryDelta) &&
			trafficDelta < uint64(threshold.TrafficDelta) &&
			cpuTimeDelta < int64(threshold.CPUTimeDelta)
	}
	return isHangConditionMet
}

func (hd *HangDetector) npuHangEventOccur(logicID int32, metrics *HangMetrics) {
	state := hd.getOrCreateHangState(logicID)
	threshold := common.GetHangDetectionThreshold()

	hangStateMapMu.Lock()
	state.HangCount++
	hwlog.RunLog.Infof("npu hang condition met, logicID=%d, hangCount=%d, metrics=%v, preMetrics=%v",
		logicID, state.HangCount, metrics, state.Metrics)
	state.Metrics = metrics
	if state.HangCount >= threshold.DetectDuration && !state.IsFault {
		state.IsFault = true
		hd.reportHangFault(logicID)
	}
	hangStateMapMu.Unlock()

}

func (hd *HangDetector) npuHangEventDisappear(logicID int32, metrics *HangMetrics) {
	state := hd.getOrCreateHangState(logicID)

	hangStateMapMu.Lock()
	if state.IsFault {
		state.IsFault = false
		hd.reportHangRecover(logicID)
	}
	if state.HangCount > 0 {
		hwlog.RunLog.Infof("npu hang condition not met, reset logicID=%d hangCount=%d to 0, metrics=%v, preMetrics=%v",
			logicID, state.HangCount, metrics, state.Metrics)
	}
	state.HangCount = 0
	state.Metrics = metrics
	hangStateMapMu.Unlock()
}

func (hd *HangDetector) reportHangFault(logicID int32) {
	faultInfo := npuCommon.DevFaultInfo{
		EventID:         npuCommon.HangFaultCode,
		LogicID:         logicID,
		Assertion:       npuCommon.FaultOccur,
		AlarmRaisedTime: time.Now().Unix(),
	}
	hwlog.RunLog.Infof("report NPU hang fault, logicID=%d, faultCode=0x%X", logicID, npuCommon.HangFaultCode)
	appendHangFaultCache(&faultInfo)
}

func (hd *HangDetector) reportHangRecover(logicID int32) {
	faultInfo := npuCommon.DevFaultInfo{
		EventID:         npuCommon.HangFaultCode,
		LogicID:         logicID,
		Assertion:       npuCommon.FaultRecover,
		AlarmRaisedTime: time.Now().Unix(),
	}
	hwlog.RunLog.Infof("report NPU hang recover, logicID=%d, faultCode=0x%X", logicID, npuCommon.HangFaultCode)
	appendHangFaultCache(&faultInfo)
}

func (hd *HangDetector) getOrCreateHangState(logicID int32) *HangState {
	hangStateMapMu.Lock()
	defer hangStateMapMu.Unlock()
	state, ok := hangStateMap[logicID]
	if !ok {
		state = &HangState{LogicID: logicID}
		hangStateMap[logicID] = state
	}
	return state
}

func getProcessCPUTime(pid int32) (int64, error) {
	if clkTck <= 0 {
		return 0, fmt.Errorf("invalid clkTck: %d", clkTck)
	}
	data, err := utils.LoadFile(fmt.Sprintf(procStatPath, pid))
	if err != nil {
		return 0, fmt.Errorf("load proc stat file failed: %v", err)
	}
	statStr := string(data)

	lastRightParen := strings.LastIndex(statStr, rightParenthesis)
	if lastRightParen < 0 {
		return 0, fmt.Errorf("invalid proc stat format for pid %d: missing comm field", pid)
	}
	fields := strings.Fields(statStr[lastRightParen+1:])
	if len(fields) <= postCommStimeIndex {
		return 0, fmt.Errorf("invalid proc stat format for pid %d", pid)
	}
	utime, err := strconv.ParseInt(fields[postCommUtimeIndex], common.BaseDec, common.BitSize)
	if err != nil {
		return 0, fmt.Errorf("parse utime failed: %v", err)
	}
	stime, err := strconv.ParseInt(fields[postCommStimeIndex], common.BaseDec, common.BitSize)
	if err != nil {
		return 0, fmt.Errorf("parse stime failed: %v", err)
	}
	return (utime + stime) / clkTck, nil
}

func setClkTck() {
	sysClkTck, err := getSysClockTicks()
	if err != nil {
		hwlog.RunLog.Errorf("get clkTck failed: %v", err)
		clkTck = 0
		return
	}
	hwlog.RunLog.Infof("set clkTck to %d", sysClkTck)
	clkTck = sysClkTck
}

func getSysClockTicks() (int64, error) {
	data, err := utils.LoadFile(fmt.Sprintf(procAuxvPath, os.Getpid()))
	if err != nil {
		return 0, fmt.Errorf("load proc auxv file failed: %v", err)
	}
	reader := bytes.NewReader(data)

	var entry struct {
		Type  uint64
		Value uint64
	}
	const at_clktck = 17
	for {
		err = binary.Read(reader, binary.NativeEndian, &entry)
		if err != nil {
			break
		}
		if entry.Type == 0 {
			err = fmt.Errorf("cant find AT_CLKTCK entry")
			break
		}
		if entry.Type == at_clktck {
			return int64(entry.Value), nil
		}
	}
	return 0, err
}

// GetAndCleanAllHangFaultCache returns all DevFaultInfo cache for the given logicID
func GetAndCleanAllHangFaultCache() []*npuCommon.DevFaultInfo {
	npuFaultCacheMu.Lock()
	defer npuFaultCacheMu.Unlock()
	ret := npuFaultCache
	npuFaultCache = make([]*npuCommon.DevFaultInfo, 0, len(ret))
	return ret
}

func appendHangFaultCache(faultInfo *npuCommon.DevFaultInfo) {
	npuFaultCacheMu.Lock()
	defer npuFaultCacheMu.Unlock()
	npuFaultCache = append(npuFaultCache, faultInfo)
}

// RegisterLogicIDForProducer registers the given logicID for hang detection producer
func RegisterLogicIDForProducer(logicID int32) {
	logicIdMap.Store(logicID, struct{}{})
}

func snapshotRegisteredLogicIDs() []int32 {
	logicIDs := make([]int32, 0)
	logicIdMap.Range(func(key, value any) bool {
		logicIDs = append(logicIDs, key.(int32))
		return true
	})
	return logicIDs
}
