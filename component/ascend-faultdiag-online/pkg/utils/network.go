/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package utils provides some network tools
package utils

import (
	"fmt"
	"net"
	"os"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/constants"
)

// GetNodeIp get the ip address of node pod
func GetNodeIp() (string, error) {
	xdlIp := os.Getenv(constants.XdlIpField)
	if xdlIp != "" {
		if checkIp := net.ParseIP(xdlIp); checkIp != nil && checkIp.To4() != nil {
			return xdlIp, nil
		}
	}

	// no env, output the warn log and get local ip
	hwlog.RunLog.Warnf("[FD-OL]%v environment variable isn't set or isn't a valid IPv4 address", constants.XdlIpField)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		// check the ip address is valid or not
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}

// GetClusterIp get the ip address of cluster pod
func GetClusterIp() string {
	podIP := os.Getenv(constants.PodIP)
	return podIP
}
