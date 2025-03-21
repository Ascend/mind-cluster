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
Package contextdata provides some test for environment.
*/
package contextdata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/enum"
)

func TestNewEnvironment(t *testing.T) {
	assert.NotNil(t, NewEnvironment())
}

func TestQueryNodeStatus(t *testing.T) {
	nodeStatus := queryNodeStatus()
	assert.Equal(t, nodeStatus.ChipType, enum.Ascend910A2)
	// TODO 补全收集服务器信息
}

func TestQueryClusterStatus(t *testing.T) {
	status := queryClusterStatus()
	assert.NotNil(t, status)
	// TODO 补全收集集群信息
}
