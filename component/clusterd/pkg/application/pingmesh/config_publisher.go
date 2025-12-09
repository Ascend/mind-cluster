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

// Package pingmesh for a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/fdapi"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/superpod"
)

// ConfigPingMeshCmManager ras feature net fault detect configmap manager info
type ConfigPingMeshCmManager struct {
	sync.RWMutex
	configCMInfo constant.ConfigPingMesh
	cacheStatus  *constant.CacheStatus
}

// ConfigPingMeshInst the config info of pingmesh-config
var ConfigPingMeshInst *ConfigPingMeshCmManager

func init() {
	ConfigPingMeshInst = &ConfigPingMeshCmManager{
		RWMutex:      sync.RWMutex{},
		configCMInfo: constant.ConfigPingMesh{},
		cacheStatus: &constant.CacheStatus{
			Inited: false,
		},
	}
}

// UpdateConfigData to update new pingmesh config
func (cf *ConfigPingMeshCmManager) UpdateConfigData(newInfo constant.ConfigPingMesh) {
	cf.Lock()
	cf.configCMInfo = newInfo
	cf.Unlock()
	if cf.cacheStatus.Inited {
		cf.updateConfigFileWhenCmUpdated()
	}
}

// UpdateConfig to update new pingmehs config
func (cf *ConfigPingMeshCmManager) UpdateConfig(newInfo constant.ConfigPingMesh) error {
	if newInfo == nil {
		hwlog.RunLog.Warnf("newInfo is nil, skip updating")
		return fmt.Errorf("parameters of updating config is invalid")
	}
	changed := false
	cf.Lock()
	oldConfig := cf.configCMInfo
	cf.configCMInfo = newInfo
	changed = cf.checkConfChanged(oldConfig, newInfo)
	cf.Unlock()
	if cf.cacheStatus.Inited {
		cf.updateConfigFileWhenCmUpdated()
	}
	if changed {
		cf.startOrReloadController()
	}
	return nil
}

func (cf *ConfigPingMeshCmManager) checkConfChanged(oldInfo, newInfo constant.ConfigPingMesh) bool {
	if !rasNetDetectInst.CheckIsOn() {
		hwlog.RunLog.Info("ping mesh config detect is inactive, no need to reload or start controller")
		return false
	}
	for _, deviceInfo := range superpod.ListClusterDevice() {
		superPodID := deviceInfo.SuperPodID
		newConf, errNew := getConfigItemBySuperPodId(newInfo, superPodID)
		if errNew != nil {
			hwlog.RunLog.Errorf("get new conf by SuperPodId failed, err:%v, skip this loop", errNew)
			continue
		}
		oldConf, errOld := getConfigItemBySuperPodId(oldInfo, superPodID)
		if errOld != nil {
			hwlog.RunLog.Warnf("get old conf by SuperPodId failed, err:%v", errOld)
		}
		// check whether the old config is equal to new config
		// if true update config file
		if configItemEqual(oldConf, newConf) {
			hwlog.RunLog.Infof("the new config of superpod ID<%s> is not changed", superPodID)
			continue
		}
		return true
	}
	return handleRoceConfigChange(oldInfo, newInfo)
}

func handleRoceConfigChange(oldInfo, newInfo constant.ConfigPingMesh) bool {
	newConf, newExist := newInfo[constant.RasRoceKey]
	if !newExist {
		hwlog.RunLog.Warnf("roce key is not exist in the new config, do nothing")
		return false
	}
	oldConf, oldExist := oldInfo[constant.RasRoceKey]
	if !oldExist {
		return true
	}
	// check whether the old config is equal to new config
	if configItemEqual(oldConf, newConf) {
		hwlog.RunLog.Infof("the new config of <%s> is not changed", constant.RasRoceKey)
		return false
	}
	hwlog.RunLog.Debugf("the old config of key<%s> is changed", constant.RasRoceKey)
	return true
}

func (cf *ConfigPingMeshCmManager) getRasConfigBySuperPodId(superPodID string) *constant.CathelperConf {
	cf.RLock()
	cfgCM, err := getConfigItemBySuperPodId(cf.configCMInfo, superPodID)
	cf.RUnlock()
	if err != nil {
		hwlog.RunLog.Errorf("get config item by super pod id failed, will use default config: %v, err: %v",
			rasConfig, err)
		return &rasConfig
	}
	cfg := NewCathelperConf()
	const collectPeriodFactor = 10
	cfg.Period = cfgCM.TaskInterval * collectPeriodFactor
	if cfgCM.Activate == constant.RasNetDetectOnStr {
		cfg.NetFault = constant.RasNetDetectOnStr
	}
	return &cfg
}

func (cf *ConfigPingMeshCmManager) updateConfigFileWhenCmUpdated() {
	if !rasNetDetectInst.CheckIsOn() {
		hwlog.RunLog.Info("ping mesh config detect is inactive, no need to reload or start controller")
		return
	}
	for _, deviceInfo := range superpod.ListClusterDevice() {
		superPodID := deviceInfo.SuperPodID
		cfg := cf.getRasConfigBySuperPodId(superPodID)
		err := saveConfigToFile(superPodID, cfg)
		if err != nil {
			hwlog.RunLog.Errorf("save config to file failed, err=%v, superPodID=%s", err, superPodID)
			return
		}
		hwlog.RunLog.Infof("update config file successfully, superPodID=%s", superPodID)
	}

	handlerSuperPodRoce(superpod.GetAllSuperPodIDWithAcceleratorType())
}

func (cf *ConfigPingMeshCmManager) startOrReloadController() {
	if !rasNetDetectInst.CheckIsOn() {
		hwlog.RunLog.Info("ping mesh config detect is inactive, no need to reload or start controller")
		return
	}
	hwlog.RunLog.Info("config changed and will reload or start controller")
	if !cf.cacheStatus.Inited {
		cf.cacheStatus.Inited = true
		fdapi.StartController()
		return
	}
	fdapi.ReloadController()
}

func getConfigItemBySuperPodId(configInfo constant.ConfigPingMesh,
	superPodID string) (*constant.HccspingMeshItem, error) {
	if configInfo == nil {
		return nil, fmt.Errorf("configInfo is nil")
	}
	config, ok := configInfo[superPodID]
	if ok {
		return config, nil
	}
	config, ok = configInfo[constant.RasGlobalKey]
	if !ok {
		return nil, fmt.Errorf("get config failed")
	}
	return config, nil
}

func configItemEqual(oldConfig, newConfig *constant.HccspingMeshItem) bool {
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}
	return *oldConfig == *newConfig
}
