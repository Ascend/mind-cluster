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
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl/v2"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
)

type containerdClient struct {
	*ociClient
}

func NewContainerdClient(criEndpoint string, ociEndpoint string) (Client, error) {
	if err := checkSockFile(ociEndpoint); err != nil {
		return nil, err
	}
	if criEndpoint == "" {
		criEndpoint = ociEndpoint
	} else {
		if err := checkSockFile(criEndpoint); err != nil {
			return nil, err
		}
	}

	cli, err := containerd.New(criEndpoint)
	if err != nil {
		return nil, err
	}
	return &containerdClient{
		ociClient: &ociClient{client: cli},
	}, nil
}

func (c *containerdClient) ParseAllContainers(ctx context.Context) (map[string]*types.ContainerNPUInfo, error) {
	nss, err := c.client.NamespaceService().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	containerInfos := make(map[string]*types.ContainerNPUInfo)
	for _, ns := range nss {
		nsCtx := namespaces.WithNamespace(ctx, ns)
		containers, err := c.client.Containers(nsCtx)
		if err != nil {
			hwlog.RunLog.Warnf("failed to list containers in namespace %s: %v", ns, err)
			continue
		}
		for _, ctr := range containers {
			info, err := c.ParseSingleContainer(nsCtx, ctr.ID())
			if err != nil {
				hwlog.RunLog.Warnf("failed to parse container %s: %v", ctr.ID(), err)
				continue
			}
			info.Namespace = ns
			containerInfos[ctr.ID()] = info
		}
	}

	return containerInfos, nil
}

func (c *containerdClient) ParseSingleContainer(ctx context.Context, containerID string) (*types.ContainerNPUInfo, error) {
	return c.parseSingleContainer(ctx, containerID)
}

func (c *containerdClient) WatchContainerEvents(ctx context.Context, handler types.EventHandler) {
	eventChan, errChan := c.client.EventService().Subscribe(ctx,
		`topic~="/tasks/start"`,
		`topic~="/tasks/exit"`,
	)
	for {
		select {
		case <-ctx.Done():
			if err := c.client.Close(); err != nil {
				hwlog.RunLog.Errorf("failed to close containerd client: %v", err)
			}
			return
		case envelope := <-eventChan:
			if envelope.Event == nil {
				continue
			}
			v, err := typeurl.UnmarshalAny(envelope.Event)
			if err != nil {
				hwlog.RunLog.Warnf("failed to unmarshal event: %v", err)
				continue
			}
			switch event := v.(type) {
			case *events.TaskStart:
				handler(types.ContainerEvent{
					Type:        types.ContainerEventCreate,
					ContainerID: event.ContainerID,
					Namespace:   envelope.Namespace,
					Timestamp:   time.Now(),
				})
			case *events.TaskExit:
				handler(types.ContainerEvent{
					Type:        types.ContainerEventDestroy,
					ContainerID: event.ContainerID,
					Namespace:   envelope.Namespace,
					Timestamp:   time.Now(),
				})
			default:
				hwlog.RunLog.Warnf("unknown event type: %T", event)
			}
		case err := <-errChan:
			hwlog.RunLog.Warnf("error receiving event: %v", err)
		}
	}
}
