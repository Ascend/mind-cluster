package fault

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/switchinfo"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"sync"
)

type SwitchFaultProcessCenter struct {
	BaseFaultCenter
	infos map[string]*constant.SwitchInfo
	mutex sync.RWMutex
}

func NewSwitchFaultProcessCenter() *SwitchFaultProcessCenter {
	return &SwitchFaultProcessCenter{
		infos:           make(map[string]*constant.SwitchInfo),
		mutex:           sync.RWMutex{},
		BaseFaultCenter: newBaseFaultCenter(),
	}
}

func (switchCenter *SwitchFaultProcessCenter) GetSwitchInfos() map[string]*constant.SwitchInfo {
	switchCenter.mutex.RLock()
	defer switchCenter.mutex.RUnlock()
	return switchinfo.DeepCopyInfos(switchCenter.infos)
}

func (switchCenter *SwitchFaultProcessCenter) setSwitchInfos(infos map[string]*constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	switchCenter.infos = switchinfo.DeepCopyInfos(infos)
}

func (switchCenter *SwitchFaultProcessCenter) InformerAddCallback(oldInfo, newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	length := len(switchCenter.infos)
	if length > constant.MaxSupportNodeNum {
		hwlog.RunLog.Errorf("SwitchInfo length=%d > %d, SwitchInfo cm name=%s save failed",
			length, constant.MaxSupportNodeNum, newInfo.CmName)
	}
	switchCenter.infos[newInfo.CmName] = newInfo
}

func (switchCenter *SwitchFaultProcessCenter) InformerDelCallback(newInfo *constant.SwitchInfo) {
	switchCenter.mutex.Lock()
	defer switchCenter.mutex.Unlock()
	delete(switchCenter.infos, newInfo.CmName)
}
