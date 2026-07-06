// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package node a series of node function
package node

import (
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	// maxCmDataSize is the max data size for a single ConfigMap (~1MB limit, using 800KB for safety margin)
	maxCmDataSize = 800 * 1024
)

// ParseNodeInfoCM get node info from configmap obj
func ParseNodeInfoCM(obj interface{}) (*constant.NodeInfo, error) {
	nodeCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return &constant.NodeInfo{}, fmt.Errorf("not node info configmap")
	}
	nodeInfoCM := constant.NodeInfoCM{}
	data, ok := nodeCm.Data[api.NodeInfoCMDataKey]
	if !ok {
		return &constant.NodeInfo{},
			fmt.Errorf("configmap %s has no key: %s", nodeCm.Name, api.NodeInfoCMDataKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &nodeInfoCM); unmarshalErr != nil {
		return &constant.NodeInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, nodeCm.Name)
	}
	if !util.EqualDataHash(nodeInfoCM.CheckCode, nodeInfoCM.NodeInfo) {
		return &constant.NodeInfo{}, fmt.Errorf("node info configmap %s is not valid", nodeCm.Name)
	}

	var node constant.NodeInfo
	node.NodeStatus = nodeInfoCM.NodeInfo.NodeStatus
	node.FaultDevList = nodeInfoCM.NodeInfo.FaultDevList
	node.CmName = nodeCm.Name
	node.UpdateTime = parseUpdateTime(nodeCm.Data["updateTime"])
	return &node, nil
}

func parseUpdateTime(updateTimeStr string) int64 {
	if updateTimeStr == "" {
		return 0
	}
	parsed, err := time.Parse(time.RFC3339, updateTimeStr)
	if err != nil {
		hwlog.RunLog.Warnf("parse updateTime %s failed: %v", updateTimeStr, err)
		return 0
	}
	return parsed.UnixMilli()
}

// DeepCopy deep copy NodeInfo
func DeepCopy(info *constant.NodeInfo) *constant.NodeInfo {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal node failed , err is %v", err)
		return nil
	}
	newNodeInfo := &constant.NodeInfo{}
	if err := json.Unmarshal(data, newNodeInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal node failed , err is %v", err)
		return nil
	}
	return newNodeInfo
}

// GetSafeData splits nodeInfos into chunks that fit within K8s ConfigMap size limit (~1MB).
// Each chunk is as close to maxCmDataSize (800KB) as possible.
func GetSafeData(nodeInfos map[string]*constant.NodeInfo) []string {
	return util.SplitMapToSafeChunks(nodeInfos, maxCmDataSize,
		func(m map[string]*constant.NodeInfo) string {
			return util.ObjToString(m)
		})
}

// GetData get data from NodeInfo
func GetData(nodeInfos map[string]*constant.NodeInfo) []string {
	if len(nodeInfos) == 0 {
		return []string{}
	}
	return []string{util.ObjToString(nodeInfos)}
}
