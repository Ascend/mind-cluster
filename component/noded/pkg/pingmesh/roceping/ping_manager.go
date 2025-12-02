/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmesh/types"
)

var csvColumnNames = []string{
	"pingTaskId", "srcType", "srcAddr", "dstType", "dstAddr", "minDelay", "maxDelay", "avgDelay",
	"minLossRate", "maxLossRate", "avgLossRate", "timestamp",
}

// PingManager for ping data manager
type PingManager struct {
	superPodId   uint32
	rackId       uint32
	serverIndex  uint32
	nodeName     string
	nodeType     string
	devType      string
	k8sClient    *kubeclient.ClientK8s
	wg           *sync.WaitGroup
	executors    []*IcmpPingExecutor
	pingList     []types.PingItem
	curPolicy    *types.HccspingMeshPolicy
	commandChan  chan *types.HccspingMeshPolicy
	recordChan   chan statisticData
	lastSaveTime int64
	writer       *hwlog.CustomLogger
}

// NewPingManager create PingManager instance
func NewPingManager(superPodId, rackId, serverIndex uint32, client *kubeclient.ClientK8s, devType string) *PingManager {
	if client == nil {
		hwlog.RunLog.Error("the k8s client is empty, cannot new ping manager for roce ping")
		return nil
	}
	nodeInfo, errInfo := client.GetNodeWithCache()
	if errInfo != nil {
		hwlog.RunLog.Errorf("get node resource info from k8s failed, err: %v", errInfo)
		return nil
	}
	acType, exist := nodeInfo.Labels[acceleratorTypeKey]
	if !exist {
		hwlog.RunLog.Warnf("the node label %s is not exist, cannot decide node type", acceleratorTypeKey)
		// the first time the expansion scenario is added, the nodes may not have been manually labeled yet
	}
	if exist && !strings.Contains(strings.ToLower(acType), labelPrefix900SuperPodA5) {
		hwlog.RunLog.Warnf("the node label %s is not for super pod a5 scene, roce ping task is not support", acType)
		return nil
	}
	return &PingManager{
		superPodId:  superPodId,
		rackId:      rackId,
		serverIndex: serverIndex,
		nodeName:    client.NodeName,
		nodeType:    acType,
		devType:     devType,
		k8sClient:   client,
		wg:          &sync.WaitGroup{},
		executors:   make([]*IcmpPingExecutor, 0),
		commandChan: make(chan *types.HccspingMeshPolicy, 1),
		recordChan:  make(chan statisticData),
	}
}

// SetFileWriter set file log writer for roce ping
func (m *PingManager) SetFileWriter(writer *hwlog.CustomLogger) {
	if m == nil {
		return
	}
	m.writer = writer
}

// GetDevType get device type
func (m *PingManager) GetDevType() string {
	if m == nil {
		return ""
	}
	return m.devType
}

// CheckNodeLabelSupported for check cur node label accelerator-type is supported
func (m *PingManager) CheckNodeLabelSupported() bool {
	if m == nil || m.k8sClient == nil {
		hwlog.RunLog.Warnf("pingManager or k8sClient is empty")
		return false
	}
	nodeInfo, errInfo := m.k8sClient.GetNodeWithCache()
	if errInfo != nil {
		hwlog.RunLog.Warnf("get node resource info from k8s failed, err: %v", errInfo)
		return false
	}

	acType, exist := nodeInfo.Labels[acceleratorTypeKey]
	if !exist {
		hwlog.RunLog.Warnf("the node label %s is not exist, cannot decide node type", acceleratorTypeKey)
		return false
	}

	if !strings.Contains(strings.ToLower(acType), labelPrefix900SuperPodA5) {
		hwlog.RunLog.Warnf("the node label %s is not for super pod a5 scene, roce ping task is not support", acType)
		return false
	}
	return true
}

// IsInPingListRange for check cur node is in ping list range
func (m *PingManager) IsInPingListRange() bool {
	pingRangeFile, err := slownet.GetPingListRangePath()
	if err != nil {
		hwlog.RunLog.Errorf("get ping list range file path failed, err: %v", err)
		return false
	}

	retry := 0
	for ; retry < maxRetryTimes; retry++ {
		err = slownet.CheckIsExistAndValid(pingRangeFile)
		if err == nil {
			break
		}
		hwlog.RunLog.Errorf("retry=%d, err: %v", retry, err)
		time.Sleep(waitTimesForGenerate * time.Second)
	}
	if retry >= maxRetryTimes {
		hwlog.RunLog.Error("ping list range file is not valid, waiting timed out")
		return false
	}

	data, err := utils.ReadLimitBytes(pingRangeFile, maxFileSize)
	if err != nil {
		hwlog.RunLog.Errorf("read data from ping list range file %s failed, err: %v", pingRangeFile, err)
		return false
	}
	pingRange := make(map[string][]string)
	if err = json.Unmarshal(data, &pingRange); err != nil {
		hwlog.RunLog.Errorf("unmarshal data from ping list range file %s failed, err: %v", pingRangeFile, err)
		return false
	}

	curSuperPodId := strconv.Itoa(int(m.superPodId))
	curServerIndex := strconv.Itoa(int(m.serverIndex))
	for superPodId, serverList := range pingRange {
		if superPodId != curSuperPodId {
			continue
		}
		for _, serverIndex := range serverList {
			if serverIndex == curServerIndex {
				return true
			}
		}
	}

	return false
}

// GetCurPolicy get current ping policy
func (m *PingManager) GetCurPolicy() *types.HccspingMeshPolicy {
	if m == nil {
		return nil
	}
	return m.curPolicy
}

// Start run ping executor
func (m *PingManager) Start(stopCh <-chan struct{}) {
	var currentStop chan struct{} = nil
	for {
		select {
		case <-stopCh:
			if currentStop != nil {
				close(currentStop)
				m.wg.Wait()
				currentStop = nil
				hwlog.RunLog.Info("wait for all tasks stop done success")
			}
			return
		case cmd := <-m.commandChan:
			hwlog.RunLog.Infof("received new cmd: %v", cmd)
			if currentStop != nil {
				close(currentStop)
				m.wg.Wait()
				currentStop = nil
				hwlog.RunLog.Info("wait for all tasks stop done success")
			}
			m.stopPingExecutors(currentStop)
			if cmd.Config.Activate == types.ActivateOff {
				currentStop = nil
				continue
			}
			m.curPolicy = cmd
			currentStop = make(chan struct{})
			m.startPingExecutors(currentStop)
			m.wg.Add(1)
			go m.startCollect(currentStop)
		}
	}
}

func (m *PingManager) startPingExecutors(stopCh chan struct{}) {
	allPingList := make([]types.PingItem, 0)
	for _, pingList := range m.curPolicy.DestAddrMap {
		allPingList = append(allPingList, pingList...)
	}
	const maxPingListSize = 16
	size := len(allPingList)
	if size > maxPingListSize {
		hwlog.RunLog.Warnf("currnet node ping list item size is %d, which exceeds the limit %d, will drop them",
			len(allPingList), maxPingListSize)
		size = maxPingListSize
	}

	m.executors = make([]*IcmpPingExecutor, size)
	m.pingList = allPingList[:size]

	for i := 0; i < len(m.executors); i++ {
		pingItem := m.pingList[i]
		operator := NewOperator(pingItem.DstAddr, pingItem.SrcAddr, m.curPolicy.Config.TaskInterval)
		m.executors[i] = NewIcmpPingExecutor(stopCh, 0, operator)
	}
	for _, executor := range m.executors {
		m.wg.Add(1)
		go executor.startPingTask(m.wg)
	}
}

func (m *PingManager) stopPingExecutors(stopCh chan struct{}) {
	if len(m.executors) == 0 {
		hwlog.RunLog.Info("the executor list is empty, no need stop")
		return
	}
	if stopCh != nil {
		close(stopCh)
	}
}

func (m *PingManager) startCollect(stopCh chan struct{}) {
	defer m.wg.Done()
	logHeader := fmt.Sprintf("ping manager collect task")
	for _, executor := range m.executors {
		m.wg.Add(1)
		go executor.getPingResultInfo(m.wg, m.recordChan)
	}
	for {
		select {
		case <-stopCh:
			hwlog.RunLog.Infof("%s received stop signal, stop startCollect", logHeader)
			return
		case data := <-m.recordChan:
			m.writeToCsv(data.record)
			m.writeToLog(data.result)
		}
	}
}

// UpdateConfig for updating the ping policy
func (m *PingManager) UpdateConfig(cfg *types.HccspingMeshPolicy) {
	if cfg == nil {
		return
	}
	m.commandChan <- cfg
}

func (m *PingManager) writeToLog(data string) {
	if m.writer == nil {
		hwlog.RunLog.Warnf("roce ping data cannot write to log, writer is empty")
		return
	}
	m.writer.Info(data)
}
func (m *PingManager) writeToCsv(record []string) {
	const defaultPerm = 0644
	appendMode, openFlag := m.calcAppendModeAndOpenFlag()
	pingResultCsv, _, err := m.prepareResultFilePaths(appendMode)
	if err != nil {
		hwlog.RunLog.Errorf("get result file path failed, err: %v", err)
		return
	}
	f, err := os.OpenFile(pingResultCsv, openFlag, defaultPerm)
	defer func() {
		if f == nil {
			return
		}
		if errClose := f.Close(); errClose != nil {
			hwlog.RunLog.Errorf("close file %s failed, err: %v", pingResultCsv, errClose)
			return
		}
	}()
	if err != nil {
		hwlog.RunLog.Errorf("open file %s failed, err:%v", pingResultCsv, err)
		return
	}
	err = f.Chmod(defaultPerm)
	if err != nil {
		hwlog.RunLog.Errorf("chmod file %s failed, err:%v", pingResultCsv, err)
		return
	}

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()
	if !appendMode {
		if errWrite := csvWriter.Write(csvColumnNames); errWrite != nil {
			hwlog.RunLog.Errorf("write record csv column title to file %s failed, err: %v",
				pingResultCsv, errWrite)
			return
		}
		hwlog.RunLog.Infof("write record csv column title to file %s success", pingResultCsv)
	}

	if errWrite := csvWriter.Write(record); errWrite != nil {
		hwlog.RunLog.Errorf("write record to csv file failed, err: %v", errWrite)
		return
	}
	hwlog.RunLog.Infof("write record to csv file %s success", pingResultCsv)
}

func (m *PingManager) calcAppendModeAndOpenFlag() (bool, int) {
	appendMode := true
	curTimeMilliSec := time.Now().UnixMilli()
	if m.lastSaveTime == 0 || m.lastSaveTime+savePeriodMillSec <= curTimeMilliSec {
		m.lastSaveTime = curTimeMilliSec
		appendMode = false
	}
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if appendMode {
		openFlag = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	}
	hwlog.RunLog.Infof("append mode: %v, open flag: %v, lastSaveTime: %d, Now: %d", appendMode, openFlag,
		m.lastSaveTime, curTimeMilliSec)
	return appendMode, openFlag
}

func (m *PingManager) prepareResultFilePaths(appendMode bool) (csvFile, csvBackFile string, err error) {
	rasNetRootPath, err := slownet.GetRasNetRootPath()
	if err != nil {
		hwlog.RunLog.Errorf("get ras net fault root path failed, err: %v", err)
		return "", "", fmt.Errorf("get ras net fault root path failed")
	}
	csvFileName := fmt.Sprintf("ping_result_%d_%d.csv", m.superPodId, m.serverIndex)
	csvFileBackName := fmt.Sprintf("ping_result_%d_%d.csv-bak", m.superPodId, m.serverIndex)
	pingResultCsv := filepath.Join(rasNetRootPath, rasNetSubPath, roceSubPath, csvFileName)
	pingResultCsvBack := filepath.Join(rasNetRootPath, rasNetSubPath, roceSubPath, csvFileBackName)
	if _, err = utils.CheckPath(pingResultCsvBack); err != nil {
		hwlog.RunLog.Errorf("file path %s is invalid, err: %v", pingResultCsvBack, err)
		return "", "", fmt.Errorf("file path is invalid")
	}
	if utils.IsLexist(pingResultCsvBack) && !appendMode {
		if err = os.Remove(pingResultCsvBack); err != nil {
			hwlog.RunLog.Errorf("remove file %s failed, err: %v", pingResultCsvBack, err)
			return "", "", fmt.Errorf("remove file failed")
		}
	}
	if _, err = utils.CheckPath(pingResultCsv); err != nil {
		hwlog.RunLog.Errorf("file path %s is invalid, err: %v", pingResultCsv, err)
		return "", "", fmt.Errorf("file path invalid")
	}
	if utils.IsLexist(pingResultCsv) && !appendMode {
		if err = os.Rename(pingResultCsv, pingResultCsvBack); err != nil {
			hwlog.RunLog.Errorf("backup file %s failed, err := %v", pingResultCsv, err)
			return "", "", fmt.Errorf("backup file failed")
		}
	}
	return pingResultCsv, pingResultCsvBack, nil
}
