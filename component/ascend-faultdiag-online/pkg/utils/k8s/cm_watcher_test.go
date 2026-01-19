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

// Package k8s is a DT collection for func in cm_watcher.go
package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	testNamespace1 = "testNamespace1"
	testCmName1    = "testCmName1"
	testNamespace2 = "testNamespace2"
	testCmName2    = "testCmName2"

	subCount = 10
)

func TestSubAndUnSub(t *testing.T) {
	GetCmWatcher()
	var count = 0
	patch := gomonkey.ApplyPrivateMethod(
		CmWatcher,
		"runInformer",
		func(*cmWatcher, context.Context, string, string) {
			count++
		},
	)
	defer patch.Reset()
	var f = func(oldObj, newObj *corev1.ConfigMap, op watch.EventType) {}
	var registerIds []string
	convey.Convey("test sub and unsub", t, func() {
		for i := 0; i < subCount; i++ {
			registerIds = append(registerIds, CmWatcher.Subscribe(testNamespace1, testNamespace1, f))
		}
		for i := 0; i < subCount; i++ {
			registerIds = append(registerIds, CmWatcher.Subscribe(testNamespace2, testNamespace2, f))
		}
		var length = 2
		convey.So(len(CmWatcher.callbackMap), convey.ShouldEqual, length)
		convey.So(len(CmWatcher.watchers), convey.ShouldEqual, length)
		convey.So(len(registerIds), convey.ShouldEqual, length*subCount)
		time.Sleep(time.Millisecond)
		convey.So(count, convey.ShouldEqual, length)
		for _, registerId := range registerIds {
			CmWatcher.Unsubscribe(testNamespace1, testNamespace1, registerId)
			CmWatcher.Unsubscribe(testNamespace2, testNamespace2, registerId)
		}
		convey.So(len(CmWatcher.callbackMap), convey.ShouldEqual, 0)
		convey.So(len(CmWatcher.watchers), convey.ShouldEqual, 0)
	})
}

func TestCmProcessor(t *testing.T) {
	GetCmWatcher()
	patch := gomonkey.ApplyPrivateMethod(
		CmWatcher,
		"runInformer",
		func(*cmWatcher, context.Context, string, string) {},
	)
	defer patch.Reset()
	convey.Convey("test cm processor", t, func() {
		// convert to configMap failed
		wrongOldData := "xxx"
		wrongNewData := "xxx"
		CmWatcher.cmProcessor(wrongOldData, wrongNewData, watch.Added)
		CmWatcher.cmProcessor(nil, wrongNewData, watch.Added)
		newData := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCmName1,
				Namespace: testNamespace1,
			},
		}
		CmWatcher.cmProcessor(nil, newData, watch.Added)
		time.Sleep(time.Millisecond)

		// sub
		f := func(oldObj, newObj *corev1.ConfigMap, op watch.EventType) {
			assert.Equal(t, newObj.Name, testCmName1)
			assert.Equal(t, newObj.Namespace, testNamespace1)
		}

		registerId := CmWatcher.Subscribe(testNamespace1, testCmName1, f)
		time.Sleep(time.Millisecond)
		// unsub no data in storage
		CmWatcher.Unsubscribe(testNamespace1, testCmName1, registerId)
		_, ok := storage.Load(CmWatcher.keyGenerator(testNamespace1, testCmName1))
		convey.So(ok, convey.ShouldBeFalse)
	})
}
