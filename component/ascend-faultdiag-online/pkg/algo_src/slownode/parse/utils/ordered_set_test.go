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

// Package utils provides some common utils
package utils

import (
	"fmt"
	"testing"
)

func TestNewOrderedIDSet(t *testing.T) {
	var count10 = 10
	var count5 = 5
	orderedIDSet := NewOrderedIDSet()
	for i := 0; i < count10; i++ {
		if !orderedIDSet.Contains(int64(i + 1)) {
			orderedIDSet.Add(int64(i + 1))
		}
		fmt.Println("add:", orderedIDSet)

		if len(orderedIDSet.ids) > count5 {
			id, found := orderedIDSet.GetByIndex(0)
			if found {
				fmt.Println("====")
				fmt.Println("need remove id:", id)
				orderedIDSet.Remove(id)
				fmt.Println("remove:", orderedIDSet)
			}
		}
	}
}

func TestRemove(t *testing.T) {
	var count = 5
	orderedIDSet := NewOrderedIDSet()
	for i := 0; i < count; i++ {
		if !orderedIDSet.Contains(int64(i + 1)) {
			orderedIDSet.Add(int64(i + 1))
		}
	}
	fmt.Println("add:", orderedIDSet)

	orderedIDSet.Remove(int64(count))
	fmt.Println("remove:", orderedIDSet)

}
