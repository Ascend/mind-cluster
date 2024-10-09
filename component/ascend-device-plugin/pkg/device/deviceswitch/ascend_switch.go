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

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/common-utils/utils"
	devmanagercommon "huawei.com/npu-exporter/v6/devmanager/common"

	"Ascend-device-plugin/pkg/common"
)

/*
   #cgo LDFLAGS: -ldl
   #cgo CFLAGS: -I/usr/local/Ascend/driver

    #include <stddef.h>
    #include <dlfcn.h>
    #include <stdlib.h>
    #include <stdio.h>

    #include "library.h"

    void *dcmiHandle;
    #define SO_NOT_FOUND  -99999
    #define FUNCTION_NOT_FOUND  -99998
    #define SUCCESS  0
    #define ERROR_UNKNOWN  -99997
    // dcmi
    int (*lq_dcmi_init_func)();
    static int dcmi_init_lq(){
		return lq_dcmi_init_func();
	}

	int (*lq_dcmi_get_fault_info_func)(unsigned int list_len, unsigned int *event_list_len, struct LqDcmiEvent *event_list);
	static int lq_dcmi_get_fault_info(unsigned int list_len, unsigned int *event_list_len, struct LqDcmiEvent *event_list){
		return lq_dcmi_get_fault_info_func(list_len,event_list_len,event_list);
	}

	void goFaultEventHandler(struct LqDcmiEvent *fault_event);
	static void event_handler(struct LqDcmiEvent *fault_event){
		goFaultEventHandler(fault_event);
	}

	int(*lq_dcmi_subscribe_fault_event_func)(struct lq_dcmi_event_filter filter,lq_dcmi_fault_event_callback handler);
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
	subTypeBase = 1000
)

// SwitchFaultEvent is the struct for switch reported fault
type SwitchFaultEvent struct {
	EventType uint
	// SubType fault subtype used for id a fault
	SubType uint
	// PeerPortDevice used to tell what kind of device connected to
	PeerPortDevice uint
	PeerPortId     uint
	SwitchChipId   uint
	SwitchPortId   uint
	// Severity used to tell how serious is the fault
	Severity uint
	// Assertion tell what kind of fault, recover, happen or once
	Assertion       uint
	EventSerialNum  int
	NotifySerialNum int
	AlarmRaisedTime int64
	AdditionalParam string
	AdditionalInfo  string
}

// SwitchDevManager is the manager for switch
type SwitchDevManager struct {
}

var (
	switchInitOnce sync.Once
	// fault code with subtype like 8,has 3 different kind, with different connect device:NPU CPU Switch,
	// this kind of fault code will be faultcode * 1000 + PeerPortDevice
	duplicateSubEventMap = map[int]bool{8: true, 10: true, 11: true, 21: true, 22: true, 23: true, 25: true}
)

// UpdateSwitchFaultLevel update the map recording fault code and it's level, as long as deviceinfo changed
func UpdateSwitchFaultLevel() {
	common.SwitchFaultLevelMapLock.Lock()
	defer common.SwitchFaultLevelMapLock.Unlock()
	common.SwitchFaultLevelMap = make(map[int64]int, common.GeneralMapSize)
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
	// faultEventHandler callback function for subscribe mod, witch will receive fault code when fault happens
	faultEvent := convertFaultEvent(event)
	hwlog.RunLog.Warnf("switch subscribe got fault:%#v, hex:%v", faultEvent,
		fmt.Sprintf("%08x", faultEvent.SubType))
	// for recovered fault, delete them from current fault codes
	if int8(faultEvent.Assertion) == devmanagercommon.FaultRecover {
		newFaultCodes := make([]int64, 0)
		for _, code := range common.GetSwitchFaultCode() {
			if code != int64(faultEvent.SubType) {
				newFaultCodes = append(newFaultCodes, code)
			}
		}
		common.SetSwitchFaultCode(newFaultCodes)
		return
	}
	currentFault := common.GetSwitchFaultCode()
	common.SetSwitchFaultCode(append(currentFault, int64(faultEvent.SubType)))
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
func GetSwitchFaults() ([]int64, error) {
	var errCount C.uint
	var errInfoArray [maxFaultNum]C.struct_LqDcmiEvent
	if retCode := C.lq_dcmi_get_fault_info(C.uint(maxFaultNum), &errCount,
		&errInfoArray[0]); int32(retCode) != devmanagercommon.Success {
		return []int64{}, fmt.Errorf("failed to get switch device errorcodes, errCode:%v", retCode)
	}
	if int32(errCount) < 0 || int32(errCount) > maxFaultNum {
		return []int64{}, fmt.Errorf("failed to get switch device errcodes, cause errcodes nums %d is illegal", errCount)
	}

	retErrores := make([]int64, 0)
	for i := 0; i < len(errInfoArray); i++ {
		faultEvent := convertFaultEvent(&errInfoArray[i])
		if faultEvent.SubType == 0 {
			continue
		}
		if int8(faultEvent.Assertion) == devmanagercommon.FaultRecover {
			continue
		}
		retErrores = append(retErrores, int64(faultEvent.SubType))
	}
	hwlog.RunLog.Warnf("get fault:%#v", retErrores)
	return retErrores, nil
}

// convertFaultEvent convert event getting from driver to go struct
func convertFaultEvent(event *C.struct_LqDcmiEvent) SwitchFaultEvent {
	fault := SwitchFaultEvent{
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
	fault.SubType = getEventType(fault)
	return fault
}

func getEventType(event SwitchFaultEvent) uint {
	if duplicate, ok := duplicateSubEventMap[int(event.SubType)]; ok && duplicate {
		return event.SubType*subTypeBase + event.PeerPortDevice
	}
	return event.SubType
}
