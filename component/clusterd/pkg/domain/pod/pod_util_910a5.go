// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package pod a series of pod util function
package pod

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

// defaultScaleOutType defines the default order of scale-out network types.
var defaultScaleOutType = []string{api.ScaleOutTypeRoCE, api.ScaleOutTypeUBoE, api.ScaleOutTypeUBG}

// adjustScaleOutTypeForStacking enforces RoCE for stacking servers, overriding UBoE if necessary.
func adjustScaleOutTypeForStacking(normalized string, podsInJob map[string]v1.Pod) string {
	for _, pod := range podsInJob {
		if pod.Labels[api.LabelReplicaType] != api.ReplicaTypeMaster ||
			pod.Spec.NodeSelector[api.AcceleratorTypeKey] != api.Ascend800ia5Stacking {
			continue
		}
		if normalized == api.ScaleOutTypeUBoE {
			hwlog.RunLog.Warnf("pob(%s) custom scale-out type is UBoE, but stacking servers only support RoCE",
				pod.Name)
			return api.ScaleOutTypeRoCE
		}
	}
	return normalized
}

func getScaleOutType(podGroup v1beta1.PodGroup, podsInJob map[string]v1.Pod) (string, error) {
	scaleOutType, ok := podGroup.Labels[api.ScaleOutType]
	if !ok {
		hwlog.RunLog.Debugf("label %s is not set, use default %s logic", api.ScaleOutType, api.ScaleOutType)
		return adjustScaleOutTypeForStacking("", podsInJob), nil
	}

	normalized := strings.ToUpper(scaleOutType)

	switch normalized {
	case "", api.ScaleOutTypeRoCE, api.ScaleOutTypeUBoE:
		hwlog.RunLog.Infof("the value of label %s is %s", api.ScaleOutType, normalized)
		return adjustScaleOutTypeForStacking(normalized, podsInJob), nil
	default:
		errMsg := fmt.Sprintf("the value of label %s is invalid, should be %s or %s",
			api.ScaleOutType, api.ScaleOutTypeRoCE, api.ScaleOutTypeUBoE)
		hwlog.RunLog.Errorf(errMsg)
		return "", fmt.Errorf(errMsg)
	}
}

// setScaleOutNetwork sets the ScaleOutNetwork of serverDev based on device info and scaleOutType
func setScaleOutNetwork(dev constant.Device, scaleOutType string, serverDev *constant.Device) {
	if len(dev.LevelList) == 0 {
		hwlog.RunLog.Debugf("empty LevelList, skip setting scale_out_network, deviceID=%s", dev.DeviceID)
		return
	}

	portMap := collectPorts(dev)
	selectScaleOutNetwork(portMap, scaleOutType, serverDev)
}

// collectPorts gathers all valid RankAddrItem addresses by netType from the device LevelList.
func collectPorts(dev constant.Device) map[string][]api.RankAddrItem {
	portMap := make(map[string][]api.RankAddrItem)

	for _, level := range dev.LevelList {
		for netTypeRaw, elem := range level.Info {
			netType := strings.ToUpper(netTypeRaw)

			if !isValidNetType(level.Level, netType) {
				continue
			}

			// collect valid addresses
			validAddrs := collectValidAddrs(level.Level, netType, elem.RankAddrList)
			portMap[netType] = append(portMap[netType], validAddrs...)
		}
	}
	return portMap
}

// collectValidAddrs filters and returns valid RankAddrItem addresses for a given level/netType.
func collectValidAddrs(level int, netType string, addrList []api.RankAddrItem) []api.RankAddrItem {
	var result []api.RankAddrItem
	for _, addr := range addrList {
		if addr.Addr == "" {
			hwlog.RunLog.Warnf("skip empty addr: level=%d, netType=%s, addrType=%s",
				level, netType, addr.AddrType)
			continue
		}
		result = append(result, addr)
		hwlog.RunLog.Debugf("collect port: level=%d, netType=%s, addrType=%s, addr=%s",
			level, netType, addr.AddrType, addr.Addr)
	}
	return result
}

// isValidNetType checks whether the given netType should be included for the level
func isValidNetType(level int, netType string) bool {
	switch level {
	case api.Level2:
		return netType == api.ScaleOutTypeUBoE || netType == api.ScaleOutTypeUBG
	case api.Level3:
		return netType == api.ScaleOutTypeRoCE
	default:
		return false
	}
}

// selectScaleOutNetwork chooses the proper network from portMap based on scaleOutType
func selectScaleOutNetwork(portMap map[string][]api.RankAddrItem, scaleOutType string, serverDev *constant.Device) {
	switch strings.ToUpper(scaleOutType) {
	case "":
		for _, scaleOut := range defaultScaleOutType {
			if ports, ok := portMap[scaleOut]; ok {
				handleScaleOutNetworkInfo(serverDev, ports)
				return
			}
		}
	case api.ScaleOutTypeRoCE:
		if ports, ok := portMap[api.ScaleOutTypeRoCE]; ok {
			handleScaleOutNetworkInfo(serverDev, ports)
			return
		}
	case api.ScaleOutTypeUBoE:
		if ports, ok := portMap[api.ScaleOutTypeUBoE]; ok {
			handleScaleOutNetworkInfo(serverDev, ports)
			return
		}
		if ports, ok := portMap[api.ScaleOutTypeUBG]; ok {
			handleScaleOutNetworkInfo(serverDev, ports)
			return
		}
	default:
		hwlog.RunLog.Errorf("invalid scaleOutType=%s", scaleOutType)
		return
	}

	hwlog.RunLog.Warnf("no suitable port netType found, keys=%v, custom label %s",
		func(m map[string][]api.RankAddrItem) []string {
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			return keys
		}(portMap), scaleOutType)
}

func handleScaleOutNetworkInfo(serverDev *constant.Device, ports []api.RankAddrItem) {
	if len(ports) == 0 {
		hwlog.RunLog.Warnf("handleScaleOutNetworkInfo called with empty ports")
		return
	}

	addrs := make([]string, 0, len(ports))
	addrType := api.AddrTypeIPV4

	for _, port := range ports {
		if strings.ToUpper(port.AddrType) == api.AddrTypeEID {
			addrType = api.AddrTypeEID
		}
		addrs = append(addrs, port.Addr)
	}

	serverDev.ScaleOutNetwork = &constant.ScaleOutNetwork{
		AddrType: addrType,
		Addrs:    addrs,
	}

	hwlog.RunLog.Infof("set scale_out_network: addr_type=%s addrs=%v", addrType, addrs)
}
