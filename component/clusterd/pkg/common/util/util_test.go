// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util test function
package util

import (
	"context"
	"fmt"
	"syscall"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

const (
	testHashCodeStr   = "testHashCodeStr"
	testHashCode      = "9f9046a094b85b781aca276c46c36956ed07f6de1ec1f94268bfce764e0a9606"
	testNilHashCode   = "74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b"
	testHashCodeError = "2343245345"
	testMapString     = `{"0":1}`
)

func init() {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return
	}
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

func TestDeepCopyForObejct(t *testing.T) {
	type args struct {
		Data map[string][]int
		Map  map[string]map[string]string
	}

	t.Run("TestDeepCopy for object", func(t *testing.T) {
		a := args{Data: make(map[string][]int)}
		a.Data["111"] = []int{1, 2, 3, 4}
		a.Data["222"] = []int{5, 6, 7, 8}
		b := new(args)
		if err := DeepCopy(b, a); err != nil {
			t.Errorf("DeepCopy() error = %v", err)
		}
		a.Data["111"] = append(a.Data["111"], 5, 6, 7, 8)
		b.Data["222"] = append(b.Data["222"], 1, 2, 3, 4)

		if len(a.Data["111"]) == len(b.Data["111"]) || len(a.Data["222"]) == len(b.Data["222"]) {
			t.Errorf("DeepCopy() failed")
		}

		a.Map = make(map[string]map[string]string)
		a.Map["111"] = make(map[string]string)
		a.Map["111"]["111"] = "111"
		DeepCopy(b, a)
		b.Map["222"] = make(map[string]string)
		if _, ok := a.Map["222"]; ok {
			t.Errorf("DeepCopy() failed")
		}
	})
}

func TestDeepCopyForMap(t *testing.T) {
	type args struct {
		Data map[string][]int
		Map  map[string]map[string]string
	}

	t.Run("TestDeepCopy for map", func(t *testing.T) {
		a := make(map[string]string)
		a["111"] = "111"
		b := new(map[string]string)
		if err := DeepCopy(b, a); err != nil {
			t.Errorf("DeepCopy() error = %v", err)
		}
		(*b)["222"] = "222"
		if _, ok := a["222"]; ok {
			t.Errorf("DeepCopy() failed")
		}
	})

	t.Run("TestDeepCopy for map with pointer value", func(t *testing.T) {
		a := make(map[string]*args)
		a["111"] = &args{
			Data: make(map[string][]int),
			Map:  make(map[string]map[string]string),
		}
		b := new(map[string]*args)
		if err := DeepCopy(b, a); err != nil {
			t.Errorf("DeepCopy() error = %v", err)
		}
		c := *b
		c["111"].Data["111"] = []int{1, 2, 3, 4}
		if len(a["111"].Data["111"]) > 0 {
			t.Errorf("DeepCopy() failed")
		}
	})
}

func TestRemoveDuplicates(t *testing.T) {
	convey.Convey("test func RemoveDuplicates", t, func() {
		const expLen = 3
		oriSli := []int{0, 2, 1, 2}
		res := RemoveDuplicates(oriSli)
		convey.So(len(res), convey.ShouldEqual, expLen)
	})
}
