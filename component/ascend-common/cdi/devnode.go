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

// Package cdi — devnode.go
//
// Device-node generation (per-device davinci* and shared manager / UB
// nodes), device path constants, and low-level stat helpers.
package cdi

import (
	"ascend-common/api"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"ascend-common/common-utils/hwlog"
)

// ---------------------------------------------------------------------------
// Device path constants
// ---------------------------------------------------------------------------

const (
	devicePath           = "/dev/"
	davinciName          = "davinci"
	virtualDavinciName   = "vdavinci"
	davinciManager       = "davinci_manager"
	davinciManagerDocker = "davinci_manager_docker"
	devmmSvm             = "devmm_svm"
	hisiHdc              = "hisi_hdc"
	dvppCmdList          = "dvpp_cmdlist"

	// Ascend310B‑specific manager devices
	svm0    = "svm0"
	tsAisle = "ts_aisle"
	upgrade = "upgrade"
	sys     = "sys"
	vdec    = "vdec"
	vpc     = "vpc"
	pngd    = "pngd"
	venc    = "venc"
	logDrv  = "log_drv"
	acodec  = "acodec"
	aiDev   = "ai"
	ao      = "ao"
	vo      = "vo"
	hdmi    = "hdmi"

	// UB device directories
	uburma = "uburma"
	ummu   = "ummu"

	// Device type
	deviceTypeChar = "c"
)

const (
	Atlas200ISoc = "Atlas 200I SoC A1"
	Atlas200     = "Atlas 200 Model 3000"
)

var ascend310BManagerDevices = []string{
	svm0, tsAisle, upgrade, sys, vdec, vpc, pngd, venc,
	logDrv, acodec, aiDev, ao, vo, hdmi,
}

// Overridable in tests to mock device files on non-Ascend machines.
var (
	unixStat  = unix.Stat
	osStat    = os.Stat
	osReadDir = os.ReadDir
)

// GeneratePerDeviceNodes returns per-device /dev/davinci{id} nodes, each wrapped
// in a cdispec.Device entry with its own ContainerEdits.
// hostFsPrefix changes the stat path, not the CDI host/container paths.
func GeneratePerDeviceNodes(deviceIDs []int, useVirtual bool, hostFsPrefix string) ([]cdispec.Device, error) {
	if len(deviceIDs) == 0 {
		return nil, fmt.Errorf("cdi: no logic IDs provided")
	}
	hostName := davinciName
	if useVirtual {
		hostName = virtualDavinciName
	}
	devices := make([]cdispec.Device, len(deviceIDs))
	for i, id := range deviceIDs {
		if id < 0 {
			return nil, fmt.Errorf("cdi: invalid logic ID %d", id)
		}
		hostDevPath := devicePath + hostName + strconv.Itoa(id)
		containerDevPath := devicePath + davinciName + strconv.Itoa(id)
		dev, err := statDevice(hostDevPath, containerDevPath, hostFsPrefix)
		if err != nil {
			return nil, fmt.Errorf("cdi: failed to get device info for %s: %w", hostDevPath, err)
		}
		devices[i] = cdispec.Device{
			Name: strconv.Itoa(id),
			ContainerEdits: cdispec.ContainerEdits{
				DeviceNodes: []*cdispec.DeviceNode{dev},
			},
		}
	}
	return devices, nil
}

// GenerateSharedNodes returns all shared (non-per-device) manager and UB
// device nodes for the given device type and product type.
func GenerateSharedNodes(devType, productType, hostFsPrefix string) ([]*cdispec.DeviceNode, error) {
	if devType == "" {
		return nil, fmt.Errorf("cdi: devType is empty")
	}

	mgr, err := resolveDavinciManager(devType, hostFsPrefix)
	if err != nil {
		return nil, err
	}
	devices := []*cdispec.DeviceNode{mgr}

	// dvpp_cmdlist: all types except Ascend910A5
	if devType != api.Ascend910A5 {
		appendDeviceNode(&devices, dvppCmdList, hostFsPrefix)
	}

	// Ascend310B: 14 special manager devices
	if devType == api.Ascend310B {
		for _, name := range ascend310BManagerDevices {
			appendDeviceNode(&devices, name, hostFsPrefix)
		}
	}

	// Common manager devices (devmm_svm, hisi_hdc)
	// Skip: Atlas200 products, Ascend310B
	// Ascend910A5: hisi_hdc only; others: both
	if devType != api.Ascend310B && productType != Atlas200ISoc && productType != Atlas200 {
		if devType == api.Ascend910A5 {
			appendDeviceNode(&devices, hisiHdc, hostFsPrefix)
		} else {
			appendDeviceNode(&devices, devmmSvm, hostFsPrefix)
			appendDeviceNode(&devices, hisiHdc, hostFsPrefix)
		}
	}

	addUBDevicesFromDir(&devices, uburma, hostFsPrefix)
	addUBDevicesFromDir(&devices, ummu, hostFsPrefix)

	return devices, nil
}

// appendDeviceNode stats /dev/<name> and appends a DeviceNode to devices.
// Errors are logged as warnings; the device is silently skipped.
func appendDeviceNode(devices *[]*cdispec.DeviceNode, name, hostFsPrefix string) {
	devPath := devicePath + name
	d, err := statDevice(devPath, devPath, hostFsPrefix)
	if err != nil {
		hwlog.RunLog.Warnf("failed to add %s to spec: %v", name, err)
		return
	}
	*devices = append(*devices, d)
}

// resolveDavinciManager resolves the davinci_manager device path.
// Ascend310B tries the docker variant first, falling back to the standard path.
// The container path is always /dev/davinci_manager to match legacy behavior
// (getMountPath rewrites the docker variant to davinci_manager).
// Returns an error if the device cannot be found (davinci_manager is mandatory).
func resolveDavinciManager(devType, hostFsPrefix string) (*cdispec.DeviceNode, error) {
	hostPath := devicePath + davinciManager
	if devType == api.Ascend310B {
		dockerPath := devicePath + davinciManagerDocker
		if _, err := osStat(deviceStatPath(hostFsPrefix, dockerPath)); err == nil {
			hostPath = dockerPath
		}
	}
	return statDevice(hostPath, devicePath+davinciManager, hostFsPrefix)
}

// addUBDevicesFromDir reads devices from /dev/<dir>/* and appends them.
func addUBDevicesFromDir(devices *[]*cdispec.DeviceNode, dir, hostFsPrefix string) {
	entries, err := addDevicesFromDir(devicePath+dir, hostFsPrefix)
	if err != nil {
		hwlog.RunLog.Warnf("failed to add %s devices: %v", dir, err)
		return
	}
	*devices = append(*devices, entries...)
}

// ---------------------------------------------------------------------------
// Low-level device helpers
// ---------------------------------------------------------------------------

// deviceStatPath maps the host device path to its location inside the spec builder.
func deviceStatPath(hostFsPrefix, hostPath string) string {
	if hostFsPrefix == "" {
		return hostPath
	}
	return filepath.Join(hostFsPrefix, strings.TrimPrefix(hostPath, "/"))
}

// statDevice creates a cdispec.DeviceNode by calling unix.Stat on the node
// visible to the spec builder, while preserving the actual host path in CDI.
// Uses unix.Major and unix.Minor on the Rdev field to get the device
// major/minor numbers.  Does NOT use containerd/oci.DeviceFromPath.
func statDevice(hostPath, containerPath, hostFsPrefix string) (*cdispec.DeviceNode, error) {
	statPath := deviceStatPath(hostFsPrefix, hostPath)
	var st unix.Stat_t
	if err := unixStat(statPath, &st); err != nil {
		return nil, fmt.Errorf("stat %s: %w", statPath, err)
	}

	mode := os.FileMode(st.Mode)

	return &cdispec.DeviceNode{
		Path:     containerPath,
		HostPath: hostPath,
		Type:     deviceTypeChar,
		Major:    int64(unix.Major(st.Rdev)),
		Minor:    int64(unix.Minor(st.Rdev)),
		FileMode: &mode,
	}, nil
}

// addDevicesFromDir reads all non-directory entries from a device directory
// and returns them as cdispec.DeviceNode entries (mirrors process.go addDevicesInDir).
func addDevicesFromDir(dirPath, hostFsPrefix string) ([]*cdispec.DeviceNode, error) {
	statDir := deviceStatPath(hostFsPrefix, dirPath)
	entries, err := osReadDir(statDir)
	if err != nil {
		return nil, fmt.Errorf("read device dir %s: %w", statDir, err)
	}

	var devices []*cdispec.DeviceNode
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(dirPath, entry.Name())
		dev, err := statDevice(fullPath, fullPath, hostFsPrefix)
		if err != nil {
			return nil, fmt.Errorf("add %s to spec: %w", fullPath, err)
		}
		devices = append(devices, dev)
	}
	return devices, nil
}
