/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package sdk provides node parse
package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

var (
	name137 = `{"streamId": "10","count": "4194304","dataType": "bfp16","opName":` +
		` "HcclAllReduce","groupName": "group_name_137"}`
	nameHCCLSend136 = `{"streamId": "10","count": "4194304","dataType": "bfp16","opName":` +
		` "HcclSend","groupName": "group_name_136"}`
	perStartTime = time.Now().UnixNano()
)

const (
	dataRows       = 40
	perStep        = 5
	secondDataRows = 4

	maxAge        = 7
	maxBackups    = 7
	maxLineLength = 512

	testDirPermission = 0755

	sleepWaiteTime    = 2  // 2
	sendStopSleepTime = 45 // 45
	writeSleepTime    = 20 // 20
)

func TestParse(t *testing.T) {
	// 输入文件构造
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	jobDir := filepath.Join(currentDir, "job_dir")
	rankDir := filepath.Join(jobDir, "0")
	err = os.MkdirAll(rankDir, testDirPermission)
	assert.NoError(t, err)
	// init hwlog
	err = initLog(jobDir)
	assert.NoError(t, err)
	// test_database.db
	_, err = getFilePath("database.db")
	assert.NoError(t, err)
	// 并行域数据文件：parallel_group.json
	tpOrNot := "tp" // 修改此值，测试有无tp的情况
	parallelGroupPath := createParallelGroupFile(t, tpOrNot)

	cg := config.DataParseModel{
		FilePath:     currentDir,
		JobName:      "test_parse",
		JobId:        "job_dir",
		Traffic:      int64(2),
		RankIds:      []string{"0"},
		JobStartTime: time.Now().Unix(),
	}

	go StartParse(cg)
	time.Sleep(sleepWaiteTime * time.Second)

	go ReloadParse(cg)

	// profiling数据文件：1747645582
	go syncWriteProfile(t)

	sendStopInfo(cg)

	checkFileExist([]string{parallelGroupPath, filepath.Join(jobDir, "test.log")}, t)
	err = os.RemoveAll(jobDir)
	assert.NoError(t, err)
}

/*
start -> stop -> start(等待上一个stop执行完后再开启任务) -> stop
[SLOWNODE PARSE]Succeeded in initializing parse job: job_dir
[SLOWNODE PARSE]Job job_dir is in the stopping state and waiting to be stopped
[SLOWNODE PARSE]Succeeded in stopping parsed
[SLOWNODE PARSE]Succeeded in stopping the executing job, start to execute a new task
[SLOWNODE PARSE]Succeeded in initializing parse job: job_dir
[SLOWNODE PARSE]Succeeded in stopping parsed
*/
func TestStartStopStartStopParse(t *testing.T) {
	// 输入文件构造
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	jobDir := filepath.Join(currentDir, "job_dir")
	rankDir := filepath.Join(jobDir, "0")
	err = os.MkdirAll(rankDir, testDirPermission)
	assert.NoError(t, err)
	rankDir1 := filepath.Join(jobDir, "1")
	err = os.MkdirAll(rankDir1, testDirPermission)
	assert.NoError(t, err)
	// init hwlog
	err = initLog(jobDir)
	assert.NoError(t, err)
	// test_database.db
	_, err = getFilePath("database.db")
	assert.NoError(t, err)
	// 并行域数据文件：parallel_group.json
	tpOrNot := "tp" // 修改此值，测试有无tp的情况
	parallelGroupPath := createParallelGroupFile(t, tpOrNot)

	cg := config.DataParseModel{
		FilePath:     currentDir,
		JobName:      "test_parse",
		JobId:        "job_dir",
		Traffic:      int64(2),
		RankIds:      []string{"0", "1", "99"},
		JobStartTime: time.Now().Unix(),
	}
	// 先start后stop
	go StartParse(cg)
	time.Sleep(sleepWaiteTime * time.Second)
	go StopParse(cg)

	// 等待上一个stop执行完后再开启任务
	go StartParse(cg)
	go syncWriteProfile(t)

	sendStopInfo(cg)

	checkFileExist([]string{parallelGroupPath, filepath.Join(jobDir, "test.log")}, t)
	err = os.RemoveAll(jobDir)
	assert.NoError(t, err)
}

func sendStopInfo(cg config.DataParseModel) {
	time.Sleep(sendStopSleepTime * time.Second)
	StopParse(cg)
}

func checkFileExist(filePaths []string, t *testing.T) {
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			hwlog.RunLog.Errorf("[TEST]Assert: file %s not exist", filePath)
			assert.NoError(t, err)
		}
	}
}

func writeProfile(filePath string, t *testing.T) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.DefaultFilePermission)
	assert.NoError(t, err)
	defer file.Close()
	const (
		randAdd = 800000
		randEnd = 100000
	)
	now := time.Now().UnixNano()
	rand.Seed(now)
	var jsonData []model.JsonData
	for i := perStep * dataRows; i < perStep*(dataRows+secondDataRows); i += perStep {
		var start int64
		if i == perStep*dataRows {
			start = perStartTime
		} else {
			start = now + int64(i)
		}
		randomGap := int64(rand.Intn(randEnd+1) + randAdd) // [800000, 900000]
		end := start + randomGap
		jsonData = append(jsonData, createOneRow(int64(i), start, end)...)
		perStartTime = end
	}

	encoder := json.NewEncoder(file)
	for _, item := range jsonData {
		err := encoder.Encode(item)
		assert.NoError(t, err)
	}
	assert.NoError(t, err)
}

func syncWriteProfile(t *testing.T) {
	profiles := createProfiles(t)
	time.Sleep(writeSleepTime * time.Second)
	go writeProfile(profiles[len(profiles)-1], t)
}

func createProfiles(t *testing.T) []string {
	const (
		randAdd = 100000
		randEnd = 200000
	)
	// 模拟profiling数据文件
	now := time.Now().UnixNano()
	rand.Seed(now)
	var jsonData []model.JsonData
	for i := 0; i < perStep*dataRows; i += perStep {
		var start int64
		if i == 0 {
			start = perStartTime
		} else {
			start = now + int64(i)
		}
		randomGap := int64(rand.Intn(randEnd+1) + randAdd) // [100000, 300000]
		end := start + randomGap

		jsonData = append(jsonData, createOneRow(int64(i), start, end)...)
		perStartTime = end
	}
	filePath, err := getFilePath(fmt.Sprintf("%d", time.Now().Unix()))
	assert.NoError(t, err)
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	defer file.Close()
	err = file.Chmod(constants.DefaultFilePermission)
	assert.NoError(t, err)

	encoder := json.NewEncoder(file)
	for _, item := range jsonData {
		err := encoder.Encode(item)
		assert.NoError(t, err)
	}
	return []string{filePath}
}

func createOneRow(index int64, startTime int64, endTime int64) []model.JsonData {
	const (
		idOne   = 1
		idTwo   = 2
		idThree = 3
		idFour  = 4
		idFive  = 5
	)
	jsonData := append(
		createData("default", index+idOne, fmt.Sprintf("step %d", (index+idOne)/perStep), startTime, endTime),
		createData("communication", index+idTwo, name137, startTime, endTime)...)
	jsonData = append(jsonData, createData("default", index+idThree, "dataloader", startTime, endTime)...)
	jsonData = append(jsonData, createData("communication", index+idFour, name137, startTime, endTime)...)
	// 构造pp算子数据
	jsonData = append(jsonData, createData("communication", index+idFive, nameHCCLSend136, startTime, endTime)...)
	return jsonData
}

func createData(domain string, id int64, name string, startTime int64, endTime int64) []model.JsonData {
	redundantTime := int64(rand.Intn(1) + 100000) // 增加一个冗余，区分device和host
	startTimeFlag := 16
	endTimeFlag := 32
	if name == "dataloader" || name == "save_checkpoint" {
		startTimeFlag = 2
		endTimeFlag = 4
	}
	data := []model.JsonData{
		// device
		{Kind: 1, Flag: startTimeFlag, SourceKind: 1, Timestamp: startTime, Id: id,
			MSPTIObjectId: model.ObjectId{
				Pt: model.Pt{ProcessId: 0, ThreadId: 6}, Ds: model.Ds{DeviceId: 0, StreamId: 6},
			},
			Name:   name,
			Domain: domain,
		},
		{Kind: 1, Flag: endTimeFlag, SourceKind: 1, Timestamp: endTime, Id: id,
			MSPTIObjectId: model.ObjectId{
				Pt: model.Pt{ProcessId: 0, ThreadId: 6}, Ds: model.Ds{DeviceId: 0, StreamId: 6},
			},
			Name: "", Domain: "",
		},
		// host
		{Kind: 1, Flag: startTimeFlag, SourceKind: 0, Timestamp: startTime, Id: id,
			MSPTIObjectId: model.ObjectId{
				Pt: model.Pt{ProcessId: 58463, ThreadId: 61196}, Ds: model.Ds{DeviceId: 58463, StreamId: 61196},
			},
			Name: "", Domain: "",
		},
		{Kind: 1, Flag: endTimeFlag, SourceKind: 0, Timestamp: endTime + redundantTime, Id: id,
			MSPTIObjectId: model.ObjectId{
				Pt: model.Pt{ProcessId: 58463, ThreadId: 61196}, Ds: model.Ds{DeviceId: 58463, StreamId: 61196},
			},
			Name: "", Domain: "",
		},
	}
	return data
}

func createParallelGroupFile(t *testing.T, tpOrNot string) string {
	info := map[string]*model.OpGroupInfo{
		"group_name_136": {
			GroupName:   "pp",
			GroupRank:   0,
			GlobalRanks: []int64{1, 2},
		},
		"group_name_137": {
			GroupName:   tpOrNot,
			GroupRank:   1,
			GlobalRanks: []int64{5, 6, 7},
		},
		"group_name_1": {
			GroupName:   "cp",
			GroupRank:   2,
			GlobalRanks: []int64{1, 3, 9},
		},
	}

	data, err := json.MarshalIndent(info, "", "  ")
	assert.NoError(t, err)

	filePath, err := getFilePath("parallel_group.json")
	assert.NoError(t, err)
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	defer file.Close()
	err = file.Chmod(constants.DefaultFilePermission)
	assert.NoError(t, err)

	_, err = file.Write(data)
	assert.NoError(t, err)

	return filePath
}

func getFilePath(fileName string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dbFilePath := filepath.Join(currentDir, "job_dir", "0", fileName)
	return dbFilePath, nil
}

func initLog(jobDir string) error {
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   filepath.Join(jobDir, "test.log"),
		LogLevel:      0, // "Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)"
		MaxAge:        maxAge,
		MaxBackups:    maxBackups,
		MaxLineLength: maxLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.TODO()); err != nil {
		return fmt.Errorf("hwlog init failed, error is %v", err)
	}
	return nil
}
