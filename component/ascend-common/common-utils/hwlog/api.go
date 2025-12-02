/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// CustomLoggerWriter implements io.Writer interface for customizing log time format
type CustomLoggerWriter struct {
	writer io.Writer
	mutex  sync.Mutex
}

// NewCustomLoggerWriter creates a new custom logger writer
func NewCustomLoggerWriter(w io.Writer) *CustomLoggerWriter {
	return &CustomLoggerWriter{
		writer: w,
	}
}

// Write implements io.Writer interface with custom time format (YYYY-MM-DD HH:MM:SS.000000)
// Note: In Go, time formatting uses reference time "2006-01-02 15:04:05.000000" as layout
// If current time is in Daylight Saving Time, append "DST" flag
func (w *CustomLoggerWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Get current time
	now := time.Now()
	// Format time with microsecond precision
	timestamp := now.Format("2006-01-02 15:04:05.000000")

	// Check if current time is in Daylight Saving Time
	isDST := now.IsDST()

	// Calculate buffer size: timestamp length + space + (DST flag length if needed) + log content length
	bufferSize := len(timestamp) + len(p) + 2 // +2 for '[' and ']'
	if isDST {
		bufferSize += 3 // +3 for "DST" flag
	}

	// Pre-allocate buffer to avoid multiple memory allocations
	buffer := make([]byte, 0, bufferSize)
	buffer = append(buffer, '[')
	buffer = append(buffer, timestamp...)

	// Append DST flag if needed
	if isDST {
		buffer = append(buffer, "DST"...)
	}

	buffer = append(buffer, ']')
	buffer = append(buffer, p...)

	return w.writer.Write(buffer)
}

const (
	logDebugLv = iota - 1
	logInfoLv
	logWarnLv
	logErrorLv
	logCriticalLv
)

type logger struct {
	lgDebug    *log.Logger
	lgInfo     *log.Logger
	lgWarn     *log.Logger
	lgError    *log.Logger
	lgCritical *log.Logger
	lgCtrl     *LogLimiter
	lgLevel    int
	lgMaxLine  int
}

func (lg *logger) initLogWriter(w io.Writer) {
	// Use custom logger writer, note that we don't use log.Ldate|log.Lmicroseconds flag to avoid duplicate timestamps
	// Custom writer will handle timestamp formatting
	customWriter := NewCustomLoggerWriter(w)
	// Only set prefix, not set time flag
	lg.lgDebug = log.New(customWriter, "[DEBUG]    ", 0)
	lg.lgInfo = log.New(customWriter, "[INFO]     ", 0)
	lg.lgWarn = log.New(customWriter, "[WARN]     ", 0)
	lg.lgError = log.New(customWriter, "[ERROR]    ", 0)
	lg.lgCritical = log.New(customWriter, "[CRITICAL] ", 0)
}

func (lg *logger) setLoggerLevel(lv int) {
	if lv < minLogLevel || lv > maxLogLevel {
		lg.lgLevel = 0
		return
	}
	lg.lgLevel = lv
}

func (lg *logger) setLoggerMaxLine(lml int) {
	if lml <= 0 || lml > maxEachLineLen {
		lg.lgMaxLine = defaultMaxEachLineLen
		return
	}
	lg.lgMaxLine = lml
}

func (lg *logger) setLoggerWriter(config *LogConfig) {
	rollLogger := &Logs{
		FileName:   config.LogFileName,
		Capacity:   config.FileMaxSize, // megabytes
		SaveVolume: config.MaxBackups,
		SaveTime:   config.MaxAge, // days
	}
	logWriter := &LogLimiter{
		Logs:        rollLogger,
		ExpiredTime: config.ExpiredTime, // seconds
		CacheSize:   config.CacheSize,
	}
	if config.OnlyToStdout {
		lg.initLogWriter(os.Stdout)
		return
	}
	if config.OnlyToFile {
		lg.initLogWriter(logWriter)
		return
	}
	writer := io.MultiWriter(os.Stdout, logWriter)
	lg.initLogWriter(writer)
	lg.lgCtrl = logWriter
}

func (lg *logger) setLogger(config *LogConfig) error {
	if err := validateLogConfigFiled(config); err != nil {
		return err
	}
	lg.setLoggerWriter(config)
	lg.setLoggerLevel(config.LogLevel)
	lg.setLoggerMaxLine(config.MaxLineLength)
	msg := fmt.Sprintf("%s's logger init success", path.Base(config.LogFileName))
	// skip change file mode and fs notify
	if config.OnlyToStdout {
		msg = fmt.Sprintf("%s, only to stdout", msg)
		return nil
	}
	lg.Info(msg)
	if err := os.Chmod(config.LogFileName, LogFileMode); err != nil {
		lg.Errorf("change file mode failed: %v", err)
		return fmt.Errorf("set log file mode failed")
	}
	return nil
}

func (lg *logger) isInit() bool {
	return lg.lgDebug != nil && lg.lgInfo != nil && lg.lgWarn != nil && lg.lgError != nil && lg.lgCritical != nil
}

// Debug record debug not format
func (lg *logger) Debug(args ...interface{}) {
	lg.DebugWithCtx(nil, args...)
}

// Debugf record debug
func (lg *logger) Debugf(format string, args ...interface{}) {
	lg.DebugfWithCtx(nil, format, args...)
}

// DebugWithCtx record Debug not format
func (lg *logger) DebugWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logDebugLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgDebug, fmt.Sprint(args...), lg.lgMaxLine, ctx)
	}
}

// DebugfWithCtx record Debug  format
func (lg *logger) DebugfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logDebugLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgDebug, fmt.Sprintf(format, args...), lg.lgMaxLine, ctx)
	}
}

// Info record info not format
func (lg *logger) Info(args ...interface{}) {
	lg.InfoWithCtx(nil, args...)
}

// Infof record info
func (lg *logger) Infof(format string, args ...interface{}) {
	lg.InfofWithCtx(nil, format, args...)
}

// InfoWithCtx record Info not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) InfoWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logInfoLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgInfo, fmt.Sprint(args...), lg.lgMaxLine, ctx)
	}
}

// InfofWithCtx record Info  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) InfofWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logInfoLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgInfo, fmt.Sprintf(format, args...), lg.lgMaxLine, ctx)
	}
}

// Warn record warn not format
func (lg *logger) Warn(args ...interface{}) {
	lg.WarnWithCtx(nil, args...)
}

// Warnf record warn
func (lg *logger) Warnf(format string, args ...interface{}) {
	lg.WarnfWithCtx(nil, format, args...)
}

// WarnWithCtx record Warn not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) WarnWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logWarnLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgWarn, fmt.Sprint(args...), lg.lgMaxLine, ctx)
	}
}

// WarnfWithCtx record Warn  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) WarnfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logWarnLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgWarn, fmt.Sprintf(format, args...), lg.lgMaxLine, ctx)
	}
}

// Error record error not format
func (lg *logger) Error(args ...interface{}) {
	lg.ErrorWithCtx(nil, args...)
}

// Errorf record error
func (lg *logger) Errorf(format string, args ...interface{}) {
	lg.ErrorfWithCtx(nil, format, args...)
}

// ErrorfWithLimit record error for default times (default 3),domain is for logType of msg,
// id is a unique identifier of this logType, you can reset the counter by call ResetErrCnt
func (lg *logger) ErrorfWithLimit(domain string, id interface{}, format string, args ...interface{}) {
	if needPrint, extraErrLog := IsNeedPrintWithSpecifiedCounts(domain, id, ProblemOccurMaxNumbers); needPrint {
		format = fmt.Sprintf("%s %s", format, extraErrLog)
		lg.ErrorfWithCtx(nil, format, args...)
	}
}

// ErrorfWithSpecifiedCounts record error for specified times,domain is for logType of msg,
// id is a unique identifier of this logType,maxCounts is for max print counts,
// you can reset the counter by call ResetErrCnt
func (lg *logger) ErrorfWithSpecifiedCounts(domain string, id interface{}, maxCounts int,
	format string, args ...interface{}) {
	if needPrint, extraErrLog := IsNeedPrintWithSpecifiedCounts(domain, id, maxCounts); needPrint {
		format = fmt.Sprintf("%s %s", format, extraErrLog)
		lg.ErrorfWithCtx(nil, format, args...)
	}
}

// ErrorWithCtx record Error not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) ErrorWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logErrorLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgError, fmt.Sprint(args...), lg.lgMaxLine, ctx)
	}
}

// ErrorfWithCtx record Error  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) ErrorfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logErrorLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgError, fmt.Sprintf(format, args...), lg.lgMaxLine, ctx)
	}
}

// Critical record critical not format
func (lg *logger) Critical(args ...interface{}) {
	lg.CriticalWithCtx(nil, args...)
}

// Criticalf record Critical log format
func (lg *logger) Criticalf(format string, args ...interface{}) {
	lg.CriticalfWithCtx(nil, format, args...)
}

// CriticalWithCtx record Critical not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) CriticalWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logCriticalLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgCritical, fmt.Sprint(args...), lg.lgMaxLine, ctx)
	}
}

// CriticalfWithCtx record Critical format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) CriticalfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logCriticalLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgCritical, fmt.Sprintf(format, args...), lg.lgMaxLine, ctx)
	}
}

func (lg *logger) validate() bool {
	if lg == nil || !lg.isInit() {
		fmt.Println("Fatal function's logger is nil")
		return false
	}
	return true
}

// FlushMem writes the contents of the memory to the disk
func (lg *logger) FlushMem() error {
	return lg.lgCtrl.Flush()
}
