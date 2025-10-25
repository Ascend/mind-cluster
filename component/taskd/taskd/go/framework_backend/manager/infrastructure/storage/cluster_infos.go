/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain c copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package storage for taskd manager backend data type
package storage

import (
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/utils"
	"taskd/toolkit_backend/net/common"
)

// ClusterInfos all cluster infos
type ClusterInfos struct {
	Clusters  map[string]*ClusterInfo
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// ClusterInfo the cluster info
type ClusterInfo struct {
	Command   map[string]string
	Business  []int32
	HeartBeat time.Time
	FaultInfo map[string]string
	Pos       *common.Position
	RWMutex   sync.RWMutex
}

func (c *ClusterInfos) registerCluster(clusterName string) *ClusterInfo {
	c.RWMutex.Lock()
	clusterInfo := &ClusterInfo{
		Command:   make(map[string]string),
		Business:  make([]int32, 0),
		HeartBeat: time.Now(),
		FaultInfo: make(map[string]string),
		Pos:       &common.Position{},
		RWMutex:   sync.RWMutex{},
	}
	c.Clusters[clusterName] = clusterInfo
	c.RWMutex.Unlock()
	hwlog.RunLog.Infof("register cluster name:%v agentInfo:%v", clusterName, utils.ObjToString(clusterInfo))
	return clusterInfo
}

func (c *ClusterInfos) getCluster(clusterName string) (*ClusterInfo, error) {
	c.RWMutex.RLock()
	if cluster, exists := c.Clusters[clusterName]; exists {
		clusterInfo := cluster.getCluster()
		c.RWMutex.RUnlock()
		return clusterInfo, nil
	}
	c.RWMutex.RUnlock()
	return nil, fmt.Errorf("cluster name is unregistered : %v", clusterName)
}

func (c *ClusterInfo) getCluster() *ClusterInfo {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	return &ClusterInfo{
		Command:   c.Command,
		Business:  c.Business,
		HeartBeat: c.HeartBeat,
		FaultInfo: c.FaultInfo,
		Pos:       c.Pos,
		RWMutex:   sync.RWMutex{},
	}
}

func (c *ClusterInfos) updateCluster(clusterName string, newCluster *ClusterInfo) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.Clusters[clusterName] = &ClusterInfo{
		Command:   newCluster.Command,
		Business:  newCluster.Business,
		HeartBeat: newCluster.HeartBeat,
		FaultInfo: newCluster.FaultInfo,
		Pos:       newCluster.Pos,
	}
	return nil
}

// GetCluster get the cluster info by cluster name
func (c *ClusterInfos) GetCluster(clusterName string) (*ClusterInfo, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if cluster, exists := c.Clusters[clusterName]; exists {
		return cluster, nil
	}
	return nil, fmt.Errorf("cluster name is unregistered : %v", clusterName)
}
