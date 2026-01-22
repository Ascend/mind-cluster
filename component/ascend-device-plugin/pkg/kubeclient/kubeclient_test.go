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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-helpers/node/util"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
)

// TestNewClientK8s test create k8s client
func TestNewClientK8s(t *testing.T) {
	convey.Convey("test create k8s client when build client config error", t, func() {
		mockBuildConfig := gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil,
			fmt.Errorf("build config error"))
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
	convey.Convey("test create k8s client when get node name from env error", t, func() {
		mockGetNodeNameFromEnv := gomonkey.ApplyFuncReturn(GetNodeNameFromEnv, nil,
			fmt.Errorf("get node name from env error"))
		defer mockGetNodeNameFromEnv.Reset()
		client, err := NewClientK8s()
		convey.So(client, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get node name from env error")
	})
	convey.Convey("test create k8s client success", t, func() {
		nodeName := os.Getenv(api.NodeNameEnv)
		mockCheckNodeName := gomonkey.ApplyFuncReturn(checkNodeName, nil)
		defer mockCheckNodeName.Reset()
		client, err := NewClientK8s()
		expectclient := &ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       nodeName,
			DeviceInfoName: common.DeviceInfoCMNamePrefix + nodeName,
			Queue:          client.Queue,
			IsApiErr:       false,
			KltClient:      mockKltClient(),
		}
		convey.So(client, convey.ShouldResemble, expectclient)
		convey.So(err, convey.ShouldBeNil)
	})
}

func mockKltClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			CipherSuites:       defaultSafeCipherSuites,
			MinVersion:         tls.VersionTLS13,
		},
	}
	return &http.Client{Transport: transport}
}

// TestGetNode test get node
func TestGetNode(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetNode init kubernetes failed")
	}
	convey.Convey("test get node success", t, func() {
		mockGetNode := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Nodes(), "Get", &v1.Node{}, nil)
		defer mockGetNode.Reset()
		node, err := client.GetNode()
		convey.So(node, convey.ShouldResemble, &v1.Node{})
		convey.So(err, convey.ShouldEqual, nil)
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

// TestGetPod test get pod by namespace and name
func TestGetPod(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetPod init kubernetes failed")
	}
	convey.Convey("test get pod failed when param pod is nil", t, func() {
		pod, err := client.GetPod(nil)
		convey.So(pod, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "param pod is nil")
	})
	testPod := getMockPod(api.HuaweiAscend910, npuChip910PhyID0)
	convey.Convey("test get pod success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll), "Get",
			&v1.Pod{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		pod, err := client.GetPod(testPod)
		convey.So(pod, convey.ShouldResemble, &v1.Pod{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
	})
}

// TestPatchPod test patch pod information
func TestPatchPod(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestPatchPod init kubernetes failed")
	}
	testPod := getMockPod(api.HuaweiAscend910, npuChip910PhyID0)
	convey.Convey("test patch pod information success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll), "Patch",
			&v1.Pod{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		pod, err := client.PatchPod(testPod, []byte{})
		convey.So(pod, convey.ShouldResemble, &v1.Pod{})
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
			_ *ClientK8s) (*v1.PodList, error) {
			return &v1.PodList{}, nil
		})
		defer mockPodListByCondition.Reset()
		pods, err := client.GetActivePodList()
		convey.So(pods, convey.ShouldResemble, []v1.Pod{})
		convey.So(err, convey.ShouldBeNil)
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
	convey.Convey("test get pod list by field selector when pod list count invalid", t, func() {
		mockPodListByCondition := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPodListByCondition", func(
			_ *ClientK8s) (*v1.PodList, error) {
			return &v1.PodList{Items: make([]v1.Pod, common.MaxPodLimit+1)}, nil
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

// TestGetPodListByCondition test get pod list by field selector
func TestGetPodListByCondition(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetPodListByCondition init kubernetes failed")
	}
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": "NodeName"})
	convey.Convey("test get pod list success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Pods(v1.NamespaceAll), "List",
			&v1.PodList{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		podList, err := client.getPodListByCondition(selector)
		convey.So(podList, convey.ShouldResemble, &v1.PodList{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
	})
}

// TestCheckPodList test check each pod and return podList
func TestCheckPodList(t *testing.T) {
	testPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "testUid",
			Name:      "name",
			Namespace: "namespace",
		},
	}
	testPodList := &v1.PodList{Items: []v1.Pod{testPod}}
	convey.Convey("test check each pod when pod list is invalid", t, func() {
		pods, err := checkPodList(nil)
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "pod list is invalid")
	})
	convey.Convey("test check each pod when the number of pods exceeds the upper limit", t, func() {
		testPodList := &v1.PodList{Items: make([]v1.Pod, common.MaxPodLimit+1)}
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "the number of pods exceeds the upper limit")
	})
	convey.Convey("test check each pod when check pod name err", t, func() {
		testPod.Name = "Name"
		testPodList := &v1.PodList{Items: []v1.Pod{testPod}}
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldResemble, make([]v1.Pod, 0))
		convey.So(err, convey.ShouldBeNil)
		testPod.Name = "name"
	})
	convey.Convey("test check each pod when check pod namespace err", t, func() {
		testPod.Namespace = "Namespace"
		testPodList := &v1.PodList{Items: []v1.Pod{testPod}}
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldResemble, make([]v1.Pod, 0))
		convey.So(err, convey.ShouldBeNil)
		testPod.Namespace = "namespace"
	})
	mockCheckPodNameAndSpace := gomonkey.ApplyFuncReturn(common.CheckPodNameAndSpace, nil)
	defer mockCheckPodNameAndSpace.Reset()
	convey.Convey("test check each pod success", t, func() {
		pods, err := checkPodList(testPodList)
		convey.So(pods, convey.ShouldResemble, testPodList.Items)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCreateConfigMap test create device info which is cm
func TestCreateConfigMap(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestCreateConfigMap init kubernetes failed")
	}
	mockCM := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
	convey.Convey("test create config map failed when param cm is nil", t, func() {
		cm, err := client.CreateConfigMap(nil)
		convey.So(cm, convey.ShouldEqual, nil)
		convey.So(err.Error(), convey.ShouldEqual, "param cm is nil")
	})
	convey.Convey("test create config map success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().ConfigMaps(v1.NamespaceAll),
			"Create",
			&v1.ConfigMap{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		cm, err := client.CreateConfigMap(mockCM)
		convey.So(cm, convey.ShouldResemble, &v1.ConfigMap{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
	})
}

// TestGetConfigMap test get config map by name and namespace
func TestGetConfigMap(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestGetConfigMap init kubernetes failed")
	}
	mockCM := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
	convey.Convey("test get config map success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().ConfigMaps(v1.NamespaceAll), "Get",
			&v1.ConfigMap{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		cm, err := client.GetConfigMap(mockCM.Name, mockCM.Namespace)
		convey.So(cm, convey.ShouldResemble, &v1.ConfigMap{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
	})
}

// TestUpdateConfigMap test update device info which is cm
func TestUpdateConfigMap(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestUpdateConfigMap init kubernetes failed")
	}
	mockCM := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
	convey.Convey("test update device info failed when param cm is nil", t, func() {
		cm, err := client.UpdateConfigMap(nil)
		convey.So(cm, convey.ShouldEqual, nil)
		convey.So(err.Error(), convey.ShouldEqual, "param cm is nil")
	})
	convey.Convey("test update device info which is cm success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().ConfigMaps(v1.NamespaceAll),
			"Update",
			&v1.ConfigMap{}, fmt.Errorf(common.ApiServerPort))
		defer mockGetPod.Reset()
		cm, err := client.UpdateConfigMap(mockCM)
		convey.So(cm, convey.ShouldResemble, &v1.ConfigMap{})
		convey.So(err.Error(), convey.ShouldEqual, common.ApiServerPort)
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
		mockWriteResetInfoDataIntoCM := gomonkey.ApplyMethodReturn(&ClientK8s{}, "WriteResetInfoDataIntoCM", testCM,
			nil)
		defer mockWriteResetInfoDataIntoCM.Reset()
		err := client.ClearResetInfo("taskName", "testNamespace")
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCreateEvent test create event resource
func TestCreateEvent(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestCreateEvent init kubernetes failed")
	}
	mockEvent := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
	convey.Convey("test create event resource failed when param event is nil", t, func() {
		cm, err := client.CreateEvent(nil)
		convey.So(cm, convey.ShouldEqual, nil)
		convey.So(err.Error(), convey.ShouldEqual, "param event is nil")
	})
	convey.Convey("test update device info which is cm success", t, func() {
		mockGetPod := gomonkey.ApplyMethodReturn((&kubernetes.Clientset{}).CoreV1().Events(v1.NamespaceAll), "Create",
			&v1.Event{}, nil)
		defer mockGetPod.Reset()
		event, err := client.CreateEvent(mockEvent)
		convey.So(event, convey.ShouldResemble, &v1.Event{})
		convey.So(err, convey.ShouldEqual, nil)
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
	convey.Convey(`test check node name failed when node name is ""`, t, func() {
		nodeName := ""
		err := checkNodeName(nodeName)
		convey.So(err.Error(), convey.ShouldEqual, "the env variable whose key is NODE_NAME must be set")
	})
	convey.Convey("test check node name failed when node name length bigger than KubeEnvMaxLength", t, func() {
		nodeName := strings.Repeat("a", common.KubeEnvMaxLength+1)
		err := checkNodeName(nodeName)
		convey.So(err.Error(), convey.ShouldEqual, "node name length 231 is bigger than 230")
	})
	convey.Convey("test check node name failed when node name testName is illegal", t, func() {
		nodeName := "testName"
		err := checkNodeName(nodeName)
		convey.So(err.Error(), convey.ShouldEqual, "node name testName is illegal")
	})
	convey.Convey("test check node name success", t, func() {
		nodeName := "name"
		err := checkNodeName(nodeName)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

// TestResourceEventHandler test handle the configmap resource event
func TestResourceEventHandler(t *testing.T) {
	client, err := newTestClientK8s()
	if err != nil {
		t.Fatal("TestResourceEventHandler init kubernetes failed")
	}
	testObj := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{api.HuaweiAscend910: "test"}}}
	testOldObj := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{api.HuaweiAscend910: "testOld"}}}
	client.Queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	convey.Convey("test handle the configmap resource event when resource type is pod", t, func() {
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
	convey.Convey("test handle the configmap resource event when resource type is cm", t, func() {
		mockMetaNamespaceKeyFunc := gomonkey.ApplyFuncReturn(cache.MetaNamespaceKeyFunc, "testKey", fmt.Errorf("error"))
		defer mockMetaNamespaceKeyFunc.Reset()
		handler := client.ResourceEventHandler(CMResource, checkPod)
		expectHandler := cache.FilteringResourceEventHandler{
			FilterFunc: checkPod,
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    handler.OnAdd,
				UpdateFunc: handler.OnUpdate,
				DeleteFunc: handler.OnDelete,
			},
		}
		handler.OnAdd(testObj)
		handler.OnUpdate(testObj, testObj)
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

// TestAddAnnotation test add annotation
func TestAddAnnotation(t *testing.T) {
	nodeName := "test-node"
	newTestClient := func(node *v1.Node, patchErrs []error) *ClientK8s {
		clientset := fake.NewSimpleClientset(node)
		actionIndex := 0
		clientset.Fake.PrependReactor("patch", "nodes",
			func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				defer func() { actionIndex++ }()
				if actionIndex < len(patchErrs) {
					return true, nil, patchErrs[actionIndex]
				}
				return true, node, nil
			})
		return &ClientK8s{
			Clientset: clientset,
			NodeName:  nodeName,
		}
	}

	convey.Convey("Given a ClientK8s instance for testing AddAnnotation", t, func() {
		baseNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        nodeName,
				Annotations: map[string]string{},
			},
		}

		convey.Convey("When all patch requests fail", func() {
			patchErrs := []error{fmt.Errorf("api server timeout"), fmt.Errorf("connection refused"),
				fmt.Errorf("network error")}
			ki := newTestClient(baseNode, patchErrs)
			err := ki.AddAnnotation("anno", "val")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, patchErrs[len(patchErrs)-1].Error())
		})

		convey.Convey("When first patch request succeeds", func() {
			ki := newTestClient(baseNode, nil)
			err := ki.AddAnnotation("anno", "val")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("When retry then succeed", func() {
			patchErrs := []error{fmt.Errorf("first error"), fmt.Errorf("second error"), nil}
			ki := newTestClient(baseNode, patchErrs)
			err := ki.AddAnnotation("anno", "val")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestResetDeviceInfo(t *testing.T) {
	// after test reset -> recover
	mockNodeDeviceInfoCache := nodeDeviceInfoCache
	mock := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)),
		"WriteDeviceInfoDataIntoCM", func(_ *ClientK8s,
			nodeDeviceData *common.NodeDeviceInfoCache, manuallySeparateNPU string,
			_ common.SwitchFaultInfo, _ common.DpuInfo) (*common.NodeDeviceInfoCache, error) {
			return nodeDeviceData, nil
		})
	defer func() {
		nodeDeviceInfoCache = mockNodeDeviceInfoCache
		mock.Reset()
	}()
	convey.Convey("case 1: A5 reset device info cache", t, func() {
		client, err := newTestClientK8s()
		if err != nil {
			t.Fatal("Test reset device info cache: init kubernetes failed")
		}
		mockRealCardType := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = api.Ascend910A5
		defer func() {
			common.ParamOption.RealCardType = mockRealCardType
		}()
		client.ResetDeviceInfo()
		convey.So(*nodeDeviceInfoCache.RackID, convey.ShouldEqual, -1)
	})

	convey.Convey("case 1: not A5 reset device info cache", t, func() {
		client, err := newTestClientK8s()
		if err != nil {
			t.Fatal("Test reset device info cache: init kubernetes failed")
		}
		mockRealCardType := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = api.Ascend910
		defer func() {
			common.ParamOption.RealCardType = mockRealCardType
		}()
		client.ResetDeviceInfo()
		convey.So(nodeDeviceInfoCache.RackID, convey.ShouldBeNil)
	})
}
