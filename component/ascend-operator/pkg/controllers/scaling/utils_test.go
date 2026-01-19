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
Package scaling is using for scale AscendJobs.
*/

package scaling

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestConvertRule(t *testing.T) {
	convey.Convey("test convertRule", t, func() {
		convey.Convey("01-rule with invalid group-num should return error", func() {
			_, err := convertRule(&Rule{
				ElasticScalingList: []Item{
					{GroupList: []Group{{GroupName: "group0", GroupNum: "xxx", ServerNumPerGroup: "8"}}},
				}})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-rule with invalid server-group-num should return error", func() {
			_, err := convertRule(&Rule{
				ElasticScalingList: []Item{
					{GroupList: []Group{{GroupName: "group0", GroupNum: "2", ServerNumPerGroup: "xxx"}}},
				}})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-rule with increase group-num should return error", func() {
			_, err := convertRule(&Rule{
				ElasticScalingList: []Item{
					{GroupList: []Group{{GroupName: "group0", GroupNum: "2", ServerNumPerGroup: "8"}}},
					{GroupList: []Group{{GroupName: "group0", GroupNum: "4", ServerNumPerGroup: "8"}}},
				},
			})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-rule with increase server-group-num should return error", func() {
			_, err := convertRule(&Rule{
				ElasticScalingList: []Item{
					{GroupList: []Group{{GroupName: "group0", GroupNum: "2", ServerNumPerGroup: "4"}}},
					{GroupList: []Group{{GroupName: "group0", GroupNum: "1", ServerNumPerGroup: "8"}}},
				},
			})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-invalid rule should return correct result", func() {
			res, err := convertRule(&Rule{
				ElasticScalingList: []Item{
					{GroupList: []Group{{GroupName: "group0", GroupNum: "2", ServerNumPerGroup: "8"}}},
					{GroupList: []Group{{GroupName: "group0", GroupNum: "1", ServerNumPerGroup: "8"}}},
				},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, []map[string]*groupInfo{
				{"group0": {2, 8}},
				{"group0": {1, 8}},
			})
		})
	})
}

func TestGetDestOfCurrentState(t *testing.T) {
	convey.Convey("test getDestOfCurrentState", t, func() {
		convey.Convey("01-rule cover cur state should return write index", func() {
			rule := []map[string]*groupInfo{
				{"group0": {2, 8}, "group1": {1, 8}},
				{"group0": {1, 8}, "group1": {1, 8}},
			}
			cur := map[string]int{"group0": 1, "group1": 1}
			convey.So(getDestOfCurrentState(rule, cur), convey.ShouldEqual, 0)
		})
		convey.Convey("02-rule can not cover cur state should return write index", func() {
			rule := []map[string]*groupInfo{
				{"group0": {2, 8}, "group1": {1, 8}},
				{"group0": {1, 8}, "group1": {1, 8}},
			}
			cur := map[string]int{"group0": 2, "group1": 1}
			convey.So(getDestOfCurrentState(rule, cur), convey.ShouldEqual, invalidIndex)
		})
	})
}
