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

// Package ub_device for ub device info
package ub_device

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"context"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/cdi"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
)

// UbResourceServer is gRPC server implements K8s device plugin api for UB devices
type UbResourceServer interface {
	types.ResourceServer
	// Additional UB-specific methods can be added here
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
}

// resourcesServerPort implements types.ResourceServerPort interface
type resourcesServerPort struct {
	server *grpc.Server
}

// NewUbResourceServer returns an initialized UB device server
func NewUbResourceServer(config *types.UserConfig, devices []types.Device, watcherMode bool,
	socketSuffix string, useCdi bool) (UbResourceServer, error) {

	if config.RdmaHcaMax < 0 {
		return nil, fmt.Errorf("error: Invalid value for rdmaHcaMax < 0: %d", config.RdmaHcaMax)
	}
	if config.ResourcePrefix == "" {
		return nil, fmt.Errorf("error: Empty resourcePrefix")
	}

	deviceSpec := getUbDevicesSpec(devices)
	devs := createUbVirtualDevices(config.RdmaHcaMax, deviceSpec, config.ResourceName)

	sockDir := common.DeprecatedSockDir
	if watcherMode {
		sockDir = common.ActiveSockDir
	}

	resourceName := fmt.Sprintf("%s/%s", config.ResourcePrefix, config.ResourceName)
	socketName := fmt.Sprintf("%s.%s", filepath.Base(config.ResourceName), socketSuffix)
	socketPath := filepath.Join(sockDir, socketName)
	log.Printf("socketPath: %s", socketPath)

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
	}, nil
}

func createUbVirtualDevices(rdmaHcaMax int, deviceSpec []*pluginapi.DeviceSpec, resourceName string) []*pluginapi.Device {
	if len(deviceSpec) == 0 {
		log.Printf("Warning: no devicesSpec, create empty resource server for %s", resourceName)
		return []*pluginapi.Device{}
	}

	devs := make([]*pluginapi.Device, 0, rdmaHcaMax)
	for n := 0; n < rdmaHcaMax; n++ {
		devs = append(devs, &pluginapi.Device{
			ID:     strconv.Itoa(n),
			Health: pluginapi.Healthy,
		})
	}
	return devs
}

// Start starts the UB device server
func (rs *ubResourceServer) Start() error {
	_ = rs.cleanup()
	log.Printf("starting %s device plugin endpoint at: %s\n", rs.resourceName, rs.socketName)
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

	log.Printf("%s device plugin endpoint started serving", rs.resourceName)

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
	log.Printf("stopping %s device plugin server...", rs.resourceName)
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
	log.Printf("restarting %s device plugin server...", rs.resourceName)
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
				log.Printf("kubelet watcher stopped for server %s", rs.socketPath)
				return
			}
		default:
			_, err := os.Lstat(rs.socketPath)
			if err != nil {
				log.Printf("warning: server endpoint not found %s", rs.socketName)
				log.Printf("warning: most likely Kubelet restarted")
				if err := rs.Restart(); err != nil {
					log.Printf("error: unable to restart server %v", err)
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
	if !common.DevicesChanged(rs.deviceSpec, newDeviceSpec) {
		rs.deviceSpec = newDeviceSpec
		needUpdate = true
	}

	rs.ubDevices = devices
}

// gRPC Device Plugin API implementations

// ListAndWatch lists available UB devices and watches for changes
func (rs *ubResourceServer) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Printf("ListAndWatch called by kubelet for: %s", rs.resourceName)

	rs.mutex.RLock()
	devs := rs.devs
	log.Printf("Updating \"%s\" devices", rs.resourceName)
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
		rs.mutex.RUnlock()
		return err
	}
	log.Printf("exposing \"%d\" devices", len(devs))
	rs.mutex.RUnlock()

	// Watch for device updates
	for {
		select {
		case <-rs.updateResource:
			rs.mutex.RLock()
			devs = rs.devs
			log.Printf("Updating \"%s\" devices", rs.resourceName)
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devs}); err != nil {
				rs.mutex.RUnlock()
				return err
			}
			log.Printf("exposing \"%d\" devices", len(devs))
			rs.mutex.RUnlock()
		case <-s.Context().Done():
			log.Printf("ListAndWatch stream closed for: %s", rs.resourceName)
			return nil
		}
	}
}

// GetDevicePluginOptions returns plugin options
func (rs *ubResourceServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

// Allocate allocates UB devices to pods
func (rs *ubResourceServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	for _, _ = range reqs.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{},
		}

		// Add RDMA device specs
		response.Devices = append(response.Devices, rs.deviceSpec...)

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// PreStartContainer performs pre-start operations on containers (not used for UB devices)
func (rs *ubResourceServer) PreStartContainer(ctx context.Context, _ *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// GetPreferredAllocation returns preferred allocation of UB devices
func (rs *ubResourceServer) GetPreferredAllocation(ctx context.Context, _ *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
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
		log.Printf("UB device plugin %s registration failed: %s", rs.resourceName, status.Error)
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
				log.Printf("Warning: non-Rdma UB Device %s\n", ubDevice.GetUbID())
			} else {
				log.Printf("Warning: non-Rdma Device %s\n", device.GetName())
			}
		}
		devicesSpec = append(devicesSpec, rdmaDeviceSpec...)
	}
	return devicesSpec
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
		log.Printf("warning: gRPC server Stop() timed out after 5s, forcing shutdown")
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
