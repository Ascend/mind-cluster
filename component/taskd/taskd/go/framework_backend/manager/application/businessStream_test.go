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

// Package application implements the taskd manager application
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

func init() {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
}

// MockMsgHandler implements the MsgHandlerInterface
type MockMsgHandler struct {
	SendMsgUseGrpcCalls []SendMsgUseGrpcCall
	SendMsgToMgrCalls   []SendMsgToMgrCall
	GetDataPoolFunc     func() *storage.DataPool
}

type SendMsgUseGrpcCall struct {
	MsgType string
	MsgBody string
	Dst     *common.Position
}
type SendMsgToMgrCall struct {
	Uuid    string
	BizType string
	Src     *common.Position
	MsgBody storage.MsgBody
}

func (m *MockMsgHandler) SendMsgUseGrpc(msgType string, msgBody string, dst *common.Position) {
	m.SendMsgUseGrpcCalls = append(m.SendMsgUseGrpcCalls, SendMsgUseGrpcCall{
		MsgType: msgType,
		MsgBody: msgBody,
		Dst:     dst,
	})
}

func (m *MockMsgHandler) SendMsgToMgr(uuid string, bizType string, src *common.Position, msgBody storage.MsgBody) {
	m.SendMsgToMgrCalls = append(m.SendMsgToMgrCalls, SendMsgToMgrCall{
		Uuid:    uuid,
		BizType: bizType,
		Src:     src,
		MsgBody: msgBody,
	})
}

func (m *MockMsgHandler) GetDataPool() *storage.DataPool {
	if m.GetDataPoolFunc != nil {
		return m.GetDataPoolFunc()
	}
	return nil
}

// MockPluginHandler implements the PluginHandlerInterface
type MockPluginHandler struct {
	InitFunc      func() error
	GetPluginFunc func(pluginName string) (infrastructure.ManagerPlugin, error)
	RegisterFunc  func(pluginName string, plugin infrastructure.ManagerPlugin) error
	HandleFunc    func(pluginName string) (infrastructure.HandleResult, error)
	PredicateFunc func(snapshot *storage.SnapShot) []infrastructure.PredicateResult
	PullMsgFunc   func(pluginName string) ([]infrastructure.Msg, error)

	PredicateResults []infrastructure.PredicateResult
	HandleResult     infrastructure.HandleResult
	PullMsgResult    []infrastructure.Msg
	RegisterCalls    []string
}

func (m *MockPluginHandler) Init() error {
	if m.InitFunc != nil {
		return m.InitFunc()
	}
	return nil
}

func (m *MockPluginHandler) GetPlugin(pluginName string) (infrastructure.ManagerPlugin, error) {
	if m.GetPluginFunc != nil {
		return m.GetPluginFunc(pluginName)
	}
	return nil, nil
}

func (m *MockPluginHandler) Register(pluginName string, plugin infrastructure.ManagerPlugin) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(pluginName, plugin)
	}
	m.RegisterCalls = append(m.RegisterCalls, pluginName)
	return nil
}

func (m *MockPluginHandler) Handle(pluginName string) (infrastructure.HandleResult, error) {
	if m.HandleFunc != nil {
		return m.HandleFunc(pluginName)
	}
	return m.HandleResult, nil
}

func (m *MockPluginHandler) Predicate(snapshot *storage.SnapShot) []infrastructure.PredicateResult {
	if m.PredicateFunc != nil {
		return m.PredicateFunc(snapshot)
	}
	return m.PredicateResults
}

func (m *MockPluginHandler) PullMsg(pluginName string) ([]infrastructure.Msg, error) {
	if m.PullMsgFunc != nil {
		return m.PullMsgFunc(pluginName)
	}
	return m.PullMsgResult, nil
}

// MockStreamHandler implements the StreamHandlerInterface
type MockStreamHandler struct {
	Streams           map[string]*infrastructure.Stream
	InitFunc          func() error
	SetStreamFunc     func(stream *infrastructure.Stream) error
	GetStreamFunc     func(streamName string) *infrastructure.Stream
	GetStreamsFunc    func() map[string]*infrastructure.Stream
	AllocateTokenFunc func(streamName, plugin string) error
	ReleaseTokenFunc  func(streamName, pluginName string) error
	ResetTokenFunc    func(streamName string) error
	PrioritizeFunc    func(streamName string, requestList []string) ([]string, error)
	IsStreamWorkFunc  func(streamName string) (bool, error)

	AllocateTokenCalls []AllocateTokenCall
	ReleaseTokenCalls  []ReleaseTokenCall
	ResetTokenCalls    []string
	PrioritizeCalls    []PrioritizeCall
	IsStreamWorkCalls  []string
}

type AllocateTokenCall struct {
	StreamName string
	Plugin     string
}

type ReleaseTokenCall struct {
	StreamName string
	PluginName string
}

type PrioritizeCall struct {
	StreamName  string
	RequestList []string
}

func (m *MockStreamHandler) Init() error {
	if m.InitFunc != nil {
		return m.InitFunc()
	}
	return nil
}

func (m *MockStreamHandler) SetStream(stream *infrastructure.Stream) error {
	if m.SetStreamFunc != nil {
		return m.SetStreamFunc(stream)
	}
	m.Streams[stream.GetName()] = stream
	return nil
}

func (m *MockStreamHandler) GetStream(streamName string) *infrastructure.Stream {
	if m.GetStreamFunc != nil {
		return m.GetStreamFunc(streamName)
	}
	return m.Streams[streamName]
}

func (m *MockStreamHandler) GetStreams() map[string]*infrastructure.Stream {
	if m.GetStreamsFunc != nil {
		return m.GetStreamsFunc()
	}
	return m.Streams
}

func (m *MockStreamHandler) AllocateToken(streamName, plugin string) error {
	if m.AllocateTokenFunc != nil {
		return m.AllocateTokenFunc(streamName, plugin)
	}
	m.AllocateTokenCalls = append(m.AllocateTokenCalls, AllocateTokenCall{
		StreamName: streamName,
		Plugin:     plugin,
	})

	stream := m.GetStream(streamName)
	if stream == nil {
		return fmt.Errorf("stream %s not found", streamName)
	}
	return stream.Bind(plugin)
}

func (m *MockStreamHandler) ReleaseToken(streamName, pluginName string) error {
	if m.ReleaseTokenFunc != nil {
		return m.ReleaseTokenFunc(streamName, pluginName)
	}
	m.ReleaseTokenCalls = append(m.ReleaseTokenCalls, ReleaseTokenCall{
		StreamName: streamName,
		PluginName: pluginName,
	})

	stream := m.GetStream(streamName)
	if stream == nil {
		return fmt.Errorf("stream %s not found", streamName)
	}
	return stream.Release(pluginName)
}

func (m *MockStreamHandler) ResetToken(streamName string) error {
	if m.ResetTokenFunc != nil {
		return m.ResetTokenFunc(streamName)
	}
	m.ResetTokenCalls = append(m.ResetTokenCalls, streamName)

	stream := m.GetStream(streamName)
	if stream == nil {
		return fmt.Errorf("stream %s not found", streamName)
	}
	return stream.Reset()
}

func (m *MockStreamHandler) Prioritize(streamName string, requestList []string) ([]string, error) {
	if m.PrioritizeFunc != nil {
		return m.PrioritizeFunc(streamName, requestList)
	}
	m.PrioritizeCalls = append(m.PrioritizeCalls, PrioritizeCall{
		StreamName:  streamName,
		RequestList: requestList,
	})

	stream := m.GetStream(streamName)
	if stream == nil {
		return nil, fmt.Errorf("stream %s not found", streamName)
	}

	// Return the original list by default
	return requestList, nil
}

func (m *MockStreamHandler) IsStreamWork(streamName string) (bool, error) {
	if m.IsStreamWorkFunc != nil {
		return m.IsStreamWorkFunc(streamName)
	}
	m.IsStreamWorkCalls = append(m.IsStreamWorkCalls, streamName)

	stream := m.GetStream(streamName)
	if stream == nil {
		return false, fmt.Errorf("stream %s not found", streamName)
	}

	return stream.GetTokenOwner() != "", nil
}

// Test BusinessStreamProcessor struct methods
func TestBusinessStreamProcessor_New(t *testing.T) {
	msgHandler := &MockMsgHandler{}
	bsp := NewBusinessStreamProcessor(msgHandler)

	if bsp == nil {
		t.Fatal("NewBusinessStreamProcessor returned nil")
	}

	if bsp.MsgHandler != msgHandler {
		t.Error("MsgHandler not set correctly")
	}

	if bsp.PluginHandler == nil {
		t.Error("PluginHandler not initialized")
	}

	if bsp.StreamHandler == nil {
		t.Error("StreamHandler not initialized")
	}
}

func TestBusinessStreamProcessor_Init(t *testing.T) {
	msgHandler := &MockMsgHandler{}
	pluginHandler := &MockPluginHandler{}
	streamHandler := &MockStreamHandler{
		Streams: make(map[string]*infrastructure.Stream),
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler
	bsp.StreamHandler = streamHandler

	err := bsp.Init()
	if err != nil {
		t.Errorf("Init() returned error: %v", err)
	}
}

func TestBusinessStreamProcessor_Init_Error(t *testing.T) {
	tests := []struct {
		name                string
		streamInitErr       error
		pluginInitErr       error
		expectedError       bool
		expectedErrorString string
	}{
		{
			name:                "stream_init_error",
			streamInitErr:       errors.New("init stream handler failed"),
			expectedError:       true,
			expectedErrorString: "init stream handler failed",
		},
		{
			name:                "plugin_init_error",
			pluginInitErr:       errors.New("init plugin handler failed"),
			expectedError:       true,
			expectedErrorString: "init plugin handler failed",
		},
		{
			name:                "both_errors",
			streamInitErr:       errors.New("init stream handler failed"),
			pluginInitErr:       errors.New("init plugin handler failed"),
			expectedError:       true,
			expectedErrorString: "init stream handler failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgHandler := &MockMsgHandler{}
			pluginHandler := &MockPluginHandler{
				InitFunc: func() error { return tt.pluginInitErr },
			}
			streamHandler := &MockStreamHandler{
				InitFunc: func() error { return tt.streamInitErr },
			}

			bsp := NewBusinessStreamProcessor(msgHandler)
			bsp.PluginHandler = pluginHandler
			bsp.StreamHandler = streamHandler

			err := bsp.Init()
			if (err != nil) != tt.expectedError {
				t.Errorf("Init() error = %v, expectedError = %v", err, tt.expectedError)
			}

			if err != nil && !strings.Contains(err.Error(), tt.expectedErrorString) {
				t.Errorf("Init() error = %v, expected substring %q", err, tt.expectedErrorString)
			}
		})
	}
}

func TestBusinessStreamProcessor_AllocateToken(t *testing.T) {
	streamName := "testStream"
	pluginName := "testPlugin"

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{pluginName: 1})
	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
	}

	pluginHandler := &MockPluginHandler{
		PredicateResults: []infrastructure.PredicateResult{
			{
				PluginName:      pluginName,
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{streamName: ""},
			},
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler
	bsp.StreamHandler = streamHandler

	snapshot := &storage.SnapShot{}
	bsp.AllocateToken(snapshot)

	// Verify token allocation
	if stream.GetTokenOwner() != pluginName {
		t.Errorf("Token owner is %q, expected %q", stream.GetTokenOwner(), pluginName)
	}

	// Verify AllocateToken was called
	if len(streamHandler.AllocateTokenCalls) != 1 {
		t.Errorf("AllocateToken called %d times, expected 1", len(streamHandler.AllocateTokenCalls))
	} else if streamHandler.AllocateTokenCalls[0].StreamName != streamName ||
		streamHandler.AllocateTokenCalls[0].Plugin != pluginName {
		t.Errorf("AllocateToken called with wrong parameters")
	}
}

func TestBusinessStreamProcessor_StreamRun(t *testing.T) {
	streamName := constant.ProfilingStream
	pluginName := constant.ProfilingPluginName

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{pluginName: 1})
	_ = stream.Bind(pluginName) // Set token owner

	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
	}

	pluginHandler := &MockPluginHandler{
		HandleResult: infrastructure.HandleResult{
			Stage:    constant.HandleStageFinal,
			ErrorMsg: "",
		},
		PullMsgResult: []infrastructure.Msg{
			{
				Receiver: []string{common.MgrRole},
				Body:     storage.MsgBody{Message: "message"},
			},
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler
	bsp.StreamHandler = streamHandler

	err := bsp.StreamRun()
	if err != nil {
		t.Errorf("StreamRun() returned error: %v", err)
	}

	// Verify message was sent
	if len(msgHandler.SendMsgToMgrCalls) != 1 {
		t.Errorf("SendMsgToMgr called %d times, expected 1", len(msgHandler.SendMsgToMgrCalls))
	}
}

func TestBusinessStreamProcessor_StreamRun_Error(t *testing.T) {
	streamName := constant.ProfilingStream
	pluginName := constant.ProfilingPluginName

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{pluginName: 1})
	_ = stream.Bind(pluginName) // Set token owner

	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
		ReleaseTokenFunc: func(streamName, pluginName string) error {
			return fmt.Errorf("test error")
		},
	}

	pluginHandler := &MockPluginHandler{
		HandleResult: infrastructure.HandleResult{
			Stage: constant.HandleStageFinal,
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler
	bsp.StreamHandler = streamHandler

	err := bsp.StreamRun()
	if err == nil {
		t.Error("StreamRun() should return error")
	} else if !strings.Contains(err.Error(), "test error") {
		t.Errorf("StreamRun() error = %v, expected 'test error'", err)
	}
}

func TestBusinessStreamProcessor_DistributeMsg(t *testing.T) {
	msgHandler := &MockMsgHandler{}
	pluginHandler := &MockPluginHandler{}
	streamHandler := &MockStreamHandler{}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler
	bsp.StreamHandler = streamHandler

	msgs := []infrastructure.Msg{
		{
			Receiver: []string{common.MgrRole},
			Body:     storage.MsgBody{Message: "test"},
		},
		{
			Receiver: []string{common.AgentRole},
			Body:     storage.MsgBody{Message: "status"},
		},
	}

	msgHandler.GetDataPoolFunc = func() *storage.DataPool {
		return &storage.DataPool{
			Snapshot: &storage.SnapShot{
				AgentInfos: &storage.AgentInfos{
					Agents: map[string]*storage.AgentInfo{
						common.AgentRole: {
							Pos: &common.Position{
								Role:        common.AgentRole,
								ServerRank:  "0",
								ProcessRank: "",
							},
							RWMutex: sync.RWMutex{},
						},
					},
					RWMutex: sync.RWMutex{},
				},
			},
			RWMutex: sync.RWMutex{},
		}
	}

	err := bsp.DistributeMsg(msgs)
	if err != nil {
		t.Errorf("DistributeMsg() returned error: %v", err)
	}

	// Verify messages were sent
	if len(msgHandler.SendMsgToMgrCalls) != 1 {
		t.Errorf("SendMsgToMgr called %d times, expected 1", len(msgHandler.SendMsgToMgrCalls))
	}

	if len(msgHandler.SendMsgUseGrpcCalls) != 1 {
		t.Errorf("SendMsgUseGrpc called %d times, expected 1", len(msgHandler.SendMsgUseGrpcCalls))
	}
}

func TestBusinessStreamProcessor_ResetStreamToken(t *testing.T) {
	streamName := constant.ProfilingStream

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{})
	_ = stream.Bind("test_plugin") // Set token owner

	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.StreamHandler = streamHandler

	err := bsp.StreamHandler.ResetToken(streamName)
	if err != nil {
		t.Errorf("ResetStreamToken() returned error: %v", err)
	}

	// Verify token was reset
	if stream.GetTokenOwner() != "" {
		t.Errorf("Token owner is %q, expected empty", stream.GetTokenOwner())
	}

	// Verify ResetToken was called
	if len(streamHandler.ResetTokenCalls) != 1 {
		t.Errorf("ResetToken called %d times, expected 1", len(streamHandler.ResetTokenCalls))
	}
}

func TestBusinessStreamProcessor_IsStreamWorking(t *testing.T) {
	streamName := constant.ProfilingStream

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{})

	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.StreamHandler = streamHandler

	// Test when stream is not working
	working, err := bsp.StreamHandler.IsStreamWork(streamName)
	if err != nil {
		t.Errorf("IsStreamWorking() returned error: %v", err)
	}

	if working {
		t.Error("IsStreamWorking() returned true, expected false")
	}

	// Set stream to working state
	_ = stream.Bind("test_plugin")

	// Test when stream is working
	working, err = bsp.StreamHandler.IsStreamWork(streamName)
	if err != nil {
		t.Errorf("IsStreamWorking() returned error: %v", err)
	}

	if !working {
		t.Error("IsStreamWorking() returned false, expected true")
	}
}

func TestBusinessStreamProcessor_PrioritizeRequests(t *testing.T) {
	streamName := constant.ProfilingStream
	requestList := []string{"plugin1", "plugin2", "plugin3"}

	msgHandler := &MockMsgHandler{}
	stream := infrastructure.NewStream(streamName, map[string]int{})

	// Set custom Prioritize method
	streamHandler := &MockStreamHandler{
		Streams: map[string]*infrastructure.Stream{
			streamName: stream,
		},
		PrioritizeFunc: func(streamName string, requests []string) ([]string, error) {
			// Simple list reversal for demonstration
			result := make([]string, len(requests))
			copy(result, requests)
			for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
				result[i], result[j] = result[j], result[i]
			}
			return result, nil
		},
	}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.StreamHandler = streamHandler

	result, err := bsp.StreamHandler.Prioritize(streamName, requestList)
	if err != nil {
		t.Errorf("PrioritizeRequests() returned error: %v", err)
	}

	// Verify reversed list
	expected := []string{"plugin3", "plugin2", "plugin1"}
	if len(result) != len(expected) {
		t.Errorf("Result length = %d, expected %d", len(result), len(expected))
	} else {
		for i := range result {
			if result[i] != expected[i] {
				t.Errorf("Result[%d] = %q, expected %q", i, result[i], expected[i])
			}
		}
	}
}

func TestBusinessStreamProcessor_RegisterPlugin(t *testing.T) {
	pluginName := "new_plugin"

	msgHandler := &MockMsgHandler{}
	pluginHandler := &MockPluginHandler{}

	bsp := NewBusinessStreamProcessor(msgHandler)
	bsp.PluginHandler = pluginHandler

	// Create a simple plugin implementation
	plugin := &struct {
		infrastructure.ManagerPlugin
	}{}

	err := bsp.PluginHandler.Register(pluginName, plugin)
	if err != nil {
		t.Errorf("RegisterPlugin() returned error: %v", err)
	}

	// Verify plugin was registered
	if len(pluginHandler.RegisterCalls) != 1 || pluginHandler.RegisterCalls[0] != pluginName {
		t.Errorf("Plugin not registered correctly")
	}
}
