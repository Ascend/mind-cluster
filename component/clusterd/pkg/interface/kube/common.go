// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
)

var (
	annotationsFormat = `{"metadata":{"annotations":%s}}`
	labelsFormat      = `{"metadata":{"labels":%s}}`
)

// GetNode get node from cache or api-server
func GetNode(name string) *v1.Node {
	node, err := GetNodeFromIndexer(name)
	if err == nil {
		return node
	}
	hwlog.RunLog.Warnf("get node %s from cache failed, err: %v", name, err)

	node, err = k8sClient.ClientSet.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("get node %s from client failed, err: %v", name, err)
		return nil
	}
	return node
}
