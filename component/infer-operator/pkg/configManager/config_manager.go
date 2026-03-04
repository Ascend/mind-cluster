/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package configManager

import (
	"encoding/json"
	"sync"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
	util "infer-operator/pkg/common/client-go"
)

// ConfigManager manages all configuration information
type ConfigManager struct {
	// key: configMap name, value: data key-value of the configMap
	configCache sync.Map
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// Start starts the ConfigManager and registers event callback functions
func (cfgMgr *ConfigManager) Start() {
	util.AddInferOperatorCfgEventFuncs(cfgMgr.faultStrategyHandler)
	hwlog.RunLog.Info("ConfigManager faultStrategyHandler registered")
}

func (cfgMgr *ConfigManager) faultStrategyHandler(oldObj, newObj interface{}, operator string) {
	switch operator {
	case common.AddOperator:
		cfgMgr.handleAdd(newObj)
	case common.UpdateOperator:
		cfgMgr.handleUpdate(oldObj, newObj)
	case common.DeleteOperator:
		cfgMgr.handleDelete(newObj)
	default:
		hwlog.RunLog.Errorf("Unknown operator %s", operator)
	}
}

func (cfgMgr *ConfigManager) handleAdd(obj interface{}) {
	configMap, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}
	innerCache := &sync.Map{}
	for key, value := range configMap.Data {
		innerCache.Store(key, value)
	}

	cfgMgr.configCache.Store(configMap.Name, innerCache)

	hwlog.RunLog.Debugf("ConfigMap added: %s/%s", configMap.Namespace, configMap.Name)
}

func (cfgMgr *ConfigManager) handleUpdate(oldObj, newObj interface{}) {
	newCM, ok := newObj.(*v1.ConfigMap)
	if !ok {
		return
	}
	innerCache, exists := cfgMgr.configCache.Load(newCM.Name)
	if !exists {
		innerCache = &sync.Map{}
		cfgMgr.configCache.Store(newCM.Name, innerCache)
	}

	innerMap, ok := innerCache.(*sync.Map)
	if !ok {
		return
	}
	innerMap.Range(func(key, value interface{}) bool {
		innerMap.Delete(key)
		return true
	})
	for key, value := range newCM.Data {
		innerMap.Store(key, value)
	}
	hwlog.RunLog.Debugf("ConfigMap updated: %s/%s", newCM.Namespace, newCM.Name)
}

func (cfgMgr *ConfigManager) handleDelete(obj interface{}) {
	configMap, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}
	cfgMgr.configCache.Delete(configMap.Name)
	hwlog.RunLog.Debugf("ConfigMap deleted: %s/%s", configMap.Namespace, configMap.Name)
}

// GetConfig returns all configuration information as a JSON string
func (cfgMgr *ConfigManager) GetConfig() string {
	result := make(map[string]map[string]interface{})
	cfgMgr.configCache.Range(func(cmName, innerCache interface{}) bool {
		cmNameStr, ok := cmName.(string)
		if !ok {
			return true
		}
		innerMap, ok := innerCache.(*sync.Map)
		if !ok {
			return true
		}

		cmData := make(map[string]interface{})
		innerMap.Range(func(key, value interface{}) bool {
			if keyStr, ok := key.(string); ok {
				cmData[keyStr] = value
			}
			return true
		})
		result[cmNameStr] = cmData
		return true
	})
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to marshal config to JSON: %v", err)
		return "{}"
	}
	return string(jsonBytes)
}

// GetConfigByCMName returns the configuration information for the specified ConfigMap name as a JSON string
func (cfgMgr *ConfigManager) GetConfigByCMName(cmName string) string {
	innerCache, exists := cfgMgr.configCache.Load(cmName)
	if !exists {
		return "{}"
	}
	innerMap, ok := innerCache.(*sync.Map)
	if !ok {
		return "{}"
	}
	cmData := make(map[string]interface{})
	innerMap.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			cmData[keyStr] = value
		}
		return true
	})
	jsonBytes, err := json.Marshal(cmData)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to marshal CM %s config: %v", cmName, err)
		return "{}"
	}
	return string(jsonBytes)
}

// GetConfigByCMAndKey returns the configuration value for the specified key of the specific ConfigMap
func (cfgMgr *ConfigManager) GetConfigByCMAndKey(cmName, key string) (interface{}, bool) {
	innerCache, exists := cfgMgr.configCache.Load(cmName)
	if !exists {
		return nil, false
	}
	innerMap, ok := innerCache.(*sync.Map)
	if !ok {
		return nil, false
	}
	value, exists := innerMap.Load(key)
	return value, exists
}
