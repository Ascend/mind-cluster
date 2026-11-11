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

package process

import (
	"errors"
	"fmt"
	"os"

	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/parser"
	"ascend-common/rdma"
)

func processRDMARequest(spec *specs.Spec) (bool, error) {
	hcaNames, err := rdma.ParseRDMADevices(parser.GetEnvValue(spec.Process.Env, api.AscendRdmaDevicesEnv))
	if err != nil {
		return false, err
	}
	if err := addRDMADevices(spec, hcaNames); err != nil {
		return false, err
	}
	return len(hcaNames) != 0, nil
}

// addRDMADevices injects selected character devices and their exact cgroup rules.
func addRDMADevices(spec *specs.Spec, hcaNames []string) error {
	if len(hcaNames) == 0 {
		return nil
	}
	if spec.Linux == nil {
		spec.Linux = &specs.Linux{}
	}
	if spec.Linux.Resources == nil {
		spec.Linux.Resources = &specs.LinuxResources{}
	}
	seen := make(map[string]bool)
	for _, hcaName := range hcaNames {
		uverbsNodePath, err := rdma.ResolveRdmaDeviceNode(hcaName)
		if err != nil {
			return fmt.Errorf("%s HCA %s: %w", api.AscendRdmaDevicesEnv, hcaName, err)
		}
		if seen[uverbsNodePath] {
			continue
		}
		seen[uverbsNodePath] = true
		if err := addRDMACharacterDevice(spec, uverbsNodePath); err != nil {
			return fmt.Errorf("%s HCA %s: %w", api.AscendRdmaDevicesEnv, hcaName, err)
		}
	}
	// rdma_cm is shared and may be absent on hosts that do not expose it.
	rdmaCMNodePath := "/dev/infiniband/rdma_cm"
	if err := addRDMACharacterDevice(spec, rdmaCMNodePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s: %w", api.AscendRdmaDevicesEnv, err)
	}
	return nil
}

func addRDMACharacterDevice(spec *specs.Spec, nodePath string) error {
	rdmaDevice, err := oci.DeviceFromPath(nodePath)
	if err != nil {
		return fmt.Errorf("read RDMA node %s: %w", nodePath, err)
	}
	if rdmaDevice.Type != "c" {
		return fmt.Errorf("RDMA node %s is not a character device (type %s)", nodePath, rdmaDevice.Type)
	}
	appendDeviceToSpec(spec, rdmaDevice)
	return nil
}
