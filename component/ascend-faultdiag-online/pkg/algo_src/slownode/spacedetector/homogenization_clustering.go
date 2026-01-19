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

// Package spacedetector performs space dimension detection by homogenizing the data.
package spacedetector

import (
	"sort"
)

// 常量定义
const (
	// 最小相邻差异的比例，用于判断是否分为两个类别
	listSpacingThreshold = 2
)

// IndexAndValue 构建值和下标的结构体
type IndexAndValue struct {
	Index int
	Value float64
}

// 空间维度检测，使用聚类算法找出异常数据点, 返回异常值的index列表
func oneDimensionalClustering(dataList []float64, clusterMeanDistance float64) ([]int, []float64) {
	/* 升序排列计算相邻差，记录最大差值下标 */
	indexAndValueList := sortDataByIndexAndValue(dataList)
	differenceList, maxDiffIndex := calculateDifferences(indexAndValueList)
	listSpacing := 0.0
	/* 差值总和 */
	for _, value := range differenceList {
		listSpacing += value
	}
	/* maxdiffIndex有效 */
	if !(len(differenceList) > maxDiffIndex && maxDiffIndex >= 0) {
		return nil, nil
	}
	/* 分类条件1：最大差值（a->b）大于差值总和(总长)的一半 */
	if !(differenceList[maxDiffIndex] >= listSpacing/listSpacingThreshold) {
		return nil, nil
	}
	/* 计算均值大的组和均值小的组的均值 */
	bigMean := calculateMean(indexAndValueList[maxDiffIndex+1:])
	littleMean := calculateMean(indexAndValueList[:maxDiffIndex+1])
	// 判断均值比
	if littleMean == 0 {
		return nil, nil
	}
	/* 分类条件2：大类均值>小类均值的conf.clusterMeanDistance倍数 */
	if bigMean/littleMean > clusterMeanDistance {
		/* 若分成两类将左边的数据视为异类，并取出对应的npu卡索引 */
		return collectIndices(indexAndValueList[:maxDiffIndex+1])
	}
	return nil, nil
}

func recurseDimensionalClustering(dataList []float64, clusterMeanDistance float64) []int {
	var result []int = nil
	var input []float64 = dataList
	for {
		tmpResult, nextDataList := oneDimensionalClustering(input, clusterMeanDistance)
		if tmpResult == nil {
			break
		}
		input = nextDataList
		result = tmpResult
	}
	return result
}

// sortDataByIndexAndValue 根据数据排序并保留原索引
func sortDataByIndexAndValue(dataList []float64) []IndexAndValue {
	var indexAndValueList []IndexAndValue
	for index, value := range dataList {
		indexAndValueList = append(indexAndValueList, IndexAndValue{Index: index, Value: value})
	}
	sort.Slice(indexAndValueList, func(i, j int) bool {
		return indexAndValueList[i].Value < indexAndValueList[j].Value
	})
	return indexAndValueList
}

// calculateDifferences 计算相邻值的差异，并返回差异列表和最大差异的索引
func calculateDifferences(indexAndValueList []IndexAndValue) ([]float64, int) {
	var differenceList []float64
	var maxDiff float64
	maxDiffIndex := -1
	for i := 1; i < len(indexAndValueList); i++ {
		diff := indexAndValueList[i].Value - indexAndValueList[i-1].Value
		differenceList = append(differenceList, diff)
		if diff > maxDiff {
			maxDiff = diff
			maxDiffIndex = i - 1
		}
	}
	return differenceList, maxDiffIndex
}

// calculateMean 计算数据列表的均值
func calculateMean(dataList []IndexAndValue) float64 {
	if len(dataList) == 0 {
		return 0.0
	}
	var sum float64
	for _, value := range dataList {
		sum += value.Value
	}
	return sum / float64(len(dataList))
}

// 收集符合条件的索引
func collectIndices(indexAndValueList []IndexAndValue) ([]int, []float64) {
	var resValue []int
	var abnormalDatas []float64
	for _, value := range indexAndValueList {
		resValue = append(resValue, value.Index)
		abnormalDatas = append(abnormalDatas, value.Value)
	}
	return resValue, abnormalDatas
}
