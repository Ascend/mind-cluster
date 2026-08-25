/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package label defines NPU node labeler interfaces.
package label

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelKeyNew     = "huawei.com/npu.new"
	labelKeyOld     = "huawei.com/npu.old"
	labelValue1     = "val1"
	labelValue2     = " val2 "
	labelValue2Trim = "val2"
)

func TestGetLabelValue(t *testing.T) {
	convey.Convey("should return empty when labels is nil", t, func() {
		val, ok := GetLabelValue(nil, labelKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when key not found", t, func() {
		labels := map[string]string{"other": "val"}
		val, ok := GetLabelValue(labels, labelKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when value is empty string", t, func() {
		labels := map[string]string{labelKeyNew: ""}
		val, ok := GetLabelValue(labels, labelKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return first key value when found", t, func() {
		labels := map[string]string{labelKeyNew: labelValue1, labelKeyOld: labelValue2}
		val, ok := GetLabelValue(labels, labelKeyNew, labelKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, labelValue1)
	})
	convey.Convey("should fallback to second key when first not found", t, func() {
		labels := map[string]string{labelKeyOld: labelValue1}
		val, ok := GetLabelValue(labels, labelKeyNew, labelKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, labelValue1)
	})
	convey.Convey("should trim spaces from value", t, func() {
		labels := map[string]string{labelKeyNew: labelValue2}
		val, ok := GetLabelValue(labels, labelKeyNew)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, labelValue2Trim)
	})
}

func TestGetNodeLabel(t *testing.T) {
	convey.Convey("should return empty when node is nil", t, func() {
		val, ok := GetNodeLabel(nil, labelKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when node labels is nil", t, func() {
		node := &v1.Node{}
		val, ok := GetNodeLabel(node, labelKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return value when key found in node labels", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{labelKeyNew: labelValue1},
			},
		}
		val, ok := GetNodeLabel(node, labelKeyNew)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, labelValue1)
	})
	convey.Convey("should fallback to second key in node labels", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{labelKeyOld: labelValue1},
			},
		}
		val, ok := GetNodeLabel(node, labelKeyNew, labelKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, labelValue1)
	})
}
