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

// Package config business configuration service for grpc client
package config

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/config"
)

func fakePublisher() *ConfigPublisher[*config.RankTableStream] {
	ctx := context.Background()
	publisher := NewConfigPublisher[*config.RankTableStream](job1, ctx, constant.RankTableDataType, nil)
	return publisher
}

// TestSelectChanAndContext for test selectChanAndContext
func TestSelectChanAndContext(t *testing.T) {
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
			close(publisher.sendChan)
			ret := publisher.selectChanAndContext(&mockConfigSubscribeRankTableServer{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("04-send rank table success, should return true", func() {
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send", nil)
			defer patch.Reset()
			table := &config.RankTableStream{
				JobId:     job1,
				RankTable: rankTable,
			}
			publisher := fakePublisher()
			publisher.sendChan <- table
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
		convey.Convey("01-send data to client success", func() {
			patch := gomonkey.ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send", nil)
			defer patch.Reset()
			isSent, _ := sendDataToClient[*config.RankTableStream](&mockConfigSubscribeRankTableServer{}, table,
				job1, constant.RankTableDataType)
			convey.So(isSent, convey.ShouldBeTrue)
		})
		convey.Convey("02-send data to client failed", func() {
			patch := gomonkey.ApplyFuncReturn(time.Sleep)
			defer patch.Reset()
			isSent, _ := sendDataToClient[*config.RankTableStream](&mockConfigSubscribeRankTableServer{}, table,
				job1, constant.RankTableDataType)
			convey.So(isSent, convey.ShouldBeFalse)
		})
	})
}

// TestSaveData for test SaveData
func TestSaveData(t *testing.T) {
	convey.Convey("test SaveData", t, func() {
		data := &config.RankTableStream{
			JobId:     job1,
			RankTable: rankTable,
		}
		convey.Convey("01-send on closed channel, should return false", func() {
			publisher := fakePublisher()
			close(publisher.sendChan)
			ret := publisher.SaveData(data)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-send success, should return true", func() {
			publisher := fakePublisher()
			ret := publisher.SaveData(data)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestStop for test stop
func TestStop(t *testing.T) {
	convey.Convey("test stop", t, func() {
		publisher := fakePublisher()
		close(publisher.sendChan)
		stopFunc := func() {
			publisher.sendChan <- &config.RankTableStream{
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
		publisher.SetSubscribe(true)
		convey.So(publisher.IsSubscribed(), convey.ShouldBeTrue)
	})
}

// TestSetAndGetData for test setSentData and GetSentData
func TestSetAndGetData(t *testing.T) {
	convey.Convey("test setSentData and GetSentData", t, func() {
		publisher := fakePublisher()
		data := &config.RankTableStream{JobId: job1, RankTable: rankTable}
		publisher.SetSentData(data)
		convey.So(publisher.GetSentData(), convey.ShouldResemble,
			&config.RankTableStream{JobId: job1, RankTable: rankTable})
	})
}
