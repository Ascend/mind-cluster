// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault public fault collector
package publicfault

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/statistics"
)

var pubFaultInitOnce sync.Once

// PubFaultCollector collect public fault info to cache
func PubFaultCollector(newPubFault *api.PubFaultInfo) error {
	pubFaultInitOnce.Do(func() {
		if err := publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCustomizationPath); err == nil {
			hwlog.RunLog.Infof("load fault config from <%s> success", constant.PubFaultCustomizationName)
			UpdateLimiter()
			return
		}
		hwlog.RunLog.Warnf("load fault config from <%s> failed, begin load from <%s>",
			constant.PubFaultCustomizationName, constant.PubFaultCodeFileName)

		const retryTime = 3
		for i := 0; i < retryTime; i++ {
			var err error
			if err = publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCodeFilePath); err == nil {
				hwlog.RunLog.Infof("load fault config from <%s> success", constant.PubFaultCodeFileName)
				break
			}
			hwlog.RunLog.Warnf("load fault config from <%s> failed, error: %v",
				constant.PubFaultCodeFileName, err)
			time.Sleep(1 * time.Second)
		}
		UpdateLimiter()
	})

	if err := LimitByResource(newPubFault.Resource); err != nil {
		hwlog.RunLog.Errorf("limiter work by resource failed, error: %v", err)
		return errors.New("limiter work by resource failed")
	}
	if err := NewPubFaultInfoChecker(newPubFault).Check(); err != nil {
		hwlog.RunLog.Errorf("check public fault info failed, error: %v", err)
		return fmt.Errorf("check public fault info failed, error: %v", err)
	}
	hwlog.RunLog.Infof("receive public fault, id: %s, resource: %s", newPubFault.Id, newPubFault.Resource)
	for _, fault := range newPubFault.Faults {
		for _, influence := range fault.Influence {
			newFault := convertPubFaultInfoToCache(fault, influence)
			nodeName := getNodeName(influence)
			faultKey := newPubFault.Resource + fault.FaultId
			dealFault(fault.Assertion, nodeName, faultKey, newFault)
		}
	}
	return nil
}

func convertPubFaultInfoToCache(fault api.Fault, influence api.Influence) *constant.PubFaultCache {
	return &constant.PubFaultCache{
		FaultDevIds: influence.DeviceIds,
		FaultId:     fault.FaultId,
		FaultType:   fault.FaultType,
		FaultCode:   fault.FaultCode,
		FaultLevel:  publicfault.GetFaultLevelByCode(fault.FaultCode),
		FaultTime:   fault.FaultTime,
		Assertion:   fault.Assertion,
	}
}

func getNodeName(influence api.Influence) string {
	if influence.NodeName != "" {
		return influence.NodeName
	}
	name, ok := statistics.GetNodeNameBySN(influence.NodeSN)
	if !ok {
		hwlog.RunLog.Error("get node name by sn failed, sn does not exist")
		return ""
	}
	return name
}

func dealFault(assertion, nodeName, faultKey string, newFault *constant.PubFaultCache) {
	const diffTime = 5

	faultExisted, addTime := publicfault.PubFaultCache.FaultExisted(nodeName, faultKey)
	switch assertion {
	case constant.AssertionOccur:
		if !faultExisted {
			publicfault.PubFaultCache.AddPubFaultToCache(newFault, nodeName, faultKey)
		}
	case constant.AssertionRecover:
		if !faultExisted {
			// deal 'recover' after 5 seconds
			dealTime := time.Now().Unix() + diffTime
			publicfault.PubFaultNeedDelete.Push(dealTime, nodeName, faultKey)
			return
		}
		// 5 seconds have passed, delete 'occur'
		if time.Now().Unix()-addTime >= diffTime {
			publicfault.PubFaultCache.DeleteOccurFault(nodeName, faultKey)
			return
		}
		// delete 'recover' after 5 seconds
		deleteTime := addTime + diffTime
		publicfault.PubFaultNeedDelete.Push(deleteTime, nodeName, faultKey)
	default:
		return
	}
}
