/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package service is to provide other service tools, i.e. clusterd
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"ascend-common/common-utils/hwlog"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

func newDataPool() *storage.DataPool {
	return &storage.DataPool{
		Snapshot: &storage.SnapShot{
			AgentInfos: &storage.AgentInfos{
				Agents:    make(map[string]*storage.AgentInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			WorkerInfos: &storage.WorkerInfos{
				Workers:   make(map[string]*storage.WorkerInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			ClusterInfos: &storage.ClusterInfos{
				Clusters:  make(map[string]*storage.ClusterInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
		},
		RWMutex: sync.RWMutex{},
	}
}

func createBaseMessage(src *common.Position, msgType string, code int32, message string) storage.BaseMessage {
	return storage.BaseMessage{
		Header: storage.MsgHeader{Src: src},
		Body: storage.MsgBody{
			MsgType: msgType,
			Code:    code,
			Message: message,
		},
	}
}

func capturePanic(f func()) error {
	var err error
	defer func() {
		err = nil
		if recovered := recover(); recovered != nil {
			err = errors.New("panic error")
		}
	}()
	f()
	return err
}

// TestMain test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return err
	}
	return nil
}
