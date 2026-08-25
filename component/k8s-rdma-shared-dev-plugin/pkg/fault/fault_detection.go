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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

var hinicHeaderRe = regexp.MustCompile(`(hinic\d+)`)
var nicLineRe = regexp.MustCompile(`NIC:([^\s)]+)`)

func validateSysfsPath(resolvedPath string) bool {
	return strings.HasPrefix(resolvedPath, "/sys/")
}

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
	Hca     string
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
			hca := ""
			if len(hcas) > 0 {
				hca = hcas[0]
			}
			result, details := runCheck(fault, hca)
			results = append(results, FaultResult{
				Fault:   fault,
				Hca:     "",
				Result:  result,
				Details: details,
			})
			continue
		}

		for _, hca := range hcas {
			result, details := runCheck(fault, hca)
			results = append(results, FaultResult{
				Fault:   fault,
				Hca:     hca,
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

	state, err := utils.ReadLimitBytesWithSymlink(statePath, readLimitBytes, validateSysfsPath)
	if err != nil {
		return "false", fmt.Sprintf("port state: UNKNOWN, port phys_state: UNKNOWN")
	}

	physState, err := utils.ReadLimitBytesWithSymlink(physStatePath, readLimitBytes, validateSysfsPath)
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
	ethName := GetHcaEthName(hca)
	if ethName == "" {
		return "false", fmt.Sprintf("cannot get eth name for hca %s", hca)
	}
	siblings, err := GetSiblingEthFor1825(ethName)
	if err != nil {
		return "false", fmt.Sprintf("get sibling eth failed: %v", err)
	}

	for _, eth := range siblings {
		// 判断是否属于bond, 如果master不存在， 则说明不是bond成员, 如果不配置bond，则全部跳过
		if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s/master", eth)); err != nil {
			continue
		}
		if isEthPortDown(eth) {
			return "true", fmt.Sprintf("bonding down: %v", eth)
		}
	}
	return "false", fmt.Sprintf("bond for %s is ok or not exist", hca)
}

/**
 *
 *hinicadm5 info output lis like:
 *     Card       UB Entity
 * |---- hinic0(CAL_2X400G_UB_EXE)
 *       |------- 0000f(NIC:ens0p0)
 *       |------- 0000f(NIC:ens0p1)
 * |---- hinic1(CAL_2X400G_UB_EXE)
 *       |------- 0000f(NIC:ens1p0)
 * 从中解析出hinic0中包含的 eth网卡列表
 */
func GetSiblingEthFor1825(ethName string) ([]string, error) {
	output, err := exec.Command("bash", "-c", "hinicadm5 info", ethName).CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("failed to run hinicadm5 info: %v, output: %s", err, output)
	}
	result, err1 := parseHinicNICMap(string(output))
	if err1 != nil {
		return []string{}, fmt.Errorf("parse hinic nic map failed: %v", err1)
	}

	for _, value := range result {
		if utils.Contains(value, ethName) {
			return value, nil
		}
	}

	return []string{}, fmt.Errorf("eth %s not found in any hinic", ethName)
}

func parseHinicNICMap(output string) (map[string][]string, error) {
	result := make(map[string][]string)
	var currentKey string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 命中新的 hinicX 头
		if m := hinicHeaderRe.FindStringSubmatch(trimmed); m != nil {
			currentKey = m[1]
			if _, ok := result[currentKey]; !ok {
				result[currentKey] = make([]string, 0)
			}
			continue
		}

		// 命中 NIC 子项：归入当前 hinic
		if m := nicLineRe.FindStringSubmatch(trimmed); m != nil {
			if currentKey == "" {
				// 没有对应的 hinic 头，跳过孤立 NIC 行
				continue
			}
			result[currentKey] = append(result[currentKey], m[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan output failed: %v", err)
	}
	return result, nil
}

func findBondByEthName(ethName string) (string, []string, error) {
	netDirs, err := os.ReadDir(common.SysClassNet)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read %s: %v", common.SysClassNet, err)
	}

	for _, bondDir := range netDirs {
		bondName := bondDir.Name()

		bondingPath := filepath.Join(common.SysClassNet, bondName, "bonding")
		if _, err := os.Stat(bondingPath); err != nil {
			// bonding dir not exist, not a bond device, skip
			continue
		}

		slaves, err := getBondSlaves(bondName)
		if err != nil {
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
	slavesContent, err := utils.ReadLimitBytesWithSymlink(slavesPath, readLimitBytes, validateSysfsPath)
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
	operstate, err := utils.ReadLimitBytesWithSymlink(operstatePath, readLimitBytes, validateSysfsPath)
	if err != nil {
		return false
	}
	// operstate state down mearns ethernet status down
	if strings.TrimSpace(string(operstate)) == "down" {
		return true
	}

	carrierPath := filepath.Join(common.SysClassNet, ifName, "carrier")
	carrierState, err := utils.ReadLimitBytesWithSymlink(carrierPath, readLimitBytes, validateSysfsPath)
	if err != nil {
		hwlog.RunLog.Debugf("Read carrier state for %s failed, carrier error", ifName)
		return true
	}
	// carrier state 0 mearns phy link down
	return strings.TrimSpace(string(carrierState)) == "0"
}

func checkDpuCardDrop(hca string) (string, string) {
	return runShellFunction(FaultScriptPath, "check_dpu_card_drop", hca)
}

// StartFaultDetection starts the fault detection loop
// It periodically runs fault checks and sends results to the faultResultChan for reporting
// getHcaList is called on startup and whenever rediscoverCh signals to refresh the cached HCA list
func StartFaultDetection(ctx context.Context, getHcaList func() []string, rediscoverCh <-chan struct{},
	faultDetectPeriod int, healthCallback func(hcaNames []string)) {
	ticker := time.NewTicker(time.Duration(faultDetectPeriod) * time.Second)
	defer ticker.Stop()

	hcaList := getHcaList()
	hwlog.RunLog.Infof("Fault detection started with %d HCAs", len(hcaList))

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("Fault detection goroutine stopped")
			return
		case <-rediscoverCh:
			hcaList = getHcaList()
			hwlog.RunLog.Infof("HCA list refreshed, now has %d HCAs", len(hcaList))
		case <-ticker.C:
			hwlog.RunLog.Debug("Starting fault detection...")
			config, err := LoadFaultConfig()
			if err != nil {
				hwlog.RunLog.Errorf("Failed to load fault config: %v", err)
				continue
			}

			results := RunFaultChecks(config, hcaList)
			dpuCfg := BuildDPUInfoCfg(results)
			if healthCallback != nil {
				unhealthyHcas := make([]string, 0)
				for _, result := range results {
					if result.Result != "true" || result.Hca == "" {
						continue
					}
					level := result.Fault.FaultLevel
					if level == NotHandleFault || level == SubHealthFault {
						continue
					}
					hwlog.RunLog.Warnf("Fault on HCA %s (code=%s, level=%s) added to unhealthy batch",
						result.Hca, result.Fault.FaultCode, level)
					unhealthyHcas = append(unhealthyHcas, result.Hca)
				}
				healthCallback(unhealthyHcas)
			}
			select {
			case faultResultChan <- dpuCfg:
			default:
				hwlog.RunLog.Warn("Fault result channel full, dropping oldest result")
				select {
				case <-faultResultChan:
				default:
				}
				select {
				case faultResultChan <- dpuCfg:
				case <-ctx.Done():
					hwlog.RunLog.Info("Fault detection goroutine stopped during send")
					return
				}
			}
		}
	}
}
