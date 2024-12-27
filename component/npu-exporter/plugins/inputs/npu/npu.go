/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package npu this for parse and pack
package npu

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultLogPath = "/var/log/mindx-dl/npu-exporter/npu-plugin.log"

	aiCore  = common.DeviceType(2)
	hbm     = common.DeviceType(6)
	overall = common.DeviceType(13)

	mega                = 1024 * 1024
	maxLogBackups       = 2
	defaultLogCacheSize = 2 * 1024
	defaultLogFileSize  = 2

	receivedFieldsNil = "received fields is incorrect, fields is nil"
	dcmiHccsMaxCounts = 8
)

const (
	txPower0 = "Tx_Power0"
	txPower1 = "Tx_Power1"
	txPower2 = "Tx_Power2"
	txPower3 = "Tx_Power3"

	rxPower0 = "Rx_Power0"
	rxPower1 = "Rx_Power1"
	rxPower2 = "Rx_Power2"
	rxPower3 = "Rx_Power3"

	present     = "present"
	temperature = "temperature"
	voltage     = "Vcc"
)

//go:embed sample.conf
var sampleConfig string

// WatchNPU npu watch struct
type WatchNPU struct {
	NpuLogPath  string `toml:"npu_log_path"`
	NpuLogLevel int    `toml:"npu_log_level"`
	devManager  devmanager.DeviceInterface
}

// SampleConfig used to return sampleConfig
func (*WatchNPU) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (npu *WatchNPU) Init() error {
	if npu.NpuLogPath == "" {
		npu.NpuLogPath = defaultLogPath
	}
	var hwLogConfig = &hwlog.LogConfig{
		LogFileName: npu.NpuLogPath,
		ExpiredTime: hwlog.DefaultExpiredTime,
		CacheSize:   defaultLogCacheSize,
		FileMaxSize: defaultLogFileSize,
		LogLevel:    npu.NpuLogLevel,
		MaxAge:      hwlog.DefaultMinSaveAge,
		MaxBackups:  maxLogBackups}

	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	dmgr, err := devmanager.AutoInit("")
	if err != nil {
		return fmt.Errorf("init dev manager failed: %v", err)
	}
	npu.devManager = dmgr
	return nil
}

// parseOptInfoForCTYun parse optical info of NPU for CT Yun
func parseOptInfoForCTYun(opticalInfo map[string]string) map[string]interface{} {
	ctYunOpticalInfo := make(map[string]interface{})
	var ctYunFloatDataKeys = []string{
		txPower0,
		txPower1,
		txPower2,
		txPower3,
		rxPower0,
		rxPower1,
		rxPower2,
		rxPower3,
		voltage,
		temperature,
	}
	var ctYunTelegrafKeys = []string{
		"npu_chip_optical_tx_power_0",
		"npu_chip_optical_tx_power_1",
		"npu_chip_optical_tx_power_2",
		"npu_chip_optical_tx_power_3",
		"npu_chip_optical_rx_power_0",
		"npu_chip_optical_rx_power_1",
		"npu_chip_optical_rx_power_2",
		"npu_chip_optical_rx_power_3",
		"npu_chip_optical_vcc",
		"npu_chip_optical_temp",
	}

	for i, ctYunOpticalKey := range ctYunFloatDataKeys {
		floatData := hccn.GetFloatDataFromStr(opticalInfo[ctYunOpticalKey], ctYunOpticalKey)
		if floatData == common.RetError {
			continue
		}
		ctYunOpticalInfo[ctYunTelegrafKeys[i]] = floatData
	}

	optState := 0
	if opticalInfo[present] == present {
		optState = 1
	}
	ctYunOpticalInfo["npu_chip_optical_state"] = optState

	return ctYunOpticalInfo
}

func (npu *WatchNPU) packDcmiInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	npu.collectHealthStatus(devID, fields, acc)
	if info, err := npu.devManager.GetDevProcessInfo(devID); err != nil {
		acc.AddError(fmt.Errorf("get npu process info failed: %v", err))
	} else {
		fields["npu_chip_info_process_info_num"] = info.ProcNum
	}
	if temp, err := npu.devManager.GetDeviceTemperature(devID); err != nil {
		acc.AddError(fmt.Errorf("get npu temperature failed: %v", err))
	} else {
		fields["npu_chip_info_temperature"] = float64(temp)
	}
	npu.collectUtilizationRate(devID, fields, acc)
	if hbmInfo, err := npu.devManager.GetDeviceHbmInfo(devID); err != nil {
		acc.AddError(fmt.Errorf("get hbm info of npu failed: %v", err))
	} else {
		fields["npu_chip_info_hbm_used_memory"] = hbmInfo.Usage * mega
	}
	if power, err := npu.devManager.GetDevicePowerInfo(devID); err != nil {
		acc.AddError(fmt.Errorf("get power of npu failed: %v", err))
	} else {
		fields["npu_chip_info_power"] = power
	}
	npu.collectSioInfo(devID, fields, acc)
	npu.collectHccsInfo(devID, fields, acc)
	npu.collectVoltageInfo(devID, fields, acc)
	npu.collectNpuFreqInfo(devID, fields, acc)
	npu.collectNpuProcessInfo(devID, fields, acc)
	codeNum, errCodes, err := npu.devManager.GetDeviceAllErrorCode(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get err code failed: %v", err))
		return
	}
	if len(errCodes) > 0 {
		fields["npu_chip_info_error_code"] = errCodes[0]
	}
	// conversion of "codeNum" here is safe because codeNum <= 128
	for i := 1; i < int(codeNum); i++ {
		errCodeKey := "npu_chip_info_error_code_" + strconv.Itoa(i)
		fields[errCodeKey] = errCodes[i]
	}
}

func (npu *WatchNPU) collectSioInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	if npu.devManager.GetDevType() == common.Ascend910A3 {
		if sioInfo, err := npu.devManager.GetSioInfo(devID); err != nil {
			acc.AddError(fmt.Errorf("get sio info of npu failed: %v", err))
		} else {
			fields["npu_chip_info_sio_crc_tx_err_cnt"] = sioInfo.TxErrCnt
			fields["npu_chip_info_sio_crc_rx_err_cnt"] = sioInfo.RxErrCnt
		}
	}
}

// collectVoltageInfo collect voltage of npu
func (npu *WatchNPU) collectVoltageInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	vol, err := npu.devManager.GetDeviceVoltage(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get voltage of npu failed: %v", err))
		return
	}
	fields["npu_chip_info_voltage"] = vol
}

// collectNpuFreqInfo collect current freq of npu
func (npu *WatchNPU) collectNpuFreqInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	freq, err := npu.devManager.GetDeviceFrequency(devID, common.AICoreCurrentFreq)
	if err != nil {
		acc.AddError(fmt.Errorf("get current freq of npu failed: %v", err))
		return
	}
	fields["npu_chip_info_aicore_current_freq"] = freq

}

// collectNpuProcessInfo collect current process memeory of npu
func (npu *WatchNPU) collectNpuProcessInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	devProcessInfo, err := npu.devManager.GetDevProcessInfo(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get current process info of npu failed: %v", err))
		return
	}
	if devProcessInfo.ProcNum == 0 {
		fields["npu_chip_info_process_info"] = 0
		return
	}
	for i := int32(0); i < devProcessInfo.ProcNum; i++ {
		procInfo := devProcessInfo.DevProcArray[i]
		fields["npu_chip_info_process_info_"+strconv.Itoa(int(procInfo.Pid))] = procInfo.MemUsage
	}
}

// collectHccsInfo collect hccs info
func (npu *WatchNPU) collectHccsInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	devType := npu.devManager.GetDevType()
	if devType != common.Ascend910B && devType != common.Ascend910A3 {
		return
	}
	hccsStatisticInfo, err := npu.devManager.GetHccsStatisticInfo(devID)
	if err != nil {
		if needPrint, extraErrLog := hwlog.IsNeedPrint(common.DomainForHccs, devID); needPrint {
			acc.AddError(fmt.Errorf("get hccs statistic info of npu failed: %v %s", err, extraErrLog))
		}
	} else {
		hwlog.ResetErrCnt(common.DomainForHccs, devID)
	}
	var hccsBeginIndex int
	if devType == common.Ascend910B || common.IsA900A3SuperPod(npu.devManager.GetMainBoardId()) {
		// 910B or A900A3SuperPod begin at 1st bit
		hccsBeginIndex = 1
	} else if common.IsA9000A3SuperPod(npu.devManager.GetMainBoardId()) {
		// A9000A3SuperPod begin at 2nd bit
		hccsBeginIndex = 2
	}
	for i := hccsBeginIndex; i < dcmiHccsMaxCounts; i++ {
		doUpdateFields(acc, fields, "npu_chip_info_hccs_statistic_info_tx_cnt_"+fmt.Sprintf("%d", i),
			hccsStatisticInfo.TxCnt[i])
		doUpdateFields(acc, fields, "npu_chip_info_hccs_statistic_info_rx_cnt_"+fmt.Sprintf("%d", i),
			hccsStatisticInfo.RxCnt[i])
		doUpdateFields(acc, fields, "npu_chip_info_hccs_statistic_info_crc_err_cnt_"+fmt.Sprintf("%d", i),
			hccsStatisticInfo.CrcErrCnt[i])
		doUpdateFields(acc, fields, "npu_chip_info_hccs_statistic_info_crc_err_cnt_"+fmt.Sprintf("%d", i),
			hccsStatisticInfo.CrcErrCnt[i])
	}
	hccsBandwidthInfo, err := npu.devManager.GetHccsBandwidthInfo(devID)
	if err != nil {
		if needPrint, extraErrLog := hwlog.IsNeedPrint(common.DomainForHccsBW, devID); needPrint {
			acc.AddError(fmt.Errorf("get hccs bandwidth info of npu failed: %v %s", err, extraErrLog))
		}
	} else {
		hwlog.ResetErrCnt(common.DomainForHccsBW, devID)
	}
	doUpdateFields(acc, fields, "npu_chip_info_hccs_bandwidth_info_profiling_time", hccsBandwidthInfo.ProfilingTime)
	doUpdateFields(acc, fields, "npu_chip_info_hccs_bandwidth_info_total_tx", hccsBandwidthInfo.TotalTxbw)
	doUpdateFields(acc, fields, "npu_chip_info_hccs_bandwidth_info_total_rx", hccsBandwidthInfo.TotalTxbw)
	for i := hccsBeginIndex; i < dcmiHccsMaxCounts; i++ {
		doUpdateFields(acc, fields, "npu_chip_info_hccs_bandwidth_info_tx_"+fmt.Sprintf("%d", i),
			hccsBandwidthInfo.TxBandwidth[i])
		doUpdateFields(acc, fields, "npu_chip_info_hccs_bandwidth_info_rx_"+fmt.Sprintf("%d", i),
			hccsBandwidthInfo.RxBandwidth[i])
	}
}

// doUpdateFields update fields
func doUpdateFields(acc telegraf.Accumulator, fields map[string]interface{}, key string, value interface{}) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}
	var finalValue float64

	switch value.(type) {
	case float64:
		finalValue = value.(float64)
	case uint32:
		finalValue = float64(value.(uint32))
	default:
		hwlog.RunLog.Warn("Invalid param in function doUpdateHccsMetric")
	}

	if finalValue == common.FailedValue {
		finalValue = common.FailedMetricValue
	}
	fields[key] = finalValue
}

func (npu *WatchNPU) collectUtilizationRate(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	if aiCoreUtil, err := npu.devManager.GetDeviceUtilizationRate(devID, aiCore); err != nil {
		if needPrint, extraErrLog := hwlog.IsNeedPrint(common.DomainForAICoreUtilization, devID); needPrint {
			acc.AddError(fmt.Errorf("failed to get ai core rate for device(logicID:%d): %v %s", devID, err, extraErrLog))
		}
	} else {
		fields["npu_chip_info_utilization"] = float64(aiCoreUtil)
		hwlog.ResetErrCnt(common.DomainForAICoreUtilization, devID)
	}

	if hbmUtil, err := npu.devManager.GetDeviceUtilizationRate(devID, hbm); err != nil {
		if needPrint, extraErrLog := hwlog.IsNeedPrint(common.DomainForHbmUtilization, devID); needPrint {
			acc.AddError(fmt.Errorf("failed to get HBM(High Bandwidth Memory) utilization for device(logicID:%d): %v %s",
				devID, err, extraErrLog))
		}
	} else {
		hwlog.ResetErrCnt(common.DomainForHbmUtilization, devID)
		fields["npu_chip_info_hbm_utilization"] = float64(hbmUtil)
	}
	if overallUtil, err := npu.devManager.GetDeviceUtilizationRate(devID, overall); err != nil {
		if needPrint, extraErrLog := hwlog.IsNeedPrint(common.DomainForOverallUtilization, devID); needPrint {
			acc.AddError(fmt.Errorf("failed to get overall utilization rate for device(logicID:%d): %v %s",
				devID, err, extraErrLog))
		}
	} else {
		hwlog.ResetErrCnt(common.DomainForOverallUtilization, devID)
		fields["npu_chip_info_overall_utilization"] = float64(overallUtil)
	}
}

func (npu *WatchNPU) collectHealthStatus(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	if health, err := npu.devManager.GetDeviceHealth(devID); err != nil {
		acc.AddError(fmt.Errorf("get health of npu failed: %v", err))
	} else {
		fields["npu_chip_info_health_status"] = hccn.GetHealthCode(health)
	}

	if netCode, err := npu.devManager.GetDeviceNetWorkHealth(devID); err != nil {
		acc.AddError(fmt.Errorf("get npu Net health failed: %v", err))
	} else {
		fields["npu_chip_info_network_status"] = hccn.GetNetworkHealthy(netCode)
	}
}

func (npu *WatchNPU) packHccnInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) error {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return nil
	}

	phyID, err := npu.devManager.GetPhysicIDFromLogicID(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get devID of npu failed: %v", err))
		return err
	}
	if linkStatus, err := hccn.GetNPULinkStatus(phyID); err != nil {
		acc.AddError(fmt.Errorf("get link status of npu failed: %v", err))
	} else {
		fields["npu_chip_info_link_status"] = hccn.GetLinkStatusCode(linkStatus)
	}
	if tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID); err != nil {
		acc.AddError(fmt.Errorf("get bandwidth of npu failed: %v", err))
	} else {
		fields["npu_chip_info_bandwidth_rx"] = rx * mega
		fields["npu_chip_info_bandwidth_tx"] = tx * mega
	}
	if speed, err := hccn.GetNPULinkSpeed(phyID); err != nil {
		acc.AddError(fmt.Errorf("get link speed of npu failed: %v", err))
	} else {
		fields["npu_chip_link_speed"] = speed * mega
	}
	if linkUpCnt, err := hccn.GetNPULinkUpNum(phyID); err != nil {
		acc.AddError(fmt.Errorf("get link up count of npu failed: %v", err))
	} else {
		fields["npu_chip_link_up_num"] = linkUpCnt
	}
	collectNPUStatInfo(phyID, fields, acc)
	opticalInfo, err := hccn.GetNPUOpticalInfo(phyID)
	if err != nil {
		acc.AddError(fmt.Errorf("get optical info of npu failed: %v", err))
		return err
	}
	ctYunOpticalInfo := parseOptInfoForCTYun(opticalInfo)
	if ctYunOpticalInfo == nil {
		errMsg := fmt.Errorf("parse optical info of NPU for CT Yun failed, ctYun optical info map is nil")
		acc.AddError(errMsg)
		return errMsg
	}
	for k, v := range ctYunOpticalInfo {
		fields[k] = v
	}
	return nil
}

func collectNPUStatInfo(phyID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	if fields == nil {
		acc.AddError(fmt.Errorf(receivedFieldsNil))
		return
	}

	statInfo, err := hccn.GetNPUStatInfo(phyID)
	if err != nil {
		acc.AddError(fmt.Errorf("get stat info of npu failed: %v", err))
	} else {
		fields["npu_chip_mac_rx_pause_num"] = statInfo["mac_rx_mac_pause_num"]
		fields["npu_chip_mac_tx_pause_num"] = statInfo["mac_tx_mac_pause_num"]
		fields["npu_chip_mac_rx_pfc_pkt_num"] = statInfo["mac_rx_pfc_pkt_num"]
		fields["npu_chip_mac_tx_pfc_pkt_num"] = statInfo["mac_tx_pfc_pkt_num"]
		fields["npu_chip_mac_rx_bad_pkt_num"] = statInfo["mac_rx_bad_pkt_num"]
		fields["npu_chip_mac_tx_bad_pkt_num"] = statInfo["mac_tx_bad_pkt_num"]
		fields["npu_chip_roce_rx_all_pkt_num"] = statInfo["roce_rx_all_pkt_num"]
		fields["npu_chip_roce_tx_all_pkt_num"] = statInfo["roce_tx_all_pkt_num"]

		fields["npu_chip_roce_rx_err_pkt_num"] = statInfo["roce_rx_err_pkt_num"]
		fields["npu_chip_roce_tx_err_pkt_num"] = statInfo["roce_tx_err_pkt_num"]

		fields["npu_chip_roce_rx_cnp_pkt_num"] = statInfo["roce_rx_cnp_pkt_num"]
		fields["npu_chip_roce_tx_cnp_pkt_num"] = statInfo["roce_tx_cnp_pkt_num"]

		fields["npu_chip_mac_tx_bad_oct_num"] = statInfo["mac_tx_bad_oct_num"]
		fields["npu_chip_mac_rx_bad_oct_num"] = statInfo["mac_rx_bad_oct_num"]

		fields["npu_chip_roce_unexpected_ack_num"] = statInfo["roce_unexpected_ack_num"]
		fields["npu_chip_roce_out_of_order_num"] = statInfo["roce_out_of_order_num"]
		fields["npu_chip_roce_verification_err_num"] = statInfo["roce_verification_err_num"]
		fields["npu_chip_roce_qp_status_err_num"] = statInfo["roce_qp_status_err_num"]
		fields["npu_chip_roce_new_pkt_rty_num"] = statInfo["roce_new_pkt_rty_num"]
	}
}

// Gather used to gather information from dcmi info and hccn tool info
func (npu *WatchNPU) Gather(acc telegraf.Accumulator) error {
	if npu.devManager == nil {
		return errors.New("empty dev object")
	}
	devNum, devList, err := npu.devManager.GetDeviceList()
	if err != nil {
		acc.AddError(fmt.Errorf("get npu list failed: %s", err))
		return err
	}

	const devName = "ascend"
	devTag := make(map[string]string)
	devTagValue := "unsupported"
	if cardType := npu.devManager.GetDevType(); cardType == common.Ascend910A3 || cardType == common.Ascend910B ||
		cardType == common.Ascend910 {
		devTagValue = common.Chip910
	}

	for i := int32(0); i < devNum; i++ {
		fields := make(map[string]interface{})

		npu.packDcmiInfo(devList[i], fields, acc)
		if err := npu.packHccnInfo(devList[i], fields, acc); err != nil {
			acc.AddError(fmt.Errorf("get hccn tool info failed: %s", err))
		}

		devTag["device"] = devTagValue + "-" + strconv.Itoa(int(devList[i]))
		acc.AddFields(devName, fields, devTag)
	}

	return nil
}

func init() {
	inputs.Add("npu", func() telegraf.Input { return &WatchNPU{} })
}
