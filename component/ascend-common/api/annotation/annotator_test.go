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
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api/label"
)

const (
	annoKey1 = "test-key-1"
	annoKey2 = "test-key-2"
	annoVal1 = "test-val-1"
	annoVal2 = "test-val-2"
)

type mockAnnotator struct {
	writeFn func(annotations map[string]string, ctx *label.NodeContext) error
}

func (m *mockAnnotator) Write(annotations map[string]string, ctx *label.NodeContext) error {
	return m.writeFn(annotations, ctx)
}

func TestNewAnnotationGroup(t *testing.T) {
	convey.Convey("should create annotation group with annotators", t, func() {
		mock := &mockAnnotator{}
		group := NewAnnotationGroup(mock)
		convey.So(group, convey.ShouldNotBeNil)
	})
	convey.Convey("should create annotation group with multiple annotators", t, func() {
		group := NewAnnotationGroup(&mockAnnotator{}, &mockAnnotator{})
		convey.So(group, convey.ShouldNotBeNil)
	})
	convey.Convey("should create annotation group with no annotators", t, func() {
		group := NewAnnotationGroup()
		convey.So(group, convey.ShouldNotBeNil)
	})
}

func TestGroup_WriteAll(t *testing.T) {
	convey.Convey("should collect annotations from all annotators", t, func() {
		a1 := &mockAnnotator{
			writeFn: func(annotations map[string]string, ctx *label.NodeContext) error {
				annotations[annoKey1] = annoVal1
				return nil
			},
		}
		a2 := &mockAnnotator{
			writeFn: func(annotations map[string]string, ctx *label.NodeContext) error {
				annotations[annoKey2] = annoVal2
				return nil
			},
		}
		group := NewAnnotationGroup(a1, a2)
		annotations, err := group.WriteAll(&label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annoKey1], convey.ShouldEqual, annoVal1)
		convey.So(annotations[annoKey2], convey.ShouldEqual, annoVal2)
	})
	convey.Convey("should return error when annotator fails", t, func() {
		a1 := &mockAnnotator{
			writeFn: func(annotations map[string]string, ctx *label.NodeContext) error {
				return fmt.Errorf("mock error")
			},
		}
		group := NewAnnotationGroup(a1)
		annotations, err := group.WriteAll(&label.NodeContext{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(annotations, convey.ShouldBeNil)
	})
	convey.Convey("should return empty map when no annotators", t, func() {
		group := NewAnnotationGroup()
		annotations, err := group.WriteAll(&label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldNotBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
}
