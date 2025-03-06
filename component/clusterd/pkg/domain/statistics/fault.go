// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics statistic funcs about fault
package statistics

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/interface/kube"
)

// UpdateFaultToCM update fault cache to configmap statistic-fault-info
func UpdateFaultToCM(faults, faultNum string, exceedsLimiter bool) error {
	cmData := make(map[string]string)
	cmData[constant.StatisticFaultNumKey] = faultNum
	cmData[constant.StatisticPubFaultKey] = faults
	if exceedsLimiter {
		const faultDesc = "The current total number of faults is too large, and only a portion of the detailed fault data is displayed"
		hwlog.RunLog.Errorf("public fault number exceeds the upper limit of %d. "+
			"Will not update the detailed info of the faults this time, only the 'FaultNum' will be updated", constant.MaxFaultNum)
		patchData := map[string]string{constant.StatisticFaultNumKey: faultNum, constant.StatisticFaultDescKey: faultDesc}
		_, err := kube.PatchCMData(constant.StatisticFaultCMName, constant.ClusterNamespace, patchData)
		if err != nil {
			hwlog.RunLog.Errorf("patch cm <%s> data failed, error: %v", constant.StatisticFaultCMName, err)
			return fmt.Errorf("patch cm <%s> data failed", constant.StatisticFaultCMName)
		}
		return nil
	}

	label := map[string]string{constant.CmStatisticFault: constant.CmConsumerValue}
	if err := kube.UpdateOrCreateConfigMap(constant.StatisticFaultCMName, constant.ClusterNamespace,
		cmData, label); err != nil {
		hwlog.RunLog.Errorf("update or create cm <%s> failed, error: %v", constant.StatisticFaultCMName, err)
		return fmt.Errorf("update or create cm <%s> failed", constant.StatisticFaultCMName)
	}
	return nil
}

// LoadFaultFromCM load fault from configmap statistic-fault-info
func LoadFaultFromCM() error {
	cm, err := kube.GetConfigMap(constant.StatisticFaultCMName, constant.ClusterNamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			// If there are no faults in the cluster, cm does not exist
			hwlog.RunLog.Warnf("cm <%s> does not exist, skip loading in cache", constant.StatisticFaultCMName)
			return nil
		}
		hwlog.RunLog.Errorf("get cm <%s> failed, error: %v", constant.StatisticFaultCMName, err)
		return fmt.Errorf("get cm <%s> failed", constant.StatisticFaultCMName)
	}
	data, ok := cm.Data[constant.StatisticPubFaultKey]
	if !ok {
		return fmt.Errorf("statistic fault cm <%s> has no key '%s'",
			constant.StatisticFaultCMName, constant.StatisticPubFaultKey)
	}

	var faults map[string][]constant.NodeFault
	if err = json.Unmarshal([]byte(data), &faults); err != nil {
		hwlog.RunLog.Errorf("unmarshal node fault from cm <%s> failed, error: %v",
			constant.StatisticFaultCMName, err)
		return fmt.Errorf("unmarshal node fault from cm <%s> failed", constant.StatisticFaultCMName)
	}

	publicfault.PubFaultCache.LoadFaultToCache(faults)
	return nil
}
