// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/kube"
)

func TestBuildCmData(t *testing.T) {
	mgr := &OutputMgr{}
	curJobStatistic := constant.CurrJobStatistic{
		JobStatistic: map[string]constant.JobStatistic{
			"job1": {K8sJobID: "job1"},
			"job2": {K8sJobID: "job2"},
		},
	}

	cmData := mgr.BuildCmData(curJobStatistic)

	assert.Equal(t, "2", cmData[statistics.TotalJobsCmKey])
	assert.Contains(t, cmData[statistics.JobDataCmKey], `{"ID":"job1"}`)
	assert.Contains(t, cmData[statistics.JobDataCmKey], `{"ID":"job2"}`)
}

func TestCurrJobStcOutputWithDataChange(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock time ticker
	tickCh := make(chan time.Time)
	patches.ApplyFunc(time.NewTicker, func(time.Duration) *time.Ticker {
		return &time.Ticker{C: tickCh}
	})

	// Mock kube client
	var updateCalled bool
	patches.ApplyFunc(kube.UpdateOrCreateConfigMap, func(string, string, map[string]string, map[string]string) error {
		updateCalled = true
		return nil
	})

	// Mock statistic data
	var version int64 = statistics.InitVersion
	patches.ApplyMethodFunc(statistics.JobStcMgrInst,
		"GetAllJobStatistic", func() (constant.CurrJobStatistic, int64) {
			version++
			return constant.CurrJobStatistic{}, version
		})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go GlobalJobOutputMgr.JobOutput(ctx)

	// Simulate two ticks
	tickCh <- time.Now()
	time.Sleep(time.Second)
	tickCh <- time.Now()
	time.Sleep(time.Second)

	assert.True(t, updateCalled)
}

func TestCurrJobStcOutputNoDataChange(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock time ticker
	tickCh := make(chan time.Time)
	patches.ApplyFunc(time.NewTicker, func(time.Duration) *time.Ticker {
		return &time.Ticker{C: tickCh}
	})

	// Mock kube client
	var updateCalled bool
	patches.ApplyFunc(kube.UpdateConfigMap, func(*v1.ConfigMap) (*v1.ConfigMap, error) {
		updateCalled = true
		return nil, nil
	})

	// Mock statistic data with fixed version
	patches.ApplyMethodFunc(statistics.JobStcMgrInst,
		"GetAllJobStatistic", func() (constant.CurrJobStatistic, int64) {
			return constant.CurrJobStatistic{}, statistics.InitVersion
		})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go GlobalJobOutputMgr.JobOutput(ctx)

	// Simulate two ticks
	tickCh <- time.Now()
	time.Sleep(time.Second)
	tickCh <- time.Now()
	time.Sleep(time.Second)

	assert.False(t, updateCalled)
}
