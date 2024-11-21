// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"errors"
	"fmt"
)

type handleFunc func() (nextEvent string, code RespCode, err error)

// TransRule is type of state change rules
type TransRule struct {
	Src     string
	Event   string
	Dst     string
	Handler handleFunc
}

/*
StateMachine is state machine.
Notice:

	this state machine implement is thread unsafe,
	you should handle data race issues in concurrent scene on your own
*/

// StateMachine is type of fsm
type StateMachine struct {
	state     string
	initState string
	rules     []TransRule
	path      []string
	pathGraph string // src(event)-->dst
}

// NewStateMachine return a new state machine
func NewStateMachine(initState string, rules []TransRule) *StateMachine {
	return &StateMachine{
		state:     initState,
		initState: initState,
		rules:     rules,
		pathGraph: initState,
		path:      []string{},
	}
}

func (m *StateMachine) changeState(state string) {
	m.state = state
}

// GetState return current state
func (m *StateMachine) GetState() string {
	return m.state
}

// Reset reset state machine
func (m *StateMachine) Reset() {
	m.state = m.initState
	m.path = []string{m.initState}
	m.pathGraph = m.initState
}

// RuleMatching return rule for event when origin state is src
func (m *StateMachine) ruleMatching(src, event string) (bool, *TransRule) {
	for _, v := range m.rules {
		if v.Src == src && v.Event == event {
			return true, &v
		}
	}
	return false, nil
}

// RuleCheck return rule for event when origin state is src
func (m *StateMachine) RuleCheck(src, event string) bool {
	for _, v := range m.rules {
		if v.Src == src && v.Event == event {
			return true
		}
	}
	return false
}

// GetPathGraph return state change path
func (m *StateMachine) GetPathGraph() string {
	return m.pathGraph
}

// Trigger trigger state change by event
func (m *StateMachine) Trigger(event string, args ...interface{}) (string, RespCode, error) {
	matching, rule := m.ruleMatching(m.state, event)
	if !matching {
		m.pathGraph = fmt.Sprintf("%s(%s)-->orderMixed", m.pathGraph, event)
		return "", OrderMix, errors.New("rule match error, change order may mixed")
	}
	m.path = append(m.path, rule.Dst)
	m.pathGraph = fmt.Sprintf("%s(%s)-->%s", m.pathGraph, event, rule.Dst)
	m.changeState(rule.Dst)
	return rule.Handler()
}
