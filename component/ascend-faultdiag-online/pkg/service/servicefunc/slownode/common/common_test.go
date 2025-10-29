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

// Package common is a DT collection for fun in common
package common

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/model/slownode"
)

func TestConvertMaptoStruct(t *testing.T) {
	convey.Convey("Test ConvertMaptoStruct", t, func() {
		convey.Convey("test normal case", func() {
			resorce := `{"slownode_default-test-pytorch-2pod-16npu":{"127.0.0.1":{"isSlow":0,"degradationLevel":` +
				`"0.0%","jobName":"default-test-pytorch-2pod-16npu","nodeRank":"127.0.0.1","slowCalculateRanks":` +
				`null,"slowCommunicationDomains":null,"slowSendRanks":null,"slowHostNodes":null,"slowIORanks":null}}}`
			var data = map[string]map[string]any{}
			err := json.Unmarshal([]byte(resorce), &data)
			convey.So(err, convey.ShouldBeNil)
			var result = &slownode.NodeAlgoResult{}
			err = ConvertMaptoStruct(data, result)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result.JobName, convey.ShouldEqual, "default-test-pytorch-2pod-16npu")
		})
		convey.Convey("test error case", func() {
			var data = map[string]map[string]any{}
			var result = &slownode.NodeAlgoResult{}
			err := ConvertMaptoStruct(data, result)
			convey.So(err.Error(), convey.ShouldEqual, "callback data is empty: map[]")

			data = map[string]map[string]any{"test": {}}
			err = ConvertMaptoStruct(data, result)
			convey.So(err.Error(), convey.ShouldEqual, "callback data is empty: map[test:map[]]")
		})
		convey.Convey("test unmarhsal failed", func() {
			data := map[string]map[string]any{"test": {"127.0.0.1": `{`}}
			var result = &slownode.NodeAlgoResult{}
			err := ConvertMaptoStruct(data, result)
			convey.So(err.Error(), convey.ShouldContainSubstring, "json: cannot unmarshal string ")
		})
		convey.Convey("test marshal failed", func() {
			patch := gomonkey.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
				return nil, fmt.Errorf("mock marshal error")
			})
			defer patch.Reset()
			data := map[string]map[string]any{"test": {"127.0.0.1": `{}`}}
			var result = &slownode.NodeAlgoResult{}
			err := ConvertMaptoStruct(data, result)
			convey.So(err.Error(), convey.ShouldContainSubstring, "mock marshal error")
		})
	})
}

func TestAreServersEqual(t *testing.T) {
	convey.Convey("Test AreServersEqual", t, func() {
		testAreServersEqualWithTrue()
		testAreServersEqualWithFalse()
	})
}

func testAreServersEqualWithTrue() {
	convey.Convey("test true case", func() {
		servers1 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
			{Sn: "sn2", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
		}
		servers2 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
			{Sn: "sn2", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
		}
		convey.So(AreServersEqual(servers1, servers2), convey.ShouldBeTrue)

		servers2 = []slownode.Server{
			{Sn: "sn2", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
		}
		convey.So(AreServersEqual(servers1, servers2), convey.ShouldBeTrue)
	})
	convey.Convey("test sn same", func() {
		servers1 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
			{Sn: "sn1", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
		}
		servers2 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
		}
		convey.So(AreServersEqual(servers1, servers2), convey.ShouldBeTrue)
	})
	convey.Convey("test sn, ip sme", func() {
		servers1 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2", "3"}},
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
		}
		servers2 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2", "3"}},
		}
		convey.So(AreServersEqual(servers1, servers2), convey.ShouldBeTrue)
	})
}

func testAreServersEqualWithFalse() {
	convey.Convey("test false case", func() {
		servers1 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
			{Sn: "sn2", Ip: "127.0.0.2", RankIds: []string{"1", "2"}},
		}
		servers2 := []slownode.Server{
			{Sn: "sn1", Ip: "127.0.0.1", RankIds: []string{"1", "2"}},
		}
		convey.So(AreServersEqual(servers1, servers2), convey.ShouldBeFalse)
	})
}

func TestNodeRankValidator(t *testing.T) {
	convey.Convey("Test NodeRankValidator", t, func() {
		convey.Convey("test normal case", func() {
			err := NodeRankValidator("normal-node-rank_123")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test empty node rank", func() {
			err := NodeRankValidator("")
			convey.So(err.Error(), convey.ShouldEqual, "node rank is empty")
		})
		convey.Convey("test invalid character in node rank", func() {
			specialChars := []string{" ", ".", "/", "\\"}
			for _, char := range specialChars {
				nodeRank := fmt.Sprintf("node-rank%s", char)
				err := NodeRankValidator(nodeRank)
				convey.So(err.Error(), convey.ShouldEqual, "contains invalid character: ' ', '.', '/', '\\'")
			}
		})
	})
}

func TestJobIdValidator(t *testing.T) {
	convey.Convey("Test JobIdValidator", t, func() {
		convey.Convey("test normal case", func() {
			err := JobIdValidator("normal-jobid-123")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test empty job id", func() {
			err := JobIdValidator("")
			convey.So(err.Error(), convey.ShouldEqual, "job id is empty")
		})
		convey.Convey("test special character in job id", func() {
			specialChars := []string{" ", "/", "\\"}
			for _, char := range specialChars {
				jobId := fmt.Sprintf("jobid%s", char)
				err := JobIdValidator(jobId)
				convey.So(err.Error(), convey.ShouldEqual, fmt.Sprintf("contains invalid character: %s", char))
			}
		})
	})
}
