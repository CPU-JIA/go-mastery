package calculator

import "math"

// 数学常量
const (
	// Pi 圆周率
	Pi = math.Pi

	// E 自然常数
	E = math.E

	// GoldenRatio 黄金比例
	GoldenRatio = 1.618033988749895
)

// 计算精度设置
const (
	// DefaultPrecision 默认精度
	DefaultPrecision = 1e-9

	// HighPrecision 高精度
	HighPrecision = 1e-15
)

// 操作类型常量
const (
	OpAdd = iota
	OpSubtract
	OpMultiply
	OpDivide
	OpPower
)
