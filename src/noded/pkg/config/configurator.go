/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package config for the fault config
package config

import (
	"context"
	"encoding/json"
	"fmt"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/common-utils/utils"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

// FaultConfigurator manage dynamically configuration information
type FaultConfigurator struct {
	client              *kubeclient.ClientK8s
	configManager       common.ConfigManager
	configCache         *common.FaultConfig
	nextConfigProcessor common.ConfigProcessor
	updateChan          chan struct{}
	stopChan            chan struct{}
	initFromCMFlag      bool
}

// NewFaultConfigurator create a configurator
func NewFaultConfigurator(client *kubeclient.ClientK8s) *FaultConfigurator {
	return &FaultConfigurator{
		client:        client,
		configManager: common.NewConfigManager(),
		configCache:   &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{}},
		updateChan:    make(chan struct{}, 1),
		stopChan:      make(chan struct{}, 1),
	}
}

// Run start working loop
func (c *FaultConfigurator) Run(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Error("ctx stop channel is closed")
			}
			hwlog.RunLog.Info("receive ctx stop signal, config manager shut down")
			return
		case _, ok := <-c.stopChan:
			if !ok {
				hwlog.RunLog.Error("stop channel is closed")
			}
			hwlog.RunLog.Info("receive stop signal, config manager shut down")
			return
		case _, ok := <-c.updateChan:
			if !ok {
				hwlog.RunLog.Error("update config channel is closed")
				return
			}
			if err := c.UpdateConfig(c.configCache); err != nil {
				hwlog.RunLog.Warn("update fault config filed, original fault config will be maintained")
			}
		}
	}
}

// Stop terminate working loop
func (c *FaultConfigurator) Stop() {
	c.stopChan <- struct{}{}
}

// UpdateConfig update config and send message to next config processor
func (c *FaultConfigurator) UpdateConfig(faultConfig *common.FaultConfig) error {
	c.configManager.SetFaultConfig(faultConfig)
	if err := c.nextConfigProcessor.UpdateConfig(c.configManager.GetFaultConfig()); err != nil {
		return err
	}
	hwlog.RunLog.Info("update fault config success")
	return nil
}

// SetNextConfigProcessor set next config processor
func (c *FaultConfigurator) SetNextConfigProcessor(configProcessor common.ConfigProcessor) {
	c.nextConfigProcessor = configProcessor
}

// Init initialize configuration information and start informer for fault config map
func (c *FaultConfigurator) Init() error {
	if err := c.initFaultConfigFromCM(); err != nil {
		hwlog.RunLog.Info("init fault config from config map failed, start load local json file")
		if err := c.loadFaultConfigFromFile(); err != nil {
			hwlog.RunLog.Errorf("load fault config from local file failed, err is %v", err)
			return err
		}
		hwlog.RunLog.Info("init config from local json file success")
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(c.client.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.Set{
				common.MetaDataNameSpace: common.FaultConfigCMNameSpace,
				common.MetaDataName:      common.FaultConfigCMName,
			}.String()
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()
	cmInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.AddConfigCM,
			UpdateFunc: c.UpdateConfigCM,
			DeleteFunc: c.DeleteConfigCM,
		})
	informerFactory.Start(wait.NeverStop)
	return nil
}

// AddConfigCM update config when add fault config map
func (c *FaultConfigurator) AddConfigCM(obj interface{}) {
	if c == nil {
		hwlog.RunLog.Error("config manager is nil when add config cm")
		return
	}
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Error("failed convert cm when add cm")
		return
	}
	// prevent update fault config repeatedly
	if c.initFromCMFlag {
		c.initFromCMFlag = false
		return
	}
	if err := c.UpdateConfigCache(cm); err != nil {
		hwlog.RunLog.Warn("update config failed when add cm, original fault config will be maintained")
		return
	}
	c.updateChan <- struct{}{}
}

// UpdateConfigCM update config when update fault config map
func (c *FaultConfigurator) UpdateConfigCM(old, new interface{}) {
	if c == nil {
		hwlog.RunLog.Error("config manager is nil when update config cm")
		return
	}
	cm, ok := new.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Error("failed convert cm when update cm")
		return
	}
	if err := c.UpdateConfigCache(cm); err != nil {
		hwlog.RunLog.Warn("update config failed when update cm, original fault config will be maintained")
		return
	}
	c.updateChan <- struct{}{}
}

// DeleteConfigCM warn when update fault config map
func (c *FaultConfigurator) DeleteConfigCM(obj interface{}) {
	if c == nil {
		hwlog.RunLog.Error("config manager is nil when delete config cm")
		return
	}
	hwlog.RunLog.Warn("fault config cm is deleted")
}

// getFaultConfigFromCM get fault config from config map
func (c *FaultConfigurator) getFaultConfigFromCM(cm *v1.ConfigMap) (*common.FaultConfig, error) {
	faultConfig, ok := cm.Data[common.FaultConfigKey]
	if !ok {
		hwlog.RunLog.Errorf("can not find the key '%s' in cm, failed to get fault config from cm",
			common.FaultConfigKey)
		return nil, fmt.Errorf("can not find the key '%s' in cm", common.FaultConfigKey)
	}
	var config common.FaultConfig
	if err := json.Unmarshal([]byte(faultConfig), &config); err != nil {
		hwlog.RunLog.Errorf("unmarshal fault config failed, err is %v", err)
		return nil, fmt.Errorf("unmarshal fault config failed: %v", err)
	}
	if err := c.filterAndCheckFaultCodes(&config); err != nil {
		hwlog.RunLog.Error("check fault codes failed when get fault config cm")
		return nil, err
	}
	return &config, nil
}

// UpdateConfigCache update config cache
func (c *FaultConfigurator) UpdateConfigCache(cm *v1.ConfigMap) error {
	newConfigCache, err := c.getFaultConfigFromCM(cm)
	if err != nil {
		return err
	}
	c.configCache = newConfigCache
	return nil
}

// initFaultConfigFromCM init fault config from config map
func (c *FaultConfigurator) initFaultConfigFromCM() error {
	c.initFromCMFlag = true
	configCM, err := c.client.GetConfigMap(common.FaultConfigCMName, common.FaultConfigCMNameSpace)
	if err != nil {
		hwlog.RunLog.Info("get config cm failed when init, may be not create, load from local json file")
		return err
	}
	if err := c.UpdateConfigCache(configCM); err != nil {
		hwlog.RunLog.Errorf("update config cache failed, please check config map content, err is %v", err)
		return err
	}
	if err := c.UpdateConfig(c.configCache); err != nil {
		hwlog.RunLog.Error("update fault config failed when init fault config from config map")
		return err
	}
	hwlog.RunLog.Info("init fault config from config map success")
	return nil
}

// loadFaultConfigFromFile load fault config from json file
func (c *FaultConfigurator) loadFaultConfigFromFile() error {
	faultConfigBytes, err := utils.LoadFile(common.FaultConfigFilePath)
	if err != nil {
		return fmt.Errorf("load local fault config json file failed: %v", err)
	}
	var fileConfig common.FaultConfig
	if err := json.Unmarshal(faultConfigBytes, &fileConfig); err != nil {
		return fmt.Errorf("unmarshal fault config byte failed: %v", err)
	}
	if err := c.filterAndCheckFaultCodes(&fileConfig); err != nil {
		hwlog.RunLog.Error("check fault codes failed when load local json file")
		return err
	}
	if err := c.UpdateConfig(&fileConfig); err != nil {
		hwlog.RunLog.Errorf("update fault config failed when load local file: %v", err)
		return err
	}
	return nil
}

// filterAndCheckFaultCodes filter conflict fault code at same level and check  whether fault code str is illegal
func (c *FaultConfigurator) filterAndCheckFaultCodes(faultConfig *common.FaultConfig) error {
	common.ToUpperFaultCodesStr(faultConfig.FaultTypeCode)
	common.FilterDuplicateFaultCodes(faultConfig.FaultTypeCode)
	if err := common.CheckFaultCodes(faultConfig.FaultTypeCode.NotHandleFaultCodes); err != nil {
		hwlog.RunLog.Errorf("check not handle fault code, %s", err.Error())
		return err
	}
	if err := common.CheckFaultCodes(faultConfig.FaultTypeCode.PreSeparateFaultCodes); err != nil {
		hwlog.RunLog.Errorf("check pre separate fault code, %s", err.Error())
		return err
	}
	if err := common.CheckFaultCodes(faultConfig.FaultTypeCode.SeparateFaultCodes); err != nil {
		hwlog.RunLog.Errorf("check separate fault code, %s", err.Error())
		return err
	}
	return nil
}
