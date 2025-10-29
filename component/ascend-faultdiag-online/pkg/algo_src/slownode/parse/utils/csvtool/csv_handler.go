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

// Package csvtool Package utils provides some common utils
package csvtool

import (
	"encoding/csv"
	"os"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/enum"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

// CSVHandler 封装 CSV 文件操作
type CSVHandler struct {
	openFile    *OpenFile
	csvFilePath string
	mode        enum.FileMode
}

// OpenFile 保存文件打开内容
type OpenFile struct {
	file   *os.File
	writer *csv.Writer
	mu     sync.Mutex
}

// NewCSVHandler 初始化 CSV 文件处理器
func NewCSVHandler(filePath string, mode enum.FileMode, perm os.FileMode) (*CSVHandler, error) {
	var (
		file *os.File
		err  error
	)

	absFilePath, err := fileutils.CheckPath(filePath)
	if err != nil {
		return nil, err
	}

	switch mode {
	case enum.ReadMode: // mode: "r"=只读
		file, err = os.Open(absFilePath)
	case enum.WriteMode: // mode: "w"=只写(覆盖)
		file, err = os.OpenFile(absFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	case enum.AppendMode: // mode: "a"=追加写
		file, err = os.OpenFile(absFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm)
	default:
		return nil, os.ErrInvalid
	}

	if err != nil {
		return nil, err
	}

	return &CSVHandler{
		openFile: &OpenFile{
			file:   file,
			writer: csv.NewWriter(file),
		},
		csvFilePath: absFilePath,
		mode:        mode,
	}, nil
}

// WriteRow 写入单行 CSV 数据（自动加锁）
func (ch *CSVHandler) WriteRow(record []string) error {
	var err error = nil
	ch.openFile.mu.Lock()
	defer func(err error) {
		if err != nil {
			if closeErro := ch.Close(); closeErro != nil {
				hwlog.RunLog.Errorf("failed to close csv file: %s, error: %v", ch.csvFilePath, closeErro)
			}
		}
		ch.openFile.mu.Unlock()
	}(err)

	if ch.mode == enum.WriteMode {
		// 移动到文件开头并清空内容
		if _, err = ch.openFile.file.Seek(0, 0); err != nil {
			return err
		}
		if err = ch.openFile.file.Truncate(0); err != nil {
			return err
		}
	}
	err = ch.openFile.writer.Write(record)
	return err
}

// WriteAll 写入多行 CSV 数据（原子操作）
func (ch *CSVHandler) WriteAll(records [][]string) error {
	var err error = nil
	ch.openFile.mu.Lock()
	defer func(err error) {
		if err != nil {
			if closeErro := ch.Close(); closeErro != nil {
				hwlog.RunLog.Errorf("failed to close csv file: %s, error: %v", ch.csvFilePath, closeErro)
			}
		}
		ch.openFile.mu.Unlock()
	}(err)

	if ch.mode == enum.WriteMode {
		// 移动到文件开头并清空内容
		if _, err = ch.openFile.file.Seek(0, 0); err != nil {
			return err
		}
		if err = ch.openFile.file.Truncate(0); err != nil {
			return err
		}
	}
	err = ch.openFile.writer.WriteAll(records)
	return err
}

// Flush 强制刷盘（确保数据写入磁盘）
func (ch *CSVHandler) Flush() error {
	ch.openFile.mu.Lock()
	defer ch.openFile.mu.Unlock()

	ch.openFile.writer.Flush()
	if err := ch.openFile.writer.Error(); err != nil {
		if closeErro := ch.Close(); closeErro != nil {
			hwlog.RunLog.Errorf("failed to close csv file: %s, error: %v", ch.csvFilePath, closeErro)
		}
		return err
	}
	return nil
}

// Close 关闭文件（自动 Flush）
func (ch *CSVHandler) Close() error {
	ch.openFile.mu.Lock()
	defer ch.openFile.mu.Unlock()

	ch.openFile.writer.Flush() // 确保最终刷新
	if err := ch.openFile.writer.Error(); err != nil {
		return ch.openFile.file.Close()
	}
	return ch.openFile.file.Close()
}
