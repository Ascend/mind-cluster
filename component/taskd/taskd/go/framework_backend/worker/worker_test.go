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

// Package worker for taskd worker backend
package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/worker/monitor/profiling"
	"taskd/framework_backend/worker/om"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return err
	}
	return nil
}

func TestInitMonitor(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(utils.InitHwLog, func(logFileName string, ctx context.Context) error {
		return nil
	})

	patches.ApplyFunc(profiling.InitMspti, func() error {
		return nil
	})
	called := atomic.Bool{}
	patches.ApplyFunc(monitorInitNotify, func() {
		called.Store(true)
	})
	InitMonitor(context.Background(), 0, 0)
	convey.ShouldBeTrue(called.Load())
}

func TestInitNetwork(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	netTool = &net.NetInstance{}
	patches.ApplyFunc(net.InitNetwork, func(conf *common.TaskNetConfig) (*net.NetInstance, error) {
		return netTool, nil
	})
	called := atomic.Bool{}
	patches.ApplyFunc(netToolInitNotify, func() {
		called.Store(true)
	})
	InitNetwork(0, 0)
	convey.ShouldBeTrue(called.Load())
}

func TestRegisterAndLoopRecv(t *testing.T) {
	netTool = &net.NetInstance{}
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(netTool, "SyncSendMessage",
		func(nt *net.NetInstance, uuid, mtype, msgBody string, dst *common.Position) (*common.Ack, error) {
			return nil, nil
		})
	patches.ApplyMethod(netTool, "ReceiveMessage", func(nt *net.NetInstance) *common.Message {
		time.Sleep(time.Second)
		return &common.Message{
			Body: utils.ObjToString(storage.MsgBody{
				Code: constant.ProfilingAllOnCmdCode,
			}),
		}
	})
	patches.ApplyFunc(waitNetToolInit, func() bool {
		return true
	})
	called := false
	patches.ApplyFunc(profiling.ProcessMsg, func(globalRank int, msg *common.Message) {
		called = true
		return
	})
	patches.ApplyFunc(om.SwitchNicProcessMsg, func(msg *common.Message) {
		return
	})
	patches.ApplyFunc(om.StressTestProcessMsg, func(msg *common.Message) {
		return
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	registerAndLoopRecv(ctx)
	convey.ShouldBeTrue(called)
}
