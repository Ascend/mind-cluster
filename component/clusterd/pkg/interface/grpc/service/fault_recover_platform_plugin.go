// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

func platFormStrategy(name, namespace string) ([]string, error) {
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("get pg err: %v", err)
		return nil, err
	}
	value, ok := pg.Annotations[constant.ProcessRecoverStrategy]
	if !ok {
		return nil, fmt.Errorf("plat strategy key not exist, job=%s, key=%s",
			name, constant.ProcessRecoverStrategy)
	}
	if value == constant.ProcessRetryStrategyName || value == constant.ProcessRecoverStrategyName ||
		value == constant.ProcessDumpStrategyName {
		strategySlice := strings.Split(value, ",")
		var res []string
		for _, strategy := range strategySlice {
			if common.StrategySupported(strategy) {
				res = append(res, strategy)
			}
		}
		res = append(res, constant.ProcessExitStrategyName)
		return util.RemoveSliceDuplicateElement(res), nil
	}
	return nil, fmt.Errorf("wait plat strategy = retry/recover/dump for job=%s", name)
}

// WaitPlatFormStrategyReady block process until processContinue return true
func WaitPlatFormStrategyReady(name, namespace string) ([]string, error) {
	startTime := time.Now().Unix()
	strategy, err := platFormStrategy(name, namespace)
	for err != nil {
		time.Sleep(constant.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > constant.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s process continue timeout, timeUse=%d > %d second, err=%v",
				name, timeUse, constant.ProcessControlTimeout, err)
			break
		}
		strategy, err = platFormStrategy(name, namespace)
	}
	return strategy, err
}

// UpdateProcessConfirmFault update UpdateProcessConfirmFault which store fault ranks
func UpdateProcessConfirmFault(name, namespace string, cacheRanks []*pb.FaultRank) error {
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when update UpdateProcessConfirmFault, err:%v, name:%s", err, name)
		return err
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	rankStr, _ := pg.Annotations[constant.ProcessConfirmFaultKey]
	if len(rankStr) > 0 {
		return fmt.Errorf("plat not clear pre confirm fault, pgName=%s", name)
	}
	allFaultRanks := common.RemoveSliceDuplicateFaults(cacheRanks)
	newConfirm := map[string]string{
		constant.ProcessConfirmFaultKey: strings.Trim(common.Faults2String(allFaultRanks), ","),
	}
	_, err = kube.RetryPatchPodGroupAnnotations(name, namespace, constant.UpdatePodGroupTimes, newConfirm)
	if err != nil {
		hwlog.RunLog.Errorf("failed to update pg when UpdateProcessConfirmFault, err:%v, name:%s", err, name)
		return err
	}
	return nil
}

// UpdateRecoverStatus update recover status
func UpdateRecoverStatus(name, namespace, value string) {
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when update UpdateRecoverStatus, err:%v, name:%s", err, name)
		return
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	pg.Annotations[constant.ProcessRecoverStatusKey] = value
	_, err = kube.RetryUpdatePodGroup(pg, constant.UpdatePodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to update pg when UpdateRecoverStatus, err:%v, name:%s", err, name)
	}
	hwlog.RunLog.Infof("update name=%s recover status=%s, success", name, value)
}

func pullProcessResultFault(name, namespace string) ([]*pb.FaultRank, []*pb.FaultRank, error) {
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when pullProcessResultFault, err:%v, name:%s", err, name)
		return nil, nil, err
	}
	if pg.Annotations == nil {
		hwlog.RunLog.Warnf("pod annotation is nil, name:%s", name)
		return nil, nil, fmt.Errorf("pod annotation is nil, name:%s", name)
	}
	resultRanks, ok := pg.Annotations[constant.ProcessResultFaultKey]
	if !ok {
		hwlog.RunLog.Warnf("can not fiind ProcessResultFaultKey, name:%s", name)
		return nil, nil, fmt.Errorf("processResultFaultKey not exist, name:%s", name)
	}
	rankSlice := strings.Split(resultRanks, ",")
	if len(rankSlice) == 0 {
		err = errors.New("processResultFault lenth is 0")
	}
	confirmRanks, ok := pg.Annotations[constant.ProcessConfirmFaultKey]
	if !ok {
		confirmRanks = ""
	}
	return common.String2Faults(resultRanks), common.String2Faults(confirmRanks), err
}

// WaitProcessResultFault block process until ProcessResultFaultKey's ranks not empty
func WaitProcessResultFault(name, namespace string) ([]*pb.FaultRank, error) {
	startTime := time.Now().Unix()
	resultRanks, confirmRanks, err := pullProcessResultFault(name, namespace)
	for len(resultRanks) == 0 {
		time.Sleep(constant.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > constant.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s ProcessResultFault timeout, timeUse=%d > %d second",
				name, timeUse, constant.ProcessControlTimeout)
			break
		}
		resultRanks, confirmRanks, err = pullProcessResultFault(name, namespace)
	}
	return common.RemoveSliceDuplicateFaults(append(resultRanks, confirmRanks...)), err
}

func rankTableReady(name, namespace string) bool {
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when check rankTableReady, err:%s,name:%s", err, name)
		return false
	}
	if pg.Annotations == nil {
		pg.Annotations = make(map[string]string)
	}
	ready, ok := pg.Annotations[constant.RankTableReadyKey]
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
		time.Sleep(constant.CheckPeriod * time.Second)
		timeUse := time.Now().Unix() - startTime
		if timeUse > constant.ProcessControlTimeout {
			hwlog.RunLog.Warnf("check %s RankTableReady timeout, timeUse=%d > %d second",
				name, startTime, constant.ProcessControlTimeout)
			return fmt.Errorf("check %s RankTableReady timeout, timeUse=%d > %d second",
				name, startTime, constant.ProcessControlTimeout)
		}
		ready = rankTableReady(name, namespace)
	}
	return nil
}
