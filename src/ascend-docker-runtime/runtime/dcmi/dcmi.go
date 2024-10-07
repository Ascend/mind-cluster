/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package dcmi
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

const (
	// RetError return error when the function failed
	retError = -1
	// hiAIMaxCardNum is the max number of cards
	hiAIMaxCardNum = 64
	// hiAIMaxCardID max card id for Ascend chip
	hiAIMaxCardID = math.MaxInt32
	// hiAIMaxDeviceNum is the max number of devices in a card
	hiAIMaxDeviceNum = 4
	maxChipNameLen   = 32
	productTypeLen   = 64
	notSupportCode   = -8255

	coreNumLen = 32
	vfgID      = 4294967295 // vfg_id表示指定虚拟设备所属的虚拟分组ID，默认自动分配，默认值为0xFFFFFFFF，转换成10进制为4294967295。
)

// ChipInfo chip info
type ChipInfo struct {
	Type    string `json:"chip_type"`
	Name    string `json:"chip_name"`
	Version string `json:"chip_version"`
}

// NpuWorker Dcmi worker
type NpuWorker struct {
}

// isValidCardID valid card id
func isValidCardID(cardID int32) bool {
	// for cardID, please watch the maximum value of the driver
	return cardID >= 0 && cardID < hiAIMaxCardID
}

// isValidDeviceID valid device id
func isValidDeviceID(deviceID int32) bool {
	return deviceID >= 0 && deviceID < hiAIMaxDeviceNum
}

// isValidCardIDAndDeviceID check two params both needs meet the requirement
func isValidCardIDAndDeviceID(cardID, deviceID int32) bool {
	return isValidCardID(cardID) && isValidDeviceID(deviceID)
}

// isValidChipInfo valid chip info is or not empty
func isValidChipInfo(chip *ChipInfo) bool {
	return chip.Name != "" || chip.Type != "" || chip.Version != ""
}

func convertUCharToCharArr(cgoArr [maxChipNameLen]C.uchar) []byte {
	var charArr []byte
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, byte(v))
	}
	return charArr
}

// Initialize dcmi lib init
func (w *NpuWorker) Initialize() error {
	cDlPath := C.CString(string(make([]byte, int32(C.PATH_MAX))))
	defer C.free(unsafe.Pointer(cDlPath))
	if err := C.DcmiInitDl(cDlPath); err != C.SUCCESS {
		errInfo := fmt.Errorf("dcmi lib load failed, , error code: %d", int32(err))
		return errInfo
	}
	dlPath := C.GoString(cDlPath)
	if _, err := mindxcheckutils.RealFileChecker(dlPath, true, false, mindxcheckutils.DefaultSize); err != nil {
		return err
	}
	if err := C.DcmiInit(); err != C.SUCCESS {
		errInfo := fmt.Errorf("dcmi init failed, , error code: %d", int32(err))
		return errInfo
	}
	return nil
}

// ShutDown shutdown dcmi lib
func (w *NpuWorker) ShutDown() {
	if err := C.DcmiShutDown(); err != C.SUCCESS {
		println(fmt.Errorf("dcmi shut down failed, error code: %d", int32(err)))
	}
}

// GetCardList  list all cards on system
func GetCardList() (int32, []int32, error) {
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
func GetDeviceNumInCard(cardID int32) (int32, error) {
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
func GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
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

// CreateVDevice create virtual device
func (w *NpuWorker) CreateVDevice(cardID, deviceID int32, coreNum string) (int32, error) {
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
	err := C.DcmiCreateVdevice(C.int(cardID), C.int(deviceID), &deviceCreateStr, &createInfo)
	if err != 0 {
		errInfo := fmt.Errorf("create virtual device failed, error code: %d", int32(err))
		return math.MaxInt32, errInfo
	}
	if createInfo.vdevId > math.MaxInt32 {
		return math.MaxInt32, fmt.Errorf("create virtual device failed, vdeviceId too large")
	}
	return int32(createInfo.vdevId), nil
}

// DestroyVDevice destroy virtual device
func (w *NpuWorker) DestroyVDevice(cardID, deviceID int32, vDevID int32) error {
	if vDevID < 0 {
		return fmt.Errorf("param error on vDevID")
	}
	if err := C.DcmiSetDestroyVdevice(C.int(cardID), C.int(deviceID), C.uint(vDevID)); err != 0 {
		errInfo := fmt.Errorf("destroy virtual device failed, error code: %d", int32(err))
		return errInfo
	}
	return nil
}

// FindDevice find device by phyical id
func (w *NpuWorker) FindDevice(visibleDevice int32) (int32, int32, error) {
	var dcmiLogicID C.uint
	if err := C.DcmiGetDeviceLogicidFromPhyid(C.uint(visibleDevice), &dcmiLogicID); err != 0 {
		return 0, 0, fmt.Errorf("phy id can not be converted to logic id : %v", err)
	}
	if int32(dcmiLogicID) < 0 || int32(dcmiLogicID) >= hiAIMaxCardNum*hiAIMaxDeviceNum {
		return 0, 0, fmt.Errorf("logic id too large")
	}
	targetLogicID := int32(dcmiLogicID)
	_, cardList, err := GetCardList()
	if err != nil {
		return 0, 0, fmt.Errorf("get card list err : %v", err)
	}
	targetDeviceID, targetCardID := int32(math.MaxInt32), int32(math.MaxInt32)
	for _, cardID := range cardList {
		deviceCount, err := GetDeviceNumInCard(cardID)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot get device num in card : %v", err)
		}
		for deviceID := int32(0); deviceID < deviceCount; deviceID++ {
			logicID, err := GetDeviceLogicID(cardID, deviceID)
			if err != nil {
				return 0, 0, fmt.Errorf("cannot get logic id : %v", err)
			}
			if logicID == int32(targetLogicID) {
				targetCardID, targetDeviceID = cardID, deviceID
			}
		}
	}
	return targetDeviceID, targetCardID, nil
}

// GetProductType get type of product by dcmi interface
func (w *NpuWorker) GetProductType(cardID, deviceID int32) (string, error) {
	cProductType := C.CString(string(make([]byte, productTypeLen)))
	defer C.free(unsafe.Pointer(cProductType))
	if err := C.DcmiGetProductType(C.int(cardID), C.int(deviceID),
		(*C.char)(cProductType), productTypeLen); err != 0 {
		if err == notSupportCode {
			// device which does not support querying product, such as Ascend 910A/B
			return "not support", nil
		}
		return "", fmt.Errorf("get product type failed, errCode: %d", err)
	}
	return C.GoString(cProductType), nil
}

// GetChipInfo get the chip info by cardID and deviceID
func (w *NpuWorker) GetChipInfo(cardID, deviceID int32) (*ChipInfo, error) {
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
