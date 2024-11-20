/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"strconv"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

const (
	logEnvPattern = "set pod<%s> env: %v"
	taskIDEnvKey  = "MINDX_TASK_ID"
)

func addEnvValue(pod *corev1.PodTemplateSpec, envKey, envValue string, index int) {
	pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
		Name:  envKey,
		Value: envValue,
	})
}

func (r *ASJobReconciler) setMindSporeEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	msRoleMap := map[commonv1.ReplicaType]string{
		mindxdlv1.MindSporeReplicaTypeScheduler: msSchedulerRole,
		mindxdlv1.ReplicaTypeWorker:             msWorkerRole,
	}
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {
			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}
			if pi.rtype == mindxdlv1.MindSporeReplicaTypeScheduler {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name: msSchedHost,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: statusPodIPDownwardAPI,
						},
					},
				})
			} else {
				addEnvValue(podTemplate, msSchedHost, pi.ip, i)
			}
			if !pi.isDynamicCutJob {
				addEnvValue(podTemplate, msLocalWorker, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, msWorkerNum, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
			}
			addEnvValue(podTemplate, taskIDEnvKey, string(pi.job.UID), i)
			addEnvValue(podTemplate, mindxServerIPEnv, pi.clusterdSvcIp, i)
			addEnvValue(podTemplate, msNodeRank, strconv.Itoa(pi.rank), i)
			addEnvValue(podTemplate, msSchedPort, pi.port, i)
			addEnvValue(podTemplate, msServerNum, "0", i)
			addEnvValue(podTemplate, msRole, msRoleMap[pi.rtype], i)
			addEnvValue(podTemplate, hostNetwork, strconv.FormatBool(pi.spec.Template.Spec.HostNetwork), i)

			addEnvValue(podTemplate, npuPod, strconv.FormatBool(checkNpuPod(pi)), i)
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}

func (r *ASJobReconciler) setPytorchEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {

			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}
			if !pi.isDynamicCutJob {
				addEnvValue(podTemplate, ptLocalWorldSize, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, ptWorldSize, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
				addEnvValue(podTemplate, ptLocalRank, localRankStr(pi.ctReq), i)
			}
			addEnvValue(podTemplate, ptMasterAddr, pi.ip, i)
			addEnvValue(podTemplate, ptMasterPort, pi.port, i)
			addEnvValue(podTemplate, ptRank, strconv.Itoa(pi.rank), i)
			addEnvValue(podTemplate, taskIDEnvKey, string(pi.job.UID), i)
			addEnvValue(podTemplate, mindxServerIPEnv, pi.clusterdSvcIp, i)
			addEnvValue(podTemplate, hostNetwork, strconv.FormatBool(pi.spec.Template.Spec.HostNetwork), i)
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}

func (r *ASJobReconciler) setTensorflowEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == mindxdlv1.DefaultContainerName {
			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}
			if pi.rtype == mindxdlv1.TensorflowReplicaTypeChief {
				podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
					Name: tfChiefIP,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: statusPodIPDownwardAPI,
						},
					},
				})
			} else {
				addEnvValue(podTemplate, tfChiefIP, pi.ip, i)
			}
			if !pi.isDynamicCutJob {
				addEnvValue(podTemplate, tfLocalWorker, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, tfWorkerSize, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
			}
			addEnvValue(podTemplate, tfChiefPort, pi.port, i)
			addEnvValue(podTemplate, tfRank, strconv.Itoa(pi.rank), i)
			addEnvValue(podTemplate, taskIDEnvKey, string(pi.job.UID), i)
			addEnvValue(podTemplate, mindxServerIPEnv, pi.clusterdSvcIp, i)
			addEnvValue(podTemplate, tfChiefDevice, "0", i)
			addEnvValue(podTemplate, hostNetwork, strconv.FormatBool(pi.spec.Template.Spec.HostNetwork), i)
			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, corev1.EnvVar{
				Name: tfWorkerIP,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: statusPodIPDownwardAPI,
					},
				},
			})
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}
