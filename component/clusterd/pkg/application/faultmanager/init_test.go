// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/yaml"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/collector"
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
	collector.InitCmCollectBuffer()
	NewFaultProcessCenter()
	code := m.Run()
	os.Exit(code)
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
	map[string]*constant.DeviceInfo, constant.JobServerInfoMap, map[string]JobFaultInfo, error) {

	var testDataPath = "../../../testdata/resource/job_fault_rank_test.yaml"
	var cmDeviceInfos = make(map[string]*constant.DeviceInfo)
	var err error
	var fileSize int64
	var decoder *yaml.YAMLOrJSONDecoder
	var jobDevices = make(map[string]map[string]constant.ServerHccl)
	var jobServerInfoMap = constant.JobServerInfoMap{}
	var expectFaultRanks = make(map[string]JobFaultInfo)
	var open *os.File
	maxFileSize := 10000

	fileInfo, err := os.Stat(testDataPath)
	if err != nil {
		err = fmt.Errorf("testDataPath invalid")
		return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
	}
	fileSize = fileInfo.Size()
	if fileSize > int64(maxFileSize) {
		err = fmt.Errorf("testData file size too big")
		return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
	}

	open, err = os.Open(testDataPath)
	if err != nil {
		err = fmt.Errorf("open testData file failed")
		return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
	}

	decoder = yaml.NewYAMLOrJSONDecoder(open, maxFileSize)

	return extractContentForJob(decoder, cmDeviceInfos, jobServerInfoMap, expectFaultRanks, jobDevices)
}

func extractContentForJob(decoder *yaml.YAMLOrJSONDecoder, cmDeviceInfos map[string]*constant.DeviceInfo,
	jobServerInfoMap constant.JobServerInfoMap, expectFaultRanks map[string]JobFaultInfo,
	jobDevices map[string]map[string]constant.ServerHccl) (map[string]*constant.DeviceInfo,
	constant.JobServerInfoMap, map[string]JobFaultInfo, error) {
	err := decoder.Decode(&cmDeviceInfos)
	if err != nil {
		err = fmt.Errorf("cmDeviceInfos decode failed")
		return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
	}

	err = decoder.Decode(&jobDevices)
	if err != nil {
		err = fmt.Errorf("jobs decode failed")
		return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
	}
	jobServerInfoMap.InfoMap = jobDevices

	err = decoder.Decode(&expectFaultRanks)
	if err != nil {
		err = fmt.Errorf("expectFaultRanks decode failed")
	}
	return cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err
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
