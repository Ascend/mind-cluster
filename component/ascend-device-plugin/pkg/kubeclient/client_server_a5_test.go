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

// Package kubeclient a series of k8s function ut
package kubeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestClientK8sMethodWriteDeviceInfoDataIntoCMA5(t *testing.T) {
	convey.Convey("test ClientK8s method WriteDeviceInfoDataIntoCMA5", t, func() {
		utKubeClient, err := initK8S()
		if err != nil {
			t.Fatal("test WriteDeviceInfoDataIntoCMA5 init kubernetes failed")
		}
		nodeDeivceData := common.NodeDeviceInfoCache{
			DeviceInfo: common.NodeDeviceInfo{DeviceList: map[string]string{}},
		}
		convey.Convey("01-should return err when common.MarshalData failed", func() {
			patch := gomonkey.ApplyFunc(common.MarshalData, func(data interface{}) []byte {
				return nil
			})
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMA5(&nodeDeivceData, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return success when all is called success", func() {
			patch := gomonkey.ApplyPrivateMethod(utKubeClient, "createOrUpdateDeviceCM",
				func(cm *v1.ConfigMap) error {
					return nil
				})
			defer patch.Reset()
			ret := utKubeClient.WriteDeviceInfoDataIntoCMA5(&nodeDeivceData, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}

func TestWriteDpuDataIntoCM(t *testing.T) {
	convey.Convey("Test WriteDpuDataIntoCM", t, func() {
		ki := &ClientK8s{NodeName: "work1"}
		busType := "test-bus"
		dpuList := []common.DpuCMData{
			{DeviceID: "dpu1", Operstate: "up"},
			{DeviceID: "dpu2", Operstate: "up"},
		}
		npuToDpusMap := map[string][]string{
			"npu1": {"dpu1", "dpu2"},
		}
		dpuListJson, err0 := json.Marshal(dpuList)
		convey.So(err0, convey.ShouldBeNil)
		npuToDpusMapJson, err1 := json.Marshal(npuToDpusMap)
		convey.So(err1, convey.ShouldBeNil)
		patches := gomonkey.ApplyFunc(common.MarshalData, func(v interface{}) []byte {
			switch v.(type) {
			case []common.DpuCMData:
				return dpuListJson
			case map[string][]string:
				return npuToDpusMapJson
			default:
				return []byte{}
			}
		})
		defer patches.Reset()
		convey.Convey("When createOrUpdateDeviceCM succeeds", func() {
			patches.ApplyPrivateMethod(ki, "createOrUpdateDeviceCM",
				func(_ *ClientK8s, cm *v1.ConfigMap) error {
					convey.So(cm.Name, convey.ShouldEqual, api.DpuInfoCMNamePrefix+ki.NodeName)
					convey.So(cm.Namespace, convey.ShouldEqual, api.KubeNS)
					convey.So(cm.Labels[api.CIMCMLabelKey], convey.ShouldEqual, common.CmConsumerValue)
					convey.So(cm.Data[api.DpuInfoCMBusTypeKey], convey.ShouldEqual, busType)
					return nil
				})
			err := ki.WriteDpuDataIntoCM(busType, dpuList, npuToDpusMap)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("When createOrUpdateDeviceCM fails", func() {
			expectedErr := fmt.Errorf("create configmap failed")
			patches.ApplyPrivateMethod(ki, "createOrUpdateDeviceCM",
				func(_ *ClientK8s, _ *v1.ConfigMap) error {
					return expectedErr
				})
			err := ki.WriteDpuDataIntoCM(busType, dpuList, npuToDpusMap)
			convey.So(err, convey.ShouldEqual, expectedErr)
		})
	})
}

func TestWriteDpuDataIntoCM2(t *testing.T) {
	convey.Convey("Test WriteDpuDataIntoCM2", t, func() {
		ki := &ClientK8s{NodeName: "work1"}
		busType := "test-bus"
		var emptyDpuList []common.DpuCMData
		emptyNpuMap := map[string][]string{}
		emptyDpuListJson, err0 := json.Marshal(emptyDpuList)
		convey.So(err0, convey.ShouldBeNil)
		emptyNpuMapJson, err1 := json.Marshal(emptyNpuMap)
		convey.So(err1, convey.ShouldBeNil)
		patches := gomonkey.ApplyFunc(common.MarshalData, func(v interface{}) []byte {
			switch v.(type) {
			case []common.DpuCMData:
				return emptyDpuListJson
			case map[string][]string:
				return emptyNpuMapJson
			default:
				return []byte{}
			}
		})
		defer patches.Reset()
		patches.ApplyPrivateMethod(ki, "createOrUpdateDeviceCM",
			func(_ *ClientK8s, cm *v1.ConfigMap) error {
				convey.So(cm.Data[api.DpuInfoCMDataKey], convey.ShouldEqual, string(emptyDpuListJson))
				convey.So(cm.Data[api.DpuInfoCMNpuToDpusMapKey], convey.ShouldEqual, string(emptyNpuMapJson))
				return nil
			})
		err := ki.WriteDpuDataIntoCM(busType, emptyDpuList, emptyNpuMap)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestWriteDpuDataIntoCM3(t *testing.T) {
	convey.Convey("Test WriteDpuDataIntoCM3", t, func() {
		ki := &ClientK8s{NodeName: "work1"}
		busType := "test-bus"
		dpuList := []common.DpuCMData{
			{DeviceID: "dpu1", Operstate: "up"},
			{DeviceID: "dpu2", Operstate: "up"},
		}
		npuToDpusMap := map[string][]string{
			"npu1": {"dpu1", "dpu2"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		npuToDpusMapJson, err1 := json.Marshal(npuToDpusMap)
		convey.So(err1, convey.ShouldBeNil)
		patches.ApplyFunc(common.MarshalData, func(v interface{}) []byte {
			switch v.(type) {
			case []common.DpuCMData:
				return nil
			case map[string][]string:
				return npuToDpusMapJson
			default:
				return []byte{}
			}
		})
		patches.ApplyPrivateMethod(ki, "createOrUpdateDeviceCM",
			func(_ *ClientK8s, cm *v1.ConfigMap) error {
				convey.So(cm.Name, convey.ShouldEqual, api.DpuInfoCMNamePrefix+ki.NodeName)
				convey.So(cm.Namespace, convey.ShouldEqual, api.KubeNS)
				convey.So(cm.Labels[api.CIMCMLabelKey], convey.ShouldEqual, common.CmConsumerValue)
				convey.So(cm.Data[api.DpuInfoCMBusTypeKey], convey.ShouldEqual, busType)
				convey.So(cm.Data[api.DpuInfoCMDataKey], convey.ShouldBeEmpty)
				return fmt.Errorf("unable to create configmap")
			})
		err := ki.WriteDpuDataIntoCM(busType, dpuList, npuToDpusMap)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unable to create configmap")
	})
}

func TestWriteDpuDataIntoCM4(t *testing.T) {
	convey.Convey("Test WriteDpuDataIntoCM4", t, func() {
		ki := &ClientK8s{NodeName: "work1"}
		busType := "test-bus"
		dpuList := []common.DpuCMData{
			{DeviceID: "dpu1", Operstate: "up"},
			{DeviceID: "dpu2", Operstate: "up"},
		}
		npuToDpusMap := map[string][]string{
			"npu1": {"dpu1", "dpu2"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		dpuListJson, err0 := json.Marshal(dpuList)
		convey.So(err0, convey.ShouldBeNil)
		patches.ApplyFunc(common.MarshalData, func(v interface{}) []byte {
			switch v.(type) {
			case []common.DpuCMData:
				return dpuListJson
			case map[string][]string:
				return nil
			default:
				return []byte{}
			}
		})
		patches.ApplyPrivateMethod(ki, "createOrUpdateDeviceCM",
			func(_ *ClientK8s, cm *v1.ConfigMap) error {
				convey.So(cm.Name, convey.ShouldEqual, api.DpuInfoCMNamePrefix+ki.NodeName)
				convey.So(cm.Namespace, convey.ShouldEqual, api.KubeNS)
				convey.So(cm.Labels[api.CIMCMLabelKey], convey.ShouldEqual, common.CmConsumerValue)
				convey.So(cm.Data[api.DpuInfoCMBusTypeKey], convey.ShouldEqual, busType)
				convey.So(cm.Data[api.DpuInfoCMNpuToDpusMapKey], convey.ShouldBeEmpty)
				return fmt.Errorf("unable to create configmap")
			})
		err := ki.WriteDpuDataIntoCM(busType, dpuList, npuToDpusMap)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
