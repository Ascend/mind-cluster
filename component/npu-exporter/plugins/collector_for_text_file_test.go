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

package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	testFilePath           = "/test/path/file.json"
	testAbsFilePath        = "/abs/test/path/file.json"
	testMetricName         = "test_metric"
	testMetricDesc         = "test description"
	testMetricVersion      = "1.0"
	testLabelKey           = "label_key"
	testLabelValue         = "label_value"
	testErrorMsg           = "test error"
	noSuchFileErrorMsg     = "no such file or directory"
	otherErrorMsg          = "permission denied"
	testMaxDataListSize    = 129
	testMaxLabelSize       = 11
	testMaxMetricNameSize  = 129
	testMaxDescSize        = 1025
	testInvalidVersion     = "2.0"
	testEmptyString        = ""
	testZeroTimestamp      = int64(0)
	testValidTimestamp     = int64(1234567890)
	testNewMetricName      = "new_metric"
	testNewMetricDesc      = "new description"
	testNewMetricVersion   = "2.0"
	testLogFlagName        = "name"
	testLogFlagVersion     = "version"
	testLogFlagDesc        = "desc"
	testDirPath            = "/test/dir"
	num2                   = 2
	testFilePath2          = "/test/path/file2.json"
	testFilePaths          = "/test/path/file1.json,/test/path/file2.json"
	testFilePathsExceedMax = "/test/path/file1.json,/test/path/file2.json," +
		"/test/path/file3.json,/test/path/file4.json,/test/path/file5.json," +
		"/test/path/file6.json,/test/path/file7.json,/test/path/file8.json," +
		"/test/path/file9.json,/test/path/file10.json,/test/path/file11.json"
	testInvalidJSON   = "invalid json"
	testEmptyFileData = ""
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
}
func TestIsDataOk(t *testing.T) {
	convey.Convey("should return nil when all fields are valid", t, func() {
		metricsData := createValidTextMetricData()
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestIsDataOkDataList(t *testing.T) {
	convey.Convey("should return error when dataList is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.DataList = []DataItem{}
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "dataList is empty")
	})

	convey.Convey("should return error when dataList size exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.DataList = make([]DataItem, testMaxDataListSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "size of dataList")
	})

	convey.Convey("should return error when label size exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		label := make(map[string]string)
		for i := 0; i < testMaxLabelSize; i++ {
			label[testLabelKey+string(rune(i))] = testLabelValue
		}
		metricsData.DataList[0].Label = label
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "size of first item's Label")
	})
}

func TestIsDataOkName(t *testing.T) {
	convey.Convey("should return error when name is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Name = testEmptyString
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "name field is empty")
	})

	convey.Convey("should return error when name length exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Name = createLongString(testMaxMetricNameSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "length of metric name")
	})
}

func TestIsDataOkDesc(t *testing.T) {
	convey.Convey("should return error when desc is empty", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Desc = testEmptyString
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "desc field is empty")
	})

	convey.Convey("should return error when desc length exceeds max", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Desc = createLongString(testMaxDescSize)
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "length of metric desc")
	})
}

func TestIsDataOkVersion(t *testing.T) {
	convey.Convey("should return error when version is not 1.0", t, func() {
		metricsData := createValidTextMetricData()
		metricsData.Version = testInvalidVersion
		err := isDataOk(metricsData, testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "version should be 1.0")
	})
}

func TestIsDataOkTimestamp(t *testing.T) {
	convey.Convey("should return error when timestamp is zero or negative", t, func() {
		testCases := []struct {
			name      string
			timestamp int64
		}{
			{"zero timestamp", testZeroTimestamp},
			{"negative timestamp", int64(-1)},
		}
		for _, tc := range testCases {
			metricsData := createValidTextMetricData()
			metricsData.Timestamp = tc.timestamp
			err := isDataOk(metricsData, testFilePath)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "timestamp field")
		}
	})
}

func TestIsNotExist(t *testing.T) {
	convey.Convey("should return false when error is nil", t, func() {
		result := isNotExistOrEmpty(nil)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("should return true when error contains no such file message", t, func() {
		err := errors.New(noSuchFileErrorMsg)
		result := isNotExistOrEmpty(err)
		convey.So(result, convey.ShouldBeTrue)
	})

	convey.Convey("should return false when error does not contain no such file message", t, func() {
		err := errors.New(otherErrorMsg)
		result := isNotExistOrEmpty(err)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestCheckFile(t *testing.T) {
	convey.Convey("should return nil when file check passes", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(utils.CheckPath, testAbsFilePath, nil)
		patches.ApplyFuncReturn(os.Getuid, 0)
		patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, nil)

		err := checkFilePermission(testFilePath)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("should return error when CheckPath fails", t, func() {
		testErr := errors.New(testErrorMsg)
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(utils.CheckPath, "", testErr)

		err := checkFilePermission(testFilePath)
		convey.So(err, convey.ShouldEqual, testErr)
	})

	convey.Convey("should return error when DoCheckOwnerAndPermission fails", t, func() {
		testErr := errors.New(testErrorMsg)
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFuncReturn(utils.CheckPath, testAbsFilePath, nil)
		patches.ApplyFuncReturn(os.Getuid, 0)
		patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, testErr)

		err := checkFilePermission(testFilePath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testErrorMsg)
	})
}

func createValidTextMetricData() *TextMetricData {
	return &TextMetricData{
		Version:   testMetricVersion,
		Desc:      testMetricDesc,
		Name:      testMetricName,
		Timestamp: testValidTimestamp,
		DataList: []DataItem{
			{
				Label: map[string]string{
					testLabelKey: testLabelValue,
				},
				Value: 1.0,
			},
		},
	}
}

func createLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func resetGlobalMaps() {
	metricStructInfosMap = make(map[string]metricStructInfo)
	existMetrics = make(map[string]string)
	validPaths = make([]string, 0)
}

func setupFileCheckPatches(patches *gomonkey.Patches, isDir bool, checkPathErr error,
	permissionErr error) {
	patches.ApplyFuncReturn(utils.IsDir, isDir)
	patches.ApplyFuncReturn(utils.CheckPath, testAbsFilePath, checkPathErr)
	patches.ApplyFuncReturn(os.Getuid, 0)
	patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, permissionErr)
}

func setupValidFilePatches(patches *gomonkey.Patches) {
	setupFileCheckPatches(patches, false, nil, nil)
	validData := createValidTextMetricData()
	fileData, err := json.Marshal(validData)
	if err != nil {
		fmt.Println("marshal err: ", err)
	}
	patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
}

type checkAndProcessFileTestCase struct {
	name                  string
	path                  string
	setupPatches          func(*gomonkey.Patches)
	expectedResult        bool
	expectedValidPathsLen int
	needResetGlobalMaps   bool
}

func getCheckAndProcessFileTestCases() []checkAndProcessFileTestCase {
	return []checkAndProcessFileTestCase{
		{name: "should return false when path is empty", path: testEmptyString,
			setupPatches:   nil,
			expectedResult: false},
		{name: "should return false when path is directory",
			path:           testDirPath,
			setupPatches:   func(patches *gomonkey.Patches) { patches.ApplyFuncReturn(utils.IsDir, true) },
			expectedResult: false},
		{name: "should return true when file does not exist", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				notExistErr := errors.New(noSuchFileErrorMsg)
				setupFileCheckPatches(patches, false, nil, notExistErr)
			},
			expectedResult: true},
		{name: "should return false when checkFilePermission fails", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				testErr := errors.New(testErrorMsg)
				setupFileCheckPatches(patches, false, nil, testErr)
			},
			expectedResult: false},
		{name: "should return true when file is empty", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				patches.ApplyFuncReturn(utils.ReadLimitBytes, []byte(testEmptyFileData), nil)
			},
			expectedResult: true},
		{name: "should return true when read file returns EOF", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				eofErr := errors.New("EOF")
				patches.ApplyFuncReturn(utils.ReadLimitBytes, nil, eofErr)
			},
			expectedResult: true},
		{name: "should return false when read file fails", path: testFilePath,
			setupPatches: func(patches *gomonkey.Patches) {
				setupFileCheckPatches(patches, false, nil, nil)
				testErr := errors.New(testErrorMsg)
				patches.ApplyFuncReturn(utils.ReadLimitBytes, nil, testErr)
			},
			expectedResult: false},
		{name: "should return false when processFileData succeeds", path: testFilePath,
			setupPatches:          func(patches *gomonkey.Patches) { setupValidFilePatches(patches) },
			expectedResult:        false,
			expectedValidPathsLen: 1,
			needResetGlobalMaps:   true},
	}
}

func TestChecker(t *testing.T) {
	convey.Convey("should return false when structInfo equals newData", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(hwlog.ResetErrCnt, func(domain string, id interface{}) {})

		result := checker(testFilePath, testMetricName, testMetricName, testLogFlagName)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("should return true and log when structInfo differs from newData", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		result := checker(testFilePath, testMetricName, testNewMetricName, testLogFlagName)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestCheckerWithTable(t *testing.T) {
	testCases := []struct {
		name         string
		jsonFilePath string
		structInfo   string
		newData      string
		logFlag      string
		shouldChange bool
	}{{name: "should return false when name matches",
		jsonFilePath: testFilePath,
		structInfo:   testMetricName,
		newData:      testMetricName,
		logFlag:      testLogFlagName,
		shouldChange: false},
		{name: "should return true when name differs",
			jsonFilePath: testFilePath,
			structInfo:   testMetricName,
			newData:      testNewMetricName,
			logFlag:      testLogFlagName,
			shouldChange: true},
		{name: "should return true when version differs",
			jsonFilePath: testFilePath,
			structInfo:   testMetricVersion,
			newData:      testNewMetricVersion,
			logFlag:      testLogFlagVersion,
			shouldChange: true},
		{name: "should return true when desc differs",
			jsonFilePath: testFilePath,
			structInfo:   testMetricDesc,
			newData:      testNewMetricDesc,
			logFlag:      testLogFlagDesc,
			shouldChange: true},
	}

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(hwlog.ResetErrCnt, func(domain string, id interface{}) {})

			result := checker(tc.jsonFilePath, tc.structInfo, tc.newData, tc.logFlag)
			if tc.shouldChange {
				convey.So(result, convey.ShouldBeTrue)
			} else {
				convey.So(result, convey.ShouldBeFalse)
			}
		})
	}
}

func TestIsStructInfoChangedForFile(t *testing.T) {
	convey.Convey("should return false when structInfo does not exist", t, func() {
		metricStructInfosMap = make(map[string]metricStructInfo)
		data := createValidTextMetricData()
		result := isStructInfoChangedForFile(testFilePath, *data)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("should return false when all fields match", t, func() {
		metricStructInfosMap = make(map[string]metricStructInfo)
		metricStructInfosMap[testFilePath] = metricStructInfo{
			name:    testMetricName,
			version: testMetricVersion,
			desc:    testMetricDesc,
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(checker, func(jsonFilePath, structInfo, newData, logFlag string) bool {
			return false
		})

		data := createValidTextMetricData()
		result := isStructInfoChangedForFile(testFilePath, *data)
		convey.So(result, convey.ShouldBeFalse)
		metricStructInfosMap = make(map[string]metricStructInfo)
	})
}

func TestIsStructInfoChangedForFileWithChanges(t *testing.T) {
	convey.Convey("should return true when name changes", t, func() {
		metricStructInfosMap = make(map[string]metricStructInfo)
		metricStructInfosMap[testFilePath] = metricStructInfo{
			name:    testMetricName,
			version: testMetricVersion,
			desc:    testMetricDesc,
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		callCount := 0
		patches.ApplyFunc(checker, func(jsonFilePath, structInfo, newData, logFlag string) bool {
			callCount++
			if callCount == 1 {
				return true
			}
			return false
		})

		data := createValidTextMetricData()
		result := isStructInfoChangedForFile(testFilePath, *data)
		convey.So(result, convey.ShouldBeTrue)
		metricStructInfosMap = make(map[string]metricStructInfo)
	})

	convey.Convey("should return true when version changes", t, func() {
		metricStructInfosMap = make(map[string]metricStructInfo)
		metricStructInfosMap[testFilePath] = metricStructInfo{
			name:    testMetricName,
			version: testMetricVersion,
			desc:    testMetricDesc,
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		callCount := 0
		patches.ApplyFunc(checker, func(jsonFilePath, structInfo, newData, logFlag string) bool {
			callCount++
			if callCount == num2 {
				return true
			}
			return false
		})

		data := createValidTextMetricData()
		result := isStructInfoChangedForFile(testFilePath, *data)
		convey.So(result, convey.ShouldBeTrue)
		metricStructInfosMap = make(map[string]metricStructInfo)
	})
}

func TestProcessFileData(t *testing.T) {
	convey.Convey("should return true when file data is valid", t, func() {
		resetGlobalMaps()
		validData := createValidTextMetricData()
		fileData, err := json.Marshal(validData)
		if err != nil {
			fmt.Println("marshal err: ", err)
		}

		err = processFileData(testFilePath, fileData)
		convey.So(err, convey.ShouldBeNil)
		convey.So(metricStructInfosMap[testFilePath].name, convey.ShouldEqual, testMetricName)
		resetGlobalMaps()
	})

	convey.Convey("should return false when unmarshal fails", t, func() {
		invalidData := []byte(testInvalidJSON)
		err := processFileData(testFilePath, invalidData)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("should return false when metric name already exists", t, func() {
		resetGlobalMaps()
		existMetrics[testMetricName] = testFilePath2
		validData := createValidTextMetricData()
		fileData, err := json.Marshal(validData)
		if err != nil {
			fmt.Println("marshal err: ", err)
		}
		err = processFileData(testFilePath, fileData)
		convey.So(err, convey.ShouldNotBeNil)
		resetGlobalMaps()
	})

	convey.Convey("should return false when isDataOk fails", t, func() {
		resetGlobalMaps()
		invalidData := &TextMetricData{
			Version:   testMetricVersion,
			Desc:      testMetricDesc,
			Name:      testMetricName,
			Timestamp: testZeroTimestamp,
			DataList:  []DataItem{},
		}
		fileData, err := json.Marshal(invalidData)
		if err != nil {
			fmt.Println("marshal err: ", err)
		}
		err = processFileData(testFilePath, fileData)
		convey.So(err, convey.ShouldNotBeNil)
		resetGlobalMaps()
	})
}

func TestCheckAndProcessFile(t *testing.T) {
	testCases := getCheckAndProcessFileTestCases()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			if tc.needResetGlobalMaps {
				resetGlobalMaps()
			}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			if tc.setupPatches != nil {
				tc.setupPatches(patches)
			}

			result := checkAndProcessFile(tc.path)
			convey.So(result, convey.ShouldEqual, tc.expectedResult)
			if tc.expectedValidPathsLen > 0 {
				convey.So(len(validPaths), convey.ShouldEqual, tc.expectedValidPathsLen)
			}
			if tc.needResetGlobalMaps {
				resetGlobalMaps()
			}
		})
	}
}

func TestPreCheckPaths(t *testing.T) {
	convey.Convey("should exit immediately when all files exist", t, func() {
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		patches.ApplyFuncReturn(time.NewTicker, ticker)

		preCheckPaths([]string{testFilePath})
		convey.So(len(validPaths), convey.ShouldEqual, 1)
		resetGlobalMaps()
	})

	convey.Convey("should wait for missing files and timeout", t, func() {
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		notExistErr := errors.New(noSuchFileErrorMsg)
		setupFileCheckPatches(patches, false, nil, notExistErr)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		preCheckPaths([]string{testFilePath})
		convey.So(len(validPaths), convey.ShouldEqual, 0)
		resetGlobalMaps()
	})
}

func TestIsSupported(t *testing.T) {
	convey.Convey("should return false when filePath is empty", t, func() {
		filePath = testEmptyString
		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeFalse)
	})

	convey.Convey("should return true when files are valid", t, func() {
		filePath = testFilePaths
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeTrue)
		resetGlobalMaps()
	})

	convey.Convey("should truncate paths when exceeds maxFileNumber", t, func() {
		filePath = testFilePathsExceedMax
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		setupValidFilePatches(patches)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeTrue)
		resetGlobalMaps()
	})

	convey.Convey("should return false when no valid paths found", t, func() {
		filePath = testFilePaths
		resetGlobalMaps()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		notExistErr := errors.New(noSuchFileErrorMsg)
		setupFileCheckPatches(patches, false, nil, notExistErr)
		patches.ApplyFuncReturn(time.Time.Add, time.Now())

		collector := &TextMetricsInfoCollector{}
		result := collector.IsSupported(nil)
		convey.So(result, convey.ShouldBeFalse)
		resetGlobalMaps()
	})
}

type collectToCacheTestCase struct {
	name              string
	validPaths        []string
	setupPatches      func(*gomonkey.Patches)
	setupStructInfo   func()
	expectedCacheSize int
	needReset         bool
}

func getCollectToCacheSuccessTestCases() []collectToCacheTestCase {
	return []collectToCacheTestCase{
		{name: "should cache data when file read succeeds",
			validPaths: []string{testFilePath},
			setupPatches: func(patches *gomonkey.Patches) {
				validData := createValidTextMetricData()
				fileData, err := json.Marshal(validData)
				if err != nil {
					fmt.Println("marshal err: ", err)
				}
				patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
			},
			expectedCacheSize: 1,
			needReset:         true},
		{name: "should process multiple files",
			validPaths: []string{testFilePath, testFilePath2},
			setupPatches: func(patches *gomonkey.Patches) {
				validData := createValidTextMetricData()
				fileData, err := json.Marshal(validData)
				if err != nil {
					fmt.Println("marshal err: ", err)
				}
				patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
			},
			expectedCacheSize: 2,
			needReset:         true},
	}
}

func getCollectToCacheFailTestCases() []collectToCacheTestCase {
	return []collectToCacheTestCase{
		{name: "should skip when read file fails",
			validPaths: []string{testFilePath},
			setupPatches: func(patches *gomonkey.Patches) {
				testErr := errors.New(testErrorMsg)
				patches.ApplyFuncReturn(utils.ReadLimitBytes, nil, testErr)
			},
			expectedCacheSize: 0,
			needReset:         true},
		{name: "should skip when unmarshal fails",
			validPaths: []string{testFilePath},
			setupPatches: func(patches *gomonkey.Patches) {
				patches.ApplyFuncReturn(utils.ReadLimitBytes, []byte(testInvalidJSON), nil)
			},
			expectedCacheSize: 0,
			needReset:         true},
		{name: "should skip when data validation fails",
			validPaths: []string{testFilePath},
			setupPatches: func(patches *gomonkey.Patches) {
				invalidData := &TextMetricData{Version: testInvalidVersion}
				fileData, err := json.Marshal(invalidData)
				if err != nil {
					fmt.Println("marshal err: ", err)
				}
				patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
			},
			expectedCacheSize: 0,
			needReset:         true},
	}
}

func getCollectToCacheStructChangedTestCases() []collectToCacheTestCase {
	return []collectToCacheTestCase{
		{name: "should skip when struct info changed",
			validPaths: []string{testFilePath},
			setupPatches: func(patches *gomonkey.Patches) {
				validData := createValidTextMetricData()
				fileData, err := json.Marshal(validData)
				if err != nil {
					fmt.Println("marshal err: ", err)
				}
				patches.ApplyFuncReturn(utils.ReadLimitBytes, fileData, nil)
			},
			expectedCacheSize: 0,
			needReset:         true,
			setupStructInfo: func() {
				metricStructInfosMap[testFilePath] = metricStructInfo{
					name:    testNewMetricName,
					version: testMetricVersion,
					desc:    testMetricDesc,
				}
			}},
	}
}

func getCollectToCacheTestCases() []collectToCacheTestCase {
	var testCases []collectToCacheTestCase
	testCases = append(testCases, getCollectToCacheSuccessTestCases()...)
	testCases = append(testCases, getCollectToCacheFailTestCases()...)
	testCases = append(testCases, getCollectToCacheStructChangedTestCases()...)
	return testCases
}

func setupDefaultStructInfo(validPaths []string) {
	metricStructInfosMap[testFilePath] = metricStructInfo{
		name:    testMetricName,
		version: testMetricVersion,
		desc:    testMetricDesc,
	}
	if len(validPaths) > 1 {
		metricStructInfosMap[testFilePath2] = metricStructInfo{
			name:    testMetricName,
			version: testMetricVersion,
			desc:    testMetricDesc,
		}
	}
}

func getCacheCount(collector *TextMetricsInfoCollector) int {
	cacheCount := 0
	collector.Cache.Range(func(key, value interface{}) bool {
		cacheCount++
		return true
	})
	return cacheCount
}

func TestCollectToCache(t *testing.T) {
	testCases := getCollectToCacheTestCases()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			if tc.needReset {
				resetGlobalMaps()
			}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(logger.Debugf, func(format string, args ...interface{}) {})
			if tc.setupPatches != nil {
				tc.setupPatches(patches)
			}

			collector := &TextMetricsInfoCollector{}
			validPaths = tc.validPaths
			if tc.setupStructInfo != nil {
				tc.setupStructInfo()
			} else {
				setupDefaultStructInfo(tc.validPaths)
			}

			collector.CollectToCache(nil, nil)
			convey.So(getCacheCount(collector), convey.ShouldEqual, tc.expectedCacheSize)

			if tc.needReset {
				resetGlobalMaps()
			}
		})
	}
}

type updateTestCase struct {
	name              string
	setupCache        func(*TextMetricsInfoCollector)
	setupStructInfo   func()
	expectedCallCount int
	needReset         bool
}

func getUpdateSuccessTestCases() []updateTestCase {
	return []updateTestCase{
		{name: "should call doUpdate when cache exists",
			setupCache: func(collector *TextMetricsInfoCollector) {
				validData := createValidTextMetricData()
				cacheKey := fmt.Sprintf("%s-%s", baseCacheKey, testFilePath)
				collector.Cache.Store(cacheKey, *validData)
			},
			setupStructInfo: func() {
				metricStructInfosMap[testFilePath] = metricStructInfo{
					name:       testMetricName,
					metricDesc: prometheus.NewDesc(testMetricName, testMetricDesc, []string{testLabelKey}, nil),
					labels:     []string{testLabelKey},
					desc:       testMetricDesc,
					version:    testMetricVersion,
				}
			},
			expectedCallCount: 1,
			needReset:         true},
		{name: "should process multiple data items",
			setupCache: func(collector *TextMetricsInfoCollector) {
				validData := createValidTextMetricData()
				validData.DataList = append(validData.DataList, DataItem{
					Label: map[string]string{testLabelKey: testLabelValue},
					Value: 2.0,
				})
				cacheKey := fmt.Sprintf("%s-%s", baseCacheKey, testFilePath)
				collector.Cache.Store(cacheKey, *validData)
			},
			setupStructInfo: func() {
				metricStructInfosMap[testFilePath] = metricStructInfo{
					name:       testMetricName,
					metricDesc: prometheus.NewDesc(testMetricName, testMetricDesc, []string{testLabelKey}, nil),
					labels:     []string{testLabelKey},
					desc:       testMetricDesc,
					version:    testMetricVersion,
				}
			},
			expectedCallCount: 2,
			needReset:         true},
	}
}

func getUpdateSkipTestCases() []updateTestCase {
	return []updateTestCase{
		{name: "should skip when cache key not found",
			setupCache:        func(collector *TextMetricsInfoCollector) {},
			setupStructInfo:   func() {},
			expectedCallCount: 0,
			needReset:         true},
		{name: "should skip when cache data type mismatch",
			setupCache: func(collector *TextMetricsInfoCollector) {
				cacheKey := fmt.Sprintf("%s-%s", baseCacheKey, testFilePath)
				collector.Cache.Store(cacheKey, "invalid_type")
			},
			setupStructInfo:   func() {},
			expectedCallCount: 0,
			needReset:         true},
	}
}

func getUpdateTestCases() []updateTestCase {
	var testCases []updateTestCase
	testCases = append(testCases, getUpdateSuccessTestCases()...)
	testCases = append(testCases, getUpdateSkipTestCases()...)
	return testCases
}

func TestUpdate(t *testing.T) {
	testCases := getUpdateTestCases()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			if tc.needReset {
				resetGlobalMaps()
			}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(logger.Debugf, func(format string, args ...interface{}) {})
			patches.ApplyFunc(logger.Warnf, func(format string, args ...interface{}) {})

			collector := &TextMetricsInfoCollector{}
			tc.setupCache(collector)
			tc.setupStructInfo()

			callCount := 0
			collector.update(func(jsonFilePath string, structInfo metricStructInfo,
				timestamp time.Time, item DataItem, index int) {
				callCount++
			})

			convey.So(callCount, convey.ShouldEqual, tc.expectedCallCount)

			if tc.needReset {
				resetGlobalMaps()
			}
		})
	}
}
