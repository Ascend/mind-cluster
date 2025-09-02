// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics main test for om
package om

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

func TestSwitchNicProcessMsg(t *testing.T) {
	patches := gomonkey.NewPatches()
	t.Run("get msg is nil", func(t *testing.T) {
		called := false
		patches.ApplyFunc(doSwitchNic, func(ranks []int, ops []bool) (string, error) {
			called = true
			return "", nil
		})
		defer patches.Reset()
		SwitchNicProcessMsg(nil)
		assert.False(t, called)
	})
	t.Run("get msg is ok", func(t *testing.T) {
		defer patches.Reset()
		called := false
		patches.ApplyFunc(doSwitchNic, func(ranks []int, ops []bool) (string, error) {
			called = true
			return "", nil
		}).ApplyFunc(notifySwitchNicResult, func(result, uid string) {})
		msgBody := &storage.MsgBody{
			Extension: map[string]string{
				constant.SwitchNicUUID: "123",
				constant.GlobalRankKey: utils.ObjToString([]string{"0", "1"}),
				constant.GlobalOpKey:   utils.ObjToString([]bool{true, true}),
			},
		}
		msg := &common.Message{
			Body: utils.ObjToString(msgBody),
		}
		SwitchNicProcessMsg(msg)
		assert.True(t, called)
	})
}
