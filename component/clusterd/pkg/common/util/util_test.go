// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util test function
package util

import (
	"context"
	"syscall"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

const (
	testHashCodeStr   = "testHashCodeStr"
	testHashCode      = "9f9046a094b85b781aca276c46c36956ed07f6de1ec1f94268bfce764e0a9606"
	testNilHashCode   = "74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b"
	testHashCodeError = "2343245345"
	testMapString     = `{"0":1}`
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestNewSignalWatcher(t *testing.T) {
	convey.Convey("TestNewSignalWatcher", t, func() {
		signalChan := NewSignalWatcher(syscall.SIGINT)
		convey.So(len(signalChan), convey.ShouldEqual, 0)
	})
}

func TestEqualDataHash(t *testing.T) {
	convey.Convey("TestEqualDataHash", t, func() {
		convey.Convey("checkCode is nil", func() {
			result := EqualDataHash("", nil)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("data not equals checkCode", func() {
			result := EqualDataHash(testHashCode, testHashCodeError)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("data equals checkCode", func() {
			result := EqualDataHash(testHashCode, testHashCodeStr)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestMakeDataHash(t *testing.T) {
	convey.Convey("TestMakeDataHash", t, func() {
		convey.Convey("data is nil", func() {
			result := MakeDataHash(nil)
			convey.So(result, convey.ShouldEqual, testNilHashCode)
		})
		convey.Convey("data is function", func() {
			result := MakeDataHash(func() {})
			convey.So(result, convey.ShouldEqual, "")
		})
		convey.Convey("data equals checkCode", func() {
			result := MakeDataHash(testHashCodeStr)
			convey.So(result, convey.ShouldEqual, testHashCode)
		})
	})
}

func TestObjToString(t *testing.T) {
	convey.Convey("TestObjToString", t, func() {
		convey.Convey("data is nil", func() {
			result := ObjToString(nil)
			convey.So(result, convey.ShouldEqual, "null")
		})
		convey.Convey("data is emtpy", func() {
			result := ObjToString("")
			convey.So(result, convey.ShouldEqual, `""`)
		})
		convey.Convey("data is string", func() {
			result := ObjToString(testHashCodeStr)
			convey.So(result, convey.ShouldEqual, `"`+testHashCodeStr+`"`)
		})
		convey.Convey("data is map", func() {
			data := map[int]int{
				0: 1,
			}
			result := ObjToString(data)
			convey.So(result, convey.ShouldEqual, testMapString)
		})
		convey.Convey("data is function", func() {
			result := ObjToString(func() {})
			convey.So(result, convey.ShouldEqual, "")
		})
	})
}

func TestMaxInt(t *testing.T) {
	convey.Convey("test MaxInt", t, func() {
		convey.Convey("x < y", func() {
			convey.So(MaxInt(0, 1), convey.ShouldEqual, 1)
		})
	})
	convey.Convey("test MaxInt", t, func() {
		convey.Convey("x >= y", func() {
			convey.So(MaxInt(1, 0), convey.ShouldEqual, 1)
		})
	})
}

// TestStringSliceToIntSlice test the func of string slice to int slice
func TestStringSliceToIntSlice(t *testing.T) {
	convey.Convey("test StringSliceToIntSlice ", t, func() {
		convey.Convey("nil slice", func() {
			strSlice := []string{"0", "1", "2"}
			result := StringSliceToIntSlice(strSlice)
			compareIntSliceIsSame(result, []int{0, 1, 2})
		})
		convey.Convey("failed convert str slice to int slice", func() {
			strSlice := []string{"xx", "yy", "zz"}
			result := StringSliceToIntSlice(strSlice)
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func compareIntSliceIsSame(slice1, slice2 []int) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i, _ := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}

func TestMarshalData(t *testing.T) {
	convey.Convey("Test marshalData", t, func() {
		convey.Convey("data is function", func() {
			result := marshalData(func() {})
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("data is string", func() {
			result := marshalData(testHashCodeStr)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestRemoveSliceDuplicateElement(t *testing.T) {
	convey.Convey("Test RemoveSliceDuplicateElement", t, func() {
		mockSlice := []string{"1", "2", "2", "3", "4", "4", "5"}
		result := RemoveSliceDuplicateElement(mockSlice)
		compareStringSliceIsSame(result, []string{"1", "2", "3", "4", "5"})
	})
}

func compareStringSliceIsSame(slice1, slice2 []string) {
	convey.So(len(slice1), convey.ShouldEqual, len(slice2))
	if len(slice1) != len(slice2) {
		return
	}
	for i, _ := range slice1 {
		convey.So(slice1[i], convey.ShouldEqual, slice2[i])
	}
}
