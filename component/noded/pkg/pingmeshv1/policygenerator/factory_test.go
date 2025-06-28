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
Package policygenerator is policy generator for pingmeshv1
*/

package policygenerator

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/pingmeshv1/policygenerator/fullmesh"
	_ "nodeD/pkg/testtool"
)

func TestRule(t *testing.T) {
	convey.Convey("TestRule", t, func() {
		factory := NewFactory()
		convey.Convey("01-register and get", func() {
			factory.Register(fullmesh.Rule, fullmesh.New("node"))
			gen := factory.Rule(fullmesh.Rule)
			convey.So(gen, convey.ShouldNotBeNil)
		})
	})
}
