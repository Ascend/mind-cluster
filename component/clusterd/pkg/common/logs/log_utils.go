// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package logs common func about logs
package logs

import (
	"context"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	jobEventLog              = "/var/log/mindx-dl/clusterd/event_job.log"
	jobEventMaxBackupLogs    = 5
	jobEventMaxLogLineLength = 2048
	jobEventMaxAge           = 40

	grpcEventLog           = "/var/log/mindx-dl/clusterd/grpc/grpc_event.log"
	grpcEventMaxBackupLogs = 14
	// grpcEventMaxLogLineLength to support 256 server
	grpcEventMaxLogLineLength = 524288
	grpcEventMaxAge           = 30
)

var (
	jobEventHwLogConfig = &hwlog.LogConfig{LogFileName: jobEventLog, MaxBackups: jobEventMaxBackupLogs,
		MaxLineLength: jobEventMaxLogLineLength, MaxAge: jobEventMaxAge, OnlyToFile: true}
	// JobEventLog is used to log job event
	JobEventLog        *hwlog.CustomLogger
	grpcEventLogConfig = &hwlog.LogConfig{LogFileName: grpcEventLog, MaxBackups: grpcEventMaxBackupLogs,
		MaxLineLength: grpcEventMaxLogLineLength, MaxAge: grpcEventMaxAge, OnlyToFile: true}
	// GrpcEventLogger is used to log grpc event
	GrpcEventLogger *hwlog.CustomLogger
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

// RecordLog record log
func RecordLog(role, event, result string) {
	switch result {
	case constant.Start, constant.Success:
		hwlog.RunLog.Infof("role[%s] %s %s", role, event, result)
	case constant.Failed:
		hwlog.RunLog.Errorf("role[%s] %s %s", role, event, result)
	default:
		hwlog.RunLog.Error("invalid event result")
	}
}

// InitGrpcEventLogger init GrpcEventLog
func InitGrpcEventLogger(ctx context.Context) error {
	grpcLog, err := hwlog.NewCustomLogger(grpcEventLogConfig, ctx)
	if err != nil {
		return err
	}
	GrpcEventLogger = grpcLog
	return nil
}
