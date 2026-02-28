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

// Package dcmi
package dcmi

import "C"
import (
	"fmt"
	"math"

	"ascend-common/common-utils/hwlog"
)

const (
	// RetError return error when the function failed
	retError = -1
	// hiAIMaxCardNum is the max number of cards
	hiAIMaxCardNum = 64
	// hiAIMaxCardID max card id for Ascend chip
	hiAIMaxCardID   = math.MaxInt32
	hiAIMaxDeviceID = math.MaxInt32
	// hiAIMaxDeviceNum is the max number of devices in a card
	hiAIMaxDeviceNum = 4
	maxChipNameLen   = 32
	productTypeLen   = 64
	notSupportCode   = -8255

	coreNumLen = 32
	// vfg_id indicates the virtual group ID to which the specified virtual device belongs.
	// It is automatically assigned by default, with a default value of 0xFFFFFFFF, which converts to 4294967295 in decimal.
	vfgID            = 4294967295
	notSupportString = "[not support]"
)

// Calling the V1 version of the initialization function on the Ascend950 generation (V2) will fail.
var managerList = []WorkerInterface{
	&NpuV1Worker{DcMgr: &DcV1Manager{}},
	&NpuV2Worker{DcMgr: &DcV2Manager{}},
}

// ChipInfo chip info
type ChipInfo struct {
	Type    string `json:"chip_type"`
	Name    string `json:"chip_name"`
	Version string `json:"chip_version"`
}

// DcDriverV1Interface dcmi v1 interface
type DcDriverV1Interface interface {
	DcInitialize() error
	DcShutDown()
	DcGetCardList() (int32, []int32, error)
	DcGetDeviceNumInCard(cardID int32) (int32, error)
	DcGetDeviceLogicID(cardID, deviceID int32) (int32, error)
	DcGetProductType(cardID, deviceID int32) (string, error)
	DcCreateVDevice(cardID, deviceID int32, coreNum string) (int32, error)
	DcDestroyVDevice(cardID, deviceID int32, vDevID int32) error
	DcGetChipInfo(cardID, deviceID int32) (*ChipInfo, error)
	DcGetDeviceLogicidFromPhyid(phyID int32) (int32, error)
}

// DcDriverV1Interface dcmi v2 interface
type DcDriverV2Interface interface {
	DcInitialize() error
	DcShutDown()
	DcCreateVDevice(deviceID int32, coreNum string) (int32, error)
	DcDestroyVDevice(deviceID int32, vDevID int32) error
	DcGetChipInfo(deviceID int32) (*ChipInfo, error)
	DcGetDeviceList() (int32, []int32, error)
}

// NpuV1Worker Dcmi V1 worker
type NpuV1Worker struct {
	DcMgr DcDriverV1Interface
}

// NpuV2Worker Dcmi V2 worker
type NpuV2Worker struct {
	DcMgr DcDriverV2Interface
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

// isValidA950DeviceID valid device id at ascend 950
func isValidA950DeviceID(deviceID int32) bool {
	return deviceID >= 0 && deviceID < hiAIMaxDeviceID
}

// isValidChipInfo valid chip info is or not empty
func isValidChipInfo(chip *ChipInfo) bool {
	if chip == nil {
		return false
	}
	return chip.Name != "" || chip.Type != "" || chip.Version != ""
}

// GetMatchingNpuWorker Obtain the NPU worker that matches the generation.
func GetMatchingNpuWorker() (WorkerInterface, error) {
	for _, manager := range managerList {
		err := manager.Initialize()
		if err == nil {
			manager.ShutDown()
			hwlog.RunLog.Infof("worker type: %T", manager)
			return manager, nil
		}
	}
	return nil, fmt.Errorf("failed to find a valid manager")
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
func (w *NpuV1Worker) Initialize() error {
	return w.DcMgr.DcInitialize()
}

// ShutDown shutdown dcmi lib
func (w *NpuV1Worker) ShutDown() {
	w.DcMgr.DcShutDown()
}

// CreateVDevice create virtual device
func (w *NpuV1Worker) CreateVDevice(uniqueID int32, coreNum string) (VDeviceInfo, error) {
	deviceID, cardID, err := w.FindDevice(uniqueID)
	if err != nil {
		return VDeviceInfo{CardID: -1, DeviceID: -1, VdeviceID: -1}, err
	}
	vdevId, err := w.DcMgr.DcCreateVDevice(cardID, deviceID, coreNum)
	if err != nil {
		return VDeviceInfo{CardID: -1, DeviceID: -1, VdeviceID: math.MaxInt32}, err
	}
	if vdevId > math.MaxInt32 {
		return VDeviceInfo{CardID: cardID, DeviceID: deviceID, VdeviceID: math.MaxInt32}, fmt.Errorf("create virtual device failed, vdeviceId too large")
	}
	return VDeviceInfo{CardID: cardID, DeviceID: deviceID, VdeviceID: int32(vdevId)}, nil
}

// DestroyVDevice destroy virtual device
func (w *NpuV1Worker) DestroyVDevice(uniqueID int32, vDevID int32) error {
	if vDevID < 0 {
		return fmt.Errorf("param error on vDevID")
	}
	deviceID, cardID, err := w.FindDevice(uniqueID)
	if err != nil {
		return err
	}
	return w.DcMgr.DcDestroyVDevice(cardID, deviceID, vDevID)
}

// FindDevice find device by phyical id
func (w *NpuV1Worker) FindDevice(visibleDevice int32) (int32, int32, error) {
	targetLogicID, err := w.DcMgr.DcGetDeviceLogicidFromPhyid(visibleDevice)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert phy id to logic id, err: %v", err)
	}
	_, cardList, err := w.DcMgr.DcGetCardList()
	if err != nil {
		return 0, 0, fmt.Errorf("get card list err : %v", err)
	}
	targetDeviceID, targetCardID := int32(math.MaxInt32), int32(math.MaxInt32)
	for _, cardID := range cardList {
		deviceCount, err := w.DcMgr.DcGetDeviceNumInCard(cardID)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot get device num in card : %v", err)
		}
		for deviceID := int32(0); deviceID < deviceCount; deviceID++ {
			logicID, err := w.DcMgr.DcGetDeviceLogicID(cardID, deviceID)
			if err != nil {
				return 0, 0, fmt.Errorf("cannot get logic id : %v", err)
			}
			if logicID == targetLogicID {
				targetCardID, targetDeviceID = cardID, deviceID
			}
		}
	}
	return targetDeviceID, targetCardID, nil
}

// GetProductType get type of product
func (w *NpuV1Worker) GetProductType() (string, error) {
	invalidType := ""
	if err := w.Initialize(); err != nil {
		return invalidType, fmt.Errorf("cannot init dcmi : %v", err)
	}
	defer w.ShutDown()

	cardNum, cardList, err := w.DcMgr.DcGetCardList()
	if cardNum == 0 || err != nil {
		hwlog.RunLog.Errorf("failed to get card list, err: %#v", err)
		return invalidType, err
	}
	for _, cardID := range cardList {
		devNum, err := w.DcMgr.DcGetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Debugf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		if devNum == 0 {
			hwlog.RunLog.Debugf("not found device on card %d", cardID)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			productType, err := w.DcMgr.DcGetProductType(cardID, devID)
			if err != nil {
				hwlog.RunLog.Debugf("get product type by card %d deviceID %d failed, err: %#v", cardID, devID, err)
				continue
			}
			return productType, nil
		}
	}
	return invalidType, nil
}

// GetChipName get name of chip
func (w *NpuV1Worker) GetChipName() (string, error) {
	invalidName := ""

	if err := w.Initialize(); err != nil {
		return invalidName, fmt.Errorf("cannot init dcmi : %v", err)
	}
	defer w.ShutDown()

	cardNum, cardList, err := w.DcMgr.DcGetCardList()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get card list, err: %#v", err)
		return invalidName, err
	}
	if cardNum == 0 {
		return invalidName, fmt.Errorf("get chip info failed, no card found")
	}

	// get device in card, then get chip info by cardID and deviceID
	for _, cardID := range cardList {
		devNum, err := w.DcMgr.DcGetDeviceNumInCard(cardID)
		if err != nil || devNum == 0 {
			hwlog.RunLog.Warnf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			chipInfo, err := w.DcMgr.DcGetChipInfo(cardID, devID)
			if err != nil {
				hwlog.RunLog.Warnf("get chip info failed by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			if !isValidChipInfo(chipInfo) {
				hwlog.RunLog.Warnf("invalid chip info by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			return (*chipInfo).Name, nil
		}
	}

	return invalidName, fmt.Errorf("cannot get valid chip info")
}

// Initialize dcmi lib init via DcMgr
func (a *NpuV2Worker) Initialize() error {
	return a.DcMgr.DcInitialize()
}

// ShutDown shutdown dcmi lib via DcMgr
func (a *NpuV2Worker) ShutDown() {
	a.DcMgr.DcShutDown()
}

// CreateVDevice create virtual device
func (a *NpuV2Worker) CreateVDevice(uniqueID int32, coreNum string) (VDeviceInfo, error) {
	vdevId, err := a.DcMgr.DcCreateVDevice(uniqueID, coreNum)
	if err != nil {
		return VDeviceInfo{
			CardID:    -1,
			DeviceID:  -1,
			VdeviceID: -1,
		}, err
	}
	// can not reach here now
	return VDeviceInfo{VdeviceID: vdevId}, nil
}

// DestroyVDevice destroy virtual device
func (a *NpuV2Worker) DestroyVDevice(uniqueID int32, vDevID int32) error {
	return a.DcMgr.DcDestroyVDevice(uniqueID, vDevID)
}

// GetProductType get type of product
func (a *NpuV2Worker) GetProductType() (string, error) {
	hwlog.RunLog.Infof("dcmi v2 not support GetProductType")
	return notSupportString, nil
}

// GetChipName get name of chip
func (a *NpuV2Worker) GetChipName() (string, error) {
	invalidName := ""

	if err := a.Initialize(); err != nil {
		return invalidName, fmt.Errorf("cannot init dcmi : %v", err)
	}
	defer a.ShutDown()

	deviceNum, deviceList, err := a.DcMgr.DcGetDeviceList()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get device list, err: %#v", err)
		return invalidName, err
	}
	if deviceNum == 0 {
		return invalidName, fmt.Errorf("get chip info failed, no card found")
	}

	// get device in card, then get chip info by cardID and deviceID
	for _, devID := range deviceList {
		chipInfo, err := a.DcMgr.DcGetChipInfo(devID)
		if err != nil {
			hwlog.RunLog.Warnf("get chip info failed by deviceID(%d), error: %#v", devID, err)
			continue
		}
		if !isValidChipInfo(chipInfo) {
			hwlog.RunLog.Warnf("invalid chip info by deviceID(%d), error: %#v", devID, err)
			continue
		}
		return (*chipInfo).Name, nil
	}
	return invalidName, fmt.Errorf("cannot get valid chip info")
}
