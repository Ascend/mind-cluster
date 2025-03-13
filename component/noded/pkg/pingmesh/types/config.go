/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package types is all type of hccsping mesh
*/
package types

import "ascend-common/devmanager/common"

// HccspingMeshConfig is the configuration for the pingmesh component
type HccspingMeshConfig struct {
	Activate     string `json:"activate"`
	TaskInterval int    `json:"task_interval"`
}

// HccspingMeshPolicy is the policy for the pingmesh component
type HccspingMeshPolicy struct {
	Config   *HccspingMeshConfig
	Address  map[string]SuperDeviceIDs
	DestAddr map[string]DestinationAddress
	UID      string
}

// DeepCopy creates a deep copy of the HccspingMeshPolicy
func (p *HccspingMeshPolicy) DeepCopy() *HccspingMeshPolicy {
	np := &HccspingMeshPolicy{
		Config: &HccspingMeshConfig{
			p.Config.Activate,
			p.Config.TaskInterval,
		},
		Address:  make(map[string]SuperDeviceIDs, len(p.Address)),
		DestAddr: make(map[string]DestinationAddress, len(p.DestAddr)),
		UID:      p.UID,
	}
	for k, v := range p.Address {
		np.Address[k] = v.DeepCopy()
	}
	for k, v := range p.DestAddr {
		np.DestAddr[k] = v.DeepCopy()
	}
	return np
}

// PingMeshRule is the rule for generating pingmesh destination addresses
type PingMeshRule string

const (
	// ActivateOn is the value for the activate field when the pingmesh component is enabled
	ActivateOn = "on"
	// ActivateOff is the value for the activate field when the pingmesh component is disabled
	ActivateOff = "off"
)

// SuperDeviceIDs is a map of super device physicID to superDeviceID
type SuperDeviceIDs map[string]string

// DeepCopy creates a deep copy of the SuperDeviceIDs
func (s SuperDeviceIDs) DeepCopy() SuperDeviceIDs {
	ns := make(SuperDeviceIDs, len(s))
	for k, v := range s {
		ns[k] = v
	}
	return ns
}

// DestinationAddress is a map of hccsping mesh taskID to destination address
type DestinationAddress map[uint]string

// DeepCopy creates a deep copy of the DestinationAddress
func (d DestinationAddress) DeepCopy() DestinationAddress {
	nd := make(DestinationAddress, len(d))
	for k, v := range d {
		nd[k] = v
	}
	return nd
}

// HccspingMeshResult hccsping-mesh result
type HccspingMeshResult struct {
	Policy  *HccspingMeshPolicy
	Results map[string]map[uint]*common.HccspingMeshInfo
}
