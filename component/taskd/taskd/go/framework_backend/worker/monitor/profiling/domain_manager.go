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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import "C"
import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

const maxRegisterTime = 10

const (
	maxCmdChanLen      = 10
	maxWaitNetInitTime = 180 * time.Second
	printErrDuration   = 30 * time.Second
)

// MsSubscribed whether ms light profiling is subscribed
var MsSubscribed atomic.Bool

// NetTool from worker
var NetTool *net.NetInstance

// MgrProfilingCmd from worker
var MgrProfilingCmd atomic.Bool

// GlobalRank of this work
var GlobalRank int

// NodeRank of this work
var NodeRank int

// CmdChan for SwitchProfiling
var CmdChan chan constant.ProfilingDomainCmd

func init() {
	CmdChan = make(chan constant.ProfilingDomainCmd, maxCmdChanLen)
}

// ManageDomainEnableStatus dead loop for manage domain status
func ManageDomainEnableStatus(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			hwlog.RunLog.Errorf("manager of changing domain manager has paniced, err: %v", r)
			fmt.Printf("[ERROR] %s manager of changing domain manager has paniced, err: %v\n", time.Now(), r)
		}
	}()
	hwlog.RunLog.Infof("start to watch for domain config changes")
	lastStatus := constant.ProfilingDomainCmd{
		DefaultDomainAble: false,
		CommDomainAble:    false,
	}
	firstRun := true
	go loopWatchProfilingFile()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("domain config received exit signal")
			return
		case newStatus := <-CmdChan:
			if lastStatus == newStatus && !firstRun {
				hwlog.RunLog.Debug("status not changed will not call mspti")
				continue
			}
			hwlog.RunLog.Infof("recv profiling cmd %v", newStatus)
			firstRun = false
			changeProfileSwitchStatus(newStatus)
			lastStatus = newStatus
		}
	}
}

func loopWatchProfilingFile() {
	circleTicker := time.NewTicker(constant.DomainCheckInterval)
	lastPrintErr := time.Now()
	for {
		select {
		case <-circleTicker.C:
			if MgrProfilingCmd.Load() {
				hwlog.RunLog.Info("MgrProfilingCmd load, return")
				return
			}
			lastPrintErr = getCmd(lastPrintErr)
		}
	}
}

func getCmd(lastPrintErr time.Time) time.Time {
	profilingSwitches, err := utils.GetProfilingSwitch(constant.ProfilingSwitchFilePath)
	if err != nil {
		if time.Since(lastPrintErr) > printErrDuration {
			hwlog.RunLog.Errorf("GetProfilingSwitch err: %v", err)
			lastPrintErr = time.Now()
		}
	} else {
		profilingDomainCmd := utils.PfSwitchToPfDomainSwitch(profilingSwitches)
		CmdChan <- profilingDomainCmd
	}
	return lastPrintErr
}

func changeProfileSwitchStatus(profilingDomainCmd constant.ProfilingDomainCmd) {
	result := constant.ProfilingResult{
		DefaultDomain: constant.ProfilingUnknownStatus,
		CommDomain:    constant.ProfilingUnknownStatus,
	}
	defer func() {
		hwlog.RunLog.Infof("exec cmd %v result %v", profilingDomainCmd, result)
		if MgrProfilingCmd.Load() {
			notifyMgrSwitchChange(result)
		}
	}()
	if !profilingDomainCmd.DefaultDomainAble {
		result.DefaultDomain = constant.ProfilingOffStatus
		if err := DisableMsptiActivity(); err != nil {
			hwlog.RunLog.Errorf("failed to disable MsptiActivity: %v", err)
			result.DefaultDomain = constant.ProfilingExpStatus
		}
		if MsSubscribed.Load() == true {
			if err := MsptiUnsubscribe(); err != nil {
				hwlog.RunLog.Error(err.Error())
				result.DefaultDomain = constant.ProfilingExpStatus
			} else {
				MsSubscribed.Store(false)
			}
		}
	} else {
		if MsSubscribed.Load() == false {
			if err := MsptiSubscribe(); err != nil {
				hwlog.RunLog.Error(err.Error())
				result.DefaultDomain = constant.ProfilingExpStatus
				return
			}
			MsSubscribed.Store(true)
		}
		result.DefaultDomain = constant.ProfilingOnStatus
		if err := EnableMsptiMarkerActivity(); err != nil {
			result.DefaultDomain = constant.ProfilingExpStatus
			hwlog.RunLog.Errorf("failed to change default marker domain status, err: %v", err)
		}
	}
	if !profilingDomainCmd.CommDomainAble {
		result.CommDomain = constant.ProfilingOffStatus
	} else {
		result.CommDomain = constant.ProfilingOnStatus
	}
	if err := EnableMarkerDomain(constant.CommunicationDomainName, profilingDomainCmd.CommDomainAble); err != nil {
		result.CommDomain = constant.ProfilingExpStatus
		hwlog.RunLog.Errorf("failed to change communication marker domain status, err: %v", err)
	}
}

func notifyMgrSwitchChange(result constant.ProfilingResult) {
	if NetTool == nil {
		hwlog.RunLog.Errorf("NetTool for worker is nil?")
		return
	}
	msg := storage.MsgBody{
		MsgType: constant.STATUS,
		Code:    utils.ProfilingResultToBizCode(result),
	}
	_, err := NetTool.SyncSendMessage(uuid.New().String(), "default", utils.ObjToString(msg), &common.Position{
		Role:       common.MgrRole,
		ServerRank: "0",
	})

	if err != nil {
		hwlog.RunLog.Errorf("send result to mgr err: %v", err)
		return
	}
	hwlog.RunLog.Infof("notify mgr result %v succeeded", result)
}

func ProcessMsg(globalRank int, msg *common.Message) {
	hwlog.RunLog.Infof("worker %d recv msg %v", globalRank, msg)
	profilingSwitch, err := getProfilingSwitch(msg)
	if err != nil {
		hwlog.RunLog.Errorf("getSwitchProfiling err: %v", err)
		return
	}
	CmdChan <- profilingSwitch
}

func getProfilingSwitch(msg *common.Message) (constant.ProfilingDomainCmd, error) {
	body, err := utils.StringToObj[storage.MsgBody](msg.Body)
	if err != nil {
		err = fmt.Errorf("get msgBody err: %v, msgBody is %v", err, body)
		return constant.ProfilingDomainCmd{}, err
	}
	return utils.BizCodeToProfilingCmd(body.Code)
}
