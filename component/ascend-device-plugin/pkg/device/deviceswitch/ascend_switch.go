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
	"sync"
	"time"
	"unsafe"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
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
)

// SwitchDevManager is the manager for switch
type SwitchDevManager struct {
}

var (
	switchInitOnce sync.Once
	// eventTypeToFaultIDMapper 5/449 5/448 need to calculate in code
	eventTypeToFaultIDMapper = map[uint]uint{13: 155907, 12: 155649, 11: 155904, 10: 132134, 8: 155910, 9: 155911,
		7: 155908, 6: 155909, 5: 155912, 4: 155913, 3: 155914}
	// faultIdToAlarmIdMapper  5/8  need alarmID
	faultIdToAlarmIdMapper = map[uint]string{
		155907: "0x00f103b0", 155649: "0x00f103b0", 155904: "0x00f103b0",
		132134: "0x00f1ff06", 155910: "0x00f1ff06", 155911: "0x00f1ff06",
		155908: "0x00f103b6", 155909: "0x00f103b6",
		155912: "0x00f1ff09", 155913: "0x00f1ff09", 155914: "0x00f1ff09",
		132332: "0x00f10509", 132333: "0x00f10509",
	}
)

// UpdateSwitchFaultLevel update the map recording fault code and it's level, as long as deviceinfo changed
func UpdateSwitchFaultLevel() {
	common.SwitchFaultLevelMapLock.Lock()
	defer common.SwitchFaultLevelMapLock.Unlock()
	common.SwitchFaultLevelMap = make(map[string]int, common.GeneralMapSize)
	for _, code := range common.NotHandleFaultCodes {
		common.SwitchFaultLevelMap[code] = common.NotHandleFaultLevel
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
		return fmt.Errorf("failed to find switch library so, err:%s", err.Error())
	}
	cDcmiTemplateName := C.CString(dcmiLibPath)
	defer C.free(unsafe.Pointer(cDcmiTemplateName))
	if retCode := C.dcmiInit_lq(cDcmiTemplateName); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi lib load failed, error code: %d", int32(retCode))
	}
	if retCode := C.dcmi_init_lq(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi init call failed, error code: %d", int32(retCode))
	}
	hwlog.RunLog.Info("init switch library succeed")

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
	// faultEventHandler callback function for subscribe mod, which will receive fault code when fault happens
	faultEvent := convertFaultEvent(event)
	hwlog.RunLog.Warnf("switch subscribe got fault:%#v, faultCode:%v", faultEvent, faultEvent.AssembledFaultCode)
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
	// DO NOT edit this log, if necessary DO inform fault dialog
	hwlog.RunLog.Warnf("switch of 910A3 get fault codes: %#v", errorCodes)
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
	if err := setExtraFaultInfo(&fault); err != nil {
		hwlog.RunLog.Error(err)
	}
	hwlog.RunLog.Debugf("convert switch fault finish, EventType:%v,SubType:%v,FaultID:%v,"+
		"AssembledFaultCode:%v,PeerPortDevice:%v,AlarmRaisedTime:%v",
		fault.EventType, fault.SubType, fault.FaultID, fault.AssembledFaultCode, fault.PeerPortDevice, fault.AlarmRaisedTime)
	return fault
}

// setExtraFaultInfo to convert fault event struct to a standard fault code as [0x00f1ff09,155912,cpu,na]
func setExtraFaultInfo(event *common.SwitchFaultEvent) error {
	faultID, ok, alarmID := uint(0), false, ""
	// while eventType is 5, it depends on subtype to tell its faultID
	if event.EventType == common.EventTypeOfSwitchPortFault {
		switch event.SubType {
		case common.SubTypeOfPortDown:
			faultID = uint(0)
			alarmID = "0x08520003"
		case common.SubTypeOfPortLaneReduceQuarter:
			faultID = uint(common.FaultIdOfPortLaneReduceQuarter)
		case common.SubTypeOfPortLaneReduceHalf:
			faultID = uint(common.FaultIdOfPortLaneReduceHalf)
		default:
			faultID = uint(common.FaultIdOfPortFailOnForwardingChip)
		}
	} else {
		faultID, ok = eventTypeToFaultIDMapper[event.EventType]
		if !ok {
			hwlog.RunLog.Warnf("failed to find faultID for switch fault event: %v", event)
		}
	}
	if alarmID == "" {
		alarmID, ok = faultIdToAlarmIdMapper[faultID]
		if !ok {
			hwlog.RunLog.Warnf("failed to find alarm id for switch fault event: %v", event)
		}
	}
	PeerDeviceType, PeerDeviceName := int(event.PeerPortDevice), ""
	if isPortLevelFault(int(event.EventType)) {
		switch PeerDeviceType {
		case common.PeerDeviceChipOrCpuPort:
			PeerDeviceName = "cpu"
		case common.PeerDeviceNpuPort:
			PeerDeviceName = "npu"
		case common.PeerDeviceL2Port:
			PeerDeviceName = "L2"
		default:
			PeerDeviceName = "na"
		}
	} else {
		PeerDeviceName = "na"
	}
	// currently the last part means ports, remain na for now
	if faultID == uint(0) {
		event.AssembledFaultCode = fmt.Sprintf("[%s,,%s,na]", alarmID, PeerDeviceName)
	} else {
		event.AssembledFaultCode = fmt.Sprintf("[%s,%d,%s,na]", alarmID, faultID, PeerDeviceName)
	}
	event.FaultID = faultID
	if event.AlarmRaisedTime == int64(0) {
		event.AlarmRaisedTime = time.Now().Unix()
	}
	return nil
}

// isPortLevelFault to judge if a fault is related to whole chip or any of its ports
// if it is whole chip peer deviceType will be "na" while peerdeivce==0
// else it is for its chip, peer deviceType will be "cpu" while peerdeivce==0
func isPortLevelFault(eventType int) bool {
	if eventType == common.PortFaultInvalidPkgEventType || eventType == common.PortFaultUnstableEventType ||
		eventType == common.PortFaultFailEventType || eventType == common.PortFaultTimeoutLpEventType ||
		eventType == common.PortFaultTimeoutRpEventType {
		return true
	}
	return false
}

// isFaultRecoveredEvent to judge if recoverEvent is the recover event for faultEvent
func isFaultRecoveredEvent(faultEvent, recoverEvent common.SwitchFaultEvent) bool {
	if int8(recoverEvent.Assertion) != devmanagercommon.FaultRecover || recoverEvent.Assertion == faultEvent.Assertion {
		return false
	}
	faultEventInfo := fmt.Sprintf("EventType:%v,SubType:%v,FaultID:%v,AssembledFaultCode:%v,"+
		"PeerPortDevice:%v,PeerPortId:%v,SwitchChipId:%v,SwitchPortId:%v", faultEvent.EventType, faultEvent.SubType,
		faultEvent.FaultID, faultEvent.AssembledFaultCode, faultEvent.PeerPortDevice, faultEvent.PeerPortId,
		faultEvent.SwitchChipId, faultEvent.SwitchPortId)
	recoveredEventInfo := fmt.Sprintf("EventType:%v,SubType:%v,FaultID:%v,AssembledFaultCode:%v,"+
		"PeerPortDevice:%v,PeerPortId:%v,SwitchChipId:%v,SwitchPortId:%v", recoverEvent.EventType, recoverEvent.SubType,
		recoverEvent.FaultID, recoverEvent.AssembledFaultCode, recoverEvent.PeerPortDevice, recoverEvent.PeerPortId,
		recoverEvent.SwitchChipId, recoverEvent.SwitchPortId)
	return faultEventInfo == recoveredEventInfo
}
