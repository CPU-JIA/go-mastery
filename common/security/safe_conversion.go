// Package security 提供安全的数据类型转换和安全编程工具
//
// 本包专门解决Go代码中的安全漏洞，特别是：
// - G115: 整数溢出转换
// - 提供安全的类型转换函数
// - 建立项目范围内的安全编程标准
package security

import (
	"fmt"
	"math"
)

// SafeUint64ToInt64 安全地将uint64转换为int64
// 如果uint64值超过int64的最大值，返回错误
func SafeUint64ToInt64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("value %d exceeds int64 maximum (%d)", value, math.MaxInt64)
	}
	return int64(value), nil
}

// MustSafeUint64ToInt64 安全转换uint64到int64，溢出时返回MaxInt64
// 用于性能监控等场景，溢出时提供合理的上限值而不是错误
func MustSafeUint64ToInt64(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}

// SafeUint32ToInt32 安全地将uint32转换为int32
func SafeUint32ToInt32(value uint32) (int32, error) {
	if value > math.MaxInt32 {
		return 0, fmt.Errorf("value %d exceeds int32 maximum (%d)", value, math.MaxInt32)
	}
	return int32(value), nil
}

// MustSafeUint32ToInt32 安全转换uint32到int32，溢出时返回MaxInt32
func MustSafeUint32ToInt32(value uint32) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

// SafeIntToUint 安全地将int转换为uint
func SafeIntToUint(value int) (uint, error) {
	if value < 0 {
		return 0, fmt.Errorf("negative value %d cannot be converted to uint", value)
	}
	return uint(value), nil
}

// SafeInt64ToUint64 安全地将int64转换为uint64
func SafeInt64ToUint64(value int64) (uint64, error) {
	if value < 0 {
		return 0, fmt.Errorf("negative value %d cannot be converted to uint64", value)
	}
	return uint64(value), nil
}

// ConversionStats 记录转换统计信息，用于监控和调试
type ConversionStats struct {
	TotalConversions uint64
	OverflowCount    uint64
	UnderflowCount   uint64
}

// NewConversionStats 创建新的转换统计实例
func NewConversionStats() *ConversionStats {
	return &ConversionStats{}
}

// TrackingUint64ToInt64 带统计追踪的uint64到int64转换
func (cs *ConversionStats) TrackingUint64ToInt64(value uint64) int64 {
	cs.TotalConversions++

	if value > math.MaxInt64 {
		cs.OverflowCount++
		return math.MaxInt64
	}
	return int64(value)
}

// GetOverflowRate 获取溢出率
func (cs *ConversionStats) GetOverflowRate() float64 {
	if cs.TotalConversions == 0 {
		return 0
	}
	return float64(cs.OverflowCount) / float64(cs.TotalConversions)
}
