// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node funcs about node
package node

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/util"
)

const (
	serverIndexKey   = "serverIndex"
	serverTypeKey    = "serverType"
	baseDevInfoAnno  = "baseDeviceInfos"
	superPodIDKey    = "superPodID"
	maxNodeDeviceNum = 128
	formatBase       = 10
)

var cache = nodeCache{}

func init() {
	cache = nodeCache{
		nodeInfoCache:      make(map[string]nodeInfo),
		nodeSNAndNameCache: make(map[string]string),
		mutex:              sync.RWMutex{},
	}
}

type nodeCache struct {
	nodeInfoCache      map[string]nodeInfo // key: node name
	nodeSNAndNameCache map[string]string   // key: node sn; value: node name
	mutex              sync.RWMutex        // lock to ensure synchronous updates of caches
}

type nodeInfo struct {
	deviceType   string
	spIndex      string
	nodeName     string
	nodeSN       string
	nodeIP       string
	superPodID   string
	baseDevInfos map[string]*api.NpuBaseInfo // key: node name, value: baseDevInfos
	nodeDevice   *api.NodeDevice
}

// SaveNodeToCache save node info to cache
func SaveNodeToCache(node *v1.Node) {
	if node == nil {
		hwlog.RunLog.Error("node is nil")
		return
	}
	nodeName := node.Name
	nodeSN := getNodeSN(node)
	nodeIP := getNodeIP(node)
	superPodID := getSuerPodID(node)
	baseDevInfos := getBaseDevInfos(node)
	devType := getDeviceType(node)
	spIndex := getServerID(node)
	nodeDeviceInfo := getNodeDevice(baseDevInfos, nodeName, devType, spIndex)

	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.nodeSNAndNameCache[nodeSN] = nodeName
	cache.nodeInfoCache[nodeName] = nodeInfo{
		nodeName:     nodeName,
		nodeSN:       nodeSN,
		nodeIP:       nodeIP,
		superPodID:   superPodID,
		baseDevInfos: baseDevInfos,
		nodeDevice:   nodeDeviceInfo,
		deviceType:   devType,
		spIndex:      spIndex,
	}
}

// DeleteNodeFromCache delete node info from cache
func DeleteNodeFromCache(node *v1.Node) {
	if node == nil {
		hwlog.RunLog.Error("node is nil")
		return
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	delete(cache.nodeSNAndNameCache, getNodeSN(node))
	delete(cache.nodeInfoCache, node.Name)
}

func getServerID(node *v1.Node) string {
	serverID, hasServerIdKey := node.Annotations[serverIndexKey]
	serverID = strings.Trim(serverID, " ")
	if !hasServerIdKey || len(serverID) == 0 {
		hwlog.RunLog.Debugf("empty server id, nodeName=%s", node.Name)
		return ""
	}
	return serverID
}

func getNodeSN(node *v1.Node) string {
	return node.Annotations[api.NodeSNAnnotation]
}

func getSuerPodID(node *v1.Node) string {
	superPodID, ok := node.Annotations[superPodIDKey]
	superPodID = strings.Trim(superPodID, " ")
	if !ok || len(superPodID) == 0 {
		hwlog.RunLog.Debugf("empty super pod id, nodeName=%s", node.Name)
		return ""
	}
	return superPodID
}

func getNodeIP(node *v1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}

func getBaseDevInfos(node *v1.Node) map[string]*api.NpuBaseInfo {
	baseDeviceMap := make(map[string]*api.NpuBaseInfo)
	deviceStr, ok := node.Annotations[baseDevInfoAnno]
	if !ok || len(deviceStr) == 0 {
		hwlog.RunLog.Debugf("empty device info, nodeName=%s", node.Name)
		return nil
	}
	if err := json.Unmarshal([]byte(deviceStr), &baseDeviceMap); err != nil {
		hwlog.RunLog.Errorf("unmarshal device info error, err=%v, nodeName=%s",
			err, node.Name)
		return nil
	}
	if len(baseDeviceMap) == 0 || len(baseDeviceMap) > maxNodeDeviceNum {
		hwlog.RunLog.Errorf("illegal device length, deviceLen=%d, nodeName=%s",
			len(baseDeviceMap), node.Name)
		return nil
	}
	return baseDeviceMap
}

func getNodeDevice(baseDevInfos map[string]*api.NpuBaseInfo, nodeName, devType, serverIndex string) *api.NodeDevice {
	if baseDevInfos == nil {
		return nil
	}
	nodeDevice := &api.NodeDevice{
		NodeName:   nodeName,
		ServerID:   serverIndex,
		ServerType: devType,
		DeviceMap:  make(map[string]string, len(baseDevInfos)),
	}
	for device, info := range baseDevInfos {
		physicID := strings.TrimPrefix(device, api.Ascend910MinuxPrefix)
		_, err := strconv.Atoi(physicID)
		if err != nil {
			hwlog.RunLog.Warnf("illegal device name, deviceName=%s, nodeName=%s",
				device, nodeName)
			return nil
		}
		superDeviceID := strconv.FormatUint(uint64(info.SuperDeviceID), formatBase)
		nodeDevice.DeviceMap[physicID] = superDeviceID
	}
	return nodeDevice
}

// GetNodeNameBySN get node name by sn
func GetNodeNameBySN(nodeSN string) (string, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	name, ok := cache.nodeSNAndNameCache[nodeSN]
	if !ok {
		return "", false
	}
	return name, true
}

// GetNodeSNByName get node sn by name
func GetNodeSNByName(nodeName string) string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	info, ok := cache.nodeInfoCache[nodeName]
	if !ok {
		hwlog.RunLog.Warnf("node[%s] does not exist in cache", nodeName)
		return ""
	}
	return info.nodeSN
}

// GetNodeIpByName get node ip by name
func GetNodeIpByName(nodeName string) string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	info, ok := cache.nodeInfoCache[nodeName]
	if !ok {
		hwlog.RunLog.Warnf("node[%s] does not exist in cache", nodeName)
		return ""
	}
	return info.nodeIP
}

// GetNodeIPAndSNMap get node ip and sn map
func GetNodeIPAndSNMap() map[string]string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	ipSNMap := make(map[string]string)
	for _, info := range cache.nodeInfoCache {
		ipSNMap[info.nodeIP] = info.nodeSN
	}
	return ipSNMap
}

// GetNodeDeviceAndSuperPodID get node device and super pod id
func GetNodeDeviceAndSuperPodID(node *v1.Node) (*api.NodeDevice, string) {
	if node == nil {
		hwlog.RunLog.Error("node is nil")
		return nil, ""
	}
	nodeName := node.Name
	cache.mutex.RLock()
	info, ok := cache.nodeInfoCache[nodeName]
	cache.mutex.RUnlock()
	if !ok {
		hwlog.RunLog.Warnf("node[%s] does not exist in cache, get from node info", nodeName)
		// If deletion operation, node info will not exist in the cache, but needs to be notified to pingmesh.
		// so we need to obtain it again
		spId := getSuerPodID(node)
		baseDevInfos := getBaseDevInfos(node)
		deviceType := getDeviceType(node)
		spIndex := getServerID(node)
		nodeDevice := getNodeDevice(baseDevInfos, nodeName, deviceType, spIndex)
		return nodeDevice, spId
	}

	// internal data protection
	oldDev := info.nodeDevice
	if oldDev == nil {
		return nil, info.superPodID
	}
	var newDev *api.NodeDevice
	if err := util.DeepCopy(&newDev, oldDev); err != nil {
		hwlog.RunLog.Errorf("deep copy node device info failed, error: %v", err)
		return nil, info.superPodID
	}
	return newDev, info.superPodID
}

func getDeviceType(node *v1.Node) string {
	devType, hasVersionKey := node.Annotations[serverTypeKey]
	devType = strings.Trim(devType, " ")
	if !hasVersionKey || len(devType) == 0 {
		hwlog.RunLog.Debugf("empty version, nodeName=%s", node.Name)
		return ""
	}
	return devType
}
