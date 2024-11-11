package fault

import (
	"clusterd/pkg/common/util"
	"testing"
	"time"
)

// ======= Test uceAccompanyFaultProcessor

func Test_uceAccompanyFaultProcessor_process(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceAccompanyFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceAccompanyFaultProcessor_process", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyProcessorTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.uceAccompanyFaultInQue(cmDeviceInfos)
		currentTime := 95 * time.Second.Milliseconds()
		filteredFaultInfos := processor.filterFaultInfos(currentTime, cmDeviceInfos)
		if !isEqualFaultInfos(filteredFaultInfos, expectProcessedDeviceInfos) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(deviceFaultProcessCenter.GetDeviceInfos()), util.ObjToString(expectProcessedDeviceInfos))
		}

		if len(processor.uceAccompanyFaultQue["node1"]["Ascend910-1"]) != 1 &&
			processor.uceAccompanyFaultQue["node1"]["Ascend910-1"][0].FaultCode == "80C98009" {
			t.Errorf("processor.uceAccompanyFaultQue() is wrong")
		}
	})
}
