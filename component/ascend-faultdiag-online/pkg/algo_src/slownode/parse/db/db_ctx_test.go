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

// Package db is a test collection for func in package db
package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSqliteDbCtx(t *testing.T) {
	dbName := "test_db.db"
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	dbPath := filepath.Join(currentDir, dbName)
	ctx := NewSqliteDbCtx(dbPath)
	err = ctx.Conn()
	if err != nil {
		assert.Error(t, err)
	}
	assert.NotNil(t, ctx)

	err = ctx.Close()
	assert.NoError(t, err)
	// remove db
	err = os.Remove(dbName)
	assert.NoError(t, err)
}
