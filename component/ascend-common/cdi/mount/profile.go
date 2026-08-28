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

// Package mount — profile.go
//
// Mount profile (mounts.json) schema: the generation → full mount list map
// with a "default" fallback key, the builtin default profile, atomic file
// writing, and the per-entry UB filter used by Build.
package mount

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ascend-common/common-utils/hwlog"
)

const (
	// mountsJSONName is the JSON mount config file name under Dir.
	mountsJSONName = "mounts.json"

	// defaultGenerationKey is the fallback key used when the DevType has no
	// dedicated generation entry in the mount config.
	defaultGenerationKey = "default"

	// ascend950Gen is the mount-config generation key for the Ascend 950
	// hardware generation.
	ascend950Gen = "Ascend950"

	// ascend910A5DevType is the DevType the device manager reports for the
	// Ascend 950 generation; it is converted to ascend950Gen on lookup.
	ascend910A5DevType = "Ascend910A5"
)

// MountProfile is the full mount config: a map keyed by generation (e.g.
// "Ascend950") with the reserved "default" key used as fallback when the
// DevType has no dedicated entry. Each value is that generation's complete
// mount list — not a delta over a shared base.
type MountProfile map[string][]MountEntry

// DefaultMountProfile returns the builtin mount profile migrated from the
// legacy base.list / ub_driver.list:
//
//   - "default" holds the base driver paths shared by all generations.
//   - "Ascend950" holds the base driver paths (except /var/queue_schedule,
//     which the 950 generation does not need), the HCCL topology paths, and
//     the UB user-space files (marked Type: ubType, controlled by
//     MountUBDrv).
func DefaultMountProfile() MountProfile {
	defaultPaths := []string{
		"/usr/local/Ascend/driver/lib64",
		"/usr/local/Ascend/driver/include",
		"/usr/local/dcmi",
		"/usr/local/bin/npu-smi",
		"/var/queue_schedule",
	}

	ubPaths := []string{
		"/usr/lib64/libummu*",
		"/usr/lib64/liburma*",
		"/usr/lib64/urma",
		"/usr/lib64/libnl*",
		"/usr/bin/urma_admin",
		"/usr/bin/urma_perftest",
		"/usr/bin/urma_ping",
	}

	// The 950 generation keeps the base driver paths but drops
	// /var/queue_schedule and adds the HCCL topology paths.
	a950Paths := []string{
		"/usr/local/Ascend/driver/lib64",
		"/usr/local/Ascend/driver/include",
		"/usr/local/dcmi",
		"/usr/local/bin/npu-smi",
		hcclRootInfoPath,
		topoDirPath,
	}

	return MountProfile{
		defaultGenerationKey: []MountEntry{{Paths: defaultPaths}},
		ascend950Gen: []MountEntry{
			{Paths: ubPaths, Type: ubType},
			{Paths: a950Paths},
		},
	}
}

// generationKey maps a DevType to the generation key used in mounts.json. The
// Ascend 950 generation reports DevType "Ascend910A5" but is keyed "Ascend950"
// in the config; every other DevType maps to itself.
func generationKey(devType string) string {
	if devType == ascend910A5DevType {
		return ascend950Gen
	}
	return devType
}

// WriteMountProfile atomically writes cfg as JSON to <dir>/mounts.json. The
// directory is created when missing; the JSON is written to a temporary file
// first and renamed into place (the temp file is removed on failure).
func WriteMountProfile(dir string, cfg MountProfile) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create mount config directory %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mount config: %w", err)
	}
	tmpPath := filepath.Join(dir, mountsJSONName+".tmp")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write temp mount config %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, filepath.Join(dir, mountsJSONName)); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename mount config %s: %w", tmpPath, err)
	}
	return nil
}

// loadMountProfile reads and unmarshals <dir>/mounts.json. A missing file is
// reported as the raw os.IsNotExist error so the caller can treat it as an
// empty config.
func loadMountProfile(dir string) (*MountProfile, error) {
	path := filepath.Join(dir, mountsJSONName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("read mount config %s: %w", path, err)
	}
	var cfg MountProfile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse mount config %s: %w", path, err)
	}
	return &cfg, nil
}

// readProfileEntries reads mounts.json from dir and returns the generation's full
// mount list: the entry keyed by the DevType's generation key (falling back to
// "default" when it has no dedicated entry). Entries marked Type ubType are
// dropped when mountUBDrv is false; ubIncluded reports whether any UB
// entry survived (so the engine forces allow-link for the symlink entries). A
// missing mounts.json yields empty entries without error (mirrors the list
// mode missing-directory behavior).
func readProfileEntries(dir, devType string, mountUBDrv bool) ([]MountEntry, bool, error) {
	cfg, err := loadMountProfile(dir)
	if err != nil {
		if os.IsNotExist(err) {
			hwlog.RunLog.Warnf("Mount: mount config not found: %s", filepath.Join(dir, mountsJSONName))
			return nil, false, nil
		}
		return nil, false, err
	}

	entries := (*cfg)[generationKey(devType)]
	if entries == nil {
		entries = (*cfg)[defaultGenerationKey]
	}

	result := make([]MountEntry, 0, len(entries))
	ubIncluded := false
	for _, e := range entries {
		if e.Type == ubType && !mountUBDrv {
			continue
		}
		result = append(result, e)
		if e.Type == ubType {
			ubIncluded = true
		}
	}
	return result, ubIncluded, nil
}
