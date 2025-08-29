/*
Copyright(C) 2023-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/core"
	commonutil "github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/common/pkg/util/labels"
	"github.com/kubeflow/common/pkg/util/train"
	"github.com/kubeflow/training-operator/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/generator"
	"ascend-operator/pkg/ranktable/utils"
	mindxdlutils "ascend-operator/pkg/utils"
)

// ReconcilePods checks and updates pods for each given ReplicaSpec.
// It will requeue the ascendJob in case of an error while creating/deleting pods.
func (r *ASJobReconciler) ReconcilePods(
	job interface{},
	jobStatus *commonv1.JobStatus,
	pods []*corev1.Pod,
	rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
) error {
	if r == nil {
		return errors.New("nil pointer")
	}
	hwlog.RunLog.Debugf("reconcile type<%s> pods start", rtype)

	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendJob)
	}

	// Convert ReplicaType to lower string.
	rt := strings.ToLower(string(rtype))
	// Get all pods for the type rt.
	filterPods := filterPodsByReplicaType(pods, rt)
	hwlog.RunLog.Debugf("filter type<%s> pods: %d", rtype, len(filterPods))

	initializeReplicaStatuses(jobStatus, rtype)

	frame, err := mindxdlv1.GetJobFramework(ascendJob)
	if err != nil {
		return err
	}

	pi, err := r.newPodInfo(ascendJob, rtype, spec, frame)
	if err != nil {
		return err
	}

	return r.reconcilePods(pi, filterPods, jobStatus, replicas)
}

func (r *ASJobReconciler) newPodInfo(job *mindxdlv1.AscendJob, rtype commonv1.ReplicaType, spec *commonv1.ReplicaSpec,
	frame string) (*podInfo, error) {
	svcIp, svcPort, err := r.getMngSvcIpAndPort(job, frame, rtype)
	if err != nil {
		return nil, err
	}

	npuName, ctReq := getNpuReqInfoPerPod(job)

	npuReplicas := getTotalNpuReplicas(job)

	return &podInfo{
		isDynamicCutJob: npuName == npuCoreName,
		frame:           frame,
		job:             job,
		spec:            spec,
		ip:              svcIp,
		port:            svcPort,
		ctReq:           ctReq,
		npuReplicas:     npuReplicas,
		rtype:           rtype,
		clusterdSvcIp:   r.getClusterDSvcIp(),
	}, nil
}

func (r *ASJobReconciler) reconcilePods(pi *podInfo, pods []*corev1.Pod, jobStatus *commonv1.JobStatus,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	// GetPodSlices will return enough information here to make decision to add/remove/update resources.
	//
	// For example, let's assume we have pods with replica-index 0, 1, 2
	// If replica is 4, return a slice with size 4. [[0],[1],[2],[]], a pod with replica-index 3 will be created.
	//
	// If replica is 1, return a slice with size 3. [[0],[1],[2]], pod with replica-index 1 and 2 are out of range
	// and will be deleted.
	podSlices := r.GetPodSlices(pods, int(*pi.spec.Replicas))
	var podToCreate []*podInfo
	for index, podSlice := range podSlices {
		if len(podSlice) > 1 {
			hwlog.RunLog.Warnf("We have too many pods for %s %d", pi.rtype, index)
		} else if len(podSlice) == 0 {
			hwlog.RunLog.Debugf("Need to create new pod: %s-%d", pi.rtype, index)
			p := pi.DeepCopy()
			p.index = index
			podToCreate = append(podToCreate, p)
		} else {
			hwlog.RunLog.Debugf("Need to check pod: %s-%d", pi.rtype, index)
			if err := r.checkExistPod(pi, index, podSlice[0], jobStatus); err != nil {
				return err
			}
		}
	}

	return r.createPods(podToCreate, replicas)
}

func (r *ASJobReconciler) updateRandIndex(allocatedPods []*corev1.Pod) {
	for _, p := range allocatedPods {
		if _, rankExist := p.Annotations[api.PodRankIndexAnno]; rankExist {
			hwlog.RunLog.Info("rank index exist")
			return
		}
	}
	var rankIndex uint64 = 0
	for _, p := range allocatedPods {
		if p.Annotations == nil {
			p.Annotations = make(map[string]string, 1)
		}
		p.Annotations[api.PodRankIndexAnno] = strconv.FormatUint(rankIndex, decimal)
		err := r.Update(context.TODO(), p)
		if err != nil {
			hwlog.RunLog.Warnf("update pod %s/%s rank index %d error: %s", p.Namespace, p.Name, rankIndex, err)
		} else {
			hwlog.RunLog.Infof("update pod %s/%s rank index %d success", p.Namespace, p.Name, rankIndex)
		}
		rankIndex++
	}
	hwlog.RunLog.Info("write rank index success")
}

func (r *ASJobReconciler) genRankTable(ji *jobInfo) {
	hwlog.RunLog.Infof("generating rank table for job %s", ji.name)
	rtg, ok := r.rtGenerators[ji.mtObj.GetUID()]
	if !ok {
		hwlog.RunLog.Warnf("rank table generator not found for job %s", ji.name)
		return
	}
	if rtg.GetStatus() == utils.CompletedRTStatus {
		hwlog.RunLog.Debugf("rank table already generated for job %s", ji.name)
		r.checkPodDelete(rtg, ji)
		return
	}

	var allocatedPods []*corev1.Pod
	for _, p := range ji.pods {
		if utils.PodHasAllocated(p) {
			allocatedPods = append(allocatedPods, p)
		}
	}
	hwlog.RunLog.Infof("allocatedPods: %d, total replicas: %d, total pods: %d",
		len(allocatedPods), ji.totalReplicas, len(ji.pods))
	if int(ji.totalReplicas) == 0 || len(allocatedPods) != int(ji.totalReplicas) {
		return
	}
	r.updateRandIndex(allocatedPods)
	rtg.DeletePod()
	errs := &sync.Map{}
	errCount := int32(0)
	wg := &sync.WaitGroup{}
	for _, pod := range allocatedPods {
		wg.Add(1)
		go func(p *corev1.Pod) {
			defer wg.Done()
			if err := rtg.AddPod(p); err != nil {
				errs.Store(p.Name, err)
				atomic.AddInt32(&errCount, 1)
			}
		}(pod)
	}
	wg.Wait()

	if errCount > 0 {
		hwlog.RunLog.Errorf("failed to cache %d pods, err: %v", errCount, errs)
		return
	}
	rtg.SetStatus(utils.CompletedRTStatus)
	rtg.GatherServerList()
	r.saveRankTable(rtg, ji.mtObj.GetName(), ji.mtObj.GetNamespace(), ji.mtObj.GetUID())
}

func (r *ASJobReconciler) checkPodDelete(rtg generator.RankTableGenerator, ji *jobInfo) {
	if len(ji.pods) >= int(ji.totalReplicas) {
		return
	}
	r.deletePodForCmFile(rtg, ji.mtObj.GetUID(), ji.mtObj.GetName(), ji.mtObj.GetNamespace())
}

func (r *ASJobReconciler) deletePodForCmFile(rtg generator.RankTableGenerator,
	uid types.UID, jobName, namespace string) {
	rtg.DeletePod()
	r.saveRankTable(rtg, jobName, namespace, uid)
}

func (r *ASJobReconciler) saveRankTable(rtg generator.RankTableGenerator,
	jobName, namespace string, uid types.UID) {
	r.saveRankTableFile(rtg)
	r.saveRankTableConfigmap(rtg, jobName, namespace, uid)
	rtg.SetTimeStamp(uint64(time.Now().Unix()))
}

func (r *ASJobReconciler) saveRankTableFile(rtg generator.RankTableGenerator) {
	rtg.Lock()
	defer rtg.Unlock()
	curStatus := rtg.GetStatus()
	if rtg.GetFileStatus() == curStatus {
		return
	}
	if err := rtg.WriteToFile(); err != nil {
		hwlog.RunLog.Errorf("failed to write rank table to file, err: %v", err)
		rtg.SetFileStatus(utils.UnknownStatus)
		return
	}
	rtg.SetFileStatus(curStatus)
}

func (r *ASJobReconciler) saveRankTableConfigmap(rtg generator.RankTableGenerator,
	jobName, namespace string, uid types.UID) {
	rtg.Lock()
	defer rtg.Unlock()
	if !r.configmapExist(rtg, jobName, namespace) {
		return
	}
	curStatus := rtg.GetStatus()
	if rtg.GetConfigmapStatus() == curStatus {
		return
	}
	if err := r.tryWriteCm(jobName, namespace, uid); err != nil {
		hwlog.RunLog.Errorf("failed to write rank table to configmap, err: %v", err)
		rtg.SetConfigmapStatus(utils.UnknownStatus)
		return
	}
	rtg.SetConfigmapStatus(curStatus)
}

func (r *ASJobReconciler) configmapExist(rtg generator.RankTableGenerator, jobName, namespace string) bool {
	configmapExist := rtg.GetConfigmapExist()
	if configmapExist == utils.ConfigmapExsit {
		return true
	}
	if configmapExist == utils.ConfigmapNotExist {
		return false
	}
	configmapName := configmapPrefix + jobName
	if _, err := r.getConfigmapFromApiserver(namespace, configmapName); err != nil {
		rtg.SetConfigmapExist(utils.ConfigmapNotExist)
		return false
	}
	rtg.SetConfigmapExist(utils.ConfigmapExsit)
	return true
}

func (r *ASJobReconciler) tryWriteCm(jobName, namespace string, uid types.UID) error {
	hwlog.RunLog.Infof("start write rank table to configmap<%s> in namespace<%s>", configmapPrefix+jobName, namespace)
	var err error
	for i := 0; i < cmRetryTime; i++ {
		if err = r.writeRanktableToCm(jobName, namespace, uid); err == nil {
			hwlog.RunLog.Info("write rank table to configmap success")
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

func (r *ASJobReconciler) checkExistPod(pi *podInfo, index int, pod *corev1.Pod, jobStatus *commonv1.JobStatus) error {
	// check if the index is in the valid range, if not, we should kill the pod
	if index < 0 || index >= int(*pi.spec.Replicas) {
		if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, pi.job); err != nil {
			return err
		}
	}
	if err := r.checkPodStatus(pi, pod, jobStatus); err != nil {
		return err
	}
	updateJobReplicaStatuses(jobStatus, pi.rtype, pod)
	return nil
}

func (r *ASJobReconciler) createPods(pods []*podInfo, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	if len(pods) == 0 {
		return nil
	}
	job := pods[0].job
	if !r.scaler.CanCreatePod(job) {
		return fmt.Errorf("job %s/%s can't create pod, try later", job.Namespace, job.Name)
	}
	if r.batchMgr.tryBatchCreate() {
		if r.batchCreatePods(pods, replicas, job) == nil {
			return nil
		}
		r.batchMgr.updateUnavailableStatus()
	}
	hwlog.RunLog.Info("the batch creation interface from k8s api-server is unavailable, start create pod singly")

	appendMutex := sync.RWMutex{}
	var createErr []error
	appendErr := func(err error) {
		appendMutex.Lock()
		defer appendMutex.Unlock()
		createErr = append(createErr, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pods))
	start := time.Now()
	for _, pInfo := range pods {
		go func(p *podInfo) {
			defer wg.Done()
			if err := r.createNewPod(p, replicas); err != nil {
				appendErr(err)
			}
		}(pInfo)
	}
	wg.Wait()
	hwlog.RunLog.Infof("create job all pods use time (%v)", time.Since(start))
	if len(createErr) > 0 {
		return fmt.Errorf("failed to create pods: %v", createErr)
	}
	return nil
}

// getPodsSlice split pods to default batches
func (r *ASJobReconciler) getPodsSlice(pods []*podInfo,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) ([]corev1.PodList, error) {
	batches := len(pods) / batchCreatePodsDefaultSize
	if len(pods)%batchCreatePodsDefaultSize != 0 {
		batches += 1
	}
	var podCollections = make([]corev1.PodList, batches)
	var podIdx = 0
	for locate := 0; locate < batches; locate++ {
		podCollections[locate].TypeMeta = metav1.TypeMeta{
			Kind:       "PodLists",
			APIVersion: "v1",
		}
		for ; podIdx < batchCreatePodsDefaultSize*(locate+1) && podIdx < len(pods); podIdx++ {
			template, err := r.createPodSpec(pods[podIdx], replicas)
			if err != nil {
				return nil, err
			}
			controllerRef := r.GenOwnerReference(pods[podIdx].job)
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      template.Labels,
					Annotations: template.Annotations,
					Name:        template.Name,
					Namespace:   pods[podIdx].job.Namespace,
					Finalizers:  template.Finalizers,
				},
			}
			pod.OwnerReferences = append(pod.OwnerReferences, *controllerRef)
			pod.Spec = template.Spec
			podCollections[locate].Items = append(podCollections[locate].Items, pod)
		}
	}
	return podCollections, nil
}

func (r *ASJobReconciler) batchCreatePods(pods []*podInfo, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	job *mindxdlv1.AscendJob) error {
	if job == nil {
		hwlog.RunLog.Error("batchCreatePods error: job is nil")
		return errors.New("batchCreatePods job error: is nil")
	}
	appendMutex := sync.RWMutex{}
	var createErr []error
	appendErr := func(err error) {
		appendMutex.Lock()
		defer appendMutex.Unlock()
		createErr = append(createErr, err)
	}
	podCollections, err := r.getPodsSlice(pods, replicas)
	if err != nil {
		hwlog.RunLog.Errorf("batchCreatePods get pod slice error: %v, jobName=%v", err, job.Name)
		return err
	}
	wg := sync.WaitGroup{}

	wg.Add(len(podCollections))
	start := time.Now()
	for locate := range podCollections {
		go func(batchIndex int) {
			defer wg.Done()
			podList := podCollections[batchIndex]
			hwlog.RunLog.Infof("start collectionCreate, batchIndex=%d, batchSize=%d", batchIndex, len(podList.Items))
			request := r.JobController.KubeClientSet.CoreV1().RESTClient().Post().Namespace(job.Namespace).
				Resource("pods").Param(batchCreateParam, "true")
			if err := request.Body(&podList).Do(context.TODO()).Error(); err != nil {
				hwlog.RunLog.Errorf("batchCreatePods collectionCreate error: %v", err)
				appendErr(err)
			}
		}(locate)
	}
	wg.Wait()
	if len(createErr) != 0 {
		return fmt.Errorf("batchCreatePods request error: %v", createErr)
	}
	hwlog.RunLog.Infof("batchCreatePods successfully, total used time: %v", time.Since(start))
	return nil
}

func (r *ASJobReconciler) checkPodStatus(pi *podInfo, pod *corev1.Pod, jobStatus *commonv1.JobStatus) error {
	// Get the exit code of the container.
	var exitCode int32 = 0xbeef // magic number
	for _, status := range pod.Status.ContainerStatuses {
		state := status.State
		if status.Name == r.GetDefaultContainerName() && state.Terminated != nil {
			exitCode = state.Terminated.ExitCode
			hwlog.RunLog.Infof("Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
			r.Recorder.Eventf(pi.job, corev1.EventTypeNormal, exitedWithCodeReason,
				"Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
		}
	}
	// Check if the pod is retryable.
	if pi.spec.RestartPolicy == commonv1.RestartPolicyExitCode {
		if pod.Status.Phase == corev1.PodFailed && train.IsRetryableExitCode(exitCode) {
			hwlog.RunLog.Infof("Need to restart the pod: %v.%v", pod.Namespace, pod.Name)
			if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, pi.job); err != nil {
				return err
			}

			// with common library framework, we have to handle restart status here
			// or we won't know which replica has been restarted in updateJobStatus after reconciling all replicas
			msg := fmt.Sprintf("AscendJob %s is restarting because %s replica(s) failed.",
				pi.job.Name, pi.rtype)
			r.Recorder.Event(pi.job, corev1.EventTypeWarning, jobRestartingReason, msg)
			err := commonutil.UpdateJobConditions(jobStatus, commonv1.JobRestarting, jobRestartingReason, msg)
			if err != nil {
				hwlog.RunLog.Errorf("Append ascendJob<%s> condition error: %v", pi.job.Name, err)
				return err
			}
		}
	}
	return nil
}

func (r *ASJobReconciler) createNewPod(pi *podInfo, replicas map[commonv1.ReplicaType]*commonv1.
	ReplicaSpec) error {
	if r == nil {
		return errors.New("nil pointer")
	}
	job := pi.job
	podTemplate, err := r.createPodSpec(pi, replicas)
	if err != nil {
		return err
	}
	err = r.PodControl.CreatePodsWithControllerRef(job.Namespace, podTemplate, job, r.GenOwnerReference(job))
	if err != nil && k8serr.IsTimeout(err) {
		// Pod is created but its initialization has timed out.
		// If the initialization is successful eventually, the
		// controller will observe the creation via the informer.
		// If the initialization fails, or if the pod keeps
		// uninitialized for a long time, the informer will not
		// receive any update, and the controller will create a new
		// pod when the expectation expires.
		return nil
	} else if err != nil {
		// Decrement the expected number of creates because the informer won't observe this pod
		hwlog.RunLog.Debugf("Failed creation, decrementing expectations for ascendjob %s/%s",
			job.Namespace, job.Name)
		return err
	}
	hwlog.RunLog.Infof("create pod: %s/%s success", job.Namespace, podTemplate.Name)
	return nil
}

func (r *ASJobReconciler) createPodSpec(pi *podInfo,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) (*corev1.PodTemplateSpec, error) {
	podTemplate := pi.spec.Template.DeepCopy()
	job := pi.job
	rtypeStr := strings.ToLower(string(pi.rtype))

	if job.Spec.ReplicaSpecs == nil {
		return nil, fmt.Errorf("job or job specs is nil")
	}

	indexStr := strconv.Itoa(pi.index)
	status := getNonWorkerPodMountChipStatus(pi.job)
	if status && pi.rtype == mindxdlv1.ReplicaTypeWorker {
		if pi.index == math.MaxInt {
			return nil, errors.New("rank is the max int")
		}
		pi.rank = pi.index + 1
	} else {
		pi.rank = pi.index
	}
	// Set name for the template.
	podTemplate.Name = common.GenGeneralName(job.Name, strings.ToLower(string(pi.rtype)), indexStr)

	err := r.setEnv(pi, podTemplate)
	if err != nil {
		return nil, err
	}
	r.setPodLabels(job, podTemplate, pi.rtype, indexStr)

	err = r.setPodAnnotation(job, podTemplate, rtypeStr, indexStr)
	if err != nil {
		return nil, err
	}

	r.setRestartPolicy(job, podTemplate, pi.spec)

	// if gang-scheduling is enabled:
	// 1. if user has specified other scheduler, we report a warning without overriding any fields.
	// 2. if no SchedulerName is set for pods, then we set the SchedulerName to "volcano".
	if r.Config.EnableGangScheduling {
		r.setGangScheduleInfo(job, podTemplate, replicas, rtypeStr)
	}
	return podTemplate, nil
}

func (r *ASJobReconciler) setEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) error {
	if mindxdlutils.IsMindIEEPJob(pi.job) {
		hwlog.RunLog.Debugf("Set mindIEEP AscendJob<%s-%s> env", pi.job.Namespace, pi.job.Name)
		r.setInferEnv(pi, podTemplate)
		return nil
	}
	if pi.frame == mindxdlv1.MindSporeFrameworkName && len(pi.job.Spec.ReplicaSpecs) == 1 &&
		pi.rtype == mindxdlv1.ReplicaTypeWorker {
		return nil
	}
	hwlog.RunLog.Debugf("Set AscendJob<%s-%s> framework<%s> env start", pi.job.Namespace, pi.job.Name, pi.frame)
	r.setCommonEnv(pi, podTemplate)
	if pi.ctReq == 0 {
		return nil
	}

	switch pi.frame {
	case mindxdlv1.MindSporeFrameworkName:
		r.setMindSporeEnv(pi, podTemplate)
	case mindxdlv1.PytorchFrameworkName:
		r.setPytorchEnv(pi, podTemplate)
	case mindxdlv1.TensorflowFrameworkName:
		r.setTensorflowEnv(pi, podTemplate)
	default:
		return fmt.Errorf("frameworke<%s> is not support", pi.frame)
	}
	return nil
}

func (r *ASJobReconciler) setGangScheduleInfo(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rt string) {
	jobSchedulerName := job.Spec.SchedulerName
	if len(jobSchedulerName) == 0 || strings.Compare(jobSchedulerName, gangSchedulerName) == 0 {
		jobSchedulerName = gangSchedulerName
	} else {
		errMsg := "Another job scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, corev1.EventTypeWarning, jobSchedulerNameReason, errMsg)
	}
	podSchedulerName := util.GetSchedulerName(replicas)
	if len(podSchedulerName) == 0 {
		podTemplate.Spec.SchedulerName = jobSchedulerName
	} else if strings.Compare(podSchedulerName, gangSchedulerName) != 0 {
		errMsg := "Another scheduler is specified when gang-scheduling is enabled and it will not be overwritten"
		hwlog.RunLog.Warn(errMsg)
		r.Recorder.Event(job, corev1.EventTypeWarning, podTemplateSchedulerNameReason, errMsg)
	}
	if podTemplate.Annotations == nil {
		podTemplate.Annotations = map[string]string{}
	}
	podTemplate.Annotations[gangSchedulingPodGroupAnnotation] = job.GetName() + "-" + string(job.GetUID())
	podTemplate.Annotations[volcanoTaskSpecKey] = rt
}

// GetPodSlices sorts the list of pods by label
func (r *ASJobReconciler) GetPodSlices(pods []*corev1.Pod, replicas int) [][]*corev1.Pod {
	if r == nil {
		return nil
	}
	sliceLen := core.CalculatePodSliceSize(pods, replicas)
	if replicas == 0 {
		sliceLen = 0
	}
	podSlices := make([][]*corev1.Pod, sliceLen)
	for _, pod := range pods {
		index, err := labels.ReplicaIndex(pod.Labels)
		if err != nil {
			hwlog.RunLog.Warnf("Error obtaining replica index from Pod %s/%s: %v", pod.Namespace, pod.Name, err)
			continue
		}
		if index < 0 || index >= replicas {
			hwlog.RunLog.Warnf("The label index is not expected: %d, pod: %s/%s", index, pod.Namespace, pod.Name)
			continue
		}

		podSlices[index] = append(podSlices[index], pod)
	}
	return podSlices
}

func (r *ASJobReconciler) setRestartPolicy(job *mindxdlv1.AscendJob, podTemplateSpec *corev1.PodTemplateSpec,
	spec *commonv1.ReplicaSpec) {
	// Submit a warning event if the user specifies restart policy for
	// the pod template. We recommend to set it from the replica level.
	if podTemplateSpec.Spec.RestartPolicy != corev1.RestartPolicy("") {
		errMsg := "Restart policy in pod template will be overwritten by restart policy in replica spec"
		hwlog.RunLog.Warnf(errMsg)
		r.Recorder.Event(job, corev1.EventTypeWarning, podTemplateRestartPolicyReason, errMsg)
	}

	// This is necessary since restartPolicyExitCode is not supported in v1.PodTemplateSpec
	if spec.RestartPolicy == commonv1.RestartPolicyExitCode {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicy(spec.RestartPolicy)
	}
}
