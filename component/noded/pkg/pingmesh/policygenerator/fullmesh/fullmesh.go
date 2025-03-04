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
Package fullmesh is one of policy generator for pingmesh
*/

package fullmesh

import (
	"sort"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

type generatorImp struct {
	local string
}

const Rule types.PingMeshRule = "full-mesh"

// New create a new generator
func New(node string) *generatorImp {
	return &generatorImp{
		local: node,
	}
}

// Generate generate pingmesh dest addresses
func (g *generatorImp) Generate(addrs map[string]types.SuperDeviceIDs) map[string]types.DestinationAddress {
	if g == nil {
		return nil
	}

	local, ok := addrs[g.local]
	if !ok {
		hwlog.RunLog.Errorf("local node %s not found in addrs", g.local)
		return nil
	}

	destAddresses := make(map[string]types.DestinationAddress)
	for this := range local {
		destAddresses[this] = make(map[uint]string)
		var internalDest []string
		for other, otherIp := range local {
			if this == other {
				continue
			}
			internalDest = append(internalDest, otherIp)
		}
		sort.Strings(internalDest)
		hwlog.RunLog.Debugf("internal dest for %s is %v", this, internalDest)
		destAddresses[this][common.InternalPingMeshTaskID] = strings.Join(internalDest, ",")
	}

	if len(addrs) == 1 {
		return destAddresses
	}

	for this := range local {
		var externalDest []string
		for node, ips := range addrs {
			if g.local == node {
				continue
			}
			if ip, ok := ips[this]; ok {
				externalDest = append(externalDest, ip)

			}
		}
		sort.Strings(externalDest)
		hwlog.RunLog.Debugf("external dest for %s is %v", this, externalDest)
		destAddresses[this][common.ExternalPingMeshTaskID] = strings.Join(externalDest, ",")
	}
	return destAddresses
}
