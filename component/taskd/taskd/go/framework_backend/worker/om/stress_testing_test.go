// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

func TestStressTestProcessMsg(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		StressTestProcessMsg(nil)
		assert.Empty(t, msgChan)
	})
	t.Run("invalid message body", func(t *testing.T) {
		msg := &common.Message{Body: "invalid json"}
		StressTestProcessMsg(msg)
		assert.Empty(t, msgChan)
	})
	t.Run("missing parameters", func(t *testing.T) {
		body := storage.MsgBody{
			Extension: map[string]string{
				constant.StressTestUUID:      "",
				constant.StressTestRankOPStr: "",
			},
		}
		msg := &common.Message{Body: utils.ObjToString(body)}
		StressTestProcessMsg(msg)
		assert.Empty(t, msgChan)
	})
	t.Run("valid message", func(t *testing.T) {
		body := storage.MsgBody{
			Extension: map[string]string{
				constant.StressTestUUID:      "uuid",
				constant.StressTestRankOPStr: "123",
			},
		}
		msg := &common.Message{Body: utils.ObjToString(body)}
		StressTestProcessMsg(msg)
		assert.Equal(t, 1, len(msgChan))
		body2 := <-msgChan
		assert.Equal(t, body.Extension[constant.StressTestUUID], body2.Extension[constant.StressTestUUID])
	})
}

func TestHandleStressTestMsg(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		HandleStressTestMsg(ctx, 0)
		assert.Empty(t, hbChan)
	})
	t.Run("invalid OPstr", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go HandleStressTestMsg(ctx, 0)
		time.Sleep(time.Second)
		msgChan <- storage.MsgBody{
			Extension: map[string]string{constant.StressTestUUID: "uid", constant.StressTestRankOPStr: "string"}}
		assert.Empty(t, hbChan)
	})
	t.Run("invalid rank", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go HandleStressTestMsg(ctx, 1)
		time.Sleep(time.Second)
		opStr := map[string]*pb.StressOpList{
			"0": {Ops: []int64{1}}}
		msgChan <- storage.MsgBody{
			Extension: map[string]string{
				constant.StressTestUUID: "uid", constant.StressTestRankOPStr: utils.ObjToString(opStr),
			},
		}
		assert.Empty(t, hbChan)
	})
	t.Run("invalid operations", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go HandleStressTestMsg(ctx, 0)
		time.Sleep(time.Second)
		opStr := map[string]*pb.StressOpList{
			"0": {Ops: []int64{1, 2, 3, 4}}}
		msgChan <- storage.MsgBody{
			Extension: map[string]string{
				constant.StressTestUUID: "uid", constant.StressTestRankOPStr: utils.ObjToString(opStr),
			},
		}
		assert.Empty(t, hbChan)
	})
}

func TestHandleStressTestMsgOK(t *testing.T) {
	t.Run("valid operations", func(t *testing.T) {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFunc(sendHeartBeatMsg, func(ctx context.Context) {
			return
		}).ApplyFunc(doStressTest, func(ops []int64) string {
			return "ok"
		}).ApplyFunc(notifyStressTestResult, func(result, uid string) {
			return
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go HandleStressTestMsg(ctx, 0)
		time.Sleep(time.Second)
		opStr := map[string]*pb.StressOpList{
			"0": {Ops: []int64{1}}}
		msgChan <- storage.MsgBody{
			Extension: map[string]string{
				constant.StressTestUUID: "uid", constant.StressTestRankOPStr: utils.ObjToString(opStr),
			},
		}
		<-hbChan
		assert.Empty(t, hbChan)
	})
}

func TestSendHeartBeatMsg(t *testing.T) {
	patches := gomonkey.NewPatches()
	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		called := false
		patches.ApplyMethodFunc(StressTestNetTool, "SyncSendMessage", func(_, _, _ string, _ *common.Position) (*common.Ack, error) {
			called = true
			return &common.Ack{}, nil
		})
		defer patches.Reset()
		cancel()
		sendHeartBeatMsg(ctx)
		assert.False(t, called)
	})
	t.Run("stop signal received", func(t *testing.T) {
		called := false
		patches.ApplyMethodFunc(StressTestNetTool, "SyncSendMessage", func(_, _, _ string, _ *common.Position) (*common.Ack, error) {
			called = true
			return &common.Ack{}, nil
		})
		defer func() {
			patches.Reset()
			close(hbChan)
			hbChan = make(chan struct{}, 1)
		}()
		close(hbChan)
		hbChan = make(chan struct{}, 1)
		hbChan <- struct{}{}
		sendHeartBeatMsg(context.Background())
		assert.False(t, called)
	})
	t.Run("normal heart beat", func(t *testing.T) {
		lock := sync.Mutex{}
		ctx, cancel := context.WithCancel(context.Background())
		StressTestNetTool = &net.NetInstance{}
		called := false
		patches.ApplyMethodFunc(StressTestNetTool, "SyncSendMessage", func(_, _, _ string, _ *common.Position) (*common.Ack, error) {
			lock.Lock()
			called = true
			lock.Unlock()
			return &common.Ack{}, nil
		})
		go sendHeartBeatMsg(ctx)
		time.Sleep(time.Second)
		lock.Lock()
		assert.True(t, called)
		lock.Unlock()
		cancel()
		patches.Reset()
	})
}

func TestNotifyStressTestResult(t *testing.T) {
	patches := gomonkey.NewPatches()
	t.Run("net tool not initialized", func(t *testing.T) {
		originTool := StressTestNetTool
		StressTestNetTool = nil
		defer func() {
			StressTestNetTool = originTool
		}()
		called := false
		patches.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
			called = true
		})
		defer patches.Reset()
		notifyStressTestResult("success", "test-uid")
		assert.False(t, called)
	})
	t.Run("notify success", func(t *testing.T) {
		originTool := StressTestNetTool
		StressTestNetTool = &net.NetInstance{}
		defer func() {
			StressTestNetTool = originTool
		}()
		called := false
		patches.ApplyMethodFunc(StressTestNetTool, "SyncSendMessage", func(uuid, mtype, msgBody string,
			dst *common.Position) (*common.Ack, error) {
			called = true
			return &common.Ack{}, nil
		})
		defer patches.Reset()
		notifyStressTestResult("success", "test-uid")
		assert.True(t, called)
	})
}
