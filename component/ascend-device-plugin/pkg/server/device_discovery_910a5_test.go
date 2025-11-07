/*
 *  Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/devmanager"
	devcommon "ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

var (
	testDevM *devmanager.DeviceManager
	testHdm  *HwDevManager
	err      error

	mockDcmiVersion           = "24.0.rc2"
	mockCardNum         int32 = 16
	mockDeviceNumInCard int32 = 1
	mockCardList              = []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	mockProductType           = ""
	mockErr                   = errors.New("test error")
	mockChipInfo              = &devcommon.ChipInfo{
		Type:    "Ascend",
		Name:    "Ascend910 7591",
		Version: "V1",
	}
	mockBoardInfo = devcommon.BoardInfo{
		BoardId: devcommon.A900A5SuperPodBin1BoardId,
	}
	mockChipAICore     float32 = 25
	mockVirtualDevInfo         = devcommon.VirtualDevInfo{
		TotalResource: devcommon.CgoSocTotalResource{
			Computing: devcommon.CgoComputingResource{
				Aic: mockChipAICore,
			},
		},
	}
)

func setK8sPatch() *gomonkey.Patches {
	patch := gomonkey.ApplyFuncReturn(kubeclient.NewClientK8s, &kubeclient.ClientK8s{
		Clientset:      &kubernetes.Clientset{},
		NodeName:       "node-test",
		DeviceInfoName: common.DeviceInfoCMNamePrefix + "node-test",
		IsApiErr:       false,
	}, nil).
		ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetNode", &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}, Annotations: map[string]string{}},
		}, nil).
		ApplyMethod(&kubeclient.ClientK8s{}, "InitPodInformer", func(*kubeclient.ClientK8s) {}).
		ApplyMethodReturn(&kubeclient.ClientK8s{}, "PatchNodeState", &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
		}, []byte{}, nil)
	return patch
}

func setDcmiPatch() *gomonkey.Patches {
	patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcInit", nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDcmiVersion", mockDcmiVersion, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceNumInCard", mockDeviceNumInCard, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetChipInfo", mockChipInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetDeviceBoardInfo", mockBoardInfo, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetProductType", mockProductType, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetLogicIDList", mockCardNum, mockCardList, nil).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetVDeviceInfo", mockVirtualDevInfo, nil).
		ApplyMethod(&dcmi.DcManager{}, "DcGetPhysicIDFromLogicID",
			func(d *dcmi.DcManager, logicID int32) (int32, error) {
				return logicID, nil
			}).
		ApplyMethod(&dcmi.DcManager{}, "DcGetCardIDDeviceID",
			func(d *dcmi.DcManager, logicID int32) (int32, int32, error) {
				return logicID, 0, nil
			}).
		ApplyMethodReturn(&dcmi.DcManager{}, "DcGetSuperPodInfo", devcommon.CgoSuperPodInfo{}, mockErr)
	return patch
}

// TestDeviceDiscovery for Ascend910A5
func TestDeviceDiscovery(t *testing.T) {
	common.ParamOption.PresetVDevice = true
	common.ParamOption.GetFdFlag = false
	common.ParamOption.UseVolcanoType = true
	if _, err := os.Stat(common.HiAIManagerDevice); err != nil {
		if err = createFile(common.HiAIManagerDevice); err != nil {
			t.Fatal("create device file Failed")
		}
	}
	dcmiPatch := setDcmiPatch()
	defer dcmiPatch.Reset()
	k8sPatch := setK8sPatch()
	defer k8sPatch.Reset()

	convey.Convey("test auto init", t, testAutoInit)
	convey.Convey("test new hw device manager", t, testNewHwDevManager)
	convey.Convey("test label node", t, testLabelNode)
	convey.Convey("test list and watch", t, testListAndWatch)
}

func testAutoInit() {
	convey.Convey("test auto init success", func() {
		testDevM, err = devmanager.AutoInit("", api.DefaultDeviceResetTimeout)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test auto init failed, get card list failed", func() {
		patch := gomonkey.ApplyMethodReturn(&dcmi.DcManager{}, "DcGetCardList", mockCardNum, mockCardList, mockErr)
		defer patch.Reset()

		expectErr := fmt.Errorf("get card list failed for init")
		expectErr2 := fmt.Errorf("auto init failed, err: %v", expectErr)
		errDevM, err := devmanager.AutoInit("", api.DefaultDeviceResetTimeout)
		convey.So(err, convey.ShouldResemble, expectErr2)
		convey.So(errDevM, convey.ShouldBeNil)
	})
}

func testNewHwDevManager() {
	testHdm = NewHwDevManager(testDevM)
	convey.So(testHdm.allInfo.AllDevTypes, convey.ShouldResemble, []string{api.Ascend910})
	convey.So(testHdm.RunMode, convey.ShouldEqual, api.Ascend910)
	convey.So(testHdm.WorkMode, convey.ShouldEqual, devcommon.AMPMode)
}

func testLabelNode() {
	convey.Convey("serverType label correct", func() {
		labelServerType := common.ParamOption.RealCardType + common.MiddelLine +
			strconv.Itoa(int(common.ParamOption.AiCoreCount))
		convey.So(common.ParamOption.AiCoreCount, convey.ShouldEqual, mockChipAICore)
		convey.So(labelServerType, convey.ShouldEqual, "Ascend910A5-25")
	})

	convey.Convey("chipName label correct", func() {
		newLabelMap, err := testHdm.updateChipNameToNode()
		convey.So(err, convey.ShouldBeNil)
		convey.So(newLabelMap[common.ChipNameLabel], convey.ShouldEqual, "Ascend910 7591")
	})
}

func testListAndWatch() {
	pluginServer := testHdm.ServerMap[api.Ascend910]
	ps, ok := pluginServer.(*PluginServer)
	if !ok {
		panic("get plugin server Failed")
	}
	mockSend := gomonkey.ApplyFunc(sendToKubelet, func(stream v1beta1.DevicePlugin_ListAndWatchServer,
		resp *v1beta1.ListAndWatchResponse) error {
		return nil
	})

	convey.Convey("Notify failed", func() {
		ret := ps.Notify(testHdm.groupDevice[api.Ascend910])
		convey.So(ret, convey.ShouldBeFalse)
	})

	mockSend.Reset()
}
