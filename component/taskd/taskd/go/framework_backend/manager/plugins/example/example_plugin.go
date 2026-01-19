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

// Package example for taskd manager plugin
package example

import (
	"ascend-common/common-utils/hwlog"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	// ExampleName indicate the example plugin name
	ExampleName = "example"
)

// ExamplePlugin define plugin example
type ExamplePlugin struct {
}

// NewExamplePlugin return a example plugin instance
func NewExamplePlugin() infrastructure.ManagerPlugin {
	return &ExamplePlugin{}
}

// Name return the plugin name
func (e *ExamplePlugin) Name() string {
	return ExampleName
}

// Register register plugin to manager
func (e *ExamplePlugin) Register() (string, error) {
	hwlog.RunLog.Infof("register example plugin success")
	return "", nil
}

// Predicate return the stream request
func (e *ExamplePlugin) Predicate(snapShot storage.SnapShot) (infrastructure.PredicateResult, error) {
	return infrastructure.PredicateResult{}, nil
}

// Release give up token in a stream
func (e *ExamplePlugin) Release() error {
	return nil
}

// Handle business process
func (e *ExamplePlugin) Handle() (infrastructure.HandleResult, error) {
	return infrastructure.HandleResult{}, nil
}

// PullMsg pull msg to others
func (e *ExamplePlugin) PullMsg() ([]infrastructure.Msg, error) {
	return []infrastructure.Msg{}, nil
}
