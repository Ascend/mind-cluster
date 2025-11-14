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
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
)

func TestClientK8sMethodWriteDeviceInfoDataIntoCMCacheA5(t *testing.T) {
	convey.Convey("test ClientK8s method WriteDeviceInfoDataIntoCMCacheA5", t, func() {
		utKubeClient, err := initK8S()
		if err != nil {
			t.Fatal("test WriteDeviceInfoDataIntoCMA5 init kubernetes failed")
		}
		oldCache := nodeDeviceInfoCache
		defer func() {
			nodeDeviceInfoCache = oldCache
		}()
		nodeDeviceData := common.NodeDeviceInfoCache{
			DeviceInfo: common.NodeDeviceInfo{DeviceList: map[string]string{}},
		}
		convey.Convey("01-should return err when write into cm failed", func() {
			patch := gomonkey.ApplyMethodReturn(utKubeClient, "WriteDeviceInfoDataIntoCMA5",
				errors.New("write device info data into cm A5 failed"))
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMCacheA5(&nodeDeviceData, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return nil when write into cm success", func() {
			patch := gomonkey.ApplyMethodReturn(utKubeClient, "WriteDeviceInfoDataIntoCMA5", nil)
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMCacheA5(&nodeDeviceData, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}
