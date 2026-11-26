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

package containerruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd/oci"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/parser"
)

const (
	defaultCrioAddr    = "/var/run/crio/crio.sock"
	crioNamespace      = "cri-o"
	crioRequestTimeout = 5 * time.Second
	crioPollInterval   = time.Second

	kubeContainerNameLabel = "io.kubernetes.container.name"
	kubePodNameLabel       = "io.kubernetes.pod.name"
	kubePodNamespaceLabel  = "io.kubernetes.pod.namespace"
)

type crioClient struct {
	client            runtimeapi.RuntimeServiceClient
	conn              *grpc.ClientConn
	initialContainers map[string]struct{}
}

// NewCrioClient creates a CRI v1 client without requiring containerd services.
func NewCrioClient(endpoint string) (Client, error) {
	if endpoint == "" {
		endpoint = defaultCrioAddr
	}
	path, err := normalizeCrioSocketPath(strings.TrimPrefix(endpoint, unixPre))
	if err != nil {
		return nil, err
	}
	if err := checkSockFile(path); err != nil {
		return nil, fmt.Errorf("invalid CRI-O endpoint: %w", err)
	}
	conn, err := grpc.NewClient(unixPre+path, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create CRI-O client: %w", err)
	}
	return &crioClient{client: runtimeapi.NewRuntimeServiceClient(conn), conn: conn}, nil
}

func (c *crioClient) runningContainers(ctx context.Context) (map[string]struct{}, error) {
	ctx, cancel := context.WithTimeout(ctx, crioRequestTimeout)
	defer cancel()
	resp, err := c.client.ListContainers(ctx, &runtimeapi.ListContainersRequest{
		Filter: &runtimeapi.ContainerFilter{
			State: &runtimeapi.ContainerStateValue{State: runtimeapi.ContainerState_CONTAINER_RUNNING},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list CRI-O containers: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("empty CRI-O container list response")
	}
	ids := make(map[string]struct{})
	for _, ctr := range resp.Containers {
		if ctr != nil && ctr.Id != "" && ctr.State == runtimeapi.ContainerState_CONTAINER_RUNNING {
			ids[ctr.Id] = struct{}{}
		}
	}
	return ids, nil
}

// ParseAllContainers returns information for running CRI-O containers.
func (c *crioClient) ParseAllContainers(ctx context.Context) (map[string]*types.ContainerNPUInfo, error) {
	ids, err := c.runningContainers(ctx)
	if err != nil {
		return nil, err
	}
	infos := make(map[string]*types.ContainerNPUInfo)
	skipped := 0
	for id := range ids {
		info, err := c.ParseSingleContainer(ctx, id)
		// A status RPC has its own timeout. Only cancellation of the scan context stops the scan.
		if scanErr := ctx.Err(); scanErr != nil {
			return nil, fmt.Errorf("CRI-O container scan interrupted: %w", scanErr)
		}
		if err != nil {
			if IsNotFound(err) {
				skipped++
				hwlog.RunLog.Debugf("skip CRI-O container %s that is no longer running: %v", id, err)
				continue
			}
			hwlog.RunLog.Warnf("failed to parse CRI-O container %s: %v", id, err)
			continue
		}
		infos[id] = info
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("CRI-O container scan interrupted: %w", err)
	}
	if skipped > 0 {
		hwlog.RunLog.Infof("skipped %d CRI-O containers that are no longer running", skipped)
	}
	c.initialContainers = make(map[string]struct{}, len(infos))
	for id := range infos {
		c.initialContainers[id] = struct{}{}
	}
	return infos, nil
}

// ParseSingleContainer reads labels and the verbose OCI specification via CRI.
func (c *crioClient) ParseSingleContainer(ctx context.Context, id string) (*types.ContainerNPUInfo, error) {
	if id == "" {
		return nil, fmt.Errorf("empty CRI-O container ID")
	}
	ctx, cancel := context.WithTimeout(ctx, crioRequestTimeout)
	defer cancel()
	resp, err := c.client.ContainerStatus(ctx, &runtimeapi.ContainerStatusRequest{ContainerId: id, Verbose: true})
	if err != nil {
		return nil, fmt.Errorf("get CRI-O container %s status: %w", id, err)
	}
	if resp == nil || resp.Status == nil || resp.Status.Id == "" {
		return nil, fmt.Errorf("missing CRI-O container status for %s", id)
	}
	if resp.Status.State != runtimeapi.ContainerState_CONTAINER_RUNNING {
		return nil, status.Error(codes.NotFound, "CRI-O container is not running")
	}
	var detail struct {
		RuntimeSpec *oci.Spec `json:"runtimeSpec"`
	}
	if err := json.Unmarshal([]byte(resp.Info["info"]), &detail); err != nil {
		return nil, fmt.Errorf("decode CRI-O container %s info: %w", id, err)
	}
	if detail.RuntimeSpec == nil {
		return nil, fmt.Errorf("missing CRI-O runtimeSpec for %s", id)
	}
	labels := resp.Status.Labels
	info := &types.ContainerNPUInfo{
		ID: resp.Status.Id, Name: labels[kubeContainerNameLabel], Namespace: crioNamespace,
		PodName: labels[kubePodNameLabel], PodNS: labels[kubePodNamespaceLabel],
	}
	if info.Name == "" {
		info.Name = resp.Status.Metadata.GetName()
	}
	if detail.RuntimeSpec.Process != nil {
		envs := detail.RuntimeSpec.Process.Env
		for i := len(envs) - 1; i >= 0; i-- {
			env := strings.TrimSpace(envs[i])
			if strings.HasPrefix(env, api.AscendDeviceInfo+"=") {
				info.Devices = parser.ParseAscendDeviceInfo(env, id)
				break
			}
		}
	}
	if len(info.Devices) == 0 {
		info.Devices = filterNPUDevices(detail.RuntimeSpec)
	}
	return info, nil
}

// WatchContainerEvents polls running containers and emits start/exit transitions.
func (c *crioClient) WatchContainerEvents(ctx context.Context, handler types.EventHandler) {
	if c.conn != nil {
		defer func() {
			if err := c.conn.Close(); err != nil {
				hwlog.RunLog.Warnf("close CRI-O client: %v", err)
			}
		}()
	}
	if handler == nil {
		return
	}
	ticker := time.NewTicker(crioPollInterval)
	defer ticker.Stop()
	c.pollContainerEvents(ctx, handler, ticker.C)
}

func (c *crioClient) pollContainerEvents(ctx context.Context, handler types.EventHandler, ticks <-chan time.Time) {
	// Reconcile against the manager's initial scan so startup transitions are not lost.
	previous := c.initialContainers
	for {
		if ctx.Err() != nil {
			return
		}
		current, err := c.runningContainers(ctx)
		if err != nil {
			hwlog.RunLog.Warnf("poll CRI-O containers: %v", err)
		} else {
			previous = c.reconcileContainers(ctx, previous, current, handler)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticks:
		}
	}
}

func (c *crioClient) reconcileContainers(ctx context.Context, previous, current map[string]struct{},
	handler types.EventHandler) map[string]struct{} {
	// Keep raw lifecycle state separate from parsing success. Temporary status errors
	// must not remove a live container, and failed new containers are retried next poll.
	next := make(map[string]struct{})
	for id := range previous {
		if _, exists := current[id]; exists {
			next[id] = struct{}{}
			continue
		}
		handler(types.ContainerEvent{Type: types.ContainerEventDestroy, ContainerID: id,
			Namespace: crioNamespace, Timestamp: time.Now()})
	}
	for id := range current {
		if _, exists := previous[id]; exists {
			continue
		}
		info, err := c.ParseSingleContainer(ctx, id)
		if err != nil {
			hwlog.RunLog.Warnf("retry new CRI-O container %s on next poll: %v", id, err)
			continue
		}
		handler(types.ContainerEvent{Type: types.ContainerEventCreate, ContainerID: id,
			Namespace: crioNamespace, Timestamp: time.Now(), ContainerInfo: info})
		next[id] = struct{}{}
	}
	return next
}

// normalizeCrioSocketPath only permits the standard root-owned /var/run alias.
// The returned path must still pass checkSockFile before it is used.
func normalizeCrioSocketPath(path string) (string, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("CRI-O endpoint must be a clean absolute Unix socket path")
	}
	if !strings.HasPrefix(path, "/var/run/") {
		return path, nil
	}
	info, err := os.Lstat("/var/run")
	if os.IsNotExist(err) || (err == nil && info.Mode()&os.ModeSymlink == 0) {
		return path, nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect /var/run: %w", err)
	}
	target, err := os.Readlink("/var/run")
	if err != nil || (target != "/run" && target != "../run") {
		return "", fmt.Errorf("CRI-O endpoint only supports the standard /var/run to /run alias")
	}
	owner, ok := info.Sys().(*syscall.Stat_t)
	parent, parentErr := os.Lstat("/var")
	if !ok || owner.Uid != 0 || parentErr != nil || !parent.IsDir() || parent.Mode().Perm()&0022 != 0 {
		return "", fmt.Errorf("unsafe /var/run alias or parent")
	}
	parentOwner, ok := parent.Sys().(*syscall.Stat_t)
	if !ok || parentOwner.Uid != 0 {
		return "", fmt.Errorf("unsafe /var/run parent owner")
	}
	return "/run/" + strings.TrimPrefix(path, "/var/run/"), nil
}
