// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package filewriter is using for pingmesh result writing to file
package filewriter

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ascend-common/api"
	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"

	"nodeD/pkg/pingmesh/types"
)

const (
	maxLineLength = 2047
	csvComma      = ','
	defaultPerm   = 0644

	rasNetSubPath          = "cluster"
	superPodPrefix         = "super-pod"
	float64FormatType      = 'f'
	float64FormatPrecision = 3
	float64BitSize         = 64
	digitalBase            = 10
	savePeriodMillSec      = 45 * 1000

	serverIDLeftMove = 22
	serverIDMask     = 0x3FF
	dieIDLeftMove    = 16
	dieIDMask        = 0x3
	deviceIDMask     = 0xFFFF
	dieIDOffset      = 2
	deviceIDMinuend  = 199
)

func superDeviceIDToIP(s string) string {
	sdid, err := strconv.Atoi(s)
	if err != nil {
		hwlog.RunLog.Errorf("convert sdid to int error: %v", err)
		return ""
	}

	serverID := (sdid >> serverIDLeftMove) & serverIDMask
	dieID := (sdid >> dieIDLeftMove) & dieIDMask
	deviceID := sdid & deviceIDMask
	return "192." + strconv.Itoa(serverID) + "." +
		strconv.Itoa(dieIDOffset+dieID) + "." + strconv.Itoa(deviceIDMinuend-deviceID)
}

type resultInfo struct {
	SourceAddr   string `json:"source_addr"`
	TargetAddr   string `json:"target_addr"`
	SucPktNum    uint   `json:"suc_pkt_num"`
	FailPktNum   uint   `json:"fail_pkt_num"`
	MaxTime      int    `json:"max_time"`
	MinTime      int    `json:"min_time"`
	AvgTime      int    `json:"avg_time"`
	TP95Time     int    `json:"tp95_time"`
	ReplyStatNum int    `json:"reply_stat_num"`
	PingTotalNum int    `json:"ping_total_num"`
}

// New creates a new manager
func New(cfg *Config) *manager {
	if cfg == nil || cfg.Path == "" {
		hwlog.RunLog.Warn("pingmesh result config is nil or dir is empty")
		return nil
	}
	hwlog.RunLog.Infof("create file writer, config: %v", cfg)
	w, err := hwlog.NewCustomLogger(&hwlog.LogConfig{
		LogFileName:   cfg.Path,
		OnlyToStdout:  false,
		OnlyToFile:    true,
		LogLevel:      0,
		FileMaxSize:   hwlog.DefaultFileMaxSize,
		MaxLineLength: maxLineLength,
		MaxBackups:    hwlog.DefaultMaxBackups,
		MaxAge:        cfg.MaxAge,
		IsCompress:    false,
		ExpiredTime:   0,
		CacheSize:     hwlog.DefaultCacheSize,
	}, context.TODO())
	if err != nil {
		hwlog.RunLog.Errorf("create logger error: %v", err)
		return nil
	}

	m := &manager{
		writer:      w,
		superPodId:  cfg.SuperPodId,
		serverIndex: cfg.ServerIndex,
	}
	m.CsvColumnNames = []string{
		"pingTaskId", "srcType", "srcAddr", "dstType", "dstAddr", "minDelay", "maxDelay", "avgDelay",
		"minLossRate", "maxLossRate", "avgLossRate", "timestamp",
	}
	return m
}

type manager struct {
	writer         *hwlog.CustomLogger
	superPodId     string
	serverIndex    string
	lastSaveTime   int64
	CsvColumnNames []string
}

// HandlePingMeshInfo handle pingmesh result
func (m *manager) HandlePingMeshInfo(res *types.HccspingMeshResult) error {
	if m == nil || res == nil || res.Policy == nil || res.Results == nil {
		return errors.New("manager or result is nil")
	}
	m.writer.Infof("uid: %s, config: %#v", res.Policy.UID, res.Policy.Config)
	appendMode, openFlag := m.calcAppendModeAndOpenFlag()
	pingResultCsv, _, err := m.prepareResultFilePaths(appendMode)
	if err != nil {
		hwlog.RunLog.Errorf("get result file path failed, err: %v", err)
		return errors.New("prepare result file paths failed")
	}
	f, err := os.OpenFile(pingResultCsv, openFlag, defaultPerm)
	defer func() {
		if f == nil {
			return
		}
		if errClose := f.Close(); errClose != nil {
			hwlog.RunLog.Errorf("close file %s failed, err: %v", pingResultCsv, errClose)
			return
		}
	}()
	if err != nil {
		hwlog.RunLog.Errorf("open file %s failed, err:%v", pingResultCsv, err)
		return errors.New("open file failed")
	}
	err = f.Chmod(defaultPerm)
	if err != nil {
		hwlog.RunLog.Errorf("chmod file %s failed, err:%v", pingResultCsv, err)
		return err
	}

	return m.writeRecord(f, res, pingResultCsv, appendMode)
}

func (m *manager) writeRecord(f *os.File, res *types.HccspingMeshResult,
	pingCsvStr string, appendMode bool) error {
	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()
	csvWriter.Comma = csvComma
	if !appendMode {
		if err := csvWriter.Write(m.CsvColumnNames); err != nil {
			hwlog.RunLog.Errorf("write record csv column title to file %s failed, err: %v", pingCsvStr, err)
			return errors.New("write record csv column title failed")
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
		m.writeForCard(physicID, res.Policy.DestAddrMap[physicID], infos)
		m.writeForCardToCsv(csvWriter, res.Policy.DestAddrMap[physicID], infos)
	}
	hwlog.RunLog.Info("write record to csv file success")
	return nil
}

func (m *manager) writeForCard(physicID string, destAddrList []types.PingItem,
	infos map[uint]*common.HccspingMeshInfo) {
	for taskID, info := range infos {
		if info == nil {
			continue
		}
		m.writer.Infof("physicID: %s, taskID: %d, DestNum: %d", physicID, taskID, info.DestNum)
		for i := 0; i < info.DestNum; i++ {
			pingItem, errFound := getPingItemByDestAddr(destAddrList, info.DstAddr[i])
			if errFound != nil {
				hwlog.RunLog.Errorf("get ping item is empty by dstAddr %s, err: %s", info.DstAddr[i], errFound)
				continue
			}
			ri := resultInfo{
				SourceAddr:   superDeviceIDToIP(pingItem.SrcAddr),
				TargetAddr:   info.DstAddr[i],
				SucPktNum:    info.SucPktNum[i],
				FailPktNum:   info.FailPktNum[i],
				MaxTime:      info.MaxTime[i],
				MinTime:      info.MinTime[i],
				AvgTime:      info.AvgTime[i],
				TP95Time:     info.TP95Time[i],
				ReplyStatNum: info.ReplyStatNum[i],
				PingTotalNum: info.PingTotalNum[i],
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

func (m *manager) writeForCardToCsv(csvWriter *csv.Writer, destAddrList []types.PingItem,
	infos map[uint]*common.HccspingMeshInfo) {
	for taskID, info := range infos {
		if info == nil {
			continue
		}
		for i := 0; i < info.DestNum; i++ {
			// keep the corresponding columns of the array [filewriter.CsvColumnNames]
			avgLossRateStr := calcAvgLossRate(info.SucPktNum[i], info.FailPktNum[i])
			pingItem, errFound := getPingItemByDestAddr(destAddrList, info.DstAddr[i])
			if errFound != nil {
				hwlog.RunLog.Errorf("get ping item is empty by dstAddr %s, err: %s", info.DstAddr[i], errFound)
				continue
			}
			record := []string{
				strconv.Itoa(int(taskID)),      // taskID
				strconv.Itoa(pingItem.SrcType), // srcType
				pingItem.SrcAddr,               // srcAddr
				strconv.Itoa(pingItem.DstType), // dstType
				pingItem.DstAddr,               // dstAddr
				strconv.Itoa(info.MinTime[i]),  // minDelay
				strconv.Itoa(info.MaxTime[i]),  // maxDelay
				strconv.Itoa(info.AvgTime[i]),  // avgDelay
				avgLossRateStr,                 // minLossRate use the avgLossRate value
				avgLossRateStr,                 // maxLossRate use the avgLossRate value
				avgLossRateStr,                 // avgLossRate
				strconv.FormatInt(time.Now().UnixMilli(), digitalBase), // timestamp use the write time stamp
			}
			if errWrite := csvWriter.Write(record); errWrite != nil {
				hwlog.RunLog.Errorf("write record to csv file failed, err: %v", errWrite)
				continue
			}
		}
	}
}

func (m *manager) calcAppendModeAndOpenFlag() (bool, int) {
	appendMode := true
	curTimeMilliSec := time.Now().UnixMilli()
	if m.lastSaveTime == 0 || m.lastSaveTime+savePeriodMillSec <= curTimeMilliSec {
		m.lastSaveTime = curTimeMilliSec
		appendMode = false
	}
	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if appendMode {
		openFlag = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	}
	hwlog.RunLog.Infof("append mode: %v, open flag: %v, lastSaveTime: %d, Now: %d", appendMode, openFlag,
		m.lastSaveTime, curTimeMilliSec)
	return appendMode, openFlag
}

func (m *manager) prepareResultFilePaths(appendMode bool) (csvFile, csvBackFile string, err error) {
	rasNetRootPath, err := slownet.GetRasNetRootPath()
	if err != nil {
		hwlog.RunLog.Errorf("get ras net fault root path failed, err: %v", err)
		return "", "", errors.New("get ras net fault root path failed")
	}
	csvFileName := fmt.Sprintf("ping_result_%s.csv", m.serverIndex)
	csvFileBackName := fmt.Sprintf("ping_result_%s.csv-bak", m.serverIndex)
	superPodSubPath := fmt.Sprintf("%s-%s", superPodPrefix, m.superPodId)
	pingResultCsv := filepath.Join(rasNetRootPath, rasNetSubPath, superPodSubPath, csvFileName)
	pingResultCsvBack := filepath.Join(rasNetRootPath, rasNetSubPath, superPodSubPath, csvFileBackName)
	if _, err = utils.CheckPath(pingResultCsvBack); err != nil {
		hwlog.RunLog.Errorf("file path %s is invalid, err: %v", pingResultCsvBack, err)
		return "", "", errors.New("file path is invalid")
	}
	if utils.IsLexist(pingResultCsvBack) && !appendMode {
		if err = os.Remove(pingResultCsvBack); err != nil {
			hwlog.RunLog.Errorf("remove file %s failed, err: %v", pingResultCsvBack, err)
			return "", "", errors.New("remove file failed")
		}
	}
	if _, err = utils.CheckPath(pingResultCsv); err != nil {
		hwlog.RunLog.Errorf("file path %s is invalid, err: %v", pingResultCsv, err)
		return "", "", errors.New("file path invalid")
	}
	if utils.IsLexist(pingResultCsv) && !appendMode {
		if err = os.Rename(pingResultCsv, pingResultCsvBack); err != nil {
			hwlog.RunLog.Errorf("backup file %s failed, err: %v", pingResultCsv, err)
			return "", "", errors.New("backup file failed")
		}
	}
	return pingResultCsv, pingResultCsvBack, nil
}

func calcAvgLossRate(sucPktNum, failPktNum uint) string {
	var avgLossRate float64
	totalPkgNum := sucPktNum + failPktNum
	if totalPkgNum != 0 {
		avgLossRate = float64(failPktNum) / float64(totalPkgNum)
	}
	avgLossRateStr := strconv.FormatFloat(avgLossRate, float64FormatType, float64FormatPrecision,
		float64BitSize)
	return avgLossRateStr
}

func getPingItemByDestAddr(dstAddrList []types.PingItem, dstAddr string) (types.PingItem, error) {
	for _, item := range dstAddrList {
		if item.DstAddr == dstAddr || superDeviceIDToIP(item.DstAddr) == dstAddr {
			return item, nil
		}
	}
	return types.PingItem{}, errors.New("not found it")
}
