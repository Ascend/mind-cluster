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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

	"Ascend-device-plugin/pkg/duplicatedetector/cache"
	"Ascend-device-plugin/pkg/duplicatedetector/containerruntime"
	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/common-utils/hwlog"
)

const (
	timeout   = 100 * time.Millisecond
	sleepTime = 50 * time.Millisecond
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

type mockClient struct {
	containers map[string]*types.ContainerNPUInfo
	parseCalls int
	onWatch    func(types.EventHandler)
}

func (m *mockClient) ParseAllContainers(ctx context.Context) (map[string]*types.ContainerNPUInfo, error) {
	return m.containers, nil
}

func (m *mockClient) ParseSingleContainer(ctx context.Context, containerID string) (*types.ContainerNPUInfo, error) {
	m.parseCalls++
	if info, ok := m.containers[containerID]; ok {
		return info, nil
	}
	return nil, nil
}

func (m *mockClient) WatchContainerEvents(ctx context.Context, handler types.EventHandler) {
	if m.onWatch != nil {
		m.onWatch(handler)
	}
}

func TestManager_ContainerEventSnapshot(t *testing.T) {
	for _, wantCalls := range []int{0, 1} {
		snapshot := wantCalls == 0
		info := &types.ContainerNPUInfo{ID: "worker", Devices: []int{6}}
		client := &mockClient{containers: map[string]*types.ContainerNPUInfo{info.ID: info}}
		event := types.ContainerEvent{Type: types.ContainerEventCreate, ContainerID: info.ID, Namespace: "cri-o"}
		if snapshot {
			event.ContainerInfo = info
		}
		manager := &Manager{client: client, cache: cache.NewContainerCache()}
		client.onWatch = func(handler types.EventHandler) {
			handler(event)
			probe := &types.ContainerNPUInfo{ID: "probe", Devices: []int{6}}
			duplicates := manager.cache.StoreSingleAndFindDuplicates(probe)
			if len(duplicates) != 1 || len(duplicates[0].Containers) != 2 || info.Namespace != event.Namespace {
				t.Fatalf("snapshot=%v: created container was not cached: %+v", snapshot, duplicates)
			}
			manager.cache.RemoveContainer(probe.ID)
			handler(types.ContainerEvent{Type: types.ContainerEventDestroy, ContainerID: info.ID})
			if duplicates := manager.cache.StoreSingleAndFindDuplicates(probe); len(duplicates) != 0 {
				t.Fatalf("snapshot=%v: destroyed container remained cached", snapshot)
			}
		}
		manager.watchContainerEvents(context.Background())
		if client.parseCalls != wantCalls {
			t.Fatalf("snapshot=%v: got %d repeated queries, want %d", snapshot, client.parseCalls, wantCalls)
		}
	}
}

func TestNewManager_NilConfig(t *testing.T) {
	_, err := NewManager(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewManager_ValidConfig(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "unix:///run/docker.sock",
		RuntimeType: "docker",
	}
	patch := gomonkey.ApplyFuncReturn(containerruntime.NewClient, nil, errors.New("mock error"))
	defer patch.Reset()
	_, err := NewManager(config)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestManager_Start(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "unix:///run/docker.sock",
		RuntimeType: "docker",
	}
	patch := gomonkey.ApplyFuncReturn(containerruntime.NewClient, &mockClient{}, nil)
	defer patch.Reset()

	manager, _ := NewManager(config)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	manager.Start(ctx)

	if !manager.isRunning {
		t.Error("manager should be running")
	}
}

func TestManager_StartAlreadyRunning(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "docker",
	}
	patch := gomonkey.ApplyFuncReturn(containerruntime.NewClient, &mockClient{}, nil)
	defer patch.Reset()
	manager, _ := NewManager(config)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	manager.Start(ctx)
	manager.Start(ctx)
}

func TestManager_HandleNewContainer_Success(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "docker",
	}
	patch := gomonkey.ApplyFuncReturn(containerruntime.NewClient, &mockClient{
		containers: map[string]*types.ContainerNPUInfo{
			"test-container": {
				ID:      "test-container",
				Devices: []int{0},
			},
		}}, nil)
	defer patch.Reset()

	manager, _ := NewManager(config)

	ctx := context.Background()
	err := manager.HandleNewContainer(ctx, "test-container", "moby")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestManager_HandleNewContainer_NoDevices(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "docker",
	}
	patch := gomonkey.ApplyFuncReturn(containerruntime.NewClient, &mockClient{
		containers: map[string]*types.ContainerNPUInfo{
			"test-container": {
				ID:      "test-container",
				Devices: []int{},
			},
		}}, nil)
	defer patch.Reset()

	manager, _ := NewManager(config)

	ctx := context.Background()
	err := manager.HandleNewContainer(ctx, "test-container", "moby")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckDuplicateDevices_NilConfig(t *testing.T) {
	ctx := context.Background()
	CheckDuplicateDevices(ctx, nil)
}

func TestCheckDuplicateDevices_ValidConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "docker",
	}
	CheckDuplicateDevices(ctx, config)

	time.Sleep(sleepTime)
}

func TestCheckDuplicateDevices_MultipleCalls(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "docker",
	}
	CheckDuplicateDevices(ctx, config)
	CheckDuplicateDevices(ctx, config)

	time.Sleep(sleepTime)
}
