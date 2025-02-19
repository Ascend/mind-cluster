// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics statistic funcs about node
package statistics

// key: node sn; value: node name
// the map is not locked, because its simple and ability to modify
var nodeSNAndNameCache map[string]string

func init() {
	nodeSNAndNameCache = make(map[string]string)
}

// GetNodeNameBySN get node name by sn
func GetNodeNameBySN(nodeSN string) (string, bool) {
	name, ok := nodeSNAndNameCache[nodeSN]
	if !ok {
		return "", false
	}
	return name, true
}

// GetNodeSNAndNameCache get node sn and name cache
func GetNodeSNAndNameCache() map[string]string {
	return nodeSNAndNameCache
}
