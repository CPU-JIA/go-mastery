/*
Package resource 提供跨平台的系统资源管理工具。

本包支持以下功能：
  - 文件描述符管理和监控
  - 内存使用监控
  - CPU 使用率监控
  - 磁盘使用监控
  - 系统资源限制查询和设置

跨平台支持：
  - Windows: 使用 Windows API
  - Linux/Unix: 使用 /proc 文件系统和系统调用
*/
package resource

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ===================
// 错误定义
// ===================

var (
	// ErrResourceNotAvailable 资源不可用
	ErrResourceNotAvailable = errors.New("resource not available")
	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")
	// ErrLimitExceeded 超出限制
	ErrLimitExceeded = errors.New("resource limit exceeded")
)

// ===================
// 内存信息
// ===================

// MemoryInfo 内存信息
type MemoryInfo struct {
	// Total 总内存（字节）
	Total uint64
	// Available 可用内存（字节）
	Available uint64
	// Used 已用内存（字节）
	Used uint64
	// UsedPercent 使用百分比
	UsedPercent float64
	// Free 空闲内存（字节）
	Free uint64
	// Buffers 缓冲区（字节，仅Linux）
	Buffers uint64
	// Cached 缓存（字节，仅Linux）
	Cached uint64
	// SwapTotal 交换空间总量（字节）
	SwapTotal uint64
	// SwapUsed 已用交换空间（字节）
	SwapUsed uint64
	// SwapFree 空闲交换空间（字节）
	SwapFree uint64
}

// ProcessMemoryInfo 进程内存信息
type ProcessMemoryInfo struct {
	// RSS 常驻内存大小（字节）
	RSS uint64
	// VMS 虚拟内存大小（字节）
	VMS uint64
	// Shared 共享内存（字节）
	Shared uint64
	// Data 数据段（字节）
	Data uint64
	// Stack 栈（字节）
	Stack uint64
	// HeapAlloc 堆分配（字节）- Go runtime
	HeapAlloc uint64
	// HeapSys 堆系统内存（字节）- Go runtime
	HeapSys uint64
	// HeapIdle 堆空闲（字节）- Go runtime
	HeapIdle uint64
	// HeapInuse 堆使用中（字节）- Go runtime
	HeapInuse uint64
	// StackInuse 栈使用中（字节）- Go runtime
	StackInuse uint64
	// NumGC GC次数 - Go runtime
	NumGC uint32
}

// ===================
// CPU信息
// ===================

// CPUInfo CPU信息
type CPUInfo struct {
	// NumCPU CPU核心数
	NumCPU int
	// NumLogicalCPU 逻辑CPU数
	NumLogicalCPU int
	// ModelName CPU型号
	ModelName string
	// Vendor CPU厂商
	Vendor string
	// Family CPU家族
	Family string
	// Model CPU型号ID
	Model string
	// MHz CPU频率
	MHz float64
	// CacheSize 缓存大小（KB）
	CacheSize int
}

// CPUUsage CPU使用率
type CPUUsage struct {
	// User 用户态时间百分比
	User float64
	// System 内核态时间百分比
	System float64
	// Idle 空闲时间百分比
	Idle float64
	// IOWait I/O等待时间百分比（仅Linux）
	IOWait float64
	// IRQ 硬中断时间百分比（仅Linux）
	IRQ float64
	// SoftIRQ 软中断时间百分比（仅Linux）
	SoftIRQ float64
	// Steal 虚拟化偷取时间百分比（仅Linux）
	Steal float64
	// Total 总使用率
	Total float64
}

// ===================
// 磁盘信息
// ===================

// DiskInfo 磁盘信息
type DiskInfo struct {
	// Path 挂载路径
	Path string
	// Device 设备名
	Device string
	// FSType 文件系统类型
	FSType string
	// Total 总空间（字节）
	Total uint64
	// Used 已用空间（字节）
	Used uint64
	// Free 可用空间（字节）
	Free uint64
	// UsedPercent 使用百分比
	UsedPercent float64
	// InodesTotal Inode总数
	InodesTotal uint64
	// InodesUsed 已用Inode
	InodesUsed uint64
	// InodesFree 可用Inode
	InodesFree uint64
}

// DiskIOStats 磁盘I/O统计
type DiskIOStats struct {
	// Device 设备名
	Device string
	// ReadCount 读取次数
	ReadCount uint64
	// WriteCount 写入次数
	WriteCount uint64
	// ReadBytes 读取字节数
	ReadBytes uint64
	// WriteBytes 写入字节数
	WriteBytes uint64
	// ReadTime 读取时间（毫秒）
	ReadTime uint64
	// WriteTime 写入时间（毫秒）
	WriteTime uint64
	// IOTime I/O时间（毫秒）
	IOTime uint64
}

// ===================
// 文件描述符管理
// ===================

// FDInfo 文件描述符信息
type FDInfo struct {
	// FD 文件描述符号
	FD int
	// Path 文件路径
	Path string
	// Type 类型（file, socket, pipe等）
	Type string
	// Mode 模式
	Mode string
	// Flags 标志
	Flags int
}

// FDLimits 文件描述符限制
type FDLimits struct {
	// SoftLimit 软限制
	SoftLimit uint64
	// HardLimit 硬限制
	HardLimit uint64
	// Current 当前使用数
	Current uint64
}

// FDTracker 文件描述符追踪器
type FDTracker struct {
	// 追踪的文件描述符
	tracked map[int]*TrackedFD
	mu      sync.RWMutex
	// 统计信息
	stats FDStats
	// 警告阈值（百分比）
	warnThreshold float64
	// 回调函数
	onWarning func(current, limit uint64)
}

// TrackedFD 被追踪的文件描述符
type TrackedFD struct {
	FD         int
	Path       string
	Type       string
	OpenTime   time.Time
	LastAccess time.Time
	Stack      string // 打开时的调用栈
}

// FDStats 文件描述符统计
type FDStats struct {
	// TotalOpened 总打开数
	TotalOpened int64
	// TotalClosed 总关闭数
	TotalClosed int64
	// CurrentOpen 当前打开数
	CurrentOpen int64
	// MaxOpen 最大同时打开数
	MaxOpen int64
	// LeakSuspects 疑似泄漏数
	LeakSuspects int64
}

// NewFDTracker 创建文件描述符追踪器
func NewFDTracker() *FDTracker {
	return &FDTracker{
		tracked:       make(map[int]*TrackedFD),
		warnThreshold: 80.0, // 默认80%警告
	}
}

// SetWarningThreshold 设置警告阈值
func (t *FDTracker) SetWarningThreshold(percent float64) {
	t.warnThreshold = percent
}

// SetWarningCallback 设置警告回调
func (t *FDTracker) SetWarningCallback(callback func(current, limit uint64)) {
	t.onWarning = callback
}

// Track 追踪文件描述符
func (t *FDTracker) Track(fd int, path, fdType string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tracked[fd] = &TrackedFD{
		FD:         fd,
		Path:       path,
		Type:       fdType,
		OpenTime:   time.Now(),
		LastAccess: time.Now(),
		Stack:      getCallStack(),
	}

	atomic.AddInt64(&t.stats.TotalOpened, 1)
	current := atomic.AddInt64(&t.stats.CurrentOpen, 1)

	// 更新最大值
	for {
		max := atomic.LoadInt64(&t.stats.MaxOpen)
		if current <= max {
			break
		}
		if atomic.CompareAndSwapInt64(&t.stats.MaxOpen, max, current) {
			break
		}
	}

	// 检查是否需要警告
	t.checkWarning()
}

// Untrack 取消追踪文件描述符
func (t *FDTracker) Untrack(fd int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.tracked[fd]; exists {
		delete(t.tracked, fd)
		atomic.AddInt64(&t.stats.TotalClosed, 1)
		atomic.AddInt64(&t.stats.CurrentOpen, -1)
	}
}

// GetStats 获取统计信息
func (t *FDTracker) GetStats() FDStats {
	return FDStats{
		TotalOpened:  atomic.LoadInt64(&t.stats.TotalOpened),
		TotalClosed:  atomic.LoadInt64(&t.stats.TotalClosed),
		CurrentOpen:  atomic.LoadInt64(&t.stats.CurrentOpen),
		MaxOpen:      atomic.LoadInt64(&t.stats.MaxOpen),
		LeakSuspects: atomic.LoadInt64(&t.stats.LeakSuspects),
	}
}

// GetTracked 获取所有被追踪的文件描述符
func (t *FDTracker) GetTracked() []*TrackedFD {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*TrackedFD, 0, len(t.tracked))
	for _, fd := range t.tracked {
		result = append(result, fd)
	}
	return result
}

// FindLeaks 查找可能的泄漏
// maxAge: 超过此时间未访问的文件描述符被视为可疑
func (t *FDTracker) FindLeaks(maxAge time.Duration) []*TrackedFD {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var suspects []*TrackedFD
	threshold := time.Now().Add(-maxAge)

	for _, fd := range t.tracked {
		if fd.LastAccess.Before(threshold) {
			suspects = append(suspects, fd)
		}
	}

	atomic.StoreInt64(&t.stats.LeakSuspects, int64(len(suspects)))
	return suspects
}

// checkWarning 检查是否需要发出警告
func (t *FDTracker) checkWarning() {
	if t.onWarning == nil {
		return
	}

	limits, err := GetFDLimits()
	if err != nil {
		return
	}

	current := uint64(atomic.LoadInt64(&t.stats.CurrentOpen))
	threshold := uint64(float64(limits.SoftLimit) * t.warnThreshold / 100)

	if current >= threshold {
		t.onWarning(current, limits.SoftLimit)
	}
}

// getCallStack 获取调用栈
func getCallStack() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// ===================
// 资源监控器
// ===================

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	// 监控间隔
	interval time.Duration
	// 停止信号
	stopCh chan struct{}
	// 运行状态
	running bool
	mu      sync.Mutex
	// 回调函数
	onMemory func(MemoryInfo)
	onCPU    func(CPUUsage)
	onDisk   func([]DiskInfo)
	onFD     func(FDLimits)
	// 历史数据
	memoryHistory []MemoryInfo
	cpuHistory    []CPUUsage
	historySize   int
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(interval time.Duration) *ResourceMonitor {
	return &ResourceMonitor{
		interval:    interval,
		stopCh:      make(chan struct{}),
		historySize: 60, // 默认保留60个历史记录
	}
}

// SetHistorySize 设置历史记录大小
func (m *ResourceMonitor) SetHistorySize(size int) {
	m.historySize = size
}

// OnMemory 设置内存监控回调
func (m *ResourceMonitor) OnMemory(callback func(MemoryInfo)) {
	m.onMemory = callback
}

// OnCPU 设置CPU监控回调
func (m *ResourceMonitor) OnCPU(callback func(CPUUsage)) {
	m.onCPU = callback
}

// OnDisk 设置磁盘监控回调
func (m *ResourceMonitor) OnDisk(callback func([]DiskInfo)) {
	m.onDisk = callback
}

// OnFD 设置文件描述符监控回调
func (m *ResourceMonitor) OnFD(callback func(FDLimits)) {
	m.onFD = callback
}

// Start 启动监控
func (m *ResourceMonitor) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
	return nil
}

// Stop 停止监控
func (m *ResourceMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)
}

// GetMemoryHistory 获取内存历史数据
func (m *ResourceMonitor) GetMemoryHistory() []MemoryInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]MemoryInfo, len(m.memoryHistory))
	copy(result, m.memoryHistory)
	return result
}

// GetCPUHistory 获取CPU历史数据
func (m *ResourceMonitor) GetCPUHistory() []CPUUsage {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]CPUUsage, len(m.cpuHistory))
	copy(result, m.cpuHistory)
	return result
}

func (m *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

func (m *ResourceMonitor) collectMetrics() {
	// 收集内存信息
	if memInfo, err := GetMemoryInfo(); err == nil {
		m.mu.Lock()
		m.memoryHistory = append(m.memoryHistory, *memInfo)
		if len(m.memoryHistory) > m.historySize {
			m.memoryHistory = m.memoryHistory[1:]
		}
		m.mu.Unlock()

		if m.onMemory != nil {
			m.onMemory(*memInfo)
		}
	}

	// 收集CPU信息
	if cpuUsage, err := GetCPUUsage(); err == nil {
		m.mu.Lock()
		m.cpuHistory = append(m.cpuHistory, *cpuUsage)
		if len(m.cpuHistory) > m.historySize {
			m.cpuHistory = m.cpuHistory[1:]
		}
		m.mu.Unlock()

		if m.onCPU != nil {
			m.onCPU(*cpuUsage)
		}
	}

	// 收集磁盘信息
	if m.onDisk != nil {
		if diskInfo, err := GetDiskInfo(); err == nil {
			m.onDisk(diskInfo)
		}
	}

	// 收集文件描述符信息
	if m.onFD != nil {
		if fdLimits, err := GetFDLimits(); err == nil {
			m.onFD(*fdLimits)
		}
	}
}

// ===================
// 资源限制
// ===================

// ResourceLimits 资源限制
type ResourceLimits struct {
	// MaxOpenFiles 最大打开文件数
	MaxOpenFiles uint64
	// MaxProcesses 最大进程数
	MaxProcesses uint64
	// MaxMemory 最大内存（字节）
	MaxMemory uint64
	// MaxCPU CPU时间限制（秒）
	MaxCPU uint64
	// MaxFileSize 最大文件大小（字节）
	MaxFileSize uint64
	// MaxStack 最大栈大小（字节）
	MaxStack uint64
}

// ===================
// Go Runtime 信息
// ===================

// RuntimeInfo Go运行时信息
type RuntimeInfo struct {
	// Version Go版本
	Version string
	// NumCPU CPU数量
	NumCPU int
	// NumGoroutine Goroutine数量
	NumGoroutine int
	// GOMAXPROCS GOMAXPROCS值
	GOMAXPROCS int
	// MemStats 内存统计
	MemStats runtime.MemStats
}

// GetRuntimeInfo 获取Go运行时信息
func GetRuntimeInfo() *RuntimeInfo {
	info := &RuntimeInfo{
		Version:      runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
	}

	runtime.ReadMemStats(&info.MemStats)
	return info
}

// GetProcessMemoryInfo 获取当前进程内存信息
func GetProcessMemoryInfo() *ProcessMemoryInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &ProcessMemoryInfo{
		HeapAlloc:  memStats.HeapAlloc,
		HeapSys:    memStats.HeapSys,
		HeapIdle:   memStats.HeapIdle,
		HeapInuse:  memStats.HeapInuse,
		StackInuse: memStats.StackInuse,
		NumGC:      memStats.NumGC,
	}
}

// ForceGC 强制执行垃圾回收
func ForceGC() {
	runtime.GC()
}

// FreeOSMemory 释放内存给操作系统
func FreeOSMemory() {
	runtime.GC()
	// debug.FreeOSMemory() 可以更积极地释放内存
}

// ===================
// 工具函数
// ===================

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ParseBytes 解析字节数字符串
func ParseBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, fmt.Errorf("empty string")
	}

	// 查找数字和单位的分界
	var numStr string
	var unit string
	for i, c := range s {
		if c < '0' || c > '9' {
			if c != '.' {
				numStr = s[:i]
				unit = strings.TrimSpace(s[i:])
				break
			}
		}
	}

	if numStr == "" {
		numStr = s
	}

	var num float64
	if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}

	unit = strings.ToUpper(unit)
	multiplier := uint64(1)

	switch unit {
	case "K", "KB", "KIB":
		multiplier = 1024
	case "M", "MB", "MIB":
		multiplier = 1024 * 1024
	case "G", "GB", "GIB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB", "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "B", "":
		multiplier = 1
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return uint64(num * float64(multiplier)), nil
}
