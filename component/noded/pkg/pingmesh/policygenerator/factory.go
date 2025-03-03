/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package policygenerator is policy generator for pingmesh
*/

package policygenerator

import "nodeD/pkg/pingmesh/types"

// Factory is a factory for policy generator.
type Factory interface {
	Register(rule types.PingMeshRule, generator Interface) Factory
	Rule(rule types.PingMeshRule) Interface
}

type generatorFactoryImpl struct {
	generators map[types.PingMeshRule]Interface
}

// NewFactory returns a new factory.
func NewFactory() Factory {
	return &generatorFactoryImpl{generators: make(map[types.PingMeshRule]Interface)}
}

// Rule returns a generator.
func (gi *generatorFactoryImpl) Rule(rule types.PingMeshRule) Interface {
	return gi.generators[rule]
}

// Register registers a generator.
func (gi *generatorFactoryImpl) Register(rule types.PingMeshRule, generator Interface) Factory {
	gi.generators[rule] = generator
	return gi
}
