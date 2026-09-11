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
	devcommon "ascend-common/devmanager/common"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// CdiSpecInterface abstracts per-claim CDI spec generation and removal.
//
// The plugin only knows a claim UID and the allocated device names reported
// by the scheduler. Translating those names into the /dev device IDs CDI must
// expose and hardware-type metadata requires device knowledge that lives in
// the driver layer (DraGenerationInterface + cdi public library), so the
// plugin depends on this interface and lets the driver supply the
// implementation.
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
// Device names follow the "npu-<id>" convention (e.g. "npu-0"), so the
// numeric suffix is the physical device ID. The CDI spec generator must
// expose the same /dev node the container actually mounts, so toMountID
// (supplied per-generation) is invoked on every
// WriteClaimSpec to honor runtime device-state changes. A nil toMountID
// means the mount ID equals the phyID and no conversion is performed.
type cdiSpecManager struct {
	devType      string
	productTypes []string
	toMountID    func(int32) (int32, error)
}

const (
	// mountConfigDir is the host directory where the JSON mount profile
	// (mounts.json) is written by PrepareMountProfileFile and read by CDI when
	// generating per-claim specs. It is backed by a hostPath volume (see
	// build/ascend-dra-driver.yaml).
	mountConfigDir = "/etc/ascend-dra/mounts"
	// defaultHostFsPrefix is the host directory where the plugin container mounts /usr /etc .etc
	defaultHostFsPrefix = "/hostRoot"
)

// NewCDISpecManager constructs a cdiSpecManager. devType and productTypes
// come from the generation once it has been handed a device manager.
// cdiRoot configures the default CDI cache's Spec directory so
// GenerateClaimSpec writes files there. toMountID converts the physical ID
// (parsed from a device name) to the ID of the /dev node CDI must expose;
// pass nil when the mount ID equals the phyID for this generation.
func NewCDISpecManager(
	devType string,
	productTypes []string,
	cdiRoot string,
	toMountID func(int32) (int32, error),
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
		toMountID:    toMountID,
	}
}

// Compile-time check: cdiSpecManager satisfies CdiSpecInterface.
var _ CdiSpecInterface = (*cdiSpecManager)(nil)

// WriteClaimSpec resolves physical and static-vNPU device names, then asks the
// CDI public library to build and persist a CDI spec file. Returns the
// fully-qualified CDI device IDs so the plugin can fill them into the prepared
// devices handed back to kubelet.
func (m *cdiSpecManager) WriteClaimSpec(claimUID string, deviceNames []string) ([]string, error) {
	if len(deviceNames) == 0 {
		return nil, fmt.Errorf("cdi: claim %s has no devices", claimUID)
	}
	useVirtual := strings.HasPrefix(deviceNames[0], staticVNPUNamePrefix)
	ids := make([]int, 0, len(deviceNames))
	for _, name := range deviceNames {
		deviceID, virtual, err := parseCDIDeviceName(name)
		if err != nil {
			return nil, fmt.Errorf("cdi: parse device ID for %q: %w", name, err)
		}
		if virtual != useVirtual {
			return nil, fmt.Errorf("cdi: physical and virtual devices cannot share one claim spec")
		}
		if !useVirtual && m.toMountID != nil {
			mountID, err := m.toMountID(int32(deviceID))
			if err != nil {
				return nil, fmt.Errorf("cdi: convert phyID %d to mountID: %w", deviceID, err)
			}
			deviceID = int(mountID)
		}
		ids = append(ids, deviceID)
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
			UseVirtual:  useVirtual,
		},
		MountConfig: mount.MountConfig{
			Dir:            mountConfigDir,
			HostFsPrefix:   defaultHostFsPrefix,
			// Mount UB driver files by default for Ascend 950-generation devices.
			MountUBDrv: true,
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

const (
	staticVNPUNamePrefix       = consts.StaticVNPUDeviceKind + devcommon.Minus
	staticVNPUNameSegmentCount = 4
	staticVNPUVDevIDSegment    = 2
)

func parseCDIDeviceName(name string) (int, bool, error) {
	if !strings.HasPrefix(name, staticVNPUNamePrefix) {
		id, err := parseDeviceIDSuffix(name)
		return id, false, err
	}
	parts := strings.Split(name, devcommon.Minus)
	if len(parts) != staticVNPUNameSegmentCount {
		return 0, true, fmt.Errorf("invalid static vNPU device name %q", name)
	}
	id, err := strconv.Atoi(parts[staticVNPUVDevIDSegment])
	if err != nil {
		return 0, true, fmt.Errorf("vDevID %q in %q is not an integer: %w",
			parts[staticVNPUVDevIDSegment], name, err)
	}
	return id, true, nil
}

// parseDeviceIDSuffix returns the trailing integer from a physical device name.
func parseDeviceIDSuffix(name string) (int, error) {
	idx := strings.LastIndex(name, devcommon.Minus)
	if idx < 0 || idx+len(devcommon.Minus) == len(name) {
		return 0, fmt.Errorf("no '-' separator or empty suffix in %q", name)
	}
	suffix := name[idx+len(devcommon.Minus):]
	id, err := strconv.Atoi(suffix)
	if err != nil {
		return 0, fmt.Errorf("suffix %q in %q is not an integer: %w", suffix, name, err)
	}
	return id, nil
}
