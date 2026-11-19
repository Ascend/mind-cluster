/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package devmanager

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/dcmi"
)

// TestIsNotSupportedFallback verifies the transition fallback: the structured code
// is preferred, and a plain error carrying "-8255" text is still recognized.
func TestIsNotSupportedFallback(t *testing.T) {
	convey.Convey("IsNotSupported recognizes structured -8255", t, func() {
		err := dcmi.Errorf(dcmi.NotSupportErrorCodeInt, "device type %s is not supported (cached)", "AICore")
		convey.So(IsNotSupported(err), convey.ShouldBeTrue)
	})

	convey.Convey("IsNotSupported falls back to the legacy string match", t, func() {
		plain := errors.New("get device info failed, error code: -8255")
		convey.So(IsNotSupported(plain), convey.ShouldBeTrue)
	})

	convey.Convey("IsNotSupported does not match other codes, plain texts or nil", t, func() {
		convey.So(IsNotSupported(dcmi.Errorf(dcmi.FuncNotFoundErrorCodeInt, "function missing")), convey.ShouldBeFalse)
		convey.So(IsNotSupported(errors.New("something else")), convey.ShouldBeFalse)
		convey.So(IsNotSupported(nil), convey.ShouldBeFalse)
	})
}

// TestErrorCodeReExports verifies the devmanager-level error code alias and constants.
func TestErrorCodeReExports(t *testing.T) {
	convey.Convey("constants match the dcmi return codes", t, func() {
		convey.So(NotSupportErrorCodeInt, convey.ShouldEqual, dcmi.NotSupportErrorCodeInt)
		convey.So(FuncNotFoundErrorCodeInt, convey.ShouldEqual, dcmi.FuncNotFoundErrorCodeInt)
	})

	convey.Convey("GetErrorCode extracts the code through the devmanager API", t, func() {
		code, ok := GetErrorCode(dcmi.Errorf(dcmi.NotSupportErrorCodeInt, "not supported"))
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(code, convey.ShouldEqual, NotSupportErrorCodeInt)
	})
}
