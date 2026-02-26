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

package duplicatedetector

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd/namespaces"

	"Ascend-device-plugin/pkg/duplicatedetector/cache"
	"Ascend-device-plugin/pkg/duplicatedetector/containerruntime"
	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
)

var (
	manager *Manager
	once    sync.Once
)

// CheckDuplicateDevices checks for duplicate NPU devices and logs the results
func CheckDuplicateDevices(ctx context.Context, config *types.DetectorConfig) {
	once.Do(func() {
		var err error
		manager, err = NewManager(config)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to create manager: %v", err)
		}
	})
	if manager == nil {
		hwlog.RunLog.Error("manager is nil")
		return
	}
	go manager.Start(ctx)
}

// Manager manages the duplicate NPU device detection functionality
type Manager struct {
	client    containerruntime.Client
	cache     *cache.ContainerCache
	isRunning bool
}

// NewManager creates a new duplicate detection Manager
func NewManager(config *types.DetectorConfig) (*Manager, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	client, err := containerruntime.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container runtime client: %w", err)
	}

	return &Manager{
		client: client,
		cache:  cache.NewContainerCache(),
	}, nil
}

// Start starts the duplicate detection
func (m *Manager) Start(ctx context.Context) {
	if m.isRunning {
		hwlog.RunLog.Error("manager is already running")
		return
	}

	hwlog.RunLog.Info("starting duplicate NPU device detection manager...")

	if err := m.scanAllContainers(ctx); err != nil {
		hwlog.RunLog.Errorf("failed to initialize detector: %v", err)
		return
	}

	go m.watchContainerEvents(ctx)

	m.isRunning = true
	hwlog.RunLog.Info("duplicate NPU device detection manager started successfully")
}

func (m *Manager) watchContainerEvents(ctx context.Context) {
	m.client.WatchContainerEvents(ctx, func(event types.ContainerEvent) {
		switch event.Type {
		case types.ContainerEventCreate:
			hwlog.RunLog.Infof("new container detected: %s", event.ContainerID)
			if err := m.HandleNewContainer(ctx, event.ContainerID, event.Namespace); err != nil {
				hwlog.RunLog.Warnf("failed to handle new container %s: %v", event.ContainerID, err)
			}
		case types.ContainerEventDestroy:
			hwlog.RunLog.Infof("container removed: %s", event.ContainerID)
			m.HandleContainerRemoval(event.ContainerID)
		default:
			hwlog.RunLog.Warnf("unknown event type: %s", event.Type)
		}
	})
}

func (m *Manager) scanAllContainers(ctx context.Context) error {
	hwlog.RunLog.Info("initializing duplicate NPU device detector...")
	result, err := m.client.ParseAllContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse containers: %w", err)
	}
	duplicates := m.cache.StoreAllAndFindDuplicates(result)
	for _, dup := range duplicates {
		m.logDuplicate(dup)
	}

	hwlog.RunLog.Infof("duplicate NPU device detector initialized. Found %d duplicate mount(s)", len(duplicates))
	return nil
}

// HandleNewContainer handles a newly created container
func (m *Manager) HandleNewContainer(ctx context.Context, containerID string, namespace string) error {
	const maxRetries = 5
	const retryDelay = 100 * time.Millisecond

	ctx = namespaces.WithNamespace(ctx, namespace)
	var info *types.ContainerNPUInfo
	var err error

	for i := 0; i < maxRetries; i++ {
		info, err = m.client.ParseSingleContainer(ctx, containerID)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to parse container %s after %d retries: %w", containerID, maxRetries, err)
	}
	if info == nil || len(info.Devices) == 0 {
		return nil
	}
	info.Namespace = namespace

	for _, dup := range m.cache.StoreSingleAndFindDuplicates(info) {
		m.logDuplicate(dup)
	}

	return nil
}

// HandleContainerRemoval handles a container removal event
func (m *Manager) HandleContainerRemoval(containerID string) {
	m.cache.RemoveContainer(containerID)
}

// logDuplicate logs a duplicate mount detection
func (m *Manager) logDuplicate(dup *types.DuplicateMountInfo) {
	const maxPrintLength = 12
	var containerInfos []string
	for _, c := range dup.Containers {
		info := fmt.Sprintf("ID=%s, Name=%s, Namespace=%s", c.ID[:min(maxPrintLength, len(c.ID))], c.Name, c.Namespace)
		if c.PodName != "" {
			info += fmt.Sprintf(", Pod=%s/%s", c.PodNS, c.PodName)
		}
		containerInfos = append(containerInfos, info)
	}

	hwlog.RunLog.Warnf("detected duplicate NPU device mount: device /dev/davinci%d is mounted by multiple containers"+
		": %s",
		dup.DeviceID, strings.Join(containerInfos, "; "))
}
