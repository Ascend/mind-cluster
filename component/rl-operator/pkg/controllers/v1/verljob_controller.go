/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	flowcommon "github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "rl-operator/pkg/api/v1"
	"rl-operator/pkg/common"
)

// VerlJobReconciler reconciles a VerlJob object
type VerlJobReconciler struct {
	*JobReconciler
	EnableGangScheduling bool
}

type VerlConfig struct {
	autoSubmit     bool
	verlPath       string
	verlExec       string
	verlConfig     string
	extraConfig    map[string]string
	checkRayStatus bool
}

type VerlJobInfo struct {
	verlJob        *mindxdlv1.VerlJob
	rayClusterList []*mindxdlv1.RayCluster
	podList        []*v1.Pod
}

const checkRayCommand = `
mkdir -p /tmp
stderr_file=$(mktemp) || { echo "Failed to create temp file"; exit 1; }
while true; do
    if ! ray_output=$(PYTHONWARNINGS=ignore ray list nodes --format=json 2>"$stderr_file"); then
        error_message=$(cat "$stderr_file")
		echo "Error: ray list nodes command failed with exit code $?"
		[ -n "$error_message" ] && echo "Error details: $error_message"
		rm -f "$stderr_file"  
        exit 1
    fi

    if ! nodes=$(echo "$ray_output" | python3 -c 'import json,sys; data=json.load(sys.stdin); print(len(data))'); then
        echo "Error: Python processing failed with exit code $?"
        echo "Python error output: $nodes"
        echo "Ray command output was: $ray_output"
        exit 1
    fi
    
    if ! [[ "$nodes" =~ ^[0-9]+$ ]]; then
        echo "Error: Invalid node count received: '$nodes'"
        echo "Ray command output: $ray_output"
        exit 1
    fi
    
    echo "Waiting for ray cluster start, Current ray node count: $nodes"
    
    if [ "$nodes" -ge "$EXPECTED_NODES" ]; then
        break
    fi
    
    sleep 2
done
`

// NewVerlJobReconciler new reconciler for RayCluster
func NewVerlJobReconciler(mgr manager.Manager, enableGangScheduling bool) *VerlJobReconciler {
	baseReconciler := NewJobReconciler(mgr, common.VerlJobControllerName)
	r := &VerlJobReconciler{
		JobReconciler:        baseReconciler,
		EnableGangScheduling: enableGangScheduling,
	}
	r.ConfigInfo = r
	return r
}

func (r *VerlJobReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return mindxdlv1.GroupVersion.WithKind(common.VerlJobResourceKind)
}

func (r *VerlJobReconciler) GetAPIGroupVersion() schema.GroupVersion {
	return mindxdlv1.GroupVersion
}

func (r *VerlJobReconciler) GetRunPolicy(obj client.Object) (*commonv1.RunPolicy, error) {
	verlJob, ok := obj.(*mindxdlv1.VerlJob)
	if !ok {
		return nil, fmt.Errorf("expect VerlJob, but got %v", reflect.TypeOf(obj))
	}
	return &verlJob.Spec.RunPolicy, nil
}

func (r *VerlJobReconciler) GetStatus(obj client.Object) (*commonv1.JobStatus, error) {
	verlJob, ok := obj.(*mindxdlv1.VerlJob)
	if !ok {
		return nil, fmt.Errorf("expect VerlJob, but got %v", reflect.TypeOf(obj))
	}
	return &verlJob.Status, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerlJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return fmt.Errorf("RayClusterReconciler nil pointer")
	}

	c, err := controller.New(common.RayClusterControllerName, mgr, controller.Options{
		Reconciler: r,
		RateLimiter: workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(common.WorkQueueBaseDelay, common.WorkQueueMaxDelay),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(common.WorkQueueQps), common.WorkQueueBurst)},
		),
	})
	if err != nil {
		return err
	}

	return r.watchRelatedResource(c)
}

func (r *VerlJobReconciler) watchRelatedResource(c controller.Controller) error {
	if err := c.Watch(&source.Kind{Type: &mindxdlv1.VerlJob{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{CreateFunc: r.onOwnerCreateFunc(), DeleteFunc: r.onOwnerDeleteFunc()},
	); err != nil {
		return err
	}
	resourceOptions := []*common.ResourceOption{
		{Kind: &source.Kind{Type: &mindxdlv1.RayCluster{}}},
		{Kind: &source.Kind{Type: &v1.Pod{}},
			PredicateFunc: predicate.Funcs{DeleteFunc: r.onPodDeleteFunc()}},
	}

	for _, src := range resourceOptions {
		if err := c.Watch(src.Kind, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mindxdlv1.VerlJob{},
		}, src.PredicateFunc); err != nil {
			return err
		}
	}
	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the VerlJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *VerlJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r == nil {
		return ctrl.Result{}, fmt.Errorf("VerlJobReconciler nil pointer")
	}
	verlJob := &mindxdlv1.VerlJob{}
	if err := r.FetchResource(ctx, req.NamespacedName, verlJob); err != nil {
		return ctrl.Result{}, err
	}
	if common.IsObjectEmpty(verlJob) {
		hwlog.RunLog.Warn("VerlJobReconciler object is empty. Skip reconcile.")
		return ctrl.Result{}, nil
	}
	if verlJob.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Infof("reconcile cancelled，job<%s> has been deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	oldStatus := verlJob.Status.DeepCopy()
	defer func() {
		var err error
		if !reflect.DeepEqual(*oldStatus, verlJob.Status) {
			hwlog.RunLog.Debugf("VerlJob Update status, from %v to %v", *oldStatus, verlJob.Status)
			if verlJob.Status.ReplicaStatuses == nil {
				verlJob.Status.ReplicaStatuses = map[commonv1.ReplicaType]*commonv1.ReplicaStatus{}
			}
			err = r.UpdateResourceStatus(verlJob)
		}
		if err != nil {
			r.Recorder.Eventf(verlJob, v1.EventTypeWarning, common.StatusUpdateFailedReason, err.Error())
			hwlog.RunLog.Warnf("failed to update VerlJob status in api server, err: %s", err)
		}
	}()

	err := r.ReconcileVerlJob(verlJob)
	if err == nil {
		return ctrl.Result{}, nil
	}
	if k8serr.IsConflict(err) || common.IfNeedRequeue(err) {
		return ctrl.Result{Requeue: true}, nil
	}
	r.Recorder.Eventf(verlJob, v1.EventTypeWarning, util.JobFailedReason, err.Error())
	hwlog.RunLog.Errorf("Reconcile VerlJob<%s> failed, %s", req.NamespacedName, err)
	err = util.UpdateJobConditions(&verlJob.Status, commonv1.JobFailed,
		common.StatusUpdateFailedReason, err.Error())
	if err != nil {
		hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
	}
	return ctrl.Result{}, nil
}

func (r *VerlJobReconciler) ReconcileVerlJob(verlJob *mindxdlv1.VerlJob) error {
	if err := r.ValidateJob(*verlJob); err != nil {
		return fmt.Errorf("VerlJob %s Validation Error: %s", verlJob.Name, err.Error())
	}
	// Set default priorities to VerlJob
	r.Scheme.Default(verlJob)

	ji, err := r.getVerlJobInfo(verlJob)
	if err != nil {
		return err
	}

	if util.IsSucceeded(verlJob.Status) || util.IsFailed(verlJob.Status) {
		err = r.handleFinishedJob(ji, false, common.ConditionInfo{})
		return err
	}

	r.Mutex.RLock()
	version, ok := r.Versions[verlJob.GetUID()]
	backoffLimit, backoffLimitOk := r.BackoffLimits[verlJob.GetUID()]
	r.Mutex.RUnlock()
	if !ok || (backoffLimitOk && backoffLimit > 0 && version > backoffLimit) {
		hwlog.RunLog.Warnf("Job %s has failed because it has reached the specified backoff limit", verlJob.Name)
		err := r.handleFinishedJob(ji, true, common.ConditionInfo{
			CondType: commonv1.JobFailed,
			Reason:   util.JobFailedReason,
			Message: fmt.Sprintf("Job %s has failed because it has reached the specified backoff limit",
				verlJob.Name),
		})
		return err
	}

	isReady, err := r.ReconcileRayCluster(ji)
	if err != nil {
		return err
	}
	if !isReady {
		hwlog.RunLog.Info("RayCluster is not ready")
		return nil
	}

	return r.ReconcileSubmitPod(ji)
}

func (r *VerlJobReconciler) getVerlJobInfo(verlJob *mindxdlv1.VerlJob) (*VerlJobInfo, error) {
	rayClusterList := &mindxdlv1.RayClusterList{}
	if err := r.FetchChildrenForResource(verlJob, rayClusterList, common.VerlJobIdentification); err != nil {
		hwlog.RunLog.Errorf("list rayClusterList of resource<%s/%s> error: %v", verlJob.GetNamespace(),
			verlJob.GetName(), err)
		return nil, err
	}

	podList := &v1.PodList{}
	if err := r.FetchChildrenForResource(verlJob, podList, common.VerlJobIdentification); err != nil {
		hwlog.RunLog.Errorf("list PodList of resource<%s/%s> error: %v", verlJob.GetNamespace(),
			verlJob.GetName(), err)
		return nil, err
	}

	return &VerlJobInfo{
		verlJob:        verlJob,
		rayClusterList: common.ConvertToPointerSlice(rayClusterList.Items),
		podList:        common.ConvertToPointerSlice(podList.Items),
	}, nil
}

func (r *VerlJobReconciler) ReconcileRayCluster(ji *VerlJobInfo) (bool, error) {
	rayClusters := ji.rayClusterList
	verlJob := ji.verlJob
	if common.IsSliceEmpty(rayClusters) {
		hwlog.RunLog.Infof("RayCluster not exists, create it")
		createRayCluster := newRayCluster(verlJob)
		if err := r.CreateChildForResource(verlJob, createRayCluster, common.VerlJobIdentification); err != nil {
			hwlog.RunLog.Errorf("failed to create resource %s<%s>, err: %s",
				reflect.TypeOf(createRayCluster).Name(), createRayCluster.GetName(), err)
			return false, err
		}
		return false, nil
	}

	if len(rayClusters) > 1 {
		hwlog.RunLog.Warn("more than 1 RayCluster exists, maybe there is something wrong")
	}

	rayCluster := rayClusters[0]
	if !common.IsRunning(rayCluster.Status) {
		hwlog.RunLog.Debugf("expect RayCluster<%s> is running, but %s",
			rayCluster.Name, common.GetStatusString(rayCluster.Status))
		common.SyncStatus(&verlJob.Status, &rayCluster.Status)
		return false, nil
	}
	return true, nil
}

func newRayCluster(verlJob *mindxdlv1.VerlJob) *mindxdlv1.RayCluster {
	verlJobCopy := verlJob.Spec.DeepCopy()
	return &mindxdlv1.RayCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.RayClusterResourceKind,
			APIVersion: mindxdlv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        verlJob.Name,
			Namespace:   verlJob.GetNamespace(),
			Labels:      common.DeepCopyMap(verlJob.Labels),
			Annotations: common.DeepCopyMap(verlJob.Annotations),
		},
		Spec: mindxdlv1.RayClusterSpec{
			ReplicaSpecs:  verlJobCopy.ReplicaSpecs,
			SchedulerName: verlJob.Spec.SchedulerName,
			RunPolicy:     verlJob.Spec.RunPolicy,
		},
	}
}

func (r *VerlJobReconciler) ReconcileSubmitPod(ji *VerlJobInfo) error {
	pods := ji.podList
	verlJob := ji.verlJob
	if common.IsSliceEmpty(pods) {
		hwlog.RunLog.Infof("VerlJob not exists, create it")
		podTemplate, err := r.createPodSpec(verlJob)
		if err != nil {
			return err
		}
		err = r.PodControl.CreatePodsWithControllerRef(verlJob.GetNamespace(), podTemplate,
			verlJob, r.GenOwnerReference(verlJob))
		if err != nil {
			hwlog.RunLog.Errorf("Failed create pod of VerlJob %s/%s",
				verlJob.GetNamespace(), verlJob.GetName())
			return err
		}
		return nil
	}

	if len(pods) > 1 {
		hwlog.RunLog.Infof("more than 1 verl submit pod exists, maybe there is something wrong")
	}

	return r.HandleSubmitPodByStatus(pods[0], verlJob)
}

func (r *VerlJobReconciler) HandleSubmitPodByStatus(submitPod *v1.Pod, verlJob *mindxdlv1.VerlJob) error {
	if submitPod.GetDeletionTimestamp() != nil {
		hwlog.RunLog.Warnf("Pod %s/%s is being deleted", submitPod.Namespace, submitPod.Name)
		return &common.ReQueueError{
			Message: fmt.Sprintf("Pod %s/%s is Terminating, try to requeue",
				submitPod.Namespace, submitPod.Name),
		}
	}
	if submitPod.Status.Phase == v1.PodSucceeded {
		hwlog.RunLog.Infof("pod<%s> of VerlJob succeed", verlJob.Name)
		return util.UpdateJobConditions(&verlJob.Status, commonv1.JobSucceeded, util.JobSucceededReason,
			"verl task complete running")
	}
	if submitPod.Status.Phase == v1.PodFailed {
		hwlog.RunLog.Infof("pod<%s> of VerlJob failed, restarting...", verlJob.Name)
		if r.IsReachBackoffLimit(verlJob) {
			return fmt.Errorf("VerlJob pod<%s> reach backoff limit, it will not restart again", verlJob.Name)
		}
		err := r.PodControl.DeletePod(submitPod.Namespace, submitPod.Name, verlJob)
		if err != nil {
			hwlog.RunLog.Warnf("error delete Pod<%s> of VerlJob when restarting: %v", verlJob.Name, err)
			return err
		}
		if err = util.UpdateJobConditions(&verlJob.Status, commonv1.JobRestarting,
			util.JobRestartingReason, "verl task Restarting"); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
		// if ray cluster down before, VerlJob's status will not be changed,
		// therefore, return ReQueueError to make sure that pod will be rescheduled in the future
		return &common.ReQueueError{
			Message: fmt.Sprintf("Pod %s/%s will be deleted, try to requeue",
				submitPod.Namespace, submitPod.Name),
		}
	}
	if submitPod.Status.Phase == v1.PodRunning {
		if err := util.UpdateJobConditions(&verlJob.Status, commonv1.JobRunning,
			util.JobRunningReason, ""); err != nil {
			hwlog.RunLog.Errorf("failed to update job status conditions: %v", err)
			return err
		}
		return nil
	}

	hwlog.RunLog.Debugf("expect pod<%s> of VerlJob is running, but %s", verlJob.Name, submitPod.Status.Phase)
	return nil
}

func (r *VerlJobReconciler) handleFinishedJob(ji *VerlJobInfo, needUpdateCond bool,
	cond common.ConditionInfo) error {
	// If the Job is succeed or failed, delete all pods and services.
	verlJob := ji.verlJob
	if err := r.clearSubResources(ji); err != nil {
		hwlog.RunLog.Errorf("job<%s> delete subresources failed, err: %s", verlJob.Name, err)
		return err
	}

	if !needUpdateCond {
		return nil
	}

	r.Recorder.Event(verlJob, v1.EventTypeNormal, cond.Reason, cond.Message)
	if err := util.UpdateJobConditions(&verlJob.Status, cond.CondType, cond.Reason, cond.Message); err != nil {
		hwlog.RunLog.Errorf("Append job condition error: %v", err)
		return err
	}
	return nil
}

func (r *VerlJobReconciler) clearSubResources(ji *VerlJobInfo) error {
	verlJob := ji.verlJob
	runPolicy := verlJob.Spec.RunPolicy
	// Delete nothing when the cleanPodPolicy is None.
	if *runPolicy.CleanPodPolicy == commonv1.CleanPodPolicyNone {
		return nil
	}

	pods := ji.podList
	for _, pod := range pods {
		// Note that pending pod will turn into running once schedulable,
		// not cleaning it may leave orphan running pod in the future,
		// we should treat it equivalent to running phase here.
		if *runPolicy.CleanPodPolicy == commonv1.CleanPodPolicyRunning &&
			pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodPending {
			continue
		}
		if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, verlJob); err != nil {
			return err
		}
	}

	for _, rayCluster := range ji.rayClusterList {
		if err := r.DeleteJob(rayCluster); err != nil {
			return err
		}
	}
	return nil
}

func (r *VerlJobReconciler) createPodSpec(verlJob *mindxdlv1.VerlJob) (*v1.PodTemplateSpec, error) {
	podTemplate := verlJob.Spec.Template.DeepCopy()
	podTemplate.Name = flowcommon.GenGeneralName(verlJob.Name, common.VerlReplicaType, common.DefaultPodIndex)

	if podTemplate.Spec.RestartPolicy == "" {
		podTemplate.Spec.RestartPolicy = v1.RestartPolicyNever
	}

	podTemplate.Labels = r.GenLabels(verlJob, common.VerlJobIdentification)
	podTemplate.Labels[commonv1.ReplicaTypeLabel] = common.VerlReplicaType
	r.SetCommonPodLabel(podTemplate, verlJob)
	setSubmitScript(podTemplate, verlJob)

	// if gang-scheduling is enabled:
	// 1. if user has specified other scheduler, we report a warning without overriding any fields.
	// 2. if no SchedulerName is set for pods, then we set the SchedulerName to "volcano".
	if r.EnableGangScheduling {
		r.setGangScheduleInfo(verlJob, podTemplate)
	}
	return podTemplate, nil
}

func (r *VerlJobReconciler) setGangScheduleInfo(job *mindxdlv1.VerlJob, podTemplate *v1.PodTemplateSpec) {
	jobSchedulerName := job.Spec.SchedulerName
	if len(jobSchedulerName) == 0 || strings.Compare(jobSchedulerName, common.GangSchedulerName) == 0 {
		jobSchedulerName = common.GangSchedulerName
	} else {
		errMsg := "Another job scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, v1.EventTypeWarning, common.JobSchedulerNameReason, errMsg)
	}
	podSchedulerName := podTemplate.Spec.SchedulerName
	if len(podSchedulerName) == 0 {
		podTemplate.Spec.SchedulerName = jobSchedulerName
	} else if strings.Compare(podSchedulerName, common.GangSchedulerName) != 0 {
		errMsg := "Another scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, v1.EventTypeWarning, common.PodTemplateSchedulerNameReason, errMsg)
	}
	if podTemplate.Annotations == nil {
		podTemplate.Annotations = map[string]string{}
	}
	podTemplate.Annotations[common.GangSchedulingPodGroupAnnotation] =
		common.GetPodGroupName(job.Namespace, job.Name, common.RayHeadType)
	podTemplate.Annotations[common.VolcanoTaskSpecKey] = common.RayHeadType
}

func setSubmitScript(template *v1.PodTemplateSpec, verlJob *mindxdlv1.VerlJob) {
	submitContainerIndex := -1
	for index, container := range template.Spec.Containers {
		if container.Name == common.DefaultContainerName {
			submitContainerIndex = index
		}
	}
	if submitContainerIndex == -1 {
		hwlog.RunLog.Warnf("must have container %s to submit verl task, but not found",
			common.DefaultContainerName)
		return
	}

	submitContainer := &template.Spec.Containers[submitContainerIndex]
	verlConfig := genVerlConfig(verlJob)
	serviceHost := common.GetServiceName(verlJob.Namespace, verlJob.Name)
	clientPort := verlJob.Labels[common.RayClientPortLabelKey]
	common.AddEnvValue(submitContainer, common.RayAddrEnvKey,
		fmt.Sprintf("ray://%s:%s", serviceHost, clientPort))
	submitCommand := ""
	if verlConfig.autoSubmit {
		hwlog.RunLog.Info("autoSubmit is true, verl script will be auto generated")
		submitCommand = genSubmitScript(verlConfig)
	}
	if verlConfig.checkRayStatus {
		hwlog.RunLog.Info("checkRayStatus is true, ray check script will be auto added into initContainer")
		replicas := common.GetTotalReplicas(verlJob.Spec.ReplicaSpecs)
		setInitContainer(template, submitContainer, replicas)
	}
	setContainerCommand(submitContainer, submitCommand)
}

func setContainerCommand(submitContainer *v1.Container, command string) {
	var newArgs string
	if !common.IsOneStringCommandMode(submitContainer) {
		newArgs = strings.Join(submitContainer.Command, " ") + strings.Join(submitContainer.Args, " ")
		submitContainer.Command = []string{"/bin/bash", "-c"}
	} else if len(submitContainer.Args) > 0 {
		newArgs = submitContainer.Args[0]
	}
	newArgs += "\n"

	submitContainer.Args = []string{newArgs + command}
}

func setInitContainer(template *v1.PodTemplateSpec, submitContainer *v1.Container, replicas int32) {
	initContainers := template.Spec.InitContainers
	syncCommand := "EXPECTED_NODES=" + strconv.Itoa(int(replicas)) + "; " + checkRayCommand
	RayCheckContainer := v1.Container{
		Name:    common.DefaultRayCheckContainerName,
		Image:   submitContainer.Image,
		Env:     submitContainer.Env,
		Command: []string{"/bin/bash", "-c"},
		Args:    []string{syncCommand},
	}
	template.Spec.InitContainers = append(initContainers, RayCheckContainer)
}

func genVerlConfig(verlJob *mindxdlv1.VerlJob) VerlConfig {
	annos := verlJob.GetAnnotations()
	return VerlConfig{
		autoSubmit:     annos[common.AutoSubmitLabelKey] == "true",
		verlPath:       annos[common.VerlPathLabelKey],
		verlExec:       annos[common.VerlExecLabelKey],
		verlConfig:     annos[common.VerlConfigLabelKey],
		extraConfig:    common.DeepCopyMap(verlJob.Spec.ExtraConfig),
		checkRayStatus: annos[common.CheckRayStatusLabelKey] == "true",
	}
}

func genSubmitScript(config VerlConfig) string {
	var command string
	if config.verlPath != "" {
		command += fmt.Sprintf("cd %s;\n", config.verlPath)
	}
	configDir := filepath.Dir(config.verlConfig)
	configFile := filepath.Base(config.verlConfig)
	command += fmt.Sprintf("ray job submit --working-dir %s -- python3 -m %s "+
		"\\\n--config-path=%s \\\n--config-name=%s",
		config.verlPath, config.verlExec, configDir, configFile)
	for k, v := range config.extraConfig {
		command += fmt.Sprintf(" \\\n%s=%s", k, v)
	}
	command += "exit $?"
	return command
}
