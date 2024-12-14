// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process

package faultmanager

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

// FaultJob contain some fault info about a fault job
type FaultJob struct {
	IsA3Job             bool
	NameSpace           string
	PodNames            map[string]string
	RelationFaults      []*faultInfo
	TriggerFault        []faultInfo
	processedFaultInfo  []faultInfo
	FaultStrategy       FaultStrategy
	SeparateNodes       sets.String
	AllFaultCode        sets.String
	ProcessingFaultCode sets.String
	PodStrategiesMaps   map[string]string
	FindNPUUnderSwitch  bool
}

type faultInfo struct {
	FaultUid         string
	FaultType        string
	NodeName         string
	NPUName          string
	FaultCode        string
	FaultLevel       string
	FaultTime        int64
	ExecutedStrategy string
	DealMaxTime      int64
}

// FaultStrategy fault strategies
type FaultStrategy struct {
	NodeLvList   map[string]string
	DeviceLvList map[string][]DeviceStrategy
}

// RelationFaultStrategy relation fault strategy
type RelationFaultStrategy struct {
	TriggerFault   string
	RelationFaults []string
	FaultStrategy  string
}

// FaultDuration fault duration config
type FaultDuration struct {
	FaultCode       string
	FaultType       string
	TimeOutInterval int64
}

// DeviceStrategy device fault strategy
type DeviceStrategy struct {
	Strategy string
	NPUName  string
}
