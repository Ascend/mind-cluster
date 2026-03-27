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

// Package metrics for general collector
package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/hccn"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
)

const (
	isUboePort = "is_uboe_port"
	isUboe     = 1

	ubIpv4PktCntRx    = "ub_ipv4_pkt_cnt_rx"
	ubIpv6PktCntRx    = "ub_ipv6_pkt_cnt_rx"
	unicIpv4PktCntRx  = "unic_ipv4_pkt_cnt_rx"
	unicIpv6PktCntRx  = "unic_ipv6_pkt_cnt_rx"
	ubCompactPktCntRx = "ub_compact_pkt_cnt_rx"
	ubUmocCtphCntRx   = "ub_umoc_ctph_cnt_rx"
	ubUmocNtphCntRx   = "ub_umoc_ntph_cnt_rx"
	ubMemPktCntRx     = "ub_mem_pkt_cnt_rx"
	unknownPktCntRx   = "unknown_pkt_cnt_rx"
	dropIndCntRx      = "drop_ind_cnt_rx"
	errIndCntRx       = "err_ind_cnt_rx"
	toHostPktCntRx    = "to_host_pkt_cnt_rx"
	toImpPktCntRx     = "to_imp_pkt_cnt_rx"
	toMarPktCntRx     = "to_mar_pkt_cnt_rx"
	toLinkPktCntRx    = "to_link_pkt_cnt_rx"
	toNocPktCntRx     = "to_noc_pkt_cnt_rx"
	routeErrCntRx     = "route_err_cnt_rx"
	outErrCntRx       = "out_err_cnt_rx"
	lengthErrCntRx    = "length_err_cnt_rx"
	rxBusiFlitNum     = "rx_busi_flit_num"
	rxSendAckFlit     = "rx_send_ack_flit"

	ubIpv4PktCntTx    = "ub_ipv4_pkt_cnt_tx"
	ubIpv6PktCntTx    = "ub_ipv6_pkt_cnt_tx"
	unicIpv4PktCntTx  = "unic_ipv4_pkt_cnt_tx"
	unicIpv6PktCntTx  = "unic_ipv6_pkt_cnt_tx"
	ubCompactPktCntTx = "ub_compact_pkt_cnt_tx"
	ubUmocCtphCntTx   = "ub_umoc_ctph_cnt_tx"
	ubUmocNtphCntTx   = "ub_umoc_ntph_cnt_tx"
	ubMemPktCntTx     = "ub_mem_pkt_cnt_tx"
	unknownPktCntTx   = "unknown_pkt_cnt_tx"
	dropIndCntTx      = "drop_ind_cnt_tx"
	errIndCntTx       = "err_ind_cnt_tx"
	lpbkIndCntTx      = "lpbk_ind_cnt_tx"
	outErrCntTx       = "out_err_cnt_tx"
	lengthErrCntTx    = "length_err_cnt_tx"
	txBusiFlitNum     = "tx_busi_flit_num"
	txRecvAckFlit     = "tx_recv_ack_flit"

	retryReqSum = "retry_req_sum"
	retryAckSum = "retry_ack_sum"
	crcErrorSum = "crc_error_sum"

	coreMibRxpausepkts = "core_mib_rxpausepkts"
	coreMibTxpausepkts = "core_mib_txpausepkts"
	coreMibRxpfcpkts   = "core_mib_rxpfcpkts"
	coreMibTxpfcpkts   = "core_mib_txpfcpkts"
	coreMibRxbadpkts   = "core_mib_rxbadpkts"
	coreMibTxbadpkts   = "core_mib_txbadpkts"
	coreMibRxbadoctets = "core_mib_rxbadoctets"
	coreMibTxbadoctets = "core_mib_txbadoctets"
)

var (
	// rx
	ubIpv4PktCntRxDesc    []*prometheus.Desc
	ubIpv6PktCntRxDesc    []*prometheus.Desc
	unicIpv4PktCntRxDesc  []*prometheus.Desc
	unicIpv6PktCntRxDesc  []*prometheus.Desc
	ubCompactPktCntRxDesc []*prometheus.Desc
	ubUmocCtphCntRxDesc   []*prometheus.Desc
	ubUmocNtphCntRxDesc   []*prometheus.Desc
	ubMemPktCntRxDesc     []*prometheus.Desc
	unknownPktCntRxDesc   []*prometheus.Desc
	dropIndCntRxDesc      []*prometheus.Desc
	errIndCntRxDesc       []*prometheus.Desc
	toHostPktCntRxDesc    []*prometheus.Desc
	toImpPktCntRxDesc     []*prometheus.Desc
	toMarPktCntRxDesc     []*prometheus.Desc
	toLinkPktCntRxDesc    []*prometheus.Desc
	toNocPktCntRxDesc     []*prometheus.Desc
	routeErrCntRxDesc     []*prometheus.Desc
	outErrCntRxDesc       []*prometheus.Desc
	lengthErrCntRxDesc    []*prometheus.Desc
	rxBusiFlitNumDesc     []*prometheus.Desc
	rxSendAckFlitNumDesc  []*prometheus.Desc
	// tx
	ubIpv4PktCntTxDesc    []*prometheus.Desc
	ubIpv6PktCntTxDesc    []*prometheus.Desc
	unicIpv4PktCntTxDesc  []*prometheus.Desc
	unicIpv6PktCntTxDesc  []*prometheus.Desc
	ubCompactPktCntTxDesc []*prometheus.Desc
	ubUmocCtphCntTxDesc   []*prometheus.Desc
	ubUmocNtphCntTxDesc   []*prometheus.Desc
	ubMemPktCntTxDesc     []*prometheus.Desc
	unknownPktCntTxDesc   []*prometheus.Desc
	dropIndCntTxDesc      []*prometheus.Desc
	errIndCntTxDesc       []*prometheus.Desc
	lpbkIndCntTxDesc      []*prometheus.Desc
	outErrCntTxDesc       []*prometheus.Desc
	lengthErrCntTxDesc    []*prometheus.Desc
	txBusiFlitNumDesc     []*prometheus.Desc
	txRecvAckFlitDesc     []*prometheus.Desc
	// sum
	retryReqSumDesc []*prometheus.Desc
	retryAckSumDesc []*prometheus.Desc
	crcErrorSumDesc []*prometheus.Desc
	// uboe
	coreMibRxpausepktsDesc []*prometheus.Desc
	coreMibTxpausepktsDesc []*prometheus.Desc
	coreMibRxpfcpktsDesc   []*prometheus.Desc
	coreMibTxpfcpktsDesc   []*prometheus.Desc
	coreMibRxbadpktsDesc   []*prometheus.Desc
	coreMibTxbadpktsDesc   []*prometheus.Desc
	coreMibRxbadoctetsDesc []*prometheus.Desc
	coreMibTxbadoctetsDesc []*prometheus.Desc

	supportedUbDevices = map[uint32]bool{
		api.Atlas3504PMainBoardID: true,
		api.Atlas9501DMainBoardID: true,
		api.Atlas950MainBoardID:   true,
		api.Atlas850MainBoardID:   true,
		api.Atlas850MainBoardID2:  true,
	}
)

func init() {
	for dieID := 0; dieID < common.MaxDieID; dieID++ {
		for portID := 0; portID < common.MaxPortID; portID++ {
			// rx
			initBuildDescRx(dieID, portID)
			// tx
			initBuildDescTx(dieID, portID)
			// sum
			buildDesc(dieID, portID, &retryReqSumDesc, retryReqSum, "number of retransmission attempts initiated on ")
			buildDesc(dieID, portID, &retryAckSumDesc, retryAckSum, "number of response retransmissions on ")
			buildDesc(dieID, portID, &crcErrorSumDesc, crcErrorSum, "number of crc check errors ")
			// uboe
			buildDesc(dieID, portID, &coreMibRxpausepktsDesc, coreMibRxpausepkts, "uboe total number of rx pause frames on ")
			buildDesc(dieID, portID, &coreMibTxpausepktsDesc, coreMibTxpausepkts, "uboe total number of tx pause frames on ")
			buildDesc(dieID, portID, &coreMibRxpfcpktsDesc, coreMibRxpfcpkts, "uboe total number of rx pfc frames on ")
			buildDesc(dieID, portID, &coreMibTxpfcpktsDesc, coreMibTxpfcpkts, "uboe total number of tx pfc frames on ")
			buildDesc(dieID, portID, &coreMibRxbadpktsDesc, coreMibRxbadpkts, "uboe total number of rx bad packets on ")
			buildDesc(dieID, portID, &coreMibTxbadpktsDesc, coreMibTxbadpkts, "uboe total number of tx bad packets on ")
			buildDesc(dieID, portID, &coreMibRxbadoctetsDesc, coreMibRxbadoctets, "uboe total number of bytes in rx bad packets on ")
			buildDesc(dieID, portID, &coreMibTxbadoctetsDesc, coreMibTxbadoctets, "uboe total number of bytes in tx bad packets on ")
		}
	}
}

type ubCache struct {
	chip      colcommon.HuaWeiAIChip
	timestamp time.Time
	// extInfo the statistics about packets
	ubInfo []*common.UBInfo
}

// UbCollector collect ub info
type UbCollector struct {
	colcommon.MetricsCollectorAdapter
}

// IsSupported check whether the collector is supported
func (c *UbCollector) IsSupported(n *colcommon.NpuCollector) bool {
	devType := n.Dmgr.GetDevType()
	mainBoardID := n.Dmgr.GetMainBoardId()
	isSupport := devType == api.Ascend910A5 && supportedUbDevices[mainBoardID]
	logForUnSupportDevice(isSupport, devType, colcommon.GetCacheKey(c),
		fmt.Sprint("this mainBoardId:", mainBoardID, " is not supported"))
	return isSupport
}

// Describe description of the metric
func (c *UbCollector) Describe(ch chan<- *prometheus.Desc) {
	// ub rx
	initUbRxDesc(ch)
	// ub tx
	initUbTxDesc(ch)
	// sum
	initDesc(ch, retryReqSumDesc)
	initDesc(ch, retryAckSumDesc)
	initDesc(ch, crcErrorSumDesc)
	// uboe
	initUboeDesc(ch)
}

// CollectToCache collect the metric to cache
func (c *UbCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		ubInfo := collectUbInfo(chip.LogicID)
		c.LocalCache.Store(chip.PhyId, ubCache{
			chip:      chip,
			timestamp: time.Now(),
			ubInfo:    ubInfo},
		)
	}
	colcommon.UpdateCache[ubCache](n, colcommon.GetCacheKey(c), &c.LocalCache)
}

// UpdatePrometheus update prometheus metrics
func (c *UbCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) {
	updateSingleChip := func(chipWithVnpu colcommon.HuaWeiAIChip, cache ubCache, cardLabel []string) {
		timestamp := cache.timestamp
		promUpdateUbInfo(ch, cache, timestamp, cardLabel)
	}
	updateFrame[ubCache](colcommon.GetCacheKey(c), n, containerMap, chips, updateSingleChip)
}

func promUpdateUbInfo(ch chan<- prometheus.Metric, cache ubCache,
	timestamp time.Time, cardLabel []string) {
	ubInfo := cache.ubInfo
	if ubInfo == nil {
		return
	}
	for i := 0; i < (common.MaxDieID * common.MaxPortID); i++ {
		if ubInfo[i].UBCommonStats != nil {
			// rx
			promUpdateUbRx(ch, timestamp, ubInfo, cardLabel, i)
			// tx
			promUpdateUbTx(ch, timestamp, ubInfo, cardLabel, i)
			// sum
			promUpdateUbSum(ch, timestamp, ubInfo, cardLabel, i)
		}
		if ubInfo[i].UboeExtensions != nil {
			// uboe
			promUpdateUbUboe(ch, timestamp, ubInfo, cardLabel, i)
		}
	}
}

// UpdateTelegraf update telegraf
func (c *UbCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *colcommon.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []colcommon.HuaWeiAIChip) map[string]map[string]interface{} {
	caches := colcommon.GetInfoFromCache[ubCache](n, colcommon.GetCacheKey(c))
	for _, chip := range chips {
		cache, ok := caches[chip.PhyId]
		if !ok {
			continue
		}
		fieldMap := getFieldMap(fieldsMap, cache.chip.PhyId)
		telegrafUpdateUbInfo(cache, fieldMap)
	}
	return fieldsMap
}

func promUpdateUbRx(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv4PktCntRx, cardLabel, ubIpv4PktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv6PktCntRx, cardLabel, ubIpv6PktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv4PktCntRx, cardLabel, unicIpv4PktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv6PktCntRx, cardLabel, unicIpv6PktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbCompactPktCntRx, cardLabel, ubCompactPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocCtphCntRx, cardLabel, ubUmocCtphCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocNtphCntRx, cardLabel, ubUmocNtphCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbMemPktCntRx, cardLabel, ubMemPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnknownPktCntRx, cardLabel, unknownPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.DropIndCntRx, cardLabel, dropIndCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ErrIndCntRx, cardLabel, errIndCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToHostPktCntRx, cardLabel, toHostPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToImpPktCntRx, cardLabel, toImpPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToMarPktCntRx, cardLabel, toMarPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToLinkPktCntRx, cardLabel, toLinkPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToNocPktCntRx, cardLabel, toNocPktCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RouteErrCntRx, cardLabel, routeErrCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.OutErrCntRx, cardLabel, outErrCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LengthErrCntRx, cardLabel, lengthErrCntRxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RxBusiFlitNum, cardLabel, rxBusiFlitNumDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RxSendAckFlit, cardLabel, rxSendAckFlitNumDesc[i])
}

func promUpdateUbTx(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv4PktCntTx, cardLabel, ubIpv4PktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv6PktCntTx, cardLabel, ubIpv6PktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv4PktCntTx, cardLabel, unicIpv4PktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv6PktCntTx, cardLabel, unicIpv6PktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbCompactPktCntTx, cardLabel, ubCompactPktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocCtphCntTx, cardLabel, ubUmocCtphCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocNtphCntTx, cardLabel, ubUmocNtphCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbMemPktCntTx, cardLabel, ubMemPktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnknownPktCntTx, cardLabel, unknownPktCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.DropIndCntTx, cardLabel, dropIndCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ErrIndCntTx, cardLabel, errIndCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LpbkIndCntTx, cardLabel, lpbkIndCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.OutErrCntTx, cardLabel, outErrCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LengthErrCntTx, cardLabel, lengthErrCntTxDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.TxBusiFlitNum, cardLabel, txBusiFlitNumDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.TxRecvAckFlit, cardLabel, txRecvAckFlitDesc[i])
}

func promUpdateUbSum(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RetryReqSum, cardLabel, retryReqSumDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RetryAckSum, cardLabel, retryAckSumDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.CrcErrorSum, cardLabel, crcErrorSumDesc[i])
}

func promUpdateUbUboe(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxPausePkts, cardLabel, coreMibRxpausepktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxPausePkts, cardLabel, coreMibTxpausepktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxPfcPkts, cardLabel, coreMibRxpfcpktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxPfcPkts, cardLabel, coreMibTxpfcpktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxBadPkts, cardLabel, coreMibRxbadpktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxBadPkts, cardLabel, coreMibTxbadpktsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxBadOctets, cardLabel, coreMibRxbadoctetsDesc[i])
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxBadOctets, cardLabel, coreMibTxbadoctetsDesc[i])
}

func telegrafUpdateUbInfo(cache ubCache, fieldMap map[string]interface{}) {
	ubInfo := cache.ubInfo
	if ubInfo == nil {
		return
	}
	for i := 0; i < (common.MaxDieID * common.MaxPortID); i++ {
		if ubInfo[i].UBCommonStats != nil {
			// rx
			telegrafUpdateUbRx(fieldMap, ubInfo, i)
			// tx
			telegrafUpdateUbTx(fieldMap, ubInfo, i)
			// sum
			telegrafUpdateUbSum(fieldMap, ubInfo, i)
		}
		if ubInfo[i].UboeExtensions != nil {
			// uboe
			telegrafUpdateUbUboe(fieldMap, ubInfo, i)
		}
	}
}

func telegrafUpdateUbRx(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int) {
	doUpdateTelegraf(fieldMap, ubIpv4PktCntRxDesc[i], ubInfo[i].UBCommonStats.UbIpv4PktCntRx, "")
	doUpdateTelegraf(fieldMap, ubIpv6PktCntRxDesc[i], ubInfo[i].UBCommonStats.UbIpv6PktCntRx, "")
	doUpdateTelegraf(fieldMap, unicIpv4PktCntRxDesc[i], ubInfo[i].UBCommonStats.UnicIpv4PktCntRx, "")
	doUpdateTelegraf(fieldMap, unicIpv6PktCntRxDesc[i], ubInfo[i].UBCommonStats.UnicIpv6PktCntRx, "")
	doUpdateTelegraf(fieldMap, ubCompactPktCntRxDesc[i], ubInfo[i].UBCommonStats.UbCompactPktCntRx, "")
	doUpdateTelegraf(fieldMap, ubUmocCtphCntRxDesc[i], ubInfo[i].UBCommonStats.UbUmocCtphCntRx, "")
	doUpdateTelegraf(fieldMap, ubUmocNtphCntRxDesc[i], ubInfo[i].UBCommonStats.UbUmocNtphCntRx, "")
	doUpdateTelegraf(fieldMap, ubMemPktCntRxDesc[i], ubInfo[i].UBCommonStats.UbMemPktCntRx, "")
	doUpdateTelegraf(fieldMap, unknownPktCntRxDesc[i], ubInfo[i].UBCommonStats.UnknownPktCntRx, "")
	doUpdateTelegraf(fieldMap, dropIndCntRxDesc[i], ubInfo[i].UBCommonStats.DropIndCntRx, "")
	doUpdateTelegraf(fieldMap, errIndCntRxDesc[i], ubInfo[i].UBCommonStats.ErrIndCntRx, "")
	doUpdateTelegraf(fieldMap, toHostPktCntRxDesc[i], ubInfo[i].UBCommonStats.ToHostPktCntRx, "")
	doUpdateTelegraf(fieldMap, toImpPktCntRxDesc[i], ubInfo[i].UBCommonStats.ToImpPktCntRx, "")
	doUpdateTelegraf(fieldMap, toMarPktCntRxDesc[i], ubInfo[i].UBCommonStats.ToMarPktCntRx, "")
	doUpdateTelegraf(fieldMap, toLinkPktCntRxDesc[i], ubInfo[i].UBCommonStats.ToLinkPktCntRx, "")
	doUpdateTelegraf(fieldMap, toNocPktCntRxDesc[i], ubInfo[i].UBCommonStats.ToNocPktCntRx, "")
	doUpdateTelegraf(fieldMap, routeErrCntRxDesc[i], ubInfo[i].UBCommonStats.RouteErrCntRx, "")
	doUpdateTelegraf(fieldMap, outErrCntRxDesc[i], ubInfo[i].UBCommonStats.OutErrCntRx, "")
	doUpdateTelegraf(fieldMap, lengthErrCntRxDesc[i], ubInfo[i].UBCommonStats.LengthErrCntRx, "")
	doUpdateTelegraf(fieldMap, rxBusiFlitNumDesc[i], ubInfo[i].UBCommonStats.RxBusiFlitNum, "")
	doUpdateTelegraf(fieldMap, rxSendAckFlitNumDesc[i], ubInfo[i].UBCommonStats.RxSendAckFlit, "")
}

func telegrafUpdateUbTx(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int) {
	doUpdateTelegraf(fieldMap, ubIpv4PktCntTxDesc[i], ubInfo[i].UBCommonStats.UbIpv4PktCntTx, "")
	doUpdateTelegraf(fieldMap, ubIpv6PktCntTxDesc[i], ubInfo[i].UBCommonStats.UbIpv6PktCntTx, "")
	doUpdateTelegraf(fieldMap, unicIpv4PktCntTxDesc[i], ubInfo[i].UBCommonStats.UnicIpv4PktCntTx, "")
	doUpdateTelegraf(fieldMap, unicIpv6PktCntTxDesc[i], ubInfo[i].UBCommonStats.UnicIpv6PktCntTx, "")
	doUpdateTelegraf(fieldMap, ubCompactPktCntTxDesc[i], ubInfo[i].UBCommonStats.UbCompactPktCntTx, "")
	doUpdateTelegraf(fieldMap, ubUmocCtphCntTxDesc[i], ubInfo[i].UBCommonStats.UbUmocCtphCntTx, "")
	doUpdateTelegraf(fieldMap, ubUmocNtphCntTxDesc[i], ubInfo[i].UBCommonStats.UbUmocNtphCntTx, "")
	doUpdateTelegraf(fieldMap, ubMemPktCntTxDesc[i], ubInfo[i].UBCommonStats.UbMemPktCntTx, "")
	doUpdateTelegraf(fieldMap, unknownPktCntTxDesc[i], ubInfo[i].UBCommonStats.UnknownPktCntTx, "")
	doUpdateTelegraf(fieldMap, dropIndCntTxDesc[i], ubInfo[i].UBCommonStats.DropIndCntTx, "")
	doUpdateTelegraf(fieldMap, errIndCntTxDesc[i], ubInfo[i].UBCommonStats.ErrIndCntTx, "")
	doUpdateTelegraf(fieldMap, lpbkIndCntTxDesc[i], ubInfo[i].UBCommonStats.LpbkIndCntTx, "")
	doUpdateTelegraf(fieldMap, outErrCntTxDesc[i], ubInfo[i].UBCommonStats.OutErrCntTx, "")
	doUpdateTelegraf(fieldMap, lengthErrCntTxDesc[i], ubInfo[i].UBCommonStats.LengthErrCntTx, "")
	doUpdateTelegraf(fieldMap, txBusiFlitNumDesc[i], ubInfo[i].UBCommonStats.TxBusiFlitNum, "")
	doUpdateTelegraf(fieldMap, txRecvAckFlitDesc[i], ubInfo[i].UBCommonStats.TxRecvAckFlit, "")
}

func telegrafUpdateUbSum(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int) {
	doUpdateTelegraf(fieldMap, retryReqSumDesc[i], ubInfo[i].UBCommonStats.RetryReqSum, "")
	doUpdateTelegraf(fieldMap, retryAckSumDesc[i], ubInfo[i].UBCommonStats.RetryAckSum, "")
	doUpdateTelegraf(fieldMap, crcErrorSumDesc[i], ubInfo[i].UBCommonStats.CrcErrorSum, "")
}

func telegrafUpdateUbUboe(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int) {
	doUpdateTelegraf(fieldMap, coreMibRxpausepktsDesc[i], ubInfo[i].UboeExtensions.CoreMibRxPausePkts, "")
	doUpdateTelegraf(fieldMap, coreMibTxpausepktsDesc[i], ubInfo[i].UboeExtensions.CoreMibTxPausePkts, "")
	doUpdateTelegraf(fieldMap, coreMibRxpfcpktsDesc[i], ubInfo[i].UboeExtensions.CoreMibRxPfcPkts, "")
	doUpdateTelegraf(fieldMap, coreMibTxpfcpktsDesc[i], ubInfo[i].UboeExtensions.CoreMibTxPfcPkts, "")
	doUpdateTelegraf(fieldMap, coreMibRxbadpktsDesc[i], ubInfo[i].UboeExtensions.CoreMibRxBadPkts, "")
	doUpdateTelegraf(fieldMap, coreMibTxbadpktsDesc[i], ubInfo[i].UboeExtensions.CoreMibTxBadPkts, "")
	doUpdateTelegraf(fieldMap, coreMibRxbadoctetsDesc[i], ubInfo[i].UboeExtensions.CoreMibRxBadOctets, "")
	doUpdateTelegraf(fieldMap, coreMibTxbadoctetsDesc[i], ubInfo[i].UboeExtensions.CoreMibTxBadOctets, "")
}

func initUboeDesc(ch chan<- *prometheus.Desc) {
	initDesc(ch, coreMibRxpausepktsDesc)
	initDesc(ch, coreMibTxpausepktsDesc)
	initDesc(ch, coreMibRxpfcpktsDesc)
	initDesc(ch, coreMibTxpfcpktsDesc)
	initDesc(ch, coreMibRxbadpktsDesc)
	initDesc(ch, coreMibTxbadpktsDesc)
	initDesc(ch, coreMibRxbadoctetsDesc)
	initDesc(ch, coreMibTxbadoctetsDesc)
}

func initUbRxDesc(ch chan<- *prometheus.Desc) {
	initDesc(ch, ubIpv4PktCntRxDesc)
	initDesc(ch, ubIpv6PktCntRxDesc)
	initDesc(ch, unicIpv4PktCntRxDesc)
	initDesc(ch, unicIpv6PktCntRxDesc)
	initDesc(ch, ubCompactPktCntRxDesc)
	initDesc(ch, ubUmocCtphCntRxDesc)
	initDesc(ch, ubUmocNtphCntRxDesc)
	initDesc(ch, ubMemPktCntRxDesc)
	initDesc(ch, unknownPktCntRxDesc)
	initDesc(ch, dropIndCntRxDesc)
	initDesc(ch, errIndCntRxDesc)
	initDesc(ch, toHostPktCntRxDesc)
	initDesc(ch, toImpPktCntRxDesc)
	initDesc(ch, toMarPktCntRxDesc)
	initDesc(ch, toLinkPktCntRxDesc)
	initDesc(ch, toNocPktCntRxDesc)
	initDesc(ch, routeErrCntRxDesc)
	initDesc(ch, outErrCntRxDesc)
	initDesc(ch, lengthErrCntRxDesc)
	initDesc(ch, rxBusiFlitNumDesc)
	initDesc(ch, rxSendAckFlitNumDesc)
}

func initUbTxDesc(ch chan<- *prometheus.Desc) {
	initDesc(ch, ubIpv4PktCntTxDesc)
	initDesc(ch, ubIpv6PktCntTxDesc)
	initDesc(ch, unicIpv4PktCntTxDesc)
	initDesc(ch, unicIpv6PktCntTxDesc)
	initDesc(ch, ubCompactPktCntTxDesc)
	initDesc(ch, ubUmocCtphCntTxDesc)
	initDesc(ch, ubUmocNtphCntTxDesc)
	initDesc(ch, ubMemPktCntTxDesc)
	initDesc(ch, unknownPktCntTxDesc)
	initDesc(ch, dropIndCntTxDesc)
	initDesc(ch, errIndCntTxDesc)
	initDesc(ch, lpbkIndCntTxDesc)
	initDesc(ch, outErrCntTxDesc)
	initDesc(ch, lengthErrCntTxDesc)
	initDesc(ch, txBusiFlitNumDesc)
	initDesc(ch, txRecvAckFlitDesc)
}

func initBuildDescRx(dieID, portID int) {
	buildDesc(dieID, portID, &ubIpv4PktCntRxDesc, ubIpv4PktCntRx, "number of ipv4 ub packets received by rx on ")
	buildDesc(dieID, portID, &ubIpv6PktCntRxDesc, ubIpv6PktCntRx, "number of ipv6 ub packets received by rx on ")
	buildDesc(dieID, portID, &unicIpv4PktCntRxDesc, unicIpv4PktCntRx, "number of ipv4 unic packets received by rx on ")
	buildDesc(dieID, portID, &unicIpv6PktCntRxDesc, unicIpv6PktCntRx, "number of ipv6 unic packets received by rx on ")
	buildDesc(dieID, portID, &ubCompactPktCntRxDesc, ubCompactPktCntRx, "number of cfg6 packets received by rx on ")
	buildDesc(dieID, portID, &ubUmocCtphCntRxDesc, ubUmocCtphCntRx, "number of cfg7 clan packets received by rx on ")
	buildDesc(dieID, portID, &ubUmocNtphCntRxDesc, ubUmocNtphCntRx, "number of cfg7 not clan packets received by rx on ")
	buildDesc(dieID, portID, &ubMemPktCntRxDesc, ubMemPktCntRx, "number of ub mem packets received by rx on ")
	buildDesc(dieID, portID, &unknownPktCntRxDesc, unknownPktCntRx, "number of unknown packets received by rx on ")
	buildDesc(dieID, portID, &dropIndCntRxDesc, dropIndCntRx, "number of packet with drop_ind received by rx on ")
	buildDesc(dieID, portID, &errIndCntRxDesc, errIndCntRx, "number of err packets received by rx on ")
	buildDesc(dieID, portID, &toHostPktCntRxDesc, toHostPktCntRx, "number of landed packets after routing on the rx ")
	buildDesc(dieID, portID, &toImpPktCntRxDesc, toImpPktCntRx, "number of landed enumeration configuration and management packets after routing on the rx ")
	buildDesc(dieID, portID, &toMarPktCntRxDesc, toMarPktCntRx, "number of landed ub memory packets after routing on the rx ")
	buildDesc(dieID, portID, &toLinkPktCntRxDesc, toLinkPktCntRx, "number of packets forward from the rx to the tx of the same port after routing ")
	buildDesc(dieID, portID, &toNocPktCntRxDesc, toNocPktCntRx, "number of p2p packets received on the rx after routing ")
	buildDesc(dieID, portID, &routeErrCntRxDesc, routeErrCntRx, "number of packets with routing lookup errors after processing received on the rx ")
	buildDesc(dieID, portID, &outErrCntRxDesc, outErrCntRx, "total number of erroneous packets after validation of packets received on the rx ")
	buildDesc(dieID, portID, &lengthErrCntRxDesc, lengthErrCntRx, "number of packets with length errors after validation of packets received on the rx ")
	buildDesc(dieID, portID, &rxBusiFlitNumDesc, rxBusiFlitNum, "number of flits of service packets received from the mac on the rx ")
	buildDesc(dieID, portID, &rxSendAckFlitNumDesc, rxSendAckFlit, "cumulative number of acks released to the peer on the rx ")
}

func initBuildDescTx(dieID, portID int) {
	buildDesc(dieID, portID, &ubIpv4PktCntTxDesc, ubIpv4PktCntTx, "number of ipv4 ub packets sent by tx on ")
	buildDesc(dieID, portID, &ubIpv6PktCntTxDesc, ubIpv6PktCntTx, "number of ipv6 ub packets sent by tx on ")
	buildDesc(dieID, portID, &unicIpv4PktCntTxDesc, unicIpv4PktCntTx, "number of ipv4 unic packets sent by tx on ")
	buildDesc(dieID, portID, &unicIpv6PktCntTxDesc, unicIpv6PktCntTx, "number of ipv6 unic packets sent by tx on ")
	buildDesc(dieID, portID, &ubCompactPktCntTxDesc, ubCompactPktCntTx, "number of cfg6 packets sent by tx on ")
	buildDesc(dieID, portID, &ubUmocCtphCntTxDesc, ubUmocCtphCntTx, "number of cfg7 clan packets sent by tx on ")
	buildDesc(dieID, portID, &ubUmocNtphCntTxDesc, ubUmocNtphCntTx, "number of cfg7 not clan packets sent by tx on ")
	buildDesc(dieID, portID, &ubMemPktCntTxDesc, ubMemPktCntTx, "number of ub mem packets sent by tx on ")
	buildDesc(dieID, portID, &unknownPktCntTxDesc, unknownPktCntTx, "number of unknown packets sent by tx on ")
	buildDesc(dieID, portID, &dropIndCntTxDesc, dropIndCntTx, "number of packet with drop_ind sent by tx on ")
	buildDesc(dieID, portID, &errIndCntTxDesc, errIndCntTx, "number of err packets sent by tx on ")
	buildDesc(dieID, portID, &lpbkIndCntTxDesc, lpbkIndCntTx, "number of packets looped back at nl by tx on ")
	buildDesc(dieID, portID, &outErrCntTxDesc, outErrCntTx, "total number of erroneous packets after validation of packets sent on the tx ")
	buildDesc(dieID, portID, &lengthErrCntTxDesc, lengthErrCntTx, "number of packets with length errors after validation of packets sent on the tx ")
	buildDesc(dieID, portID, &txBusiFlitNumDesc, txBusiFlitNum, "number of flits of service packets sent from the mac on the tx ")
	buildDesc(dieID, portID, &txRecvAckFlitDesc, txRecvAckFlit, "cumulative number of acks released to the peer on the tx ")
}

func collectUbInfo(logicID int32) []*common.UBInfo {
	var newUbInfos []*common.UBInfo
	for dieID := 0; dieID < common.MaxDieID; dieID++ {
		for portID := 0; portID < common.MaxPortID; portID++ {
			newUbInfos = append(newUbInfos, getUBStatInfo(logicID, dieID, portID))
		}
	}
	return newUbInfos
}

func getUBStatInfo(logicID int32, dieID, portID int) *common.UBInfo {
	ubInfos := common.UBInfo{
		UBCommonStats:  &common.UBCommonStats{},
		UboeExtensions: &common.UBOEExtensions{},
	}
	if ubInfo, err := hccn.GetNPUUbStatInfo(logicID, int32(dieID), int32(portID)); err == nil {
		if result, err := strconv.Atoi(ubInfo[isUboePort]); err == nil && result == isUboe {
			hwlog.RunLog.Debug("is uboe port")
			convertUboeExtensions(&ubInfos, ubInfo)
		}
		convertUBCommonStats(&ubInfos, ubInfo)
		hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForUb, dieID, portID), logicID)
	} else {
		ubInfos.UBCommonStats = nil
		ubInfos.UboeExtensions = nil
		logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForUb, dieID, portID), logicID, dieID, portID, err)
	}
	return &ubInfos
}

func convertUboeExtensions(ubInfos *common.UBInfo, ubInfo map[string]string) {
	ubInfos.UboeExtensions.CoreMibRxPausePkts = hccn.GetIntDataFromStr(ubInfo[coreMibRxpausepkts], coreMibRxpausepkts)
	ubInfos.UboeExtensions.CoreMibTxPausePkts = hccn.GetIntDataFromStr(ubInfo[coreMibTxpausepkts], coreMibTxpausepkts)
	ubInfos.UboeExtensions.CoreMibRxPfcPkts = hccn.GetIntDataFromStr(ubInfo[coreMibRxpfcpkts], coreMibRxpfcpkts)
	ubInfos.UboeExtensions.CoreMibTxPfcPkts = hccn.GetIntDataFromStr(ubInfo[coreMibTxpfcpkts], coreMibTxpfcpkts)
	ubInfos.UboeExtensions.CoreMibRxBadPkts = hccn.GetIntDataFromStr(ubInfo[coreMibRxbadpkts], coreMibRxbadpkts)
	ubInfos.UboeExtensions.CoreMibTxBadPkts = hccn.GetIntDataFromStr(ubInfo[coreMibTxbadpkts], coreMibTxbadpkts)
	ubInfos.UboeExtensions.CoreMibRxBadOctets = hccn.GetIntDataFromStr(ubInfo[coreMibRxbadoctets], coreMibRxbadoctets)
	ubInfos.UboeExtensions.CoreMibTxBadOctets = hccn.GetIntDataFromStr(ubInfo[coreMibTxbadoctets], coreMibTxbadoctets)
}

func convertUBCommonStats(ubInfos *common.UBInfo, ubInfo map[string]string) {
	ubInfos.UBCommonStats.UbIpv4PktCntRx = hccn.GetIntDataFromStr(ubInfo[ubIpv4PktCntRx], ubIpv4PktCntRx)
	ubInfos.UBCommonStats.UbIpv6PktCntRx = hccn.GetIntDataFromStr(ubInfo[ubIpv6PktCntRx], ubIpv6PktCntRx)
	ubInfos.UBCommonStats.UnicIpv4PktCntRx = hccn.GetIntDataFromStr(ubInfo[unicIpv4PktCntRx], unicIpv4PktCntRx)
	ubInfos.UBCommonStats.UnicIpv6PktCntRx = hccn.GetIntDataFromStr(ubInfo[unicIpv6PktCntRx], unicIpv6PktCntRx)
	ubInfos.UBCommonStats.UbCompactPktCntRx = hccn.GetIntDataFromStr(ubInfo[ubCompactPktCntRx], ubCompactPktCntRx)
	ubInfos.UBCommonStats.UbUmocCtphCntRx = hccn.GetIntDataFromStr(ubInfo[ubUmocCtphCntRx], ubUmocCtphCntRx)
	ubInfos.UBCommonStats.UbUmocNtphCntRx = hccn.GetIntDataFromStr(ubInfo[ubUmocNtphCntRx], ubUmocNtphCntRx)
	ubInfos.UBCommonStats.UbMemPktCntRx = hccn.GetIntDataFromStr(ubInfo[ubMemPktCntRx], ubMemPktCntRx)
	ubInfos.UBCommonStats.UnknownPktCntRx = hccn.GetIntDataFromStr(ubInfo[unknownPktCntRx], unknownPktCntRx)
	ubInfos.UBCommonStats.DropIndCntRx = hccn.GetIntDataFromStr(ubInfo[dropIndCntRx], dropIndCntRx)
	ubInfos.UBCommonStats.ErrIndCntRx = hccn.GetIntDataFromStr(ubInfo[errIndCntRx], errIndCntRx)
	ubInfos.UBCommonStats.ToHostPktCntRx = hccn.GetIntDataFromStr(ubInfo[toHostPktCntRx], toHostPktCntRx)
	ubInfos.UBCommonStats.ToImpPktCntRx = hccn.GetIntDataFromStr(ubInfo[toImpPktCntRx], toImpPktCntRx)
	ubInfos.UBCommonStats.ToMarPktCntRx = hccn.GetIntDataFromStr(ubInfo[toMarPktCntRx], toMarPktCntRx)
	ubInfos.UBCommonStats.ToLinkPktCntRx = hccn.GetIntDataFromStr(ubInfo[toLinkPktCntRx], toLinkPktCntRx)
	ubInfos.UBCommonStats.ToNocPktCntRx = hccn.GetIntDataFromStr(ubInfo[toNocPktCntRx], toNocPktCntRx)
	ubInfos.UBCommonStats.RouteErrCntRx = hccn.GetIntDataFromStr(ubInfo[routeErrCntRx], routeErrCntRx)
	ubInfos.UBCommonStats.OutErrCntRx = hccn.GetIntDataFromStr(ubInfo[outErrCntRx], outErrCntRx)
	ubInfos.UBCommonStats.LengthErrCntRx = hccn.GetIntDataFromStr(ubInfo[lengthErrCntRx], lengthErrCntRx)
	ubInfos.UBCommonStats.RxBusiFlitNum = hccn.GetIntDataFromStr(ubInfo[rxBusiFlitNum], rxBusiFlitNum)
	ubInfos.UBCommonStats.RxSendAckFlit = hccn.GetIntDataFromStr(ubInfo[rxSendAckFlit], rxSendAckFlit)

	ubInfos.UBCommonStats.UbIpv4PktCntTx = hccn.GetIntDataFromStr(ubInfo[ubIpv4PktCntTx], ubIpv4PktCntTx)
	ubInfos.UBCommonStats.UbIpv6PktCntTx = hccn.GetIntDataFromStr(ubInfo[ubIpv6PktCntTx], ubIpv6PktCntTx)
	ubInfos.UBCommonStats.UnicIpv4PktCntTx = hccn.GetIntDataFromStr(ubInfo[unicIpv4PktCntTx], unicIpv4PktCntTx)
	ubInfos.UBCommonStats.UnicIpv6PktCntTx = hccn.GetIntDataFromStr(ubInfo[unicIpv6PktCntTx], unicIpv6PktCntTx)
	ubInfos.UBCommonStats.UbCompactPktCntTx = hccn.GetIntDataFromStr(ubInfo[ubCompactPktCntTx], ubCompactPktCntTx)
	ubInfos.UBCommonStats.UbUmocCtphCntTx = hccn.GetIntDataFromStr(ubInfo[ubUmocCtphCntTx], ubUmocCtphCntTx)
	ubInfos.UBCommonStats.UbUmocNtphCntTx = hccn.GetIntDataFromStr(ubInfo[ubUmocNtphCntTx], ubUmocNtphCntTx)
	ubInfos.UBCommonStats.UbMemPktCntTx = hccn.GetIntDataFromStr(ubInfo[ubMemPktCntTx], ubMemPktCntTx)
	ubInfos.UBCommonStats.UnknownPktCntTx = hccn.GetIntDataFromStr(ubInfo[unknownPktCntTx], unknownPktCntTx)
	ubInfos.UBCommonStats.DropIndCntTx = hccn.GetIntDataFromStr(ubInfo[dropIndCntTx], dropIndCntTx)
	ubInfos.UBCommonStats.ErrIndCntTx = hccn.GetIntDataFromStr(ubInfo[errIndCntTx], errIndCntTx)
	ubInfos.UBCommonStats.LpbkIndCntTx = hccn.GetIntDataFromStr(ubInfo[lpbkIndCntTx], lpbkIndCntTx)
	ubInfos.UBCommonStats.OutErrCntTx = hccn.GetIntDataFromStr(ubInfo[outErrCntTx], outErrCntTx)
	ubInfos.UBCommonStats.LengthErrCntTx = hccn.GetIntDataFromStr(ubInfo[lengthErrCntTx], lengthErrCntTx)
	ubInfos.UBCommonStats.TxBusiFlitNum = hccn.GetIntDataFromStr(ubInfo[txBusiFlitNum], txBusiFlitNum)
	ubInfos.UBCommonStats.TxRecvAckFlit = hccn.GetIntDataFromStr(ubInfo[txRecvAckFlit], txRecvAckFlit)

	ubInfos.UBCommonStats.RetryReqSum = hccn.GetIntDataFromStr(ubInfo[retryReqSum], retryReqSum)
	ubInfos.UBCommonStats.RetryAckSum = hccn.GetIntDataFromStr(ubInfo[retryAckSum], retryAckSum)
	ubInfos.UBCommonStats.CrcErrorSum = hccn.GetIntDataFromStr(ubInfo[crcErrorSum], crcErrorSum)
}

func initDesc(ch chan<- *prometheus.Desc, descs []*prometheus.Desc) {
	for _, desc := range descs {
		ch <- desc
	}
}

func buildDesc(dieID, portID int, desc *[]*prometheus.Desc, metricName, help string) {
	colcommon.BuildDescSlice(desc, fmt.Sprint(api.MetricsPrefix, metricName, "_",
		strconv.Itoa(dieID), "_", strconv.Itoa(portID)),
		fmt.Sprint(help, "dieId:", strconv.Itoa(dieID), " portId:", strconv.Itoa(portID)))
}
