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

// Package main is the noded entry test
package main

import (
	"context"
	"flag"
	"testing"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		return
	}
}

// TestMonitorPeriodAlias tests the deprecated -monitorPeriod flag: it stays equivalent to
// -ipmiMonitorPeriod and takes precedence when both flags are explicitly set.
func TestMonitorPeriodAlias(t *testing.T) {
	// defaults: neither flag is explicitly set
	if !checkParameters() {
		t.Fatalf("checkParameters failed with default parameters")
	}
	setParameters()
	if common.ParamOption.IpmiMonitorPeriod != defaultIpmiMonitorPeriod {
		t.Fatalf("default ipmi monitor period should be %d, got %d",
			defaultIpmiMonitorPeriod, common.ParamOption.IpmiMonitorPeriod)
	}

	// only -ipmiMonitorPeriod is set
	if err := flag.Set("ipmiMonitorPeriod", "120"); err != nil {
		t.Fatalf("set ipmiMonitorPeriod failed: %v", err)
	}
	if !checkParameters() {
		t.Fatalf("checkParameters failed with valid ipmiMonitorPeriod")
	}
	setParameters()
	if common.ParamOption.IpmiMonitorPeriod != 120 {
		t.Fatalf("ipmiMonitorPeriod should take effect, got %d", common.ParamOption.IpmiMonitorPeriod)
	}

	// out-of-range ipmiMonitorPeriod blocks startup
	if err := flag.Set("ipmiMonitorPeriod", "999"); err != nil {
		t.Fatalf("set ipmiMonitorPeriod failed: %v", err)
	}
	if checkParameters() {
		t.Fatalf("checkParameters should fail when ipmiMonitorPeriod is out of range")
	}
	if err := flag.Set("ipmiMonitorPeriod", "120"); err != nil {
		t.Fatalf("set ipmiMonitorPeriod failed: %v", err)
	}

	// when both flags are set, monitorPeriod takes precedence
	if err := flag.Set("monitorPeriod", "90"); err != nil {
		t.Fatalf("set monitorPeriod failed: %v", err)
	}
	if !checkParameters() {
		t.Fatalf("checkParameters failed with valid monitorPeriod")
	}
	setParameters()
	if common.ParamOption.IpmiMonitorPeriod != 90 {
		t.Fatalf("monitorPeriod should take precedence, got %d", common.ParamOption.IpmiMonitorPeriod)
	}

	// out-of-range monitorPeriod blocks startup
	if err := flag.Set("monitorPeriod", "999"); err != nil {
		t.Fatalf("set monitorPeriod failed: %v", err)
	}
	if checkParameters() {
		t.Fatalf("checkParameters should fail when monitorPeriod is out of range")
	}
	if err := flag.Set("monitorPeriod", "90"); err != nil {
		t.Fatalf("set monitorPeriod failed: %v", err)
	}

	// monitorPeriod takes precedence even if ipmiMonitorPeriod is out of range
	if err := flag.Set("ipmiMonitorPeriod", "999"); err != nil {
		t.Fatalf("set ipmiMonitorPeriod failed: %v", err)
	}
	if !checkParameters() {
		t.Fatalf("checkParameters should pass since monitorPeriod takes precedence")
	}
	setParameters()
	if common.ParamOption.IpmiMonitorPeriod != 90 {
		t.Fatalf("monitorPeriod should take precedence, got %d", common.ParamOption.IpmiMonitorPeriod)
	}
}
