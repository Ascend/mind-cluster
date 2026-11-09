/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package main
package main

import (
	"errors"

	"ascend-common/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/collector/container"
)

const (
	containerModeDocker     = "docker"
	containerModeContainerd = "containerd"
	containerModeIsula      = "isula"
	containerModeCrio       = "crio"
)

var (
	containerMode string
	containerd    string
	endpoint      string
)

func readCntMonitoringFlags() container.CntNpuMonitorOpts {
	opts := container.CntNpuMonitorOpts{UseOciBackup: true, UseCriBackup: true}
	switch containerMode {
	case containerModeDocker:
		opts.EndpointType = container.EndpointTypeDockerd
		opts.OciEndpoint = container.DefaultDockerAddr
		opts.CriEndpoint = container.DefaultDockerShim
	case containerModeContainerd:
		opts.EndpointType = container.EndpointTypeContainerd
		opts.OciEndpoint = container.DefaultContainerdAddr
		opts.CriEndpoint = container.DefaultContainerdAddr
	case containerModeCrio:
		opts.EndpointType = container.EndpointTypeCrio
		opts.CriEndpoint = container.DefaultCrioAddr
		opts.UseOciBackup = false
		opts.UseCriBackup = false
	case containerModeIsula:
		opts.EndpointType = container.EndpointTypeIsula
		opts.OciEndpoint = container.DefaultIsuladAddr
		opts.CriEndpoint = container.DefaultIsuladAddr
	default:
		hwlog.RunLog.Error("invalid container mode setting,reset to docker")
		opts.EndpointType = container.EndpointTypeDockerd
		opts.OciEndpoint = container.DefaultDockerAddr
		opts.CriEndpoint = container.DefaultDockerShim
	}
	if containerd != "" && containerMode != containerModeCrio {
		opts.OciEndpoint = containerd
		opts.UseOciBackup = false
	}
	if endpoint != "" {
		opts.CriEndpoint = endpoint
		opts.UseCriBackup = false
	}
	return opts
}

func checkCntMonitoringFlags() error {
	if containerMode == containerModeCrio && containerd != "" {
		return errors.New("-containerd is not supported when -containerMode=crio; use -endpoint instead")
	}
	return nil
}
