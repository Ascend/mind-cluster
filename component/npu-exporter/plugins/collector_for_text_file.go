/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package plugins for custom metrics
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
	"huawei.com/npu-exporter/v6/utils/logger"
)

var (
	filePath             = ""
	validPaths           = make([]string, 0) // list of file paths
	existMetrics         = make(map[string]string)
	metricStructInfosMap = make(map[string]metricStructInfo)
	baseCacheKey         = ""
)

const (
	size100k                 = 100 * 1024
	maxLabelSize             = 10
	maxFileNumber            = 10
	num1000                  = 1000
	tickerDuration           = 100
	maxDataListSize          = 128
	maxMetricNameSize        = 128
	maxDescSize              = 1024
	fileDisabledMsg          = "will not collect this file any more"
	fileMetricsDisabledMsg   = "file metrics collection will be disabled"
	skipCurrentCollectionMsg = "will skip current collection and report cached metrics"
	excludedPermission       = 0111 // file should not have any execute permission
	logDomain                = "textMetrics"
)

func init() {
	baseCacheKey = common.GetCacheKey(&TextMetricsInfoCollector{})
}

type metricStructInfo struct {
	name       string
	metricDesc *prometheus.Desc
	labels     []string
	desc       string
	version    string
}

// TextMetricData represents the JSON structure
type TextMetricData struct {
	Version   string     `json:"version"`
	Desc      string     `json:"desc"`
	Name      string     `json:"name"`
	Timestamp int64      `json:"timestamp"`
	DataList  []DataItem `json:"data_list"`
}

// DataItem represents each item in data_list
type DataItem struct {
	Label map[string]string `json:"label"`
	Value float64           `json:"value"`
}

// SetTextMetricsFilePath init text metric
func SetTextMetricsFilePath(metricsFilePath string) {
	filePath = metricsFilePath
}

func isDataOk(metricsData *TextMetricData, filePath string) error {
	if len(metricsData.DataList) == 0 {
		return fmt.Errorf("dataList is empty in json file %s", filePath)
	}
	if len(metricsData.DataList) > maxDataListSize {
		return fmt.Errorf("size of dataList(%d) is more than max allowed dataList size(%d) in json file %s",
			len(metricsData.DataList), maxDataListSize, filePath)
	}
	if len(metricsData.DataList[0].Label) > maxLabelSize {
		return fmt.Errorf("size of first item's Label(%d) is more than max allowed label size(%d) in json file %s",
			len(metricsData.DataList[0].Label), maxLabelSize, filePath)
	}
	if metricsData.Name == "" {
		return fmt.Errorf("name field is empty in json file %s", filePath)
	}
	if len(metricsData.Name) > maxMetricNameSize {
		return fmt.Errorf("length of metric name should not larger than %d, but current is %d, file: %s",
			maxMetricNameSize, len(metricsData.Name), filePath)
	}
	if metricsData.Desc == "" {
		return fmt.Errorf("desc field is empty in json file %s", filePath)
	}
	if len(metricsData.Desc) > maxDescSize {
		return fmt.Errorf("length of metric desc should not larger than %d, but current is %d, file: %s",
			maxDescSize, len(metricsData.Desc), filePath)
	}
	// only support 1.0 version currently
	if metricsData.Version != "1.0" {
		return fmt.Errorf("version should be 1.0, but current is %s, file: %s", metricsData.Version, filePath)
	}
	if metricsData.Timestamp <= 0 {
		return fmt.Errorf("timestamp field is empty or not correct in json file %s", filePath)
	}
	return nil
}

func isNotExistOrEmpty(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no such file or directory") ||
		strings.Contains(err.Error(), "EOF")
}

func checkFilePermission(filePath string) error {
	absFilePath, err := utils.CheckPath(filePath)
	if err != nil {
		return err
	}
	currentUID := uint32(os.Getuid())
	if err = utils.DoCheckOwnerAndPermission(absFilePath, excludedPermission, currentUID); err != nil {
		return fmt.Errorf("file permission check failed: file should be owned by current process (uid: %d) "+
			"and should not have execute permission (%04o) , error: %v", currentUID, excludedPermission, err)
	}
	return nil
}

// TextMetricsInfoCollector collect custom plugin info
type TextMetricsInfoCollector struct {
	common.MetricsCollectorAdapter
	Cache sync.Map
}

// Describe description of the metric
func (c *TextMetricsInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricStructInfosMap {
		if metric.metricDesc != nil {
			ch <- metric.metricDesc
		}
	}
}

// CollectToCache collect the metric to cache
func (c *TextMetricsInfoCollector) CollectToCache(n *common.NpuCollector, chipList []common.HuaWeiAIChip) {
	logger.Debugf("TextMetricsInfoCollector CollectToCache")

	for _, jsonFilePath := range validPaths {
		fileData, err := utils.ReadLimitBytes(jsonFilePath, size100k)
		logId := jsonFilePath + "readFileErr"
		if err != nil {
			logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: logDomain, ID: logId},
				"read json file %s failed: %v", jsonFilePath, err)
			continue
		}
		hwlog.ResetErrCnt(logDomain, logId)

		var metricsData TextMetricData
		logId = jsonFilePath + "unmarshalFileErr"
		if err := json.Unmarshal(fileData, &metricsData); err != nil {
			logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: logDomain, ID: logId},
				"unmarshal json file %s failed: %v, "+
					"Possible causes:\n1. The file is not in JSON format\n2. File size is more than 100KB ", jsonFilePath, err)
			continue
		}
		hwlog.ResetErrCnt(logDomain, logId)

		if isStructInfoChangedForFile(jsonFilePath, metricsData) {
			continue
		}

		logId = jsonFilePath + "dataNotOk"
		if err := isDataOk(&metricsData, jsonFilePath); err != nil {
			logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: logDomain, ID: logId},
				"%v, %s", err, skipCurrentCollectionMsg)
			continue
		}
		hwlog.ResetErrCnt(logDomain, logId)

		c.Cache.Store(fmt.Sprintf("%s-%s", baseCacheKey, jsonFilePath), metricsData)
	}
}

func isStructInfoChangedForFile(jsonFilePath string, data TextMetricData) bool {
	structInfo, exists := metricStructInfosMap[jsonFilePath]
	if !exists {
		return false
	}
	return checker(jsonFilePath, structInfo.name, data.Name, "name") ||
		checker(jsonFilePath, structInfo.version, data.Version, "version") ||
		checker(jsonFilePath, structInfo.desc, data.Desc, "desc")
}

func checker(jsonFilePath string, structInfo string, newData string, logFlag string) bool {
	if structInfo != newData {
		logger.LogfWithOptions(logger.WarnLevel, logger.LogOptions{Domain: logDomain, ID: jsonFilePath + logFlag},
			"[%v] of json file [%v] changed, %s, old: %v, new: %v",
			logFlag, jsonFilePath, skipCurrentCollectionMsg, structInfo, newData)
		return true
	}
	hwlog.ResetErrCnt(logDomain, jsonFilePath+logFlag)
	return false
}

// UpdatePrometheus update prometheus metric
func (c *TextMetricsInfoCollector) UpdatePrometheus(ch chan<- prometheus.Metric, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) {
	logger.Debug("TextMetricsInfoCollector UpdatePrometheus")
	c.update(func(jsonFilePath string, structInfo metricStructInfo, timestamp time.Time, item DataItem, index int) {
		labelValues := make([]string, len(structInfo.labels))
		for i, key := range structInfo.labels {
			if value, ok := item.Label[key]; ok {
				labelValues[i] = value
			} else {
				labelValues[i] = ""
			}
		}

		ch <- prometheus.NewMetricWithTimestamp(timestamp,
			prometheus.MustNewConstMetric(structInfo.metricDesc, prometheus.GaugeValue, item.Value, labelValues...))
	})
}

// UpdateTelegraf update telegraf metric
func (c *TextMetricsInfoCollector) UpdateTelegraf(fieldsMap map[string]map[string]interface{}, n *common.NpuCollector,
	containerMap map[int32]container.DevicesInfo, chips []common.HuaWeiAIChip) map[string]map[string]interface{} {
	logger.Debug("TextMetricsInfoCollector UpdateTelegraf")
	if fieldsMap[common.KeyForTextMetrics] == nil {
		fieldsMap[common.KeyForTextMetrics] = make(map[string]interface{})
	}
	c.update(func(jsonFilePath string, structInfo metricStructInfo, timestamp time.Time, item DataItem, index int) {
		labelsMap := make(map[string]string)
		for _, key := range structInfo.labels {
			if value, ok := item.Label[key]; ok {
				labelsMap[key] = value
			} else {
				labelsMap[key] = ""
			}
		}
		tetegrafData := common.TelegrafData{
			Labels:    labelsMap,
			Metrics:   map[string]interface{}{structInfo.name: item.Value},
			Timestamp: timestamp,
		}
		fieldsMap[common.KeyForTextMetrics][jsonFilePath+"-"+strconv.Itoa(index)] = tetegrafData
	})

	return fieldsMap
}

// processFileData processes file data and initializes its metrics
func processFileData(path string, fileData []byte) error {
	var metricsData TextMetricData
	if err := json.Unmarshal(fileData, &metricsData); err != nil {
		return fmt.Errorf("unmarshal json file %s failed, %s: %v, "+
			"Possible causes:\t1. The file is not in correct format;\t2. File size is more than 100KB;",
			path, fileDisabledMsg, err)
	}
	if mPath, ok := existMetrics[metricsData.Name]; ok {
		return fmt.Errorf("metric [%s] already described in file [%s], ignore file [%s]", metricsData.Name, mPath, path)
	}
	if err := isDataOk(&metricsData, path); err != nil {
		return fmt.Errorf("%v, %s", err, fileDisabledMsg)
	}
	desc := metricsData.Desc
	labelKeys := make([]string, 0, len(metricsData.DataList[0].Label))
	for key := range metricsData.DataList[0].Label {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys)
	logger.Infof("init text metric succeeded for file %s, metricName: %v, version: %v, desc: %v, labels: %v",
		path, metricsData.Name, metricsData.Version, desc, labelKeys)

	metricStructInfosMap[path] = metricStructInfo{
		metricDesc: prometheus.NewDesc(metricsData.Name, desc, labelKeys, nil),
		labels:     labelKeys,
		name:       metricsData.Name,
		desc:       desc,
		version:    metricsData.Version,
	}
	existMetrics[metricsData.Name] = path
	return nil
}

// checkAndProcessFile checks a single file and processes it if exists
func checkAndProcessFile(path string) bool {
	if path == "" {
		return false
	}
	if utils.IsDir(path) {
		logger.Errorf("file path %s is a directory, only support specify file path", path)
		return false
	}

	err := checkFilePermission(path)
	if isNotExistOrEmpty(err) {
		return true
	}
	if err != nil {
		logger.Warnf("check file %s failed: %v", path, err)
		return false
	}

	fileData, err := utils.ReadLimitBytes(path, size100k)
	if isNotExistOrEmpty(err) {
		return true
	}
	if err != nil {
		logger.Warnf("read file %s failed: %v", path, err)
		return false
	}
	if len(fileData) == 0 {
		return true
	}
	err = processFileData(path, fileData)
	if err == nil {
		validPaths = append(validPaths, path)
	} else {
		logger.Error(err)
	}
	return false
}

// preCheckPaths waits for missing files to appear
func preCheckPaths(paths []string) {
	var deadline time.Time
	var once = sync.Once{}
	ticker := time.NewTicker(tickerDuration * time.Millisecond)
	defer ticker.Stop()

	for {
		remainingMissing := make([]string, 0)
		for _, path := range paths {
			trimmedPath := strings.TrimSpace(path)
			isNeedWait := checkAndProcessFile(trimmedPath)
			if isNeedWait {
				remainingMissing = append(remainingMissing, trimmedPath)
			}
		}
		if len(remainingMissing) == 0 {
			break
		}
		paths = remainingMissing
		once.Do(func() {
			logger.Warnf("found %d file(s) that don't exist yet, will wait 1 minute for them: %v",
				len(paths), paths)
			deadline = time.Now().Add(time.Minute)
		})
		if time.Now().After(deadline) {
			if len(paths) > 0 {
				logger.Warnf("timeout (%v) exceeded, %d file(s) still not found, will ignore them: %v",
					time.Minute, len(paths), paths)
			}
			break
		}

		<-ticker.C
		continue
	}
}

// IsSupported Check whether the current hardware supports this metric
func (c *TextMetricsInfoCollector) IsSupported(n *common.NpuCollector) bool {
	if filePath == "" {
		return false
	}
	paths := strings.Split(filePath, ",")
	if len(paths) > maxFileNumber {
		logger.Warnf("the number of files is more than max allowed number(%d), only the first %d files will be collected",
			maxFileNumber, maxFileNumber)
		paths = paths[0:maxFileNumber]
	}

	preCheckPaths(paths)

	if len(validPaths) == 0 {
		logger.Warnf("no valid file paths found in filePath: %s, %s", filePath, fileMetricsDisabledMsg)
		return false
	}
	logger.Infof("successfully initialized %d text metric file(s)", len(validPaths))
	return true
}

func (c *TextMetricsInfoCollector) update(doUpdate func(string, metricStructInfo, time.Time, DataItem, int)) {
	for jsonFilePath, structInfo := range metricStructInfosMap {
		cacheKey := fmt.Sprintf("%s-%s", baseCacheKey, jsonFilePath)
		data, ok := c.Cache.Load(cacheKey)
		if !ok {
			logger.Debugf("cache key %s not found", cacheKey)
			continue
		}

		textMetricsData, ok := data.(TextMetricData)
		if !ok {
			logger.Warnf("cache data type mismatch for key %s", cacheKey)
			continue
		}

		timestamp := time.Unix(0, textMetricsData.Timestamp*num1000)

		for index, item := range textMetricsData.DataList {
			doUpdate(jsonFilePath, structInfo, timestamp, item, index)
		}
	}
}
