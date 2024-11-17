/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) setPodEnvironment(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	frame, err := mindxdlv1.GetJobFramework(job)
	if err != nil {
		return err
	}

	switch frame {
	case mindxdlv1.MindSporeFrameworkName:
		return r.setMindSporeEnv(job, podTemplate, rtype, index)
	case mindxdlv1.PytorchFrameworkName:
		return r.setPytorchEnv(job, podTemplate, rtype, index)
	case mindxdlv1.TensorflowFrameworkName:
		return r.setTensorflowEnv(job, podTemplate, rtype, index)
	default:
		return fmt.Errorf("frameworke<%s> is not support", frame)
	}
}

func (r *ASJobReconciler) setMindSporeEnv(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	hwlog.RunLog.Debugf("Set AscendJob<%s-%s> framework<%s> env start",
		job.Namespace, job.Name, mindxdlv1.MindSporeFrameworkName)

	if len(job.Spec.ReplicaSpecs) == 1 {
		return nil
	}

	svcIp, svcPort, err := r.getMngSvcIpAndPort(job)
	if err != nil {
		return err
	}

	ctReq := getNpuReqPerPod(job)
	if ctReq == 0 {
		return fmt.Errorf("job<%s/%s> not req npu", job.Namespace, job.Name)
	}

	npuReplicas := getTotalNpuReplicas(job)
	if npuReplicas == 0 {
		return fmt.Errorf("job<%s/%s> npu pod is 0", job.Namespace, job.Name)
	}

	msRoleMap := map[string]string{
		"scheduler": msSchedulerRole,
		"worker":    msWorkerRole,
	}

	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {
			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}

			if rtype == strings.ToLower(string(mindxdlv1.MindSporeReplicaTypeScheduler)) {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name: msSchedHost,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: statusPodIPDownwardAPI,
						},
					},
				})
			} else {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name:  msSchedHost,
					Value: svcIp,
				})
			}

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msNodeRank,
				Value: index,
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msSchedPort,
				Value: svcPort,
			})
			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msServerNum,
				Value: "0",
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msLocalWorker,
				Value: strconv.Itoa(ctReq),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msWorkerNum,
				Value: strconv.Itoa(ctReq * npuReplicas),
			})
			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  msRole,
				Value: msRoleMap[rtype],
			})
			hwlog.RunLog.Infof("set pod<%s> env: %v", podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
	return nil
}

func (r *ASJobReconciler) setPytorchEnv(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	hwlog.RunLog.Debugf("Set AscendJob<%s-%s> framework<%s> env start",
		job.Namespace, job.Name, mindxdlv1.PytorchFrameworkName)

	svcIp, svcPort, err := r.getMngSvcIpAndPort(job)
	if err != nil {
		return err
	}

	ctReq := getNpuReqPerPod(job)
	if ctReq == 0 {
		return fmt.Errorf("job<%s/%s> not req npu", job.Namespace, job.Name)
	}

	npuReplicas := getTotalNpuReplicas(job)
	if npuReplicas == 0 {
		return fmt.Errorf("job<%s/%s> npu pod is 0", job.Namespace, job.Name)
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

	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {

			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptMasterAddr,
				Value: svcIp,
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptMasterPort,
				Value: svcPort,
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptLocalRank,
				Value: localRankStr(ctReq),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptRank,
				Value: strconv.Itoa(rank),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptLocalWorldSize,
				Value: strconv.Itoa(ctReq),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  ptWorldSize,
				Value: strconv.Itoa(ctReq * npuReplicas),
			})
			hwlog.RunLog.Infof("set pod<%s> env: %v", podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
	return nil
}

func (r *ASJobReconciler) setTensorflowEnv(job *mindxdlv1.AscendJob, podTemplate *corev1.PodTemplateSpec, rtype,
	index string) error {
	hwlog.RunLog.Debugf("Set AscendJob<%s-%s> framework<%s> env start",
		job.Namespace, job.Name, mindxdlv1.TensorflowFrameworkName)

	svcIp, svcPort, err := r.getMngSvcIpAndPort(job)
	if err != nil {
		return err
	}

	ctReq := getNpuReqPerPod(job)
	if ctReq == 0 {
		return fmt.Errorf("job<%s/%s> not req npu", job.Namespace, job.Name)
	}

	npuReplicas := getTotalNpuReplicas(job)
	if npuReplicas == 0 {
		return fmt.Errorf("job<%s/%s> npu pod is 0", job.Namespace, job.Name)
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

	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {
			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}

			if rtype == strings.ToLower(string(mindxdlv1.TensorflowReplicaTypeChief)) {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name: tfChiefIP,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: statusPodIPDownwardAPI,
						},
					},
				})
			} else {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name:  tfChiefIP,
					Value: svcIp,
				})
			}

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  tfChiefDevice,
				Value: "0",
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  tfChiefPort,
				Value: svcPort,
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  tfRank,
				Value: strconv.Itoa(rank),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  tfLocalWorker,
				Value: strconv.Itoa(ctReq),
			})

			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name:  tfWorkerSize,
				Value: strconv.Itoa(ctReq * npuReplicas),
			})
			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name: tfWorkerIP,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: statusPodIPDownwardAPI,
					},
				},
			})
			hwlog.RunLog.Infof("set pod<%s> env: %v", podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
	return nil
}
