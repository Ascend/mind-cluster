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

/*
Package utils is used for file reading and writing, as well as data processing.
*/
package utils

import (
	"encoding/json"
	"os"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

// getNodeRanksFromRanktable 解析ranktable文件，构架ip2Ranks字典，根据本节点IP地址读取Ranks（IP从环境变量中读取）
func getNodeRanksFromRanktable(rankTablePath string) []int {
	// 判断ranktable文件是否存在
	_, err := os.Stat(rankTablePath)
	if err != nil {
		hwlog.RunLog.Error("ranktable is not exist")
		return []int{}
	}

	// 读取文件内容
	data, err := fileutils.ReadLimitBytes(rankTablePath, constants.Size10M)
	if err != nil {
		hwlog.RunLog.Errorf("read ranktable error: %v", err)
		return []int{}
	}

	// 定义rankTable为嵌套字典
	var rankTable map[string]any

	// 解析JSON数据到rankTable
	if err := json.Unmarshal(data, &rankTable); err != nil {
		hwlog.RunLog.Errorf("ranktable parse error: %v", err)
		return []int{}
	}

	var serverList []any
	serverListx, ok := rankTable[serverListField]
	if !ok {
		// Handle the case where the key "server_list" does not exist
		hwlog.RunLog.Error("Error: 'server_list' key does not exist in the rankTable.")
		return []int{}
	}
	serverList, ok = serverListx.([]any)
	if !ok {
		hwlog.RunLog.Error("server_list parse error")
		return []int{}
	}

	// Construct the ip2Ranks mapping
	ip2Ranks := buildIp2Ranks(serverList)

	// 获取环境变量 XDL_IP
	xdlIp := os.Getenv(xdlIpField)

	// 检查 XDL_IP 是否在 ip2Ranks 中
	if ranksForServer, exists := ip2Ranks[xdlIp]; exists {
		// 如果 XDL_IP 存在，返回对应的 ranks
		return ranksForServer
	} else {
		// 如果 XDL_IP 不存在
		hwlog.RunLog.Info("XDL_IP not in ip2Ranks")
		return []int{}
	}

}

// buildIp2Ranks构建IP地址和Ranks列表之间的映射
func buildIp2Ranks(serverList []any) map[string][]int {
	// An example of serverList: [{"server_id": "string", "device": [{"rank_d":"123"}]}]
	ip2Ranks := make(map[string][]int, len(serverList))

	for _, server := range serverList {
		serverData, ok := server.(map[string]any)
		if !ok {
			hwlog.RunLog.Error("server data parse error")
			continue
		}

		serverID, ok := serverData[serverIdField].(string)
		if !ok {
			hwlog.RunLog.Error("server_id parse error")
			continue
		}

		deviceList, ok := serverData[deviceField].([]any)
		if !ok {
			hwlog.RunLog.Error("device list parse error")
			continue
		}

		// 提取 rank_id
		var rankIDs = []int{}
		for _, device := range deviceList {
			deviceData, ok := device.(map[string]any)
			if !ok {
				hwlog.RunLog.Error("device data parse error")
				continue
			}

			// 获取 rank_id 并转换为 int
			rankID, ok := deviceData[rankIdField].(string)
			if !ok {
				hwlog.RunLog.Error("rank_id parse error")
				continue
			}

			// 将 rank_id 转换为整数并添加到 rankIDs 列表中
			rankIDs = append(rankIDs, stringToInt(rankID))
		}

		// 将结果存储到字典中
		ip2Ranks[serverID] = rankIDs
	}
	return ip2Ranks
}

// stringToInt将字符串转换为整数
func stringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		// 如果转换失败，打印错误并返回 0
		hwlog.RunLog.Errorf("error stringToInt : %v", err)
		return 0
	}
	return i
}
