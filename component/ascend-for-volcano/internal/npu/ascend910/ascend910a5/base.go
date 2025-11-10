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

// Package ascend910a5 is using for HuaWei Ascend 910A5 pin affinity schedule.
package ascend910a5

// SetArch Set the job arch to distinguish between jobs. A+X 16P,A+K 8p.
func (tp *Base910A5) SetArch(value string) {
	tp.arch = value
}

// GetArch Get the job arch to distinguish between jobs. A+X 16P,A+K 8p.
func (tp *Base910A5) GetArch() string {
	return tp.arch
}
