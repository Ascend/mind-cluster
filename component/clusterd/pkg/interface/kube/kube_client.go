// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"clusterd/pkg/common/constant"
)

// CreateConfigMap create configMap here
func CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}

	return k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(),
		cm, metav1.CreateOptions{})
}

// CreateOrUpdateConfigMap create or update configMap.
func CreateOrUpdateConfigMap(cmName, nameSpace string, data, label map[string]string) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: nameSpace,
			Labels:    label},
		Data: data}
	_, cErr := k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(),
		cm, metav1.CreateOptions{})
	if cErr == nil {
		return nil
	}
	if !errors.IsAlreadyExists(cErr) {
		hwlog.RunLog.Errorf("create cm failed, err: %v", cErr)
		return fmt.Errorf("unable to create ConfigMap: %v", cErr)
	}
	_, err := k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(),
		cm, metav1.UpdateOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("unable to update ConfigMap: %v", err)
	}
	return err
}

// UpdateOrCreateConfigMap update or create configMap.
func UpdateOrCreateConfigMap(cmName, nameSpace string, data, label map[string]string) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: nameSpace,
			Labels:    label},
		Data: data}
	_, err := k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(),
		cm, metav1.UpdateOptions{})
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("update cm failed, err: %v", err)
		return fmt.Errorf("unable to update ConfigMap: %v", err)
	}
	_, cErr := k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(),
		cm, metav1.CreateOptions{})
	if cErr != nil {
		hwlog.RunLog.Errorf("unable to create ConfigMap: %v", cErr)
	}
	return cErr
}

// UpdateConfigMap update device info, which is cm
func UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	return k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(),
		cm, metav1.UpdateOptions{})
}

// GetConfigMap get configMap
func GetConfigMap(cmName, cmNamespace string) (*v1.ConfigMap, error) {
	return k8sClient.ClientSet.CoreV1().ConfigMaps(cmNamespace).Get(context.TODO(),
		cmName, metav1.GetOptions{})
}

// DeleteConfigMap delete configMap
func DeleteConfigMap(cmName, cmNamespace string) error {
	return k8sClient.ClientSet.CoreV1().ConfigMaps(cmNamespace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
}

// RetryPatchPodLabels retry patch pod labels
func RetryPatchPodLabels(pod *v1.Pod, retryTimes int, labels map[string]string) (*v1.Pod, error) {
	pod, err := PatchPodLabel(pod.Name, pod.Namespace, labels)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		pod, err = PatchPodLabel(pod.Name, pod.Namespace, labels)
	}
	return pod, err
}

// PatchPodLabel path pod label
func PatchPodLabel(podName, podNamespace string, labels map[string]string) (*v1.Pod, error) {
	labelStr, err := json.Marshal(labels)
	if err != nil {
		hwlog.RunLog.Errorf("marshal labels failed when path pod, err is %v", err)
		return nil, err
	}
	patchBody := fmt.Sprintf(labelsFormat, labelStr)
	return k8sClient.ClientSet.CoreV1().Pods(podNamespace).Patch(context.TODO(),
		podName, types.MergePatchType, []byte(patchBody), metav1.PatchOptions{})
}

// CheckVolcanoExist check volcano is existed
func CheckVolcanoExist(vcClient *versioned.Clientset) bool {
	if vcClient == nil {
		hwlog.RunLog.Error("vcK8sClient.ClientSet is nil")
		return false
	}
	_, err := vcClient.SchedulingV1beta1().PodGroups(constant.DefaultNamespace).Get(context.Background(),
		constant.TestName, metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), constant.NoResourceOnServer) {
		return false
	}
	return true
}
