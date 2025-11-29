// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package slownet for net fault detect common
package slownet

import (
	"fmt"
	"os"
	"path/filepath"

	"ascend-common/common-utils/utils"
)

const (
	roceSubPath       = "super-pod-roce"
	pingListRangeFile = "ping_list_range.json"
)

// GetRoCEPingListFilePath get ping list task info file for RoCE ping
func GetRoCEPingListFilePath(superPodId, serverIndex string) (string, error) {
	rootPath, err := GetRasNetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(rootPath, netFaultSubPath, roceSubPath, fmt.Sprintf("ping_list_%s_%s.json",
		superPodId, serverIndex)), nil
}

// GetPingListRangePath get ping list range file for roce ping
func GetPingListRangePath() (string, error) {
	rootPath, err := GetRasNetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(rootPath, netFaultSubPath, roceSubPath, pingListRangeFile), nil
}

// CheckIsExistAndValid check file is exist and modify time is new enough
func CheckIsExistAndValid(filePath string) error {
	if !utils.IsLexist(filePath) {
		return fmt.Errorf("%s file is not exist", filePath)
	}
	fileInfo, errStat := os.Stat(filePath)
	if errStat != nil {
		return fmt.Errorf("get %s file status info failed, err: %v", filePath, errStat)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("%s file content is empty", filePath)
	}
	return nil
}

// GetRackTopologyFilePath get the file path of rack topology in ras feature
func GetRackTopologyFilePath(superPodId, rackId, serverIndex int32) (string, error) {
	rootPath, err := GetRasNetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(rootPath, netFaultSubPath, fmt.Sprintf("super-pod-%d", superPodId),
		fmt.Sprintf("rack-%d", rackId), fmt.Sprintf("topo_%d.json", serverIndex)), nil
}
