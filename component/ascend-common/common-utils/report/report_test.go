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

// Package report for test
package report

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/version"
)

type (
	mockStatusWriter      struct{}
	mockSubResourceClient struct{}
)

type mockClient struct {
	createdObj *corev1.ConfigMap
	patchedObj *corev1.ConfigMap
	createCall int
	patchCall  int
}

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func (m *mockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (m *mockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (m *mockSubResourceClient) Get() error {
	return nil
}

func (m *mockSubResourceClient) Create() error {
	return nil
}

func (m *mockSubResourceClient) Update() error {
	return nil
}

func (m *mockSubResourceClient) Patch() error {
	return nil
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return nil
}

func (m *mockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (m *mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.createCall++
	cm, ok := obj.(*corev1.ConfigMap)
	if ok {
		m.createdObj = cm
	}
	return nil
}

func (m *mockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (m *mockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (m *mockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	m.patchCall++
	cm, ok := obj.(*corev1.ConfigMap)
	if ok {
		m.patchedObj = cm
	}
	return nil
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (m *mockClient) Status() client.StatusWriter {
	return &mockStatusWriter{}
}

func (m *mockClient) Scheme() *runtime.Scheme {
	return nil
}

func (m *mockClient) RESTMapper() meta.RESTMapper {
	var rm meta.RESTMapper
	return rm
}

func TestReportVersionToConfigMap(t *testing.T) {
	ctx := context.Background()
	testInfo := version.Info{
		Version:   "v1.0.0",
		GitCommit: "abc123",
		GitBranch: "main",
		BuildOS:   "linux",
		BuildArch: "amd64",
		GoVersion: "go1.21",
	}
	testComponent := "test-component"

	convey.Convey("Test ReportVersionToConfigMap", t, func() {
		convey.Convey("create new configmap successfully on first attempt", func() {
			mc := &mockClient{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})

			ReportVersionToConfigMap(mc, ctx, testInfo, testComponent)

			convey.So(mc.createCall, convey.ShouldEqual, 1)
			convey.So(mc.patchCall, convey.ShouldEqual, 0)
			convey.So(mc.createdObj, convey.ShouldNotBeNil)
			convey.So(mc.createdObj.Name, convey.ShouldEqual, api.VersionName)
			convey.So(mc.createdObj.Namespace, convey.ShouldEqual, api.DLNamespace)
			convey.So(mc.createdObj.Data[testComponent], convey.ShouldNotBeEmpty)
		})

		convey.Convey("configmap already exists, patch successfully on first attempt", func() {
			mc := &mockClient{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethod(mc, "Create",
				func(_ *mockClient, _ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return k8serrors.NewAlreadyExists(schema.GroupResource{}, api.VersionName)
				})
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})

			ReportVersionToConfigMap(mc, ctx, testInfo, testComponent)

			convey.So(mc.createCall, convey.ShouldEqual, 0)
			convey.So(mc.patchCall, convey.ShouldEqual, 1)
			convey.So(mc.patchedObj, convey.ShouldNotBeNil)
			convey.So(mc.patchedObj.Name, convey.ShouldEqual, api.VersionName)
			convey.So(mc.patchedObj.Namespace, convey.ShouldEqual, api.DLNamespace)
		})

		convey.Convey("create fails with non-AlreadyExists error, stop retrying", func() {
			mc := &mockClient{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			internalErr := k8serrors.NewInternalError(errors.New("create failed"))
			patches.ApplyMethod(mc, "Create",
				func(_ *mockClient, _ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return internalErr
				})
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})

			ReportVersionToConfigMap(mc, ctx, testInfo, testComponent)

			convey.So(mc.createCall, convey.ShouldEqual, 0)
			convey.So(mc.patchCall, convey.ShouldEqual, 0)
		})

		convey.Convey("all 3 attempts fail with AlreadyExists and patch failure", func() {
			mc := &mockClient{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethod(mc, "Create",
				func(_ *mockClient, _ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return k8serrors.NewAlreadyExists(schema.GroupResource{}, api.VersionName)
				})
			patches.ApplyMethod(mc, "Patch",
				func(_ *mockClient, _ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
					return errors.New("patch always fails")
				})
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})

			ReportVersionToConfigMap(mc, ctx, testInfo, testComponent)

			convey.So(mc.createCall, convey.ShouldEqual, 0)
			convey.So(mc.patchCall, convey.ShouldEqual, 0)
		})
	})
}
