// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node main test for node
package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

var testErr = errors.New("test error")

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	return constructNodeInfo()
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

func constructNodeInfo() error {
	baseDevInfo, err := json.Marshal(baseDeviceMap)
	if err != nil {
		return err
	}
	node = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName1,
			Annotations: map[string]string{
				api.NodeSNAnnotation: nodeSN1,
				superPodIDKey:        superPodIDStr,
				baseDevInfoAnno:      string(baseDevInfo),
			},
		},
	}
	return nil
}
