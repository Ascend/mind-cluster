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

// Package app main test for pkg
package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"container-manager/pkg/common"
	"container-manager/pkg/devmgr"
)

const (
	testFilePath = "./testCfg.json"
	mode644      = 0644

	len0 = 0
	len1 = 1
	len2 = 2
	len3 = 3
	len4 = 4

	devId0   = 0
	devId1   = 1
	eventId0 = 0x123
	eventId1 = 0x456
	eventId2 = 0x789
	eventId3 = 0xabc

	moduleId0      = 0
	moduleId1      = 1
	mockModuleKey0 = "0000"
	mockModuleKey1 = "1111"
	mockModuleKey2 = "2222"

	invalidQueueLen      = 100000
	faultExistedDuration = 301
)

var (
	testErr = errors.New("test error")

	mockFaultMgr = &FaultMgr{}
	mockDevMgr   = &devmgr.HwDevMgr{}
	newItem1     = common.DevFaultInfo{EventID: eventId0}
	newItem2     = common.DevFaultInfo{EventID: eventId1}
	newItem3     = common.DevFaultInfo{EventID: eventId2}
	newItem4     = common.DevFaultInfo{EventID: eventId3}
	newItem5     = common.DevFaultInfo{EventID: eventId3, FaultLevel: common.NormalNPU}
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
	mockFaultMgr = NewFaultMgr()
	mockDevMgr.SetDmgr(&devmanager.DeviceManagerMock{})
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

func resetQueueCache() {
	QueueCache = &FaultQueue{
		faults: make([]common.DevFaultInfo, 0),
		mutex:  sync.Mutex{},
	}
}
