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

// Package dtfsmonitor for monitor fault by dtfs on the server
package dtfsmonitor

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
	dtfsStatus common.DtfsStatus
	// dtfs file match: DTFS_PROCESS_ERROR: XX
	errRegex       = regexp.MustCompile(`^(\w+):\s*(-?\d+)$`)
	instRegex      = regexp.MustCompile(`^\[instidx=(\d+)\]$`)
	lastUploadTime int64
)

// DtfsEventMonitor monitor fault on server by dtfs
type DtfsEventMonitor struct {
	stopChan chan struct{}
	ctx      context.Context
}

// NewDtfsEventMonitor create dtfs monitor
func NewDtfsEventMonitor(ctx context.Context) *DtfsEventMonitor {
	return &DtfsEventMonitor{
		stopChan: make(chan struct{}, 1),
		ctx:      ctx,
	}
}

// Init dtfs tool
func (i *DtfsEventMonitor) Init() error {
	return nil
}

// Stop terminate working loop
func (i *DtfsEventMonitor) Stop() {
	hwlog.RunLog.Info("stop dtfs status check")
	i.stopChan <- struct{}{}
}

// Name get monitor name
func (i *DtfsEventMonitor) Name() string {
	return common.PluginMonitorDtfs
}

// Monitoring start monitor
func (i *DtfsEventMonitor) Monitoring() {
	for {
		select {
		case _, ok := <-i.stopChan:
			if !ok {
				hwlog.RunLog.Error("stop channel is closed")
				return
			}
			hwlog.RunLog.Info("receive stop signal, dtfs monitor shut down...")
			return
		default:
			time.Sleep(common.CheckPeriod)
			newDtfsStatus, err := getStatusFromFile()
			if err != nil {
				hwlog.RunLog.ErrorfWithLimit(common.DtfsLogDomain,
					common.DtfsLogDomainId, "get dtfs status failed, err is %v", err)
				continue
			}
			hwlog.ResetErrCnt(common.DtfsLogDomain, common.DtfsLogDomainId)
			if isSame(newDtfsStatus) {
				continue
			}
			lastUploadTime = time.Now().UnixMilli()
			dtfsStatus = newDtfsStatus
			common.TriggerUpdate(common.DtfsProcess)
		}
	}
}

func isSame(newDtfsStatus common.DtfsStatus) bool {
	if lastUploadTime == 0 {
		return false
	}
	if dtfsStatus.ProcessError == newDtfsStatus.ProcessError && dtfsStatus.LinkError == newDtfsStatus.LinkError {
		return true
	}
	return false
}

func getStatusFromFile() (common.DtfsStatus, error) {
	absPath, err := utils.CheckOwnerAndPermission(common.DtfsFilePath, common.ExcludePermissions, common.RootUID)
	if err != nil {
		return common.DtfsStatus{}, fmt.Errorf("the filePath is invalid: %v", err)
	}
	f, err := os.Open(absPath)
	if err != nil {
		return common.DtfsStatus{}, fmt.Errorf("open file failed: %v", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	s := bufio.NewScanner(f)
	count := 0
	var newDtfsStatus common.DtfsStatus
	for s.Scan() {
		if count > common.MaxInstNumber {
			break
		}
		count++
		_, dtfsStatus, err := readInstStatus(s)
		if err != nil {
			return common.DtfsStatus{}, err
		}
		if dtfsStatus.ProcessError {
			newDtfsStatus.ProcessError = true
		}
		if dtfsStatus.LinkError {
			newDtfsStatus.LinkError = true
		}
	}
	if count == 0 {
		return common.DtfsStatus{}, errors.New("the file is empty")
	}
	return newDtfsStatus, nil
}

func readInstStatus(s *bufio.Scanner) (int, common.DtfsStatus, error) {
	var inst int
	var err error
	var newDtfsStatus common.DtfsStatus
	instStr := s.Text()
	if instMatch := instRegex.FindStringSubmatch(instStr); len(instMatch) > common.DtfsInstResultIndex {
		inst, err = strconv.Atoi(instMatch[common.DtfsInstResultIndex])
		if err != nil {
			return 0, common.DtfsStatus{}, err
		}
	} else {
		return 0, common.DtfsStatus{}, errors.New("get inst failed")
	}
	if !s.Scan() {
		return 0, common.DtfsStatus{}, errors.New("get status failed")
	}
	processErrorStr := s.Text()
	if status, err := getStatusByText(processErrorStr, common.DtfsProcessErrorKey); err != nil {
		return 0, common.DtfsStatus{}, err
	} else {
		newDtfsStatus.ProcessError = status
	}
	if !s.Scan() {
		return 0, common.DtfsStatus{}, errors.New("get status failed")
	}
	linkErrorStr := s.Text()
	if status, err := getStatusByText(linkErrorStr, common.DtfsLinkErrorKey); err != nil {
		return 0, common.DtfsStatus{}, err
	} else {
		newDtfsStatus.LinkError = status
	}
	return inst, newDtfsStatus, nil
}

func getStatusByText(text string, key string) (bool, error) {
	errMatch := errRegex.FindStringSubmatch(text)
	if len(errMatch) <= common.DtfsErrorResultIndex {
		return false, errors.New("get status failed, not match regex")
	}
	fileKey := errMatch[common.DtfsErrorTypeIndex]
	value, err := strconv.Atoi(errMatch[common.DtfsErrorResultIndex])
	if err != nil {
		return false, err
	}
	if fileKey != key {
		return false, errors.New("get status failed, key is invalid")
	}
	if value == common.DtfsError {
		return true, nil
	} else if value == common.DtfsHealthy {
		return false, nil
	}
	return false, errors.New("get status failed")
}

// GetMonitorData get monitor data
func (i *DtfsEventMonitor) GetMonitorData() *common.FaultAndConfigInfo {
	return &common.FaultAndConfigInfo{
		DtfsStatus: dtfsStatus,
	}
}
