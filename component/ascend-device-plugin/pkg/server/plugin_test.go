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
	"io/fs"
	"os"
	"reflect"
	"strings"
	"sync"
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
	"ascend-common/devmanager"
	"ascend-common/devmanager/dcmi"
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
	mockAllNpuInfo = common.NpuAllInfo{
		AllDevs: []common.NpuDevice{
			{DevType: api.Ascend910, DeviceName: "Ascend910-0", Health: "Healthy"},
			{DevType: api.Ascend910, DeviceName: "Ascend910-1", Health: "Healthy"},
		},
		AICoreDevs: []*common.NpuDevice{
			{DevType: api.Ascend910, DeviceName: "Ascend910-0", Health: "Healthy"},
			{DevType: api.Ascend910, DeviceName: "Ascend910-1", Health: "Healthy"},
		},
		AllDevTypes: []string{api.Ascend910},
	}
	mockUBDevices = []os.DirEntry{
		FakeDirEntry{name: "udma1", typ: fs.ModeDevice},
		FakeDirEntry{name: "udma2", typ: fs.ModeDevice},
		FakeDirEntry{name: "udma3", typ: fs.ModeDevice},
	}
	fakeErr = errors.New("fake error")
)

const (
	mockPerfDumpPath       = "/root/a"
	mockPerfDumpConfig     = "step:true,time=4"
	slowNodeStepTimeEnvNum = 2
	intNum10               = 10
	intNum2                = 2
	mockValidVNPUId        = "vnpu-001"
	mockValidAicoreQuota   = "50"
	mockValidHbmQuota      = "2048"
	mockValidPolicyStr     = "Elastic"
	mockConvertedPolicy    = "2"
	testPhysicalID         = 0
	testJobName            = "test-job"
	testConfigDir          = "/etc/ascend/config/test-job/0_vnpu-001"
	testLogicID            = int32(0)
	testDieId              = "die-001"
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

type FakeDirEntry struct {
	name string
	typ  fs.FileMode
	info fs.FileInfo
}

func (f FakeDirEntry) Name() string { return f.name }

func (f FakeDirEntry) IsDir() bool { return false }

func (f FakeDirEntry) Type() fs.FileMode { return f.typ }

func (f FakeDirEntry) Info() (fs.FileInfo, error) { return f.info, nil }

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

// TestAllocateUBDevice for test the Allocate when ub devices exist
func TestAllocateUBDevice(t *testing.T) {
	ps := NewPluginServer(api.Ascend910, devices, nil, device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = false
	common.ParamOption.UseAscendDocker = false
	var requests v1beta1.AllocateRequest
	convey.Convey("test allocate ub devices", t, func() {
		mockGetNPUsFunc := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend910Manager)), "GetNPUs",
			func(_ *device.HwAscend910Manager) (common.NpuAllInfo, error) {
				return mockAllNpuInfo, nil
			})
		defer mockGetNPUsFunc.Reset()
		mockSlowNodeFunc := mockSetSlowNodeNoticeEnv()
		defer mockSlowNodeFunc.Reset()
		ps.deepCopyDevice(devices)
		requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{
			{DevicesIDs: []string{"Ascend910-0", "Ascend910-2"}},
		}

		convey.Convey("ub devices not exist", func() {
			mockOsStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			})
			defer mockOsStat.Reset()
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(len(resp.ContainerResponses[0].Devices), convey.ShouldEqual,
				len(requests.ContainerRequests[0].DevicesIDs))
		})

		convey.Convey("ub devices exist", func() {
			mockOsStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, nil
			})
			defer mockOsStat.Reset()
			mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return mockUBDevices, nil
			})
			defer mockReadDir.Reset()
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			mountDeviceNum := 8
			convey.So(len(resp.ContainerResponses[0].Devices), convey.ShouldEqual, mountDeviceNum)
		})
	})
}

// TestAddDevicesInDir tests the function addDevicesInDir
func TestAddDevicesInDir(t *testing.T) {
	convey.Convey("test addDevicesInDir", t, func() {
		convey.Convey("test addDevicesInDir should add ub devices when read ub directory normally", func() {
			resp := new(v1beta1.ContainerAllocateResponse)
			mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					FakeDirEntry{name: "udma1", typ: fs.ModeDevice},
				}, nil
			})
			defer mockReadDir.Reset()

			err := addDevicesInDir(resp, "/dev/uburma")
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp.Devices, convey.ShouldNotBeEmpty)
		})
		convey.Convey("test addDevicesInDir should return error when readDir error", func() {
			resp := new(v1beta1.ContainerAllocateResponse)
			mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return nil, fmt.Errorf("dir read error")
			})
			defer mockReadDir.Reset()

			err := addDevicesInDir(resp, "/dev/uburma")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(resp.Devices, convey.ShouldBeEmpty)
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
			_, _, err := ps.useVolcano(devicesIDs)
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

// TestExitSelfProcess test exit self process
func TestExitSelfProcess(t *testing.T) {
	convey.Convey("test exitSelfProcess case 1", t, func() {
		mock1 := gomonkey.ApplyFunc(os.Getpid, func() int {
			return 1
		})
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(os.FindProcess, func(_ int) (*os.Process, error) {
			return nil, errors.New("fake error 1")
		})
		defer mock2.Reset()
		convey.So(exitSelfProcess().Error(), convey.ShouldEqual, "fake error 1")
	})
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

func TestIsValidPhyID(t *testing.T) {
	convey.Convey("test isValidPhyID case 1", t, func() {
		cacheDevices := []common.NpuDevice{
			{PhyID: 1},
		}
		ps := PluginServer{
			cachedDevices: cacheDevices,
		}
		convey.So(ps.isValidPhyID("1"), convey.ShouldBeTrue)
	})

	convey.Convey("test isValidPhyID case 2", t, func() {
		cacheDevices := []common.NpuDevice{
			{PhyID: 1},
		}
		ps := PluginServer{
			cachedDevices: cacheDevices,
		}
		convey.So(ps.isValidPhyID("2"), convey.ShouldBeFalse)
	})
}

func TestIsValidRequestID(t *testing.T) {
	convey.Convey("test isValidRequestID case 1", t, func() {
		cacheDevices := []common.NpuDevice{
			{PhyID: 1},
		}
		ps := PluginServer{
			cachedDevices: cacheDevices,
		}
		ids := []string{"1"}
		convey.So(len(ps.isValidRequestID(ids)), convey.ShouldEqual, 0)
	})

	convey.Convey("test isValidRequestID case 2", t, func() {
		cacheDevices := []common.NpuDevice{
			{PhyID: 2},
		}
		ps := PluginServer{
			cachedDevices: cacheDevices,
		}
		ids := []string{"1"}
		convey.So(len(ps.isValidRequestID(ids)), convey.ShouldEqual, 1)
	})
}

func TestAddSoftShareDev(t *testing.T) {
	convey.Convey("Test addSoftShareDev", t, func() {
		ps := &PluginServer{}
		mockOption := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{ShareCount: intNum10})
		defer mockOption.Reset()
		resp := &v1beta1.ListAndWatchResponse{
			Devices: []*v1beta1.Device{},
		}
		ps.addSoftShareDev(resp, ascend910LogicID0, common.NpuDevice{Health: v1beta1.Healthy})
		convey.So(len(resp.Devices), convey.ShouldEqual, intNum10)
	})
}

type updatePresetAllocMapTestCase struct {
	name                string
	supportSoftShare    bool
	realAlloc           []string
	kltAlloc            []string
	preKlt2RealDevMap   map[string]string
	expectedKlt2RealDev map[string]string
}

func buildUpdatePresetAllocMapTestCases() []updatePresetAllocMapTestCase {
	return []updatePresetAllocMapTestCase{
		{
			name:                "soft share enabled, realAlloc non-empty, kltAlloc non-empty",
			supportSoftShare:    true,
			realAlloc:           []string{"npu0"},
			kltAlloc:            []string{"npu0-0", "npu0-1"},
			preKlt2RealDevMap:   map[string]string{"npu0-0": "old-npu", "other": "npu1"},
			expectedKlt2RealDev: map[string]string{"other": "npu1", "npu0-0": "npu0", "npu0-1": "npu0"},
		},
		{
			name:                "soft share enabled, realAlloc empty",
			supportSoftShare:    true,
			realAlloc:           []string{},
			kltAlloc:            []string{"npu0-0"},
			preKlt2RealDevMap:   map[string]string{"npu0-0": "old-npu"},
			expectedKlt2RealDev: map[string]string{},
		},
		{
			name:                "soft share disabled, len(realAlloc) == len(kltAlloc)",
			supportSoftShare:    false,
			realAlloc:           []string{"npu0", "npu1"},
			kltAlloc:            []string{"klt0", "klt1"},
			preKlt2RealDevMap:   map[string]string{"klt0": "old", "klt2": "npu2"},
			expectedKlt2RealDev: map[string]string{"klt2": "npu2", "klt0": "npu0", "klt1": "npu1"},
		},
		{
			name:                "soft share disabled, len(realAlloc) != len(kltAlloc)",
			supportSoftShare:    false,
			realAlloc:           []string{"npu0"},
			kltAlloc:            []string{"klt0", "klt1"},
			preKlt2RealDevMap:   map[string]string{"klt0": "old"},
			expectedKlt2RealDev: map[string]string{"klt0": "old"},
		},
	}
}

func TestUpdatePresetAllocMap(t *testing.T) {
	convey.Convey("Test updatePresetAllocMap", t, func() {
		ps := &PluginServer{
			klt2RealDevMap: make(map[string]string),
			allocMapLock:   sync.RWMutex{},
		}
		tests := buildUpdatePresetAllocMapTestCases()
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				ps.klt2RealDevMap = make(map[string]string)
				for k, v := range tt.preKlt2RealDevMap {
					ps.klt2RealDevMap[k] = v
				}
				patchSoftShare := gomonkey.ApplyFuncReturn(common.IsSupportSoftShareDevice, tt.supportSoftShare)
				defer patchSoftShare.Reset()
				ps.updatePresetAllocMap(tt.realAlloc, tt.kltAlloc)
				convey.So(ps.klt2RealDevMap, convey.ShouldResemble, tt.expectedKlt2RealDev)
			})
		}
	})
}

type getValidLogicDeviceIDTestCase struct {
	name                string
	devices             []string
	mockGetDeviceListID func() (map[int]int, []int, error)
	expectedID          int
	expectedErr         bool
}

func buildGetValidLogicDeviceIDTestCases() []getValidLogicDeviceIDTestCase {
	return []getValidLogicDeviceIDTestCase{
		{
			name:    "GetDeviceListID returns error",
			devices: []string{"npu0"},
			mockGetDeviceListID: func() (map[int]int, []int, error) {
				return nil, nil, errors.New("device list fetch failed")
			},
			expectedID:  0,
			expectedErr: true,
		},
		{
			name:    "visible devices length is 0",
			devices: []string{"npu0"},
			mockGetDeviceListID: func() (map[int]int, []int, error) {
				return map[int]int{}, []int{}, nil
			},
			expectedID:  0,
			expectedErr: true,
		},
		{
			name:    "visible devices length is 2",
			devices: []string{"npu0", "npu1"},
			mockGetDeviceListID: func() (map[int]int, []int, error) {
				return map[int]int{}, []int{0, 1}, nil
			},
			expectedID:  0,
			expectedErr: true,
		},
		{
			name:    "visible devices length is 1, return correct ID",
			devices: []string{"npu0"},
			mockGetDeviceListID: func() (map[int]int, []int, error) {
				return map[int]int{}, []int{intNum2}, nil
			},
			expectedID:  intNum2,
			expectedErr: false,
		},
	}
}

func TestGetValidLogicDeviceID(t *testing.T) {
	convey.Convey("Test getValidLogicDeviceID", t, func() {
		ps := &PluginServer{}
		tests := buildGetValidLogicDeviceIDTestCases()
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				patch := gomonkey.ApplyFunc(common.GetDeviceListID,
					func(devices []string, opts string) (map[int]int, []int, error) {
						return tt.mockGetDeviceListID()
					})
				defer patch.Reset()
				gotID, err := ps.getValidLogicDeviceID(tt.devices)
				convey.So(gotID, convey.ShouldEqual, tt.expectedID)
				if tt.expectedErr {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
			})
		}
	})
}

type extractPodAnnotationsTestCase struct {
	name                string
	pod                 *v1.Pod
	mockConvertPolicy   func() string
	expectedAnnotations softShareDevAnnotations
	expectedErr         bool
}

func buildExtractPodAnnotationsTestCases1() []extractPodAnnotationsTestCase {
	return []extractPodAnnotationsTestCase{
		{
			name: "all annotations exist and non-empty, policy converted",
			pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				api.SchedulerSoftShareDevVNPUIdKey:      mockValidVNPUId,
				api.SchedulerSoftShareDevAicoreQuotaKey: mockValidAicoreQuota,
				api.SchedulerSoftShareDevHbmQuotaKey:    mockValidHbmQuota,
				api.SchedulerSoftShareDevPolicyKey:      mockValidPolicyStr,
			}}},
			mockConvertPolicy: func() string {
				return mockConvertedPolicy
			},
			expectedAnnotations: softShareDevAnnotations{
				vNPUId:           mockValidVNPUId,
				aicoreQuota:      mockValidAicoreQuota,
				hbmQuota:         mockValidHbmQuota,
				schedulingPolicy: mockConvertedPolicy,
			},
			expectedErr: false,
		},
		{
			name: "missing single annotation (vNPUId)",
			pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				api.SchedulerSoftShareDevAicoreQuotaKey: mockValidAicoreQuota,
				api.SchedulerSoftShareDevHbmQuotaKey:    mockValidHbmQuota,
				api.SchedulerSoftShareDevPolicyKey:      mockValidPolicyStr,
			}}},
			mockConvertPolicy: func() string {
				return mockConvertedPolicy
			},
			expectedAnnotations: softShareDevAnnotations{},
			expectedErr:         true,
		},
		{
			name: "multiple annotations value empty",
			pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				api.SchedulerSoftShareDevVNPUIdKey:      mockValidVNPUId,
				api.SchedulerSoftShareDevAicoreQuotaKey: "",
				api.SchedulerSoftShareDevHbmQuotaKey:    "",
				api.SchedulerSoftShareDevPolicyKey:      mockValidPolicyStr,
			}}},
			mockConvertPolicy: func() string {
				return mockConvertedPolicy
			},
			expectedAnnotations: softShareDevAnnotations{},
			expectedErr:         true,
		},
	}
}

func buildExtractPodAnnotationsTestCases2() []extractPodAnnotationsTestCase {
	return []extractPodAnnotationsTestCase{
		{
			name: "pod has no annotations",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			mockConvertPolicy: func() string {
				return mockConvertedPolicy
			},
			expectedAnnotations: softShareDevAnnotations{},
			expectedErr:         true,
		},
		{
			name:                "pod is nil",
			pod:                 nil,
			mockConvertPolicy:   func() string { return "" },
			expectedAnnotations: softShareDevAnnotations{},
			expectedErr:         true,
		},
		{
			name: "policy annotation value empty",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						api.SchedulerSoftShareDevVNPUIdKey:      mockValidVNPUId,
						api.SchedulerSoftShareDevAicoreQuotaKey: mockValidAicoreQuota,
						api.SchedulerSoftShareDevHbmQuotaKey:    mockValidHbmQuota,
						api.SchedulerSoftShareDevPolicyKey:      "",
					},
				},
			},
			mockConvertPolicy: func() string {
				return mockConvertedPolicy
			},
			expectedAnnotations: softShareDevAnnotations{},
			expectedErr:         true,
		},
	}
}

func TestExtractPodAnnotations(t *testing.T) {
	convey.Convey("Test extractPodAnnotations", t, func() {
		ps := &PluginServer{}
		tests := append(buildExtractPodAnnotationsTestCases1(), buildExtractPodAnnotationsTestCases2()...)
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				if tt.pod == nil {
					convey.So(func() {
						_, _ = ps.extractPodAnnotations(tt.pod)
					}, convey.ShouldPanic)
					return
				}
				patch := gomonkey.ApplyFunc(common.ConvertSchedulingPolicyToIntStr,
					func(policy string) string {
						return tt.mockConvertPolicy()
					})
				defer patch.Reset()

				gotAnnotations, err := ps.extractPodAnnotations(tt.pod)
				if tt.expectedErr {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
				convey.So(gotAnnotations.vNPUId, convey.ShouldEqual, tt.expectedAnnotations.vNPUId)
				convey.So(gotAnnotations.aicoreQuota, convey.ShouldEqual, tt.expectedAnnotations.aicoreQuota)
				convey.So(gotAnnotations.hbmQuota, convey.ShouldEqual, tt.expectedAnnotations.hbmQuota)
				convey.So(gotAnnotations.schedulingPolicy, convey.ShouldEqual, tt.expectedAnnotations.schedulingPolicy)
			})
		}
	})
}

type configDirPathTestCase struct {
	name          string
	pod           *v1.Pod
	jobName       string
	logicID       int
	vNPUId        string
	mockParentDir string
	expectedPath  string
}

func buildConfigDirPathTestCases() []configDirPathTestCase {
	testNamespace := "default"
	testJobName := "test-job"
	return []configDirPathTestCase{
		{
			name: "all parameters valid, return correct path",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
				},
			},
			jobName:      testJobName,
			logicID:      0,
			vNPUId:       "1",
			expectedPath: common.SoftShareDevNPUInfoConfigParentDirPath + "default.test-job/0_1",
		},
		{
			name: "jobName is empty, return empty string and log error",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
				},
			},
			jobName:      "",
			logicID:      0,
			vNPUId:       "1",
			expectedPath: "",
		},
		{
			name:         "pod is nil, access Namespace panic",
			pod:          nil,
			jobName:      testJobName,
			logicID:      0,
			vNPUId:       "1",
			expectedPath: "",
		},
	}
}

func TestBuildConfigDirPath(t *testing.T) {
	convey.Convey("Test buildConfigDirPath", t, func() {
		ps := &PluginServer{}
		tests := buildConfigDirPathTestCases()
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				gotPath := ps.buildConfigDirPath(tt.pod, tt.jobName, tt.logicID, tt.vNPUId)
				convey.So(gotPath, convey.ShouldEqual, tt.expectedPath)
			})
		}
	})
}

type npuConfigFileTestCase struct {
	name               string
	configDir          string
	physicalID         int
	dieId              string
	annotations        softShareDevAnnotations
	mockWriteFileErr   error
	expectedErr        bool
	expectedConfigData string
}

func getTestBaseConfig() (string, int, string) {
	configDir := "/etc/ascend/config"
	physicalID := 0
	dieId := "die-001"
	return configDir, physicalID, dieId
}

func buildValidAnnotations() softShareDevAnnotations {
	return softShareDevAnnotations{
		vNPUId:           "vnpu-001",
		aicoreQuota:      "50",
		hbmQuota:         "2048",
		schedulingPolicy: "2",
	}
}

func buildPartialEmptyAnnotations() softShareDevAnnotations {
	return softShareDevAnnotations{
		vNPUId:           "",
		aicoreQuota:      "50",
		hbmQuota:         "",
		schedulingPolicy: "2",
	}
}

func buildValidConfigData() string {
	return strings.Join([]string{
		api.SoftShareDeviceConfigPhysicalNPUId + "=0",
		api.SoftShareDeviceConfigVirtualNPUId + "=vnpu-001",
		api.SoftShareDeviceConfigAICoreQuota + "=50",
		api.SoftShareDeviceConfigHbmQuota + "=2048",
		api.SoftShareDeviceConfigShmId + "=die-001",
		api.SoftShareDeviceConfigSchedulingPolicy + "=2",
	}, "\n")
}

func buildPartialEmptyConfigData() string {
	return strings.Join([]string{
		api.SoftShareDeviceConfigPhysicalNPUId + "=0",
		api.SoftShareDeviceConfigAICoreQuota + "=50",
		api.SoftShareDeviceConfigShmId + "=die-001",
		api.SoftShareDeviceConfigSchedulingPolicy + "=2",
	}, "\n")
}

func buildNpuConfigFileTestCases() []npuConfigFileTestCase {
	configDir := "/etc/ascend/config"
	physicalID := 0
	dieId := "die-001"
	validAnnot := buildValidAnnotations()
	partialEmptyAnnot := buildPartialEmptyAnnotations()
	validConfigData := buildValidConfigData()
	partialEmptyConfigData := buildPartialEmptyConfigData()
	return []npuConfigFileTestCase{
		{
			name:               "all config items non-empty, write file success",
			configDir:          configDir,
			physicalID:         physicalID,
			dieId:              dieId,
			annotations:        validAnnot,
			mockWriteFileErr:   nil,
			expectedErr:        false,
			expectedConfigData: validConfigData,
		},
		{
			name:               "partial config items empty, skip empty and write valid",
			configDir:          configDir,
			physicalID:         physicalID,
			dieId:              dieId,
			annotations:        partialEmptyAnnot,
			mockWriteFileErr:   nil,
			expectedErr:        false,
			expectedConfigData: partialEmptyConfigData,
		},
		{
			name:               "write file failed, return wrapped error",
			configDir:          configDir,
			physicalID:         physicalID,
			dieId:              dieId,
			annotations:        validAnnot,
			mockWriteFileErr:   errors.New("permission denied"),
			expectedErr:        true,
			expectedConfigData: validConfigData,
		},
	}
}

func TestWriteNPUConfigFile(t *testing.T) {
	convey.Convey("Test writeNPUConfigFile", t, func() {
		ps := &PluginServer{}
		tests := buildNpuConfigFileTestCases()
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				var actualConfigData string
				var actualFileFullName string
				patchWriteFile := gomonkey.ApplyFunc(common.WriteToFileWithPerm,
					func(data, filename string, dirPerm, filePerm os.FileMode) error {
						actualConfigData = data
						actualFileFullName = filename
						return tt.mockWriteFileErr
					})
				defer patchWriteFile.Reset()
				err := ps.writeNPUConfigFile(tt.configDir, tt.physicalID, tt.dieId, tt.annotations)
				if tt.expectedErr {
					convey.So(err, convey.ShouldNotBeNil)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
				if len(tt.expectedConfigData) > 0 {
					convey.So(actualConfigData, convey.ShouldEqual, tt.expectedConfigData)
					convey.So(actualFileFullName, convey.ShouldEqual,
						tt.configDir+"/"+api.SoftShareDeviceConfigFileName)
				}
			})
		}
	})
}

type getNPUInfoConfigDirTestCase struct {
	name                string
	pod                 *v1.Pod
	devices             []string
	mockIsSoftShareDev  bool
	mockExtractAnnot    func() (softShareDevAnnotations, error)
	mockGetValidLogicID func() (int, error)
	mockGetJobName      string
	mockBuildConfigDir  string
	mockGetLogicID      func() (int32, error)
	mockGetDieID        func() (string, error)
	mockWriteConfigFile error
	expectedDir         string
}

func buildGetNPUInfoConfigDirTestCases1() []getNPUInfoConfigDirTestCase {
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-pod"}}
	testDevices := []string{"npu0"}
	testAnnotations := softShareDevAnnotations{vNPUId: "vnpu-001"}
	return []getNPUInfoConfigDirTestCase{
		{
			name:                "pod is nil",
			pod:                 nil,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
		{
			name:                "pod is not soft share job",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  false,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
		{
			name:                "extract annotations failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return softShareDevAnnotations{}, errors.New("missing annot") },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
	}
}

func buildGetNPUInfoConfigDirTestCases2() []getNPUInfoConfigDirTestCase {
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-pod"}}
	testDevices := []string{"npu0"}
	testAnnotations := softShareDevAnnotations{vNPUId: "vnpu-001"}
	return []getNPUInfoConfigDirTestCase{
		{
			name:                "get valid logic device ID failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return 0, errors.New("device not found") },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
		{
			name:                "build config dir failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  "",
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
		{
			name:                "get logic ID from physical ID failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return 0, errors.New("logic ID not found") },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
	}
}

func buildGetNPUInfoConfigDirTestCases3() []getNPUInfoConfigDirTestCase {
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-pod"}}
	testDevices := []string{"npu0"}
	testAnnotations := softShareDevAnnotations{vNPUId: "vnpu-001"}
	return []getNPUInfoConfigDirTestCase{
		{
			name:                "get die ID failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return "", errors.New("die ID not found") },
			mockWriteConfigFile: nil,
			expectedDir:         "",
		},
		{
			name:                "write NPU config file failed",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: errors.New("write failed"),
			expectedDir:         "",
		},
		{
			name:                "all steps success",
			pod:                 testPod,
			devices:             testDevices,
			mockIsSoftShareDev:  true,
			mockExtractAnnot:    func() (softShareDevAnnotations, error) { return testAnnotations, nil },
			mockGetValidLogicID: func() (int, error) { return testPhysicalID, nil },
			mockGetJobName:      testJobName,
			mockBuildConfigDir:  testConfigDir,
			mockGetLogicID:      func() (int32, error) { return testLogicID, nil },
			mockGetDieID:        func() (string, error) { return testDieId, nil },
			mockWriteConfigFile: nil,
			expectedDir:         testConfigDir,
		},
	}
}

func TestGetNPUInfoConfigDirFromPod(t *testing.T) {
	convey.Convey("Test getNPUInfoConfigDirFromPod", t, func() {
		ps := &PluginServer{manager: device.NewHwAscend910Manager()}
		tests := append(buildGetNPUInfoConfigDirTestCases1(), append(buildGetNPUInfoConfigDirTestCases2(),
			buildGetNPUInfoConfigDirTestCases3()...)...)
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				patchIsSoftShare := gomonkey.ApplyFuncReturn(common.IsSoftShareDevJob, tt.mockIsSoftShareDev)
				defer patchIsSoftShare.Reset()
				patchExtractAnnot := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ps), "extractPodAnnotations",
					func(_ *PluginServer, pod *v1.Pod) (softShareDevAnnotations, error) { return tt.mockExtractAnnot() })
				defer patchExtractAnnot.Reset()
				patchGetValidLogicID := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ps), "getValidLogicDeviceID",
					func(_ *PluginServer, devices []string) (int, error) { return tt.mockGetValidLogicID() })
				defer patchGetValidLogicID.Reset()
				patchGetJobName := gomonkey.ApplyFuncReturn(common.GetJobNameOfPod, tt.mockGetJobName)
				defer patchGetJobName.Reset()
				patchBuildConfigDir := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ps), "buildConfigDirPath",
					func(_ *PluginServer, pod *v1.Pod, jobName string, logicID int, vNPUId string) string {
						return tt.mockBuildConfigDir
					})
				defer patchBuildConfigDir.Reset()
				mockGetDmgr := gomonkey.ApplyMethod(ps.manager, "GetDmgr",
					func(_ *device.HwAscend910Manager) devmanager.DeviceInterface { return &devmanager.DeviceManagerMock{} })
				defer mockGetDmgr.Reset()
				patchGetLogicID := gomonkey.ApplyMethod(ps.manager.GetDmgr(), "GetLogicIDFromPhysicID",
					func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) { return tt.mockGetLogicID() })
				defer patchGetLogicID.Reset()
				patchGetDieID := gomonkey.ApplyMethod(ps.manager.GetDmgr(), "GetDieID",
					func(_ *devmanager.DeviceManagerMock, logicID int32, dieType dcmi.DieType) (string, error) {
						return tt.mockGetDieID()
					})
				defer patchGetDieID.Reset()
				patchWriteConfigFile := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ps), "writeNPUConfigFile",
					func(_ *PluginServer, configDir string, physicalID int, dieId string,
						annotations softShareDevAnnotations) error {
						return tt.mockWriteConfigFile
					})
				defer patchWriteConfigFile.Reset()
				gotDir := ps.getNPUInfoConfigDirFromPod(tt.pod, tt.devices)
				convey.So(gotDir, convey.ShouldEqual, tt.expectedDir)
			})
		}
	})
}
