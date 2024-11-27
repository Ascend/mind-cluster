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

// Package cmreporter for fault device info by configmap
package cmreporter

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

// ConfigMapReporter report fault device info by config map
type ConfigMapReporter struct {
	client        *kubeclient.ClientK8s
	nodeInfoCache common.NodeInfoCM
}

// NewConfigMapReporter create a config map reporter
func NewConfigMapReporter(client *kubeclient.ClientK8s) *ConfigMapReporter {
	return &ConfigMapReporter{
		client:        client,
		nodeInfoCache: common.NodeInfoCM{},
	}
}

// Report send fault device info by config map
func (c *ConfigMapReporter) Report(faultDevInfo *common.FaultDevInfo) {
	c.nodeInfoCache = common.NodeInfoCM{
		NodeInfo: *faultDevInfo,
	}
	c.nodeInfoCache.CheckCode = common.MakeDataHash(c.nodeInfoCache.NodeInfo)

	data, err := json.Marshal(c.nodeInfoCache)
	if err != nil {
		hwlog.RunLog.Errorf("marshal node info cache failed, err is %v", err)
		return
	}
	nodeInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.client.NodeInfoName,
			Namespace: common.NodeInfoCMNameSpace,
			Labels:    map[string]string{common.CmConsumer: common.CmConsumerValue},
		},
		Data: map[string]string{
			common.NodeInfoCMDataKey: string(data),
		},
	}
	if err := c.client.CreateOrUpdateConfigMap(nodeInfoCM); err != nil {
		hwlog.RunLog.Errorf("report node fault device info to k8s by configmap failed, err is %v", err)
		return
	}
	return
}

// Init initialize node info config map
func (c *ConfigMapReporter) Init() error {
	nodeInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.client.NodeInfoName,
			Namespace: common.NodeInfoCMNameSpace,
		},
		Data: map[string]string{},
	}
	if err := c.client.CreateOrUpdateConfigMap(nodeInfoCM); err != nil {
		hwlog.RunLog.Errorf("init node info config map failed, err is %v", err)
		return err
	}
	return nil
}
