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

// Package dpcmonitor for monitor the fault by dpc on the server
package dpcmonitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
)

var (
	dpcMap map[int]common.DpcStatus = nil
	// dpc file match: [instidx=N]
	instRegex = regexp.MustCompile(`^\[instidx=(\d+)\]$`)
	// dpc file match: DPC_INTERNAL_ERROR: XX
	errRegex = regexp.MustCompile(`^(\w+):\s*(-?\d+)$`)

	lastUploadTime int64
)

// DpcEventMonitor monitor fault on server by dpc
type DpcEventMonitor struct {
	stopChan chan struct{}
	ctx      context.Context
}

// NewDpcEventMonitor create dpc monitor
func NewDpcEventMonitor(ctx context.Context) *DpcEventMonitor {
	return &DpcEventMonitor{
		stopChan: make(chan struct{}, 1),
		ctx:      ctx,
	}
}

// Init dpc tool
func (i *DpcEventMonitor) Init() error {
	return nil
}

// Stop terminate working loop
func (i *DpcEventMonitor) Stop() {
	hwlog.RunLog.Info("stop dpc status check")
	i.stopChan <- struct{}{}
}

// Name get monitor name
func (i *DpcEventMonitor) Name() string {
	return common.PluginMonitorDpc
}

// Monitoring start monitor
func (i *DpcEventMonitor) Monitoring() {
	for {
		select {
		case _, ok := <-i.stopChan:
			if !ok {
				hwlog.RunLog.Error("stop channel is closed")
				return
			}
			hwlog.RunLog.Info("receive stop signal, ipmi monitor shut down...")
			return
		default:
			time.Sleep(common.CheckPeriod)
			newDpcMap, err := getStatusFromFile()
			if err != nil {
				hwlog.RunLog.ErrorfWithLimit(common.DpcLogDomain,
					common.DpcLogDomainId, "get dpc status failed, err is %v", err)
				continue
			}
			hwlog.ResetErrCnt(common.DpcLogDomain, common.DpcLogDomainId)
			newDpcMap = setNewDpcMapTime(newDpcMap)
			if isSame(newDpcMap) {
				continue
			}
			lastUploadTime = time.Now().UnixMilli()
			dpcMap = newDpcMap
			common.TriggerUpdate(common.DpcProcess)
		}
	}
}

func setNewDpcMapTime(newDpcMap map[int]common.DpcStatus) map[int]common.DpcStatus {
	dpcMapWithTime := make(map[int]common.DpcStatus, len(newDpcMap))
	for i, status := range newDpcMap {
		oldStatus, ok := dpcMap[i]
		if !ok {
			status.ProcessErrorTime = time.Now().UnixMilli()
			status.MemoryErrorTime = time.Now().UnixMilli()
			dpcMapWithTime[i] = status
			continue
		}
		if oldStatus.ProcessError == status.ProcessError {
			status.ProcessErrorTime = oldStatus.ProcessErrorTime
		} else {
			status.ProcessErrorTime = time.Now().UnixMilli()
		}
		if oldStatus.MemoryError == status.MemoryError {
			status.MemoryErrorTime = oldStatus.MemoryErrorTime
		} else {
			status.MemoryErrorTime = time.Now().UnixMilli()
		}
		dpcMapWithTime[i] = status
	}
	return dpcMapWithTime
}

func isSame(newDpcMap map[int]common.DpcStatus) bool {
	if lastUploadTime == 0 || time.Now().UnixMilli()-lastUploadTime > common.DpcMonitorTimeOut.Milliseconds() {
		return false
	}
	if len(newDpcMap) != len(dpcMap) {
		return false
	}
	for i, status := range dpcMap {
		newStatus, ok := newDpcMap[i]
		if !ok {
			return false
		}
		if status.ProcessError != newStatus.ProcessError || status.MemoryError != newStatus.MemoryError {
			return false
		}
	}
	return true
}

func getStatusFromFile() (map[int]common.DpcStatus, error) {
	absPath, err := utils.CheckOwnerAndPermission(common.DpcFilePath, common.ExcludePermissions, common.RootUID)
	if err != nil {
		return nil, fmt.Errorf("the filePath is invalid: %v", err)
	}
	f, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	s := bufio.NewScanner(f)
	count := 0
	newDpcMap := make(map[int]common.DpcStatus)
	for s.Scan() {
		if count > common.MaxInstNumber {
			break
		}
		count++
		inst, dpcStatus, err := readInstStatus(s)
		if err != nil {
			return nil, fmt.Errorf("file not valid: %v", err)
		}
		newDpcMap[inst] = dpcStatus
	}
	return newDpcMap, nil
}

func readInstStatus(s *bufio.Scanner) (int, common.DpcStatus, error) {
	var inst int
	var err error
	var dpcStatus common.DpcStatus
	text := s.Text()
	if instMatch := instRegex.FindStringSubmatch(text); len(instMatch) > common.DpcInstResultIndex {
		inst, err = strconv.Atoi(instMatch[common.DpcInstResultIndex])
		if err != nil {
			return 0, common.DpcStatus{}, err
		}
	} else {
		return 0, common.DpcStatus{}, errors.New("get inst failed")
	}
	if !s.Scan() {
		return 0, common.DpcStatus{}, errors.New("get status failed")
	}
	text2 := s.Text()
	if status, err := getStatusByText(text2, common.DpcInternalErrorKey); err != nil {
		return 0, common.DpcStatus{}, err
	} else {
		dpcStatus.MemoryError = status
	}
	if !s.Scan() {
		return 0, common.DpcStatus{}, errors.New("get status failed")
	}
	text3 := s.Text()
	if status, err := getStatusByText(text3, common.DpcProcessErrorKey); err != nil {
		return 0, common.DpcStatus{}, err
	} else {
		dpcStatus.ProcessError = status
	}
	return inst, dpcStatus, nil
}

func getStatusByText(text string, key string) (bool, error) {
	errMatch := errRegex.FindStringSubmatch(text)
	if len(errMatch) <= common.DpcErrorResultIndex {
		return false, errors.New("get status failed, not match regex")
	}
	fileKey := errMatch[common.DpcErrorTypeIndex]
	value, err := strconv.Atoi(errMatch[common.DpcErrorResultIndex])
	if err != nil {
		return false, err
	}
	if fileKey != key {
		return false, errors.New("get status failed, key is invalid")
	}
	switch key {
	case common.DpcInternalErrorKey:
		if value == common.DpcInternalError {
			return true, nil
		} else if value == common.DpcInternalHealthy {
			return false, nil
		} else {
			return false, errors.New("get DPC_INTERNAL_ERROR failed, value is invalid")
		}
	case common.DpcProcessErrorKey:
		if value == common.DpcProcessError {
			return true, nil
		} else if value == common.DpcProcessHealthy {
			return false, nil
		} else {
			return false, errors.New("get DPC_PROCESS_ERROR failed, value is invalid")
		}
	default:
		return false, errors.New("get status failed, key is invalid")
	}

}

// GetMonitorData get monitor data
func (i *DpcEventMonitor) GetMonitorData() *common.FaultAndConfigInfo {
	return &common.FaultAndConfigInfo{
		DpcStatusMap: dpcMap,
	}
}
