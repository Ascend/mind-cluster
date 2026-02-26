/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of device function
package device

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/devmanager"
)

func createFake310pManager() *HwAscend310PManager {
	manager := NewHwAscend310PManager()
	manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return manager
}

func TestHwAscend310PManagerGetNPUs(t *testing.T) {
	convey.Convey("310p get npu", t, func() {
		manager := createFake310pManager()
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, api.Ascend310P)
		convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", api.Ascend310P, allInfo.AllDevs[0].PhyID))
	})
}

func TestDoWithVolcanoListAndWatch310p(t *testing.T) {
	convey.Convey("310p DoWithVolcanoListAndWatch", t, func() {
		manager := createFake310pManager()
		fakeKubeInteractor := &kubeclient.ClientK8s{Clientset: nil, NodeName: "NODE_NAME"}
		manager.SetKubeClient(fakeKubeInteractor)
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		groupDevice := ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)
		mockGetPodsUsedNpu := mockGetPodsUsedNpuByCommon()
		mockGetConfigMap := mockGetDeviceInfoCMCache(map[string]string{api.Ascend310P: api.Ascend310P + "-1"})
		mockCreateConfigMap := mockWriteDeviceInfoDataIntoCM()
		defer func() {
			mockGetPodsUsedNpu.Reset()
			mockGetConfigMap.Reset()
			mockCreateConfigMap.Reset()
		}()
		manager.client.SetNodeDeviceInfoCache(createFakeDeviceInfo())
		manager.DoWithVolcanoListAndWatch(groupDevice, 0)
	})
}
