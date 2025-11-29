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
Package api common const.
*/
package api

// CheckIsVersionA5 check card type is A5
func CheckIsVersionA5(cardTypeVersion string) bool {
	return cardTypeVersion == VersionA5
}

// IsA5InferServer check device is A5InferServer
func IsA5InferServer(acceleratorType string) bool {
	switch acceleratorType {
	case Ascend800ia5x8, Ascend800ta5x8, Ascend800ia5SuperPod, Ascend800ta5SuperPod, Ascend800ia5Stacking:
		return true
	default:
		return false
	}
}
