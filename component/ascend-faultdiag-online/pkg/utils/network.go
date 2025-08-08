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
	// 获取环境变量 XDL_IP
	xdlIp := os.Getenv(constants.XdlIpField)
	// 如果环境变量存在，直接返回
	if xdlIp != "" {
		return xdlIp, nil
	}

	// 如果没有环境变量，输出警告并调用 GetLocalIP 获取本地 IP
	hwlog.RunLog.Warnf("%v environment variable not set", constants.XdlIpField)
	// 获取本地所有网络接口的地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 遍历所有地址，查找 IPv4 地址
	for _, addr := range addrs {
		// 检查是否为 IP 地址，并且是否为 IPv4 地址
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			// 返回第一个找到的非 loopback 的 IPv4 地址
			return ipNet.IP.String(), nil
		}
	}

	// 如果没有找到有效的 IPv4 地址，返回错误
	return "", fmt.Errorf("no valid IP address found")
}

// GetClusterIp get the ip address of cluster pod
func GetClusterIp() string {
	podIP := os.Getenv(constants.PodIP)
	return podIP
}
