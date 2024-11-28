// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/yaml"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{LogFileName: "../../../testdata/clusterd.log"}
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

	var testDataPath = "../../../testdata/resource/uce_fault_processor_test.yaml"
	var maxFileSize = 10000
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectProcessedDeviceInfos = make(map[string]*constant.DeviceInfo)
	var uceNodesInfos = make(map[string]uceNodeInfo)
	var expectUceJobsInfo = make(map[string]uceJobInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var jobIsUce = make(map[string]map[string]map[string]string)
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

	err = decoder.Decode(&jobIsUce)
	if err != nil {
		err = fmt.Errorf("jobIsUce decode failed")
		goto ReturnLabel
	}

	for key, value := range jobs {
		worker := job.Worker{
			WorkerInfo: job.WorkerInfo{
				CMData: value["CMData"],
			},
			Info: job.Info{
				PGLabels: jobIsUce[key]["PGLabels"],
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
	map[string]job.PodWorker, *reportInfosForAllJobs, error) {

	var testDataPath = "../../../testdata/resource/uce_scenario_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var expectDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var josIsUce = make(map[string]map[string]map[string]string)
	var jobsPodWorkers = make(map[string]job.PodWorker)
	var open *os.File
	var reportInfos reportInfosForAllJobs
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		return cmDeviceInfos, expectDeviceInfos, jobsPodWorkers, &reportInfos, err
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		return cmDeviceInfos, expectDeviceInfos, jobsPodWorkers, &reportInfos, err
	}
	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		return cmDeviceInfos, expectDeviceInfos, jobsPodWorkers, &reportInfos, err
	}
	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)
	return extractContent(decoder, cmDeviceInfos, expectDeviceInfos, jobsPodWorkers, reportInfos, jobs, josIsUce)
}

func extractContent(decoder *yaml.YAMLOrJSONDecoder, cmDeviceInfos map[string]*constant.DeviceInfo,
	expectProcessedDeviceInfos map[string]*constant.DeviceInfo, jobsPodWorkers map[string]job.PodWorker,
	reportInfos reportInfosForAllJobs, jobs map[string]map[string]*job.RankTable,
	josIsUce map[string]map[string]map[string]string) (map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo,
	map[string]job.PodWorker, *reportInfosForAllJobs, error) {
	err := decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
	}

	err = decoder.Decode(&expectProcessedDeviceInfos)
	if err != nil {
		err = fmt.Errorf("expectProcessedDeviceInfos decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
	}

	err = decoder.Decode(&jobs)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
	}

	err = decoder.Decode(&josIsUce)
	if err != nil {
		err = fmt.Errorf("josIsUce decode failed")
		return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
	}

	for key, value := range jobs {
		worker := job.Worker{
			WorkerInfo: job.WorkerInfo{
				CMData: value["CMData"],
			},
			Info: job.Info{PGLabels: josIsUce[key]["PGLabels"]},
		}
		jobsPodWorkers[key] = &worker
	}

	err = decoder.Decode(&reportInfos)
	if err != nil {
		err = fmt.Errorf("reportInfos decode failed")
	}
	return cmDeviceInfos, expectProcessedDeviceInfos, jobsPodWorkers, &reportInfos, err
}

func readObjectFromUceAccompanyProcessorTestYaml() (
	map[string]*constant.DeviceInfo, map[string]*constant.DeviceInfo, error) {

	var testDataPath = "../../../testdata/resource/uce_accompany_processor_test.yaml"
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
	map[string]*constant.DeviceInfo, map[string]job.PodWorker, map[string]JobFaultInfo, error) {

	var testDataPath = "../../../testdata/resource/job_fault_rank_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobs = make(map[string]map[string]*job.RankTable)
	var jobsPodWorkers = make(map[string]job.PodWorker)
	var expectFaultRanks = make(map[string]JobFaultInfo)
	var open *os.File
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	return extractContentForJob(decoder, cmDeviceInfos, jobsPodWorkers, expectFaultRanks, jobs)
}

func extractContentForJob(decoder *yaml.YAMLOrJSONDecoder, cmDeviceInfos map[string]*constant.DeviceInfo,
	jobsPodWorkers map[string]job.PodWorker, expectFaultRanks map[string]JobFaultInfo,
	jobs map[string]map[string]*job.RankTable) (map[string]*constant.DeviceInfo,
	map[string]job.PodWorker, map[string]JobFaultInfo, error) {
	err := decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
	}

	err = decoder.Decode(&jobs)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		return cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err
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
	}
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

func isFaultRankMapEqual(faultRankMap1, faultRankMap2 map[string]JobFaultInfo) bool {
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
