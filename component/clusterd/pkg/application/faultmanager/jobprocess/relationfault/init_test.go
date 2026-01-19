// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package relationfault contain relation fault process
package relationfault

import (
	"testing"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	initConfig()
	m.Run()
}

func initConfig() {
	relationFilePath := "../../../../../build/relationFaultCustomization.json"
	if fileBytes := LoadConfigFromFile(relationFilePath); fileBytes != nil {
		initRelationFaultStrategies(fileBytes)
		initRelationFaultCodesMap()
	}
	durationFilePath := "../../../../../build/faultDuration.json"
	if fileBytes := LoadConfigFromFile(durationFilePath); fileBytes != nil {
		initFaultDuration(fileBytes)
		initFaultCodeTimeOutMap()
	}
}
