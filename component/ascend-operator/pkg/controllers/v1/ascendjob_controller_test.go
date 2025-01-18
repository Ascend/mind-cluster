/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/training-operator/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/scheduling/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corelisters "k8s.io/client-go/listers/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func newCommonReconciler() *ASJobReconciler {
	rc := &ASJobReconciler{
		Client:        &fakeClient{},
		recorder:      &fakeRecorder{},
		versions:      make(map[types.UID]int32),
		backoffLimits: make(map[types.UID]int32),
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

func newCommonPodInfo() *podInfo {
	return &podInfo{
		rtype: mindxdlv1.ReplicaTypeWorker,
		ip:    "127.0.0.1",
		job: &mindxdlv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				UID: "123456",
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
				Name:          mindxdlv1.DefaultPortName,
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

func newCommonAscendJob() *mindxdlv1.AscendJob {
	return &mindxdlv1.AscendJob{
		TypeMeta: metav1.TypeMeta{
			Kind: "AscendJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ascendjob-test",
			UID:         "1111",
			Annotations: map[string]string{},
		},
		Spec: mindxdlv1.AscendJobSpec{},
	}
}

func newReplicas(i int) *int32 {
	x := int32(i)
	return &x
}

// TestIsVcjobOrDeploy This function is a test function for the isVcjobOrDeploy method of the AscendJobReconciler struct.
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
	patches := gomonkey.ApplyPrivateMethod(r, "ranktablePipeline", func(job *mindxdlv1.AscendJob) { return })
	defer patches.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.isVcjobOrDeploy(tt.args.ctx, tt.args.req); got != tt.want {
				t.Errorf("isVcjobOrDeploy() = %v, want %v", got, tt.want)
			}
		})
	}
}
