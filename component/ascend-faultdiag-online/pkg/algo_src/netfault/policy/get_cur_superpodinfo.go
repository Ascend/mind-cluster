/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package policy is used for processing superpod information
package policy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

const numSplits = 2

const configFile = "cathelper.conf"

/* 解析server级别topo需要的参数 */
type superPodParam struct {
	superPodId    string
	protocol      string
	rack          *RackInfo
	server        *ServerInfo
	npu           *NpuInfo
	protocolPorts *PortInfo
}

// DiagVersionA3 故障检测A3字符串
const DiagVersionA3 = "A3"

const logPrintInterval = 10

// NpuType 网络拓扑类型
const NpuType = "npu_type"

// ServerIdMap nodeName与serverId的映射
const ServerIdMap = "serverIdMap"

type npuMapParam struct {
	superPodInfo   *SuperPodInfo
	typeStr        string
	rackNpuMap     map[string]bool
	serverTopology *RackTopology
}

func spliceSuperPodFilePath(superPodPath string) string {
	copyPath := superPodPath
	/* 从/xx/xx/super-pod-0/ 获取当前超节点id */
	// 获取路径的最后一部分
	lastLevel := filepath.Base(copyPath)

	// 去除末尾的斜杠
	superPodJsonFile := strings.TrimSuffix(lastLevel, "/")

	fileName := superPodJsonFile + ".json"
	retStr := superPodPath + "/" + fileName
	retStr = filepath.Clean(retStr)
	hwlog.RunLog.Infof("Read superPodJsonFile:%s", retStr)
	return retStr
}

/* 获取超级点内探测pingList和super-pod-i.json内容 */
func getCurrentSuperPodInfo(
	superPodPath string,
	detectObj *algo.NetDetect) (*SuperPodInfo, map[string]any) {
	if superPodPath == "" {
		hwlog.RunLog.Errorf("Invalid config path")
		return nil, nil
	}

	superPodJsonFile := spliceSuperPodFilePath(superPodPath)
	superPodInfo, fullMesh, linkPath := processSuperPodJson(superPodJsonFile, superPodPath)
	if superPodInfo == nil && linkPath == nil &&
		len(fullMesh) == 0 {
		return nil, nil
	}
	/* 拼接算法生成pingList接口的入参 */
	algoPingListInput := spliceAlgorithmInput(fullMesh, linkPath)
	if algoPingListInput == nil {
		return nil, nil
	}
	jsonPingList := detectObj.GenPingStrategy(algoPingListInput)
	if jsonPingList == nil {
		return nil, nil
	}
	return superPodInfo, jsonPingList
}

// 解析superPodJsonFile文件
func processSuperPodJson(superPodJsonFile string, superPodPath string) (*SuperPodInfo, []string, map[string][]string) {
	if !loopWaitFile(superPodJsonFile, superPodPath) {
		return nil, nil, nil
	}
	superPodInfo := readConfigMap(superPodJsonFile)
	if superPodInfo == nil {
		return nil, nil, nil
	}
	switch superPodInfo.Version {
	case DiagVersionA3:
		fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(superPodInfo)
		return superPodInfo, fullMesh, linkPath
	default:
		hwlog.RunLog.Errorf(" %s version info error,the value %s", superPodJsonFile, superPodInfo.Version)
		return nil, nil, nil
	}
}

// SetCallAlgorithmParamInfo 设置算法参数
func SetCallAlgorithmParamInfo(superPodId int, superPodFilePath string,
	callAlgorithmParam map[string]any) error {
	if callAlgorithmParam == nil {
		return errors.New("callAlgorithmParam is nullptr")
	}

	superPodFile := fmt.Sprintf("super-pod-%d.json", superPodId)
	superPodFile = superPodFilePath + "/" + superPodFile
	superPodFile = filepath.Clean(superPodFile)

	if !loopWaitFile(superPodFile, superPodFilePath) {
		return errors.New("loop wait failed")
	}
	superPodInfo := readConfigMap(superPodFile)
	if superPodInfo == nil {
		return errors.New("super pod info is nil")
	}

	if superPodInfo.Version != DiagVersionA3 {
		return fmt.Errorf("unexpected %s version", superPodFile)
	}
	callAlgorithmParam[NpuType] = superPodInfo.Version
	/* A3 */
	callAlgorithmParam["pingObjType"] = 1

	// A3网络结构设置nodeName与serverId的映射
	return getWorKMapping(callAlgorithmParam, superPodInfo)
}

func getWorKMapping(callAlgorithmParam map[string]any, superPodInfo *SuperPodInfo) error {
	if superPodInfo == nil {
		return errors.New("the superPodInfo is empty")
	}
	if superPodInfo.NodeDeviceMap == nil {
		return errors.New("the NodeDeviceMap is empty")
	}
	serverIdMap, ok := callAlgorithmParam[ServerIdMap].(map[string]string)
	if !ok {
		return errors.New("callAlgorithmParam ServerId Map format error")
	}
	for workId, workInfo := range superPodInfo.NodeDeviceMap {
		if workInfo == nil || len(workInfo.NodeName) == 0 {
			return fmt.Errorf("get work %s NodeName error", workId)
		}
		if len(workInfo.ServerID) == 0 {
			return fmt.Errorf("get work %s ServerId error", workId)
		}
		serverIdMap[workInfo.ServerID] = workInfo.NodeName
	}
	return nil
}

// GetTargetSuperPodNpuMap get target super pod npu eid or ip map
func GetTargetSuperPodNpuMap(superPodFilePath string,
	superPodId int) (bool, map[string]algo.NpuInfo) {
	superPodFile := fmt.Sprintf("super-pod-%d.json", superPodId)
	superPodFile = superPodFilePath + "/" + superPodFile
	superPodFile = filepath.Clean(superPodFile)
	var superPodInfo *SuperPodInfo
	var npuNetplaneInfo map[string][]string
	if !loopWaitFile(superPodFile, superPodFilePath) {
		return false, nil
	}
	superPodInfo = readConfigMap(superPodFile)
	if superPodInfo == nil {
		hwlog.RunLog.Error("[NETFAULT ALGO]read config map failed: superPodInfo is nil")
		return false, nil
	}
	var npuInfoMap = make(map[string]algo.NpuInfo)
	switch superPodInfo.Version {
	case DiagVersionA3:
		_, npuNetplaneInfo = GetCurSuperPodInfoFromMapA3(superPodInfo)
		if len(npuNetplaneInfo) == 0 {
			hwlog.RunLog.Error("npu netplane link info is empty!")
			return false, nil
		}
		npuInfoMap = ExtractNPUMapA3(npuNetplaneInfo)
	default:
		hwlog.RunLog.Errorf(" %s version info error,the value %s", superPodFile, superPodInfo.Version)
		return false, nil

	}
	return true, npuInfoMap
}

/* loop wait file */
func loopWaitFile(filePath string, superPodDirPath string) bool {
	for i := 0; i < maxRetryTime && !controllerflags.IsControllerExited.GetState() &&
		CheckCurSuperPodConfigSwitch(superPodDirPath); i++ {
		_, err := os.Stat(filePath)
		/* 不管错误类型 */
		if err != nil && os.IsNotExist(err) {
			if i == maxRetryTime-1 {
				hwlog.RunLog.Errorf("%s retry max time failed!", filePath)
				return false
			}
			if i%logPrintInterval == 0 {
				hwlog.RunLog.Warn(err, " retry:", i+1)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	/* 总体开关检查 */
	if controllerflags.IsControllerExited.GetState() {
		hwlog.RunLog.Info("network detection off")
		return false
	}
	/* 当前超节点开关检查 */
	if !CheckCurSuperPodConfigSwitch(superPodDirPath) {
		hwlog.RunLog.Infof("%s detection switch(off)", superPodDirPath)
		return false
	}
	return true
}

func isAlphanumeric(s string) bool {
	curPattern := "^[a-zA-Z0-9]+$"
	regex := regexp.MustCompile(curPattern)
	return regex.MatchString(s)
}

func containsElement(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func isPureNumber(str string) bool {
	matched, err := regexp.MatchString(`^\d+$`, str)
	if err != nil {
		return false
	}
	return matched
}

func isPureLetter(str string) bool {
	matched, err := regexp.MatchString(`^[a-zA-Z]+$`, str)
	if err != nil {
		return false
	}
	return matched
}

// ReadConfigFromFile 从key=value配置文件中获取指定的所有key
func ReadConfigFromFile(fileContent []byte, targetKeys []string) map[string]any {
	callAlgorithmParam := make(map[string]any)
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过空行和注释行
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析键值对
		parts := strings.SplitN(line, "=", numSplits)
		if len(parts) != numSplits {
			hwlog.RunLog.Errorf("Invalid line format: %v", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		/* 非预期值 */
		if !isAlphanumeric(value) {
			continue
		}
		if isPureNumber(value) && containsElement(targetKeys, key) {
			intValue, err := strconv.Atoi(value)
			if err == nil {
				callAlgorithmParam[key] = intValue
			}
		} else if isPureLetter(value) && containsElement(targetKeys, key) {
			callAlgorithmParam[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		hwlog.RunLog.Errorf("Error reading file: %v", err)
		return nil
	}
	return callAlgorithmParam
}

// CheckCurSuperPodConfigSwitch 判断某个超节点检测开关
func CheckCurSuperPodConfigSwitch(superPodPath string) bool {
	configPath := filepath.Join(superPodPath, configFile)
	/* 需要文件权限、存在、软链接检查等 */
	fileContent, err := fileutils.ReadLimitBytes(configPath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("Open:%v", err)
		return false
	}
	target := []string{"netFault"}
	configParam := ReadConfigFromFile(fileContent, target)
	if len(configParam) == 0 {
		hwlog.RunLog.Errorf("netfault field is not exist in %s", configPath)
		return false
	}
	/* 检查开关, 上面接口中取的是唯一目标 */
	flag := configParam["netFault"]
	if value, ok := flag.(string); ok && value == "on" {
		return true
	}
	return false
}
