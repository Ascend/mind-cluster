/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package ub_device for ub device info
package ub_device

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"ascend-common/api"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

const (
	testResourceName = "huawei.com/ub_rdma"
	testContainer    = "inference"
	testNodeName     = "node-1"
)

func newNpuPod(annotations map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod-1",
			Namespace:   "default",
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			NodeName: testNodeName,
			Containers: []v1.Container{
				{
					Name: testContainer,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{v1.ResourceName(testResourceName): resource.MustParse("1")},
					},
				},
			},
		},
	}
}

func newNpuEnvPod(annotations map[string]string, visibleDevices string) v1.Pod {
	pod := newNpuPod(annotations)
	pod.Spec.Containers[0].Env = []v1.EnvVar{{Name: api.AscendVisibleDevicesEnv, Value: visibleDevices}}
	return pod
}

func newNpuServer(devices []types.Device) *ubResourceServer {
	return &ubResourceServer{
		resourceName: testResourceName,
		ubDevices:    devices,
		k8sClient:    &kubernetes.Clientset{},
	}
}

func TestParsePredicateTime(t *testing.T) {
	convey.Convey("Given a pod with a valid predicate-time", t, func() {
		pod := newNpuPod(map[string]string{podPredicateTimeAnnotation: "12345"})
		convey.Convey("Then the timestamp and ok should be returned", func() {
			ts, ok := parsePredicateTime(&pod)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(ts, convey.ShouldEqual, 12345)
		})
	})
	convey.Convey("Given a pod without predicate-time", t, func() {
		pod := newNpuPod(nil)
		convey.Convey("Then ok should be false", func() {
			_, ok := parsePredicateTime(&pod)
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
	convey.Convey("Given a pod with an over-length predicate-time", t, func() {
		pod := newNpuPod(map[string]string{podPredicateTimeAnnotation: "123456789"})
		convey.Convey("Then ok should be false", func() {
			_, ok := parsePredicateTime(&pod)
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
	convey.Convey("Given a pod with a non-numeric predicate-time", t, func() {
		pod := newNpuPod(map[string]string{podPredicateTimeAnnotation: "abc"})
		convey.Convey("Then ok should be false", func() {
			_, ok := parsePredicateTime(&pod)
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

func TestPodIsOlder(t *testing.T) {
	convey.Convey("Given two pods with predicate times", t, func() {
		a := newNpuPod(map[string]string{podPredicateTimeAnnotation: "10"})
		b := newNpuPod(map[string]string{podPredicateTimeAnnotation: "20"})
		convey.Convey("Then the one with the smaller time is older", func() {
			convey.So(podIsOlder(&a, &b), convey.ShouldBeTrue)
			convey.So(podIsOlder(&b, &a), convey.ShouldBeFalse)
		})
	})
	convey.Convey("Given only a has a predicate time", t, func() {
		a := newNpuPod(map[string]string{podPredicateTimeAnnotation: "10"})
		b := newNpuPod(nil)
		convey.Convey("Then a is considered older", func() {
			convey.So(podIsOlder(&a, &b), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given only b has a predicate time", t, func() {
		a := newNpuPod(nil)
		b := newNpuPod(map[string]string{podPredicateTimeAnnotation: "10"})
		convey.Convey("Then a is not considered older", func() {
			convey.So(podIsOlder(&a, &b), convey.ShouldBeFalse)
		})
	})
	convey.Convey("Given pods without predicate times but different creation time", t, func() {
		a := newNpuPod(nil)
		b := newNpuPod(nil)
		a.CreationTimestamp = metav1.NewTime(ts(100))
		b.CreationTimestamp = metav1.NewTime(ts(200))
		convey.Convey("Then the earlier created pod is older", func() {
			convey.So(podIsOlder(&a, &b), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given pods equal in all but UID", t, func() {
		a := newNpuPod(nil)
		b := newNpuPod(nil)
		a.UID = "aaa"
		b.UID = "bbb"
		convey.Convey("Then the pod with the smaller UID is older", func() {
			convey.So(podIsOlder(&a, &b), convey.ShouldBeTrue)
		})
	})
}

// TestPendingNpuPods tests pendingNpuPods filtering out allocated pods and sorting the rest by priority.
func TestPendingNpuPods(t *testing.T) {
	convey.Convey("Given pods where some are already allocated", t, func() {
		pending1 := newNpuPod(map[string]string{podPredicateTimeAnnotation: "20"})
		allocated := newNpuPod(map[string]string{deviceStatusAnnotation: `{"c":["ub0"]}`})
		pending2 := newNpuPod(map[string]string{podPredicateTimeAnnotation: "10"})
		pods := []v1.Pod{pending1, allocated, pending2}
		convey.Convey("Then allocated pods are filtered out and the rest sorted by priority", func() {
			pending := pendingNpuPods(pods)
			convey.So(len(pending), convey.ShouldEqual, 2)
			convey.So(pending[0].Name, convey.ShouldEqual, "pod-1")
			convey.So(pending[0].Annotations[podPredicateTimeAnnotation], convey.ShouldEqual, "10")
			convey.So(pending[1].Annotations[podPredicateTimeAnnotation], convey.ShouldEqual, "20")
		})
	})
}

func TestPodResourceContainerName(t *testing.T) {
	convey.Convey("Given a pod requesting the resource", t, func() {
		pod := newNpuPod(nil)
		convey.Convey("Then the requesting container name is returned", func() {
			convey.So(podResourceContainerName(&pod, testResourceName), convey.ShouldEqual, testContainer)
		})
	})
	convey.Convey("Given a pod not requesting the resource", t, func() {
		pod := v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Name: "c", Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{"cpu": resource.MustParse("1")}}}}}}
		convey.Convey("Then an empty name is returned", func() {
			convey.So(podResourceContainerName(&pod, testResourceName), convey.ShouldEqual, "")
		})
	})
	convey.Convey("Given only the second container requests the resource", t, func() {
		pod := v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{
			{Name: "c1", Resources: v1.ResourceRequirements{Requests: v1.ResourceList{"cpu": resource.MustParse("1")}}},
			{Name: "c2", Resources: v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceName(testResourceName): resource.MustParse("1")}}},
		}}}
		convey.Convey("Then its name is returned", func() {
			convey.So(podResourceContainerName(&pod, testResourceName), convey.ShouldEqual, "c2")
		})
	})
}

func TestMarkPodAllocated(t *testing.T) {
	convey.Convey("Given a pod without the requesting container", t, func() {
		pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"}}
		rs := newNpuServer(nil)
		convey.Convey("Then an error should be returned", func() {
			convey.So(rs.markPodAllocated(context.Background(), &pod, []string{"ub0"}), convey.ShouldNotBeNil)
		})
	})
	convey.Convey("Given a pod requesting the resource", t, func() {
		pod := newNpuPod(nil)
		f := fake.NewSimpleClientset(&pod)
		rs := newNpuServer(nil)
		patches := gomonkey.ApplyMethodReturn((*kubernetes.Clientset)(nil), "CoreV1", f.CoreV1())
		defer patches.Reset()
		convey.Convey("Then the annotations should be patched", func() {
			err := rs.markPodAllocated(context.Background(), &pod, []string{"ub0"})
			convey.So(err, convey.ShouldBeNil)
			got, getErr := f.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
			convey.So(getErr, convey.ShouldBeNil)
			convey.So(got.Annotations[podPredicateTimeAnnotation], convey.ShouldEqual, "18446744073709551615")
			var status map[string][]string
			convey.So(json.Unmarshal([]byte(got.Annotations[deviceStatusAnnotation]), &status), convey.ShouldBeNil)
			convey.So(status[testContainer], convey.ShouldResemble, []string{"ub0"})
		})
	})
	convey.Convey("Given a patch failure", t, func() {
		pod := newNpuPod(nil)
		f := fake.NewSimpleClientset() // pod not present, patch returns an error
		rs := newNpuServer(nil)
		patches := gomonkey.ApplyMethodReturn((*kubernetes.Clientset)(nil), "CoreV1", f.CoreV1())
		defer patches.Reset()
		convey.Convey("Then an error should be returned", func() {
			convey.So(rs.markPodAllocated(context.Background(), &pod, []string{"ub0"}), convey.ShouldNotBeNil)
		})
	})
}

func TestParseNpuIds(t *testing.T) {
	convey.Convey("Given various NPU annotation values", t, func() {
		tests := []struct {
			value    string
			expected []int
			wantErr  bool
		}{
			{"0", []int{0}, false},
			{"2,0,2", []int{0, 2}, false},
			{"", []int{}, false},
			{" 1 ", []int{1}, false},
			{"Ascend910-0,Ascend910-1", []int{0, 1}, false},
			{"npu-1,npu-2", []int{1, 2}, false},
			{"a", nil, true},
			{"0-2", nil, true},
			{"Ascend310-7", nil, true},
			{"Ascend910-2c-180-3", nil, true},
		}
		for _, tt := range tests {
			ids, err := parseNpuIds(tt.value)
			if tt.wantErr {
				convey.So(err, convey.ShouldNotBeNil)
				continue
			}
			convey.So(err, convey.ShouldBeNil)
			convey.So(ids, convey.ShouldResemble, tt.expected)
		}
	})
}

func TestPodRequestsResource(t *testing.T) {
	convey.Convey("Given a pod requesting the resource", t, func() {
		pod := newNpuPod(nil)
		convey.Convey("Then true is returned", func() {
			convey.So(podRequestsResource(&pod, testResourceName), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given a pod not requesting the resource", t, func() {
		pod := v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{"cpu": resource.MustParse("1")}}}}}}
		convey.Convey("Then false is returned", func() {
			convey.So(podRequestsResource(&pod, testResourceName), convey.ShouldBeFalse)
		})
	})
}

// TestGetNpuPodsOnNode tests getNpuPodsOnNode with a successful kubelet response and various failure modes.
func TestGetNpuPodsOnNode(t *testing.T) {
	convey.Convey("Given an active matching pod on the node, it is returned", t, func() {
		newKubeletPodsEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
			list := v1.PodList{Items: []v1.Pod{
				newNpuPod(nil),
				podWithNodeName("node-2", v1.PodPending),
				podWithNodeName(testNodeName, v1.PodSucceeded),
				podWithNodeNameAndResource(testNodeName, v1.PodRunning, false),
			}}
			_ = json.NewEncoder(w).Encode(list)
		}, nil)
		pods := getNpuPodsOnNode(testNodeName, testResourceName)
		convey.So(len(pods), convey.ShouldEqual, 1)
		convey.So(pods[0].Name, convey.ShouldEqual, "pod-1")
	})
	convey.Convey("Given various failure modes, no pods are returned", t, func() {
		tests := []struct {
			name    string
			handler http.HandlerFunc
			cfgErr  error
		}{
			{"in-cluster config failure", func(w http.ResponseWriter, r *http.Request) {}, errTest},
			{"non-200 response", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, nil},
			{"malformed response", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not-json")) }, nil},
		}
		for _, tt := range tests {
			newKubeletPodsEndpoint(t, tt.handler, tt.cfgErr)
			convey.So(getNpuPodsOnNode(testNodeName, testResourceName), convey.ShouldBeNil)
		}
	})
	convey.Convey("Given a kubelet URL build failure, no pods are returned", t, func() {
		patches := gomonkey.ApplyFuncReturn(kubeletPodsURL, "", errTest)
		defer patches.Reset()
		convey.So(getNpuPodsOnNode(testNodeName, testResourceName), convey.ShouldBeNil)
	})
	convey.Convey("Given a kubelet request failure, no pods are returned", t, func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		srv.Close() // closed server makes the request fail
		patches := gomonkey.ApplyFuncReturn(kubeletPodsURL, srv.URL, nil)
		defer patches.Reset()
		patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{BearerToken: "token"}, nil)
		convey.So(getNpuPodsOnNode(testNodeName, testResourceName), convey.ShouldBeNil)
	})
}

func TestKubeletPodsURL(t *testing.T) {
	convey.Convey("Given HOST_IP is not set", t, func() {
		t.Setenv(hostIPEnv, "")
		convey.Convey("Then an error is returned", func() {
			_, err := kubeletPodsURL()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("Given HOST_IP is set without a custom port", t, func() {
		t.Setenv(hostIPEnv, "1.2.3.4")
		t.Setenv(kubeletPortEnv, "")
		convey.Convey("Then the default kubelet port is used", func() {
			url, err := kubeletPodsURL()
			convey.So(err, convey.ShouldBeNil)
			convey.So(url, convey.ShouldEqual, "https://1.2.3.4:10250/pods")
		})
	})
	convey.Convey("Given HOST_IP and a custom KUBELET_PORT", t, func() {
		t.Setenv(hostIPEnv, "1.2.3.4")
		t.Setenv(kubeletPortEnv, "11251")
		convey.Convey("Then the custom port is used", func() {
			url, err := kubeletPodsURL()
			convey.So(err, convey.ShouldBeNil)
			convey.So(url, convey.ShouldEqual, "https://1.2.3.4:11251/pods")
		})
	})
}

// TestSameDeviceSet tests sameDeviceSet order-insensitive matching for equal, different and sized sets.
func TestSameDeviceSet(t *testing.T) {
	convey.Convey("Given a matching device-status annotation", t, func() {
		convey.Convey("Then the set matches order-insensitively", func() {
			convey.So(sameDeviceSet(`{"c":["ub0","ub1"]}`, []string{"ub1", "ub0"}), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given a device-status annotation with different ids", t, func() {
		convey.Convey("Then the set does not match", func() {
			convey.So(sameDeviceSet(`{"c":["ub0","ub2"]}`, []string{"ub0", "ub1"}), convey.ShouldBeFalse)
		})
	})
	convey.Convey("Given a device-status annotation with a different size", t, func() {
		convey.Convey("Then the set does not match", func() {
			convey.So(sameDeviceSet(`{"c":["ub0"]}`, []string{"ub0", "ub1"}), convey.ShouldBeFalse)
		})
	})
}

func TestDeviceStatusSet(t *testing.T) {
	convey.Convey("Given a valid device-status annotation", t, func() {
		set := deviceStatusSet(`{"c1":["ub0"],"c2":["ub1"]}`)
		convey.Convey("Then all device ids are flattened into the set", func() {
			convey.So(len(set), convey.ShouldEqual, 2)
			convey.So(set["ub0"], convey.ShouldNotBeNil)
			convey.So(set["ub1"], convey.ShouldNotBeNil)
		})
	})
	convey.Convey("Given an invalid device-status annotation", t, func() {
		convey.Convey("Then an empty set is returned", func() {
			convey.So(len(deviceStatusSet("not-json")), convey.ShouldEqual, 0)
		})
	})
}

func TestDeviceSpecByID(t *testing.T) {
	convey.Convey("Given a server with a known device", t, func() {
		rs := newNpuServer([]types.Device{&ubDevice{ubID: "ub0", rdmaSpec: []*pluginapi.DeviceSpec{{HostPath: "/dev/ub0"}}}})
		convey.Convey("Then the spec is found by id", func() {
			spec, ok := rs.deviceSpecByID("ub0")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(spec[0].HostPath, convey.ShouldEqual, "/dev/ub0")
		})
		convey.Convey("Then an unknown id is not found", func() {
			_, ok := rs.deviceSpecByID("ub9")
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

func TestGetPodNpuIds(t *testing.T) {
	convey.Convey("Given pods with various NPU annotations", t, func() {
		tests := []struct {
			annotations map[string]string
			env         string
			want        []int
			wantErr     bool
		}{
			{map[string]string{api.HuaweiNPU: "2,3"}, "", []int{2, 3}, false},
			{map[string]string{api.HuaweiNPU: "npu-1,npu-2"}, "", []int{1, 2}, false},
			{map[string]string{api.PodAnnotationAscendReal: "2,3"}, "", nil, true},
			{map[string]string{api.HuaweiAscend910: "0,1"}, "", nil, true},
			{map[string]string{api.HuaweiNPU: "Ascend910-2c-180-3"}, "", nil, true},
			{nil, "0-2", nil, true},
			{nil, "", nil, true},
		}
		for _, tt := range tests {
			var pod v1.Pod
			if tt.env != "" {
				pod = newNpuEnvPod(tt.annotations, tt.env)
			} else {
				pod = newNpuPod(tt.annotations)
			}
			ids, err := getPodNpuIds(&pod)
			if tt.wantErr {
				convey.So(err, convey.ShouldNotBeNil)
				continue
			}
			convey.So(err, convey.ShouldBeNil)
			convey.So(ids, convey.ShouldResemble, tt.want)
		}
	})
}

func TestNpuMappedDeviceIDs(t *testing.T) {
	convey.Convey("Given NPU IDs with various npu-nic mappings", t, func() {
		nicByNpu := func(npuID int) ([]string, error) {
			if npuID == 0 {
				return []string{"enp0s1"}, nil
			}
			return []string{"enp0s2"}, nil
		}
		devs := []types.Device{
			&ubDevice{ubID: "ub0", ifName: "enp0s1"},
			&ubDevice{ubID: "ub1", ifName: "enp0s2"},
		}
		tests := []struct {
			nicNames func(int) ([]string, error)
			npuIds   []int
			want     []string
			wantErr  bool
		}{
			{nicByNpu, []int{0, 1}, []string{"ub0", "ub1"}, false},
			{nicByNpu, []int{0, 0}, []string{"ub0"}, false},
			{func(int) ([]string, error) { return nil, errTest }, []int{0}, nil, true},
			{func(int) ([]string, error) { return []string{}, nil }, []int{0}, nil, true},
			{func(int) ([]string, error) { return []string{"enp0s9"}, nil }, []int{0}, nil, true},
		}
		for _, tt := range tests {
			patches := gomonkey.ApplyFunc(utils.GetNicNames, tt.nicNames)
			ids, err := newNpuServer(devs).npuMappedDeviceIDs(tt.npuIds, "pod-1")
			patches.Reset()
			if tt.wantErr {
				convey.So(err, convey.ShouldNotBeNil)
				continue
			}
			convey.So(err, convey.ShouldBeNil)
			convey.So(ids, convey.ShouldResemble, tt.want)
		}
	})
}

func TestFindDeviceByIfName(t *testing.T) {
	convey.Convey("Given a server with a known ifname", t, func() {
		rs := newNpuServer([]types.Device{&ubDevice{ubID: "ub0", ifName: "enp0s1"}})
		convey.Convey("Then the device is found", func() {
			convey.So(rs.findDeviceByIfName("enp0s1"), convey.ShouldNotBeNil)
		})
		convey.Convey("Then an unknown ifname is not found", func() {
			convey.So(rs.findDeviceByIfName("enp0s9"), convey.ShouldBeNil)
		})
	})
}

// TestPreferredDeviceIDs tests preferredDeviceIDs for zero size, NPU mapping priority and failure propagation.
func TestPreferredDeviceIDs(t *testing.T) {
	rs := newNpuServer([]types.Device{&ubDevice{ubID: "ub0", ifName: "enp0s1"}})
	convey.Convey("Given a zero allocation size, no ids are returned", t, func() {
		ids, err := rs.preferredDeviceIDs(context.Background(), &pluginapi.ContainerPreferredAllocationRequest{AllocationSize: 0})
		convey.So(err, convey.ShouldBeNil)
		convey.So(ids, convey.ShouldBeNil)
	})
	convey.Convey("Given a successful NPU mapping, preferred ids come first", t, func() {
		mockPreferredEnv(t, []string{"enp0s1"}, nil, true)
		req := &pluginapi.ContainerPreferredAllocationRequest{AllocationSize: 4, MustIncludeDeviceIDs: []string{"x"}, AvailableDeviceIDs: []string{"ub0", "a", "b"}}
		ids, err := rs.preferredDeviceIDs(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ids, convey.ShouldResemble, []string{"x", "ub0", "a", "b"})
	})
	convey.Convey("Given mapped ids not in the available set, preferred is skipped", t, func() {
		mockPreferredEnv(t, []string{"enp0s1"}, nil, true)
		req := &pluginapi.ContainerPreferredAllocationRequest{AllocationSize: 4, MustIncludeDeviceIDs: []string{"x"}, AvailableDeviceIDs: []string{"a", "b"}}
		ids, err := rs.preferredDeviceIDs(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ids, convey.ShouldResemble, []string{"x", "a", "b"})
	})
	convey.Convey("Given no pending pods, only the requested ids are returned", t, func() {
		mockPreferredEnv(t, nil, nil, false)
		req := &pluginapi.ContainerPreferredAllocationRequest{AllocationSize: 2, MustIncludeDeviceIDs: []string{"x"}, AvailableDeviceIDs: []string{"a"}}
		ids, err := rs.preferredDeviceIDs(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ids, convey.ShouldResemble, []string{"x", "a"})
	})
	convey.Convey("Given an NPU mapping failure, the error is propagated", t, func() {
		mockPreferredEnv(t, nil, errTest, true)
		req := &pluginapi.ContainerPreferredAllocationRequest{AllocationSize: 1, AvailableDeviceIDs: []string{"a"}}
		_, err := rs.preferredDeviceIDs(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestStringSet(t *testing.T) {
	convey.Convey("Given a device id slice", t, func() {
		set := stringSet([]string{"ub0", "ub1", ""})
		convey.Convey("Then the set is built without empty ids", func() {
			convey.So(len(set), convey.ShouldEqual, 2)
			convey.So(set["ub0"], convey.ShouldNotBeNil)
			convey.So(set["ub1"], convey.ShouldNotBeNil)
		})
	})
}

func TestSetSubset(t *testing.T) {
	convey.Convey("Given a proper subset", t, func() {
		sub := stringSet([]string{"a"})
		super := stringSet([]string{"a", "b"})
		convey.Convey("Then it is a subset", func() {
			convey.So(setSubset(sub, super), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given a non-subset", t, func() {
		sub := stringSet([]string{"c"})
		super := stringSet([]string{"a", "b"})
		convey.Convey("Then it is not a subset", func() {
			convey.So(setSubset(sub, super), convey.ShouldBeFalse)
		})
	})
}

// TestAllocateByNpu tests allocateByNpu client checks, pod state handling and allocation recording.
func TestAllocateByNpu(t *testing.T) {
	convey.Convey("Given a nil k8s client, an error is returned", t, func() {
		rs := &ubResourceServer{resourceName: testResourceName}
		convey.So(rs.allocateByNpu(context.Background(), []string{"ub0"}), convey.ShouldNotBeNil)
	})
	convey.Convey("Given a GetNodeName failure, an error is returned", t, func() {
		patches := gomonkey.ApplyFuncReturn(utils.GetNodeName, "", errTest)
		defer patches.Reset()
		convey.So(newNpuServer(nil).allocateByNpu(context.Background(), []string{"ub0"}), convey.ShouldNotBeNil)
	})
	convey.Convey("Given various pod states, allocation behaves accordingly", t, func() {
		tests := []struct {
			withPod   bool
			pod       v1.Pod
			clientset kubernetes.Interface
			wantErr   bool
		}{
			{true, newNpuPod(map[string]string{deviceStatusAnnotation: `{"inference":["ub0"]}`}), nil, false},
			{true, newNpuPod(nil), nil, true},
			{true, newNpuPod(map[string]string{deviceStatusAnnotation: `{"inference":["ub1"]}`}), nil, false},
			{false, v1.Pod{}, nil, false},
			{true, newNpuPod(map[string]string{api.HuaweiNPU: "0"}), fake.NewSimpleClientset(), true},
		}
		for _, tt := range tests {
			pods := []v1.Pod{}
			if tt.withPod {
				pods = []v1.Pod{tt.pod}
			}
			mockAllocateEnv(t, pods, tt.clientset)
			err := newNpuServer(nil).allocateByNpu(context.Background(), []string{"ub0"})
			if tt.wantErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		}
	})
	convey.Convey("Given a pending pod, the allocation is recorded", t, func() {
		pod := newNpuPod(map[string]string{api.HuaweiNPU: "0"})
		f := fake.NewSimpleClientset(&pod)
		mockAllocateEnv(t, []v1.Pod{pod}, f)
		convey.So(newNpuServer(nil).allocateByNpu(context.Background(), []string{"ub0"}), convey.ShouldBeNil)
		got, getErr := f.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		convey.So(getErr, convey.ShouldBeNil)
		var status map[string][]string
		convey.So(json.Unmarshal([]byte(got.Annotations[deviceStatusAnnotation]), &status), convey.ShouldBeNil)
		convey.So(status[testContainer], convey.ShouldResemble, []string{"ub0"})
	})
}

var errTest = errors.New("test error")

func ts(sec int64) time.Time { return time.Unix(sec, 0) }

func podWithNodeName(node string, phase v1.PodPhase) v1.Pod {
	return podWithNodeNameAndResource(node, phase, true)
}

func podWithNodeNameAndResource(node string, phase v1.PodPhase, withResource bool) v1.Pod {
	pod := newNpuPod(nil)
	pod.Name = node + "-" + string(phase)
	pod.Spec.NodeName = node
	pod.Status.Phase = phase
	if !withResource {
		pod.Spec.Containers[0].Resources.Requests = nil
	}
	return pod
}

func newKubeletPodsEndpoint(t *testing.T, handler http.HandlerFunc, cfgErr error) {
	srv := httptest.NewServer(handler)
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(kubeletPodsURL, srv.URL, nil)
	if cfgErr != nil {
		patches.ApplyFuncReturn(rest.InClusterConfig, nil, cfgErr)
	} else {
		patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{BearerToken: "token"}, nil)
	}
	t.Cleanup(func() {
		patches.Reset()
		srv.Close()
	})
}

func mockPreferredEnv(t *testing.T, nicNames []string, nicErr error, withPending bool) {
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(utils.GetNodeName, testNodeName, nil)
	patches.ApplyFuncReturn(getNpuPodsOnNode, []v1.Pod{})
	if withPending {
		patches.ApplyFuncReturn(getNpuPodsOnNode, []v1.Pod{newNpuPod(map[string]string{api.HuaweiNPU: "0"})})
		patches.ApplyFuncReturn(getPodNpuIds, []int{0}, nil)
		// npuMappedDeviceIDs is unexported and cannot be patched by gomonkey, so mock its dependency instead.
		patches.ApplyFuncReturn(utils.GetNicNames, nicNames, nicErr)
	}
	t.Cleanup(patches.Reset)
}

func mockAllocateEnv(t *testing.T, pods []v1.Pod, clientset kubernetes.Interface) {
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(utils.GetNodeName, testNodeName, nil)
	patches.ApplyFuncReturn(getNpuPodsOnNode, pods)
	if clientset != nil {
		patches.ApplyMethodReturn((*kubernetes.Clientset)(nil), "CoreV1", clientset.CoreV1())
	}
	t.Cleanup(patches.Reset)
}
