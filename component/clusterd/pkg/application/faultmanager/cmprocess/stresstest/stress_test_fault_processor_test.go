// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package stresstest contain filtering fault handling method for stress test faults
package stresstest

import (
	"clusterd/pkg/domain/faultdomain"
	"github.com/agiledragon/gomonkey/v2"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func TestMain(m *testing.M) {
	err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	if err != nil {
		return
	}
	m.Run()
}

func TestStressTestFaultProcess(t *testing.T) {
	t.Run("TestStressTestFaultProcess, data is err case", func(t *testing.T) {
		ori := constant.OneConfigmapContent[*constant.SwitchInfo]{}
		res := StressTestProcessor.Process(ori)
		assert.NotNil(t, res)
	})
	t.Run("TestStressTestFaultProcess, data is normal case", func(t *testing.T) {
		oriDevInfo1 := make(map[string]*constant.DeviceInfo)
		ori := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
			AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
			UpdateConfigmap: nil,
		}
		StressTestProcessor.jobFilterFault = map[string][]string{
			"job1": {"node1"},
		}
		oriDevInfo1["node1"] = &constant.DeviceInfo{}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(StressTestProcessor, "processNodeFaults", func(constant.AdvanceDeviceFaultCm, string) {
			return
		})
		res := StressTestProcessor.Process(ori)
		assert.Equal(t, res, ori)
	})
}

func TestProcessNodeFaults(t *testing.T) {
	t.Run("TestProcessNodeFaults, has aic fault case", func(t *testing.T) {
		rank := "rank1"
		devInfo := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				rank: {
					{
						FaultCode: constant.StressTestHighLevelCode,
						NPUName:   rank,
					},
				},
			},
		}
		StressTestProcessor.processNodeFaults(devInfo, "node")
		assert.Equal(t, 0, len(devInfo.FaultDeviceList[rank]))
	})
}

func TestSetFilterAicFault(t *testing.T) {
	jobID := "job1"
	t.Run("SetFilterAicFault, filter is true", func(t *testing.T) {
		StressTestProcessor.SetFilterAicFault(jobID, []string{"node1"}, true)
		defer StressTestProcessor.SetFilterAicFault(jobID, []string{"node1"}, false)
		assert.Equal(t, 1, len(StressTestProcessor.jobFilterFault[jobID]))
	})
	t.Run("SetFilterAicFault, filter is true", func(t *testing.T) {
		StressTestProcessor.SetFilterAicFault(jobID, []string{"node1"}, true)
		StressTestProcessor.SetFilterAicFault(jobID, []string{"node1"}, false)
		assert.Equal(t, 0, len(StressTestProcessor.jobFilterFault[jobID]))
	})
}
