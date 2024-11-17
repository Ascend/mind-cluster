/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/controller.v1/expectation"
	commoncore "github.com/kubeflow/common/pkg/core"
	commonutil "github.com/kubeflow/common/pkg/util"
	utillabels "github.com/kubeflow/common/pkg/util/labels"
	train_util "github.com/kubeflow/common/pkg/util/train"
	"github.com/kubeflow/training-operator/pkg/common/util"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// reconcilePods checks and updates pods for each given MSReplicaSpec.
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
	hwlog.RunLog.Debugf("reconcile pods start")
	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return fmt.Errorf("%v is not a type of AscendJob", ascendJob)
	}

	// Convert ReplicaType to lower string.
	rt := strings.ToLower(string(rtype))
	// Get all pods for the type rt.
	filterPods, err := r.FilterPodsForReplicaType(pods, rt)
	if err != nil {
		hwlog.RunLog.Errorf("filter job<%s> replica-type<%v> pods failed", ascendJob.Name, rtype)
		return err
	}

	numReplicas := int(*spec.Replicas)
	initializeReplicaStatuses(jobStatus, rtype)

	// GetPodSlices will return enough information here to make decision to add/remove/update resources.
	//
	// For example, let's assume we have pods with replica-index 0, 1, 2
	// If replica is 4, return a slice with size 4. [[0],[1],[2],[]], a pod with replica-index 3 will be created.
	//
	// If replica is 1, return a slice with size 3. [[0],[1],[2]], pod with replica-index 1 and 2 are out of range and will be deleted.
	podSlices := r.GetPodSlices(filterPods, numReplicas)
	for index, podSlice := range podSlices {
		if len(podSlice) > 1 {
			hwlog.RunLog.Warnf("We have too many pods for %s %d", rt, index)
		} else if len(podSlice) == 0 {
			hwlog.RunLog.Infof("Need to create new pod: %s-%d", rt, index)

			// check if this replica is the master role
			masterRole := r.Controller.IsMasterRole(replicas, rtype, index)
			err = r.CreateNewPod(ascendJob, rt, strconv.Itoa(index), spec, masterRole, replicas)
			if err != nil {
				return err
			}
		} else {
			// Check the status of the current pod.
			pod := podSlice[0]

			// check if the index is in the valid range, if not, we should kill the pod
			if index < 0 || index >= numReplicas {
				err = r.PodControl.DeletePod(pod.Namespace, pod.Name, ascendJob)
				if err != nil {
					return err
				}
			}
			// Get the exit code of the container.
			var exitCode int32 = 0xbeef // magic number
			for _, status := range pod.Status.ContainerStatuses {
				state := status.State
				if status.Name == r.GetDefaultContainerName() && state.Terminated != nil {
					exitCode = state.Terminated.ExitCode
					hwlog.RunLog.Infof("Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
					r.Recorder.Eventf(ascendJob, corev1.EventTypeNormal, exitedWithCodeReason, "Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
				}
			}
			// Check if the pod is retryable.
			if spec.RestartPolicy == commonv1.RestartPolicyExitCode {
				if pod.Status.Phase == corev1.PodFailed && train_util.IsRetryableExitCode(exitCode) {
					hwlog.RunLog.Infof("Need to restart the pod: %v.%v", pod.Namespace, pod.Name)
					if err := r.PodControl.DeletePod(pod.Namespace, pod.Name, ascendJob); err != nil {
						return err
					}

					// with common library framework, we have to handle restart status here
					// or we won't know which replica has been restarted in updateJobStatus after reconciling all replicas
					msg := fmt.Sprintf("AscendJob %s is restarting because %s replica(s) failed.",
						ascendJob.Name, rtype)
					r.Recorder.Event(ascendJob, corev1.EventTypeWarning, jobRestartingReason, msg)
					err := commonutil.UpdateJobConditions(jobStatus, commonv1.JobRestarting, jobRestartingReason, msg)
					if err != nil {
						hwlog.RunLog.Errorf("Append ascendJob<%s> condition error: %v", ascendJob.Name, err)
						return err
					}
				}
			}

			updateJobReplicaStatuses(jobStatus, rtype, pod)
		}
	}
	return nil
}

func (r *ASJobReconciler) CreateNewPod(job *mindxdlv1.AscendJob, rt, index string, spec *commonv1.ReplicaSpec,
	masterRole bool, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	jobKey, err := common.KeyFunc(job)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for AscendJob object %#v: %v", job, err))
		return err
	}
	expectationPodsKey := expectation.GenExpectationPodsKey(jobKey, rt)
	err = r.Expectations.ExpectCreations(expectationPodsKey, 1)
	if err != nil {
		return err
	}

	// Create OwnerReference.
	controllerRef := r.GenOwnerReference(job)

	// Set type and index for the worker.
	labels := r.GenLabels(job.Name)

	utillabels.SetReplicaType(labels, rt)
	utillabels.SetReplicaIndexStr(labels, index)

	frame, _ := mindxdlv1.GetJobFramework(job)
	if frame == string(mindxdlv1.TensorflowReplicaTypeChief) || frame == mindxdlv1.PytorchFrameworkName {
		if masterRole {
			labels[commonv1.JobRoleLabel] = "master"
		}
	}

	podTemplate := spec.Template.DeepCopy()

	// Set name for the template.
	podTemplate.Name = common.GenGeneralName(job.Name, rt, index)

	if podTemplate.Labels == nil {
		podTemplate.Labels = make(map[string]string)
	}

	if podTemplate.Annotations == nil {
		podTemplate.Annotations = make(map[string]string)
	}

	podTemplate.Labels[podVersionLabel] = strconv.FormatInt(int64(defaultPodVersion), decimal)
	if version, ok := r.versions[job.GetUID()]; ok {
		podTemplate.Labels[podVersionLabel] = strconv.FormatInt(int64(version), decimal)
	}
	for key, value := range labels {
		podTemplate.Labels[key] = value
	}

	if err := r.SetClusterSpec(job, podTemplate, rt, index); err != nil {
		return err
	}

	// Submit a warning event if the user specifies restart policy for
	// the pod template. We recommend to set it from the replica level.
	if podTemplate.Spec.RestartPolicy != corev1.RestartPolicy("") {
		errMsg := "Restart policy in pod template will be overwritten by restart policy in replica spec"
		hwlog.RunLog.Warnf(errMsg)
		r.Recorder.Event(job, corev1.EventTypeWarning, podTemplateRestartPolicyReason, errMsg)
	}
	setRestartPolicy(podTemplate, spec)

	// if gang-scheduling is enabled:
	// 1. if user has specified other scheduler, we report a warning without overriding any fields.
	// 2. if no SchedulerName is set for pods, then we set the SchedulerName to "volcano".
	if r.Config.EnableGangScheduling {
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

	err = r.PodControl.CreatePodsWithControllerRef(job.Namespace, podTemplate, job, controllerRef)
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
		hwlog.RunLog.Errorf(
			"Failed creation, decrementing expectations for ascnedjob %s/%s, key %s",
			job.Namespace, job.Name, expectationPodsKey)
		r.Expectations.CreationObserved(expectationPodsKey)
		return err
	}
	return nil

}

func (r *ASJobReconciler) GetPodSlices(pods []*corev1.Pod, replicas int) [][]*corev1.Pod {
	if r == nil {
		return nil
	}
	podSlices := make([][]*corev1.Pod, commoncore.CalculatePodSliceSize(pods, replicas))
	for _, pod := range pods {
		index, err := utillabels.ReplicaIndex(pod.Labels)
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
