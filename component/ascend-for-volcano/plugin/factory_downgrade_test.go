/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import "testing"

type schedulerDowngradeTimeoutTest struct {
	name string
	conf map[string]string
	want int
}

func buildSchedulerDowngradeTimeoutTestCases() []schedulerDowngradeTimeoutTest {
	return []schedulerDowngradeTimeoutTest{
		{name: "01 nil conf falls back to default", conf: nil, want: DefaultSchedulerDowngradeTimeout},
		{name: "02 absent key falls back to default", conf: map[string]string{}, want: DefaultSchedulerDowngradeTimeout},
		{name: "03 invalid value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "abc"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "04 non-positive value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "-5"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "05 zero value falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "0"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "06 value over the upper bound falls back to default",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "86401"}, want: DefaultSchedulerDowngradeTimeout},
		{name: "07 lower bound value", conf: map[string]string{schedulerDowngradeTimeoutKey: "1"}, want: 1},
		{name: "08 upper bound value",
			conf: map[string]string{schedulerDowngradeTimeoutKey: "86400"}, want: 86400},
		{name: "09 valid value", conf: map[string]string{schedulerDowngradeTimeoutKey: "30"}, want: 30},
	}
}

func TestGetSchedulerDowngradeTimeout(t *testing.T) {
	for _, tt := range buildSchedulerDowngradeTimeoutTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSchedulerDowngradeTimeout(tt.conf); got != tt.want {
				t.Errorf("getSchedulerDowngradeTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
