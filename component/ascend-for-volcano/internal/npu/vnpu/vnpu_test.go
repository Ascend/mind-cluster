/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/framework"

	test2 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type preStartActionTestCase struct {
	name    string
	env     *plugin.ScheduleEnv
	ssn     *framework.Session
	wantErr bool
}

func buildPreStartActionTestCase01() preStartActionTestCase {
	return preStartActionTestCase{
		name:    "01 will return err when ssn is nil",
		env:     &plugin.ScheduleEnv{},
		ssn:     nil,
		wantErr: true,
	}
}

func buildPreStartActionTestCase02() preStartActionTestCase {
	return preStartActionTestCase{
		name:    "02 will return nil when ssn is not nil",
		env:     test2.FakeScheduleEnv(),
		ssn:     test.FakeNormalSSN(nil),
		wantErr: false,
	}
}

func buildPreStartActionTestCases() []preStartActionTestCase {
	return []preStartActionTestCase{
		buildPreStartActionTestCase01(),
		buildPreStartActionTestCase02(),
	}
}

func TestVirtualNPUPreStartAction(t *testing.T) {
	for _, tt := range buildPreStartActionTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tp := &VirtualNPU{}
			if err := tp.PreStartAction(tt.env, tt.ssn); (err != nil) != tt.wantErr {
				t.Errorf("PreStartAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
