/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"fmt"
	"strconv"
	"strings"

	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"

	"ascend-common/cdi"
	"ascend-common/cdi/mount"
	"ascend-common/common-utils/hwlog"
)

// CdiSpecInterface abstracts per-claim CDI spec generation and removal.
//
// The plugin only knows a claim UID and the allocated device names reported
// by the scheduler. Translating those names into NPU logical IDs and
// hardware-type metadata requires device knowledge that lives in the driver
// layer (DraGenerationInterface + cdi public library), so the plugin depends
// on this interface and lets the driver supply the implementation.
type CdiSpecInterface interface {
	// WriteClaimSpec generates and persists a CDI spec file for the claim
	// and returns the fully-qualified CDI device IDs to be injected into
	// the requesting container.
	WriteClaimSpec(claimUID string, deviceNames []string) (cdiDeviceIDs []string, err error)

	// DeleteClaimSpec removes a previously generated CDI spec file. It is
	// idempotent: a missing spec is not an error.
	DeleteClaimSpec(claimUID string) error
}

// cdiSpecManager implements plugin.CdiSpecManager by delegating to the
// ascend-common/cdi public library. It is the bridge between the plugin
// (which only knows claim UIDs and allocated device names) and the cdi
// library (which needs NPU device IDs, devType and productType).
//
// Device names follow the "<releasedName>-<id>" convention (e.g.
// "Ascend910-0"), so the numeric suffix is the device ID. No allInfo
// lookup is needed.
type cdiSpecManager struct {
	devType      string
	productTypes []string
}

const (
	// mountConfigDir is the host directory where the JSON mount profile
	// (mounts.json) is written by PrepareMountProfileFile and read by CDI when
	// generating per-claim specs. It is backed by a hostPath volume (see
	// build/ascend-dra-driver.yaml).
	mountConfigDir = "/etc/ascend-dra/mounts"
	// defaultHostRoot is the host directory where the plugin container mounts /usr /etc .etc
	defaultHostRoot = "/hostRoot"
)

// NewCDISpecManager constructs a cdiSpecManager. devType and productTypes
// come from the generation once it has been handed a device manager.
// cdiRoot configures the default CDI cache's Spec directory so
// GenerateClaimSpec writes files there.
func NewCDISpecManager(
	devType string,
	productTypes []string,
	cdiRoot string,
) *cdiSpecManager {
	// The cdi public library uses the global default cache; configure its
	// Spec directory once at construction time. Safe to call before the
	// cache is first touched by GenerateClaimSpec.
	_ = cdiapi.Configure(cdiapi.WithSpecDirs(cdiRoot))

	// Publish the builtin mount config (generation + UB partitions) as JSON so
	// the CDI spec generation can read it from the shared host directory.
	if err := cdi.PrepareMountConfigFile(mountConfigDir); err != nil {
		hwlog.RunLog.Warnf("prepare mount profile file failed: %v", err)
	}

	return &cdiSpecManager{
		devType:      devType,
		productTypes: productTypes,
	}
}

// Compile-time check: cdiSpecManager satisfies CdiSpecInterface.
var _ CdiSpecInterface = (*cdiSpecManager)(nil)

// WriteClaimSpec parses the numeric suffix from each device name, then asks
// the cdi public library to build and persist a CDI spec file for the claim.
// Returns the fully-qualified CDI device IDs so the plugin can fill them
// into the prepared devices handed back to kubelet.
func (m *cdiSpecManager) WriteClaimSpec(claimUID string, deviceNames []string) ([]string, error) {
	ids := make([]int, 0, len(deviceNames))
	for _, name := range deviceNames {
		id, err := parseDeviceIDSuffix(name)
		if err != nil {
			return nil, fmt.Errorf("cdi: parse device ID for %q: %w", name, err)
		}
		ids = append(ids, id)
	}

	// cdi.DeviceConfig.ProductType is a single string; it is only used to
	// detect Atlas200 products so the first entry is sufficient. Empty for
	// non-Atlas hardware where GetProductTypes returns an empty slice.
	productType := ""
	if len(m.productTypes) > 0 {
		productType = m.productTypes[0]
	}

	_, cdiIDs, err := cdi.GenerateClaimSpec(cdi.BuildSpecConfig{
		DeviceConfig: cdi.DeviceConfig{
			DeviceIDs:   ids,
			DevType:     m.devType,
			ProductType: productType,
		},
		MountConfig: mount.MountConfig{
			Dir:      mountConfigDir,
			HostRoot: defaultHostRoot,
		},
	}, claimUID)
	if err != nil {
		return nil, fmt.Errorf("cdi: generate claim spec: %w", err)
	}
	hwlog.RunLog.Debugf("CDI spec written, claimUID=%s, cdiIDs=%v", claimUID, cdiIDs)
	return cdiIDs, nil
}

// DeleteClaimSpec asks the cdi public library to remove the per-claim CDI
// spec file. Idempotent.
func (m *cdiSpecManager) DeleteClaimSpec(claimUID string) error {
	return cdi.DeleteClaimSpec("", claimUID)
}

// parseDeviceIDSuffix splits a "<name>-<id>" device name and returns the
// trailing integer. e.g. "Ascend910-12" -> 12, nil.
func parseDeviceIDSuffix(name string) (int, error) {
	idx := strings.LastIndex(name, "-")
	if idx < 0 || idx == len(name)-1 {
		return 0, fmt.Errorf("no '-' separator or empty suffix in %q", name)
	}
	id, err := strconv.Atoi(name[idx+1:])
	if err != nil {
		return 0, fmt.Errorf("suffix %q in %q is not an integer: %w", name[idx+1:], name, err)
	}
	return id, nil
}
