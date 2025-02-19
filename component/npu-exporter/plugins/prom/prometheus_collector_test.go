package prom

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/metrics"
)

func initChain() {
	common.ChainForSingleGoroutine = []common.MetricsCollector{
		&metrics.BaseInfoCollector{},
		&metrics.SioCollector{},
		&metrics.VersionCollector{},
		&metrics.HbmCollector{},
		&metrics.DdrCollector{},
		&metrics.VnpuCollector{},
		&metrics.PcieCollector{},
	}
	common.ChainForMultiGoroutine = []common.MetricsCollector{
		&metrics.NetworkCollector{},
		&metrics.RoceCollector{},
		&metrics.OpticalCollector{},
	}
}

func TestDescribe(t *testing.T) {

	convey.Convey("test prometheus desc ", t, func() {
		initChain()

		collector := NewPrometheusCollector(nil)
		ch := make(chan *prometheus.Desc, 1000)

		collector.Describe(ch)
		t.Logf("Describe len(ch):%v", len(ch))

		convey.So(ch, convey.ShouldNotBeEmpty)
	})
}

/*func TestCollect(t *testing.T) {

}
*/
