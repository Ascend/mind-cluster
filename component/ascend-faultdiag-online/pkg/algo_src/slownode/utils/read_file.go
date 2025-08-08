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

// Package utils is used for file reading and writing, as well as data processing.
package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/model"
)

// 定义结构体，映射 JSON 格式
type dataJson struct {
	TP [][]int `json:"TP"`
	PP [][]int `json:"PP"`
}

const (
	defaultSize = 100 // in megabytes
)

// loadData 从 JSON 文件中读取并解析指定字段的数据
func loadData(filePath string, field string) ([][]int, error) {
	// 读取文件内容
	data, err := utils.LoadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 数据到结构体
	var jsonData dataJson
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}

	// 根据字段名返回对应的数据
	switch field {
	case tp:
		return jsonData.TP, nil
	case pp:
		return jsonData.PP, nil
	default:
		return nil, fmt.Errorf("invalid field: %s", field)
	}
}

// getTPParallel 读取 JSON 文件并返回 TP 数据
func getTPParallel(filePath string) ([][]int, error) {
	return loadData(filePath, tp)
}

// GetPpParallel 读取 JSON 文件并返回 PP 数据
func GetPpParallel(filePath string) ([][]int, error) {
	return loadData(filePath, pp)
}

// 读取CSV文件并返回PP列数据
func readLocalDataFromCSV(filePath string, columnName string) ([]float64, error) {
	if _, err := utils.RealFileChecker(filePath, true, true, defaultSize); err != nil {
		return nil, err
	}

	// 打开CSV文件
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)

	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// 处理关闭文件时的错误
			hwlog.RunLog.Errorf("error occurred while closing the file: %v", err)
		}
	}()

	// 读取CSV数据
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file %s: %v", filePath, err)
	}

	// 获取标题行（第一行）
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file %s is empty", filePath)
	}
	headers := records[0]

	// 查找指定的列名对应的索引
	var columnIndex int
	columnFound := false
	for i, header := range headers {
		// 去除下划线并转换为小写后进行比较
		if strings.ToLower(strings.ReplaceAll(header, "_", "")) == strings.ToLower(strings.ReplaceAll(columnName, "_", "")) {
			columnIndex = i
			columnFound = true
			break
		}
	}

	// 提取列数据
	var columnData = []float64{}

	// 如果未找到列名，返回错误
	if !columnFound {
		// 输出警告信息，不应该直接报错
		hwlog.RunLog.Warnf("Column %s not found in file %s", columnName, filePath)
		return columnData, nil
	}

	for i, record := range records {
		// 跳过标题行
		if i == 0 {
			continue
		}

		value, err := strconv.ParseFloat(record[columnIndex], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing value in file %s: %v", filePath, err)
		}
		columnData = append(columnData, value)
	}

	return columnData, nil
}

// 对齐多个切片数据
func alignData(localDataSlices [][]float64) [][]float64 {
	// 假设所有切片的长度相同，基于第一个切片的长度进行对齐
	if len(localDataSlices) == 0 {
		return nil
	}

	// 获取最小的切片长度
	minLength := len(localDataSlices[0])
	for _, ppData := range localDataSlices {
		if len(ppData) < minLength {
			minLength = len(ppData)
		}
	}

	// 对齐数据，取最小长度
	var alignedData [][]float64
	for _, ppData := range localDataSlices {
		var row = []float64{}
		row = append(row, ppData[:minLength]...)
		alignedData = append(alignedData, row)
	}

	return alignedData
}

// ReadLocalDataAndAlign 读取 CSV 文件中的 ZPPP 列并返回 []int64 类型的切片
func ReadLocalDataAndAlign(fileRanks []int, tpPpFilePath string, columnName string) ([][]float64, []int) {
	var localDataSlices = [][]float64{} // 读取每个文件的数据
	var haveDataRanks = []int{}         // 可能某些TP并没有采集到数据，剩下的卡也需要完成检测
	for _, rank := range fileRanks {
		fileName := fmt.Sprintf("global_rank_%d.csv", rank)
		filePath := tpPpFilePath + "/" + fileName
		ppData, err := readLocalDataFromCSV(filePath, columnName)
		if err != nil {
			hwlog.RunLog.Errorf("Error reading file %s: %v", fileName, err)
		}
		if ppData == nil || len(ppData) == 0 {
			hwlog.RunLog.Warnf("RankDir %d not collected data!", rank)
		} else {
			localDataSlices = append(localDataSlices, ppData) // 读取每个文件的数据
			haveDataRanks = append(haveDataRanks, rank)
		}
	}
	// 判断 columnName 中是否包含 "ZP"
	if strings.Contains(strings.ToUpper(columnName), zp) {
		alignedData := alignData(localDataSlices) // 对齐数据
		return alignedData, haveDataRanks
	}

	return localDataSlices, haveDataRanks
}

// ReadStepTimeCSV 读取 CSV 文件中的 durationtime 列并返回 []int64 类型的切片
func ReadStepTimeCSV(steptimepath string) ([]float64, error) {
	if _, err := utils.RealFileChecker(steptimepath, true, true, defaultSize); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(steptimepath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// 处理关闭文件时的错误
			hwlog.RunLog.Errorf("close file error: %v", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// 跳过表头
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}
	records = records[1:] // 去掉表头行

	var stepTimeList []float64
	for _, record := range records {
		if len(record) < minColumns {
			return nil, fmt.Errorf("invalid CSV format: missing columns")
		}

		duration, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse durationtime as int64: %v", err)
		}

		stepTimeList = append(stepTimeList, duration)
	}

	return stepTimeList, nil
}

// 读取文件内容并解析 JSON
func readJSONFile(filePath string) (*model.NodeResult, error) {
	// 读取文件内容
	fileContent, err := utils.LoadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// 创建 FileData 实例
	var data model.NodeResult
	// 解析 JSON 数据
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON from file %s: %v", filePath, err)
	}
	return &data, nil
}

// MergeDataFromFiles 遍历文件夹下的所有文件并合并数据
func MergeDataFromFiles(directoryPath string) (*model.ClusterResult, []int, error) {
	// 用来存储合并后的数据
	mergedData := model.ClusterResult{
		SlowCalculateRanks:       []int{},
		SlowCommunicationDomains: [][]int{},
		SlowHostNodes:            []string{},
		SlowIORanks:              []int{},
	}
	slowSendRanks := make([]int, 0)

	// 遍历目录下的所有文件
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			hwlog.RunLog.Warnf("Error accessing path: %v", err)
			return nil
		}

		// 只处理文件，忽略文件夹
		if info.IsDir() {
			return nil
		}
		// 读取并解析每个 JSON 文件
		data, err := readJSONFile(path)
		if err != nil {
			hwlog.RunLog.Errorf("Error reading JSON file: %v", err)
			return nil
		}
		// 合并数据
		mergedData.SlowCalculateRanks = append(mergedData.SlowCalculateRanks, data.SlowCalculateRank...)
		slowSendRanks = append(slowSendRanks, data.SlowSendRanks...)
		mergedData.SlowIORanks = append(mergedData.SlowIORanks, data.SlowIORanks...)

		// 合并二维切片 slowCommunicationDomains
		mergedData.SlowCommunicationDomains = append(
			mergedData.SlowCommunicationDomains,
			data.SlowCommunicationDomain...,
		)
		mergedData.SlowHostNodes = append(
			mergedData.SlowHostNodes,
			data.SlowHostNodes...,
		)
		return nil
	})

	if err != nil {
		return &model.ClusterResult{}, nil, err
	}
	return &mergedData, slowSendRanks, nil
}
