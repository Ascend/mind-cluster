/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"Ascend-device-plugin/pkg/common"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-helpers/node/util"
	"os"
	"reflect"
	"testing"
)

// TestNewClientK8s test create k8s client
func TestNewClientK8s(t *testing.T) {
	convey.Convey("test create k8s client when build client config error", t, func() {
		mockBuildConfig := gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil, fmt.Errorf("build config error"))
		defer mockBuildConfig.Reset()
		client, err := NewClientK8s()
		convey.So(client, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "build config error")
	})
	mockBuildConfig := gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, &rest.Config{UserAgent: ""}, nil)
	defer mockBuildConfig.Reset()
	convey.Convey("test create k8s client when get client error", t, func() {
		mockNewForConfig := gomonkey.ApplyFuncReturn(kubernetes.NewForConfig, nil, fmt.Errorf("get client error"))
		defer mockNewForConfig.Reset()
		client, err := NewClientK8s()
		convey.So(client, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get client error")
	})
	mockNewForConfig := gomonkey.ApplyFuncReturn(kubernetes.NewForConfig, &kubernetes.Clientset{}, nil)
	defer mockNewForConfig.Reset()
	convey.Convey("test create k8s client success", t, func() {
		nodeName := os.Getenv("NODE_NAME")
		mockCheckNodeName := gomonkey.ApplyFuncReturn(checkNodeName, nil)
		defer mockCheckNodeName.Reset()
		client, err := NewClientK8s()
		expectclient := &ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       nodeName,
			DeviceInfoName: common.DeviceInfoCMNamePrefix + nodeName,
			Queue:          client.Queue,
			IsApiErr:       false,
		}
		convey.So(client, convey.ShouldResemble, expectclient)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestPatchNodeState test patch node state
func TestPatchNodeState(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestPatchNodeState init kubernetes failed")
	}
	convey.Convey("test patch node state when patch node status error", t, func() {
		mockPatchNodeStatus := gomonkey.ApplyFuncReturn(util.PatchNodeStatus, &v1.Node{}, []byte{},
			fmt.Errorf(common.ApiServerPort))
		defer mockPatchNodeStatus.Reset()
		node, patchBytes, err := client.PatchNodeState(&v1.Node{}, &v1.Node{})
		convey.So(node, convey.ShouldResemble, &v1.Node{})
		convey.So(patchBytes, convey.ShouldResemble, []byte{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
	})
}

// TestGetActivePodList test get active pod list
func TestGetActivePodList(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetActivePodList init kubernetes failed")
	}
	convey.Convey("test get active pod list when selector parse error", t, func() {
		mockParseSelector := gomonkey.ApplyFuncReturn(fields.ParseSelector, nil, fmt.Errorf("selector parse error"))
		defer mockParseSelector.Reset()
		pods, err := client.GetActivePodList()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "selector parse error")
	})
	mockParseSelector := gomonkey.ApplyFuncReturn(fields.ParseSelector, fields.Nothing(), nil)
	defer mockParseSelector.Reset()
	convey.Convey("test get active pod list when getPodListByCondition error", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) (*v1.PodList, error) {
			return &v1.PodList{}, fmt.Errorf("getPodListByCondition error")
		})
		defer mockPodListByCondition.Reset()
		pods, err := client.GetActivePodList()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "getPodListByCondition error")
	})
	convey.Convey("test get active pod list when the number of pods exceeds the upper limit", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) ([]v1.Pod, error) {
			return []v1.Pod{}, nil
		})
		defer mockPodListByCondition.Reset()
		pods, err := client.GetActivePodList()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "the number of pods exceeds the upper limit")
	})
}

// TestGetAllPodList test get pod list by field selector
func TestGetAllPodList(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetActivePodList init kubernetes failed")
	}
	mockSelectorFromSet := gomonkey.ApplyFuncReturn(fields.SelectorFromSet, fields.Nothing())
	defer mockSelectorFromSet.Reset()
	convey.Convey("test get pod list by field selector when getPodListByCondition error", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) (*v1.PodList, error) {
			return &v1.PodList{}, fmt.Errorf("getPodListByCondition error")
		})
		defer mockPodListByCondition.Reset()
		podList, err := client.GetAllPodList()
		convey.So(podList, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "getPodListByCondition error")
	})
	convey.Convey("test get pod list by field selector when the number of pods exceeds the upper limit", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) ([]v1.Pod, error) {
			return []v1.Pod{}, nil
		})
		defer mockPodListByCondition.Reset()
		podList, err := client.GetAllPodList()
		convey.So(podList, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "pod list count invalid")
	})
	convey.Convey("test get pod list by field selector success", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) (*v1.PodList, error) {
			return &v1.PodList{}, nil
		})
		defer mockPodListByCondition.Reset()
		podList, err := client.GetAllPodList()
		convey.So(podList, convey.ShouldResemble, &v1.PodList{})
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCheckPodList test check each pod and return podList
func TestCheckPodList(t *testing.T) {
	testPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "testUid",
			Name:      "testName",
			Namespace: "testNamespace",
		},
	}
	testPodList := &v1.PodList{Items: []v1.Pod{testPod}}
	convey.Convey("test check each pod when pod list is invalid", t, func() {
		pods, err := checkPodList(nil)
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "pod list is invalid")
	})
	convey.Convey("test check each pod when check pod name and space err", t, func() {
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldResemble, make([]v1.Pod, 0))
		convey.So(err, convey.ShouldBeNil)
	})
	mockCheckPodNameAndSpace := gomonkey.ApplyFuncReturn(common.CheckPodNameAndSpace, nil)
	defer mockCheckPodNameAndSpace.Reset()
	convey.Convey("test check each pod success", t, func() {
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldResemble, testPodList.Items)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestClearResetInfo test clear reset info
func TestClearResetInfo(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestClearResetInfo init kubernetes failed")
	}
	testCM := getMockCreateCM(common.ResetInfoCMDataKey, common.ResetInfoCMNamePrefix+"node")
	defer testCM.Reset()
	convey.Convey("test clear reset info when write reset info data into cm error", t, func() {
		mockWriteResetInfoDataIntoCM := gomonkey.ApplyMethodReturn(&ClientK8s{},
			"WriteResetInfoDataIntoCM", testCM, fmt.Errorf("write reset info data into cm error"))
		defer mockWriteResetInfoDataIntoCM.Reset()
		err := client.ClearResetInfo("taskName", "testNamespace")
		convey.So(err.Error(), convey.ShouldEqual, "write reset info data into cm error")
	})
	convey.Convey("test clear reset info success", t, func() {
		mockWriteResetInfoDataIntoCM := gomonkey.ApplyMethodReturn(&ClientK8s{}, "WriteResetInfoDataIntoCM", testCM, nil)
		defer mockWriteResetInfoDataIntoCM.Reset()
		err := client.ClearResetInfo("taskName", "testNamespace")
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetNodeNameFromEnv test get current node name from env
func TestGetNodeNameFromEnv(t *testing.T) {
	convey.Convey("test get current node name from env when check node name error", t, func() {
		mockCheckNodeName := gomonkey.ApplyFuncReturn(checkNodeName, fmt.Errorf("checkNodeName error"))
		defer mockCheckNodeName.Reset()
		nodeName, err := GetNodeNameFromEnv()
		convey.So(nodeName, convey.ShouldEqual, "")
		convey.So(err.Error(), convey.ShouldEqual, "check node name failed: checkNodeName error")
	})
}

// TestCheckNodeName test check node name
func TestCheckNodeName(t *testing.T) {
	convey.Convey("test check node name when node name is \"\"", t, func() {
		nodeName := ""
		err := checkNodeName(nodeName)
		convey.So(err.Error(), convey.ShouldEqual, "the env variable whose key is NODE_NAME must be set")
	})
	convey.Convey("test check node name when node name testName is illegal", t, func() {
		nodeName := "testName"
		err := checkNodeName(nodeName)
		convey.So(err.Error(), convey.ShouldEqual, "node name testName is illegal")
	})
}

// TestResourceEventHandler test handle the configmap resource event
func TestResourceEventHandler(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestResourceEventHandler init kubernetes failed")
	}
	convey.Convey("test handle the configmap resource event", t, func() {
		testObj := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{common.HuaweiAscend910: "test"}}}
		testOldObj := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{common.HuaweiAscend910: "testOld"}}}
		client.Queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		mockMetaNamespaceKeyFunc := gomonkey.ApplyFuncReturn(cache.MetaNamespaceKeyFunc, "testKey", nil)
		defer mockMetaNamespaceKeyFunc.Reset()
		mockDeepEqual := gomonkey.ApplyFuncReturn(reflect.DeepEqual, false)
		defer mockDeepEqual.Reset()
		handler := client.ResourceEventHandler(PodResource, checkPod)
		expectHandler := cache.FilteringResourceEventHandler{
			FilterFunc: checkPod,
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    handler.OnAdd,
				UpdateFunc: handler.OnUpdate,
				DeleteFunc: handler.OnDelete,
			},
		}
		handler.OnAdd(testObj)
		handler.OnUpdate(testOldObj, testObj)
		handler.OnDelete(testObj)
		convey.So(handler, convey.ShouldHaveSameTypeAs, expectHandler)
		convey.So(handler, convey.ShouldNotResemble, expectHandler)
	})
}

// TestFlushPodCacheNextQuerying test flush cache
func TestFlushPodCacheNextQuerying(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestFlushPodCacheNextQuerying init kubernetes failed")
	}
	convey.Convey("test flush pod cache success", t, func() {
		client.FlushPodCacheNextQuerying()
		convey.So(client.IsApiErr, convey.ShouldBeTrue)
	})
}
