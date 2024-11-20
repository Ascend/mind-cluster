/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// TestGetServerUsageLabelCache01 test case for get pod has timeout
func TestCheckPodInCache01(t *testing.T) {
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "xxxxxxxxx1",
			Namespace: "default",
			Name:      "pod1",
		},
	}
	patch := gomonkey.
		ApplyFuncReturn(NewClientK8s, &ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       "node",
			DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
			IsApiErr:       false,
		}, nil)
	defer patch.Reset()

	patch1 := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPod", func(_ *ClientK8s,
		_ context.Context, _, _ string) (*v1.PodList, error) {
		return &v1.PodList{Items: []v1.Pod{*pod1}}, nil
	})
	defer patch1.Reset()
	convey.Convey("test check pod in cache", t, func() {
		client, _ := NewClientK8s()
		pod1UpdateTime := time.Now().Add(-time.Hour).Add(-time.Minute)
		expectNewPodCache := map[types.UID]*podInfo{}
		podCache = map[types.UID]*podInfo{
			"xxxxxxxxx1": {
				Pod:        &v1.Pod{},
				updateTime: pod1UpdateTime,
			},
		}
		client.checkPodInCache(context.TODO())
		convey.ShouldEqual(podCache, expectNewPodCache)
	})
}

// TestGetServerUsageLabelCache02 test case for get pod has not timeout
func TestCheckPodInCache02(t *testing.T) {
	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "xxxxxxxxx2",
			Namespace: "default",
			Name:      "pod2",
		},
	}
	patch := gomonkey.
		ApplyFuncReturn(NewClientK8s, &ClientK8s{
			Clientset:      &kubernetes.Clientset{},
			NodeName:       "node",
			DeviceInfoName: common.DeviceInfoCMNamePrefix + "node",
			IsApiErr:       false,
		}, nil)
	defer patch.Reset()

	patch1 := gomonkey.ApplyPrivateMethod(&ClientK8s{}, "getPod", func(_ *ClientK8s,
		_ context.Context, _, _ string) (*v1.PodList, error) {
		return &v1.PodList{Items: []v1.Pod{*pod2}}, nil
	})
	defer patch1.Reset()
	convey.Convey("test check pod in cache", t, func() {
		client, _ := NewClientK8s()
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
