// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
)

// IsNSAndNameMatched check whether its namespace and name match the configmap
func IsNSAndNameMatched(obj interface{}, namespace string, namePrefix string) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Error("cannot convert to ConfigMap")
		return false
	}
	return cm.Namespace == namespace && strings.HasPrefix(cm.Name, namePrefix)
}
