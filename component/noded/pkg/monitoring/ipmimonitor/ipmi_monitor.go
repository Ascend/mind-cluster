/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package ipmimonitor for monitor the fault by ipmi on the server
package ipmimonitor

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/u-root/u-root/pkg/ipmi"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/common/manager"
)

var currentAlarmReqPrefix = []byte{0x30, 0x94, 0xDB, 0x07, 0x00, 0x40, 0x00}
var currentAlarmReqSuffix = []byte{0x0E, 0xFF}

const maxIpmiLoop = 100

// IpmiEventMonitor monitor fault on server by ipmi
type IpmiEventMonitor struct {
	ipmiTool     *ipmi.IPMI
	faultManager manager.FaultManager
	stopChan     chan struct{}
}

// NewIpmiEventMonitor create ipmi monitor
func NewIpmiEventMonitor() *IpmiEventMonitor {
	return &IpmiEventMonitor{
		ipmiTool:     &ipmi.IPMI{},
		faultManager: manager.NewFaultManager(),
		stopChan:     make(chan struct{}, 1),
	}
}

// Monitoring monitor working loop
func (i *IpmiEventMonitor) Monitoring() {
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
			time.Sleep(time.Duration(common.ParamOption.MonitorPeriod) * time.Second)
			if err := i.UpdateFaultDevList(); err != nil {
				hwlog.RunLog.Errorf("ipmi monitor update fault device list failed, error: %v", err)
			}
		}
	}
}

// Init initialize ipmi tool and get fault device list
func (i *IpmiEventMonitor) Init() error {
	ipmiTool, err := ipmi.Open(0)
	if err != nil {
		hwlog.RunLog.Errorf("open ipmi device failed, error: %v", err)
		return err
	}
	i.ipmiTool = ipmiTool
	if err := i.UpdateFaultDevList(); err != nil {
		hwlog.RunLog.Errorf("ipmi monitor update fault device list failed, please check bmc version and server type, "+
			"error: %v", err)
		return err
	}
	return nil
}

// Stop terminate working loop
func (i *IpmiEventMonitor) Stop() {
	if err := i.ipmiTool.Close(); err != nil {
		hwlog.RunLog.Errorf("ipmi tool close device failed, err is %v", err)
	}
	i.stopChan <- struct{}{}
}

// Name get monitor name
func (i *IpmiEventMonitor) Name() string {
	return common.PluginMonitorIpmi
}

// GetMonitorData get monitor data
func (i *IpmiEventMonitor) GetMonitorData() *common.FaultAndConfigInfo {
	fcInfo := &common.FaultAndConfigInfo{
		FaultDevInfo: i.faultManager.GetFaultDevInfo(),
	}
	return fcInfo
}

// UpdateFaultDevList update fault device list
func (i *IpmiEventMonitor) UpdateFaultDevList() error {
	currentAlarmFaultEvents, err := i.GetCurrentAlarmFaultEvents()
	if err != nil {
		hwlog.RunLog.Errorf("get current alarm fault events failed, error: %v", err)
		return errors.New("get current alarm fault events failed")
	}
	printFaultEvents(currentAlarmFaultEvents)
	i.faultManager.SetFaultDevList(GetFaultDevList(currentAlarmFaultEvents))
	common.TriggerUpdate(common.IpmiProcess)
	return nil
}

// GetCurrentAlarmFaultEvents get current fault events by ipmi query current msg
func (i *IpmiEventMonitor) GetCurrentAlarmFaultEvents() ([]*common.FaultEvent, error) {
	var alarmEvents []*common.FaultEvent
	var eventIndex int64 = 0
	firstAlarmMsg, err := i.ipmiTool.RawCmd(GetCurrentAlarmReq(eventIndex))
	if err != nil || len(firstAlarmMsg) < common.EventFieldStartIndex {
		hwlog.RunLog.Errorf("get first alarm msg from ipmi failed, error: %v", err)
		return nil, errors.New("get first alarm msg from ipmi failed")
	}
	alarmEvents = append(alarmEvents, GetFaultEvents(firstAlarmMsg[common.EventFieldStartIndex:])...)
	totalNumEvents := GetTotalEventsNum(firstAlarmMsg[common.TotalEventsStartIndex:common.TotalEventsEndIndex])
	msgNumEvents := int64(firstAlarmMsg[common.MsgEventsIndex])
	eventIndex += msgNumEvents
	totalNumEvents -= msgNumEvents
	currentIpmiLoop := 0
	for totalNumEvents > 0 && currentIpmiLoop < maxIpmiLoop {
		currentIpmiLoop++
		if eventIndex < 0 || eventIndex > common.MaxSixTeenBitIntValue {
			break
		}
		nextAlarmMsg, err := i.ipmiTool.RawCmd(GetCurrentAlarmReq(eventIndex))
		if err != nil {
			hwlog.RunLog.Errorf("get another alarm msg from ipmi failed, error: %v", err)
			return nil, errors.New("get another alarm msg from ipmi failed")
		}
		msgNumEvents = int64(nextAlarmMsg[common.MsgEventsIndex])
		eventIndex += msgNumEvents
		totalNumEvents -= msgNumEvents
		alarmEvents = append(alarmEvents, GetFaultEvents(nextAlarmMsg[common.EventFieldStartIndex:])...)
	}
	return alarmEvents, nil
}

// GetCurrentAlarmReq get current alarm request
func GetCurrentAlarmReq(eventIndex int64) []byte {
	return append(append(currentAlarmReqPrefix, common.ConvertIntToTwoByteSlice(eventIndex)...),
		currentAlarmReqSuffix...)
}

// GetFaultDevList get fault device list by ipmi fault event
func GetFaultDevList(faultEvents []*common.FaultEvent) []*common.FaultDev {
	faultDevErrCodeMap := make(map[string]map[int64][]string, 0)
	for _, faultEvent := range faultEvents {
		if _, ok := faultDevErrCodeMap[faultEvent.DeviceType]; !ok {
			faultDevErrCodeMap[faultEvent.DeviceType] = make(map[int64][]string, 0)
		}
		faultDevMap := faultDevErrCodeMap[faultEvent.DeviceType]
		if _, ok := faultDevMap[faultEvent.DeviceId]; !ok {
			faultDevMap[faultEvent.DeviceId] = make([]string, 0)
		}
		faultDevMap[faultEvent.DeviceId] = append(faultDevMap[faultEvent.DeviceId], faultEvent.ErrorCode)
	}
	var faultDevList []*common.FaultDev
	for deviceType, deviceMap := range faultDevErrCodeMap {
		for deviceId, errorCodeList := range deviceMap {
			faultDevList = append(faultDevList, &common.FaultDev{
				DeviceType: deviceType,
				DeviceId:   deviceId,
				FaultCode:  common.RemoveDuplicateString(errorCodeList),
			})
		}
	}
	return faultDevList
}

// GetTotalEventsNum get total events number
func GetTotalEventsNum(totalEventNumByte []byte) int64 {
	if len(totalEventNumByte) != common.TotalEventsByteLength {
		hwlog.RunLog.Error("input wrong total event num byte!")
		return 0
	}
	return int64(totalEventNumByte[1])*common.HexByteBase + int64(totalEventNumByte[0])
}

// GetFaultEvents get all fault events in a msg
func GetFaultEvents(eventsByte []byte) []*common.FaultEvent {
	if len(eventsByte)%common.SingleEventBytes != 0 {
		hwlog.RunLog.Errorf("events byte length is illegal, length is %d", len(eventsByte))
		return nil
	}
	var faultEvents []*common.FaultEvent
	for i := 0; i < len(eventsByte)-1; i += common.SingleEventBytes {
		faultEvents = append(faultEvents, GetFaultEvent(eventsByte[i:i+common.SingleEventBytes]))
	}
	return faultEvents
}

// GetFaultEvent get a fault event
func GetFaultEvent(eventByte []byte) *common.FaultEvent {
	if len(eventByte) <= common.DeviceIdIndex {
		return nil
	}
	return &common.FaultEvent{
		ErrorCode: strings.ToUpper(hex.EncodeToString(
			common.RevertByteSlice(eventByte[common.ErrorCodeStartIndex:common.ErrorCodeEndIndex]))),
		Severity:   int64(eventByte[common.SeverityIndex]),
		DeviceType: common.GetDeviceType(eventByte[common.DeviceTypeIndex]),
		DeviceId:   int64(eventByte[common.DeviceIdIndex]),
	}
}

func printFaultEvents(faultEvents []*common.FaultEvent) {
	for _, faultEvent := range faultEvents {
		hwlog.RunLog.Infof("get fault event, [device type: %s]-[device id: %d]-[error code: %s]",
			faultEvent.DeviceType, faultEvent.DeviceId, faultEvent.ErrorCode)
	}
}
