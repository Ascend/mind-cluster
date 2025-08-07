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
Package db.
*/
package db

import (
	"database/sql"
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

// SnpDbContext 数据库上下文
type SnpDbContext struct {
	DbType string
	DbPath string
	dbConn *sql.DB
	// 在 WAL 模式下，可以多个读协程一起读取，只需要给写入加锁
	DbWriteLock sync.Mutex
}

// Conn 连接
func (dbCtx *SnpDbContext) Conn() error {
	if err := utils.CheckDBFilePerm(dbCtx.DbPath); err != nil {
		return err
	}
	// 启用 WAL 模式，提高读写并发性
	params := "_journal_mode=WAL&_foreign_keys=on"
	dsn := fmt.Sprintf("file:%s?%s", dbCtx.DbPath, params)
	dbConn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %s", err)
	}
	rawDb, err := dbConn.DB()
	if err != nil {
		return fmt.Errorf("failed to transform database: %s", err)
	}
	// 测试数据库连接
	err = rawDb.Ping()
	if err != nil {
		if err := dbCtx.Close(); err != nil {
			return err
		}
		return fmt.Errorf("failed to ping database: %s", err)
	}
	// 限制最大连接数为 1，避免 SQLite 多连接冲突
	rawDb.SetMaxOpenConns(1)
	dbCtx.dbConn = rawDb
	return nil
}

// Close 关闭连接
func (dbCtx *SnpDbContext) Close() error {
	return dbCtx.dbConn.Close()
}

// Query 查询: query all records, got a array of records, no records queried, return a empty array.
func Query[T any](dbCtx *SnpDbContext, sqlStr string, params []any,
	ptrListFunc func(t *T) []any) ([]*T, error) {
	rows, err := dbCtx.dbConn.Query(sqlStr, params...)
	if err != nil {
		return nil, err
	}
	var rowsErr error
	var results []*T
	for rows.Next() {
		var t T
		rowsErr = rows.Scan(ptrListFunc(&t)...)
		if rowsErr != nil {
			break
		}
		results = append(results, &t)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if rowsErr != nil {
		return nil, rowsErr
	}
	return results, nil
}

// QuerySingleLine query the first record, is no record queried, return error or nil
func QuerySingleLine[T any](dbCtx *SnpDbContext, sqlStr string, params []any,
	ptrListFunc func(t *T) []any) (*T, error) {
	query, err := Query(dbCtx, sqlStr, params, ptrListFunc)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		return query[0], nil
	}
	return nil, nil
}

// Insert 插入
func Insert(ctx *SnpDbContext, sqlStr string, params []any) (int64, error) {
	result, err := ctx.dbConn.Exec(sqlStr, params...)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Delete 删除
func Delete(ctx *SnpDbContext, sqlStr string, params []any) error {
	result, err := ctx.dbConn.Exec(sqlStr, params...)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

// CreateTable 创建表
func CreateTable(dbCtx *SnpDbContext, sqlStr string) error {
	if _, err := dbCtx.dbConn.Exec(sqlStr); err != nil {
		return err
	}
	return nil
}
