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

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	EmptyLength = 0
)

func IsSliceEmpty[T any](slice []T) bool {
	return slice != nil && len(slice) == EmptyLength
}

func IsObjectEmpty(obj client.Object) bool {
	if obj == nil {
		return true
	}
	if obj.GetName() == "" || obj.GetNamespace() == "" {
		hwlog.RunLog.Debugf("object<%v> is empty", reflect.TypeOf(obj))
		return true
	}
	return false
}

func IfNeedRequeue(err error) bool {
	var reQueueError *ReQueueError
	ok := errors.As(err, &reQueueError)
	return ok
}

func IsRunning(status commonv1.JobStatus) bool {
	return HasCondition(status, commonv1.JobRunning)
}

func GetStatusString(status commonv1.JobStatus) string {
	for _, condition := range status.Conditions {
		if condition.Status == corev1.ConditionTrue {
			return string(condition.Type)
		}
	}
	return ""
}

func HasCondition(status commonv1.JobStatus, condType commonv1.JobConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == condType && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func ConvertToPointerSlice[T any](slice []T) []*T {
	pointerSlice := make([]*T, len(slice))
	for index := range slice {
		pointerSlice[index] = &slice[index]
	}
	return pointerSlice
}

func FilterPodsByReplicaType(pods []*corev1.Pod, rt string) []*corev1.Pod {
	var filtered []*corev1.Pod
	for _, pod := range pods {
		if pod.Labels[commonv1.ReplicaTypeLabel] == rt {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

func GetServiceName(jobNamespace, jobName string) string {
	return GetPodGroupName(jobNamespace, jobName, RayHeadType)
}

func GetPodGroupName(jobNamespace, jobName, rType string) string {
	jobNamespace = strings.ToLower(jobNamespace)
	jobName = strings.ToLower(jobName)
	rType = strings.ToLower(rType)
	return fmt.Sprintf("%s-%s-%s", jobNamespace, jobName, rType)
}

func DeepCopyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	copyMap := make(map[string]string)
	for k, v := range m {
		copyMap[k] = v
	}
	return copyMap
}

func SyncStatus(target *commonv1.JobStatus, source *commonv1.JobStatus) {
	sourceConditionCopy := source.DeepCopy().Conditions
	target.Conditions = sourceConditionCopy
}

func GenHeadParamMap(labels map[string]string) map[string]string {
	paramMap := make(map[string]string)
	paramMap[RayPortParamKey] = labels[RayGcsPortLabelKey]
	paramMap[RayDashHostParamKey] = DefaultDashboardHost
	paramMap[RayDashPortParamKey] = labels[RayDashboardPortLabelKey]
	paramMap[RayNodeManagerParamKey] = DefaultNodeManagerPort
	paramMap[RayObjectManagerParamKey] = DefaultObjectManagerPort
	paramMap[RayLabelParamKey] = fmt.Sprintf("\"$%s\"", RayLabelEnvKey)
	paramMap[RayResourceParamKey] = fmt.Sprintf("\"$%s\"", RayResourcesEnvKey)
	return paramMap
}

func GenWorkerParamMap(labels map[string]string, svcName string) map[string]string {
	paramMap := make(map[string]string)
	paramMap[RayAddressParamKey] = fmt.Sprintf("%s:%s", svcName, labels[RayGcsPortLabelKey])
	paramMap[RayNodeManagerParamKey] = DefaultNodeManagerPort
	paramMap[RayObjectManagerParamKey] = DefaultObjectManagerPort
	paramMap[RayLabelParamKey] = fmt.Sprintf("\"$%s\"", RayLabelEnvKey)
	paramMap[RayResourceParamKey] = fmt.Sprintf("\"$%s\"", RayResourcesEnvKey)
	return paramMap
}

func genRayLabel(annos map[string]string) RayLabel {
	instance, err := parseDeviceInfo(annos)
	if err != nil {
		hwlog.RunLog.Warnf("GenRayLabel parse device info failed: %v, ray label will be ignored", err)
		return RayLabel{}
	}
	if instance == nil {
		hwlog.RunLog.Warnf("GenRayLabel: pod annotation<%s> not found, ray label will be ignored",
			api.Pod910DeviceAnno)
		return RayLabel{}
	}
	var levelList []string
	levelList = append(levelList, DefaultCluster)
	if instance.SuperPodId != nil && *instance.SuperPodId != -1 {
		superPodId := strconv.Itoa(int(*instance.SuperPodId))
		levelList = append(levelList, SuperPodPrefix+superPodId)
	}
	if instance.RackID != nil && *instance.RackID != -1 {
		rackId := strconv.Itoa(int(*instance.RackID))
		levelList = append(levelList, RackPrefix+rackId)
	}
	nodeId := instance.ServerID
	levelList = append(levelList, NodePrefix+nodeId)
	rayLabel := getRayLabelFromList(levelList)
	rayLabel.NPU = len(instance.Devices)
	return rayLabel
}

// SetRayEnvToCM add ray label env into configmap, and return value means if
// we need to update configmap later
func SetRayEnvToCM(pod *corev1.Pod, cm *corev1.ConfigMap) bool {
	needUpdate := false
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	envKey := GenRayEnvScriptName(pod.Name)
	if _, ok := cm.Data[envKey]; !ok {
		// make sure each pod has its field in configmap
		cm.Data[envKey] = ""
		needUpdate = true
	}
	if cm.Data[envKey] != "" {
		return needUpdate
	}
	hwlog.RunLog.Infof("Set Env pod: %s/%s", pod.Namespace, pod.Name)
	// add RAY_RESOURCES env
	rayLabel := genRayLabel(pod.Annotations)
	labelString, err := marshalFromValue(rayLabel)
	if err != nil {
		hwlog.RunLog.Warnf("marshal ray label failed: %v, RAY_LABEL will be ignored", err)
		return needUpdate
	}
	if labelString == "" {
		hwlog.RunLog.Warnf("Set ray label Env: generated ray label is zero value, RAY_LABEL will be ignored")
		return needUpdate
	}
	envValue := fmt.Sprintf("export %s='%s'; ", RayLabelEnvKey, string(labelString))

	// add RAY_RESOURCES env
	rayResources := fmt.Sprintf("{\"NPU\":%d,\"%s\":1}", rayLabel.NPU, rayLabel.L0)
	envValue += fmt.Sprintf("export %s='%s'; ", RayResourcesEnvKey, rayResources)
	cm.Data[envKey] = envValue
	return true
}

func SetCommonEnv(pod *corev1.PodTemplateSpec) {
	// add ASCEND_SUPERPOD_BLOCK_SIZE env
	ascendSpBlock := AscendSuperpodBlock{
		SpBlock: pod.Annotations[SpBlockAnnotationKey],
		TpBlock: pod.Annotations[TpBlockAnnotationKey],
	}
	blockString, err := marshalFromValue(ascendSpBlock)
	if err != nil {
		hwlog.RunLog.Warnf("marshal sp block failed: %v, ASCEND_SUPERPOD_BLOCK_SIZE will be ignored", err)
		return
	}
	if blockString == "" {
		hwlog.RunLog.Info("Set sp block Env: generated ray label is zero value, " +
			"ASCEND_SUPERPOD_BLOCK_SIZE will be ignored")
		return
	}
	for index, _ := range pod.Spec.Containers {
		pod.Spec.Containers[index].Env = append(pod.Spec.Containers[index].Env, corev1.EnvVar{
			Name:  AscendSuperpodEnvKey,
			Value: blockString,
		})
	}
}

func SetVolumes(podTemplate *corev1.PodTemplateSpec, cmIdentification string) {
	scriptFileName := GenRayEnvScriptName(podTemplate.Name)
	cmName := GenConfigInfoConfigMapName(cmIdentification)
	newVolume := corev1.Volume{
		Name: RayScriptVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  scriptFileName,
						Path: scriptFileName,
					},
				},
			},
		},
	}
	podTemplate.Spec.Volumes = append(podTemplate.Spec.Volumes, newVolume)

	for index, _ := range podTemplate.Spec.Containers {
		newVolumeMount := corev1.VolumeMount{
			Name:      RayScriptVolumeName,
			MountPath: RayEnvMountPath,
		}
		podTemplate.Spec.Containers[index].VolumeMounts =
			append(podTemplate.Spec.Containers[index].VolumeMounts, newVolumeMount)
	}
}

func marshalFromValue[T any](value T) (string, error) {
	var zero T
	if reflect.DeepEqual(value, zero) {
		return "", nil
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func GenConfigInfoConfigMapName(identification string) string {
	return fmt.Sprintf(RayInfoConfigMapName, identification)
}

func GenRayEnvScriptName(identification string) string {
	return fmt.Sprintf(ConfigMapRayEnvKey, identification)
}

func GetTotalReplicas(replicaMap map[commonv1.ReplicaType]*commonv1.ReplicaSpec) int32 {
	jobReplicas := int32(0)
	for _, spec := range replicaMap {
		jobReplicas += *spec.Replicas
	}
	return jobReplicas
}

func parseDeviceInfo(annos map[string]string) (*Instance, error) {
	deviceInfo, ok := annos[api.Pod910DeviceAnno]
	if !ok {
		return nil, nil
	}
	var instance Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		hwlog.RunLog.Errorf("unmarshal deviceInfo failed: %v", err)
		return nil, err
	}
	hwlog.RunLog.Debugf("instance: %v", instance)
	return &instance, nil
}

func getRayLabelFromList(levelList []string) RayLabel {
	switch len(levelList) {
	// A2
	case 2:
		return RayLabel{
			L2: levelList[0],
			L1: DefaultLabelL1,
			L0: levelList[1],
		}
	// A3
	case 3:
		return RayLabel{
			L2: levelList[0],
			L1: levelList[1],
			L0: levelList[2],
		}
	// A5
	case 4:
		return RayLabel{
			L3: levelList[0],
			L2: levelList[1],
			L1: levelList[2],
			L0: levelList[3],
		}
	default:
		{
			hwlog.RunLog.Warnf("error generate ray label, maybe level info is wrong in pod annotation<%s> ",
				api.Pod910DeviceAnno)
			return RayLabel{}
		}
	}
}

// InitializeReplicaStatuses initializes the ReplicaStatuses for replica.
func InitializeReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType) {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = make(map[commonv1.ReplicaType]*commonv1.ReplicaStatus)
	}

	jobStatus.ReplicaStatuses[rtype] = &commonv1.ReplicaStatus{}
}

// UpdateJobReplicaStatuses updates the JobReplicaStatuses according to the pod.
func UpdateJobReplicaStatuses(jobStatus *commonv1.JobStatus, rtype commonv1.ReplicaType, pod *corev1.Pod) {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = make(map[commonv1.ReplicaType]*commonv1.ReplicaStatus)
	}
	hwlog.RunLog.Debugf("before updateJobReplicaStatuses  status<%#v> by pod<%s> phase<%s>",
		jobStatus.ReplicaStatuses[rtype], pod.Name, pod.Status.Phase)
	defer hwlog.RunLog.Debugf("after updateJobReplicaStatuses status<%#v>", jobStatus.ReplicaStatuses[rtype])
	switch pod.Status.Phase {
	case corev1.PodRunning:
		jobStatus.ReplicaStatuses[rtype].Active++
	case corev1.PodSucceeded:
		jobStatus.ReplicaStatuses[rtype].Succeeded++
	case corev1.PodFailed:
		if pod.DeletionTimestamp != nil {
			hwlog.RunLog.Infof("pod<%s> is deleting, so it can not be treat as failed", pod.Name)
			return
		}
		jobStatus.ReplicaStatuses[rtype].Failed++
	default:
	}
}

func GetSchedulerName(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) string {
	for _, spec := range replicas {
		if len(spec.Template.Spec.SchedulerName) > 0 {
			return spec.Template.Spec.SchedulerName
		}
	}
	return ""
}

func GenerateRayStartCommand(nodeType string, params map[string]string) string {
	switch nodeType {
	case RayHeadType:
		return fmt.Sprintf("ray start --head %s --block", ConvertParamMapToString(params))
	default:
		return fmt.Sprintf("ray start %s --block", ConvertParamMapToString(params))
	}
}

func AddCommandToSpec(podTemplate *corev1.PodTemplateSpec, command string) {
	defaultContainerIndex := findDefaultContainerIndex(podTemplate.Spec.Containers)
	if defaultContainerIndex == -1 {
		return
	}
	containers := podTemplate.Spec.Containers
	var args []string
	envScriptName := GenRayEnvScriptName(podTemplate.Name)
	args = append(args, fmt.Sprintf("source %s/%s", RayEnvMountPath, envScriptName))
	args = append(args, fmt.Sprintf("%s=${%s:-'{}'}", RayLabelEnvKey, RayLabelEnvKey))
	args = append(args, fmt.Sprintf("%s=${%s:-'{}'}", RayResourcesEnvKey, RayResourcesEnvKey))
	args = append(args, command)
	argString := strings.Join(args, ";\n")

	var newArgs string
	defaultContainer := &containers[defaultContainerIndex]
	if !IsOneStringCommandMode(defaultContainer) {
		newArgs = strings.Join(defaultContainer.Command, " ") + strings.Join(defaultContainer.Args, " ")
		defaultContainer.Command = []string{"/bin/bash", "-c"}
	} else if len(defaultContainer.Args) > 0 {
		newArgs = defaultContainer.Args[0]
	}
	newArgs += "\n"

	containers[defaultContainerIndex].Args = []string{newArgs + argString}
}

// IsOneStringCommandMode means command config like below:
// command: ["/bin/bash", "-c"]
// args: ["echo starting app && exec my-app"]
func IsOneStringCommandMode(container *corev1.Container) bool {
	commandField := container.Command
	return commandField != nil && len(commandField) > 0 &&
		(strings.Contains(commandField[0], "sh") ||
			strings.Contains(commandField[0], "bash")) &&
		commandField[1] == "-c"
}

func ConvertParamMapToString(m map[string]string) string {
	var str string
	for k, v := range m {
		str = str + fmt.Sprintf("--%s=%s ", k, v)
	}
	return str
}

func AddEnvValue(container *corev1.Container, envKey, envValue string) {
	container.Env = append(container.Env, corev1.EnvVar{
		Name:  envKey,
		Value: envValue,
	})
}

func findDefaultContainerIndex(containers []corev1.Container) int {
	defaultContainerIndex := -1
	for index, container := range containers {
		if container.Name == DefaultContainerName {
			defaultContainerIndex = index
		}
	}
	if defaultContainerIndex == -1 {
		hwlog.RunLog.Warnf("must have container %s to submit verl task or start ray cluster, but not found",
			DefaultContainerName)
	}
	return defaultContainerIndex
}
