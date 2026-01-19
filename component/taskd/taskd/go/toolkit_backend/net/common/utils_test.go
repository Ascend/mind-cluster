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
// Package common defines common constants and types used by the toolkit backend.
package common

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"taskd/toolkit_backend/net/proto"
)

const (
	emptyStr    = ""
	demoIp      = "127.0.0.1"
	correctAddr = demoIp + ":8899"
	wrongAddr   = demoIp + "-8899"
)

func TestExtractAckFrame(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, ExtractAckFrame(nil))
	})

	t.Run("valid input", func(t *testing.T) {
		ack := &proto.Ack{
			Uuid: "test-uuid",
			Code: 0,
			Src: &proto.Position{
				Role:        "test-role",
				ServerRank:  "test-server-rank",
				ProcessRank: "test-process-rank",
			},
		}
		result := ExtractAckFrame(ack)
		assert.Equal(t, ack.Uuid, result.Uuid)
		assert.Equal(t, ack.Code, result.Code)
		assert.Equal(t, ack.Src.Role, result.Src.Role)
		assert.Equal(t, ack.Src.ServerRank, result.Src.ServerRank)
		assert.Equal(t, ack.Src.ProcessRank, result.Src.ProcessRank)
	})
}

func TestExtractDataFrame(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, ExtractDataFrame(nil))
	})

	t.Run("valid input", func(t *testing.T) {
		msg := &proto.Message{
			Header: &proto.MessageHeader{
				Uuid:  "test-uuid",
				Mtype: "test-type",
				Src: &proto.Position{
					Role:        "src-role",
					ServerRank:  "src-server-rank",
					ProcessRank: "src-process-rank",
				},
				Dst: &proto.Position{
					Role:        "dst-role",
					ServerRank:  "dst-server-rank",
					ProcessRank: "dst-process-rank",
				},
			},
			Body: "test-body",
		}
		result := ExtractDataFrame(msg)
		assert.Equal(t, msg.Header.Uuid, result.Uuid)
		assert.Equal(t, msg.Header.Mtype, result.BizType)
		assert.Equal(t, msg.Header.Src.Role, result.Src.Role)
		assert.Equal(t, msg.Header.Dst.Role, result.Dst.Role)
		assert.Equal(t, msg.Body, result.Body)
	})
}

func TestDataFrame(t *testing.T) {
	src := &Position{
		Role:        "src-role",
		ServerRank:  "src-server-rank",
		ProcessRank: "src-process-rank",
	}
	dst := &Position{
		Role:        "dst-role",
		ServerRank:  "dst-server-rank",
		ProcessRank: "dst-process-rank",
	}

	msg := DataFrame("test-uuid", "test-type", "test-body", src, dst)
	assert.Equal(t, "test-uuid", msg.Header.Uuid)
	assert.Equal(t, "test-type", msg.Header.Mtype)
	assert.Equal(t, src.Role, msg.Header.Src.Role)
	assert.Equal(t, dst.Role, msg.Header.Dst.Role)
	assert.Equal(t, "test-body", msg.Body)
}

func TestRegisterReqFrame(t *testing.T) {
	src := &Position{
		Role:        "test-role",
		ServerRank:  "test-server-rank",
		ProcessRank: "test-process-rank",
	}
	req := RegisterReqFrame(src)
	assert.NotEmpty(t, req.Uuid)
	assert.Equal(t, src.Role, req.Pos.Role)
	assert.Equal(t, src.ServerRank, req.Pos.ServerRank)
	assert.Equal(t, src.ProcessRank, req.Pos.ProcessRank)
}

func TestAckFrame(t *testing.T) {
	src := &Position{
		Role:        "test-role",
		ServerRank:  "test-server-rank",
		ProcessRank: "test-process-rank",
	}
	ack := AckFrame("test-uuid", 0, src)
	assert.Equal(t, "test-uuid", ack.Uuid)
	assert.Equal(t, uint32(0), ack.Code)
	assert.Equal(t, src.Role, ack.Src.Role)
	assert.Equal(t, src.ServerRank, ack.Src.ServerRank)
	assert.Equal(t, src.ProcessRank, ack.Src.ProcessRank)
}

func TestGetContextMetaData(t *testing.T) {
	t.Run("no metadata", func(t *testing.T) {
		assert.Empty(t, GetContextMetaData(context.Background(), "test-key"))
	})

	t.Run("with metadata", func(t *testing.T) {
		md := metadata.New(map[string]string{"test-key": "test-value"})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		assert.Equal(t, "test-value", GetContextMetaData(ctx, "test-key"))
	})
}

func TestSetContextMetaData(t *testing.T) {
	kv := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	ctx := SetContextMetaData(context.Background(), kv)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "value1", md.Get("key1")[0])
	assert.Equal(t, "value2", md.Get("key2")[0])
}

func TestIsNaturalNumberCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid single number", "123", true},
		{"valid multiple numbers", "1,2,3", true},
		{"invalid negative", "-1", false},
		{"invalid string", "abc", false},
		{"mixed valid and invalid", "1,a,3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNaturalNumberCommaSeparated(tt.input))
		})
	}
}

func TestValidateAndCorrectFrame(t *testing.T) {
	tests := []struct {
		name        string
		msg         *proto.Message
		expectedErr error
	}{
		{
			"nil message",
			nil,
			errors.New("message nil"),
		},
		{
			"nil header",
			&proto.Message{},
			errors.New("message header nil"),
		},
		{
			"nil positions",
			&proto.Message{Header: &proto.MessageHeader{}},
			errors.New("dst or src position nil"),
		},
		{
			"invalid dst role",
			&proto.Message{
				Header: &proto.MessageHeader{
					Src: &proto.Position{},
					Dst: &proto.Position{Role: "invalid"},
				},
			},
			errors.New("dst role illegal"),
		},
		{
			"valid message",
			&proto.Message{
				Header: &proto.MessageHeader{
					Src: &proto.Position{},
					Dst: &proto.Position{Role: MgrRole, ServerRank: "0"},
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAndCorrectFrame(tt.msg)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *TaskNetConfig
		expectedErr error
	}{
		{
			"nil config",
			nil,
			errors.New("config nil"),
		},
		{
			"invalid role",
			&TaskNetConfig{
				Pos: Position{Role: "invalid"},
			},
			errors.New("config position illegal"),
		},
		{
			"invalid server rank",
			&TaskNetConfig{
				Pos: Position{Role: MgrRole, ServerRank: "-1"},
			},
			errors.New("config position illegal"),
		},
		{
			"invalid process rank for worker",
			&TaskNetConfig{
				Pos: Position{Role: WorkerRole, ServerRank: "0", ProcessRank: "-1"},
			},
			errors.New("config position illegal"),
		},
		{
			"valid config",
			&TaskNetConfig{
				Pos: Position{Role: MgrRole, ServerRank: "0"},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckConfig(tt.config)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestGetIpFromAddr(t *testing.T) {
	convey.Convey("get ip should right", t, func() {
		addr := GetHostFromAddr(correctAddr)
		convey.ShouldEqual(addr, demoIp)
		addr = GetHostFromAddr(wrongAddr)
		convey.ShouldEqual(addr, emptyStr)
	})
}

func TestDstCase(t *testing.T) {
	convey.Convey("Test DstCase function", t, func() {
		cur := &Position{Role: MgrRole}
		dst := &Position{Role: WorkerRole, ServerRank: "0", ProcessRank: "0"}
		dst2 := &Position{Role: WorkerRole, ServerRank: "1", ProcessRank: "1"}
		broadcast := &Position{Role: WorkerRole, ServerRank: BroadCastPos, ProcessRank: BroadCastPos}

		convey.Convey("When either position is nil", func() {
			convey.So(DstCase(nil, dst), convey.ShouldEqual, "unknown")
			convey.So(DstCase(cur, nil), convey.ShouldEqual, "unknown")
		})

		convey.Convey("When positions are equal", func() {
			convey.So(DstCase(cur, cur), convey.ShouldEqual, Dst2Self)
		})

		convey.Convey("When destination is higher level", func() {
			convey.So(DstCase(dst, cur), convey.ShouldEqual, Dst2UpperLevel)
		})

		convey.Convey("When destination is lower level", func() {
			convey.So(DstCase(cur, dst), convey.ShouldEqual, Dst2LowerLevel)
		})

		convey.Convey("When destination is same level", func() {
			convey.So(DstCase(dst, dst2), convey.ShouldEqual, Dst2SameLevel)
		})

		convey.Convey("When destination is broadcast", func() {
			convey.So(DstCase(dst, broadcast), convey.ShouldEqual, Dst2Self)
		})
	})
}

func TestIsBroadCast(t *testing.T) {
	convey.Convey("judge broadcast should right", t, func() {
		broadcast := &proto.Position{Role: WorkerRole, ServerRank: BroadCastPos, ProcessRank: BroadCastPos}
		notBroadcast := &proto.Position{Role: WorkerRole, ServerRank: "0", ProcessRank: "0"}
		convey.ShouldBeTrue(IsBroadCast(broadcast))
		convey.ShouldBeFalse(IsBroadCast(notBroadcast))
	})
}
