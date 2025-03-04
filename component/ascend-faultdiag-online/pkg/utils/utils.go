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
Package utils 提供了一些工具函数
*/
package utils

import (
	"strconv"
)

// ToFloat64 interface转float64
func ToFloat64(val interface{}, defaultValue float64) float64 {
	switch v := val.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		float, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue
		}
		return float
	default:
		return defaultValue
	}
}

// ToString interface转string
func ToString(val interface{}) string {
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}
