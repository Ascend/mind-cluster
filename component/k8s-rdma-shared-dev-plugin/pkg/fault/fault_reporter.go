/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package fault for fault check and fault report
package fault

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"ascend-common/common-utils/hwlog"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

const (
	faultReporterNamespace = "kube-system"
	cmNamePrefix           = "dpuinfo-"
	dpuInfoCfgKey          = "DpuInfoCfg"
	cmLabelKey             = "huawei.com/consumer.clusterd"
	cmLabelValue           = "true"
	forceUpdateInterval    = 5 * time.Minute
)

var (
	lastReportedData   *DpuInfoCfg
	lastReportedDataMu sync.RWMutex
	k8sClientset       *kubernetes.Clientset
	appCtx             context.Context
)

// ReportToConfigMap reports DPU information to a Kubernetes ConfigMap
func ReportToConfigMap(dpuCfg DpuInfoCfg) error {
	nodeName, err := utils.GetNodeName()
	if err != nil {
		return fmt.Errorf("failed to get node name: %v", err)
	}
	lastReportedDataMu.RLock()
	needUpdate := lastReportedData == nil ||
		time.Since(lastReportedData.UpdateTime) >= forceUpdateInterval ||
		!isDpuInfoCfgEqual(lastReportedData, &dpuCfg)
	lastReportedDataMu.RUnlock()

	if !needUpdate {
		hwlog.RunLog.Debugf("No changes detected, skipping update for ConfigMap %s/%s",
			faultReporterNamespace, cmNamePrefix+nodeName)
		return nil
	}

	cfgJSONStr, err := json.Marshal(dpuCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal dpu info: %v", err)
	}

	cmName := cmNamePrefix + nodeName
	if err := CreateOrUpdateConfigMap(cmName, string(cfgJSONStr)); err != nil {
		return err
	}

	updateLastReportedData(&dpuCfg)
	return nil
}

// CreateOrUpdateConfigMap creates a new ConfigMap or updates an existing one
// Following the clusterd pattern: Create first, then Update if already exists
func CreateOrUpdateConfigMap(cmName, cfgJSONStr string) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cmName,
			Labels: map[string]string{cmLabelKey: cmLabelValue},
		},
		Data: map[string]string{dpuInfoCfgKey: cfgJSONStr},
	}

	cmClient := k8sClientset.CoreV1().ConfigMaps(faultReporterNamespace)
	_, cErr := cmClient.Create(appCtx, cm, metav1.CreateOptions{})
	if cErr == nil {
		hwlog.RunLog.Infof("Created ConfigMap %s/%s", faultReporterNamespace, cmName)
		return nil
	}
	if !apierrors.IsAlreadyExists(cErr) {
		hwlog.RunLog.Errorf("Failed to create ConfigMap %s/%s: %v", faultReporterNamespace, cmName, cErr)
		return fmt.Errorf("failed to create configmap %s/%s: %v", faultReporterNamespace, cmName, cErr)
	}

	_, err := cmClient.Update(appCtx, cm, metav1.UpdateOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("Failed to update ConfigMap %s/%s: %v", faultReporterNamespace, cmName, err)
		return fmt.Errorf("failed to update configmap %s/%s: %v", faultReporterNamespace, cmName, err)
	}
	hwlog.RunLog.Infof("Updated ConfigMap %s/%s", faultReporterNamespace, cmName)
	return nil
}

// isDpuInfoCfgEqual compares two DpuInfoCfg structures for equality
func isDpuInfoCfgEqual(a, b *DpuInfoCfg) bool {
	if len(a.DPUInfo.DPUList) != len(b.DPUInfo.DPUList) {
		return false
	}

	for i := range a.DPUInfo.DPUList {
		if !isDPUItemEqual(&a.DPUInfo.DPUList[i], &b.DPUInfo.DPUList[i]) {
			return false
		}
	}

	return true
}

// isDPUItemEqual compares two DPUItem structures for equality
func isDPUItemEqual(a, b *DPUItem) bool {
	if a.HcaName != b.HcaName ||
		a.EthName != b.EthName ||
		a.IpAddr != b.IpAddr ||
		a.DeviceID != b.DeviceID ||
		a.VendorID != b.VendorID {
		return false
	}

	return reflect.DeepEqual(a.FaultList, b.FaultList)
}

// updateLastReportedData updates the cached last reported DPU information
func updateLastReportedData(dpuCfg *DpuInfoCfg) {
	lastReportedDataMu.Lock()
	defer lastReportedDataMu.Unlock()

	if dpuCfg == nil {
		lastReportedData = nil
		return
	}

	data, err := json.Marshal(dpuCfg)
	if err != nil {
		hwlog.RunLog.Errorf("Failed to marshal dpu info for caching: %v", err)
		return
	}

	var cachedCfg DpuInfoCfg
	if err := json.Unmarshal(data, &cachedCfg); err != nil {
		hwlog.RunLog.Errorf("Failed to unmarshal dpu info for caching: %v", err)
		return
	}

	lastReportedData = &cachedCfg
}

// StartFaultReporting starts the fault reporting loop
// It receives fault detection results from faultResultChan and reports them to ConfigMap
func StartFaultReporting(ctx context.Context, clientset *kubernetes.Clientset) {
	k8sClientset = clientset
	appCtx = ctx

	hwlog.RunLog.Infof("Fault reporting goroutine started, force update interval: %v", forceUpdateInterval)

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("Fault reporting goroutine stopped")
			return
		case dpuCfg := <-faultResultChan:
			if err := ReportToConfigMap(dpuCfg); err != nil {
				hwlog.RunLog.Errorf("Failed to report faults to ConfigMap: %v", err)
			}
		}
	}
}
