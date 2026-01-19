// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/api/ascend-operator/client/clientset/versioned"
	"ascend-common/common-utils/hwlog"
)

var operatorClient *OperatorClient

// OperatorClient is the client of volcano
type OperatorClient struct {
	ClientSet *versioned.Clientset
}

// InitOperatorClient init operator client
func InitOperatorClient() (*OperatorClient, error) {
	var err error
	if operatorClient == nil || operatorClient.ClientSet == nil {
		operatorClient, err = newOperatorClient()
	}
	return operatorClient, err
}

// GetOperatorClient get operator client
func GetOperatorClient() *OperatorClient {
	return operatorClient
}

// newOperatorClient new operator client
func newOperatorClient() (*OperatorClient, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build  operator client config err: %v", err)
		return nil, err
	}
	client, err := versioned.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get operator client err: %v", err)
		return nil, err
	}
	return &OperatorClient{
		ClientSet: client,
	}, nil
}
