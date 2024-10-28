// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/kube"
)

// GetPlatStrategy get plat strategy
func GetPlatStrategy(name, namespace string) string {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		return ""
	}
	value, ok := pg.Annotations[common.ProcessRecoverStrategy]
	if !ok {
		return ""
	}
	if value == common.PlatFormArfStrategyName {
		return common.ProcessArfStrategy
	}
	if value == common.PlatFormDumpStrategyName {
		return common.ProcessDumpStrategy
	}
	if value == common.PlatFormExitStrategyName {
		return common.ProcessExitStrategy
	}
	hwlog.RunLog.Warnf("ProcessRecoverStrategy value=%s not supported, use exit strategy, name=%s",
		value, name)
	return common.ProcessExitStrategy
}

func processContinue(name, namespace string) (bool, string, error) {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("get pg err: %v", err)
		return false, "", err
	}
	value, ok := pg.Annotations[common.ProcessRecoverStrategy]
	if !ok {
		return false, "", nil
	}
	if value == common.PlatFormArfStrategyName || value == common.PlatFormDumpStrategyName {
		return true, value, nil
	}
	return true, value, errors.New("wait ProcessArfStrategy=recover/dump")
}

// WaitProcessContinue block process until processContinue return true
func WaitProcessContinue(name, namespace string) (bool, string, error) {
	startTime := time.Now().Unix()
	platForm, strategy, err := processContinue(name, namespace)
	for platForm && err != nil {
		time.Sleep(common.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > common.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s process continue timeout, timeUse=%d > %d second, err=%v",
				name, timeUse, common.ProcessControlTimeout, err)
			break
		}
		platForm, strategy, err = processContinue(name, namespace)
	}
	return platForm, strategy, err
}

// UpdateProcessConfirmFault update UpdateProcessConfirmFault which store fault ranks
func UpdateProcessConfirmFault(name, namespace string, cacheRanks []string) error {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when update UpdateProcessConfirmFault, err:%v, name:%s", err, name)
		return err
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	rankStr, ok := pg.Annotations[common.ProcessConfirmFaultKey]
	if !ok {
		rankStr = ""
		hwlog.RunLog.Warnf("can not find ProcessConfirmFaultKey, name:%s", name)
	}
	rankList := strings.Split(rankStr, ",")
	allFaultRanks := util.RemoveSliceDuplicateElement(append(rankList, cacheRanks...))
	pg.Annotations[common.ProcessConfirmFaultKey] = strings.Join(allFaultRanks, ",")
	_, err = kube.RetryUpdatePodGroup(pg, common.UpdatePodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to update pg when UpdateProcessConfirmFault, err:%v, name:%s", err, name)
		return err
	}
	return nil
}

// UpdateRecoverStatus update recover status
func UpdateRecoverStatus(name, namespace, value string) {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when update UpdateRecoverStatus, err:%v, name:%s", err, name)
		return
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	pg.Annotations[common.ProcessRecoverStatusKey] = value
	_, err = kube.RetryUpdatePodGroup(pg, common.UpdatePodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to update pg when UpdateRecoverStatus, err:%v, name:%s", err, name)
	}
	hwlog.RunLog.Infof("update name=%s recover status=%s, success", name, value)
}

func pullProcessResultFault(name, namespace string) ([]string, []string, error) {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when pullProcessResultFault, err:%v, name:%s", err, name)
		return nil, nil, err
	}
	if pg.Annotations == nil {
		hwlog.RunLog.Warnf("pod annotation is nil, name:%s", name)
		return nil, nil, fmt.Errorf("pod annotation is nil, name:%s", name)
	}
	resultRanks, ok := pg.Annotations[common.ProcessResultFaultKey]
	if !ok {
		hwlog.RunLog.Warnf("can not fiind ProcessResultFaultKey, name:%s", name)
		return nil, nil, fmt.Errorf("processResultFaultKey not exist, name:%s", name)
	}
	rankSlice := strings.Split(resultRanks, ",")
	if len(rankSlice) == 0 {
		err = errors.New("processResultFault lenth is 0")
	}
	confirmRanks, ok := pg.Annotations[common.ProcessConfirmFaultKey]
	if !ok {
		confirmRanks = ""
	}
	return strings.Split(resultRanks, ","), strings.Split(confirmRanks, ","), err
}

// WaitProcessResultFault block process until ProcessResultFaultKey's ranks not empty
func WaitProcessResultFault(name, namespace string) ([]string, error) {
	startTime := time.Now().Unix()
	resultRanks, confirmRanks, err := pullProcessResultFault(name, namespace)
	for len(resultRanks) == 0 {
		time.Sleep(common.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > common.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s ProcessResultFault timeout, timeUse=%d > %d second",
				name, timeUse, common.ProcessControlTimeout)
			break
		}
		resultRanks, confirmRanks, err = pullProcessResultFault(name, namespace)
	}
	return util.RemoveSliceDuplicateElement(append(resultRanks, confirmRanks...)), err
}

func rankTableReady(name, namespace string) bool {
	pg, err := kube.RetryGetPodGroup(name, namespace, common.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when check rankTableReady, err:%s,name:%s", err, name)
		return false
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	ready, ok := pg.Annotations[common.RankTableReadyKey]
	if !ok {
		return false
	}
	return ready == strconv.FormatBool(true)
}

// WaitRankTableReady block process until RankTableReady is true
func WaitRankTableReady(name, namespace string) error {
	startTime := time.Now().Unix()
	ready := rankTableReady(name, namespace)
	for !ready {
		time.Sleep(common.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > common.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s RankTableReady timeout, timeUse=%d > %d second",
				name, startTime, common.ProcessControlTimeout)
			return fmt.Errorf("check %s RankTableReady timeout, timeUse=%d > %d second",
				name, startTime, common.ProcessControlTimeout)
		}
		ready = rankTableReady(name, namespace)
	}
	return nil
}
