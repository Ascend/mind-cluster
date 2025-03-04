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

/*
Package filewriter is using for pingmesh result writing to file
*/

package filewriter

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	nodedcommon "nodeD/pkg/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	maxLineLength    = 2047
	serverIDLeftMove = 22
	serverIDMask     = 0x3FF
	dieIDLeftMove    = 16
	dieIDMask        = 0x3
	deviceIDMask     = 0xFFFF
	dieIDOffset      = 2
	deviceIDMinuend  = 199
)

func New(cfg *Config) *manager {
	if cfg == nil || cfg.Path == "" {
		hwlog.RunLog.Warnf("pingmesh result config is nil or dir is empty")
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
		writer: w,
	}
	return m
}

type manager struct {
	writer *hwlog.CustomLogger
}

// HandlePingMeshInfo handle pingmesh result
func (m *manager) HandlePingMeshInfo(res *types.HccspingMeshResult) error {
	m.writer.Infof("uid: %s, config: %#v", res.Policy.UID, res.Policy.Config)
	for physicID, infos := range res.Results {
		devices, ok := res.Policy.Address[os.Getenv(nodedcommon.ENVNodeNameKey)]
		if !ok {
			continue
		}
		sdid, ok := devices[physicID]
		if !ok {
			continue
		}
		m.writeForCard(physicID, superDeviceIDToIP(sdid), infos)
	}
	return nil
}

func (m *manager) writeForCard(physicID string, source string, infos map[uint]*common.HccspingMeshInfo) {
	for taskID, info := range infos {
		m.writer.Infof("physicID: %s, taskID: %d, DestNum: %d", physicID, taskID, info.DestNum)
		for i := 0; i < info.DestNum; i++ {
			ri := resultInfo{
				SourceAddr:   source,
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
