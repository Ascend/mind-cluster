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

// Package devmgr main test for pkg
package devmgr

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	ascommon "ascend-common/devmanager/common"
	"container-manager/pkg/common"
)

var (
	testErr    = errors.New("test error")
	mockDevMgr = &HwDevMgr{}
)

const (
	dev0 = 0
	dev1 = 1
	dev2 = 2
	dev3 = 3
	dev4 = 4
	dev5 = 5
	dev6 = 6
	dev7 = 7

	len8 = 8
)

func TestMain(m *testing.M) {
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
	resetDevMgr()
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

func resetDevMgr() {
	mockDevMgr = &HwDevMgr{
		devType:  api.Ascend910B,
		boardId:  0,
		devUsage: common.Infer,
		workMode: ascommon.AMPMode,
		npuInfos: map[int32]*common.NPUInfo{
			dev0: {PhyID: dev0, LogicID: dev0},
			dev1: {PhyID: dev1, LogicID: dev1},
			dev2: {PhyID: dev2, LogicID: dev2},
			dev3: {PhyID: dev3, LogicID: dev3},
			dev4: {PhyID: dev4, LogicID: dev4},
			dev5: {PhyID: dev5, LogicID: dev5},
			dev6: {PhyID: dev6, LogicID: dev6},
			dev7: {PhyID: dev7, LogicID: dev7},
		},
		dmgr: &devmanager.DeviceManagerMock{},
	}
}
