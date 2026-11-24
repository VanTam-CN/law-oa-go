// Package safeconv 提供带边界检查的整型转换，供监控/健康检查等
// 对运行时统计值做 uint64/int64 互转时使用，避免无符号溢出回绕。
//
// 引入背景（t_745060c0，gosec G115）：runtime.MemStats 与 syscall.Statfs
// 字段为 uint64/int64，直接互转可能溢出（内存/磁盘统计在物理上远小于
// int64 上限，但 gosec 静态标记该转换模式）。本包以饱和/钳制语义收敛：
// 溢出时不回绕，而是钳制到可表示边界，保证监控值始终非负且不产生
// 负数读数。仅用于监控/统计类转换；业务数值转换请使用 strconv。
package safeconv

import "math"

// Int64 将 uint64 饱和转换为 int64：超过 MaxInt64 时返回 MaxInt64，
// 避免无符号大数回绕为负数。
func Int64(u uint64) int64 {
	if u > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(u)
}

// Uint64 将 int64 钳制转换为 uint64：负数（理论上的非法块大小等）
// 归零，避免负数回绕为极大无符号数。
func Uint64(v int64) uint64 {
	if v < 0 {
		return 0
	}
	return uint64(v)
}
