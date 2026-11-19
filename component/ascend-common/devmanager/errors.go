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
	"strings"

	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

// The devmanager layer re-exports the dcmi structured error APIs, so business
// code should classify errors through the devmanager package (IsNotSupported,
// GetErrorCode, Classify) instead of importing dcmi for error classification.
// Tests may still import dcmi directly to construct structured errors.

// ErrorCode is the return code of the bottom libdcmi.so interface, such as -8255, -99998, -8001...
type ErrorCode = dcmi.ErrorCode

// Key return code constants, consistent with dcmiErrMap.
const (
	// NotSupportErrorCodeInt device ID/function is not supported
	NotSupportErrorCodeInt = dcmi.NotSupportErrorCodeInt
	// FuncNotFoundErrorCodeInt the called function is missing, please upgrade the driver
	FuncNotFoundErrorCodeInt = dcmi.FuncNotFoundErrorCodeInt
)

// ErrCategory is the error category for unified log-level / report-policy decisions.
type ErrCategory = dcmi.ErrCategory

const (
	// CategoryUnknown non-DcmiError or unknown code
	CategoryUnknown = dcmi.CategoryUnknown
	// CategoryNotSupported -8255: device/interface not supported (normal)
	CategoryNotSupported = dcmi.CategoryNotSupported
	// CategoryFuncNotFound -99998: function missing, need to upgrade driver
	CategoryFuncNotFound = dcmi.CategoryFuncNotFound
	// CategoryFault the rest: real hardware/driver fault
	CategoryFault = dcmi.CategoryFault
)

// IsNotSupported reports whether err carries the -8255 "device/function not supported" code.
// It prefers the structured return code and falls back to the legacy string match, so
// plain fmt.Errorf errors from not-yet-structured producers are still recognized during
// the transition; remove the fallback once the whole chain is structured.
func IsNotSupported(err error) bool {
	if dcmi.IsNotSupported(err) {
		return true
	}
	return err != nil && strings.Contains(err.Error(), common.NotSupportErrorCode)
}

// IsFuncNotFound reports whether err carries the -99998 "function missing, need to upgrade driver" code.
func IsFuncNotFound(err error) bool {
	return dcmi.IsFuncNotFound(err)
}

// GetErrorCode extracts the return code from err; ok is false if err is not a *DcmiError.
func GetErrorCode(err error) (ErrorCode, bool) {
	return dcmi.GetErrorCode(err)
}

// Classify classifies the error by its category based on the return code.
func Classify(err error) ErrCategory {
	return dcmi.Classify(err)
}
