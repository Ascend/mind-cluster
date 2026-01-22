/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package app implement for interface ContainerClient
package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
)

// DockerClient docker client
type DockerClient struct {
	client *client.Client
}

// NewDockerClient new docker client
func NewDockerClient() *DockerClient {
	return &DockerClient{}
}

func (d *DockerClient) init() error {
	sockURL := fmt.Sprintf("unix://%s", common.ParamOption.SockPath)
	dClient, err := client.NewClientWithOpts(client.WithHost(sockURL), client.WithAPIVersionNegotiation())
	if err != nil {
		hwlog.RunLog.Errorf("connect to container runtime failed, error: %v", err)
		return errors.New("connect to container runtime failed")
	}
	d.client = dClient
	return nil
}

func (d *DockerClient) close() error {
	return d.client.Close()
}

func (d *DockerClient) doStop(containerID, ns string) error {
	return d.client.ContainerStop(context.Background(), containerID, container.StopOptions{})
}

func (d *DockerClient) doStart(containerID, ns string) error {
	return d.client.ContainerStart(context.Background(), containerID, container.StartOptions{})
}

func (d *DockerClient) getAllContainers() (interface{}, error) {
	ctrs, err := d.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		hwlog.RunLog.Errorf("failed to get container list, error: %v", err)
		return nil, errors.New("failed to get container list")
	}
	return ctrs, nil
}

func (d *DockerClient) getUsedDevs(containerObj interface{}, ctx context.Context) ([]int32, error) {
	switch cs := containerObj.(type) {
	case types.Container:
		return d.doGetUsedDevs(cs)
	default:
		return nil, nil
	}
}

func (d *DockerClient) doGetUsedDevs(cs types.Container) ([]int32, error) {
	if strings.Contains(cs.Status, "Exited") {
		return []int32{}, nil
	}
	containerJSON, err := d.client.ContainerInspect(context.Background(), cs.ID)
	if err != nil {
		return nil, err
	}
	for i := len(containerJSON.Config.Env) - 1; i >= 0; i-- {
		env := containerJSON.Config.Env[i]
		if strings.Contains(env, api.AscendDeviceInfo) {
			usedDevs, err := getUsedDevsWithAscendRuntime(env)
			if err != nil {
				return nil, fmt.Errorf("parse env %s failed, error: %v", api.AscendDeviceInfo, err)
			}
			return usedDevs, nil
		}
	}
	hwlog.RunLog.Debugf("get container %s used devs by env %s failed, not used ascend docker runtime",
		cs.ID, api.AscendDeviceInfo)
	usedDevs, err := getUsedDevsWithoutAscendRuntimeForDocker(containerJSON.HostConfig.Resources)
	if err != nil {
		return nil, fmt.Errorf("get container %s device ids failed, error: %v", cs.ID, err)
	}
	return usedDevs, nil
}

func getUsedDevsWithoutAscendRuntimeForDocker(resources container.Resources) ([]int32, error) {
	phyIds := make([]int32, 0, sliceLen16)
	for _, dev := range resources.Devices {
		path := strings.TrimPrefix(dev.PathInContainer, "/dev/")
		id := strings.TrimPrefix(path, "davinci")
		if strings.Contains(id, "_") {
			continue
		}
		phyId, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("get container %s device id failed, error: %v", resources.Devices, err)
		}
		phyIds = append(phyIds, int32(phyId))
	}
	return phyIds, nil
}
