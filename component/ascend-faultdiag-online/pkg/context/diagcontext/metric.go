package diagcontext

import (
	"ascend-faultdiag-online/pkg/utils/constants"
)

// Metric 指标结构体, 抽象的指标，包含指标域和指标名，不包含具体的值
type Metric struct {
	Domain *Domain // 指标域
	Name   string  // 指标名
}

func (item *Metric) GetMetricKey() string {
	return item.Domain.GetDomainKey() + constants.TypeSeparator + item.Name
}
