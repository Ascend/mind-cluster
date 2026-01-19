/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package filewriter is using for pingmesh result writing to file
package filewriter

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
	_ "nodeD/pkg/testtool"
)

const (
	superPod0    = "0"
	serverIndex0 = "0"
	nodeName0    = "node0"
	nodeName1    = "node1"
	physicID1    = "1"
	physicID1Int = 1
	physicID1Ip1 = "ip1"
	physicID2Ip1 = "ip2"
	taskID0      = uint(0)
	sdid1        = "4259841"
)

func patchNewCustomLogger(log *hwlog.CustomLogger, err error) *gomonkey.Patches {
	mockedFunc := func(*hwlog.LogConfig, context.Context) (*hwlog.CustomLogger, error) {
		return log, err
	}
	patch := gomonkey.ApplyFunc(hwlog.NewCustomLogger, mockedFunc)
	return patch
}

// TestNew for all testing cases entry of New func
func TestNew(t *testing.T) {
	convey.Convey("test New func", t, func() {
		testNewShouldReturnNilWhenCfgIsNil()
		testNewShouldReturnNilWhenCfgWithEmptyPath()
		testNewShouldReturnNilWhenNewLogFailed()
		testNewShouldReturnManagerWhenNewLogSuccess()
	})
}

func testNewShouldReturnNilWhenCfgIsNil() {
	convey.Convey("should return nil when cfg is nil", func() {
		var cfg *Config = nil
		fm := New(cfg)
		convey.So(fm, convey.ShouldBeNil)
	})
}

func testNewShouldReturnNilWhenCfgWithEmptyPath() {
	convey.Convey("should return nil when cfg with empty path", func() {
		cfg := &Config{Path: ""}
		fm := New(cfg)
		convey.So(fm, convey.ShouldBeNil)
	})
}

func testNewShouldReturnNilWhenNewLogFailed() {
	convey.Convey("should return nil when new log failed", func() {
		patch := patchNewCustomLogger(nil, errors.New("new custom logger failed"))
		defer patch.Reset()
		cfg := &Config{Path: "testPath"}
		fm := New(cfg)
		convey.So(fm, convey.ShouldBeNil)
	})
}

func testNewShouldReturnManagerWhenNewLogSuccess() {
	convey.Convey("should return manager when new log success", func() {
		logger := &hwlog.CustomLogger{}
		patch := patchNewCustomLogger(logger, nil)
		defer patch.Reset()
		cfg := &Config{Path: "testPath", SuperPodId: superPod0, ServerIndex: serverIndex0}
		fm := New(cfg)
		convey.So(fm, convey.ShouldNotBeNil)
		convey.So(fm.CsvColumnNames, convey.ShouldNotBeNil)
		convey.So(fm.superPodId, convey.ShouldEqual, superPod0)
		convey.So(fm.serverIndex, convey.ShouldEqual, serverIndex0)
	})
}

func TestCalcAppendModeAndOpenFlag(t *testing.T) {
	convey.Convey("test calcAppendModeAndOpenFlag func", t, func() {
		testCalcAppendModeAndOpenFlagShouldBeFalseWhenFirstCalled()
		testCalcAppendModeAndOpenFlagShouldBeTrueWhenSecondCalled()
		testCalcAppendModeAndOpenFlagShouldBeFalseWhenNextTimePeriod()
	})
}

func testCalcAppendModeAndOpenFlagShouldBeFalseWhenFirstCalled() {
	m := makeManager()
	expectAppendMode := false
	expectedOpenFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	convey.Convey("should appendMode be false when the first called", func() {
		appendMode, openFlag := m.calcAppendModeAndOpenFlag()
		convey.So(appendMode, convey.ShouldEqual, expectAppendMode)
		convey.So(openFlag, convey.ShouldEqual, expectedOpenFlag)
	})
}

func testCalcAppendModeAndOpenFlagShouldBeTrueWhenSecondCalled() {
	m := makeManager()
	expectedOpenFlagFirst := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	expectedOpenFlagSecond := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	convey.Convey("should appendMode be True when the second called", func() {
		appendMode, openFlag := m.calcAppendModeAndOpenFlag()
		convey.So(appendMode, convey.ShouldBeFalse)
		convey.So(openFlag, convey.ShouldEqual, expectedOpenFlagFirst)

		appendMode, openFlag = m.calcAppendModeAndOpenFlag()
		convey.So(appendMode, convey.ShouldBeTrue)
		convey.So(openFlag, convey.ShouldEqual, expectedOpenFlagSecond)
	})
}

func testCalcAppendModeAndOpenFlagShouldBeFalseWhenNextTimePeriod() {
	m := makeManager()
	expectedOpenFlagFirst := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	convey.Convey("should appendMode be False when the next time period called", func() {
		appendMode, openFlag := m.calcAppendModeAndOpenFlag()
		convey.So(appendMode, convey.ShouldBeFalse)
		convey.So(openFlag, convey.ShouldEqual, expectedOpenFlagFirst)
		m.lastSaveTime -= savePeriodMillSec
		appendMode, openFlag = m.calcAppendModeAndOpenFlag()
		convey.So(appendMode, convey.ShouldBeFalse)
		convey.So(openFlag, convey.ShouldEqual, expectedOpenFlagFirst)
	})
}

func TestPrepareResultFilePaths(t *testing.T) {
	convey.Convey("Test prepareResultFilePaths func", t, func() {
		testPrepareResultFilePathsShouldErrWhenRootPathErr()
		testPrepareResultFilePathsShouldErrWhenInvalidPath()
		testPrepareResultFilePathsShouldSuccessWhenAllPathValid()
	})
}

func testPrepareResultFilePathsShouldErrWhenRootPathErr() {
	m := makeManager()
	appendMode := false
	convey.Convey("should return error when ras net root path is invalid", func() {
		patch := gomonkey.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) {
			return "", errors.New("ras net root path invalid")
		})
		defer patch.Reset()
		csvFile, csvBackFile, err := m.prepareResultFilePaths(appendMode)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(csvFile, convey.ShouldBeEmpty)
		convey.So(csvBackFile, convey.ShouldBeEmpty)
	})
}

func testPrepareResultFilePathsShouldErrWhenInvalidPath() {
	m := makeManager()
	appendMode := false
	tmpDir := os.TempDir()
	convey.Convey("should return error when ras net result path is invalid", func() {
		patch := gomonkey.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) {
			return tmpDir, nil
		})
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFunc(utils.CheckPath, func(path string) (string, error) {
			return "", errors.New(fmt.Sprintf("%s path is invalid", path))
		})
		defer patchCheck.Reset()
		csvFile, csvBackFile, err := m.prepareResultFilePaths(appendMode)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(csvFile, convey.ShouldBeEmpty)
		convey.So(csvBackFile, convey.ShouldBeEmpty)
	})
}

func testPrepareResultFilePathsShouldSuccessWhenAllPathValid() {
	m := makeManager()
	appendMode := false
	tmpDir := os.TempDir()
	expectedCsv := filepath.Join(tmpDir, rasNetSubPath, fmt.Sprintf("super-pod-%s", m.superPodId),
		fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	expectedCsvBak := filepath.Join(tmpDir, rasNetSubPath, fmt.Sprintf("super-pod-%s", m.superPodId),
		fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))
	convey.Convey("should success when all path valid and appendMode is false", func() {
		patch := gomonkey.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) {
			return tmpDir, nil
		})
		defer patch.Reset()

		checkPathOutputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"path", nil}},
			{Values: gomonkey.Params{"path", nil}},
		}
		patchCheck := gomonkey.ApplyFuncSeq(utils.CheckPath, checkPathOutputs)
		defer patchCheck.Reset()

		existOutputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{true}},
			{Values: gomonkey.Params{true}},
		}
		patchExists := gomonkey.ApplyFuncSeq(utils.IsLexist, existOutputs)
		defer patchExists.Reset()

		patchRemove := gomonkey.ApplyFunc(os.Remove, func(name string) error {
			return nil
		})
		defer patchRemove.Reset()
		patchRename := gomonkey.ApplyFuncReturn(os.Rename, nil)
		defer patchRename.Reset()
		csvFile, csvBackFile, err := m.prepareResultFilePaths(appendMode)
		convey.So(err, convey.ShouldBeNil)
		convey.So(csvFile, convey.ShouldEqual, expectedCsv)
		convey.So(csvBackFile, convey.ShouldEqual, expectedCsvBak)
	})
}

func patchGetEnv(node string) *gomonkey.Patches {
	return gomonkey.ApplyFunc(os.Getenv, func(key string) string {
		return node
	})
}

func patchCalcAppendModeAndOpenFlag(mgr *manager, appendMode bool, openFlag int) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(mgr, "calcAppendModeAndOpenFlag", func() (bool, int) {
		return appendMode, openFlag
	})
}

func patchPrepareResultFilePaths(m *manager, appendMode bool, csvFile, csvBackFile string) *gomonkey.Patches {
	return gomonkey.ApplyPrivateMethod(m, "prepareResultFilePaths",
		func(appendMode bool) (string, string, error) {
			return csvFile, csvBackFile, nil
		},
	)
}

func TestHandlePingMeshInfo(t *testing.T) {
	convey.Convey("Test HandlePingMeshInfo func", t, func() {
		testHandlePingMeshInfoShouldErrWhenInputIsNil()
		testHandlePingMeshInfoShouldErrWhenInputPolicyIsNil()
		testHandlePingMeshInfoShouldErrWhenInputResultIsNil()
		testHandlePingMeshInfoShouldErrWhenPrepareCsvFailed()
		testHandlePingMeshInfoShouldErrWhenOpenCsvFailed()

		testHandlePingMeshInfoShouldDoNothingWhenInputWithoutLocalHost()
		testHandlePingMeshInfoShouldDoNothingWhenInputWithoutCard()
		testHandlePingMeshInfoShouldDoNothingWhenInputPolicyDestAddrMapWithoutCard()
		testHandlePingMeshInfoShouldWriteSuccessWhenInputDataIsGood()
	})
}

func testHandlePingMeshInfoShouldErrWhenInputIsNil() {
	m := makeManager()
	convey.Convey("should return error when hccsping mesh result is nil", func() {
		var result *types.HccspingMeshResult = nil
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingMeshInfoShouldErrWhenInputPolicyIsNil() {
	m := makeManager()
	convey.Convey("should return error when hccsping mesh policy is nil", func() {
		result := &types.HccspingMeshResult{Policy: nil}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingMeshInfoShouldErrWhenInputResultIsNil() {
	m := makeManager()
	convey.Convey("should return error when results of hccsping mesh result is nil", func() {
		result := &types.HccspingMeshResult{
			Policy:  &types.HccspingMeshPolicy{},
			Results: nil,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingMeshInfoShouldErrWhenPrepareCsvFailed() {
	m := makeManager()
	convey.Convey("should return error when prepare result csv failed", func() {
		patch := gomonkey.ApplyPrivateMethod(m, "prepareResultFilePaths",
			func(appendMode bool) (string, string, error) {
				return "", "", errors.New("file path is invalid")
			},
		)
		defer patch.Reset()

		result := &types.HccspingMeshResult{
			Policy:  &types.HccspingMeshPolicy{},
			Results: nil,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingMeshInfoShouldErrWhenOpenCsvFailed() {
	m := makeManager()
	appendMode := false
	tmpDir := os.TempDir()
	expectedCsv := filepath.Join(tmpDir, rasNetSubPath, fmt.Sprintf("super-pod-%s", m.superPodId),
		fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	expectedCsvBak := filepath.Join(tmpDir, rasNetSubPath, fmt.Sprintf("super-pod-%s", m.superPodId),
		fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))

	convey.Convey("should return error when open result csv file failed", func() {
		patch := patchPrepareResultFilePaths(m, appendMode, expectedCsv, expectedCsvBak)
		defer patch.Reset()

		patchOpen := gomonkey.ApplyFuncReturn(os.OpenFile, nil, errors.New("open file failed"))
		defer patchOpen.Reset()

		result := &types.HccspingMeshResult{
			Policy:  &types.HccspingMeshPolicy{},
			Results: nil,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func cleanTempFiles(tempFiles []string) {
	for _, tempFile := range tempFiles {
		if exist := utils.IsLexist(tempFile); exist {
			if errRemove := os.Remove(tempFile); errRemove != nil {
				hwlog.RunLog.Errorf("remove temp file %s failed: err: %v", tempFile, errRemove)
				continue
			}
		}
	}
}

func testHandlePingMeshInfoShouldDoNothingWhenInputWithoutLocalHost() {
	m := makeManager()
	appendMode := false
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	tmpDir := os.TempDir()
	tempCsv := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	tempCsvBak := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))
	defer cleanTempFiles([]string{tempCsv, tempCsvBak})

	convey.Convey("should do nothing when input result data without localhost", func() {
		patchCalc := patchCalcAppendModeAndOpenFlag(m, appendMode, openFlag)
		defer patchCalc.Reset()
		patchPrepare := patchPrepareResultFilePaths(m, appendMode, tempCsv, tempCsvBak)
		defer patchPrepare.Reset()
		patchEnv := patchGetEnv(nodeName0)
		defer patchEnv.Reset()
		callCnt := 0
		expectedCallCnt := 0
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCard",
			func(physicID string, destAddrList []types.PingItem, infos map[uint]*common.HccspingMeshInfo) {
				callCnt++
				return
			},
		)
		defer patchWriteForCard.Reset()
		m.serverIndex = nodeName0
		policy := &types.HccspingMeshPolicy{
			DestAddr: map[string]types.DestinationAddress{
				nodeName1: {},
			},
		}

		resultData := map[string]map[uint]*common.HccspingMeshInfo{
			physicID1: {
				taskID0: {},
			},
		}
		result := &types.HccspingMeshResult{
			Policy:  policy,
			Results: resultData,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingMeshInfoShouldDoNothingWhenInputWithoutCard() {
	m := makeManager()
	appendMode := false
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	tmpDir := os.TempDir()
	tempCsv := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	tempCsvBak := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))
	defer cleanTempFiles([]string{tempCsv, tempCsvBak})

	convey.Convey("should do nothing when input result data without card", func() {
		patchCalc := patchCalcAppendModeAndOpenFlag(m, appendMode, openFlag)
		defer patchCalc.Reset()
		patchPrepare := patchPrepareResultFilePaths(m, appendMode, tempCsv, tempCsvBak)
		defer patchPrepare.Reset()
		patchEnv := patchGetEnv(nodeName0)
		defer patchEnv.Reset()
		callCnt := 0
		expectedCallCnt := 0
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCard",
			func(physicID string, destAddrList []types.PingItem, infos map[uint]*common.HccspingMeshInfo) {
				callCnt++
				return
			},
		)
		defer patchWriteForCard.Reset()
		m.serverIndex = nodeName0
		policy := &types.HccspingMeshPolicy{
			Address: map[string]types.SuperDeviceIDs{
				nodeName0: {},
			},
		}

		resultData := map[string]map[uint]*common.HccspingMeshInfo{
			physicID1: {
				taskID0: {},
			},
		}
		result := &types.HccspingMeshResult{
			Policy:  policy,
			Results: resultData,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingMeshInfoShouldDoNothingWhenInputPolicyDestAddrMapWithoutCard() {
	m := makeManager()
	appendMode := false
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	tmpDir := os.TempDir()
	tempCsv := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	tempCsvBak := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))
	defer cleanTempFiles([]string{tempCsv, tempCsvBak})
	convey.Convey("should do nothing when input policy destAddrMap without card", func() {
		patchCalc := patchCalcAppendModeAndOpenFlag(m, appendMode, openFlag)
		defer patchCalc.Reset()
		patchPrepare := patchPrepareResultFilePaths(m, appendMode, tempCsv, tempCsvBak)
		defer patchPrepare.Reset()
		patchEnv := patchGetEnv(nodeName0)
		defer patchEnv.Reset()
		callCnt := 0
		expectedCallCnt := 0
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCard",
			func(physicID string, destAddrList []types.PingItem, infos map[uint]*common.HccspingMeshInfo) {
				callCnt++
				return
			},
		)
		defer patchWriteForCard.Reset()
		m.serverIndex = nodeName0
		policy := &types.HccspingMeshPolicy{
			Address: map[string]types.SuperDeviceIDs{
				nodeName0: {
					physicID1: sdid1,
				},
			},
			DestAddrMap: map[string][]types.PingItem{},
		}

		resultData := map[string]map[uint]*common.HccspingMeshInfo{
			physicID1: {
				taskID0: {},
			},
		}
		result := &types.HccspingMeshResult{
			Policy:  policy,
			Results: resultData,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingMeshInfoShouldWriteSuccessWhenInputDataIsGood() {
	m := makeManager()
	appendMode := false
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	tmpDir := os.TempDir()
	tempCsv := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv", m.serverIndex))
	tempCsvBak := filepath.Join(tmpDir, fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex))
	defer cleanTempFiles([]string{tempCsv, tempCsvBak})

	convey.Convey("should write success when input data is all good", func() {
		patchCalc := patchCalcAppendModeAndOpenFlag(m, appendMode, openFlag)
		defer patchCalc.Reset()
		patchPrepare := patchPrepareResultFilePaths(m, appendMode, tempCsv, tempCsvBak)
		defer patchPrepare.Reset()
		patches := patchGetEnv(nodeName0)
		defer patches.Reset()
		callCnt := 0
		patches.ApplyFunc(getPingItemByDestAddr,
			func(_ []types.PingItem, _ string) (types.PingItem, error) {
				callCnt++
				return getPingItem(), nil
			},
		)
		callCntCsv := 0
		patches.ApplyMethod(&csv.Writer{}, "Write",
			func(_ *csv.Writer, _ []string) error {
				callCntCsv++
				return nil
			},
		)
		m.serverIndex = nodeName0
		policy := makePolicyData()

		resultData := map[string]map[uint]*common.HccspingMeshInfo{
			physicID1: {
				taskID0: mockHccspingMeshInfo(),
			},
		}
		result := &types.HccspingMeshResult{
			Policy:  policy,
			Results: resultData,
		}
		err := m.HandlePingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldBeGreaterThan, 0)
		convey.So(callCntCsv, convey.ShouldBeGreaterThan, 0)
	})
}

func getPingItem() types.PingItem {
	return types.PingItem{
		SrcType:      0,
		DstType:      0,
		PktSize:      common.MinPktSize,
		SrcCardPhyId: physicID1Int,
		SrcAddr:      physicID1Ip1,
		DstAddr:      physicID2Ip1,
	}
}

func makePolicyData() *types.HccspingMeshPolicy {
	policy := &types.HccspingMeshPolicy{
		Address: map[string]types.SuperDeviceIDs{
			nodeName0: {
				physicID1: sdid1,
			},
		},
		DestAddrMap: map[string][]types.PingItem{
			physicID1: {
				{
					SrcType:      0,
					DstType:      0,
					PktSize:      common.MinPktSize,
					SrcCardPhyId: physicID1Int,
					SrcAddr:      physicID1Ip1,
					DstAddr:      physicID2Ip1,
				},
			},
		},
	}
	return policy
}
func makeManager() *manager {
	const defaultMaxAge = 7
	cfg := &Config{
		Path:        "test",
		MaxAge:      defaultMaxAge,
		SuperPodId:  superPod0,
		ServerIndex: serverIndex0,
	}
	return New(cfg)
}

func mockHccspingMeshInfo() *common.HccspingMeshInfo {
	return &common.HccspingMeshInfo{
		DstAddr:      []string{"111"},
		SucPktNum:    []uint{1},
		FailPktNum:   []uint{1},
		MaxTime:      []int{1},
		MinTime:      []int{1},
		AvgTime:      []int{1},
		TP95Time:     []int{1},
		ReplyStatNum: []int{1},
		PingTotalNum: []int{1},
		DestNum:      1,
	}
}

func TestGetPingItemByDestAddr(t *testing.T) {
	convey.Convey("Test getPingItemByDestAddr func", t, func() {
		convey.Convey("when data is valid, err should be nil", func() {
			dstAddrList := []types.PingItem{
				{
					SrcType:      0,
					DstType:      0,
					PktSize:      common.MinPktSize,
					SrcCardPhyId: physicID1Int,
					SrcAddr:      physicID1Ip1,
					DstAddr:      physicID2Ip1,
				},
			}
			_, err := getPingItemByDestAddr(dstAddrList, physicID2Ip1)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSuperDeviceIDToIP(t *testing.T) {
	convey.Convey("Test superDeviceIDToIP", t, func() {
		convey.Convey("should return empty string for invalid input", func() {
			convey.So(superDeviceIDToIP("abc"), convey.ShouldEqual, "")
		})
	})
}
