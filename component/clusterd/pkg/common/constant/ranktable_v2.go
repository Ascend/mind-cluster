// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

package constant

import (
	"errors"
	"fmt"

	"ascend-common/common-utils/hwlog"
)

// Values copied from component/ascend-operator/pkg/ranktable/v2dot0/ranktable.go and
// pkg/api/v1/constants_v2.go. Keep in sync with device-plugin's UPPERCASE values.

const (
	// ScaleOutTypeLabel is the label key that surfaces the large mesh type (e.g. RoCE, UBoE), case-insensitive.
	ScaleOutTypeLabel = "scaleout-type"
	// ScaleOutTypeRoCE is the label value of task for RoCE, kept in uppercase format.
	ScaleOutTypeRoCE = "ROCE"
	// ScaleOutTypeUBoE is the label value of task for UBoE, kept in uppercase format.
	ScaleOutTypeUBoE = "UBOE"

	// PortAddrTypeRoCE is the RoCE port address type, kept in uppercase format to match device-plugin.
	PortAddrTypeRoCE = "ROCE"
	// PortAddrTypeUBoE is the UBoE port address type, kept in uppercase format to match device-plugin.
	PortAddrTypeUBoE = "UBOE"
	// PortAddrTypeUBG is the UBG port address type, kept in uppercase format to match device-plugin.
	PortAddrTypeUBG = "UBG"

	// RankAddrTypeIP is the rank address type IP, kept in uppercase format.
	RankAddrTypeIP = "IP"
	// RankAddrTypeEID is the rank address type EID, kept in uppercase format.
	RankAddrTypeEID = "EID"
)

type NetInfo struct {
	PortAddrType string
	ScaleOutType string
	RankAddrType string
}

var defaultPriorityScaleOutType = []string{PortAddrTypeRoCE, PortAddrTypeUBoE, PortAddrTypeUBG}

var portTypeMappings = map[string]NetInfo{
	PortAddrTypeRoCE: {PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP},
	PortAddrTypeUBoE: {PortAddrTypeUBoE, ScaleOutTypeUBoE, RankAddrTypeIP},
	PortAddrTypeUBG:  {PortAddrTypeUBG, ScaleOutTypeUBoE, RankAddrTypeEID},
}

var scaleOutTypeMappings = map[string][]NetInfo{
	ScaleOutTypeRoCE: {{PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP}},
	ScaleOutTypeUBoE: {
		{PortAddrTypeUBoE, ScaleOutTypeUBoE, RankAddrTypeIP},
		{PortAddrTypeUBG, ScaleOutTypeUBoE, RankAddrTypeEID},
	},
}

// GetNetInfoByDefault selects the network info policy by default priority (RoCE > UBoE > UBG).
func GetNetInfoByDefault(portAddrTypes map[string]struct{}) (NetInfo, error) {
	for _, sot := range defaultPriorityScaleOutType {
		if _, exist := portAddrTypes[sot]; !exist {
			continue
		}
		if netInfo, exist := portTypeMappings[sot]; exist {
			return netInfo, nil
		}
	}
	hwlog.RunLog.Warn("no suitable port addr type found")
	return NetInfo{}, nil
}

// GetNetInfoByCustom selects the network info policy by user-specified scale-out type.
func GetNetInfoByCustom(portAddrTypes map[string]struct{}, customScaleOutType string) (NetInfo, error) {
	infoList, exist := scaleOutTypeMappings[customScaleOutType]
	if !exist {
		errMsg := fmt.Sprintf("the value of label %s is invalid, which should be %s or %s",
			ScaleOutTypeLabel, ScaleOutTypeRoCE, ScaleOutTypeUBoE)
		hwlog.RunLog.Error(errMsg)
		return NetInfo{}, errors.New(errMsg)
	}
	for _, item := range infoList {
		if _, ok := portAddrTypes[item.PortAddrType]; ok {
			return item, nil
		}
	}
	hwlog.RunLog.Warnf("no suitable port addr type found in device for the custom %v label value %s",
		ScaleOutTypeLabel, customScaleOutType)
	return NetInfo{}, nil
}

// GetNetInfo selects the network info policy (custom if set, else default priority).
func GetNetInfo(portAddrTypes map[string]struct{}, customScaleOutType string) (NetInfo, error) {
	if len(customScaleOutType) == 0 {
		return GetNetInfoByDefault(portAddrTypes)
	}
	return GetNetInfoByCustom(portAddrTypes, customScaleOutType)
}
