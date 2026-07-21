/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/common"
)

func TestGetChannel(t *testing.T) {
	client := &ClientK8s{}
	originParamOption := common.ParamOption
	defer func() { common.ParamOption = originParamOption }()

	convey.Convey("default channel should get pods from kubelet", t, func() {
		common.ParamOption = common.Option{GetPodFromKubelet: true}
		_, ok := client.getChannel().(*Kubelet)
		convey.So(ok, convey.ShouldBeTrue)
	})

	convey.Convey("disabled switch should get pods from apiserver", t, func() {
		common.ParamOption = common.Option{GetPodFromKubelet: false}
		_, ok := client.getChannel().(*Apiserver)
		convey.So(ok, convey.ShouldBeTrue)
	})
}

func TestKubeletGetPod(t *testing.T) {
	client := &ClientK8s{NodeName: "node"}
	kubelet := &Kubelet{client: client}
	podList := &v1.PodList{Items: []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"},
			Spec:       v1.PodSpec{NodeName: "node"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "default"},
			Spec:       v1.PodSpec{NodeName: "node"},
		},
	}}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) {
			return podList, nil
		})
	defer patch.Reset()

	convey.Convey("kubelet channel should get pod from local pod list", t, func() {
		pod, err := kubelet.GetPod(context.Background(), &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default"}})
		convey.So(err, convey.ShouldBeNil)
		convey.So(pod.Name, convey.ShouldEqual, "target")
	})
}

func TestKubeletGetPodList(t *testing.T) {
	client := &ClientK8s{NodeName: "node"}
	kubelet := &Kubelet{client: client}
	podList := &v1.PodList{Items: []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "running", Namespace: "default"},
			Spec:       v1.PodSpec{NodeName: "node"},
			Status:     v1.PodStatus{Phase: v1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "succeeded", Namespace: "default"},
			Spec:       v1.PodSpec{NodeName: "node"},
			Status:     v1.PodStatus{Phase: v1.PodSucceeded},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "other-node", Namespace: "default"},
			Spec:       v1.PodSpec{NodeName: "other"},
			Status:     v1.PodStatus{Phase: v1.PodRunning},
		},
	}}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) {
			return podList, nil
		})
	defer patch.Reset()

	convey.Convey("kubelet channel should filter pod list by current node", t, func() {
		pods, err := kubelet.GetAllPodList()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(pods.Items), convey.ShouldEqual, 2)
	})

	convey.Convey("kubelet channel should filter active pod list", t, func() {
		pods, err := kubelet.GetActivePodList()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(pods), convey.ShouldEqual, 1)
		convey.So(pods[0].Name, convey.ShouldEqual, "running")
	})
}

func TestKubeletInitPodInformer(t *testing.T) {
	// InitPodInformer starts a background poll goroutine whose stopCh is an
	// inline make(chan struct{}) with no retained reference, so the loop cannot
	// be stopped from the test. After this test returns, defer patch.Reset()
	// restores the real getPodsByKltPort and a leaked poll could reach it and
	// call ki.KltClient.Do(req). A non-nil KltClient turns that into a benign
	// connection error (logged) instead of a nil-pointer panic that crashes the
	// whole test binary.
	client := &ClientK8s{KltClient: &http.Client{}}
	kubelet := &Kubelet{client: client}

	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) { return &v1.PodList{}, nil })
	defer patch.Reset()

	convey.Convey("kubelet channel should create a kubelet-backed pod informer", t, func() {
		kubelet.InitPodInformer()
		convey.So(client.PodInformer, convey.ShouldNotBeNil)
		// it should be a kubeletPodInformer, not an apiserver informer
		_, ok := client.PodInformer.(*kubeletPodInformer)
		convey.So(ok, convey.ShouldBeTrue)
	})
}
