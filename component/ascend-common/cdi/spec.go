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

// Package cdi provides CDI spec generation for Ascend NPU devices.
//
// # Package cdi — spec.go
//
// BuildSpec, GenerateClaimSpec, DeleteClaimSpec, and atomic file I/O.
package cdi

import (
	"fmt"
	"strings"

	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/cdi/mount"
)

// ---------------------------------------------------------------------------
// Configuration structs
// ---------------------------------------------------------------------------

// DeviceConfig holds device identification fields shared by BuildSpec and ClaimSpec.
type DeviceConfig struct {
	// DeviceIDs is the list of logical NPU device IDs to include in the spec.
	// Each ID produces one per-device davinci node in the output.
	DeviceIDs []int

	// DevType identifies the Ascend hardware variant.
	// Examples: "Ascend310", "Ascend310B", "Ascend310P", "Ascend910", "Ascend910A5",
	// "Atlas 200I SoC A1", "Atlas 200 Model 3000".
	DevType string

	// ProductType is the hardware-level product identifier.
	// Used to skip common manager devices (devmm_svm, hisi_hdc) for Atlas200 products.
	// May be empty for non-Atlas200 hardware.
	ProductType string

	// UseVirtual toggles vdavinci device naming for virtual NPU support.
	// When true, device HostPath uses "vdavinci" prefix while ContainerPath uses "davinci".
	UseVirtual bool
}

// BuildSpecConfig holds all configuration for BuildSpec.
type BuildSpecConfig struct {
	DeviceConfig // embed shared device fields

	// MountConfig holds the mount-generation configuration passed to Build.
	// Use mount.MountConfig{Dir: "/etc/ascend-docker-runtime.d",
	// IsAscendDockerRuntime: true} for runtime CDI (list mode), or
	// mount.MountConfig{Dir: "/etc/ascend-dra/mounts"} for the JSON mount
	// profile written by DRA (default JSON mode).
	mount.MountConfig

	// Version is the CDI spec version. Empty falls back to the package
	// default (cdiVersion).
	Version string

	// Kind is the CDI spec kind (device vendor/class identifier). Empty
	// falls back to the package default (cdiKind).
	Kind string
}

// ---------------------------------------------------------------------------
// CDI spec metadata
// ---------------------------------------------------------------------------

const (
	cdiVersion = "0.8.0"
	cdiKind    = "ascend.com/npu"
)

// ---------------------------------------------------------------------------
// Ascend driver library paths
// ---------------------------------------------------------------------------

// ascendDriverLibPaths lists Ascend driver library directories scanned for
// LD_LIBRARY_PATH injection.  Mirrors ascend-docker-runtime's
// ascendDriverLibPaths.
var ascendDriverLibPaths = []string{
	"/usr/local/Ascend/driver/lib64/common",
	"/usr/local/Ascend/driver/lib64/driver",
	"/usr/lib64", // UB driver user-space library dir; mirrors legacy runtime/process/process.go
}

const ldLibraryPathKey = "LD_LIBRARY_PATH"

// collectAscendLibPaths returns the subset of ascendDriverLibPaths that
// exist on the filesystem. hostFsPrefix, when non-empty, prefixes the stat
// target for non-/dev paths (see mount.StatHostPath).
func collectAscendLibPaths(hostFsPrefix string) []string {
	var paths []string
	for _, libPath := range ascendDriverLibPaths {
		if err := mount.StatHostPath(hostFsPrefix, libPath); err == nil {
			paths = append(paths, libPath)
		}
	}
	return paths
}

// ---------------------------------------------------------------------------
// BuildSpec
// ---------------------------------------------------------------------------

// PrepareMountConfigFile writes the builtin mount profile to <dir>/mounts.json.
// DRA calls this at startup so CDI can read the JSON when generating specs.
func PrepareMountConfigFile(dir string) error {
	return mount.WriteMountProfile(dir, mount.DefaultMountProfile())
}

// buildMountsFn is the seam through which BuildSpec generates container
// mounts. Tests override it to capture arguments or inject failures; the
// default is mount.Build.
var buildMountsFn = mount.Build

// BuildSpec constructs a CDI v0.8.0 Spec.  Per-device davinci nodes go into
// Device[i].ContainerEdits; shared manager/UB nodes, mounts, and HCCL
// topology go into Spec.ContainerEdits.
//
// Used by both GenerateEdits (memory CDI) and GenerateClaimSpec (file CDI).
func BuildSpec(cfg BuildSpecConfig) (*cdispec.Spec, error) {
	version := cfg.Version
	if version == "" {
		version = cdiVersion
	}
	kind := cfg.Kind
	if kind == "" {
		kind = cdiKind
	}

	mounts, err := buildMountsFn(cfg.MountConfig, cfg.DevType)
	if err != nil {
		return nil, err
	}
	specEdits := cdispec.ContainerEdits{Mounts: mounts}
	if libPaths := collectAscendLibPaths(cfg.MountConfig.HostFsPrefix); len(libPaths) > 0 {
		specEdits.Env = append(specEdits.Env, ldLibraryPathKey+"="+strings.Join(libPaths, ":"))
	}

	if len(cfg.DeviceIDs) == 0 {
		return &cdispec.Spec{Version: version, Kind: kind, ContainerEdits: specEdits}, nil
	}

	devices, err := GeneratePerDeviceNodes(cfg.DeviceIDs, cfg.UseVirtual)
	if err != nil {
		return nil, err
	}
	sharedNodes, err := GenerateSharedNodes(cfg.DevType, cfg.ProductType)
	if err != nil {
		return nil, err
	}
	specEdits.DeviceNodes = sharedNodes

	spec := &cdispec.Spec{
		Version:        version,
		Kind:           kind,
		ContainerEdits: specEdits,
		Devices:        devices,
	}
	if err := validateSpec(spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// validateSpec performs semantic validation on a CDI Spec.
// It ensures:
//   - The version is a valid, known CDI spec version.
//   - The kind field is non-empty.
//   - At least one device is defined.
//   - Device names are non-empty and unique within the spec.
func validateSpec(spec *cdispec.Spec) error {
	if err := cdispec.ValidateVersion(spec); err != nil {
		return fmt.Errorf("invalid CDI spec: %w", err)
	}

	if spec.Kind == "" {
		return fmt.Errorf("invalid CDI spec: kind is empty")
	}

	if len(spec.Devices) == 0 {
		return fmt.Errorf("invalid CDI spec: no devices")
	}

	seen := make(map[string]bool)
	for _, dev := range spec.Devices {
		if dev.Name == "" {
			return fmt.Errorf("invalid CDI spec: device name is empty")
		}
		if seen[dev.Name] {
			return fmt.Errorf("invalid CDI spec: duplicate device name %q", dev.Name)
		}
		seen[dev.Name] = true
	}

	return nil
}

// ---------------------------------------------------------------------------
// GenerateClaimSpec
// ---------------------------------------------------------------------------

// GenerateClaimSpec builds a CDI specification for the given claim and
// list of logical device IDs, then writes it to the CDI cache's configured
// Spec directory using the standard CDI library API.
//
// claimUID is the unique identifier for the claim; it is used as the CDI
// spec file name suffix. cfg.Version and cfg.Kind, when empty, fall back to
// the package defaults.
//
// Returns:
//   - specName:  the generated Spec file name (without extension), used for later removal
//   - cdidIDs:   fully-qualified CDI device names ("ascend.com/npu={id}")
//   - err:       non-nil on validation, I/O, or serialisation failure
func GenerateClaimSpec(cfg BuildSpecConfig, claimUID string) (string, []string, error) {
	if claimUID == "" {
		return "", nil, fmt.Errorf("cdi: claimUID must not be empty")
	}

	spec, err := BuildSpec(cfg)
	if err != nil {
		return "", nil, fmt.Errorf("cdi: build spec: %w", err)
	}

	name, err := cdiapi.GenerateNameForTransientSpec(spec, claimUID)
	if err != nil {
		return "", nil, fmt.Errorf("cdi: generate spec name: %w", err)
	}

	if err := cdiapi.GetDefaultCache().WriteSpec(spec, name); err != nil {
		return "", nil, fmt.Errorf("cdi: write spec: %w", err)
	}

	cdidIDs := make([]string, 0, len(cfg.DeviceIDs))
	for _, id := range cfg.DeviceIDs {
		cdidIDs = append(cdidIDs, fmt.Sprintf("%s=%d", spec.Kind, id))
	}

	return name, cdidIDs, nil
}

// DeleteClaimSpec removes a previously generated per-claim CDI spec
// from the CDI cache's configured Spec directory.
//
// kind is the same CDI kind passed to GenerateClaimSpec (empty falls back to
// the package default); claimUID is the same unique identifier passed to
// GenerateClaimSpec.
// Returns nil when the Spec does not exist (the operation is idempotent).
func DeleteClaimSpec(kind, claimUID string) error {
	if kind == "" {
		kind = cdiKind
	}
	name, err := cdiapi.GenerateNameForTransientSpec(&cdispec.Spec{Kind: kind}, claimUID)
	if err != nil {
		return fmt.Errorf("cdi: invalid kind %q: %w", kind, err)
	}
	return cdiapi.GetDefaultCache().RemoveSpec(name)
}
