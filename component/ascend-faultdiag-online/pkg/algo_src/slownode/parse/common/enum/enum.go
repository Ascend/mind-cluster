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

// Package enum provides some constant value
package enum

// OpType 算子类型
type OpType string

const (
	// Tp tp算子类型
	Tp OpType = "tp"
	// Pp pp算子类型
	Pp OpType = "pp"
)

// FileMode 文件mod
type FileMode string

const (
	// ReadMode "r"=只读
	ReadMode FileMode = "r"
	// WriteMode "w"=只写(覆盖)
	WriteMode FileMode = "w"
	// AppendMode "a"=追加写
	AppendMode FileMode = "a"
)
