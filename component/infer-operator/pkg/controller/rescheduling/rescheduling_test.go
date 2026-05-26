package rescheduling

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func createMockWorkLoadHandlerWithError() workload.WorkLoadHandler {
	return &mockWorkLoadHandler{
		returnError:  true,
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

		convey.Convey("Should return false when pod has no instanceIndex label", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
			})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeFalse)
		})

		convey.Convey("Should return true for valid infer pod", func() {
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			result := rescheduler.isValidInferPod(pod)
			convey.So(result, convey.ShouldBeTrue)
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

		convey.Convey("Should return false for failed pod with negative retry times", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.OperatorNameKey:          common.TrueBool,
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "-1",
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

		convey.Convey("Should not overwrite existing retry times", func() {
			pod := createTestPod("test-pod", "default", map[string]string{
				common.PodStatusAnnotationKey: common.CommonUnhealthyStatus + "-" + common.PodFailed,
			}, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
				common.FaultRetryTimesLabelKey:  "3",
			})
			workLoadName := "test-service-test-role-0"
			instanceSetName := "test-service-test-role"
			expectedFaultWorkLoad := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: workLoadName},
				instanceSetName: instanceSetName,
			}
			rescheduler.faultRetryTimesMap[expectedFaultWorkLoad] = 10

			done := rescheduler.recordWorkLoadFault(pod, workLoadName, instanceSetName)

			convey.So(done, convey.ShouldBeFalse)

			rescheduler.Lock()
			retryTimes, exists := rescheduler.faultRetryTimesMap[expectedFaultWorkLoad]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(retryTimes, convey.ShouldEqual, 10)
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

func TestRescheduler_getFaultWorkLoad(t *testing.T) {
	convey.Convey("Test getFaultWorkLoad", t, func() {
		convey.Convey("Should return empty map when no fault workloads", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)

			faultMap, err := rescheduler.getFaultWorkLoad(context.Background(), instanceSet, createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(faultMap), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return fault workloads", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)

			mockHandler := &mockWorkLoadHandler{
				workLoadList: []workload.WorkLoadInterface{
					&mockWorkLoadInterface{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-workload",
							Namespace: "default",
						},
						replicas: 1,
						ready:    true,
					},
				},
				returnError: false,
			}

			faultWorkLoadKey := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-workload"},
				instanceSetName: instanceSet.Name,
			}
			rescheduler.faultWorkLoadMap[faultWorkLoadKey] = common.CommonUnhealthyStatus

			faultMap, err := rescheduler.getFaultWorkLoad(context.Background(), instanceSet, mockHandler)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(faultMap), convey.ShouldEqual, 1)
			convey.So(faultMap[faultWorkLoadKey], convey.ShouldEqual, common.CommonUnhealthyStatus)
		})

		convey.Convey("Should return error when ListWorkLoad fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)

			mockHandler := &mockWorkLoadHandler{
				listWorkLoadError: fmt.Errorf("failed to list workloads"),
				returnError:       true,
			}

			_, err := rescheduler.getFaultWorkLoad(context.Background(), instanceSet, mockHandler)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to list all workload")
		})
	})
}

func TestRescheduler_processPrioritySetting(t *testing.T) {
	convey.Convey("Test processPrioritySetting", t, func() {
		convey.Convey("Should return empty when no priority", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)

			result, err := rescheduler.processPrioritySetting(context.Background(), instanceSet, createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return empty when priority strategy not set", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			priority := int32(100)
			instanceSet := createTestInstanceSet("test-is", "default", &priority)

			result, err := rescheduler.processPrioritySetting(context.Background(), instanceSet, createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return error when no infer service label", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			priority := int32(100)
			instanceSet := createTestInstanceSet("test-is", "default", &priority)
			instanceSet.Labels = map[string]string{
				common.PrioritySchedulingStrategyLabelKey: common.SchedulingStrategyPriority,
			}

			_, err := rescheduler.processPrioritySetting(context.Background(), instanceSet, createMockWorkLoadHandler())

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "instance set does not have infer service label")
		})

		convey.Convey("Should process priority successfully", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			priority := int32(100)
			instanceSet := createTestInstanceSet("test-is", "default", &priority)
			instanceSet.Labels[common.PrioritySchedulingStrategyLabelKey] = common.SchedulingStrategyPriority

			result, err := rescheduler.processPrioritySetting(context.Background(), instanceSet, createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestRescheduler_deleteOtherWorkLoad(t *testing.T) {
	convey.Convey("Test deleteOtherWorkLoad", t, func() {
		convey.Convey("Should return error when listing InstanceSet fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			patches := gomonkey.ApplyMethodReturn(fakeClient, "List", fmt.Errorf("failed to list InstanceSet"))
			defer patches.Reset()
			_, err := rescheduler.deleteOtherWorkLoad(context.Background(), 100, "default", "test-service", createMockWorkLoadHandler())

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should skip when no unready low priority instance sets", func() {
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			result, err := rescheduler.deleteOtherWorkLoad(context.Background(), 100, "default", "test-service", createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should delete unready low priority workloads", func() {
			lowPriority := int32(200)
			instanceSet := createTestInstanceSet("test-is", "default", &lowPriority)
			instanceSet.Status.ReadyReplicas = 0
			instanceSet.Status.Replicas = 1
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			result, err := rescheduler.deleteOtherWorkLoad(context.Background(), 100, "default", "test-service", createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)
		})

		convey.Convey("Should skip when other priority is nil", func() {
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			instanceSet.Status.ReadyReplicas = 0
			instanceSet.Status.Replicas = 1
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			result, err := rescheduler.deleteOtherWorkLoad(context.Background(), 100, "default", "test-service", createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should skip when other priority is lower but instance is ready", func() {
			lowPriority := int32(200)
			instanceSet := createTestInstanceSet("test-is", "default", &lowPriority)
			instanceSet.Status.ReadyReplicas = 1
			instanceSet.Status.Replicas = 1
			fakeClient := newFakeClient(instanceSet)
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

			result, err := rescheduler.deleteOtherWorkLoad(context.Background(), 100, "default", "test-service", createMockWorkLoadHandler())

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestRescheduler_deleteFaultWorkLoad(t *testing.T) {
	convey.Convey("Test deleteFaultWorkLoad", t, func() {
		convey.Convey("Should return error when deleting workload fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			faultMap := make(map[faultWorkLoad]string)
			err := rescheduler.deleteFaultWorkLoad(context.Background(), instanceSet, createMockWorkLoadHandlerWithError(), faultMap)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should delete fault workloads successfully", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			faultMap := make(map[faultWorkLoad]string)
			faultWorkLoadKey := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-workload"},
				instanceSetName: instanceSet.Name,
			}
			faultMap[faultWorkLoadKey] = common.CommonUnhealthyStatus

			err := rescheduler.deleteFaultWorkLoad(context.Background(), instanceSet, createMockWorkLoadHandler(), faultMap)

			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Should decrement retry times for failed pod", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			faultMap := make(map[faultWorkLoad]string)
			faultWorkLoadKey := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-workload"},
				instanceSetName: instanceSet.Name,
			}
			faultMap[faultWorkLoadKey] = common.CommonUnhealthyStatus + "-" + common.PodFailed
			rescheduler.faultRetryTimesMap[faultWorkLoadKey] = 3

			err := rescheduler.deleteFaultWorkLoad(context.Background(), instanceSet, createMockWorkLoadHandler(), faultMap)

			convey.So(err, convey.ShouldBeNil)
			rescheduler.Lock()
			retryTimes, exists := rescheduler.faultRetryTimesMap[faultWorkLoadKey]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(retryTimes, convey.ShouldEqual, 2)
		})

		convey.Convey("Should not decrement retry times when retry times is zero", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			faultMap := make(map[faultWorkLoad]string)
			faultWorkLoadKey := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-workload"},
				instanceSetName: instanceSet.Name,
			}
			faultMap[faultWorkLoadKey] = common.CommonUnhealthyStatus + "-" + common.PodFailed
			rescheduler.faultRetryTimesMap[faultWorkLoadKey] = 0

			err := rescheduler.deleteFaultWorkLoad(context.Background(), instanceSet, createMockWorkLoadHandler(), faultMap)

			convey.So(err, convey.ShouldBeNil)
			rescheduler.Lock()
			retryTimes, exists := rescheduler.faultRetryTimesMap[faultWorkLoadKey]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(retryTimes, convey.ShouldEqual, 0)
		})
	})
}

func TestRescheduler_DoRescheduling(t *testing.T) {
	convey.Convey("Test DoRescheduling", t, func() {
		convey.Convey("Should return error when WorkLoadTypeToGVK fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			instanceSet.Spec.WorkloadTypeMeta = v1.WorkloadType{Kind: "Invalid"}
			patch := gomonkey.ApplyFuncReturn(schema.ParseGroupVersion, nil, fmt.Errorf("failed to parse group version"))
			defer patch.Reset()
			_, err := rescheduler.DoRescheduling(context.Background(), instanceSet)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when GetWorkLoadHandler fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			factory := &workload.WorkLoadHandlerFactory{}
			rescheduler.SetWorkLoadHandlerFactory(factory)
			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", nil, fmt.Errorf("failed to get WorkLoadHandler"))
			defer patches.Reset()

			_, err := rescheduler.DoRescheduling(context.Background(), instanceSet)

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return empty when no fault workloads", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			factory := &workload.WorkLoadHandlerFactory{}

			mockHandler := createMockWorkLoadHandler()
			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", mockHandler, nil)
			defer patches.Reset()

			rescheduler.SetWorkLoadHandlerFactory(factory)

			result, err := rescheduler.DoRescheduling(context.Background(), instanceSet)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should process rescheduling successfully", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			factory := &workload.WorkLoadHandlerFactory{}

			mockHandler := &mockWorkLoadHandler{
				workLoadList: []workload.WorkLoadInterface{
					&mockWorkLoadInterface{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-workload",
							Namespace: "default",
						},
						replicas: 1,
						ready:    true,
					},
				},
				returnError: false,
			}

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", mockHandler, nil)
			defer patches.Reset()

			rescheduler.SetWorkLoadHandlerFactory(factory)

			faultWorkLoadKey := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "test-workload"},
				instanceSetName: instanceSet.Name,
			}
			rescheduler.faultWorkLoadMap[faultWorkLoadKey] = common.CommonUnhealthyStatus

			result, err := rescheduler.DoRescheduling(context.Background(), instanceSet)

			convey.So(err, convey.ShouldBeNil)
			convey.So(len(result), convey.ShouldEqual, 1)

			// Verify fault was removed from map
			rescheduler.Lock()
			_, exists := rescheduler.faultWorkLoadMap[faultWorkLoadKey]
			rescheduler.Unlock()
			convey.So(exists, convey.ShouldBeFalse)
		})

	})
}

func TestGetNamespacedNameList(t *testing.T) {
	convey.Convey("Test getNamespacedNameList", t, func() {
		convey.Convey("Should return empty map for empty workload list", func() {
			result := getNamespacedNameList([]workload.WorkLoadInterface{})
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("Should return namespaced name map", func() {
			mockWorkLoad := &mockWorkLoadInterface{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload",
					Namespace: "default",
				},
			}
			result := getNamespacedNameList([]workload.WorkLoadInterface{mockWorkLoad})
			convey.So(len(result), convey.ShouldEqual, 1)
			_, exists := result[types.NamespacedName{Namespace: "default", Name: "test-workload"}]
			convey.So(exists, convey.ShouldBeTrue)
		})
	})
}

func TestRescheduler_triggerInstanceSetReconcile(t *testing.T) {
	convey.Convey("Test triggerInstanceSetReconcile", t, func() {
		convey.Convey("Should return error when WorkLoadTypeToGVK fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			instanceSet.Spec.WorkloadTypeMeta = v1.WorkloadType{Kind: "Invalid"}
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			patch := gomonkey.ApplyFuncReturn(schema.ParseGroupVersion, nil, fmt.Errorf("failed to parse group version"))
			defer patch.Reset()
			err := rescheduler.triggerInstanceSetReconcile(context.Background(), instanceSet, pod, "test-workload")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when GetWorkLoadHandler fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			factory := &workload.WorkLoadHandlerFactory{}
			rescheduler.SetWorkLoadHandlerFactory(factory)

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", nil, fmt.Errorf("failed to get workload handler"))
			defer patches.Reset()

			err := rescheduler.triggerInstanceSetReconcile(context.Background(), instanceSet, pod, "test-workload")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should return error when UpdateWorkLoad fails", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			factory := &workload.WorkLoadHandlerFactory{}

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", createMockWorkLoadHandlerWithError(), nil)
			defer patches.Reset()

			rescheduler.SetWorkLoadHandlerFactory(factory)

			err := rescheduler.triggerInstanceSetReconcile(context.Background(), instanceSet, pod, "test-workload")

			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("Should trigger reconcile successfully", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)
			instanceSet := createTestInstanceSet("test-is", "default", nil)
			pod := createTestPod("test-pod", "default", nil, map[string]string{
				common.InferServiceNameLabelKey: "test-service",
				common.InstanceSetNameLabelKey:  "test-role",
				common.InstanceIndexLabelKey:    "0",
			})
			factory := &workload.WorkLoadHandlerFactory{}

			patches := gomonkey.ApplyMethodReturn(factory, "GetWorkLoadHandler", createMockWorkLoadHandler(), nil)
			defer patches.Reset()

			rescheduler.SetWorkLoadHandlerFactory(factory)

			err := rescheduler.triggerInstanceSetReconcile(context.Background(), instanceSet, pod, "test-workload")

			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRescheduler_retryTimesMapPeriodicCleanup(t *testing.T) {
	convey.Convey("Test retryTimesMapPeriodicCleanup", t, func() {
		convey.Convey("Should start periodic cleanup and clean up when ticker fires", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, 100*time.Millisecond)

			// Setup test data
			testFaultWorkLoad1 := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload1"},
				instanceSetName: "set1",
			}
			testFaultWorkLoad2 := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload2"},
				instanceSetName: "set2",
			}
			testFaultWorkLoad3 := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload3"},
				instanceSetName: "set3",
			}

			// Add entries to both maps
			rescheduler.faultWorkLoadMap[testFaultWorkLoad1] = common.CommonUnhealthyStatus
			rescheduler.faultWorkLoadMap[testFaultWorkLoad2] = common.CommonUnhealthyStatus
			rescheduler.faultRetryTimesMap[testFaultWorkLoad1] = 3
			rescheduler.faultRetryTimesMap[testFaultWorkLoad2] = 2
			rescheduler.faultRetryTimesMap[testFaultWorkLoad3] = 1

			ctx, cancel := context.WithCancel(context.Background())

			// Start the periodic cleanup in a goroutine
			go rescheduler.retryTimesMapPeriodicCleanup(ctx)

			// Wait for at least one cleanup cycle
			time.Sleep(250 * time.Millisecond)

			// Cancel context to stop the goroutine
			cancel()

			// Give some time for the goroutine to exit
			time.Sleep(50 * time.Millisecond)

			// Verify cleanup results
			rescheduler.Lock()
			defer rescheduler.Unlock()

			// workload3 should be removed from faultRetryTimesMap
			_, exists := rescheduler.faultRetryTimesMap[testFaultWorkLoad3]
			convey.So(exists, convey.ShouldBeFalse)

			// workload1 and workload2 should still exist
			_, exists = rescheduler.faultRetryTimesMap[testFaultWorkLoad1]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(rescheduler.faultRetryTimesMap[testFaultWorkLoad1], convey.ShouldEqual, 3)

			_, exists = rescheduler.faultRetryTimesMap[testFaultWorkLoad2]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(rescheduler.faultRetryTimesMap[testFaultWorkLoad2], convey.ShouldEqual, 2)

			// faultWorkLoadMap should remain unchanged
			_, exists = rescheduler.faultWorkLoadMap[testFaultWorkLoad1]
			convey.So(exists, convey.ShouldBeTrue)
			_, exists = rescheduler.faultWorkLoadMap[testFaultWorkLoad2]
			convey.So(exists, convey.ShouldBeTrue)
		})

		convey.Convey("Should stop when context is cancelled", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, 10*time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			// This should return quickly without panic
			rescheduler.retryTimesMapPeriodicCleanup(ctx)
		})
	})
}

func TestRescheduler_performCleanup(t *testing.T) {
	convey.Convey("Test cleanupWithTimeout", t, func() {
		convey.Convey("Should clean up entries not present in faultWorkLoadMap", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, 100*time.Millisecond)

			// Setup test data
			workloadInBothMaps := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload-in-both"},
				instanceSetName: "set1",
			}
			workloadOnlyInRetryMap := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload-only-in-retry"},
				instanceSetName: "set2",
			}
			workloadOnlyInFaultMap := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload-only-in-fault"},
				instanceSetName: "set3",
			}

			// Setup faultWorkLoadMap
			rescheduler.faultWorkLoadMap[workloadInBothMaps] = common.CommonUnhealthyStatus
			rescheduler.faultWorkLoadMap[workloadOnlyInFaultMap] = common.CommonUnhealthyStatus

			// Setup faultRetryTimesMap
			rescheduler.faultRetryTimesMap[workloadInBothMaps] = 3
			rescheduler.faultRetryTimesMap[workloadOnlyInRetryMap] = 5

			// Execute cleanup
			rescheduler.cleanupWithTimeout()

			// Verify results
			_, existsInRetry := rescheduler.faultRetryTimesMap[workloadInBothMaps]
			convey.So(existsInRetry, convey.ShouldBeTrue)
			convey.So(rescheduler.faultRetryTimesMap[workloadInBothMaps], convey.ShouldEqual, 3)

			_, existsInRetry = rescheduler.faultRetryTimesMap[workloadOnlyInRetryMap]
			convey.So(existsInRetry, convey.ShouldBeFalse)

			// faultWorkLoadMap should remain unchanged
			_, existsInFault := rescheduler.faultWorkLoadMap[workloadInBothMaps]
			convey.So(existsInFault, convey.ShouldBeTrue)
			_, existsInFault = rescheduler.faultWorkLoadMap[workloadOnlyInFaultMap]
			convey.So(existsInFault, convey.ShouldBeTrue)
		})

		convey.Convey("Should handle empty maps gracefully", func() {
			fakeClient := newFakeClient()
			rescheduler := NewRescheduler(fakeClient, 100*time.Millisecond)

			// Both maps are empty
			convey.So(len(rescheduler.faultWorkLoadMap), convey.ShouldEqual, 0)
			convey.So(len(rescheduler.faultRetryTimesMap), convey.ShouldEqual, 0)

			// Execute cleanup - should not panic
			rescheduler.cleanupWithTimeout()

			// Maps should still be empty
			convey.So(len(rescheduler.faultWorkLoadMap), convey.ShouldEqual, 0)
			convey.So(len(rescheduler.faultRetryTimesMap), convey.ShouldEqual, 0)
		})
	})
}

func TestRescheduler_CleanupWithInstanceSetDeletion(t *testing.T) {
	convey.Convey("Test CleanupWithInstanceSetDeletion", t, func() {
		fakeClient := newFakeClient()
		rescheduler := NewRescheduler(fakeClient, common.FaultRetryTimesCleanupInterval)

		convey.Convey("Should cleanup faultWorkLoadMap and faultRetryTimesMap for the given instanceSet", func() {
			// Setup test data for instanceSet "test-set-1"
			testWorkLoad1 := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload1"},
				instanceSetName: "test-set-1",
			}

			// Setup test data for instanceSet "test-set-2" (should remain)
			testWorkLoad3 := faultWorkLoad{
				NamespacedName:  types.NamespacedName{Namespace: "default", Name: "workload3"},
				instanceSetName: "test-set-2",
			}

			// Add entries to both maps
			rescheduler.faultWorkLoadMap[testWorkLoad1] = common.CommonUnhealthyStatus
			rescheduler.faultWorkLoadMap[testWorkLoad3] = common.CommonUnhealthyStatus

			rescheduler.faultRetryTimesMap[testWorkLoad1] = 3
			rescheduler.faultRetryTimesMap[testWorkLoad3] = 5

			// Execute cleanup for instanceSet "test-set-1"
			rescheduler.CleanupWithInstanceSetDeletion("test-set-1")

			// Verify entries for "test-set-1" are removed from faultWorkLoadMap
			_, exists := rescheduler.faultWorkLoadMap[testWorkLoad1]
			convey.So(exists, convey.ShouldBeFalse)

			// Verify entries for "test-set-1" are removed from faultRetryTimesMap
			_, exists = rescheduler.faultRetryTimesMap[testWorkLoad1]
			convey.So(exists, convey.ShouldBeFalse)

			// Verify entries for "test-set-2" remain
			_, exists = rescheduler.faultWorkLoadMap[testWorkLoad3]
			convey.So(exists, convey.ShouldBeTrue)

			_, exists = rescheduler.faultRetryTimesMap[testWorkLoad3]
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(rescheduler.faultRetryTimesMap[testWorkLoad3], convey.ShouldEqual, 5)
		})

	})
}
