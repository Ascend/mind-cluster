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
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"

	"Ascend-device-plugin/pkg/common"
)

const (
	sockMode = 0755
)

func init() {
	if _, err := os.Stat(socketPath); err == nil {
		return
	}
	if _, err := os.OpenFile(socketPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sockMode); err != nil {
		fmt.Errorf("err: %v", err)
		return
	}
	if err := os.Chmod(socketPath, os.ModeSocket); err != nil {
		fmt.Errorf("err: %v", err)
		return
	}
}

// TestPodResourceStart1 for test the interface Start part 2
func TestPodResourceStart1(t *testing.T) {
	pr := NewPodResource()
	convey.Convey("test start", t, func() {
		convey.Convey("VerifyPath failed", func() {
			mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string,
				waitSecond int) (string, bool) {
				return "", false
			})
			defer mockVerifyPath.Reset()
			convey.So(pr.start(), convey.ShouldNotBeNil)
		})
		convey.Convey("VerifyPath ok", func() {
			mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string,
				waitSecond int) (string, bool) {
				return "", true
			})
			defer mockVerifyPath.Reset()
			err := pr.start()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestPodResourceStart2 for test the interface Start part 2
func TestPodResourceStart2(t *testing.T) {
	pr := NewPodResource()
	convey.Convey("test start", t, func() {
		convey.Convey("GetClient failed", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetV1alpha1Client, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, fmt.Errorf("err")
			})
			defer mockGetClient.Reset()
			convey.So(pr.start(), convey.ShouldNotBeNil)
		})
		convey.Convey("start ok", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetV1alpha1Client, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, nil
			})
			defer mockGetClient.Reset()
			funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission,
				func(verifyPathAndPermission string, waitSecond int) (string, bool) {
					return verifyPathAndPermission, true
				})
			defer funcStub.Reset()
			convey.So(pr.start(), convey.ShouldBeNil)
		})
	})
}

// TestPodResourceStart for test the interface Stop
func TestPodResourceStop(t *testing.T) {
	convey.Convey("test start", t, func() {
		convey.Convey("close failed", func() {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			mockClose := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)), "Close",
				func(_ *grpc.ClientConn) error { return fmt.Errorf("err") })
			defer mockClose.Reset()
			pr.stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
		convey.Convey("close ok", func() {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			mockClose := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)), "Close",
				func(_ *grpc.ClientConn) error { return nil })
			defer mockClose.Reset()
			pr.stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
	})
}

type FakeClient struct{}

// List is to get pod resource
func (c *FakeClient) List(ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
	opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
	out := new(v1alpha1.ListPodResourcesResponse)
	return out, nil
}

// getContainerResourceTestCase getContainerResource test case
type getContainerResourceTestCase struct {
	Name              string
	containerResource *v1alpha1.ContainerResources
	WantDevices       []string
	WantErr           error
}

func buildGetContainerResourceTestCases() []getContainerResourceTestCase {
	return []getContainerResourceTestCase{
		{
			Name:              "01-ContainerResources is nil, should return empty resource and error",
			containerResource: nil,
			WantDevices:       nil,
			WantErr:           errors.New("invalid container resource"),
		},
		{
			Name: "02-device is list is empty, should return empty resource and error",
			containerResource: &v1alpha1.ContainerResources{
				Devices: []*v1alpha1.ContainerDevices{
					nil,
					{ResourceName: "notExistResourceName"},
					{ResourceName: common.HuaweiUnHealthAscend910, DeviceIds: []string{}},
				},
			},
			WantDevices: nil,
			WantErr:     errors.New("container device num 0 exceeds the upper limit"),
		},
		{
			Name: "02-get container resource success, should return resource and nil",
			containerResource: &v1alpha1.ContainerResources{
				Devices: []*v1alpha1.ContainerDevices{
					{ResourceName: common.HuaweiUnHealthAscend910, DeviceIds: []string{"0", "1"}},
				},
			},
			WantDevices: []string{"0", "1"},
			WantErr:     nil,
		},
	}
}

// TestGetContainerResource for test getContainerResource
func TestGetContainerResource(t *testing.T) {
	testCases := buildGetContainerResourceTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			_, deviceIds, err := pr.getContainerResource(tt.containerResource)
			if !reflect.DeepEqual(deviceIds, tt.WantDevices) {
				t.Errorf("getContainerResource() deviceIds = %v, WantDevices = %v", deviceIds, tt.WantDevices)
			}
			if !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("getContainerResource() err = %v, wantErr = %v", err, tt.WantErr)
			}
		})
	}
}

// getDeviceFromPodTestCase getDeviceFromPod test case
type getDeviceFromPodTestCase struct {
	Name         string
	podResources *v1alpha1.PodResources
	WantDevices  []string
	WantErr      error
}

func buildGetDeviceFromPodTestCases() []getDeviceFromPodTestCase {
	return []getDeviceFromPodTestCase{
		{
			Name:         "01-podResources is nil, should return empty resource and error",
			podResources: nil,
			WantDevices:  nil,
			WantErr:      errors.New("invalid podReousrces"),
		},
		{
			Name: "02-get device from pod success, should return resource and nil",
			podResources: &v1alpha1.PodResources{
				Containers: []*v1alpha1.ContainerResources{
					{
						Devices: []*v1alpha1.ContainerDevices{
							{ResourceName: common.HuaweiUnHealthAscend910, DeviceIds: []string{"0", "1"}},
						},
					},
					{
						Devices: []*v1alpha1.ContainerDevices{
							{ResourceName: common.HuaweiUnHealthAscend910, DeviceIds: []string{"3", "4"}},
						},
					},
				},
			},
			WantDevices: []string{"0", "1", "3", "4"},
			WantErr:     nil,
		},
	}
}

// TestGetDeviceFromPod for test getDeviceFromPod
func TestGetDeviceFromPod(t *testing.T) {
	testCases := buildGetDeviceFromPodTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			_, deviceIds, err := pr.getDeviceFromPod(tt.podResources)
			if !reflect.DeepEqual(deviceIds, tt.WantDevices) {
				t.Errorf("getDeviceFromPod() deviceIds = %v, WantDevices = %v", deviceIds, tt.WantDevices)
			}
			if !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("getDeviceFromPod() err = %v, wantErr = %v", err, tt.WantErr)
			}
		})
	}
}

// assemblePodResourceTestCase assemblePodResource test case
type assemblePodResourceTestCase struct {
	Name        string
	resp        *v1alpha1.ListPodResourcesResponse
	listErr     error
	WantDevices map[string]PodDevice
	WantErr     error
}

func buildAssemblePodResourceTestCases() []assemblePodResourceTestCase {
	return []assemblePodResourceTestCase{
		{
			Name:        "01-client list error, should return empty pod resource and error",
			resp:        nil,
			listErr:     errors.New("fake client list error"),
			WantDevices: nil,
			WantErr:     errors.New("list pod resource failed, err: fake client list error"),
		},
		{
			Name: "02-assemble pod success, should return pod resource and nil",
			resp: &v1alpha1.ListPodResourcesResponse{
				PodResources: []*v1alpha1.PodResources{
					nil,
					{Name: "INVALID-NAME", Namespace: "namespace"},
					{Name: "valid-name", Namespace: ".invalid"},
					{Name: "valid-name", Namespace: "namespace", Containers: []*v1alpha1.ContainerResources{
						{
							Devices: []*v1alpha1.ContainerDevices{
								{ResourceName: common.HuaweiUnHealthAscend910, DeviceIds: []string{"0", "1"}},
							},
						},
					}},
				},
			},
			listErr: nil,
			WantDevices: map[string]PodDevice{
				"namespace_valid-name": {
					ResourceName: common.HuaweiUnHealthAscend910,
					DeviceIds:    []string{"0", "1"},
				},
			},
			WantErr: nil,
		},
	}
}

// TestAssemblePodResource for test assemblePodResource
func TestAssemblePodResource(t *testing.T) {
	testCases := buildAssemblePodResourceTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			pr := &PodResource{conn: &grpc.ClientConn{}, client: &FakeClient{}}
			patch := gomonkey.ApplyMethodReturn(&FakeClient{}, "List", tt.resp, tt.listErr)
			podResourceList, err := pr.assemblePodResource()
			patch.Reset()
			if !reflect.DeepEqual(podResourceList, tt.WantDevices) {
				t.Errorf("assemblePodResource() podResourceList = %v, "+
					"WantDevices = %v", podResourceList, tt.WantDevices)
			}
			if !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("assemblePodResource() err = %v, wantErr = %v", err, tt.WantErr)
			}
		})
	}
}

// TestStop for test stop
func TestStop(t *testing.T) {
	pr := NewPodResource()
	pr.conn = &grpc.ClientConn{}
	pr.client = &FakeClient{}
	patch := gomonkey.ApplyMethodReturn(&grpc.ClientConn{}, "Close",
		errors.New("fake close error"))
	defer patch.Reset()
	pr.stop()
	if pr.conn != nil || pr.client != nil {
		t.Errorf("conn: %v and client: %v should be nilisCompleted", pr.conn, pr.client)
	}
}

type isPodMoveCompleteTestCase struct {
	Name            string
	podDevices      map[string]PodDevice
	mockErr         error
	podList         []v1.Pod
	wantIsCompleted bool
}

func buildIsPodMoveCompleteTestCases() []isPodMoveCompleteTestCase {
	podList := []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "pod0"}},
		{ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "pod1"}},
	}
	return []isPodMoveCompleteTestCase{
		{
			Name:            "01-get pod resources error, should return false",
			podDevices:      nil,
			mockErr:         errors.New("fake get valid pod resources error"),
			podList:         podList,
			wantIsCompleted: false,
		},
		{
			Name: "02-pod move completed, should return true",
			podDevices: map[string]PodDevice{
				"namespace_pod0": {
					ResourceName: common.HuaweiUnHealthAscend910,
					DeviceIds:    []string{},
				},
				"namespace_pod1": {
					ResourceName: common.HuaweiUnHealthAscend910,
					DeviceIds:    []string{"0", "1"},
				},
			},
			mockErr:         nil,
			podList:         podList,
			wantIsCompleted: true,
		},
	}
}

// TestIsPodMoveComplete for test IsPodMoveComplete
func TestIsPodMoveComplete(t *testing.T) {
	testCases := buildIsPodMoveCompleteTestCases()
	deviceName := "device0"
	ps := &PluginServer{
		klt2RealDevMap: map[string]string{
			"2": deviceName,
		},
	}
	patch := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{UseVolcanoType: true})
	defer patch.Reset()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			pr := &PodResource{conn: &grpc.ClientConn{}, client: &FakeClient{}}
			patch1 := gomonkey.ApplyMethodReturn(&PodResource{}, "GetPodResource",
				tt.podDevices, tt.mockErr)
			isCompleted := pr.IsPodMoveComplete(deviceName, tt.podList, ps)
			patch1.Reset()
			if isCompleted != tt.wantIsCompleted {
				t.Errorf("IsPodMoveComplete() isCompleted = %v, "+
					"wantIsCompleted = %v", isCompleted, tt.wantIsCompleted)
			}
		})
	}
}
