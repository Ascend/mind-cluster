package diagtemplate

import (
	"math"

	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/context/diagcontext/metricpool"
	"ascend-faultdiag-online/pkg/utils"
)

// FloatMetricCompareFunc 指标比较函数
type FloatMetricCompareFunc func(metric, threshold float64) *diagcontext.CompareRes

// StringMetricCompareFunc 指标比较函数
type StringMetricCompareFunc func(metric, threshold string) *diagcontext.CompareRes

func buildMetricDiagRes(domainMetrics []*metricpool.DomainMetrics, metricName string,
	compareFunc diagcontext.MetricCompareFunc, threshold *diagcontext.MetricThreshold,
	results []*diagcontext.MetricDiagRes) []*diagcontext.MetricDiagRes {
	for _, metric := range domainMetrics {
		itemGroup, ok := metric.ItemGroupMap[metricName]
		if !ok {
			continue
		}
		poolItem := itemGroup.GetLatestMetricPoolItem()
		compareRes := compareFunc(poolItem.Value, threshold.Value)
		results = append(results, &diagcontext.MetricDiagRes{
			Metric:      itemGroup.Metric,
			Value:       poolItem.Value,
			Threshold:   threshold.Value,
			Unit:        threshold.Unit,
			Time:        poolItem.Timestamp,
			IsAbnormal:  compareRes.IsAbnormal,
			Description: compareRes.Description,
		})
	}
	return results
}

// SingleMetricDiagFunc 单个指标诊断事件
func SingleMetricDiagFunc(targetThreshold *diagcontext.MetricThreshold,
	compareFunc diagcontext.MetricCompareFunc) diagcontext.DiagFunc {
	return func(diagItem *diagcontext.DiagItem, thresholds []*diagcontext.MetricThreshold,
		domainMetrics []*metricpool.DomainMetrics) []*diagcontext.MetricDiagRes {
		var results []*diagcontext.MetricDiagRes
		for _, threshold := range thresholds {
			if threshold.Name != targetThreshold.Name {
				continue
			}
			results = buildMetricDiagRes(domainMetrics, threshold.Name, compareFunc, threshold, results)
		}
		return results
	}
}

// SingleFloat64MetricDiagFunc 单个指标诊断事件， 请保证参数正确
func SingleFloat64MetricDiagFunc(threshold *diagcontext.MetricThreshold,
	float64CompareFunc FloatMetricCompareFunc) diagcontext.DiagFunc {
	compareFunc := func(metric, threshold interface{}) *diagcontext.CompareRes {
		return float64CompareFunc(utils.ToFloat64(metric, math.MaxFloat64), utils.ToFloat64(threshold, math.MaxFloat64))
	}
	return SingleMetricDiagFunc(threshold, compareFunc)
}

// SingleStringMetricDiagFunc 单个指标诊断事件， 请保证参数正确
func SingleStringMetricDiagFunc(threshold *diagcontext.MetricThreshold,
	stringCompareFunc StringMetricCompareFunc) diagcontext.DiagFunc {
	compareFunc := func(metric, threshold interface{}) *diagcontext.CompareRes {
		return stringCompareFunc(utils.ToString(metric), utils.ToString(threshold))
	}
	return SingleMetricDiagFunc(threshold, compareFunc)
}
