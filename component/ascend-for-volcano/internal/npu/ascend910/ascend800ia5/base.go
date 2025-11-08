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
Package ascend800ia5 is using for HuaWei ascend 800I A5 affinity schedule.
*/
package ascend800ia5

// SetAcceleratorValue Set the acceleratorValue to distinguish between task types.
func (ab *Base800ia5) SetAcceleratorValue(value string) {
	ab.acceleratorValue = value
}

// GetAcceleratorValue Get the acceleratorValue to distinguish between task types.
func (ab *Base800ia5) GetAcceleratorValue() string {
	return ab.acceleratorValue
}

// SetArch Set the job arch to distinguish between jobs.
func (ab *Base800ia5) SetArch(value string) {
	ab.arch = value
}

// GetArch Get the job arch to distinguish between jobs. A+X 16P,A+K 8p.
func (ab *Base800ia5) GetArch() string {
	return ab.arch
}
