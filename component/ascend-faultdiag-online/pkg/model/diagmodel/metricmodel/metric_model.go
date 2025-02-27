package metricmodel

import (
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils/constants"
)

// DomainItem 指标域单项
type DomainItem struct {
	DomainType enum.MetricDomainType `json:"domain_type"`
	Value      string                `json:"value"`
}

func (item *DomainItem) GetDomainItemKey() string {
	return string(item.DomainType) + constants.ValueSeparator + item.Value
}

// MetricReqModel 指标请求数据模型
type MetricReqModel struct {
	Domain    []*DomainItem        `json:"domain"`
	Name      string               `json:"name"`
	ValueType enum.MetricValueType `json:"value_type"`
	Value     string               `json:"value"`
}

// MetricReqData 指标请求data
type MetricReqData struct {
	Metrics []*MetricReqModel `json:"metrics"`
}
