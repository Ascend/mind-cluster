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

package app

import (
	"errors"
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

var errCoordNotReady = errors.New("coordinator not initialized")

// --------------------
// coordinate Initiator
// --------------------

// RequestStopJobs implements coordinator.DistributedCoord.
func (c *Coordinator) RequestStopJobs(jobIDs []string, ctrIds []string) error {
	return c.requestJobs(jobIDs, ctrIds, common.ActionStop)
}

// RequestStartJobs implements coordinator.DistributedCoord.
func (c *Coordinator) RequestStartJobs(jobIDs []string, ctrIds []string) error {
	return c.requestJobs(jobIDs, ctrIds, common.ActionStart)
}

// requestJobs builds a CoordinateReq and routes it to the leader path or the
// ordinary path based on the local role.
func (c *Coordinator) requestJobs(jobIDs, ctrIds []string, action string) error {
	if len(jobIDs) == 0 {
		return nil
	}
	if !c.enabled() {
		return errCoordNotReady
	}
	req := &proto.CoordinateReq{
		Uuid:   utils.NewUUID(),
		NodeId: common.ParamOption.LocalNodeID,
		JobIds: jobIDs,
		CtrIds: ctrIds,
		Action: action,
	}
	hwlog.RunLog.Infof("start a coordinate req: %s", req.String())
	return c.callAllLeadersForCoord(req)
}

// callAllLeadersForCoord is the fail-over wrapper for CoordinateReq.
func (c *Coordinator) callAllLeadersForCoord(req *proto.CoordinateReq) error {
	leaders := common.ParamOption.LeaderAddrs
	if len(leaders) == 0 {
		return fmt.Errorf("call all leaders failed: no leader found")
	}
	start := int(c.leaderIdx.Load())
	for i := 0; i < len(leaders); i++ {
		idx := (start + i) % len(leaders)
		addr := leaders[idx]
		if err := c.tryLeader(addr, req); err != nil {
			hwlog.RunLog.Warnf("leader %s handle failed, try next: %v", addr, err)
			continue
		}
		// remember the working leader so the next request starts from it
		c.leaderIdx.Store(int32(idx))
		return nil
	}
	return fmt.Errorf("call all leaders failed: all leader handle failed")
}

// tryLeader forwards req to the given leader address.
func (c *Coordinator) tryLeader(addr string, req *proto.CoordinateReq) error {
	if c.isLocalLeader(addr) {
		return c.broadcastToOrdinary(req)
	}
	if c.client == nil {
		return errCoordNotReady
	}
	return c.client.callLeaderForCoord(addr, req)
}

// --------------------
// coordinate broadcaster
// --------------------

// broadcast runs on the leader
func (c *Coordinator) broadcastToOrdinary(req *proto.CoordinateReq) error {
	if c.server == nil {
		return errCoordNotReady
	}
	nodeIDs := c.containerStore.GetNodesByJob(req.JobIds)
	hwlog.RunLog.Infof("broadcast coordinate req %s to nodes %s", req.Uuid, nodeIDs)
	if len(nodeIDs) == 0 {
		return fmt.Errorf("coordinate %s: no hosting nodes found", req.Uuid)
	}
	if !c.isStreamsAlive(nodeIDs) {
		return fmt.Errorf("coordinate %s: streams are not alive", req.Uuid)
	}
	if err := c.validateJobsForCoord(req); err != nil {
		return err
	}
	resp := c.broadcastCoordReq(req, nodeIDs)
	if resp.GetCode() != 0 {
		return fmt.Errorf("coordinate %s failed: %s", req.Uuid, resp.GetMessage())
	}
	return nil
}

// isStreamsAlive checks if all nodeIDs have active streams
func (c *Coordinator) isStreamsAlive(nodeIDs []string) bool {
	for _, nodeID := range nodeIDs {
		if nodeID == common.ParamOption.LocalNodeID { // skip the broadcastor
			hwlog.RunLog.Debugf("will broadcast coordinate req to the node itself: %s", nodeID)
			continue
		}
		if _, ok := c.server.getStream(nodeID); !ok {
			hwlog.RunLog.Warnf("broadcast coordinate req to node %s failed: stream not found", nodeID)
			return false
		}
	}
	return true
}

// broadcastCoordReq is the leader-side broadcastCoordReq orchestrator
func (c *Coordinator) broadcastCoordReq(req *proto.CoordinateReq, nodeIDs []string) *proto.Response {
	hwlog.RunLog.Infof("direct broadcast req %s to ordinary node, init by %s", req.Uuid, req.NodeId)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failed []string
	for _, nodeID := range nodeIDs {
		f := func(nid string) {
			hwlog.RunLog.Debugf("broadcast coordinate req %s to node %s", req.Uuid, nid)
			defer wg.Done()
			var err error
			if nid == common.ParamOption.LocalNodeID {
				err = c.executeLocal(req)
			} else {
				err = c.server.sendCoordReq(req, nid)
			}
			if err != nil {
				mu.Lock()
				failed = append(failed, nid)
				mu.Unlock()
				hwlog.RunLog.Errorf("broadcast %s to node %s failed: %v", req.Uuid, nid, err)
			}
		}
		wg.Add(1)
		go f(nodeID)
	}
	wg.Wait()
	if len(failed) > 0 {
		return &proto.Response{Uuid: req.Uuid, Code: 1, Message: fmt.Sprintf("broadcast partially failed on nodes: %v", failed)}
	}
	return &proto.Response{Uuid: req.Uuid, Code: 0}
}

// validateJobsForCoord is a leader-side pre-broadcast sanity check for a stop/start coordination
func (s *Coordinator) validateJobsForCoord(req *proto.CoordinateReq) error {
	infos := s.containerStore.GetContainersByJob(req.JobIds)
	if len(infos) == 0 {
		return fmt.Errorf("no container info found for jobs %v", req.JobIds)
	}
	byJob := make(map[string][]*proto.ContainerInfo)
	for _, c := range infos {
		if !c.EnableRecover {
			return fmt.Errorf("job %s container %s does not enable recovery", c.JobId, c.ContainerId)
		}
		if req.Action == common.ActionStart {
			if c.PausedByPeer != "" && c.PausedByPeer != req.NodeId {
				return fmt.Errorf("job %s container %s paused by node %s cannot be started by node %s",
					c.JobId, c.ContainerId, c.PausedByPeer, req.NodeId)
			}
			if c.Status != common.StatusPaused {
				return fmt.Errorf("job %s container %s status %s cannot be started", c.JobId, c.ContainerId, c.Status)
			}
		}
		byJob[c.JobId] = append(byJob[c.JobId], c)
	}
	for jobID, cs := range byJob {
		replica := int(cs[0].JobReplica)
		if replica <= 0 {
			// replica not configured on the reported containers: skip the strict count check
			continue
		}
		if len(cs) != replica {
			return fmt.Errorf("job %s container count %d does not match replica %d", jobID, len(cs), replica)
		}
	}
	return nil
}

// --------------------
// coordinate executor
// --------------------

// executeLocal runs the requested stop/start action on the local node via
// ContainerOps.
func (c *Coordinator) executeLocal(req *proto.CoordinateReq) error {
	if c.ops == nil {
		return fmt.Errorf("container ops not injected")
	}

	var err error = nil
	switch req.Action {
	case common.ActionStop:
		err = c.ops.PauseJobContainers(req.JobIds, req.CtrIds, req.NodeId)
	case common.ActionStart:
		err = c.ops.ResumeJobContainers(req.JobIds, req.CtrIds, req.NodeId)
	default:
		err = fmt.Errorf("unknown action: %s", req.Action)
	}
	return err
}

func (c *Coordinator) isLocalLeader(addr string) bool {
	return addr == common.ParamOption.ListenAddr
}
