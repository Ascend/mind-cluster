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
// This file contains legacy metrics with _X_Y suffix for Atlas 350 backward compatibility.
// It will be removed after the compatibility period ends.
package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/api"
	"ascend-common/devmanager/common"
	colcommon "huawei.com/npu-exporter/v6/collector/common"
)

var ubLegacyDescMap = make(map[string]map[string]*prometheus.Desc)

func buildLegacyDescMap(baseName string, help string) map[string]*prometheus.Desc {
	descs := make(map[string]*prometheus.Desc)
	dieIDs := []int{0, 1}
	for _, dieID := range dieIDs {
		portIDs, ok := colcommon.NpuDevPortInfos.GetPortMap()[dieID]
		if !ok || len(portIDs) == 0 {
			continue
		}
		for _, port := range portIDs {
			name := fmt.Sprint(api.MetricsPrefix, baseName, "_", dieID, "_", port.PortID)
			key := fmt.Sprintf("%d_%d", dieID, port.PortID)
			descs[key] = colcommon.BuildDescWithLabel(name, help, colcommon.CardLabel)
		}
	}
	return descs
}

func buildLegacyDescSlice(baseName string, help string) []*prometheus.Desc {
	var descs []*prometheus.Desc
	dieIDs := []int{0, 1}
	for _, dieID := range dieIDs {
		portIDs, ok := colcommon.NpuDevPortInfos.GetPortMap()[dieID]
		if !ok || len(portIDs) == 0 {
			continue
		}
		for _, port := range portIDs {
			name := fmt.Sprint(api.MetricsPrefix, baseName, "_", dieID, "_", port.PortID)
			descs = append(descs, colcommon.BuildDescWithLabel(name, help, colcommon.CardLabel))
		}
	}
	return descs
}

func initUbLegacyDesc() {
	ubLegacyDescMap[ubIpv4PktCntRx] = buildLegacyDescMap(ubIpv4PktCntRx, "rx side IPv4 UB packet count on ub port")
	ubLegacyDescMap[ubIpv6PktCntRx] = buildLegacyDescMap(ubIpv6PktCntRx, "rx side IPv6 UB packet count on ub port")
	ubLegacyDescMap[unicIpv4PktCntRx] = buildLegacyDescMap(unicIpv4PktCntRx, "rx side IPv4 UNIC packet count on ub port")
	ubLegacyDescMap[unicIpv6PktCntRx] = buildLegacyDescMap(unicIpv6PktCntRx, "rx side IPv6 UNIC packet count on ub port")
	ubLegacyDescMap[ubCompactPktCntRx] = buildLegacyDescMap(ubCompactPktCntRx, "rx side CFG6 packet count on ub port")
	ubLegacyDescMap[ubUmocCtphCntRx] = buildLegacyDescMap(ubUmocCtphCntRx, "rx side CFG7 CLAN packet count on ub port")
	ubLegacyDescMap[ubUmocNtphCntRx] = buildLegacyDescMap(ubUmocNtphCntRx, "rx side CFG7 non-CLAN packet count on ub port")
	ubLegacyDescMap[ubMemPktCntRx] = buildLegacyDescMap(ubMemPktCntRx, "rx side UB mem packet count on ub port")
	ubLegacyDescMap[unknownPktCntRx] = buildLegacyDescMap(unknownPktCntRx, "rx side unknown packet count on ub port")
	ubLegacyDescMap[dropIndCntRx] = buildLegacyDescMap(dropIndCntRx, "rx side drop_ind packet count on ub port")
	ubLegacyDescMap[errIndCntRx] = buildLegacyDescMap(errIndCntRx, "rx side ERR packet count on ub port")
	ubLegacyDescMap[toHostPktCntRx] = buildLegacyDescMap(toHostPktCntRx, "rx side to-host packet count on ub port")
	ubLegacyDescMap[toImpPktCntRx] = buildLegacyDescMap(toImpPktCntRx, "rx side to-imp packet count on ub port")
	ubLegacyDescMap[toMarPktCntRx] = buildLegacyDescMap(toMarPktCntRx, "rx side to-mar packet count on ub port")
	ubLegacyDescMap[toLinkPktCntRx] = buildLegacyDescMap(toLinkPktCntRx, "rx side to-link packet count on ub port")
	ubLegacyDescMap[toNocPktCntRx] = buildLegacyDescMap(toNocPktCntRx, "rx side to-noc packet count on ub port")
	ubLegacyDescMap[routeErrCntRx] = buildLegacyDescMap(routeErrCntRx, "rx side route error count on ub port")
	ubLegacyDescMap[outErrCntRx] = buildLegacyDescMap(outErrCntRx, "rx side output error count on ub port")
	ubLegacyDescMap[lengthErrCntRx] = buildLegacyDescMap(lengthErrCntRx, "rx side length error count on ub port")
	ubLegacyDescMap[rxBusiFlitNum] = buildLegacyDescMap(rxBusiFlitNum, "rx busi flit number on ub port")
	ubLegacyDescMap[rxSendAckFlit] = buildLegacyDescMap(rxSendAckFlit, "rx send ack flit on ub port")
	ubLegacyDescMap[ubIpv4PktCntTx] = buildLegacyDescMap(ubIpv4PktCntTx, "tx side IPv4 UB packet count on ub port")
	ubLegacyDescMap[ubIpv6PktCntTx] = buildLegacyDescMap(ubIpv6PktCntTx, "tx side IPv6 UB packet count on ub port")
	ubLegacyDescMap[unicIpv4PktCntTx] = buildLegacyDescMap(unicIpv4PktCntTx, "tx side IPv4 UNIC packet count on ub port")
	ubLegacyDescMap[unicIpv6PktCntTx] = buildLegacyDescMap(unicIpv6PktCntTx, "tx side IPv6 UNIC packet count on ub port")
	ubLegacyDescMap[ubCompactPktCntTx] = buildLegacyDescMap(ubCompactPktCntTx, "tx side CFG6 packet count on ub port")
	ubLegacyDescMap[ubUmocCtphCntTx] = buildLegacyDescMap(ubUmocCtphCntTx, "tx side CFG7 CLAN packet count on ub port")
	ubLegacyDescMap[ubUmocNtphCntTx] = buildLegacyDescMap(ubUmocNtphCntTx, "tx side CFG7 non-CLAN packet count on ub port")
	ubLegacyDescMap[ubMemPktCntTx] = buildLegacyDescMap(ubMemPktCntTx, "tx side UB mem packet count on ub port")
	ubLegacyDescMap[unknownPktCntTx] = buildLegacyDescMap(unknownPktCntTx, "tx side unknown packet count on ub port")
	ubLegacyDescMap[dropIndCntTx] = buildLegacyDescMap(dropIndCntTx, "tx side drop_ind packet count on ub port")
	ubLegacyDescMap[errIndCntTx] = buildLegacyDescMap(errIndCntTx, "tx side ERR packet count on ub port")
	ubLegacyDescMap[lpbkIndCntTx] = buildLegacyDescMap(lpbkIndCntTx, "tx side loopback packet count on ub port")
	ubLegacyDescMap[outErrCntTx] = buildLegacyDescMap(outErrCntTx, "tx side output error count on ub port")
	ubLegacyDescMap[lengthErrCntTx] = buildLegacyDescMap(lengthErrCntTx, "tx side length error count on ub port")
	ubLegacyDescMap[txBusiFlitNum] = buildLegacyDescMap(txBusiFlitNum, "tx busi flit number on ub port")
	ubLegacyDescMap[txRecvAckFlit] = buildLegacyDescMap(txRecvAckFlit, "tx recv ack flit on ub port")
	ubLegacyDescMap[retryReqSum] = buildLegacyDescMap(retryReqSum, "number of retransmission attempts initiated on ub port")
	ubLegacyDescMap[retryAckSum] = buildLegacyDescMap(retryAckSum, "number of response retransmissions on ub port")
	ubLegacyDescMap[crcErrorSum] = buildLegacyDescMap(crcErrorSum, "number of crc check errors on ub port")
	ubLegacyDescMap[coreMibRxpausepkts] = buildLegacyDescMap(coreMibRxpausepkts, "uboe total number of rx pause frames on ub port")
	ubLegacyDescMap[coreMibTxpausepkts] = buildLegacyDescMap(coreMibTxpausepkts, "uboe total number of tx pause frames on ub port")
	ubLegacyDescMap[coreMibRxpfcpkts] = buildLegacyDescMap(coreMibRxpfcpkts, "uboe total number of rx pfc frames on ub port")
	ubLegacyDescMap[coreMibTxpfcpkts] = buildLegacyDescMap(coreMibTxpfcpkts, "uboe total number of tx pfc frames on ub port")
	ubLegacyDescMap[coreMibRxbadpkts] = buildLegacyDescMap(coreMibRxbadpkts, "uboe total number of rx bad packets on ub port")
	ubLegacyDescMap[coreMibTxbadpkts] = buildLegacyDescMap(coreMibTxbadpkts, "uboe total number of tx bad packets on ub port")
	ubLegacyDescMap[coreMibRxbadoctets] = buildLegacyDescMap(coreMibRxbadoctets, "uboe total number of bytes in rx bad packets on ub port")
	ubLegacyDescMap[coreMibTxbadoctets] = buildLegacyDescMap(coreMibTxbadoctets, "uboe total number of bytes in tx bad packets on ub port")
}

func tryEmitUbLegacyMetric(ch chan<- prometheus.Metric, timestamp time.Time,
	value interface{}, extendedLabel []string, legacyKey string, udie, port int) {
	if descs, ok := ubLegacyDescMap[legacyKey]; ok {
		key := fmt.Sprintf("%d_%d", udie, port)
		if desc, ok := descs[key]; ok {
			doUpdateMetric(ch, timestamp, value, extendedLabel[:len(extendedLabel)-2], desc)
		}
	}
}

func promUpdateUbRxLegacy(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo *common.UBInfo,
	cardLabel []string) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	udie, port := ubInfo.Udie, ubInfo.Port
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbIpv4PktCntRx, cardLabel, ubIpv4PktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbIpv6PktCntRx, cardLabel, ubIpv6PktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnicIpv4PktCntRx, cardLabel, unicIpv4PktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnicIpv6PktCntRx, cardLabel, unicIpv6PktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbCompactPktCntRx, cardLabel, ubCompactPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbUmocCtphCntRx, cardLabel, ubUmocCtphCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbUmocNtphCntRx, cardLabel, ubUmocNtphCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbMemPktCntRx, cardLabel, ubMemPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnknownPktCntRx, cardLabel, unknownPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.DropIndCntRx, cardLabel, dropIndCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ErrIndCntRx, cardLabel, errIndCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ToHostPktCntRx, cardLabel, toHostPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ToImpPktCntRx, cardLabel, toImpPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ToMarPktCntRx, cardLabel, toMarPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ToLinkPktCntRx, cardLabel, toLinkPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ToNocPktCntRx, cardLabel, toNocPktCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.RouteErrCntRx, cardLabel, routeErrCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.OutErrCntRx, cardLabel, outErrCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.LengthErrCntRx, cardLabel, lengthErrCntRx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.RxBusiFlitNum, cardLabel, rxBusiFlitNum, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.RxSendAckFlit, cardLabel, rxSendAckFlit, udie, port)
}

func promUpdateUbTxLegacy(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo *common.UBInfo,
	cardLabel []string) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	udie, port := ubInfo.Udie, ubInfo.Port
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbIpv4PktCntTx, cardLabel, ubIpv4PktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbIpv6PktCntTx, cardLabel, ubIpv6PktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnicIpv4PktCntTx, cardLabel, unicIpv4PktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnicIpv6PktCntTx, cardLabel, unicIpv6PktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbCompactPktCntTx, cardLabel, ubCompactPktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbUmocCtphCntTx, cardLabel, ubUmocCtphCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbUmocNtphCntTx, cardLabel, ubUmocNtphCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UbMemPktCntTx, cardLabel, ubMemPktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.UnknownPktCntTx, cardLabel, unknownPktCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.DropIndCntTx, cardLabel, dropIndCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.ErrIndCntTx, cardLabel, errIndCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.LpbkIndCntTx, cardLabel, lpbkIndCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.OutErrCntTx, cardLabel, outErrCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.LengthErrCntTx, cardLabel, lengthErrCntTx, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.TxBusiFlitNum, cardLabel, txBusiFlitNum, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.TxRecvAckFlit, cardLabel, txRecvAckFlit, udie, port)
}

func promUpdateUbSumLegacy(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo *common.UBInfo,
	cardLabel []string) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	udie, port := ubInfo.Udie, ubInfo.Port
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.RetryReqSum, cardLabel, retryReqSum, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.RetryAckSum, cardLabel, retryAckSum, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UBCommonStats.CrcErrorSum, cardLabel, crcErrorSum, udie, port)
}

func promUpdateUbUboeLegacy(ch chan<- prometheus.Metric, timestamp time.Time, ubInfo *common.UBInfo,
	cardLabel []string) {
	if !colcommon.EnableLegacyMetrics {
		return
	}
	udie, port := ubInfo.Udie, ubInfo.Port
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibRxPausePkts, cardLabel, coreMibRxpausepkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibTxPausePkts, cardLabel, coreMibTxpausepkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibRxPfcPkts, cardLabel, coreMibRxpfcpkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibTxPfcPkts, cardLabel, coreMibTxpfcpkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibRxBadPkts, cardLabel, coreMibRxbadpkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibTxBadPkts, cardLabel, coreMibTxbadpkts, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibRxBadOctets, cardLabel, coreMibRxbadoctets, udie, port)
	tryEmitUbLegacyMetric(ch, timestamp, ubInfo.UboeExtensions.CoreMibTxBadOctets, cardLabel, coreMibTxbadoctets, udie, port)
}

func addUbLegacyMetricsDesc(ch chan<- *prometheus.Desc) {
	// Send legacy Desc for backward compatibility
	if colcommon.EnableLegacyMetrics {
		for _, descMap := range ubLegacyDescMap {
			for _, desc := range descMap {
				ch <- desc
			}
		}
	}
}
