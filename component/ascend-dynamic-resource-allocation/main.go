/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	draDriver "ascend-dynamic-resource-allocation/internal/driver"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
)

func main() {
	draConfig := draFlags.NewDraConfig()
	draConfig.RegisterFlags()
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := draConfig.HwLogConfig.InitLogModule(ctx); err != nil {
		fmt.Printf("init log module failed: %v\n", err)
		return
	}
	hwlog.RunLog.Info("init log module successfully.")

	if err := draConfig.DraOption.Validate(); err != nil {
		hwlog.RunLog.Errorf("validate dra option failed: %v", err)
		return
	}
	hwlog.RunLog.Info("validate dra option successfully.")

	dmgr, err := devmanager.AutoInit("", draConfig.DraOption.DeviceResetTimeout)
	if err != nil {
		hwlog.RunLog.Errorf("auto-init device manager failed: %v", err)
		return
	}
	hwlog.RunLog.Info("auto-init device manager successfully.")

	ctx, cancel := context.WithCancelCause(ctx)
	ascendDraManager, err := draDriver.NewAscendDraManager(cancel, dmgr, draConfig)
	if err != nil {
		hwlog.RunLog.Errorf("new ascend dra manager failed: %v", err)
		return
	}
	hwlog.RunLog.Info("new ascend dra manager successfully.")

	if err := ascendDraManager.Start(ctx); err != nil {
		hwlog.RunLog.Errorf("start ascend dra manager failed: %v", err)
		return
	}
	hwlog.RunLog.Info("ascend dra manager started successfully.")

	<-ctx.Done()
	stop()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		hwlog.RunLog.Errorf("error from context: %v", err)
	}

	ascendDraManager.Stop()
	return
}
