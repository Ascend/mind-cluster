// Copyright 2025 NVIDIA CORPORATION & AFFILIATES
// Modified by Huawei Technologies Co.,Ltd in 2026
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package ub_device for ub device info
package ub_device

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"

	"ascend-common/common-utils/hwlog"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/cdi"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// UbResourceServer implements the K8s device plugin API for UB devices
type UbResourceServer interface {
	types.ResourceServer
	// Additional UB-specific methods can be added here
	MarkDevicesUnhealthy(hcaNames []string) int
}

// ubResourceServer implements UbResourceServer interface
type ubResourceServer struct {
	resourceName   string
	watchMode      bool
	socketName     string
	socketPath     string
	stopWatcher    chan bool
	updateResource chan bool
	health         chan *pluginapi.Device
	rsConnector    types.ResourceServerPort
	rdmaHcaMax     int
	// Mutex protects devs and deviceSpec
	mutex           sync.RWMutex
	devs            []*pluginapi.Device
	deviceSpec      []*pluginapi.DeviceSpec
	ubDevices       []types.Device
	useCdi          bool
	cdi             cdi.CDI
	cdiResourceName string
	exclMode        bool
	// k8sClient queries pod info for NPU-based DPU allocation
	k8sClient *kubernetes.Clientset
}

// resourcesServerPort implements types.ResourceServerPort interface
type resourcesServerPort struct {
	server *grpc.Server
}

// NewUbResourceServer returns an initialized UB device server
func NewUbResourceServer(config *types.UserConfig, devices []types.Device, watcherMode bool,
	socketSuffix string, useCdi bool, exclMode bool, k8sClient *kubernetes.Clientset) (UbResourceServer, error) {

	if config.RdmaHcaMax < 0 {
		return nil, fmt.Errorf("invalid value for rdmaHcaMax < 0: %d", config.RdmaHcaMax)
	}
	if config.ResourcePrefix == "" {
		return nil, fmt.Errorf("empty resourcePrefix")
	}

	deviceSpec := getUbDevicesSpec(devices)

	var devs []*pluginapi.Device
	if exclMode {
		devs = createUbExclDevices(devices)
	} else {
		devs = createUbVirtualDevices(len(devices), deviceSpec, config.ResourceName)
	}

	sockDir := common.DeprecatedSockDir
	if watcherMode {
		sockDir = common.ActiveSockDir
	}

	resourceName := fmt.Sprintf("%s/%s", config.ResourcePrefix, config.ResourceName)
	socketName := fmt.Sprintf("%s.%s", filepath.Base(config.ResourceName), socketSuffix)
	socketPath := filepath.Join(sockDir, socketName)
	hwlog.RunLog.Infof("socketPath: %s", socketPath)

	return &ubResourceServer{
		resourceName:    resourceName,
		watchMode:       watcherMode,
		socketName:      socketName,
		socketPath:      socketPath,
		stopWatcher:     make(chan bool, 1),
		updateResource:  make(chan bool, 1),
		health:          make(chan *pluginapi.Device),
		rsConnector:     &resourcesServerPort{},
		rdmaHcaMax:      config.RdmaHcaMax,
		devs:            devs,
		deviceSpec:      deviceSpec,
		ubDevices:       devices,
		useCdi:          useCdi,
		cdi:             cdi.New(),
		cdiResourceName: config.ResourceName,
		exclMode:        exclMode,
		k8sClient:       k8sClient,
	}, nil
}

func createUbVirtualDevices(deviceCount int, deviceSpec []*pluginapi.DeviceSpec, resourceName string) []*pluginapi.Device {
	if len(deviceSpec) == 0 {
		hwlog.RunLog.Warnf("no devicesSpec, create empty resource server for %s", resourceName)
		return []*pluginapi.Device{}
	}

	devs := make([]*pluginapi.Device, 0, deviceCount)
	for n := 0; n < deviceCount; n++ {
		devs = append(devs, &pluginapi.Device{
			ID:     strconv.Itoa(n),
			Health: pluginapi.Healthy,
		})
	}
	return devs
}

// createUbExclDevices creates devices for exclusive mode.
func createUbExclDevices(devices []types.Device) []*pluginapi.Device {
	devs := make([]*pluginapi.Device, 0, len(devices))
	for _, dev := range devices {
		devs = append(devs, &pluginapi.Device{
			ID:     dev.GetName(),
			Health: pluginapi.Healthy,
		})
	}
	hwlog.RunLog.Infof("exclusive mode: created %d devices from real UB devices", len(devs))
	return devs
}

// reconcileDevs adjusts the virtual device list to match the real device list while
// preserving the Health state of devices that are still present.
func (rs *ubResourceServer) reconcileDevs(newDevices []types.Device) []*pluginapi.Device {
	oldHealth := make(map[string]string, len(rs.ubDevices))
	for i, dev := range rs.ubDevices {
		if i >= len(rs.devs) {
			break
		}
		ubDev, ok := dev.(types.UbDevice)
		if !ok {
			continue
		}
		oldHealth[ubDev.GetDeviceName()] = rs.devs[i].Health
	}

	newDevs := make([]*pluginapi.Device, 0, len(newDevices))
	for i, dev := range newDevices {
		health := pluginapi.Healthy
		if ubDev, ok := dev.(types.UbDevice); ok {
			if h, exist := oldHealth[ubDev.GetDeviceName()]; exist {
				health = h
			}
		}
		newDevs = append(newDevs, &pluginapi.Device{
			ID:     strconv.Itoa(i),
			Health: health,
		})
	}
	return newDevs
}

// Start starts the UB device server
func (rs *ubResourceServer) Start() error {
	_ = rs.cleanup()
	hwlog.RunLog.Infof("starting %s device plugin endpoint at: %s\n", rs.resourceName, rs.socketName)
	rs.rsConnector.CreateServer()
	sock, err := rs.rsConnector.Listen("unix", rs.socketPath)
	if err != nil {
		return err
	}

	if rs.watchMode {
		registerapi.RegisterRegistrationServer(rs.rsConnector.GetServer(), rs)
	}
	pluginapi.RegisterDevicePluginServer(rs.rsConnector.GetServer(), rs)

	rs.rsConnector.Serve(sock)

	conn, err := rs.rsConnector.GetClientConn(rs.socketPath)
	if err != nil {
		return err
	}
	rs.rsConnector.Close(conn)

	hwlog.RunLog.Infof("%s device plugin endpoint started serving", rs.resourceName)

	if !rs.watchMode {
		if err = rs.register(); err != nil {
			rs.rsConnector.Stop()
			return err
		}
	}

	return nil
}

// Stop stops the UB device server
func (rs *ubResourceServer) Stop() error {
	hwlog.RunLog.Infof("stopping %s device plugin server...", rs.resourceName)
	if rs.rsConnector == nil || rs.rsConnector.GetServer() == nil {
		return nil
	}

	if !rs.watchMode {
		select {
		case rs.stopWatcher <- true:
		default:
		}
	}

	rs.rsConnector.Stop()
	rs.rsConnector.DeleteServer()

	return rs.cleanup()
}

// Restart restarts the UB device server
func (rs *ubResourceServer) Restart() error {
	hwlog.RunLog.Infof("restarting %s device plugin server...", rs.resourceName)
	if rs.rsConnector == nil || rs.rsConnector.GetServer() == nil {
		return fmt.Errorf("grpc server instance not found for %s", rs.resourceName)
	}

	rs.rsConnector.Stop()
	rs.rsConnector.DeleteServer()

	return rs.Start()
}

func (rs *ubResourceServer) cleanup() error {
	if err := os.Remove(rs.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (rs *ubResourceServer) register() error {
	kubeletEndpoint := filepath.Join(common.DeprecatedSockDir, common.KubeEndPoint)
	conn, err := rs.rsConnector.GetClientConn(kubeletEndpoint)
	if err != nil {
		return err
	}
	defer rs.rsConnector.Close(conn)

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     rs.socketName,
		ResourceName: rs.resourceName,
	}

	return rs.rsConnector.Register(client, reqt)
}

// Watch for Kubelet socket file; if not present restart server
func (rs *ubResourceServer) Watch() {
	for {
		select {
		case stop := <-rs.stopWatcher:
			if stop {
				hwlog.RunLog.Infof("kubelet watcher stopped for server %s", rs.socketPath)
				return
			}
		default:
			_, err := os.Lstat(rs.socketPath)
			if err != nil {
				hwlog.RunLog.Warnf("server endpoint not found %s", rs.socketName)
				hwlog.RunLog.Warn("most likely Kubelet restarted")
				if err := rs.Restart(); err != nil {
					hwlog.RunLog.Errorf("unable to restart server %v", err)
				}
			}
		}
		time.Sleep(common.WatchWaitTime)
	}
}

// UpdateDevices updates the list of UB devices
func (rs *ubResourceServer) UpdateDevices(devices []types.Device) {
	var needUpdate bool

	rs.mutex.Lock()
	defer func() {
		rs.mutex.Unlock()
		if needUpdate {
			select {
			case rs.updateResource <- true:
			default:
			}
		}
	}()

	newDeviceSpec := getUbDevicesSpec(devices)
	if common.DevicesChanged(rs.deviceSpec, newDeviceSpec) {
		rs.deviceSpec = newDeviceSpec
		// Recreate devs when device spec changes
		if rs.exclMode {
			rs.devs = createUbExclDevices(devices)
		} else {
			rs.devs = rs.reconcileDevs(devices)
		}
		needUpdate = true
	}

	rs.ubDevices = devices
}

// ListAndWatch lists available UB devices and watches for changes
func (rs *ubResourceServer) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	hwlog.RunLog.Infof("ListAndWatch called by kubelet for: %s", rs.resourceName)

	rs.mutex.RLock()
	devs := rs.devs
	hwlog.RunLog.Infof("Updating \"%s\" devices", rs.resourceName)
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
		rs.mutex.RUnlock()
		return err
	}
	hwlog.RunLog.Infof("exposing \"%d\" devices", len(devs))
	rs.mutex.RUnlock()

	// Watch for device updates
	for {
		select {
		case <-rs.updateResource:
			rs.mutex.RLock()
			devs = rs.devs
			hwlog.RunLog.Infof("Updating \"%s\" devices", rs.resourceName)
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
				rs.mutex.RUnlock()
				return err
			}
			hwlog.RunLog.Infof("exposing \"%d\" devices", len(devs))
			rs.mutex.RUnlock()
		case <-rs.health:
			rs.mutex.RLock()
			devs = rs.devs
			hwlog.RunLog.Warnf("Health update triggered for %s, pushing %d devices snapshot",
				rs.resourceName, len(devs))
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
				rs.mutex.RUnlock()
				return err
			}
			rs.mutex.RUnlock()
		case <-s.Context().Done():
			hwlog.RunLog.Infof("ListAndWatch stream closed for: %s", rs.resourceName)
			return nil
		}
	}
}

// GetDevicePluginOptions returns plugin options
func (rs *ubResourceServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired:                false,
		GetPreferredAllocationAvailable: rs.exclMode,
	}, nil
}

// Allocate allocates UB devices to pods
func (rs *ubResourceServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	// Write lock prevents concurrent annotation updates by allocateByNpu.
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	responses := pluginapi.AllocateResponse{}
	if rs.exclMode {
		// Validate devices exist before recording, so a failed Allocate leaves no stale state.
		for _, container := range reqs.ContainerRequests {
			for _, id := range container.DevicesIDs {
				if _, ok := rs.deviceSpecByID(id); !ok {
					return nil, fmt.Errorf("device %q not found in %s devices", id, rs.resourceName)
				}
			}
		}
		// Record the allocation, then mount exactly the devices kubelet assigned.
		if err := rs.allocateByNpu(ctx, allRequestedDeviceIDs(reqs)); err != nil {
			return nil, err
		}
		for _, container := range reqs.ContainerRequests {
			response, err := rs.allocateExclContainer(container)
			if err != nil {
				return nil, err
			}
			responses.ContainerResponses = append(responses.ContainerResponses, response)
		}
	} else {
		// shared mode: mount all devices
		for range reqs.ContainerRequests {
			responses.ContainerResponses = append(responses.ContainerResponses, rs.allocateSharedContainer())
		}
	}
	return &responses, nil
}

// allRequestedDeviceIDs collects the device IDs kubelet assigned to the containers
// of this Allocate request.
func allRequestedDeviceIDs(reqs *pluginapi.AllocateRequest) []string {
	var ids []string
	for _, container := range reqs.ContainerRequests {
		ids = append(ids, container.DevicesIDs...)
	}
	return ids
}

// allocateExclContainer allocates the DPU devices for one container in exclusive mode.
func (rs *ubResourceServer) allocateExclContainer(container *pluginapi.ContainerAllocateRequest) (*pluginapi.ContainerAllocateResponse, error) {
	response := &pluginapi.ContainerAllocateResponse{Envs: map[string]string{}}
	for _, id := range container.DevicesIDs {
		specs, ok := rs.deviceSpecByID(id)
		if !ok {
			return nil, fmt.Errorf("device %q not found in %s devices", id, rs.resourceName)
		}
		response.Devices = append(response.Devices, specs...)
	}
	return response, nil
}

// allocateSharedContainer mounts all UB devices to one container in shared mode
func (rs *ubResourceServer) allocateSharedContainer() *pluginapi.ContainerAllocateResponse {
	response := &pluginapi.ContainerAllocateResponse{Envs: map[string]string{}}
	response.Devices = append(response.Devices, rs.deviceSpec...)
	return response
}

// PreStartContainer performs pre-start operations on containers (not used for UB devices)
func (rs *ubResourceServer) PreStartContainer(ctx context.Context, _ *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// GetPreferredAllocation returns the preferred DPU device IDs to allocate for each container.
func (rs *ubResourceServer) GetPreferredAllocation(ctx context.Context, reqs *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	response := &pluginapi.PreferredAllocationResponse{}
	for _, req := range reqs.ContainerRequests {
		ids, err := rs.preferredDeviceIDs(ctx, req)
		if err != nil {
			return nil, err
		}
		response.ContainerResponses = append(response.ContainerResponses, &pluginapi.ContainerPreferredAllocationResponse{
			DeviceIDs: ids,
		})
	}
	return response, nil
}

// gRPC Plugin Registration API implementations (for watch mode)

// GetInfo returns plugin info
func (rs *ubResourceServer) GetInfo(ctx context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	return &registerapi.PluginInfo{
		Type:              "DevicePlugin",
		Name:              rs.resourceName,
		Endpoint:          rs.socketName,
		SupportedVersions: []string{pluginapi.Version},
	}, nil
}

// NotifyRegistrationStatus notifies plugin of registration status
func (rs *ubResourceServer) NotifyRegistrationStatus(ctx context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	if !status.PluginRegistered {
		hwlog.RunLog.Infof("UB device plugin %s registration failed: %s", rs.resourceName, status.Error)
	}
	return &registerapi.RegistrationStatusResponse{}, nil
}

// Helper functions

// getUbDevicesSpec returns device specs for UB devices
func getUbDevicesSpec(devices []types.Device) []*pluginapi.DeviceSpec {
	devicesSpec := make([]*pluginapi.DeviceSpec, 0)
	for _, device := range devices {
		rdmaDeviceSpec := device.GetRdmaSpec()
		if len(rdmaDeviceSpec) == 0 {
			// Use type assertion to get UB device-specific information
			if ubDevice, ok := device.(types.UbDevice); ok {
				hwlog.RunLog.Warnf("non-Rdma UB Device %s\n", ubDevice.GetUbID())
			} else {
				hwlog.RunLog.Warnf("non-Rdma Device %s\n", device.GetName())
			}
		}
		devicesSpec = append(devicesSpec, rdmaDeviceSpec...)
	}
	return devicesSpec
}

// MarkDevicesUnhealthy marks virtual devices as Unhealthy and triggers ListAndWatch refresh via the health channel.
func (rs *ubResourceServer) MarkDevicesUnhealthy(hcaNames []string) int {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	faultySet := make(map[string]bool, len(hcaNames))
	for _, hca := range hcaNames {
		faultySet[hca] = true
	}

	changed := 0
	for i, dev := range rs.ubDevices {
		ubDev, ok := dev.(types.UbDevice)
		if !ok {
			continue
		}
		isFaulty := faultySet[ubDev.GetDeviceName()]
		if isFaulty {
			if rs.devs[i].Health != pluginapi.Unhealthy {
				rs.devs[i].Health = pluginapi.Unhealthy
				changed++
				hwlog.RunLog.Warnf("Device %s marked as Unhealthy for resource %s (HCA=%s)",
					rs.devs[i].ID, rs.resourceName, ubDev.GetDeviceName())
			}
		} else {
			if rs.devs[i].Health != pluginapi.Healthy {
				rs.devs[i].Health = pluginapi.Healthy
				changed++
				hwlog.RunLog.Infof("Device %s restored to Healthy for resource %s (HCA=%s)",
					rs.devs[i].ID, rs.resourceName, ubDev.GetDeviceName())
			}
		}
	}

	if changed == 0 {
		return 0
	}

	select {
	case rs.health <- &pluginapi.Device{}:
		hwlog.RunLog.Warnf("ListAndWatch refresh signaled for %d device state changes", changed)
	default:
		hwlog.RunLog.Warn("Health channel full, batched state already visible to next ListAndWatch refresh")
	}
	return changed
}

// GetServer listener methods for resourcesServerPort
func (rsc *resourcesServerPort) GetServer() *grpc.Server {
	return rsc.server
}

func (rsc *resourcesServerPort) CreateServer() {
	rsc.server = grpc.NewServer([]grpc.ServerOption{}...)
}

func (rsc *resourcesServerPort) DeleteServer() {
	rsc.server = nil
}

func (rsc *resourcesServerPort) Listen(socketType, socketPath string) (net.Listener, error) {
	// Remove existing socket file
	os.Remove(socketPath)
	// Create Unix socket listener
	listener, err := net.Listen(socketType, socketPath)
	if err != nil {
		return nil, err
	}
	// Set socket permissions
	if err := os.Chmod(socketPath, 0660); err != nil {
		listener.Close()
		return nil, err
	}
	return listener, nil
}

func (rsc *resourcesServerPort) Serve(listener net.Listener) {
	go func() {
		_ = rsc.server.Serve(listener)
	}()
}

func (rsc *resourcesServerPort) Stop() {
	if rsc.server == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		rsc.server.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		hwlog.RunLog.Warnf("gRPC server Stop() timed out after 5s, forcing shutdown")
	}
}

func (rsc *resourcesServerPort) CloseListener(listener net.Listener) {
	if listener != nil {
		listener.Close()
	}
}

func (rsc *resourcesServerPort) Close(conn *grpc.ClientConn) {
	if conn != nil {
		conn.Close()
	}
}

func (rsc *resourcesServerPort) Register(client pluginapi.RegistrationClient, req *pluginapi.RegisterRequest) error {
	_, err := client.Register(context.Background(), req)
	return err
}

func (rsc *resourcesServerPort) GetClientConn(unixSocketPath string) (*grpc.ClientConn, error) {
	var c *grpc.ClientConn
	var err error

	c, err = grpc.NewClient(
		"unix://"+unixSocketPath, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client connection for %s, %w", unixSocketPath, err)
	}

	return c, nil
}
