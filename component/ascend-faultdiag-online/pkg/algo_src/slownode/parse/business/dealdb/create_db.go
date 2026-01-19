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
Package dealdb.
*/
package dealdb

import (
	"fmt"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/db"
)

// CreateDBTable 创建数据库表
func CreateDBTable(ctx *db.SnpDbContext) error {
	if err := CreatCAnnApiTable(ctx); err != nil {
		return err
	}
	if err := CreatCommOpTable(ctx); err != nil {
		return err
	}
	if err := CreatMSTXEventsTable(ctx); err != nil {
		return err
	}
	if err := CreatStepTimeTable(ctx); err != nil {
		return err
	}
	if err := CreatTaskTable(ctx); err != nil {
		return err
	}
	if err := CreatStringIdsTable(ctx); err != nil {
		return err
	}
	return nil
}

// CreatCAnnApiTable 创建表 CANN_API
func CreatCAnnApiTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS CANN_API (
			startNs      INTEGER,
			endNs        INTEGER,
			type         INTEGER,
			globalTid    INTEGER,
			connectionId INTEGER PRIMARY KEY,
			name         INTEGER
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table CANN_API error: %v", err)
	}
	return nil
}

// CreatCommOpTable 创建表 COMMUNICATION_OP
func CreatCommOpTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS COMMUNICATION_OP (
			opName       INTEGER,
			startNs      INTEGER,
			endNs        INTEGER,
			connectionId INTEGER,
			groupName    INTEGER,
			opId         INTEGER PRIMARY KEY,
			relay        INTEGER,
			retry        INTEGER,
			dataType     INTEGER,
			algType      INTEGER,
			count        INTEGER,
			opType       INTEGER
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table COMMUNICATION_OP error: %v", err)
	}
	return nil
}

// CreatMSTXEventsTable 创建表 MSTX_EVENTS
func CreatMSTXEventsTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS MSTX_EVENTS (
			startNs      INTEGER,
			endNs        INTEGER,
			eventType    INTEGER,
			rangeId      INTEGER,
			category     INTEGER,
			message      INTEGER,
			globalTid    INTEGER,
			endGlobalTid INTEGER,
			domainId     INTEGER,
			connectionId INTEGER PRIMARY KEY
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table MSTX_EVENTS error: %v", err)
	}
	return nil
}

// CreatStepTimeTable 创建表 STEP_TIME
func CreatStepTimeTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS STEP_TIME (
			id      INTEGER PRIMARY KEY,
			startNs INTEGER,
			endNs   INTEGER
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table STEP_TIME error: %v", err)
	}
	return nil
}

// CreatTaskTable 创建表 TASK
func CreatTaskTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS TASK (
			startNs      INTEGER,
			endNs        INTEGER,
			deviceId     INTEGER,
			connectionId INTEGER PRIMARY KEY,
			globalTaskId INTEGER,
			globalPid    INTEGER,
			taskType     INTEGER,
			contextId    INTEGER,
			streamId     INTEGER,
			taskId       INTEGER,
			modelId      INTEGER
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table TASK error: %v", err)
	}
	return nil
}

// CreatStringIdsTable 创建表 STRING_IDS
func CreatStringIdsTable(ctx *db.SnpDbContext) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS STRING_IDS (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT
		);
	`
	if err := db.CreateTable(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table STRING_IDS error: %v", err)
	}
	return nil
}
