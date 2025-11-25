/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package domain main test for pkg
package domain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
)

const (
	len0 = 0
	len1 = 1
	len2 = 2
	len3 = 3

	devId0   = 0
	devId1   = 1
	eventId0 = 0x123
	eventId1 = 0x456

	moduleId0 = 0
	moduleId1 = 1
	moduleId2 = 2
)

var (
	testErr = errors.New("test error")

	// dev0, event0, module0, occur
	mockFault1 = &common.DevFaultInfo{
		EventID:       eventId0,
		PhyID:         devId0,
		ModuleType:    moduleId0,
		ModuleID:      moduleId0,
		SubModuleType: moduleId0,
		SubModuleID:   moduleId0,
		Assertion:     common.FaultOccur,
	}
	// dev0, event1, module0, occur
	mockFault2 = &common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId0,
		ModuleType:    moduleId0,
		ModuleID:      moduleId0,
		SubModuleType: moduleId0,
		SubModuleID:   moduleId0,
		Assertion:     common.FaultOccur,
	}
	// dev0, event1, module1, occur
	mockFault3 = &common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId0,
		ModuleType:    moduleId1,
		ModuleID:      moduleId1,
		SubModuleType: moduleId1,
		SubModuleID:   moduleId1,
		Assertion:     common.FaultOccur,
	}
	// dev1, event1, module2, occur
	mockFault4 = &common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId1,
		ModuleType:    moduleId2,
		ModuleID:      moduleId2,
		SubModuleType: moduleId2,
		SubModuleID:   moduleId2,
		Assertion:     common.FaultOccur,
	}
	// dev0, event1, module1, recover
	mockFault5 = &common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId0,
		ModuleType:    moduleId1,
		ModuleID:      moduleId1,
		SubModuleType: moduleId1,
		SubModuleID:   moduleId1,
		Assertion:     common.FaultRecover,
	}

	mockFault6 = &common.DevFaultInfo{
		EventID:       eventId1,
		PhyID:         devId1,
		ModuleType:    moduleId2,
		ModuleID:      moduleId2,
		SubModuleType: moduleId2,
		SubModuleID:   moduleId2,
		Assertion:     common.FaultRecover,
	}
	mockFault7 = &common.DevFaultInfo{
		Assertion: common.FaultOnce,
	}
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	teardown()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	mockFaultCache = GetFaultCache()
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

func teardown() {
	err := os.Remove(testFilePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		fmt.Printf("remove file %s failed, %v\n", testFilePath, err)
	}
}
