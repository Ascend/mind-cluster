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

// Package grpc is a DT collection for func in grpc
package grpc

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/grpc/profiling"
)

var (
	connectFailed = false
)

func TestMain(m *testing.M) {
	// mock grpc.NewClient
	mockNewClient := gomonkey.ApplyFunc(grpc.Dial, func(string, ...grpc.DialOption) (*grpc.ClientConn, error) {
		if connectFailed {
			return nil, fmt.Errorf("failed to connect to grpc server")
		}
		return &grpc.ClientConn{}, nil
	})

	defer mockNewClient.Reset()

	// Run the tests
	code := m.Run()

	// Exit with the test result code
	fmt.Printf("exit_code = %v\n", code)
}

func TestGrpc(t *testing.T) {
	// test connect
	connectFailed = true
	_, err := GetClient()
	assert.NotNil(t, err)
	connectFailed = false

	// test multiple connect
	client1, err := GetClient()
	assert.NotNil(t, err)
	client2, err := GetClient()
	assert.NotNil(t, err)
	assert.Equal(t, client1, client2)
}

func TestProfiling(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.GetClusterIp, func() string {
		return "127.0.0.1"
	})
	defer patches.Reset()
	connErr = nil
	client = nil
	client, err := GetClient()
	assert.Nil(t, err)

	// mock profilingSwitch
	patches.ApplyPrivateMethod(reflect.TypeOf(client), "profilingSwitch",
		func(*Client, *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
			return &profiling.DataTypeRes{
				Message: "success",
				Code:    0,
			}, nil
		})

	// test start all profiling
	err = client.StartAllProfiling("job1", "ns1")
	assert.Nil(t, err)

	// test stop all profiling
	err = client.StopAllProfiling("job1", "ns1")
	assert.Nil(t, err)

	// test start heavy profiling
	err = client.StartHeavyProfiling("job1", "ns1")
	assert.Nil(t, err)

	// test stop heavy profiling
	err = client.StopHeavyProfiling("job1", "ns1")
	assert.Nil(t, err)
}
