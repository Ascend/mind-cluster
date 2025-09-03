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

// Package service for taskd manager backend service
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/manager/plugins/faultdig"
)

// Defines mock interface implementation
type MockManagerPlugin struct {
	mock.Mock
}

func (m *MockManagerPlugin) Capture() (string, error) {
	return m.Called().String(0), nil
}

func (m *MockManagerPlugin) Release() error {
	return nil
}

func (m *MockManagerPlugin) Name() string {
	return m.Called().String(0)
}

func (m *MockManagerPlugin) Handle() (infrastructure.HandleResult, error) {
	args := m.Called()
	return args.Get(0).(infrastructure.HandleResult), args.Error(1)
}

func (m *MockManagerPlugin) Predicate(snapshot storage.SnapShot) (infrastructure.PredicateResult, error) {
	args := m.Called(snapshot)
	return args.Get(0).(infrastructure.PredicateResult), args.Error(1)
}

func (m *MockManagerPlugin) PullMsg() ([]infrastructure.Msg, error) {
	args := m.Called()
	return args.Get(0).([]infrastructure.Msg), args.Error(1)
}

// Test cases
func TestNewPluginHandler(t *testing.T) {
	handler := NewPluginHandler()
	assert.NotNil(t, handler)
	assert.Empty(t, handler.Plugins)
}

func TestRegisterSuccess(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")

	err := handler.Register("test-plugin", plugin)
	assert.NoError(t, err)
	assert.Len(t, handler.Plugins, 1)
}

func TestRegisterDuplicate(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")

	// First registration succeeds
	err := handler.Register("test-plugin", plugin)
	assert.NoError(t, err)

	// Second registration fails
	err = handler.Register("test-plugin", plugin)
	assert.Error(t, err)
	assert.Equal(t, "register failed: plugin test-plugin has already register", err.Error())
}

func TestGetPluginFound(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")

	_ = handler.Register("test-plugin", plugin)
	retPlugin, err := handler.GetPlugin("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, plugin, retPlugin)
}

func TestGetPluginNotFound(t *testing.T) {
	handler := NewPluginHandler()
	plugin, err := handler.GetPlugin("non-existent")
	assert.Nil(t, plugin)
	assert.Error(t, err)
	assert.Equal(t, "can not find plugin non-existent", err.Error())
}

func TestHandleSuccess(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")
	expectedResult := infrastructure.HandleResult{Stage: "success"}
	plugin.On("Handle").Return(expectedResult, nil)

	_ = handler.Register("test-plugin", plugin)
	result, err := handler.Handle("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestHandlePluginNotFound(t *testing.T) {
	handler := NewPluginHandler()
	result, err := handler.Handle("non-existent")
	assert.Error(t, err)
	assert.Equal(t, infrastructure.HandleResult{}, result)
}

func TestHandlePluginError(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")
	plugin.On("Handle").Return(infrastructure.HandleResult{}, errors.New("handle error"))

	_ = handler.Register("test-plugin", plugin)
	result, err := handler.Handle("test-plugin")
	assert.Error(t, err)
	assert.Equal(t, "handle error", err.Error())
	assert.Equal(t, infrastructure.HandleResult{}, result)
}

func TestPredicateSuccess(t *testing.T) {
	handler := NewPluginHandler()
	snapshot := storage.SnapShot{}

	// Plugin1: normal return
	plugin1 := &MockManagerPlugin{}
	plugin1.On("Name").Return("plugin1")
	result1 := infrastructure.PredicateResult{CandidateStatus: "candidate"}
	plugin1.On("Predicate", snapshot).Return(result1, nil)

	// Plugin2: returns error (should be skipped)
	plugin2 := &MockManagerPlugin{}
	plugin2.On("Name").Return("plugin2")
	plugin2.On("Predicate", snapshot).Return(infrastructure.PredicateResult{}, errors.New("predicate error"))

	// Plugin3: normal return
	plugin3 := &MockManagerPlugin{}
	plugin3.On("Name").Return("plugin3")
	result3 := infrastructure.PredicateResult{CandidateStatus: ""}
	plugin3.On("Predicate", snapshot).Return(result3, nil)

	_ = handler.Register("plugin1", plugin1)
	_ = handler.Register("plugin2", plugin2)
	_ = handler.Register("plugin3", plugin3)

	results := handler.Predicate(&snapshot)
	assert.Len(t, results, 2) // Only two plugins return results (plugin2 error skipped)
	assert.Contains(t, results, result1)
	assert.Contains(t, results, result3)
}

func TestPullMsgSuccess(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")
	expectedMsgs := []infrastructure.Msg{{Receiver: []string{"worker1"}}}
	plugin.On("PullMsg").Return(expectedMsgs, nil)

	_ = handler.Register("test-plugin", plugin)
	msgs, err := handler.PullMsg("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgs, msgs)
}

func TestPullMsgPluginNotFound(t *testing.T) {
	handler := NewPluginHandler()
	msgs, err := handler.PullMsg("non-existent")
	assert.Error(t, err)
	assert.Nil(t, msgs)
}

func TestPullMsgPluginError(t *testing.T) {
	handler := NewPluginHandler()
	plugin := &MockManagerPlugin{}
	plugin.On("Name").Return("test-plugin")
	plugin.On("PullMsg").Return([]infrastructure.Msg{}, errors.New("pull error"))

	_ = handler.Register("test-plugin", plugin)
	msgs, err := handler.PullMsg("test-plugin")
	assert.Error(t, err)
	assert.Equal(t, []infrastructure.Msg{}, msgs)
}

func TestInitSuccess(t *testing.T) {
	handler := NewPluginHandler()
	err := handler.Init()
	assert.NoError(t, err)
	assert.NotEqual(t, len(handler.Plugins), 0) // Ensure example plugin is registered
}

func TestInitRegisterFailure(t *testing.T) {
	handler := NewPluginHandler()

	// Manually inject plugin with same name to trigger error
	profilingPlugin := faultdig.NewProfilingPlugin()
	handler.Plugins[profilingPlugin.Name()] = &MockManagerPlugin{}

	err := handler.Init()
	assert.Error(t, err)
}
