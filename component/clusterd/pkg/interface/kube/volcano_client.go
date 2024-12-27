// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"ascend-common/common-utils/hwlog"
)

var vcK8sClient *VcK8sClient

// VcK8sClient is the client of volcano
type VcK8sClient struct {
	ClientSet *versioned.Clientset
}

// InitClientVolcano init volcano client
func InitClientVolcano() (*VcK8sClient, error) {
	var err error
	if vcK8sClient == nil || vcK8sClient.ClientSet == nil {
		vcK8sClient, err = newVCClientK8s()
	}
	return vcK8sClient, err
}

// GetClientVolcano get client volcano
func GetClientVolcano() *VcK8sClient {
	return vcK8sClient
}

// newVCClientK8s new vcjob client
func newVCClientK8s() (*VcK8sClient, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build vcClient config err: %v", err)
		return nil, err
	}
	client, err := versioned.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get vcClient err: %v", err)
		return nil, err
	}
	return &VcK8sClient{
		ClientSet: client,
	}, nil
}

// RetryGetPodGroup call GetPodGroup retryTimes
func RetryGetPodGroup(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
	pg, err := GetPodGroup(name, namespace)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		pg, err = GetPodGroup(name, namespace)
	}
	return pg, err
}

// GetPodGroup return pod group according pod group name
func GetPodGroup(name, namespace string) (*v1beta1.PodGroup, error) {
	if vcK8sClient != nil {
		return vcK8sClient.ClientSet.SchedulingV1beta1().PodGroups(namespace).Get(context.TODO(),
			name, v1.GetOptions{})
	}
	return nil, fmt.Errorf("vcK8sClient is nil")
}

// RetryUpdatePodGroup call UpdatePod
func RetryUpdatePodGroup(pg *v1beta1.PodGroup, retryTimes int) (*v1beta1.PodGroup, error) {
	pg, err := UpdatePodGroup(pg)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		pg, err = UpdatePodGroup(pg)
	}
	return pg, err
}

// UpdatePodGroup update pod group
func UpdatePodGroup(pg *v1beta1.PodGroup) (*v1beta1.PodGroup, error) {
	if vcK8sClient != nil {
		return vcK8sClient.ClientSet.SchedulingV1beta1().PodGroups(pg.ObjectMeta.Namespace).Update(context.TODO(),
			pg, v1.UpdateOptions{})
	}
	return nil, fmt.Errorf("vcK8sClient is nil")
}

// RetryPatchPodGroupAnnotations retry patch pod group annotations
func RetryPatchPodGroupAnnotations(pgName, pgNamespace string, retryTimes int,
	annotations map[string]string) (*v1beta1.PodGroup, error) {
	pg, err := patchPodGroupAnnotation(pgName, pgNamespace, annotations)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		pg, err = patchPodGroupAnnotation(pgName, pgNamespace, annotations)
	}
	return pg, err
}

func patchPodGroupAnnotation(pgName, pgNamespace string, annotations map[string]string) (*v1beta1.PodGroup, error) {
	if vcK8sClient == nil || vcK8sClient.ClientSet == nil {
		hwlog.RunLog.Errorf("client set is nil")
		return nil, fmt.Errorf("client set is nil")
	}
	annotationStr, err := json.Marshal(annotations)
	if err != nil {
		hwlog.RunLog.Errorf("marshal labels failed when path pod, err is %v", err)
		return nil, err
	}
	patchBody := fmt.Sprintf(annotationsFormat, annotationStr)
	hwlog.RunLog.Infof("prepare patch pg annotation, pgName=%s, pgNamespace=%s", pgName, pgNamespace)
	return vcK8sClient.ClientSet.SchedulingV1beta1().PodGroups(pgNamespace).Patch(context.TODO(),
		pgName, types.MergePatchType, []byte(patchBody), v1.PatchOptions{})
}

// RetryPatchPodGroupLabel retry patch pod group label
func RetryPatchPodGroupLabel(pgName, nameSpace string, retryTimes int,
	labels map[string]string) (*v1beta1.PodGroup, error) {
	pg, err := patchPodGroupLabel(pgName, nameSpace, labels)
	retry := 0
	for err != nil && retry < retryTimes {
		retry++
		time.Sleep(time.Second * time.Duration(retry))
		pg, err = patchPodGroupLabel(pgName, nameSpace, labels)
	}
	return pg, err
}

func patchPodGroupLabel(pgName, pgNamespace string, labels map[string]string) (*v1beta1.PodGroup, error) {
	if vcK8sClient == nil || vcK8sClient.ClientSet == nil {
		hwlog.RunLog.Errorf("client set is nil")
		return nil, fmt.Errorf("client set is nil")
	}
	labelStr, err := json.Marshal(labels)
	if err != nil {
		hwlog.RunLog.Errorf("marshal labels failed when path pod, err is %v", err)
		return nil, err
	}
	patchBody := fmt.Sprintf(labelsFormat, labelStr)
	hwlog.RunLog.Infof("prepare patch pg label, pgName=%s, pgNamespace=%s", pgName, pgNamespace)
	return vcK8sClient.ClientSet.SchedulingV1beta1().PodGroups(pgNamespace).Patch(context.TODO(),
		pgName, types.MergePatchType, []byte(patchBody), v1.PatchOptions{})
}
