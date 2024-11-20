/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"errors"
	"math"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) setPodAnnotation(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	return r.setHcclRankIndex(job, podTemplate, rtype, index)
}

func (r *ASJobReconciler) setHcclRankIndex(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	_, existMaster := job.Spec.ReplicaSpecs[mindxdlv1.PytorchReplicaTypeMaster]
	_, existChief := job.Spec.ReplicaSpecs[mindxdlv1.TensorflowReplicaTypeChief]
	if !existMaster && !existChief {
		podTemplate.Annotations[rankIndexKey] = index
		return nil
	}

	rank, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	if rtype == strings.ToLower(string(mindxdlv1.ReplicaTypeWorker)) {
		if rank == math.MaxInt {
			return errors.New("rank is the max int")
		}
		rank = rank + 1
	}

	podTemplate.Annotations[rankIndexKey] = strconv.Itoa(rank)
	return nil
}
