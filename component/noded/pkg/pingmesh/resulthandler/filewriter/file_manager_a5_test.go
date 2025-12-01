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

// Package filewriter for
package filewriter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	byte1 = byte(0x00)
	byte2 = byte(0x01)
)

var (
	mockSrcEID = [16]byte{byte1, byte1, byte1, byte1, byte1, byte1, byte1, byte1,
		byte1, byte1, byte1, byte1, byte1, byte1, byte1, byte2}
	mockDstEID1 = [16]byte{byte1, byte1, byte1, byte1, byte1, byte1, byte1, byte1,
		byte1, byte1, byte1, byte1, byte1, byte1, byte2, byte2}
	destNum = 2
)

func TestHandleUBPingMeshInfo(t *testing.T) {
	convey.Convey("Test HandlePingMeshInfo func", t, func() {
		testHandlePingUBMeshInfoShouldErrWhenInputIsNil()
		testHandlePingUBMeshInfoShouldErrWhenInputPolicyIsNil()
		testHandlePingUBMeshInfoShouldErrWhenInputResultIsNil()
		testHandlePingUBMeshInfoShouldErrWhenPrepareCsvFailed()
		testHandlePingUBMeshInfoShouldErrWhenOpenCsvFailed()

		testHandlePingUBMeshInfoShouldDoNothingWhenInputWithoutLocalHost()
		testHandlePingUBMeshInfoShouldDoNothingWhenInputWithoutCard()
		testHandlePingUBMeshInfoShouldDoNothingWhenInputPolicyDestAddrMapWithoutCard()
		testHandlePingUBMeshInfoShouldWriteSuccessWhenInputDataIsGood()
	})
}

func testHandlePingUBMeshInfoShouldErrWhenInputIsNil() {
	m := makeManager()
	convey.Convey("should return error when hccsping mesh result is nil", func() {
		var result *types.HccspingMeshResult = nil
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingUBMeshInfoShouldErrWhenInputPolicyIsNil() {
	m := makeManager()
	convey.Convey("should return error when hccsping mesh policy is nil", func() {
		result := &types.HccspingMeshResult{Policy: nil}
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingUBMeshInfoShouldErrWhenInputResultIsNil() {
	m := makeManager()
	convey.Convey("should return error when results of hccsping mesh result is nil", func() {
		result := &types.HccspingMeshResult{
			Policy:  &types.HccspingMeshPolicy{},
			Results: nil,
		}
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingUBMeshInfoShouldErrWhenPrepareCsvFailed() {
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
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingUBMeshInfoShouldErrWhenOpenCsvFailed() {
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
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testHandlePingUBMeshInfoShouldDoNothingWhenInputWithoutLocalHost() {
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
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCardA5",
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
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingUBMeshInfoShouldDoNothingWhenInputWithoutCard() {
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
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCardA5",
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
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingUBMeshInfoShouldDoNothingWhenInputPolicyDestAddrMapWithoutCard() {
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
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCardA5",
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
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldEqual, expectedCallCnt)
	})
}

func testHandlePingUBMeshInfoShouldWriteSuccessWhenInputDataIsGood() {
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
		patchEnv := patchGetEnv(nodeName0)
		defer patchEnv.Reset()
		callCnt := 0
		patchWriteForCard := gomonkey.ApplyPrivateMethod(m, "writeForCardA5",
			func(physicID string, destAddrList []types.PingItem, infos map[uint]*common.HccspingMeshInfo) {
				callCnt++
				return
			},
		)
		defer patchWriteForCard.Reset()
		callCntCsv := 0
		patchWriteForCardCsv := gomonkey.ApplyPrivateMethod(m, "writeForCardToCsvA5",
			func(csvWriter *csv.Writer, destAddrList []types.PingItem, infos map[uint]*common.HccspingMeshInfo) {
				callCntCsv++
				return
			},
		)
		defer patchWriteForCardCsv.Reset()
		m.serverIndex = nodeName0
		policy := makePolicyData()

		resultData := map[string]map[uint]*common.HccspingMeshInfo{
			physicID1: {
				taskID0: mockUBHccspingMeshInfo(),
			},
		}
		result := &types.HccspingMeshResult{
			Policy:  policy,
			Results: resultData,
		}
		err := m.HandleUBPingMeshInfo(result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(callCnt, convey.ShouldBeGreaterThan, 0)
		convey.So(callCntCsv, convey.ShouldBeGreaterThan, 0)
	})
}

func mockUBHccspingMeshInfo() *common.HccspingMeshInfo {
	return &common.HccspingMeshInfo{
		UBPingMeshInfoList: []common.UBPingMeshInfo{
			{
				SrcEIDs:      common.Eid{Raw: mockSrcEID},
				DstEIDList:   []common.Eid{{Raw: mockDstEID1}},
				SucPktNum:    []uint{1},
				FailPktNum:   []uint{1},
				MaxTime:      []int{1},
				MinTime:      []int{1},
				AvgTime:      []int{1},
				Tp95Time:     []int{1},
				ReplyStatNum: []int{1},
				PingTotalNum: []int{1},
				OccurTime:    1,
				DestNum:      1,
			},
		},
	}
}

func TestLoopUBPingMeshListForCard(t *testing.T) {
	convey.Convey("Test loopUBPingMeshListForCard", t, func() {
		testSrcEID := common.Eid{Raw: [common.EidByteSize]byte{1, 0}}
		testDstEID := common.Eid{Raw: [common.EidByteSize]byte{0, 1}}

		info := &common.HccspingMeshInfo{
			UBPingMeshInfoList: []common.UBPingMeshInfo{
				{
					DestNum: 1,
					SrcEIDs: testSrcEID,
					DstEIDList: []common.Eid{
						testDstEID,
					},
					SucPktNum:    []uint{0},
					FailPktNum:   []uint{0},
					MaxTime:      []int{0},
					MinTime:      []int{0},
					AvgTime:      []int{0},
					Tp95Time:     []int{0},
					ReplyStatNum: []int{0},
					PingTotalNum: []int{0},
				},
			},
		}
		m := &manager{}
		count := 0
		patch := gomonkey.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
			count++
			return []byte{}, errors.New("json failed")
		})
		defer patch.Reset()
		m.loopUBPingMeshListForCard(info)
		convey.So(count, convey.ShouldEqual, 1)
	})
}

func TestLoopUBPingMeshListForCardCsv(t *testing.T) {
	convey.Convey("Test loopUBPingMeshListForCardCsv", t, func() {
		info := &common.HccspingMeshInfo{
			UBPingMeshInfoList: []common.UBPingMeshInfo{
				{
					DestNum: destNum,
					SrcEIDs: common.Eid{Raw: [common.EidByteSize]byte{0, 1}},
					DstEIDList: []common.Eid{
						{Raw: [common.EidByteSize]byte{0, 1}},
						{Raw: [common.EidByteSize]byte{0, 1}},
					},
					SucPktNum:  []uint{0, 1},
					FailPktNum: []uint{0, 1},
					MinTime:    []int{0, 1},
					MaxTime:    []int{0, 1},
					AvgTime:    []int{0, 1},
				},
			},
		}

		var callCount int
		patches := gomonkey.ApplyMethod(&csv.Writer{}, "Write", func(_ *csv.Writer, record []string) error {
			callCount++
			return nil
		})
		defer patches.Reset()

		m := &manager{}
		csvWriter := csv.NewWriter(nil)

		m.loopUBPingMeshListForCardCsv(info, 1, csvWriter)

		convey.So(callCount, convey.ShouldEqual, destNum)
	})
}
