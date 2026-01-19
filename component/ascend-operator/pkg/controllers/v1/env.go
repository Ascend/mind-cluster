/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"fmt"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/utils"
)

const (
	logEnvPattern = "set pod<%s> env: %v"

	taskIDEnvKey  = "MINDX_TASK_ID"
	appTypeEnvKey = "APP_TYPE"
)

func addEnvValue(pod *corev1.PodTemplateSpec, envKey, envValue string, index int) {
	pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
		Name:  envKey,
		Value: envValue,
	})
}

func addEnvValueForSoftStrategy(pod *corev1.PodTemplateSpec, envKey string, index int) {
	pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
		Name:      envKey,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: utils.SuperPodEnvPath}},
	})
}

// isVirtualResourceReq return true when pod request virtual resource, otherwise return false
func (r *ASJobReconciler) isVirtualResourceReq(requests *corev1.ResourceList) bool {
	if requests == nil {
		return false
	}
	nonVirtualResources := map[corev1.ResourceName]struct{}{
		api.HuaweiAscend310:  {},
		api.HuaweiAscend310P: {},
		api.HuaweiAscend910:  {},
	}
	for name := range *requests {
		if _, ok := nonVirtualResources[name]; !ok {
			hwlog.RunLog.Debugf("virtual resource name detected: %s", name)
			return true
		}
	}
	return false
}

func (r *ASJobReconciler) setInferEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name != api.DefaultContainerName {
			continue
		}
		if len(podTemplate.Spec.Containers[i].Env) == 0 {
			podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
		}
		addEnvValue(podTemplate, taskIDEnvKey, pi.job.Labels[mindxdlv1.JodIdLabelKey], i)
		addEnvValue(podTemplate, appTypeEnvKey, pi.job.Labels[mindxdlv1.AppLabelKey], i)
		addEnvValue(podTemplate, mindxServerIPEnv, pi.clusterdSvcIp, i)
		hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
	}
}

func (r *ASJobReconciler) setCommonEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == api.DefaultContainerName {
			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}
			if !r.isVirtualResourceReq(&podTemplate.Spec.Containers[i].Resources.Requests) {
				r.setAscendVisibleDevicesEnv(&podTemplate.Spec.Containers[i])
			}
			addEnvValue(podTemplate, taskIDEnvKey, string(pi.job.UID), i)
			addEnvValue(podTemplate, mindxServerIPEnv, pi.clusterdSvcIp, i)
			addEnvValue(podTemplate, hostNetwork, strconv.FormatBool(pi.spec.Template.Spec.HostNetwork), i)
			addHcclSuperPodIdEnv(pi, podTemplate, i)
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}

func (r *ASJobReconciler) setAscendVisibleDevicesEnv(container *corev1.Container) {
	for resourceAnnoKey := range container.Resources.Requests {
		if strings.Contains(string(resourceAnnoKey), api.ResourceNamePrefix) {
			container.Env = append(container.Env, corev1.EnvVar{
				Name: api.AscendVisibleDevicesEnv,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: fmt.Sprintf("metadata.annotations['%s']", resourceAnnoKey),
					},
				},
			})
			return
		}
	}
}

func (r *ASJobReconciler) setMindSporeEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	msRoleMap := map[commonv1.ReplicaType]string{
		mindxdlv1.MindSporeReplicaTypeScheduler: msSchedulerRole,
		mindxdlv1.ReplicaTypeWorker:             msWorkerRole,
	}
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == api.DefaultContainerName {
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
				addEnvValue(podTemplate, api.MsLocalWorkerEnv, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, api.MsWorkerNumEnv, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
			}
			addEnvValue(podTemplate, msNodeRank, strconv.Itoa(pi.rank), i)
			addEnvValue(podTemplate, msSchedPort, pi.port, i)
			addEnvValue(podTemplate, msServerNum, "0", i)
			addEnvValue(podTemplate, msRole, msRoleMap[pi.rtype], i)

			addEnvValue(podTemplate, npuPod, strconv.FormatBool(checkNpuPod(pi)), i)
			addProcessRecoverEnv(pi, podTemplate, i, api.MindSporeFramework)
			addMSPodScheduleEnv(pi, podTemplate, i)
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}

func (r *ASJobReconciler) setPytorchEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == api.DefaultContainerName {

			if len(podTemplate.Spec.Containers[i].Env) == 0 {
				podTemplate.Spec.Containers[i].Env = make([]corev1.EnvVar, 0)
			}
			if !pi.isDynamicCutJob {
				addEnvValue(podTemplate, api.PtLocalWorldSizeEnv, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, api.PtWorldSizeEnv, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
				addEnvValue(podTemplate, api.PtLocalRankEnv, localRankStr(pi.ctReq), i)
			}
			addEnvValue(podTemplate, ptMasterAddr, pi.ip, i)
			addEnvValue(podTemplate, ptMasterPort, pi.port, i)
			addEnvValue(podTemplate, ptRank, strconv.Itoa(pi.rank), i)
			addProcessRecoverEnv(pi, podTemplate, i, api.PytorchFramework)
			hwlog.RunLog.Debugf(logEnvPattern, podTemplate.Name, podTemplate.Spec.Containers[i].Env)
		}
	}
}

func (r *ASJobReconciler) setTensorflowEnv(pi *podInfo, podTemplate *corev1.PodTemplateSpec) {
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == api.DefaultContainerName {
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
				addEnvValue(podTemplate, api.TfLocalWorkerEnv, strconv.Itoa(pi.ctReq), i)
				addEnvValue(podTemplate, api.TfWorkerSizeEnv, strconv.Itoa(pi.ctReq*pi.npuReplicas), i)
			}
			addEnvValue(podTemplate, tfChiefPort, pi.port, i)
			addEnvValue(podTemplate, tfRank, strconv.Itoa(pi.rank), i)
			addEnvValue(podTemplate, tfChiefDevice, "0", i)
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

// addHcclSuperPodIdEnv add HCCL_LOGIC_SUPERPOD_ID env to build hccs network
func addHcclSuperPodIdEnv(pi *podInfo, pod *corev1.PodTemplateSpec, index int) {
	for name, res := range pod.Spec.Containers[index].Resources.Requests {
		if strings.Contains(string(name), api.ResourceNamePrefix) {
			chipsPerNode := int(res.Value())
			superPodId := strconv.Itoa(utils.GetLogicSuperPodId(pi.rank, utils.GetSpBlock(pi.job), chipsPerNode))
			hwlog.RunLog.Debugf("pod<%s> resource<%v=%v> pod-rank=%v sp-block=%v set %s=%v",
				pod.Name, name, chipsPerNode, pi.rank, utils.GetSpBlock(pi.job), hcclSuperPodLogicId, superPodId)
			if pi.job.Labels[utils.SuperPodAffinity] == utils.SoftStrategy {
				addEnvValueForSoftStrategy(pod, hcclSuperPodLogicId, index)
			} else {
				addEnvValue(pod, hcclSuperPodLogicId, superPodId, index)
			}
			break
		}
	}
}

func addMSPodScheduleEnv(pi *podInfo, pod *corev1.PodTemplateSpec, containerIndex int) {
	if !isPodScheduleStrategy(pi.job) {
		return
	}
	addEnvValue(pod, api.MsRecoverEnv, `'{`+api.MsRscStrategy+`}'`, containerIndex)
}

func addProcessRecoverEnv(pi *podInfo, pod *corev1.PodTemplateSpec, containerIndex int, framework string) {
	strategies := getJobRecoverStrategy(pi.job)
	if strategies == "" {
		return
	}
	env := make(map[string]string)
	trainEnv := make(sets.String)
	for _, strategy := range strings.Split(strategies, ",") {
		addEnvByStrategy(env, trainEnv, strategy, framework)
	}
	if framework == api.PytorchFramework {
		env[api.HighAvailableEnv] = strings.Join(trainEnv.List(), ",")
	}
	if framework == api.MindSporeFramework {
		if isPodScheduleStrategy(pi.job) {
			trainEnv.Insert(api.MsRscStrategy)
		}
		env[api.MsRecoverEnv] = `'{` + strings.Join(trainEnv.List(), ",") + `}'`
		env[api.EnableMS] = api.EnableFlag
	}
	for k, v := range env {
		addEnvValue(pod, k, v, containerIndex)
	}
	hwlog.RunLog.Infof("set process reschedule pod<%s> env: %v", pod.Name, env)
}

func addEnvByStrategy(env map[string]string, trainEnv sets.String, strategy string, framework string) {
	switch strategy {
	case api.RecoverStrategy:
		addRecoverEnv(env, trainEnv, framework)
	case api.RetryStrategy:
		addRetryEnv(env, trainEnv, framework)
	case api.DumpStrategy:
		addDumpEnv(env, trainEnv, framework)
	case api.InPlaceStrategy:
		addRecoverInPlaceEnv(env, trainEnv, framework)
	case api.ExitStrategy:
		addRecoverEnv(env, trainEnv, framework)
	case api.ElasticTraining:
		addElasticTrainingEnv(env, trainEnv, framework)
	default:
		return
	}
}

func addRecoverEnv(env map[string]string, trainEnv sets.String, framework string) {
	if env == nil || trainEnv == nil {
		return
	}
	if framework == api.PytorchFramework {
		trainEnv.Insert(api.RecoverStrategy)
	}
	if framework == api.MindSporeFramework {
		trainEnv.Insert(api.MsArfStrategy)
	}
	env[api.ProcessRecoverEnv] = api.EnableFunc
	env[api.ElasticRecoverEnv] = api.EnableFlag
}

func addRetryEnv(env map[string]string, trainEnv sets.String, framework string) {
	if env == nil || trainEnv == nil {
		return
	}
	if framework == api.PytorchFramework {
		trainEnv.Insert(api.RetryStrategy)
	}
	if framework == api.MindSporeFramework {
		trainEnv.Insert(api.MsUceStrategy)
		trainEnv.Insert(api.MsHcceStrategy)
	}
	env[api.ProcessRecoverEnv] = api.EnableFunc
	env[api.ElasticRecoverEnv] = api.EnableFlag
}

func addRecoverInPlaceEnv(env map[string]string, trainEnv sets.String, framework string) {
	if env == nil || trainEnv == nil {
		return
	}
	if framework == api.PytorchFramework {
		trainEnv.Insert(api.RecoverStrategy)
	}
	if framework == api.MindSporeFramework {
		trainEnv.Insert(api.MsArfStrategy)
	}
	env[api.ProcessRecoverEnv] = api.EnableFunc
	env[api.EnableRestartEnv] = api.EnableFunc
	env[api.ElasticRecoverEnv] = api.EnableFlag
}

func addDumpEnv(env map[string]string, trainEnv sets.String, framework string) {
	if env == nil || trainEnv == nil {
		return
	}
	if framework == api.PytorchFramework {
		trainEnv.Insert(api.DumpStrategy)
	}
	if framework == api.MindSporeFramework {
		trainEnv.Insert(api.MsDumpStrategy)
	}
	env[api.ElasticRecoverEnv] = api.EnableFlag
	env[api.ProcessRecoverEnv] = api.EnableFunc
}

func addElasticTrainingEnv(env map[string]string, trainEnv sets.String, framework string) {
	if env == nil || trainEnv == nil {
		hwlog.RunLog.Warnf("env or trainEnv is nil, env: %v, trainEnv: %v", env, trainEnv)
		return
	}
	if framework != api.PytorchFramework {
		hwlog.RunLog.Warn("elastic-training strategy only support pytorch framework")
		return
	}
	env[api.ProcessRecoverEnv] = api.EnableFunc
	trainEnv.Insert(api.ElasticTraining)
}
