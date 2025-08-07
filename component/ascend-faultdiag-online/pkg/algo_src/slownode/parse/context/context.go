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

/*
Package context.
*/
package context

import (
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/context/contextdata"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

const dataQueMaxSize = 2000
const sqlQueMaxSize = 1000

// SnpRankContext 慢节点单个Rank清洗上下文
type SnpRankContext struct {
	ContextData  *contextdata.SnpRankContextData
	JsonDataQue  chan *model.JsonData
	InsertSqlQue chan string
	JobId        string
	RankId       string
}

// NewSnpContext 上下文构造函数
func NewSnpContext(configPath string) (*SnpRankContext, error) {
	parserConfig, err := config.NewSlowNodeParserConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &SnpRankContext{ContextData: &contextdata.SnpRankContextData{
		DbCtx:     db.NewSqliteDbCtx(parserConfig.DbFilePath),
		Config:    parserConfig,
		StepCount: 0,
	},
		JsonDataQue:  make(chan *model.JsonData, dataQueMaxSize),
		InsertSqlQue: make(chan string, sqlQueMaxSize),
	}, nil
}
