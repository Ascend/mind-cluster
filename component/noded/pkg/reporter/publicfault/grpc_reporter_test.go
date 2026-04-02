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
			reporter := NewGrpcReporter()
			convey.So(reporter, convey.ShouldNotBeNil)
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
			}).ApplyMethodReturn(&grpcclient.Client{}, "SendToPubFaultCenter",
				&pubfault.RespStatus{}, errors.New(""))
			reporter.Report(&common.FaultAndConfigInfo{
				PubFaultInfo: &pubfault.PublicFaultRequest{},
			})
			convey.So(clientNewCalled, convey.ShouldBeTrue)
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
