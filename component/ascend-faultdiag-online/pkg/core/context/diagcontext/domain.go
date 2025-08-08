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
Package diagcontext provides the function to generate a factory class of domain.
*/
package diagcontext

import (
	"strings"

	"ascend-faultdiag-online/pkg/core/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

func buildDomainItemsKey(domainItems []*metricmodel.DomainItem) string {
	keys := slicetool.MapToValue(domainItems, func(item *metricmodel.DomainItem) string {
		return item.GetDomainItemKey()
	})
	return strings.Join(keys, constants.TypeSeparator)
}

// Domain is a collection of Domain items
type Domain struct {
	DomainItems []*metricmodel.DomainItem
}

// GetDomainKey get the key of Domain
func (domain *Domain) GetDomainKey() string {
	if domain == nil {
		return ""
	}
	return buildDomainItemsKey(domain.DomainItems)
}

// Size get the size of domain items.
func (domain *Domain) Size() int {
	if domain == nil {
		return 0
	}
	return len(domain.DomainItems)
}

// DomainFactory 域工厂类，生成不重复实例
type DomainFactory struct {
	domainMap map[string]*Domain
}

// NewDomainFactory 创建一个工厂实例
func NewDomainFactory() *DomainFactory {
	return &DomainFactory{domainMap: make(map[string]*Domain)}
}

// GetInstance 获取实例
func (factory *DomainFactory) GetInstance(domainItems []*metricmodel.DomainItem) *Domain {
	if factory == nil {
		return &Domain{}
	}
	key := buildDomainItemsKey(domainItems)
	domain, ok := factory.domainMap[key]
	if !ok {
		domain = &Domain{
			DomainItems: domainItems,
		}
		factory.domainMap[key] = domain
	}
	return domain
}
