/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.

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
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

type jobInfo struct {
	job           interface{}
	jobKey        string
	name          string
	rtObj         runtime.Object
	mtObj         metav1.Object
	pods          []*corev1.Pod
	status        *commonv1.JobStatus
	runPolicy     *commonv1.RunPolicy
	rpls          map[commonv1.ReplicaType]*commonv1.ReplicaSpec
	totalReplicas int32
}

func (r *ASJobReconciler) newJobInfo(
	job interface{},
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
	jobStatus *commonv1.JobStatus,
	runPolicy *commonv1.RunPolicy) (*jobInfo, error) {
	metaObject, ok := job.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("job is not of type metav1.Object")
	}

	runtimeObject, ok := job.(runtime.Object)
	if !ok {
		return nil, fmt.Errorf("job is not of type runtime.Object")
	}

	ascendJob, ok := job.(*mindxdlv1.AscendJob)
	if !ok {
		return nil, fmt.Errorf("job is not of type AscendJob")
	}

	jobKey, err := common.KeyFunc(job)
	if err != nil {
		hwlog.RunLog.Errorf("couldn't get key for job<%s/%s> object: %v", metaObject.GetNamespace(),
			metaObject.GetName(), err)
		return nil, err
	}

	pods, err := r.getPodsForJob(job)
	if err != nil {
		hwlog.RunLog.Warnf("GetPodsForJob error %v", err)
		return nil, err
	}

	return &jobInfo{
		job:           job,
		jobKey:        jobKey,
		name:          metaObject.GetName(),
		rtObj:         runtimeObject,
		mtObj:         metaObject,
		pods:          pods,
		status:        jobStatus,
		runPolicy:     runPolicy,
		rpls:          replicas,
		totalReplicas: getTotalReplicas(ascendJob),
	}, nil
}

func genLabels(jobObj interface{}, jobName string) map[string]string {
	acjob, ok := jobObj.(*mindxdlv1.AscendJob)
	if !ok {
		hwlog.RunLog.Errorf("job not found")
		return map[string]string{
			"NoPodShouldAdd": "NoPodShouldAdd",
		}
	}
	switch acjob.Kind {
	case "AscendJob":
		return map[string]string{
			commonv1.JobNameLabel: jobName,
		}
	case "Job":
		return map[string]string{
			"volcano.sh/job-name": jobName,
		}
	default:
		hwlog.RunLog.Errorf("job kind %s is invalid", acjob.Kind)
		return map[string]string{
			"NoPodShouldAdd": "NoPodShouldAdd",
		}
	}
}

func (r *ASJobReconciler) getPodsForJob(jobObject interface{}) ([]*corev1.Pod, error) {
	job, ok := jobObject.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("job is not of type metav1.Object")
	}

	// Create selector.
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: genLabels(jobObject, job.GetName()),
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't convert Job selector: %v", err)
	}
	pods := &corev1.PodList{}
	if err = r.List(context.TODO(), pods, client.MatchingLabelsSelector{Selector: selector},
		client.InNamespace(job.GetNamespace())); err != nil {
		hwlog.RunLog.Errorf("list job<%s/%s> pods error %v", job.GetNamespace(), job.GetName(), err)
		return nil, err
	}

	podSlice := make([]*corev1.Pod, len(pods.Items))
	for i := range pods.Items {
		podSlice[i] = &pods.Items[i]
	}
	return podSlice, nil
}

func (r *ASJobReconciler) getOrCreateSvc(job *mindxdlv1.AscendJob) (*corev1.Service, error) {
	var rtype commonv1.ReplicaType
	var spec *commonv1.ReplicaSpec
	for rt, sp := range job.Spec.ReplicaSpecs {
		if r.IsMasterRole(job.Spec.ReplicaSpecs, rt, 0) {
			rtype = rt
			spec = sp
			break
		}
	}

	name := common.GenGeneralName(job.GetName(), strings.ToLower(string(rtype)), "0")
	svc := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: job.GetNamespace(), Name: name}, svc)
	if err == nil {
		return svc, nil
	}

	if errors.IsNotFound(err) {
		newSvc, gerr := r.genService(job, rtype, spec)
		if gerr != nil {
			return nil, gerr
		}
		_, createErr := r.KubeClientSet.CoreV1().Services(job.GetNamespace()).Create(context.TODO(), newSvc,
			metav1.CreateOptions{})
		if createErr != nil {
			return nil, err
		}
	}
	return nil, err
}
