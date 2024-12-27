// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"errors"
	"fmt"
	"sync"
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
	lock      sync.RWMutex
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
	m.lock.Lock()
	defer m.lock.Unlock()
	m.state = state
}

// GetState return current state
func (m *StateMachine) GetState() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.state
}

// Reset reset state machine
func (m *StateMachine) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.state = m.initState
	m.path = []string{m.initState}
	m.pathGraph = m.initState
}

// RuleMatching return rule for event when origin state is src
func (m *StateMachine) ruleMatching(src, event string) (bool, *TransRule) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, v := range m.rules {
		if v.Src == src && v.Event == event {
			return true, &v
		}
	}
	return false, nil
}

// RuleCheck return rule for event when origin state is src
func (m *StateMachine) RuleCheck(src, event string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, v := range m.rules {
		if v.Src == src && v.Event == event {
			return true
		}
	}
	return false
}

// GetPathGraph return state change path
func (m *StateMachine) GetPathGraph() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.pathGraph
}

func (m *StateMachine) appendPath(match bool, event, dst string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if match {
		m.path = append(m.path, dst)
		m.pathGraph = fmt.Sprintf("%s(%s)-->%s", m.pathGraph, event, dst)
		return
	}
	m.pathGraph = fmt.Sprintf("%s(%s)-->orderMixed", m.pathGraph, event)
}

// Trigger state change by event
func (m *StateMachine) Trigger(event string, args ...interface{}) (string, RespCode, error) {
	matching, rule := m.ruleMatching(m.state, event)
	dstState := ""
	if matching {
		dstState = rule.Dst
	}
	m.appendPath(matching, event, dstState)
	if !matching {
		return "", OrderMix, errors.New("rule match error, change order may mixed")
	}
	m.changeState(rule.Dst)
	return rule.Handler()
}
