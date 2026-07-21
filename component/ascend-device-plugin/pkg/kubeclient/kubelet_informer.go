/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language or permissions and
   limitations under the License.
*/

// Package kubeclient a series of k8s function
package kubeclient

import (
	"context"
	"reflect"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
)

// kubeletPodResyncInterval is the polling interval for the kubelet-based pod
// informer. The kubelet /pods endpoint does not support watch, so we poll and
// diff against the local cache to synthesize add/update/delete events.
const kubeletPodResyncInterval = 5 * time.Second

// kubeletPodInformer implements cache.SharedIndexInformer by polling the
// kubelet /pods endpoint. It exists so that kubelet mode (where the apiserver
// pod informer is unavailable) can still provide a non-nil PodInformer and
// drive the same pod event handlers (handlePodAddEvent, soft-share delete,
// etc.) that apiserver mode uses.
type kubeletPodInformer struct {
	client *ClientK8s

	mu       sync.RWMutex
	store    map[string]*v1.Pod // key: namespace/name
	handlers []cache.ResourceEventHandler
	synced   bool
	stopped  bool

	cancel context.CancelFunc
	doneCh chan struct{}
}

// newKubeletPodInformer creates a kubelet-backed pod informer. It does not
// start polling until Run is called.
func newKubeletPodInformer(client *ClientK8s) *kubeletPodInformer {
	return &kubeletPodInformer{
		client: client,
		store:  make(map[string]*v1.Pod),
		doneCh: make(chan struct{}),
	}
}

// AddEventHandler registers a handler that will receive synthesized pod
// events. Handlers may be added before Run; events are only delivered after
// Run starts. It implements cache.SharedInformer.
func (k *kubeletPodInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.handlers = append(k.handlers, handler)
	return &kubeletReg{informer: k}, nil
}

// AddEventHandlerWithResyncPeriod registers a handler with a resync period.
// The kubelet informer ignores resync requests because it already polls at a
// fixed interval; every poll acts as a partial resync via diff.
func (k *kubeletPodInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return k.AddEventHandler(handler)
}

// RemoveEventHandler is not supported; returns nil for compatibility.
func (k *kubeletPodInformer) RemoveEventHandler(_ cache.ResourceEventHandlerRegistration) error {
	return nil
}

// GetStore returns the informer's local cache as a cache.Store.
func (k *kubeletPodInformer) GetStore() cache.Store {
	return &kubeletPodStore{informer: k}
}

// GetController is deprecated; returns a no-op controller.
func (k *kubeletPodInformer) GetController() cache.Controller {
	return &kubeletController{}
}

// Run starts the poll-and-diff loop. It blocks until stopCh is closed.
func (k *kubeletPodInformer) Run(stopCh <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	k.mu.Lock()
	k.cancel = cancel
	k.mu.Unlock()

	// Do an immediate full sync so HasSynced becomes true quickly.
	k.pollOnce(true)

	ticker := time.NewTicker(kubeletPodResyncInterval)
	defer ticker.Stop()
	defer close(k.doneCh)
	defer cancel()

	for {
		select {
		case <-stopCh:
			k.mu.Lock()
			k.stopped = true
			k.mu.Unlock()
			return
		case <-ctx.Done():
			k.mu.Lock()
			k.stopped = true
			k.mu.Unlock()
			return
		case <-ticker.C:
			k.pollOnce(false)
		}
	}
}

// HasSynced returns true after the first successful poll has populated the
// local cache.
func (k *kubeletPodInformer) HasSynced() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.synced
}

// LastSyncResourceVersion returns an empty string; the kubelet /pods endpoint
// does not expose a watchable resource version.
func (k *kubeletPodInformer) LastSyncResourceVersion() string {
	return ""
}

// SetWatchErrorHandler is a no-op; polling errors are logged directly.
func (k *kubeletPodInformer) SetWatchErrorHandler(_ cache.WatchErrorHandler) error {
	return nil
}

// SetTransform is a no-op.
func (k *kubeletPodInformer) SetTransform(_ cache.TransformFunc) error {
	return nil
}

// IsStopped reports whether the informer has been stopped.
func (k *kubeletPodInformer) IsStopped() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.stopped
}

// AddIndexers is not supported by the kubelet informer.
func (k *kubeletPodInformer) AddIndexers(_ cache.Indexers) error {
	return nil
}

// GetIndexer returns the local cache as a cache.Indexer.
func (k *kubeletPodInformer) GetIndexer() cache.Indexer {
	return &kubeletPodStore{informer: k}
}

// pollOnce fetches the current pod list from the kubelet /pods endpoint and
// diffs it against the local cache, dispatching OnAdd/OnUpdate/OnDelete to all
// registered handlers. When isInitialList is true, all pods are delivered as
// OnAdd with isInInitialList=true.
func (k *kubeletPodInformer) pollOnce(isInitialList bool) {
	podList, err := k.client.getPodsByKltPort()
	if err != nil {
		hwlog.RunLog.Errorf("kubelet pod informer poll failed: %v", err)
		return
	}

	latest := make(map[string]*v1.Pod, len(podList.Items))
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Spec.NodeName != "" && pod.Spec.NodeName != k.client.NodeName {
			continue
		}
		key := pod.Namespace + "/" + pod.Name
		latest[key] = pod
	}

	// Compute the diff and update the store under the write lock, collecting
	// the events to dispatch. Dispatching must happen *outside* the lock
	// because handlers (e.g. updatePodList -> GetByKey) re-enter the informer
	// via RLock; holding the write lock while calling RLock in the same
	// goroutine would deadlock.
	type event struct {
		kind int // 0=add, 1=update, 2=delete
		new  *v1.Pod
		old  *v1.Pod
	}
	var events []event

	k.mu.Lock()
	handlers := k.handlers
	for key, newPod := range latest {
		oldPod, exists := k.store[key]
		if !exists {
			k.store[key] = newPod
			events = append(events, event{kind: 0, new: newPod})
			continue
		}
		if !podEqual(oldPod, newPod) {
			k.store[key] = newPod
			events = append(events, event{kind: 1, old: oldPod, new: newPod})
		}
	}
	for key, oldPod := range k.store {
		if _, exists := latest[key]; !exists {
			delete(k.store, key)
			events = append(events, event{kind: 2, old: oldPod})
		}
	}
	if !k.synced {
		k.synced = true
	}
	k.mu.Unlock()

	// Dispatch events to all handlers outside the lock.
	for _, e := range events {
		for _, h := range handlers {
			switch e.kind {
			case 0:
				h.OnAdd(e.new, isInitialList)
			case 1:
				h.OnUpdate(e.old, e.new)
			case 2:
				h.OnDelete(e.old)
			}
		}
	}
}

// podEqual compares two pods by ResourceVersion first (cheap), falling back to
// deep equality. The kubelet /pods response includes ResourceVersion per pod,
// so this avoids spurious update events on unchanged pods.
func podEqual(a, b *v1.Pod) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.UID != b.UID {
		return false
	}
	if a.ResourceVersion != "" && a.ResourceVersion == b.ResourceVersion {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// kubeletPodStore implements cache.Indexer backed by the kubelet informer's
// local cache. Only the methods used by consumers (GetByKey, List, ListKeys,
// Get) have real implementations; the rest are no-ops.
type kubeletPodStore struct {
	informer *kubeletPodInformer
}

func (s *kubeletPodStore) Add(_ interface{}) error                 { return nil }
func (s *kubeletPodStore) Update(_ interface{}) error              { return nil }
func (s *kubeletPodStore) Delete(_ interface{}) error              { return nil }
func (s *kubeletPodStore) Replace(_ []interface{}, _ string) error { return nil }
func (s *kubeletPodStore) Resync() error                           { return nil }
func (s *kubeletPodStore) AddIndexers(_ cache.Indexers) error      { return nil }
func (s *kubeletPodStore) GetIndexers() cache.Indexers             { return cache.Indexers{} }

func (s *kubeletPodStore) GetByKey(key string) (interface{}, bool, error) {
	s.informer.mu.RLock()
	defer s.informer.mu.RUnlock()
	pod, exists := s.informer.store[key]
	return pod, exists, nil
}

func (s *kubeletPodStore) Get(obj interface{}) (interface{}, bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return nil, false, err
	}
	return s.GetByKey(key)
}

func (s *kubeletPodStore) List() []interface{} {
	s.informer.mu.RLock()
	defer s.informer.mu.RUnlock()
	list := make([]interface{}, 0, len(s.informer.store))
	for _, pod := range s.informer.store {
		list = append(list, pod)
	}
	return list
}

func (s *kubeletPodStore) ListKeys() []string {
	s.informer.mu.RLock()
	defer s.informer.mu.RUnlock()
	keys := make([]string, 0, len(s.informer.store))
	for key := range s.informer.store {
		keys = append(keys, key)
	}
	return keys
}

// ByIndex returns an empty list; the kubelet informer does not support
// secondary indexes.
func (s *kubeletPodStore) ByIndex(_, _ string) ([]interface{}, error) {
	return nil, nil
}

// Index returns an empty list; the kubelet informer does not support secondary
// indexes.
func (s *kubeletPodStore) Index(_ string, _ interface{}) ([]interface{}, error) {
	return nil, nil
}

// ListIndexFuncValues returns an empty list; no secondary indexes are maintained.
func (s *kubeletPodStore) ListIndexFuncValues(_ string) []string {
	return nil
}

// IndexKeys returns an empty list.
func (s *kubeletPodStore) IndexKeys(_, _ string) ([]string, error) {
	return nil, nil
}

// kubeletReg is a trivial ResourceEventHandlerRegistration. It holds a
// back-reference to the informer so HasSynced reports the live sync state
// rather than the (stale) value captured at registration time.
type kubeletReg struct {
	informer *kubeletPodInformer
}

func (r *kubeletReg) HasSynced() bool { return r.informer.HasSynced() }

// kubeletController is a no-op cache.Controller.
type kubeletController struct{}

func (c *kubeletController) Run(_ <-chan struct{})           {}
func (c *kubeletController) HasSynced() bool                 { return true }
func (c *kubeletController) LastSyncResourceVersion() string { return "" }

// podUIDKey returns namespace/name for a pod, used as the store key.
func podUIDKey(pod *v1.Pod) string {
	return pod.Namespace + "/" + pod.Name
}

// compile-time interface checks
var (
	_ cache.SharedIndexInformer              = (*kubeletPodInformer)(nil)
	_ cache.Indexer                          = (*kubeletPodStore)(nil)
	_ cache.ResourceEventHandlerRegistration = (*kubeletReg)(nil)
	_ cache.Controller                       = (*kubeletController)(nil)
)
