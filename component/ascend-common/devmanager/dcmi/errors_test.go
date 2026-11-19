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

package dcmi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestDcmiErrorf verifies dcmiErrorf builds a *DcmiError carrying the return code,
// the semantic description and the unchanged rendered text.
func TestDcmiErrorf(t *testing.T) {
	convey.Convey("dcmiErrorf builds a structured DcmiError", t, func() {
		err := dcmiErrorf(NotSupportErrorCodeInt, "get device info failed, error code: %d", int32(NotSupportErrorCodeInt))

		dcmiErr, ok := AsDcmiError(err)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(dcmiErr.code, convey.ShouldEqual, NotSupportErrorCodeInt)
		convey.So(dcmiErr.desc, convey.ShouldEqual, "Device ID/function is not supported")
		convey.So(dcmiErr.Error(), convey.ShouldEqual, "get device info failed, error code: -8255")
	})
}

// TestErrorf verifies the exported Errorf builds a structured DcmiError as well,
// and falls back to "unknown error code" for codes not in dcmiErrMap.
func TestErrorf(t *testing.T) {
	convey.Convey("Errorf builds a structured DcmiError", t, func() {
		err := Errorf(FuncNotFoundErrorCodeInt, "function missing, please upgrade the driver")

		dcmiErr, ok := AsDcmiError(err)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(dcmiErr.code, convey.ShouldEqual, FuncNotFoundErrorCodeInt)
		convey.So(dcmiErr.desc, convey.ShouldEqual, "The called function is missing,please upgrade the driver")
		convey.So(dcmiErr.Error(), convey.ShouldEqual, "function missing, please upgrade the driver")
	})

	convey.Convey("Errorf with unknown code gets fallback desc", t, func() {
		err := Errorf(ErrorCode(12345), "unknown code message")

		dcmiErr, ok := AsDcmiError(err)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(dcmiErr.code, convey.ShouldEqual, ErrorCode(12345))
		convey.So(dcmiErr.desc, convey.ShouldEqual, "unknown error code")
	})
}

// TestDcmiDesc verifies dcmiDesc resolves known return codes and falls back for unknown ones.
func TestDcmiDesc(t *testing.T) {
	convey.Convey("dcmiDesc resolves known codes", t, func() {
		convey.So(dcmiDesc(int32(NotSupportErrorCodeInt)), convey.ShouldEqual, "Device ID/function is not supported")
		convey.So(dcmiDesc(int32(FuncNotFoundErrorCodeInt)), convey.ShouldEqual, "The called function is missing,please upgrade the driver")
	})

	convey.Convey("dcmiDesc falls back for unknown codes", t, func() {
		convey.So(dcmiDesc(12345), convey.ShouldEqual, "unknown error code")
	})
}

// TestAsDcmiErrorAndGetErrorCode verifies extraction only succeeds on a *DcmiError,
// returns false for nil/plain errors, and penetrates %w wrap chains.
func TestAsDcmiErrorAndGetErrorCode(t *testing.T) {
	convey.Convey("nil and plain errors are not extracted", t, func() {
		dcmiErr, ok := AsDcmiError(nil)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(dcmiErr, convey.ShouldBeNil)

		code, ok := GetErrorCode(nil)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(code, convey.ShouldEqual, ErrorCode(0))

		plain := errors.New("plain error")
		_, ok = AsDcmiError(plain)
		convey.So(ok, convey.ShouldBeFalse)
		code, ok = GetErrorCode(plain)
		convey.So(ok, convey.ShouldBeFalse)
	})

	convey.Convey("errors.As penetrates %w wrap chains", t, func() {
		wrapped := fmt.Errorf("get device total resource failed, error is: %w",
			dcmiErrorf(NotSupportErrorCodeInt, "not supported"))

		dcmiErr, ok := AsDcmiError(wrapped)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(dcmiErr.code, convey.ShouldEqual, NotSupportErrorCodeInt)

		code, ok := GetErrorCode(wrapped)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(code, convey.ShouldEqual, NotSupportErrorCodeInt)
	})
}

// TestIsNotSupportedAndIsFuncNotFound verifies the code-based classifiers.
func TestIsNotSupportedAndIsFuncNotFound(t *testing.T) {
	convey.Convey("IsNotSupported matches only -8255", t, func() {
		notSupported := dcmiErrorf(NotSupportErrorCodeInt, "not supported")
		funcNotFound := dcmiErrorf(FuncNotFoundErrorCodeInt, "function not found")
		plain := errors.New("plain error")

		convey.So(IsNotSupported(notSupported), convey.ShouldBeTrue)
		convey.So(IsNotSupported(funcNotFound), convey.ShouldBeFalse)
		convey.So(IsNotSupported(plain), convey.ShouldBeFalse)
		convey.So(IsNotSupported(nil), convey.ShouldBeFalse)
	})

	convey.Convey("IsFuncNotFound matches only -99998", t, func() {
		notSupported := dcmiErrorf(NotSupportErrorCodeInt, "not supported")
		funcNotFound := dcmiErrorf(FuncNotFoundErrorCodeInt, "function not found")
		plain := errors.New("plain error")

		convey.So(IsFuncNotFound(funcNotFound), convey.ShouldBeTrue)
		convey.So(IsFuncNotFound(notSupported), convey.ShouldBeFalse)
		convey.So(IsFuncNotFound(plain), convey.ShouldBeFalse)
		convey.So(IsFuncNotFound(nil), convey.ShouldBeFalse)
	})
}

// TestClassify verifies the four classification branches.
func TestClassify(t *testing.T) {
	convey.Convey("Classify maps codes to categories", t, func() {
		convey.So(Classify(dcmiErrorf(NotSupportErrorCodeInt, "x")), convey.ShouldEqual, CategoryNotSupported)
		convey.So(Classify(dcmiErrorf(FuncNotFoundErrorCodeInt, "x")), convey.ShouldEqual, CategoryFuncNotFound)
		convey.So(Classify(dcmiErrorf(ErrorCode(-8001), "x")), convey.ShouldEqual, CategoryFault)
		convey.So(Classify(dcmiErrorf(ErrorCode(12345), "x")), convey.ShouldEqual, CategoryFault)
		convey.So(Classify(errors.New("plain error")), convey.ShouldEqual, CategoryUnknown)
		convey.So(Classify(nil), convey.ShouldEqual, CategoryUnknown)
	})
}

// TestClassifyThroughWrapChain guards the %w wrap contract: classification must
// still work when a *DcmiError is wrapped by several layers of fmt.Errorf.
func TestClassifyThroughWrapChain(t *testing.T) {
	convey.Convey("classification penetrates multi-layer %w wrap chains", t, func() {
		base := dcmiErrorf(NotSupportErrorCodeInt, "not supported")
		wrapped := fmt.Errorf("get device total resource failed, error is: %w", base)
		top := fmt.Errorf("get virtual device info failed, error is: %w", wrapped)

		convey.So(IsNotSupported(top), convey.ShouldBeTrue)
		convey.So(Classify(top), convey.ShouldEqual, CategoryNotSupported)
	})
}
