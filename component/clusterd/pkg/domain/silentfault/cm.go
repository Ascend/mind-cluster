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

package silentfault

import (
	"sync"
	"time"

	"clusterd/pkg/common/constant"
)

// FaultSource identifies the occur-message source (internal = clusterd/silent-fault-<node>, external = its own resource and faultId)
type FaultSource struct {
	Resource string
	FaultId  string
}

// NodeSilentInfo node-level silent fault information
type NodeSilentInfo struct {
	FaultCode        string        // silent fault code
	FaultLevel       string        // silent fault level (looked up in publicFaultConfiguration.json)
	LastSeparateTime int64         // milliseconds, latest occur-message time (auto-release timer base; refreshed/deferred when an occur happens again)
	DevList          []string      // all cards of the node (device list of the latest occur message)
	Sources          []FaultSource // currently active occur-message sources of the node (a recover message is sent for each on release)
}

// SilentFaultCmInfo silent fault result cache: node -> silent fault info
type SilentFaultCmInfo struct {
	nodes map[string]NodeSilentInfo
	mutex sync.RWMutex
}

// NewSilentFaultCmInfo create a silent fault result cache
func NewSilentFaultCmInfo() *SilentFaultCmInfo {
	return &SilentFaultCmInfo{
		nodes: make(map[string]NodeSilentInfo),
		mutex: sync.RWMutex{},
	}
}

// Upsert handles an occur message: refresh LastSeparateTime and DevList if the node exists; append the source when Sources lacks the faultKey (supports coexisting internal/external sources).
// devList is the device-name list of all cards on the node (the caller converts DeviceIds in the message back to device names before passing in).
func (c *SilentFaultCmInfo) Upsert(node, faultId, resource string, devList []string) {
	now := time.Now().UnixMilli()
	faultKey := resource + faultId

	c.mutex.Lock()
	defer c.mutex.Unlock()

	info, ok := c.nodes[node]
	if !ok {
		c.nodes[node] = NodeSilentInfo{
			FaultCode:        constant.SilentFaultCode,
			FaultLevel:       constant.SilentFault,
			LastSeparateTime: now,
			DevList:          append([]string{}, devList...),
			Sources:          []FaultSource{{Resource: resource, FaultId: faultId}},
		}
		return
	}
	info.LastSeparateTime = now
	info.DevList = append([]string{}, devList...)
	if !sourceExists(info.Sources, faultKey) {
		info.Sources = append(info.Sources, FaultSource{Resource: resource, FaultId: faultId})
	}
	c.nodes[node] = info
}

// Restore rebuilds the node silent fault entry from persisted data on restart. Unlike Upsert,
// it uses the passed historical value instead of the current time, so auto-release timing survives restarts.
// When multiple sources coexist on a node, LastSeparateTime keeps the larger value (the later occur time).
func (c *SilentFaultCmInfo) Restore(node, faultId, resource string, devList []string, lastSeparateTime int64) {
	faultKey := resource + faultId

	c.mutex.Lock()
	defer c.mutex.Unlock()

	info, ok := c.nodes[node]
	if !ok {
		c.nodes[node] = NodeSilentInfo{
			FaultCode:        constant.SilentFaultCode,
			FaultLevel:       constant.SilentFault,
			LastSeparateTime: lastSeparateTime,
			DevList:          append([]string{}, devList...),
			Sources:          []FaultSource{{Resource: resource, FaultId: faultId}},
		}
		return
	}
	info.DevList = append([]string{}, devList...)
	if !sourceExists(info.Sources, faultKey) {
		info.Sources = append(info.Sources, FaultSource{Resource: resource, FaultId: faultId})
	}
	if lastSeparateTime > info.LastSeparateTime {
		info.LastSeparateTime = lastSeparateTime
	}
	c.nodes[node] = info
}

// RemoveSource handles a recover message: removes the source; deletes the node entry when Sources becomes empty
func (c *SilentFaultCmInfo) RemoveSource(node, faultKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	info, ok := c.nodes[node]
	if !ok {
		return
	}
	info.Sources = removeSource(info.Sources, faultKey)
	if len(info.Sources) == 0 {
		delete(c.nodes, node)
		return
	}
	c.nodes[node] = info
}

// GetAll returns a deep copy of all node silent fault info
func (c *SilentFaultCmInfo) GetAll() map[string]NodeSilentInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make(map[string]NodeSilentInfo, len(c.nodes))
	for node, info := range c.nodes {
		res[node] = NodeSilentInfo{
			FaultCode:        info.FaultCode,
			FaultLevel:       info.FaultLevel,
			LastSeparateTime: info.LastSeparateTime,
			DevList:          append([]string{}, info.DevList...),
			Sources:          append([]FaultSource{}, info.Sources...),
		}
	}
	return res
}

// Get returns the silent fault info of the specified node
func (c *SilentFaultCmInfo) Get(node string) (NodeSilentInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	info, ok := c.nodes[node]
	info.Sources = append([]FaultSource{}, info.Sources...)
	info.DevList = append([]string{}, info.DevList...)
	return info, ok
}

// Nodes returns a snapshot of all node names
func (c *SilentFaultCmInfo) Nodes() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	nodes := make([]string, 0, len(c.nodes))
	for node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// Len returns the number of nodes
func (c *SilentFaultCmInfo) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.nodes)
}

// Expired returns the list of expired nodes (for auto release); releaseSeconds < 0 means no auto release
func (c *SilentFaultCmInfo) Expired(nowMilli, releaseSeconds int64) []string {
	if releaseSeconds < 0 {
		return nil
	}
	cutoff := nowMilli - releaseSeconds*constant.SecondsToMilliseconds
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	expired := make([]string, 0)
	for node, info := range c.nodes {
		if info.LastSeparateTime > 0 && info.LastSeparateTime <= cutoff {
			expired = append(expired, node)
		}
	}
	return expired
}

// Reset clears the result cache
func (c *SilentFaultCmInfo) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.nodes = make(map[string]NodeSilentInfo)
}

func sourceExists(sources []FaultSource, faultKey string) bool {
	for _, s := range sources {
		if s.Resource+s.FaultId == faultKey {
			return true
		}
	}
	return false
}

func removeSource(sources []FaultSource, faultKey string) []FaultSource {
	kept := sources[:0]
	for _, s := range sources {
		if s.Resource+s.FaultId != faultKey {
			kept = append(kept, s)
		}
	}
	return kept
}
