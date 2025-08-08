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

// Package slownode is a DT collection for func in starter
package slownode

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-faultdiag-online/pkg/core/model/enum"
)

func TestStartSlowNode(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(registerHandlers, func(func(any) bool, func(any, any, watch.EventType)) {})
	patches.ApplyFunc(initCMInformer, func() {})

	convey.Convey("Test StartSlowNode", t, func() {
		StartSlowNode(enum.Cluster)
		StartSlowNode(enum.Node)
		StartSlowNode("wrong-target")
	})
}
