package fault

import (
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"fmt"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"testing"
	"time"
)

// ======= Test uceAccompanyFaultProcessor
func readObjectFromUceAccompanyTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo, error) {

	var testDataPath = "../../../../testdata/resource/uce_accompany_processor_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var open *os.File
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		goto RetureLabel
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		goto RetureLabel
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		goto RetureLabel
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	err = decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		goto RetureLabel
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		goto RetureLabel
	}

RetureLabel:
	return cmDeviceInfos, expectProcessedDeviceInfos, err
}
func Test_uceAccompanyFaultProcessor_process(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceAccompanyFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceAccompanyFaultProcessor_process", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, testFileErr := readObjectFromUceAccompanyTestYaml()
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
	})
}
