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

// Package deviceswitch functions of getting switch faults code
package deviceswitch

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"

	"Ascend-device-plugin/pkg/common"
	devmanagercommon "ascend-common/devmanager/common"
)

/*
   #cgo LDFLAGS: -ldl
   #cgo CFLAGS: -I/usr/local/Ascend/driver

    #include <stddef.h>
    #include <dlfcn.h>
    #include <stdlib.h>
    #include <stdio.h>

    #include "library.h"

    static void *dcmiHandle;
    #define SO_NOT_FOUND  -99999
    #define FUNCTION_NOT_FOUND  -99998
    #define SUCCESS  0
    #define ERROR_UNKNOWN  -99997
    // dcmi
    static int (*lq_dcmi_init_func)();
    static int dcmi_init_lq(){
		return lq_dcmi_init_func();
	}

	static int (*lq_dcmi_get_fault_info_func)(unsigned int listLen, unsigned int *eventListLen, struct LqDcmiEvent *eventList);
	static int lq_dcmi_get_fault_info(unsigned int listLen, unsigned int *eventListLen, struct LqDcmiEvent *eventList){
		return lq_dcmi_get_fault_info_func(listLen,eventListLen,eventList);
	}

	void goFaultEventHandler(struct LqDcmiEvent *fault_event);
	static void event_handler(struct LqDcmiEvent *fault_event){
		goFaultEventHandler(fault_event);
	}

	static int (*lq_dcmi_subscribe_fault_event_func)(struct lq_dcmi_event_filter filter,LqDcmiFaultEventCallback handler);
	static int lq_dcmi_subscribe_fault_event(struct lq_dcmi_event_filter filter){
		return lq_dcmi_subscribe_fault_event_func(filter,event_handler);
	}

	 // load .so files and functions
	static int dcmiInit_lq(const char* dcmiLibPath){
		if (dcmiLibPath == NULL) {
			fprintf (stderr,"lib path is null\n");
			return SO_NOT_FOUND;
		}
		dcmiHandle = dlopen(dcmiLibPath,RTLD_LAZY | RTLD_GLOBAL);
		if (dcmiHandle == NULL){
			fprintf (stderr,"%s\n",dlerror());
			return SO_NOT_FOUND;
		}


		lq_dcmi_init_func = dlsym(dcmiHandle,"lq_dcmi_init");
		lq_dcmi_subscribe_fault_event_func = dlsym(dcmiHandle,"lq_dcmi_subscribe_fault_event");
		lq_dcmi_get_fault_info_func = dlsym(dcmiHandle,"lq_dcmi_get_fault_info");
		return SUCCESS;
	}

	static int lqDcmiShutDown(void){
		if (dcmiHandle == NULL) {
			return SUCCESS;
		}
		return (dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
	}
*/
import "C"

const (
	maxFaultNum = 128
	invalidNum  = 0xFFFFFFFF
)

// SwitchDevManager is the manager for switch
type SwitchDevManager struct {
}

// UpdateSwitchFaultLevel update the map recording fault code and it's level, as long as deviceinfo changed
func UpdateSwitchFaultLevel() {
	common.SwitchFaultLevelMapLock.Lock()
	defer common.SwitchFaultLevelMapLock.Unlock()
	common.SwitchFaultLevelMap = make(map[string]int, common.GeneralMapSize)
	for _, code := range common.NotHandleFaultCodes {
		common.SwitchFaultLevelMap[code] = common.NotHandleFaultLevel
	}
	for _, code := range common.RestartRequestCodes {
		common.SwitchFaultLevelMap[code] = common.RestartRequestFaultLevel
	}
	for _, code := range common.PreSeparateFaultCodes {
		common.SwitchFaultLevelMap[code] = common.PreSeparateFaultLevel
	}
	for _, code := range common.SeparateFaultCodes {
		common.SwitchFaultLevelMap[code] = common.SeparateFaultLevel
	}
}

// NewSwitchDevManager create a new SwitchDevManager
func NewSwitchDevManager() *SwitchDevManager {
	return &SwitchDevManager{}
}

// InitSwitchDev try to call init func of driver, before call any other function
func (sdm *SwitchDevManager) InitSwitchDev() error {
	// path is not determined yet
	dcmiLibName := "liblingqu-dcmi.so"
	dcmiLibPath, err := utils.GetDriverLibPath(dcmiLibName)
	if err != nil {
		return fmt.Errorf("failed to find switch library so, err: %s", err.Error())
	}
	cDcmiTemplateName := C.CString(dcmiLibPath)
	defer C.free(unsafe.Pointer(cDcmiTemplateName))
	if retCode := C.dcmiInit_lq(cDcmiTemplateName); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi lib load failed, error code: %d", int32(retCode))
	}
	if retCode := C.dcmi_init_lq(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi init call failed, error code: %d", int32(retCode))
	}
	hwlog.RunLog.Info("init switch library succeeded")

	return nil
}

// ShutDownSwitch need to be called before dp exit
func (sdm *SwitchDevManager) ShutDownSwitch() {
	if retCode := C.lqDcmiShutDown(); retCode != C.SUCCESS {
		hwlog.RunLog.Error("failed to shutdown switch library")
		return
	}
	hwlog.RunLog.Info("switch library has been shutdown")
}

//export goFaultEventHandler
func goFaultEventHandler(event *C.struct_LqDcmiEvent) {
	defer func() {
		common.TriggerUpdate("switch fault occur")
	}()
	// faultEventHandler callback function for subscribe mod, which will receive fault code when fault happens
	faultEvent := convertFaultEvent(event)
	hwlog.RunLog.Warnf("switch subscribe got fault:%s, faultCode:%v",
		fmt.Sprintf("AlarmId: 0x%08x, FaultID: %v, AssembledFaultCode: %v, PeerPortDevice: %v, PeerPortId: %v, "+
			"SwitchChipId: %v, SwitchPortId: %v, Assertion: %v, AlarmRaisedTime: %v",
			faultEvent.EventType, faultEvent.FaultID, faultEvent.AssembledFaultCode, faultEvent.PeerPortDevice,
			faultEvent.PeerPortId, faultEvent.SwitchChipId, faultEvent.SwitchPortId, faultEvent.Assertion,
			faultEvent.AlarmRaisedTime), faultEvent.AssembledFaultCode)
	// for recovered fault, delete them from current fault codes
	if int8(faultEvent.Assertion) == devmanagercommon.FaultRecover {
		newFaultCodes := make([]common.SwitchFaultEvent, 0)
		for _, errInfo := range common.GetSwitchFaultCode() {
			// only in faultEvent and recoverEvent major info is the same it will be thought as recover
			if !isFaultRecoveredEvent(errInfo, faultEvent) {
				newFaultCodes = append(newFaultCodes, errInfo)
			}
		}
		common.SetSwitchFaultCode(newFaultCodes)
		return
	}
	currentFault := common.GetSwitchFaultCode()
	common.SetSwitchFaultCode(append(currentFault, faultEvent))
}

// GetSwitchFaultCodeByInterval start a none stop loop to query and update switch fault code
func (sdm *SwitchDevManager) GetSwitchFaultCodeByInterval(ctx context.Context, interval time.Duration) {
	runtime.LockOSThread()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal channel closed")
			}
			hwlog.RunLog.Info("query switch fault by interval stopped")
			return
		default:
			hwlog.RunLog.Debug("will start to query all switch fault codes")
			errCodes, err := GetSwitchFaults()
			if err != nil {
				hwlog.RunLog.Error(err)
				time.Sleep(interval)
				continue
			}
			common.SetSwitchFaultCode(errCodes)
			time.Sleep(interval)
		}
	}
}

// SubscribeSwitchFaults will start to subscribe fault from switch,
// and the callback function is faultEventHandler(event *C.struct_fault_event)
func (sdm *SwitchDevManager) SubscribeSwitchFaults() error {
	var filter C.struct_lq_dcmi_event_filter
	if retCode := C.lq_dcmi_subscribe_fault_event(filter); int32(retCode) != 0 {
		hwlog.RunLog.Errorf("failed to subscribe switch fault, errCode: %v", retCode)
		return fmt.Errorf("failed to subscribe switch fault, errCode: %v", retCode)
	}
	hwlog.RunLog.Info("succeed to subscribe switch fault")
	return nil
}

// GetSwitchFaults will try to get all fault
func GetSwitchFaults() ([]common.SwitchFaultEvent, error) {
	var errCount C.uint
	var errInfoArray [maxFaultNum]C.struct_LqDcmiEvent
	if retCode := C.lq_dcmi_get_fault_info(C.uint(maxFaultNum), &errCount,
		&errInfoArray[0]); int32(retCode) != devmanagercommon.Success {
		return []common.SwitchFaultEvent{}, fmt.Errorf("failed to get switch device errorcodes, "+
			"errCode:%v", retCode)
	}
	if int32(errCount) < 0 || int32(errCount) > maxFaultNum {
		return []common.SwitchFaultEvent{}, fmt.Errorf("failed to get switch device errcodes, "+
			"cause errcodes nums %v is illegal", errCount)
	}

	errorCodes := make([]string, 0)
	retErrorInfo := make([]common.SwitchFaultEvent, 0)
	for i := 0; i < int(errCount); i++ {
		faultEvent := convertFaultEvent(&errInfoArray[i])
		if int8(faultEvent.Assertion) == devmanagercommon.FaultRecover {
			continue
		}
		errorCodes = append(errorCodes, faultEvent.AssembledFaultCode)
		retErrorInfo = append(retErrorInfo, faultEvent)
	}
	if len(errorCodes) > 0 {
		// DO NOT edit this log, if necessary DO inform fault dialog
		hwlog.RunLog.Warnf("switch of 910A3 get fault codes: %#v", errorCodes)
	}
	return retErrorInfo, nil
}

// convertFaultEvent convert event getting from driver to go struct
func convertFaultEvent(event *C.struct_LqDcmiEvent) common.SwitchFaultEvent {
	fault := common.SwitchFaultEvent{
		EventType:       uint(event.eventType),
		SubType:         uint(event.subType),
		PeerPortDevice:  uint(event.peerportDevice),
		PeerPortId:      uint(event.peerportId),
		SwitchChipId:    uint(event.switchChipid),
		SwitchPortId:    uint(event.switchPortid),
		Severity:        uint(event.severity),
		Assertion:       uint(event.assertion),
		EventSerialNum:  int(event.eventSerialNum),
		NotifySerialNum: int(event.notifySerialNum),
		AlarmRaisedTime: int64(event.alarmRaisedTime),
	}
	setExtraFaultInfo(&fault)
	hwlog.RunLog.Debugf("convert switch fault finish, AlarmId:%v, FaultID:%v,"+
		"AssembledFaultCode:%v, PeerPortDevice:%v, AlarmRaisedTime:%v",
		fault.EventType, fault.FaultID, fault.AssembledFaultCode,
		fault.PeerPortDevice, fault.AlarmRaisedTime)
	return fault
}

// setExtraFaultInfo to convert fault event struct to a standard fault code as [0x00f1ff09,155912,cpu,na]
func setExtraFaultInfo(event *common.SwitchFaultEvent) {
	// to get peer device type, if switch port id is 0xFFFFFFFF, then it is not port level fault
	PeerDeviceType, PeerDeviceName := int(event.PeerPortDevice), ""
	if isPortLevelFault(int(event.SwitchPortId)) {
		PeerDeviceName = getPeerDeviceName(PeerDeviceType)
	} else {
		PeerDeviceName = common.PeerDeviceNAPortName
	}
	alarmID, faultID := event.EventType, event.SubType
	// for linkdown fault event, the faultID will be na, received 0xFFFFFFFF from driver
	if faultID == invalidNum {
		event.FaultID = common.PeerDeviceNAPortName
		event.AssembledFaultCode = fmt.Sprintf("[0x%08x,na,%s,na]", alarmID, PeerDeviceName)
	} else {
		event.FaultID = strconv.Itoa(int(faultID))
		event.AssembledFaultCode = fmt.Sprintf("[0x%08x,%d,%s,na]", alarmID, faultID, PeerDeviceName)
	}
	if event.AlarmRaisedTime == int64(0) {
		event.AlarmRaisedTime = time.Now().UnixMilli()
	}
}

func getPeerDeviceName(PeerDeviceType int) string {
	switch PeerDeviceType {
	case common.PeerDeviceChipOrCpuPort:
		return common.PeerDeviceChipOrCpuPortName
	case common.PeerDeviceNpuPort:
		return common.PeerDeviceNpuPortName
	case common.PeerDeviceL2Port:
		return common.PeerDeviceL2PortName
	default:
		return common.PeerDeviceNAPortName
	}
}

// isPortLevelFault to judge if a fault is related to whole chip or any of its ports
// if it is whole chip peer deviceType will be "na" while peerdeivce==0
// else it is for its chip, peer deviceType will be "cpu" while peerdeivce==0,
// only when the switchPortId value is not 0xFFFFFFFF, it is port level fault
func isPortLevelFault(switchPortId int) bool {
	return switchPortId != invalidNum
}

// isFaultRecoveredEvent to judge if recoverEvent is the recover event for faultEvent
func isFaultRecoveredEvent(faultEvent, recoverEvent common.SwitchFaultEvent) bool {
	if int8(recoverEvent.Assertion) != devmanagercommon.FaultRecover || recoverEvent.Assertion == faultEvent.Assertion {
		return false
	}
	faultEventInfo := fmt.Sprintf("EventType:%v,FaultID:%v,AssembledFaultCode:%v,"+
		"PeerPortDevice:%v,PeerPortId:%v,SwitchChipId:%v,SwitchPortId:%v", faultEvent.EventType, faultEvent.SubType,
		faultEvent.AssembledFaultCode, faultEvent.PeerPortDevice, faultEvent.PeerPortId,
		faultEvent.SwitchChipId, faultEvent.SwitchPortId)
	recoveredEventInfo := fmt.Sprintf("EventType:%v,FaultID:%v,AssembledFaultCode:%v,"+
		"PeerPortDevice:%v,PeerPortId:%v,SwitchChipId:%v,SwitchPortId:%v", recoverEvent.EventType, recoverEvent.SubType,
		recoverEvent.AssembledFaultCode, recoverEvent.PeerPortDevice, recoverEvent.PeerPortId,
		recoverEvent.SwitchChipId, recoverEvent.SwitchPortId)
	return faultEventInfo == recoveredEventInfo
}
