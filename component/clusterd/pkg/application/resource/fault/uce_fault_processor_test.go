package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"reflect"
	"testing"
	"time"
)

const UceFaultTime = int64(100 * time.Second)

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// =============Test canFilterUceDeviceFaultInfo===========
// Test current time not exceed (UceFaultTime + JobReportRecoverTimeout), should filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation1(t *testing.T) {
	want := true
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation1-"+string(rune(i)), func(t *testing.T) {
			var currentTime int64
			if i == 0 {
				currentTime = UceFaultTime
			} else if i == 1 {
				currentTime = UceFaultTime + constant.JobReportRecoverTimeout
			} else {
				currentTime = rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1)
			}
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  constant.JobNotRecover,
				CompleteTime: constant.JobNotRecoverComplete,
			}
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v. uceDevice = %v, currentTime = %v",
					got, want, uceDevice, currentTime)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// but don't receive job recover info, should not filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation2(t *testing.T) {
	want := false
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive job recover info
	t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation2-1", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1 // exceed 1ns
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  constant.JobNotRecover,
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// receive recover info, but exceed (UceFaultTime + JobReportRecoverTimeout)
	t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation2-2", func(t *testing.T) {
		currentTime := UceFaultTime + constant.JobReportRecoverTimeout + 1
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  UceFaultTime + constant.JobReportRecoverTimeout + 1, // exceed 1ns
			CompleteTime: constant.JobNotRecoverComplete,
		}
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// but not exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, should filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation3(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation3-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			var currentTime int64
			if i == 0 {
				currentTime = uceDevice.RecoverTime
			} else if i == 1 {
				currentTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				currentTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but don't receive job recover complete info, should not filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive recover complete info
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4-1-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}

	// receive recover complete info, but exceed (RecoverTime + JobReportCompleteTimeout)
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation4-2-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobReportRecoverTimeout + 1, // exceed 1s
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, should filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation5(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation5-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
				CompleteTime: constant.JobNotRecoverComplete,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.RecoverTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, but receive invalid job recover complete info, should not filter uce fault
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation6(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	// CompleteTime smaller than FaultTime
	t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation6-1", func(t *testing.T) {
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: UceFaultTime - 1,
		}

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// CompleteTime smaller than RecoverTime
	t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation6-2", func(t *testing.T) {
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+constant.JobReportRecoverTimeout+1),
			CompleteTime: constant.JobNotRecoverComplete,
		}
		uceDevice.CompleteTime = uceDevice.RecoverTime - 1

		currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid, should filter
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation7(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := true
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation7-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - 1,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.FaultTime, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid,
// but complete time - recover time exceeds  JobReportCompleteTimeout, should not filter
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation8(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	want := false
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation8-"+string(rune(i)), func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - int64(31*time.Second),
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + constant.JobReportCompleteTimeout
			}
			currentTime := max(UceFaultTime+constant.JobReportRecoverTimeout+1, uceDevice.RecoverTime+constant.JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// ========= Test mindio callback ===========
func Test_uceFaultProcessor_CallbackForReportUceInfo(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceFaultProcessor_CallbackForReportUceInfo", func(t *testing.T) {
		_, _, _, jobsPodWorkers, _, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		if err != nil {
			t.Errorf("%v", err)
		}
		type testsuit struct {
			jobId       string
			rankId      string
			recoverTime int64
		}
		ts1Success := testsuit{
			jobId:       "job1",
			rankId:      "1",
			recoverTime: 10,
		}

		ts2Fault := testsuit{
			jobId:       "job2",
			rankId:      "9",
			recoverTime: 10,
		}
		defer func() {
			processor.mindIoReportInfo.Infos = make(map[string]map[string]map[string]mindIoReportInfo)
		}()

		err = deviceFaultProcessCenter.CallbackForReportUceInfo(ts1Success.jobId, ts1Success.rankId, ts1Success.recoverTime)
		nodeName, deviceId, _ := kube.JobMgr.GetNodeAndDeviceFromJobIdAndRankId(ts1Success.jobId, ts1Success.rankId)
		deviceName := deviceID2DeviceKey(deviceId)
		info, ok := processor.mindIoReportInfo.Infos[ts1Success.jobId][nodeName][deviceName]
		if err != nil || !ok || info.RecoverTime != ts1Success.recoverTime {
			t.Errorf("test CallbackForReportUceInfo success failed.")
		}
		err = deviceFaultProcessCenter.CallbackForReportUceInfo(ts2Fault.jobId, ts2Fault.rankId, ts2Fault.recoverTime)
		if err == nil {
			t.Errorf("test CallbackForReportUceInfo fault failed.")
		}
	})
}

// =============Test scenario===========
func readObjectFromBaseTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]uceNodeInfo, map[string]job.PodWorker, map[string]uceJobInfo, error) {

	var testDataPath = "../../../../testdata/resource/uce_processor_test.yaml"
	var maxFileSize = 10000
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var uceNodesInfos = make(map[string]uceNodeInfo)
	var expectUceJobsInfo = make(map[string]uceJobInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var jobsPodWorkers = make(map[string]job.PodWorker)
	var open *os.File

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		goto ReturnLabel
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		goto ReturnLabel
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		goto ReturnLabel
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	err = decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		goto ReturnLabel
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		goto ReturnLabel
	}

	err = decoder.Decode(&uceNodesInfos)
	if err != nil {
		err = fmt.Errorf("uceNodesInfos decode failed")
		goto ReturnLabel
	}

	err = decoder.Decode(&jobs)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		goto ReturnLabel
	}

	for key, value := range jobs {
		worker := job.Worker{
			WorkerInfo: job.WorkerInfo{
				CMData: value["CMData"],
			},
		}
		jobsPodWorkers[key] = &worker
	}

	err = decoder.Decode(&expectUceJobsInfo)
	if err != nil {
		err = fmt.Errorf("expectUceJobsInfo decode failed")
		goto ReturnLabel
	}

ReturnLabel:
	return cmDeviceInfos, expectProcessedDeviceInfos, uceNodesInfos, jobsPodWorkers, expectUceJobsInfo, err
}

func readObjectFromScenarioTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]job.PodWorker, *mindIoReportInfosForAllJobs, error) {

	var testDataPath = "../../../../testdata/resource/uce_scenario.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var jobsPodWorkers = make(map[string]job.PodWorker)
	var open *os.File
	var reportInfos mindIoReportInfosForAllJobs
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

	err = decoder.Decode(&jobs)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		goto RetureLabel
	}

	for key, value := range jobs {
		worker := job.Worker{
			WorkerInfo: job.WorkerInfo{
				CMData: value["CMData"],
			},
		}
		jobsPodWorkers[key] = &worker
	}

	err = decoder.Decode(&reportInfos)
	if err != nil {
		err = fmt.Errorf("reportInfos decode failed")
		goto RetureLabel
	}

RetureLabel:
	return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
}

func Test_uceFaultProcessor_getUceDeviceOfNodes(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceFaultProcessor_getUceDeviceOfNodes", func(t *testing.T) {
		cmDeviceInfos, _, uceNodesInfos, _, _, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.SetDeviceInfos(cmDeviceInfos)
		deviceInfos := deviceFaultProcessCenter.GetDeviceInfos()
		deviceOfNodes := processor.getUceDeviceOfNodes(deviceInfos)
		if !reflect.DeepEqual(deviceOfNodes, uceNodesInfos) {
			t.Errorf("getUceDeviceOfNodes() = %v, want %v",
				util.ObjToString(deviceOfNodes), util.ObjToString(uceNodesInfos))
		}
	})
}

func Test_uceFaultProcessor_getUceDevicesForUceTolerateJobs(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceFaultProcessor_getUceDevicesForUceTolerateJobs", func(t *testing.T) {
		cmDeviceInfos, _, _, jobsPodWorkers, expectUceJobsInfo, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.SetDeviceInfos(cmDeviceInfos)
		deviceInfos := deviceFaultProcessCenter.GetDeviceInfos()
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs(deviceInfos)
		if !reflect.DeepEqual(processor.uceDevicesOfUceJob, expectUceJobsInfo) {
			t.Errorf("getUceDevicesForUceTolerateJobs() = %v, want %v",
				util.ObjToString(processor.uceDevicesOfUceJob), util.ObjToString(expectUceJobsInfo))
		}
	})
}

func isEqualFaultInfos(one, other map[string]*constant.DeviceInfo) bool {
	if len(one) != len(other) {
		return false
	}

	for nodeName, expect := range one {
		actual, ok := other[nodeName]
		if !ok {
			return false
		}
		if !reflect.DeepEqual(device.GetFaultMap(actual), device.GetFaultMap(expect)) {
			return false
		}
	}
	return true
}
func Test_uceFaultProcessor_processUceFaultInfo(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceFaultProcessor_processUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, _, jobsPodWorkers, _, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.SetDeviceInfos(cmDeviceInfos)
		deviceInfos := deviceFaultProcessCenter.GetDeviceInfos()
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs(deviceInfos)
		currentTime := 109 * time.Second.Milliseconds()
		processUceFaultInfo := processor.processUceFaultInfo(deviceInfos, currentTime)
		if !isEqualFaultInfos(processUceFaultInfo, expectProcessedDeviceInfos) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(processUceFaultInfo), util.ObjToString(expectProcessedDeviceInfos))
		}
	})
}

func Test_uceFaultProcessor_Scenario1(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetUceFaultProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Run("Test_uceFaultProcessor_Scenario1", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, reportInfos, testFileErr := readObjectFromScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.SetDeviceInfos(cmDeviceInfos)
		processor.mindIoReportInfo = reportInfos
		deviceInfos := deviceFaultProcessCenter.GetDeviceInfos()
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes(deviceInfos)
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs(deviceInfos)
		currentTime := 100 * time.Second.Milliseconds()
		processUceFaultInfo := processor.processUceFaultInfo(deviceInfos, currentTime)
		if !isEqualFaultInfos(processUceFaultInfo, expectProcessedDeviceInfos) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(processUceFaultInfo), util.ObjToString(expectProcessedDeviceInfos))
		}
	})
}
