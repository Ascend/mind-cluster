// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/common-utils/hwlog"
)

var k8sClient *K8sClient

// K8sClient include name of node, config map and client of k8s
type K8sClient struct {
	ClientSet       kubernetes.Interface
	ClusterInfoName string
}

// InitClientK8s init k8s client
func InitClientK8s() error {
	if k8sClient == nil || k8sClient.ClientSet == nil {
		var err error
		k8sClient, err = newClientK8s()
		return err
	}
	return nil
}

// GetClientK8s get client k8s
func GetClientK8s() *K8sClient {
	return k8sClient
}

// newClientK8s create k8s client
func newClientK8s() (*K8sClient, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build client config err: %v", err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get client err: %v", err)
		return nil, err
	}

	return &K8sClient{
		ClientSet:       client,
		ClusterInfoName: "",
	}, nil
}
