// Copyright (c) Huawei Technologies Co., Ltd. 2025-2026. All rights reserved.

// Package faultcodeevent contain fault code event log processor
package faultcodeevent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func init() {
	hwLogConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitRunLogger(hwLogConfig, nil); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
}

func TestBuildFaultDigest(t *testing.T) {
	t.Run("empty fault list", func(t *testing.T) {
		digest := buildFaultDigest(nil)
		assert.Equal(t, "", digest)
	})

	t.Run("single fault", func(t *testing.T) {
		faultList := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
		}
		digest := buildFaultDigest(faultList)
		assert.NotEmpty(t, digest)
	})

	t.Run("same faults produce same digest", func(t *testing.T) {
		faultList1 := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
			{ServerId: "s2", FaultCode: "80C98009", FaultLevel: "major", FaultTime: 200, DeviceId: "1"},
		}
		faultList2 := []constant.FaultDevice{
			{ServerId: "s2", FaultCode: "80C98009", FaultLevel: "major", FaultTime: 200, DeviceId: "1"},
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
		}
		assert.Equal(t, buildFaultDigest(faultList1), buildFaultDigest(faultList2))
	})

	t.Run("different faults produce different digest", func(t *testing.T) {
		faultList1 := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
		}
		faultList2 := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
			{ServerId: "s2", FaultCode: "80C98009", FaultLevel: "major", FaultTime: 200, DeviceId: "1"},
		}
		assert.NotEqual(t, buildFaultDigest(faultList1), buildFaultDigest(faultList2))
	})

	t.Run("subset of previous faults produces different digest", func(t *testing.T) {
		faultList1 := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
			{ServerId: "s2", FaultCode: "80C98009", FaultLevel: "major", FaultTime: 200, DeviceId: "1"},
		}
		faultList2 := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
		}
		assert.NotEqual(t, buildFaultDigest(faultList1), buildFaultDigest(faultList2))
	})

	t.Run("digest contains serverId, faultCode, faultLevel, faultTime and deviceId", func(t *testing.T) {
		faultList := []constant.FaultDevice{
			{ServerId: "s1", FaultCode: "80E01801", FaultLevel: "critical", FaultTime: 100, DeviceId: "0"},
		}
		digest := buildFaultDigest(faultList)
		assert.Contains(t, digest, "s1")
		assert.Contains(t, digest, "80E01801")
		assert.Contains(t, digest, "critical")
		assert.Contains(t, digest, "100")
		assert.Contains(t, digest, "0")
	})
}

func TestCollectActiveFaultCodes(t *testing.T) {
	t.Run("empty fault device list", func(t *testing.T) {
		jobFaultInfo := &constant.JobFaultInfo{
			FaultDevice: nil,
		}
		result := collectActiveFaultCodes(jobFaultInfo)
		assert.Equal(t, 0, len(result))
	})

	t.Run("collect with valid fault codes and timestamps", func(t *testing.T) {
		jobFaultInfo := &constant.JobFaultInfo{
			FaultDevice: []constant.FaultDevice{
				{ServerName: "node1", DeviceId: "0", FaultCode: "80E01801", FaultTime: 1700000000, FaultLevel: "critical"},
				{ServerName: "node1", DeviceId: "1", FaultCode: "80C98009", FaultTime: 1700000001, FaultLevel: "major"},
			},
		}
		result := collectActiveFaultCodes(jobFaultInfo)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "80E01801", result[0].FaultCode)
		assert.Equal(t, "node1", result[0].NodeName)
		assert.Equal(t, "0", result[0].DeviceId)
		assert.Equal(t, int64(1700000000), result[0].Timestamp)
		assert.Equal(t, "critical", result[0].FaultLevel)
	})

	t.Run("deduplicate same fault code and device", func(t *testing.T) {
		jobFaultInfo := &constant.JobFaultInfo{
			FaultDevice: []constant.FaultDevice{
				{ServerName: "node1", DeviceId: "0", FaultCode: "80E01801", FaultTime: 1700000000},
				{ServerName: "node1", DeviceId: "0", FaultCode: "80E01801", FaultTime: 1700000000},
			},
		}
		result := collectActiveFaultCodes(jobFaultInfo)
		assert.Equal(t, 1, len(result))
	})

	t.Run("fallback to current time when fault time is zero", func(t *testing.T) {
		jobFaultInfo := &constant.JobFaultInfo{
			FaultDevice: []constant.FaultDevice{
				{ServerName: "node1", DeviceId: "0", FaultCode: "80E01801", FaultLevel: "critical"},
			},
		}
		result := collectActiveFaultCodes(jobFaultInfo)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "80E01801", result[0].FaultCode)
		assert.Equal(t, "critical", result[0].FaultLevel)
		assert.Greater(t, result[0].Timestamp, int64(0))
	})
}

func TestAppendFaultCodes(t *testing.T) {
	t.Run("append to empty slice", func(t *testing.T) {
		newFaults := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "critical"},
		}
		result := appendFaultCodes("", nil, newFaults)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "80E01801", result[0].FaultCode)
	})

	t.Run("deduplicate existing fault codes", func(t *testing.T) {
		existing := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "critical"},
		}
		newFaults := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "critical"},
		}
		result := appendFaultCodes("", existing, newFaults)
		assert.Equal(t, 1, len(result))
	})

	t.Run("same fault code and timestamp with different fault level should append", func(t *testing.T) {
		existing := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "NotHandleFault"},
		}
		newFaults := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "PreSeparateFault"},
		}
		result := appendFaultCodes("", existing, newFaults)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "NotHandleFault", result[0].FaultLevel)
		assert.Equal(t, "PreSeparateFault", result[1].FaultLevel)
	})

	t.Run("append new fault codes", func(t *testing.T) {
		existing := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80E01801", Timestamp: 100, NodeName: "node1", DeviceId: "0", FaultLevel: "critical"},
		}
		newFaults := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80C98009", Timestamp: 200, NodeName: "node1", DeviceId: "1", FaultLevel: "major"},
		}
		result := appendFaultCodes("", existing, newFaults)
		assert.Equal(t, 2, len(result))
	})

	t.Run("overflow sliding window", func(t *testing.T) {
		var existing []constant.FaultCodeAndTimestamp
		for i := int64(0); i < constant.MaxTimestampRecords; i++ {
			existing = append(existing, constant.FaultCodeAndTimestamp{
				FaultCode: "80E01801", Timestamp: i, NodeName: "node1", DeviceId: "0",
			})
		}
		newFaults := []constant.FaultCodeAndTimestamp{
			{FaultCode: "80C98009", Timestamp: constant.MaxTimestampRecords, NodeName: "node1", DeviceId: "1"},
		}
		result := appendFaultCodes("", existing, newFaults)
		assert.Equal(t, constant.MaxTimestampRecords, len(result))
		assert.Equal(t, int64(1), result[0].Timestamp)
	})
}
