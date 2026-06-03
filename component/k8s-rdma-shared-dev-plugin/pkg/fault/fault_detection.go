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

// Package fault for fault check and fault report
package fault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
)

var faultResultChan = make(chan DpuInfoCfg, 1)

const FaultScriptPath = "/etc/rdma-plugin/fault_detection.sh"
const FaultConfigPath = "/etc/rdma-plugin/fault_code.json"

const (
	readLimitBytes = 1024
)

const (
	CheckUbPort      = "check_ub_port"
	CheckUbLane      = "check_ub_lane"
	CheckHcaPort     = "check_hca_port"
	CheckBondMember  = "check_bond_member"
	CheckDpuCardDrop = "check_dpu_card_drop"
)

// FaultConfig represents a fault configuration item
type FaultConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FaultCode   string `json:"faultcode"`
	FaultLevel  string `json:"faultlevel"`
	CheckMethod string `json:"check_method"`
	DependsOn   string `json:"depends_on,omitempty"`
}

// FaultConfigList represents a list of fault configurations
type FaultConfigList struct {
	Faults []FaultConfig `json:"faults"`
}

// FaultResult represents the result of a fault check
type FaultResult struct {
	Fault   FaultConfig
	HCA     string
	Result  string
	Details string
}

// CheckFunc defines the signature for fault check functions
type CheckFunc func(hca string) (string, string)

var checkFuncMap = map[string]CheckFunc{
	CheckUbPort:      checkUbPort,
	CheckUbLane:      checkUbLane,
	CheckHcaPort:     checkHcaPort,
	CheckBondMember:  checkBondMember,
	CheckDpuCardDrop: checkDpuCardDrop,
}

// LoadFaultConfig loads the fault configuration from the config file
func LoadFaultConfig() (*FaultConfigList, error) {
	configData, err := utils.LoadFile(FaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fault config file: %v", err)
	}
	if configData == nil {
		return nil, fmt.Errorf("fault config file not found")
	}

	var config FaultConfigList
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fault config: %v", err)
	}

	return &config, nil
}

func runShellCommand(cmd string) (string, string) {
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return "false", fmt.Sprintf("command failed: %v, output: %s", err, output)
	}

	trimmedOutput := strings.TrimSpace(string(output))
	parts := strings.SplitN(trimmedOutput, ":", 2)
	if len(parts) < 2 {
		return "false", fmt.Sprintf("invalid output format: %s", trimmedOutput)
	}

	return parts[0], parts[1]
}

func runShellFunction(scriptPath, functionName string, args ...string) (string, string) {
	shellCmd := fmt.Sprintf("%s %s", scriptPath, functionName)
	for _, arg := range args {
		shellCmd += " " + arg
	}
	return runShellCommand(shellCmd)
}

// RunFaultChecks runs all fault checks specified in the config against the given HCA devices
func RunFaultChecks(config *FaultConfigList, hcas []string) []FaultResult {
	results := []FaultResult{}

	for _, fault := range config.Faults {
		if fault.Name == "dpu_card_drop" {
			result, details := runCheck(fault, "")
			results = append(results, FaultResult{
				Fault:   fault,
				HCA:     "",
				Result:  result,
				Details: details,
			})
			continue
		}

		for _, hca := range hcas {
			result, details := runCheck(fault, hca)
			results = append(results, FaultResult{
				Fault:   fault,
				HCA:     hca,
				Result:  result,
				Details: details,
			})
		}
	}

	return results
}

func runCheck(fault FaultConfig, hca string) (string, string) {
	checkFunc, exists := checkFuncMap[fault.CheckMethod]
	if !exists {
		return "false", fmt.Sprintf("check method %s not found", fault.CheckMethod)
	}
	return checkFunc(hca)
}

func checkUbPort(hca string) (string, string) {
	return runShellFunction(FaultScriptPath, "check_ub_port", hca)
}

func checkUbLane(hca string) (string, string) {
	return runShellFunction(FaultScriptPath, "check_ub_lane", hca)
}

func checkHcaPort(hca string) (string, string) {
	statePath := fmt.Sprintf("%s/%s/ports/1/state", common.SysClassInfiniband, hca)
	physStatePath := fmt.Sprintf("%s/%s/ports/1/phys_state", common.SysClassInfiniband, hca)

	state, err := utils.ReadLimitBytes(statePath, readLimitBytes)
	if err != nil {
		return "false", fmt.Sprintf("port state: UNKNOWN, port phys_state: UNKNOWN")
	}

	physState, err := utils.ReadLimitBytes(physStatePath, readLimitBytes)
	if err != nil {
		return "false", fmt.Sprintf("port state: %s, port phys_state: UNKNOWN", strings.TrimSpace(string(state)))
	}

	stateStr := strings.TrimSpace(string(state))
	physStateStr := strings.TrimSpace(string(physState))

	if !strings.Contains(stateStr, "ACTIVE") || !strings.Contains(physStateStr, "LinkUp") {
		return "true", fmt.Sprintf("port state: %s, port phys_state: %s", stateStr, physStateStr)
	}

	return "false", fmt.Sprintf("port state: %s, port phys_state: %s", stateStr, physStateStr)
}

func checkBondMember(hca string) (string, string) {
	ethName := GetHCAEthName(hca)
	if ethName == "" {
		return "false", fmt.Sprintf("cannot get eth name for hca %s", hca)
	}

	bondName, slaves, err := findBondByEthName(ethName)
	if err != nil {
		return "false", fmt.Sprintf("bond not found for eth %s: %v", ethName, err)
	}

	if bondName == "" {
		return "false", fmt.Sprintf("no bond contains eth %s", ethName)
	}

	return checkBondSlavesState(bondName, slaves, hca)
}

func findBondByEthName(ethName string) (string, []string, error) {
	netDirs, err := os.ReadDir(common.SysClassNet)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read %s: %v", common.SysClassNet, err)
	}

	for _, bondDir := range netDirs {
		bondName := bondDir.Name()
		if !strings.HasPrefix(bondName, "bond") {
			continue
		}

		slaves, err := getBondSlaves(bondName)
		if err != nil || len(slaves) != 2 {
			continue
		}

		if utils.Contains(slaves, ethName) {
			return bondName, slaves, nil
		}
	}

	return "", nil, nil
}

func getBondSlaves(bondName string) ([]string, error) {
	slavesPath := filepath.Join(common.SysClassNet, bondName, "bonding", "slaves")
	slavesContent, err := utils.ReadLimitBytes(slavesPath, readLimitBytes)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(slavesContent)), nil
}

func checkBondSlavesState(bondName string, slaves []string, hca string) (string, string) {
	downCount := 0
	var failedSlave string

	for _, slave := range slaves {
		if isEthPortDown(slave) {
			downCount++
			failedSlave = slave
		}
	}

	if downCount == 1 {
		return "true", fmt.Sprintf("bond %s: one member %s down, hca=%s",
			bondName, failedSlave, hca)
	}

	if downCount == len(slaves) {
		return "false", fmt.Sprintf("bond %s: all members down, hca=%s", bondName, hca)
	}

	return "false", fmt.Sprintf("no bond member failure detected for hca %s", hca)
}

func isEthPortDown(ifName string) bool {
	operstatePath := filepath.Join(common.SysClassNet, ifName, "operstate")
	operstate, err := utils.ReadLimitBytes(operstatePath, readLimitBytes)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(operstate)) == "down"
}

func checkDpuCardDrop(hca string) (string, string) {
	return runShellFunction(FaultScriptPath, "check_dpu_card_drop", hca)
}

// LogFaults logs the detected faults at ERROR level
func LogFaults(results []FaultResult) {
	for _, result := range results {
		if result.Result == "true" {
			hcaLabel := result.HCA
			if hcaLabel == "" {
				hcaLabel = "GLOBAL"
			}
			hwlog.RunLog.Errorf("FAULT DETECTED: HCA=%s, Code=%s, Level=%s, Description=%s",
				hcaLabel, result.Fault.FaultCode, result.Fault.FaultLevel, result.Fault.Description)
			hwlog.RunLog.Errorf("Details: %s", result.Details)
		}
	}
}

// StartFaultDetection starts the fault detection loop
// It periodically runs fault checks and sends results to the faultResultChan for reporting
func StartFaultDetection(ctx context.Context, hcaList []string, faultDetectPeriod int) {
	ticker := time.NewTicker(time.Duration(faultDetectPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("Fault detection goroutine stopped")
			return
		case <-ticker.C:
			hwlog.RunLog.Info("Starting fault detection...")
			config, err := LoadFaultConfig()
			if err != nil {
				hwlog.RunLog.Errorf("Failed to load fault config: %v", err)
				continue
			}

			results := RunFaultChecks(config, hcaList)

			LogFaults(results)

			dpuCfg := BuildDPUInfoCfg(results)
			// Non-blocking send to reporter goroutine
			select {
			case faultResultChan <- dpuCfg:
			default:
				hwlog.RunLog.Warn("Fault result channel full, dropping oldest result")
				// Clear old data and send new
				select {
				case <-faultResultChan:
				default:
				}
				faultResultChan <- dpuCfg
			}
		}
	}
}
