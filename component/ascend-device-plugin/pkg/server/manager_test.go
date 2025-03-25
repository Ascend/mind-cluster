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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

const (
	serverNum  = 2
	rqtTaskNum = 4
)

var testErr = errors.New("test")

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
	mockGetDmgr := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend310Manager)), "GetDmgr",
		func(_ *device.HwAscend310Manager) devmanager.DeviceInterface { return &devmanager.DeviceManagerMock{} })
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
			common.ChipNameLabel: "testName", api.NPUChipMemoryKey: "0G"})
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

type mockDevManager struct {
	device.DevManager
}

// GetKubeClient mocks GetKubeClient
func (mdm mockDevManager) GetKubeClient() *kubeclient.ClientK8s {
	return nil
}

// GetNPUs mocks GetNPUs
func (mdm mockDevManager) GetNPUs() (common.NpuAllInfo, error) {
	return common.NpuAllInfo{}, testErr
}

// TestCheckNodeResetInfo tests the function checkNodeResetInfo
func TestCheckNodeResetInfo(t *testing.T) {
	hdm := HwDevManager{}
	flag := false
	convey.Convey("test checkNodeResetInfo", t, func() {
		patch := gomonkey.ApplyFunc(device.WriteResetInfo,
			func(resetInfo device.ResetInfo, writeMode device.WriteMode, update bool) {
				flag = true
			})
		patch.ApplyFuncReturn(checkOverRetryDev, device.ResetInfo{})
		defer patch.Reset()
		const id0 = 0
		hdm.manager = mockDevManager{}
		convey.Convey("01-client nil, flag should be false", func() {
			hdm.checkNodeResetInfo()
			convey.So(flag, convey.ShouldBeFalse)
		})
		patch.ApplyMethodReturn(&mockDevManager{}, "GetKubeClient", &kubeclient.ClientK8s{})
		patch.ApplyFuncReturn(device.GetResetInfoMgr, &device.ResetInfoMgr{})
		patch.ApplyFuncReturn(device.ReadResetInfo,
			device.ResetInfo{ThirdPartyResetDevs: []device.ResetDevice{
				{PhyID: id0},
			}})
		convey.Convey("02-get npus error, flag should be false", func() {
			hdm.checkNodeResetInfo()
			convey.So(flag, convey.ShouldBeFalse)
		})
		patch.ApplyMethodReturn(&mockDevManager{}, "GetNPUs", common.NpuAllInfo{}, nil)
		convey.Convey("03-success, flag should be true", func() {
			patch1 := gomonkey.ApplyFuncReturn(checkDeviceStatus, []device.ResetDevice{}, true)
			defer patch1.Reset()
			hdm.checkNodeResetInfo()
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}

// TestCheckDeviceStatus tests the function checkDeviceStatus
func TestCheckDeviceStatus(t *testing.T) {
	convey.Convey("test checkDeviceStatus", t, func() {
		const id1, id2, id3 = 1, 2, 3
		convey.Convey("01-status change, should return true", func() {
			patch := gomonkey.ApplyFunc(device.FreeBusyDev, func(cardID, deviceID int32) {
				return
			})
			defer patch.Reset()
			allInfo := map[string][]*common.NpuDevice{
				common.Ascend910: {
					&common.NpuDevice{
						PhyID:  int32(id1),
						Health: v1beta1.Healthy,
					},
				},
			}
			failDevs := []device.ResetDevice{
				{
					PhyID: id1,
				},
				{
					PhyID: id2,
				},
				{
					PhyID: id3,
				},
			}
			_, isChange := checkDeviceStatus(failDevs, allInfo)
			convey.So(isChange, convey.ShouldBeTrue)
		})
	})
}

// TestSetContainerdClient for test setContainerdClient
func TestSetContainerdClient(t *testing.T) {
	convey.Convey("test setContainerdClient", t, func() {
		hdm := &HwDevManager{
			manager: device.NewHwAscend310Manager(),
			allInfo: common.NpuAllInfo{
				AllDevs: []common.NpuDevice{{LogicID: 0}},
			},
		}
		convey.Convey("when not exist containerd sock file, result return err", func() {
			mock := gomonkey.ApplyFuncReturn(utils.IsExist, false)
			defer mock.Reset()
			err := hdm.setContainerdClient()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when containerd client create failed, result return err", func() {
			mock := gomonkey.ApplyFuncReturn(utils.IsExist, true).
				ApplyFuncReturn(containerd.New, nil, fmt.Errorf("test error"))
			defer mock.Reset()
			err := hdm.setContainerdClient()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when containerd client create success, return error is nil", func() {
			mock := gomonkey.ApplyFuncReturn(utils.IsExist, true).
				ApplyFuncReturn(containerd.New, &containerd.Client{}, nil)
			defer mock.Reset()
			err := hdm.setContainerdClient()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCheckOverRetryDev test the function checkOverRetryDev
func TestCheckOverRetryDev(t *testing.T) {
	const id = 0
	const numOne, numZero = 1, 0
	convey.Convey("test checkOverRetryDev", t, func() {
		input := device.ResetInfo{
			ThirdPartyResetDevs: []device.ResetDevice{{
				CardId: id,
			}},
			ManualResetDevs: make([]device.ResetDevice, 0),
		}
		convey.Convey("01-over retry time, dev should be add to manualDev", func() {
			patch1 := gomonkey.ApplyFuncReturn(device.GetResetCnt, common.MaxResetTimes+numOne)
			defer patch1.Reset()
			ret := checkOverRetryDev(input)
			convey.So(len(ret.ManualResetDevs), convey.ShouldEqual, numOne)
		})
		convey.Convey("02-not over retry times, dev be add to third party", func() {
			patch1 := gomonkey.ApplyFuncReturn(device.GetResetCnt, common.MaxResetTimes-numOne)
			defer patch1.Reset()
			ret := checkOverRetryDev(input)
			convey.So(len(ret.ManualResetDevs), convey.ShouldEqual, numZero)
		})
	})
}

// TestFlattenMap test the function flattenMap
func TestFlattenMap(t *testing.T) {
	const id = 0
	const targetLen = 2
	m := map[string][]*common.NpuDevice{
		common.Ascend910: {
			&common.NpuDevice{
				PhyID:  int32(id),
				Health: v1beta1.Healthy,
			},
		},
		common.Ascend310P: {
			&common.NpuDevice{
				PhyID:  int32(id),
				Health: v1beta1.Healthy,
			},
		},
	}
	ret := flattenMap(m)
	if len(ret) != targetLen {
		t.Errorf("expect len %v, got %v", targetLen, len(ret))
	}
}

// TestIsSupportGraceTolerance test isSupportGraceTolerance
func TestIsSupportGraceTolerance(t *testing.T) {
	var hdm HwDevManager
	tmpGraceToleranceOn := common.ParamOption.GraceToleranceOn
	tmpHotReset := common.ParamOption.HotReset

	common.ParamOption.GraceToleranceOn = false
	convey.Convey("test isSupportGraceTolerance when hot reset mode error", t, func() {
		common.ParamOption.HotReset = common.HotResetInfer
		hdm.isSupportGraceTolerance()
		convey.So(common.ParamOption.GraceToleranceOn, convey.ShouldNotEqual, true)
	})

	common.ParamOption.HotReset = common.HotResetTrainOnLine
	convey.Convey("test isSupportGraceTolerance when run mode is not Ascend910", t, func() {
		hdm.RunMode = common.Ascend310P
		hdm.isSupportGraceTolerance()
		convey.So(common.ParamOption.GraceToleranceOn, convey.ShouldNotEqual, true)
	})

	hdm.RunMode = common.Ascend910
	tmpRealCardType := common.ParamOption.RealCardType
	common.ParamOption.RealCardType = common.Ascend910
	convey.Convey("test isSupportGraceTolerance when SMP chip mode is not for Ascend910", t, func() {
		hdm.WorkMode = common.AMPMode
		hdm.isSupportGraceTolerance()
		convey.So(common.ParamOption.GraceToleranceOn, convey.ShouldNotEqual, true)
	})

	hdm.WorkMode = common.SMPMode
	convey.Convey("test isSupportGraceTolerance when GraceToleranceOn is true", t, func() {
		hdm.isSupportGraceTolerance()
		convey.So(common.ParamOption.GraceToleranceOn, convey.ShouldEqual, true)
	})

	common.ParamOption.HotReset = tmpHotReset
	common.ParamOption.RealCardType = tmpRealCardType
	common.ParamOption.GraceToleranceOn = tmpGraceToleranceOn
}

// TestUpdateAllInfo test updateAllInfo
func TestUpdateAllInfo(t *testing.T) {
	hdm := &HwDevManager{}
	tmpPresetVDevice := common.ParamOption.PresetVDevice
	convey.Convey("test updateAllInfo when PresetVDevice is true return nil", t, func() {
		common.ParamOption.PresetVDevice = true
		err := hdm.updateAllInfo()
		convey.So(err, convey.ShouldBeNil)
	})
	common.ParamOption.PresetVDevice = false
	convey.Convey("test updateAllInfo when not found npu-core in server map return error", t, func() {
		err := hdm.updateAllInfo()
		convey.So(err.Error(), convey.ShouldEqual, "not found npu-core plugin server")
	})
	hdm = &HwDevManager{ServerMap: map[string]InterfaceServer{common.AiCoreResourceName: &PluginServer{}},
		manager: &device.HwAscend310Manager{}}
	convey.Convey("test updateAllInfo when DestroyNotUsedVNPU is error return error", t, func() {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
			func(_ *PluginServer) error { return fmt.Errorf("error") })
		defer patch.Reset()
		err := hdm.updateAllInfo()
		convey.So(err.Error(), convey.ShouldEqual, "error")
	})
	patch := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
		func(_ *PluginServer) error { return nil })
	defer patch.Reset()
	mockCheckDeviceTypeLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "CheckDeviceTypeLabel",
		func(_ *device.AscendTools) error { return fmt.Errorf("error") })
	defer mockCheckDeviceTypeLabel.Reset()
	convey.Convey("test updateAllInfo when GetNPUs is error return error", t, func() {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend310Manager)), "GetNPUs", func(
			_ *device.HwAscend310Manager) (common.NpuAllInfo, error) {
			return common.NpuAllInfo{}, fmt.Errorf("error")
		})
		defer patch.Reset()
		err := hdm.updateAllInfo()
		convey.So(err.Error(), convey.ShouldEqual, "error")
	})
	mockGetNPUs := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend310Manager)), "GetNPUs",
		func(_ *device.HwAscend310Manager) (common.NpuAllInfo, error) {
			return common.NpuAllInfo{AllDevs: []common.NpuDevice{}, AllDevTypes: []string{}}, nil
		})
	defer mockGetNPUs.Reset()
	convey.Convey("test updateAllInfo success return nil", t, func() {
		err := hdm.updateAllInfo()
		convey.So(err, convey.ShouldBeNil)
	})
	common.ParamOption.PresetVDevice = tmpPresetVDevice
}

func mockGetConfigMap(configmap *v1.ConfigMap, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetConfigMap",
		func(_ *kubeclient.ClientK8s, _, _ string) (*v1.ConfigMap, error) {
			return configmap, err
		})
}

// TestTryToClearResetInfoCM1 test tryToClearResetInfoCM
func TestTryToClearResetInfoCM1(t *testing.T) {
	hdm := &HwDevManager{manager: &device.HwAscend310Manager{}}
	convey.Convey("test tryToClearResetInfoCM when get task name failed return error", t, func() {
		pod := v1.Pod{}
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual, "failed to get task name by task key")
	})
	pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{common.ResetTaskNameKey: "taskName"}}}
	mockGetKubeClient := gomonkey.ApplyMethodReturn(&mockDevManager{}, "GetKubeClient", &kubeclient.ClientK8s{})
	defer mockGetKubeClient.Reset()
	convey.Convey("test tryToClearResetInfoCM when get reset cm failed return error", t, func() {
		mockGetConfigMap := mockGetConfigMap(nil, errors.New("error"))
		defer mockGetConfigMap.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual, "error")
	})
	convey.Convey("test tryToClearResetInfoCM when reset.json not exist return error", t, func() {
		mockGetConfigMap := mockGetConfigMap(&v1.ConfigMap{Data: map[string]string{}}, nil)
		defer mockGetConfigMap.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual, "reset.json not exist")
	})
	convey.Convey("test tryToClearResetInfoCM when cm data size out of memory return error", t, func() {
		mockGetConfigMap := mockGetConfigMap(&v1.ConfigMap{Data: map[string]string{
			common.ResetInfoCMDataKey: string(make([]byte, common.CMDataMaxLength+1))}}, nil)
		defer mockGetConfigMap.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual, "configmap data size is out of memory")
	})
}

// TestTryToClearResetInfoCM2 test tryToClearResetInfoCM
func TestTryToClearResetInfoCM2(t *testing.T) {
	hdm := &HwDevManager{manager: &device.HwAscend310Manager{}}
	pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{common.ResetTaskNameKey: "taskName"}}}
	mockGetKubeClient := gomonkey.ApplyMethodReturn(&mockDevManager{}, "GetKubeClient", &kubeclient.ClientK8s{})
	defer mockGetKubeClient.Reset()
	mockGetConfigMap := mockGetConfigMap(&v1.ConfigMap{Data: map[string]string{common.ResetInfoCMDataKey: "key"}}, nil)
	defer mockGetConfigMap.Reset()
	convey.Convey("test tryToClearResetInfoCM when unmarshal configmap data failed return error", t, func() {
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual,
			"unmarshal configmap data failed, err: invalid character 'k' looking for beginning of value")
	})
	convey.Convey("test tryToClearResetInfoCM when reset info config map is initialized return nil", t, func() {
		mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, nil)
		defer mockUnmarshal.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err, convey.ShouldBeNil)
	})
	mockUnmarshal := gomonkey.ApplyFunc(json.Unmarshal, func(_ []byte, value any) error {
		taskResetInfo, ok := value.(*common.TaskResetInfo)
		if !ok {
			return errors.New("error")
		}
		taskResetInfo.UpdateTime = 1
		return nil
	})
	defer mockUnmarshal.Reset()
	convey.Convey("test tryToClearResetInfoCM when clear reset info failed return error", t, func() {
		mockClearResetInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "ClearResetInfo",
			func(_ *kubeclient.ClientK8s, _, _ string) error { return errors.New("error") })
		defer mockClearResetInfo.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err.Error(), convey.ShouldEqual, "clear reset configMap failed err is: error")
	})
	convey.Convey("test tryToClearResetInfoCM when clear reset info success return nil", t, func() {
		mockClearResetInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "ClearResetInfo",
			func(_ *kubeclient.ClientK8s, _, _ string) error { return nil })
		defer mockClearResetInfo.Reset()
		err := hdm.tryToClearResetInfoCM(pod)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestResetHccsServer tests the ResetHccsServer function.
func TestResetHccsServer(t *testing.T) {
	hdm := &HwDevManager{manager: &device.HwAscend310Manager{}}
	convey.Convey("Test ResetHccsServer", t, func() {
		convey.Convey("When all devices are healthy, do nothing", func() {
			devices := []*common.NpuDevice{{Health: v1beta1.Healthy}, {Health: v1beta1.Healthy}}
			hdm.ResetHccsServer("devType", devices, &PodResource{})
		})
		devices := []*common.NpuDevice{{Health: v1beta1.Unhealthy}}
		convey.Convey("When reset failed times exceed max limit, log warning and return", func() {
			patch := gomonkey.ApplyMethodReturn(hdm.manager, "GetResetFailedTimes", common.MaxResetTimes+1)
			defer patch.Reset()
			hdm.ResetHccsServer("devType", devices, &PodResource{})
		})
		convey.Convey("When cards are in resetting, do nothing", func() {
			patch := gomonkey.ApplyMethodReturn(hdm.manager, "GetIfCardsInResetting", true)
			defer patch.Reset()
			hdm.ResetHccsServer("devType", devices, &PodResource{})
		})
		patch1 := gomonkey.ApplyMethodReturn(hdm.manager, "GetResetFailedTimes", 0).ApplyMethodReturn(hdm.manager,
			"GetIfCardsInResetting", false)
		defer patch1.Reset()
		convey.Convey("When index out of range, log error and return", func() {
			devices := make([]*common.NpuDevice, 0)
			hdm.ResetHccsServer("devType", devices, &PodResource{})
		})
	})
}

// TestSubscribeNpuFaultEvent tests the subscribeNpuFaultEvent function.
func TestSubscribeNpuFaultEvent(t *testing.T) {
	hdm := &HwDevManager{manager: &device.HwAscend310Manager{}, RunMode: common.Ascend910}
	convey.Convey("Test subscribeNpuFaultEvent", t, func() {
		convey.Convey("When LoadFaultCodeFromFile fails, set SubscribeFailed and log error", func() {
			patch := gomonkey.ApplyFunc(common.LoadFaultCodeFromFile,
				func() error { return errors.New("load faultCode.json failed") })
			defer patch.Reset()
			hdm.subscribeNpuFaultEvent()
			convey.So(common.SubscribeFailed, convey.ShouldBeTrue)
		})
		patch1 := gomonkey.ApplyFunc(common.LoadFaultCodeFromFile, func() error { return nil })
		defer patch1.Reset()
		convey.Convey("When RunMode is not Ascend910, set SubscribeFailed and log debug", func() {
			hdm.RunMode = "otherMode"
			hdm.subscribeNpuFaultEvent()
			convey.So(common.SubscribeFailed, convey.ShouldBeTrue)
			hdm.RunMode = common.Ascend910
		})
		mockDmgr := gomonkey.ApplyMethodReturn(hdm.manager, "GetDmgr", &devmanager.DeviceManager{})
		defer mockDmgr.Reset()
		convey.Convey("When SetFaultEventCallFunc fails, set SubscribeFailed and log error", func() {
			patch2 := gomonkey.ApplyMethodReturn(hdm.manager.GetDmgr(), "SetFaultEventCallFunc",
				errors.New("set callback failed"))
			defer patch2.Reset()
			hdm.subscribeNpuFaultEvent()
			convey.So(common.SubscribeFailed, convey.ShouldBeTrue)
		})
		patch2 := gomonkey.ApplyMethodReturn(hdm.manager.GetDmgr(), "SetFaultEventCallFunc", nil)
		defer patch2.Reset()
		tmpSubscribeFailed := common.SubscribeFailed
		common.SubscribeFailed = false
		convey.Convey("When SubscribeDeviceFaultEvent succeeds, return directly", func() {
			patch3 := gomonkey.ApplyMethodReturn(hdm.manager.GetDmgr(), "SubscribeDeviceFaultEvent", nil)
			defer patch3.Reset()
			hdm.subscribeNpuFaultEvent()
			convey.So(common.SubscribeFailed, convey.ShouldBeFalse)
		})
		common.SubscribeFailed = tmpSubscribeFailed
		convey.Convey("When SubscribeDeviceFaultEvent fails after retries, set SubscribeFailed and log error", func() {
			patch3 := gomonkey.ApplyMethod(hdm.manager.GetDmgr(), "SubscribeDeviceFaultEvent",
				func(_ *devmanager.DeviceManager, _ int32) error { return errors.New("subscribe failed") })
			patch4 := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer patch3.Reset()
			defer patch4.Reset()
			hdm.subscribeNpuFaultEvent()
			convey.So(common.SubscribeFailed, convey.ShouldBeTrue)
		})
	})
}

// TestHotReset tests the hotReset function.
func TestHotReset(t *testing.T) {
	hdm := &HwDevManager{manager: &device.HwAscend310Manager{}, RunMode: common.Ascend910}
	npuDevice := &common.NpuDevice{DeviceName: "name", LogicID: 0}
	convey.Convey("Test hotReset", t, func() {
		patch := gomonkey.ApplyMethod(hdm.manager, "SetCardsInResetting",
			func(_ *device.HwAscend310Manager, _ int32, _ bool) {}).ApplyMethod(hdm.manager, "SetResetFailedTimes",
			func(_ *device.HwAscend310Manager, _ int32, _ int) {})
		defer patch.Reset()
		convey.Convey("When PollImmediate error log warn and return", func() {
			mockPollImmediate := gomonkey.ApplyFuncReturn(wait.PollImmediate, errors.New("error"))
			defer mockPollImmediate.Reset()
			hdm.hotReset(npuDevice)
		})
		mockPollImmediate := gomonkey.ApplyFuncReturn(wait.PollImmediate, nil)
		defer mockPollImmediate.Reset()
		convey.Convey("When PollImmediate return nil   hot rest success", func() {
			hdm.hotReset(npuDevice)
		})
	})
}
