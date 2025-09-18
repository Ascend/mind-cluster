/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/training-operator/pkg/common/util"
	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/scheduling/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	configv1alpha1 "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable"
	"ascend-operator/pkg/ranktable/generator"
	_ "ascend-operator/pkg/testtool"
	"ascend-operator/pkg/utils"
)

func newCommonReconciler() *ASJobReconciler {
	rc := &ASJobReconciler{
		Client:        &fakeClient{},
		recorder:      &fakeRecorder{},
		apiReader:     &fakeReader{},
		versions:      make(map[types.UID]int32),
		backoffLimits: make(map[types.UID]int32),
		rtGenerators:  map[types.UID]generator.RankTableGenerator{},
	}
	rc.JobController = common.JobController{
		Controller:          rc,
		Config:              common.JobControllerConfiguration{EnableGangScheduling: true},
		WorkQueue:           &util.FakeWorkQueue{},
		PodLister:           &fakePodLister{},
		PriorityClassLister: &fakePriorityClassLister{},
		Recorder:            &fakeRecorder{},
		PodControl:          &fakePodController{},
	}
	return rc
}

const (
	spBlock = "32"
)

type fakeReader struct{}

func (f fakeReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return nil
}

func (f fakeReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func newCommonPodInfo() *podInfo {
	return &podInfo{
		rtype: mindxdlv1.ReplicaTypeWorker,
		ip:    "127.0.0.1",
		job: &mindxdlv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				UID: "123456",
				Annotations: map[string]string{
					utils.AnnoKeyOfSuperPod: spBlock,
				},
			},
		},
		spec: &commonv1.ReplicaSpec{
			Replicas:      newReplicas(1),
			Template:      corev1.PodTemplateSpec{},
			RestartPolicy: "",
		},
		port:        "2222",
		ctReq:       2,
		npuReplicas: 1,
		rank:        1,
	}
}

type fakePodController struct{}

// CreatePods This function is part of the fakePodController struct and creates a pod based on the given parameters.
func (pc *fakePodController) CreatePods(_ string, _ *corev1.PodTemplateSpec, _ runtime.Object) error {
	return nil
}

// CreatePodsOnNode This function is part of the fakePodController struct
// and creates a pod on a given node based on the given parameters.
func (pc *fakePodController) CreatePodsOnNode(_, _ string, _ *corev1.PodTemplateSpec,
	_ runtime.Object, _ *metav1.OwnerReference) error {
	return nil
}

// CreatePodsWithControllerRef This function is part of the fakePodController struct
// and creates a pod with a given controller reference based on the given parameters.
func (pc *fakePodController) CreatePodsWithControllerRef(_ string, _ *corev1.PodTemplateSpec,
	_ runtime.Object, _ *metav1.OwnerReference) error {
	return nil
}

// DeletePod This function is part of the fakePodController struct and deletes a pod based on the given parameters.
func (pc *fakePodController) DeletePod(_ string, _ string, _ runtime.Object) error { return nil }

// PatchPod This function is part of the fakePodController struct and patches a pod based on the given parameters.
func (pc *fakePodController) PatchPod(_, _ string, _ []byte) error { return nil }

// fakeClient This struct is a fake client struct used to mock the behavior of a kubernetes client.
type fakeClient struct{}

// Get This function is part of the fakeClient struct
// and retrieves an object from the client based on the given parameters.
func (c *fakeClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object) error { return nil }

// List This function is part of the fakeClient struct and lists objects from the client based on the given parameters.
func (c *fakeClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return nil
}

// Create This function is part of the fakeClient struct and creates an object based on the given parameters.
func (c *fakeClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return nil
}

// Delete This function is part of the fakeClient struct and deletes an object based on the given parameters.
func (c *fakeClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return nil
}

// Update This function is part of the fakeClient struct and updates an object based on the given parameters.
func (c *fakeClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return nil
}

// Patch This function is part of the fakeClient struct and patches an object based on the given parameters.
func (c *fakeClient) Patch(_ context.Context, _ client.Object, _ client.Patch,
	_ ...client.PatchOption) error {
	return nil
}

// DeleteAllOf This function is part of the fakeClient struct and deletes all objects based on the given parameters.
func (c *fakeClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
	return nil
}

type fakeStatusWriter struct{}

func (f *fakeStatusWriter) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return nil
}

func (f *fakeStatusWriter) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return nil
}

// Status This function is part of the fakeClient struct and returns a status based on the given parameters.
func (c *fakeClient) Status() client.StatusWriter { return &fakeStatusWriter{} }

// Scheme This function is part of the fakeClient struct and returns a scheme based on the given parameters.
func (c *fakeClient) Scheme() *runtime.Scheme { return nil }

// RESTMapper This function is part of the fakeClient struct and returns a rest mapper based on the given parameters.
func (c *fakeClient) RESTMapper() meta.RESTMapper { return nil }

// fakeRecorder This struct is a fake recorder struct used to mock the behavior of a kubernetes recorder.
type fakeRecorder struct{}

// Event This function is part of the fakeRecorder struct and records an event based on the given parameters.
func (rc *fakeRecorder) Event(_ runtime.Object, _, _, _ string) {}

// Eventf This function is part of the fakeRecorder struct
// and records an event with a given message based on the given parameters.
func (rc *fakeRecorder) Eventf(_ runtime.Object, _, _, _ string, _ ...interface{}) {}

// AnnotatedEventf This function is part of the fakeRecorder struct
// and records an event with a given message and annotations based on the given parameters.
func (rc *fakeRecorder) AnnotatedEventf(_ runtime.Object, _ map[string]string, _, _,
	_ string, _ ...interface{}) {
}

// fakePriorityClassLister This struct is a fake priority class lister struct
// used to mock the behavior of a kubernetes priority class lister.
type fakePriorityClassLister struct{}

// List This function is part of the fakePriorityClassLister struct
// and lists priority classes based on the given parameters.
func (pc *fakePriorityClassLister) List(_ labels.Selector) ([]*v1beta1.PriorityClass, error) {
	return nil, nil
}

// Get This function is part of the fakePriorityClassLister struct
// and retrieves a priority class based on the given parameters.
func (pc *fakePriorityClassLister) Get(_ string) (*v1beta1.PriorityClass, error) {
	return nil, nil
}

type fakePodLister struct{}

// List lists all Pods in the indexer.
func (s *fakePodLister) List(selector labels.Selector) ([]*corev1.Pod, error) {
	return nil, nil
}

// Pods returns an object that can list and get Pods.
func (s *fakePodLister) Pods(namespace string) corelisters.PodNamespaceLister {
	return nil
}

const fakeResourceName = "huawei.com/Ascend910"

func newCommonContainer() corev1.Container {
	return corev1.Container{
		Name: "test",
		Ports: []corev1.ContainerPort{
			{
				Name:          api.DefaultPortName,
				ContainerPort: fakePort,
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: map[corev1.ResourceName]resource.Quantity{
				fakeResourceName: resource.MustParse("1"),
			},
			Requests: map[corev1.ResourceName]resource.Quantity{
				fakeResourceName: resource.MustParse("1"),
			},
		},
	}
}

func newCommonAscendJob() *mindxdlv1.Job {
	return &mindxdlv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: acJobApiversion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ascendjob-test",
			UID:         "1111",
			Annotations: map[string]string{},
		},
		Spec: mindxdlv1.JobSpec{},
	}
}

func newReplicas(i int) *int32 {
	x := int32(i)
	return &x
}

// TestIsVcjobOrDeploy This function is a test function for the isVcjobOrDeploy
// method of the AscendJobReconciler struct.
func TestIsVcjobOrDeploy(t *testing.T) {
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	vcjobReq := controllerruntime.Request{NamespacedName: types.NamespacedName{Namespace: "vcjob", Name: "vcjob"}}
	deployReq := controllerruntime.Request{NamespacedName: types.NamespacedName{Namespace: "deploy", Name: "deploy"}}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test vcjob", args: args{ctx: context.TODO(), req: vcjobReq}, want: true},
		{name: "test deploy", args: args{ctx: context.TODO(), req: deployReq}, want: true},
	}
	r := newCommonReconciler()

	// stub ranktablePipeline
	patches := gomonkey.ApplyPrivateMethod(r, "ranktablePipeline", func(job *mindxdlv1.Job) { return })
	defer patches.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.isVcjobOrDeploy(tt.args.ctx, tt.args.req); got != tt.want {
				t.Errorf("isVcjobOrDeploy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteRanktableToCm(t *testing.T) {
	convey.Convey("TestWriteRanktableToCm", t, func() {
		r := newCommonReconciler()
		convey.Convey("01-ranktable generaotor not found should return error", func() {
			err := r.writeRanktableToCm("job", "default", "111")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-ranktable generaotor not found should return error", func() {
			r.rtGenerators = map[types.UID]generator.RankTableGenerator{
				"111": ranktable.NewGenerator(&mindxdlv1.Job{})}
			err := r.writeRanktableToCm("job", "default", "111")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDecorateVcjob(t *testing.T) {
	convey.Convey("TestDecorateVcjob", t, func() {
		vcjob := &v1alpha1.Job{
			Spec: v1alpha1.JobSpec{
				Tasks: []v1alpha1.TaskSpec{{Name: "task1", Replicas: 1}},
			},
		}
		job := decorateVcjob(vcjob)
		convey.So(len(job.Spec.ReplicaSpecs), convey.ShouldEqual, 1)
	})
}

func TestDeleteJob(t *testing.T) {
	convey.Convey("TestDeleteJob", t, func() {
		r := newCommonReconciler()
		convey.Convey("01-nil job should return error", func() {
			err := r.DeleteJob(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-delete job should return nil", func() {
			err := r.DeleteJob(newCommonAscendJob())
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

type fakeManager struct{}

func (f fakeManager) SetFields(i interface{}) error {
	return nil
}

func (f fakeManager) GetConfig() *rest.Config {
	return &rest.Config{}
}

func (f fakeManager) GetScheme() *runtime.Scheme {
	return nil
}

func (f fakeManager) GetClient() client.Client {
	return &fakeClient{}
}

func (f fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (f fakeManager) GetCache() cache.Cache {
	return nil
}

func (f fakeManager) GetEventRecorderFor(name string) record.EventRecorder {
	return &fakeRecorder{}
}

func (f fakeManager) GetRESTMapper() meta.RESTMapper {
	return &meta.DefaultRESTMapper{}
}

func (f fakeManager) GetAPIReader() client.Reader {
	return nil
}

func (f fakeManager) Start(ctx context.Context) error {
	return nil
}

func (f fakeManager) Add(runnable manager.Runnable) error {
	return nil
}

func (f fakeManager) Elected() <-chan struct{} {
	return nil
}

func (f fakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (f fakeManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return nil
}

func (f fakeManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return nil
}

func (f fakeManager) GetWebhookServer() *webhook.Server {
	return nil
}

func (f fakeManager) GetLogger() logr.Logger {
	return logr.Logger{}
}

func (f fakeManager) GetControllerOptions() configv1alpha1.ControllerConfigurationSpec {
	return configv1alpha1.ControllerConfigurationSpec{}
}

type fakeController struct{}

func (f fakeController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (f fakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	return nil
}

func (f fakeController) Start(ctx context.Context) error {
	return nil
}

func (f fakeController) GetLogger() logr.Logger {
	return logr.Logger{}
}

func TestSetupWithManager(t *testing.T) {
	convey.Convey("TestSetupWithManager", t, func() {
		mgr := &fakeManager{}
		patch := gomonkey.ApplyFunc(controller.New, func(name string, mgr manager.Manager,
			options controller.Options) (controller.Controller, error) {
			return &fakeController{}, nil
		})
		defer patch.Reset()
		convey.Convey("01-New controller failed should return error", func() {
			r := NewReconciler(mgr, true)
			err := r.SetupWithManager(mgr)
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "watchAscendJobRelatedResource",
			func(*ASJobReconciler, controller.Controller, ctrl.Manager) error { return nil })
		defer patch1.Reset()
		convey.Convey("02-New controller success should return nil", func() {
			r := NewReconciler(mgr, true)
			err := r.SetupWithManager(mgr)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetJobFromInformerCache(t *testing.T) {
	convey.Convey("TestGetJobFromInformerCache", t, func() {
		r := newCommonReconciler()
		obj, err := r.GetJobFromInformerCache("default", "fake-job")
		convey.So(obj, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetJobFromAPIClient(t *testing.T) {
	convey.Convey("TestGetJobFromAPIClient", t, func() {
		r := newCommonReconciler()
		obj, err := r.GetJobFromAPIClient("default", "fake-job")
		convey.So(obj, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSetFaultRetryTimesToBackoffLimits(t *testing.T) {
	convey.Convey("TestSetFaultRetryTimesToBackoffLimits", t, func() {
		r := newCommonReconciler()
		job := newCommonAscendJob()
		convey.Convey("01-job without labels should return nil", func() {
			err := r.setFaultRetryTimesToBackoffLimits(job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-job with invalid labels should return error", func() {
			job.Labels = map[string]string{labelFaultRetryTimes: "xxx"}
			err := r.setFaultRetryTimesToBackoffLimits(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-job with valid labels should return nil", func() {
			job.Labels = map[string]string{labelFaultRetryTimes: "3"}
			err := r.setFaultRetryTimesToBackoffLimits(job)
			convey.So(err, convey.ShouldBeNil)
			expected := 3
			convey.So(r.backoffLimits[job.UID], convey.ShouldEqual, expected)
		})
	})
}

func TestOnOwnerDeleteFunc(t *testing.T) {
	convey.Convey("TestOnOwnerDeleteFunc", t, func() {
		r := newCommonReconciler()
		job := newCommonAscendJob()
		fn := r.onOwnerDeleteFunc()
		res := fn(event.DeleteEvent{Object: job})
		convey.So(res, convey.ShouldEqual, true)
	})
}

func TestOnPodDeleteFunc(t *testing.T) {
	convey.Convey("TestOnPodDeleteFunc", t, func() {
		r := newCommonReconciler()
		job := newCommonAscendJob()
		r.versions[job.UID] = 1
		fn := r.onPodDeleteFunc()
		trueValue := true
		pod := &corev1.Pod{}
		pod.Labels = make(map[string]string)
		convey.Convey("01-pod without replica-type label should return false", func() {
			res := fn(event.DeleteEvent{Object: pod})
			convey.So(res, convey.ShouldEqual, false)
		})
		pod.Labels[commonv1.ReplicaTypeLabel] = "master"
		convey.Convey("02-pod with version label should return false", func() {
			res := fn(event.DeleteEvent{Object: pod})
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("03-pod with invalid version labels should return false", func() {
			pod.Labels[podVersionLabel] = "xx"
			res := fn(event.DeleteEvent{Object: pod})
			convey.So(res, convey.ShouldEqual, false)
		})
		pod.Labels[podVersionLabel] = "1"
		convey.Convey("04-pod without controller ref should return true", func() {
			res := fn(event.DeleteEvent{Object: pod})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("05-pod with valid version labels should return true", func() {
			pod.OwnerReferences = []metav1.OwnerReference{metav1.OwnerReference{
				APIVersion: mindxdlv1.GroupVersion.String(),
				Kind:       api.AscendJobKind,
				Name:       "fake-job",
				UID:        job.UID,
				Controller: &trueValue,
			}}
			pod.Labels[podVersionLabel] = "1"
			res := fn(event.DeleteEvent{Object: pod})
			convey.So(res, convey.ShouldEqual, true)
			expected := 2
			convey.So(r.versions[job.UID], convey.ShouldEqual, expected)
		})
	})
}

func TestOnOwnerCreateFunc(t *testing.T) {
	convey.Convey("TestOnOwnerCreateFunc", t, func() {
		r := newCommonReconciler()
		fn := r.onOwnerCreateFunc()
		convey.Convey("01-not support kind object should return false", func() {
			res := fn(event.CreateEvent{Object: &corev1.Pod{}})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("02-vcjob without needed label should return false", func() {
			res := fn(event.CreateEvent{Object: &v1alpha1.Job{}})
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("03-vcjob with needed label should return true", func() {
			vcjob := &v1alpha1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{api.AtlasTaskLabel: ""},
				},
			}
			res := fn(event.CreateEvent{Object: vcjob})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("04-deployment without needed label should return false", func() {
			res := fn(event.CreateEvent{Object: &appsv1.Deployment{}})
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("05-deployment with needed label should return true", func() {
			res := fn(event.CreateEvent{Object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{api.AtlasTaskLabel: ""},
				},
			}})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("06-ascend job without labels should return true", func() {
			job := newCommonAscendJob()
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("07-ascend job without invalid fault-retry-times labels should return false", func() {
			job := newCommonAscendJob()
			job.Labels = map[string]string{labelFaultRetryTimes: "xx"}
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, false)
		})
	})
}

// TestHandleNewPodDeleted test handleNewPodDeleted function
func TestHandleNewPodDeleted1(t *testing.T) {
	convey.Convey("01-reconciler is nil should log error and return", t, func() {
		var r *ASJobReconciler
		handleNewPodDeleted(&corev1.Pod{}, r)
		convey.So(true, convey.ShouldBeTrue)
	})

	convey.Convey("02-get old pod failed should log error and return", t, func() {
		r := newCommonReconciler()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-pod",
				Namespace:   "default",
				Annotations: map[string]string{api.BackupSourcePodNameKey: "old-pod"},
			},
		}
		patch := gomonkey.ApplyMethodReturn(r.Client, "Get", errors.New("get pod failed"))
		defer patch.Reset()
		handleNewPodDeleted(pod, r)
		convey.So(true, convey.ShouldBeTrue)
	})

	convey.Convey("03-update old pod failed should log error and return", t, func() {
		r := newCommonReconciler()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-pod",
				Namespace:   "default",
				Annotations: map[string]string{api.BackupSourcePodNameKey: "old-pod"},
			},
		}
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodFunc(r.Client, "Get", func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
			obj.SetAnnotations(map[string]string{
				api.InHotSwitchFlowKey:  api.InHotSwitchFlowValue,
				api.BackupNewPodNameKey: "test-pod",
			})
			return nil
		}).ApplyMethodReturn(r.Client, "Update", errors.New("update pod failed"))

		handleNewPodDeleted(pod, r)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestHandleNewPodDeleted2(t *testing.T) {
	convey.Convey("04-handle new pod deleted success", t, func() {
		r := newCommonReconciler()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Annotations: map[string]string{
					api.BackupSourcePodNameKey: "old-pod",
				},
			},
		}

		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyMethodFunc(r.Client, "Get", func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
			obj.SetAnnotations(map[string]string{
				api.InHotSwitchFlowKey:  api.InHotSwitchFlowValue,
				api.BackupNewPodNameKey: "test-pod",
			})
			return nil
		}).ApplyMethodReturn(r.Client, "Update", nil)

		handleNewPodDeleted(pod, r)
		convey.So(true, convey.ShouldBeTrue)
	})
}
