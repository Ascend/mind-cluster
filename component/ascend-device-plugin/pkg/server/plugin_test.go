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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/metadata"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

var (
	devices = []*common.NpuDevice{
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-0", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-1", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-2", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-3", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-4", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-5", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-6", Health: "Healthy"},
		{DevType: api.Ascend910, DeviceName: api.Ascend910 + "-7", Health: "Healthy"},
	}
	mockPods = []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "test1", Namespace: "test1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test2", Namespace: "test2",
			Annotations: map[string]string{common.PodPredicateTime: "abcdef"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3", Annotations: map[string]string{common.
			PodPredicateTime: "1", api.HuaweiAscend910: api.Ascend910 + "-1"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test4", Namespace: "test4", Annotations: map[string]string{common.
			PodPredicateTime: "4", api.HuaweiAscend910: api.Ascend910 + "-2"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test5", Namespace: "test5", Annotations: map[string]string{common.
			PodPredicateTime: "5", api.ResourceNamePrefix + common.Ascend910vir2: api.Ascend910 + "-2c-180-3"}}},
	}
	fakeErr = errors.New("fake error")
)

const (
	mockPerfDumpPath       = "/root/a"
	mockPerfDumpConfig     = "step:true,time=4"
	slowNodeStepTimeEnvNum = 2
	intNum10               = 10
	intNum2                = 2
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
	common.ParamOption.PresetVDevice = true
}

type fakeGrpcStream struct{}

func (stream *fakeGrpcStream) SetHeader(md metadata.MD) error { return nil }

func (stream *fakeGrpcStream) SendHeader(md metadata.MD) error { return nil }

func (stream *fakeGrpcStream) SetTrailer(md metadata.MD) {}

func (stream *fakeGrpcStream) Context() context.Context { return context.Background() }

func (stream *fakeGrpcStream) SendMsg(m interface{}) error { return nil }

func (stream *fakeGrpcStream) RecvMsg(m interface{}) error { return nil }

func (stream *fakeGrpcStream) Send(*v1beta1.ListAndWatchResponse) error { return nil }

// TestListAndWatch for test the interface ListAndWatch
func TestListAndWatch(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, nil, nil, device.NewHwAscend910Manager())
	convey.Convey("test ListAndWatch", t, func() {
		mockSend := gomonkey.ApplyFunc(sendToKubelet, func(stream v1beta1.DevicePlugin_ListAndWatchServer,
			resp *v1beta1.ListAndWatchResponse) error {
			return nil
		})
		convey.Convey("Notify false", func() {
			ret := ps.Notify(devices)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("Notify true", func() {
			stream := fakeGrpcStream{}
			go ps.ListAndWatch(&v1beta1.Empty{}, &stream)
			time.Sleep(time.Second)
			ret := ps.Notify(devices)
			convey.So(ret, convey.ShouldBeTrue)
			convey.So(len(ps.cachedDevices), convey.ShouldEqual, len(devices))
			for i, id := range devices {
				convey.So(id.DeviceName, convey.ShouldEqual, ps.cachedDevices[i].DeviceName)
				convey.So(id.Health, convey.ShouldEqual, ps.cachedDevices[i].Health)
			}
			ps.stopListAndWatch()
		})
		mockSend.Reset()
	})
}

// TestUpdateAllocMap for test the updateAllocMap
func TestUpdateAllocMap(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, nil)
	convey.Convey("length no equal", t, func() {
		realAlloc := []string{api.Ascend910 + "-0", api.Ascend910 + "-2", api.Ascend910 + "-1"}
		kltAlloc := []string{api.Ascend910 + "-2", api.Ascend910 + "-7", api.Ascend910 + "-0", api.Ascend910 + "-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, 0)
	})
	convey.Convey("update map", t, func() {
		realAlloc := []string{api.Ascend910 + "-0", api.Ascend910 + "-2", api.Ascend910 + "-1", api.Ascend910 + "-3"}
		kltAlloc := []string{api.Ascend910 + "-2", api.Ascend910 + "-7", api.Ascend910 + "-0", api.Ascend910 + "-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, len(realAlloc))
		for i, id := range kltAlloc {
			v, exist := ps.klt2RealDevMap[id]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, realAlloc[i])
		}
	})
	convey.Convey("update duplicate device", t, func() {
		lastLength := len(ps.klt2RealDevMap)
		realAlloc := []string{"Ascend910-4"}
		kltAlloc := []string{"Ascend910-2"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, lastLength)
		for i, id := range kltAlloc {
			v, exist := ps.klt2RealDevMap[id]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, realAlloc[i])
		}
	})
}

// TestGenerateAllDeviceMap for test the generateAllDeviceMap
func TestGenerateAllDeviceMap(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, nil)
	convey.Convey("length no equal", t, func() {
		ps.deepCopyDevice(devices)
		realAlloc := []string{api.Ascend910 + "-0", api.Ascend910 + "-2", api.Ascend910 + "-1", api.Ascend910 + "-3"}
		kltAlloc := []string{api.Ascend910 + "-2", api.Ascend910 + "-7", api.Ascend910 + "-0", api.Ascend910 + "-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		expectMap := map[string]string{
			api.Ascend910 + "-4": api.Ascend910 + "-3", api.Ascend910 + "-5": api.Ascend910 + "-4",
			api.Ascend910 + "-6": api.Ascend910 + "-5", api.Ascend910 + "-7": api.Ascend910 + "-6",
		}
		actualMap := ps.generateAllDeviceMap()
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, len(expectMap))
		for k, v := range expectMap {
			id, exist := actualMap[k]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(id, convey.ShouldEqual, v)
		}
	})
}

// TestResponseToKubelet for test the responseToKubelet
func TestResponseToKubelet(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
	convey.Convey("use volcano", t, func() {
		common.ParamOption.UseVolcanoType = true
		ps.deepCopyDevice(devices)
		ps.klt2RealDevMap = map[string]string{
			api.Ascend910 + "-4": api.Ascend910 + "-3", api.Ascend910 + "-5": api.Ascend910 + "-4",
			api.Ascend910 + "-6": api.Ascend910 + "-5", api.Ascend910 + "-7": api.Ascend910 + "-6",
			api.Ascend910 + "-0": api.Ascend910 + "-7", api.Ascend910 + "-1": api.Ascend910 + "-2",
			api.Ascend910 + "-2": api.Ascend910 + "-1", api.Ascend910 + "-3": api.Ascend910 + "-0",
		}
		resp := ps.responseToKubelet()
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.Devices), convey.ShouldEqual, len(ps.cachedDevices))
		for i, id := range ps.cachedDevices {
			convey.So(id.DeviceName, convey.ShouldEqual, ps.klt2RealDevMap[resp.Devices[i].ID])
			convey.So(id.Health, convey.ShouldEqual, ps.cachedDevices[i].Health)
		}
	})
}

// TestAllocateRequestPhysicalDevice for test the Allocate request physical device
func TestAllocateRequestPhysicalDevice(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = false
	var requests v1beta1.AllocateRequest
	convey.Convey("invalid request", t, func() {
		mockGetNPUsFunc := mockGetNPUs()
		defer mockGetNPUsFunc.Reset()
		convey.Convey("input nil", func() {
			_, err := ps.Allocate(context.Background(), nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("container num exceeds the upper limit", func() {
			requests.ContainerRequests = make([]*v1beta1.ContainerAllocateRequest, common.MaxContainerLimit+1)
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("devices num exceeds the upper limit", func() {
			requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: make([]string,
				common.MaxDevicesNum+1)}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request physical device not exist", func() {
			ps.deepCopyDevice(devices)
			requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-8"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request physical device exist", func() {
			mockSlowNodeFunc := mockSetSlowNodeNoticeEnv()
			defer mockSlowNodeFunc.Reset()
			ps.deepCopyDevice(devices)
			deviceID := "1"
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{api.Ascend910 + "-" + deviceID}}}
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_VISIBLE_DEVICES"], convey.ShouldEqual, "")
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_RUNTIME_OPTIONS"], convey.ShouldBeEmpty)
		})
	})
}

// TestAllocateRequestVirtualDevice for test the Allocate request virtual device
func TestAllocateRequestVirtualDevice(t *testing.T) {
	common.ParamOption.UseVolcanoType = false
	ps := NewPluginServer(common.Ascend910vir2, devices, nil, device.NewHwAscend910Manager())
	var requests v1beta1.AllocateRequest
	convey.Convey("invalid request", t, func() {
		mockGetNPUsFunc := mockGetNPUs()
		defer mockGetNPUsFunc.Reset()
		convey.Convey("request more than 1 virtual device", func() {
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910vir2, DeviceName: "Ascend910-2c-100-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-2c-100-0", "Ascend910-2c-100-1"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request virtual device not exist", func() {
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910vir2, DeviceName: "Ascend910-2c-100-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-2c-100-1"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request virtual device exist", func() {
			mockSlowNodeFunc := mockSetSlowNodeNoticeEnv()
			defer mockSlowNodeFunc.Reset()
			deviceID := "100"
			ps := NewPluginServer(common.Ascend910vir2, devices, nil, device.NewHwAscend910Manager())
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910vir2,
				DeviceName: api.Ascend910 + "-2c-" + deviceID + "-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{api.Ascend910 + "-2c-" + deviceID + "-0"}}}
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs[api.AscendVisibleDevicesEnv], convey.ShouldEqual, "")
			convey.So(resp.ContainerResponses[0].Envs[api.AscendVisibleDevicesEnv], convey.ShouldEqual, "")
		})
	})
}

// TestAllocateWithVolcano1 for test the Allocate request physical device with volcano, not get valid oldest pod
func TestAllocateWithVolcano1(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-0"}}}
	convey.Convey("with volcano", t, func() {
		mockGetNPUsFunc := mockGetNPUs()
		defer mockGetNPUsFunc.Reset()
		convey.Convey("GetPodList failed", func() {
			mockActivePodList := mockGetActivePodListCache(nil)
			defer mockActivePodList.Reset()
			mockActivePod := mockGetActivePodList(nil, nil)
			defer mockActivePod.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("oldestPod is nil", func() {
			mockActivePodList := mockGetActivePodListCache(mockPods)
			defer mockActivePodList.Reset()
			mockActivePod := mockGetActivePodList(mockPods, nil)
			defer mockActivePod.Reset()
			mockFilter := mockFilterPods(nil)
			defer mockFilter.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAllocateWithVolcano2 for test the Allocate request physical device with volcano, get oldest pod
func TestAllocateWithVolcano2(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-0"}}}
	convey.Convey("test AllocateWithVolcano", t, func() {
		mockActivePodList := mockGetActivePodListCache(mockPods)
		defer mockActivePodList.Reset()
		mockGetNPUsFunc := mockGetNPUs()
		defer mockGetNPUsFunc.Reset()
		convey.Convey("TryUpdatePodAnnotation failed", func() {
			mockPodSlice := []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
				Annotations: map[string]string{common.PodPredicateTime: "5",
					api.HuaweiAscend910: api.Ascend910 + "-0"}}}}
			mockFilter := mockFilterPods(mockPodSlice)
			defer mockFilter.Reset()
			mockUpdatePod := mockTryUpdatePodAnnotation(fmt.Errorf("err"))
			defer mockUpdatePod.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("common.GetDeviceFromPodAnnotation failed", func() {
			mockPodSlice := []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
				Annotations: map[string]string{common.PodPredicateTime: "5",
					api.ResourceNamePrefix + common.Ascend910vir2: api.Ascend910 + "-2c-180-3"}}}}
			mockFilter := mockFilterPods(mockPodSlice)
			defer mockFilter.Reset()
			mockUpdatePod := mockTryUpdatePodAnnotation(nil)
			defer mockUpdatePod.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAllocateWithVolcano3 for test the Allocate request physical device with volcano, part 3
func TestAllocateWithVolcano3(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{api.Ascend910 + "-0"}}}
	convey.Convey("test AllocateWithVolcano", t, func() {
		mockActivePodList := mockGetActivePodListCache(mockPods)
		defer mockActivePodList.Reset()
		mockUpdatePod := mockTryUpdatePodAnnotation(nil)
		defer mockUpdatePod.Reset()
		mockSlowNodeFunc := mockSetSlowNodeNoticeEnv()
		defer mockSlowNodeFunc.Reset()
		mockGetNPUsFunc := mockGetNPUs()
		defer mockGetNPUsFunc.Reset()
		convey.Convey("with volcano GetDeviceListID failed", func() {
			mockPodSlice := []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
				Annotations: map[string]string{common.PodPredicateTime: "5",
					api.HuaweiAscend910: api.Ascend910}}}}
			mockFilter := mockFilterPods(mockPodSlice)
			defer mockFilter.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("with volcano run ok", func() {
			mockFilter := mockFilterPods(mockPods)
			defer mockFilter.Reset()
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs[api.AscendVisibleDevicesEnv], convey.ShouldEqual, "")
			_, err = ps.GetRealAllocateDevicesFromMap([]string{api.Ascend910 + "-2"})
			convey.So(err, convey.ShouldNotBeNil)
			realAllocate, err := ps.GetRealAllocateDevicesFromMap([]string{api.Ascend910 + "-0"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(realAllocate), convey.ShouldEqual, 1)
			convey.So(realAllocate[0], convey.ShouldEqual, api.Ascend910+"-1")
		})
	})
}

// TestSetSlowNodeNoticeEnv
func TestSetSlowNodeNoticeEnv(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	convey.Convey("test environment variable", t, func() {
		mockGetCM := mockGetCM()
		defer mockGetCM.Reset()
		resp := v1beta1.ContainerAllocateResponse{}
		resp.Envs = make(map[string]string, slowNodeStepTimeEnvNum)
		common.ParamOption.EnableSlowNode = true
		ps.SetSlowNodeNoticeEnv(&resp)
		convey.So(resp.Envs[common.PerfDumpPathEnv], convey.ShouldEqual, mockPerfDumpPath)
		convey.So(resp.Envs[common.PerfDumpConfigEnv], convey.ShouldEqual, mockPerfDumpConfig)
	})
}

// TestGetUnhealthyAICore for testGetUnhealthyAICore
func TestGetUnhealthyAICore(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	ps.klt2RealDevMap[api.Ascend910+"-0"] = api.Ascend910 + "-0"
	common.ParamOption.AiCoreCount = common.MinAICoreNum
	mockGetAiCore := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "GetRealUsedAICore",
		func(_ *PluginServer) (map[string]string, error) { return nil, nil })
	defer mockGetAiCore.Reset()
	convey.Convey("test GetUnhealthyAICore", t, func() {
		convey.Convey("GetUnhealthyAICore success", func() {
			unhealthyDev := ps.getUnhealthyAICore()
			convey.So(len(unhealthyDev), convey.ShouldEqual, 0)
		})
	})
}

// TestDestroyNotUsedVNPU for testDestroyNotUsedVNPU
func TestDestroyNotUsedVNPU(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	ps.klt2RealDevMap[api.Ascend910+"-0"] = api.Ascend910 + "-0"
	common.ParamOption.AiCoreCount = common.MinAICoreNum
	mockGetNPUsFunc := mockGetNPUs()
	mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "DestroyVirtualDevice",
		func(_ *device.AscendTools, _ string) error {
			return nil
		})
	mockAllocateDev := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "GetKltAndRealAllocateDev",
		func(_ *PluginServer, _ []v1.Pod) ([]*common.PodDeviceInfo, error) {
			return []*common.PodDeviceInfo{}, nil
		})
	mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetAllPodListCache",
		func(_ *kubeclient.ClientK8s) []v1.Pod {
			return []v1.Pod{}
		})
	defer mockPodList.Reset()
	defer mockDestroy.Reset()
	defer mockAllocateDev.Reset()
	defer mockGetNPUsFunc.Reset()
	convey.Convey("test DestroyNotUsedVNPU", t, func() {
		convey.Convey("DestroyNotUsedVNPU success", func() {
			err := ps.DestroyNotUsedVNPU()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDoWithVolcanoSchedule for testDoWithVolcanoSchedule
func TestDoWithVolcanoSchedule(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	devicesIDs := []string{""}
	podList := getMockPodList()
	common.ParamOption.PresetVDevice = false
	mockActivePodList := mockGetActivePodListCache(podList)
	mockUpdatePod := mockTryUpdatePodAnnotation(nil)
	mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
		func(_ *PluginServer) error {
			return nil
		})
	mockCreate := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "CreateVirtualDevice",
		func(_ *device.AscendTools, phyID int32, templateName string) (string, error) {
			return "Ascend910-2c-100-0", nil
		})
	defer mockCreate.Reset()
	defer mockDestroy.Reset()
	defer mockUpdatePod.Reset()
	defer mockActivePodList.Reset()
	convey.Convey("test DoWithVolcanoSchedule", t, func() {
		convey.Convey("DoWithVolcanoSchedule success", func() {
			_, err := ps.useVolcano(devicesIDs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
	common.ParamOption.PresetVDevice = true
}

// TestStrategyForSendStats for strategyForSendStats
func TestStrategyForSendStats(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	convey.Convey("test strategyForSendStats", t, func() {
		convey.Convey("case last send success, expect EmptyStrategy", func() {
			ps.deviceSyncStat.RecordSendResult(true)
			convey.So(ps.strategyForSendStats(), convey.ShouldEqual, common.EmptyStrategy)
		})
		convey.Convey("case send failure count >= threshold for reRegistry, expect ReRegistryStrategy",
			func() {
				for i := 0; i < intNum10/intNum2; i++ {
					ps.deviceSyncStat.RecordSendResult(false)
				}
				convey.So(ps.strategyForSendStats(), convey.ShouldEqual, common.ReRegistryStrategy)
			})
		convey.Convey("case send failure count >= threshold for restart, expect ReStartDevicePluginStrategy",
			func() {
				for i := 0; i < intNum10; i++ {
					ps.deviceSyncStat.RecordSendResult(false)
				}
				convey.So(ps.strategyForSendStats(), convey.ShouldEqual, common.ReStartDevicePluginStrategy)
			})
	})
}

// func (ps *PluginServer) responseToKubelet() *v1beta1.ListAndWatchResponse
// TestReportDeviceInfo for reportDeviceInfo
func TestReportDeviceInfo(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	convey.Convey("test reportDeviceInfo", t, func() {
		convey.Convey("case sendToKubelet success, expect last send success", func() {
			patch := gomonkey.ApplyFuncReturn(sendToKubelet, nil).
				ApplyPrivateMethod(reflect.TypeOf(ps), "responseToKubelet", func() *v1beta1.ListAndWatchResponse {
					return nil
				})
			defer patch.Reset()
			ps.reportDeviceInfo(nil)
			convey.So(ps.deviceSyncStat.GetLastSendStatus(), convey.ShouldBeTrue)
		})
		convey.Convey("case sendToKubelet failed, expect last send failed", func() {
			patch := gomonkey.ApplyFuncReturn(sendToKubelet, fakeErr).
				ApplyPrivateMethod(reflect.TypeOf(ps), "responseToKubelet", func() *v1beta1.ListAndWatchResponse {
					return nil
				})
			defer patch.Reset()
			ps.reportDeviceInfo(nil)
			convey.So(ps.deviceSyncStat.GetLastSendStatus(), convey.ShouldBeFalse)
		})
	})
}

// TestHandleConsecutiveErrorStrategy for handleConsecutiveErrorStrategy
func TestHandleConsecutiveErrorStrategy(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	ps.isRunning.Store(true)
	convey.Convey("test handleConsecutiveErrorStrategy", t, func() {
		convey.Convey("case restart device plugin, expect isRunning=false",
			func() {
				patch := gomonkey.ApplyFuncReturn(exitSelfProcess, nil)
				defer patch.Reset()
				ps.handleConsecutiveErrorStrategy(common.ReStartDevicePluginStrategy, 0)
				convey.So(ps.isRunning.Load(), convey.ShouldBeFalse)
			})
		ps.isRunning.Store(true)
		convey.Convey("case reRegistry strategy, isRunning=false",
			func() {
				ps.handleConsecutiveErrorStrategy(common.ReRegistryStrategy, 0)
				convey.So(ps.isRunning.Load(), convey.ShouldBeFalse)
			})
		ps.isRunning.Store(true)
		convey.Convey("case empty strategy, isRunning=true",
			func() {
				ps.handleConsecutiveErrorStrategy(common.EmptyStrategy, 0)
				convey.So(ps.isRunning.Load(), convey.ShouldBeTrue)
			})
	})
}

type getRealAllocateDevicesFromEnvTestCase struct {
	Name    string
	pod     v1.Pod
	WantDev []string
}

func buildGetRealAllocateDevicesFromEnvTestCases() []getRealAllocateDevicesFromEnvTestCase {
	fieldPath := fmt.Sprintf("%s['%s%s']",
		common.MetaDataAnnotation, api.ResourceNamePrefix, api.Ascend910)
	annotationTag := fmt.Sprintf("%s%s", api.ResourceNamePrefix, api.Ascend910)
	return []getRealAllocateDevicesFromEnvTestCase{
		{
			Name:    "01-containers len is zero, should return nil",
			pod:     v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{}}},
			WantDev: nil,
		},
		{
			Name:    "02-all env is empty, should return nil",
			pod:     v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Env: []v1.EnvVar{}}}}},
			WantDev: nil,
		},
		{
			Name: "03-get device from pod annotation failed, should return nil",
			pod: v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{{Env: []v1.EnvVar{
					{Name: "fakeName", ValueFrom: nil},
					{Name: common.AscendVisibleDevicesEnv,
						ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "fakePath"}},
					},
					{Name: common.AscendVisibleDevicesEnv,
						ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: fieldPath}},
					},
				}}}},
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			},
			WantDev: nil,
		},
		{
			Name: "04-get real dev from env success, should return devices",
			pod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{
				{Env: []v1.EnvVar{
					{Name: common.AscendVisibleDevicesEnv,
						ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: fieldPath}},
					},
				}}}},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationTag: "0,1"},
				}},
			WantDev: []string{"0", "1"},
		},
	}
}

// TestGetRealAllocateDevicesFromEnv for test GetRealAllocateDevicesFromEnv
func TestGetRealAllocateDevicesFromEnv(t *testing.T) {
	testCases := buildGetRealAllocateDevicesFromEnvTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			ps := NewPluginServer(api.Ascend910, devices, []string{common.HiAIManagerDevice},
				device.NewHwAscend910Manager())
			deviceList := ps.GetRealAllocateDevicesFromEnv(tt.pod)
			if !reflect.DeepEqual(deviceList, tt.WantDev) {
				t.Errorf("GetRealAllocateDevicesFromEnv() Devices = %v, WantDevices = %v",
					deviceList, tt.WantDev)
			}
		})
	}
}

func getMockPodList() []v1.Pod {
	return []v1.Pod{
		getMockPod(),
	}
}

func mockGetCM() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetConfigMap", func(_ *kubeclient.ClientK8s, _ string, _ string) (*v1.ConfigMap, error) {
			nodeCMData := stepTimeCM{
				Data: stepTimeData{
					PerfDumpPath:   mockPerfDumpPath,
					PerfDumpConfig: mockPerfDumpConfig,
				},
			}
			return &v1.ConfigMap{Data: map[string]string{
				common.SlowNodeNoticeCMName: string(common.MarshalData(nodeCMData)),
			},
			}, nil
		})
}

func mockSetSlowNodeNoticeEnv() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)),
		"SetSlowNodeNoticeEnv", func(_ *PluginServer, _ *v1beta1.ContainerAllocateResponse) {
			return
		})
}

func mockGetNPUs() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend910Manager)), "GetNPUs",
		func(_ *device.HwAscend910Manager) (common.NpuAllInfo, error) {
			return common.NpuAllInfo{}, nil
		})
}

func mockGetActivePodListCache(mockPods []v1.Pod) *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodListCache",
		func(_ *kubeclient.ClientK8s) []v1.Pod { return mockPods })
}

func mockTryUpdatePodAnnotation(err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "TryUpdatePodAnnotation",
		func(_ *kubeclient.ClientK8s, _ *v1.Pod, _ map[string]string) error { return err })
}

func mockGetActivePodList(mockPods []v1.Pod, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
		func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) { return mockPods, err })
}

func mockFilterPods(mockPods []v1.Pod) *gomonkey.Patches {
	return gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
		conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
		return mockPods
	})
}

const (
	virDevType     = api.Ascend910 + "-16c"
	devType        = api.Ascend910 + "-16"
	realResNameVir = api.ResourceNamePrefix + virDevType
	realResName    = api.ResourceNamePrefix + devType
)

type getKltAndRealAllocateDevArgs struct {
	mockPodDevice map[string]PodDevice
	mockErr       error
	podList       []v1.Pod
	deviceType    string
}

func getFakePodList() []v1.Pod {
	return []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "pod1",
			Annotations: map[string]string{api.PodAnnotationAscendReal: "0,1,2,3"}}},
		{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "pod2",
			Annotations: map[string]string{api.PodAnnotationAscendReal: "4,5,6,7"}}},
		{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "pod3",
			Annotations: map[string]string{}}},
	}
}

// getKltAndRealAllocateDevTestCase GetKltAndRealAllocateDev test case
type getKltAndRealAllocateDevTestCase struct {
	Name        string
	args        getKltAndRealAllocateDevArgs
	wantRealDev []string
	wantErr     error
}

func buildGetKltAndRealAllocateDevTestCaseTestCases() []getKltAndRealAllocateDevTestCase {
	podList := getFakePodList()
	return []getKltAndRealAllocateDevTestCase{
		{
			Name: "01-get pod resource failed, should return empty pod device info and error",
			args: getKltAndRealAllocateDevArgs{mockErr: fakeErr,
				podList: []v1.Pod{}, mockPodDevice: map[string]PodDevice{}, deviceType: virDevType},
			wantRealDev: nil,
			wantErr:     errors.New("get pod resource failed, fake error"),
		},
		{
			Name: "02-get virtual dev info success, should return pod virtual device info and nil",
			args: getKltAndRealAllocateDevArgs{mockErr: nil, podList: podList, deviceType: virDevType,
				mockPodDevice: map[string]PodDevice{"ns_pod1": {ResourceName: "fakeName", DeviceIds: []string{"0"}},
					"ns_pod2": {ResourceName: realResNameVir, DeviceIds: []string{"4"}}}},
			wantRealDev: []string{"4"},
			wantErr:     nil,
		},
		{
			Name: "03-get dev info success, should return pod device info and nil",
			args: getKltAndRealAllocateDevArgs{mockErr: nil, podList: podList, deviceType: devType,
				mockPodDevice: map[string]PodDevice{"ns_pod1": {ResourceName: realResName, DeviceIds: []string{"0"}},
					"ns_pod3": {ResourceName: realResName, DeviceIds: []string{"8"}}}},
			wantRealDev: []string{"0", "1", "2", "3"},
			wantErr:     nil,
		},
	}
}

func TestGetKltAndRealAllocateDev(t *testing.T) {
	testCases := buildGetKltAndRealAllocateDevTestCaseTestCases()
	patch := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{PresetVDevice: true}).
		ApplyPrivateMethod(&PluginServer{}, "updateAllocMap", func(*PluginServer, []string, []string) {}).
		ApplyMethod(&PluginServer{}, "GetRealAllocateDevicesFromMap",
			func(*PluginServer, []string) ([]string, error) { return nil, fakeErr }).
		ApplyMethod(&PluginServer{}, "GetRealAllocateDevicesFromEnv",
			func(*PluginServer, v1.Pod) []string { return nil })
	defer patch.Reset()

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			ps := NewPluginServer(tt.args.deviceType, devices, []string{common.HiAIManagerDevice},
				device.NewHwAscend910Manager())
			patch1 := gomonkey.ApplyMethodReturn(&PodResource{}, "GetPodResource",
				tt.args.mockPodDevice, tt.args.mockErr)
			info, err := ps.GetKltAndRealAllocateDev(tt.args.podList)
			patch1.Reset()
			if len(info) == 0 && len(tt.wantRealDev) > 0 {
				t.Error("GetKltAndRealAllocateDev() failed")
			}
			if len(info) > 0 && !reflect.DeepEqual(info[0].RealDevice, tt.wantRealDev) {
				t.Errorf("GetKltAndRealAllocateDev() realDev = %v, "+
					"wantRealDev = %v", info[0].RealDevice, tt.wantRealDev)
			}
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetKltAndRealAllocateDev() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetPreferredAllocation test get preferred allocation
func TestGetPreferredAllocation(t *testing.T) {
	convey.Convey("Test GetPreferredAllocation", t, func() {
		ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
		_, err := ps.GetPreferredAllocation(nil, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetDevicePluginOptions(t *testing.T) {
	convey.Convey("Test GetDevicePluginOptions", t, func() {
		ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
		_, err := ps.GetDevicePluginOptions(nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPreStartContainer(t *testing.T) {
	convey.Convey("Test PreStartContainer", t, func() {
		ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
		_, err := ps.PreStartContainer(nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
