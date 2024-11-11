package fault

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
)

// nodeFaultProcessCenter
type NodeFaultProcessCenter struct {
	BaseFaultCenter
	infos map[string]*constant.NodeInfo
	mutex sync.RWMutex
}

func NewNodeFaultProcessCenter() *NodeFaultProcessCenter {
	return &NodeFaultProcessCenter{
		infos:           make(map[string]*constant.NodeInfo),
		mutex:           sync.RWMutex{},
		BaseFaultCenter: newBaseFaultCenter(),
	}
}

func (nodeCenter *NodeFaultProcessCenter) GetNodeInfos() map[string]*constant.NodeInfo {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	return node.DeepCopyInfos(nodeCenter.infos)
}

func (nodeCenter *NodeFaultProcessCenter) SetNodeInfos(infos map[string]*constant.NodeInfo) {
	nodeCenter.mutex.RLock()
	defer nodeCenter.mutex.RUnlock()
	nodeCenter.infos = node.DeepCopyInfos(infos)
}

func (nodeCenter *NodeFaultProcessCenter) InformerAddCallback(oldInfo, newInfo *constant.NodeInfo) bool {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	length := len(nodeCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
		return false
	}
	oldInfo, found := nodeCenter.infos[newInfo.CmName]
	nodeCenter.infos[newInfo.CmName] = newInfo
	return found && node.BusinessDataIsNotEqual(oldInfo, newInfo)
}

func (nodeCenter *NodeFaultProcessCenter) InformerDelCallback(oldInfo, newInfo *constant.NodeInfo) bool {
	nodeCenter.mutex.Lock()
	defer nodeCenter.mutex.Unlock()
	oldInfo, found := nodeCenter.infos[newInfo.CmName]
	delete(nodeCenter.infos, newInfo.CmName)
	nodeCenter.infos[newInfo.CmName] = newInfo
	return found
}
