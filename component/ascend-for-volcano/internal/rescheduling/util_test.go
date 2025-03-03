/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import "testing"

// TestJudgePublicFaultInReason pass when public fault in reason
func TestJudgePublicFaultInReason(t *testing.T) {
	t.Run("TestJudgePublicFaultInReason", func(t *testing.T) {
		faultTask := miniFaultTask{
			Reason: []FaultReasonList{
				{
					FaultDeviceList: FaultDeviceList{
						FaultType: PublicFaultType,
					},
				}, {
					FaultDeviceList: FaultDeviceList{
						FaultType: CardUnhealthy,
					},
				},
			},
		}
		if !JudgePublicFaultInReason(&faultTask) {
			t.Error("TestJudgePublicFaultInReason failed")
		}
	})
}
