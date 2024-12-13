// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"strings"
)

func isContainsAny(str string, subStrs ...string) bool {
	for _, subStr := range subStrs {
		if strings.Contains(str, subStr) {
			return true
		}
	}
	return false
}
