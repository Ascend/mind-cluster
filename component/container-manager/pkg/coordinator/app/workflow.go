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
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/coordinator"
	"container-manager/pkg/coordinator/domain"
)

// Coordinator is the top-level coordinator Module.
type Coordinator struct {
	ops coordinator.ContainerOps
	// Shared domain state.
	containerStore *domain.ContainerStore
	// Role endpoints.
	server *serverEndpoint // leader: owns gRPC server + broadcast fan-out
	client *clientEndpoint // ordinary: owns leader conns + broadcast streams + fail-over
	// leaderIdx records the index
	leaderIdx atomic.Int32
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewCoordinator constructs a Coordinator with the given config.
func NewCoordinator(ctx context.Context, ops coordinator.ContainerOps) *Coordinator {
	c := &Coordinator{
		ops:            ops,
		containerStore: domain.NewContainerStore(),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return c
}

// Name implements workflow.Module.
func (c *Coordinator) Name() string { return "coordinator" }

// Init implements workflow.Module.
func (c *Coordinator) Init() error {
	var err error
	if common.ParamOption.ListenAddr != "" {
		hwlog.RunLog.Infof("coordinator starting in LEADER role, local=%s", common.ParamOption.LocalNodeID)
		if c.server, err = newServerEndpoint(c); err != nil {
			return err
		}
	}
	if len(common.ParamOption.LeaderAddrs) > 0 {
		if common.ParamOption.LocalNodeID == "" {
			return errors.New("localNodeID must be set")
		}
		hwlog.RunLog.Infof("coordinator starting in ORDINARY role, local=%s, leaders=%v",
			common.ParamOption.LocalNodeID, common.ParamOption.LeaderAddrs)
		if c.client, err = newClientEndpoint(c); err != nil {
			hwlog.RunLog.Warnf("initial leader connect incomplete, retrying in background: %v", err)
		}
	}
	return nil
}

// Work implements workflow.Module.
func (c *Coordinator) Work(ctx context.Context) {
	if !c.enabled() {
		return
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.dataSyncLoop(c.ctx)
	}()
}

// ShutDown implements workflow.Module.
func (c *Coordinator) ShutDown() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.server != nil {
		c.server.close()
	}
	if c.client != nil {
		c.client.close()
	}
	c.wg.Wait()
	hwlog.RunLog.Infof("coordinator shut down")
}

// enabled implements coordinator.DistributedCoord.
func (c *Coordinator) enabled() bool {
	return len(common.ParamOption.LeaderAddrs) > 0
}
