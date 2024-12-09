/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package v1 is using for reconcile AscendJob.
package v1

import (
	"errors"
	"math"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) setPodAnnotation(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	return r.setHcclRankIndex(job, podTemplate, rtype, index)
}

func (r *ASJobReconciler) setHcclRankIndex(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	if podTemplate.Annotations == nil {
		podTemplate.Annotations = make(map[string]string)
	}

	rank, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	status := getNonWorkerPodMountChipStatus(job)
	if !status {
		podTemplate.Annotations[rankIndexKey] = index
		return nil
	}

	if rtype == strings.ToLower(string(mindxdlv1.ReplicaTypeWorker)) {
		if rank == math.MaxInt {
			return errors.New("rank is the max int")
		}
		rank = rank + 1
	}

	podTemplate.Annotations[rankIndexKey] = strconv.Itoa(rank)
	hwlog.RunLog.Debugf("set rank index<%d> to pod<%s>", rank, podTemplate.Name)
	return nil
}
