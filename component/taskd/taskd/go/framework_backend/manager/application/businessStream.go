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

// Package application for taskd manager application
package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/framework_backend/manager/service"
	"taskd/toolkit_backend/net/common"
)

// BusinessStreamProcessor define the class of business stream processor
type BusinessStreamProcessor struct {
	MsgHandler    *MsgHandler
	PluginHandler *service.PluginHandler
	StreamHandler *service.StreamHandler
}

// NewBusinessStreamProcessor return a business stream handler
func NewBusinessStreamProcessor(msgHandler *MsgHandler) *BusinessStreamProcessor {
	return &BusinessStreamProcessor{
		MsgHandler:    msgHandler,
		PluginHandler: service.NewPluginHandler(),
		StreamHandler: service.NewStreamHandler(),
	}
}

// Init init all handler
func (b *BusinessStreamProcessor) Init() error {
	if err := b.StreamHandler.Init(); err != nil {
		hwlog.RunLog.Errorf("init stream handler failed, err: %v", err)
		return err
	}
	if err := b.PluginHandler.Init(); err != nil {
		hwlog.RunLog.Errorf("init plugin handler failed, err: %v", err)
		return err
	}
	return nil
}

// AllocateToken allocate stream token to plugin
func (b *BusinessStreamProcessor) AllocateToken(snapShot *storage.SnapShot) {
	predicateResult := b.PluginHandler.Predicate(snapShot)
	streamTokenRequest := make(map[string][]string, len(b.StreamHandler.Streams))
	for _, pluginRequest := range predicateResult {
		if pluginRequest.CandidateStatus != constant.CandidateStatus {
			continue
		}
		for streamName := range pluginRequest.PredicateStream {
			if _, ok := streamTokenRequest[streamName]; !ok {
				streamTokenRequest[streamName] = make([]string, 0)
			}
			streamTokenRequest[streamName] = append(streamTokenRequest[streamName], pluginRequest.PluginName)
		}
	}
	for streamName, StreamRequestList := range streamTokenRequest {
		if workStatus, err := b.StreamHandler.IsStreamWork(streamName); workStatus || err != nil {
			continue
		}
		sortedRequestList, err := b.StreamHandler.Prioritize(streamName, StreamRequestList)
		if err != nil {
			hwlog.RunLog.Errorf("sorted stream %s stream request failed", streamName)
			continue
		}
		if len(sortedRequestList) == 0 {
			hwlog.RunLog.Debugf("sort stream %s request list failed: get null list", streamName)
			continue
		}
		if err := b.StreamHandler.AllocateToken(streamName, sortedRequestList[0]); err != nil {
			hwlog.RunLog.Errorf("stream %s allocate token to plugin %s failed, err: %v",
				streamName, StreamRequestList[0], err)
		}
	}
}

// StreamRun run all stream whose token is allocated
func (b *BusinessStreamProcessor) StreamRun() error {
	for _, stream := range b.StreamHandler.GetStreams() {
		pluginName := stream.GetTokenOwner()
		if pluginName == "" {
			continue
		}
		handResult, err := b.PluginHandler.Handle(pluginName)
		if handResult.ErrorMsg != "" || err != nil {
			if err := b.StreamHandler.ReleaseToken(stream.GetName(), pluginName); err != nil {
				hwlog.RunLog.Errorf("release stream %s token failed: %v", stream.GetName(), err)
			}
			continue
		}
		if handResult.Stage == constant.HandleStageFinal {
			if err := b.StreamHandler.ReleaseToken(stream.GetName(), pluginName); err != nil {
				hwlog.RunLog.Errorf("release stream %s token failed on final: %v", stream.GetName(), err)
				return fmt.Errorf("release stream %s token failed on final: %v", stream.GetName(), err)
			}
			if err := b.StreamHandler.ResetToken(stream.GetName()); err != nil {
				hwlog.RunLog.Errorf("reset stream %s token failed: %v", stream.GetName(), err)
				return fmt.Errorf("reset stream %s token failed on final: %v", stream.GetName(), err)
			}
		}
		msg, err := b.PluginHandler.PullMsg(pluginName)
		if err != nil {
			if err := b.StreamHandler.ReleaseToken(stream.GetName(), pluginName); err != nil {
				hwlog.RunLog.Errorf("release stream %s token failed: %v", stream.GetName(), err)
			}
			continue
		}
		hwlog.RunLog.Debugf("plugin %s pull msg: %v", pluginName, msg)
	}
	return nil
}

// DistributeMsg business handler pull msg to others
func (b *BusinessStreamProcessor) DistributeMsg(msgs []infrastructure.Msg) error {
	for _, msg := range msgs {
		if len(msg.Receiver) == 0 {
			continue
		}
		for _, receiver := range msg.Receiver {
			if receiver == common.MgrRole {
				b.DistributedMsgToMgr(msg)
				continue
			}
			sendMsg, err := json.Marshal(msg.Body)
			if err != nil {
				hwlog.RunLog.Errorf("business handler send msg marshal failed, err: %v", err)
				break
			}
			b.DistributedMsgToOthers(receiver, sendMsg)
		}
	}
	return nil
}

// DistributedMsgToMgr distributed message to manager
func (b *BusinessStreamProcessor) DistributedMsgToMgr(msg infrastructure.Msg) {
	b.MsgHandler.SendMsgToMgr(uuid.New().String(), constant.DefaultDomainName,
		&common.Position{
			Role:        common.MgrRole,
			ServerRank:  "0",
			ProcessRank: "-1",
		}, msg.Body)
	hwlog.RunLog.Debugf("business handler send msg %v to mgr", msg.Body)
}

// DistributedMsgToOthers distributed message to others
func (b *BusinessStreamProcessor) DistributedMsgToOthers(receiver string, sendMsg []byte) {
	var dst *common.Position
	var err error
	if strings.Contains(receiver, common.WorkerRole) {
		dst, err = b.MsgHandler.DataPool.GetPos(common.WorkerRole, receiver)
		if err != nil {
			hwlog.RunLog.Errorf("get worker pos failed: err %v", err)
			return
		}
	}
	if strings.Contains(receiver, common.AgentRole) {
		dst, err = b.MsgHandler.DataPool.GetPos(common.AgentRole, receiver)
		if err != nil {
			hwlog.RunLog.Errorf("get Agent pos failed: err %v", err)
			return
		}
	}
	b.MsgHandler.SendMsgUseGrpc(constant.DefaultDomainName, string(sendMsg), dst)
	hwlog.RunLog.Debugf("business handler send msg %s to others", string(sendMsg))
}
