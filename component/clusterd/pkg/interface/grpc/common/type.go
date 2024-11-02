// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"clusterd/pkg/interface/grpc/pb"
)

// Notifier notify job manager publish signal
type Notifier struct {
	CreateTimeStamp int64
	pb.ProcessManageSignal
}

// JobHealthyMgr interface for job healthy status management
type JobHealthyMgr interface {
	GetJobHealthy(jobId string) (bool, []string)
	GetJobDeviceNumPerNode(jobId string) int
	NotifySignalSend(notifier *Notifier)
	ListenTaskScheduleResult(jobId string, strategy string)
	GetJobInfo(jobId string) (string, string, string)
	IsJobRunning(jobId string) bool
	JobExist(jobId string) bool
}

// Publisher publish signal and handle job schedule result
type Publisher interface {
	PublishSignal(signal *pb.ProcessManageSignal, expectStates MachineStates)
	NotifyJobSchedulerResult(success bool, taskId string, strategy string)
}

// SignalRetrySender have a method send
type SignalRetrySender interface {
	Send(signal *pb.ProcessManageSignal) error
}

// MachineStates a slice type of MachineState
type MachineStates []MachineState

// TaskResetInfo record task reset device information
type TaskResetInfo struct {
	RankList   []*TaskDevInfo
	UpdateTime int64
	RetryTime  int
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
