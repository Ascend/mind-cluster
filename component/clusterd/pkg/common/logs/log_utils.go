// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package logs a series of statistic function
package logs

import (
	"ascend-common/common-utils/hwlog"
	"context"
)

const (
	jobEventLog              = "/var/log/mindx-dl/clusterd/event_job.log"
	jobEventMaxBackupLogs    = 5
	jobEventMaxLogLineLength = 2048
	jobEventMaxAge           = 40
)

var (
	jobEventHwLogConfig = &hwlog.LogConfig{LogFileName: jobEventLog, MaxBackups: jobEventMaxBackupLogs,
		MaxLineLength: jobEventMaxLogLineLength, MaxAge: jobEventMaxAge}
	JobEventLog *hwlog.CustomLogger
)

// InitJobEventLogger init JobEventLog
func InitJobEventLogger(ctx context.Context) error {
	customLog, err := hwlog.NewCustomLogger(jobEventHwLogConfig, ctx)
	if err != nil {
		return err
	}
	JobEventLog = customLog
	return nil
}
