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

// Package config is used for file reading and writing, as well as data processing.
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
)

type checkStatusFunc func(string) bool

// CheckExistDirectoryOrFile 检查文件或目录是否存在(fileOrDir:dir true, file false)
func CheckExistDirectoryOrFile(path string, fileOrDir bool, level string, jobName string) bool {
	/* 确认路径是否存在 */
	var err error
	/* 不存在重试 */
	flag := false
	var checkFunc checkStatusFunc
	if level == "cluster" {
		checkFunc = jobdetectionmanager.GetDetectionLoopStatusClusterLevel
	} else if level == "node" {
		checkFunc = jobdetectionmanager.GetDetectionLoopStatusNodeLevel
	}
	/* should check loop status */
	var fileInfo os.FileInfo
	for i := 0; i < fileNotExistRetryNums && checkFunc(jobName); i++ {
		fileInfo, err = os.Lstat(path)
		if err == nil {
			flag = true
			break
		}
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%v (retry)", err)
		time.Sleep(time.Duration(1) * time.Second)
	}
	if !flag {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s not exist!(return)", path)
		return false
	}
	/* 文件权限检查 */
	if !CheckFileOrDirectoryReadMode(path) || CheckFileOrDirectoryIsSoftLink(path) {
		return false
	}

	if fileOrDir {
		if !fileInfo.IsDir() {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s is not a directory", path)
			return false
		}
	} else {
		if fileInfo.IsDir() {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s is a directory", path)
			return false
		}
	}
	return true
}

// GetLocalIP 获取本地IP地址
func GetLocalIP() (string, error) {
	// 获取环境变量 XDL_IP
	xdlIp := os.Getenv(xdlIpField)
	// 如果环境变量存在且格式正确，则直接返回
	if xdlIp != "" {
		if checkIp := net.ParseIP(xdlIp); checkIp != nil && checkIp.To4() != nil {
			return xdlIp, nil
		}
	}
	// 如果没有环境变量或者格式不正确，输出警告并调用 GetLocalIP 获取本地 IP
	hwlog.RunLog.Warnf("[SLOWNODE ALGO]environment variable isn't set or isn't a valid IPv4 address:%v", xdlIpField)
	// 获取本地所有网络接口的地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 遍历所有地址，查找 IPv4 地址
	for _, addr := range addrs {
		// 检查是否为 IP 地址，并且是否为 IPv4 地址
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			// 返回第一个找到的非 loop back 的 IPv4 地址
			return ipNet.IP.String(), nil
		}
	}
	// 如果没有找到有效的 IPv4 地址，返回错误
	return "", fmt.Errorf("[SLOWNODE ALGO]no valid IP address found")
}

func checkExistCpOrEp(rankJson map[string]any, jobPath string) bool {
	for _, group := range rankJson {
		groupDetail, ok := group.(map[string]any)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]error parallel formation!")
			return true
		}
		groupName, exist := groupDetail[parallelGroupName]
		if !exist {
			hwlog.RunLog.Error("[SLOWNODE ALGO]error parallel formation(without group_name)!")
			return true
		}
		parallelName, ok := groupName.(string)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]error parallel formation(group_name not string)!")
			return true
		}
		if parallelName != parallelCpField && parallelName != parallelEpField {
			continue
		}
		/* 获取global_ranks */
		globalRanks, exist := groupDetail[parallelGlobalRanks]
		if !exist {
			hwlog.RunLog.Error("[SLOWNODE ALGO]error parallel formation(without global_ranks)!")
			return true
		}
		ranksArray, ok := globalRanks.([]any)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]error parallel formation(global_ranks)!")
			return true
		}
		if len(ranksArray) > 1 {
			hwlog.RunLog.Errorf("[SLOWNODE ALGO]%s open cp or ep parallel", jobPath)
			return true
		}
	}
	return false
}

/* 检查节点侧任务数据文件是否开启了CP或EP */
func checkNodeLevelJobOpenEpOrCp(conf AlgoInputConfig) bool {
	jobpath := filepath.Join(conf.FilePath, conf.JobName)
	/* 节点侧仅需要检测一个rank即可(获取当前目录下所有目录，未使用正则) */
	curJobRanksPath := make([]string, 0)
	regRank := regexp.MustCompile(TargetRankDir)
	/* 文件或目录权限、软链接检查 */
	if CheckFileOrDirectoryIsSoftLink(jobpath) || !CheckFileOrDirectoryReadMode(jobpath) {
		return true
	}
	/* 遍历当前job所有rankI */
	dir, err := os.Open(jobpath)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return true
	}
	defer dir.Close()
	entrys, err := dir.Readdirnames(-1)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return true
	}
	for _, entry := range entrys {
		if !regRank.MatchString(entry) {
			continue
		}
		curJobRanksPath = append(curJobRanksPath, filepath.Join(jobpath, entry, rankTopofileName))
	}
	if len(curJobRanksPath) == 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]No job ranks found")
		return true
	}
	data, err := utils.LoadFile(curJobRanksPath[0])
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return true
	}
	var rankJson map[string]any
	err = json.Unmarshal(data, &rankJson)
	if err != nil {
		hwlog.RunLog.Errorf("[SLONODE ALGO]%v", err)
		return true
	}
	return checkExistCpOrEp(rankJson, curJobRanksPath[0])
}

// CheckCurJobOpenCpOrEp 检测任务是否开启了CP或EP
func CheckCurJobOpenCpOrEp(conf AlgoInputConfig, level string) bool {
	filePath := filepath.Join(conf.FilePath, conf.JobName)
	if !CheckExistDirectoryOrFile(filePath, true, level, conf.JobName) {
		/* 等待过后路径仍然未存在，返回true不进行检测 */
		return true
	}
	/* 若存在tp并行域则一定不存在cp或ep */
	return checkNodeLevelJobOpenEpOrCp(conf)
}

// TransferFloatArrayToInt 将json数字字符数组转换为int数组
func TransferFloatArrayToInt(npuIds []any) []int {
	if len(npuIds) == 0 {
		return nil
	}
	npus := make([]int, len(npuIds))
	for i, num := range npuIds {
		npuId, ok := num.(float64)
		if !ok {
			hwlog.RunLog.Error("[SLOWNODE ALGO]Transfer npu id failed!")
			return nil
		}
		npus[i] = int(npuId)
	}
	return npus
}

// LoopDetectionIntervalCheckSwitch 循环检查sleep
func LoopDetectionIntervalCheckSwitch(detectionUsed int64, detectionInterval int,
	jobName string, level string) {
	if detectionUsed >= int64(detectionInterval) {
		return
	}
	sleepTime := int(int64(detectionInterval) - detectionUsed)
	for i := 0; i < sleepTime; i++ {
		if level == "cluster" && !jobdetectionmanager.GetDetectionLoopStatusClusterLevel(jobName) {
			break
		} else if level == "node" && !jobdetectionmanager.GetDetectionLoopStatusNodeLevel(jobName) {
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

// CheckFileOrDirectoryReadMode 检查文件或目录权限是否可读
func CheckFileOrDirectoryReadMode(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return false
	}
	if fileInfo.Mode()&readMode != 0 {
		return true
	}
	hwlog.RunLog.Errorf("[SLOWNODE ALGO]file or directory can't be read:%s", path)
	return false
}

// CheckFileOrDirectoryIsSoftLink 检查文件或目录是否是软链接
func CheckFileOrDirectoryIsSoftLink(path string) bool {
	linkInfo, err := os.Lstat(path)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]%v", err)
		return false
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 {
		hwlog.RunLog.Warnf("[SLOWNODE ALGO]%s is a soft symlink", path)
		return true
	}
	return false
}
