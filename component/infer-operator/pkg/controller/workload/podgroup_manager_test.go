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

package workload

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"infer-operator/pkg/common"
)

// TestNewVolcanoPodGroupManager tests the NewVolcanoPodGroupManager function.
func TestNewVolcanoPodGroupManager(t *testing.T) {
	convey.Convey("Test NewVolcanoPodGroupManager function", t, func() {
		convey.Convey("Should create a new VolcanoPodGroupManager", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			convey.So(manager, convey.ShouldNotBeNil)
			convey.So(volcanoManager, convey.ShouldNotBeNil)
			convey.So(volcanoManager.client, convey.ShouldEqual, fakeClient)
		})
	})
}

// TestVolcanoPodGroupManagerGetPodGroupForInstance tests the GetPodGroupForInstance method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerGetPodGroupForInstance(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager GetPodGroupForInstance method", t, func() {
		convey.Convey("Should successfully get PodGroup", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			podGroup := &v1beta1.PodGroup{
				ObjectMeta: v1.ObjectMeta{
					Name:      common.GetPGNameFromIndexer(indexer),
					Namespace: instanceSet.Namespace,
				},
			}
			fakeClient := NewFakeClient().WithObjects(podGroup).Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			ctx := context.Background()
			result, err := manager.GetPodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.Name, convey.ShouldEqual, common.GetPGNameFromIndexer(indexer))
		})

		convey.Convey("Should return not found error when PodGroup does not exist", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			ctx := context.Background()
			_, err := manager.GetPodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when getting PodGroup fails", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			mockErr := errors.New("failed to get podgroup")
			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Get", mockErr)
			defer patches.Reset()
			ctx := context.Background()
			_, err := manager.GetPodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "failed to get podgroup")
		})
	})
}

// TestVolcanoPodGroupManagerDeletePodGroupForInstance tests the DeletePodGroupForInstance
// method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerDeletePodGroupForInstance(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager DeletePodGroupForInstance method", t, func() {
		convey.Convey("Should successfully delete PodGroup", func() {
			instanceSet := CreateTestInstanceSet("test-instance", "default", int32(1))
			indexer := GetTestIndexer("test-service", "test-role", "0")
			podGroup := &v1beta1.PodGroup{
				ObjectMeta: v1.ObjectMeta{
					Name:      common.GetPGNameFromIndexer(indexer),
					Namespace: instanceSet.Namespace,
				},
			}
			fakeClient := NewFakeClient().WithObjects(podGroup).Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			patches2 := gomonkey.ApplyMethodReturn(volcanoManager.client, "Delete", nil)
			defer patches2.Reset()

			ctx := context.Background()
			err := manager.DeletePodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should skip deletion when PodGroup does not exist", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")

			ctx := context.Background()
			err := manager.DeletePodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestVolcanoPodGroupManagerDeletePodGroupForInstance2 tests the DeletePodGroupForInstance
// method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerDeletePodGroupForInstance2(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager DeletePodGroupForInstance method", t, func() {
		convey.Convey("Should return error when deleting PodGroup fails", func() {
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			podGroup := &v1beta1.PodGroup{
				ObjectMeta: v1.ObjectMeta{
					Name:      common.GetPGNameFromIndexer(indexer),
					Namespace: instanceSet.Namespace,
				},
			}
			fakeClient := NewFakeClient().WithObjects(podGroup).Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			mockErr := errors.New("failed to delete podgroup")
			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Delete", mockErr)
			defer patches.Reset()

			ctx := context.Background()
			err := manager.DeletePodGroupForInstance(ctx, instanceSet, indexer)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "failed to delete podgroup")
		})
	})
}

// TestVolcanoPodGroupManagerGetOrCreatePodGroupForInstance tests the GetOrCreatePodGroupForInstance
// method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerGetOrCreatePodGroupForInstance(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager GetOrCreatePodGroupForInstance method", t, func() {
		convey.Convey("Should return existing PodGroup when it exists", func() {
			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			podGroup := &v1beta1.PodGroup{
				ObjectMeta: v1.ObjectMeta{
					Name:      common.GetPGNameFromIndexer(indexer),
					Namespace: instanceSet.Namespace,
				},
			}
			fakeClient := NewFakeClient().WithObjects(podGroup).Build()
			manager := NewVolcanoPodGroupManager(fakeClient)

			ctx := context.Background()
			exists, err := manager.GetOrCreatePodGroupForInstance(ctx, instanceSet, indexer, v1beta1.PodGroupSpec{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(exists, convey.ShouldBeTrue)
		})

		convey.Convey("Should create PodGroup when it does not exist", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Create", nil)
			defer patches.Reset()

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			exists, err := manager.GetOrCreatePodGroupForInstance(ctx, instanceSet, indexer, v1beta1.PodGroupSpec{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(exists, convey.ShouldBeFalse)
		})
	})
}

// TestVolcanoPodGroupManagerGetOrCreatePodGroupForInstance2 tests the GetOrCreatePodGroupForInstance
// method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerGetOrCreatePodGroupForInstance2(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager GetOrCreatePodGroupForInstance method", t, func() {
		convey.Convey("Should return error when getting PodGroup fails", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			mockErr := errors.New("failed to get podgroup")
			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Get", mockErr)
			defer patches.Reset()

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			exists, err := manager.GetOrCreatePodGroupForInstance(ctx, instanceSet, indexer, v1beta1.PodGroupSpec{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
			convey.So(exists, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when creating PodGroup fails", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldBeTrue)

			mockErr := errors.New("failed to create podgroup")
			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Create", mockErr)
			defer patches.Reset()

			replicas := int32(1)
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			exists, err := manager.GetOrCreatePodGroupForInstance(ctx, instanceSet, indexer, v1beta1.PodGroupSpec{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(exists, convey.ShouldBeFalse)
		})
	})
}

// TestVolcanoPodGroupManagerCreatePodGroupForInstance tests the createPodGroupForInstance
// method of VolcanoPodGroupManager.
func TestVolcanoPodGroupManagerCreatePodGroupForInstance(t *testing.T) {
	convey.Convey("Test VolcanoPodGroupManager createPodGroupForInstance method", t, func() {
		convey.Convey("Should successfully create PodGroup", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldNotBeNil)

			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Create", nil)
			defer patches.Reset()

			replicas := int32(1)
			spec := v1beta1.PodGroupSpec{
				MinMember: replicas,
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := volcanoManager.createPodGroupForInstance(ctx, instanceSet, indexer, spec)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should not add inferserviceid to PodGroup labels", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldNotBeNil)

			replicas := int32(1)
			spec := v1beta1.PodGroupSpec{
				MinMember: replicas,
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			instanceSet.Labels[common.InferServiceIDLabelKey] = "test-uid"
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := volcanoManager.createPodGroupForInstance(ctx, instanceSet, indexer, spec)
			convey.So(err, convey.ShouldBeNil)
			created := &v1beta1.PodGroup{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Name: common.GetPGNameFromIndexer(indexer), Namespace: "default"}, created)
			convey.So(err, convey.ShouldBeNil)
			_, exist := created.Labels[common.InferServiceIDLabelKey]
			convey.So(exist, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when creating PodGroup fails", func() {
			fakeClient := NewFakeClient().Build()
			manager := NewVolcanoPodGroupManager(fakeClient)
			volcanoManager, ok := manager.(*VolcanoPodGroupManager)
			convey.So(ok, convey.ShouldNotBeNil)

			mockErr := errors.New("failed to create podgroup")
			patches := gomonkey.ApplyMethodReturn(volcanoManager.client, "Create", mockErr)
			defer patches.Reset()

			replicas := int32(1)
			spec := v1beta1.PodGroupSpec{
				MinMember: replicas,
			}
			instanceSet := CreateTestInstanceSet("test-instance", "default", replicas)
			indexer := GetTestIndexer("test-service", "test-role", "0")
			ctx := context.Background()
			err := volcanoManager.createPodGroupForInstance(ctx, instanceSet, indexer, spec)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "failed to create podgroup")
		})
	})
}
