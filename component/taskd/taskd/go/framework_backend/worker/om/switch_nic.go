// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

/*
#include <stdlib.h>
#include "switch.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

// SwitchNicNetTool worker net tool
var SwitchNicNetTool *net.NetInstance

// switchNicCallback switch callback func
var switchNicCallback C.callbackfunc

// RegisterSwitchNicCallback register switch callback func
func RegisterSwitchNicCallback(ptr uintptr) {
	switchNicCallback = (C.callbackfunc)(unsafe.Pointer(ptr))
}

// SwitchNicProcessMsg process switch nic msg
func SwitchNicProcessMsg(msg *common.Message) {
	if msg == nil {
		hwlog.RunLog.Error("msg is nil")
		return
	}
	body, err := utils.StringToObj[storage.MsgBody](msg.Body)
	if err != nil {
		hwlog.RunLog.Errorf("get msgBody err: %v, msgBody is %v", err, body)
		return
	}
	uid := body.Extension[constant.SwitchNicUUID]
	rankStr := body.Extension[constant.GlobalRankKey]
	opStr := body.Extension[constant.GlobalOpKey]
	if uid == "" || rankStr == "" || opStr == "" {
		hwlog.RunLog.Debugf("failed to get param, uid: %v, rankStr: %v, opStr: %v, skip", uid, rankStr, opStr)
		return
	}
	var ranks []string
	var ops []bool
	err = json.Unmarshal([]byte(rankStr), &ranks)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal, err: %v", err)
		return
	}
	err = json.Unmarshal([]byte(opStr), &ops)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal, err: %v", err)
		return
	}
	ranksInt := make([]int, len(ranks))
	for i, rank := range ranks {
		rankInt, err := strconv.Atoi(rank)
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert rank to int, err: %v", err)
			return
		}
		ranksInt[i] = rankInt
	}
	hwlog.RunLog.Infof("worker recv uuid: %v, ranks: %v, ops: %v", uid, ranksInt, ops)
	result, err := doSwitchNic(ranksInt, ops)
	if err != nil {
		hwlog.RunLog.Errorf("failed to do switch nic, err: %v", err)
	}
	notifySwitchNicResult(result, uid)
}

func notifySwitchNicResult(result, uid string) {
	if SwitchNicNetTool == nil {
		hwlog.RunLog.Error("SwitchNicNetTool for worker is nil")
		return
	}
	msg := storage.MsgBody{
		MsgType: constant.Action,
		Code:    constant.SwitchNicCode,
		Message: result,
		Extension: map[string]string{
			constant.SwitchNicUUID: uid,
		},
	}
	_, err := SwitchNicNetTool.SyncSendMessage(uuid.New().String(), "default", utils.ObjToString(msg), &common.Position{
		Role:       common.MgrRole,
		ServerRank: "0",
	})

	if err != nil {
		hwlog.RunLog.Errorf("send result to mgr err: %v", err)
		return
	}
	hwlog.RunLog.Infof("notify mgr result %v succeeded, msg: %s", result, utils.ObjToString(msg))
}

func doSwitchNic(ranks []int, ops []bool) (string, error) {
	if switchNicCallback == nil {
		return constant.SwitchFail, fmt.Errorf("switchNicCallback is nil")
	}
	if len(ranks) != len(ops) {
		return constant.SwitchFail, fmt.Errorf("ranks and ops length not match")
	}
	cRanks := make([]C.int, len(ranks))
	for i, r := range ranks {
		cRanks[i] = C.int(r)
	}

	cOps := make([]C.bool, len(ops))
	for i, b := range ops {
		cOps[i] = C.bool(b)
	}

	cRanksPtr := (*C.int)(C.malloc(C.size_t(len(cRanks)) * C.sizeof_int))
	defer C.free(unsafe.Pointer(cRanksPtr))
	cOpsPtr := (*C.bool)(C.malloc(C.size_t(len(cOps)) * C.sizeof_bool))
	defer C.free(unsafe.Pointer(cOpsPtr))

	ranksSlice := unsafe.Slice(cRanksPtr, len(cRanks))
	opsSlice := unsafe.Slice(cOpsPtr, len(cOps))
	copy(ranksSlice, cRanks)
	copy(opsSlice, cOps)

	res := C.callbackfuncwrap(switchNicCallback, cRanksPtr, cOpsPtr, C.int(len(cRanks)))
	hwlog.RunLog.Infof("callback func exex success result: %v", res)
	if !bool(res) {
		return constant.SwitchFail, fmt.Errorf("switch nic failed")
	}
	return constant.SwitchOK, nil
}
