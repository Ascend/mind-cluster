/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2021-2024. All rights reserved.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package collector for Prometheus
package collector

import (
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	pcieBwLabel = append(cardLabel, pcieBwType)
)

func pcieBwLabelVal(cardLabels []string, pcieBwType string) []string {
	return append(cardLabels, pcieBwType)
}

func metricWithPcieBw(labelsVal []string, metrics *prometheus.Desc, val float64, valType string) prometheus.Metric {
	return prometheus.MustNewConstMetric(metrics, prometheus.GaugeValue, val, pcieBwLabelVal(labelsVal, valType)...)
}

func updatePcieBwInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cardLabelsVal []string) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	chipInfo := common.DeepCopyChipInfo(chip.ChipIfo)
	boardInfo := common.DeepCopyBoardInfo(chip.BoardInfo)
	vDevActivityInfo := common.DeepCopyVDevActivityInfo(chip.VDevActivityInfo)

	if ((chipInfo != nil && !common.Is910BChip(chipInfo.Name)) &&
		(boardInfo != nil && !common.Is910A3Chip(boardInfo.BoardId))) ||
		(vDevActivityInfo != nil && common.IsValidVDevID(vDevActivityInfo.VDevID)) {
		hwlog.RunLog.Debug("only 910B or Atlas 900 A3 SuperPod supports pcie info query")
		return
	}

	updateAvgPcieBwInfo(ch, npu, chip, cardLabelsVal)
	updateMinPcieBwInfo(ch, npu, chip, cardLabelsVal)
	updateMaxPcieBwInfo(ch, npu, chip, cardLabelsVal)
}

func updateAvgPcieBwInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cardLabelsVal []string) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	pcieBwInfo := common.DeepCopyPcieBwInfo(chip.PcieBwInfo)
	if !validate(ch, npu, chip, pcieBwInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateAvgPcieBwInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxPBW, float64(pcieBwInfo.PcieTxPBw.PcieAvgBw), avgPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxNpBW, float64(pcieBwInfo.PcieTxNPBw.PcieAvgBw), avgPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxCplBW, float64(pcieBwInfo.PcieTxCPLBw.PcieAvgBw), avgPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxPBW, float64(pcieBwInfo.PcieRxPBw.PcieAvgBw), avgPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxNpBW, float64(pcieBwInfo.PcieRxNPBw.PcieAvgBw), avgPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxCplBW, float64(pcieBwInfo.PcieRxCPLBw.PcieAvgBw), avgPcieBw))
}

func updateMinPcieBwInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cardLabelsVal []string) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	pcieBwInfo := common.DeepCopyPcieBwInfo(chip.PcieBwInfo)
	if !validate(ch, npu, chip, pcieBwInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateMinPcieBwInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxPBW, float64(pcieBwInfo.PcieTxPBw.PcieMinBw), minPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxNpBW, float64(pcieBwInfo.PcieTxNPBw.PcieMinBw), minPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxCplBW, float64(pcieBwInfo.PcieTxCPLBw.PcieMinBw), minPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxPBW, float64(pcieBwInfo.PcieRxPBw.PcieMinBw), minPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxNpBW, float64(pcieBwInfo.PcieRxNPBw.PcieMinBw), minPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxCplBW, float64(pcieBwInfo.PcieRxCPLBw.PcieMinBw), minPcieBw))
}

func updateMaxPcieBwInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, cardLabelsVal []string) {
	// use deep copy to prevent the pointer structure from being assigned nil by other goroutine
	pcieBwInfo := common.DeepCopyPcieBwInfo(chip.PcieBwInfo)
	if !validate(ch, npu, chip, pcieBwInfo) {
		hwlog.RunLog.Warnf("Invalid param in function updateMaxPcieBwInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxPBW, float64(pcieBwInfo.PcieTxPBw.PcieMaxBw), maxPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxNpBW, float64(pcieBwInfo.PcieTxNPBw.PcieMaxBw), maxPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescTxCplBW, float64(pcieBwInfo.PcieTxCPLBw.PcieMaxBw), maxPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxPBW, float64(pcieBwInfo.PcieRxPBw.PcieMaxBw), maxPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxNpBW, float64(pcieBwInfo.PcieRxNPBw.PcieMaxBw), maxPcieBw))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		metricWithPcieBw(cardLabelsVal, npuChipInfoDescRxCplBW, float64(pcieBwInfo.PcieRxCPLBw.PcieMaxBw), maxPcieBw))
}
