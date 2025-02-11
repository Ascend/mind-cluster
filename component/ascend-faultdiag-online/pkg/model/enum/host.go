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
Package enum 提供枚举类
*/
package enum

// ChipType 定义芯片类型枚举类
type ChipType string

const (
	// Ascend310 ascend 310 chip
	Ascend310 ChipType = "Ascend310"
	// Ascend310B ascend 310B chip
	Ascend310B ChipType = "Ascend310B"
	// Ascend310P ascend 310P chip
	Ascend310P ChipType = "Ascend310P"
	// Ascend910 ascend 910 chip
	Ascend910 ChipType = "Ascend910"
	// Ascend910A2 ascend 910A2 chip
	Ascend910A2 ChipType = "Ascend910A2"
	// Ascend910A3 ascend Ascend910A3 chip
	Ascend910A3 ChipType = "Ascend910A3"
	// Atlas200ISoc 200 soc env
	Atlas200ISoc ChipType = "Atlas 200I SoC A1"
)

var (
	// Ascend910SerialChips 910系列芯片
	Ascend910SerialChips = []ChipType{Ascend910, Ascend910A2, Ascend910A3}
)
