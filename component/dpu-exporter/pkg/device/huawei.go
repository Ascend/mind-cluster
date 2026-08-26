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

package device

import (
	"ascend-common/common-utils/utils"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"huawei.com/dpu-exporter/utils/logger"
)

const (
	// defaultHinicadm5Path is the default path to the hinicadm5 binary
	defaultHinicadm5Path = "/usr/sbin/hinicadm5"
	// cmdTimeout is the max duration for a hinicadm5 command
	cmdTimeout = 30 * time.Second
	// infoSubCmd is the subcommand for querying device info
	infoSubCmd = "info"
	// sysfsNetBase is the base path for network interface sysfs entries
	sysfsNetBase = "/sys/class/net"
	// maxBinarySizeMB is the max allowed binary file size in MB
	maxBinarySizeMB int64 = 200
	// splitParts is the number of parts for strings.SplitN
	splitParts = 2
)

// HwDpuManager implements DeviceManager for Huawei DPU hardware.
// It uses hinicadm5 CLI for global metrics and sysfs for interface-level metrics.
// Per the class diagram, HwDpuManager has ExecHinicadm5 as a Huawei-specific method
// and implements the generic DeviceManager interface.
type HwDpuManager struct {
	hinicadm5Path string
	sysfsBasePath string
	dpuList       []DPU
}

// NewHwDpuManager creates a HwDpuManager with default paths
func NewHwDpuManager() (*HwDpuManager, error) {
	path, err := exec.LookPath("hinicadm5")
	if err != nil {
		path = defaultHinicadm5Path
		logger.Warnf("hinicadm5 not found in PATH, using default: %s", path)
	}
	if err := validateBinaryPath(path); err != nil {
		return nil, fmt.Errorf("invalid hinicadm5 path %s: %v", path, err)
	}
	return &HwDpuManager{
		hinicadm5Path: path,
		sysfsBasePath: sysfsNetBase,
	}, nil
}

// validateBinaryPath checks that a binary path is a regular file owned by root and not a symlink.
func validateBinaryPath(path string) error {
	// RealFileChecker: reject symlinks (allowLink=false), non-regular files, oversized files
	if _, err := utils.RealFileChecker(path, false, false, maxBinarySizeMB); err != nil {
		return fmt.Errorf("real file check failed: %v", err)
	}
	// VerifyFile: check file is not symlink and owner uid matches current euid (root)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()
	if err := utils.VerifyFile(file, maxBinarySizeMB); err != nil {
		return fmt.Errorf("verify file failed: %v", err)
	}
	return nil
}

// NewHwDpuManagerWithPath creates a HwDpuManager with custom paths (for testing)
func NewHwDpuManagerWithPath(hinicadm5Path, sysfsBasePath string) *HwDpuManager {
	return &HwDpuManager{
		hinicadm5Path: hinicadm5Path,
		sysfsBasePath: sysfsBasePath,
	}
}

// AutoInit discovers DPU cards and their interfaces by running `hinicadm5 info`
func (m *HwDpuManager) AutoInit() error {
	output, err := m.ExecHinicadm5(infoSubCmd)
	if err != nil {
		return fmt.Errorf("failed to execute hinicadm5 info: %w", err)
	}
	dpuList := parseHinicadm5Info(output)
	if len(dpuList) == 0 {
		return fmt.Errorf("no DPU devices detected")
	}
	m.dpuList = dpuList
	logger.Infof("discovered %d DPU cards", len(dpuList))
	return nil
}

// GetDpuList returns the discovered DPU list
func (m *HwDpuManager) GetDpuList() []DPU {
	return m.dpuList
}

// ExecCommand implements DeviceManager.ExecCommand — delegates to hinicadm5
func (m *HwDpuManager) ExecCommand(args ...string) (string, error) {
	return m.ExecHinicadm5(args...)
}

// ExecHinicadm5 executes the hinicadm5 command with given args and returns stdout.
// This is the Huawei-specific method per the class diagram.
func (m *HwDpuManager) ExecHinicadm5(args ...string) (string, error) {
	cmd := exec.Command(m.hinicadm5Path, args...)
	output, err := execWithTimeout(cmd, cmdTimeout)
	if err != nil {
		logger.Errorf("hinicadm5 %s failed: %v, output: %s", strings.Join(args, " "),
			err, output)
	}
	return output, err
}

// ReadSysfs implements DeviceManager.ReadSysfs — reads a file at the given path
func (m *HwDpuManager) ReadSysfs(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read sysfs %s failed: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// ListDir implements DeviceManager.ListDir — lists file names in a directory
func (m *HwDpuManager) ListDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("list dir %s failed: %w", path, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// ReadSysfsMetric is a convenience method that constructs the sysfs path and reads the file.
// Path: <sysfsBasePath>/<ethName>/statistics/<fileName>
func (m *HwDpuManager) ReadSysfsMetric(ethName, fileName string) (string, error) {
	path := filepath.Join(m.sysfsBasePath, ethName, "statistics", fileName)
	return m.ReadSysfs(path)
}

// GetCardType implements DeviceManager.GetCardType
func (m *HwDpuManager) GetCardType() string {
	return CardTypeHuawei
}

// parseHinicadm5Info parses the output of `hinicadm5 info` into a DPU list.
// Output format:
//
//	Card num:4
//	Device Information:
//	     Card         UB Entity
//	|----hinic0(CAL_2X400G_UB_EXP)
//	          |--------0000f(NIC:ens1f0)
//	          |--------00010(NIC:ens1p1)
func parseHinicadm5Info(output string) []DPU {
	var dpuList []DPU
	lines := strings.Split(output, "\n")

	var currentDPU *DPU
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Interface line must be checked BEFORE card line
		// because |-------- also starts with |----
		if strings.HasPrefix(line, "|--------") && currentDPU != nil {
			ethName := parseInterfaceLine(line)
			if ethName != "" {
				iface := NewInterface(ethName, currentDPU.CardName,
					"/sys/class/net/"+ethName)
				currentDPU.Interfaces = append(currentDPU.Interfaces, iface)
			}
			continue
		}

		// DPU card line: |----hinic0(CAL_2X400G_UB_EXP)
		if strings.HasPrefix(line, "|----") {
			cardName, cardType := parseCardLine(line)
			if cardName != "" {
				dpu := NewDPU(cardName, cardType, cardName)
				dpuList = append(dpuList, dpu)
				currentDPU = &dpuList[len(dpuList)-1]
			}
			continue
		}
	}
	return dpuList
}

// parseCardLine extracts card name and type from "|----hinic0(CAL_2X400G_UB_EXP)"
func parseCardLine(line string) (cardName, cardType string) {
	rest := strings.TrimPrefix(line, "|----")
	parts := strings.SplitN(rest, "(", splitParts)
	if len(parts) < 2 {
		return "", ""
	}
	cardName = strings.TrimSpace(parts[0])
	cardType = strings.TrimSuffix(parts[1], ")")
	return cardName, cardType
}

// parseInterfaceLine extracts interface name from "|--------0000f(NIC:ens1f0)"
func parseInterfaceLine(line string) string {
	parts := strings.SplitN(line, "NIC:", splitParts)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
}

// execWithTimeout runs a command with a timeout and returns combined output.
// Uses context.Context to avoid the race between timer and cmd.Process being nil.
func execWithTimeout(cmd *exec.Cmd, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
