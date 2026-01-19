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

// Package infrastructure for taskd manager backend infrastructure
package infrastructure

import (
	"fmt"
	"sort"
)

// StreamInterface define the interface of Stream
type StreamInterface interface {
	Bind(string) error
	Release() error
	Reset() error
}

// Stream define the struct of Stream
type Stream struct {
	// Name indicate the name of stream
	Name string
	// OwnerMap indicate the plugins that have obtained tokens from this stream
	OwnerMap map[string]string
	// TokenOwner indicate the current token owner of this stream
	TokenOwner string
	// PluginPriority indicate the priority of different plugins obtaining this stream
	PluginPriority map[string]int
}

// NewStream return a stream
func NewStream(name string, priorityConf map[string]int) *Stream {
	return &Stream{
		Name:           name,
		OwnerMap:       make(map[string]string, 0),
		TokenOwner:     "",
		PluginPriority: priorityConf,
	}
}

// GetName return the name of this stream
func (s *Stream) GetName() string {
	return s.Name
}

// Bind function of binding the token of this stream to the plugin
func (s *Stream) Bind(plugin string) error {
	if s.TokenOwner != "" {
		return fmt.Errorf("tokenOwer is occupied")
	}
	s.TokenOwner = plugin
	s.OwnerMap[s.TokenOwner] = s.TokenOwner
	return nil
}

// Release function of releasing the plugin that currently holds the token for this stream
func (s *Stream) Release(ownerName string) error {
	if s.TokenOwner == "" {
		return fmt.Errorf("stream %s is free,release failed", s.Name)
	}
	if s.TokenOwner != ownerName {
		return fmt.Errorf("stream %s is not belong to owner %s,release failed", s.Name, ownerName)
	}
	s.TokenOwner = ""
	return nil
}

// Reset function of resetting token allocation for this stream
func (s *Stream) Reset() error {
	s.OwnerMap = make(map[string]string, 0)
	s.TokenOwner = ""
	return nil
}

// GetTokenOwner return the token owner in this stream
func (s *Stream) GetTokenOwner() string {
	return s.TokenOwner
}

// Prioritize sort the plugins applying for tokens in this stream based on stream priority
func (s *Stream) Prioritize(requestList []string) []string {
	filterRequest := s.filterAllocated(requestList)
	sort.Slice(filterRequest, func(i, j int) bool {
		s1, s2 := filterRequest[i], filterRequest[j]
		p1, hasP1 := s.PluginPriority[s1]
		p2, hasP2 := s.PluginPriority[s2]
		switch {
		case hasP1 && hasP2:
			// both have priority: sort by priority
			return p1 < p2
		case hasP1:
			// only s1 has priority: s1 ranks first
			return true
		case hasP2:
			// only s2 has priority: s2 ranks last
			return false
		default:
			// none of them have priority: sorted by dictionary order
			return s1 < s2
		}
	})
	return filterRequest
}

// filterAllocated ilter all plugins that have obtained or obtained tokens for this stream
func (s *Stream) filterAllocated(requestList []string) []string {
	// create filter mapping for quick search
	filterMap := make(map[string]bool, len(s.OwnerMap))
	for n := range s.OwnerMap {
		filterMap[n] = true
	}

	// filter out unnecessary values
	filtered := make([]string, 0, len(requestList))
	for _, r := range requestList {
		if !filterMap[r] {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
