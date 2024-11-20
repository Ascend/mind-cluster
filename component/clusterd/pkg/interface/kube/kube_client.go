// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CreateConfigMap create configMap here
func CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}

	return k8sClient.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(),
		cm, metav1.CreateOptions{})
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
