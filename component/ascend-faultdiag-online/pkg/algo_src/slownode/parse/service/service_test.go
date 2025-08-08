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

// Package service is a DT collection for func in service
package service

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/business/writecsv"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

var (
	newSnpContextError         = false
	readParallelGroupInfoError = false
	collectGlobalRankError     = false
	writeGlobalRankError       = false
	collectIterateDelayError   = false
	writeIterateDelayError     = false
)

func TestSlowNodeParse(t *testing.T) {

	mockFunctions := append(append(defMockFunc1(), defMockFunc2()...), defMockFunc3()...)

	// get snpContext error
	newSnpContextError = true
	assert.Equal(t, "newSnpContextError", SlowNodeParse("").Error())
	newSnpContextError = false

	// get readParallelGroupInfo error
	readParallelGroupInfoError = true
	assert.Equal(t, "readParallelGroupInfoError", SlowNodeParse("").Error())
	readParallelGroupInfoError = false

	// get globalRank error
	collectGlobalRankError = true
	assert.Equal(t, "collectGlobalRankError", SlowNodeParse("").Error())
	collectGlobalRankError = false

	// write csv file error
	writeGlobalRankError = true
	assert.Equal(t, "writeGlobalRankError", SlowNodeParse("").Error())
	writeGlobalRankError = false

	// collecting iteration delay error
	collectIterateDelayError = true
	assert.Equal(t, "collectIterateDelayError", SlowNodeParse("").Error())
	collectIterateDelayError = false

	// writing iteration delay to a CSV file error
	writeIterateDelayError = true
	assert.Equal(t, "writeIterateDelayError", SlowNodeParse("").Error())
	writeIterateDelayError = false

	// no error
	assert.Nil(t, SlowNodeParse(""))

	// restore the original functions
	for _, function := range mockFunctions {
		function.Reset()
	}
}

func defMockFunc1() []*gomonkey.Patches {
	mockFunc1 := gomonkey.ApplyFunc(context.NewSnpContext, func(string) (*context.SnpRankContext, error) {
		if newSnpContextError {
			return nil, errors.New("newSnpContextError")
		}
		return &context.SnpRankContext{
			ContextData: &contextdata.SnpRankContextData{
				Config: &config.SlowNodeParserConfig{},
			},
		}, nil
	})

	mockFunc2 := gomonkey.ApplyFunc(readParallelGroupInfo, func(string) (map[string]*model.OpGroupInfo, error) {
		if readParallelGroupInfoError {
			return nil, errors.New("readParallelGroupInfoError")
		}
		return make(map[string]*model.OpGroupInfo), nil
	})
	return []*gomonkey.Patches{mockFunc1, mockFunc2}
}

func defMockFunc2() []*gomonkey.Patches {
	mockFunc1 := gomonkey.ApplyFunc(CollectGlobalRank, func(
		*contextdata.SnpRankContextData, map[string]*model.OpGroupInfo, int64) ([]*model.StepGlobalRank, error) {
		if collectGlobalRankError {
			return nil, errors.New("collectGlobalRankError")
		}
		return []*model.StepGlobalRank{{}}, nil
	})

	mockFunc2 := gomonkey.ApplyFunc(writecsv.WriteGlobalRank, func(
		[]*model.StepGlobalRank, string) error {
		if writeGlobalRankError {
			return errors.New("writeGlobalRankError")
		}
		return nil
	})
	return []*gomonkey.Patches{mockFunc1, mockFunc2}
}

func defMockFunc3() []*gomonkey.Patches {
	mockFunc1 := gomonkey.ApplyFunc(CollectIterateDelay, func(
		*db.SnpDbContext) ([]*model.StepIterateDelay, error) {
		if collectIterateDelayError {
			return nil, errors.New("collectIterateDelayError")
		}
		return []*model.StepIterateDelay{}, nil
	})

	mockFunc2 := gomonkey.ApplyFunc(writecsv.WriteIterateDelay, func(
		[]*model.StepIterateDelay, string) error {
		if writeIterateDelayError {
			return errors.New("writeIterateDelayError")
		}
		return nil
	})
	return []*gomonkey.Patches{mockFunc1, mockFunc2}
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
