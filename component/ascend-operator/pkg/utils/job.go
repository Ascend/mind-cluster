// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package utils is common utils
package utils

import "ascend-operator/pkg/api/v1"

// IsMindIEEPJob judge mindIEEP job
func IsMindIEEPJob(job *v1.AscendJob) bool {
	if job == nil || job.Labels == nil {
		return false
	}
	if _, ok := job.Labels[v1.JodIdLabelKey]; !ok {
		return false
	}
	if _, ok := job.Labels[v1.AppLabelKey]; !ok {
		return false
	}
	return true
}

// IsSoftStrategyJob judge soft strategy job
func IsSoftStrategyJob(job *v1.AscendJob) bool {
	if job == nil || job.Labels == nil {
		return false
	}
	return job.Labels[SuperPodAffinity] == SoftStrategy
}
