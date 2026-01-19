// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube main test for kube
package kube

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/common-utils/hwlog"
)

var (
	testErr       = errors.New("test error")
	testK8sClient *K8sClient
)

func TestMain(m *testing.M) {
	var patches = gomonkey.ApplyFuncReturn(newClientK8s,
		&K8sClient{
			ClientSet: fake.NewSimpleClientset(),
		}, nil)
	defer patches.Reset()
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	if err := initK8sClient(); err != nil {
		return err
	}
	return nil
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func initK8sClient() error {
	err := InitClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("init k8s client failed when start, err: %v", err)
		return err
	}
	testK8sClient = GetClientK8s()
	return nil
}
