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
	"fmt"

	"ascend-common/common-utils/hwlog"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/manager/plugins/faultdig"
	"taskd/framework_backend/manager/plugins/om"
)

// PluginHandlerInterface define the interface of plugin handler
type PluginHandlerInterface interface {
	Init() error
	GetPlugin(string) (infrastructure.ManagerPlugin, error)
	Register(string, infrastructure.ManagerPlugin) error
	Handle(string) (infrastructure.HandleResult, error)
	Predicate(*storage.SnapShot) []infrastructure.PredicateResult
	PullMsg(string) ([]infrastructure.Msg, error)
}

// PluginHandler is defined to handle plugins operation
type PluginHandler struct {
	Plugins map[string]infrastructure.ManagerPlugin
}

// Init register all plugin
func (p *PluginHandler) Init() error {
	pluginList := []infrastructure.ManagerPlugin{
		faultdig.NewProfilingPlugin(),
		om.NewOmSwitchNicPlugin(),
		om.NewOmStressTestPlugin(),
	}
	for _, plugin := range pluginList {
		if err := p.Register(plugin.Name(), plugin); err != nil {
			hwlog.RunLog.Errorf("register plugin %s failed!", plugin.Name())
			return fmt.Errorf("register plugin %s failed", plugin.Name())
		}
	}
	return nil
}

// NewPluginHandler return a plugin handler instance
func NewPluginHandler() *PluginHandler {
	return &PluginHandler{
		Plugins: make(map[string]infrastructure.ManagerPlugin, 0),
	}
}

// GetPlugin return a plugin instance
func (p *PluginHandler) GetPlugin(pluginName string) (infrastructure.ManagerPlugin, error) {
	plugin, ok := p.Plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("can not find plugin %s", pluginName)
	}
	return plugin, nil
}

// Register register a plugin in handler
func (p *PluginHandler) Register(pluginName string, plugin infrastructure.ManagerPlugin) error {
	if _, ok := p.Plugins[pluginName]; ok {
		return fmt.Errorf("register failed: plugin %s has already register", pluginName)
	}
	p.Plugins[pluginName] = plugin
	return nil
}

// Handle execute the handle function of plugin
func (p *PluginHandler) Handle(pluginName string) (infrastructure.HandleResult, error) {
	var result infrastructure.HandleResult
	plugin, err := p.GetPlugin(pluginName)
	if err != nil {
		return result, err
	}
	result, err = plugin.Handle()
	if err != nil {
		return result, err
	}
	return result, nil
}

// Predicate execute the predicate function of all registered plugin
func (p *PluginHandler) Predicate(snapshot *storage.SnapShot) []infrastructure.PredicateResult {
	var predicateResults []infrastructure.PredicateResult
	for _, plugin := range p.Plugins {
		result, err := plugin.Predicate(*snapshot)
		if err != nil {
			continue
		}
		predicateResults = append(predicateResults, result)
	}
	return predicateResults
}

// PullMsg execute the PullMsg function of plugin
func (p *PluginHandler) PullMsg(pluginName string) ([]infrastructure.Msg, error) {
	var pullMsg []infrastructure.Msg
	plugin, err := p.GetPlugin(pluginName)
	if err != nil {
		return pullMsg, err
	}
	pullMsg, err = plugin.PullMsg()
	if err != nil {
		return pullMsg, err
	}
	return pullMsg, nil
}
