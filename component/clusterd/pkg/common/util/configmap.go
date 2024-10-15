// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"clusterd/pkg/common/constant"
)

// CreateOrUpdateConfigMap Create or update configMap.
func CreateOrUpdateConfigMap(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, nameSpace string) error {
	hwlog.RunLog.Infof("cmName: %s, cmNamespace: %s", cmName, cm.ObjectMeta.Namespace)
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !errors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap: %v", cErr)
		}

		// To reduce the cm write operations
		if !IsConfigMapChanged(k8s, cm, cmName, nameSpace) {
			hwlog.RunLog.Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap: %v", err)
		}
	}
	return nil
}

// IsConfigMapChanged judge the cm wither is same. true is no change.
func IsConfigMapChanged(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, nameSpace string) bool {
	cmData, getErr := GetConfigMapWithRetry(k8s, nameSpace, cmName)
	if getErr != nil {
		return true
	}
	if reflect.DeepEqual(cmData, cm) {
		return false
	}

	return true
}

// GetConfigMapWithRetry Get config map from k8s.
func GetConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var err error

	for i := 0; i < constant.RetryTime; i++ {
		// There can be no delay or blocking operations in a session.
		if cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{}); err != nil {
			time.Sleep(constant.RetrySleepTime)
			continue
		}
		return cm, nil
	}
	return nil, err
}

// IsNSAndNameMatched check whether its namespace and name match the configmap
func IsNSAndNameMatched(obj interface{}, namespace string, namePrefix string) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("Cannot convert to ConfigMap:%v", obj)
		return false
	}
	return cm.Namespace == namespace && strings.HasPrefix(cm.Name, namePrefix)
}

// CreateOrUpdateCm Create or update configMap.
func CreateOrUpdateCm(k8s kubernetes.Interface, cm *v1.ConfigMap) error {
	tmpCm, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Get(context.TODO(), cm.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			cm.ResourceVersion = "0"
			_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			if cErr != nil {
				return fmt.Errorf("unable to create ConfigMap:%v", cErr)
			}
			return nil
		}
		return err
	}
	tmpCm.Data = cm.Data
	_, upErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), tmpCm, metav1.UpdateOptions{})
	return upErr
}

// GetAndUpdateCmByTotalNum get and update cm by total cut num
func GetAndUpdateCmByTotalNum(total, name, ns string, m map[string]string, k8s kubernetes.Interface) error {
	totalNum, cgErr := strconv.Atoi(total)
	if cgErr != nil {
		return cgErr
	}
	for i := 1; i < totalNum; i++ {
		cm, err := k8s.CoreV1().ConfigMaps(ns).Get(context.TODO(), name+"-"+strconv.Itoa(i), metav1.GetOptions{})
		if err != nil {
			return err
		}
		for k, v := range m {
			cm.Data[k] = v
		}
		_, upErr := k8s.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if upErr != nil {
			return upErr
		}
	}
	return nil
}
