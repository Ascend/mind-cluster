// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"context"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const backOffTime = 2 * time.Second

// GetServiceIpWithRetry get service ip with retry
func GetServiceIpWithRetry(k kubernetes.Interface, nameSpace, name string) string {
	retryTimes := 3
	for i := 0; i < retryTimes; i++ {
		svc, err := k.CoreV1().Services(nameSpace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			time.Sleep(backOffTime)
			hwlog.RunLog.Errorf("get svc from api server failed by:%v", err)
			continue
		}
		return svc.Spec.ClusterIP
	}
	return ""
}
