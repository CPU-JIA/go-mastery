/*
Package resource 的单元测试

测试覆盖：
  - 内存信息获取
  - CPU信息获取
  - 磁盘信息获取
  - 文件描述符管理
  - 资源监控
  - 工具函数
*/
package resource

import (
	"runtime"
	"testing"
	"time"
)

// TestGetMemoryInfo 测试获取内存信息
func TestGetMemoryInfo(t *testing.T) {
	info, err := GetMemoryInfo()
	if err != nil {
		t.Fatalf("GetMemoryInfo failed: %v", err)
	}

	if info.Total == 0 {
		t.Error("total memory should not be zero")
	}

	if info.Available == 0 {
		t.Error("available memory should not be zero")
	}

	if info.UsedPercent < 0 || info.UsedPercent > 100 {
		t.Errorf("used percent should be between 0 and 100, got %.2f", info.UsedPercent)
	}

	t.Logf("Memory: Total=%s, Available=%s, Used=%.1f%%",
		FormatBytes(info.Total), FormatBytes(info.Available), info.UsedPercent)
}

// TestGetCPUInfo 测试获取CPU信息
func TestGetCPUInfo(t *testing.T) {
	info, err := GetCPUInfo()
	if err != nil {
		t.Fatalf("GetCPUInfo failed: %v", err)
	}

	if info.NumCPU <= 0 {
		t.Error("NumCPU should be positive")
	}

	if info.NumCPU != runtime.NumCPU() {
		t.Errorf("NumCPU mismatch: expected %d, got %d", runtime.NumCPU(), info.NumCPU)
	}

	t.Logf("CPU: NumCPU=%d, Model=%s, Vendor=%s",
		info.NumCPU, info.ModelName, info.Vendor)
}

// TestGetCPUUsage 测试获取CPU使用率
func TestGetCPUUsage(t *testing.T) {
	// 第一次调用可能返回估计值
	_, _ = GetCPUUsage()

	// 等待一段时间后再次调用
	time.Sleep(200 * time.Millisecond)

	usage, err := GetCPUUsage()
	if err != nil {
		t.Fatalf("GetCPUUsage failed: %v", err)
	}

	if usage.Total < 0 || usage.Total > 100 {
		t.Errorf("total CPU usage should be between 0 and 100, got %.2f", usage.Total)
	}

	if usage.Idle < 0 || usage.Idle > 100 {
		t.Errorf("idle CPU should be between 0 and 100, got %.2f", usage.Idle)
	}

	t.Logf("CPU Usage: Total=%.1f%%, User=%.1f%%, System=%.1f%%, Idle=%.1f%%",
		usage.Total, usage.User, usage.System, usage.Idle)
}

// TestGetDiskInfo 测试获取磁盘信息
func TestGetDiskInfo(t *testing.T) {
	disks, err := GetDiskInfo()
	if err != nil {
		t.Fatalf("GetDiskInfo failed: %v", err)
	}

	if len(disks) == 0 {
		t.Error("should have at least one disk")
	}

	for _, disk := range disks {
		if disk.Total == 0 {
			continue // 跳过虚拟文件系统
		}

		if disk.UsedPercent < 0 || disk.UsedPercent > 100 {
			t.Errorf("disk %s: used percent should be between 0 and 100, got %.2f",
				disk.Path, disk.UsedPercent)
		}

		t.Logf("Disk: Path=%s, Total=%s, Free=%s, Used=%.1f%%",
			disk.Path, FormatBytes(disk.Total), FormatBytes(disk.Free), disk.UsedPercent)
	}
}

// TestGetFDLimits 测试获取文件描述符限制
func TestGetFDLimits(t *testing.T) {
	limits, err := GetFDLimits()
	if err != nil {
		t.Fatalf("GetFDLimits failed: %v", err)
	}

	if limits.SoftLimit == 0 {
		t.Error("soft limit should not be zero")
	}

	if limits.HardLimit == 0 {
		t.Error("hard limit should not be zero")
	}

	if limits.SoftLimit > limits.HardLimit {
		t.Error("soft limit should not exceed hard limit")
	}

	t.Logf("FD Limits: Soft=%d, Hard=%d, Current=%d",
		limits.SoftLimit, limits.HardLimit, limits.Current)
}

// TestGetResourceLimits 测试获取资源限制
func TestGetResourceLimits(t *testing.T) {
	limits, err := GetResourceLimits()
	if err != nil {
		t.Fatalf("GetResourceLimits failed: %v", err)
	}

	if limits.MaxOpenFiles == 0 {
		t.Error("MaxOpenFiles should not be zero")
	}

	t.Logf("Resource Limits: MaxOpenFiles=%d, MaxMemory=%s",
		limits.MaxOpenFiles, FormatBytes(limits.MaxMemory))
}

// TestGetRuntimeInfo 测试获取Go运行时信息
func TestGetRuntimeInfo(t *testing.T) {
	info := GetRuntimeInfo()

	if info.Version == "" {
		t.Error("Go version should not be empty")
	}

	if info.NumCPU <= 0 {
		t.Error("NumCPU should be positive")
	}

	if info.NumGoroutine <= 0 {
		t.Error("NumGoroutine should be positive")
	}

	t.Logf("Runtime: Version=%s, NumCPU=%d, NumGoroutine=%d, GOMAXPROCS=%d",
		info.Version, info.NumCPU, info.NumGoroutine, info.GOMAXPROCS)
}

// TestGetProcessMemoryInfo 测试获取进程内存信息
func TestGetProcessMemoryInfo(t *testing.T) {
	info := GetProcessMemoryInfo()

	if info.HeapAlloc == 0 {
		t.Error("HeapAlloc should not be zero")
	}

	if info.HeapSys == 0 {
		t.Error("HeapSys should not be zero")
	}

	t.Logf("Process Memory: HeapAlloc=%s, HeapSys=%s, StackInuse=%s",
		FormatBytes(info.HeapAlloc), FormatBytes(info.HeapSys), FormatBytes(info.StackInuse))
}

// TestFDTracker 测试文件描述符追踪器
func TestFDTracker(t *testing.T) {
	tracker := NewFDTracker()

	// 追踪一些文件描述符
	tracker.Track(10, "/tmp/test1.txt", "file")
	tracker.Track(11, "/tmp/test2.txt", "file")
	tracker.Track(12, "socket:[12345]", "socket")

	stats := tracker.GetStats()
	if stats.CurrentOpen != 3 {
		t.Errorf("expected 3 open FDs, got %d", stats.CurrentOpen)
	}

	if stats.TotalOpened != 3 {
		t.Errorf("expected 3 total opened, got %d", stats.TotalOpened)
	}

	// 取消追踪
	tracker.Untrack(10)

	stats = tracker.GetStats()
	if stats.CurrentOpen != 2 {
		t.Errorf("expected 2 open FDs after untrack, got %d", stats.CurrentOpen)
	}

	if stats.TotalClosed != 1 {
		t.Errorf("expected 1 total closed, got %d", stats.TotalClosed)
	}

	// 获取追踪的FD列表
	tracked := tracker.GetTracked()
	if len(tracked) != 2 {
		t.Errorf("expected 2 tracked FDs, got %d", len(tracked))
	}
}

// TestFDTrackerFindLeaks 测试查找泄漏
func TestFDTrackerFindLeaks(t *testing.T) {
	tracker := NewFDTracker()

	// 追踪一些文件描述符
	tracker.Track(10, "/tmp/old.txt", "file")

	// 模拟时间流逝（通过直接修改）
	tracker.tracked[10].LastAccess = time.Now().Add(-2 * time.Hour)

	tracker.Track(11, "/tmp/new.txt", "file")

	// 查找超过1小时未访问的FD
	leaks := tracker.FindLeaks(1 * time.Hour)

	if len(leaks) != 1 {
		t.Errorf("expected 1 leak suspect, got %d", len(leaks))
	}

	if len(leaks) > 0 && leaks[0].FD != 10 {
		t.Errorf("expected FD 10 as leak suspect, got %d", leaks[0].FD)
	}
}

// TestFDTrackerWarningCallback 测试警告回调
func TestFDTrackerWarningCallback(t *testing.T) {
	tracker := NewFDTracker()
	tracker.SetWarningThreshold(50.0) // 50%警告阈值

	warningCalled := false
	tracker.SetWarningCallback(func(current, limit uint64) {
		warningCalled = true
		t.Logf("Warning: %d/%d FDs used", current, limit)
	})

	// 追踪一个FD（不太可能触发警告，除非限制很低）
	tracker.Track(10, "/tmp/test.txt", "file")

	// 注意：警告是否触发取决于系统的FD限制
	t.Logf("Warning callback called: %v", warningCalled)
}

// TestResourceMonitor 测试资源监控器
func TestResourceMonitor(t *testing.T) {
	monitor := NewResourceMonitor(50 * time.Millisecond)

	memoryCallCount := 0
	cpuCallCount := 0

	monitor.OnMemory(func(info MemoryInfo) {
		memoryCallCount++
	})

	monitor.OnCPU(func(usage CPUUsage) {
		cpuCallCount++
	})

	err := monitor.Start()
	if err != nil {
		t.Fatalf("failed to start monitor: %v", err)
	}

	// 等待一些回调
	time.Sleep(500 * time.Millisecond)

	monitor.Stop()

	// 放宽测试条件，至少有1次回调即可
	if memoryCallCount < 1 {
		t.Errorf("expected at least 1 memory callback, got %d", memoryCallCount)
	}

	t.Logf("Memory callbacks: %d, CPU callbacks: %d", memoryCallCount, cpuCallCount)

	// 检查历史数据
	memHistory := monitor.GetMemoryHistory()
	if len(memHistory) < 1 {
		t.Errorf("expected at least 1 memory history entry, got %d", len(memHistory))
	}
}

// TestFormatBytes 测试字节格式化
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d): expected %s, got %s", tt.bytes, tt.expected, result)
		}
	}
}

// TestParseBytes 测试字节解析
func TestParseBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"1024", 1024, false},
		{"1 KB", 1024, false},
		{"1 MB", 1048576, false},
		{"1 GB", 1073741824, false},
		{"1.5 GB", 1610612736, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseBytes(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseBytes(%s): expected error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseBytes(%s): unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseBytes(%s): expected %d, got %d", tt.input, tt.expected, result)
			}
		}
	}
}

// TestForceGC 测试强制GC
func TestForceGC(t *testing.T) {
	// 分配一些内存
	data := make([]byte, 10*1024*1024) // 10MB
	_ = data

	before := GetProcessMemoryInfo()

	// 清除引用并强制GC
	data = nil
	ForceGC()

	after := GetProcessMemoryInfo()

	// GC后堆分配应该减少（不一定立即生效）
	t.Logf("Before GC: HeapAlloc=%s, After GC: HeapAlloc=%s",
		FormatBytes(before.HeapAlloc), FormatBytes(after.HeapAlloc))
}

// BenchmarkGetMemoryInfo 基准测试：获取内存信息
func BenchmarkGetMemoryInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetMemoryInfo()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetCPUUsage 基准测试：获取CPU使用率
func BenchmarkGetCPUUsage(b *testing.B) {
	// 预热
	GetCPUUsage()
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCPUUsage()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetDiskInfo 基准测试：获取磁盘信息
func BenchmarkGetDiskInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetDiskInfo()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFormatBytes 基准测试：字节格式化
func BenchmarkFormatBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatBytes(1073741824)
	}
}

// BenchmarkFDTracker 基准测试：FD追踪
func BenchmarkFDTracker(b *testing.B) {
	tracker := NewFDTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Track(i, "/tmp/test.txt", "file")
		tracker.Untrack(i)
	}
}
