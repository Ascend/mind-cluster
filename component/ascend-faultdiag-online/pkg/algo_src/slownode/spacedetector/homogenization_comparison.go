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

// Package spacedetector performs space dimension detection by homogenizing the data
package spacedetector

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
)

// transpose 对矩阵数据进行转置
func transpose(data [][]float64) [][]float64 {
	rows := len(data)
	cols := len(data[0])
	transposed := make([][]float64, cols)
	for i := 0; i < cols; i++ {
		transposed[i] = make([]float64, rows)
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			transposed[j][i] = data[i][j]
		}
	}
	return transposed
}

// HomogenizationComparisonFunc 空间维度检测，使用聚类算法找出TP域内的异常数据点
func HomogenizationComparisonFunc(
	sndConfig config.AlgoInputConfig,
	fileRanks []int,
	alignedData [][]float64) []int {
	oneTpAbnormalGlobalRanks := make([]int, 0)
	/* nConsecAnomalies 表示最低需要多少个数据进行检测 */
	nConsecAnomalies := sndConfig.NconsecAnomaliesSignifySlow
	/* 判断是否分为两个类的距离阈值 */
	clusterMeanDistance := sndConfig.ClusterMeanDistance
	if len(alignedData) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]alignedData is nil or empty")
		return oneTpAbnormalGlobalRanks
	}

	for i, zpData := range alignedData {
		if len(zpData) < nConsecAnomalies {
			hwlog.RunLog.Warnf("[SLOWNODE ALGO]data at index %d is not enough for detection, length: %d, required: %d",
				i, len(zpData), nConsecAnomalies)
			return oneTpAbnormalGlobalRanks
		}
	}
	var alignedNCData [][]float64
	/* 只取最后几个连续的数据进行聚类 */
	for _, zpData := range alignedData {
		var row = []float64{}
		row = append(row, zpData[len(zpData)-nConsecAnomalies:]...)
		alignedNCData = append(alignedNCData, row)
	}
	/* 转置矩阵 */
	transposed := transpose(alignedNCData)
	var tpAbnormalIndexss = [][]int{}
	/* 转置后每一个[]int{} 为多张npu卡同字段列同一个step的时延数据 */
	for _, value := range transposed {
		tpAbnormalIndexs := recurseDimensionalClustering(value, clusterMeanDistance)
		tpAbnormalIndexss = append(tpAbnormalIndexss, tpAbnormalIndexs)
	}
	/* 找出异常卡 */
	abnormalIndexs, isAllCommon := findCommonAndCheck(tpAbnormalIndexss)
	if !isAllCommon {
		return oneTpAbnormalGlobalRanks
	}
	for _, value := range abnormalIndexs {
		if value >= 0 && value < len(fileRanks) {
			oneTpAbnormalGlobalRanks = append(oneTpAbnormalGlobalRanks, fileRanks[value])
		}
	}
	return oneTpAbnormalGlobalRanks
}
