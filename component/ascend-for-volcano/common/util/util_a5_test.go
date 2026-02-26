/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"reflect"
	"testing"
)

func TestMergeUnique(t *testing.T) {
	var list1 = []int{1, 2}
	var list2 = []int{3, 4}
	var list3 = []int{1, 2, 3, 4}
	t.Run("test func MergeUnique success", func(t *testing.T) {
		result := MergeUnique(list1, list2)
		if !reflect.DeepEqual(result, list3) {
			t.Errorf("test MergeUnique fail ,expect:%v,result:%v", list3, result)
		}
	})
}

func TestHasCommonElement(t *testing.T) {
	var list1 = []int{1, 2, 3, 4}
	var list2 = []int{3, 4, 5, 6}
	var list3 = []int{3, 4}
	var resultTrue = true
	t.Run("test func MergeUnique success", func(t *testing.T) {
		result, hasCommon := HasCommonElement(list1, list2)
		if !reflect.DeepEqual(result, list3) || !reflect.DeepEqual(resultTrue, hasCommon) {
			t.Errorf("test HasCommonElement fail ,expect:%v,result:%v", list3, result)
		}
	})
}
