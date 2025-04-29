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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/config"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	job1      = "job1"
	job2      = "job2"
	rankTable = "rankTable"
)

func fakeService() *BusinessConfigServer {
	ctx := context.Background()
	return NewBusinessConfigServer(ctx)
}

// TestRankTableChange for test rankTableChange
func TestRankTableChange(t *testing.T) {
	convey.Convey("test rankTableChange", t, func() {
		convey.Convey("01-publisher not exist, should save cache and not save work queue", func() {
			service := fakeService()
			resent, err := service.rankTableChange(job1, rankTable)
			convey.So(resent, convey.ShouldBeTrue)
			convey.So(err, convey.ShouldResemble, errors.New("job not registered or not subscribed"))
		})
		convey.Convey("02-publisher exist, should save channel", func() {
			service := fakeService()
			service.addPublisher(job1)
			publisher, _ := service.getPublisher(job1)
			publisher.setSubscribe(true)
			resent, err := service.rankTableChange(job1, rankTable)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resent, convey.ShouldBeFalse)
			publisher, _ = service.getPublisher(job1)
			data, ok := <-publisher.rankTableChan
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(data.RankTable, convey.ShouldEqual, rankTable)
			close(publisher.rankTableChan)
		})
		data := &config.RankTableStream{}
		hwlog.RunLog.Infof("send rank table success, data=%v", data)
	})
}

// TestRegister for test Register
func TestRegister(t *testing.T) {
	convey.Convey("test Register", t, func() {
		service := fakeService()
		service.configPublisher[job1] = NewConfigPublisher(job1, context.Background())
		convey.Convey("01-publisher not exist, should register", func() {
			req := &config.ClientInfo{
				JobId: job2,
				Role:  "",
			}
			status, err := service.Register(nil, req)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldNotBeNil)
			convey.So(service.configPublisher[job2], convey.ShouldNotBeNil)
		})
	})
}

func TestSubscribeRankTable(t *testing.T) {
	convey.Convey("test rankTableChange", t, func() {
		req := &config.ClientInfo{
			JobId: job1,
			Role:  "",
		}
		patch := gomonkey.ApplyPrivateMethod(&ConfigPublisher{}, "listenRankTableChange",
			func(*ConfigPublisher, config.Config_SubscribeRankTableServer) { return })
		defer patch.Reset()
		convey.Convey("01-publisher not exist, should return error", func() {
			service := fakeService()
			err := service.SubscribeRankTable(req, nil)
			convey.So(err, convey.ShouldResemble, errors.New("jobId=job1 not registered, role="))
		})
		convey.Convey("02-subscribe rank table service success, should return nil", func() {
			service := fakeService()
			service.addPublisher(job1)
			err := service.SubscribeRankTable(req, &mockConfigSubscribeRankTableServer{})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetPublisher for test getPublisher
func TestGetPublisher(t *testing.T) {
	convey.Convey("test getPublisher", t, func() {
		service := fakeService()
		publisher, ok := service.getPublisher(job1)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(publisher, convey.ShouldBeNil)
		service.configPublisher[job1] = NewConfigPublisher(job1, context.Background())
		publisher, ok = service.getPublisher(job1)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(publisher, convey.ShouldNotBeNil)
	})
}

// TestDeletePublisher for test deletePublisher
func TestDeletePublisher(t *testing.T) {
	convey.Convey("test deletePublisher", t, func() {
		service := fakeService()
		service.configPublisher[job1] = NewConfigPublisher(job1, context.Background())
		publisher, _ := service.getPublisher(job1)
		service.deletePublisher(job1, publisher.getCreateTime())
		publisher, ok := service.getPublisher(job1)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(publisher, convey.ShouldBeNil)
	})
}

// TestAddPublisher for test addPublisher
func TestAddPublisher(t *testing.T) {
	convey.Convey("test addPublisher", t, func() {
		service := fakeService()
		service.addPublisher(job1)
		publisher, ok := service.getPublisher(job1)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(publisher, convey.ShouldNotBeNil)
	})
}

type mockConfigSubscribeRankTableServer struct {
	grpc.ServerStream
}

func (x *mockConfigSubscribeRankTableServer) Send(m *config.RankTableStream) error {
	return errors.New("send failed")
}

func (x *mockConfigSubscribeRankTableServer) Context() context.Context {
	return context.Background()
}
