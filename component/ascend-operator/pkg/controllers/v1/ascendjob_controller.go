/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

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
	"github.com/kubeflow/common/pkg/controller.v1/expectation"
	commonutil "github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/training-operator/pkg/common/util"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	volcanoclient "volcano.sh/apis/pkg/client/clientset/versioned"

	mindxdlv1 "ascend-operator/pkg/api/v1"
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
	}

	cfg := mgr.GetConfig()
	kubeClientSet := kubeclientset.NewForConfigOrDie(cfg)
	volcanoClientSet := volcanoclient.NewForConfigOrDie(cfg)
	sharedInformers := informers.NewSharedInformerFactory(kubeClientSet, 0)
	priorityClassInformer := sharedInformers.Scheduling().V1beta1().PriorityClasses()

	r.JobController = common.JobController{
		Controller:                  r,
		Expectations:                expectation.NewControllerExpectations(),
		Config:                      common.JobControllerConfiguration{EnableGangScheduling: enableGangScheduling},
		WorkQueue:                   &util.FakeWorkQueue{},
		Recorder:                    r.recorder,
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
	versions      map[types.UID]int32
	backoffLimits map[types.UID]int32
}

//+kubebuilder:rbac:groups=mindxdl.gitee.com,resources=msjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mindxdl.gitee.com,resources=msjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mindxdl.gitee.com,resources=msjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AscendJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ASJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r == nil {
		return ctrl.Result{}, errors.New("nil pointer")
	}
	ascendjob := &mindxdlv1.AscendJob{}
	err := r.Get(ctx, req.NamespacedName, ascendjob)
	if err != nil {
		hwlog.RunLog.Warnf("unable to fetch AscendJob<%s>, err: %s", req.NamespacedName.String(), err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err = r.validateV1ReplicaSpecs(ascendjob, ascendjob.Spec.ReplicaSpecs); err != nil {
		hwlog.RunLog.Errorf("AscendJob<%s> failed validation, err: %s", req.NamespacedName.String(), err)
		return ctrl.Result{}, r.UpdateJobStatusInApiServer(ascendjob, &ascendjob.Status)
	}

	// Check if reconciliation is needed
	jobKey, err := common.KeyFunc(ascendjob)
	if err != nil {
		hwlog.RunLog.Errorf("couldn't get jobKey for job object %#v: %v", ascendjob, err)
		utilruntime.HandleError(fmt.Errorf("couldn't get jobKey for job object %#v: %v", ascendjob, err))
	}

	replicaTypes := util.GetReplicaTypes(ascendjob.Spec.ReplicaSpecs)
	needReconcile := util.SatisfiedExpectations(r.Expectations, jobKey, replicaTypes)

	if !needReconcile || ascendjob.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled, job<%s> does not need to do reconcile or has been deleted",
			"need sync: %v, have deleted: %v", req.NamespacedName.String(), needReconcile,
			ascendjob.GetDeletionTimestamp() != nil)
		delete(r.versions, ascendjob.UID)
		delete(r.backoffLimits, ascendjob.UID)
		return ctrl.Result{}, nil
	}

	// Set default priorities to ascendJob
	r.Scheme.Default(ascendjob)

	// Use common to reconcile the job related pod and service
	err = r.ReconcileJobs(ascendjob, ascendjob.Spec.ReplicaSpecs, ascendjob.Status, &ascendjob.Spec.RunPolicy)
	if err != nil {
		if k8serr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		hwlog.RunLog.Warnf("Reconcile AscendJob<%s> failed err: %s", req.NamespacedName.String(), err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
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

	// using onOwnerCreateFunc is easier to set defaults
	if err = c.Watch(&source.Kind{Type: &mindxdlv1.AscendJob{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}

	// inject watching for job related pod
	if err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mindxdlv1.AscendJob{},
	}, predicate.Funcs{
		CreateFunc: util.OnDependentCreateFunc(r.Expectations),
		UpdateFunc: util.OnDependentUpdateFunc(&r.JobController),
		DeleteFunc: r.onPodDeleteFunc(r.Expectations),
	}); err != nil {
		return err
	}

	// inject watching for job related service
	if err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mindxdlv1.AscendJob{},
	}, predicate.Funcs{
		CreateFunc: util.OnDependentCreateFunc(r.Expectations),
		UpdateFunc: util.OnDependentUpdateFunc(&r.JobController),
		DeleteFunc: util.OnDependentDeleteFunc(r.Expectations),
	}); err != nil {
		return err
	}

	if r.Config.EnableGangScheduling {
		_, err = mgr.GetRESTMapper().RESTMapping(schema.GroupKind{Group: v1beta1.SchemeGroupVersion.Group, Kind: "PodGroup"},
			v1beta1.SchemeGroupVersion.Version)
		if err != nil {
			hwlog.RunLog.Infof("enableGangScheduling is true, but PodGroup is not in cluster")
			return err
		}
		if err = c.Watch(&source.Kind{Type: &v1beta1.PodGroup{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.AscendJob{},
		}, predicate.Funcs{
			CreateFunc: util.OnDependentCreateFunc(r.Expectations),
			UpdateFunc: util.OnDependentUpdateFuncGeneric(&r.JobController),
			DeleteFunc: util.OnDependentDeleteFunc(r.Expectations),
		}); err != nil {
			return err
		}
	}

	return nil
}
func (r *ASJobReconciler) onOwnerCreateFunc() func(event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		ascendJob, ok := e.Object.(*mindxdlv1.AscendJob)
		if !ok {
			return true
		}
		msg := fmt.Sprintf("AscendJob %s is create.", e.Object.GetName())
		hwlog.RunLog.Info(msg)
		err := commonutil.UpdateJobConditions(&ascendJob.Status, commonv1.JobCreated, "AscendCreated", msg)
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
		hwlog.RunLog.Debugf("now backoffLimits: %v", r.backoffLimits)
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
		hwlog.RunLog.Warnf("assigns the value of fault-retry-times to backoffLimits, fault-retry-times: %v", faultRetryTimes)
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
		return true
	}
}

// onPodDeleteFunc does some necessary processing logic when a pod is deleted.
func (r *ASJobReconciler) onPodDeleteFunc(
	exp expectation.ControllerExpectationsInterface) func(event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		replicaType, ok := e.Object.GetLabels()[commonv1.ReplicaTypeLabel]
		if !ok || len(replicaType) == 0 {
			return false
		}

		if controllerRef := metav1.GetControllerOf(e.Object); controllerRef != nil {
			jobKey := fmt.Sprintf("%s/%s", e.Object.GetNamespace(), controllerRef.Name)
			var expectKey string
			if _, ok := e.Object.(*corev1.Pod); ok {
				expectKey = expectation.GenExpectationPodsKey(jobKey, replicaType)
			}

			if _, ok := e.Object.(*corev1.Service); ok {
				expectKey = expectation.GenExpectationServicesKey(jobKey, replicaType)
			}

			exp.DeletionObserved(expectKey)
			return r.dealPodVersion(e, controllerRef)
		}
		return true
	}
}

func (r *ASJobReconciler) dealPodVersion(e event.DeleteEvent, controllerRef *metav1.OwnerReference) bool {
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
	currentVersion, ok := r.versions[controllerRef.UID]
	if ok && int32(versionNumber) == currentVersion {
		r.versions[controllerRef.UID]++
	}
	return true
}

func (r *ASJobReconciler) ControllerName() string {
	return controllerName
}

func (r *ASJobReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(mindxdlv1.Kind)
}

func (r *ASJobReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

func (r *ASJobReconciler) GetGroupNameLabelValue() string {
	return mindxdlv1.GroupVersion.Group
}

func (r *ASJobReconciler) GetJobFromInformerCache(namespace, name string) (metav1.Object, error) {
	ascendjob := &mindxdlv1.AscendJob{}
	err := r.Get(context.Background(), types.NamespacedName{
		Namespace: namespace, Name: name,
	}, ascendjob)
	return ascendjob, err
}

func (r *ASJobReconciler) GetJobFromAPIClient(namespace, name string) (metav1.Object, error) {
	job := &mindxdlv1.AscendJob{}

	err := r.apiReader.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, job)
	if err != nil {
		if k8serr.IsNotFound(err) {
			hwlog.RunLog.Errorf("AscendJob<%s-%s> not found, err: %s", namespace, name, err)
		} else {
			hwlog.RunLog.Errorf("failed to get AscendJob from api-server", namespace, name, err)
		}
		return nil, err
	}
	return job, nil
}

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

func (r *ASJobReconciler) UpdateJobStatus(
	job interface{},
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus *commonv1.JobStatus) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendJob)
	}

	jobKey, err := common.KeyFunc(ascendJob)
	if err != nil {
		hwlog.RunLog.Errorf("couldn't get key for ascendJob object %#v: %v", ascendJob, err)
		utilruntime.HandleError(fmt.Errorf("couldn't get key for ascendJob object %#v: %v", ascendJob, err))
		return err
	}

	worker0Completed, err := r.IsWorker0Completed(ascendJob, replicas)
	if err != nil {
		hwlog.RunLog.Warnf("check if worker 0 completed error %s", err)
		return err
	}

	// Set StartTime.
	if jobStatus.StartTime == nil {
		now := metav1.Now()
		jobStatus.StartTime = &now
		// enqueue a sync to check if job past ActiveDeadlineSeconds
		if ascendJob.Spec.RunPolicy.ActiveDeadlineSeconds != nil {
			hwlog.RunLog.Infof("Job with ActiveDeadlineSeconds will sync after %d seconds",
				*ascendJob.Spec.RunPolicy.ActiveDeadlineSeconds)
			r.WorkQueue.AddAfter(jobKey, time.Duration(*ascendJob.Spec.RunPolicy.ActiveDeadlineSeconds)*time.Second)
		}
	}

	status := commonv1.ReplicaStatus{}
	for _, st := range jobStatus.ReplicaStatuses {
		status.Active += st.Active
		status.Succeeded += st.Succeeded
		status.Failed += st.Failed
	}

	var existingRestartingCondition *commonv1.JobCondition
	for _, condition := range jobStatus.Conditions {
		if condition.Type == commonv1.JobRestarting {
			existingRestartingCondition = &commonv1.JobCondition{
				Reason:  condition.Reason,
				Message: condition.Message,
			}
		}
	}

	if status.Failed > 0 {
		if err = r.reconcileWithFailed(ascendJob, jobStatus, existingRestartingCondition); err != nil {
			return err
		}
	} else if status.Succeeded == getTotalTrainReplicas(ascendJob) || (worker0Completed &&
		*ascendJob.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyAllWorkers) {
		msg := fmt.Sprintf("AscendJob<%s-%s> successfully completed.",
			ascendJob.Namespace, ascendJob.Name)
		r.recorder.Event(ascendJob, corev1.EventTypeNormal, commonutil.JobRunningReason, msg)
		if jobStatus.CompletionTime == nil {
			now := metav1.Now()
			jobStatus.CompletionTime = &now
		}
		updateErr := commonutil.UpdateJobConditions(jobStatus,
			commonv1.JobSucceeded, commonutil.JobRunningReason, msg)
		if updateErr != nil {
			hwlog.RunLog.Errorf("Append ascendJob<%s-%s> condition err: %v",
				ascendJob.Namespace, ascendJob.Name, updateErr)
			return updateErr
		}
	} else {
		msg := fmt.Sprintf("AscendJob %s/%s is running.", ascendJob.Namespace, ascendJob.Name)
		updateErr := commonutil.UpdateJobConditions(jobStatus, commonv1.JobRunning,
			commonutil.JobRunningReason, msg)
		if updateErr != nil {
			hwlog.RunLog.Errorf("Append ascendJob<%s-%s> condition err: %v",
				ascendJob.Namespace, ascendJob.Name, err)
			return updateErr
		}
	}

	// we assign the jobStatus to the msJob.Status for testing purpose
	// it won't effect the main reconcile logic
	// because we already use oldStatus := jobStatus.DeepCopy() to record the oldStatus
	// and use !reflect.DeepEqual(*oldStatus, jobStatus) to decide whether to update the msJob or not
	ascendJob.Status = *jobStatus.DeepCopy()

	return nil
}

func (r *ASJobReconciler) reconcileWithFailed(ascendJob *mindxdlv1.AscendJob,
	jobStatus *commonv1.JobStatus, restartCondition *commonv1.JobCondition) error {
	if restartCondition != nil {
		err := commonutil.UpdateJobConditions(jobStatus, commonv1.JobRestarting,
			restartCondition.Reason, restartCondition.Message)
		if err != nil {
			hwlog.RunLog.Errorf("Append ascendJob condition failed error: %v",
				ascendJob.Namespace, ascendJob.Name, err)
			return err
		}
		return nil
	}

	if !r.isUnconditionalRetryJob(ascendJob) {
		msg := fmt.Sprintf("AscendJob <%s/%s> has failed because has pod failed.", ascendJob.Namespace,
			ascendJob.Name)
		return r.setJobFailedCondition(ascendJob, jobStatus, msg)
	}

	rt, err := r.getJobRemainRetryTimes(ascendJob)
	if err != nil {
		hwlog.RunLog.Warnf("getJobRemainRetryTimes failed, err %s", err)
		return nil
	}

	if rt == 0 {
		msg := fmt.Sprintf("AscendJob <%s/%s> has failed because pod failed and remain retry times is 0.",
			ascendJob.Namespace, ascendJob.Name)
		return r.setJobFailedCondition(ascendJob, jobStatus, msg)
	}
	return nil
}

func (r *ASJobReconciler) setJobFailedCondition(ascendJob *mindxdlv1.AscendJob, jobStatus *commonv1.JobStatus,
	msg string) error {
	r.recorder.Event(ascendJob, corev1.EventTypeNormal, commonutil.JobFailedReason, msg)
	if jobStatus.CompletionTime == nil {
		now := metav1.Now()
		jobStatus.CompletionTime = &now
	}
	err := commonutil.UpdateJobConditions(jobStatus,
		commonv1.JobFailed, commonutil.JobFailedReason, msg)
	if err != nil {
		hwlog.RunLog.Errorf("Append ascendjob<%s/%s> condition error: %v", ascendJob.Namespace, ascendJob.Name, err)
		return err
	}
	return nil
}

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
func (r *ASJobReconciler) SetClusterSpec(job interface{}, podTemplate *corev1.PodTemplateSpec, rtype, index string) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	ascendjob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendjob)
	}

	if ascendjob == nil || ascendjob.Spec.ReplicaSpecs == nil {
		return fmt.Errorf("job or job specs is nil")
	}

	if err := r.setPodEnvironment(ascendjob, podTemplate, rtype, index); err != nil {
		return err
	}
	return r.setPodAnnotation(ascendjob, podTemplate, rtype, index)
}

func (r *ASJobReconciler) GetDefaultContainerName() string {
	return mindxdlv1.DefaultContainerName
}

func (r *ASJobReconciler) GetDefaultContainerPortName() string {
	return mindxdlv1.DefaultPortName
}

func (r *ASJobReconciler) IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	rtype commonv1.ReplicaType, index int) bool {
	if ContainsChiefOrMasterSpec(replicas) {
		return rtype == mindxdlv1.TensorflowReplicaTypeChief || rtype == mindxdlv1.PytorchReplicaTypeMaster
	}
	return rtype == mindxdlv1.ReplicaTypeWorker && index == 0
}

// IsWorker0Completed returns true if pod of worker0 succeeded and exited with 0
func (r *ASJobReconciler) IsWorker0Completed(ascendJob *mindxdlv1.AscendJob, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) (bool, error) {
	if r == nil {
		return false, errors.New("nil pointer")
	}
	worker0Completed := false
	_, ok := replicas[mindxdlv1.ReplicaTypeWorker]
	if !ok {
		return true, nil
	}
	podSlices, err := r.getPodSlices(ascendJob, replicas[mindxdlv1.ReplicaTypeWorker].Replicas)
	if err != nil {
		return false, err
	}
	for index, podSlice := range podSlices {
		if len(podSlice) == 1 {
			pod := podSlice[0]
			exitCode := getContainerExitCode(pod)
			if index == 0 && exitCode == 0 && pod.Status.Phase == v1.PodSucceeded {
				worker0Completed = true
			}
		}
	}
	return worker0Completed, nil
}

// getPodSlices returns a slice, which element is the slice of pod.
// It gives enough information to caller to make decision to up/down scale resources.
func (r *ASJobReconciler) getPodSlices(job *mindxdlv1.AscendJob, replicasNum *int32) ([][]*v1.Pod, error) {
	pods, err := r.GetPodsForJob(job)
	if err != nil {
		hwlog.RunLog.Errorf("get job<%s-%s> pods failed error %v", job.Namespace, job.Name, err)
		return nil, err
	}

	// Get all pods for the type rt.
	pods, err = r.JobController.FilterPodsForReplicaType(pods, strings.ToLower(string(mindxdlv1.ReplicaTypeWorker)))
	if err != nil {
		return nil, err
	}

	podSlices := r.GetPodSlices(pods, int(*replicasNum))
	return podSlices, nil
}
