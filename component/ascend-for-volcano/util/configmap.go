/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package util is using for HuaWei infer common Ascend pin affinity schedule.
*/
package util

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// GetConfigMapWithRetry  Get config map from k8s.
func GetConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	var cm *v1.ConfigMap
	var err error

	// There can be no delay or blocking operations in a session.
	if cm, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{}); err != nil {
		return nil, err
	}

	return cm, nil
}

func DelConfigMapWithRetry(client kubernetes.Interface, namespace, cmName string) error {
	var err error
	if err = client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), cmName, metav1.DeleteOptions{}); err != nil {
		return err
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

// CreateOrUpdateConfigMap Create or update configMap.
func CreateOrUpdateConfigMap(k8s kubernetes.Interface, cm *v1.ConfigMap, cmName, nameSpace string) error {
	klog.V(LogDebugLev).Infof("cmName: %s, cmNamespace: %s", cmName, cm.ObjectMeta.Namespace)
	_, cErr := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if cErr != nil {
		if !errors.IsAlreadyExists(cErr) {
			return fmt.Errorf("unable to create ConfigMap:%s", SafePrint(cErr))
		}

		// To reduce the cm write operations
		if !IsConfigMapChanged(k8s, cm, cmName, nameSpace) {
			klog.V(LogInfoLev).Infof("configMap not changed,no need update")
			return nil
		}

		_, err := k8s.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update ConfigMap:%s", SafePrint(cErr))
		}
	}
	return nil
}

// UpdateConfigmapIncrementally update configmap Map data but keep the key value pair that new data does not have
func UpdateConfigmapIncrementally(kubeClient kubernetes.Interface, ns, name string,
	newData map[string]string) (map[string]string, error) {
	if len(newData) == 0 {
		return newData, fmt.Errorf("newData is empty")
	}
	oldCM, err := GetConfigMapWithRetry(kubeClient, ns, name)
	if err != nil || oldCM == nil {
		upCmErr := fmt.Errorf("get old configmap from kubernetes failed err:%s", SafePrint(err))
		if name != CmName {
			return newData, upCmErr
		}
		return updateFaultCMCheckCode(newData), upCmErr
	}
	oldCMData := oldCM.Data
	if oldCMData != nil {
		for key, value := range oldCMData {
			_, ok := newData[key]
			if !ok {
				newData[key] = value // place the key-value pairs from kubernetes back
				continue
			}
		}
	}
	if name != CmName {
		return newData, nil
	}
	return updateFaultCMCheckCode(newData), nil
}

func InformerConfigmapFilter(obj interface{}) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		klog.V(LogErrorLev).Infof("Cannot convert to ConfigMap:%#v", obj)
		return false
	}
	return CheckConfigMapIsDeviceInfo(cm) || CheckConfigMapIsNodeInfo(cm)
}

// CheckConfigMapIsDeviceInfo check configmap is device info
func CheckConfigMapIsDeviceInfo(cm *v1.ConfigMap) bool {
	return cm.Namespace == DevInfoNameSpace && strings.HasPrefix(cm.Name, DevInfoPreName)
}

// CheckConfigMapIsNodeInfo check whether the configmap is kube-system/node-info-
func CheckConfigMapIsNodeInfo(cm *v1.ConfigMap) bool {
	return cm.Namespace == MindXDlNameSpace && strings.HasPrefix(cm.Name, NodeDCmInfoNamePrefix)
}

// MakeDataHash check code for configmap
func MakeDataHash(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	h := sha256.New()
	if _, err := h.Write(dataBuffer); err != nil {
		klog.V(LogErrorLev).Infof("hash data error")
		return ""
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(LogErrorLev).Infof("marshal data err: %s", SafePrint(err))
		return nil
	}
	return dataBuffer
}

func updateFaultCMCheckCode(newData map[string]string) map[string]string {
	_, ok := newData[CmCheckCode]
	if ok {
		delete(newData, CmCheckCode) // if check code exists, delete and create new
	}
	checkCode := MakeDataHash(newData)
	newData[CmCheckCode] = checkCode
	return newData
}
