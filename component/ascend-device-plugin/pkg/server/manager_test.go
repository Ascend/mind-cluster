/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
+   Licensed under the Apache License, Version 2.0 (the "License");
+   you may not use this file except in compliance with the License.
+   You may obtain a copy of the License at
+
+   http://www.apache.org/licenses/LICENSE-2.0
+
+   Unless required by applicable law or agreed to in writing, software
+   distributed under the License is distributed on an "AS IS" BASIS,
+   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
+   See the License for the specific language governing permissions and
+   limitations under the License.
+*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

const (
	serverNum  = 2
	rqtTaskNum = 4
)

func setPatch() *gomonkey.Patches {
	patch := gomonkey.ApplyFuncReturn(kubeclient.NewClientK8s, &kubeclient.ClientK8s{
		Clientset:      &kubernetes.Clientset{},
		NodeName:       "node",
		DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
		IsApiErr:       false,
	}, nil).
		ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Nodes(), "Get", &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
		}, nil)
	return patch
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Chmod(common.SocketChmod)
}

// TestNewHwDevManager for testNewHwDevManager
func TestNewHwDevManager(t *testing.T) {
	patch := setPatch()
	defer patch.Reset()
	convey.Convey("test NewHwDevManager", t, func() {
		if _, err := os.Stat(common.HiAIManagerDevice); err != nil {
			if err = createFile(common.HiAIManagerDevice); err != nil {
				t.Fatal("TestGetDefaultDevices Run Failed")
			}
		}
		mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
			func(_ *device.AscendTools) (int32, error) {
				return common.DeviceNotSupport, nil
			})
		defer mockGetChipAiCoreCount.Reset()
		mockUpdateNodeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateNode",
			func(_ *HwDevManager) error {
				return nil
			})
		defer mockUpdateNodeLabel.Reset()
		convey.Convey("init HwDevManager", func() {
			common.ParamOption.UseVolcanoType = true
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldNotBeNil)
		})
		convey.Convey("init HwDevManager get device type failed", func() {
			mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDevType",
				func(_ *devmanager.DeviceManagerMock) string {
					return "errorType"
				})
			defer mockGetDevType.Reset()
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("test NewHwDevManager, product type is not supported", func() {
			common.ParamOption.ProductTypes = []string{common.Atlas300IDuo}
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldNotBeNil)
		})
	})
}

// TestSetAscendManager for testSetAscendManager
func TestSetAscendManager(t *testing.T) {
	var hdm HwDevManager
	devM := &devmanager.DeviceManagerMock{}
	mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
		func(_ *device.AscendTools) (int32, error) {
			return common.DeviceNotSupport, nil
		})
	defer mockGetChipAiCoreCount.Reset()
	convey.Convey("test devType is Ascend310", t, func() {
		mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDevType",
			func(_ *devmanager.DeviceManagerMock) string {
				return common.Ascend310
			})
		defer mockGetDevType.Reset()
		err := hdm.setAscendManager(devM)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test devType is Ascend310P", t, func() {
		mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDevType",
			func(_ *devmanager.DeviceManagerMock) string {
				return common.Ascend310P
			})
		defer mockGetDevType.Reset()
		err := hdm.setAscendManager(devM)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test GetChipAiCoreCount return error", t, func() {
		mockGetChipAiCoreCount = gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
			func(_ *device.AscendTools) (int32, error) {
				return 0, fmt.Errorf("getChipAiCoreCount error")
			})
		defer mockGetChipAiCoreCount.Reset()
		err := hdm.setAscendManager(devM)
		convey.So(err.Error(), convey.ShouldEqual, "getChipAiCoreCount error")
	})
}

// TestUpdateNode for test update node
func TestUpdateNode(t *testing.T) {
	var hdm HwDevManager
	hdm.manager = device.NewHwAscend310Manager()
	convey.Convey("test update node when scene is edge", t, func() {
		tmpBuildScene := common.ParamOption.BuildScene
		common.ParamOption.BuildScene = common.EdgeScene
		err := hdm.UpdateNode()
		convey.So(err, convey.ShouldBeNil)
		common.ParamOption.BuildScene = tmpBuildScene
	})
	mockInitPodInformer := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "InitPodInformer",
		func(_ *kubeclient.ClientK8s) {})
	defer mockInitPodInformer.Reset()
	convey.Convey("test update node when get node error", t, func() {
		mockGetNode := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "GetNode", func(_ *kubeclient.ClientK8s) (
			*v1.Node, error) {
			return &v1.Node{}, fmt.Errorf("getNode error")
		})
		defer mockGetNode.Reset()
		err := hdm.UpdateNode()
		convey.So(err.Error(), convey.ShouldEqual, "getNode error")
	})
	mockMarshal := gomonkey.ApplyFuncReturn(json.Marshal, []byte{0}, nil)
	defer mockMarshal.Reset()
	convey.Convey("test update node when update node label success", t, func() {
		testLabel := map[string]string{"testKey": "testValue"}
		mockGetNewNodeLabel := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(HwDevManager)), "getNewNodeLabel",
			func(_ *HwDevManager, _ *v1.Node) (map[string]string, error) { return testLabel, nil })
		defer mockGetNewNodeLabel.Reset()
		mockGetNode := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "GetNode", func(_ *kubeclient.ClientK8s) (
			*v1.Node, error) {
			return &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: make(map[string]string),
					Labels:      make(map[string]string),
					Name:        "node",
				},
				Status: v1.NodeStatus{Addresses: getAddresses()},
			}, nil
		})
		defer mockGetNode.Reset()
		mockPatchNodeState := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "PatchNodeState", func(
			_ *kubeclient.ClientK8s, _, _ *v1.Node) (*v1.Node, []byte, error) {
			return &v1.Node{}, []byte{}, nil
		})
		defer mockPatchNodeState.Reset()
		err := hdm.UpdateNode()
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetNewNodeLabel for test getNewNodeLabel
func TestGetNewNodeLabel(t *testing.T) {
	hdm := &HwDevManager{
		manager: device.NewHwAscend310Manager(),
		allInfo: common.NpuAllInfo{
			AllDevs: []common.NpuDevice{{LogicID: 0}},
		},
	}
	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{common.ServerTypeLabelKey: "test server type"},
			Name:   "node",
		}}
	mockGetDmgr := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetDmgr",
		func(_ *device.AscendTools) devmanager.DeviceInterface { return &devmanager.DeviceManagerMock{} })
	defer mockGetDmgr.Reset()
	convey.Convey("test getNewNodeLabel when chip info error", t, func() {
		mockGetValidChipInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetValidChipInfo", func(_ *devmanager.DeviceManagerMock) (npuCommon.ChipInfo, error) {
				return npuCommon.ChipInfo{}, fmt.Errorf("chip info error")
			})
		defer mockGetValidChipInfo.Reset()
		labelMap, err := hdm.getNewNodeLabel(testNode)
		convey.So(labelMap, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "chip info error")
	})
	convey.Convey("test getNewNodeLabel success", t, func() {
		mockGetDeviceUsage := gomonkey.ApplyMethod(&device.AscendTools{}, "GetDeviceUsage",
			func(_ *device.AscendTools) string {
				return common.Infer
			})
		defer mockGetDeviceUsage.Reset()
		mockGetBoardInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetBoardInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (npuCommon.BoardInfo, error) {
				return npuCommon.BoardInfo{BoardId: common.A300IA2BoardId}, nil
			})
		defer mockGetBoardInfo.Reset()
		mockGetValidChipInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetValidChipInfo", func(_ *devmanager.DeviceManagerMock) (npuCommon.ChipInfo, error) {
				return npuCommon.ChipInfo{Name: "testName"}, nil
			})
		defer mockGetValidChipInfo.Reset()
		mockIsContainAll300IDuo := gomonkey.ApplyFuncReturn(common.IsContainAll300IDuo, true)
		defer mockIsContainAll300IDuo.Reset()
		labelMap, err := hdm.getNewNodeLabel(testNode)
		convey.So(labelMap, convey.ShouldResemble, map[string]string{common.InferCardKey: common.A300IDuoLabel,
			common.ChipNameLabel: "testName"})
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestStartAllServer for testStartAllServer
func TestStartAllServer(t *testing.T) {
	mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
		func(_ *device.AscendTools) (int32, error) {
			return common.DeviceNotSupport, nil
		})
	defer mockGetChipAiCoreCount.Reset()
	mockUpdateNodeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateNode",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockUpdateNodeLabel.Reset()
	convey.Convey("test startAllServer", t, func() {
		patch := setPatch()
		defer patch.Reset()
		mockStart := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "Start",
			func(_ *PluginServer, socketWatcher *common.FileWatch) error {
				return fmt.Errorf("error")
			})
		defer mockStart.Reset()
		common.ParamOption.PresetVDevice = true
		hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
		res := hdm.startAllServer(&common.FileWatch{})
		convey.So(res, convey.ShouldBeFalse)
	})
}

// TestUpdatePodAnnotation for testUpdatePodAnnotation
func TestUpdatePodAnnotation(t *testing.T) {
	node := getMockNode(common.Ascend310P)
	podDeviceInfo := getMockDeviceInfo()
	mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
		func(_ *device.AscendTools) (int32, error) {
			return common.DeviceNotSupport, nil
		})
	defer mockGetChipAiCoreCount.Reset()
	mockUpdateNodeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateNode",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockUpdateNodeLabel.Reset()
	convey.Convey("test updatePodAnnotation", t, func() {
		convey.Convey("updatePodAnnotation success", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return node, nil
				})
			mockPodDeviceInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "GetKltAndRealAllocateDev",
				func(_ *PluginServer, _ []v1.Pod) ([]*common.PodDeviceInfo, error) {
					return podDeviceInfo, nil
				})
			mockManager := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "AddPodAnnotation",
				func(_ *device.AscendTools, _ *common.PodDeviceInfo, _ string, _ string, _ []common.NpuDevice) error {
					return nil
				})
			mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodListCache",
				func(_ *kubeclient.ClientK8s) []v1.Pod {
					return []v1.Pod{}
				})
			mockClearCM := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(HwDevManager)), "tryToClearResetInfoCM",
				func(_ *HwDevManager, _ v1.Pod) error {
					return nil
				})
			patch := setPatch()
			defer patch.Reset()
			defer mockPodList.Reset()
			defer mockManager.Reset()
			defer mockNode.Reset()
			defer mockPodDeviceInfo.Reset()
			defer mockClearCM.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestUpdateDevice for testUpdateDevice
func TestUpdateDevice(t *testing.T) {
	mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
		func(_ *device.AscendTools) (int32, error) {
			return common.DeviceNotSupport, nil
		})
	defer mockGetChipAiCoreCount.Reset()
	mockUpdateNodeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateNode",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockUpdateNodeLabel.Reset()
	convey.Convey("test UpdateDevice", t, func() {
		convey.Convey("UpdateDevice success", func() {
			mockCheckLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)),
				"CheckDeviceTypeLabel",
				func(_ *device.AscendTools) error {
					return nil
				})
			mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
				func(_ *PluginServer) error {
					return nil
				})
			patch := setPatch()
			defer patch.Reset()
			defer mockDestroy.Reset()
			defer mockCheckLabel.Reset()
			common.ParamOption.PresetVDevice = true
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.Ascend310P, nil, nil, nil)
			err := hdm.updateAllInfo()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestNotifyToK8s for testNotifyToK8s
func TestNotifyToK8s(t *testing.T) {
	mockGetChipAiCoreCount := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChipAiCoreCount",
		func(_ *device.AscendTools) (int32, error) {
			return common.DeviceNotSupport, nil
		})
	defer mockGetChipAiCoreCount.Reset()
	mockUpdateNodeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateNode",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockUpdateNodeLabel.Reset()
	convey.Convey("test NotifyToK8s", t, func() {
		convey.Convey("NotifyToK8s success", func() {
			mockUpdateHealth := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "UpdateHealth",
				func(_ *device.AscendTools, _ map[string][]*common.NpuDevice, _ []*common.NpuDevice, _ string) {
					return
				})
			mockGrace := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(HwDevManager)), "graceTolerance",
				func(_ *HwDevManager, _ map[string][]*common.NpuDevice) {
					return
				})
			mockChange := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "GetChange",
				func(_ *device.AscendTools, _ map[string][]*common.NpuDevice, _ map[string][]*common.NpuDevice) map[string]bool {
					return map[string]bool{common.Ascend310P: true, common.Ascend310: false}
				})
			patch := setPatch()
			defer patch.Reset()
			defer mockUpdateHealth.Reset()
			defer mockGrace.Reset()
			defer mockChange.Reset()
			common.ParamOption.PresetVDevice = true
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.Ascend310P, nil, nil, nil)
			initTime := time.Now()
			hdm.notifyToK8s(&initTime)
			convey.So(len(hdm.ServerMap), convey.ShouldEqual, serverNum)
		})
	})
}

func getMockPod() v1.Pod {
	limitValue := v1.ResourceList{
		common.HuaweiAscend910: *resource.NewQuantity(rqtTaskNum, resource.BinarySI),
	}
	annotation := map[string]string{
		common.HuaweiAscend910: "0-vir01",
	}
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-dls-npu-1p-default-2p-0",
			Namespace:   "btg-test",
			Annotations: annotation,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Resources: v1.ResourceRequirements{
					Limits: limitValue,
				}},
			},
		},
		Status: v1.PodStatus{
			Reason: "UnexpectedAdmissionError1",
			ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}},
			},
		},
	}
}

func getMockNode(ascendType string) *v1.Node {
	return &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceName(ascendType): resource.Quantity{},
			},
			Addresses: getAddresses(),
		},
	}
}

func getAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    v1.NodeHostName,
			Address: common.DefaultDeviceIP,
		},
	}
}

func getMockDeviceInfo() []*common.PodDeviceInfo {
	return []*common.PodDeviceInfo{
		{
			Pod:        getMockPod(),
			KltDevice:  []string{},
			RealDevice: []string{},
		},
		{
			Pod:        getMockPod(),
			KltDevice:  []string{""},
			RealDevice: []string{""},
		},
	}
}
