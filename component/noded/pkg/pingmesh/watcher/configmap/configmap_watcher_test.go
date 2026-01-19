/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package configmap is using for watching configmap changes.
*/

package configmap

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"nodeD/pkg/kubeclient"
	_ "nodeD/pkg/testtool"
)

func TestNewWatcher(t *testing.T) {
	convey.Convey("Testing NewWatcher", t, func() {
		client := &kubeclient.ClientK8s{
			ClientSet: fake.NewSimpleClientset(),
		}
		options := []Option{
			WithLabelSector("test-label"),
			WithNamespace("test-namespace"),
		}

		watcher := NewWatcher(client, options...)

		convey.So(watcher, convey.ShouldNotBeNil)
		convey.So(watcher.client, convey.ShouldEqual, client)
		convey.So(watcher.labelSelector, convey.ShouldEqual, "test-label")
		convey.So(watcher.namespace, convey.ShouldEqual, "test-namespace")
		convey.So(watcher.handlers, convey.ShouldNotBeNil)
		convey.So(watcher.queue, convey.ShouldNotBeNil)
	})

}

func TestInit(t *testing.T) {
	convey.Convey("Test Init", t, func() {
		watcher := NewWatcher(&kubeclient.ClientK8s{
			ClientSet: fake.NewSimpleClientset(),
		})
		watcher.Init()

		convey.So(watcher.informerFactory, convey.ShouldNotBeNil)
		convey.So(watcher.indexer, convey.ShouldNotBeNil)
	})
}

func TestWatch(t *testing.T) {
	convey.Convey("Test Watch", t, func() {
		var expected atomic.Int32
		client := fake.NewSimpleClientset()
		watcher := NewWatcher(&kubeclient.ClientK8s{
			ClientSet: client,
		}, WithNamespace("test-namespace"), WithNamedHandlers(NamedHandler{
			Name: "test",
			Handle: func(cm *corev1.ConfigMap) {
				expected.Add(1)
			},
		}))
		watcher.Init()

		stopCh := make(chan struct{})
		cm := &corev1.ConfigMap{}
		cm.Namespace = "test-namespace"
		cm.Name = "test"
		_, err := client.CoreV1().ConfigMaps("test-namespace").Create(context.TODO(), cm, metav1.CreateOptions{})
		convey.So(err, convey.ShouldBeNil)
		go watcher.Watch(stopCh)
		time.Sleep(time.Second)
		close(stopCh)
		convey.So(expected.Load(), convey.ShouldEqual, 1)
	})
}
