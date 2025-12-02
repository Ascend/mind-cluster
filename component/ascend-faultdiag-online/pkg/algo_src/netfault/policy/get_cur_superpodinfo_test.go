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

// Package policy is used for processing super pod information
package policy

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestIsAlphanumeric test for func isAlphanumeric
func TestIsAlphanumeric(t *testing.T) {
	convey.Convey("Test isAlphanumeric", t, func() {
		convey.Convey("should return false when request is invalid", func() {
			s := "abc123$#"
			ret := isAlphanumeric(s)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when request is valid", func() {
			s := "a1b2c3"
			ret := isAlphanumeric(s)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestContainsElement test for func containsElement
func TestContainsElement(t *testing.T) {
	convey.Convey("Test containsElement", t, func() {
		convey.Convey("should return false when request is invalid", func() {
			slice := []string{"abc123$#"}
			str := "1"
			ret := containsElement(slice, str)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return true when request is valid", func() {
			slice := []string{"abc123$#", "1"}
			str := "1"
			ret := containsElement(slice, str)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}
