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

// Package kubeclient a series of k8s function ut
package kubeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

const (
	nodeNameKey     = "NODE_NAME"
	nodeNameValue   = "master"
	invalidNodeName = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz" +
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn" +
		"opqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"

	npuChip310PhyID0  = "Ascend310-0"
	npuChip910PhyID0  = "Ascend910-0"
	npuChip310PPhyID0 = "Ascend310P-0"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func initK8S() (*ClientK8s, error) {
	return &ClientK8s{}, nil
}

// TestAnnotationReset test device info reset
func TestAnnotationReset(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestAnnotationReset init kubernetes failed")
	}
	common.ParamOption.AutoStowingDevs = true
	convey.Convey("annotation reset with no error", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(nil, nil, nil)
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("annotation reset with no error", t, func() {
		mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
			func(_ *ClientK8s) (*v1.Node, error) {
				return nil, nil
			})
		defer mockNode.Reset()
		err := utKubeClient.AnnotationReset()
		convey.So(err.Error(), convey.ShouldEqual, "invalid node")
	})
	convey.Convey("annotation reset with error", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(fmt.Errorf("can not found device info cm"),
			fmt.Errorf("patch node state failed"), nil)
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err.Error(), convey.ShouldEqual, "patch node state failed")
	})
	convey.Convey("annotation reset with get node failed", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(nil, nil, fmt.Errorf("get node failed"))
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err.Error(), convey.ShouldEqual, "get node failed")
	})
}

// TestGetNodeServerID test get node server id
func TestGetNodeServerID(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestGetNodeServerID init kubernetes failed")
	}
	node := getMockNode(common.HuaweiAscend910, npuChip910PhyID0)
	convey.Convey("get node server id without get node", t, func() {
		mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
			func(_ *ClientK8s) (*v1.Node, error) {
				return nil, fmt.Errorf("failed to get node")
			})
		defer mockNode.Reset()
		_, err := utKubeClient.GetNodeServerID()
		convey.So(err.Error(), convey.ShouldEqual, "failed to get node")
	})
	convey.Convey("get node server id", t, func() {
		mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
			func(_ *ClientK8s) (*v1.Node, error) {
				return node, nil
			})
		defer mockNode.Reset()
		serverID, err := utKubeClient.GetNodeServerID()
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(serverID, convey.ShouldEqual, common.DefaultDeviceIP)
	})
}

// TestGetPodsUsedNpu test used npu devices on pod
func TestGetPodsUsedNpu(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestGetPodsUsedNpu init kubernetes failed")
	}
	podList := getMockPodList(common.ResourceNamePrefix+common.PodRealAlloc, npuChip310PhyID0)
	convey.Convey("get used npu on pods without get pod list", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetActivePodListCache",
			func(_ *ClientK8s) []v1.Pod {
				return nil
			})
		defer mockPodList.Reset()
		useNpu := utKubeClient.GetPodsUsedNpu()
		convey.So(useNpu, convey.ShouldEqual, sets.String{})
	})
	convey.Convey("get used npu on pods", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetActivePodListCache",
			func(_ *ClientK8s) []v1.Pod {
				return podList
			})
		defer mockPodList.Reset()
		useNpu := utKubeClient.GetPodsUsedNpu()
		convey.So(strings.Join(useNpu.List(), ","), convey.ShouldEqual, npuChip310PhyID0)
	})
}

// TestWriteDeviceInfoDataIntoCM get cm write operation
func TestWriteDeviceInfoDataIntoCM(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestWriteDeviceInfoDataIntoCM init kubernetes failed")
	}
	updateCM := getMockCreateCM(common.HuaweiAscend310P, npuChip310PPhyID0)
	deviceInfo := getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0)
	mockCreateCM, mockUpdateCM := mockCMOpr(updateCM)
	defer resetMock(mockCreateCM, mockUpdateCM)
	convey.Convey("write device info (cm) when marshal node device data failed", t, func() {
		mockMarshalData := gomonkey.ApplyFuncReturn(common.MarshalData, []byte{})
		defer mockMarshalData.Reset()
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(deviceInfo, "", common.SwitchFaultInfo{}, -1, -1)
		convey.So(err.Error(), convey.ShouldEqual, "marshal nodeDeviceData failed")
	})
	convey.Convey("write device info (cm) when real card type is Ascend910A3", t, func() {
		mockRealCardType := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = common.Ascend910A3
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(deviceInfo, "", common.SwitchFaultInfo{}, -1, -1)
		common.ParamOption.RealCardType = mockRealCardType
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("get write device info (cm) when get cm success", t, func() {
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(deviceInfo, "", common.SwitchFaultInfo{}, -1, -1)
		convey.So(err, convey.ShouldEqual, nil)
	})
	mockIsNotFound := gomonkey.ApplyFuncReturn(errors.IsNotFound, true)
	defer mockIsNotFound.Reset()
	convey.Convey("write device info (cm) when create cm success", t, func() {
		mockCreateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "CreateConfigMap",
			func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) { return nil, nil })
		defer mockCreateCM.Reset()
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(deviceInfo, "", common.SwitchFaultInfo{}, -1, -1)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("write device info (cm) when update cm error", t, func() {
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(deviceInfo, "", common.SwitchFaultInfo{}, -1, -1)
		convey.So(err.Error(), convey.ShouldEqual, "unable to create configmap, already exists")
	})
}

// TestWriteResetInfoDataIntoCM get cm write reset info
func TestWriteResetInfoDataIntoCM(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestWriteResetInfoDataIntoCM init kubernetes failed")
	}
	oldCM := getMockCreateCM(common.ResetInfoCMDataKey, common.ResetInfoCMNamePrefix+"node")
	defer oldCM.Reset()
	testPod := getMockPod(common.HuaweiAscend910, npuChip910PhyID0)
	defer testPod.Reset()
	testTaskResetInfo := getTaskResetInfo()
	mockGetCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) {
			return oldCM, nil
		})
	defer mockGetCM.Reset()
	convey.Convey("write reset info when invalid reset info data", t, func() {
		_, err := utKubeClient.WriteResetInfoDataIntoCM("taskName", testPod.Namespace, testTaskResetInfo, true)
		convey.So(err.Error(), convey.ShouldEqual,
			"failed to unmarshal reset info data, err: invalid character 'r' looking for beginning of value")
	})
	mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, nil)
	defer mockUnmarshal.Reset()
	convey.Convey("write reset info when marshal task reset data failed", t, func() {
		mockMarshalData := gomonkey.ApplyFuncReturn(common.MarshalData, nil)
		defer mockMarshalData.Reset()
		_, err := utKubeClient.WriteResetInfoDataIntoCM("taskName", testPod.Namespace, testTaskResetInfo, true)
		convey.So(err.Error(), convey.ShouldEqual, "marshal task reset data failed")
	})
	convey.Convey("write reset info when update cm", t, func() {
		mockUpdateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "UpdateConfigMap",
			func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
				return oldCM, nil
			})
		defer mockUpdateCM.Reset()
		mockMarshalData := gomonkey.ApplyFuncReturn(common.MarshalData, []byte{1})
		defer mockMarshalData.Reset()
		_, err := utKubeClient.WriteResetInfoDataIntoCM("taskName", testPod.Namespace, testTaskResetInfo, true)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestWriteFaultInfoDataIntoCM get cm write fault info
func TestWriteFaultInfoDataIntoCM(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestWriteFaultInfoDataIntoCM init kubernetes failed")
	}
	oldCM := getMockCreateCM(common.ResetInfoCMDataKey, common.ResetInfoCMNamePrefix+"node")
	defer oldCM.Reset()
	testPod := getMockPod(common.HuaweiAscend910, npuChip910PhyID0)
	defer testPod.Reset()
	testTaskFaultInfo := getTaskFaultInfo()
	mockGetCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) {
			return nil, fmt.Errorf("failed to get fault cm")
		})
	defer mockGetCM.Reset()
	convey.Convey("write fault info when failed to get fault cm", t, func() {
		_, err := utKubeClient.WriteFaultInfoDataIntoCM("taskName", testPod.Namespace, testTaskFaultInfo)
		convey.So(err.Error(), convey.ShouldEqual, "failed to get fault cm")
	})
	mockGetCM = gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) {
			return oldCM, nil
		})
	defer mockGetCM.Reset()
	mockMakeDataHash := gomonkey.ApplyFuncReturn(common.MakeDataHash, "testCheckCode")
	defer mockMakeDataHash.Reset()
	convey.Convey("write fault info when marshal task reset data failed", t, func() {
		mockMarshalData := gomonkey.ApplyFuncReturn(common.MarshalData, nil)
		defer mockMarshalData.Reset()
		_, err := utKubeClient.WriteFaultInfoDataIntoCM("taskName", testPod.Namespace, testTaskFaultInfo)
		convey.So(err.Error(), convey.ShouldEqual, "marshal task reset data failed")
	})
	convey.Convey("write fault info when update cm", t, func() {
		mockUpdateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "UpdateConfigMap",
			func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
				return oldCM, nil
			})
		defer mockUpdateCM.Reset()
		mockMarshalData := gomonkey.ApplyFuncReturn(common.MarshalData, []byte{1})
		defer mockMarshalData.Reset()
		_, err := utKubeClient.WriteFaultInfoDataIntoCM("taskName", testPod.Namespace, testTaskFaultInfo)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestTryUpdatePodAnnotation try update pod annotation
func TestTryUpdatePodAnnotation(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestTryUpdatePodAnnotation init kubernetes failed")
	}
	testPod := getMockPod(common.HuaweiAscend910, npuChip910PhyID0)
	annotation := getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0)
	defer testPod.Reset()
	mockPatchPod := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "PatchPod",
		func(_ *ClientK8s, _ *v1.Pod, _ []byte) (*v1.Pod, error) {
			return nil, fmt.Errorf("test function errors")
		})
	defer mockPatchPod.Reset()
	convey.Convey("try update pod annotation when get pod is nil", t, func() {
		err := utKubeClient.TryUpdatePodAnnotation(nil, annotation)
		convey.So(err.Error(), convey.ShouldEqual, "param pod is nil")
	})
	convey.Convey("try update pod annotation when get invalid annotation", t, func() {
		err := utKubeClient.TryUpdatePodAnnotation(testPod, nil)
		convey.So(err.Error(), convey.ShouldEqual, "invalid annotation")
	})
	convey.Convey("try update pod annotation when get pod is not nil", t, func() {
		err := utKubeClient.TryUpdatePodAnnotation(testPod, annotation)
		convey.So(err.Error(), convey.ShouldEqual, "patch pod annotation failed, exceeded max number of retries")
	})
	convey.Convey("try update pod annotation when failed to marshal error", t, func() {
		mockMarshal := gomonkey.ApplyFuncReturn(json.Marshal, []byte{0}, fmt.Errorf("marshal error"))
		defer mockMarshal.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod, annotation)
		convey.So(err.Error(), convey.ShouldEqual, "marshal error")
	})
	convey.Convey("try update pod annotation success", t, func() {
		mockPatchPod := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "PatchPod",
			func(_ *ClientK8s, _ *v1.Pod, _ []byte) (*v1.Pod, error) {
				return nil, nil
			})
		defer mockPatchPod.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod, annotation)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("try update pod annotation when patch pod error not found", t, func() {
		mockIsNotFound := gomonkey.ApplyFuncReturn(errors.IsNotFound, true)
		defer mockIsNotFound.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod, annotation)
		convey.So(err.Error(), convey.ShouldEqual, "test function errors")
	})
}

// TestTryUpdatePodCacheAnnotation try update pod annotation in both api server and cache
func TestTryUpdatePodCacheAnnotation(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestTryUpdatePodCacheAnnotation init kubernetes failed")
	}
	testPod := getMockPod(common.HuaweiAscend910, npuChip910PhyID0)
	mockPatchPod := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "PatchPod",
		func(_ *ClientK8s, _ *v1.Pod, _ []byte) (*v1.Pod, error) {
			return nil, fmt.Errorf("test function errors")
		})
	defer mockPatchPod.Reset()
	convey.Convey("try update pod annotation when get pod is nil", t, func() {
		err := utKubeClient.TryUpdatePodCacheAnnotation(nil, getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err.Error(), convey.ShouldEqual, "param pod is nil")
	})
	convey.Convey("try update pod annotation when update pod annotation in api server failed", t, func() {
		err := utKubeClient.TryUpdatePodCacheAnnotation(testPod, getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err.Error(), convey.ShouldEqual, "patch pod annotation failed, exceeded max number of retries")
	})
	mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "TryUpdatePodAnnotation",
		func(_ *ClientK8s, _ *v1.Pod, _ map[string]string) error { return nil })
	defer mockTryUpdatePodAnnotation.Reset()
	convey.Convey("try update pod annotation when update annotation in pod cache success", t, func() {
		podCache = map[types.UID]*podInfo{
			"xxxxxxxxx1": {
				Pod:        testPod,
				updateTime: time.Now(),
			},
		}
		err := utKubeClient.TryUpdatePodCacheAnnotation(testPod, getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err, convey.ShouldBeNil)
		podCache = make(map[types.UID]*podInfo)
	})
	convey.Convey("try update pod annotation when no pod found in cache", t, func() {
		err := utKubeClient.TryUpdatePodCacheAnnotation(testPod, getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetDeviceInfoManuallySeparateNPUData returns the ManuallySeparateNPU from device info
func TestGetDeviceInfoManuallySeparateNPUData(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestGetDeviceInfoManuallySeparateNPUData init kubernetes failed")
	}
	mockCreateCM := getMockCreateCM(common.DeviceInfoCMManuallySeparateNPUKey, common.DeviceInfoCMNamePrefix+"node")
	defer mockCreateCM.Reset()
	mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) { return mockCreateCM, nil })
	defer mockGetConfigMap.Reset()
	convey.Convey("failed to get cm", t, func() { getCMError(utKubeClient) })
	convey.Convey("failed when cm ascendType is not ManuallySeparateNPU", t, func() { ascendTypeIsNil(utKubeClient) })
	convey.Convey("failed to get device run mode", t, func() {
		phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
			common.DeviceInfoCMNameSpace)
		convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
	})
	mockGetDeviceRunMode := gomonkey.ApplyFuncReturn(common.GetDeviceRunMode, common.Ascend910, nil)
	defer mockGetDeviceRunMode.Reset()
	convey.Convey("failed when npu cache is empty", t, func() { npuCacheIsEmpty(utKubeClient) })
	convey.Convey("manuallySeparateNPU will be ignored", t, func() {
		phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
			common.DeviceInfoCMNameSpace)
		convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
	})
	mockCheckDeviceName := gomonkey.ApplyFuncReturn(common.CheckDeviceName, true)
	defer mockCheckDeviceName.Reset()
	convey.Convey("failed to convert string phyIDStr type to int type", t, func() {
		phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
			common.DeviceInfoCMNameSpace)
		convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
	})
	convey.Convey("get device info manually separate npu success", t, func() { appendPhyIdSuccess(utKubeClient) })
}

func getCMError(utKubeClient *ClientK8s) {
	mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) { return nil, fmt.Errorf("get cm error") })
	defer mockGetConfigMap.Reset()
	phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
		common.DeviceInfoCMNameSpace)
	convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
}

func ascendTypeIsNil(utKubeClient *ClientK8s) {
	mockCreateCM := &v1.ConfigMap{
		Data:       make(map[string]string),
		ObjectMeta: metav1.ObjectMeta{Name: "testName"},
	}
	mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) { return mockCreateCM, nil })
	defer mockGetConfigMap.Reset()
	phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
		common.DeviceInfoCMNameSpace)
	convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
}

func npuCacheIsEmpty(utKubeClient *ClientK8s) {
	mockCreateCM := &v1.ConfigMap{
		Data: map[string]string{common.DeviceInfoCMManuallySeparateNPUKey: ""},
	}
	mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
		func(_ *ClientK8s, _ string, _ string) (*v1.ConfigMap, error) { return mockCreateCM, nil })
	defer mockGetConfigMap.Reset()
	phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
		common.DeviceInfoCMNameSpace)
	convey.So(phyIDs, convey.ShouldResemble, make([]int32, 0))
}

func appendPhyIdSuccess(utKubeClient *ClientK8s) {
	mockAtoI := gomonkey.ApplyFuncReturn(strconv.Atoi, 1, nil)
	defer mockAtoI.Reset()
	phyIDs := utKubeClient.GetManuallySeparateNPUIDFromDeviceInfo(utKubeClient.DeviceInfoName,
		common.DeviceInfoCMNameSpace)
	convey.So(phyIDs, convey.ShouldResemble, []int32{1})
}

func getMockCreateCM(ascendType, ascendValue string) *v1.ConfigMap {
	return &v1.ConfigMap{
		Data: map[string]string{
			ascendType: ascendValue,
		},
	}
}

func getDeviceInfo(ascendType, ascendValue string) map[string]string {
	return map[string]string{
		ascendType: ascendValue,
	}
}

func getMockPod(ascendType, ascendValue string) *v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	annotations["predicate-time"] = "1626785193048251590"
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-dls-npu-1p-default-2p-0",
			Namespace:   "btg-test",
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						common.HuaweiAscend910: resource.Quantity{},
					},
				}},
			},
		},
		Status: v1.PodStatus{
			Reason: "UnexpectedAdmissionError",
			ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}},
			},
		},
	}
}

func getMockNode(ascendType, ascendValue string) *v1.Node {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	labels := make(map[string]string, 1)
	labels[common.HuaweiRecoverAscend910] = "0"
	return &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceName(ascendType): resource.Quantity{},
			},
			Addresses: getAddresses(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Labels:      labels,
		},
	}
}

func getAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    v1.NodeHostName,
			Address: common.DefaultDeviceIP,
		},
		{
			Type:    v1.NodeInternalIP,
			Address: common.DefaultDeviceIP,
		},
	}
}

func getMockPodList(devType, ascendValue string) []v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[devType] = ascendValue
	annotations[common.PodPredicateTime] = strconv.FormatUint(math.MaxUint64, common.BaseDec)
	containers := getContainers(devType)
	return []v1.Pod{
		getPodUTOne(annotations, containers),
		getPodUTTwo(annotations),
		getPodUTThree(),
	}
}

func getPodUTOne(annotations map[string]string, containers []v1.Container) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-1",
			Namespace:   "btg-test1",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
	}
}

func getPodUTTwo(annotations map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-2",
			Namespace:   "btg-test2",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
}

func getPodUTThree() v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[common.HuaweiAscend310] = ""
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-3",
			Namespace:   "btg-test3",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
}

func getContainers(devType string) []v1.Container {
	limits := resource.NewQuantity(1, resource.DecimalExponent)
	container := v1.Container{
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceName(devType): *limits,
			},
		},
	}
	return []v1.Container{
		container,
	}
}

func getTaskFaultInfo() *common.TaskFaultInfo {
	return &common.TaskFaultInfo{
		FaultRank:  []int{0},
		UpdateTime: time.Now().Unix(),
	}
}

func getTaskResetInfo() *common.TaskResetInfo {
	rankList := []*common.TaskDevInfo{
		{
			RankId:       0,
			DevFaultInfo: mockResetErrDevFaultInfo(0),
		},
		{
			RankId:       1,
			DevFaultInfo: mockEmptyErrDevFaultInfo(1),
		},
	}
	return &common.TaskResetInfo{
		RankList: rankList,
	}
}

// mockResetErrDevFaultInfo create a fake dev fault info with reset error
func mockResetErrDevFaultInfo(logicId int32) common.DevFaultInfo {
	return common.DevFaultInfo{
		LogicId:       logicId,
		Status:        common.UnrecoveredStatus,
		Policy:        common.ResetError,
		InitialPolicy: common.ResetError,
		ErrorCode:     []int64{2350927360},
		ErrorCodeHex:  "0x8C204E00",
	}
}

// mockEmptyErrDevFaultInfo create a fake dev fault info with empty error
func mockEmptyErrDevFaultInfo(logicId int32) common.DevFaultInfo {
	return common.DevFaultInfo{
		LogicId:       logicId,
		Status:        common.UnrecoveredStatus,
		Policy:        common.EmptyError,
		InitialPolicy: common.EmptyError,
		ErrorCode:     []int64{},
		ErrorCodeHex:  "",
	}
}
func mockCMOpr(updateCM *v1.ConfigMap) (*gomonkey.Patches, *gomonkey.Patches) {
	mockCreateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "CreateConfigMap",
		func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
			return nil, fmt.Errorf("already exists")
		})
	mockUpdateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "UpdateConfigMap",
		func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
			return updateCM, nil
		})
	return mockCreateCM, mockUpdateCM
}

func resetMock(resetMockList ...*gomonkey.Patches) {
	for _, resetMock := range resetMockList {
		resetMock.Reset()
	}
}

func annotationResetMock(devErr, stateErr, nodeErr error) (*gomonkey.Patches, *gomonkey.Patches, *gomonkey.Patches) {
	node := getMockNode(common.HuaweiAscend910, npuChip910PhyID0)
	mockWrite := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "WriteDeviceInfoDataIntoCM",
		func(_ *ClientK8s, _ map[string]string, _ string, _ common.SwitchFaultInfo, _, _ int32) (
			*common.NodeDeviceInfoCache, error) {
			return nil, devErr
		})
	mockPatchNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "PatchNodeState",
		func(_ *ClientK8s, _ *v1.Node, _ *v1.Node) (*v1.Node, []byte, error) {
			return nil, nil, stateErr
		})
	mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
		func(_ *ClientK8s) (*v1.Node, error) {
			return node, nodeErr
		})
	return mockWrite, mockPatchNode, mockNode
}
