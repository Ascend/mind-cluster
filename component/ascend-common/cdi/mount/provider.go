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

// Package mount provides the mount Provider interface, mount‑type
// constants, HCCL topology injection, and OS‑detection helpers.
//
// Package mount — provider.go
//
// Provider interface, Build helper, HCCL topology, and OS detection.
package mount

import (
	"fmt"
	"os"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

// ---------------------------------------------------------------------------
// Mount type constants
// ---------------------------------------------------------------------------

const (
	mountTypeBind    = "bind"
	mountOptRBind    = "rbind"
	mountOptRPrivate = "rprivate"
	mountOptReadOnly = "ro"
)

// ---------------------------------------------------------------------------
// Provider interface
// ---------------------------------------------------------------------------

// Provider abstracts the source of device mounts.
// Implementations return device-appropriate mounts for CDI spec generation.
type Provider interface {
	// GetMounts returns the mount configurations for the device.
	// HostPath and ContainerPath in the returned Mounts must remain absolute host paths.
	GetMounts() ([]*cdispec.Mount, error)
}

// ---------------------------------------------------------------------------
// HCCL topology injection
// ---------------------------------------------------------------------------

const (
	hcclRootInfoPath = "/etc/hccl_rootinfo.json"
	topoDirPath      = "/usr/local/Ascend/driver/topo"
)

// TopologyItems lists host paths that should be bind-mounted into the
// container when present.  Exported so tests can substitute temporary files.
var TopologyItems = []TopologyItem{
	{hcclRootInfoPath, []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}},
	{topoDirPath, []string{mountOptRBind, mountOptRPrivate, mountOptReadOnly}},
}

// TopologyItem pairs a host path with its bind-mount options.
type TopologyItem struct {
	HostPath string
	Options  []string
}

// Build fetches mounts from the provider and appends HCCL topology mounts.
// This is the primary entry point called by cdi.BuildSpec.
func Build(provider Provider) ([]*cdispec.Mount, error) {
	mounts, err := provider.GetMounts()
	if err != nil {
		return nil, fmt.Errorf("cdi: get mounts: %w", err)
	}
	return appendTopology(mounts), nil
}

// appendTopology appends bind-mounts for HCCL topology artifacts when they
// exist on the host.
func appendTopology(mounts []*cdispec.Mount) []*cdispec.Mount {
	for _, item := range TopologyItems {
		if _, err := os.Stat(item.HostPath); err != nil {
			continue
		}
		mounts = append(mounts, &cdispec.Mount{
			ContainerPath: item.HostPath,
			HostPath:      item.HostPath,
			Type:          mountTypeBind,
			Options:       item.Options,
		})
	}
	return mounts
}
