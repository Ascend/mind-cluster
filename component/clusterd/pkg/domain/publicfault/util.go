// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault utils for public fault
package publicfault

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

// ParsePubFaultCM parse obj to PubFaultCM
func ParsePubFaultCM(obj interface{}) (*api.PubFaultInfo, error) {
	pubFaultCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return nil, errors.New("input is not a valid cm")
	}

	data, ok := pubFaultCm.Data[constant.PubFaultCMKey]
	if !ok {
		return nil, fmt.Errorf("public fault cm <%s> has no key '%s'", pubFaultCm.Name, constant.PubFaultCMKey)
	}

	var pubFaultInfo api.PubFaultInfo
	if err := json.Unmarshal([]byte(data), &pubFaultInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal public fault cm <%s> failed, error: %v", pubFaultCm.Name, err)
		return nil, fmt.Errorf("unmarshal public fault cm <%s> failed", pubFaultCm.Name)
	}
	return &pubFaultInfo, nil
}

// DeepCopy deep copy public fault info
func DeepCopy(pubFaultInfo *api.PubFaultInfo) *api.PubFaultInfo {
	if pubFaultInfo == nil {
		return nil
	}
	data, err := json.Marshal(pubFaultInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal public fault info failed when deepcopy, error: %v", err)
		return nil
	}
	var newPubFaultInfo api.PubFaultInfo
	if err = json.Unmarshal(data, &newPubFaultInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal public fault info failed when deepcopy, error: %v", err)
		return nil
	}
	return &newPubFaultInfo
}

// GetFaultLevelByCode get fault level by fault code
// if one fault code corresponds to multiple fault levels, the most severe one will be dealt with
func GetFaultLevelByCode(faultCode string) string {
	_, ok := PubFaultCodeCfg.SeparateNPUCodes[faultCode]
	if ok {
		return constant.SeparateNPU
	}
	_, ok = PubFaultCodeCfg.SubHealthFaultCodes[faultCode]
	if ok {
		return constant.SubHealthFault
	}
	_, ok = PubFaultCodeCfg.NotHandleFaultCodes[faultCode]
	if ok {
		return constant.NotHandleFault
	}
	return ""
}
