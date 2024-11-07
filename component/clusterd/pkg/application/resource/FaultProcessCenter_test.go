package resource

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"fmt"
	rand2 "golang.org/x/exp/rand"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"reflect"
	"testing"
)

const JobReportRecoverTimeout = int64(10)
const JobReportCompleteTimeout = int64(30)
const UceFaultTime = int64(100)

func getUceFaultProcessor(center *FaultProcessCenter) (*uceFaultProcessor, error) {
	for _, processor := range center.deviceFaultProcessor {
		if uceProcessor, ok := processor.(*uceFaultProcessor); ok {
			return uceProcessor, nil
		}
	}
	return nil, fmt.Errorf("can not find uceFaultProcessor in FaultProcessCenter")
}

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

	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation1", func(t *testing.T) {
			processor, err := getUceFaultProcessor(faultProcessCenter)
			if err != nil {
				t.Errorf("%v", err)
			}
			var currentTime int64
			if i == 0 {
				currentTime = UceFaultTime
			} else if i == 1 {
				currentTime = UceFaultTime + JobReportRecoverTimeout
			} else {
				currentTime = rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1)
			}
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  JobNotRecover,
				CompleteTime: JobNotRecoverComplete,
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
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive job recover info
	t.Run("canFilterUceDeviceFaultInfoSituation2", func(t *testing.T) {
		currentTime := UceFaultTime + JobReportRecoverTimeout + 1
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  JobNotRecover,
			CompleteTime: JobNotRecoverComplete,
		}
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// receive recover info, but exceed (UceFaultTime + JobReportRecoverTimeout)
	t.Run("canFilterUceDeviceFaultInfoSituation1", func(t *testing.T) {
		currentTime := UceFaultTime + JobReportRecoverTimeout + rand2.Int63n(10)
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  UceFaultTime + JobReportRecoverTimeout + 1, // exceed 1s
			CompleteTime: JobNotRecoverComplete,
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
	want := true
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation3", func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
				CompleteTime: JobNotRecoverComplete,
			}
			var currentTime int64
			if i == 0 {
				currentTime = uceDevice.RecoverTime
			} else if i == 1 {
				currentTime = uceDevice.RecoverTime + JobReportCompleteTimeout
			} else {
				currentTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
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
	want := false
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	// don't receive recover complete info
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation4", func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
				CompleteTime: JobNotRecoverComplete,
			}
			currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}

	// receive recover complete info, but exceed (RecoverTime + JobReportCompleteTimeout)
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation1", func(t *testing.T) {
			processor := faultProcessCenter.deviceFaultProcessor[0].(*uceFaultProcessor)
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
				CompleteTime: JobReportRecoverTimeout + 1, // exceed 1s
			}
			currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
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
	want := true
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation5", func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:   "test-device",
				FaultTime:    UceFaultTime,
				RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
				CompleteTime: JobNotRecoverComplete,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.RecoverTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.RecoverTime, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
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
	want := false
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	// CompleteTime smaller than FaultTime
	t.Run("canFilterUceDeviceFaultInfoSituation6-1", func(t *testing.T) {
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
			CompleteTime: UceFaultTime - 1,
		}

		currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})

	// CompleteTime smaller than RecoverTime
	t.Run("canFilterUceDeviceFaultInfoSituation6-2", func(t *testing.T) {
		processor := faultProcessCenter.deviceFaultProcessor[0].(*uceFaultProcessor)
		uceDevice := uceDeviceInfo{
			DeviceName:   "test-device",
			FaultTime:    UceFaultTime,
			RecoverTime:  rand.Int63nRange(UceFaultTime, UceFaultTime+JobReportRecoverTimeout+1),
			CompleteTime: JobNotRecoverComplete,
		}
		uceDevice.CompleteTime = uceDevice.RecoverTime - 1

		currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
		if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
			t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
		}
	})
}

// Test current time exceed (UceFaultTime + JobReportRecoverTimeout),
// and exceed (RecoverTime + JobReportCompleteTimeout)
// and receive job recover info, and receive job recover complete info, and complete time valid, should filter
func Test_uceFaultProcessor_canFilterUceDeviceFaultInfoSituation7(t *testing.T) {
	want := true
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation7-1", func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - 1,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + JobReportCompleteTimeout
			} else {
				uceDevice.CompleteTime = rand.Int63nRange(uceDevice.FaultTime, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
			}
			currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
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
	want := false
	processor, err := getUceFaultProcessor(faultProcessCenter)
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		t.Run("canFilterUceDeviceFaultInfoSituation7-2", func(t *testing.T) {
			uceDevice := uceDeviceInfo{
				DeviceName:  "test-device",
				FaultTime:   UceFaultTime,
				RecoverTime: UceFaultTime - 31,
			}
			if i == 0 {
				uceDevice.CompleteTime = uceDevice.FaultTime
			} else if i == 1 {
				uceDevice.CompleteTime = uceDevice.RecoverTime + JobReportCompleteTimeout
			}
			currentTime := max(UceFaultTime+JobReportRecoverTimeout+1, uceDevice.RecoverTime+JobReportCompleteTimeout+1)
			if got := processor.canFilterUceDeviceFaultInfo(uceDevice, currentTime); got != want {
				t.Errorf("canFilterUceDeviceFaultInfo() = %v, want %v", got, want)
			}
		})
	}
}

// =============Test canFilterUceDeviceFaultInfo===========
func readObjectFromBaseTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]uceNodeInfo, map[string]job.PodWorker, map[string]uceJobInfo, error) {

	var testDataPath = "../../../testdata/resource/base_test.yaml"
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

	var testDataPath = "../../../testdata/resource/uce_scenario.yaml"
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
	t.Run("getUceDeviceOfNodes", func(t *testing.T) {
		cmDeviceInfos, _, uceNodesInfos, _, _, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		faultProcessCenter.deviceInfos = cmDeviceInfos
		processor, err := getUceFaultProcessor(faultProcessCenter)
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceOfNodes := processor.getUceDeviceOfNodes()
		if !reflect.DeepEqual(deviceOfNodes, uceNodesInfos) {
			t.Errorf("getUceDeviceOfNodes() = %v, want %v", deviceOfNodes, uceNodesInfos)
		}
	})
}

func Test_uceFaultProcessor_getUceDevicesForUceTolerateJobs(t *testing.T) {
	t.Run("getUceDevicesForUceTolerateJobs", func(t *testing.T) {
		cmDeviceInfos, _, _, jobsPodWorkers, expectUceJobsInfo, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		faultProcessCenter.deviceInfos = cmDeviceInfos
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor, err := getUceFaultProcessor(faultProcessCenter)
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		if !reflect.DeepEqual(processor.uceDevicesOfUceJob, expectUceJobsInfo) {
			t.Errorf("getUceDevicesForUceTolerateJobs() = %v, want %v",
				processor.uceDevicesOfUceJob, expectUceJobsInfo)
		}
	})
}

func Test_uceFaultProcessor_processUceFaultInfo(t *testing.T) {
	t.Run("processUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, _, jobsPodWorkers, _, testFileErr := readObjectFromBaseTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		faultProcessCenter.deviceInfos = cmDeviceInfos
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor, err := getUceFaultProcessor(faultProcessCenter)
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		currentTime := int64(109)
		processUceFaultInfo := processor.processUceFaultInfo(cmDeviceInfos, currentTime)
		if !reflect.DeepEqual(processUceFaultInfo, expectProcessedDeviceInfos) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				processUceFaultInfo, expectProcessedDeviceInfos)
		}
	})
}

func Test_uceFaultProcessor_Scenario1(t *testing.T) {
	t.Run("processUceFaultInfo", func(t *testing.T) {
		cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, reportInfos, testFileErr := readObjectFromScenarioTestYaml()
		if testFileErr != nil {
			t.Errorf("init data failed. %v", testFileErr)
		}
		faultProcessCenter.deviceInfos = cmDeviceInfos

		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor, err := getUceFaultProcessor(faultProcessCenter)
		if err != nil {
			t.Errorf("%v", err)
		}
		processor.mindIoReportInfo = reportInfos
		processor.uceDeviceOfNode = processor.getUceDeviceOfNodes()
		processor.uceDevicesOfUceJob = processor.getUceDevicesForUceTolerateJobs()
		currentTime := int64(100)
		processUceFaultInfo := processor.processUceFaultInfo(cmDeviceInfos, currentTime)

		if len(processUceFaultInfo) != len(expectProcessedDeviceInfos) {
			t.Errorf("processUceFaultInfo() = %v, want %v",
				util.ObjToString(processUceFaultInfo), util.ObjToString(expectProcessedDeviceInfos))
		}

		for nodeName, expect := range expectProcessedDeviceInfos {
			actual, ok := processUceFaultInfo[nodeName]
			if !ok {
				t.Errorf("processUceFaultInfo() = %v, want %v",
					util.ObjToString(processUceFaultInfo), util.ObjToString(expectProcessedDeviceInfos))
			}
			if !reflect.DeepEqual(device.GetFaultMap(actual), device.GetFaultMap(expect)) {
				t.Errorf("processUceFaultInfo() = %v, want %v",
					util.ObjToString(processUceFaultInfo), util.ObjToString(expectProcessedDeviceInfos))
			}
		}
	})
}
