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

// Package command test for run command
package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	app2 "container-manager/pkg/container/app"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/app"
	"container-manager/pkg/workflow"
)

func TestRunCmdBasicMethods(t *testing.T) {
	convey.Convey("test cmd 'run' basic methods", t, func() {
		cmd := RunCmd()
		convey.So(cmd.Name(), convey.ShouldEqual, "run")
		convey.So(cmd.Description(), convey.ShouldEqual, "Run container-manager")
		err := cmd.InitLog(context.Background())
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestRunCmdCheckParam(t *testing.T) {
	convey.Convey("test cmd 'status' methods CheckParam, param valid", t, testValidParam)
	convey.Convey("test cmd 'status' methods CheckParam, invalid ctrStrategy", t, testErrCtrStrategy)
	convey.Convey("test cmd 'status' methods CheckParam, invalid sockPath", t, testErrSockPath)
	convey.Convey("test cmd 'status' methods CheckParam, invalid runtimeType", t, testErrRuntimeType)
}

func testValidParam() {
	var p1 = gomonkey.ApplyFuncReturn(utils.IsExist, true).ApplyFuncReturn(utils.CheckPath, nil, nil)
	defer p1.Reset()
	stCmd := runCmd{
		ctrStrategy: common.NeverStrategy,
		sockPath:    defaultSockPath,
		runtimeType: common.ContainerDType,
	}
	convey.So(stCmd.CheckParam(), convey.ShouldBeNil)
}

func testErrCtrStrategy() {
	var p1 = gomonkey.ApplyFuncReturn(utils.IsExist, true).ApplyFuncReturn(utils.CheckPath, nil, nil)
	defer p1.Reset()
	stCmd := runCmd{
		ctrStrategy: "invalid ctrStrategy",
		sockPath:    defaultSockPath,
		runtimeType: common.ContainerDType,
	}
	err := stCmd.CheckParam()
	expErr := fmt.Errorf("invalid ctrStrategy, should be between %s, %s and %s",
		common.NeverStrategy, common.SingleStrategy, common.RingStrategy)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testErrSockPath() {
	stCmd := runCmd{
		ctrStrategy: common.NeverStrategy,
		sockPath:    defaultSockPath,
		runtimeType: common.ContainerDType,
	}
	var p2 = gomonkey.ApplyFuncReturn(utils.IsExist, false)
	err := stCmd.CheckParam()
	expErr := errors.New("socket file not exist")
	convey.So(err, convey.ShouldResemble, expErr)
	p2.Reset()

	var p3 = gomonkey.ApplyFuncReturn(utils.IsExist, true).ApplyFuncReturn(utils.CheckPath, nil, testErr)
	err = stCmd.CheckParam()
	expErr = fmt.Errorf("invalid sockPath, %v", testErr)
	convey.So(err, convey.ShouldResemble, expErr)
	p3.Reset()
}

func testErrRuntimeType() {
	var p1 = gomonkey.ApplyFuncReturn(utils.IsExist, true).ApplyFuncReturn(utils.CheckPath, nil, nil)
	defer p1.Reset()
	stCmd := runCmd{
		ctrStrategy: common.NeverStrategy,
		sockPath:    defaultSockPath,
		runtimeType: "invalid runtimeType",
	}
	err := stCmd.CheckParam()
	expErr := fmt.Errorf("invalid runtimeType, should be between %s and %s", common.DockerType, common.ContainerDType)
	convey.So(err, convey.ShouldResemble, expErr)
}

func TestRunCmdExecute(t *testing.T) {
	cmd := RunCmd()
	cmd.BindFlag()
	if err := flag.Set("runtimeMode", common.DockerType); err != nil {
		t.Errorf("set flag err: %v", err)
	}
	if err := flag.Set("sockPath", defaultSockPath); err != nil {
		t.Errorf("set flag err: %v", err)
	}
	if err := flag.Set("ctrStrategy", common.SingleStrategy); err != nil {
		t.Errorf("set flag err: %v", err)
	}
	flag.Parse()

	var patches = gomonkey.ApplyMethodReturn(&workflow.ModuleMgr{}, "Init", nil).
		ApplyMethod(&workflow.ModuleMgr{}, "Work", func(_ *workflow.ModuleMgr, _ context.Context) {}).
		ApplyMethod(&workflow.ModuleMgr{}, "ShutDown", func(_ *workflow.ModuleMgr) {})
	defer patches.Reset()
	convey.Convey("test method 'Run' success", t, func() {
		p1 := gomonkey.ApplyFuncReturn(devmgr.NewHwDevMgr, nil).
			ApplyFuncReturn(app.NewFaultMgr, nil).
			ApplyFuncReturn(app2.NewCtrCtl, &app2.CtrCtl{}, nil)
		defer p1.Reset()
		err := cmd.Execute(context.Background())
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test method 'Run' failed, new hwdevmgr failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(devmgr.NewHwDevMgr, testErr)
		defer p1.Reset()
		err := cmd.Execute(context.Background())
		expErr := errors.New("new dev manager failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test method 'Run' failed, new ctrmgr failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(devmgr.NewHwDevMgr, nil).
			ApplyFuncReturn(app2.NewCtrCtl, nil, testErr)
		defer p1.Reset()
		err := cmd.Execute(context.Background())
		expErr := errors.New("new container controller failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
