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

// Package context is a DT collection for funcs in context
package context

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/config"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
)

func TestNewSnpContext(t *testing.T) {
	mockFunc1 := gomonkey.ApplyFunc(config.NewSlowNodeParserConfig, func(configPath string) (*config.SlowNodeParserConfig, error) {
		if configPath == "wrongPath" {
			return nil, errors.New("wrong path")
		}
		return &config.SlowNodeParserConfig{}, nil
	})

	mockFunc2 := gomonkey.ApplyFunc(db.NewSqliteDbCtx, func(dbPath string) *db.SnpDbContext {
		return &db.SnpDbContext{}
	})

	defer mockFunc1.Reset()
	defer mockFunc2.Reset()

	_, err := NewSnpContext("wrongPath")
	assert.Equal(t, "wrong path", err.Error())

	_, err = NewSnpContext("normal")
	assert.Nil(t, err)

}
