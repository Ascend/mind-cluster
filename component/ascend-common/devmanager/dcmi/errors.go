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
)

// ErrorCode is the return code of the bottom libdcmi.so interface, such as -8255, -99998, -8001...
type ErrorCode int32

// Key return code constants, consistent with dcmiErrMap.
const (
	// NotSupportErrorCodeInt device ID/function is not supported
	NotSupportErrorCodeInt ErrorCode = -8255
	// FuncNotFoundErrorCodeInt the called function is missing, please upgrade the driver
	FuncNotFoundErrorCodeInt ErrorCode = -99998
)

// DcmiError is the unified structured error returned by the dcmi interface wrappers.
// It carries the structured return code and context, so callers can make precise
// decisions by code or category instead of fragile string matching.
// All fields are unexported: the type can only be constructed through the package
// builders (dcmiErrorf/Errorf/buildDcmiErr), which guarantees Error() never returns
// an empty message; inspect it via errors.As + GetErrorCode or the Classify helpers.
type DcmiError struct {
	code     ErrorCode // unique source of truth, from the bottom libdcmi interface
	desc     string    // from dcmiErrMap; "unknown error code" if not found
	rendered string    // original v1/v2 text, returned as-is by Error()
}

// Error returns the original rendered text, keeping the v1/v2 messages unchanged.
func (e *DcmiError) Error() string { return e.rendered }

// dcmiDesc returns the semantic description of the error code, or
// "unknown error code" if the code is not in dcmiErrMap.
func dcmiDesc(code int32) string {
	if desc, ok := dcmiErrMap[code]; ok {
		return desc
	}
	return "unknown error code"
}

// dcmiErrorf constructs a structured error from the return code and a format string.
// The rendered text is identical to the pre-refactor message.
// The code parameter accepts any int32-like type (C.int, int32, ErrorCode, ...), so
// call sites can pass the raw dcmi return code without an explicit ErrorCode conversion.
func dcmiErrorf[T ~int32 | ~uint32](code T, format string, args ...any) error {
	return &DcmiError{
		code:     ErrorCode(code),
		desc:     dcmiDesc(int32(code)),
		rendered: fmt.Sprintf(format, args...),
	}
}

// Errorf constructs a structured error carrying the given bottom return code.
// It is used where a readable message must be kept while the code is attached,
// e.g. devmanager-layer "not supported (cached)" short-circuits.
func Errorf(code ErrorCode, format string, args ...any) error {
	return dcmiErrorf(code, format, args...)
}

// AsDcmiError extracts the structured *DcmiError from err via errors.As.
func AsDcmiError(err error) (*DcmiError, bool) {
	var dcmiErr *DcmiError
	if errors.As(err, &dcmiErr) {
		return dcmiErr, true
	}
	return nil, false
}

// GetErrorCode extracts the return code from err; ok is false if err is not a *DcmiError.
func GetErrorCode(err error) (ErrorCode, bool) {
	dcmiErr, ok := AsDcmiError(err)
	if !ok {
		return 0, false
	}
	return dcmiErr.code, true
}

// IsNotSupported reports whether err carries the -8255 "device/function not supported" code.
func IsNotSupported(err error) bool {
	c, ok := GetErrorCode(err)
	return ok && c == NotSupportErrorCodeInt
}

// IsFuncNotFound reports whether err carries the -99998 "function missing, need to upgrade driver" code.
func IsFuncNotFound(err error) bool {
	c, ok := GetErrorCode(err)
	return ok && c == FuncNotFoundErrorCodeInt
}

// ErrCategory is the error category for unified log-level / report-policy decisions.
type ErrCategory int

const (
	// CategoryUnknown non-DcmiError or unknown code
	CategoryUnknown ErrCategory = iota
	// CategoryNotSupported -8255: device/interface not supported (normal)
	CategoryNotSupported
	// CategoryFuncNotFound -99998: function missing, need to upgrade driver
	CategoryFuncNotFound
	// CategoryFault the rest: real hardware/driver fault
	CategoryFault
)

// Classify classifies the error by its category based on the return code.
// Only the passed-through *DcmiError participates in classification.
func Classify(err error) ErrCategory {
	c, ok := GetErrorCode(err)
	if !ok {
		return CategoryUnknown
	}
	switch c {
	case NotSupportErrorCodeInt:
		return CategoryNotSupported
	case FuncNotFoundErrorCodeInt:
		return CategoryFuncNotFound
	default:
		return CategoryFault
	}
}
