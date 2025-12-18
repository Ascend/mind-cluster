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

// Package utils is a DT collection for func in storage.go
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/model"
)

func TestStorage(t *testing.T) {
	convey.Convey("test Storage by string", t, func() {
		storage := NewStorage[string]()
		// store data
		var key = "testKey"
		var value = "testValue"
		storage.Store(key, value)
		convey.So(storage.Len(), convey.ShouldEqual, 1)
		// load
		res, ok := storage.Load(key)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(res, convey.ShouldEqual, value)
		// load non-existed key
		res, ok = storage.Load("non-existed-key")
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeEmpty)
		// clear
		storage.Clear()
		// load the existed key
		res, ok = storage.Load(key)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeEmpty)
	})
	convey.Convey("test Storage by job", t, func() {
		storage := NewStorage[*model.JobSummary]()
		// store data
		var key = "testKey"
		var value = &model.JobSummary{}
		storage.Store(key, value)
		// load
		res, ok := storage.Load(key)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(res, convey.ShouldEqual, value)
		// load non-existed key
		res, ok = storage.Load("non-existed-key")
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeNil)
		// clear
		storage.Delete(key)
		// load the existed key
		res, ok = storage.Load(key)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(res, convey.ShouldBeNil)
	})
}

func TestRange(t *testing.T) {
	convey.Convey("test Storage by string", t, func() {
		const (
			key1 = "key1"
			key2 = "key2"
			val  = "value1"
		)
		storage := NewStorage[string]()
		storage.Store(key1, val)
		storage.Store(key2, val)
		cnt := 0
		storage.Range(func(key, value string) bool {
			cnt++
			return value == val
		})
		convey.So(cnt, convey.ShouldEqual, storage.Len())
	})
}
