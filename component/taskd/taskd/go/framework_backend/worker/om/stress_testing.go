// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

/*
#include <stdlib.h>
#include "stress.h"
*/
import "C"

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	autils "ascend-common/common-utils/utils"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

const fieldCount = 2
const aicOp = 0
const p2pOp = 1
const aicTimeout time.Duration = 300
const p2pTimeout time.Duration = 600

// StressTestNetTool worker net tool
var StressTestNetTool *net.NetInstance

// stressTestCallback switch callback func
var stressTestCallback C.stress_test_callback_func

// RegisterStressTestCallback register stress test callback func
func RegisterStressTestCallback(ptr uintptr) {
	stressTestCallback = (C.stress_test_callback_func)(unsafe.Pointer(ptr))
}

var hbChan chan struct{} // heart beat chan
var msgChan chan storage.MsgBody
var opTimeoutMap = map[int]time.Duration{
	aicOp: aicTimeout,
	p2pOp: p2pTimeout,
}

func init() {
	hbChan = make(chan struct{}, 1)
	msgChan = make(chan storage.MsgBody, 1)
}

// HandleStressTestMsg dead loop for handle stress test msg
func HandleStressTestMsg(ctx context.Context, globalRank int) {
	hwlog.RunLog.Infof("start to watch for stress test")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("stress test received exit signal")
			return
		case msg := <-msgChan:
			uid := msg.Extension[constant.StressTestUUID]
			rankOpStr := msg.Extension[constant.StressTestRankOPStr]
			rankOp, err := utils.StringToObj[map[string]*pb.StressOpList](rankOpStr)
			if err != nil {
				hwlog.RunLog.Errorf("failed to marshal, err: %v", err)
				continue
			}
			ops, ok := rankOp[strconv.Itoa(globalRank)]
			if !ok {
				hwlog.RunLog.Errorf("not find ops for rank %v", globalRank)
				continue
			}
			if err := autils.CheckSliceSupport(ops.Ops, []int64{aicOp, p2pOp}); err != nil {
				hwlog.RunLog.Errorf("check op support failed, err: %v", err)
				continue
			}
			hwlog.RunLog.Infof("worker recv uuid: %v, rankOp: %v", uid, rankOp)
			go sendHeartBeatMsg(ctx)
			result := doStressTest(ops.Ops)
			notifyStressTestResult(result, uid)
			hbChan <- struct{}{}
		}
	}
}

func sendHeartBeatMsg(ctx context.Context) {
	hwlog.RunLog.Info("start send heart beat msg")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("stress test received exit signal")
			return
		case <-hbChan:
			hwlog.RunLog.Info("stop send heart beat msg")
			return
		default:
			msg := &storage.MsgBody{
				MsgType: constant.KeepAlive,
			}
			_, err := StressTestNetTool.SyncSendMessage(uuid.New().String(), "default", utils.ObjToString(msg), &common.Position{
				Role:       common.MgrRole,
				ServerRank: "0",
			})
			if err != nil {
				hwlog.RunLog.Errorf("send result to mgr err: %v", err)
			} else {
				hwlog.RunLog.Debugf("send result to mgr: %v", msg)
			}
			time.Sleep(time.Second)
		}
	}
}

// StressTestProcessMsg process stress test msg
func StressTestProcessMsg(msg *common.Message) {
	if msg == nil {
		hwlog.RunLog.Error("msg is nil")
		return
	}
	body, err := utils.StringToObj[storage.MsgBody](msg.Body)
	if err != nil {
		hwlog.RunLog.Errorf("get msgBody err: %v, msgBody is %v", err, body)
		return
	}
	uid := body.Extension[constant.StressTestUUID]
	rankOpStr := body.Extension[constant.StressTestRankOPStr]
	if uid == "" || rankOpStr == "" {
		hwlog.RunLog.Debugf("failed to get param, uid: %v, rankOpStr: %#v", uid, rankOpStr)
		return
	}
	msgChan <- body
}

func notifyStressTestResult(result, uid string) {
	if StressTestNetTool == nil {
		hwlog.RunLog.Error("StressTestNetTool for worker is nil")
		return
	}
	msg := storage.MsgBody{
		MsgType: constant.Action,
		Code:    constant.StressTestCode,
		Message: result,
		Extension: map[string]string{
			constant.StressTestUUID: uid,
		},
	}
	_, err := StressTestNetTool.SyncSendMessage(uuid.New().String(), "default", utils.ObjToString(msg), &common.Position{
		Role:       common.MgrRole,
		ServerRank: "0",
	})

	if err != nil {
		hwlog.RunLog.Errorf("send result to mgr err: %v", err)
		return
	}
	hwlog.RunLog.Infof("notify mgr result %v succeeded, msg: %s", result, utils.ObjToString(msg))
}

func doStressTest(ops []int64) string {
	opsInt := make([]int, 0, len(ops))
	for _, op := range ops {
		opsInt = append(opsInt, int(op))
	}
	sort.Ints(opsInt)
	rankResult := &pb.StressTestRankResult{
		RankResult: map[string]*pb.StressTestOpResult{},
	}
	for _, op := range opsInt {
		rankResult.RankResult[strconv.Itoa(op)] = &pb.StressTestOpResult{
			Code:   constant.StressTestExecFail,
			Result: "can not exec stress test",
		}
	}
	if stressTestCallback == nil {
		return utils.ObjToString(rankResult)
	}
	goRes := ""
	for _, op := range opsInt {
		goStr := execCallback(op)
		goRes += goStr + ","
	}

	// str pattern: code1-result1,code2-result2
	goRes = strings.TrimSuffix(goRes, ",")
	hwlog.RunLog.Infof(" stress test result: %v", goRes)
	pairs := strings.Split(goRes, ",")
	if len(pairs) != len(opsInt) {
		return utils.ObjToString(rankResult)
	}
	for i, op := range opsInt {
		pa := strings.Split(pairs[i], "-")
		if len(pa) != fieldCount {
			return utils.ObjToString(rankResult)
		}
		rankResult.RankResult[strconv.Itoa(op)] = &pb.StressTestOpResult{
			Code:   pa[0],
			Result: pa[1],
		}
	}
	return utils.ObjToString(rankResult)
}

func execCallback(op int) string {
	resultChan := make(chan string, 1)
	timeout, _ := opTimeoutMap[op]
	hwlog.RunLog.Infof("callback func exec max duration: %v", timeout)
	go func() {
		cResult := C.stress_test_callback_wrap(stressTestCallback, C.int(op))
		defer C.free(unsafe.Pointer(cResult))
		goStr := C.GoString(cResult)
		hwlog.RunLog.Infof("callback func exec success result: %v", goStr)
		resultChan <- goStr
	}()
	select {
	case res := <-resultChan:
		return res
	case <-time.After(timeout * time.Second):
		hwlog.RunLog.Errorf("callback timeout for op %d", op)
		return fmt.Sprintf("%s-exec timeout", constant.StressTestTimeout)
	}
}
