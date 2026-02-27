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
	dtypes "Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type dockerClient struct {
	client *client.Client
	ociClient
}

const (
	defaultDockerAddress = "unix:///run/docker.sock"
	dockerNamespace      = "moby"
	excludePermissions   = 0002
	unixPre              = "unix://"
)

func checkSockFile(path string) error {
	absPath, err := utils.CheckPath(strings.TrimPrefix(path, unixPre))
	if err != nil {
		return err
	}
	return utils.DoCheckOwnerAndPermission(absPath, excludePermissions, 0)
}

// NewDockerClient creates a new docker client
func NewDockerClient(criEndpoint string, ociEndpoint string) (*dockerClient, error) {
	if criEndpoint == "" {
		criEndpoint = defaultDockerAddress
	}
	if err := checkSockFile(criEndpoint); err != nil {
		return nil, fmt.Errorf("invalid cri endpoint(%s): %v", criEndpoint, err)
	}
	if err := checkSockFile(ociEndpoint); err != nil {
		return nil, fmt.Errorf("invalid oci endpoint(%s): %v", ociEndpoint, err)
	}

	cli, err := client.NewClientWithOpts(client.WithHost(criEndpoint), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	ctrClient, err := containerd.New(ociEndpoint)
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		client: cli,
		ociClient: ociClient{
			client: ctrClient,
		},
	}, nil
}

// ParseAllContainers returns all containers
func (d *dockerClient) ParseAllContainers(ctx context.Context) (map[string]*dtypes.ContainerNPUInfo, error) {
	ctrs, err := d.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	containerInfos := make(map[string]*dtypes.ContainerNPUInfo)
	nsCtx := namespaces.WithNamespace(ctx, dockerNamespace)
	for _, ctr := range ctrs {
		info, err := d.ParseSingleContainer(nsCtx, ctr.ID)
		if err != nil {
			hwlog.RunLog.Debugf("Failed to parse container %s: %v", ctr.ID, err)
			continue
		}
		info.PodName = ctr.Labels["io.kubernetes.pod.name"]
		info.PodNS = ctr.Labels["io.kubernetes.pod.namespace"]
		info.Namespace = dockerNamespace
		info.Name = ctr.Labels["io.kubernetes.container.name"]
		containerInfos[ctr.ID] = info
	}
	return containerInfos, nil
}

// ParseSingleContainer returns a single container
func (d *dockerClient) ParseSingleContainer(ctx context.Context, containerID string) (*dtypes.ContainerNPUInfo, error) {
	containerJson, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}
	labels := containerJson.Config.Labels

	info, err := d.parseSingleContainer(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse single container %s: %w", containerID, err)
	}
	info.PodName = labels["io.kubernetes.pod.name"]
	info.PodNS = labels["io.kubernetes.pod.namespace"]
	info.Namespace = dockerNamespace
	info.Name = labels["io.kubernetes.container.name"]
	return info, nil
}

// WatchContainerEvents watches container events
func (d *dockerClient) WatchContainerEvents(ctx context.Context, handler dtypes.EventHandler) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("event", "start")
	filterArgs.Add("event", "die")
	eventChan, errChan := d.client.Events(ctx, types.EventsOptions{
		Filters: filterArgs,
	})
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Debugf("stopping watching container events")
			if err := d.client.Close(); err != nil {
				hwlog.RunLog.Errorf("error closing client: %v", err)
			}
			return
		case event := <-eventChan:
			switch event.Action {
			case "start":
				handler(dtypes.ContainerEvent{
					Type:        dtypes.ContainerEventCreate,
					ContainerID: event.Actor.ID,
					Namespace:   dockerNamespace,
					Timestamp:   time.Now(),
				})

			case "die":
				handler(dtypes.ContainerEvent{
					Type:        dtypes.ContainerEventDestroy,
					ContainerID: event.Actor.ID,
					Namespace:   dockerNamespace,
					Timestamp:   time.Now(),
				})
			default:
				hwlog.RunLog.Warnf("unknown event type: %T", event)
			}

		case err := <-errChan:
			hwlog.RunLog.Errorf("error receiving event: %v", err)
		}
	}
}
