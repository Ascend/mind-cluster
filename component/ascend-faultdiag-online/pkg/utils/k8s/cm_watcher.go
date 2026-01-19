/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package k8s is a tool to watch the config map
package k8s

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils"
)

var (
	// CmWatcher is a instance of cmWatcher
	CmWatcher     *cmWatcher
	cmWatcherCnce sync.Once
	storage       = utils.NewStorage[*cmData]()
)

type callbackFunc func(oldObj, newObj *corev1.ConfigMap, op watch.EventType)

type cmData struct {
	oldCm *corev1.ConfigMap
	newCm *corev1.ConfigMap
	op    watch.EventType
}

type callback struct {
	registerId string
	f          callbackFunc
}

type cmWatcher struct {
	mu          sync.Mutex
	watchers    map[string]context.CancelFunc
	callbackMap map[string][]callback
}

// GetCmWatcher return a singleton of cmWatcher
func GetCmWatcher() *cmWatcher {
	cmWatcherCnce.Do(func() {
		CmWatcher = &cmWatcher{
			watchers:    make(map[string]context.CancelFunc),
			callbackMap: make(map[string][]callback),
		}
	})
	return CmWatcher
}

func (c *cmWatcher) keyGenerator(namespace, cmName string) string {
	return fmt.Sprintf("%v/%v", namespace, cmName)
}

func (c *cmWatcher) runInformer(ctx context.Context, namespace, cmName string) {
	client, err := GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL]get k8s client failed: %v", err)
		return
	}
	factory := informers.NewFilteredSharedInformerFactory(
		client.ClientSet,
		0,
		namespace,
		func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", cmName).String()
		},
	)

	informer := factory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			go c.cmProcessor(nil, obj, watch.Added)
		},
		UpdateFunc: func(oldObj, newObj any) {
			go c.cmProcessor(oldObj, newObj, watch.Modified)
		},
		DeleteFunc: func(obj any) {
			go c.cmProcessor(nil, obj, watch.Deleted)
		},
	})

	stopChan := make(chan struct{})
	key := c.keyGenerator(namespace, cmName)
	go func() {
		<-ctx.Done()
		close(stopChan)
		hwlog.RunLog.Infof("[FD-OL]config map for %v stopped", key)
	}()
	hwlog.RunLog.Infof("[FD-OL]start to watch config map: %v", key)
	informer.Run(stopChan)
}

func (c *cmWatcher) cmProcessor(oldData, newData any, op watch.EventType) {
	// convert data to configMap object
	var oldCm, newCm *corev1.ConfigMap
	var ok bool
	if oldData != nil {
		oldCm, ok = oldData.(*corev1.ConfigMap)
		if !ok {
			hwlog.RunLog.Errorf("[FD-OL]could not convert data: %v to config map object", oldData)
			return
		}
	}
	newCm, ok = newData.(*corev1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL]could not convert data: %v to config map object", newData)
		return
	}
	key := c.keyGenerator(newCm.Namespace, newCm.Name)
	storage.Store(key, &cmData{oldCm: oldCm, newCm: newCm, op: op})
	c.mu.Lock()
	for _, cb := range c.callbackMap[key] {
		go cb.f(oldCm, newCm, op)
	}
	c.mu.Unlock()
}

// Subscribe subscribe the config map by namespace and cm name, call f if data available
func (c *cmWatcher) Subscribe(namespace, cmName string, f callbackFunc) string {
	key := c.keyGenerator(namespace, cmName)
	data, ok := storage.Load(key)
	if ok {
		go f(data.oldCm, data.newCm, data.op)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	registerId := uuid.New().String()
	if _, exists := c.watchers[key]; exists {
		callbacks := c.callbackMap[key]
		callbacks = append(callbacks, callback{registerId: registerId, f: f})
		c.callbackMap[key] = callbacks
		return registerId
	}
	c.callbackMap[key] = []callback{{registerId: registerId, f: f}}
	ctx, cancle := context.WithCancel(context.Background())
	c.watchers[key] = cancle
	go c.runInformer(ctx, namespace, cmName)
	return registerId
}

// Unsubscribe unsubscribe the config map
func (c *cmWatcher) Unsubscribe(namespace, cmName, registerId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var key = c.keyGenerator(namespace, cmName)
	callbacks := c.callbackMap[key]
	for i := 0; i < len(callbacks); i++ {
		if callbacks[i].registerId == registerId {
			callbacks[i] = callbacks[len(callbacks)-1]
			callbacks = callbacks[:len(callbacks)-1]
			c.callbackMap[key] = callbacks
			break
		}
	}
	if len(callbacks) == 0 {
		if cancel, exists := c.watchers[key]; exists {
			cancel()
		}
		delete(c.watchers, key)
		delete(c.callbackMap, key)
		storage.Delete(key)
	}
}
