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
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	"taskd/toolkit_backend/net/proto"
)

// ExtractAckFrame extracts an Ack frame from a proto.Ack.
func ExtractAckFrame(ack *proto.Ack) *Ack {
	if ack == nil {
		return nil
	}
	return &Ack{
		Uuid: ack.Uuid,
		Code: ack.Code,
		Src: &Position{
			Role:        ack.Src.Role,
			ServerRank:  ack.Src.ServerRank,
			ProcessRank: ack.Src.ProcessRank,
		},
	}
}

// ExtractDataFrame extracts a Message frame from a proto.Message.
func ExtractDataFrame(msg *proto.Message) *Message {
	if msg == nil {
		return nil
	}
	return &Message{
		Uuid:    msg.Header.Uuid,
		BizType: msg.Header.Mtype,
		Src: &Position{
			Role:        msg.Header.Src.Role,
			ServerRank:  msg.Header.Src.ServerRank,
			ProcessRank: msg.Header.Src.ProcessRank,
		},
		Dst: &Position{
			Role:        msg.Header.Dst.Role,
			ServerRank:  msg.Header.Dst.ServerRank,
			ProcessRank: msg.Header.Dst.ProcessRank,
		},
		Body: msg.Body,
	}
}

// DataFrame creates a new proto.Message with the given parameters.
func DataFrame(uid, mtype, body string, src, dst *Position) *proto.Message {
	return &proto.Message{
		Header: &proto.MessageHeader{
			Uuid:  uid,
			Mtype: mtype,
			Src: &proto.Position{
				Role:        src.Role,
				ServerRank:  src.ServerRank,
				ProcessRank: src.ProcessRank,
			},
			Dst: &proto.Position{
				Role:        dst.Role,
				ServerRank:  dst.ServerRank,
				ProcessRank: dst.ProcessRank,
			},
		},
		Body: body,
	}
}

// RegisterReqFrame creates a new proto.RegisterReq with the given source position.
func RegisterReqFrame(src *Position) *proto.RegisterReq {
	return &proto.RegisterReq{
		Uuid: uuid.New().String(),
		Pos: &proto.Position{
			Role:        src.Role,
			ServerRank:  src.ServerRank,
			ProcessRank: src.ProcessRank,
		},
	}
}

// AckFrame creates a new proto.Ack with the given parameters.
func AckFrame(uid string, code uint32, src *Position) *proto.Ack {
	return &proto.Ack{
		Uuid: uid,
		Code: code,
		Src: &proto.Position{
			Role:        src.Role,
			ServerRank:  src.ServerRank,
			ProcessRank: src.ProcessRank,
		},
	}
}

// GetContextMetaData retrieves the metadata value from the context by key.
func GetContextMetaData(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || md == nil {
		return ""
	}
	v := md.Get(key)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

// SetContextMetaData sets the metadata in the context with the given key-value pairs.
func SetContextMetaData(ctx context.Context, kv map[string]string) context.Context {
	md := metadata.New(map[string]string{})
	for k, v := range kv {
		md.Set(k, v)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// IsNaturalNumberCommaSeparated checks if the input string is a comma-separated list of natural numbers.
func IsNaturalNumberCommaSeparated(s string) bool {
	parts := strings.Split(s, ",")
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 {
			return false
		}
	}
	return true
}

// ValidateAndCorrectFrame validates and corrects a proto.Message frame.
func ValidateAndCorrectFrame(msg *proto.Message) (int, error) {
	if msg == nil {
		return NilMessage, errors.New("message nil")
	}
	if msg.Header == nil {
		return NilHeader, errors.New("message header nil")
	}
	if msg.Header.Src == nil || msg.Header.Dst == nil {
		return NilPosition, errors.New("dst or src position nil")
	}
	if RoleLevel(msg.Header.Dst.Role) == -1 {
		return DstRoleIllegal, errors.New("dst role illegal")
	}
	if !IsNaturalNumberCommaSeparated(msg.Header.Dst.ServerRank) {
		if !strings.Contains(msg.Header.Dst.ServerRank, BroadCastPos) {
			return DstSrvRankIllegal, errors.New("dst server rank illegal")
		}
		msg.Header.Dst.ServerRank = BroadCastPos
	}
	if !RoleHasProcessProperty(msg.Header.Dst.Role) {
		msg.Header.Dst.ProcessRank = "-1"
		return OK, nil
	}
	if !IsNaturalNumberCommaSeparated(msg.Header.Dst.ProcessRank) {
		if !strings.Contains(msg.Header.Dst.ProcessRank, BroadCastPos) {
			return DstProcessRankIllegal, errors.New("dst process rank illegal")
		}
		msg.Header.Dst.ProcessRank = BroadCastPos
	}
	return OK, nil
}

// CheckConfig checks the validity of a TaskNetConfig.
func CheckConfig(conf *TaskNetConfig) error {
	if conf == nil {
		return errors.New("config nil")
	}
	if RoleLevel(conf.Pos.Role) == -1 {
		return errors.New("config position illegal")
	}
	if i, err := strconv.Atoi(conf.Pos.ServerRank); err != nil || i < 0 {
		return errors.New("config position illegal")
	}
	if RoleHasProcessProperty(conf.Pos.Role) {
		if i, err := strconv.Atoi(conf.Pos.ProcessRank); err != nil || i < 0 {
			return errors.New("config position illegal")
		}
	}
	return nil
}

// IsIPValid checks if the input string is a valid IP address.
func IsIPValid(ipStr string) error {
	parsedIp := net.ParseIP(ipStr)
	if parsedIp == nil {
		return errors.New("parse to ip failed")
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		return errors.New("not a valid ipv4 or ipv6 ip")
	}
	if parsedIp.IsUnspecified() {
		return errors.New("is all zeros ip")
	}
	if parsedIp.IsMulticast() {
		return errors.New("is multicast ip")
	}
	return nil
}

// GetIpFromAddr extracts the IP address from a string in the format "host:port".
func GetIpFromAddr(addr string) string {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return ip
}

// DstCase determines the destination case based on the current and destination positions.
func DstCase(cur, dst *Position) string {
	if cur == nil || dst == nil {
		return "unknown"
	}
	myLevel := RoleLevel(cur.Role)
	dstLevel := RoleLevel(dst.Role)
	if *dst == *cur {
		return Dst2Self
	}
	if dstLevel > myLevel {
		return Dst2UpperLevel
	}
	if dstLevel < myLevel {
		return Dst2LowerLevel
	}
	if cur.ServerRank == dst.ServerRank || dst.ServerRank == BroadCastPos {
		if !RoleHasProcessProperty(cur.Role) {
			return Dst2Self
		}
		if cur.ProcessRank == dst.ProcessRank || dst.ProcessRank == BroadCastPos {
			return Dst2Self
		}
	}
	return Dst2SameLevel
}

// IsBroadCast checks if the destination position is a broadcast position.
func IsBroadCast(dst *proto.Position) bool {
	return strings.Contains(dst.ServerRank, BroadCastPos) ||
		strings.Contains(dst.ProcessRank, BroadCastPos)
}
