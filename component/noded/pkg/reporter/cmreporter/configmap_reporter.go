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
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

const (
	defaultReportInterval = 30 * time.Minute
	retryTime             = 3
)

// ConfigMapReporter report fault device info by config map
type ConfigMapReporter struct {
	client        *kubeclient.ClientK8s
	nodeInfoCache common.NodeInfoCM
	reportTime    time.Time
}

// NewConfigMapReporter create a config map reporter
func NewConfigMapReporter(client *kubeclient.ClientK8s) *ConfigMapReporter {
	return &ConfigMapReporter{
		client:        client,
		nodeInfoCache: common.NodeInfoCM{},
		reportTime:    time.Now(),
	}
}

// Report send fault device info by config map
func (c *ConfigMapReporter) Report(faultDevInfo *common.FaultDevInfo) {
	if common.DeepEqualFaultDevInfo(faultDevInfo, &c.nodeInfoCache.NodeInfo) &&
		time.Since(c.reportTime) < defaultReportInterval {
		hwlog.RunLog.Debugf("node fault device info is not changed and report time is not reached, no need to report")
		return
	}
	checkCode := common.MakeDataHash(faultDevInfo)
	c.nodeInfoCache = common.NodeInfoCM{
		NodeInfo:  *faultDevInfo,
		CheckCode: checkCode,
	}

	data, err := json.Marshal(c.nodeInfoCache)
	if err != nil {
		hwlog.RunLog.Errorf("marshal node info cache failed, err is %v", err)
		return
	}
	nodeInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.client.NodeInfoName,
			Namespace: api.DLNamespace,
			Labels:    map[string]string{api.CIMCMLabelKey: common.CmConsumerValue},
		},
		Data: map[string]string{
			api.NodeInfoCMDataKey: string(data),
			"updateTime":          time.Now().Format(time.RFC3339),
		},
	}

	var initSuc bool
	for i := 0; i < retryTime; i++ {
		if err := c.client.CreateOrUpdateConfigMap(nodeInfoCM); err != nil {
			hwlog.RunLog.Errorf("report node fault device info to k8s by configmap failed, error: %v, "+
				"retry count: %d", err, i+1)
			time.Sleep(time.Second)
			continue
		}
		initSuc = true
		break
	}
	if !initSuc {
		hwlog.RunLog.Errorf("report node fault device info to k8s by configmap failed, "+
			"the maximum number of retries (%d) has been reached", retryTime)
		return
	}
	c.reportTime = time.Now()
	hwlog.RunLog.Infof("report node fault device info to k8s by configmap success, time is %s",
		c.reportTime.Format(time.RFC3339))
	return
}

// Init initialize node info config map
func (c *ConfigMapReporter) Init() error {
	nodeInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.client.NodeInfoName,
			Namespace: api.DLNamespace,
		},
		Data: map[string]string{},
	}
	if err := c.client.CreateOrUpdateConfigMap(nodeInfoCM); err != nil {
		hwlog.RunLog.Errorf("init node info config map failed, err is %v", err)
		return err
	}
	return nil
}
