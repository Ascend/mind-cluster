/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package server

import (
	"context"
	"errors"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"Ascend-device-plugin/pkg/enpuconfig"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
	"ascend-common/enpumanager"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

type reconcileClientStub struct {
	enpuconfig.ResourceClient
	allocations []enpumanager.Allocation
	listErr     error
	listCalls   int
	released    []string
}

func (c *reconcileClientStub) ListAllocations(context.Context) ([]enpumanager.Allocation, error) {
	c.listCalls++
	return c.allocations, c.listErr
}

func (c *reconcileClientStub) Release(_ context.Context, id string) error {
	c.released = append(c.released, id)
	return nil
}

func reconcileAllocation(id, uid string) enpumanager.Allocation {
	allocation := enpumanager.Allocation{AllocationID: id}
	allocation.Metadata.PodUID = uid
	return allocation
}

type reconcileInformer struct {
	cache.SharedIndexInformer
	synced bool
}

func (i reconcileInformer) HasSynced() bool { return i.synced }

func TestReconcileManagerAllocations(t *testing.T) {
	running := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "infer", Namespace: "ns", UID: "new-uid"}}
	terminating := *running.DeepCopy()
	terminating.UID = "terminating-uid"
	now := metav1.Now()
	terminating.DeletionTimestamp = &now
	t.Cleanup(mockGetActivePodListCache([]v1.Pod{running, terminating}).Reset)
	client := &reconcileClientStub{allocations: []enpumanager.Allocation{
		reconcileAllocation("replaced", "old-uid"),
		reconcileAllocation("active", "new-uid"),
		reconcileAllocation("terminating", "terminating-uid"),
		reconcileAllocation("terminal", "failed-uid"),
		reconcileAllocation("unlinked", ""),
	}}
	m := &softShareManager{client: client}
	m.seen.Store(true)
	kubeClient := &kubeclient.ClientK8s{PodInformer: reconcileInformer{synced: true}}
	if err := m.reconcileAllocations(t.Context(), kubeClient); err != nil {
		t.Fatal(err)
	}
	want := []string{"replaced", "terminal", "unlinked"}
	if !reflect.DeepEqual(client.released, want) {
		t.Fatalf("released = %v, want %v", client.released, want)
	}
}

func TestReconcileRequiresReadyCache(t *testing.T) {
	for _, tc := range []struct {
		name    string
		listErr error
		synced  bool
	}{
		{name: "manager query fails", listErr: errors.New("list failed"), synced: true},
		{name: "cache not synchronized"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &reconcileClientStub{allocations: []enpumanager.Allocation{reconcileAllocation("old", "old-uid")}, listErr: tc.listErr}
			m := &softShareManager{client: client}
			m.seen.Store(true)
			kubeClient := &kubeclient.ClientK8s{PodInformer: reconcileInformer{synced: tc.synced}}
			if err := m.reconcileAllocations(t.Context(), kubeClient); err == nil {
				t.Fatal("expected reconciliation error")
			}
			if len(client.released) != 0 {
				t.Fatalf("unexpected releases: %v", client.released)
			}
		})
	}
}

func TestReconcileEmptyActiveCache(t *testing.T) {
	t.Cleanup(mockGetActivePodListCache(nil).Reset)
	client := &reconcileClientStub{allocations: []enpumanager.Allocation{reconcileAllocation("old", "old-uid")}}
	m := &softShareManager{client: client}
	m.seen.Store(true)
	if err := m.reconcileAllocations(t.Context(), &kubeclient.ClientK8s{PodInformer: reconcileInformer{synced: true}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(client.released, []string{"old"}) {
		t.Fatalf("unexpected releases: %v", client.released)
	}
}
