/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package v1 is using for reconcile AscendJob.
package v1

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"ascend-operator/pkg/api/v1"
)

// TestOnOwnerCreateFuncForA5 test case for TestOnOwnerCreateFunc in A5
func TestOnOwnerCreateFuncForA5(t *testing.T) {
	convey.Convey("TestOnOwnerCreateFunc for A5", t, func() {
		r := newCommonReconciler()
		fn := r.onOwnerCreateFunc()
		convey.Convey("02-ascend job with scaleout-type=roce labels should return false", func() {
			job := newCommonAscendJob()
			job.Labels = map[string]string{v1.ScaleOutTypeLabel: "roce"}
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("03-ascend job with scaleout-type=uboe  labels should return false", func() {
			job := newCommonAscendJob()
			job.Labels = map[string]string{v1.ScaleOutTypeLabel: "uboe"}
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}
