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
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        testPod,
				updateTime: time.Now(),
			},
		}
		UpdatePodList(nil, testPod, EventTypeAdd)
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
	testPod.Namespace = "testPod1"
	testPod.Namespace = "testNamespace1"
	convey.Convey("test update pod list when operator is EventTypeUpdate", t, func() {
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        testPod,
				updateTime: time.Now(),
			},
		}
		UpdatePodList(nil, testPod, EventTypeUpdate)
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
	convey.Convey("test update pod list when operator is EventTypeUpdate", t, func() {
		UpdatePodList(nil, testPod, EventTypeDelete)
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
	convey.Convey("test update pod list when operator is default", t, func() {
		UpdatePodList(nil, testPod, "default")
		convey.So(podCache, convey.ShouldResemble, make(map[types.UID]*podInfo))
	})
}

// TestRefreshPodList test get pod list by field selector with cache
func TestRefreshPodList(t *testing.T) {
	client, _ := newTestClientK8s()
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
		expectPodCache := map[types.UID]*podInfo{
			testPod.UID: {
				Pod:        &testPod,
				updateTime: time.Now(),
			},
		}
		client.refreshPodList()
		convey.So(client.IsApiErr, convey.ShouldBeFalse)
		convey.So(podCache, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetAllPodListCache test get pod list by field selector with cache
func TestGetAllPodListCache(t *testing.T) {
	expectPodCache := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "default",
			},
		},
	}
	client, _ := newTestClientK8s()
	convey.Convey("test get pod list by field selector with cache", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "default",
					},
				},
			},
		}
		testPodList := client.GetAllPodListCache()
		convey.So(testPodList, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetActivePodListCache01 test get active pod list with cache
func TestGetActivePodListCache01(t *testing.T) {
	expectPodCache := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "default",
			},
		},
	}
	client, _ := newTestClientK8s()
	convey.Convey("test get active pod list when pod name err", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "errorName",
						Namespace: "default",
					},
				},
			},
		}
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
					},
				},
			},
		}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, expectPodCache)
	})
}

// TestGetActivePodListCache02 test get active pod list with cache
func TestGetActivePodListCache02(t *testing.T) {
	client, _ := newTestClientK8s()
	convey.Convey("test get active pod list when pod namespace err", t, func() {
		podCache = map[types.UID]*podInfo{
			"testPod": {
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "errorNamespace",
					},
				},
			},
		}
		testPodList := client.GetActivePodListCache()
		convey.So(testPodList, convey.ShouldResemble, make([]v1.Pod, 0, common.GeneralMapSize))
	})
}

// TestGetNodeServerIDCache test case for get server id
func TestGetNodeServerIDCache(t *testing.T) {
	patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
	}, nil)
	defer patch.Reset()
	client, _ := newTestClientK8s()
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
	patch := gomonkey.ApplyMethodReturn(&ClientK8s{}, "GetNode", &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
	}, nil)
	defer patch.Reset()
	client, _ := newTestClientK8s()
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
		client, _ := newTestClientK8s()
		usage, err := client.GetServerUsageLabelCache()
		fmt.Printf("usage: %s\n", usage)
		convey.So(usage == common.Infer, convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCheckPodInCache01 test case for check pod in cache
func TestCheckPodInCache01(t *testing.T) {
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "xxxxxxxxx1",
			Namespace: "default",
			Name:      "pod1",
		},
	}
	patch1 := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPod", func(_ *ClientK8s,
		_ context.Context, _, _ string) (*v1.PodList, error) {
		return &v1.PodList{Items: []v1.Pod{*pod1}}, nil
	})
	defer patch1.Reset()
	convey.Convey("test check pod in cache", t, func() {
		client, _ := newTestClientK8s()
		pod1UpdateTime := time.Now().Add(-time.Hour).Add(-time.Minute)
		expectNewPodCache := map[types.UID]*podInfo{}
		podCache = map[types.UID]*podInfo{
			"xxxxxxxxx1": {
				Pod: &v1.Pod{
					Spec: v1.PodSpec{},
				},
				updateTime: pod1UpdateTime,
			},
		}
		client.checkPodInCache(context.TODO())
		convey.ShouldEqual(podCache, expectNewPodCache)
	})
}

// TestCheckPodInCache02 test case for check pod in cache
func TestCheckPodInCache02(t *testing.T) {
	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "xxxxxxxxx2",
			Namespace: "default",
			Name:      "pod2",
		},
	}
	patch1 := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPod", func(_ *ClientK8s,
		_ context.Context, _, _ string) (*v1.PodList, error) {
		return &v1.PodList{Items: []v1.Pod{*pod2}}, nil
	})
	defer patch1.Reset()
	convey.Convey("test check pod in cache", t, func() {
		client, _ := newTestClientK8s()
		pod2UpdateTime := time.Now().Add(-time.Minute)
		expectNewPodCache := map[types.UID]*podInfo{
			"xxxxxxxxx2": {
				Pod:        &v1.Pod{},
				updateTime: pod2UpdateTime,
			},
		}
		podCache = map[types.UID]*podInfo{
			"xxxxxxxxx2": {
				Pod:        &v1.Pod{},
				updateTime: pod2UpdateTime,
			},
		}
		client.checkPodInCache(context.TODO())
		convey.ShouldEqual(podCache, expectNewPodCache)
	})
}
