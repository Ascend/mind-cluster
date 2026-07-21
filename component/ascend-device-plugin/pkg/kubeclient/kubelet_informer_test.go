/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// eventCollector records OnAdd/OnUpdate/OnDelete calls for test assertions.
type eventCollector struct {
	mu   sync.Mutex
	adds []*v1.Pod
	upds [][2]*v1.Pod
	dels []*v1.Pod
}

func (e *eventCollector) OnAdd(obj interface{}, _ bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.adds = append(e.adds, obj.(*v1.Pod))
}
func (e *eventCollector) OnUpdate(oldObj, newObj interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.upds = append(e.upds, [2]*v1.Pod{oldObj.(*v1.Pod), newObj.(*v1.Pod)})
}
func (e *eventCollector) OnDelete(obj interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dels = append(e.dels, obj.(*v1.Pod))
}

var errPollFailed = errors.New("simulated kubelet /pods error")

func mkPod(name, ns, rv string) *v1.Pod {
	return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, ResourceVersion: rv}}
}

func TestKubeletPodInformerAddUpdateDelete(t *testing.T) {
	client := &ClientK8s{NodeName: "node"}

	// Phase 1: one pod exists -> OnAdd on initial list
	list1 := &v1.PodList{Items: []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default", ResourceVersion: "1"},
			Spec: v1.PodSpec{NodeName: "node"}},
	}}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) { return list1, nil })
	defer patch.Reset()

	convey.Convey("kubelet informer dispatches add/update/delete via poll diff", t, func() {
		inf := newKubeletPodInformer(client)
		col := &eventCollector{}
		inf.AddEventHandler(col)

		stopCh := make(chan struct{})
		go inf.Run(stopCh)

		// wait for initial sync
		convey.So(waitForSynced(inf, 2*time.Second), convey.ShouldBeTrue)
		convey.So(inf.HasSynced(), convey.ShouldBeTrue)

		// initial list -> 1 add
		col.mu.Lock()
		convey.So(len(col.adds), convey.ShouldEqual, 1)
		convey.So(col.adds[0].Name, convey.ShouldEqual, "p1")
		col.mu.Unlock()

		// GetByKey should return the cached pod
		obj, exists, err := inf.GetIndexer().GetByKey("default/p1")
		convey.So(err, convey.ShouldBeNil)
		convey.So(exists, convey.ShouldBeTrue)
		convey.So(obj.(*v1.Pod).Name, convey.ShouldEqual, "p1")

		// Phase 2: update p1 RV + add p2
		list2 := &v1.PodList{Items: []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default", ResourceVersion: "2"},
				Spec: v1.PodSpec{NodeName: "node"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "default", ResourceVersion: "1"},
				Spec: v1.PodSpec{NodeName: "node"}},
		}}
		patch2 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
			func(_ *ClientK8s) (*v1.PodList, error) { return list2, nil })
		inf.pollOnce(false)

		col.mu.Lock()
		convey.So(len(col.upds), convey.ShouldEqual, 1)
		convey.So(col.upds[0][1].ResourceVersion, convey.ShouldEqual, "2")
		convey.So(len(col.adds), convey.ShouldEqual, 2) // p1 + p2
		convey.So(col.adds[1].Name, convey.ShouldEqual, "p2")
		col.mu.Unlock()

		// Phase 3: remove p1 -> OnDelete
		list3 := &v1.PodList{Items: []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "default", ResourceVersion: "1"},
				Spec: v1.PodSpec{NodeName: "node"}},
		}}
		patch3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
			func(_ *ClientK8s) (*v1.PodList, error) { return list3, nil })
		inf.pollOnce(false)

		col.mu.Lock()
		convey.So(len(col.dels), convey.ShouldEqual, 1)
		convey.So(col.dels[0].Name, convey.ShouldEqual, "p1")
		col.mu.Unlock()

		// p1 should be gone from indexer
		_, exists2, _ := inf.GetIndexer().GetByKey("default/p1")
		convey.So(exists2, convey.ShouldBeFalse)

		close(stopCh)
		patch2.Reset()
		patch3.Reset()
	})
}

func TestKubeletPodInformerFiltersOtherNodes(t *testing.T) {
	client := &ClientK8s{NodeName: "this-node"}
	list := &v1.PodList{Items: []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "mine", Namespace: "default"},
			Spec: v1.PodSpec{NodeName: "this-node"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "theirs", Namespace: "default"},
			Spec: v1.PodSpec{NodeName: "other-node"}},
	}}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) { return list, nil })
	defer patch.Reset()

	convey.Convey("kubelet informer only caches pods on the current node", t, func() {
		inf := newKubeletPodInformer(client)
		stopCh := make(chan struct{})
		go inf.Run(stopCh)
		convey.So(waitForSynced(inf, 2*time.Second), convey.ShouldBeTrue)

		_, mineExists, _ := inf.GetIndexer().GetByKey("default/mine")
		convey.So(mineExists, convey.ShouldBeTrue)
		_, theirsExists, _ := inf.GetIndexer().GetByKey("default/theirs")
		convey.So(theirsExists, convey.ShouldBeFalse)

		close(stopCh)
	})
}

func TestKubeletPodInformerPollErrorKeepsCache(t *testing.T) {
	client := &ClientK8s{NodeName: "node"}
	goodList := &v1.PodList{Items: []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default", ResourceVersion: "1"},
			Spec: v1.PodSpec{NodeName: "node"}},
	}}

	callCount := 0
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) {
			callCount++
			if callCount == 1 {
				return goodList, nil
			}
			return nil, errPollFailed
		})
	defer patch.Reset()

	convey.Convey("poll error does not clear existing cache or mark unsynced", t, func() {
		inf := newKubeletPodInformer(client)
		stopCh := make(chan struct{})
		go inf.Run(stopCh)
		convey.So(waitForSynced(inf, 2*time.Second), convey.ShouldBeTrue)

		// trigger a failing poll
		inf.pollOnce(false)

		// cache should still hold p1
		obj, exists, _ := inf.GetIndexer().GetByKey("default/p1")
		convey.So(exists, convey.ShouldBeTrue)
		convey.So(obj.(*v1.Pod).Name, convey.ShouldEqual, "p1")
		convey.So(inf.HasSynced(), convey.ShouldBeTrue)

		close(stopCh)
	})
}

func TestPodEqual(t *testing.T) {
	convey.Convey("podEqual compares by RV then deep equal", t, func() {
		convey.So(podEqual(mkPod("a", "ns", "1"), mkPod("a", "ns", "1")), convey.ShouldBeTrue)
		convey.So(podEqual(mkPod("a", "ns", "1"), mkPod("a", "ns", "2")), convey.ShouldBeFalse)
		convey.So(podEqual(nil, nil), convey.ShouldBeTrue)
		convey.So(podEqual(nil, mkPod("a", "ns", "1")), convey.ShouldBeFalse)
	})
}

// waitForSynced polls HasSynced until true or timeout.
func waitForSynced(inf *kubeletPodInformer, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if inf.HasSynced() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return inf.HasSynced()
}

// TestKubeletPodInformerHandlerReentersStoreNoDeadlock is a regression test for
// the deadlock that occurred when pollOnce dispatched events while holding the
// write lock. The production handler updatePodList calls
// PodInformer.GetIndexer().GetByKey (which acquires the informer's read lock).
// If dispatch happened under the write lock, the same goroutine would try to
// acquire the read lock -> deadlock. This test simulates that path by
// registering a handler that calls GetByKey on every event and verifying the
// poll completes within a timeout.
func TestKubeletPodInformerHandlerReentersStoreNoDeadlock(t *testing.T) {
	client := &ClientK8s{NodeName: "node"}
	list := &v1.PodList{Items: []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default", ResourceVersion: "1", UID: "uid-1"},
			Spec: v1.PodSpec{NodeName: "node"}},
	}}
	patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(ClientK8s)), "getPodsByKltPort",
		func(_ *ClientK8s) (*v1.PodList, error) { return list, nil })
	defer patch.Reset()

	convey.Convey("handler that re-enters store via GetByKey must not deadlock", t, func() {
		inf := newKubeletPodInformer(client)

		// This handler mirrors the production updatePodList path: it calls
		// GetIndexer().GetByKey, which acquires the informer's RLock.
		// Under the old (buggy) dispatch-under-lock design this deadlocks.
		reentered := &reentrantGetByKeyHandler{informer: inf}
		inf.AddEventHandler(reentered)

		done := make(chan struct{})
		go func() {
			inf.pollOnce(true)
			close(done)
		}()

		select {
		case <-done:
			convey.So(reentered.calls, convey.ShouldEqual, 1)
		case <-time.After(5 * time.Second):
			t.Fatal("pollOnce deadlocked: handler GetByKey blocked on informer lock")
		}
	})
}

// reentrantGetByKeyHandler is a handler that re-enters the informer store on
// OnAdd by calling GetIndexer().GetByKey, mirroring updatePodList's behavior.
type reentrantGetByKeyHandler struct {
	informer *kubeletPodInformer
	calls    int
}

func (r *reentrantGetByKeyHandler) OnAdd(_ interface{}, _ bool) {
	_, _, _ = r.informer.GetIndexer().GetByKey("default/p1")
	r.calls++
}
func (r *reentrantGetByKeyHandler) OnUpdate(_, _ interface{}) {
	r.calls++
}
func (r *reentrantGetByKeyHandler) OnDelete(_ interface{}) {
	r.calls++
}

// ensure cache.ResourceEventHandlerFuncs adapter compiles with the collector
var _ cache.ResourceEventHandler = (*eventCollector)(nil)
