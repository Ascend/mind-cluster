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

// Package algo 网络连通性检测算法
package algo

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestReverseSlice(t *testing.T) {
	convey.Convey("Test reverseSlice", t, func() {
		convey.Convey(`when input is ["a", "b", "c"]`, func() {
			input := []string{"a", "b", "c"}
			expected := []string{"c", "b", "a"}
			result := reverseSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when input is ["1", "2", "3", "4"]`, func() {
			input := []string{"1", "2", "3", "4"}
			expected := []string{"4", "3", "2", "1"}
			result := reverseSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("when input is []", func() {
			input := make([]string, 0)
			expected := make([]string, 0)
			result := reverseSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when input is ["single"]`, func() {
			input := []string{"single"}
			expected := []string{"single"}
			result := reverseSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestUniqueSlice(t *testing.T) {
	convey.Convey("Test uniqueSlice", t, func() {
		convey.Convey(`when input is ["a", "b", "a", "c"]`, func() {
			input := []string{"a", "b", "a", "c"}
			expected := []string{"a", "b", "c"}
			result := uniqueSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when input is ["1", "2", "2", "3"]`, func() {
			input := []string{"1", "2", "2", "3"}
			expected := []string{"1", "2", "3"}
			result := uniqueSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("when input is []", func() {
			input := make([]string, 0)
			expected := make([]string, 0)
			result := uniqueSlice(input)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestAppearOnce(t *testing.T) {
	convey.Convey("Test appearOnce", t, func() {
		convey.Convey(`when input is ["a", "b", "a", "c"]`, func() {
			input := []string{"a", "b", "a", "c"}
			result := appearOnce(input)
			convey.So(result, convey.ShouldContain, "b")
			convey.So(result, convey.ShouldContain, "c")
		})

		convey.Convey(`when input is ["1", "2", "2", "3"]`, func() {
			input := []string{"1", "2", "2", "3"}
			result := appearOnce(input)
			convey.So(result, convey.ShouldContain, "1")
			convey.So(result, convey.ShouldContain, "3")
		})

		convey.Convey("when input is []", func() {
			input := make([]string, 0)
			expected := make([]string, 0)
			result := appearOnce(input)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestRemoveElements(t *testing.T) {
	convey.Convey("Test removeElements", t, func() {
		convey.Convey(`when a is ["a", "b", "c"] and b is ["b"]`, func() {
			a := []string{"a", "b", "c"}
			b := []string{"b"}
			expected := []string{"a", "c"}
			result := removeElements(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when a is ["1", "2", "3"] and b is ["4"]`, func() {
			a := []string{"1", "2", "3"}
			b := []string{"4"}
			expected := []string{"1", "2", "3"}
			result := removeElements(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when a is ["a"] and b is []`, func() {
			a := []string{"a"}
			var b []string
			expected := []string{"a"}
			result := removeElements(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey(`when a is [] and b is ["a"]`, func() {
			a := make([]string, 0)
			b := []string{"a"}
			expected := make([]string, 0)
			result := removeElements(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestContains(t *testing.T) {
	convey.Convey("Test contains", t, func() {
		convey.Convey(`when slice is ["a", "b", "c"] and item is "b"`, func() {
			slice := []string{"a", "b", "c"}
			item := "b"
			result := contains(slice, item)
			convey.So(result, convey.ShouldEqual, true)
		})

		convey.Convey(`when slice is ["1", "2", "3"] and item is "4"`, func() {
			slice := []string{"1", "2", "3"}
			item := "4"
			result := contains(slice, item)
			convey.So(result, convey.ShouldEqual, false)
		})

		convey.Convey(`when slice is [] and item is "a"`, func() {
			var slice []string
			item := "a"
			result := contains(slice, item)
			convey.So(result, convey.ShouldEqual, false)
		})
	})
}

func TestContainsKey(t *testing.T) {
	convey.Convey("Test containsKey", t, func() {
		convey.Convey(`when map is {"a": 1} and key is "a"`, func() {
			m := map[string]any{"a": 1}
			key := "a"
			result := containsKey(m, key)
			convey.So(result, convey.ShouldEqual, true)
		})

		convey.Convey(`when map is {"b": 2} and key is "a"`, func() {
			m := map[string]any{"b": 2}
			key := "a"
			result := containsKey(m, key)
			convey.So(result, convey.ShouldEqual, false)
		})

		convey.Convey(`when map is {} and key is "a"`, func() {
			m := map[string]any{}
			key := "a"
			result := containsKey(m, key)
			convey.So(result, convey.ShouldEqual, false)
		})
	})
}

func TestDeduplicateSlice(t *testing.T) {
	convey.Convey("Given a slice of maps", t, func() {
		input := []map[string]any{
			{"key1": "value1", "key2": "value2"},
			{"key1": "value1", "key2": "value2"}, // Duplicate
			{"key1": "value3", "key2": "value4"},
			{"key1": "value5", "key2": "value6"},
			{"key1": "value5", "key2": "value6"}, // Duplicate
		}

		convey.Convey("When deduplicating the slice", func() {
			result := deduplicateSlice(input)

			convey.Convey("Then the result should contain unique maps", func() {
				convey.So(len(result), convey.ShouldEqual, 3) // Expecting 3 unique items
			})

			convey.Convey("And the result should match the expected output", func() {
				expected := []map[string]any{
					{"key1": "value1", "key2": "value2"},
					{"key1": "value3", "key2": "value4"},
					{"key1": "value5", "key2": "value6"},
				}
				convey.So(result, convey.ShouldResemble, expected)
			})
		})

		convey.Convey("When the input is empty", func() {
			var input []map[string]any
			result := deduplicateSlice(input)

			convey.Convey("Then the result should also be empty", func() {
				convey.So(len(result), convey.ShouldEqual, 0)
			})
		})

		convey.Convey("When the input has no duplicates", func() {
			input := []map[string]any{
				{"key1": "value1"},
				{"key2": "value2"},
			}
			result := deduplicateSlice(input)

			convey.Convey("Then the result should match the input", func() {
				convey.So(result, convey.ShouldResemble, input)
			})
		})
	})
}

func TestMergeAndDeduplicate(t *testing.T) {
	convey.Convey("Test MergeAndDeduplicate", t, func() {
		convey.Convey(`when a is [{"key": "value1"}, {"key": "value2"}] and b is [{"key": "value2"}, 
			{"key": "value3"}]`, func() {
			a := []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
			}
			b := []map[string]any{
				{"key": "value2"},
				{"key": "value3"},
			}
			expected := []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
				{"key": "value3"},
			}
			result := MergeAndDeduplicate(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("when b is empty", func() {
			a := []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
			}
			b := make([]map[string]any, 0)
			expected := []map[string]any{
				{"key": "value1"},
				{"key": "value2"},
			}
			result := MergeAndDeduplicate(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("when both a and b are empty", func() {
			a := make([]map[string]any, 0)
			b := make([]map[string]any, 0)
			expected := make([]map[string]any, 0)
			result := MergeAndDeduplicate(a, b)
			convey.So(result, convey.ShouldResemble, expected)
		})
	})
}

func TestNewNetDetect(t *testing.T) {
	superPodId := "testSuperPodId"
	netDetect := NewNetDetect(superPodId)

	convey.Convey("Test NewNetDetect", t, func() {
		convey.Convey("curSuperPodId should match the input", func() {
			convey.So(netDetect.curSuperPodId, convey.ShouldEqual, superPodId)
		})

		convey.Convey("curFullPingFlag should be false", func() {
			convey.So(netDetect.curFullPingFlag, convey.ShouldBeFalse)
		})

		convey.Convey("curAxisStrategy should match the constant", func() {
			convey.So(netDetect.curAxisStrategy, convey.ShouldEqual, crossAxisConstant)
		})

		convey.Convey("curTopo should be empty", func() {
			convey.So(len(netDetect.curTopo), convey.ShouldEqual, 0)
		})

		convey.Convey("curPingPeriod should match the default", func() {
			convey.So(netDetect.curPingPeriod, convey.ShouldEqual, defaultPingPeriod)
		})

		convey.Convey("curSuppressedPeriod should match the default", func() {
			convey.So(netDetect.curSuppressedPeriod, convey.ShouldEqual, defaultSPeriod)
		})

		convey.Convey("curDetectParams should be empty", func() {
			convey.So(len(netDetect.curDetectParams), convey.ShouldEqual, 0)
		})

		convey.Convey("curNpuInfo should have no samples", func() {
			convey.So(len(netDetect.curNpuInfo), convey.ShouldEqual, 0)
		})

		convey.Convey("curSlideWindows should be empty", func() {
			convey.So(len(netDetect.curSlideWindows), convey.ShouldEqual, 0)
		})

		convey.Convey("curSlideWindowsMaxTs should be 0", func() {
			convey.So(netDetect.curSlideWindowsMaxTs, convey.ShouldEqual, 0)
		})
	})
}

func TestMoveSliceLeftOneStep(t *testing.T) {
	convey.Convey("Testing moveSliceLeftTwoStep function", t, func() {
		convey.Convey("When the slice is empty", func() {
			result := moveSliceLeftTwoStep([]string{})
			convey.So(result, convey.ShouldResemble, []string{})
		})

		convey.Convey("When the slice has one element", func() {
			result := moveSliceLeftTwoStep([]string{"A"})
			convey.So(result, convey.ShouldResemble, []string{"A"})
		})

		convey.Convey("When the slice has multiple elements", func() {
			result := moveSliceLeftTwoStep([]string{"A", "B", "C", "D"})
			convey.So(result, convey.ShouldResemble, []string{"C", "D", "A", "B"})
		})

		convey.Convey("When the slice has two elements", func() {
			result := moveSliceLeftTwoStep([]string{"X", "Y"})
			convey.So(result, convey.ShouldResemble, []string{"X", "Y"})
		})
	})
}

func TestGetIndex(t *testing.T) {
	convey.Convey("Given a slice of strings", t, func() {
		slice := []string{"apple", "banana", "cherry"}

		convey.Convey("When the target string is in the slice", func() {
			target := "banana"
			index := getIndex(slice, target)

			convey.Convey("Then it should return the correct index", func() {
				convey.So(index, convey.ShouldEqual, 1)
			})
		})

		convey.Convey("When the target string is not in the slice", func() {
			target := "orange"
			index := getIndex(slice, target)

			convey.Convey("Then it should return -1", func() {
				convey.So(index, convey.ShouldEqual, -1)
			})
		})
	})
}

func TestRoundToThreeDecimal(t *testing.T) {
	convey.Convey("Testing roundToThreeDecimal function", t, func() {
		convey.Convey("Given a positive float", func() {
			input1 := 123.456789 // 输入1
			output1 := 123.457   // 结果1
			result := roundToThreeDecimal(input1)
			convey.So(result, convey.ShouldEqual, output1)

			input2 := 123.456789 // 输入2
			output2 := 123.457   // 结果2
			result = roundToThreeDecimal(input2)
			convey.So(result, convey.ShouldEqual, output2)

			input3 := 123.0001 // 输入3
			output3 := 123.000 // 结果3
			result = roundToThreeDecimal(input3)
			convey.So(result, convey.ShouldEqual, output3)
		})

		convey.Convey("Given a negative float", func() {
			input1 := 200.0    // 输入1
			output1 := 200.000 // 结果1
			result := roundToThreeDecimal(input1)
			convey.So(result, convey.ShouldAlmostEqual, output1)

			input2 := -123.0    // 输入2
			output2 := -123.000 // 结果2
			result = roundToThreeDecimal(input2)
			convey.So(result, convey.ShouldEqual, output2)

			input3 := -123.55444444 // 输入3
			output3 := -123.554     // 结果3
			result = roundToThreeDecimal(input3)
			convey.So(result, convey.ShouldEqual, output3)
		})

		convey.Convey("Given zero", func() {
			input1 := 0.0  // 输入1
			output1 := 0.0 // 结果1
			result := roundToThreeDecimal(input1)
			convey.So(result, convey.ShouldEqual, output1)
		})
	})
}
