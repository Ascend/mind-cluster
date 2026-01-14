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

// Package utils is a DT collection for func in file read_file
package utils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/model"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

func TestLoadData(t *testing.T) {

	var filePath = "loadData.json"
	var sourceData = ""
	var field = "field"
	assert.Nil(t, generateFile(sourceData, filePath))

	absPath, err := filepath.Abs(filePath)
	assert.Nil(t, err)

	resoledPath, err := filepath.EvalSymlinks(absPath)
	assert.Nil(t, err)

	// loadFile failed
	mockLoadFile := gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, []byte{}, errors.New("load file failed"))
	data, err := loadData(resoledPath, field)
	assert.Equal(t, "load file failed", err.Error())
	assert.Nil(t, data)
	mockLoadFile.Reset()

	// wrong json data
	sourceData = `{`
	assert.Nil(t, generateFile(sourceData, resoledPath))
	data, err = loadData(resoledPath, field)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Nil(t, data)

	// right data with invalid field
	sourceData = `{"TP":[[1,2,3]],"PP":[[1],[2],[3]]}`
	assert.Nil(t, generateFile(sourceData, resoledPath))
	data, err = loadData(resoledPath, field)
	assert.Equal(t, "invalid field: "+field, err.Error())
	assert.Nil(t, data)

	// right data with TP
	data, err = loadData(resoledPath, tp)
	assert.Nil(t, err)
	assert.ElementsMatch(t, [][]int{{1, 2, 3}}, data)

	// right data with PP
	data, err = loadData(resoledPath, pp)
	assert.Nil(t, err)
	assert.ElementsMatch(t, [][]int{{1}, {2}, {3}}, data)

	// clear file
	assert.Nil(t, clearFile(resoledPath))
}

func TestGetTPParallel(t *testing.T) {
	// mock loadData
	mockLoadData := gomonkey.ApplyFunc(loadData, func(string, string) ([][]int, error) {
		return nil, errors.New("load data failed")
	})
	defer mockLoadData.Reset()
	res, err := getTPParallel("")
	assert.Nil(t, res)
	assert.Equal(t, "load data failed", err.Error())
}

func TestGetPpParallel(t *testing.T) {
	// mock loadData
	mockLoadData := gomonkey.ApplyFunc(loadData, func(string, string) ([][]int, error) {
		return nil, errors.New("load data failed")
	})
	defer mockLoadData.Reset()
	res, err := GetPpParallel("")
	assert.Nil(t, res)
	assert.Equal(t, "load data failed", err.Error())
}

func TestReadLocalDataFromCSV(t *testing.T) {
	var fileCheckFailed = false

	var filePath = "test_read_local_data_from_csv.csv"
	var columnName = "test"
	// mock func
	mockRealFileChecker := gomonkey.ApplyFunc(utils.RealFileChecker, func(
		string, bool, bool, int64) (string, error) {
		if fileCheckFailed {
			return "", errors.New("real file checker failed")
		}
		return "", nil
	})

	defer mockRealFileChecker.Reset()
	fileCheckFailed = true
	data, err := readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, data)
	assert.Equal(t, "real file checker failed", err.Error())
	fileCheckFailed = false
	testReadReadLocalDataFromCSVWithCase1(t, filePath, columnName)
	testReadReadLocalDataFromCSVWithCase2(t, filePath, columnName)
	// clear file
	assert.Nil(t, clearFile(filePath))
}

func testReadReadLocalDataFromCSVWithCase1(t *testing.T, filePath, columnName string) {

	var csvData = ""
	assert.Nil(t, generateFile(csvData, filePath))

	// OpenFile failed
	mockOpenFile := gomonkey.ApplyFunc(os.OpenFile, func(string, int, os.FileMode) (*os.File, error) {
		return nil, errors.New("open file failed")
	})
	data, err := readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, data)
	assert.Equal(t, fmt.Sprintf("error opening file %s: open file failed", filePath), err.Error())
	mockOpenFile.Reset()

	// csv reader, ReadAll failed
	mockReadAll := gomonkey.ApplyMethod(new(csv.Reader), "ReadAll", func(*csv.Reader) ([][]string, error) {
		return nil, errors.New("read all failed")
	})
	data, err = readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, data)
	assert.Equal(t, fmt.Sprintf("error reading CSV file %s: read all failed", filePath), err.Error())
	mockReadAll.Reset()

	// records 为 0
	data, err = readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, data)
	assert.Equal(t, fmt.Sprintf("CSV file %s is empty", filePath), err.Error())
}

func testReadReadLocalDataFromCSVWithCase2(t *testing.T, filePath, columnName string) {
	// wrong format of value in csv with no corresponding column
	csvData := "header1,header2,header3\ndata1,data2,data3\ndata4,data5,data6\ndata7,data8,data9"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err := readLocalDataFromCSV(filePath, columnName)
	assert.Equal(t, []float64{}, data)
	assert.Nil(t, err)

	// wrong format of value in csv with corresponding column
	columnName = "header2"
	csvData = "header1,header2,header3\ndata1,data2,data3\ndata4,data5,data6\ndata7,data8,data9"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err = readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, data)
	assert.Equal(t, fmt.Sprintf(
		`error parsing value in file %s: strconv.ParseFloat: parsing "data2": invalid syntax`,
		filePath), err.Error())

	// correct Data with corresponding column
	csvData = "header1,header2,header3\n1.1,0.2,1.3\n1.4,1.9999999999,1.6\n1.19,1.000000001,1.22"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err = readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, err)
	assert.Equal(t, []float64{0.2, 1.9999999999, 1.000000001}, data)

	// correct Data with corresponding capitalized columnName
	columnName = "HEADER2"
	data, err = readLocalDataFromCSV(filePath, columnName)
	assert.Nil(t, err)
	assert.Equal(t, []float64{0.2, 1.9999999999, 1.000000001}, data)
}

func TestAlignData(t *testing.T) {
	var testCases = []struct {
		dataSlice [][]float64
		expect    [][]float64
	}{
		{[][]float64{}, nil},
		{
			[][]float64{{1.1, 1.2, 1.3}, {1.2, 1.3}, {1.4, 1.1}},
			[][]float64{{1.1, 1.2}, {1.2, 1.3}, {1.4, 1.1}},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, alignData(tc.dataSlice))
	}
}

func TestReadLocalDataAndAlign(t *testing.T) {

	var readFailed = false

	var fileRanks = []int{1, 2, 3}
	var tpPpFilePath = ""
	var columnName = ""

	// mock readLocalDataFromCSV
	mockFunc := gomonkey.ApplyFunc(readLocalDataFromCSV, func(string, string) ([]float64, error) {
		if readFailed {
			return nil, errors.New("read csv file failed")
		}
		return []float64{1.1, 1.2, 1.3}, nil
	})
	defer mockFunc.Reset()

	// read csv failed
	readFailed = true
	localDataSlices, haveDataRanks := ReadLocalDataAndAlign(fileRanks, tpPpFilePath, columnName)
	assert.Equal(t, [][]float64{}, localDataSlices)
	assert.Equal(t, []int{}, haveDataRanks)
	readFailed = false

	// columnName with no ZP
	columnName = "test"
	localDataSlices, haveDataRanks = ReadLocalDataAndAlign(fileRanks, tpPpFilePath, columnName)
	assert.Equal(t, [][]float64{{1.1, 1.2, 1.3}, {1.1, 1.2, 1.3}, {1.1, 1.2, 1.3}}, localDataSlices)
	assert.Equal(t, []int{1, 2, 3}, haveDataRanks)

	// columnName with ZP
	columnName = "xxxZPxxx"
	localDataSlices, haveDataRanks = ReadLocalDataAndAlign(fileRanks, tpPpFilePath, columnName)
	assert.Equal(t, [][]float64{{1.1, 1.2, 1.3}, {1.1, 1.2, 1.3}, {1.1, 1.2, 1.3}}, localDataSlices)
	assert.Equal(t, []int{1, 2, 3}, haveDataRanks)
}

func TestReadStepTimeCSV(t *testing.T) {
	var fileCheckFailed = false

	var filePath = "test_read_step_time_csv.csv"
	// mock func
	mockRealFileChecker := gomonkey.ApplyFunc(utils.RealFileChecker, func(
		string, bool, bool, int64) (string, error) {
		if fileCheckFailed {
			return "", errors.New("real file checker failed")
		}
		return "", nil
	})

	defer mockRealFileChecker.Reset()
	fileCheckFailed = true
	data, err := ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t, "real file checker failed", err.Error())
	fileCheckFailed = false

	testReadStepTimeCSVWithCase1(t, filePath)
	testReadStepTimeCSVWithCase2(t, filePath)
	// clear file
	assert.Nil(t, clearFile(filePath))
}

func testReadStepTimeCSVWithCase1(t *testing.T, filePath string) {
	var csvData = ""
	assert.Nil(t, generateFile(csvData, filePath))

	// OpenFile failed
	mockOpenFile := gomonkey.ApplyFunc(os.OpenFile, func(string, int, os.FileMode) (*os.File, error) {
		return nil, errors.New("open file failed")
	})
	data, err := ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t, "failed to open file: open file failed", err.Error())
	mockOpenFile.Reset()

	// csv reader, ReadAll failed
	mockReadAll := gomonkey.ApplyMethod(new(csv.Reader), "ReadAll", func(*csv.Reader) ([][]string, error) {
		return nil, errors.New("read all failed")
	})
	data, err = ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t, "failed to read CSV file: read all failed", err.Error())
	mockReadAll.Reset()

	// records 为 0
	data, err = ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t, "CSV file is empty", err.Error())
}

func testReadStepTimeCSVWithCase2(t *testing.T, filePath string) {

	// missing columns
	csvData := "header1\ndata1\ndata4\ndata7"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err := ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t, "invalid CSV format: missing columns", err.Error())

	// wrong format of data
	csvData = "header1,durationtime,header3\ndata1,data2,data3\ndata4,data5,data6\ndata7,data8,data9"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err = ReadStepTimeCSV(filePath)
	assert.Nil(t, data)
	assert.Equal(t,
		`failed to parse durationtime as int64: strconv.ParseFloat: parsing "data2": invalid syntax`,
		err.Error())

	// correct data
	csvData = "header1,durationtime,header3\n1.1,0.2,1.3\n1.4,1.9999999999,1.6\n1.19,1.000000001,1.22"
	assert.Nil(t, generateFile(csvData, filePath))
	data, err = ReadStepTimeCSV(filePath)
	assert.Nil(t, err)
	assert.Equal(t, []float64{0.2, 1.9999999999, 1.000000001}, data)
}

func TestReadJSONFile(t *testing.T) {
	var filePath = "read_jsonb_file.json"
	var sourceData = ""
	assert.Nil(t, generateFile(sourceData, filePath))

	absPath, err := filepath.Abs(filePath)
	assert.Nil(t, err)
	resoledPath, err := filepath.EvalSymlinks(absPath)
	assert.Nil(t, err)

	// loadFile failed
	mockLoadFile := gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, []byte{}, errors.New("load file failed"))
	file, err := readJSONFile(resoledPath)
	assert.Nil(t, file)
	assert.Equal(t, fmt.Errorf("error reading file %s: load file failed", resoledPath).Error(), err.Error())
	mockLoadFile.Reset()

	// json unmarshal failed
	sourceData = "}"
	assert.Nil(t, generateFile(sourceData, filePath))
	file, err = readJSONFile(resoledPath)
	assert.Nil(t, file)
	assert.Equal(
		t,
		fmt.Errorf(`error unmarshalling JSON from file %s: invalid character '}' looking for beginning of value`,
			resoledPath).Error(),
		err.Error(),
	)

	// correct data
	sourceData = `{"slowCalculateRanks":[1,2],"slowIONodes":[2,3]}`
	assert.Nil(t, generateFile(sourceData, filePath))
	file, err = readJSONFile(resoledPath)
	assert.Nil(t, err)
	assert.NotNil(t, file)

	// clear file
	assert.Nil(t, clearFile(absPath))
}

func TestMergeDataFromFiles(t *testing.T) {

	// mock Walk
	mockWalk := gomonkey.ApplyFunc(filepath.Walk, func(root string, fn filepath.WalkFunc) error {
		return errors.New("filepath walk error")
	})

	mergedData, slowSendRanks, err := MergeDataFromFiles("path")
	assert.Equal(t, "filepath walk error", err.Error())
	assert.NotNil(t, mergedData)
	assert.ElementsMatch(t, []int{}, slowSendRanks)
	mockWalk.Reset()

	var readJsonFileFailed = false
	// mock readJsonFile
	mockReadJsonFile := gomonkey.ApplyFunc(readJSONFile, func(filePath string) (*model.NodeResult, error) {
		if readJsonFileFailed {
			return nil, errors.New("read json file failed")
		}
		return &model.NodeResult{
			SlowCalculateRank:       []int{1, 2, 3},
			SlowCommunicationDomain: [][]int{{1}, {2}, {3}},
			SlowSendRanks:           []int{4, 5, 6},
			SlowHostNodes:           []string{"t1", "t2"},
			SlowIORanks:             []int{7, 8, 9},
			FileName:                "file_name",
		}, nil
	})

	defer mockReadJsonFile.Reset()
	mergedData, slowSendRanks, err = MergeDataFromFiles(".")
	assert.Nil(t, err)
	assert.True(t, len(slowSendRanks) > 0)
	assert.NotNil(t, mergedData)
}
