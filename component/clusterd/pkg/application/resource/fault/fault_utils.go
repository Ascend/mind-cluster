package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"strings"
)

func getDevicesNameOfJobOnNode(nodeName string, serverList []*job.ServerHccl, jobId string) []*job.Device {
	var devices []*job.Device
	found := false
	for _, server := range serverList {
		if server.ServerName != nodeName {
			continue
		}
		found = true
		devices = server.DeviceList
		break
	}
	if !found {
		hwlog.RunLog.Warnf("Job %s may not run on node %s.", jobId, nodeName)
	}
	return devices
}

func getNodesNameFromDeviceInfo(deviceInfos map[string]*constant.DeviceInfo) []string {
	nodesName := make([]string, 0)
	for cmName, _ := range deviceInfos {
		nodeName, err := cmNameToNodeName(cmName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		nodesName = append(nodesName, nodeName)
	}
	return nodesName
}

func cmNameToNodeName(cmName string) (string, error) {
	if !strings.HasPrefix(cmName, constant.DeviceInfoPrefix) {
		return "", fmt.Errorf("cmName has not prefix %s", constant.DeviceInfoPrefix)
	}
	return strings.TrimPrefix(cmName, constant.DeviceInfoPrefix), nil
}

func nodeNameToCmName(nodeName string) string {
	return constant.DeviceInfoPrefix + nodeName
}

func deviceID2DeviceKey(deviceID string) string {
	return constant.AscendDevPrefix + deviceID
}
