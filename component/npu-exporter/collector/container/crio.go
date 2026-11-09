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

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	criv1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/v6/collector/container/v1"
)

// CRI-O encodes verbose ContainerStatus data as JSON in Info["info"].
// Keep pid raw so an unavailable PID cannot prevent device discovery.
type crioContainerInfo struct {
	RuntimeSpec *v1.Spec        `json:"runtimeSpec"`
	PID         json.RawMessage `json:"pid"`
}

func (operator *RuntimeOperatorTool) getCrioContainerInfo(ctx context.Context, id string) (*crioContainerInfo, error) {
	if operator.criConn == nil {
		return nil, errors.New("CRI-O client is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if operator.criVersion != criVersionV1alpha2 {
		resp, err := criv1.NewRuntimeServiceClient(operator.criConn).ContainerStatus(ctx,
			&criv1.ContainerStatusRequest{ContainerId: id, Verbose: true})
		if err == nil {
			return parseCrioContainerInfo(resp.GetInfo())
		}
		// Only protocol incompatibility permits a fallback. Preserve errors
		// such as NotFound, PermissionDenied and DeadlineExceeded.
		if status.Code(err) != codes.Unimplemented {
			return nil, fmt.Errorf("CRI-O ContainerStatus v1 failed: %w", err)
		}
	}
	resp, err := v1alpha2.NewRuntimeServiceClient(operator.criConn).ContainerStatus(ctx,
		&v1alpha2.ContainerStatusRequest{ContainerId: id, Verbose: true})
	if err != nil {
		return nil, fmt.Errorf("CRI-O ContainerStatus v1alpha2 failed: %w", err)
	}
	return parseCrioContainerInfo(resp.GetInfo())
}

func parseCrioContainerInfo(info map[string]string) (*crioContainerInfo, error) {
	raw, ok := info["info"]
	if !ok || raw == "" {
		return nil, errors.New("CRI-O verbose container info is empty")
	}
	var result crioContainerInfo
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("decode CRI-O container info: %w", err)
	}
	if result.RuntimeSpec == nil || result.RuntimeSpec.Process == nil ||
		result.RuntimeSpec.Linux == nil || result.RuntimeSpec.Linux.Resources == nil {
		return nil, errors.New("CRI-O runtimeSpec is missing process or Linux resources")
	}
	return &result, nil
}

// CRI has no ListPids operation. CRI-O exposes only the host init PID in
// ContainerStatus. Do not call containerd's Tasks API or claim child PID coverage.
func (operator *RuntimeOperatorTool) getCrioContainerPIDs(ctx context.Context, id string) ([]uint32, error) {
	info, err := operator.getCrioContainerInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	var pid uint32
	if err := json.Unmarshal(info.PID, &pid); err != nil {
		return nil, fmt.Errorf("decode CRI-O init PID: %w", err)
	}
	if pid == 0 || pid > math.MaxInt32 {
		return nil, errors.New("CRI-O init PID is invalid")
	}
	return []uint32{pid}, nil
}
