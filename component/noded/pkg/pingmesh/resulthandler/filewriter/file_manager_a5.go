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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	eidType = 2
)

// HandleUBPingMeshInfo handle ub pingmesh result
func (m *manager) HandleUBPingMeshInfo(res *types.HccspingMeshResult) error {
	if res == nil || res.Policy == nil || res.Results == nil {
		return fmt.Errorf("result is nil")
	}
	m.writer.Infof("uid: %s, config: %#v", res.Policy.UID, res.Policy.Config)
	appendMode, openFlag := m.calcAppendModeAndOpenFlag()
	pingResultCsv, _, err := m.prepareResultFilePaths(appendMode)
	if err != nil {
		hwlog.RunLog.Errorf("get result file path failed, err: %v", err)
		return fmt.Errorf("prepare result file paths failed")
	}
	f, err := os.OpenFile(pingResultCsv, openFlag, defaultPerm)
	if err != nil {
		hwlog.RunLog.Errorf("open file %s failed, err:%v", pingResultCsv, err)
		return fmt.Errorf("open file failed")
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			hwlog.RunLog.Errorf("close file %s failed, err: %v", pingResultCsv, errClose)
			return
		}
	}()
	err = f.Chmod(defaultPerm)
	if err != nil {
		hwlog.RunLog.Errorf("chmod file %s failed, err:%v", pingResultCsv, err)
		return err
	}

	return m.writeRecordA5(f, res, pingResultCsv, appendMode)
}

func (m *manager) writeRecordA5(f *os.File, res *types.HccspingMeshResult,
	pingCsvStr string, appendMode bool) error {
	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()
	csvWriter.Comma = csvComma
	if !appendMode {
		if err := csvWriter.Write(m.CsvColumnNames); err != nil {
			hwlog.RunLog.Errorf("write record csv column title to file %s failed, err: %v", pingCsvStr, err)
			return fmt.Errorf("write record csv column title failed")
		}
	}
	for physicID, infos := range res.Results {
		devices, ok := res.Policy.Address[os.Getenv(api.NodeNameEnv)]
		if !ok {
			continue
		}
		_, ok = devices[physicID]
		if !ok {
			continue
		}
		if _, ok = res.Policy.DestAddrMap[physicID]; !ok {
			continue
		}
		m.writeForCardA5(physicID, infos)
		m.writeForCardToCsvA5(csvWriter, infos)
	}
	hwlog.RunLog.Info("write record to csv file success")
	return nil
}

func (m *manager) writeForCardA5(physicID string, infos map[uint]*common.HccspingMeshInfo) {
	for taskID, info := range infos {
		m.writer.Infof("physicID: %s, taskID: %d", physicID, taskID)
		m.loopUBPingMeshListForCard(info)
	}
}

func (m *manager) loopUBPingMeshListForCard(info *common.HccspingMeshInfo) {
	for _, ubInfo := range info.UBPingMeshInfoList {
		destNum := ubInfo.DestNum
		for index := 0; index < destNum; index++ {
			dstEid := hex.EncodeToString(ubInfo.DstEIDList[index].Raw[:])
			srcEid := hex.EncodeToString(ubInfo.SrcEIDs.Raw[:])
			ri := resultInfo{
				SourceAddr:   srcEid,
				TargetAddr:   dstEid,
				SucPktNum:    ubInfo.SucPktNum[index],
				FailPktNum:   ubInfo.FailPktNum[index],
				MaxTime:      ubInfo.MaxTime[index],
				MinTime:      ubInfo.MinTime[index],
				AvgTime:      ubInfo.AvgTime[index],
				TP95Time:     ubInfo.Tp95Time[index],
				ReplyStatNum: ubInfo.ReplyStatNum[index],
				PingTotalNum: ubInfo.PingTotalNum[index],
			}
			b, err := json.Marshal(ri)
			if err != nil {
				hwlog.RunLog.Errorf("json marshal error: %v", err)
				continue
			}
			m.writer.Info(string(b))
		}
	}
}

func (m *manager) writeForCardToCsvA5(csvWriter *csv.Writer, infos map[uint]*common.HccspingMeshInfo) {
	for taskID, info := range infos {
		m.loopUBPingMeshListForCardCsv(info, taskID, csvWriter)
	}
}

func (m *manager) loopUBPingMeshListForCardCsv(info *common.HccspingMeshInfo, taskID uint, csvWriter *csv.Writer) {
	for _, ubInfo := range info.UBPingMeshInfoList {
		for i := 0; i < ubInfo.DestNum; i++ {
			dstEid := hex.EncodeToString(ubInfo.DstEIDList[i].Raw[:])
			srcEid := hex.EncodeToString(ubInfo.SrcEIDs.Raw[:])
			// keep the corresponding columns of the array [filewriter.CsvColumnNames]
			avgLossRateStr := calcAvgLossRate(ubInfo.SucPktNum[i], ubInfo.FailPktNum[i])
			record := []string{
				strconv.Itoa(int(taskID)),       // taskID
				strconv.Itoa(eidType),           // srcType
				srcEid,                          // srcAddr
				strconv.Itoa(eidType),           // dstType
				dstEid,                          // dstAddr
				strconv.Itoa(ubInfo.MinTime[i]), // minDelay
				strconv.Itoa(ubInfo.MaxTime[i]), // maxDelay
				strconv.Itoa(ubInfo.AvgTime[i]), // avgDelay
				avgLossRateStr,                  // minLossRate use the avgLossRate value
				avgLossRateStr,                  // maxLossRate use the avgLossRate value
				avgLossRateStr,                  // avgLossRate
				strconv.FormatInt(time.Now().UnixMilli(), digitalBase), // timestamp use the write time stamp
			}
			if errWrite := csvWriter.Write(record); errWrite != nil {
				hwlog.RunLog.Errorf("write record to csv file failed, err: %v", errWrite)
				continue
			}
		}
	}
}
