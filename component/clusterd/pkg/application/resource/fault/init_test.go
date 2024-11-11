package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../../testdata/clusterd.log"}
	hwLogConfig.MaxBackups = 30
	hwLogConfig.MaxAge = 7
	hwLogConfig.LogLevel = 0
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}

	code := m.Run()
	os.Exit(code)
}

func readObjectFromUceProcessorTestYaml() (
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

func readObjectFromUceScenarioTestYaml() (
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

func readObjectFromUceAccompanyProcessorTestYaml() (
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

func readObjectFromJobFaultRankTestYaml() (
	map[string]*constant.DeviceInfo, map[string]job.PodWorker, map[string]FaultInfo, error) {

	var testDataPath = "../../../../testdata/resource/job_fault_rank_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var jobsPodWorkers = make(map[string]job.PodWorker)
	var expectFaultRanks = make(map[string]FaultInfo)
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

	err = decoder.Decode(&expectFaultRanks)
	if err != nil {
		err = fmt.Errorf("expectFaultRanks decode failed")
		goto RetureLabel
	}

RetureLabel:
	return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
}

func isSlicesEqual[T comparable](s1, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}
	for _, x := range s1 {
		found := false
		for _, y := range s2 {
			if reflect.DeepEqual(x, y) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func isFaultRankMapEqual(faultRankMap1, faultRankMap2 map[string]FaultInfo) bool {
	if len(faultRankMap1) != len(faultRankMap2) {
		return false
	}
	for jobId, faultRank1 := range faultRankMap1 {
		faultRank2, found := faultRankMap2[jobId]
		if !found {
			return false
		}
		if faultRank1.JobId != faultRank2.JobId {
			return false
		}
		if !isSlicesEqual(faultRank1.FaultList, faultRank2.FaultList) {
			return false
		}
	}
	return true
}
