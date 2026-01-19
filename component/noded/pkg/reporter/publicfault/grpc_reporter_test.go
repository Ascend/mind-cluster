package publicfault

import (
	"context"
	"errors"
	"fmt"
	"nodeD/pkg/common"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/grpcclient"
	"nodeD/pkg/grpcclient/pubfault"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&config, context.Background()); err != nil {
		fmt.Printf("%v", err)
	}
}

func TestNewGrpcReporter(t *testing.T) {
	convey.Convey("Test NewGrpcReporter", t, func() {
		convey.Convey("test success case", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			mockClient := &grpcclient.Client{}
			patches.ApplyFunc(grpcclient.New, func(string) (*grpcclient.Client, error) {
				return mockClient, nil
			})
			reporter := NewGrpcReporter()
			convey.So(reporter.client, convey.ShouldEqual, mockClient)
		})
		convey.Convey("test error case", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(grpcclient.New, func(string) (*grpcclient.Client, error) {
				return nil, errors.New("fake error")
			})
			reporter := NewGrpcReporter()
			convey.So(reporter.client, convey.ShouldBeNil)
		})
	})
}

func TestReport(t *testing.T) {
	convey.Convey("Test Report", t, func() {
		convey.Convey("case fcInfo is nil", func() {
			reporter := &GrpcReporter{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			clientNewCalled := false
			patches.ApplyFunc(grpcclient.New, func(_ string) (*grpcclient.Client, error) {
				clientNewCalled = true
				return &grpcclient.Client{}, nil
			})
			reporter.Report(nil)
			convey.So(clientNewCalled, convey.ShouldBeFalse)
		})
		convey.Convey("case reporter client is nil", func() {
			reporter := &GrpcReporter{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			clientNewCalled := false
			patches.ApplyFunc(grpcclient.New, func(_ string) (*grpcclient.Client, error) {
				clientNewCalled = true
				return &grpcclient.Client{}, nil
			}).ApplyMethodReturn(reporter.client, "SendToPubFaultCenter",
				&pubfault.RespStatus{}, errors.New(""))
			reporter.Report(&common.FaultAndConfigInfo{
				PubFaultInfo: &pubfault.PublicFaultRequest{},
			})
			convey.So(clientNewCalled, convey.ShouldBeTrue)
		})
		convey.Convey("case new client error", func() {
			reporter := &GrpcReporter{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(grpcclient.New, func(_ string) (*grpcclient.Client, error) {
				return nil, errors.New("")
			})
			reporter.Report(&common.FaultAndConfigInfo{
				PubFaultInfo: &pubfault.PublicFaultRequest{},
			})
			convey.So(reporter.client, convey.ShouldBeNil)
		})
	})
}

func TestGrpcReporter_Init(t *testing.T) {
	convey.Convey("Test Init", t, func() {
		reporter := &GrpcReporter{}
		err := reporter.Init()
		convey.So(err, convey.ShouldBeNil)
	})
}
