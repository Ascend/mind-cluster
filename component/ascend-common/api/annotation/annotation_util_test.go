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

// Package annotation defines NPU node annotator methods
package annotation

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	annoKeyNew     = "huawei.com/npu.new"
	annoKeyOld     = "huawei.com/npu.old"
	annoValue1     = "value1"
	annoValue2     = " value2 "
	annoValue2Trim = "value2"
)

func TestGetAnnotationValue(t *testing.T) {
	convey.Convey("should return empty when annotations is nil", t, func() {
		val, ok := GetAnnotationValue(nil, annoKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when key not found", t, func() {
		annotations := map[string]string{"other": "val"}
		val, ok := GetAnnotationValue(annotations, annoKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when value is empty string", t, func() {
		annotations := map[string]string{annoKeyNew: ""}
		val, ok := GetAnnotationValue(annotations, annoKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return first key value when found", t, func() {
		annotations := map[string]string{annoKeyNew: annoValue1, annoKeyOld: annoValue2}
		val, ok := GetAnnotationValue(annotations, annoKeyNew, annoKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, annoValue1)
	})
	convey.Convey("should fallback to second key when first not found", t, func() {
		annotations := map[string]string{annoKeyOld: annoValue1}
		val, ok := GetAnnotationValue(annotations, annoKeyNew, annoKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, annoValue1)
	})
	convey.Convey("should trim spaces from value", t, func() {
		annotations := map[string]string{annoKeyNew: annoValue2}
		val, ok := GetAnnotationValue(annotations, annoKeyNew)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, annoValue2Trim)
	})
}

func TestGetNodeAnnotation(t *testing.T) {
	convey.Convey("should return empty when node is nil", t, func() {
		val, ok := GetNodeAnnotation(nil, annoKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when node annotations is nil", t, func() {
		node := &v1.Node{}
		val, ok := GetNodeAnnotation(node, annoKeyNew)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(val, convey.ShouldBeEmpty)
	})
	convey.Convey("should return value when key found in node annotations", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{annoKeyNew: annoValue1},
			},
		}
		val, ok := GetNodeAnnotation(node, annoKeyNew)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, annoValue1)
	})
	convey.Convey("should fallback to second key in node annotations", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{annoKeyOld: annoValue1},
			},
		}
		val, ok := GetNodeAnnotation(node, annoKeyNew, annoKeyOld)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(val, convey.ShouldEqual, annoValue1)
	})
}
