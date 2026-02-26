/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package containerruntime is the client for interacting with docker and containerd runtime

package containerruntime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/parser"
)

type Client interface {
	ParseAllContainers(ctx context.Context) (map[string]*types.ContainerNPUInfo, error)
	ParseSingleContainer(ctx context.Context, containerID string) (*types.ContainerNPUInfo, error)
	WatchContainerEvents(ctx context.Context, handler types.EventHandler)
}

const (
	defaultContainerdAddr = "/run/containerd/containerd.sock"
	dockerContainerdAddr  = "/run/docker/containerd/containerd.sock"
)

type ociClient struct {
	client *containerd.Client
}

func (c *ociClient) parseSingleContainer(ctx context.Context, containerID string) (*types.ContainerNPUInfo, error) {
	task, err := c.client.TaskService().Get(ctx, &tasks.GetRequest{ContainerID: containerID})
	if err != nil {
		return nil, err
	}

	if task.GetProcess() == nil {
		return nil, fmt.Errorf("task not found for container %s", containerID)
	}

	ctr, err := c.client.LoadContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	spec, err := ctr.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}
	labels, err := ctr.Labels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container labels: %w", err)
	}
	info := &types.ContainerNPUInfo{
		ID:      ctr.ID(),
		Name:    labels["io.kubernetes.container.name"],
		PodName: labels["io.kubernetes.pod.name"],
		PodNS:   labels["io.kubernetes.pod.namespace"],
	}

	if spec.Process != nil {
		for i := len(spec.Process.Env) - 1; i >= 0; i-- {
			env := strings.TrimSpace(spec.Process.Env[i])
			if strings.Contains(env, api.AscendDeviceInfo) {
				info.Devices = parser.ParseAscendDeviceInfo(env, ctr.ID())
				break
			}
		}
	}
	if len(info.Devices) != 0 {
		return info, nil
	}

	info.Devices = parser.FilterNPUDevices(spec)
	return info, nil
}

// NewClient creates a new containerd client, If endpoint is empty, it will auto-detect the containerd socket path
func NewClient(config *types.DetectorConfig) (Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	ociEndpoint, err := autoDetectOciEndpoint()
	if err != nil {
		return nil, err
	}
	if config.RuntimeType == kubeclient.DockerRuntime {
		hwlog.RunLog.Info("using docker runtime")
		return NewDockerClient(config.CriEndpoint, ociEndpoint)

	}
	if config.RuntimeType == kubeclient.ContainerdRuntime {
		hwlog.RunLog.Info("using containerd runtime")
		return NewContainerdClient(config.CriEndpoint, ociEndpoint)
	}
	return nil, fmt.Errorf("runtime type %s is not supported", config.RuntimeType)
}

func autoDetectOciEndpoint() (string, error) {
	// Check default K8s oci path
	if _, err := os.Stat(defaultContainerdAddr); err == nil {
		hwlog.RunLog.Infof("auto-detected oci socket at: %s", defaultContainerdAddr)
		return defaultContainerdAddr, nil
	}

	// Check Docker containerd path
	if _, err := os.Stat(dockerContainerdAddr); err == nil {
		hwlog.RunLog.Infof("auto-detected oci socket at: %s", dockerContainerdAddr)
		return dockerContainerdAddr, nil
	}

	return "", errors.New("failed to auto-detect oci socket path")
}
