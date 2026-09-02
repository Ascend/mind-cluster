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

package app

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/coordinator/proto"
)

func TestCoordinatorWorkflow(t *testing.T) {
	convey.Convey("test NewCoordinator and Name", t, testNewCoordinatorAndName)
	convey.Convey("test Init leader role", t, testInitLeaderRole)
	convey.Convey("test Init ordinary role", t, testInitOrdinaryRole)
	convey.Convey("test Init error paths", t, testInitErrorPaths)
	convey.Convey("test Work", t, testWork)
	convey.Convey("test ShutDown", t, testShutDown)
	convey.Convey("test enabled", t, testEnabled)
}

func resetParamOption() {
	common.ParamOption = common.Option{}
}

func testNewCoordinatorAndName() {
	resetParamOption()
	c := newMockCoordinator()
	convey.So(c, convey.ShouldNotBeNil)
	convey.So(c.ops, convey.ShouldNotBeNil)
	convey.So(c.containerStore, convey.ShouldNotBeNil)
	convey.So(c.Name(), convey.ShouldEqual, "coordinator")
	convey.So(c.cancel, convey.ShouldNotBeNil)
}

func testInitLeaderRole() {
	resetParamOption()
	common.ParamOption.ListenAddr = testListen
	common.ParamOption.LocalNodeID = testNodeID

	mockServer := &serverEndpoint{streams: map[string]*serverStream{}, acks: map[string]chan *proto.Response{}}
	var p1 = gomonkey.ApplyFuncReturn(newServerEndpoint, mockServer, nil)
	defer p1.Reset()

	c := newMockCoordinator()
	err := c.Init()
	convey.So(err, convey.ShouldBeNil)
	convey.So(c.server, convey.ShouldEqual, mockServer)
}

func testInitOrdinaryRole() {
	resetParamOption()
	common.ParamOption.LeaderAddrs = []string{testLeaderA}
	common.ParamOption.LocalNodeID = testNodeID

	mockClient := &clientEndpoint{entries: map[string]*leaderEntry{}}
	var p1 = gomonkey.ApplyFuncReturn(newClientEndpoint, mockClient, nil)
	defer p1.Reset()

	c := newMockCoordinator()
	err := c.Init()
	convey.So(err, convey.ShouldBeNil)
	convey.So(c.client, convey.ShouldEqual, mockClient)
}

func testInitErrorPaths() {
	resetParamOption()
	convey.Convey("newServerEndpoint fails", func() {
		common.ParamOption.ListenAddr = testListen
		var p1 = gomonkey.ApplyFuncReturn(newServerEndpoint, nil, testErr)
		defer p1.Reset()
		c := newMockCoordinator()
		err := c.Init()
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("ordinary role without local node id", func() {
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		common.ParamOption.LocalNodeID = ""
		c := newMockCoordinator()
		err := c.Init()
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("newClientEndpoint fails is non-fatal", func() {
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		common.ParamOption.LocalNodeID = testNodeID
		var p1 = gomonkey.ApplyFuncReturn(newClientEndpoint, nil, testErr)
		defer p1.Reset()
		c := newMockCoordinator()
		err := c.Init()
		convey.So(err, convey.ShouldBeNil)
		convey.So(c.client, convey.ShouldBeNil)
	})

	convey.Convey("no role configured", func() {
		c := newMockCoordinator()
		err := c.Init()
		convey.So(err, convey.ShouldBeNil)
	})
}

func testWork() {
	convey.Convey("disabled: returns without starting loop", func() {
		resetParamOption()
		var started bool
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "dataSyncLoop",
			func(_ *Coordinator, _ context.Context) { started = true })
		defer p1.Reset()
		c := newMockCoordinator()
		c.Work(context.Background())
		convey.So(started, convey.ShouldBeFalse)
	})

	convey.Convey("enabled: starts dataSyncLoop", func() {
		resetParamOption()
		common.ParamOption.LeaderAddrs = []string{testLeaderA}
		common.ParamOption.LocalNodeID = testNodeID
		var started bool
		var p1 = gomonkey.ApplyPrivateMethod(&Coordinator{}, "dataSyncLoop",
			func(_ *Coordinator, _ context.Context) { started = true })
		defer p1.Reset()
		c := newMockCoordinator()
		c.Work(context.Background())
		c.wg.Wait()
		convey.So(started, convey.ShouldBeTrue)
	})
}

func testShutDown() {
	resetParamOption()
	c := newMockCoordinator()
	mockServer := newMockServerEndpoint(c)
	mockClient := &clientEndpoint{entries: map[string]*leaderEntry{}}
	c.server = mockServer
	c.client = mockClient
	c.ShutDown()
	convey.So(c.server.server, convey.ShouldBeNil)
}

func testEnabled() {
	resetParamOption()
	c := newMockCoordinator()
	convey.So(c.enabled(), convey.ShouldBeFalse)
	common.ParamOption.LeaderAddrs = []string{testLeaderA}
	convey.So(c.enabled(), convey.ShouldBeTrue)
}
