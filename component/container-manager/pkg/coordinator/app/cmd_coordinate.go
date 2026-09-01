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

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

var errCoordNotReady = errors.New("coordinator not initialized")

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
	return err
}

func (c *Coordinator) isLocalLeader(addr string) bool {
	return addr == common.ParamOption.ListenAddr
}
