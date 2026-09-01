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

package domain

import (
	"sync"
	"time"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

const (
	expireFactor = 3
)

// nodeSnapshot holds the latest full snapshot of containers for a node
type nodeSnapshot struct {
	containers map[string][]*proto.ContainerInfo // job ID -> containers
	syncTime   int64
}

func (ns *nodeSnapshot) expired() bool {
	return time.Now().Unix()-ns.syncTime > int64(common.ParamOption.ScheduledSyncInterval*expireFactor)
}

// ContainerStore keeps the latest container snapshot per node
type ContainerStore struct {
	mu    sync.RWMutex
	nodes map[string]*nodeSnapshot // node ID -> snapshot
}

// NewContainerStore creates an empty store.
func NewContainerStore() *ContainerStore {
	return &ContainerStore{nodes: make(map[string]*nodeSnapshot)}
}

// groupByJob builds the job-ID -> containers map from a flat container list.
func groupByJob(containers []*proto.ContainerInfo) map[string][]*proto.ContainerInfo {
	byJob := make(map[string][]*proto.ContainerInfo)
	for _, c := range containers {
		byJob[c.JobId] = append(byJob[c.JobId], c)
	}
	return byJob
}

// ApplySync overwrites the snapshot for req.NodeId.
func (s *ContainerStore) ApplySync(req *proto.SyncDataReq) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[req.NodeId] = &nodeSnapshot{
		containers: groupByJob(req.Containers),
		syncTime:   time.Now().Unix(),
	}
}

// GetNodesByJob returns the node IDs that host any of the given jobs.
func (s *ContainerStore) GetNodesByJob(jobIDs []string) []string {
	if len(jobIDs) == 0 {
		return nil
	}
	s.mu.RLock()
	var nodeIDs, expireNodes []string
	for nodeID, snap := range s.nodes {
		if snap.expired() {
			expireNodes = append(expireNodes, nodeID)
			continue
		}
		for _, jobID := range jobIDs {
			if len(snap.containers[jobID]) == 0 {
				continue
			}
			nodeIDs = append(nodeIDs, nodeID)
			break
		}
	}
	s.mu.RUnlock()

	s.cleanContainerByNodes(expireNodes)
	return nodeIDs
}

// GetContainersByJob returns all container infos belonging to any of the given
// jobs across all node snapshots (leader only).
func (s *ContainerStore) GetContainersByJob(jobIDs []string) []*proto.ContainerInfo {
	if len(jobIDs) == 0 {
		return nil
	}
	s.mu.RLock()
	var result []*proto.ContainerInfo
	var expireNodes []string
	for nodeID, snap := range s.nodes {
		if snap.expired() {
			expireNodes = append(expireNodes, nodeID)
			continue
		}
		for _, jobID := range jobIDs {
			for _, c := range snap.containers[jobID] {
				result = append(result, cloneContainerInfo(c))
			}
		}
	}
	s.mu.RUnlock()

	s.cleanContainerByNodes(expireNodes)
	return result
}

// cloneContainerInfo returns a deep copy of the given container info.
func cloneContainerInfo(c *proto.ContainerInfo) *proto.ContainerInfo {
	if c == nil {
		return nil
	}
	clone := *c
	clone.PhyIds = append([]int32{}, c.PhyIds...)
	return &clone
}

func (s *ContainerStore) cleanContainerByNodes(nodes []string) {
	if len(nodes) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, nodeID := range nodes {
		delete(s.nodes, nodeID)
	}
}
