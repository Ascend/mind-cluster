package rescheduling

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
	"infer-operator/pkg/controller/workload"
)

var (
	testScheme = runtime.NewScheme()
)

func init() {
	_ = scheme.AddToScheme(testScheme)
	_ = v1.AddToScheme(testScheme)
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func newFakeClient(objects ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(testScheme).WithRuntimeObjects(objects...).Build()
}

func createTestPod(name, namespace string, annotations, labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      labels,
		},
	}
}

func createTestInstanceSet(name, namespace string, priority *int32) *v1.InstanceSet {
	return &v1.InstanceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			},
			UID: types.UID("test-uid-123"),
		},
		Spec: v1.InstanceSetSpec{
			Name:     "test-role",
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Priority: priority,
			WorkloadTypeMeta: v1.WorkloadType{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
		},
	}
}

func createTestDeployment(name, namespace string, annotations map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
			Labels: map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			},
		},
	}
}

type mockWorkLoadInterface struct {
	ObjectMeta metav1.ObjectMeta
	replicas   int32
	ready      bool
}

func (m *mockWorkLoadInterface) GetWorkLoadObjMeta() metav1.ObjectMeta {
	return m.ObjectMeta
}

func (m *mockWorkLoadInterface) SetWorkLoadObjMeta(objectMeta metav1.ObjectMeta) {
	m.ObjectMeta = objectMeta
}

func (m *mockWorkLoadInterface) GetWorkLoadReplicas() int32 {
	return m.replicas
}

func (m *mockWorkLoadInterface) IsWorkLoadReady() bool {
	return m.ready
}

type mockWorkLoadHandler struct {
	listWorkLoadError   error
	deleteWorkLoadError error
	updateWorkLoadError error
	workLoadList        []workload.WorkLoadInterface
	returnError         bool
}

func (m *mockWorkLoadHandler) CheckOrCreateWorkLoad(context.Context, *v1.InstanceSet,
	common.InstanceIndexer) error {
	return nil
}

func (m *mockWorkLoadHandler) DeleteExtraWorkLoad(context.Context, common.InstanceIndexer, int) error {
	return nil
}

func (m *mockWorkLoadHandler) GetWorkLoadReadyReplicas(context.Context, common.InstanceIndexer) (int, error) {
	return 1, nil
}

func (m *mockWorkLoadHandler) Validate(runtime.RawExtension) error {
	return nil
}

func (m *mockWorkLoadHandler) GetReplicas(runtime.RawExtension) (int32, error) {
	return 1, nil
}

func (m *mockWorkLoadHandler) ListWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...workload.WorkLoadFilter) ([]workload.WorkLoadInterface, error) {
	if m.listWorkLoadError != nil {
		return nil, m.listWorkLoadError
	}
	return m.workLoadList, nil
}

func (m *mockWorkLoadHandler) DeleteWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	filters ...workload.WorkLoadFilter) error {
	if m.deleteWorkLoadError != nil {
		return m.deleteWorkLoadError
	}
	if m.returnError {
		return fmt.Errorf("delete workload failed")
	}
	return nil
}

func (m *mockWorkLoadHandler) UpdateWorkLoad(
	ctx context.Context,
	selectLabels map[string]string,
	namespace string,
	updater workload.WorkloadUpdater,
	filters ...workload.WorkLoadFilter) error {
	if m.updateWorkLoadError != nil {
		return m.updateWorkLoadError
	}
	if m.returnError {
		return fmt.Errorf("update workload failed")
	}
	return nil
}

func createMockWorkLoadHandler() workload.WorkLoadHandler {
	return &mockWorkLoadHandler{
		returnError:  false,
		workLoadList: []workload.WorkLoadInterface{},
	}
}

func TestNewRescheduler(t *testing.T) {
	convey.Convey("Test NewRescheduler", t, func() {
		convey.Convey("Should create a new Rescheduler", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			convey.So(rescheduler, convey.ShouldNotBeNil)
			convey.So(rescheduler.client, convey.ShouldEqual, fakeClient)
			convey.So(rescheduler.faultRetryTimesMap, convey.ShouldNotBeNil)
			convey.So(rescheduler.faultWorkLoadMap, convey.ShouldNotBeNil)
		})
	})
}

func TestRescheduler_SetWorkLoadHandlerFactory(t *testing.T) {
	convey.Convey("Test SetWorkLoadHandlerFactory", t, func() {
		convey.Convey("Should set WorkLoadHandlerFactory", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			factory := &workload.WorkLoadHandlerFactory{}

			rescheduler.SetWorkLoadHandlerFactory(factory)

			convey.So(rescheduler.workLoadHandlerFactory, convey.ShouldEqual, factory)
		})
	})
}

func TestRescheduler_isValidInferPod(t *testing.T) {
	convey.Convey("Test isValidInferPod", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should return false when pod has no operator label", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when operator label is not true", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey: "false",
			})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when pod has no infer service name", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey:         common.TrueBool,
				common.InstanceSetNameLabelKey: "test-role",
				common.InstanceIndexLabelKey:   "0",
			})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when pod has no instanceSetName label", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceIndexLabelKey:    "0",
			})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

	})
}

func TestRescheduler_isValidFaultPod(t *testing.T) {
	convey.Convey("Test isValidFaultPod", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should return false when pod is not valid infer pod", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when pod has no unhealthy status", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when pod is being deleted", func() {
			now := metav1.Now()
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			pod.DeletionTimestamp = &now
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true for valid fault pod", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return true for failed pod with valid retry times", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "1",
			})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false for failed pod with no retry times label", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return false for failed pod with invalid retry times", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "invalid",
			})
			result := rescheduler.isValidFaultPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestRescheduler_handlePodUpdate(t *testing.T) {
	convey.Convey("Test handlePodUpdate", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should skip when new object is not a pod", func() {
			rescheduler.handlePodUpdate("not-a-pod", "not-a-pod")
		})

		convey.Convey("Should skip when pod is not valid fault pod", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{})
			rescheduler.handlePodUpdate(nil, pod)
		})
	})
}

func TestRescheduler_processFaultEvent(t *testing.T) {
	convey.Convey("Test processFaultEvent", t, func() {
		convey.Convey("Should return error when instanceSet not found", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			err := rescheduler.processFaultEvent(pod)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to get instance set")
		})

		convey.Convey("Should record fault successfully and trigger reconcile", func() {
			instanceSet := createTestInstanceSet("test-service-test-role", "default", nil)
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			factory := &workload.WorkLoadHandlerFactory{}
			rescheduler.SetWorkLoadHandlerFactory(factory)

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", createMockWorkLoadHandler(), nil)
			defer patches.Reset()

			err := rescheduler.processFaultEvent(pod)
			convey.So(err, convey.ShouldBeNil)

			// Verify fault was recorded
			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-service-test-role-0"},
				instanceSetName: "test-service-test-role",
			}
			rescheduler.Lock()
			_, exists := rescheduler.faultWorkLoadMap[expectedFaultWorkLoad]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
		})

		convey.Convey("Should skip when workload already recorded", func() {
			instanceSet := createTestInstanceSet("test-service-test-role", "default", nil)
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			// Pre-record the fault
			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-service-test-role-0"},
				instanceSetName: "test-service-test-role",
			}
			rescheduler.faultWorkLoadMap[expectedFaultWorkLoad] = common.CommonUnhealthyStatus

			err := rescheduler.processFaultEvent(pod)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should record retry times for failed pod", func() {
			instanceSet := createTestInstanceSet("test-service-test-role", "default", nil)
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "3",
			})
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			factory := &workload.WorkLoadHandlerFactory{}
			rescheduler.SetWorkLoadHandlerFactory(factory)

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", createMockWorkLoadHandler(), nil)
			defer patches.Reset()

			err := rescheduler.processFaultEvent(pod)

			convey.So(err, convey.ShouldBeNil)

			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-service-test-role-0"},
				instanceSetName: "test-service-test-role",
			}
			rescheduler.Lock()
			retryTimes, exists := rescheduler.faultRetryTimesMap[expectedFaultWorkLoad]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(retryTimes, convey.ShouldEqual, 3)
		})
	})
}

func TestRescheduler_recordWorkLoadFault(t *testing.T) {
	convey.Convey("Test recordWorkLoadFault", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should return true when workload already recorded", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			workLoadName := "test-service-test-role-0"
			instanceSetName := "test-service-test-role"
			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: workLoadName},
				instanceSetName: instanceSetName,
			}
			rescheduler.faultWorkLoadMap[expectedFaultWorkLoad] = common.CommonUnhealthyStatus

			done := rescheduler.recordWorkLoadFault(pod, workLoadName, instanceSetName)

			convey.So(done, convey.ShouldBeTrue)
		})

		convey.Convey("Should return false and record fault when not already recorded", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus,
			}, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			workLoadName := "test-service-test-role-0"
			instanceSetName := "test-service-test-role"

			done := rescheduler.recordWorkLoadFault(pod, workLoadName, instanceSetName)

			convey.So(done, convey.ShouldBeFalse)

			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: workLoadName},
				instanceSetName: instanceSetName,
			}
			rescheduler.Lock()
			_, exists := rescheduler.faultWorkLoadMap[expectedFaultWorkLoad]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
		})

		convey.Convey("Should record retry times for failed pod", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "5",
			})
			workLoadName := "test-service-test-role-0"
			instanceSetName := "test-service-test-role"

			done := rescheduler.recordWorkLoadFault(pod, workLoadName, instanceSetName)

			convey.So(done, convey.ShouldBeFalse)

			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: workLoadName},
				instanceSetName: instanceSetName,
			}
			rescheduler.Lock()
			retryTimes, exists := rescheduler.faultRetryTimesMap[expectedFaultWorkLoad]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(retryTimes, convey.ShouldEqual, 5)
		})

	})
}

func TestRescheduler_getWorkLoadNameAndInstanceSetName(t *testing.T) {
	convey.Convey("Test getWorkLoadNameAndInstanceSetName", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should return correct workload name and instanceSet name", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "42",
			})

			workLoadName, instanceSetName := rescheduler.getWorkLoadNameAndInstanceSetName(pod)

			convey.So(workLoadName, convey.ShouldEqual, "test-service-test-role-42")
			convey.So(instanceSetName, convey.ShouldEqual, "test-service-test-role")
		})
	})
}
