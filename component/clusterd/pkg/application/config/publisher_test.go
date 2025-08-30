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
	"errors"
	"sort"
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
			publisher.sendChan <- &jobDataForChan[*config.RankTableStream]{
				jobId: job1, data: table,
			}
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
			isSent, _ := sendDataToClient[*config.RankTableStream](&mockConfigSubscribeRankTableServer{}, table,
				job1, constant.RankTableDataType)
			convey.So(isSent, convey.ShouldBeTrue)
		})
		convey.Convey("02-send data to client failed", func() {
			patch := gomonkey.ApplyFuncReturn(time.Sleep).
				ApplyMethodReturn(&mockConfigSubscribeRankTableServer{}, "Send", errors.New("fake err"))
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
			ret := publisher.SaveData(job1, data)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-send success, should return true", func() {
			publisher := fakePublisher()
			ret := publisher.SaveData(job1, data)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestStop for test stop
func TestStop(t *testing.T) {
	convey.Convey("test stop", t, func() {
		publisher := fakePublisher()
		publisher.Stop()
		stopFunc := func() {
			publisher.sendChan <- &jobDataForChan[*config.RankTableStream]{
				jobId: job1, data: &config.RankTableStream{
					JobId:     job1,
					RankTable: rankTable,
				},
			}
		}
		publisher.Stop()
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
		publisher.SetSentData(job1, data)
		convey.So(publisher.GetSentData(job1), convey.ShouldResemble,
			&config.RankTableStream{JobId: job1, RankTable: rankTable})
	})
}

func TestGetCreateTime(t *testing.T) {
	publisher := fakePublisher()
	convey.Convey("test getCreateTime", t, func() {
		convey.So(publisher.GetCreateTime().Before(time.Now()), convey.ShouldBeTrue)
	})
}

func TestGetAllSentJobIdListAndClearDeletedJobIdList(t *testing.T) {
	publisher := fakePublisher()
	publisher.SaveData(job1, &config.RankTableStream{})
	publisher.SaveData(job2, &config.RankTableStream{})
	convey.Convey("test GetAllSentJobIdList", t, func() {
		ret := publisher.GetAllSentJobIdList()
		sort.Strings(ret)
		convey.ShouldResemble(ret, []string{job1, job2})
	})
	convey.Convey("test ClearDeletedJobIdList", t, func() {
		publisher.ClearDeletedJobIdList([]string{job1})
		ret := publisher.GetAllSentJobIdList()
		convey.ShouldResemble(ret, []string{job2})
	})
}
