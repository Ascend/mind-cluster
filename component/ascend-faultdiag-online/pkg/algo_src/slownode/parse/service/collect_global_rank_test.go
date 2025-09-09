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

// Package service is a DT collection for funcs in service
package service

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/dealdb"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

func TestQueryDataLoaderHostDuration(t *testing.T) {

	var queryHostStartEndTimeError = false
	var queryDataLoaderHostError = false

	mockFunc1 := gomonkey.ApplyFunc(dealdb.QueryHostStartEndTime, func(
		*db.SnpDbContext, int64) (*model.StartEndNs, error) {
		if queryHostStartEndTimeError {
			return nil, errors.New("queryHostStartEndTimeError")
		}
		return &model.StartEndNs{StartNs: 111, EndNs: 222}, nil
	})

	mockFunc2 := gomonkey.ApplyFunc(dealdb.QueryDataLoaderHost, func(
		*db.SnpDbContext, int64, int64) (*model.Duration, error) {
		if queryDataLoaderHostError {
			return nil, errors.New("queryDataLoaderHostError")
		}
		return &model.Duration{Dur: 555}, nil
	})

	defer mockFunc1.Reset()
	defer mockFunc2.Reset()

	// queryHostStartEndTimeError
	queryHostStartEndTimeError = true
	dur, err := queryDataLoaderHostDuration(nil, 1)
	assert.Nil(t, dur)
	assert.Equal(t, "queryHostStartEndTimeError", err.Error())
	queryHostStartEndTimeError = false

	// queryDataLoaderHostError
	queryDataLoaderHostError = true
	dur, err = queryDataLoaderHostDuration(nil, 1)
	assert.Nil(t, dur)
	assert.Equal(t, "queryDataLoaderHostError", err.Error())
	queryDataLoaderHostError = false

	// normal
	dur, err = queryDataLoaderHostDuration(nil, 1)
	assert.Equal(t, int64(555), dur.Dur)
	assert.Nil(t, err)
}

func TestCollectGroupRank(t *testing.T) {
	var queryPpDevDurationFailed, queryPpHostDurationFailed, queryDataLoaderHostDurationFailed bool

	patches := mockFunc(&queryPpDevDurationFailed, &queryPpHostDurationFailed, &queryDataLoaderHostDurationFailed)
	defer patches.Reset()

	var queryZpHdDurationFuncWithError QueryZpHdDurationFunc
	var queryZpHdDurationFuncWithSuccess QueryZpHdDurationFunc

	queryZpHdDurationFuncWithError = func(dbCtx *db.SnpDbContext, startNsMin int64,
		startNsMax int64) (*model.HostDeviceDuration, error) {
		return nil, errors.New("queryZpHdDurationFuncWithError")
	}

	queryZpHdDurationFuncWithSuccess = func(dbCtx *db.SnpDbContext, startNsMin int64,
		startNsMax int64) (*model.HostDeviceDuration, error) {
		return &model.HostDeviceDuration{HostDuration: 123, DeviceDuration: 321}, nil
	}

	var startEndNsList = []*model.StepStartEndNs{
		{Id: 0, StartNs: 123, EndNs: 321},
		{Id: 1, StartNs: 123, EndNs: 321},
		{Id: 2, StartNs: 123, EndNs: 321},
		{Id: 3, StartNs: 123, EndNs: 321},
	}

	// queryZpHdDurationFuncWithError
	_, err := collectGroupRank(nil, -1, startEndNsList, queryZpHdDurationFuncWithError)
	assert.Equal(t, "queryZpHdDurationFuncWithError", err.Error())

	// QueryPpDevDuration failed
	queryPpDevDurationFailed = true
	_, err = collectGroupRank(nil, -1, startEndNsList, queryZpHdDurationFuncWithSuccess)
	assert.Equal(t, "queryPpDevDurationFailed", err.Error())
	queryPpDevDurationFailed = false

	// QueryPpHostDuration failed
	queryPpHostDurationFailed = true
	_, err = collectGroupRank(nil, -1, startEndNsList, queryZpHdDurationFuncWithSuccess)
	assert.Equal(t, "queryPpHostDurationFailed", err.Error())
	queryPpHostDurationFailed = false

	// queryDataLoaderHostDuration failed
	queryDataLoaderHostDurationFailed = true
	_, err = collectGroupRank(nil, -1, startEndNsList, queryZpHdDurationFuncWithSuccess)
	assert.Equal(t, "queryDataLoaderHostDurationFailed", err.Error())
	queryDataLoaderHostDurationFailed = false

	// normal
	data, err := collectGroupRank(nil, -1, startEndNsList, queryZpHdDurationFuncWithSuccess)
	assert.Nil(t, err)
	assert.Equal(t, len(startEndNsList), len(data))
}

func mockFunc(
	queryPpDevDurationFailed *bool,
	queryPpHostDurationFailed *bool,
	queryDataLoaderHostDurationFailed *bool,
) *gomonkey.Patches {
	// mock 3 func
	patches := gomonkey.ApplyFunc(dealdb.QueryPpDevDuration, func(
		*db.SnpDbContext, int64, int64, int64) (*model.Duration, error) {
		if *queryPpDevDurationFailed {
			return nil, errors.New("queryPpDevDurationFailed")
		}
		return &model.Duration{Dur: 111}, nil
	})

	patches.ApplyFunc(dealdb.QueryPpHostDuration, func(
		*db.SnpDbContext, int64, int64, int64) (*model.Duration, error) {
		if *queryPpHostDurationFailed {
			return nil, errors.New("queryPpHostDurationFailed")
		}
		return &model.Duration{Dur: 111}, nil
	})

	patches.ApplyFunc(queryDataLoaderHostDuration, func(
		*db.SnpDbContext, int64) (*model.Duration, error) {
		if *queryDataLoaderHostDurationFailed {
			return nil, errors.New("queryDataLoaderHostDurationFailed")
		}
		return &model.Duration{Dur: 111}, nil
	})
	return patches
}

func TestReadParallelGroupInfo(t *testing.T) {
	data := map[string]*model.OpGroupInfo{
		"group_name_136": {
			GroupName:   "pp",
			GroupRank:   0,
			GlobalRanks: []int64{1, 2},
		},
		"group_name_137": {
			GroupName:   "tp",
			GroupRank:   1,
			GlobalRanks: []int64{5, 6, 7},
		},
	}
	// creating a temporary json file
	tmpFile, err := os.CreateTemp("", "parallel_group.json")
	assert.NoError(t, err)

	// encode to JSON and write to a file
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)

	_, err = tmpFile.Write(jsonBytes)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	infoMap, err := readParallelGroupInfo(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, infoMap, data)

}
