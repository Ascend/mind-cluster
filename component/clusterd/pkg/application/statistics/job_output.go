// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"context"
	"strconv"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/kube"
)

// OutputMgr output cache to cm
type OutputMgr struct{}

var (
	// GlobalJobOutputMgr is a global instance of StatisticInfo output to cm.
	GlobalJobOutputMgr *OutputMgr
)

const (
	updateStatisticFrequency = 3
)

func init() {
	GlobalJobOutputMgr = &OutputMgr{}
}

// JobOutput output cache to cm
func (c *OutputMgr) JobOutput(ctx context.Context) {
	// Create a timer that fires every 3 seconds and updates cm if the data is updated
	var lastVersion int64 = statistics.InitVersion
	ticker := time.NewTicker(updateStatisticFrequency * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("accept stop work signal")
			return
		case <-ticker.C:
			curJobStatistic, version := statistics.JobStcMgrInst.GetAllJobStatistic()
			if version == lastVersion {
				continue
			}
			lastVersion = version
			cmData := c.buildCmData(curJobStatistic)
			err := kube.UpdateOrCreateConfigMap(statistics.JobStcCMName, api.DLNamespace, cmData, nil)
			if err != nil {
				hwlog.RunLog.Errorf("update or create cm err:%v", err)
			}
		}
	}
}

// buildCmData build cm data
func (c *OutputMgr) buildCmData(curJobStatistic constant.CurrJobStatistic) map[string]string {
	tmpSlice := make([]constant.JobStatistic, 0, len(curJobStatistic.JobStatistic))
	cmData := make(map[string]string)
	for _, jobStc := range curJobStatistic.JobStatistic {
		tmpSlice = append(tmpSlice, jobStc)
	}
	cmData[statistics.JobDataCmKey] = util.ObjToString(tmpSlice)
	cmData[statistics.TotalJobsCmKey] = strconv.Itoa(len(tmpSlice))
	return cmData
}
