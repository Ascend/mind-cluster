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

// Package service for taskd manager backend service
package service

import (
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
)

// StreamHandlerInterface define the interface of stream handler
type StreamHandlerInterface interface {
	Init() error
	SetStream(*infrastructure.Stream) error
	GetStream(string) *infrastructure.Stream
	GetStreams() map[string]*infrastructure.Stream
	AllocateToken(string, string) error
	ReleaseToken(string, string) error
	ResetToken(string) error
	Prioritize(string, []string) ([]string, error)
	IsStreamWork(string) (bool, error)
}

// StreamHandler define the stream handler struct
type StreamHandler struct {
	Streams     map[string]*infrastructure.Stream
	StreamsLock *sync.RWMutex
}

// NewStreamHandler return stream handler instance
func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		Streams:     make(map[string]*infrastructure.Stream, 0),
		StreamsLock: &sync.RWMutex{},
	}
}

// Init create business stream
func (s *StreamHandler) Init() error {
	profilingCollectStream := infrastructure.NewStream(
		constant.ProfilingStream, map[string]int{constant.ProfilingPluginName: 1})
	OmSwitchNicStream := infrastructure.NewStream(constant.OMSwitchNicStreamName,
		map[string]int{constant.OMSwitchNicPluginName: 1})
	OmStressTestStream := infrastructure.NewStream(constant.OMStressTestStreamName,
		map[string]int{constant.OMStressTestPluginName: 1})
	streamList := []*infrastructure.Stream{
		profilingCollectStream,
		OmStressTestStream,
		OmSwitchNicStream,
	}
	for _, stream := range streamList {
		if err := s.SetStream(stream); err != nil {
			hwlog.RunLog.Errorf("init stream handler failed: set stream %s failed", stream.GetName())
			return err
		}
	}
	return nil
}

// SetStream set a stream in streams
func (s *StreamHandler) SetStream(stream *infrastructure.Stream) error {
	s.StreamsLock.Lock()
	defer s.StreamsLock.Unlock()
	_, ok := s.Streams[stream.GetName()]
	if ok {
		hwlog.RunLog.Errorf("stream %s set failed: conflict stream name", stream.GetName())
		return fmt.Errorf("stream %s set failed: conflict stream name", stream.GetName())
	}
	s.Streams[stream.GetName()] = stream
	return nil
}

// GetStream return a stream by name
func (s *StreamHandler) GetStream(streamName string) *infrastructure.Stream {
	s.StreamsLock.RLock()
	defer s.StreamsLock.RUnlock()
	stream, ok := s.Streams[streamName]
	if !ok {
		return nil
	}
	return stream
}

// GetStreams return all streams
func (s *StreamHandler) GetStreams() map[string]*infrastructure.Stream {
	s.StreamsLock.RLock()
	defer s.StreamsLock.RUnlock()
	return s.Streams
}

// AllocateToken allocate stream token to plugin
func (s *StreamHandler) AllocateToken(streamName, plugin string) error {
	stream := s.GetStream(streamName)
	if stream == nil {
		hwlog.RunLog.Errorf("stream %s is unregistered", streamName)
		return fmt.Errorf("stream %s is unregistered", streamName)
	}
	if err := stream.Bind(plugin); err != nil {
		hwlog.RunLog.Errorf("stream %s bind plugin failed: %v", streamName, err)
		return fmt.Errorf("stream %s bind plugin failed: %v", streamName, err)
	}
	return nil
}

// ReleaseToken release stream token by plugin name
func (s *StreamHandler) ReleaseToken(streamName, pluginName string) error {
	stream := s.GetStream(streamName)
	if stream == nil {
		hwlog.RunLog.Errorf("stream %s is unregistered", streamName)
		return fmt.Errorf("stream %s is unregistered", streamName)
	}
	return stream.Release(pluginName)
}

// ResetToken reset stream owner map and current owner to reset stream execute
func (s *StreamHandler) ResetToken(streamName string) error {
	stream := s.GetStream(streamName)
	if stream == nil {
		hwlog.RunLog.Errorf("stream %s is unregistered", streamName)
		return fmt.Errorf("stream %s is unregistered", streamName)
	}
	return stream.Reset()
}

// Prioritize sort the requests for stream application
func (s *StreamHandler) Prioritize(streamName string, requestList []string) ([]string, error) {
	var sortedRequestList []string
	stream := s.GetStream(streamName)
	if stream == nil {
		hwlog.RunLog.Errorf("prioritize failed: stream %s is not exist", streamName)
		return sortedRequestList, fmt.Errorf("stream %s is not exist", streamName)
	}
	sortedRequestList = stream.Prioritize(requestList)
	return sortedRequestList, nil
}

// IsStreamWork return stream working status
func (s *StreamHandler) IsStreamWork(streamName string) (bool, error) {
	stream := s.GetStream(streamName)
	if stream == nil {
		hwlog.RunLog.Errorf("get stream failed: stream %s is not exist", streamName)
		return false, fmt.Errorf("stream %s is not exist", streamName)
	}
	if owner := stream.GetTokenOwner(); owner != "" {
		return true, nil
	}
	return false, nil
}
