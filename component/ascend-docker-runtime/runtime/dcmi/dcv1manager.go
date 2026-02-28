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
	"math"
	"unsafe"

	"ascend-docker-runtime/mindxcheckutils"
)

// DcV1Manager dcmi v1 manager
type DcV1Manager struct {
}

// DcGetDeviceLogicidFromPhyid convert phy id to logic id
func (d *DcV1Manager) DcGetDeviceLogicidFromPhyid(phyID int32) (int32, error) {
	var dcmiLogicID C.uint
	if err := C.DcmiGetDeviceLogicidFromPhyid(C.uint(phyID), &dcmiLogicID); err != 0 {
		return 0, fmt.Errorf("phy id %v can not be converted to logic id : %v", phyID, err)
	}
	if int32(dcmiLogicID) < 0 || int32(dcmiLogicID) >= hiAIMaxCardNum*hiAIMaxDeviceNum {
		return 0, fmt.Errorf("logic id too large")
	}
	return int32(dcmiLogicID), nil
}

// Initialize dcmi lib init
func (d *DcV1Manager) DcInitialize() error {
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
	if err := C.DcmiInit(); err != C.SUCCESS {
		errInfo := fmt.Errorf("dcmi init failed, error code: %d", int32(err))
		return errInfo
	}
	return nil
}

// ShutDown shutdown dcmi lib
func (d *DcV1Manager) DcShutDown() {
	if err := C.DcmiShutDown(); err != C.SUCCESS {
		println(fmt.Errorf("dcmi shut down failed, error code: %d", int32(err)))
	}
}

// DcGetCardList  list all cards on system
func (d *DcV1Manager) DcGetCardList() (int32, []int32, error) {
	var ids [hiAIMaxCardNum]C.int
	var cNum C.int
	if err := C.DcmiGetCardNumList(&cNum, &ids[0], hiAIMaxCardNum); err != 0 {
		errInfo := fmt.Errorf("get card list failed, error code: %d", int32(err))
		return retError, nil, errInfo
	}
	// checking card's quantity
	if cNum <= 0 || cNum > hiAIMaxCardNum {
		errInfo := fmt.Errorf("get error card quantity: %d", int32(cNum))
		return retError, nil, errInfo
	}
	var cardNum = int32(cNum)
	var cardIDList []int32
	for i := int32(0); i < cardNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			continue
		}
		cardIDList = append(cardIDList, cardID)
	}
	return cardNum, cardIDList, nil
}

// GetDeviceNumInCard get device number in the npu card
func (d *DcV1Manager) DcGetDeviceNumInCard(cardID int32) (int32, error) {
	var deviceNum C.int
	if err := C.DcmiGetDeviceNumInCard(C.int(cardID), &deviceNum); err != 0 {
		errInfo := fmt.Errorf("get device count on the card failed, error code: %d", int32(err))
		return retError, errInfo
	}
	if deviceNum <= 0 || deviceNum > hiAIMaxDeviceNum {
		errInfo := fmt.Errorf("the number of chips obtained is invalid, the number is: %d", int32(deviceNum))
		return retError, errInfo
	}
	return int32(deviceNum), nil
}

// GetDeviceLogicID get device logicID
func (d *DcV1Manager) DcGetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	var logicID C.int
	if err := C.DcmiGetDeviceLogicId(&logicID, C.int(cardID), C.int(deviceID)); err != 0 {
		errInfo := fmt.Errorf("get logicID failed, error code: %d", int32(err))
		return retError, errInfo
	}

	// check whether phyID is too big
	if logicID < 0 || uint32(logicID) > uint32(math.MaxInt8) {
		errInfo := fmt.Errorf("the logicID value is invalid, logicID is: %d", logicID)
		return retError, errInfo
	}
	return int32(logicID), nil
}

// GetProductType get type of product
func (d *DcV1Manager) DcGetProductType(cardID, deviceID int32) (string, error) {
	cProductType := C.CString(string(make([]byte, productTypeLen)))
	defer C.free(unsafe.Pointer(cProductType))
	if err := C.DcmiGetProductType(C.int(cardID), C.int(deviceID),
		(*C.char)(cProductType), productTypeLen+1); err != 0 {
		if err == notSupportCode {
			// device which does not support querying product, such as Ascend 910A/B
			return "not support", nil
		}
		return "", fmt.Errorf("get product type failed, errCode: %d", err)
	}
	return C.GoString(cProductType), nil
}

// DcDestroyVDevice dcmi destroy virtual device
func (d *DcV1Manager) DcDestroyVDevice(cardID, deviceID int32, vDevID int32) error {
	if err := C.DcmiSetDestroyVdevice(C.int(cardID), C.int(deviceID), C.uint(vDevID)); err != 0 {
		errInfo := fmt.Errorf("destroy virtual device failed, error code: %d", int32(err))
		return errInfo
	}
	return nil
}

// DcCreateVDevice dcmi create virtual device
func (d *DcV1Manager) DcCreateVDevice(cardID, deviceID int32, coreNum string) (int32, error) {
	var createInfo C.struct_DcmiCreateVdevOut
	createInfo.vdevId = C.uint(math.MaxUint32)
	var deviceCreateStr C.struct_DcmiCreateVdevResStru
	deviceCreateStr = C.struct_DcmiCreateVdevResStru{
		vdevId: C.uint(vfgID),
		vfgId:  C.uint(vfgID),
	}
	deviceCreateStrArr := [coreNumLen]C.char{0}
	for i := 0; i < len(coreNum); i++ {
		if i >= coreNumLen {
			return math.MaxInt32, fmt.Errorf("wrong template")
		}
		deviceCreateStrArr[i] = C.char(coreNum[i])
	}
	deviceCreateStr.templateName = deviceCreateStrArr
	ret := C.DcmiCreateVdevice(C.int(cardID), C.int(deviceID), &deviceCreateStr, &createInfo)
	if ret != 0 {
		errInfo := fmt.Errorf("create virtual device failed, error code: %d", int32(ret))
		return math.MaxInt32, errInfo
	}
	return int32(createInfo.vdevId), nil
}

// DcGetChipInfo get the chip info by cardID and deviceID
func (d *DcV1Manager) DcGetChipInfo(cardID, deviceID int32) (*ChipInfo, error) {
	if !isValidCardIDAndDeviceID(cardID, deviceID) {
		return nil, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var chipInfo C.struct_DcmiChipInfo
	if rCode := C.DcmiGetDeviceChipInfo(C.int(cardID), C.int(deviceID), &chipInfo); int32(rCode) != 0 {
		return nil, fmt.Errorf("get device ChipInfo information failed, cardID(%d), deviceID(%d),"+
			" error code: %d", cardID, deviceID, int32(rCode))
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
			" cardID(%d), deviceID(%d)", cardID, deviceID)
	}

	return chip, nil
}
