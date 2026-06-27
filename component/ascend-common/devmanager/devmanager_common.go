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

package devmanager

import (
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type DeviceCommonSetInterface interface {
	DeviceInterface
	SetValidMainBoardInfo() error
	SetDcManger(dcMgr interface{}) error
	SetDevType(devType string)
	GetDcManager() DeviceInterface
	SetAllProductType() error
	GetDcmiApiVersion() string
	SetDcmiVersion()
}

const (
	DcmiApiV1 = "dcmi"
	// DcmiApiV2 for the dcmiv2_xxx api
	DcmiApiV2 = "dcmiv2"
)

const (
	// util index for getDeviceUtilizationRateV1Common
	aicoreUtilIndex = iota
	aivUtilIndex
	npuUtilIndex
	aicUtilIndex
)

var deviceCommonSetManagerList = []DeviceCommonSetInterface{
	&deviceCommonInitManager{
		DeviceManager: DeviceManager{
			DcMgr:          &dcmi.DcManager{},
			dcmiApiVersion: DcmiApiV1,
		},
	},
	&deviceCommonInitManagerV2{
		DeviceManagerV2: DeviceManagerV2{
			DcMgr:          &dcmi.DcV2Manager{},
			dcmiApiVersion: DcmiApiV2,
		},
	},
}

// DetectDcmiApiVersion for detect dcmi dynamic library interface api version, such as dcmi_xxx or dcmiv2_xxx,
// and return a common device set manager to set all common param within resetTimeout
func DetectDcmiApiVersion(resetTimeout int) (DeviceCommonSetInterface, error) {
	for start, retryCnt := 0, 1; start < resetTimeout; retryCnt, start = retryCnt+1, start+defaultRetryDelay {
		hwlog.RunLog.Infof("timeout is %ds, dcmi version detection at %d times: ", resetTimeout, retryCnt)
		for _, devCommonSetMgr := range deviceCommonSetManagerList {
			hwlog.RunLog.Infof("try dcmi api version: %v", devCommonSetMgr.GetDcmiApiVersion())
			if err := devCommonSetMgr.Init(); err == nil {
				if err := devCommonSetMgr.ShutDown(); err != nil {
					hwlog.RunLog.Warnf("dcmi shutdown failed, err: %v", err)
					// ignore error
				}
				hwlog.RunLog.Infof("dcmi api version is %v", devCommonSetMgr.GetDcmiApiVersion())
				return devCommonSetMgr, nil
			} else {
				hwlog.RunLog.Warnf("dcmi api version: %v, init err: %v", devCommonSetMgr.GetDcmiApiVersion(), err)
			}
		}
		time.Sleep(time.Second * time.Duration(defaultRetryDelay))
	}
	return nil, errors.New(fmt.Sprintf("after %ds, can not find an available dcmi version", resetTimeout))
}

func getDeviceInfoForInit(commonDevMgr DeviceInterface) (common.ChipInfo, common.BoardInfo, error) {
	var err error
	chipInfo, err := commonDevMgr.GetValidChipInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.ChipInfo{}, common.BoardInfo{}, err
	}
	boardInfo, err := commonDevMgr.GetValidBoardInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return chipInfo, common.BoardInfo{}, err
	}

	return chipInfo, boardInfo, nil
}

// AutoInit auto detect npu chip type and return the corresponding processing object
func AutoInit(dType string, resetTimeout int) (DeviceInterface, error) {
	var devMgr DeviceInterface
	devCommonSetMgr, err := DetectDcmiApiVersion(resetTimeout)
	if err != nil {
		return nil, fmt.Errorf("detect dcmi version failed, err: %s", err)
	}
	// reduce interface range
	devMgr = devCommonSetMgr.GetDcManager()
	devMgr.WaitDeviceOnline(resetTimeout)
	chipInfo, boardInfo, err := getDeviceInfoForInit(devCommonSetMgr)
	if err != nil {
		return nil, fmt.Errorf("auto init failed when get device info, err: %s", err)
	}
	devCommonSetMgr.SetDcmiVersion()
	hwlog.RunLog.Infof("the dcmi version is %s", devMgr.GetDcmiVersion())
	err = devCommonSetMgr.SetValidMainBoardInfo()
	if err != nil {
		// Non-blocking when the main board ID is not found
		hwlog.RunLog.Warn(err)
	}
	var devType = common.GetDevType(chipInfo.Name, boardInfo.BoardId)
	switch devType {
	case api.Ascend910A5:
		err = devCommonSetMgr.SetDcManger(&A950Manager{})
	case api.Ascend910A, api.Ascend910B, api.Ascend910A3:
		err = devCommonSetMgr.SetDcManger(&A910Manager{})
	case api.Ascend310P:
		err = devCommonSetMgr.SetDcManger(&A310PManager{})
	case api.Ascend310, api.Ascend310B:
		err = devCommonSetMgr.SetDcManger(&A310Manager{})
	default:
		return nil, fmt.Errorf("unsupported device type (%s)", devType)
	}
	if err != nil {
		return nil, fmt.Errorf("set dcManager failed, err: %s", err)
	}
	if devType == api.Ascend910A5 {
		hwlog.RunLog.Infof("chipName: %v, devType: npu", chipInfo.Name)
	} else {
		hwlog.RunLog.Infof("chipName: %v, devType: %v", chipInfo.Name, devType)
	}
	if dType != "" && devType != dType {
		return nil, fmt.Errorf("the value of dType(%s) is inconsistent with the actual chip type(%s)",
			dType, devType)
	}
	devCommonSetMgr.SetDevType(devType)
	if err := devCommonSetMgr.SetIsTrainingCard(); err != nil {
		hwlog.RunLog.Errorf("auto recognize training card failed, err: %s", err)
	}
	err = devCommonSetMgr.SetAllProductType()
	if err != nil {
		hwlog.RunLog.Debugf("auto init product types failed, err: %s", err)
	}
	return devMgr, nil
}

// unsupportedDeviceTypeCache manages unsupported device types for device managers
type unsupportedDeviceTypeCache struct {
	unsupported map[int32]bool
	mu          sync.Mutex
}

func (c *unsupportedDeviceTypeCache) isUnsupported(deviceTypeCode int32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.unsupported == nil {
		return false
	}
	return c.unsupported[deviceTypeCode]
}

func (c *unsupportedDeviceTypeCache) markAsUnsupported(deviceTypeCode int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.unsupported == nil {
		c.unsupported = make(map[int32]bool)
	}
	c.unsupported[deviceTypeCode] = true
}

// utilizationFuncCache manages utilization function cache for device managers
type utilizationFuncCache struct {
	fn func(int32) (common.DcmiMultiUtilizationInfo, error)
	mu sync.Mutex
}

func (c *utilizationFuncCache) get() func(int32) (common.DcmiMultiUtilizationInfo, error) {
	c.mu.Lock()
	fn := c.fn
	c.mu.Unlock()
	return fn
}

func (c *utilizationFuncCache) set(fn func(int32) (common.DcmiMultiUtilizationInfo, error)) {
	if fn != nil {
		c.mu.Lock()
		c.fn = fn
		c.mu.Unlock()
	}
}

// determineUtilizationFunc determines and returns the best utilization function
type utilizationCandidate struct {
	fn          func(int32) (common.DcmiMultiUtilizationInfo, error)
	dcmiApiName string
}

func determineUtilizationFunc(logicID int32, candidates []utilizationCandidate,
) (func(int32) (common.DcmiMultiUtilizationInfo, error), common.DcmiMultiUtilizationInfo, error) {
	for i, candidate := range candidates {
		res, err := candidate.fn(logicID)
		if err == nil {
			hwlog.RunLog.Infof("%s interface is available, will use this interface to get utilization rate", candidate.dcmiApiName)
			return candidate.fn, res, nil
		}
		if i < len(candidates)-1 {
			hwlog.RunLog.Warnf("utilization func %s failed, try next, err: %v", candidate.dcmiApiName, err)
		}
	}
	return nil, dcmi.BuildErrNpuMultiUtilizationInfo(), fmt.Errorf("all utilization functions failed, logicID(%d)", logicID)
}

// getDeviceUtilizationRateV1Common gets utilization by calling GetDeviceUtilizationRate 4 times
func getDeviceUtilizationRateV1Common(logicID int32,
	getRateFunc func(int32, common.DeviceType) (uint32, error)) (common.DcmiMultiUtilizationInfo, error) {

	deviceTypes := []common.DeviceType{
		common.AICore,
		common.VectorCore,
		common.Overall,
		common.AICube,
	}

	results := make([]uint32, len(deviceTypes))
	errs := make([]error, len(deviceTypes))
	allFailed := true

	for i, devType := range deviceTypes {
		results[i], errs[i] = getRateFunc(logicID, devType)
		if errs[i] == nil {
			allFailed = false
			hwlog.ResetErrCnt(devType.Name, logicID)
		} else {
			// Don't print error again if it's a cached unsupported device type
			if !strings.Contains(errs[i].Error(), "is not supported (cached)") {
				hwlog.RunLog.ErrorfWithLimit(devType.Name, logicID,
					"get %s utilization rate failed, logicID(%d), err: %v", devType.Name, logicID, errs[i])
			}
		}
	}

	if allFailed {
		return dcmi.BuildErrNpuMultiUtilizationInfo(), fmt.Errorf("all GetDeviceUtilizationRate calls failed, logicID(%d)", logicID)
	}

	return common.DcmiMultiUtilizationInfo{
		AicoreUtil: results[aicoreUtilIndex],
		AivUtil:    results[aivUtilIndex],
		NpuUtil:    results[npuUtilIndex],
		AicUtil:    results[aicUtilIndex],
	}, nil
}
