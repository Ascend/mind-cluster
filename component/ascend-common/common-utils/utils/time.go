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

// Package utils this file for time utils
package utils

import "time"

// monoBase the process-level monotonic baseline captured at startup
var monoBase = time.Now()

// NowMono returns the elapsed nanoseconds since the process baseline, measured by
// the monotonic clock. Unlike time.Now().UnixNano(), it is immune to wall clock
// jumps (e.g. NTP correction), so it is safe for measuring intervals and expiry.
func NowMono() int64 {
	return int64(time.Since(monoBase))
}
