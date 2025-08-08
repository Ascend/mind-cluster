/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package resulthandler is using for handle hccsping mesh result
*/
package resulthandler

import (
	"sync"

	"k8s.io/client-go/util/workqueue"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/pingmeshv1/types"
)

// Interface is interface for handle pingmeshv1 result
type Interface interface {
	Receive(*types.HccspingMeshResult)
	Handle(<-chan struct{})
}

// HandleFunc is function for handle pingmeshv1 result
type HandleFunc func(*types.HccspingMeshResult) error

// NewAggregatedHandler is new aggregated handler for pingmeshv1 result
func NewAggregatedHandler(hs ...HandleFunc) Interface {
	return &aggregatedHandler{
		resultQueue: workqueue.New(),
		hs:          hs,
	}
}

// aggregatedHandler is aggregated handler for pingmeshv1 result
type aggregatedHandler struct {
	resultQueue workqueue.Interface
	hs          []HandleFunc
}

// Receive is receive pingmeshv1 result
func (h *aggregatedHandler) Receive(info *types.HccspingMeshResult) {
	h.resultQueue.Add(info)
}

// Handle is handle pingmeshv1 result
func (h *aggregatedHandler) Handle(stop <-chan struct{}) {
	if stop == nil {
		hwlog.RunLog.Errorf("stop channel is nil")
		return
	}

	for {
		select {
		case <-stop:
			h.resultQueue.ShutDownWithDrain()
			return
		default:
			obj, shutdown := h.resultQueue.Get()
			if shutdown {
				return
			}
			infos, ok := obj.(*types.HccspingMeshResult)
			if !ok {
				hwlog.RunLog.Error("receive invalid pingmeshv1 info")
				h.resultQueue.Done(obj)
				return
			}
			h.handlePingMeshInfo(infos)
			h.resultQueue.Done(obj)
		}
	}
}

func (h *aggregatedHandler) handlePingMeshInfo(info *types.HccspingMeshResult) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.hs))
	for _, hand := range h.hs {
		go func(handler HandleFunc) {
			defer wg.Done()
			err := handler(info)
			if err != nil {
				hwlog.RunLog.Error(err)
			}
		}(hand)
	}
	wg.Wait()
}
