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

/*
Package timedetector is used for time dimension detection by comparing data
with itself to identify significantly abnormal data points in time series data.
*/
package timedetector

import "ascend-faultdiag-online/pkg/algo_src/slownode/config"

// AbnormalDetectionConfig 把异常检测需要的参数放到一起
type AbnormalDetectionConfig struct {
	NormalNumber                int
	NSigma                      int
	DegradationPercentage       float64
	NConsecAnomaliesSignifySlow int
}

// DetectAbnormalSend 检测异常的Send算子
func DetectAbnormalSend(
	fileRanks []int,
	alignedData [][]float64,
	conf config.AlgoInputConfig,
	column string) []int {
	var slowSendRanks = []int{}
	/* 对每一张卡的数据进行检测 */
	for index, oneRankHistoryData := range alignedData {
		if index >= len(fileRanks) {
			continue
		}
		isSlow, _ :=
			FirstNPointsShouldBeNormal(oneRankHistoryData, conf, fileRanks[index], column)
		if isSlow != 1 {
			continue
		}
		if index >= 0 && index < len(fileRanks) {
			slowSendRanks = append(slowSendRanks, fileRanks[index])
		}
	}
	return slowSendRanks
}

// DetectAbnormalDomain 检测异常的通信域，当前通信域组中所有卡都是慢的才认为是慢通信域
func DetectAbnormalDomain(
	fileRanks []int,
	alignedData [][]float64,
	conf config.AlgoInputConfig,
	column string) []int {
	var slowDomain = []int{}
	var sumNumberOfSlow int = 0
	/* 对每一张卡的数据进行检测 */
	for index, oneRankHistoryData := range alignedData {
		if index >= len(fileRanks) {
			continue
		}
		isSlow, _ :=
			FirstNPointsShouldBeNormal(oneRankHistoryData, conf, fileRanks[index], column)
		sumNumberOfSlow += isSlow
	}
	/* 当前通信域组中所有卡变慢才认为该通信域组变慢 */
	if sumNumberOfSlow == len(fileRanks) {
		slowDomain = append(slowDomain, fileRanks...)
	} else {
		slowDomain = []int{}
	}
	/* 返回变慢的npu卡ID */
	return slowDomain
}

// DetectionAbnormalCard 不存在tp并行通信域时，检测npu卡组中的慢计算卡
func DetectionAbnormalCard(cards []int,
	cardsData [][]float64,
	conf config.AlgoInputConfig,
	column string) []int {
	slowCalculatedRanks := make([]int, 0)
	length := len(cardsData)
	for i := 0; i < length; i++ {
		alignedData := cardsData[i]
		if i >= len(cards) {
			continue
		}
		isSlow, _ :=
			FirstNPointsShouldBeNormal(alignedData, conf, cards[i], column)
		if isSlow == 1 {
			slowCalculatedRanks = append(slowCalculatedRanks, cards[i])
		}
	}
	return slowCalculatedRanks
}
