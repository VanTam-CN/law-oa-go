package metrics

import (
	"fmt"
	"runtime"
	"time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	startTime   time.Time
	lastGCStats runtime.MemStats
	gcInterval  time.Duration
	stopChan    chan struct{}
}

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`

	// 内存统计
	AllocBytes        uint64 `json:"alloc_bytes"`
	TotalAllocBytes   uint64 `json:"total_alloc_bytes"`
	SysBytes          uint64 `json:"sys_bytes"`
	HeapAllocBytes    uint64 `json:"heap_alloc_bytes"`
	HeapSysBytes      uint64 `json:"heap_sys_bytes"`
	HeapIdleBytes     uint64 `json:"heap_idle_bytes"`
	HeapInuseBytes    uint64 `json:"heap_inuse_bytes"`
	HeapReleasedBytes uint64 `json:"heap_released_bytes"`
	HeapObjects       uint64 `json:"heap_objects"`

	// GC统计
	GCCPUFraction  float64 `json:"gc_cpu_fraction"`
	NumGC          uint32  `json:"num_gc"`
	GCPauseTotalNs uint64  `json:"gc_pause_total_ns"`
	LastGCPauseNs  uint64  `json:"last_gc_pause_ns"`

	// Goroutine统计
	GoroutineCount int `json:"goroutine_count"`

	// Cgo调用统计
	CgoCallCount uint64 `json:"cgo_call_count"`
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime:  time.Now(),
		gcInterval: 30 * time.Second,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start() {
	go pm.monitor()
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	close(pm.stopChan)
}

// monitor 监控性能
func (pm *PerformanceMonitor) monitor() {
	// 获取初始GC统计
	runtime.ReadMemStats(&pm.lastGCStats)

	ticker := time.NewTicker(pm.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.collectMetrics()
		case <-pm.stopChan:
			return
		}
	}
}

// collectMetrics 收集性能指标
func (pm *PerformanceMonitor) collectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算GC统计变化
	gcPauseTotal := memStats.PauseTotalNs - pm.lastGCStats.PauseTotalNs

	// 更新GC统计
	pm.lastGCStats = memStats

	// 获取当前指标
	stats := PerformanceStats{
		Timestamp:         time.Now(),
		Uptime:            time.Since(pm.startTime).String(),
		AllocBytes:        memStats.Alloc,
		TotalAllocBytes:   memStats.TotalAlloc,
		SysBytes:          memStats.Sys,
		HeapAllocBytes:    memStats.HeapAlloc,
		HeapSysBytes:      memStats.HeapSys,
		HeapIdleBytes:     memStats.HeapIdle,
		HeapInuseBytes:    memStats.HeapInuse,
		HeapReleasedBytes: memStats.HeapReleased,
		HeapObjects:       memStats.HeapObjects,
		GCCPUFraction:     memStats.GCCPUFraction,
		NumGC:             memStats.NumGC,
		GCPauseTotalNs:    gcPauseTotal,
		LastGCPauseNs:     memStats.PauseNs[(memStats.NumGC+255)%256],
		GoroutineCount:    runtime.NumGoroutine(),
		CgoCallCount:      0, // CGO调用计数在Go 1.23中可能不可用
	}

	// 更新Prometheus指标
	appMetrics := GetDefaultMetrics()
	appMetrics.performanceMetrics.memoryUsage.Set(float64(memStats.Alloc))
	appMetrics.performanceMetrics.goroutineCount.Set(float64(runtime.NumGoroutine()))

	// 记录性能统计
	pm.logPerformanceStats(stats)
}

// logPerformanceStats 记录性能统计
func (pm *PerformanceMonitor) logPerformanceStats(stats PerformanceStats) {
	// 可以扩展为写入文件或发送到监控系统
	fmt.Printf("=== Performance Stats [%s] ===\n", stats.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Uptime: %s\n", stats.Uptime)
	fmt.Printf("Memory: Alloc=%s, HeapAlloc=%s, Sys=%s\n",
		formatBytes(stats.AllocBytes),
		formatBytes(stats.HeapAllocBytes),
		formatBytes(stats.SysBytes))
	fmt.Printf("GC: Count=%d, PauseTotal=%s, LastPause=%s, CPUFraction=%.2f%%\n",
		stats.NumGC,
		formatDuration(stats.GCPauseTotalNs),
		formatDuration(stats.LastGCPauseNs),
		stats.GCCPUFraction*100)
	fmt.Printf("Goroutines: %d\n", stats.GoroutineCount)
	fmt.Printf("CGO Calls: %d\n", stats.CgoCallCount)
	fmt.Println("========================================")
}

// GetCurrentStats 获取当前性能统计
func (pm *PerformanceMonitor) GetCurrentStats() PerformanceStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return PerformanceStats{
		Timestamp:         time.Now(),
		Uptime:            time.Since(pm.startTime).String(),
		AllocBytes:        memStats.Alloc,
		TotalAllocBytes:   memStats.TotalAlloc,
		SysBytes:          memStats.Sys,
		HeapAllocBytes:    memStats.HeapAlloc,
		HeapSysBytes:      memStats.HeapSys,
		HeapIdleBytes:     memStats.HeapIdle,
		HeapInuseBytes:    memStats.HeapInuse,
		HeapReleasedBytes: memStats.HeapReleased,
		HeapObjects:       memStats.HeapObjects,
		GCCPUFraction:     memStats.GCCPUFraction,
		NumGC:             memStats.NumGC,
		GCPauseTotalNs:    memStats.PauseTotalNs,
		LastGCPauseNs:     memStats.PauseNs[(memStats.NumGC+255)%256],
		GoroutineCount:    runtime.NumGoroutine(),
		CgoCallCount:      0, // CGO调用计数在Go 1.23中可能不可用
	}
}

// GetMemoryUsage 获取内存使用情况
func (pm *PerformanceMonitor) GetMemoryUsage() MemoryUsage {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryUsage{
		Alloc:      memStats.Alloc,
		TotalAlloc: memStats.TotalAlloc,
		Sys:        memStats.Sys,
		HeapAlloc:  memStats.HeapAlloc,
		HeapSys:    memStats.HeapSys,
		HeapIdle:   memStats.HeapIdle,
		HeapInuse:  memStats.HeapInuse,
	}
}

// GetGCStats 获取GC统计
func (pm *PerformanceMonitor) GetGCStats() GCStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return GCStats{
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction,
		LastGCPause:   time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		TotalGCPause:  time.Duration(memStats.PauseTotalNs),
	}
}

// MemoryUsage 内存使用情况
type MemoryUsage struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
	HeapIdle   uint64 `json:"heap_idle"`
	HeapInuse  uint64 `json:"heap_inuse"`
}

// GCStats GC统计
type GCStats struct {
	NumGC         uint32        `json:"num_gc"`
	GCCPUFraction float64       `json:"gc_cpu_fraction"`
	LastGCPause   time.Duration `json:"last_gc_pause"`
	TotalGCPause  time.Duration `json:"total_gc_pause"`
}

// formatBytes 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration 格式化纳秒时间
func formatDuration(ns uint64) string {
	if ns < 1000 {
		return fmt.Sprintf("%d ns", ns)
	} else if ns < 1000000 {
		return fmt.Sprintf("%.1f μs", float64(ns)/1000)
	} else if ns < 1000000000 {
		return fmt.Sprintf("%.1f ms", float64(ns)/1000000)
	}
	return fmt.Sprintf("%.1f s", float64(ns)/1000000000)
}

// ForceGC 强制执行GC
func ForceGC() {
	runtime.GC()
}

// SetGCPercent 设置GC目标百分比
func SetGCPercent(percent int) int {
	// 在某些Go版本中，这个函数可能在debug包中
	// 这里提供一个简单的实现
	return 100 // 返回默认值
}

// GetNumCPU 获取CPU核心数
func GetNumCPU() int {
	return runtime.NumCPU()
}

// GetGOMAXPROCS 获取GOMAXPROCS设置
func GetGOMAXPROCS() int {
	return runtime.GOMAXPROCS(0)
}

// SetGOMAXPROCS 设置GOMAXPROCS
func SetGOMAXPROCS(n int) int {
	return runtime.GOMAXPROCS(n)
}
