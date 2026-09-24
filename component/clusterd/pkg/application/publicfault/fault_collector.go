// Copyright (c) Huawei Technologies Co., Ltd. 2025-2026. All rights reserved.

// Package publicfault public fault collector
package publicfault

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/statistics"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/publicfault"
)

var pubFaultInitOnce sync.Once

// silentFaultHandler silent-fault-level message handler (registered by application/manualfault at init,
// to avoid an import cycle from the publicfault -> silentfault reverse dependency)
var silentFaultHandler func(event SilentFaultEvent)

// SilentFaultEvent silent-fault-level message event (occur/recover)
type SilentFaultEvent struct {
	Assertion string // occur / recover / once
	NodeName  string
	Resource  string // message source (clusterd or an external source)
	FaultId   string
	DevIds    []int32
}

// RegisterSilentFaultHandler registers the handler for silent-fault-level messages
func RegisterSilentFaultHandler(handler func(event SilentFaultEvent)) {
	silentFaultHandler = handler
}

// InitPubFaultCfg ensures the public fault code/resource config is loaded exactly once.
// Call it at startup so code-level checks (e.g. silent fault detection) see a populated
// cache before the first fault is processed.
func InitPubFaultCfg() {
	pubFaultInitOnce.Do(loadPubFaultCfg)
}

// loadPubFaultCfg loads the customization file first and falls back to the default file on failure
func loadPubFaultCfg() {
	hwlog.RunLog.Infof("start trying load public fault config from file %s",
		constant.PubFaultCustomizationPath)
	if err := publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCustomizationPath); err == nil {
		UpdateLimiter()
		return
	}
	hwlog.RunLog.Warnf("load fault config from <%s> failed, begin load from <%s>",
		constant.PubFaultCustomizationName, constant.PubFaultCodeFileName)

	const retryTime = 3
	for i := 0; i < retryTime; i++ {
		var err error
		if err = publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCodeFilePath); err == nil {
			break
		}
		hwlog.RunLog.Warnf("load fault config from <%s> failed, error: %v",
			constant.PubFaultCodeFileName, err)
		time.Sleep(1 * time.Second)
	}
	UpdateLimiter()
}

// PubFaultCollector collect public fault info to cache
func PubFaultCollector(newPubFault *api.PubFaultInfo) error {
	if err := LimitByResource(newPubFault.Resource); err != nil {
		hwlog.RunLog.Errorf("limiter work by resource failed, error: %v", err)
		return errors.New("limiter work by resource failed")
	}
	if err := NewPubFaultInfoChecker(newPubFault).CheckAndFlush(); err != nil {
		hwlog.RunLog.Errorf("check public fault info failed, error: %v", err)
		return fmt.Errorf("check public fault info failed, error: %v", err)
	}
	hwlog.RunLog.Infof("receive public fault, id: %s, resource: %s, timestamp: %d",
		newPubFault.Id, newPubFault.Resource, newPubFault.TimeStamp)
	for _, fault := range newPubFault.Faults {
		// silent fault switch off: drop silent-level faults from any resource
		// (internal detection messages and external gRPC reports), do not
		// process or cache them to avoid leftover misjudged entries
		if isSilentFaultDisabled(fault.FaultCode) {
			hwlog.RunLog.Infof("skip silent fault, id: %s, code: %s, resource: %s: silent fault switch is off",
				fault.FaultId, fault.FaultCode, newPubFault.Resource)
			continue
		}
		hwlog.RunLog.Infof("faultId: %s, faultType: %s, faultCode: %s, faultTime: %d, assertion: %s",
			fault.FaultId, fault.FaultType, fault.FaultCode, fault.FaultTime, fault.Assertion)
		for _, influence := range fault.Influence {
			newFault := convertPubFaultInfoToCache(fault, influence)
			nodeName := getNodeName(influence)
			faultKey := newPubFault.Resource + fault.FaultId
			dealFault(fault.Assertion, nodeName, faultKey, newFault)
			// level routing: forward silent-fault-level occur/recover messages to the silent-side handler (the source of truth is
			// the level check at the routing site; when the switch is off, isSilentFaultDisabled above already drops them one by one, so this is a double guard)
			if newFault.FaultLevel == constant.SilentFault && conf.GetSilentFaultEnabled() && silentFaultHandler != nil {
				hwlog.RunLog.Infof("route silent fault message to handler: assertion %s, node %s, resource %s, faultId %s",
					fault.Assertion, nodeName, newPubFault.Resource, fault.FaultId)
				silentFaultHandler(SilentFaultEvent{
					Assertion: fault.Assertion,
					NodeName:  nodeName,
					Resource:  newPubFault.Resource,
					FaultId:   fault.FaultId,
					DevIds:    influence.DeviceIds,
				})
			}
		}
	}
	// a public fault message, update statistic configmap once
	statistics.StatisticFault.Notify()
	return nil
}

// isSilentFaultDisabled returns true when the fault is silent-level and the
// silent fault switch is off. In that case the fault must not be reported from
// any path (internal detection or external gRPC report)
func isSilentFaultDisabled(faultCode string) bool {
	return publicfault.GetFaultLevelByCode(faultCode) == constant.SilentFault &&
		!conf.GetSilentFaultEnabled()
}

func convertPubFaultInfoToCache(fault api.Fault, influence api.Influence) *constant.PubFaultCache {
	const convertMSToSec = 1000
	return &constant.PubFaultCache{
		FaultDevIds: influence.DeviceIds,
		FaultId:     fault.FaultId,
		FaultType:   fault.FaultType,
		FaultCode:   fault.FaultCode,
		FaultLevel:  publicfault.GetFaultLevelByCode(fault.FaultCode),
		FaultTime:   fault.FaultTime / convertMSToSec,
		Assertion:   fault.Assertion,
	}
}

func getNodeName(influence api.Influence) string {
	if influence.NodeName != "" {
		return influence.NodeName
	}
	name, ok := node.GetNodeNameBySN(influence.NodeSN)
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
			PubFaultNeedDelete.Push(dealTime, nodeName, faultKey)
			return
		}
		// 5 seconds have passed, delete 'occur'
		if time.Now().Unix()-addTime >= diffTime {
			publicfault.PubFaultCache.DeleteOccurFault(nodeName, faultKey)
			return
		}
		// delete 'recover' after 5 seconds
		deleteTime := addTime + diffTime
		PubFaultNeedDelete.Push(deleteTime, nodeName, faultKey)
	case constant.AssertionOnce:
		if !faultExisted {
			newFault.Assertion = constant.AssertionOccur
			publicfault.PubFaultCache.AddPubFaultToCache(newFault, nodeName, faultKey)
			// deal 'once' after 5 seconds
			dealTime := time.Now().Unix() + diffTime
			// push 'occur' to PubFaultNeedDelete
			PubFaultNeedDelete.Push(dealTime, nodeName, faultKey)
		}
	default:
		return
	}
}
