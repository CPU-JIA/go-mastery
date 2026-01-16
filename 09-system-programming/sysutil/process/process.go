/*
Package process 提供跨平台的进程管理工具。

本包支持以下功能：
  - 进程列表获取和过滤
  - 进程信息查询（CPU、内存、状态）
  - 进程创建和管理
  - 信号发送（跨平台抽象）
  - 进程树遍历

跨平台支持：
  - Windows: 使用 Windows API 和 WMI
  - Linux/Unix: 使用 /proc 文件系统和系统调用
*/
package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// ===================
// 错误定义
// ===================

var (
	// ErrProcessNotFound 进程未找到
	ErrProcessNotFound = errors.New("process not found")
	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")
	// ErrInvalidPID 无效的进程ID
	ErrInvalidPID = errors.New("invalid process ID")
	// ErrProcessExited 进程已退出
	ErrProcessExited = errors.New("process has exited")
	// ErrTimeout 操作超时
	ErrTimeout = errors.New("operation timed out")
)

// ===================
// 进程信息结构
// ===================

// Info 进程信息结构
// 包含进程的基本信息和资源使用情况
type Info struct {
	// PID 进程ID
	PID int
	// PPID 父进程ID
	PPID int
	// Name 进程名称
	Name string
	// Executable 可执行文件路径
	Executable string
	// CommandLine 完整命令行
	CommandLine string
	// WorkingDir 工作目录
	WorkingDir string
	// Username 运行用户
	Username string
	// State 进程状态
	State ProcessState
	// CreateTime 创建时间
	CreateTime time.Time
	// CPUPercent CPU使用率（百分比）
	CPUPercent float64
	// MemoryInfo 内存使用信息
	MemoryInfo MemoryInfo
	// NumThreads 线程数
	NumThreads int
	// NumFDs 文件描述符数量（仅Linux）
	NumFDs int
	// Nice 优先级（仅Unix）
	Nice int
	// IOCounters I/O计数器
	IOCounters IOCounters
}

// ProcessState 进程状态
type ProcessState int

const (
	// StateUnknown 未知状态
	StateUnknown ProcessState = iota
	// StateRunning 运行中
	StateRunning
	// StateSleeping 睡眠中
	StateSleeping
	// StateStopped 已停止
	StateStopped
	// StateZombie 僵尸进程
	StateZombie
	// StateIdle 空闲
	StateIdle
	// StateWaiting 等待中
	StateWaiting
)

// String 返回状态的字符串表示
func (s ProcessState) String() string {
	states := []string{"Unknown", "Running", "Sleeping", "Stopped", "Zombie", "Idle", "Waiting"}
	if int(s) < len(states) {
		return states[s]
	}
	return "Unknown"
}

// MemoryInfo 内存使用信息
type MemoryInfo struct {
	// RSS 常驻内存大小（字节）
	RSS uint64
	// VMS 虚拟内存大小（字节）
	VMS uint64
	// Shared 共享内存大小（字节）
	Shared uint64
	// Data 数据段大小（字节）
	Data uint64
	// Stack 栈大小（字节）
	Stack uint64
	// Percent 内存使用百分比
	Percent float64
}

// IOCounters I/O计数器
type IOCounters struct {
	// ReadCount 读取次数
	ReadCount uint64
	// WriteCount 写入次数
	WriteCount uint64
	// ReadBytes 读取字节数
	ReadBytes uint64
	// WriteBytes 写入字节数
	WriteBytes uint64
}

// ===================
// 进程管理器
// ===================

// Manager 进程管理器
// 提供进程的查询、监控和管理功能
type Manager struct {
	// cache 进程信息缓存
	cache     map[int]*Info
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

// NewManager 创建新的进程管理器
func NewManager() *Manager {
	return &Manager{
		cache:    make(map[int]*Info),
		cacheTTL: 5 * time.Second, // 默认缓存5秒
	}
}

// SetCacheTTL 设置缓存过期时间
func (m *Manager) SetCacheTTL(ttl time.Duration) {
	m.cacheTTL = ttl
}

// List 获取所有进程列表
// 返回系统中所有进程的信息
func (m *Manager) List() ([]*Info, error) {
	return m.ListWithFilter(nil)
}

// ListWithFilter 使用过滤器获取进程列表
// filter 函数返回 true 表示保留该进程
func (m *Manager) ListWithFilter(filter func(*Info) bool) ([]*Info, error) {
	procs, err := listProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	// 更新缓存
	m.cacheMu.Lock()
	m.cache = make(map[int]*Info)
	for _, p := range procs {
		m.cache[p.PID] = p
	}
	m.cacheTime = time.Now()
	m.cacheMu.Unlock()

	// 应用过滤器
	if filter == nil {
		return procs, nil
	}

	var filtered []*Info
	for _, p := range procs {
		if filter(p) {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}

// Get 获取指定PID的进程信息
func (m *Manager) Get(pid int) (*Info, error) {
	if pid <= 0 {
		return nil, ErrInvalidPID
	}

	// 检查缓存
	m.cacheMu.RLock()
	if time.Since(m.cacheTime) < m.cacheTTL {
		if info, ok := m.cache[pid]; ok {
			m.cacheMu.RUnlock()
			return info, nil
		}
	}
	m.cacheMu.RUnlock()

	// 直接获取进程信息
	return getProcessInfo(pid)
}

// FindByName 按名称查找进程
// 支持部分匹配（包含关系）
func (m *Manager) FindByName(name string) ([]*Info, error) {
	name = strings.ToLower(name)
	return m.ListWithFilter(func(p *Info) bool {
		return strings.Contains(strings.ToLower(p.Name), name)
	})
}

// FindByExecutable 按可执行文件路径查找进程
func (m *Manager) FindByExecutable(path string) ([]*Info, error) {
	path = strings.ToLower(path)
	return m.ListWithFilter(func(p *Info) bool {
		return strings.Contains(strings.ToLower(p.Executable), path)
	})
}

// GetChildren 获取指定进程的子进程
func (m *Manager) GetChildren(pid int) ([]*Info, error) {
	return m.ListWithFilter(func(p *Info) bool {
		return p.PPID == pid
	})
}

// GetTree 获取进程树
// 返回以指定PID为根的进程树
func (m *Manager) GetTree(pid int) (*ProcessTree, error) {
	procs, err := m.List()
	if err != nil {
		return nil, err
	}

	// 构建进程映射
	procMap := make(map[int]*Info)
	for _, p := range procs {
		procMap[p.PID] = p
	}

	// 查找根进程
	root, ok := procMap[pid]
	if !ok {
		return nil, ErrProcessNotFound
	}

	// 递归构建树
	tree := &ProcessTree{
		Process:  root,
		Children: make([]*ProcessTree, 0),
	}

	m.buildTree(tree, procs)
	return tree, nil
}

func (m *Manager) buildTree(node *ProcessTree, procs []*Info) {
	for _, p := range procs {
		if p.PPID == node.Process.PID && p.PID != node.Process.PID {
			child := &ProcessTree{
				Process:  p,
				Children: make([]*ProcessTree, 0),
			}
			node.Children = append(node.Children, child)
			m.buildTree(child, procs)
		}
	}
}

// ProcessTree 进程树结构
type ProcessTree struct {
	Process  *Info
	Children []*ProcessTree
}

// Print 打印进程树
func (t *ProcessTree) Print() string {
	var sb strings.Builder
	t.printNode(&sb, "", true)
	return sb.String()
}

func (t *ProcessTree) printNode(sb *strings.Builder, prefix string, isLast bool) {
	// 打印当前节点
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	sb.WriteString(prefix)
	sb.WriteString(connector)
	sb.WriteString(fmt.Sprintf("[%d] %s\n", t.Process.PID, t.Process.Name))

	// 打印子节点
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range t.Children {
		child.printNode(sb, childPrefix, i == len(t.Children)-1)
	}
}

// ===================
// 进程操作
// ===================

// Kill 终止进程
// 在 Windows 上使用 TerminateProcess
// 在 Unix 上发送 SIGKILL
func Kill(pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := proc.Kill(); err != nil {
		if strings.Contains(err.Error(), "process already finished") {
			return ErrProcessExited
		}
		if strings.Contains(err.Error(), "permission denied") ||
			strings.Contains(err.Error(), "Access is denied") {
			return ErrPermissionDenied
		}
		return fmt.Errorf("failed to kill process: %w", err)
	}

	return nil
}

// Signal 向进程发送信号
// 注意：Windows 仅支持 SIGKILL 和 SIGINT
func Signal(pid int, sig os.Signal) error {
	if pid <= 0 {
		return ErrInvalidPID
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := proc.Signal(sig); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}

// Exists 检查进程是否存在
func Exists(pid int) bool {
	if pid <= 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 在 Unix 上，FindProcess 总是成功的
	// 需要发送信号 0 来检查进程是否存在
	if runtime.GOOS != "windows" {
		err = proc.Signal(nil)
		return err == nil
	}

	// Windows 上 FindProcess 会检查进程是否存在
	return true
}

// WaitForExit 等待进程退出
// 返回进程的退出状态
func WaitForExit(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return ErrInvalidPID
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ErrTimeout
		case <-ticker.C:
			if !Exists(pid) {
				return nil
			}
		}
	}
}

// ===================
// 进程创建
// ===================

// StartOptions 进程启动选项
type StartOptions struct {
	// Dir 工作目录
	Dir string
	// Env 环境变量
	Env []string
	// Stdin 标准输入
	Stdin *os.File
	// Stdout 标准输出
	Stdout *os.File
	// Stderr 标准错误
	Stderr *os.File
	// Detached 是否分离运行（后台进程）
	Detached bool
}

// Start 启动新进程
// 返回进程的 PID
func Start(name string, args []string, opts *StartOptions) (int, error) {
	cmd := exec.Command(name, args...)

	if opts != nil {
		if opts.Dir != "" {
			cmd.Dir = opts.Dir
		}
		if opts.Env != nil {
			cmd.Env = opts.Env
		}
		if opts.Stdin != nil {
			cmd.Stdin = opts.Stdin
		}
		if opts.Stdout != nil {
			cmd.Stdout = opts.Stdout
		}
		if opts.Stderr != nil {
			cmd.Stderr = opts.Stderr
		}
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}

	return cmd.Process.Pid, nil
}

// Run 运行命令并等待完成
// 返回命令的输出和错误
func Run(name string, args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return string(output), ErrTimeout
	}

	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// ===================
// 进程监控
// ===================

// Monitor 进程监控器
// 用于监控进程的状态变化
type Monitor struct {
	pid      int
	interval time.Duration
	stopCh   chan struct{}
	eventCh  chan Event
	running  bool
	mu       sync.Mutex
}

// Event 进程事件
type Event struct {
	Type      EventType
	PID       int
	Timestamp time.Time
	Info      *Info
	Error     error
}

// EventType 事件类型
type EventType int

const (
	// EventStarted 进程启动
	EventStarted EventType = iota
	// EventExited 进程退出
	EventExited
	// EventCPUHigh CPU使用率过高
	EventCPUHigh
	// EventMemoryHigh 内存使用率过高
	EventMemoryHigh
	// EventError 监控错误
	EventError
)

// NewMonitor 创建进程监控器
func NewMonitor(pid int, interval time.Duration) *Monitor {
	return &Monitor{
		pid:      pid,
		interval: interval,
		stopCh:   make(chan struct{}),
		eventCh:  make(chan Event, 100),
	}
}

// Start 启动监控
func (m *Monitor) Start() <-chan Event {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return m.eventCh
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
	return m.eventCh
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)
}

func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	var lastInfo *Info
	wasRunning := false

	for {
		select {
		case <-m.stopCh:
			close(m.eventCh)
			return
		case <-ticker.C:
			info, err := getProcessInfo(m.pid)
			if err != nil {
				if wasRunning {
					// 进程退出
					m.eventCh <- Event{
						Type:      EventExited,
						PID:       m.pid,
						Timestamp: time.Now(),
						Info:      lastInfo,
					}
					wasRunning = false
				}
				continue
			}

			if !wasRunning {
				// 进程启动
				m.eventCh <- Event{
					Type:      EventStarted,
					PID:       m.pid,
					Timestamp: time.Now(),
					Info:      info,
				}
				wasRunning = true
			}

			// 检查资源使用
			if info.CPUPercent > 90 {
				m.eventCh <- Event{
					Type:      EventCPUHigh,
					PID:       m.pid,
					Timestamp: time.Now(),
					Info:      info,
				}
			}

			if info.MemoryInfo.Percent > 90 {
				m.eventCh <- Event{
					Type:      EventMemoryHigh,
					PID:       m.pid,
					Timestamp: time.Now(),
					Info:      info,
				}
			}

			lastInfo = info
		}
	}
}

// ===================
// 工具函数
// ===================

// Self 获取当前进程信息
func Self() (*Info, error) {
	return getProcessInfo(os.Getpid())
}

// Parent 获取父进程信息
func Parent() (*Info, error) {
	return getProcessInfo(os.Getppid())
}

// TopByCPU 获取CPU使用率最高的N个进程
func TopByCPU(n int) ([]*Info, error) {
	mgr := NewManager()
	procs, err := mgr.List()
	if err != nil {
		return nil, err
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].CPUPercent > procs[j].CPUPercent
	})

	if n > len(procs) {
		n = len(procs)
	}

	return procs[:n], nil
}

// TopByMemory 获取内存使用最高的N个进程
func TopByMemory(n int) ([]*Info, error) {
	mgr := NewManager()
	procs, err := mgr.List()
	if err != nil {
		return nil, err
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].MemoryInfo.RSS > procs[j].MemoryInfo.RSS
	})

	if n > len(procs) {
		n = len(procs)
	}

	return procs[:n], nil
}

// Count 获取进程总数
func Count() (int, error) {
	mgr := NewManager()
	procs, err := mgr.List()
	if err != nil {
		return 0, err
	}
	return len(procs), nil
}
