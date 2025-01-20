// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func event1Handler() (string, RespCode, error) {
	return "event2", OK, nil
}

func event2Handler() (string, RespCode, error) {
	return "", OK, nil
}

func getFakeRules() []TransRule {
	return []TransRule{
		{InitState, "event1", "state1", event1Handler},
		{"state1", "event2", InitState, event2Handler},
	}
}

func TestNewStateMachine(t *testing.T) {
	convey.Convey("Test TestNewStateMachine", t, func() {
		sm := NewStateMachine(InitState, nil)
		convey.So(sm.rules, convey.ShouldBeNil)
		convey.So(sm.pathGraph, convey.ShouldEqual, InitState)
		convey.So(sm.initState, convey.ShouldEqual, InitState)
		convey.So(len(sm.path), convey.ShouldEqual, 0)
	})
}

func TestGetPathGraph(t *testing.T) {
	convey.Convey("Test TestGetPathGraph", t, func() {
		sm := NewStateMachine(InitState, nil)
		graph := sm.GetPathGraph()
		convey.So(graph, convey.ShouldEqual, InitState)
	})
}

func TestGetState(t *testing.T) {
	convey.Convey("Test TestGetState", t, func() {
		sm := NewStateMachine(InitState, nil)
		state := sm.GetState()
		convey.So(state, convey.ShouldEqual, InitState)
	})
}

func TestReset(t *testing.T) {
	convey.Convey("Test TestReset", t, func() {
		sm := NewStateMachine(InitState, nil)
		sm.Reset()
		convey.So(sm.rules, convey.ShouldBeNil)
		convey.So(sm.pathGraph, convey.ShouldEqual, InitState)
		convey.So(sm.initState, convey.ShouldEqual, InitState)
		convey.So(len(sm.path), convey.ShouldEqual, 1)
	})
}

func TestRuleCheck(t *testing.T) {
	convey.Convey("Test TestRuleCheck", t, func() {
		sm := NewStateMachine(InitState, getFakeRules())
		convey.Convey("rule match case", func() {
			match, _ := sm.ruleMatching(InitState, "event1")
			convey.So(match, convey.ShouldBeTrue)
		})
		convey.Convey("rule not match case", func() {
			match, _ := sm.ruleMatching(InitState, "event2")
			convey.So(match, convey.ShouldBeFalse)
		})
	})
}

func TestTrigger(t *testing.T) {
	convey.Convey("Test TestTrigger", t, func() {
		convey.Convey("rule match for trigger", func() {
			sm := NewStateMachine(InitState, getFakeRules())
			nextEvent, code, err := sm.Trigger("event1")
			convey.So(nextEvent, convey.ShouldEqual, "event2")
			convey.So(sm.GetPathGraph(), convey.ShouldEqual,
				fmt.Sprintf("%s(%s)-->%s", InitState, "event1", "state1"))
			convey.So(code, convey.ShouldEqual, OK)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("rule not match for trigger", func() {
			sm := NewStateMachine(InitState, getFakeRules())
			nextEvent, code, err := sm.Trigger("event2")
			convey.So(nextEvent, convey.ShouldEqual, "")
			convey.So(sm.GetPathGraph(), convey.ShouldEqual,
				fmt.Sprintf("%s(%s)-->orderMixed", InitState, "event2"))
			convey.So(code, convey.ShouldEqual, OrderMix)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
