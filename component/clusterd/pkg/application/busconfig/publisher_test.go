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

// Package busconfig business configuration service for grpc client
package busconfig

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/interface/grpc/config"
)

func fakePublisher() *ConfigPublisher {
	ctx := context.Background()
	publisher := NewConfigPublisher(job1, ctx)
	return publisher
}

// TestSelectChanAndContext for test selectChanAndContext
func TestSelectChanAndContextCase1(t *testing.T) {
	convey.Convey("test selectChanAndContext", t, func() {
		convey.Convey("01-context canceled, should return false", func() {
			publisher := fakePublisher()
			publisher.ctxCancelFunc()
			ret := publisher.selectChanAndContext(&mockConfigSubscribeRankTableServer{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-stream context canceled, should return false", func() {
			publisher := fakePublisher()
			ctx, cancel := context.WithCancel(context.Background())
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Context", ctx)
			defer patch.Reset()
			cancel()
			ret := publisher.selectChanAndContext(&mockConfigSubscribeRankTableServer{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("03-rankTableChan closed, should return false", func() {
			publisher := fakePublisher()
			close(publisher.rankTableChan)
			ret := publisher.selectChanAndContext(&mockConfigSubscribeRankTableServer{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("04-send rank table success, should return true", func() {
			patch := gomonkey.ApplyFuncReturn(sendRankTable)
			defer patch.Reset()
			table := &config.RankTableStream{
				JobId:     job1,
				RankTable: rankTable,
			}
			publisher := fakePublisher()
			publisher.rankTableChan <- table
			ret := publisher.
				selectChanAndContext(&mockConfigSubscribeRankTableServer{})
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestSendRankTable for test sendRankTable
func TestSendRankTable(t *testing.T) {
	convey.Convey("test sendRankTable", t, func() {
		table := &config.RankTableStream{
			JobId:     job1,
			RankTable: rankTable,
		}
		convey.Convey("01-send rank table success", func() {
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send", nil)
			defer patch.Reset()
			sendRankTable(&mockConfigSubscribeRankTableServer{}, table)
		})
		convey.Convey("02-send rank table failed", func() {
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send",
				errors.New("connect failed")).
				ApplyFuncReturn(time.Sleep)
			defer patch.Reset()
			sendRankTable(&mockConfigSubscribeRankTableServer{}, table)
		})
	})
}

// TestSaveData for test SaveData
func TestSaveData(t *testing.T) {
	convey.Convey("test SaveData", t, func() {
		convey.Convey("01-send on closed channel, should return false", func() {
			publisher := fakePublisher()
			publisher.stop()
			ret := publisher.SaveData(job1, rankTable)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-send success, should return true", func() {
			publisher := fakePublisher()
			ret := publisher.SaveData(job1, rankTable)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestStop for test stop
func TestStop(t *testing.T) {
	convey.Convey("test stop", t, func() {
		publisher := fakePublisher()
		publisher.stop()
		stopFunc := func() {
			publisher.rankTableChan <- &config.RankTableStream{
				JobId:     job1,
				RankTable: rankTable,
			}
		}
		convey.So(stopFunc, convey.ShouldPanic)
	})
}

// TestSetSubscribeAndIsSubscribed for test setSubscribe and IsSubscribed
func TestSetSubscribeAndIsSubscribed(t *testing.T) {
	convey.Convey("test setSubscribe and IsSubscribed", t, func() {
		publisher := fakePublisher()
		publisher.setSubscribe(true)
		convey.So(publisher.isSubscribed(), convey.ShouldBeTrue)
	})
}
