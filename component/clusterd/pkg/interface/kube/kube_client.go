// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
