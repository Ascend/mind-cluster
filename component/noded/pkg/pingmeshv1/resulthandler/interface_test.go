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
Package resulthandler is using for handle hccsping mesh result
*/

package resulthandler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/pingmeshv1/types"
	_ "nodeD/pkg/testtool"
)

func TestHandle(t *testing.T) {
	convey.Convey("TestHandle", t, func() {
		var expected atomic.Int32
		convey.Convey("01-nil Stop chan will do nothing", func() {
			h := NewAggregatedHandler()
			h.Handle(nil)
			convey.So(expected.Load(), convey.ShouldEqual, 0)
		})
		convey.Convey("02-Stop chan will do nothing", func() {
			h := NewAggregatedHandler(func(*types.HccspingMeshResult) error {
				expected.Add(1)
				return nil
			})
			stopChan := make(chan struct{})
			go h.Handle(stopChan)
			h.Receive(&types.HccspingMeshResult{})
			time.Sleep(time.Second)
			convey.So(expected.Load(), convey.ShouldEqual, 1)
		})
	})
}
