// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"clusterd/pkg/interface/grpc/pb"
)

// SignalRetrySender have a method send
type SignalRetrySender interface {
	Send(signal *pb.ProcessManageSignal) error
}

// TaskResetInfo record task reset device information
type TaskResetInfo struct {
	RankList      []*TaskDevInfo
	UpdateTime    int64
	RetryTime     int
	FaultFlushing bool
	GracefulExit  int
}

// TaskDevInfo is the device info of a task
type TaskDevInfo struct {
	RankId int
	DevFaultInfo
}

// DevFaultInfo is the fault info of device
type DevFaultInfo struct {
	LogicId       int32
	Status        string
	Policy        string
	InitialPolicy string
	ErrorCode     []int64
	ErrorCodeHex  string
}

// RecoverConfig is config for recover service
type RecoverConfig struct {
	ProcessRescheduleOn   bool
	MindXConfigStrategies []string
	PlatFormMode          bool
}

// JobBaseInfo job base info
type JobBaseInfo struct {
	JobId     string
	JobName   string
	PgName    string
	Namespace string
	RecoverConfig
}

// RecoverResult recover result
type RecoverResult struct {
	Strategy       string
	Code           RespCode
	RecoverSuccess bool
}
