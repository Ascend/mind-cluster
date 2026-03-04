/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"infer-operator/pkg/common"
)

var (
	two   = 2
	three = 3
)

type mockLogger struct{}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
}

func TestAddInferOperatorCfgEventFuncs(t *testing.T) {
	convey.Convey("Test AddInferOperatorCfgEventFuncs", t, func() {
		defer func() {
			inferOperatorCfgEventFuncs = []func(interface{}, interface{}, string){}
		}()

		convey.Convey("add one event func", func() {
			AddInferOperatorCfgEventFuncs(func(_, _ interface{}, _ string) {})
			convey.So(len(inferOperatorCfgEventFuncs), convey.ShouldEqual, 1)
		})

		convey.Convey("add two event func", func() {
			AddInferOperatorCfgEventFuncs(func(_, _ interface{}, _ string) {})
			AddInferOperatorCfgEventFuncs(func(_, _ interface{}, _ string) {})
			convey.So(len(inferOperatorCfgEventFuncs), convey.ShouldEqual, two)
		})
	})
}

func TestFilterInferOperatorCfg(t *testing.T) {
	convey.Convey("Test filterInferOperatorCfg", t, func() {
		convey.Convey("when object is not a ConfigMap", func() {
			obj := "not a configmap"
			result := filterInferOperatorCfg(obj)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("when ConfigMap has correct namespace and name", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(getCurrentNamespace, func() string {
				return common.DefaultNamespace
			})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      common.InferOperatorCfgName,
					Namespace: common.DefaultNamespace,
				},
			}
			result := filterInferOperatorCfg(cm)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("when ConfigMap has wrong name", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(getCurrentNamespace, func() string {
				return common.DefaultNamespace
			})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-name",
					Namespace: common.DefaultNamespace,
				},
			}
			result := filterInferOperatorCfg(cm)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestInferOperatorCfgHandler(t *testing.T) {
	convey.Convey("Test inferOperatorCfgHandler", t, func() {
		defer func() {
			inferOperatorCfgEventFuncs = []func(interface{}, interface{}, string){}
		}()

		convey.Convey("when no event funcs registered", func() {
			inferOperatorCfgHandler(nil, nil, common.AddOperator)
			convey.So(len(inferOperatorCfgEventFuncs), convey.ShouldEqual, 0)
		})

		convey.Convey("when event funcs registered, should execute", func() {
			isCalled := false
			AddInferOperatorCfgEventFuncs(func(_, _ interface{}, _ string) {
				isCalled = true
			})
			inferOperatorCfgHandler(nil, nil, common.AddOperator)
			convey.So(isCalled, convey.ShouldBeTrue)
		})

		convey.Convey("when multiple event funcs registered", func() {
			callCount := 0
			AddInferOperatorCfgEventFuncs(
				func(_, _ interface{}, _ string) { callCount++ },
				func(_, _ interface{}, _ string) { callCount++ },
			)
			inferOperatorCfgHandler(nil, nil, common.AddOperator)
			convey.So(callCount, convey.ShouldEqual, two)
		})

		convey.Convey("when newObj is nil, should still execute", func() {
			isCalled := false
			AddInferOperatorCfgEventFuncs(func(_, _ interface{}, _ string) {
				isCalled = true
			})
			inferOperatorCfgHandler(nil, nil, common.DeleteOperator)
			convey.So(isCalled, convey.ShouldBeTrue)
		})

		convey.Convey("when oldObj and newObj both exist", func() {
			oldData := map[string]string{"key": "old"}
			newData := map[string]string{"key": "new"}
			oldCm := &corev1.ConfigMap{Data: oldData}
			newCm := &corev1.ConfigMap{Data: newData}

			var receivedOld, receivedNew interface{}
			AddInferOperatorCfgEventFuncs(func(old, new interface{}, _ string) {
				receivedOld = old
				receivedNew = new
			})
			inferOperatorCfgHandler(oldCm, newCm, common.UpdateOperator)
			convey.So(receivedOld, convey.ShouldNotBeNil)
			convey.So(receivedNew, convey.ShouldNotBeNil)
		})
	})
}

func TestAddInferOperatorCfgEventHandler(t *testing.T) {
	convey.Convey("Test addInferOperatorCfgEventHandler", t, func() {
		convey.Convey("when cmInformer is not nil", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			called := false
			mockInformerImpl := &mockInformer{}
			var mockInformer cache.Informer = mockInformerImpl

			patches.ApplyMethodFunc(mockInformerImpl, "AddEventHandler", func(h toolscache.ResourceEventHandler) {
				called = true
			})

			addInferOperatorCfgEventHandler(&mockInformer)
			convey.So(called, convey.ShouldBeTrue)
		})

		convey.Convey("handler filter func should work correctly", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			filterCalled := false
			patches.ApplyFunc(filterInferOperatorCfg, func(obj interface{}) bool {
				filterCalled = true
				return true
			})

			mockInformerImpl := &mockInformer{}
			var mockInformer cache.Informer = mockInformerImpl
			var capturedHandler toolscache.ResourceEventHandler

			patches.ApplyMethodFunc(mockInformerImpl, "AddEventHandler", func(h toolscache.ResourceEventHandler) {
				capturedHandler = h
			})

			addInferOperatorCfgEventHandler(&mockInformer)

			if capturedHandler != nil {
				testCM := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.InferOperatorCfgName,
						Namespace: common.DefaultNamespace,
					},
				}
				capturedHandler.OnAdd(testCM)
			}
			convey.So(filterCalled, convey.ShouldBeTrue)
		})
	})
}

func TestFilterInferOperatorCfgWithRealData(t *testing.T) {
	convey.Convey("Test filterInferOperatorCfg with real ConfigMap data", t, func() {
		convey.Convey("when ConfigMap matches InferOperator criteria", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(getCurrentNamespace, func() string {
				return common.DefaultNamespace
			})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      common.InferOperatorCfgName,
					Namespace: common.DefaultNamespace,
				},
				Data: map[string]string{
					"config": "test-data",
				},
			}
			result := filterInferOperatorCfg(cm)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("when ConfigMap namespace differs", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(getCurrentNamespace, func() string {
				return "correct-namespace"
			})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      common.InferOperatorCfgName,
					Namespace: "different-namespace",
				},
			}
			result := filterInferOperatorCfg(cm)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("when ConfigMap name differs", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(getCurrentNamespace, func() string {
				return common.DefaultNamespace
			})

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "different-name",
					Namespace: common.DefaultNamespace,
				},
			}
			result := filterInferOperatorCfg(cm)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestAddInferOperatorCfgEventFuncsMultiple(t *testing.T) {
	convey.Convey("Test AddInferOperatorCfgEventFuncs with multiple additions", t, func() {
		defer func() {
			inferOperatorCfgEventFuncs = []func(interface{}, interface{}, string){}
		}()

		convey.Convey("add funcs in batches", func() {
			AddInferOperatorCfgEventFuncs(
				func(_, _ interface{}, _ string) {},
				func(_, _ interface{}, _ string) {},
			)
			convey.So(len(inferOperatorCfgEventFuncs), convey.ShouldEqual, two)

			AddInferOperatorCfgEventFuncs(
				func(_, _ interface{}, _ string) {},
			)
			convey.So(len(inferOperatorCfgEventFuncs), convey.ShouldEqual, three)
		})

		convey.Convey("all registered funcs should be called", func() {
			callCount := 0
			AddInferOperatorCfgEventFuncs(
				func(_, _ interface{}, _ string) { callCount++ },
				func(_, _ interface{}, _ string) { callCount++ },
				func(_, _ interface{}, _ string) { callCount++ },
			)
			inferOperatorCfgHandler(nil, nil, common.AddOperator)
			convey.So(callCount, convey.ShouldEqual, three)
		})
	})
}

// mockCache implements cache.Cache interface for testing
type mockCache struct {
	shouldError bool
}

func (m *mockCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (cache.Informer, error) {
	if m.shouldError {
		return nil, fmt.Errorf("test error")
	}
	informer := &mockInformer{}
	return informer, nil
}

func (m *mockCache) GetInformer(ctx context.Context, obj client.Object) (cache.Informer, error) {
	return nil, nil
}

func (m *mockCache) Start(ctx context.Context) error {
	return nil
}

func (m *mockCache) WaitForCacheSync(ctx context.Context) bool {
	return true
}

func (m *mockCache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

func (m *mockCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return nil
}

func (m *mockCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

// mockInformer implements cache.Informer interface for testing
type mockInformer struct {
	handlers []toolscache.ResourceEventHandler
}

func (m *mockInformer) AddEventHandler(handler toolscache.ResourceEventHandler) {
	m.handlers = append(m.handlers, handler)
}

func (m *mockInformer) AddEventHandlerWithResyncPeriod(handler toolscache.ResourceEventHandler, resyncPeriod time.Duration) {
	m.handlers = append(m.handlers, handler)
}

func (m *mockInformer) HasSynced() bool {
	return true
}

func (m *mockInformer) LastSyncResourceVersion() string {
	return ""
}

func (m *mockInformer) AddIndexers(indexers toolscache.Indexers) error {
	return nil
}
