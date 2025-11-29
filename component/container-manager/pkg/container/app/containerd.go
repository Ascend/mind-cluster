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
	"math"
	"strconv"
	"strings"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	charDevice    = "c"
	maxDevicesNum = 100000
	maxEnvNum     = 10000
	base          = 10
)

// ContainerdClient containerd client
type ContainerdClient struct {
	client           *containerd.Client
	stoppedContainer map[string]containerd.Container
}

// NewContainerdClient new containerd client
func NewContainerdClient() *ContainerdClient {
	return &ContainerdClient{
		stoppedContainer: make(map[string]containerd.Container),
	}
}

func (c *ContainerdClient) init() error {
	cClient, err := containerd.New(common.ParamOption.SockPath)
	if err != nil {
		hwlog.RunLog.Errorf("connect to container runtime failed, error: %v", err)
		return errors.New("connect to container runtime failed")
	}
	c.client = cClient
	return nil
}

func (c *ContainerdClient) close() error {
	return c.client.Close()
}

func (c *ContainerdClient) getAllContainers() (interface{}, error) {
	ctx := namespaces.WithNamespace(context.Background(), "k8s.io")
	// list running containers
	containers, err := c.client.ContainerService().List(ctx)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get container list, error: %v", err)
		return nil, errors.New("failed to get container list")
	}
	var ctrs = make([]containerd.Container, 0, len(containers))
	for _, container := range containers {
		containerObj, err := c.client.LoadContainer(ctx, container.ID)
		if err != nil {
			hwlog.RunLog.Errorf("failed to load container %s, error: %v", container.ID, err)
			continue
		}
		ctrs = append(ctrs, containerObj)
	}
	return ctrs, nil
}

func (c *ContainerdClient) doStart(containerID, ns string) error {
	ctx := namespaces.WithNamespace(context.Background(), ns)
	container, ok := c.stoppedContainer[containerID]
	if !ok {
		return fmt.Errorf("container %s have not stopped", containerID)
	}
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to new task for container: %v", err)
	}
	if err = task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task for container: %v", err)
	}
	delete(c.stoppedContainer, containerID)
	return nil
}

func (c *ContainerdClient) doStop(containerID, ns string) error {
	ctx := namespaces.WithNamespace(context.Background(), ns)
	container, err := c.client.LoadContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to load container: %v", err)
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get container %s , error: %v", containerID, err)
	}
	if err = task.Kill(ctx, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to kill container %s, error: %v", containerID, err)
	}
	// force kill again to make sure the container is killed
	if err = task.Kill(ctx, syscall.SIGKILL, containerd.WithKillAll); err != nil && errors2.IsNotFound(err) {
		return fmt.Errorf("failed to kill container %s, error: %v", containerID, err)
	}
	if _, err = task.Delete(ctx, containerd.WithProcessKill); err != nil {
		return fmt.Errorf("failed to delete task for container %s, error: %v", containerID, err)
	}
	c.stoppedContainer[containerID] = container
	return nil
}

func (c *ContainerdClient) getUsedDevs(containerObj interface{}, ctx context.Context) ([]int32, error) {
	switch cs := containerObj.(type) {
	case containerd.Container:
		return c.doGetUsedDevs(cs, ctx)
	default:
		return nil, nil
	}
}

func (c *ContainerdClient) doGetUsedDevs(cs containerd.Container, ctx context.Context) ([]int32, error) {
	spec, err := getCtrValidSpec(cs, ctx)
	if err != nil {
		return nil, fmt.Errorf("get container %s valid spec failed, error: %v", cs.ID(), err)
	}
	envs := spec.Process.Env
	// for containerd, env with the same name will be appended directly and will not be overwritten.
	// To avoid the presence of environment variables ASCEND_VISIBLE_DEVICES in the image, iterate from back to front
	for i := len(envs) - 1; i >= 0; i-- {
		env := envs[i]
		if strings.Contains(env, api.AscendDeviceInfo) {
			usedDevs, err := getUsedDevsWithAscendRuntime(env)
			if err != nil {
				return nil, fmt.Errorf("parse env %s failed, error: %v", api.AscendDeviceInfo, err)
			}
			return usedDevs, nil
		}
	}
	hwlog.RunLog.Debugf("get used devs by env %s failed, not used ascend docker runtime", api.AscendDeviceInfo)
	usedDevs, err := getUsedDevsWithoutAscendRuntime(spec)
	if err != nil {
		return nil, fmt.Errorf("get container %s device ids failed, error: %v", cs.ID(), err)
	}
	return usedDevs, nil
}

func getCtrValidSpec(containerObj containerd.Container, ctx context.Context) (*oci.Spec, error) {
	spec, err := containerObj.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec:%v", err)
	}
	if spec.Linux == nil || spec.Linux.Resources == nil || len(spec.Linux.Resources.Devices) > maxDevicesNum {
		return nil, fmt.Errorf("devices in container is too much (%v) or empty", maxDevicesNum)
	}
	if spec.Process == nil || len(spec.Process.Env) > maxEnvNum {
		return nil, fmt.Errorf("env in container is too much (%v) or empty", maxEnvNum)
	}
	return spec, nil
}

func getUsedDevsWithoutAscendRuntime(spec *oci.Spec) ([]int32, error) {
	if spec.Linux == nil || spec.Linux.Resources == nil {
		return nil, errors.New("empty spec info")
	}

	phyIds := make([]int32, 0, sliceLen16)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Resources.Devices {
		if dev.Minor == nil || dev.Major == nil {
			continue
		}
		if *dev.Minor > math.MaxInt32 {
			return nil, fmt.Errorf("get wrong device ID (%v)", dev.Minor)
		}
		major := strconv.FormatInt(*dev.Major, base)
		if dev.Type == charDevice && utils.Contains(majorIDs, major) {
			phyIds = append(phyIds, int32(*dev.Minor))
		}
	}
	return phyIds, nil
}
