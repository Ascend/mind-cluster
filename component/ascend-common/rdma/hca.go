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

// Package rdma parses HCA requests and discovers RDMA driver mounts.
package rdma

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ascend-common/api"
)

const (
	maxRDMADevices     = 512
	infinibandSysfsDir = "/sys/class/infiniband"
	hcaDeviceDir       = "device"
	infinibandVerbsDir = "infiniband_verbs"
	infinibandNodeDir  = "/dev/infiniband"
)

var hcaNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// ParseRDMADevices parses HCA names in request order; empty or void disables RDMA.
func ParseRDMADevices(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "void" {
		return nil, nil
	}

	hcaNames := make([]string, 0)
	seen := make(map[string]bool)
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if err := ValidateRDMADeviceName(name); err != nil {
			return nil, fmt.Errorf("invalid %s: %w", api.AscendRdmaDevicesEnv, err)
		}
		if !seen[name] {
			seen[name] = true
			hcaNames = append(hcaNames, name)
			if len(hcaNames) > maxRDMADevices {
				return nil, fmt.Errorf("invalid %s: more than %d unique devices", api.AscendRdmaDevicesEnv, maxRDMADevices)
			}
		}
	}
	return hcaNames, nil
}

// ValidateRDMADeviceName requires an ASCII letter first, followed by ASCII letters, digits, or underscores.
func ValidateRDMADeviceName(name string) error {
	if name == "void" || !hcaNamePattern.MatchString(name) {
		return fmt.Errorf("invalid RDMA device name %q", name)
	}
	return nil
}

var rdmaSysfsDir = infinibandSysfsDir
var rdmaVerbsName = regexp.MustCompile(`^uverbs[0-9]+$`)

// ResolveRdmaDeviceNode maps an HCA to exactly one uverbs node; missing or ambiguous mappings fail.
func ResolveRdmaDeviceNode(hcaName string) (string, error) {
	hcaSysfsPath, err := selectedRdmaHCAPath(hcaName)
	if err != nil {
		return "", err
	}

	verbsPath := filepath.Join(hcaSysfsPath, hcaDeviceDir, infinibandVerbsDir)
	entries, err := os.ReadDir(verbsPath)
	if err != nil {
		return "", fmt.Errorf("HCA %s: read %s: %w", hcaName, verbsPath, err)
	}
	verbsNames := make([]string, 0)
	for _, entry := range entries {
		if rdmaVerbsName.MatchString(entry.Name()) {
			verbsNames = append(verbsNames, entry.Name())
		}
	}
	if len(verbsNames) != 1 {
		return "", fmt.Errorf("HCA %s: %s requires exactly one uverbs mapping, got %d", hcaName, verbsPath, len(verbsNames))
	}

	mapping := filepath.Join(verbsPath, verbsNames[0])
	info, err := os.Stat(mapping)
	if err != nil {
		return "", fmt.Errorf("HCA %s: stat %s: %w", hcaName, mapping, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("HCA %s: mapping %s is not a directory", hcaName, mapping)
	}
	return filepath.Join(infinibandNodeDir, verbsNames[0]), nil
}

func selectedRdmaHCAPath(hcaName string) (string, error) {
	if err := ValidateRDMADeviceName(hcaName); err != nil {
		return "", err
	}
	hcaSysfsPath := filepath.Join(rdmaSysfsDir, hcaName)
	info, err := os.Stat(hcaSysfsPath)
	if err != nil {
		return "", fmt.Errorf("HCA %s: stat %s: %w", hcaName, hcaSysfsPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("HCA %s: %s is not a directory", hcaName, hcaSysfsPath)
	}
	return hcaSysfsPath, nil
}
