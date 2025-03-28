// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault main test for public fault
package publicfault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/yaml"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

var (
	oriDevInfo1    = make(map[string]*constant.DeviceInfo)
	expDeviceInfo1 = make(map[string]*constant.DeviceInfo)
	oriDevInfo2    = make(map[string]*constant.DeviceInfo)
	expDeviceInfo2 = make(map[string]*constant.DeviceInfo)
)

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
	if err := initTestDataFromYaml(); err != nil {
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

func initTestDataFromYaml() error {
	const maxFileSize = 10000
	var testDataPath = "../../../../../testdata/resource/pub_fault_processor_test.yaml"

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		return errors.New("test data path is invalid")
	}
	if fileInfo.Size() > int64(maxFileSize) {
		return errors.New("test data path size is invalid")
	}
	open, err := os.Open(testDataPath)
	if err != nil {
		return errors.New("open test data file failed")
	}
	defer open.Close()
	decoder := yaml.NewYAMLOrJSONDecoder(open, maxFileSize)
	if err = decoder.Decode(&oriDevInfo1); err != nil {
		return errors.New("decode oriDevInfo1 failed")
	}
	if err = decoder.Decode(&expDeviceInfo1); err != nil {
		return errors.New("decode expDeviceInfo1 failed")
	}
	if err = decoder.Decode(&oriDevInfo2); err != nil {
		return errors.New("decode oriDevInfo2 failed")
	}
	if err = decoder.Decode(&expDeviceInfo2); err != nil {
		return errors.New("decode expDeviceInfo2 failed")
	}
	return nil
}
