package diagcontext

import (
	"strings"

	"ascend-faultdiag-online/pkg/model/diagmodel/metricmodel"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

func buildDomainItemsKey(domainItems []*metricmodel.DomainItem) string {
	keys := slicetool.MapToValue(domainItems, func(item *metricmodel.DomainItem) string {
		return item.GetDomainItemKey()
	})
	return strings.Join(keys, constants.TypeSeparator)
}

type Domain struct {
	DomainItems []*metricmodel.DomainItem
}

func (domain *Domain) GetDomainKey() string {
	return buildDomainItemsKey(domain.DomainItems)
}

func (domain *Domain) Size() int {
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
	key := buildDomainItemsKey(domainItems)
	domain, ok := factory.domainMap[key]
	if !ok {
		domain = &Domain{domainItems}
		factory.domainMap[key] = domain
	}
	return domain
}
