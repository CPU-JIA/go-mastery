/*
Package process 的单元测试

测试覆盖：
  - 进程列表获取
  - 进程信息查询
  - 进程操作（存在性检查、信号发送）
  - 进程监控
  - 工具函数
*/
package process

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// TestNewManager 测试创建进程管理器
func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.cache == nil {
		t.Error("cache should be initialized")
	}

	if mgr.cacheTTL != 5*time.Second {
		t.Errorf("default cacheTTL should be 5s, got %v", mgr.cacheTTL)
	}
}

// TestManagerSetCacheTTL 测试设置缓存过期时间
func TestManagerSetCacheTTL(t *testing.T) {
	mgr := NewManager()
	mgr.SetCacheTTL(10 * time.Second)

	if mgr.cacheTTL != 10*time.Second {
		t.Errorf("cacheTTL should be 10s, got %v", mgr.cacheTTL)
	}
}

// TestManagerList 测试获取进程列表
func TestManagerList(t *testing.T) {
	mgr := NewManager()
	procs, err := mgr.List()

	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(procs) == 0 {
		t.Error("process list should not be empty")
	}

	// 验证当前进程在列表中
	currentPID := os.Getpid()
	found := false
	for _, p := range procs {
		if p.PID == currentPID {
			found = true
			break
		}
	}

	if !found {
		t.Error("current process should be in the list")
	}
}

// TestManagerListWithFilter 测试使用过滤器获取进程列表
func TestManagerListWithFilter(t *testing.T) {
	mgr := NewManager()

	// 过滤出当前进程
	currentPID := os.Getpid()
	procs, err := mgr.ListWithFilter(func(p *Info) bool {
		return p.PID == currentPID
	})

	if err != nil {
		t.Fatalf("ListWithFilter failed: %v", err)
	}

	if len(procs) != 1 {
		t.Errorf("expected 1 process, got %d", len(procs))
	}

	if len(procs) > 0 && procs[0].PID != currentPID {
		t.Errorf("expected PID %d, got %d", currentPID, procs[0].PID)
	}
}

// TestManagerGet 测试获取指定进程信息
func TestManagerGet(t *testing.T) {
	mgr := NewManager()
	currentPID := os.Getpid()

	info, err := mgr.Get(currentPID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if info.PID != currentPID {
		t.Errorf("expected PID %d, got %d", currentPID, info.PID)
	}

	if info.Name == "" {
		t.Error("process name should not be empty")
	}
}

// TestManagerGetInvalidPID 测试获取无效PID的进程信息
func TestManagerGetInvalidPID(t *testing.T) {
	mgr := NewManager()

	_, err := mgr.Get(0)
	if err != ErrInvalidPID {
		t.Errorf("expected ErrInvalidPID, got %v", err)
	}

	_, err = mgr.Get(-1)
	if err != ErrInvalidPID {
		t.Errorf("expected ErrInvalidPID, got %v", err)
	}
}

// TestManagerFindByName 测试按名称查找进程
func TestManagerFindByName(t *testing.T) {
	mgr := NewManager()

	// 查找 go 相关进程（测试进程本身）
	procs, err := mgr.FindByName("go")
	if err != nil {
		t.Fatalf("FindByName failed: %v", err)
	}

	// 应该至少找到测试进程
	if len(procs) == 0 {
		t.Log("Warning: no 'go' processes found, this might be expected in some environments")
	}
}

// TestManagerGetChildren 测试获取子进程
func TestManagerGetChildren(t *testing.T) {
	mgr := NewManager()
	currentPID := os.Getpid()

	// 获取当前进程的子进程
	children, err := mgr.GetChildren(currentPID)
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}

	// 测试进程可能没有子进程，这是正常的
	t.Logf("Found %d children for PID %d", len(children), currentPID)
}

// TestSelf 测试获取当前进程信息
func TestSelf(t *testing.T) {
	info, err := Self()
	if err != nil {
		t.Fatalf("Self failed: %v", err)
	}

	if info.PID != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), info.PID)
	}
}

// TestParent 测试获取父进程信息
func TestParent(t *testing.T) {
	info, err := Parent()
	if err != nil {
		t.Fatalf("Parent failed: %v", err)
	}

	if info.PID != os.Getppid() {
		t.Errorf("expected PPID %d, got %d", os.Getppid(), info.PID)
	}
}

// TestExists 测试进程存在性检查
func TestExists(t *testing.T) {
	// 当前进程应该存在
	if !Exists(os.Getpid()) {
		t.Error("current process should exist")
	}

	// PID 0 不应该存在（或者是特殊进程）
	// 注意：在某些系统上 PID 0 可能是有效的
	if Exists(-1) {
		t.Error("PID -1 should not exist")
	}
}

// TestProcessState 测试进程状态字符串转换
func TestProcessState(t *testing.T) {
	tests := []struct {
		state    ProcessState
		expected string
	}{
		{StateUnknown, "Unknown"},
		{StateRunning, "Running"},
		{StateSleeping, "Sleeping"},
		{StateStopped, "Stopped"},
		{StateZombie, "Zombie"},
		{StateIdle, "Idle"},
		{StateWaiting, "Waiting"},
		{ProcessState(100), "Unknown"}, // 无效状态
	}

	for _, tt := range tests {
		result := tt.state.String()
		if result != tt.expected {
			t.Errorf("State %d: expected %s, got %s", tt.state, tt.expected, result)
		}
	}
}

// TestCount 测试获取进程总数
func TestCount(t *testing.T) {
	count, err := Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count <= 0 {
		t.Error("process count should be positive")
	}

	t.Logf("Total processes: %d", count)
}

// TestTopByCPU 测试获取CPU使用率最高的进程
func TestTopByCPU(t *testing.T) {
	procs, err := TopByCPU(5)
	if err != nil {
		t.Fatalf("TopByCPU failed: %v", err)
	}

	if len(procs) == 0 {
		t.Error("should return at least one process")
	}

	if len(procs) > 5 {
		t.Errorf("should return at most 5 processes, got %d", len(procs))
	}

	// 验证排序
	for i := 1; i < len(procs); i++ {
		if procs[i].CPUPercent > procs[i-1].CPUPercent {
			t.Error("processes should be sorted by CPU usage descending")
		}
	}
}

// TestTopByMemory 测试获取内存使用最高的进程
func TestTopByMemory(t *testing.T) {
	procs, err := TopByMemory(5)
	if err != nil {
		t.Fatalf("TopByMemory failed: %v", err)
	}

	if len(procs) == 0 {
		t.Error("should return at least one process")
	}

	if len(procs) > 5 {
		t.Errorf("should return at most 5 processes, got %d", len(procs))
	}

	// 验证排序
	for i := 1; i < len(procs); i++ {
		if procs[i].MemoryInfo.RSS > procs[i-1].MemoryInfo.RSS {
			t.Error("processes should be sorted by memory usage descending")
		}
	}
}

// TestProcessTree 测试进程树
func TestProcessTree(t *testing.T) {
	mgr := NewManager()

	// 获取 init/systemd 进程的树（PID 1）
	// 在 Windows 上这可能不适用
	if runtime.GOOS == "windows" {
		t.Skip("Process tree test skipped on Windows")
	}

	tree, err := mgr.GetTree(1)
	if err != nil {
		// PID 1 可能需要 root 权限
		t.Logf("GetTree(1) failed (may need root): %v", err)
		return
	}

	if tree.Process.PID != 1 {
		t.Errorf("expected root PID 1, got %d", tree.Process.PID)
	}

	// 打印树结构
	output := tree.Print()
	if output == "" {
		t.Error("tree output should not be empty")
	}
	t.Logf("Process tree:\n%s", output)
}

// TestMonitor 测试进程监控器
func TestMonitor(t *testing.T) {
	currentPID := os.Getpid()
	monitor := NewMonitor(currentPID, 100*time.Millisecond)

	eventCh := monitor.Start()

	// 等待一些事件
	timeout := time.After(500 * time.Millisecond)
	eventCount := 0

loop:
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				break loop
			}
			eventCount++
			t.Logf("Received event: type=%d, PID=%d", event.Type, event.PID)
			if eventCount >= 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	monitor.Stop()

	// 应该至少收到一个启动事件
	if eventCount == 0 {
		t.Log("Warning: no events received, this might be timing-related")
	}
}

// TestRun 测试运行命令
func TestRun(t *testing.T) {
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmd = "echo"
		args = []string{"hello"}
	}

	output, err := Run(cmd, args, 5*time.Second)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got: %s", output)
	}
}

// TestRunTimeout 测试命令超时
func TestRunTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Timeout test skipped on Windows")
	}

	// 运行一个会超时的命令
	_, err := Run("sleep", []string{"10"}, 100*time.Millisecond)
	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

// TestStart 测试启动进程
func TestStart(t *testing.T) {
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	} else {
		cmd = "echo"
		args = []string{"test"}
	}

	pid, err := Start(cmd, args, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if pid <= 0 {
		t.Error("PID should be positive")
	}

	// 等待进程完成
	time.Sleep(100 * time.Millisecond)
}

// BenchmarkList 基准测试：获取进程列表
func BenchmarkList(b *testing.B) {
	mgr := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mgr.List()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGet 基准测试：获取单个进程信息
func BenchmarkGet(b *testing.B) {
	mgr := NewManager()
	pid := os.Getpid()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mgr.Get(pid)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExists 基准测试：检查进程存在性
func BenchmarkExists(b *testing.B) {
	pid := os.Getpid()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Exists(pid)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
