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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

// WriteDeviceInfoDataIntoCMA5 write deviceinfo into config map for A5
func (ki *ClientK8s) WriteDeviceInfoDataIntoCMA5(nodeDeviceData *common.NodeDeviceInfoCache,
	manuallySeparateNPU string, switchInfo common.SwitchFaultInfo) error {

	nodeDeviceData.DeviceInfo.UpdateTime = time.Now().Unix()
	nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)

	var data, switchData []byte
	if data = common.MarshalData(nodeDeviceData); len(data) == 0 {
		return fmt.Errorf("marshal nodeDeviceData failed")
	}
	if switchData = common.MarshalData(switchInfo); len(switchData) == 0 {
		return fmt.Errorf("marshal switchDeviceData failed")
	}
	deviceInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ki.DeviceInfoName,
			Namespace: api.KubeNS,
			Labels:    map[string]string{api.CIMCMLabelKey: common.CmConsumerValue},
		},
	}
	deviceInfoCM.Data = map[string]string{
		api.DeviceInfoCMDataKey:                   string(data),
		api.SwitchInfoCMDataKey:                   string(switchData),
		common.DeviceInfoCMManuallySeparateNPUKey: manuallySeparateNPU}
	hwlog.RunLog.Debugf("write device info cache into cm: %s/%s.", deviceInfoCM.Namespace, deviceInfoCM.Name)
	return ki.createOrUpdateDeviceCM(deviceInfoCM)
}

// WriteDpuDataIntoCM write dpu info into configmap
func (ki *ClientK8s) WriteDpuDataIntoCM(busType string, dpuList []common.DpuCMData,
	npuToDpusMap map[string][]string) error {
	dpuInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      api.DpuInfoCMNamePrefix + ki.NodeName,
			Namespace: api.KubeNS,
			Labels:    map[string]string{api.CIMCMLabelKey: common.CmConsumerValue},
		},
	}
	dpuListJsonData := common.MarshalData(dpuList)
	npuToDpusMapJsonData := common.MarshalData(npuToDpusMap)
	dpuInfoCM.Data = map[string]string{
		api.DpuInfoCMBusTypeKey:      busType,
		api.DpuInfoCMDataKey:         string(dpuListJsonData),
		api.DpuInfoCMNpuToDpusMapKey: string(npuToDpusMapJsonData),
	}
	hwlog.RunLog.Debugf("%s write DPU info cache into cm: %s/%s.", api.DpuLogPrefix, dpuInfoCM.Namespace,
		dpuInfoCM.Name)

	const duration = 500 * time.Millisecond
	backoff := wait.Backoff{
		Steps:    3,
		Duration: duration,
		Factor:   2.0, // 0.5s -> 1s -> 2s
		Jitter:   0.1,
	}
	return retry.OnError(backoff,
		func(err error) bool {
			return true
		}, func() error {
			return ki.createOrUpdateDeviceCM(dpuInfoCM)
		},
	)
}
