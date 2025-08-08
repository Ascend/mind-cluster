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

// Package model is a DT collection for func in request
package model

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewRequestBodyFromJson(t *testing.T) {
	convey.Convey("Test NewRequestBodyFromJson", t, func() {
		convey.Convey("test normal case", func() {
			jsonStr := `{"component":"noded","request_type":"event","name":"test_event","data":{}}`
			body, err := NewRequestBodyFromJson(jsonStr)
			convey.So(err, convey.ShouldBeNil)
			convey.So(body.Component, convey.ShouldEqual, "noded")
			convey.So(body.RequestType, convey.ShouldEqual, "event")
			convey.So(body.Name, convey.ShouldEqual, "test_event")
		})

		convey.Convey("test error case", func() {
			jsonStr := `{"component":123}` // Invalid JSON
			body, err := NewRequestBodyFromJson(jsonStr)
			convey.So(err.Error(), convey.ShouldEqual,
				"json: cannot unmarshal number into Go struct field Body.component of type string")
			convey.So(body, convey.ShouldBeNil)
		})
	})
}

func TestNewRequestContext(t *testing.T) {
	convey.Convey("Test NewRequestContext", t, func() {
		api := "test_api"
		reqJson := `{"key":"value"}`
		ctx := NewRequestContext(api, reqJson)
		convey.So(ctx.Api, convey.ShouldEqual, api)
		convey.So(ctx.ReqJson, convey.ShouldEqual, reqJson)
		convey.So(ctx.Response, convey.ShouldNotBeNil)
		convey.So(ctx.FinishChan, convey.ShouldNotBeNil)
	})
}
