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
	"sync"
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

	ubDieIDLabel  = "udie"
	ubPortIDLabel = "port"
)

var (
	ubCardLabel     []string
	ubCardLabelOnce sync.Once
	ubDescOnce      sync.Once

	// rx
	ubIpv4PktCntRxDesc    *prometheus.Desc
	ubIpv6PktCntRxDesc    *prometheus.Desc
	unicIpv4PktCntRxDesc  *prometheus.Desc
	unicIpv6PktCntRxDesc  *prometheus.Desc
	ubCompactPktCntRxDesc *prometheus.Desc
	ubUmocCtphCntRxDesc   *prometheus.Desc
	ubUmocNtphCntRxDesc   *prometheus.Desc
	ubMemPktCntRxDesc     *prometheus.Desc
	unknownPktCntRxDesc   *prometheus.Desc
	dropIndCntRxDesc      *prometheus.Desc
	errIndCntRxDesc       *prometheus.Desc
	toHostPktCntRxDesc    *prometheus.Desc
	toImpPktCntRxDesc     *prometheus.Desc
	toMarPktCntRxDesc     *prometheus.Desc
	toLinkPktCntRxDesc    *prometheus.Desc
	toNocPktCntRxDesc     *prometheus.Desc
	routeErrCntRxDesc     *prometheus.Desc
	outErrCntRxDesc       *prometheus.Desc
	lengthErrCntRxDesc    *prometheus.Desc
	rxBusiFlitNumDesc     *prometheus.Desc
	rxSendAckFlitNumDesc  *prometheus.Desc
	// tx
	ubIpv4PktCntTxDesc    *prometheus.Desc
	ubIpv6PktCntTxDesc    *prometheus.Desc
	unicIpv4PktCntTxDesc  *prometheus.Desc
	unicIpv6PktCntTxDesc  *prometheus.Desc
	ubCompactPktCntTxDesc *prometheus.Desc
	ubUmocCtphCntTxDesc   *prometheus.Desc
	ubUmocNtphCntTxDesc   *prometheus.Desc
	ubMemPktCntTxDesc     *prometheus.Desc
	unknownPktCntTxDesc   *prometheus.Desc
	dropIndCntTxDesc      *prometheus.Desc
	errIndCntTxDesc       *prometheus.Desc
	lpbkIndCntTxDesc      *prometheus.Desc
	outErrCntTxDesc       *prometheus.Desc
	lengthErrCntTxDesc    *prometheus.Desc
	txBusiFlitNumDesc     *prometheus.Desc
	txRecvAckFlitDesc     *prometheus.Desc
	// sum
	retryReqSumDesc *prometheus.Desc
	retryAckSumDesc *prometheus.Desc
	crcErrorSumDesc *prometheus.Desc
	// uboe
	coreMibRxpausepktsDesc *prometheus.Desc
	coreMibTxpausepktsDesc *prometheus.Desc
	coreMibRxpfcpktsDesc   *prometheus.Desc
	coreMibTxpfcpktsDesc   *prometheus.Desc
	coreMibRxbadpktsDesc   *prometheus.Desc
	coreMibTxbadpktsDesc   *prometheus.Desc
	coreMibRxbadoctetsDesc *prometheus.Desc
	coreMibTxbadoctetsDesc *prometheus.Desc
)

func initUBCardLabel() {
	ubCardLabelOnce.Do(func() {
		ubCardLabel = append(ubCardLabel, colcommon.CardLabel...)
		ubCardLabel = append(ubCardLabel, ubDieIDLabel, ubPortIDLabel)
	})
}

func initBuildDesc() {
	ubDescOnce.Do(func() {
		initUBCardLabel()

		// rx
		initBuildDescRx()
		// tx
		initBuildDescTx()
		// sum
		buildUbDesc(&retryReqSumDesc, retryReqSum, "number of retransmission attempts initiated on ub port")
		buildUbDesc(&retryAckSumDesc, retryAckSum, "number of response retransmissions on ub port")
		buildUbDesc(&crcErrorSumDesc, crcErrorSum, "number of crc check errors on ub port")
		// uboe
		buildUbDesc(&coreMibRxpausepktsDesc, coreMibRxpausepkts, "uboe total number of rx pause frames on ub port")
		buildUbDesc(&coreMibTxpausepktsDesc, coreMibTxpausepkts, "uboe total number of tx pause frames on ub port")
		buildUbDesc(&coreMibRxpfcpktsDesc, coreMibRxpfcpkts, "uboe total number of rx pfc frames on ub port")
		buildUbDesc(&coreMibTxpfcpktsDesc, coreMibTxpfcpkts, "uboe total number of tx pfc frames on ub port")
		buildUbDesc(&coreMibRxbadpktsDesc, coreMibRxbadpkts, "uboe total number of rx bad packets on ub port")
		buildUbDesc(&coreMibTxbadpktsDesc, coreMibTxbadpkts, "uboe total number of tx bad packets on ub port")
		buildUbDesc(&coreMibRxbadoctetsDesc, coreMibRxbadoctets, "uboe total number of bytes in rx bad packets on ub port")
		buildUbDesc(&coreMibTxbadoctetsDesc, coreMibTxbadoctets, "uboe total number of bytes in tx bad packets on ub port")
		initUbLegacyDesc()
	})
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
	isSupport := colcommon.DevType == api.Ascend910A5
	logForUnSupportDevice(isSupport, colcommon.DevType, colcommon.GetCacheKey(c), "")
	if isSupport {
		initBuildDesc()
	}
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
	addUbLegacyMetricsDesc(ch)
}

// CollectToCache collect the metric to cache
func (c *UbCollector) CollectToCache(n *colcommon.NpuCollector, chipList []colcommon.HuaWeiAIChip) {
	for _, chip := range chipList {
		ubInfo := collectUbInfo(chip.LogicID)
		c.LocalCache.Store(chip.PhyId, ubCache{
			chip:      chip,
			timestamp: time.Now(),
			ubInfo:    ubInfo,
		})
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
	for i := 0; i < len(ubInfo); i++ {
		if ubInfo[i] == nil {
			continue
		}
		extendedLabel := append(cardLabel, strconv.Itoa(ubInfo[i].Udie), strconv.Itoa(ubInfo[i].Port))
		if ubInfo[i].UBCommonStats != nil {
			// rx
			promUpdateUbRx(ch, timestamp, ubInfo, extendedLabel, i)
			promUpdateUbRxLegacy(ch, timestamp, ubInfo[i], extendedLabel)
			// tx
			promUpdateUbTx(ch, timestamp, ubInfo, extendedLabel, i)
			promUpdateUbTxLegacy(ch, timestamp, ubInfo[i], extendedLabel)
			// sum
			promUpdateUbSum(ch, timestamp, ubInfo, extendedLabel, i)
			promUpdateUbSumLegacy(ch, timestamp, ubInfo[i], extendedLabel)
		}
		if ubInfo[i].UboeExtensions != nil {
			//uboe
			promUpdateUbUboe(ch, timestamp, ubInfo, extendedLabel, i)
			promUpdateUbUboeLegacy(ch, timestamp, ubInfo[i], extendedLabel)
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
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv4PktCntRx, cardLabel, ubIpv4PktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv6PktCntRx, cardLabel, ubIpv6PktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv4PktCntRx, cardLabel, unicIpv4PktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv6PktCntRx, cardLabel, unicIpv6PktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbCompactPktCntRx, cardLabel, ubCompactPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocCtphCntRx, cardLabel, ubUmocCtphCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocNtphCntRx, cardLabel, ubUmocNtphCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbMemPktCntRx, cardLabel, ubMemPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnknownPktCntRx, cardLabel, unknownPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.DropIndCntRx, cardLabel, dropIndCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ErrIndCntRx, cardLabel, errIndCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToHostPktCntRx, cardLabel, toHostPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToImpPktCntRx, cardLabel, toImpPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToMarPktCntRx, cardLabel, toMarPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToLinkPktCntRx, cardLabel, toLinkPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ToNocPktCntRx, cardLabel, toNocPktCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RouteErrCntRx, cardLabel, routeErrCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.OutErrCntRx, cardLabel, outErrCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LengthErrCntRx, cardLabel, lengthErrCntRxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RxBusiFlitNum, cardLabel, rxBusiFlitNumDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RxSendAckFlit, cardLabel, rxSendAckFlitNumDesc)
}

func promUpdateUbTx(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv4PktCntTx, cardLabel, ubIpv4PktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbIpv6PktCntTx, cardLabel, ubIpv6PktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv4PktCntTx, cardLabel, unicIpv4PktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnicIpv6PktCntTx, cardLabel, unicIpv6PktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbCompactPktCntTx, cardLabel, ubCompactPktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocCtphCntTx, cardLabel, ubUmocCtphCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbUmocNtphCntTx, cardLabel, ubUmocNtphCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UbMemPktCntTx, cardLabel, ubMemPktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.UnknownPktCntTx, cardLabel, unknownPktCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.DropIndCntTx, cardLabel, dropIndCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.ErrIndCntTx, cardLabel, errIndCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LpbkIndCntTx, cardLabel, lpbkIndCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.OutErrCntTx, cardLabel, outErrCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.LengthErrCntTx, cardLabel, lengthErrCntTxDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.TxBusiFlitNum, cardLabel, txBusiFlitNumDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.TxRecvAckFlit, cardLabel, txRecvAckFlitDesc)
}

func promUpdateUbSum(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RetryReqSum, cardLabel, retryReqSumDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.RetryAckSum, cardLabel, retryAckSumDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UBCommonStats.CrcErrorSum, cardLabel, crcErrorSumDesc)
}

func promUpdateUbUboe(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo []*common.UBInfo,
	cardLabel []string, i int) {
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxPausePkts, cardLabel, coreMibRxpausepktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxPausePkts, cardLabel, coreMibTxpausepktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxPfcPkts, cardLabel, coreMibRxpfcpktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxPfcPkts, cardLabel, coreMibTxpfcpktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxBadPkts, cardLabel, coreMibRxbadpktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxBadPkts, cardLabel, coreMibTxbadpktsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibRxBadOctets, cardLabel, coreMibRxbadoctetsDesc)
	doUpdateMetric(ch, timestamp, ubInfo[i].UboeExtensions.CoreMibTxBadOctets, cardLabel, coreMibTxbadoctetsDesc)
}

func telegrafUpdateUbInfo(cache ubCache, fieldMap map[string]interface{}) {
	ubInfo := cache.ubInfo
	if ubInfo == nil {
		return
	}
	for i := 0; i < len(ubInfo); i++ {
		if ubInfo[i] == nil {
			continue
		}
		extInfo := fmt.Sprint("_", ubInfo[i].Udie, "_", ubInfo[i].Port)
		if ubInfo[i].UBCommonStats != nil {
			// rx
			telegrafUpdateUbRx(fieldMap, ubInfo, i, extInfo)
			// tx
			telegrafUpdateUbTx(fieldMap, ubInfo, i, extInfo)
			// sum
			telegrafUpdateUbSum(fieldMap, ubInfo, i, extInfo)
		}
		if ubInfo[i].UboeExtensions != nil {
			//uboe
			telegrafUpdateUbUboe(fieldMap, ubInfo, i, extInfo)
		}
	}
}

func telegrafUpdateUbRx(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int, extInfo string) {
	doUpdateTelegraf(fieldMap, ubIpv4PktCntRxDesc, ubInfo[i].UBCommonStats.UbIpv4PktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, ubIpv6PktCntRxDesc, ubInfo[i].UBCommonStats.UbIpv6PktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, unicIpv4PktCntRxDesc, ubInfo[i].UBCommonStats.UnicIpv4PktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, unicIpv6PktCntRxDesc, ubInfo[i].UBCommonStats.UnicIpv6PktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, ubCompactPktCntRxDesc, ubInfo[i].UBCommonStats.UbCompactPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, ubUmocCtphCntRxDesc, ubInfo[i].UBCommonStats.UbUmocCtphCntRx, extInfo)
	doUpdateTelegraf(fieldMap, ubUmocNtphCntRxDesc, ubInfo[i].UBCommonStats.UbUmocNtphCntRx, extInfo)
	doUpdateTelegraf(fieldMap, ubMemPktCntRxDesc, ubInfo[i].UBCommonStats.UbMemPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, unknownPktCntRxDesc, ubInfo[i].UBCommonStats.UnknownPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, dropIndCntRxDesc, ubInfo[i].UBCommonStats.DropIndCntRx, extInfo)
	doUpdateTelegraf(fieldMap, errIndCntRxDesc, ubInfo[i].UBCommonStats.ErrIndCntRx, extInfo)
	doUpdateTelegraf(fieldMap, toHostPktCntRxDesc, ubInfo[i].UBCommonStats.ToHostPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, toImpPktCntRxDesc, ubInfo[i].UBCommonStats.ToImpPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, toMarPktCntRxDesc, ubInfo[i].UBCommonStats.ToMarPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, toLinkPktCntRxDesc, ubInfo[i].UBCommonStats.ToLinkPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, toNocPktCntRxDesc, ubInfo[i].UBCommonStats.ToNocPktCntRx, extInfo)
	doUpdateTelegraf(fieldMap, routeErrCntRxDesc, ubInfo[i].UBCommonStats.RouteErrCntRx, extInfo)
	doUpdateTelegraf(fieldMap, outErrCntRxDesc, ubInfo[i].UBCommonStats.OutErrCntRx, extInfo)
	doUpdateTelegraf(fieldMap, lengthErrCntRxDesc, ubInfo[i].UBCommonStats.LengthErrCntRx, extInfo)
	doUpdateTelegraf(fieldMap, rxBusiFlitNumDesc, ubInfo[i].UBCommonStats.RxBusiFlitNum, extInfo)
	doUpdateTelegraf(fieldMap, rxSendAckFlitNumDesc, ubInfo[i].UBCommonStats.RxSendAckFlit, extInfo)
}

func telegrafUpdateUbTx(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int, extInfo string) {
	doUpdateTelegraf(fieldMap, ubIpv4PktCntTxDesc, ubInfo[i].UBCommonStats.UbIpv4PktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, ubIpv6PktCntTxDesc, ubInfo[i].UBCommonStats.UbIpv6PktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, unicIpv4PktCntTxDesc, ubInfo[i].UBCommonStats.UnicIpv4PktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, unicIpv6PktCntTxDesc, ubInfo[i].UBCommonStats.UnicIpv6PktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, ubCompactPktCntTxDesc, ubInfo[i].UBCommonStats.UbCompactPktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, ubUmocCtphCntTxDesc, ubInfo[i].UBCommonStats.UbUmocCtphCntTx, extInfo)
	doUpdateTelegraf(fieldMap, ubUmocNtphCntTxDesc, ubInfo[i].UBCommonStats.UbUmocNtphCntTx, extInfo)
	doUpdateTelegraf(fieldMap, ubMemPktCntTxDesc, ubInfo[i].UBCommonStats.UbMemPktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, unknownPktCntTxDesc, ubInfo[i].UBCommonStats.UnknownPktCntTx, extInfo)
	doUpdateTelegraf(fieldMap, dropIndCntTxDesc, ubInfo[i].UBCommonStats.DropIndCntTx, extInfo)
	doUpdateTelegraf(fieldMap, errIndCntTxDesc, ubInfo[i].UBCommonStats.ErrIndCntTx, extInfo)
	doUpdateTelegraf(fieldMap, lpbkIndCntTxDesc, ubInfo[i].UBCommonStats.LpbkIndCntTx, extInfo)
	doUpdateTelegraf(fieldMap, outErrCntTxDesc, ubInfo[i].UBCommonStats.OutErrCntTx, extInfo)
	doUpdateTelegraf(fieldMap, lengthErrCntTxDesc, ubInfo[i].UBCommonStats.LengthErrCntTx, extInfo)
	doUpdateTelegraf(fieldMap, txBusiFlitNumDesc, ubInfo[i].UBCommonStats.TxBusiFlitNum, extInfo)
	doUpdateTelegraf(fieldMap, txRecvAckFlitDesc, ubInfo[i].UBCommonStats.TxRecvAckFlit, extInfo)
}

func telegrafUpdateUbSum(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int, extInfo string) {
	doUpdateTelegraf(fieldMap, retryReqSumDesc, ubInfo[i].UBCommonStats.RetryReqSum, extInfo)
	doUpdateTelegraf(fieldMap, retryAckSumDesc, ubInfo[i].UBCommonStats.RetryAckSum, extInfo)
	doUpdateTelegraf(fieldMap, crcErrorSumDesc, ubInfo[i].UBCommonStats.CrcErrorSum, extInfo)
}

func telegrafUpdateUbUboe(fieldMap map[string]interface{}, ubInfo []*common.UBInfo, i int, extInfo string) {
	doUpdateTelegraf(fieldMap, coreMibRxpausepktsDesc, ubInfo[i].UboeExtensions.CoreMibRxPausePkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibTxpausepktsDesc, ubInfo[i].UboeExtensions.CoreMibTxPausePkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibRxpfcpktsDesc, ubInfo[i].UboeExtensions.CoreMibRxPfcPkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibTxpfcpktsDesc, ubInfo[i].UboeExtensions.CoreMibTxPfcPkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibRxbadpktsDesc, ubInfo[i].UboeExtensions.CoreMibRxBadPkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibTxbadpktsDesc, ubInfo[i].UboeExtensions.CoreMibTxBadPkts, extInfo)
	doUpdateTelegraf(fieldMap, coreMibRxbadoctetsDesc, ubInfo[i].UboeExtensions.CoreMibRxBadOctets, extInfo)
	doUpdateTelegraf(fieldMap, coreMibTxbadoctetsDesc, ubInfo[i].UboeExtensions.CoreMibTxBadOctets, extInfo)
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

func initBuildDescRx() {
	buildUbDesc(&ubIpv4PktCntRxDesc, ubIpv4PktCntRx, "number of ipv4 ub packets received by rx on ub port")
	buildUbDesc(&ubIpv6PktCntRxDesc, ubIpv6PktCntRx, "number of ipv6 ub packets received by rx on ub port")
	buildUbDesc(&unicIpv4PktCntRxDesc, unicIpv4PktCntRx, "number of ipv4 unic packets received by rx on ub port")
	buildUbDesc(&unicIpv6PktCntRxDesc, unicIpv6PktCntRx, "number of ipv6 unic packets received by rx on ub port")
	buildUbDesc(&ubCompactPktCntRxDesc, ubCompactPktCntRx, "number of cfg6 packets received by rx on ub port")
	buildUbDesc(&ubUmocCtphCntRxDesc, ubUmocCtphCntRx, "number of cfg7 clan packets received by rx on ub port")
	buildUbDesc(&ubUmocNtphCntRxDesc, ubUmocNtphCntRx, "number of cfg7 not clan packets received by rx on ub port")
	buildUbDesc(&ubMemPktCntRxDesc, ubMemPktCntRx, "number of ub mem packets received by rx on ub port")
	buildUbDesc(&unknownPktCntRxDesc, unknownPktCntRx, "number of unknown packets received by rx on ub port")
	buildUbDesc(&dropIndCntRxDesc, dropIndCntRx, "number of packet with drop_ind received by rx on ub port")
	buildUbDesc(&errIndCntRxDesc, errIndCntRx, "number of err packets received by rx on ub port")
	buildUbDesc(&toHostPktCntRxDesc, toHostPktCntRx, "number of landed packets after routing on the rx ub port")
	buildUbDesc(&toImpPktCntRxDesc, toImpPktCntRx, "number of landed enumeration configuration and management packets after routing on the rx ub port")
	buildUbDesc(&toMarPktCntRxDesc, toMarPktCntRx, "number of landed ub memory packets after routing on the rx ub port")
	buildUbDesc(&toLinkPktCntRxDesc, toLinkPktCntRx, "number of packets forward from the rx to the tx of the same port after routing on ub port")
	buildUbDesc(&toNocPktCntRxDesc, toNocPktCntRx, "number of p2p packets received on the rx after routing on ub port")
	buildUbDesc(&routeErrCntRxDesc, routeErrCntRx, "number of packets with routing lookup errors after processing received on the rx ub port")
	buildUbDesc(&outErrCntRxDesc, outErrCntRx, "total number of erroneous packets after validation of packets received on the rx ub port")
	buildUbDesc(&lengthErrCntRxDesc, lengthErrCntRx, "number of packets with length errors after validation of packets received on the rx ub port")
	buildUbDesc(&rxBusiFlitNumDesc, rxBusiFlitNum, "number of flits of service packets received from the mac on the rx ub port")
	buildUbDesc(&rxSendAckFlitNumDesc, rxSendAckFlit, "cumulative number of acks released to the peer on the rx ub port")
}

func initBuildDescTx() {
	buildUbDesc(&ubIpv4PktCntTxDesc, ubIpv4PktCntTx, "number of ipv4 ub packets sent by tx on ub port")
	buildUbDesc(&ubIpv6PktCntTxDesc, ubIpv6PktCntTx, "number of ipv6 ub packets sent by tx on ub port")
	buildUbDesc(&unicIpv4PktCntTxDesc, unicIpv4PktCntTx, "number of ipv4 unic packets sent by tx on ub port")
	buildUbDesc(&unicIpv6PktCntTxDesc, unicIpv6PktCntTx, "number of ipv6 unic packets sent by tx on ub port")
	buildUbDesc(&ubCompactPktCntTxDesc, ubCompactPktCntTx, "number of cfg6 packets sent by tx on ub port")
	buildUbDesc(&ubUmocCtphCntTxDesc, ubUmocCtphCntTx, "number of cfg7 clan packets sent by tx on ub port")
	buildUbDesc(&ubUmocNtphCntTxDesc, ubUmocNtphCntTx, "number of cfg7 not clan packets sent by tx on ub port")
	buildUbDesc(&ubMemPktCntTxDesc, ubMemPktCntTx, "number of ub mem packets sent by tx on ub port")
	buildUbDesc(&unknownPktCntTxDesc, unknownPktCntTx, "number of unknown packets sent by tx on ub port")
	buildUbDesc(&dropIndCntTxDesc, dropIndCntTx, "number of packet with drop_ind sent by tx on ub port")
	buildUbDesc(&errIndCntTxDesc, errIndCntTx, "number of err packets sent by tx on ub port")
	buildUbDesc(&lpbkIndCntTxDesc, lpbkIndCntTx, "number of packets looped back at nl by tx on ub port")
	buildUbDesc(&outErrCntTxDesc, outErrCntTx, "total number of erroneous packets after validation of packets sent on the tx ub port")
	buildUbDesc(&lengthErrCntTxDesc, lengthErrCntTx, "number of packets with length errors after validation of packets sent on the tx ub port")
	buildUbDesc(&txBusiFlitNumDesc, txBusiFlitNum, "number of flits of service packets sent from the mac on the tx ub port")
	buildUbDesc(&txRecvAckFlitDesc, txRecvAckFlit, "cumulative number of acks released to the peer on the tx ub port")
}

func collectUbInfo(logicID int32) []*common.UBInfo {
	var newUbInfos []*common.UBInfo
	// udie only has 0 and 1, fixed order
	dieIDs := []int{0, 1}
	for _, dieID := range dieIDs {
		portIDs, ok := colcommon.NpuDevPortInfos.GetPortMap()[dieID]
		if !ok || len(portIDs) == 0 {
			continue
		}
		for _, port := range portIDs {
			newUbInfos = append(newUbInfos, getUBStatInfo(logicID, dieID, port.PortID))
		}
	}
	return newUbInfos
}

func getUBStatInfo(logicID int32, uDieID, portID int) *common.UBInfo {
	ubInfos := common.UBInfo{
		UBCommonStats:  &common.UBCommonStats{},
		UboeExtensions: &common.UBOEExtensions{},
		Udie:           uDieID,
		Port:           portID,
	}
	ubInfo, err := hccn.GetNPUUbStatInfo(logicID, int32(uDieID), int32(portID))
	if err != nil {
		logWarnMetricsWithLimit(fmt.Sprint(colcommon.DomainForUb, uDieID, portID), logicID, uDieID, portID, err)
		return nil
	}
	if result, err := strconv.Atoi(ubInfo[isUboePort]); err == nil && result == isUboe {
		hwlog.RunLog.Debugf("logicID:%v ,UdieID:%v, portID:%v is uboe port", logicID, uDieID, portID)
		convertUboeExtensions(&ubInfos, ubInfo)
	} else {
		ubInfos.UboeExtensions = nil
	}
	convertUBCommonStats(&ubInfos, ubInfo)
	hwlog.ResetErrCnt(fmt.Sprint(colcommon.DomainForUb, uDieID, portID), logicID)
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

func initDesc(ch chan<- *prometheus.Desc, desc *prometheus.Desc) {
	if desc != nil {
		ch <- desc
	}
}

func buildUbDesc(desc **prometheus.Desc, metricName, help string) {
	*desc = colcommon.BuildDescWithLabel(fmt.Sprint(api.MetricsPrefix, metricName), help, ubCardLabel)
}
