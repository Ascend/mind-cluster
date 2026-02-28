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

package dcmi

// #cgo LDFLAGS: -ldl
// #include "dcmi_interface_api.h"
import "C"
import (
	"fmt"
	"unsafe"

	"ascend-docker-runtime/mindxcheckutils"
)

// DcV2Manager dcmi v2 manager
type DcV2Manager struct {
}

// DcInitialize Initialize dcmi lib init
func (d *DcV2Manager) DcInitialize() error {
	cDlPath := C.CString(string(make([]byte, int32(C.PATH_MAX))))
	defer C.free(unsafe.Pointer(cDlPath))
	if err := C.DcmiInitDl(cDlPath); err != C.SUCCESS {
		errInfo := fmt.Errorf("dcmi lib load failed, error code: %d", int32(err))
		return errInfo
	}
	dlPath := C.GoString(cDlPath)
	if _, err := mindxcheckutils.RealFileChecker(dlPath, true, false, mindxcheckutils.DefaultSize); err != nil {
		return err
	}
	if err := C.DcmiV2Init(); err != C.SUCCESS {
		errInfo := fmt.Errorf("dcmi init failed, error code: %d", int32(err))
		return errInfo
	}
	return nil
}

// DcShutDown ShutDown shutdown dcmi lib
func (d *DcV2Manager) DcShutDown() {
	if err := C.DcmiShutDown(); err != C.SUCCESS {
		println(fmt.Errorf("dcmi shut down failed, error code: %d", int32(err)))
	}
}

// DcCreateVDevice dcmi create virtual device
func (d *DcV2Manager) DcCreateVDevice(deviceID int32, coreNum string) (int32, error) {
	return -1, fmt.Errorf("create error, ascend 950 not support virtual device")
}

// DcDestroyVDevice dcmi destroy virtual device
func (d *DcV2Manager) DcDestroyVDevice(deviceID int32, vDevID int32) error {
	return fmt.Errorf("destroy error, ascend 950 not support virtual device")
}

// DcGetChipInfo get the chip info by deviceID
func (d *DcV2Manager) DcGetChipInfo(deviceID int32) (*ChipInfo, error) {
	if !isValidA950DeviceID(deviceID) {
		return nil, fmt.Errorf("deviceID(%d) is invalid", deviceID)
	}
	var chipInfo C.struct_DcmiChipInfo
	if rCode := C.DcmiV2GetDeviceChipInfo(C.int(deviceID), &chipInfo); int32(rCode) != 0 {
		return nil, fmt.Errorf("get device ChipInfo information failed, deviceID(%d),"+
			" error code: %d", deviceID, int32(rCode))
	}

	name := convertUCharToCharArr(chipInfo.chipName)
	cType := convertUCharToCharArr(chipInfo.chipType)
	ver := convertUCharToCharArr(chipInfo.chipVer)

	chip := &ChipInfo{
		Name:    string(name),
		Type:    string(cType),
		Version: string(ver),
	}
	if !isValidChipInfo(chip) {
		return nil, fmt.Errorf("get device ChipInfo information failed, chip info is empty,"+
			" deviceID(%d)", deviceID)
	}

	return chip, nil
}

// DcGetDeviceList get deviceList
func (d *DcV2Manager) DcGetDeviceList() (int32, []int32, error) {
	var ids [hiAIMaxCardNum]C.int
	var cNum C.int
	if err := C.DcmiV2GetDeviceList(&ids[0], &cNum, hiAIMaxCardNum); err != 0 {
		errInfo := fmt.Errorf("get device list failed, error code: %d", int32(err))
		return retError, nil, errInfo
	}
	// checking device's quantity
	if cNum <= 0 || cNum > hiAIMaxCardNum {
		errInfo := fmt.Errorf("get error device quantity: %d", int32(cNum))
		return retError, nil, errInfo
	}
	var deviceNum = int32(cNum)
	var deviceIDList []int32
	for i := int32(0); i < deviceNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			continue
		}
		deviceIDList = append(deviceIDList, cardID)
	}
	return deviceNum, deviceIDList, nil
}
