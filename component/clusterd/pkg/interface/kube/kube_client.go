// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

var refreshCMIds sync.Map

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
func RetryPatchPodLabels(podName, podNamespace string, retryTimes int, labels map[string]string) error {
	_, err := PatchPodLabel(podName, podNamespace, labels)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		_, err = PatchPodLabel(podName, podNamespace, labels)
	}
	return err
}

// RetryPatchPodAnnotations retry patch pod annotations
func RetryPatchPodAnnotations(podName, podNamespace string, retryTimes int, labels map[string]string) error {
	_, err := PatchPodAnnotation(podName, podNamespace, labels)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		_, err = PatchPodAnnotation(podName, podNamespace, labels)
	}
	return err
}

// PatchPodLabel path pod label
func PatchPodLabel(podName, podNamespace string, labels map[string]string) (*v1.Pod, error) {
	return patchPod("label", podName, podNamespace, labels)
}

// PatchPodAnnotation path pod annotation
func PatchPodAnnotation(podName, podNamespace string, labels map[string]string) (*v1.Pod, error) {
	return patchPod("annnotation", podName, podNamespace, labels)
}

// patchPod path pod where given format
func patchPod(patchType, podName, podNamespace string, data map[string]string) (*v1.Pod, error) {
	labelStr, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data failed when path pod, err is %v", err)
		return nil, err
	}

	var patchBody string
	if patchType == "label" {
		patchBody = fmt.Sprintf(labelsFormat, labelStr)
	} else if patchType == "annnotation" {
		patchBody = fmt.Sprintf(annotationsFormat, labelStr)
	} else {
		return nil, fmt.Errorf("patchType is error")
	}

	return k8sClient.ClientSet.CoreV1().Pods(podNamespace).Patch(context.TODO(),
		podName, types.MergePatchType, []byte(patchBody), metav1.PatchOptions{})
}

// PatchCMData path configmap data
func PatchCMData(name, namespace string, data map[string]string) (*v1.ConfigMap, error) {
	dataFormat := `{"data":%s}`
	dataByte, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal cm data failed, error: %v", err)
		return nil, fmt.Errorf("marshal cm data failed")
	}
	patchBody := fmt.Sprintf(dataFormat, dataByte)
	return k8sClient.ClientSet.CoreV1().ConfigMaps(namespace).Patch(context.TODO(),
		name, types.MergePatchType, []byte(patchBody), metav1.PatchOptions{})
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

// GetJobEvent get job event
func GetJobEvent(namespace, name, jobType string) (*v1.EventList, error) {
	fieldSelector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.namespace", namespace),
		fields.OneTermEqualSelector("involvedObject.kind", jobType),
	).String()
	events, err := k8sClient.ClientSet.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		hwlog.RunLog.Errorf("get events faild: %s", err)
		return nil, err
	}
	return events, nil
}

// CreateOrUpdateSuperPodFaultInfo create or update fault job info configmap
func CreateOrUpdateSuperPodFaultInfo(jobId string, faultInfos map[int]api.SuperPodFaultInfos) {
	for i := 0; i < constant.RetryTime; i++ {
		cm, getErr := GetConfigMap(api.FaultJobCmName, api.ClusterNS)
		if getErr != nil && !errors.IsNotFound(getErr) {
			hwlog.RunLog.Errorf("get configmap fault-job-info err:%v", getErr)
			time.Sleep(time.Second * time.Duration(i))
			continue
		}
		if cm != nil && cm.Data == nil {
			cm = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      api.FaultJobCmName,
					Namespace: api.ClusterNS,
				},
				Data: map[string]string{},
			}
		}
		faultData, err := json.Marshal(faultInfos)
		if err != nil {
			hwlog.RunLog.Errorf("marshal fault info failed, err:%v", err)
			return
		}
		cm.Data[jobId] = string(faultData)
		if getErr != nil && errors.IsNotFound(getErr) {
			if _, createErr := CreateConfigMap(cm); createErr != nil {
				hwlog.RunLog.Errorf("create configmap fault-job-info err:%v", createErr)
			}
			return
		}
		if _, updateErr := UpdateConfigMap(cm); updateErr != nil {
			hwlog.RunLog.Errorf("update configmap fault-job-info err:%v", updateErr)
			time.Sleep(time.Second * time.Duration(i))
			continue
		}
		refreshCMIds.Store(jobId, struct{}{})
		return
	}
}

// RecoverFaultJobInfoCmWithSync update cm with sync
func RecoverFaultJobInfoCmWithSync(jobId string) {
	_, ok := refreshCMIds.Load(jobId)
	if !ok {
		time.Sleep(time.Minute)
	}
	RecoverFaultJobInfoCm(jobId)
}

// RecoverFaultJobInfoCm update cm
func RecoverFaultJobInfoCm(jobId string) {
	if k8sClient == nil || k8sClient.ClientSet == nil {
		return
	}
	time.Sleep(constant.ReleaseTimeOut)
	for i := 0; i < constant.RetryTime; i++ {
		cm, getErr := GetConfigMap(api.FaultJobCmName, api.ClusterNS)
		if getErr != nil {
			if errors.IsNotFound(getErr) {
				hwlog.RunLog.Warnf("get configmap fault-job-info err:%v", getErr)
				return
			}
			hwlog.RunLog.Errorf("get configmap fault-job-info err:%v", getErr)
			continue
		}
		if _, ok := cm.Data[jobId]; !ok {
			hwlog.RunLog.Errorf("configmap fault-job-info not found jobId:%s", jobId)
			return
		}
		delete(cm.Data, jobId)
		if _, updateErr := UpdateConfigMap(cm); updateErr != nil {
			hwlog.RunLog.Errorf("update configmap fault-job-info err:%v", updateErr)
			time.Sleep(time.Second)
		}
		hwlog.RunLog.Infof("delete jobId:%s from configmap fault-job-info success", jobId)
		refreshCMIds.Delete(jobId)
		return
	}
}

// PatchNodeAnnotation path node annotation
func PatchNodeAnnotation(nodeName string, annotations map[string]string) (*v1.Node, error) {
	annotationStr, err := json.Marshal(annotations)
	if err != nil {
		hwlog.RunLog.Errorf("marshal annotations failed when patch node, err is %v", err)
		return nil, err
	}

	patchBody := fmt.Sprintf(annotationsFormat, annotationStr)
	return k8sClient.ClientSet.CoreV1().Nodes().Patch(context.TODO(),
		nodeName,
		types.MergePatchType,
		[]byte(patchBody),
		metav1.PatchOptions{})
}

// RetryPatchNodeAnnotation retry patch Node annotation
func RetryPatchNodeAnnotation(nodeName string, retryTimes int, annotations map[string]string) error {
	_, err := PatchNodeAnnotation(nodeName, annotations)
	retry := 0
	for err != nil && retry < retryTimes-1 {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		_, err = PatchNodeAnnotation(nodeName, annotations)
	}
	return err
}

// DeletePodAnnotation delete pod annotations
func DeletePodAnnotation(namespace, podName string, keysToDelete []string) error {
	var lastErr error = nil
	for i := 0; i < constant.PatchPodTimes; i++ {
		patchData := map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{},
			},
		}
		for _, key := range keysToDelete {
			patchData["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[key] = nil
		}
		patchBytes, err := json.Marshal(patchData)
		if err != nil {
			hwlog.RunLog.Errorf("marshal patch data failed, attempt=%d, err=%v", i+1, err)
			return err
		}
		_, err = k8sClient.ClientSet.CoreV1().Pods(namespace).Patch(
			context.TODO(), podName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
		if err == nil {
			hwlog.RunLog.Infof("delete annotation success, namespace=%s, pod=%s, keys=%v",
				namespace, podName, keysToDelete)
			return nil
		}
		lastErr = err
		hwlog.RunLog.Warnf("delete pod annotation failed, attempt=%d, namespace=%s, pod=%s, keys=%v, err=%v",
			i+1, namespace, podName, keysToDelete, err)
		if i < constant.PatchPodTimes-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	hwlog.RunLog.Errorf("delete pod annotation failed after 3 attempts, namespace=%s, pod=%s, keys=%v, lastErr=%v",
		namespace, podName, keysToDelete, lastErr)
	return lastErr
}
