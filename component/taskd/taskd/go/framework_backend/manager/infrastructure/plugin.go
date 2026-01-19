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

// Package infrastructure for taskd manager backend infrastructure
package infrastructure

import "taskd/framework_backend/manager/infrastructure/storage"

// PredicateResult indicate predicate result from plugin
type PredicateResult struct {
	// PluginName indicate plugin name
	PluginName string // Name of the plugin that generated this result
	// CandidateStatus indicate is candidate for predicate stream
	CandidateStatus string
	// PredicateStream indicate the predicate stream of the stream
	PredicateStream map[string]string
}

// HandleResult indicate handle result from plugin
type HandleResult struct {
	// Status indicate plugin handle status
	Status string
	// Stage indicate plugin handle stage
	Stage string
	// ErrorMsg indicate plugin error message
	ErrorMsg string
}

// Msg defines the message from plugin to manager
type Msg struct {
	// Receiver indicate all message receives
	Receiver []string
	// Body indicate the message body
	Body storage.MsgBody
}

// ManagerPlugin defines the interface for management plugins
type ManagerPlugin interface {
	Name() string
	Predicate(shot storage.SnapShot) (PredicateResult, error)
	Release() error
	Handle() (HandleResult, error)
	PullMsg() ([]Msg, error)
}
