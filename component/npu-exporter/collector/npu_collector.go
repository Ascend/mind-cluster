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

// Package collector for Prometheus
package collector

import (
	"context"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/cache"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
	"ascend-common/devmanager/hccn"
	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/versions"
)

// metric label name
const (
	npuID       = "id"
	pcieBwType  = "pcie_bw_type"
	avgPcieBw   = "avgPcieBw"
	minPcieBw   = "minPcieBw"
	maxPcieBw   = "maxPcieBw"
	modelName   = "model_name"
	npuUUID     = "vdie_id"
	vNpuUUID    = "v_dev_id"
	npuPCIEInfo = "pcie_bus_info"
	namespace   = "namespace"
	podName     = "pod_name"
	cntrName    = "container_name"
	isVirtual   = "is_virtual"
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

	notPresent  = "not present"
	present     = "present"
	temperature = "temperature"
	voltage     = "Vcc"
)

const (
	macRxMacPauseNum       = "mac_rx_mac_pause_num"
	macTxMacPauseNum       = "mac_tx_mac_pause_num"
	macRxPfcPktNum         = "mac_rx_pfc_pkt_num"
	macTxPfcPktNum         = "mac_tx_pfc_pkt_num"
	macRxBadPktNum         = "mac_rx_bad_pkt_num"
	macTxBadPktNum         = "mac_tx_bad_pkt_num"
	roCERxAllPktNum        = "roce_rx_all_pkt_num"
	roCETxAllPktNum        = "roce_tx_all_pkt_num"
	roCERxErrPktNum        = "roce_rx_err_pkt_num"
	roCETxErrPktNum        = "roce_tx_err_pkt_num"
	roCERxCnpPktNum        = "roce_rx_cnp_pkt_num"
	roCETxCnpPktNum        = "roce_tx_cnp_pkt_num"
	macRxBadOctNum         = "mac_rx_bad_oct_num"
	macTxBadOctNum         = "mac_tx_bad_oct_num"
	roCEUnexpectedAckNum   = "roce_unexpected_ack_num"
	roCEOutOfOrderNum      = "roce_out_of_order_num"
	roCEVerificationErrNum = "roce_verification_err_num"
	roCEQpStatusErrNum     = "roce_qp_status_err_num"
	roCENewPktRtyNum       = "roce_new_pkt_rty_num"
	roCEEcnDBNum           = "roce_ecn_db_num"
	macRXFcsErrPktNum      = "mac_rx_fcs_err_pkt_num"
)

var (
	cardLabel = []string{npuID, modelName, npuUUID, npuPCIEInfo, namespace, podName, cntrName}
)

var (
	versionInfoDesc = prometheus.NewDesc("npu_exporter_version_info",
		"exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc = prometheus.NewDesc("machine_npu_nums",
		"Amount of npu installed on the machine.", nil, nil)
	npuChipInfoDescNpuName = prometheus.NewDesc("npu_chip_info_name",
		"the Ascend npu name with value '1'", []string{npuID, "name", npuUUID, npuPCIEInfo, namespace, podName,
			cntrName}, nil)
	npuChipInfoDescUtil = prometheus.NewDesc("npu_chip_info_utilization",
		"the ai core utilization", cardLabel, nil)
	npuChipInfoDescOverUtil = prometheus.NewDesc("npu_chip_info_overall_utilization",
		"the overall utilization of npu", cardLabel, nil)
	npuChipInfoDescVectorUtil = prometheus.NewDesc("npu_chip_info_vector_utilization",
		"the vector ai core utilization", cardLabel, nil)
	npuChipInfoDescTemp = prometheus.NewDesc("npu_chip_info_temperature",
		"the npu temperature", cardLabel, nil)
	npuChipInfoDescPower = prometheus.NewDesc("npu_chip_info_power",
		"the npu power", cardLabel, nil)
	npuChipInfoDescVoltage = prometheus.NewDesc("npu_chip_info_voltage",
		"the npu voltage", cardLabel, nil)
	npuChipInfoDescUsedMemory = prometheus.NewDesc("npu_chip_info_used_memory",
		"the npu used memory", cardLabel, nil)
	npuChipInfoDescTotalMemory = prometheus.NewDesc("npu_chip_info_total_memory",
		"the npu total memory", cardLabel, nil)
	npuChipInfoDescHealthStatus = prometheus.NewDesc("npu_chip_info_health_status",
		"the npu health status", cardLabel, nil)
	npuChipInfoDescHbmUsedMemory = prometheus.NewDesc("npu_chip_info_hbm_used_memory",
		"the npu hbm used memory", cardLabel, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory",
		"the npu hbm total memory", cardLabel, nil)
	npuChipInfoDescErrorCode = prometheus.NewDesc("npu_chip_info_error_code",
		"the First npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode1 = prometheus.NewDesc("npu_chip_info_error_code_1",
		"the Second npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode2 = prometheus.NewDesc("npu_chip_info_error_code_2",
		"the Third npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode3 = prometheus.NewDesc("npu_chip_info_error_code_3",
		"the Fourth npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode4 = prometheus.NewDesc("npu_chip_info_error_code_4",
		"the Fifth npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode5 = prometheus.NewDesc("npu_chip_info_error_code_5",
		"the Sixth npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode6 = prometheus.NewDesc("npu_chip_info_error_code_6",
		"the Seventh npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode7 = prometheus.NewDesc("npu_chip_info_error_code_7",
		"the Eighth npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode8 = prometheus.NewDesc("npu_chip_info_error_code_8",
		"the Ninth npu error code", cardLabel, nil)
	npuChipInfoDescErrorCode9 = prometheus.NewDesc("npu_chip_info_error_code_9",
		"the Tenth npu error code", cardLabel, nil)
	npuChipInfoDescLinkStatus = prometheus.NewDesc("npu_chip_info_link_status",
		"the npu link status", cardLabel, nil)
	npuChipInfoDescNetworkStatus = prometheus.NewDesc("npu_chip_info_network_status",
		"the npu network health status", cardLabel, nil)
	npuChipInfoDescBandwidthTx = prometheus.NewDesc("npu_chip_info_bandwidth_tx",
		"the npu interface transport speed, unit is 'MB/s'", cardLabel, nil)
	npuChipInfoDescBandwidthRx = prometheus.NewDesc("npu_chip_info_bandwidth_rx",
		"the npu interface receive speed, unit is 'MB/s'", cardLabel, nil)
	npuChipInfoDescRxPBW = prometheus.NewDesc("npu_chip_info_pcie_rx_p_bw",
		"the npu write bw to remote‘s speed, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescRxNpBW = prometheus.NewDesc("npu_chip_info_pcie_rx_np_bw",
		"the npu read bw's speed from remote, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescRxCplBW = prometheus.NewDesc("npu_chip_info_pcie_rx_cpl_bw",
		"the npu reply remote read operate cpl's speed, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescTxPBW = prometheus.NewDesc("npu_chip_info_pcie_tx_p_bw",
		"the npu receive remote write operate's speed, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescTxNpBW = prometheus.NewDesc("npu_chip_info_pcie_tx_np_bw",
		"the npu receive remote read operate's speed, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescTxCplBW = prometheus.NewDesc("npu_chip_info_pcie_tx_cpl_bw",
		"the npu read cpl's responese bw speed from remote, unit is 'MB/ms'", pcieBwLabel, nil)
	npuChipInfoDescRxECNNum = prometheus.NewDesc("npu_chip_info_rx_ecn_num",
		"the npu network ecn receive number", cardLabel, nil)
	npuChipInfoDescRxFCSNum = prometheus.NewDesc("npu_chip_info_rx_fcs_num",
		"the npu network fcs receive number", cardLabel, nil)
	npuChipLinkSpeed = prometheus.NewDesc("npu_chip_link_speed",
		"the npu interface receive link speed, unit is 'Mb/s'", cardLabel, nil)
	npuChipLinkUpNum = prometheus.NewDesc("npu_chip_link_up_num",
		"the npu interface receive link-up num", cardLabel, nil)
	npuChipMacRxPauseNum = prometheus.NewDesc("npu_chip_mac_rx_pause_num",
		"the npu interface receive mac-rx-pause-num", cardLabel, nil)
	npuChipMacTxPauseNum = prometheus.NewDesc("npu_chip_mac_tx_pause_num",
		"the npu interface receive mac-tx-pause-num", cardLabel, nil)
	npuChipMacRxPfcPktNum = prometheus.NewDesc("npu_chip_mac_rx_pfc_pkt_num",
		"the npu interface receive mac-rx-pfc-pkt-num", cardLabel, nil)
	npuChipMacTxPfcPktNum = prometheus.NewDesc("npu_chip_mac_tx_pfc_pkt_num",
		"the npu interface receive mac-tx-pfc-pkt-num", cardLabel, nil)
	npuChipMacRxBadPktNum = prometheus.NewDesc("npu_chip_mac_rx_bad_pkt_num",
		"the npu interface receive mac-rx-bad-pkt-num", cardLabel, nil)
	npuChipMacTxBadPktNum = prometheus.NewDesc("npu_chip_mac_tx_bad_pkt_num",
		"the npu interface receive mac-tx-bad-pkt-num", cardLabel, nil)
	npuChipRoceRxAllPktNum = prometheus.NewDesc("npu_chip_roce_rx_all_pkt_num",
		"the npu interface receive roce-rx-all-pkt-num", cardLabel, nil)
	npuChipRoceTxAllPktNum = prometheus.NewDesc("npu_chip_roce_tx_all_pkt_num",
		"the npu interface receive roce-tx-all-pkt-num", cardLabel, nil)
	npuChipRoceRxErrPktNum = prometheus.NewDesc("npu_chip_roce_rx_err_pkt_num",
		"the npu interface receive roce-rx-err-pkt-num", cardLabel, nil)
	npuChipRoceTxErrPktNum = prometheus.NewDesc("npu_chip_roce_tx_err_pkt_num",
		"the npu interface receive roce-tx-err-pkt-num", cardLabel, nil)
	npuChipRoceRxCnpPktNum = prometheus.NewDesc("npu_chip_roce_rx_cnp_pkt_num",
		"the npu interface receive roce-rx-cnp-pkt-num", cardLabel, nil)
	npuChipRoceTxCnpPktNum = prometheus.NewDesc("npu_chip_roce_tx_cnp_pkt_num",
		"the npu interface receive roce-tx-cnp-pkt-num", cardLabel, nil)
	npuChipRoceNewPktRtyNum = prometheus.NewDesc("npu_chip_roce_new_pkt_rty_num",
		"the npu interface receive roce-new-pkt-rty-num", cardLabel, nil)
	npuChipMacTxBadOctNum = prometheus.NewDesc("npu_chip_mac_tx_bad_oct_num",
		"the npu interface receive mac-tx-bad-oct-num", cardLabel, nil)
	npuChipMacRxBadOctNum = prometheus.NewDesc("npu_chip_mac_rx_bad_oct_num",
		"the npu interface receive mac-rx-bad-oct-num", cardLabel, nil)
	npuChipRoceUnexpectedAcktNum = prometheus.NewDesc("npu_chip_roce_unexpected_ack_num",
		"the npu interface receive roce-unexpected-ack-num", cardLabel, nil)
	npuChipRoceOutOfOrderNum = prometheus.NewDesc("npu_chip_roce_out_of_order_num",
		"the npu interface receive roce-out-of-order-num", cardLabel, nil)
	npuChipRoceVerificationErrNum = prometheus.NewDesc("npu_chip_roce_verification_err_num",
		"the npu interface receive roce-verification-err-num", cardLabel, nil)
	npuChipRoceQpStatusErrNum = prometheus.NewDesc("npu_chip_roce_qp_status_err_num",
		"the npu interface receive roce-qp-status-err-num", cardLabel, nil)
	npuChipOpticalState = prometheus.NewDesc("npu_chip_optical_state",
		"the npu interface receive optical-state", cardLabel, nil)
	npuChipOpticalTxPower0 = prometheus.NewDesc("npu_chip_optical_tx_power_0",
		"the npu interface receive optical-tx-power-0", cardLabel, nil)
	npuChipOpticalTxPower1 = prometheus.NewDesc("npu_chip_optical_tx_power_1",
		"the npu interface receive optical-tx-power-1", cardLabel, nil)
	npuChipOpticalTxPower2 = prometheus.NewDesc("npu_chip_optical_tx_power_2",
		"the npu interface receive optical-tx-power-2", cardLabel, nil)
	npuChipOpticalTxPower3 = prometheus.NewDesc("npu_chip_optical_tx_power_3",
		"the npu interface receive optical-tx-power-3", cardLabel, nil)
	npuChipOpticalRxPower0 = prometheus.NewDesc("npu_chip_optical_rx_power_0",
		"the npu interface receive optical-rx-power-0", cardLabel, nil)
	npuChipOpticalRxPower1 = prometheus.NewDesc("npu_chip_optical_rx_power_1",
		"the npu interface receive optical-rx-power-1", cardLabel, nil)
	npuChipOpticalRxPower2 = prometheus.NewDesc("npu_chip_optical_rx_power_2",
		"the npu interface receive optical-rx-power-2", cardLabel, nil)
	npuChipOpticalRxPower3 = prometheus.NewDesc("npu_chip_optical_rx_power_3",
		"the npu interface receive optical-rx-power-3", cardLabel, nil)
	npuChipOpticalVcc = prometheus.NewDesc("npu_chip_optical_vcc",
		"the npu interface receive optical-vcc", cardLabel, nil)
	npuChipOpticalTemp = prometheus.NewDesc("npu_chip_optical_temp",
		"the npu interface receive optical-temperature", cardLabel, nil)
	npuChipInfoDescDevProcessInfo = prometheus.NewDesc("npu_chip_info_process_info",
		"the npu process info, unit is 'MB'. if process run on host, container_id and container_name will be empty",
		[]string{npuID, modelName, npuUUID, "process_id", "container_id", cntrName, npuPCIEInfo, namespace,
			podName}, nil)
	npuChipInfoDescAICoreFreqInfo = prometheus.NewDesc("npu_chip_info_aicore_current_freq",
		"the npu ai core current frequency, unit is 'MHz'", cardLabel, nil)
	npuContainerInfo = prometheus.NewDesc("npu_container_info",
		"the container name and deviceID relationship", []string{"containerID", "containerName", "npuID", modelName,
			npuUUID, npuPCIEInfo, namespace, podName, cntrName}, nil)
	npuContainerTotalMemory = prometheus.NewDesc("container_npu_total_memory",
		"the npu total memory in container, unit is 'MB'", []string{npuID, namespace, podName, cntrName,
			modelName, npuUUID, npuPCIEInfo}, nil)
	npuContainerUsedMemory = prometheus.NewDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'", []string{npuID, namespace, podName, cntrName,
			modelName, npuUUID, npuPCIEInfo}, nil)
	npuContainerUtilization = prometheus.NewDesc("container_npu_utilization",
		"the npu ai core utilization in container, unit is '%'", []string{npuID, namespace, podName,
			cntrName, modelName, npuUUID, npuPCIEInfo}, nil)
	podAiCoreUtilizationRate = prometheus.NewDesc("vnpu_pod_aicore_utilization",
		"the vnpu aicore utilization rate, unit is '%'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, cntrName, isVirtual}, nil)
	podTotalMemory = prometheus.NewDesc("vnpu_pod_total_memory", "the vnpu total memory on pod, unit is 'KB'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, cntrName, isVirtual}, nil)
	podUsedMemory = prometheus.NewDesc("vnpu_pod_used_memory", "the vnpu used memory on pod, unit is 'KB'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, cntrName, isVirtual}, nil)
	npuChipInfoDescHbmEccEnableFlag = prometheus.NewDesc("npu_chip_info_hbm_ecc_enable_flag",
		"whether HBM ecc detection is enabled", cardLabel, nil)
	npuChipInfoDescHbmEccSingleBitErrorCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_single_bit_error_cnt",
		"HBM Single Bit Error Count", cardLabel, nil)
	npuChipInfoDescHbmEccDoubleBitErrorCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_double_bit_error_cnt",
		"HBM Double Bit Error Count", cardLabel, nil)
	npuChipInfoDescHbmEccTotalSingleBitErrorCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_total_single_bit_"+
		"error_cnt", "HBM Single Bit Aggregate Total Err Cnt", cardLabel, nil)
	npuChipInfoDescHbmEccTotalDoubleBitErrorCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_total_double_bit_"+
		"error_cnt", "HBM Double Bit Aggregate Total Err Cnt", cardLabel, nil)
	npuChipInfoDescHbmEccSingleBitIoslatedPagesCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_single_bit_"+
		"isolated_pages_cnt", "HBM Single Bit Isolated Pages Count", cardLabel, nil)
	npuChipInfoDescHbmEccDoubleBitIoslatedPagesCnt = prometheus.NewDesc("npu_chip_info_hbm_ecc_double_bit_"+
		"isolated_pages_cnt", "HBM Double Bit Isolated Pages Count", cardLabel, nil)
	npuChipInfoSioCrcTxErrCnt = prometheus.NewDesc("npu_chip_info_sio_crc_"+
		"tx_err_cnt", "sio transmitted error count between die", cardLabel, nil)
	npuChipInfoSioCrcRxErrCnt = prometheus.NewDesc("npu_chip_info_sio_crc_"+
		"rx_err_cnt", "sio received error count between die", cardLabel, nil)
	npuChipInfoHccsTxCnt0 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_0",
		"transmitted message count for hccs 0", cardLabel, nil)
	npuChipInfoHccsTxCnt1 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_1",
		"transmitted message count for hccs 1", cardLabel, nil)
	npuChipInfoHccsTxCnt2 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_2",
		"transmitted message count for hccs 2", cardLabel, nil)
	npuChipInfoHccsTxCnt3 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_3",
		"transmitted message count for hccs 3", cardLabel, nil)
	npuChipInfoHccsTxCnt4 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_4",
		"transmitted message count for hccs 4", cardLabel, nil)
	npuChipInfoHccsTxCnt5 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_5",
		"transmitted message count for hccs 5", cardLabel, nil)
	npuChipInfoHccsTxCnt6 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_6",
		"transmitted message count for hccs 6", cardLabel, nil)
	npuChipInfoHccsTxCnt7 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_tx_cnt_7",
		"transmitted message count for hccs 7", cardLabel, nil)
	npuChipInfoHccsRxCnt0 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_0",
		"received message count for hccs 0", cardLabel, nil)
	npuChipInfoHccsRxCnt1 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_1",
		"received message count for hccs 1", cardLabel, nil)
	npuChipInfoHccsRxCnt2 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_2",
		"received message count for hccs 2", cardLabel, nil)
	npuChipInfoHccsRxCnt3 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_3",
		"received message count for hccs 3", cardLabel, nil)
	npuChipInfoHccsRxCnt4 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_4",
		"received message count for hccs 4", cardLabel, nil)
	npuChipInfoHccsRxCnt5 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_5",
		"received message count for hccs 5", cardLabel, nil)
	npuChipInfoHccsRxCnt6 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_6",
		"received message count for hccs 6", cardLabel, nil)
	npuChipInfoHccsRxCnt7 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_rx_cnt_7",
		"received message count for hccs 7", cardLabel, nil)
	npuChipInfoCrcErrCnt0 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_0",
		"crc error count for hccs 0", cardLabel, nil)
	npuChipInfoCrcErrCnt1 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_1",
		"crc error count for hccs 1", cardLabel, nil)
	npuChipInfoCrcErrCnt2 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_2",
		"crc error count for hccs 2", cardLabel, nil)
	npuChipInfoCrcErrCnt3 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_3",
		"crc error count for hccs 3", cardLabel, nil)
	npuChipInfoCrcErrCnt4 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_4",
		"crc error count for hccs 4", cardLabel, nil)
	npuChipInfoCrcErrCnt5 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_5",
		"crc error count for hccs 5", cardLabel, nil)
	npuChipInfoCrcErrCnt6 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_6",
		"crc error count for hccs 6", cardLabel, nil)
	npuChipInfoCrcErrCnt7 = prometheus.NewDesc("npu_chip_info_hccs_statistic_info_crc_err_cnt_7",
		"crc error count for hccs 7", cardLabel, nil)
	npuChipInfoHccsBWProfilingTime = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_profiling_time",
		"Sampling interval for hccs bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTotalTx = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_total_tx",
		"total sent data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTotalRx = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_total_rx",
		"total received data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx0 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_0",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx1 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_1",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx2 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_2",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx3 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_3",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx4 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_4",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx5 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_5",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx6 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_6",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWTx7 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_tx_7",
		"single-link transmission data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx0 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_0",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx1 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_1",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx2 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_2",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx3 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_3",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx4 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_4",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx5 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_5",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx6 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_6",
		"single-link receive data bandwidth", cardLabel, nil)
	npuChipInfoHccsBWRx7 = prometheus.NewDesc("npu_chip_info_hccs_bandwidth_info_rx_7",
		"single-link receive data bandwidth", cardLabel, nil)

	npuContainerInfoInit sync.Once
	npuChipInfoInit      sync.Once
)

var netInfoMap sync.Map

const (
	cacheSize    = 128
	nameSpaceIdx = 0
	podNameIdx   = 1
	conNameIdx   = 2

	space             = " "
	newLine           = "\n"
	linkStatusPart    = 3
	trafficPart       = 4
	noTraffic         = 0.00
	decimalPlaces     = 2
	bitSize           = 64
	dcmiHccsMaxCounts = 8
)

type npuCollector struct {
	cache         *cache.ConcurrencyLRUCache
	devicesParser *container.DevicesParser
	updateTime    time.Duration
	cacheTime     time.Duration
}

// NewNpuCollector create an instance of prometheus Collector
func NewNpuCollector(cacheTime time.Duration, updateTime time.Duration,
	deviceParser *container.DevicesParser) *npuCollector {
	npuCollect := &npuCollector{
		cache:         cache.New(cacheSize),
		cacheTime:     cacheTime,
		updateTime:    updateTime,
		devicesParser: deviceParser,
	}
	return npuCollect
}

func setNetInfoWithMap(phyID int32, netInfo common.NpuNetInfo) {
	netInfoMap.Store(phyID, netInfo)
}

func getNetInfoFromMap(oldNetInfo map[int32]common.NpuNetInfo) map[int32]common.NpuNetInfo {
	newNetInfo := oldNetInfo
	netInfoMap.Range(func(key, value interface{}) bool {
		phyID, ok := key.(int32)
		if !ok {
			hwlog.RunLog.Warnf("failed to get phyID of netInfo from map, which is: %v", key)
			return true
		}
		netInfo, ok := value.(common.NpuNetInfo)
		if !ok {
			hwlog.RunLog.Warnf("failed to get value of netInfo from map, which is: %v", value)
			return true
		}
		newNetInfo[phyID] = netInfo
		return true
	})

	return newNetInfo
}

func startToGetNetInfo(ctx context.Context, dmgr devmanager.DeviceInterface, updateTime time.Duration) {
	cardNum, cards, err := dmgr.GetCardList()
	if err != nil || cardNum == 0 {
		hwlog.RunLog.Errorf("failed to get npu info, error is: %v", err)
		return
	}

	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num of card: %v failed: %v", cardID, err)
			continue
		}
		for i := int32(0); i < deviceNum; i++ {
			logicID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err != nil {
				hwlog.RunLog.Errorf("get logic ID of card: %v device:%v failed: %v", cardID, i, err)
				continue
			}

			phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get phy id when assemble net info: %v", err)
				continue
			}
			go assembleNPUNetInfo(ctx, phyID, dmgr, updateTime)
		}
	}
}

func getNPUInfo(dmgr devmanager.DeviceInterface) []HuaWeiNPUCard {
	npuList := make([]HuaWeiNPUCard, 0)
	cardNum, cards, err := dmgr.GetCardList()
	if err != nil || cardNum == 0 {
		hwlog.RunLog.Errorf("failed to get npu info, error is: %v", err)
		return npuList
	}

	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num of card %v failed: %v", cardID, err)
			continue
		}
		deviceList := make([]*HuaWeiAIChip, 0)
		for i := int32(0); i < deviceNum; i++ {
			var chipInfo *HuaWeiAIChip
			logicID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err != nil {
				hwlog.RunLog.Errorf("get logic ID of card %v device %v failed: %v", cardID, i, err)
				continue
			}
			chipInfo = assembleNPUInfo(cardID, logicID, dmgr)
			if chipInfo == nil {
				continue
			}
			if dmgr.GetDevType() != common.Ascend310P || chipInfo.VDevInfos == nil ||
				len(chipInfo.VDevInfos.VDevActivityInfo) == 0 {
				deviceList = append(deviceList, chipInfo)
				continue
			}
			deviceList = append(deviceList, getVNPUInfo(*chipInfo)...)
		}
		npuCard := HuaWeiNPUCard{
			CardID:     int(cardID),
			DeviceList: deviceList,
			Timestamp:  time.Now(),
		}
		npuList = append(npuList, npuCard)
	}
	return npuList
}

func assembleNPUNetInfo(ctx context.Context, phyID int32, dmgr devmanager.DeviceInterface, updateTime time.Duration) {
	if !dmgr.IsTrainingCard() {
		return
	}
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("received the stop signal, stop npu net info collect")
			return
		default:
			setNetInfoWithMap(phyID, networkPackInfo(phyID))
			time.Sleep(updateTime)
		}
	}
}

func assembleNPUInfo(cardID int32, logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	// check cardId, convert it to int type later
	if err != nil {
		hwlog.RunLog.Errorf("failed to get phy id when assemble npu info: %v", err)
		return nil
	}
	chipInfo := packChipInfo(logicID, dmgr)
	chipInfo.DeviceID = int(phyID)

	if dmgr.GetDevType() == common.Ascend310P {
		cardPower, err := dmgr.GetMcuPowerInfo(cardID)
		if err != nil {
			hwlog.RunLog.Error(err)
			cardPower = float32(common.RetError)
		}
		// Ascend310P use cardPower to replace chipPower
		chipInfo.Power = cardPower
		vDevInfos, err := dmgr.GetVirtualDeviceInfo(logicID)
		if err != nil {
			hwlog.RunLog.Warnf("failed to get virtual device info,logicID(%d),err: %v", logicID, err)
			chipInfo.VDevInfos = nil
			return chipInfo
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			chipInfo.VDevInfos = &common.VirtualDevInfo{}
			return chipInfo
		}
		chipInfo.VDevInfos = &vDevInfos
	}
	return chipInfo
}

func getVNPUInfo(chipInfo HuaWeiAIChip) []*HuaWeiAIChip {
	var aiChips []*HuaWeiAIChip
	if chipInfo.VDevInfos == nil {
		return aiChips
	}

	for _, activityVDev := range chipInfo.VDevInfos.VDevActivityInfo {
		vDevInfo := chipInfo
		vDevInfo.VDevActivityInfo = &activityVDev
		aiChips = append(aiChips, &vDevInfo)
	}
	return aiChips
}

// Start to collect npu base info, npu network info, container info
func Start(ctx context.Context, fn context.CancelFunc, n *npuCollector) {
	if n == nil {
		hwlog.RunLog.Warnf("Invalid param in function start")
		return
	}

	dmgr, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("new npu collector failed, error is %v", err)
		fn()
		return
	}

	defer func() {
		if err := dmgr.ShutDown(); err != nil {
			hwlog.RunLog.Error(err)
		}
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %v", err)
		}
		hwlog.RunLog.Info("npuCollector exit")
	}()

	if err := n.devicesParser.Init(); err != nil {
		hwlog.RunLog.Errorf("failed to init devices parser: %v", err)
	}
	defer n.devicesParser.Close()
	n.devicesParser.Timeout = n.updateTime
	hwlog.RunLog.Infof("Starting update cache every %d seconds", n.updateTime/time.Second)

	group := &sync.WaitGroup{}

	npuBaseInfoCollect(ctx, group, n, dmgr)
	npuNetworkInfoCollect(ctx, group, n, dmgr)
	containerInfoCollect(ctx, fn, group, n)

	group.Wait()

	return
}

func npuBaseInfoCollect(ctx context.Context, group *sync.WaitGroup, n *npuCollector, dmgr devmanager.DeviceInterface) {
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("received the stop signal,STOP npu base info collect")
				return
			default:
				npuInfo := getNPUInfo(dmgr)
				if err := n.cache.Set(npuListCacheKey, npuInfo, n.cacheTime); err != nil {
					hwlog.RunLog.Error(err)
				} else {
					hwlog.RunLog.Infof(updateCachePattern, npuListCacheKey)
				}
				if _, ok := <-ticker.C; !ok {
					hwlog.RunLog.Errorf(tickerFailedPattern, npuListCacheKey)
					return
				}
			}
		}
	}()
}

func npuNetworkInfoCollect(ctx context.Context, group *sync.WaitGroup, n *npuCollector,
	dmgr devmanager.DeviceInterface) {
	group.Add(1)
	netInfo := make(map[int32]common.NpuNetInfo, initSize)
	startToGetNetInfo(ctx, dmgr, n.updateTime)

	collectNetworkInfo := func() {
		obj, err := n.cache.Get(npuNetworkCacheKey)
		if err != nil {
			hwlog.RunLog.Warnf("get info of %s failed: %v, so use initial net info", npuNetworkCacheKey, err)
		} else {
			if oldNetWorkInfo, ok := obj.(map[int32]common.NpuNetInfo); ok {
				netInfo = oldNetWorkInfo
			} else {
				hwlog.RunLog.Warn("format of net info in cache is not right")
			}
		}
		// get current net info from map to update cache
		newNetInfo := getNetInfoFromMap(netInfo)
		if err := n.cache.Set(npuNetworkCacheKey, newNetInfo, n.cacheTime); err != nil {
			hwlog.RunLog.Error(err)
		} else {
			hwlog.RunLog.Infof(updateCachePattern, npuNetworkCacheKey)
		}
	}

	go func() {
		defer group.Done()
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("received the stop signal,STOP npu network info collect")
				return
			default:
				collectNetworkInfo()
				if _, ok := <-ticker.C; !ok {
					hwlog.RunLog.Errorf(tickerFailedPattern, npuNetworkCacheKey)
					return
				}

			}
		}
	}()
}

func containerInfoCollect(ctx context.Context, fn context.CancelFunc, group *sync.WaitGroup, n *npuCollector) {
	group.Add(1)

	go func() {
		defer group.Done()
		retryCount := 0
		collectContainerInfo := func() {
			hwlog.RunLog.Info("start to collect container info")
			n.devicesParser.FetchAndParse(nil)
			select {
			case result := <-n.devicesParser.RecvResult():
				if err := n.cache.Set(containersDevicesCacheKey, result, n.cacheTime); err != nil {
					hwlog.RunLog.Error(err)
				}
				hwlog.RunLog.Infof(updateCachePattern, containersDevicesCacheKey)
				retryCount = 0
			case err := <-n.devicesParser.RecvErr():
				hwlog.RunLog.Errorf("received error from device parser: %v", err)
				if strings.Contains(err.Error(), "connection refused") {
					retryCount++
					if retryCount == connectRefusedMaxRetry {
						hwlog.RunLog.Error("connection refused, task shutdown")
						fn()
					}
				}
			}
		}
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("received the stop signal,STOP container info collect")
				return
			default:
				collectContainerInfo()
				if _, ok := <-ticker.C; !ok {
					hwlog.RunLog.Errorf(tickerFailedPattern, containersDevicesCacheKey)
					return
				}
			}
		}
	}()
}

func describeBaseChipInfo(ch chan<- *prometheus.Desc) {
	ch <- versionInfoDesc
	ch <- machineInfoNPUDesc
	ch <- npuChipInfoDescUtil
	ch <- npuChipInfoDescOverUtil
	ch <- npuChipInfoDescVectorUtil
	ch <- npuChipInfoDescTemp
	ch <- npuChipInfoDescPower
	ch <- npuChipInfoDescVoltage
	ch <- npuChipInfoDescHealthStatus
	ch <- npuChipInfoDescHbmUsedMemory
	ch <- npuChipInfoDescHbmTotalMemory
	ch <- npuChipInfoDescUsedMemory
	ch <- npuChipInfoDescTotalMemory
	ch <- npuChipInfoDescNpuName
	describeChipErrCodesInfo(ch)
}

func describeChipErrCodesInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipInfoDescErrorCode
	ch <- npuChipInfoDescErrorCode1
	ch <- npuChipInfoDescErrorCode2
	ch <- npuChipInfoDescErrorCode3
	ch <- npuChipInfoDescErrorCode4
	ch <- npuChipInfoDescErrorCode5
	ch <- npuChipInfoDescErrorCode6
	ch <- npuChipInfoDescErrorCode7
	ch <- npuChipInfoDescErrorCode8
	ch <- npuChipInfoDescErrorCode9
}

func describeHBMEccInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipInfoDescHbmEccEnableFlag
	ch <- npuChipInfoDescHbmEccSingleBitErrorCnt
	ch <- npuChipInfoDescHbmEccDoubleBitErrorCnt
	ch <- npuChipInfoDescHbmEccTotalSingleBitErrorCnt
	ch <- npuChipInfoDescHbmEccTotalDoubleBitErrorCnt
	ch <- npuChipInfoDescHbmEccSingleBitIoslatedPagesCnt
	ch <- npuChipInfoDescHbmEccDoubleBitIoslatedPagesCnt
}

func describeHccsInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipInfoHccsTxCnt0
	ch <- npuChipInfoHccsTxCnt1
	ch <- npuChipInfoHccsTxCnt2
	ch <- npuChipInfoHccsTxCnt3
	ch <- npuChipInfoHccsTxCnt4
	ch <- npuChipInfoHccsTxCnt5
	ch <- npuChipInfoHccsTxCnt6
	ch <- npuChipInfoHccsTxCnt7
	ch <- npuChipInfoHccsRxCnt0
	ch <- npuChipInfoHccsRxCnt1
	ch <- npuChipInfoHccsRxCnt2
	ch <- npuChipInfoHccsRxCnt3
	ch <- npuChipInfoHccsRxCnt4
	ch <- npuChipInfoHccsRxCnt5
	ch <- npuChipInfoHccsRxCnt6
	ch <- npuChipInfoHccsRxCnt7
	ch <- npuChipInfoCrcErrCnt0
	ch <- npuChipInfoCrcErrCnt1
	ch <- npuChipInfoCrcErrCnt2
	ch <- npuChipInfoCrcErrCnt3
	ch <- npuChipInfoCrcErrCnt4
	ch <- npuChipInfoCrcErrCnt5
	ch <- npuChipInfoCrcErrCnt6
	ch <- npuChipInfoCrcErrCnt7
}

func describeOpticalInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipOpticalState
	ch <- npuChipOpticalTxPower0
	ch <- npuChipOpticalTxPower1
	ch <- npuChipOpticalTxPower2
	ch <- npuChipOpticalTxPower3
	ch <- npuChipOpticalRxPower0
	ch <- npuChipOpticalRxPower1
	ch <- npuChipOpticalRxPower2
	ch <- npuChipOpticalRxPower3
	ch <- npuChipOpticalVcc
	ch <- npuChipOpticalTemp
}

func describeRoCEInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipInfoDescNetworkStatus
	ch <- npuChipInfoDescBandwidthTx
	ch <- npuChipInfoDescBandwidthRx
	ch <- npuChipInfoDescLinkStatus
	ch <- npuChipLinkSpeed
	ch <- npuChipLinkUpNum
	ch <- npuChipMacRxPauseNum
	ch <- npuChipMacTxPauseNum
	ch <- npuChipMacRxPfcPktNum
	ch <- npuChipMacTxPfcPktNum
	ch <- npuChipMacRxBadPktNum
	ch <- npuChipMacTxBadPktNum
	ch <- npuChipRoceRxAllPktNum
	ch <- npuChipRoceTxAllPktNum
	ch <- npuChipRoceRxErrPktNum
	ch <- npuChipRoceTxErrPktNum
	ch <- npuChipRoceRxCnpPktNum
	ch <- npuChipRoceTxCnpPktNum
	ch <- npuChipRoceNewPktRtyNum
	ch <- npuChipMacTxBadOctNum
	ch <- npuChipMacRxBadOctNum
	ch <- npuChipRoceUnexpectedAcktNum
	ch <- npuChipRoceOutOfOrderNum
	ch <- npuChipRoceVerificationErrNum
	ch <- npuChipRoceQpStatusErrNum
}

// Describe implements prometheus.Collector
func (n *npuCollector) Describe(ch chan<- *prometheus.Desc) {
	if ch == nil {
		hwlog.RunLog.Warnf("Invalid param in function Describe")
		return
	}
	describeBaseChipInfo(ch)
	describeOpticalInfo(ch)
	describeRoCEInfo(ch)
	describeHBMEccInfo(ch)
	describeHccsInfo(ch)
	ch <- npuContainerInfo
	ch <- npuContainerTotalMemory
	ch <- npuContainerUsedMemory
	ch <- npuContainerUtilization
	ch <- npuChipInfoDescDevProcessInfo
	ch <- npuChipInfoDescAICoreFreqInfo
	ch <- podAiCoreUtilizationRate
	ch <- podTotalMemory
	ch <- podUsedMemory
	ch <- npuChipInfoDescRxPBW
	ch <- npuChipInfoDescTxPBW
	ch <- npuChipInfoDescRxNpBW
	ch <- npuChipInfoDescTxNpBW
	ch <- npuChipInfoDescRxCplBW
	ch <- npuChipInfoDescTxCplBW
	ch <- npuChipInfoDescRxECNNum
	ch <- npuChipInfoDescRxFCSNum
	ch <- npuChipInfoSioCrcTxErrCnt
	ch <- npuChipInfoSioCrcRxErrCnt
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		hwlog.RunLog.Warnf("Invalid param in function Collect")
		return
	}
	npuList := getNPUInfoInCache(ch, n)
	networkInfoMap := getNetworkInfoInCache(ch, n)
	containerMap := getContainerNPUInfo(ch, n)
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1,
		[]string{versions.BuildVersion}...)
	var totalCount = 0
	for _, card := range npuList {
		deviceCount := len(card.DeviceList)
		if deviceCount <= 0 {
			continue
		}
		totalCount += deviceCount
		for _, chip := range card.DeviceList {
			deviceID := chip.DeviceID
			if devNetWorkInfo, ok := networkInfoMap[int32(deviceID)]; ok {
				chip.NetInfo = &devNetWorkInfo
			} else {
				hwlog.RunLog.Warn("no network information at the moment, so use initial info")
				chip.NetInfo = &common.NpuNetInfo{}
			}

			if chip.VDevActivityInfo != nil && chip.VDevActivityInfo.IsVirtualDev {
				deviceID = int(chip.VDevActivityInfo.VDevID)
			}
			devInfo, ok := containerMap[deviceID]
			if !ok {
				devInfo = container.DevicesInfo{}
			}
			updateNPUCommonInfo(ch, &card, chip, devInfo)
			updateNPUMemoryInfo(ch, &card, chip, devInfo)
			updateNPUHBMInfo(ch, &card, chip, devInfo)
			updateNPUNetworkInfo(ch, &card, chip, devInfo)
			updateProcessInfo(ch, &card, chip, devInfo)
			updateContainerInfo(ch, &card, chip, devInfo)
			updatePodVNPUInfo(ch, &card, chip, devInfo)
			updateHBMECCInfo(ch, &card, chip, devInfo)
			updateSioInfo(ch, &card, chip, devInfo)
			updateHccsInfo(ch, &card, chip, devInfo)
		}
	}

	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(totalCount))
}

func getNPUInfoInCache(ch chan<- prometheus.Metric, n *npuCollector) []HuaWeiNPUCard {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return nil
	}
	obj, err := n.cache.Get(npuListCacheKey)
	npuChipInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Debugf("no cache, start to get npulist and rebuild cache")
			devManager, err := devmanager.GetDeviceManager()
			if err != nil {
				hwlog.RunLog.Debugf("get device manager failed, error is: %v ", err)
				return
			}
			npuInfo := getNPUInfo(devManager)
			if err = n.cache.Set(npuListCacheKey, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Errorf("no cache for prometheus, try to build cache failed, error is: %v", err)
				return
			}
			hwlog.RunLog.Debugf("rebuild cache successfully")
			obj = npuInfo
		}
	})
	npuList, ok := obj.([]HuaWeiNPUCard)
	if !ok {
		hwlog.RunLog.Error("Error npu info cache and convert failed")
		n.cache.Delete(npuListCacheKey)
		return nil
	}

	return npuList
}

func getNetworkInfoInCache(ch chan<- prometheus.Metric, n *npuCollector) map[int32]common.NpuNetInfo {
	res := make(map[int32]common.NpuNetInfo, initSize)
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return res
	}
	obj, err := n.cache.Get(npuNetworkCacheKey)
	if err != nil {
		hwlog.RunLog.Warn("npu network info not found in cache, please wait for the cache to be rebuilt")
		return res
	}
	networkInfoList, ok := obj.(map[int32]common.NpuNetInfo)
	if !ok {
		hwlog.RunLog.Error("Error npu network info cache and convert failed")
		n.cache.Delete(npuNetworkCacheKey)
		return res
	}

	return networkInfoList
}

func getContainerNPUInfo(ch chan<- prometheus.Metric, n *npuCollector) map[int]container.DevicesInfo {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return nil
	}
	obj, err := n.cache.Get(containersDevicesCacheKey)
	// only run once to prevent wait when container info get failed
	npuContainerInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Warn("containers' devices info not found in cache, rebuilding")
			resultChan := make(chan container.DevicesInfos, 1)
			n.devicesParser.FetchAndParse(resultChan)
			select {
			case obj = <-resultChan:
			case <-time.After(time.Second):
				hwlog.RunLog.Warn("rebuild container info cache timeout")
				return
			}
			hwlog.RunLog.Warn("rebuild cache successfully")
		}
	})
	cntNpuInfos, ok := obj.(container.DevicesInfos)
	if !ok {
		hwlog.RunLog.Error("Error container npu info cache and convert failed")
		n.cache.Delete(containersDevicesCacheKey)
		return nil
	}
	res := make(map[int]container.DevicesInfo, initSize)
	for _, v := range cntNpuInfos {
		for _, deviceID := range v.Devices {
			res[deviceID] = v
		}
	}
	return res
}

func validate(ch chan<- prometheus.Metric, objs ...interface{}) bool {
	if ch == nil {
		return false
	}
	for _, v := range objs {
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr {
			return false
		}
		if val.IsNil() {
			return false
		}
	}
	return true
}

func validateObj(objs ...interface{}) bool {
	for _, v := range objs {
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr {
			return false
		}
		if val.IsNil() {
			return false
		}
	}
	return true
}

func validateNum(num float64) bool {
	if num == -1 || num == math.MaxUint32 || float32(num) == math.MaxUint32 {
		return false
	}

	return true
}

func getContainerNameArray(devInfo container.DevicesInfo) []string {
	if devInfo.Name == "" {
		return nil
	}

	return strings.Split(devInfo.Name, "_")
}

func updateNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	memoryInfo := common.DeepCopyMemoryInfo(chip.Meminf)
	if !validate(ch, npu, chip, memoryInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateNPUMemoryInfo")
		return
	}

	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory,
		prometheus.GaugeValue, float64(memoryInfo.MemorySize-memoryInfo.MemoryAvailable),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue,
			float64(memoryInfo.MemorySize), collectCardLabelValue(chip, namespaceValue,
				podNameValue, containerName)...))
}

func updateNPUHBMInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	hbmInfo := common.DeepCopyHbmInfo(chip.HbmInfo.HbmInfo)
	if !validate(ch, npu, chip, hbmInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateNPUHBMInfo")
		return
	}

	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmUsedMemory, prometheus.GaugeValue, float64(hbmInfo.Usage),
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(hbmInfo.MemorySize), collectCardLabelValue(chip, namespaceValue,
				podNameValue, containerName)...))
}

func updateStatInfoOfMac(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateStatInfoOfMac")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	statInfo := common.DeepCopyStatInfo(chip.NetInfo.StatInfo)
	if !validate(ch, npu, chip, statInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateStatInfoOfMac")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxPauseNum, prometheus.GaugeValue, statInfo.MacRxPauseNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxPauseNum, prometheus.GaugeValue, statInfo.MacTxPauseNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxPfcPktNum, prometheus.GaugeValue, statInfo.MacRxPfcPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxPfcPktNum, prometheus.GaugeValue, statInfo.MacTxPfcPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxBadPktNum, prometheus.GaugeValue, statInfo.MacRxBadPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxBadPktNum, prometheus.GaugeValue, statInfo.MacTxBadPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxBadOctNum, prometheus.GaugeValue, statInfo.MacTxBadOctNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxBadOctNum, prometheus.GaugeValue, statInfo.MacRxBadOctNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescRxFCSNum, prometheus.GaugeValue, statInfo.MacRXFcsErrPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
}

func updateStatInfoOfRoCE(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateStatInfoOfRoCE")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	statInfo := common.DeepCopyStatInfo(chip.NetInfo.StatInfo)
	if !validate(ch, npu, chip, statInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateStatInfoOfRoCE")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxAllPktNum, prometheus.GaugeValue, statInfo.RoceRxAllPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxAllPktNum, prometheus.GaugeValue, statInfo.RoceTxAllPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxErrPktNum, prometheus.GaugeValue, statInfo.RoceRxErrPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxErrPktNum, prometheus.GaugeValue, statInfo.RoceTxErrPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxCnpPktNum, prometheus.GaugeValue, statInfo.RoceRxCnpPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxCnpPktNum, prometheus.GaugeValue, statInfo.RoceTxCnpPktNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceNewPktRtyNum, prometheus.GaugeValue, statInfo.RoceNewPktRtyNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceUnexpectedAcktNum, prometheus.GaugeValue, statInfo.
			RoceUnexpectedAckNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceOutOfOrderNum, prometheus.GaugeValue, statInfo.RoceOutOfOrderNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceVerificationErrNum, prometheus.GaugeValue, statInfo.
			RoceVerificationErrNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceQpStatusErrNum, prometheus.GaugeValue, statInfo.RoceQpStatusErrNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescRxECNNum, prometheus.GaugeValue, statInfo.RoceEcnDBNum,
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
}

func updateOpticalInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateOpticalInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	opticalInfo := common.DeepCopyOpticalInfo(chip.NetInfo.OpticalInfo)
	if !validate(ch, npu, chip, opticalInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateOpticalInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(opticalInfo.OpticalState) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalState, prometheus.GaugeValue, opticalInfo.OpticalState,
				collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalVcc) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalVcc, prometheus.GaugeValue, opticalInfo.OpticalVcc,
				collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalTemp) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalTemp, prometheus.GaugeValue, opticalInfo.OpticalTemp,
				collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	updateOpticalTxPower(ch, npu, chip, cNameArray)
	updateOpticalRxPower(ch, npu, chip, cNameArray)
}

func updateOpticalTxPower(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateOpticalTxPower")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	opticalInfo := common.DeepCopyOpticalInfo(chip.NetInfo.OpticalInfo)
	if !validate(ch, npu, chip, opticalInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateOpticalTxPower")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(opticalInfo.OpticalTxPower0) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalTxPower0, prometheus.GaugeValue, opticalInfo.
				OpticalTxPower0, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalTxPower1) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalTxPower1, prometheus.GaugeValue, opticalInfo.
				OpticalTxPower1, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalTxPower2) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalTxPower2, prometheus.GaugeValue, opticalInfo.
				OpticalTxPower2, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalTxPower3) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalTxPower3, prometheus.GaugeValue, opticalInfo.
				OpticalTxPower3, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateOpticalRxPower(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateOpticalRxPower")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	opticalInfo := common.DeepCopyOpticalInfo(chip.NetInfo.OpticalInfo)
	if !validate(ch, npu, chip, opticalInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateOpticalRxPower")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(opticalInfo.OpticalRxPower0) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalRxPower0, prometheus.GaugeValue, opticalInfo.
				OpticalRxPower0, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalRxPower1) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalRxPower1, prometheus.GaugeValue, opticalInfo.
				OpticalRxPower1, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalRxPower2) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalRxPower2, prometheus.GaugeValue, opticalInfo.
				OpticalRxPower2, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(opticalInfo.OpticalRxPower3) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipOpticalRxPower3, prometheus.GaugeValue, opticalInfo.
				OpticalRxPower3, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateBandwidthInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cNameArray []string) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateBandwidthInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	bandwidthInfo := common.DeepCopyBandwidthInfo(chip.NetInfo.BandwidthInfo)
	if !validate(ch, npu, chip, bandwidthInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateBandwidthInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(bandwidthInfo.TxValue) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescBandwidthTx, prometheus.GaugeValue, bandwidthInfo.
				TxValue, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}

	if validateNum(bandwidthInfo.RxValue) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescBandwidthRx, prometheus.GaugeValue, bandwidthInfo.
				RxValue, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateNPUNetworkInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Warnf("Invalid param in function updateNPUNetworkInfo")
		return
	}
	cNameArray := getContainerNameArray(devInfo)
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(float64(getHealthCode(chip.NetHealthStatus))) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescNetworkStatus, prometheus.GaugeValue,
				float64(getHealthCode(chip.NetHealthStatus)), collectCardLabelValue(chip, namespaceValue, podNameValue,
					containerName)...))
	}

	updateStatInfoOfMac(ch, npu, chip, cNameArray)
	updateStatInfoOfRoCE(ch, npu, chip, cNameArray)
	updateOpticalInfo(ch, npu, chip, cNameArray)
	updateBandwidthInfo(ch, npu, chip, cNameArray)
	updateLinkSpeedInfo(ch, npu, chip, devInfo)
	updateLinkStatInfo(ch, npu, chip, devInfo)
	updateLinkStatusInfo(ch, npu, chip, devInfo)
}

func updateLinkSpeedInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateLinkSpeedInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	linkSpeedInfo := common.DeepCopyLinkSpeedInfo(chip.NetInfo.LinkSpeedInfo)
	if !validate(ch, npu, chip, linkSpeedInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateLinkSpeedInfo")
		return
	}
	cNameArray := getContainerNameArray(devInfo)
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(linkSpeedInfo.Speed) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipLinkSpeed, prometheus.GaugeValue, linkSpeedInfo.Speed,
				collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateLinkStatInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateLinkStatusInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	linkStatInfo := common.DeepCopyLinkStatInfo(chip.NetInfo.LinkStatInfo)
	if !validate(ch, npu, chip, linkStatInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateLinkStatusInfo")
		return
	}
	cNameArray := getContainerNameArray(devInfo)
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(linkStatInfo.LinkUPNum) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipLinkUpNum, prometheus.GaugeValue, linkStatInfo.LinkUPNum,
				collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateLinkStatusInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if chip.NetInfo == nil {
		hwlog.RunLog.Error("NetInfo is nil in function updateLinkStatInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	linkStatusInfo := common.DeepCopyLinkStatusInfo(chip.NetInfo.LinkStatusInfo)
	if !validate(ch, npu, chip, linkStatusInfo) {
		hwlog.RunLog.Warn("Invalid param in function updateLinkStatInfo")
		return
	}
	cNameArray := getContainerNameArray(devInfo)
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(cNameArray)
	if validateNum(float64(hccn.GetLinkStatusCode(linkStatusInfo.LinkState))) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescLinkStatus,
			prometheus.GaugeValue, float64(hccn.GetLinkStatusCode(linkStatusInfo.LinkState)),
			collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	}
}

func updateContainerInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Warnf("Invalid param in function updateContainerInfo")
		return
	}

	containerName := getContainerNameArray(devInfo)
	if len(containerName) != containerNameLen {
		return
	}
	ch <- prometheus.MustNewConstMetric(npuContainerInfo, prometheus.GaugeValue, 1,
		[]string{devInfo.ID, strings.Join(containerName, "_"), strconv.Itoa(chip.DeviceID),
			common.GetNpuName(chipInfo), chip.VDieID, chip.PCIeBusInfo, containerName[nameSpaceIdx],
			containerName[podNameIdx], containerName[conNameIdx]}...)
	vDevActivityInfo := common.DeepCopyVDevActivityInfo(chip.VDevActivityInfo)
	if vDevActivityInfo != nil && common.IsValidVDevID(vDevActivityInfo.VDevID) {
		return
	}
	if validateNum(float64(chip.Utilization)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUtilization,
			prometheus.GaugeValue, float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx],
				common.GetNpuName(chipInfo), chip.VDieID,
				chip.PCIeBusInfo}...))
	}
	updateContainerNPUMemoryInfo(ch, npu, chip, containerName)
}

func updatePodVNPUInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	if chipInfo != nil && !strings.Contains(chipInfo.Name, "310P") {
		hwlog.RunLog.Debug("only 310P supports vNPU information query")
		return
	}

	vDevActivityInfo := common.DeepCopyVDevActivityInfo(chip.VDevActivityInfo)
	if !validate(ch, npu, chip, vDevActivityInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updatePodVNPUInfo")
		return
	}

	if !common.IsValidVDevID(vDevActivityInfo.VDevID) {
		return
	}
	containerName := getContainerNameArray(devInfo)
	if len(containerName) != containerNameLen {
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podAiCoreUtilizationRate, prometheus.GaugeValue,
			float64(vDevActivityInfo.VDevAiCoreRate), getPodDisplayInfo(chip, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podTotalMemory, prometheus.GaugeValue,
			float64(vDevActivityInfo.VDevTotalMem), getPodDisplayInfo(chip, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podUsedMemory, prometheus.GaugeValue,
			float64(vDevActivityInfo.VDevUsedMem), getPodDisplayInfo(chip, containerName)...))
}

func updateContainerNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	containerName []string) {
	if len(containerName) != containerNameLen {
		hwlog.RunLog.Errorf("container name length %v is not %v", len(containerName), containerNameLen)
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	if chipInfo != nil && strings.Contains(chipInfo.Name, common.Chip910) {
		hbmInfo := common.DeepCopyHbmInfo(chip.HbmInfo.HbmInfo)
		if !validate(ch, npu, chip, hbmInfo) {
			hwlog.RunLog.Error("Invalid hbm info param in function updateContainerNPUMemoryInfo")
			return
		}
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerTotalMemory, prometheus.GaugeValue,
				float64(hbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
					containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx],
					common.GetNpuName(chipInfo), chip.VDieID, chip.PCIeBusInfo}...))
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerUsedMemory, prometheus.GaugeValue, float64(hbmInfo.Usage),
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
					containerName[podNameIdx], containerName[conNameIdx],
					common.GetNpuName(chipInfo), chip.VDieID, chip.PCIeBusInfo}...))
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	memoryInfo := common.DeepCopyMemoryInfo(chip.Meminf)
	if !validate(ch, npu, chip, memoryInfo) {
		hwlog.RunLog.Error("Invalid mem info param in function updateContainerNPUMemoryInfo")
		return
	}

	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerTotalMemory,
		prometheus.GaugeValue, float64(memoryInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx],
			common.GetNpuName(chipInfo), chip.VDieID,
			chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUsedMemory,
		prometheus.GaugeValue, float64(memoryInfo.MemorySize-memoryInfo.MemoryAvailable),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
			containerName[podNameIdx], containerName[conNameIdx], common.GetNpuName(chipInfo),
			chip.VDieID, chip.PCIeBusInfo}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Warnf("Invalid param in function updateNpuCommonInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescNpuName,
		prometheus.GaugeValue, 1, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))

	updateChipBaseInfo(ch, npu, chip, devInfo)
	updatePcieBwInfo(ch, npu, chip, collectCardLabelValue(chip, namespaceValue, podNameValue, containerName))
}

func updateChipBaseInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	labelValue := collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)
	if validateNum(float64(chip.Utilization)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUtil,
			prometheus.GaugeValue, float64(chip.Utilization), labelValue...))
	}
	if validateNum(float64(chip.OverallUtilization)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescOverUtil,
			prometheus.GaugeValue, float64(chip.OverallUtilization), labelValue...))
	}
	if validateNum(float64(chip.VectorUtilization)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescVectorUtil,
			prometheus.GaugeValue, float64(chip.VectorUtilization), labelValue...))
	}
	if validateNum(float64(chip.Temperature)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescTemp,
			prometheus.GaugeValue, float64(chip.Temperature), labelValue...))
	}
	if validateNum(float64(chip.Power)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescPower,
			prometheus.GaugeValue, float64(chip.Power), labelValue...))
	}
	if validateNum(float64(chip.Voltage)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescVoltage,
			prometheus.GaugeValue, float64(chip.Voltage), labelValue...))
	}
	if validateNum(float64(getHealthCode(chip.HealthStatus))) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescHealthStatus, prometheus.GaugeValue,
				float64(getHealthCode(chip.HealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base),
					common.GetNpuName(chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo, namespaceValue, podNameValue,
					containerName}...))
	}
	if validateNum(float64(chip.AICoreCurrentFreq)) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescAICoreFreqInfo,
			prometheus.GaugeValue, float64(chip.AICoreCurrentFreq), labelValue...))
	}
	updateErrorCodesInfo(ch, npu, chip, labelValue)
}

// updateErrorCodesInfo update error code info, max 10
func updateErrorCodesInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, labelValue []string) {
	errCodesDesc := []*prometheus.Desc{
		npuChipInfoDescErrorCode,
		npuChipInfoDescErrorCode1, npuChipInfoDescErrorCode2, npuChipInfoDescErrorCode3, npuChipInfoDescErrorCode4,
		npuChipInfoDescErrorCode5, npuChipInfoDescErrorCode6, npuChipInfoDescErrorCode7, npuChipInfoDescErrorCode8,
		npuChipInfoDescErrorCode9,
	}
	if len(chip.ErrorCodes) > common.MaxErrorCodeLen {
		hwlog.RunLog.Warnf("Error code number is larger than %v, only the first %v will be reported, "+
			"all errorCode is: %v", common.MaxErrorCodeLen, common.MaxErrorCodeLen, chip.ErrorCodes)
	}
	for i := 0; i < len(chip.ErrorCodes) && i < len(errCodesDesc); i++ {
		value := float64(chip.ErrorCodes[i])
		if !validateNum(value) {
			continue
		}
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(errCodesDesc[i],
			prometheus.GaugeValue, value, labelValue...))
	}
}

func updateProcessInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	devProcessInfo := common.DeepCopyDevProcessInfo(chip.DevProcessInfo)
	if !validate(ch, npu, chip, devProcessInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateProcessInfo")
		return
	}
	containerName := ""
	containerID := ""
	namespaceValue := ""
	podNameValue := ""
	cNameArray := getContainerNameArray(devInfo)
	if len(cNameArray) == containerNameLen {
		namespaceValue = cNameArray[nameSpaceIdx]
		podNameValue = cNameArray[podNameIdx]
		containerName = strings.Join(cNameArray, "_")
		containerID = devInfo.ID
	}
	if devProcessInfo.ProcNum == 0 {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, 0,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(chipInfo),
					chip.VDieID, "", containerID, containerName, chip.PCIeBusInfo, namespaceValue, podNameValue}...))
		return
	}
	for i := int32(0); i < devProcessInfo.ProcNum; i++ {
		procInfo := devProcessInfo.DevProcArray[i]
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, procInfo.MemUsage,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(chipInfo),
					chip.VDieID, strconv.FormatInt(int64(procInfo.Pid), base), containerID, containerName,
					chip.PCIeBusInfo, namespaceValue, podNameValue}...))
	}
}

func updateSioInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	devSioInfo := common.DeepCopySioCrcErrStatisticInfo(chip.SioInfo)
	if !validate(ch, npu, chip, devSioInfo) {
		hwlog.RunLog.Warn("Invalid param in function updateSioInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoSioCrcTxErrCnt, prometheus.GaugeValue,
			float64(devSioInfo.TxErrCnt), collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoSioCrcRxErrCnt, prometheus.GaugeValue,
			float64(devSioInfo.RxErrCnt), collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
}

func updateHccsInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if chip.ChipIfo == nil || chip.BoardInfo == nil {
		hwlog.RunLog.Warn("Invalid param in function updateHccsInfo")
		return
	}
	devType := common.GetDevType(chip.ChipIfo.Name, chip.BoardInfo.BoardId)
	if devType != common.Ascend910B && devType != common.Ascend910A3 {
		return
	}
	var hccsBeginIndex int
	if devType == common.Ascend910B || common.IsA900A3SuperPod(chip.MainBoardId) {
		// 910B or A900A3SuperPod begin at 1st bit
		hccsBeginIndex = 1
	} else if common.IsA9000A3SuperPod(chip.MainBoardId) {
		// A9000A3SuperPod begin at 2nd bit
		hccsBeginIndex = 2
	}
	updateHccsStatisticInfo(ch, npu, chip, devInfo, hccsBeginIndex)
	updateHccsBandwidthInfo(ch, npu, chip, devInfo, hccsBeginIndex)
}

func doUpdateHccsMetric(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, value interface{},
	cardLabel []string, desc *prometheus.Desc) {
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
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, finalValue, cardLabel...))
}

func updateHccsStatisticInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo, hccsBeginIndex int) {
	hccsStatisticInfo := common.DeepCopyHccsStatisticInfo(chip.HccsStatisticInfo)
	if !validate(ch, npu, chip, hccsStatisticInfo) {
		hwlog.RunLog.Warn("Invalid param in function updateHccsStatisticInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	cardLabel := collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)
	hccsStatisticTxInfo := []*prometheus.Desc{
		npuChipInfoHccsTxCnt0, npuChipInfoHccsTxCnt1, npuChipInfoHccsTxCnt2, npuChipInfoHccsTxCnt3,
		npuChipInfoHccsTxCnt4, npuChipInfoHccsTxCnt5, npuChipInfoHccsTxCnt6, npuChipInfoHccsTxCnt7,
	}
	hccsStatisticRxInfo := []*prometheus.Desc{
		npuChipInfoHccsRxCnt0, npuChipInfoHccsRxCnt1, npuChipInfoHccsRxCnt2, npuChipInfoHccsRxCnt3,
		npuChipInfoHccsRxCnt4, npuChipInfoHccsRxCnt5, npuChipInfoHccsRxCnt6, npuChipInfoHccsRxCnt7,
	}
	hccsStatisticErrCntInfo := []*prometheus.Desc{
		npuChipInfoCrcErrCnt0, npuChipInfoCrcErrCnt1, npuChipInfoCrcErrCnt2, npuChipInfoCrcErrCnt3,
		npuChipInfoCrcErrCnt4, npuChipInfoCrcErrCnt5, npuChipInfoCrcErrCnt6, npuChipInfoCrcErrCnt7,
	}
	if hccsBeginIndex < 0 {
		hccsBeginIndex = 0
	}
	for i := hccsBeginIndex; i < dcmiHccsMaxCounts && i < len(hccsStatisticTxInfo) &&
		i < len(hccsStatisticRxInfo) && i < len(hccsStatisticErrCntInfo); i++ {
		doUpdateHccsMetric(ch, npu, hccsStatisticInfo.TxCnt[i], cardLabel, hccsStatisticTxInfo[i])
		doUpdateHccsMetric(ch, npu, hccsStatisticInfo.RxCnt[i], cardLabel, hccsStatisticRxInfo[i])
		doUpdateHccsMetric(ch, npu, hccsStatisticInfo.CrcErrCnt[i], cardLabel, hccsStatisticErrCntInfo[i])
	}
}
func updateHccsBandwidthInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo, hccsBeginIndex int) {
	hccsBandwidthInfo := common.DeepCopyHccsBandwidthInfo(chip.HccsBandwidthInfo)
	if !validate(ch, npu, chip, hccsBandwidthInfo) {
		hwlog.RunLog.Warn("Invalid param in function updateHccsStatisticInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	cardLabel := collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)
	hccsBWTxDescs := []*prometheus.Desc{
		npuChipInfoHccsBWTx0, npuChipInfoHccsBWTx1, npuChipInfoHccsBWTx2, npuChipInfoHccsBWTx3,
		npuChipInfoHccsBWTx4, npuChipInfoHccsBWTx5, npuChipInfoHccsBWTx6, npuChipInfoHccsBWTx7,
	}
	hccsBWRxDescs := []*prometheus.Desc{
		npuChipInfoHccsBWRx0, npuChipInfoHccsBWRx1, npuChipInfoHccsBWRx2, npuChipInfoHccsBWRx3,
		npuChipInfoHccsBWRx4, npuChipInfoHccsBWRx5, npuChipInfoHccsBWRx6, npuChipInfoHccsBWRx7,
	}
	if hccsBeginIndex < 0 {
		hccsBeginIndex = 0
	}
	doUpdateHccsMetric(ch, npu, hccsBandwidthInfo.ProfilingTime, cardLabel, npuChipInfoHccsBWProfilingTime)
	doUpdateHccsMetric(ch, npu, hccsBandwidthInfo.TotalTxbw, cardLabel, npuChipInfoHccsBWTotalTx)
	doUpdateHccsMetric(ch, npu, hccsBandwidthInfo.TotalRxbw, cardLabel, npuChipInfoHccsBWTotalRx)

	for i := hccsBeginIndex; i < dcmiHccsMaxCounts && i < len(hccsBWTxDescs) &&
		i < len(hccsBWRxDescs); i++ {
		doUpdateHccsMetric(ch, npu, hccsBandwidthInfo.TxBandwidth[i], cardLabel, hccsBWTxDescs[i])
		doUpdateHccsMetric(ch, npu, hccsBandwidthInfo.RxBandwidth[i], cardLabel, hccsBWRxDescs[i])
	}
}

func updateHBMECCInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if chip.HbmInfo == nil {
		hwlog.RunLog.Error("HbmInfo is nil in function updateHBMECCInfo")
		return
	}
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	eccInfo := common.DeepCopyECCInfo(chip.HbmInfo.ECCInfo)
	if !validate(ch, npu, chip, eccInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateHBMECCInfo")
		return
	}
	containerName, namespaceValue, podNameValue := getContainerInfoWithDefault(getContainerNameArray(devInfo))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccEnableFlag,
		prometheus.GaugeValue, float64(eccInfo.EnableFlag),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccSingleBitErrorCnt,
		prometheus.GaugeValue, float64(eccInfo.SingleBitErrorCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccDoubleBitErrorCnt,
		prometheus.GaugeValue, float64(eccInfo.DoubleBitErrorCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccTotalSingleBitErrorCnt,
		prometheus.GaugeValue, float64(eccInfo.TotalSingleBitErrorCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccTotalDoubleBitErrorCnt,
		prometheus.GaugeValue, float64(eccInfo.TotalDoubleBitErrorCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccSingleBitIoslatedPagesCnt,
		prometheus.GaugeValue, float64(eccInfo.SingleBitIsolatedPagesCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(
		npuChipInfoDescHbmEccDoubleBitIoslatedPagesCnt,
		prometheus.GaugeValue, float64(eccInfo.DoubleBitIsolatedPagesCnt),
		collectCardLabelValue(chip, namespaceValue, podNameValue, containerName)...))
}

var packChipInfo = func(logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	chip := &HuaWeiAIChip{}

	if info, err := dmgr.GetChipInfo(logicID); err != nil {
		hwlog.RunLog.Warnf("get chip info failed: %v", err)
		chip.ChipIfo = nil
	} else {
		chip.ChipIfo = info
	}

	if boardInfo, err := dmgr.GetBoardInfo(logicID); err != nil {
		hwlog.RunLog.Warnf("get board info failed: %v", err)
		chip.BoardInfo = nil
	} else {
		chip.BoardInfo = &boardInfo
	}

	chip.MainBoardId = dmgr.GetMainBoardId()

	packChipInfoPart2(logicID, dmgr, chip)
	packChipInfoPart1(logicID, dmgr, chip)
	return chip
}

func packChipInfoPart1(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	freq, err := dmgr.GetDeviceFrequency(logicID, common.AICoreCurrentFreq)
	if err != nil {
		freq = common.UnRetError
	}
	power, err := dmgr.GetDevicePowerInfo(logicID)
	if err != nil {
		power = common.UnRetError
	}
	temp, err := dmgr.GetDeviceTemperature(logicID)
	if err != nil {
		temp = common.RetError
	}
	vol, err := dmgr.GetDeviceVoltage(logicID)
	if err != nil {
		vol = common.UnRetError
	}
	mem, err := dmgr.GetDeviceMemoryInfo(logicID)
	if err != nil {
		mem = nil
	}
	hbmInfo, err := getAllHBMEccInfo(logicID, dmgr)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get all hbm info, err: %v", err)
	}
	sioInfo, err := dmgr.GetSioInfo(logicID)
	if err != nil {
		sioInfo = nil
	}
	hwChip.AICoreCurrentFreq = freq
	hwChip.Power = power
	hwChip.HealthStatus = getHealth(logicID, dmgr)
	hwChip.Temperature = int(temp)
	hwChip.Voltage = vol
	hwChip.Meminf = mem
	hwChip.HbmInfo = hbmInfo
	hwChip.SioInfo = sioInfo
	packHccsInfo(logicID, dmgr, hwChip)
	packHccsBandwidthInfo(logicID, dmgr, hwChip)

	// There is no PCIe link between 910A3 host and device.
	// Therefore, PCIe link bandwidth information cannot be queried.
	if dmgr.GetDevType() == common.Ascend910A3 {
		hwlog.RunLog.Debug("There is no PCIe link between Ascend910A3 host and device. " +
			"Therefore, PCIe link bandwidth information cannot be queried")
		hwChip.PcieBwInfo = nil
		return
	}

	if pcieBwInfo, err := dmgr.GetPCIEBandwidth(logicID, common.ProfilingTime); err != nil {
		hwChip.PcieBwInfo = nil
	} else {
		hwChip.PcieBwInfo = &pcieBwInfo
	}
}

func packHccsInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	if dmgr.GetDevType() != common.Ascend910B && dmgr.GetDevType() != common.Ascend910A3 {
		return
	}
	hccsStatisticInfo, err := dmgr.GetHccsStatisticInfo(logicID)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForHccs, logicID, "get hccs statistic info of npu failed: %v", err)
	} else {
		hwlog.ResetErrCnt(common.DomainForHccs, logicID)
	}
	hwChip.HccsStatisticInfo = hccsStatisticInfo
}

func packHccsBandwidthInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	if dmgr.GetDevType() != common.Ascend910B && dmgr.GetDevType() != common.Ascend910A3 {
		return
	}
	hccsBandwidthInfo, err := dmgr.GetHccsBandwidthInfo(logicID)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForHccsBW, logicID, "get hccs bandwidth info of npu failed: %v ", err)
	} else {
		hwlog.ResetErrCnt(common.DomainForHccsBW, logicID)
	}
	hwChip.HccsBandwidthInfo = hccsBandwidthInfo
}

func getAllHBMEccInfo(logicID int32, dmgr devmanager.DeviceInterface) (*common.HbmAggregateInfo, error) {
	hbmInfo := &common.HbmAggregateInfo{}
	var err error
	hbmInfo.HbmInfo, err = dmgr.GetDeviceHbmInfo(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get device ECC info, err: %v", err)
		hbmInfo.HbmInfo = nil
	}
	hbmInfo.ECCInfo, err = dmgr.GetDeviceEccInfo(logicID, common.DcmiDeviceTypeHBM)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get HBM ECC info, err: %v", err)
		hbmInfo.ECCInfo = nil
	}
	return hbmInfo, nil
}

func packChipInfoPart2(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	util, err := dmgr.GetDeviceUtilizationRate(logicID, common.AICore)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForAICoreUtilization, logicID,
			"get device(logicID:%d) AI core utilization rate failed, err is: %v", logicID, err)
		util = common.UnRetError // valid data range 0-100
	} else {
		hwlog.ResetErrCnt(common.DomainForAICoreUtilization, logicID)
	}
	overAllUtil, err := dmgr.GetDeviceUtilizationRate(logicID, common.Overall)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForOverallUtilization, logicID,
			"get device(logicID:%d) overall utilization rate of npu failed, err is: %v", logicID, err)
		overAllUtil = common.UnRetError // valid data range 0-100
	} else {
		hwlog.ResetErrCnt(common.DomainForOverallUtilization, logicID)
	}

	_, errCodes, err := dmgr.GetDeviceAllErrorCode(logicID)
	if err != nil {
		errCodes = make([]int64, 0)
	}
	vdieID, err := dmgr.GetDieID(logicID, dcmi.VDIE)
	if err != nil {
		hwlog.RunLog.Debug(err)
	}
	setNetHealthStatus(logicID, dmgr, hwChip)
	setProcessInfo(logicID, dmgr, hwChip)
	setPCIeBusInfo(logicID, dmgr, hwChip)
	hwChip.ErrorCodes = errCodes
	hwChip.Utilization = int(util)
	hwChip.OverallUtilization = int(overAllUtil)
	hwChip.VDieID = vdieID
	vecUtil, err := dmgr.GetDeviceUtilizationRate(logicID, common.VectorCore)

	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(common.DomainForVectorCoreUtilization, logicID,
			"get device(logicID:%d) vector core utilization rate failed, err is: %v", logicID, err)
		vecUtil = common.UnRetError // valid data range 0-100
	} else {
		hwlog.ResetErrCnt(common.DomainForVectorCoreUtilization, logicID)
	}
	hwChip.VectorUtilization = int(vecUtil)
}

func setNetHealthStatus(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	hwChip.NetHealthStatus = Abnormal
	if !dmgr.IsTrainingCard() {
		return
	}

	netCode, err := dmgr.GetDeviceNetWorkHealth(logicID)
	hwlog.RunLog.Debugf("chip %d network healthy code is %d", logicID, netCode)
	if err != nil {
		netCode = math.MaxUint32
	}
	hwChip.NetHealthStatus = getNetworkHealthy(netCode)
}

func setProcessInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	productTypes := dmgr.GetProductTypeArray()
	info, err := dmgr.GetDevProcessInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			hwlog.RunLog.Debugf("process info is not supported on %s", common.Atlas200ISoc)
			hwChip.DevProcessInfo = nil
			return
		}
		hwlog.RunLog.Error(err)
		info = nil
	}
	hwChip.DevProcessInfo = info
}

func setPCIeBusInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	productTypes := dmgr.GetProductTypeArray()
	pcieInfo, err := dmgr.GetPCIeBusInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			hwlog.RunLog.Debugf("pcie bus info is not supported on %s", common.Atlas200ISoc)
			hwChip.PCIeBusInfo = ""
			return
		}
		hwlog.RunLog.Error(err)
		pcieInfo = ""
	}
	hwChip.PCIeBusInfo = pcieInfo
}

func getMainOptInfo(opticalInfo map[string]string) *common.OpticalInfo {
	mainOpticalInfo := common.OpticalInfo{}
	mainOpticalInfo.OpticalTxPower0 = hccn.GetFloatDataFromStr(opticalInfo[txPower0], txPower0)
	mainOpticalInfo.OpticalTxPower1 = hccn.GetFloatDataFromStr(opticalInfo[txPower1], txPower1)
	mainOpticalInfo.OpticalTxPower2 = hccn.GetFloatDataFromStr(opticalInfo[txPower2], txPower2)
	mainOpticalInfo.OpticalTxPower3 = hccn.GetFloatDataFromStr(opticalInfo[txPower3], txPower3)
	mainOpticalInfo.OpticalRxPower0 = hccn.GetFloatDataFromStr(opticalInfo[rxPower0], rxPower0)
	mainOpticalInfo.OpticalRxPower1 = hccn.GetFloatDataFromStr(opticalInfo[rxPower1], rxPower1)
	mainOpticalInfo.OpticalRxPower2 = hccn.GetFloatDataFromStr(opticalInfo[rxPower2], rxPower2)
	mainOpticalInfo.OpticalRxPower3 = hccn.GetFloatDataFromStr(opticalInfo[rxPower3], rxPower3)
	mainOpticalInfo.OpticalVcc = hccn.GetFloatDataFromStr(opticalInfo[voltage], voltage)
	mainOpticalInfo.OpticalTemp = hccn.GetFloatDataFromStr(opticalInfo[temperature], temperature)
	var optState float64
	if opticalInfo[present] == present {
		optState = 1.0
	} else if opticalInfo[present] == notPresent {
		optState = 0.0
	} else {
		optState = common.RetError
	}
	mainOpticalInfo.OpticalState = optState

	return &mainOpticalInfo
}

func getMainStatInfo(statInfo map[string]int) *common.StatInfo {
	mainStatInfo := common.StatInfo{}
	mainStatInfo.MacRxPauseNum = float64(statInfo[macRxMacPauseNum])
	mainStatInfo.MacTxPauseNum = float64(statInfo[macTxMacPauseNum])
	mainStatInfo.MacRxPfcPktNum = float64(statInfo[macRxPfcPktNum])
	mainStatInfo.MacTxPfcPktNum = float64(statInfo[macTxPfcPktNum])
	mainStatInfo.MacRxBadPktNum = float64(statInfo[macRxBadPktNum])
	mainStatInfo.MacTxBadPktNum = float64(statInfo[macTxBadPktNum])
	mainStatInfo.RoceRxAllPktNum = float64(statInfo[roCERxAllPktNum])
	mainStatInfo.RoceTxAllPktNum = float64(statInfo[roCETxAllPktNum])
	mainStatInfo.RoceRxErrPktNum = float64(statInfo[roCERxErrPktNum])
	mainStatInfo.RoceTxErrPktNum = float64(statInfo[roCETxErrPktNum])
	mainStatInfo.RoceRxCnpPktNum = float64(statInfo[roCERxCnpPktNum])
	mainStatInfo.RoceTxCnpPktNum = float64(statInfo[roCETxCnpPktNum])
	mainStatInfo.MacRxBadOctNum = float64(statInfo[macRxBadOctNum])
	mainStatInfo.MacTxBadOctNum = float64(statInfo[macTxBadOctNum])
	mainStatInfo.RoceUnexpectedAckNum = float64(statInfo[roCEUnexpectedAckNum])
	mainStatInfo.RoceOutOfOrderNum = float64(statInfo[roCEOutOfOrderNum])
	mainStatInfo.RoceVerificationErrNum = float64(statInfo[roCEVerificationErrNum])
	mainStatInfo.RoceQpStatusErrNum = float64(statInfo[roCEQpStatusErrNum])
	mainStatInfo.RoceNewPktRtyNum = float64(statInfo[roCENewPktRtyNum])
	mainStatInfo.RoceEcnDBNum = float64(statInfo[roCEEcnDBNum])
	mainStatInfo.MacRXFcsErrPktNum = float64(statInfo[macRXFcsErrPktNum])

	return &mainStatInfo
}

func networkPackInfo(phyID int32) common.NpuNetInfo {
	newNetInfo := common.NpuNetInfo{}

	newNetInfo.LinkStatusInfo = &common.LinkStatusInfo{}
	if linkState, err := hccn.GetNPULinkStatus(phyID); err == nil {
		newNetInfo.LinkStatusInfo.LinkState = linkState
	} else {
		newNetInfo.LinkStatusInfo.LinkState = Abnormal
	}

	if tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID); err == nil {
		newNetInfo.BandwidthInfo = &common.BandwidthInfo{}
		newNetInfo.BandwidthInfo.RxValue = rx
		newNetInfo.BandwidthInfo.TxValue = tx
	} else {
		newNetInfo.BandwidthInfo = nil
	}

	if opticalInfo, err := hccn.GetNPUOpticalInfo(phyID); err == nil {
		newNetInfo.OpticalInfo = getMainOptInfo(opticalInfo)
	} else {
		newNetInfo.OpticalInfo = nil
	}

	if statInfo, err := hccn.GetNPUStatInfo(phyID); err == nil {
		newNetInfo.StatInfo = getMainStatInfo(statInfo)
	} else {
		newNetInfo.StatInfo = nil
	}

	if linkUpNum, err := hccn.GetNPULinkUpNum(phyID); err == nil {
		newNetInfo.LinkStatInfo = &common.LinkStatInfo{}
		newNetInfo.LinkStatInfo.LinkUPNum = float64(linkUpNum)
	} else {
		newNetInfo.LinkStatInfo = nil
	}

	if speed, err := hccn.GetNPULinkSpeed(phyID); err == nil {
		newNetInfo.LinkSpeedInfo = &common.LinkSpeedInfo{}
		newNetInfo.LinkSpeedInfo.Speed = float64(speed)
	} else {
		newNetInfo.LinkSpeedInfo = nil
	}

	return newNetInfo
}

func getHealth(logicID int32, dmgr devmanager.DeviceInterface) string {
	health, err := dmgr.GetDeviceHealth(logicID)
	if err != nil || health != 0 {
		return UnHealthy
	}
	return Healthy
}

func getHealthCode(health string) int {
	if health == Abnormal {
		return common.RetError
	}

	if Healthy == health {
		return 1
	}
	return 0
}

func getNetworkHealthy(netCode uint32) string {
	if netCode == math.MaxUint32 {
		return Abnormal
	}

	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return Healthy
	}

	return UnHealthy
}

func getPodDisplayInfo(chip *HuaWeiAIChip, containerName []string) []string {
	if len(containerName) != containerNameLen {
		hwlog.RunLog.Errorf("container name length %v is not %v", len(containerName), containerNameLen)
		return nil
	}

	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	vDevActivityInfo := common.DeepCopyVDevActivityInfo(chip.VDevActivityInfo)

	if !validateObj(chip) {
		hwlog.RunLog.Warnf("Invalid chip param in function getPodDisplayInfo")
		return []string{"", "", "", "",
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx], ""}
	}

	var vDevID, vDevAiCore, isVirtualDev string
	if !validateObj(vDevActivityInfo) {
		hwlog.RunLog.Warnf("Invalid vDevActivityInfo param in function getPodDisplayInfo")
		vDevID = ""
		vDevAiCore = ""
		isVirtualDev = ""
	} else {
		vDevID = strconv.Itoa(int(vDevActivityInfo.VDevID))
		vDevAiCore = strconv.FormatFloat(vDevActivityInfo.VDevAiCore, 'f', decimalPlaces, bitSize)
		isVirtualDev = strconv.FormatBool(vDevActivityInfo.IsVirtualDev)
	}

	return []string{
		strconv.Itoa(chip.DeviceID),
		common.GetNpuName(chipInfo),
		vDevID,
		vDevAiCore,
		containerName[nameSpaceIdx],
		containerName[podNameIdx],
		containerName[conNameIdx],
		isVirtualDev,
	}
}

func getContainerInfoWithDefault(cNameArray []string) (containerName, namespaceValue, podNameValue string) {
	if len(cNameArray) == containerNameLen {
		namespaceValue = cNameArray[nameSpaceIdx]
		podNameValue = cNameArray[podNameIdx]
		containerName = cNameArray[conNameIdx]
	}
	return containerName, namespaceValue, podNameValue
}

func collectCardLabelValue(chip *HuaWeiAIChip, namespaceValue, podNameValue, containerName string) []string {
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	if !validateObj(chip) {
		hwlog.RunLog.Warnf("Invalid chip param in function collectCardLabelValue")
		return []string{"", "", "", "", namespaceValue, podNameValue, containerName}
	}

	return []string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(chipInfo), chip.VDieID,
		chip.PCIeBusInfo, namespaceValue, podNameValue, containerName}
}
