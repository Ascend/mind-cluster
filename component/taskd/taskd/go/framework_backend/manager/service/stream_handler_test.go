/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// Test stream handler initialization
func TestStreamHandler_Init(t *testing.T) {
	sh := NewStreamHandler()
	err := sh.Init()
	assert.NoError(t, err)
	assert.NotNil(t, sh.Streams[constant.ProfilingStream])
	assert.Equal(t, constant.ProfilingStream, sh.Streams[constant.ProfilingStream].Name)
}

// Test getting an existing stream
func TestGetStream_Exist(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	stream := sh.GetStream(constant.ProfilingStream)
	assert.NotNil(t, stream)
	assert.Equal(t, constant.ProfilingStream, stream.Name)
}

// Test getting a non-existing stream
func TestGetStream_NotExist(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	stream := sh.GetStream("UnknownStream")
	assert.Nil(t, stream)
}

// Test successful token allocation
func TestAllocateToken_Success(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	err := sh.AllocateToken(constant.ProfilingStream, "plugin1")
	assert.NoError(t, err)
}

// Test token allocation for non-existing stream
func TestAllocateToken_StreamNotExist(t *testing.T) {
	sh := NewStreamHandler()
	err := sh.AllocateToken("UnknownStream", "plugin1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unregistered")
}

// Test duplicate token allocation (should fail)
func TestAllocateToken_Duplicate(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	_ = sh.AllocateToken(constant.ProfilingStream, "plugin1")
	err := sh.AllocateToken(constant.ProfilingStream, "plugin1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bind plugin failed")
}

// Test successful token release
func TestReleaseToken_Success(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	_ = sh.AllocateToken(constant.ProfilingStream, "plugin1")
	err := sh.ReleaseToken(constant.ProfilingStream, "plugin1")
	assert.NoError(t, err)
}

// Test releasing non-allocated token
func TestReleaseToken_NotAllocated(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	err := sh.ReleaseToken(constant.ProfilingStream, "plugin1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "release failed")
}

// Test stream token reset
func TestResetToken(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	_ = sh.AllocateToken(constant.ProfilingStream, "plugin1")
	err := sh.ResetToken(constant.ProfilingStream)
	assert.NoError(t, err)
	assert.Empty(t, sh.GetStream(constant.ProfilingStream).GetTokenOwner())
}

// Test resetting non-existing stream
func TestResetToken_StreamNotExist(t *testing.T) {
	sh := NewStreamHandler()
	err := sh.ResetToken("UnknownStream")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unregistered")
}

// Test request prioritization
func TestPrioritize(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	requests := []string{"req3", "req1", "req2"}
	sorted, err := sh.Prioritize(constant.ProfilingStream, requests)
	assert.NoError(t, err)
	expected := []string{"req1", "req2", "req3"} // Assuming Stream.Prioritize sorts alphabetically
	assert.Equal(t, expected, sorted)
}

// Test stream working status (busy)
func TestIsStreamWork_True(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	_ = sh.AllocateToken(constant.ProfilingStream, "plugin1")
	working, err := sh.IsStreamWork(constant.ProfilingStream)
	assert.NoError(t, err)
	assert.True(t, working)
}

// Test stream working status (idle)
func TestIsStreamWork_False(t *testing.T) {
	sh := NewStreamHandler()
	_ = sh.Init()
	working, err := sh.IsStreamWork(constant.ProfilingStream)
	assert.NoError(t, err)
	assert.False(t, working)
}

// Test working status for non-existing stream
func TestIsStreamWork_NotExist(t *testing.T) {
	sh := NewStreamHandler()
	working, err := sh.IsStreamWork("UnknownStream")
	assert.Error(t, err)
	assert.False(t, working)
}
