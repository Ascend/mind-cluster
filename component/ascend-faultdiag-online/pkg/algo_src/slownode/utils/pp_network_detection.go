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
Package utils is used for file reading and writing, as well as data processing.
*/
package utils

import (
	"fmt"

	"ascend-common/common-utils/hwlog"
)

// PPNetworkDetection 慢网络卡检测
func PPNetworkDetection(slowSendRanks []int, PPrankss [][]int) []int {
	var res = []int{}

	// 遍历每一个PP通信域
	for _, PPranks := range PPrankss {

		length := len(PPranks)
		if length <= 0 {
			continue
		}

		// 在此PP通信域内，初始化连接数和坏连接数
		connectNumbers, badConnectNumbers, err := initializeConnectNumbers(length)
		if err != nil {
			continue
		}

		// 基于慢send的结果，更新坏链路
		updateBadConnectNumbers(slowSendRanks, PPranks, badConnectNumbers)

		// 计算 ppValue 并打印
		ppValue := calculatePPValue(connectNumbers, badConnectNumbers)
		hwlog.RunLog.Infof("ppValue values: %v", ppValue)

		// 将 此PP域内 检测出的坏卡添加到 res 中
		res = append(res, findAbnormalNodes(PPranks, ppValue)...)

	}
	return res
}

// 初始化连接数和坏连接数
func initializeConnectNumbers(length int) ([]int, []int, error) {
	if length <= 0 {
		hwlog.RunLog.Info("length is zero")
		return nil, nil, fmt.Errorf("PPranks len is zero")
	}
	connectNumbers := make([]int, length)
	badConnectNumbers := make([]int, length)

	for index := range connectNumbers {
		// 将Rank所有数据设置左右两个连接
		connectNumbers[index] = 2 // 非两端设置值为2
	}
	connectNumbers[0] = 1        // 两端设置值为1
	connectNumbers[length-1] = 1 // 两端设置值为1

	return connectNumbers, badConnectNumbers, nil
}

// 更新坏连接数，对慢Send两边的Rank进行加1
func updateBadConnectNumbers(slowSendRanks, PPranks []int, badConnectNumbers []int) {
	for _, abnormalCard := range slowSendRanks {
		processAbnormalCard(abnormalCard, PPranks, badConnectNumbers)
	}
}

// 处理每个异常Send
func processAbnormalCard(abnormalCard int, PPranks []int, badConnectNumbers []int) {
	// Iterate through PPranks and update badConnectNumbers
	for index, value := range PPranks {
		if abnormalCard == value {
			updateBadConnectNumbersForIndex(index, badConnectNumbers)
		}
	}
}

// 更新指定索引的坏连接数
func updateBadConnectNumbersForIndex(index int, badConnectNumbers []int) {
	if index >= len(badConnectNumbers)-1 {
		return
	}
	// Update for current and next index
	badConnectNumbers[index]++
	badConnectNumbers[index+1]++
}

// 计算ppValue
func calculatePPValue(connectNumbers, badConnectNumbers []int) []float64 {
	ppValue := make([]float64, len(connectNumbers))
	for index := range connectNumbers {
		if index >= len(badConnectNumbers) {
			break
		}
		if connectNumbers[index] != 0 {
			ppValue[index] = float64(badConnectNumbers[index]) / float64(connectNumbers[index])
		} else {
			ppValue[index] = 0
		}
	}

	return ppValue
}

// 查找异常节点
func findAbnormalNodes(PPranks []int, ppValue []float64) []int {
	var res = []int{}
	var half = 0.5
	for index, value := range ppValue {
		if value != 1 {
			continue
		}
		if index == 0 && ppValue[index+1] == half {
			res = append(res, PPranks[index])
		} else if index == len(ppValue)-1 && ppValue[index-1] == half {
			res = append(res, PPranks[index])
		} else if index > 0 && index < len(ppValue)-1 && (ppValue[index-1] == half ||
			ppValue[index+1] == half) {
			res = append(res, PPranks[index])
		}
	}
	return res
}
