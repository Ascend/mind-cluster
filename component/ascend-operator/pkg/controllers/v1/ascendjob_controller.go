/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/controller.v1/control"
	"github.com/kubeflow/common/pkg/util"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable"
	"ascend-operator/pkg/ranktable/generator"
)

// NewReconciler new reconciler for AscendJob
func NewReconciler(mgr manager.Manager, enableGangScheduling bool) *ASJobReconciler {
	r := &ASJobReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		apiReader:     mgr.GetAPIReader(),
		recorder:      mgr.GetEventRecorderFor(controllerName),
		versions:      make(map[types.UID]int32),
		backoffLimits: make(map[types.UID]int32),
		rtGenerators:  make(map[types.UID]generator.RankTableGenerator),
		hasHcclCtr:    false,
	}

	cfg := mgr.GetConfig()
	kubeClientSet := kubernetes.NewForConfigOrDie(cfg)
	volcanoClientSet := versioned.NewForConfigOrDie(cfg)
	sharedInformers := informers.NewSharedInformerFactory(kubeClientSet, 0)
	priorityClassInformer := sharedInformers.Scheduling().V1beta1().PriorityClasses()

	r.JobController = common.JobController{
		Controller:                  r,
		Config:                      common.JobControllerConfiguration{EnableGangScheduling: enableGangScheduling},
		WorkQueue:                   workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		Recorder:                    r.recorder,
		PodLister:                   sharedInformers.Core().V1().Pods().Lister(),
		ServiceLister:               sharedInformers.Core().V1().Services().Lister(),
		KubeClientSet:               kubeClientSet,
		VolcanoClientSet:            volcanoClientSet,
		PriorityClassLister:         priorityClassInformer.Lister(),
		PriorityClassInformerSynced: priorityClassInformer.Informer().HasSynced,
		PodControl:                  control.RealPodControl{KubeClient: kubeClientSet, Recorder: r.recorder},
		ServiceControl:              control.RealServiceControl{KubeClient: kubeClientSet, Recorder: r.recorder},
	}

	return r
}

// ASJobReconciler reconciles a AscendJob object
type ASJobReconciler struct {
	common.JobController
	client.Client
	Scheme        *runtime.Scheme
	recorder      record.EventRecorder
	apiReader     client.Reader
	hasHcclCtr    bool
	versions      map[types.UID]int32
	backoffLimits map[types.UID]int32
	rtGenerators  map[types.UID]generator.RankTableGenerator
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the AscendJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *ASJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r == nil {
		return ctrl.Result{}, errors.New("nil pointer")
	}
	if r.isVcjobOrDeploy(ctx, req) {
		return ctrl.Result{}, nil
	}
	ascendjob := &mindxdlv1.AscendJob{}
	if err := r.Get(ctx, req.NamespacedName, ascendjob); err != nil {
		if k8serr.IsNotFound(err) {
			hwlog.RunLog.Debugf("unable to fetch AscendJob<%s>, err: %s", req.NamespacedName, err)
		} else {
			hwlog.RunLog.Errorf("unable to fetch AscendJob<%s>, err: %s", req.NamespacedName, err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.validateJob(ascendjob); err != nil {
		hwlog.RunLog.Errorf("AscendJob<%s> failed validation, err: %v", req.NamespacedName, err)
		return ctrl.Result{}, r.UpdateJobStatusInApiServer(ascendjob, &ascendjob.Status)
	}

	if ascendjob.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled，job<%s> has been deleted", req.NamespacedName)
		delete(r.versions, ascendjob.UID)
		delete(r.backoffLimits, ascendjob.UID)
		return ctrl.Result{}, nil
	}

	// Set default priorities to ascendJob
	r.Scheme.Default(ascendjob)

	// Use common to reconcile the job related pod and service
	err := r.ReconcileJobs(ascendjob, ascendjob.Spec.ReplicaSpecs, ascendjob.Status, &ascendjob.Spec.RunPolicy)
	if err != nil {
		if k8serr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		hwlog.RunLog.Warnf("Reconcile AscendJob<%s> failed err: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ASJobReconciler) isVcjobOrDeploy(ctx context.Context, req ctrl.Request) bool {
	if r.hasHcclCtr {
		hwlog.RunLog.Info("hccl controller exist, do not handle vcjob and deployment")
		return false
	}
	// try fetch deployment
	deploy := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deploy); err == nil {
		if strings.Contains(req.Name, hcclCtrName) {
			r.hasHcclCtr = true
			hwlog.RunLog.Info("hccl controller found")
			return true
		}
		r.ranktablePipeline(decorateDeploy(deploy))
		return true
	}
	// try fetch vcjob
	vcjob := &v1alpha1.Job{}
	if err := r.Get(ctx, req.NamespacedName, vcjob); err == nil {
		r.ranktablePipeline(decorateVcjob(vcjob))
		return true
	}
	return false
}

func (r *ASJobReconciler) ranktablePipeline(job *mindxdlv1.AscendJob) {
	ji, err := r.newJobInfo(job, job.Spec.ReplicaSpecs, &job.Status, &job.Spec.RunPolicy)
	if err != nil {
		hwlog.RunLog.Errorf("failed to generate ranktable for job<%s> in namespace<%s>, err: %v", job.Name, job.Namespace, err)
		return
	}
	r.genRankTable(ji)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ASJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	c, err := controller.New(r.ControllerName(), mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	return r.watchRelatedResource(c, mgr)
}

type resourceOption struct {
	kind          *source.Kind
	predicateFunc predicate.Funcs
}

func (r *ASJobReconciler) watchRelatedResource(c controller.Controller, mgr ctrl.Manager) error {
	if err := c.Watch(&source.Kind{Type: &v1alpha1.Job{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &appv1.Deployment{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &mindxdlv1.AscendJob{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}

	resourceOptions := []*resourceOption{
		{kind: &source.Kind{Type: &corev1.Pod{}}, predicateFunc: predicate.Funcs{DeleteFunc: r.onPodDeleteFunc()}},
		{kind: &source.Kind{Type: &corev1.Service{}}},
	}

	if r.Config.EnableGangScheduling {
		_, mapErr := mgr.GetRESTMapper().RESTMapping(schema.GroupKind{Group: v1beta1.SchemeGroupVersion.Group,
			Kind: "PodGroup"},
			v1beta1.SchemeGroupVersion.Version)
		if mapErr != nil {
			hwlog.RunLog.Errorf("enableGangScheduling is true, but PodGroup is not in cluster")
			return mapErr
		}
		resourceOptions = append(resourceOptions, &resourceOption{kind: &source.Kind{Type: &v1beta1.PodGroup{}}})
	}

	for _, src := range resourceOptions {
		if err := c.Watch(src.kind, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.AscendJob{},
		}, src.predicateFunc); err != nil {
			return err
		}
	}
	return nil
}

func (r *ASJobReconciler) onOwnerCreateFunc() func(event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		switch e.Object.(type) {
		case *v1alpha1.Job:
			vcjob := e.Object.(*v1alpha1.Job)
			r.rtGenerators[vcjob.UID] = ranktable.NewGenerator(decorateVcjob(vcjob))
			hwlog.RunLog.Infof("create rtGenerator for Volcano Job %s", vcjob.Name)
			return true
		case *appv1.Deployment:
			deploy := e.Object.(*appv1.Deployment)
			r.rtGenerators[deploy.UID] = ranktable.NewGenerator(decorateDeploy(deploy))
			hwlog.RunLog.Infof("create rtGenerator for Deployment %s", deploy.Name)
			return true
		default:
			hwlog.RunLog.Info("job type is not volcano job or deployment")
		}

		ascendJob, ok := e.Object.(*mindxdlv1.AscendJob)
		if !ok {
			return true
		}
		msg := fmt.Sprintf("AscendJob %s is create.", e.Object.GetName())
		hwlog.RunLog.Info(msg)
		err := util.UpdateJobConditions(&ascendJob.Status, commonv1.JobCreated, "AscendCreated", msg)
		if err != nil {
			log.Log.Error(err, "append job condition error")
			return false
		}
		r.versions[ascendJob.UID] = defaultPodVersion
		r.backoffLimits[ascendJob.UID] = unsetBackoffLimits
		if ascendJob.Spec.RunPolicy.BackoffLimit != nil {
			r.backoffLimits[ascendJob.UID] = *ascendJob.Spec.RunPolicy.BackoffLimit
		} else if err = r.setFaultRetryTimesToBackoffLimits(ascendJob); err != nil {
			hwlog.RunLog.Errorf("failed to get fault-retry-times, error: %v", err)
			return false
		}
		if frame, err := mindxdlv1.GetJobFramework(ascendJob); err == nil {
			r.rtGenerators[ascendJob.UID] = ranktable.NewGenerator(ascendJob)
			hwlog.RunLog.Infof("create rtGenerator for frame %s ascendJob %s", frame, ascendJob.Name)
		}

		return true
	}
}

// setFaultRetryTimesToBackoffLimits assigns the value of fault-retry-times to backoffLimits.
func (r *ASJobReconciler) setFaultRetryTimesToBackoffLimits(ascendJob *mindxdlv1.AscendJob) error {
	if len(ascendJob.ObjectMeta.Labels) == 0 {
		return nil
	}
	if value, ok := ascendJob.ObjectMeta.Labels[labelFaultRetryTimes]; ok && value != "" {
		faultRetryTimes, err := strconv.Atoi(value)
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert string to int, error: %v", err)
			return err
		}
		r.backoffLimits[ascendJob.UID] = int32(faultRetryTimes)
	}
	return nil
}

func (r *ASJobReconciler) onOwnerDeleteFunc() func(deleteEvent event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		ascendJob, ok := e.Object.(*mindxdlv1.AscendJob)
		if !ok {
			return false
		}
		msg := fmt.Sprintf("AscendJob %s is deleted.", e.Object.GetName())
		hwlog.RunLog.Info(msg)
		delete(r.versions, ascendJob.UID)
		delete(r.backoffLimits, ascendJob.UID)
		if rtg, ok := r.rtGenerators[ascendJob.UID]; ok {
			if err := rtg.DeleteFile(); err != nil {
				hwlog.RunLog.Errorf("failed to delete ranktable, err: %v", err)
			}
			delete(r.rtGenerators, ascendJob.UID)
		}
		return true
	}
}

// onPodDeleteFunc does some necessary processing logic when a pod is deleted.
func (r *ASJobReconciler) onPodDeleteFunc() func(event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		controllerRef := metav1.GetControllerOf(e.Object)
		if controllerRef != nil {
			if rtg, exist := r.rtGenerators[controllerRef.UID]; exist {
				rtg.DeletePod(e.Object.(*corev1.Pod))
			}
		}
		replicaType, ok := e.Object.GetLabels()[commonv1.ReplicaTypeLabel]
		if !ok || len(replicaType) == 0 {
			return false
		}
		version, ok := e.Object.GetLabels()[podVersionLabel]
		if !ok || len(version) == 0 {
			return false
		}
		versionNumber, err := strconv.Atoi(version)
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert string to int, err: %v", err)
			return false
		}
		hwlog.RunLog.Infof("deleted pod version is: %v", version)
		if controllerRef == nil {
			return true
		}
		currentVersion, ok := r.versions[controllerRef.UID]
		if ok && int32(versionNumber) == currentVersion {
			r.versions[controllerRef.UID]++
		}
		return true
	}
}

// ControllerName get controller name
func (r *ASJobReconciler) ControllerName() string {
	return controllerName
}

// GetAPIGroupVersionKind get api group version kind
func (r *ASJobReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(mindxdlv1.Kind)
}

// GetAPIGroupVersion get api group version
func (r *ASJobReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

// GetGroupNameLabelValue get group name label value
func (r *ASJobReconciler) GetGroupNameLabelValue() string {
	return mindxdlv1.GroupVersion.Group
}

// GetJobFromInformerCache get job from informer cache
func (r *ASJobReconciler) GetJobFromInformerCache(namespace, name string) (metav1.Object, error) {
	ascendjob := &mindxdlv1.AscendJob{}
	err := r.Get(context.Background(), types.NamespacedName{
		Namespace: namespace, Name: name,
	}, ascendjob)
	return ascendjob, err
}

// GetJobFromAPIClient get job from api server
func (r *ASJobReconciler) GetJobFromAPIClient(namespace, name string) (metav1.Object, error) {
	job := &mindxdlv1.AscendJob{}

	err := r.apiReader.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, job)
	if err != nil {
		if k8serr.IsNotFound(err) {
			hwlog.RunLog.Warnf("AscendJob<%s/%s> not found, err: %s", namespace, name, err)
		} else {
			hwlog.RunLog.Errorf("failed to get AscendJob<%s/%s> from api-server. err: %s", namespace, name, err)
		}
		return nil, err
	}
	return job, nil
}

// DeleteJob deletes the job.
func (r *ASJobReconciler) DeleteJob(job interface{}) error {
	ascendjob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendjob)
	}

	if err := r.Delete(context.Background(), ascendjob); err != nil {
		r.recorder.Eventf(ascendjob, v1.EventTypeWarning, FailedDeleteJobReason, "Error deleting: %v", err)
		hwlog.RunLog.Errorf("failed to delete job<%s-%s>, err: %s", ascendjob.Namespace, ascendjob.Name, err)
		return err
	}

	r.recorder.Eventf(ascendjob, v1.EventTypeNormal, SuccessfulDeleteJobReason, "Deleted job: %v", ascendjob.Name)
	hwlog.RunLog.Infof("job<%s-%s> has been deleted", ascendjob.Namespace, ascendjob.Name)
	return nil
}

// UpdateJobStatusInApiServer update job status in api-server.
func (r *ASJobReconciler) UpdateJobStatusInApiServer(job interface{}, jobStatus *commonv1.JobStatus) error {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = map[commonv1.ReplicaType]*commonv1.ReplicaStatus{}
	}
	ascendjob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendjob)
	}
	startTime := time.Now()
	defer func() {
		hwlog.RunLog.Infof("Finished updating AscendJob Status %q (%v)",
			ascendjob.Name, time.Since(startTime))
	}()

	ascendjob = ascendjob.DeepCopy()
	ascendjob.Status = *jobStatus.DeepCopy()

	return r.Status().Update(context.Background(), ascendjob)
}

// SetClusterSpec Set Envs for AscendJob
func (r *ASJobReconciler) SetClusterSpec(job interface{}, podTemplate *corev1.PodTemplateSpec,
	rtype, index string) error {
	return nil
}

// GetDefaultContainerName Get default container name
func (r *ASJobReconciler) GetDefaultContainerName() string {
	return mindxdlv1.DefaultContainerName
}

// GetDefaultContainerPortName get default container port name
func (r *ASJobReconciler) GetDefaultContainerPortName() string {
	return mindxdlv1.DefaultPortName
}

// IsMasterRole check whether the role is master
func (r *ASJobReconciler) IsMasterRole(_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	rtype commonv1.ReplicaType, _ int) bool {
	return rtype == mindxdlv1.MindSporeReplicaTypeScheduler ||
		rtype == mindxdlv1.PytorchReplicaTypeMaster ||
		rtype == mindxdlv1.TensorflowReplicaTypeChief
}

func decorateVcjob(vcjob *v1alpha1.Job) *mindxdlv1.AscendJob {
	repSpecs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
	for i, task := range vcjob.Spec.Tasks {
		repSpecs[commonv1.ReplicaType("Vcjob"+strconv.Itoa(i))] = &commonv1.ReplicaSpec{
			Template: task.Template,
			Replicas: &task.Replicas,
		}
	}
	return &mindxdlv1.AscendJob{
		TypeMeta:   vcjob.TypeMeta,
		ObjectMeta: vcjob.ObjectMeta,
		Spec: mindxdlv1.AscendJobSpec{
			ReplicaSpecs: repSpecs,
		},
	}
}

func decorateDeploy(deploy *appv1.Deployment) *mindxdlv1.AscendJob {
	repSpec := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		"Deploy": &commonv1.ReplicaSpec{
			Template: deploy.Spec.Template,
			Replicas: deploy.Spec.Replicas,
		},
	}
	return &mindxdlv1.AscendJob{
		TypeMeta:   deploy.TypeMeta,
		ObjectMeta: deploy.ObjectMeta,
		Spec: mindxdlv1.AscendJobSpec{
			ReplicaSpecs: repSpec,
		},
	}
}

func (r *ASJobReconciler) writeRanktableToCm(jobName, namespace string, ji *jobInfo) error {
	configmapName := configmapPrefix + jobName
	cm := &corev1.ConfigMap{}
	namespacedname := types.NamespacedName{Namespace: namespace, Name: configmapName}
	err := r.Get(context.TODO(), namespacedname, cm)
	if err != nil {
		hwlog.RunLog.Infof("failed to get configmap in namespace %s, err: %v", namespace, err)
		return err
	}
	rtg, ok := r.rtGenerators[ji.mtObj.GetUID()]
	if !ok {
		err = fmt.Errorf("ranktable generaotor not found for job %s", ji.name)
		hwlog.RunLog.Error(err)
		return err
	}
	cm.Data[configmapKey], err = rtg.ToString()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get ranktable string, err: %v", err)
		return err
	}
	hwlog.RunLog.Infof("start write info to configmap<%s> in namespace<%s>", configmapName, namespace)
	if err := r.Update(context.TODO(), cm); err != nil {
		hwlog.RunLog.Errorf("failed to write configmap, err: %v", err)
		return err
	}
	return nil
}
