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

// Package slownode a series of function
package slownode

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

var (
	jobFuncList         = []func(*slownode.Job, *slownode.Job, watch.EventType){}
	nodeAlgoResFuncList = []func(
		*slownode.NodeAlgoResult, *slownode.NodeAlgoResult, watch.EventType){}
	nodeDataProfilingResFuncList = []func(
		*slownode.NodeDataProfilingResult, *slownode.NodeDataProfilingResult, watch.EventType){}
	nodeRestartInfoFuncList = []func(*string, *string, watch.EventType){}
	informerCh              = make(chan struct{})

	informerHandlers = []InformerHandlerType{}
)

// InformerHandlerType register event handler for informer
type InformerHandlerType struct {
	filterFunc func(any) bool
	handler    func(any, any, watch.EventType)
}

func registerHandlers(filter func(any) bool, handler func(any, any, watch.EventType)) {
	informerHandlers = append(informerHandlers, InformerHandlerType{filter, handler})
}

// stopInformer stop informer when loss-leader
func stopInformer() {
	if informerCh != nil {
		close(informerCh)
		return
	}
	hwlog.RunLog.Warn("[FD-OL SLOWNODE]stop CM informer: channel is nil will not close it")
}

// cleanFunc clean func when loss-leader
func cleanFunc() {
	jobFuncList = jobFuncList[:0]
	nodeAlgoResFuncList = nodeAlgoResFuncList[:0]
	nodeDataProfilingResFuncList = nodeDataProfilingResFuncList[:0]
	nodeRestartInfoFuncList = nodeRestartInfoFuncList[:0]
}

// addCMHandler Add one or one more func in local funcList
func addCMHandler[T any](
	handlers *[]func(old, new *T, op watch.EventType), newHandlers ...func(old, new *T, op watch.EventType)) {
	*handlers = append(*handlers, newHandlers...)
}

// initCMInformer init configmap informer
func initCMInformer() {
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]created k8sClient failed: %v", err)
		return
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = constants.CmConsumer + "=" + constants.CmConsumerValue
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()

	for _, item := range informerHandlers {
		informerHandler := item
		cmInformer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: informerHandler.filterFunc,
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					go informerHandler.handler(nil, obj, watch.Added)
				},
				UpdateFunc: func(oldObj, newObj any) {
					go informerHandler.handler(oldObj, newObj, watch.Modified)
				},
				DeleteFunc: func(obj any) {
					go informerHandler.handler(nil, obj, watch.Deleted)
				},
			},
		})
	}
	hwlog.RunLog.Info("[FD-OL SLOWNODE]started to watch configMap")
	informerFactory.Start(informerCh)
}

func clusterJobHandler(oldObj, newObj any, operator watch.EventType) {
	genericHandler(oldObj, newObj, operator, constants.ClusterJobCMKey, jobFuncList)
}

func nodeJobHandler(oldObj, newObj any, operator watch.EventType) {
	genericHandler(oldObj, newObj, operator, constants.NodeJobCMKey, jobFuncList)
}

func nodeAlgoHandler(oldObj, newObj any, operator watch.EventType) {
	genericHandler(oldObj, newObj, operator, constants.NodeAlgoResultCMKey, nodeAlgoResFuncList)
}

func nodeDataProfilingHandler(oldObj, newObj any, operator watch.EventType) {
	genericHandler(oldObj, newObj, operator, constants.NodeDataProfilingResultCMKey, nodeDataProfilingResFuncList)
}

func nodeRestartInfoHandler(oldObj, newObj any, operator watch.EventType) {
	genericHandler(oldObj, newObj, operator, constants.NodeRestartInfoCMKey, nodeRestartInfoFuncList)
}

func genericHandler[T slownode.NodeAlgoResult | slownode.NodeDataProfilingResult | slownode.Job | string](
	oldObj, newObj any, operator watch.EventType, cmKey string, handlerFuncs []func(*T, *T, watch.EventType)) {
	var oldObjTyped, newObjTyped *T

	if oldObj != nil {
		oldObjTyped = new(T)
		if err := parseCMResult(oldObj, cmKey, oldObjTyped); err != nil {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]parsed old cm: %v failed: %v", oldObj, err)
			return
		}
	}

	newObjTyped = new(T)
	if err := parseCMResult(newObj, cmKey, newObjTyped); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]parsed new cm: %v failed: %v", newObj, err)
		return
	}
	for _, f := range handlerFuncs {
		f(oldObjTyped, newObjTyped, operator)
	}
}

// filterClusterJob filter func for job in cluster
func filterClusterJob(obj any) bool {
	return isNameMatched(obj, constants.ClusterJobPrefix)
}

// filterNodeJob filter func for job in node
func filterNodeJob(obj any) bool {
	return isNameMatched(obj, constants.NodeJobPrefix)
}

// filterSlowNodeAlgoResult filter func for slow node algo result
func filterAlgoResult(obj any) bool {
	return isNameMatched(obj, constants.NodeAlgoResultPrefix)
}

// filterDataProfilingResult filter func for data profiling result
func filterDataProfilingResult(obj any) bool {
	return isNameMatched(obj, constants.NodeDataProfilingResultPrefix)
}

// filterNodeRestartInfoResult filter func for node restart info
func filterNodeRestartInfoResult(obj any) bool {
	return isNameMatched(obj, constants.NodeRestartInfoPrefix)
}

// isNameMatched check whether its namespace and name match the configmap
func isNameMatched(obj any, namePrefix string) bool {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]cannot convert obj: %v to ConfigMap", obj)
		return false
	}
	return strings.HasPrefix(cm.Name, namePrefix)
}

// parseCMResult is a general func parsing source to result
func parseCMResult[T slownode.NodeAlgoResult | slownode.NodeDataProfilingResult | slownode.Job | string](
	source any, cmKey string, result *T) error {
	cm, ok := source.(*corev1.ConfigMap)
	if !ok {
		return errors.New("convert to ConfigMap object failed")
	}
	data, ok := cm.Data[cmKey]
	if !ok {
		return fmt.Errorf("no cmKey[%s] found", cmKey)
	}
	switch cmKey {
	case constants.NodeRestartInfoCMKey:
		// expect T is string
		resultV := reflect.ValueOf(result)
		if resultV.Kind() != reflect.Ptr {
			return fmt.Errorf("result must be a pointer")
		}
		elem := resultV.Elem()
		if elem.Type() != reflect.TypeOf("") {
			return fmt.Errorf("cmKey %s expects T to be string, but got %v", cmKey, elem.Type())
		}
		elem.Set(reflect.ValueOf(data))
		return nil
	default:
		return json.Unmarshal([]byte(data), result)
	}
}
