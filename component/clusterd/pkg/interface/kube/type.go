// Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"time"
)

// faultJobReleaseInfo fault job release info
type faultJobReleaseInfo struct {
	jobId      string
	nodeName   string
	duration   time.Duration
	createTime int64
}
