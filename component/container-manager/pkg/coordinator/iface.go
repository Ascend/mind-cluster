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

// Package coordinator provides the distributed NPU self-healing coordinator.
//
// The coordinator is an independent Module registered with the module manager.
// It is decoupled from the container controller via two boundary interfaces:
//   - DistributedCoord: implemented by the coordinator, called by CtrCtl.
//   - ContainerOps: implemented by CtrCtl, called by the coordinator.
//
// Both are injected in run.go to avoid a compile-time import cycle.
package coordinator

import (
	"container-manager/pkg/coordinator/proto"
)

// DistributedCoord exposes distributed coordination capabilities to the
// container controller. Implemented by the coordinator.
type DistributedCoord interface {
	RequestStopJobs(jobIDs []string, ctrIds []string) error
	RequestStartJobs(jobIDs []string, ctrIds []string) error
}

// ContainerOps exposes container operations needed by the coordinator.
// Implemented by CtrCtl.
type ContainerOps interface {
	GetLocalContainers() []*proto.ContainerInfo
	HasDataChanged() bool
	PauseJobContainers(jobIDs, ctrIds []string, peerNodeID string) error
	ResumeJobContainers(jobIDs, ctrIds []string, peerNodeID string) error
}
