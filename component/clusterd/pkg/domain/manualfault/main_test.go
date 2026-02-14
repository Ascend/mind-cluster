/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault main test for manual fault
package manualfault

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
)

var testErr = errors.New("test error")

const (
	node1 = "node1"
	node2 = "node2"
	node3 = "node3"

	dev1 = "dev1"
	dev2 = "dev2"
	dev3 = "dev3"

	code1 = "code1"
	code2 = "code2"

	job1 = "123"

	len0 = 0
	len1 = 1
	len2 = 2

	receiveTime0 = 1770969600000 // 2026-02-13 08:00:00
	receiveTime1 = 1771059600000 // 2026-02-14 09:00:00
	receiveTime2 = 1771059610000 // 2026-02-14 09:00:10
	receiveTime3 = 1771059620000 // 2026-02-14 09:00:20
	receiveTime4 = 1771059630000 // 2026-02-14 09:00:30
	receiveTime5 = 1771059640000 // 2026-02-14 09:00:40
	receiveTime6 = 1771149600000 // 2026-02-15 10:00:00
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
