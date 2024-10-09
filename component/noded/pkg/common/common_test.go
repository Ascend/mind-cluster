/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

import (
	"encoding/json"
	"fmt"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const ValidDiveType = 0x03
const WrongDiveType = 0xFF

// TestGetDeviceType get device type
func TestGetDeviceType(t *testing.T) {
	convey.Convey("test get device type", t, func() {
		convey.Convey("right device byte", func() {
			device := GetDeviceType(ValidDiveType)
			convey.So(device, convey.ShouldEqual, "PSU")
		})
		convey.Convey("wrong device byte", func() {
			device := GetDeviceType(WrongDiveType)
			convey.So(device, convey.ShouldEqual, UnknownDevice)
		})
	})
}

// TestGetPattern test node name pattern
func TestGetPattern(t *testing.T) {
	convey.Convey("test get node pattern", t, func() {
		convey.Convey("right node name str", func() {
			pattern := GetPattern()[RegexNodeNameKey]
			match := pattern.MatchString("node-1")
			convey.So(match, convey.ShouldBeTrue)
		})
		convey.Convey("wrong node name str", func() {
			pattern := GetPattern()[RegexNodeNameKey]
			match := pattern.MatchString("!%!#das#%")
			convey.So(match, convey.ShouldBeFalse)
		})
	})
}

// TestCopyStringSlice test copy string slice
func TestCopyStringSlice(t *testing.T) {
	convey.Convey("test copy string slice", t, func() {
		convey.Convey("copy string slice", func() {
			oldStrSlice := []string{"test1", "test2"}
			newStrSlice := CopyStringSlice(oldStrSlice)
			convey.So(len(oldStrSlice), convey.ShouldEqual, len(newStrSlice))
			for i, str := range oldStrSlice {
				if i >= len(newStrSlice) {
					return
				}
				convey.So(str, convey.ShouldEqual, newStrSlice[i])
			}
		})
	})
}

// TestMakeDataHash test make data hash
func TestMakeDataHash(t *testing.T) {
	convey.Convey("test make data hash", t, func() {
		convey.Convey("hash success", func() {
			NodeInfo := FaultDevInfo{FaultDevList: []*FaultDev{&FaultDev{
				DeviceType: "PSU",
				DeviceId:   0,
				FaultCode:  []string{"00000001"},
				FaultLevel: NotHandleFault,
			}}}
			result := MakeDataHash(NodeInfo)
			convey.So(result, convey.ShouldNotBeEmpty)
		})
		convey.Convey("Marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("marshal failed")
			})
			defer mockMarshal.Reset()
			NodeInfo := FaultDevInfo{FaultDevList: []*FaultDev{&FaultDev{
				DeviceType: "PSU",
				DeviceId:   0,
				FaultCode:  []string{"00000001"},
				FaultLevel: NotHandleFault,
			}}}
			result := MakeDataHash(NodeInfo)
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

// TestDeepCopyFaultConfig test deep copy fault config
func TestDeepCopyFaultConfig(t *testing.T) {
	convey.Convey("test deep copy fault config", t, func() {
		convey.Convey("deep copy fault config", func() {
			oldFaultConfig := &FaultConfig{FaultTypeCode: &FaultTypeCode{
				NotHandleFaultCodes:   []string{"00000001"},
				PreSeparateFaultCodes: []string{"00000002"},
				SeparateFaultCodes:    []string{"00000003"},
			}}
			newFaultConfig := &FaultConfig{FaultTypeCode: &FaultTypeCode{}}
			DeepCopyFaultConfig(oldFaultConfig, newFaultConfig)
			FaultConfigEqual(oldFaultConfig, newFaultConfig)
		})
	})
}

// TestRemoveDuplicateString test string deduplication
func TestRemoveDuplicateString(t *testing.T) {
	convey.Convey("test remove duplicate string", t, func() {
		convey.Convey("remove duplicate string", func() {
			slice1 := []string{"00000001", "00000001", "00000002", "00000003"}
			slice2 := []string{"00000001", "00000002", "00000003"}
			newSlice := RemoveDuplicateString(slice1)
			SliceStrEqual(newSlice, slice2)
		})
	})
}

// TestRevertByteSlice test revert byte slice
func TestRevertByteSlice(t *testing.T) {
	convey.Convey("test revert byte slice", t, func() {
		convey.Convey("revert byte slice", func() {
			slice1 := []byte{0x01, 0x02, 0x03}
			slice2 := []byte{0x03, 0x02, 0x01}
			RevertByteSlice(slice2)
			SliceByteEqual(slice1, slice2)
		})
	})
}

// TestConvertIntToByteSlice test convert int to byte slice
func TestConvertIntToByteSlice(t *testing.T) {
	convey.Convey("test covert positive int to byte slice", t, func() {
		convey.Convey("covert int to byte slice", func() {
			var num int64 = 257
			targetByteSlice := []byte{0x01, 0x01}
			resultByteSlice := ConvertIntToTwoByteSlice(num)
			SliceByteEqual(resultByteSlice, targetByteSlice)
		})
		convey.Convey("convert negative int to byte slice", func() {
			var num int64 = -1
			targetByteSlice := []byte{0x00, 0x00}
			resultByteSlice := ConvertIntToTwoByteSlice(num)
			SliceByteEqual(resultByteSlice, targetByteSlice)
		})
		convey.Convey("convert normal int to slice to test slice length", func() {
			var num int64 = 1
			targetByteSlice := []byte{0x01, 0x00}
			resultByteSlice := ConvertIntToTwoByteSlice(num)
			SliceByteEqual(resultByteSlice, targetByteSlice)
		})
		convey.Convey("convert out of range int to slice to byte slice", func() {
			var num int64 = 65536
			targetByteSlice := []byte{0x00, 0x00}
			resultByteSlice := ConvertIntToTwoByteSlice(num)
			SliceByteEqual(resultByteSlice, targetByteSlice)
		})
	})
}

// TestNewSignalWatcher test create a signal watcher
func TestNewSignalWatcher(t *testing.T) {
	convey.Convey("create new signal watcher", t, func() {
		signWatcher := NewSignalWatcher(syscall.SIGKILL)
		convey.So(signWatcher, convey.ShouldNotBeNil)
	})
}

// FaultConfigEqual judge if fault config is equal
func FaultConfigEqual(oldFaultConfig, newFaultConfig *FaultConfig) {
	convey.So(oldFaultConfig.FaultTypeCode, convey.ShouldNotBeNil)
	convey.So(newFaultConfig.FaultTypeCode, convey.ShouldNotBeNil)
	FaultTypeCodesEqual(oldFaultConfig.FaultTypeCode, newFaultConfig.FaultTypeCode)
}

// FaultTypeCodesEqual judge if fault type code is equal
func FaultTypeCodesEqual(oldFaultTypeCode, newFaultTypeCode *FaultTypeCode) {
	SliceStrEqual(oldFaultTypeCode.NotHandleFaultCodes, newFaultTypeCode.NotHandleFaultCodes)
	SliceStrEqual(oldFaultTypeCode.PreSeparateFaultCodes, newFaultTypeCode.PreSeparateFaultCodes)
	SliceStrEqual(oldFaultTypeCode.SeparateFaultCodes, newFaultTypeCode.SeparateFaultCodes)
}

// SliceStrEqual judge string slice is equal
func SliceStrEqual(slice1, slice2 []string) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}

// SliceByteEqual judge byte slice is equal
func SliceByteEqual(slice1, slice2 []byte) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}
