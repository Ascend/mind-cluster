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

// Package enpuconfig implements device-plugin allocation policy.
package enpuconfig

import (
	"context"
	"fmt"

	"ascend-common/enpumanager"
)

// ConfigRoot is the allowed host directory for manager allocation files.
const ConfigRoot = "/etc/enpu/vcann-rt/allocations"

// ResourceClient exposes the manager operations used to create or reuse an allocation.
type ResourceClient interface {
	Allocate(context.Context, enpumanager.AllocateRequest) (*enpumanager.AllocateResponse, error)
	ListAllocations(context.Context) ([]enpumanager.Allocation, error)
}

// Provider applies device-plugin allocation reuse and request validation.
type Provider struct{ ResourceClient }

// findPodAllocation finds a unique allocation in a complete list.
func (c *Provider) findPodAllocation(ctx context.Context, uid string) (*enpumanager.Allocation, error) {
	if uid == "" {
		return nil, fmt.Errorf("missing soft share Pod UID")
	}
	allocations, err := c.ListAllocations(ctx)
	if err != nil {
		return nil, err
	}
	var match *enpumanager.Allocation
	for i := range allocations {
		if allocations[i].Metadata.PodUID != uid {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("multiple enpu-manager allocations for Pod %s", uid)
		}
		match = &allocations[i]
	}
	if match != nil && !enpumanager.IsValidAllocationID(match.AllocationID) {
		return nil, fmt.Errorf("invalid enpu-manager allocation UUID for Pod %s", uid)
	}
	return match, nil
}

// EnsureAllocation reuses an existing allocation, or asks the manager to create it.
func (c *Provider) EnsureAllocation(ctx context.Context, input enpumanager.AllocateRequest) (*AllocationResult, error) {
	existing, err := c.findPodAllocation(ctx, input.PodUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		result, err := c.Allocate(ctx, input)
		if err != nil {
			return nil, err
		}
		if result.PhysicalID != input.PhysicalID {
			return nil, fmt.Errorf("enpu-manager returned an unexpected physical device")
		}
		return &AllocationResult{AllocateResponse: *result}, nil
	}
	if existing.PhysicalID != input.PhysicalID || existing.AICoreQuota != input.AICoreQuota ||
		existing.HBMQuota != input.HBMQuota || existing.HBMLimit != input.HBMQuota ||
		existing.SchedPolicy != input.SchedPolicy {
		return nil, fmt.Errorf("existing enpu-manager allocation does not match Pod %s request", input.PodUID)
	}
	return &AllocationResult{AllocateResponse: enpumanager.AllocateResponse{AllocationID: existing.AllocationID, ConfigPath: existing.ConfigPath,
		PhysicalID: existing.PhysicalID, VirtualID: existing.VirtualID, DieID: existing.DieID, ShmID: existing.ShmID,
	}, Reused: true}, nil
}

// AllocationResult records whether DP reused an allocation.
type AllocationResult struct {
	enpumanager.AllocateResponse
	Reused bool
}
