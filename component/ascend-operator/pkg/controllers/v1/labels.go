/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/util/labels"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) setPodLabels(job *mindxdlv1.Job, podTemplate *corev1.PodTemplateSpec,
	rt commonv1.ReplicaType, index string) {
	// Set type and index for the worker.
	labelsMap := r.GenLabels(job.Name)

	rtypeStr := strings.ToLower(string(rt))
	labels.SetReplicaType(labelsMap, rtypeStr)
	labels.SetReplicaIndexStr(labelsMap, index)

	if podTemplate.Labels == nil {
		podTemplate.Labels = make(map[string]string)
	}

	if r.IsMasterRole(job.Spec.ReplicaSpecs, rt, 0) {
		podTemplate.Labels[commonv1.JobRoleLabel] = "master"
	}
	podTemplate.Labels[podVersionLabel] = strconv.FormatInt(int64(defaultPodVersion), decimal)
	if version, ok := r.versions[job.GetUID()]; ok {
		podTemplate.Labels[podVersionLabel] = strconv.FormatInt(int64(version), decimal)
	}
	for key, value := range labelsMap {
		podTemplate.Labels[key] = value
	}
}
