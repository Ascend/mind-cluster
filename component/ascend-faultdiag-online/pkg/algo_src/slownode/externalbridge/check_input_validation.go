/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Package externalbridge for node and cluster level detection interact interface
package externalbridge

import (
	"encoding/json"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

func checkConfigValid(detectionConfig config.AlgoInputConfig) bool {
	if detectionConfig.DetectionLevel == "" ||
		(detectionConfig.DetectionLevel != "node" &&
			detectionConfig.DetectionLevel != "cluster") ||
		detectionConfig.FilePath == "" ||
		detectionConfig.JobName == "" ||
		detectionConfig.Nsigma < 0 ||
		detectionConfig.NormalNumber <= 0 ||
		detectionConfig.NconsecAnomaliesSignifySlow < 1 ||
		detectionConfig.NsecondsOneDetection < 0 ||
		detectionConfig.DegradationPercentage <= 0 ||
		detectionConfig.ClusterMeanDistance <= 1 ||
		detectionConfig.CardsOneNode <= 0 {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error input parameters")
		return false
	}
	if detectionConfig.NormalNumber > config.NormalNumberUpper {
		detectionConfig.NormalNumber = config.DefaultNormalNumber
	}
	return true
}

func checkConfigDigit(config map[string]any) bool {
	if _, exist := config["normalNumber"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without normalNumber!")
		return false
	}
	if _, exist := config["nSigma"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without nSigma!")
		return false
	}
	if _, exist := config["cardOneNode"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without cardOneNode!")
		return false
	}
	if _, exist := config["nSecondsDoOneDetection"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without nSecondsDoOneDetection!")
		return false
	}
	if _, exist := config["nConsecAnomaliesSignifySlow"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without nConsecAnomaliesSignifySlow!")
		return false
	}
	if _, exist := config["degradationPercentage"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without degradationPercentage!")
		return false
	}
	if _, exist := config["clusterMeanDistance"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without clusterMeanDistance!")
		return false
	}
	return true
}

func checkConfigExist(conf any, cmdStr enum.Command) bool {
	if conf == nil {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Invalid config!")
		return false
	}

	// convert to json bytes
	jsonBytes, err := json.Marshal(conf)
	if err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid config(%v), marshal failed: %v", conf, err)
		return false
	}
	// convert bytes to map[string]any
	var cg map[string]any
	if err = json.Unmarshal(jsonBytes, &cg); err != nil {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]Invalid config(%s), json unmarshal failed: %v", string(jsonBytes), err)
		return false
	}

	if _, exist := cg["jobId"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without jobId!")
		return false
	}
	/* 如果是stop command检查到此处即可 */
	if cmdStr == "stop" {
		return true
	}
	if _, exist := cg["filePath"]; !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without filePath!")
		return false
	}
	return checkConfigDigit(cg)

}

func checkMainKey(input map[string]any) string {
	value, exist := input["command"]
	if !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without command!")
		return ""
	}
	valueStr, ok := value.(string)
	if !ok {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error command!")
		return ""
	}
	targetStr, exist := input["target"]
	if !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without target!")
		return ""
	}
	str, ok := targetStr.(string)
	if !ok {
		hwlog.RunLog.Error("[SLOWNODE ALGO]target is not a string")
		return ""
	}
	if str != "node" && str != "cluster" {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error target!")
		return ""
	}
	eventType, exist := input["eventType"]
	if !exist {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Input without eventType!")
		return ""
	}
	eventTypeStr, ok := eventType.(string)
	if !ok {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error eventType", eventType)
		return ""
	}
	if eventTypeStr != "slownode" {
		hwlog.RunLog.Warn("[SLOWNODE ALGO]eventType:", eventTypeStr)
		return ""
	}
	return valueStr
}

func sliceContains[T comparable](array []T, value T) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

/* check input is invalid */
func checkInvalidInput(inputData *model.Input) bool {
	if !sliceContains([]enum.DeployMode{enum.Node, enum.Cluster}, inputData.Target) {
		hwlog.RunLog.Error("[SLOWNODE ALGO]target is not in input or not a string or invalid.")
		return false
	}

	if !sliceContains([]string{enum.SlowNodeAlgo, enum.DataParse}, inputData.EventType) {
		hwlog.RunLog.Errorf("[SLOWNODE ALGO]invalid eventType: %s", inputData.EventType)
		return false
	}
	if inputData.Command == enum.Register {
		if inputData.Func == nil {
			hwlog.RunLog.Error("[SLOWNODE ALGO]invalid func: should not be nil")
			return false
		}
		return true
	}
	if inputData.Command != enum.Start && inputData.Command != enum.Reload {
		return true
	}
	// check config if slownode algo
	if inputData.EventType == enum.SlowNodeAlgo {
		return checkConfigExist(inputData.Model, inputData.Command)
	}
	return true
}

/* transform model field to struct(when command is start or reload) */
func transformJsonToConfigStruct(paramIn map[string]any) (config.AlgoInputConfig, bool) {
	var configParam config.AlgoInputConfig
	/* model is map[string]any */
	model, ok := paramIn["model"].(map[string]any)
	if !ok {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error model!")
		return configParam, false
	}

	marshal, err := json.Marshal(model)
	if err != nil {
		hwlog.RunLog.Error("[SLOWNODE ALGO]", err)
		return configParam, false
	}

	err = json.Unmarshal(marshal, &configParam)
	if err != nil {
		hwlog.RunLog.Error("[SLOWNODE ALGO]", err)
		return configParam, false
	}

	configParam.DetectionLevel, ok = paramIn["target"].(string)
	if !ok {
		hwlog.RunLog.Error("[SLOWNODE ALGO]Error target")
		return configParam, false
	}

	return configParam, true
}

func transformJsonToStruct[T config.AlgoInputConfig | config.DataParseModel](input *model.Input, cg *T) error {
	dataBytes, err := json.Marshal(input.Model)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(dataBytes, &cg); err != nil {
		return err
	}
	return nil
}
