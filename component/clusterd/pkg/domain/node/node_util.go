// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package node a series of node function
package node

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeNodeSize = 2000

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
	return &node, nil
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

// DeepCopyInfos deep copy NodeInfos
func DeepCopyInfos(infos map[string]*constant.NodeInfo) map[string]*constant.NodeInfo {
	res := make(map[string]*constant.NodeInfo)
	for key, val := range infos {
		res[key] = DeepCopy(val)
	}
	return res
}

// GetSafeData get data every 2000 NodeInfo
func GetSafeData(nodeInfos map[string]*constant.NodeInfo) []string {
	if len(nodeInfos) == 0 {
		return []string{}
	}
	if len(nodeInfos) <= safeNodeSize {
		return []string{util.ObjToString(nodeInfos)}
	}
	nodeSlice := make([]string, 0, len(nodeInfos)/safeNodeSize+1)
	childNodeInfos := make(map[string]*constant.NodeInfo, safeNodeSize)
	for cmName, nodeInfo := range nodeInfos {
		childNodeInfos[cmName] = nodeInfo
		if len(childNodeInfos)%safeNodeSize == 0 {
			nodeSlice = append(nodeSlice, util.ObjToString(childNodeInfos))
			childNodeInfos = make(map[string]*constant.NodeInfo, safeNodeSize)
		}
	}
	if len(childNodeInfos) != 0 {
		nodeSlice = append(nodeSlice, util.ObjToString(childNodeInfos))
	}
	return nodeSlice
}
