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

// Package handlejson provides node parse
package handlejson

import (
	"fmt"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/utils"
)

const (
	// MaxIdNameDomainSize 缓存最大容量
	MaxIdNameDomainSize = 100
	// ItemSize2 计数为2
	ItemSize2 = 2
	// ItemSize4 计数为4
	ItemSize4 = 4
)

// domain2Items 2个item的domain
var domain2Items = []string{constants.CKPTWord, constants.DataLoaderWord}

// TimeStampFile 保存文件时间戳
type TimeStampFile struct {
	// Name 文件名字
	Name string
	// Offset 位置指针
	Offset int64
}

// CacheNameDomain 缓存name和domain字段
type CacheNameDomain struct {
	Name   string
	Domain string
	Cnt    int
}

// CacheData 缓存json数据
type CacheData struct {
	cAnnApi    *model.CAnnApi
	commOp     *model.CommOp
	mSTXEvents *model.MSTXEvents
	stepTime   *model.StepTime
	task       *model.Task
	dbName     string
}

// EventStack 缓存json数据的栈
type EventStack struct {
	JsonDataList []*model.JsonData
}

// ParseFileContext 文件清洗上下文
type ParseFileContext struct {
	CurFile  *TimeStampFile
	JsonData []*model.JsonData
	// EventStackMap 缓存开始和结束的两条profiling数据
	EventStackMap map[string]*EventStack
	// EventStackQue
	EventStackQue chan *EventStack
	// IdNameDomain 缓存id相同的profiling数据（Name和Domain）
	IdNameDomain map[int64]*CacheNameDomain
	// OrderedIDSet 记录id顺序缓存
	OrderedIDSet *utils.OrderedIDSet
}

// NewParseFileContext 创建文件清洗上下文实例
func NewParseFileContext() *ParseFileContext {
	return &ParseFileContext{
		CurFile:       &TimeStampFile{"", 0},
		JsonData:      make([]*model.JsonData, 0),
		EventStackMap: make(map[string]*EventStack),
		EventStackQue: make(chan *EventStack, MaxIdNameDomainSize),
		IdNameDomain:  make(map[int64]*CacheNameDomain),
		OrderedIDSet:  utils.NewOrderedIDSet(),
	}
}

// HandleIdDomain 处理id和Domain缓存
func (ctx *ParseFileContext) HandleIdDomain(jsonData *model.JsonData) {
	domain, ok := ctx.IdNameDomain[jsonData.Id]
	if !ok {
		ctx.IdNameDomain[jsonData.Id] = &CacheNameDomain{Domain: jsonData.Domain, Name: jsonData.Name, Cnt: 1}
		return
	}
	domain.Cnt += 1
	jsonData.Domain = domain.Domain
	jsonData.Name = domain.Name

	if (utils.InSlice(domain2Items, jsonData.Name) && domain.Cnt == ItemSize2) || domain.Cnt == ItemSize4 {
		delete(ctx.IdNameDomain, jsonData.Id)
		ctx.OrderedIDSet.Remove(jsonData.Id)
	}
}

// DealJsonData 处理json数据
func (ctx *ParseFileContext) DealJsonData(data *model.JsonData) {
	const startAndEndLen = 2
	if !ctx.OrderedIDSet.Contains(data.Id) {
		ctx.OrderedIDSet.Add(data.Id)
	}
	if len(ctx.IdNameDomain) > MaxIdNameDomainSize {
		id, found := ctx.OrderedIDSet.GetByIndex(0)
		if found {
			delete(ctx.IdNameDomain, id)
			ctx.OrderedIDSet.Remove(id)
		}
	}
	// 记录临时栈
	cacheKey := fmt.Sprintf("%d-", data.Id) + fmt.Sprintf("%d", data.SourceKind)
	stack, ok := ctx.EventStackMap[cacheKey]
	if !ok {
		stack = &EventStack{JsonDataList: make([]*model.JsonData, 0)}
		ctx.EventStackMap[cacheKey] = stack
	}
	stack.JsonDataList = append(stack.JsonDataList, data)

	// 找到两条profiling数据（开始和结束）时，加入事件队列
	if len(stack.JsonDataList) == startAndEndLen {
		ctx.EventStackQue <- stack
		delete(ctx.EventStackMap, cacheKey)
	}
}
