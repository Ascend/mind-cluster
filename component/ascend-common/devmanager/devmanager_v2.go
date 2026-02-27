// Package devmanager for device info manager
package devmanager

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
	"ascend-common/devmanager/dcmiv2"
)

// DeviceManagerV2 common device manager by dcmiv2 api
type DeviceManagerV2 struct {
	// DcMgr for common dev manager
	DcMgr dcmiv2.DcDriverInterface
	// DevType the value is the same as the device type corresponding to the DcMgr variable.
	DevType string
	// ProductTypes product type in server
	ProductTypes []string
	// isTrainingCard whether the device is used for training
	isTrainingCard bool
	// dcmiVersion for dcmi driver product version
	dcmiVersion string
	// dcmiApiVersion dcmi interface api version, v1 or empty for dcmi_xxx, v2 for dcmiv2_xxx api
	dcmiApiVersion string
	// mainBoardId used to distinguish between A900A3SuperPod and A9000A3SuperPod
	mainBoardId uint32
}

const (
	errNotSupportedInDcmiV2 = "is not supported in dcmiv2"
	// DcmiApiV1 for the dcmi_xxx api
	DcmiApiV1 = "dcmi"
	// DcmiApiV2 for the dcmiv2_xxx api
	DcmiApiV2 = "dcmiv2"
)

var (
	// for dcmiv2
	devManagerV2     *DeviceManagerV2 = nil
	devManagerV2Once sync.Once
	dcmiApiVersion   = ""
	// isContainAtlas300IDuo for reset scene
	isContainAtlas300IDuo = false
)

func initDeviceManager(mgr *DeviceManagerV2) bool {
	if mgr == nil {
		hwlog.RunLog.Error("init device manager v2 failed, mgr is empty")
		return false
	}
	if err := mgr.Init(); err != nil {
		hwlog.RunLog.Errorf("init device manager v2 failed, err: %s", err)
		return false
	}
	hwlog.RunLog.Info("init device manager v2 success")
	dcmiVer, err := mgr.DcMgr.DcGetDcmiVersion()
	if err != nil {
		hwlog.RunLog.Warnf("deviceManager get dcmi version failed, err: %v", err)
		// will continue the whole init workflow
	}
	hwlog.RunLog.Infof("the dcmi version is %s", dcmiVer)
	mgr.dcmiVersion = dcmiVer
	return true
}

func retryGetDeviceList(mgr *DeviceManagerV2, resetTimeout int) bool {
	if mgr == nil {
		hwlog.RunLog.Error("retry get device list failed, mgr is empty")
		return false
	}
	var retryDelay = defaultRetryDelay
	hwlog.RunLog.Infof("get device list from dcmi reset timeout is %d", resetTimeout)
	for currentTime, retryCount := 0, 0; currentTime <= resetTimeout; currentTime += retryDelay {
		devNum, devList, err := mgr.GetDeviceList()
		if err == nil && int(devNum) == len(devList) {
			hwlog.RunLog.Infof("deviceManager get devList is %v, devList length equal to devNum: %v", devList, devNum)
			break
		}
		if diffTime := float64(resetTimeout - currentTime); diffTime > 0 {
			retryDelay = int(math.Min(float64(defaultRetryDelay), diffTime))
		}
		retryCount++
		hwlog.RunLog.Warnf("deviceManager get device list failed (attempt %d), devNum=%d, devList=%v, err: %v",
			retryCount, devNum, devList, err)
		if currentTime+retryDelay <= resetTimeout {
			if err = mgr.ShutDown(); err != nil {
				hwlog.RunLog.Errorf("deviceManager shut down failed, err: %v", err)
				return false
			}
			time.Sleep(time.Second * time.Duration(retryDelay))
			continue
		}
		if int(devNum) != len(devList) {
			hwlog.RunLog.Warnf("deviceManager get devList is %v, but devNum is %v, "+
				"please check whether the real number of npu matches the devList", devList, devNum)
		}
	}
	return true
}

func setupDeviceInfo(mgr *DeviceManagerV2, dType string) bool {
	if mgr == nil {
		hwlog.RunLog.Error("retry get device list failed, mgr is empty")
		return false
	}
	chipInfo, err := mgr.GetValidChipInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}
	boardInfo, err := mgr.GetValidBoardInfo()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}
	_, err = mgr.GetValidMainBoardInfo()
	if err != nil {
		hwlog.RunLog.Warn(err)
	}
	devType := common.GetDevType(chipInfo.Name, boardInfo.BoardId)
	switch devType {
	case api.Ascend910A5:
	default:
		hwlog.RunLog.Errorf("unsupport device type (%s)", devType)
		return false
	}
	hwlog.RunLog.Infof("chipName: %v, devType: %v", chipInfo.Name, devType)
	if dType != "" && devType != dType {
		hwlog.RunLog.Errorf("the value of dType(%s) is inconsistent with the actual chip type(%s)", dType, devType)
		return false
	}
	mgr.DevType = devType
	if err = mgr.SetIsTrainingCard(); err != nil {
		hwlog.RunLog.Errorf("auto recognize training card failed, err: %s", err)
	}
	pTypes, err := mgr.GetAllProductType()
	if err != nil {
		hwlog.RunLog.Debugf("auto init product types failed, err: %s", err)
		// ignore the error, which does not matter
	}
	mgr.ProductTypes = pTypes
	return true
}

// Init load symbol and initialize dcmi
func (d *DeviceManagerV2) Init() error {
	return d.DcMgr.DcInit()
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManagerV2) ShutDown() error {
	return d.DcMgr.DcShutDown()
}

// GetDcmiVersion  get dcmi version
func (d *DeviceManagerV2) GetDcmiVersion() string {
	return d.dcmiVersion
}

// GetDeviceCount get npu device count
func (d *DeviceManagerV2) GetDeviceCount() (int32, error) {
	return d.DcMgr.DcGetDeviceCount()
}

// GetBrotherCardID get brother card id
func (d *DeviceManagerV2) GetBrotherCardID(logicID int32) (int32, error) {
	err := fmt.Errorf("get brother card id by logicID(%d) %s", logicID, errNotSupportedInDcmiV2)
	hwlog.RunLog.Error(err)
	return common.RetError, err
}

// PreResetSoc pre reset soc
func (d *DeviceManagerV2) PreResetSoc(logicID int32) error {
	return d.DcMgr.DcPreResetSoc(logicID)
}

// GetOutBandChannelState get out band channel state
func (d *DeviceManagerV2) GetOutBandChannelState(logicID int32) error {
	return d.DcMgr.DcGetOutBandChannelState(logicID)
}

// SetDeviceResetOutBand set device reset out band
func (d *DeviceManagerV2) SetDeviceResetOutBand(logicID int32) error {
	return d.DcMgr.DcSetDeviceResetOutBand(logicID)
}

// GetHccsStatisticInfoInU64 get hccs statistic info in u64
func (d *DeviceManagerV2) GetHccsStatisticInfoInU64(logicID int32) (*common.HccsStatisticInfo, error) {
	return nil, fmt.Errorf("getHccsStatisticInfoInU64 by logicID(%d) %s", logicID, errNotSupportedInDcmiV2)
}

// GetCardList get npu card list, is not supported in dcmiv2
func (d *DeviceManagerV2) GetCardList() (int32, []int32, error) {
	return common.RetError, nil, fmt.Errorf("getCardList %s", errNotSupportedInDcmiV2)
}

// GetDeviceNumInCard get npu device number by card, is not supported in dcmiv2
func (d *DeviceManagerV2) GetDeviceNumInCard(cardID int32) (int32, error) {
	return -1, fmt.Errorf("getDeviceNumInCard by cardID(%d) %s", cardID, errNotSupportedInDcmiV2)
}

// GetProductTypeArray get npu device product types array
func (d *DeviceManagerV2) GetProductTypeArray() []string {
	return d.ProductTypes
}

// GetDeviceBootStatus get device boot status
func (d *DeviceManagerV2) GetDeviceBootStatus(logicID int32) (int, error) {
	return d.DcMgr.DcGetDeviceBootStatus(logicID)
}

// SetFaultEventCallFunc set fault event callback function
func (d *DeviceManagerV2) SetFaultEventCallFunc(businessFunc func(common.DevFaultInfo)) error {
	if businessFunc == nil {
		return errors.New("business func can't be nil")
	}
	d.DcMgr.DcSetFaultEventCallFunc(businessFunc)
	return nil
}

// IsTrainingCard check the device is training card
func (d *DeviceManagerV2) IsTrainingCard() bool {
	return d.isTrainingCard
}

// GetMainBoardId get main board id
func (d *DeviceManagerV2) GetMainBoardId() uint32 {
	return d.mainBoardId
}

// GetValidMainBoardInfo get valid main board info
func (d *DeviceManagerV2) GetValidMainBoardInfo() (uint32, error) {
	// get device list
	deviceNum, deviceList, err := d.DcMgr.DcGetDeviceList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return 0, fmt.Errorf(common.ErrMsgInitCardListFailed)
	}
	if deviceNum == 0 {
		return 0, fmt.Errorf(common.ErrMsgGetBoardInfoFailed)
	}
	for _, deviceID := range deviceList {
		mainBoardId, err := d.DcMgr.DcGetDeviceMainBoardInfo(deviceID)
		if err != nil {
			hwlog.RunLog.Debug(err)
			continue
		}
		if !common.IsValidMainBoardInfo(mainBoardId) {
			hwlog.RunLog.Warnf("invalid mainBoardId info by deviceID(%d), error: %v", deviceID, err)
			continue
		}
		d.mainBoardId = mainBoardId
		return mainBoardId, nil
	}
	return 0, errors.New("cannot get main board id")
}

// GetValidChipInfo get valid chip info
func (d *DeviceManagerV2) GetValidChipInfo() (common.ChipInfo, error) {
	devNum, devList, err := d.DcMgr.DcGetDeviceList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.ChipInfo{}, fmt.Errorf(common.ErrMsgInitDeviceListFailed)
	}

	if devNum == 0 {
		return common.ChipInfo{}, fmt.Errorf("get chip info failed, no device found")
	}
	for _, devId := range devList {
		chipInfo, err := d.DcMgr.DcGetChipInfo(devId)
		if err != nil {
			hwlog.RunLog.Debugf("get chip info failed by deviceId(%d), error: %v", devId, err)
			continue
		}
		if !common.IsValidChipInfo(chipInfo) {
			hwlog.RunLog.Debugf("invalid chip info by deviceID(%d), error: %v", devId, err)
			continue
		}
		return *chipInfo, nil
	}

	return common.ChipInfo{}, errors.New("cannot get valid chip info")
}

// GetValidBoardInfo get valid board info
func (d *DeviceManagerV2) GetValidBoardInfo() (common.BoardInfo, error) {
	// get device list
	devNum, devList, err := d.DcMgr.DcGetDeviceList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.BoardInfo{}, fmt.Errorf(common.ErrMsgInitCardListFailed)
	}
	if devNum == 0 {
		return common.BoardInfo{}, fmt.Errorf(common.ErrMsgGetBoardInfoFailed)
	}
	for _, devId := range devList {
		boardInfo, err := d.DcMgr.DcGetDeviceBoardInfo(devId)
		if err != nil {
			hwlog.RunLog.Debugf("get board info failed by deviceID(%d), error: %v", devId, err)
			continue
		}
		if !common.IsValidBoardInfo(&boardInfo) {
			hwlog.RunLog.Debugf("invalid board info by deviceID(%d), error: %v", devId, err)
			continue
		}
		return boardInfo, nil
	}
	return common.BoardInfo{}, errors.New("cannot get valid board info")
}

// GetDevType return dev type
func (d *DeviceManagerV2) GetDevType() string {
	return d.DevType
}

// GetDeviceHealth query npu device health status
func (d *DeviceManagerV2) GetDeviceHealth(logicID int32) (uint32, error) {
	healthCode, err := d.DcMgr.DcGetDeviceHealth(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, err
	}

	return uint32(healthCode), nil
}

// GetDeviceNetWorkHealth query npu device network health status
func (d *DeviceManagerV2) GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	healthCode, err := d.DcMgr.DcGetDeviceNetWorkHealth(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, err
	}

	return healthCode, nil
}

// GetDeviceUtilizationRate get npu device utilization
func (d *DeviceManagerV2) GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	rate, err := d.DcMgr.DcGetDeviceUtilizationRate(logicID, deviceType)
	if err != nil {
		return common.UnRetError, err
	}

	return uint32(rate), nil
}

// GetDeviceUtilizationRateV2 get npu device utilization v2
func (d *DeviceManagerV2) GetDeviceUtilizationRateV2(logicID int32) (common.DcmiMultiUtilizationInfo, error) {
	return dcmi.BuildErrNpuMultiUtilizationInfo(), fmt.Errorf("getDeviceUtilizationRateV2 by logicID(%d) %s",
		logicID, errNotSupportedInDcmiV2)
}

// GetDeviceTemperature get npu device temperature
func (d *DeviceManagerV2) GetDeviceTemperature(logicID int32) (int32, error) {
	temp, err := d.DcMgr.DcGetDeviceTemperature(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get temperature by logicID(%d)", logicID)
	}

	return temp, nil
}

// GetDeviceVoltage get npu device voltage
func (d *DeviceManagerV2) GetDeviceVoltage(logicID int32) (float32, error) {
	voltage, err := d.DcMgr.DcGetDeviceVoltage(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get voltage by logicID(%d)", logicID)
	}

	return voltage, nil
}

// GetDevicePowerInfo get npu device power info
func (d *DeviceManagerV2) GetDevicePowerInfo(logicID int32) (float32, error) {
	power, err := d.DcMgr.DcGetDevicePowerInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get power by logicID(%d)", logicID)
	}

	return power, nil
}

// GetDeviceFrequency get npu device work frequency
func (d *DeviceManagerV2) GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (uint32, error) {
	frequency, err := d.DcMgr.DcGetDeviceFrequency(logicID, deviceType)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get frequency by logicID(%d)", logicID)
	}

	return frequency, nil
}

// GetDeviceMemoryInfo get npu memory information
func (d *DeviceManagerV2) GetDeviceMemoryInfo(logicID int32) (*common.MemoryInfo, error) {
	hwlog.RunLog.Infof("getDeviceMemoryInfo by logicID %d not support in dcmiv2 api", logicID)
	return nil, fmt.Errorf("not support in dcmiv2 api")
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func (d *DeviceManagerV2) GetDeviceHbmInfo(logicID int32) (*common.HbmInfo, error) {
	hbmInfo, err := d.DcMgr.DcGetHbmInfo(logicID)
	if err != nil {
		return nil, err
	}

	return hbmInfo, nil
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManagerV2) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	errCount, errCode, err := d.DcMgr.DcGetDeviceErrorCode(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, common.RetError, fmt.Errorf("failed to get device error code by logicID(%d)",
			logicID)
	}

	return errCount, errCode, nil
}

// GetChipInfo get npu chip info
func (d *DeviceManagerV2) GetChipInfo(logicID int32) (*common.ChipInfo, error) {
	chipInfo, err := d.DcMgr.DcGetChipInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get chip info code by logicID(%d)", logicID)
	}

	return chipInfo, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManagerV2) GetPhysicIDFromLogicID(logicID int32) (int32, error) {
	physicID, err := d.DcMgr.DcGetPhysicIDFromLogicID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get physicID by logicID(%d)", logicID)
	}

	return physicID, nil
}

// GetLogicIDFromPhysicID get device logic id from physic id
func (d *DeviceManagerV2) GetLogicIDFromPhysicID(physicID int32) (int32, error) {
	logicID, err := d.DcMgr.DcGetLogicIDFromPhysicID(physicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get logicID by physicID(%d)", physicID)
	}

	return logicID, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManagerV2) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return common.RetError, fmt.Errorf("getDeviceLogicID by cardID(%d) deviceID(%d) %s",
		cardID, deviceID, errNotSupportedInDcmiV2)
}

// GetDeviceIPAddress get device ip address
func (d *DeviceManagerV2) GetDeviceIPAddress(logicID, ipType int32) (string, error) {
	return d.DcMgr.DcGetDeviceIPAddress(logicID, ipType)
}

// CreateVirtualDevice create virtual device
func (d *DeviceManagerV2) CreateVirtualDevice(
	logicID int32, vDevInfo common.CgoCreateVDevRes) (common.CgoCreateVDevOut, error) {
	return common.CgoCreateVDevOut{}, fmt.Errorf("createVirtualDevice by logicID(%d) %s",
		logicID, errNotSupportedInDcmiV2)
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManagerV2) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	cgoVDevInfo, err := d.DcMgr.DcGetVDeviceInfo(logicID)
	if err != nil {
		hwlog.RunLog.Debug(err)
		return common.VirtualDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %v ", err)
	}
	for _, vDevInfo := range cgoVDevInfo.VDevInfo {
		if !common.IsValidTemplateName(d.DevType, vDevInfo.QueryInfo.Name) {
			return common.VirtualDevInfo{}, fmt.Errorf("vdevice id %d, it's template name is invalid: %s",
				vDevInfo.VDevID, vDevInfo.QueryInfo.Name)
		}
	}
	return cgoVDevInfo, nil
}

// DestroyVirtualDevice destroy virtual device
func (d *DeviceManagerV2) DestroyVirtualDevice(logicID int32, vDevID uint32) error {
	return fmt.Errorf("destroyVirtualDevice by logicID(%d) %s", logicID, errNotSupportedInDcmiV2)
}

// GetMcuPowerInfo get mcu power info for cardID
func (d *DeviceManagerV2) GetMcuPowerInfo(cardID int32) (float32, error) {
	return 0, fmt.Errorf("getMcuPowerInfo by cardID(%d) %s", cardID, errNotSupportedInDcmiV2)
}

// GetCardIDDeviceID get cardID and deviceID by logicID
func (d *DeviceManagerV2) GetCardIDDeviceID(logicID int32) (int32, int32, error) {
	return -1, -1, fmt.Errorf("getCardIDDeviceID by logicID(%d) %s", logicID, errNotSupportedInDcmiV2)
}

// GetProductType get product type by cardID and deviceID
func (d *DeviceManagerV2) GetProductType(logicID int32) (string, error) {
	return "", fmt.Errorf("getProductType by logicID(%d) %s", logicID, errNotSupportedInDcmiV2)
}

// GetDeviceList get all device logicID list
func (d *DeviceManagerV2) GetDeviceList() (int32, []int32, error) {
	return d.DcMgr.DcGetDeviceList()
}

// GetAllProductType get all product type
func (d *DeviceManagerV2) GetAllProductType() ([]string, error) {
	hwlog.RunLog.Debugf("getAllProductType %s", errNotSupportedInDcmiV2)
	return []string{}, nil
}

// GetNpuWorkMode get work mode of NPU
func (d *DeviceManagerV2) GetNpuWorkMode() string {
	hwlog.RunLog.Warnf("only AMP mode is available on %s", d.DevType)
	return common.AMPMode
}

// SetDeviceReset reset spec device
func (d *DeviceManagerV2) SetDeviceReset(logicID int32) error {
	return d.DcMgr.DcSetDeviceReset(logicID)
}

// GetOutBandChannelStateV2 get out band channel state
func (d *DeviceManagerV2) GetOutBandChannelStateV2(logicID int32) error {
	return d.DcMgr.DcGetOutBandChannelState(logicID)
}

// PreResetSocV2 pre reset soc, used before reset out band
func (d *DeviceManagerV2) PreResetSocV2(logicID int32) error {
	return d.DcMgr.DcPreResetSoc(logicID)
}

// SetDeviceResetOutBandV2 reset spec device out band
func (d *DeviceManagerV2) SetDeviceResetOutBandV2(logicID int32) error {
	return d.DcMgr.DcSetDeviceResetOutBand(logicID)
}

// RescanSoc trigger soc rescan, non-blocking
func (d *DeviceManagerV2) RescanSoc(logicID int32) error {
	return d.DcMgr.DcRescanSoc(logicID)
}

// GetDeviceAllErrorCode get npu device all error code
func (d *DeviceManagerV2) GetDeviceAllErrorCode(logicID int32) (int32, []int64, error) {
	errCount, errCodes, err := d.DcMgr.DcGetDeviceAllErrorCode(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, nil, fmt.Errorf("failed to get device error code by logicID(%d)", logicID)
	}
	return errCount, errCodes, nil
}

// GetDeviceAllErrorCodeWithTimeOut get npu device all error code with timeout
func (d *DeviceManagerV2) GetDeviceAllErrorCodeWithTimeOut(logicID int32, timeout time.Duration) (int32, []int64, error) {
	return common.RetError, nil, fmt.Errorf("getDeviceAllErrorCodeWithTimeOut by logicID(%d) %s",
		logicID, errNotSupportedInDcmiV2)
}

// SubscribeDeviceFaultEvent get npu device error code by subscribe
func (d *DeviceManagerV2) SubscribeDeviceFaultEvent(logicID int32) error {
	if err := d.DcMgr.DcSubscribeDeviceFaultEvent(logicID); err != nil {
		hwlog.RunLog.Error(err)
		return fmt.Errorf("failed to subscribe device error code by logicID(%d)", logicID)
	}
	return nil
}

// GetDieID return die id by dcmi die type, vdie id or ndie id
func (d *DeviceManagerV2) GetDieID(logicID int32, dcmiDieType dcmi.DieType) (string, error) {
	return d.DcMgr.DcGetDieID(logicID, dcmiDieType)
}

// GetDevProcessInfo get process and process memory in device side
func (d *DeviceManagerV2) GetDevProcessInfo(logicID int32) (*common.DevProcessInfo, error) {
	return d.DcMgr.DcGetDevProcessInfo(logicID)
}

// GetPCIeBusInfo pcie bus info
func (d *DeviceManagerV2) GetPCIeBusInfo(logicID int32) (string, error) {
	return d.DcMgr.DcGetPCIeBusInfo(logicID)
}

// GetBoardInfo return board info of device
func (d *DeviceManagerV2) GetBoardInfo(logicID int32) (common.BoardInfo, error) {
	return d.DcMgr.DcGetDeviceBoardInfo(logicID)
}

// GetCardElabelV2 get card elabel information
func (d *DeviceManagerV2) GetCardElabelV2(logicID int32) (common.ElabelInfo, error) {
	return d.DcMgr.DcGetCardElabel(logicID)
}

// GetPCIEBandwidth get pcie bandwidth
func (d *DeviceManagerV2) GetPCIEBandwidth(logicID int32, profilingTime int) (common.PCIEBwStat, error) {
	pciePCIEBw, err := d.DcMgr.DcGetPCIEBandwidth(logicID, profilingTime)
	if err != nil {
		return common.PCIEBwStat{}, err
	}
	return pciePCIEBw, nil
}

// SetIsTrainingCard identifies whether it is a training card according to the usage of card
func (d *DeviceManagerV2) SetIsTrainingCard() error {
	devType := d.GetDevType()
	if strings.HasPrefix(devType, api.Ascend310) {
		d.isTrainingCard = false
		return nil
	}

	boardInfo := common.BoardInfo{}
	devNum, devList, err := d.GetDeviceList()
	if err != nil || devNum == 0 {
		hwlog.RunLog.Errorf("failed to get device list when set 'IsTrainingCard' err: %v", err)
		return err
	}
	for _, deviceID := range devList {
		boardInfo, err = d.DcMgr.DcGetDeviceBoardInfo(deviceID)
		if err != nil {
			hwlog.RunLog.Warnf("get board info by deviceID %d failed, err: %v", deviceID, err)
			continue
		}
		break
	}

	if devType == api.Ascend910B &&
		(boardInfo.BoardId == common.A300IA2BoardId || boardInfo.BoardId == common.A300IA2GB64BoardId) {
		d.isTrainingCard = false
		return nil
	}

	d.isTrainingCard = true
	return nil
}

// GetDeviceEccInfo query device ECC info
func (d *DeviceManagerV2) GetDeviceEccInfo(logicID int32, dcmiDeviceType common.DcmiDeviceType) (*common.ECCInfo, error) {
	return d.DcMgr.DcGetDeviceEccInfo(logicID, dcmiDeviceType)
}

// GetSuperPodInfo  get 910A3 super pod info
func (d *DeviceManagerV2) GetSuperPodInfo(logicID int32) (common.CgoSuperPodInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.CgoSuperPodInfo{}, fmt.Errorf("input invalid logicID: %d", logicID)
	}

	cgoSuperPodInfo, err := d.DcMgr.DcGetSuperPodInfo(logicID)
	if err != nil {
		return common.CgoSuperPodInfo{}, fmt.Errorf("failed to get super pod info by logicID(%d), error: %v",
			logicID, err)
	}

	return cgoSuperPodInfo, nil
}

// GetSioInfo get SIO info
func (d *DeviceManagerV2) GetSioInfo(logicID int32) (*common.SioCrcErrStatisticInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("input invalid logicID when get sio info: %d", logicID)
	}

	cgoSPodSioInfo, err := d.DcMgr.DcGetSioInfo(logicID)
	if err != nil {
		return nil, err
	}

	return &cgoSPodSioInfo, nil
}

// GetHccsStatisticInfo get HCCS statistic info
func (d *DeviceManagerV2) GetHccsStatisticInfo(logicID int32) (*common.HccsStatisticInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return buildFailedHccsInfo(), fmt.Errorf("input invalid logicID when get hccs statistic info: %d", logicID)
	}
	return buildFailedHccsInfo(), fmt.Errorf("getHccsStatisticInfo by deviceID(%d) %s",
		logicID, errNotSupportedInDcmiV2)
}

// GetHccsBandwidthInfo get hccs bandwidth info
func (d *DeviceManagerV2) GetHccsBandwidthInfo(logicID int32) (*common.HccsBandwidthInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return buildFailedHccsBWInfo(), fmt.Errorf("input invalid logicID when get hccs bandwidth info: %d", logicID)
	}
	return buildFailedHccsBWInfo(), fmt.Errorf("getHccsBandwidthInfo by logicID(%d) %s",
		logicID, errNotSupportedInDcmiV2)
}

// GetChipBaseInfos get chip base info
func (d *DeviceManagerV2) GetChipBaseInfos() ([]*common.ChipBaseInfo, error) {
	_, devList, err := d.DcMgr.DcGetDeviceList()
	if err != nil {
		return nil, fmt.Errorf("get device list failed, error: %v", err)
	}
	var chips []*common.ChipBaseInfo
	for _, logicID := range devList {
		physicID, err := d.DcMgr.DcGetPhysicIDFromLogicID(logicID)
		if err != nil {
			return nil, fmt.Errorf("get device (logicID: %d) physic id failed, error: %v", logicID, err)
		}
		hwlog.RunLog.Infof("get chip base info, logicID: %d, physicID: %d", logicID, physicID)
		chips = append(chips, &common.ChipBaseInfo{
			PhysicID: physicID,
			LogicID:  logicID,
			CardID:   -1,
			DeviceID: -1,
		})
	}
	return chips, nil
}

// StartHccsPingMesh or UB PingMesh depending on device type
func (d *DeviceManagerV2) StartHccsPingMesh(logicID int32, portID int, operate common.HccspingMeshOperate) error {
	devType := d.GetDevType()
	if devType == common.Ascend910A5 {
		return d.DcMgr.DcStartUbPingMesh(logicID, operate)
	}
	hwlog.RunLog.Errorf("devType: %v, StartHccsPingMesh is not support", devType)
	return fmt.Errorf("devType: %v, StartHccsPingMesh is not support", devType)
}

// DcStartHccsPingMesh or UB PingMesh depending on device type
func (d *DeviceManagerV2) DcStartHccsPingMesh(logicID int32, portID int,
	operate common.HccspingMeshOperate) error {
	devType := d.GetDevType()
	if devType == common.Ascend910A5 {
		return d.DcMgr.DcStartUbPingMesh(logicID, operate)
	}
	hwlog.RunLog.Errorf("devType: %v, DcStartHccsPingMesh is not support", devType)
	return fmt.Errorf("devType: %v, DcStartHccsPingMesh is not support", devType)
}

// StopHccsPingMesh or UB PingMesh depending on device type
func (d *DeviceManagerV2) StopHccsPingMesh(logicID int32, portID int, taskID uint) error {
	devType := d.GetDevType()
	if devType == common.Ascend910A5 {
		return d.DcMgr.DcStopUbPingMesh(logicID, taskID)
	}
	hwlog.RunLog.Errorf("devType: %v, StopHccsPingMesh is not support", devType)
	return fmt.Errorf("devType: %v, StopHccsPingMesh is not support", devType)
}

// GetHccsPingMeshInfo or UB PingMesh depending on device type
func (d *DeviceManagerV2) GetHccsPingMeshInfo(logicID int32, portID int, taskID uint) (*common.HccspingMeshInfo, error) {
	devType := d.GetDevType()
	if devType == common.Ascend910A5 {
		return d.DcMgr.DcGetUbPingMeshInfo(logicID, taskID, common.UbPingMeshMaxNum)
	}
	hwlog.RunLog.Errorf("devType: %v, GetHccsPingMeshInfo is not support", devType)
	return nil, fmt.Errorf("devType: %v, GetHccsPingMeshInfo is not support", devType)
}

// GetHccsPingMeshState or UB PingMesh depending on device type
func (d *DeviceManagerV2) GetHccsPingMeshState(logicID int32, portID int, taskID uint) (int, error) {
	devType := d.GetDevType()
	if devType == common.Ascend910A5 {
		return d.DcMgr.DcGetUbPingMeshState(logicID, taskID)
	}
	hwlog.RunLog.Errorf("devType: %v, GetHccsPingMeshState is not support", devType)
	return common.RetError, fmt.Errorf("devType: %v, GetHccsPingMeshState is not support", devType)
}

// GetSuperPodStatus get super pod status
func (d *DeviceManagerV2) GetSuperPodStatus(logicID int32, sdid uint32) (int, error) {
	hwlog.RunLog.Errorf("get super pod status failed, logicID: %d, sdid: %d, error: %v",
		logicID, sdid, errNotSupportedInDcmiV2)
	return -1, fmt.Errorf("set super pod status failed, logicID: %d, sdid: %d, error: %v",
		logicID, sdid, errNotSupportedInDcmiV2)
}

// SetSuperPodStatus set super pod status
func (d *DeviceManagerV2) SetSuperPodStatus(logicID int32, sdid, status uint32) error {
	hwlog.RunLog.Errorf("set super pod status failed, logicID: %d, sdid: %d, status: %d, error: %v",
		logicID, sdid, status, errNotSupportedInDcmiV2)
	return fmt.Errorf("set super pod status failed, logicID: %d, sdid: %d, status: %d, error: %v",
		logicID, sdid, status, errNotSupportedInDcmiV2)
}
