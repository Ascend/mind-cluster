package utils

import (
	"strconv"
)

// ToFloat64 interface转float64
func ToFloat64(val interface{}, defaultValue float64) float64 {
	switch v := val.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		float, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue
		}
		return float
	default:
		return defaultValue
	}
}

// ToString interface转string
func ToString(val interface{}) string {
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}
