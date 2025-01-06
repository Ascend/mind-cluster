/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

package kubeclient

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"Ascend-device-plugin/pkg/common"
)

func newTestClientK8s() (*ClientK8s, error) {
	return &ClientK8s{
		Clientset:      &kubernetes.Clientset{},
		NodeName:       "node",
		DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
		IsApiErr:       false,
	}, nil
}

// TestUpdatePodList test update pod list by informer
func TestUpdatePodList(t *testing.T) {
	testPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "testUid",
			Name:      "testPod",
			Namespace: "testNamespace",
		},
	}
	podCache = make(map[types.UID]*podInfo)
	convey.Convey("test update pod list when operator is EventTypeAdd", t, func() {
		UpdatePodList(testPod, EventTypeAdd)
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        testPod,
				updateTime: podCache[testPod.UID].updateTime,
			},
		}
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
	testPod.Namespace = "testPod1"
	testPod.Namespace = "testNamespace1"
	convey.Convey("test update pod list when operator is EventTypeUpdate", t, func() {
		UpdatePodList(testPod, EventTypeUpdate)
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        testPod,
				updateTime: podCache[testPod.UID].updateTime,
			},
		}
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
	convey.Convey("test update pod list when operator is EventTypeDelete", t, func() {
		UpdatePodList(testPod, EventTypeDelete)
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
	convey.Convey("test update pod list when operator is default", t, func() {
		UpdatePodList(testPod, "default")
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
	convey.Convey("test update pod list failed when newPod is not ok", t, func() {
		UpdatePodList(nil, "default")
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
}

// TestRefreshPodList test get pod list by field selector with cache
func TestRefreshPodList(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestRefreshPodList init kubernetes failed")
	}
	testPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: "testUid",
		},
	}
	convey.Convey("test get pod list when podList is empty", t, func() {
		client.IsApiErr = true
		mockGetAllPodList := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetAllPodList",
			nil, fmt.Errorf("podList is empty"))
		defer mockGetAllPodList.Reset()
		client.refreshPodList()
		convey.So(client.IsApiErr, convey.ShouldBeTrue)
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
	convey.Convey("test get pod list by field selector with cache", t, func() {
		client.IsApiErr = true
		mockGetAllPodList := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetAllPodList",
			&v1.PodList{
				Items: []v1.Pod{
					testPod,
				},
			}, nil)
		defer mockGetAllPodList.Reset()
		client.refreshPodList()
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        &testPod,
				updateTime: podCache[testPod.UID].updateTime,
			},
		}
		convey.So(client.IsApiErr, convey.ShouldBeFalse)
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetAllPodListCache test get pod list by field selector with cache
func TestGetAllPodListCache(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetAllPodListCache init kubernetes failed")
	}
	expectPodCache := []v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		}}}
	convey.Convey("test get pod list by field selector with cache", t, func() {
		client.IsApiErr = true
		mockRefreshPodList := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "refreshPodList", func(_ *ClientK8s) { return })
		defer mockRefreshPodList.Reset()
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "default",
					}}}}
		testPodList := client.GetAllPodListCache()
		convey.So(testPodList, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetActivePodListCache01 test get active pod list with cache
func TestGetActivePodListCache01(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetActivePodListCache01 init kubernetes failed")
	}
	expectPodCache := []v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		}}}
	client.IsApiErr = true
	mockRefreshPodList := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "refreshPodList", func(_ *ClientK8s) { return })
	defer mockRefreshPodList.Reset()
	convey.Convey("test get active pod list when pod name err", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "errorName",
						Namespace: "default",
					}}}}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, make([]v1.Pod, 0, common.GeneralMapSize))
	})
	convey.Convey("test get active pod list with cache", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "default",
					}}}}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetActivePodListCache02 test get active pod list with cache
func TestGetActivePodListCache02(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetActivePodListCache02 init kubernetes failed")
	}
	convey.Convey("test get active pod list when pod namespace err", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "errorNamespace",
					}}}}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, make([]v1.Pod, 0, common.GeneralMapSize))
	})
	convey.Convey("test get active pod list when pod status is failed or succeeded", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "default",
					},
					Status: v1.PodStatus{Phase: v1.PodFailed},
				}}}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, make([]v1.Pod, 0, common.GeneralMapSize))
	})
}

// TestGetNodeServerIDCache test case for get server id
func TestGetNodeServerIDCache(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetNodeServerIDCache init kubernetes failed")
	}
	convey.Convey("test get server id failed", t, func() {
		patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{
			Status: v1.NodeStatus{Addresses: make([]v1.NodeAddress, common.MaxPodLimit+1)}}, nil)
		defer patch.Reset()
		id, err := client.GetNodeServerIDCache()
		convey.So(id, convey.ShouldEqual, "")
		convey.So(err.Error(), convey.ShouldEqual, "the number of node status in exceeds the upper limit")
	})
	patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
	}, nil)
	defer patch.Reset()
	convey.Convey("test server id", t, func() {
		nodeServerIp = "test server id"
		id, err := client.GetNodeServerIDCache()
		convey.So(id, convey.ShouldEqual, "test server id")
		convey.So(err, convey.ShouldBeNil)
		nodeServerIp = ""
	})
	convey.Convey("test no server id", t, func() {
		id, err := client.GetNodeServerIDCache()
		convey.So(id, convey.ShouldEqual, "")
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetServerUsageLabelCache test case for get server usage
func TestGetServerUsageLabelCache(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetServerUsageLabelCache init kubernetes failed")
	}
	convey.Convey("test get node failed", t, func() {
		patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{}, fmt.Errorf("get node error"))
		defer patch.Reset()
		usage, err := client.GetServerUsageLabelCache()
		convey.So(usage, convey.ShouldEqual, "")
		convey.So(err.Error(), convey.ShouldEqual, "get node error")
	})
	patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
	}, nil)
	defer patch.Reset()
	convey.Convey("test usage label", t, func() {
		serverUsageLabel = "test usage label"
		usage, err := client.GetServerUsageLabelCache()
		convey.So(usage, convey.ShouldEqual, "test usage label")
		convey.So(err, convey.ShouldBeNil)
		serverUsageLabel = ""
	})
	convey.Convey("test no usage label", t, func() {
		usage, err := client.GetServerUsageLabelCache()
		convey.So(usage == "unknown", convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetA800IA2Label test case for get a800 ia2 label
func TestGetA800IA2Label(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetA800IA2Label init kubernetes failed")
	}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "node",
		},
	}
	node.Labels[common.ServerUsageLabelKey] = common.Infer
	patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", node, nil)
	defer patch.Reset()
	convey.Convey("test usage label with infer", t, func() {
		serverUsageLabel = ""
		usage, err := client.GetServerUsageLabelCache()
		fmt.Printf("usage: %s\n", usage)
		convey.So(usage == common.Infer, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCheckPodInCache01 test case for check pod in cache
func TestCheckPodInCache01(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestCheckPodInCache01 init kubernetes failed")
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "uid",
			Namespace: "default",
			Name:      "pod",
		}}
	podUpdateTime := time.Now().Add(-time.Hour).Add(-time.Minute)
	expectNewPodCache := map[types.UID]*podInfo{}
	convey.Convey("test check pod in cache", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll), "Get", pod, nil)
		defer mockGetPod.Reset()
		podCache = map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{Spec: v1.PodSpec{}},
				updateTime: podUpdateTime,
			}}
		client.checkPodInCache(context.TODO())
		convey.So(podCache, convey.ShouldResemble, expectNewPodCache)
	})
	mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll),
		"Get", pod, fmt.Errorf(common.ApiServerPort))
	defer mockGetPod.Reset()
	convey.Convey("test check pod in cache when get pod error", t, func() {
		podCache = map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{Spec: v1.PodSpec{}},
				updateTime: podUpdateTime,
			}}
		client.checkPodInCache(context.TODO())
		convey.ShouldEqual(podCache, expectNewPodCache)
	})
	convey.Convey("test check pod in cache when get pod error is not found", t, func() {
		mockIsNotFound := gomonkey.ApplyFuncReturn(errors.IsNotFound, true)
		defer mockIsNotFound.Reset()
		podCache = map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{Spec: v1.PodSpec{}},
				updateTime: podUpdateTime,
			}}
		client.checkPodInCache(context.TODO())
		convey.ShouldEqual(podCache, expectNewPodCache)
	})
}

// TestCheckPodInCache02 test case for check pod in cache
func TestCheckPodInCache02(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestCheckPodInCache02 init kubernetes failed")
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "uid",
			Namespace: "default",
			Name:      "pod",
		},
		Spec: v1.PodSpec{NodeName: "node"}}
	mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll), "Get", pod, nil)
	defer mockGetPod.Reset()
	convey.Convey("test check pod in cache when update time lower than pod cache timeout", t, func() {
		podUpdateTime := time.Now().Add(-time.Minute)
		expectNewPodCache := map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{},
				updateTime: podUpdateTime,
			}}
		podCache = map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{},
				updateTime: podUpdateTime,
			}}
		client.checkPodInCache(context.TODO())
		convey.So(podCache, convey.ShouldResemble, expectNewPodCache)
	})
	convey.Convey("test check pod in cache when need refresh", t, func() {
		podUpdateTime := time.Now().Add(-time.Hour).Add(-time.Minute)
		podCache = map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{},
				updateTime: podUpdateTime,
			}}
		client.checkPodInCache(context.TODO())
		expectNewPodCache := map[types.UID]*podInfo{
			pod.UID: {
				Pod:        &v1.Pod{},
				updateTime: podCache[pod.UID].updateTime,
			}}
		convey.So(podCache, convey.ShouldResemble, expectNewPodCache)
	})
}

// TestGetDeviceInfoCMCache test get device info configMap with cache
func TestGetDeviceInfoCMCache(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetDeviceInfoCMCache init kubernetes failed")
	}
	convey.Convey("test get device info configMap with cache success", t, func() {
		var mockCache *common.NodeDeviceInfoCache
		cache := client.GetDeviceInfoCMCache()
		convey.So(cache, convey.ShouldResemble, mockCache)
	})
}

// TestSetNodeDeviceInfoCache test set device info cache
func TestSetNodeDeviceInfoCache(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestSetNodeDeviceInfoCache init kubernetes failed")
	}
	convey.Convey("test set device info cache success", t, func() {
		mockNodeDeviceInfoCache := nodeDeviceInfoCache
		client.SetNodeDeviceInfoCache(&common.NodeDeviceInfoCache{})
		convey.So(nodeDeviceInfoCache, convey.ShouldResemble, &common.NodeDeviceInfoCache{})
		nodeDeviceInfoCache = mockNodeDeviceInfoCache
	})
}
