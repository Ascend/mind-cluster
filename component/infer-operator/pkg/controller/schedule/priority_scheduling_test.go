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

// Package schedule contains the scheduling logic
package schedule

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestShouldScheduleErrors tests the ShouldSchedule function with error scenarios.
func TestShouldScheduleErrors(t *testing.T) {
	convey.Convey("Test ShouldSchedule error scenarios", t, func() {
		convey.Convey("Should return error when scheduling strategy is unknown", func() {
			instanceSet := createPriorityInstanceSet("test-role", "unknown-strategy", false)
			fakeClient := createFakeClient(instanceSet)

			ctx := context.Background()
			shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(shouldSchedule, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when InferService not found", func() {
			instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
			fakeClient := createFakeClient(instanceSet)

			ctx := context.Background()
			shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(shouldSchedule, convey.ShouldBeFalse)
		})

		convey.Convey("Should return error when InferServiceNameLabelKey is missing", func() {
			instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
			delete(instanceSet.Labels, common.InferServiceNameLabelKey)
			fakeClient := createFakeClient(instanceSet)

			ctx := context.Background()
			shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(shouldSchedule, convey.ShouldBeFalse)
		})
	})
}

// TestShouldScheduleWhenCurrentRoleHasHighestPriority tests the ShouldSchedule function.
func TestShouldScheduleWhenCurrentRoleHasHighestPriority(t *testing.T) {
	convey.Convey("Should return true when scheduling strategy is priority and current role has highest priority", t, func() {
		roles := []v1.InstanceSetSpec{
			{Name: "test-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
			{Name: "sibling-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
		}
		inferService := createPriorityInferServiceWithRoles(roles)
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
		fakeClient := createFakeClient(instanceSet, inferService)

		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(shouldSchedule, convey.ShouldBeTrue)
	})
}

// TestShouldScheduleWhenCurrentRoleHasLowerPriority tests the ShouldSchedule function.
func TestShouldScheduleWhenCurrentRoleHasLowerPriority(t *testing.T) {
	convey.Convey("Should return false with RequeueError when current role has lower priority", t, func() {
		roles := []v1.InstanceSetSpec{
			{Name: "test-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
			{Name: "sibling-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
		}
		inferService := createPriorityInferServiceWithRoles(roles)
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
		siblingInstanceSet := createPriorityInstanceSet("sibling-role", common.SchedulingStrategyPriority, false)
		siblingInstanceSet.UID = "sibling-uid"

		fakeClient := createFakeClient(instanceSet, siblingInstanceSet, inferService)
		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(common.IsRequeueError(err), convey.ShouldBeTrue)
		convey.So(shouldSchedule, convey.ShouldBeFalse)
	})
}

// TestShouldScheduleWhenCurrentRoleIsReady tests the ShouldSchedule function.
func TestShouldScheduleWhenCurrentRoleIsReady(t *testing.T) {
	convey.Convey("Should return false when current role is ready", t, func() {
		roles := []v1.InstanceSetSpec{
			{Name: "test-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
			{Name: "sibling-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
		}
		inferService := createPriorityInferServiceWithRoles(roles)
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, true)
		fakeClient := createFakeClient(instanceSet, inferService)

		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(shouldSchedule, convey.ShouldBeFalse)
	})
}

// TestShouldScheduleWhenAllSiblingInstanceSetsAreReady tests the ShouldSchedule function.
func TestShouldScheduleWhenAllSiblingInstanceSetsAreReady(t *testing.T) {
	convey.Convey("Should return true when all sibling InstanceSets are ready", t, func() {
		roles := []v1.InstanceSetSpec{
			{Name: "sibling-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
			{Name: "test-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
		}
		inferService := createPriorityInferServiceWithRoles(roles)
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
		siblingInstanceSet := createPriorityInstanceSet("sibling-role", common.SchedulingStrategyPriority, true)
		siblingInstanceSet.UID = "sibling-uid"
		fakeClient := createFakeClient(instanceSet, siblingInstanceSet, inferService)

		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(shouldSchedule, convey.ShouldBeTrue)
	})
}

// TestShouldScheduleWhenMultipleRolesHaveSameHighestPriority tests the ShouldSchedule function.
func TestShouldScheduleWhenMultipleRolesHaveSameHighestPriority(t *testing.T) {
	convey.Convey("Should return true when current role has same priority with highest not ready role", t, func() {
		roles := []v1.InstanceSetSpec{
			{Name: "test-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
			{Name: "sibling-role", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
		}
		inferService := createPriorityInferServiceWithRoles(roles)
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategyPriority, false)
		siblingInstanceSet := createPriorityInstanceSet("sibling-role", common.SchedulingStrategyPriority, false)
		siblingInstanceSet.UID = "sibling-uid"
		fakeClient := createFakeClient(instanceSet, siblingInstanceSet, inferService)

		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(shouldSchedule, convey.ShouldBeTrue)
	})
}

// TestShouldScheduleWhenSchedulingStrategyIsSequential tests the ShouldSchedule function.
func TestShouldScheduleWhenSchedulingStrategyIsSequential(t *testing.T) {
	convey.Convey("Should return true when scheduling strategy is sequential", t, func() {
		instanceSet := createPriorityInstanceSet("test-role", common.SchedulingStrategySequential, false)
		fakeClient := createFakeClient(instanceSet)

		ctx := context.Background()
		shouldSchedule, err := ShouldSchedule(ctx, fakeClient, instanceSet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(shouldSchedule, convey.ShouldBeTrue)
	})
}

// TestFindHighestPriorityOfNotReadyRole tests the findHighestPriorityOfNotReadyRole function.
func TestFindHighestPriorityOfNotReadyRole(t *testing.T) {
	convey.Convey("Test findHighestPriorityOfNotReadyRole function", t, func() {
		convey.Convey("Should return highest priority when some roles are not ready", func() {
			inferService := createTestInferService("test-service", "default")
			inferService.Spec.Roles = []v1.InstanceSetSpec{
				{Name: "test-role1", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
				{Name: "test-role2", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
				{Name: "test-role3", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(3); return &p }()},
			}
			instanceSetMap := map[string]*v1.InstanceSet{
				"test-role1": createPriorityInstanceSet("test-role1", common.PrioritySchedulingStrategyLabelKey, true),
				"test-role2": createPriorityInstanceSet("test-role2", common.PrioritySchedulingStrategyLabelKey, false),
			}

			priority := findHighestPriorityOfNotReadyRole(inferService, instanceSetMap)
			convey.So(priority, convey.ShouldEqual, int32(2))
		})

		convey.Convey("Should return AllReadyPriority when all roles are ready", func() {
			inferService := createTestInferService("test-service", "default")
			inferService.Spec.Roles = []v1.InstanceSetSpec{
				{Name: "test-role1", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
				{Name: "test-role2", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
			}
			instanceSetMap := map[string]*v1.InstanceSet{
				"test-role1": createPriorityInstanceSet("test-role1", common.PrioritySchedulingStrategyLabelKey, true),
				"test-role2": createPriorityInstanceSet("test-role2", common.PrioritySchedulingStrategyLabelKey, true),
			}

			priority := findHighestPriorityOfNotReadyRole(inferService, instanceSetMap)
			convey.So(priority, convey.ShouldEqual, AllReadyPriority)
			convey.So(AllReadyPriority > common.MaxRoleTypeCount, convey.ShouldBeTrue)
		})

		convey.Convey("Should return highest priority when role does not exist", func() {
			inferService := createTestInferService("test-service", "default")
			inferService.Spec.Roles = []v1.InstanceSetSpec{
				{Name: "test-role1", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(1); return &p }()},
				{Name: "test-role2", Replicas: func() *int32 { r := int32(1); return &r }(), Priority: func() *int32 { p := int32(2); return &p }()},
			}
			instanceSetMap := map[string]*v1.InstanceSet{
				"test-role1": createPriorityInstanceSet("test-role1", common.PrioritySchedulingStrategyLabelKey, true),
			}

			priority := findHighestPriorityOfNotReadyRole(inferService, instanceSetMap)
			convey.So(priority, convey.ShouldEqual, int32(2))
		})
	})
}

func TestGetPriority(t *testing.T) {
	convey.Convey("Test getPriority function", t, func() {
		convey.Convey("Should return priority value when priority is set", func() {
			role := &v1.InstanceSetSpec{
				Name:     "test-role",
				Priority: func() *int32 { p := int32(5); return &p }(),
			}
			priority := getPriority(role)
			convey.So(priority, convey.ShouldEqual, int32(5))
		})

		convey.Convey("Should return default priority when priority is nil", func() {
			role := &v1.InstanceSetSpec{
				Name: "test-role",
			}
			priority := getPriority(role)
			minPriority := int32(1)
			convey.So(priority, convey.ShouldEqual, common.DefaultPriority)
			convey.So(common.DefaultPriority <= common.MaxRoleTypeCount, convey.ShouldBeTrue)
			convey.So(common.DefaultPriority >= minPriority, convey.ShouldBeTrue)
		})
	})
}

func createTestInferService(name, namespace string) *v1.InferService {
	return &v1.InferService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func createFakeClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func createPriorityInferServiceWithRoles(roles []v1.InstanceSetSpec) *v1.InferService {
	inferService := createTestInferService("test-service", "default")
	inferService.Spec.SchedulingStrategy = &v1.SchedulingStrategy{
		Type: common.SchedulingStrategyPriority,
	}
	inferService.Spec.Roles = roles
	return inferService
}

func createPriorityInstanceSet(roleName, prioritySchedulingStrategy string, isReady bool) *v1.InstanceSet {
	replicas := int32(1)
	status := metav1.ConditionFalse
	if isReady {
		status = metav1.ConditionTrue
	}
	instanceSet := &v1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instanceSet-" + roleName,
			Namespace: "default",
		},
		Spec: v1.InstanceSetSpec{
			Replicas: &replicas,
		},
		Status: v1.InstanceSetStatus{
			Conditions: []metav1.Condition{
				{Type: string(common.InstanceSetReady), Status: status},
			},
		},
	}
	instanceSet.Labels = map[string]string{
		common.InferServiceNameLabelKey:           "test-service",
		common.InstanceSetNameLabelKey:            roleName,
		common.PrioritySchedulingStrategyLabelKey: prioritySchedulingStrategy,
	}
	return instanceSet
}
