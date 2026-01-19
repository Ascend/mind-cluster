// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics statistic funcs about fault
package statistics

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/statistics"
)

// StatisticFault statistic fault cache
var StatisticFault FaultCache

func init() {
	StatisticFault = FaultCache{
		updateChan: make(chan struct{}, 1),
	}
}

// FaultCache struct for statistic fault cache
type FaultCache struct {
	updateChan chan struct{}
}

// Notify notify updateChan for updating fault to configmap statistic-fault-info
func (fc *FaultCache) Notify() {
	if len(fc.updateChan) == 0 {
		fc.updateChan <- struct{}{}
	}
}

// UpdateFault update fault
func (fc *FaultCache) UpdateFault(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Error("ctx stop channel is closed")
			}
			hwlog.RunLog.Info("receive ctx stop signal")
			return
		case _, ok := <-fc.updateChan:
			if !ok {
				hwlog.RunLog.Error("updateChan is closed")
				return
			}
			fc.updateCacheToCM()
			time.Sleep(1 * time.Second)
		}
	}
}

func (fc *FaultCache) updateCacheToCM() {
	pubFaults, pubFaultsNum := publicfault.PubFaultCache.GetPubFaultsForCM()
	faults := util.ObjToString(pubFaults)
	faultNum := util.ObjToString(
		constant.FaultNum{
			PubFaultNum: pubFaultsNum,
		})
	if err := statistics.UpdateFaultToCM(faults, faultNum, pubFaultsNum > constant.MaxFaultNum); err != nil {
		hwlog.RunLog.Errorf("update fault to cm failed, error: %v", err)
		return
	}
}

// LoadFaultData load fault data from configmap statistic-fault-info
func (fc *FaultCache) LoadFaultData() {
	if err := statistics.LoadFaultFromCM(); err != nil {
		hwlog.RunLog.Errorf("load fault from cm failed, error: %v", err)
		return
	}
	hwlog.RunLog.Info("load statistic fault from cm successfully")
}
