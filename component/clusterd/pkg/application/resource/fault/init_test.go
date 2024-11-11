package fault

import (
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}

	code := m.Run()
	os.Exit(code)
}
