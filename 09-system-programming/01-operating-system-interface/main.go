/*
=== Go系统编程：操作系统接口掌控 ===

本模块专注于深度掌握Go语言的系统编程技术，探索：
1. 系统调用接口封装和优化
2. 进程管理和生命周期控制
3. 信号处理和中断管理
4. 内存映射和虚拟内存操作
5. 文件描述符和I/O多路复用
6. 进程间通信(IPC)机制
7. 系统资源监控和管理
8. 线程和协程系统级调优
9. 系统调用性能优化
10. 跨平台系统编程抽象

学习目标：
- 掌握操作系统底层接口的Go实现
- 理解系统调用的性能特征和优化
- 学会进程间通信的各种机制
- 掌握系统资源的监控和管理
*/

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// Windows compatible constants
const (
	SIGUSR1     = syscall.Signal(10) // User signal 1
	PROT_READ   = 1
	PROT_WRITE  = 2
	MAP_SHARED  = 1
	MAP_PRIVATE = 2

	// Path and security constants
	MaxPathLength       = 260   // Windows maximum path length
	MaxArgumentLength   = 1024  // Maximum argument length for security
	DefaultPermission   = 0o600 // Safe file permission
	DefaultPageSize     = 4096  // Standard page size
	DefaultTimeoutSec   = 10    // Default timeout in seconds
	DefaultTickerSec    = 1     // Default ticker interval
	MaxSemaphoreValue   = 5     // Maximum semaphore value
	MaxMessageQueueSize = 10    // Maximum message queue size
	MaxProcessDisplay   = 5     // Maximum processes to display

	// System call testing constants
	TestLoopCount         = 10
	MicrosecondsBase      = 100
	MicrosecondsIncrement = 10
	ReadMicroseconds      = 50
	ReadIncrement         = 5
	HashMultiplier        = 31
	ProcessTimeoutSec     = 30
	KBToBytes             = 1024
	SecondsToHours        = 3600
)

// ==================
// 1. 系统调用接口封装
// ==================

// SystemCallWrapper 系统调用包装器
type SystemCallWrapper struct {
	callCounts    map[string]int64
	callDurations map[string]time.Duration
	errorCounts   map[string]int64
	mutex         sync.RWMutex
}

// SystemCallInfo 系统调用信息
type SystemCallInfo struct {
	Name     string
	Number   uintptr
	Args     []uintptr
	Result   uintptr
	Error    error
	Duration time.Duration
	Caller   string
}

// SystemResourceMonitor 系统资源监控器
type SystemResourceMonitor struct {
	processes   map[int]*ProcessInfo
	memory      MemoryInfo
	fileHandles map[int]*FileHandle
	network     NetworkInfo
	mutex       sync.RWMutex
	running     bool
	stopCh      chan struct{}
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID         int
	PPID        int
	Name        string
	State       ProcessState
	StartTime   time.Time
	CPUTime     time.Duration
	MemoryUsage int64
	FileHandles []int
	Threads     int
	Priority    int
	Nice        int
	Command     string
	Arguments   []string
	Environment map[string]string
}

// ProcessState 进程状态
type ProcessState int

const (
	ProcessRunning ProcessState = iota
	ProcessSleeping
	ProcessWaiting
	ProcessZombie
	ProcessStopped
)

func (ps ProcessState) String() string {
	states := []string{"Running", "Sleeping", "Waiting", "Zombie", "Stopped"}
	if int(ps) < len(states) {
		return states[ps]
	}
	return "Unknown"
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	TotalPhysical     uint64
	AvailablePhysical uint64
	UsedPhysical      uint64
	TotalVirtual      uint64
	AvailableVirtual  uint64
	UsedVirtual       uint64
	PageSize          uint64
	PageFaults        uint64
	SwapTotal         uint64
	SwapUsed          uint64
}

// FileHandle 文件句柄信息
type FileHandle struct {
	FD       int
	Path     string
	Mode     os.FileMode
	Position int64
	Process  int
	Type     FileType
	Flags    int
}

// FileType 文件类型
type FileType int

const (
	FileTypeRegular FileType = iota
	FileTypeDirectory
	FileTypeSocket
	FileTypePipe
	FileTypeDevice
	FileTypeSymlink
)

// NetworkInfo 网络信息
type NetworkInfo struct {
	Connections []NetworkConnection
	Interfaces  []NetworkInterface
	Statistics  NetworkStatistics
}

// NetworkConnection 网络连接
type NetworkConnection struct {
	Protocol    string
	LocalAddr   string
	LocalPort   int
	RemoteAddr  string
	RemotePort  int
	State       ConnectionState
	PID         int
	ProcessName string
}

// ConnectionState 连接状态
type ConnectionState int

const (
	ConnEstablished ConnectionState = iota
	ConnListen
	ConnClosing
	ConnClosed
)

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name         string
	Index        int
	MTU          int
	Flags        net.Flags
	HardwareAddr net.HardwareAddr
	Addresses    []net.Addr
	Statistics   InterfaceStatistics
}

// InterfaceStatistics 接口统计
type InterfaceStatistics struct {
	BytesSent     uint64
	BytesReceived uint64
	PacketsSent   uint64
	PacketsRecv   uint64
	ErrorsIn      uint64
	ErrorsOut     uint64
	DroppedIn     uint64
	DroppedOut    uint64
}

// NetworkStatistics 网络统计
type NetworkStatistics struct {
	TCPConnections     int
	UDPConnections     int
	ActiveConnections  int
	PassiveConnections int
	FailedConnections  int
	ResetConnections   int
}

func NewSystemCallWrapper() *SystemCallWrapper {
	return &SystemCallWrapper{
		callCounts:    make(map[string]int64),
		callDurations: make(map[string]time.Duration),
		errorCounts:   make(map[string]int64),
	}
}

func (scw *SystemCallWrapper) WrapSyscall(name string, fn func() (uintptr, uintptr, error)) (uintptr, uintptr, error) {
	start := time.Now()

	r1, r2, err := fn()

	duration := time.Since(start)

	scw.mutex.Lock()
	scw.callCounts[name]++
	scw.callDurations[name] += duration
	if err != nil {
		scw.errorCounts[name]++
	}
	scw.mutex.Unlock()

	return r1, r2, err
}

func (scw *SystemCallWrapper) GetStatistics() map[string]SystemCallStats {
	scw.mutex.RLock()
	defer scw.mutex.RUnlock()

	stats := make(map[string]SystemCallStats)
	for name, count := range scw.callCounts {
		stats[name] = SystemCallStats{
			Name:        name,
			CallCount:   count,
			TotalTime:   scw.callDurations[name],
			AverageTime: scw.callDurations[name] / time.Duration(count),
			ErrorCount:  scw.errorCounts[name],
			ErrorRate:   float64(scw.errorCounts[name]) / float64(count),
		}
	}

	return stats
}

// SystemCallStats 系统调用统计
type SystemCallStats struct {
	Name        string
	CallCount   int64
	TotalTime   time.Duration
	AverageTime time.Duration
	ErrorCount  int64
	ErrorRate   float64
}

// ==================
// 2. 进程管理系统
// ==================

// ProcessManager 进程管理器
type ProcessManager struct {
	processes map[int]*ManagedProcess
	mutex     sync.RWMutex
}

// ManagedProcess 被管理的进程
type ManagedProcess struct {
	*os.Process
	Command     *exec.Cmd
	StartTime   time.Time
	Status      ProcessStatus
	LastCPU     time.Duration
	LastMemory  int64
	Restarts    int
	MaxRestarts int
	AutoRestart bool
	HealthCheck func() bool
}

// ProcessStatus 进程状态
type ProcessStatus struct {
	State     ProcessState
	ExitCode  int
	Signal    os.Signal
	Timestamp time.Time
}

// G204安全修复：输入验证函数防止命令注入
func validateExecutablePath(path string) error {
	// 清理路径，移除相对路径引用
	cleanPath := filepath.Clean(path)

	// 检查路径是否包含目录遍历序列
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("可执行文件路径不能包含目录遍历序列")
	}

	// 检查路径长度，防止缓冲区溢出
	if len(cleanPath) > MaxPathLength {
		return fmt.Errorf("可执行文件路径过长")
	}

	// 使用正则表达式验证路径格式（允许字母数字、路径分隔符、点、连字符、下划线）
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\\/._-]+$`, cleanPath)
	if !matched {
		return fmt.Errorf("可执行文件路径包含非法字符")
	}

	return nil
}

func validateProcessArgs(args []string) error {
	// 危险字符列表，防止shell命令注入
	dangerousChars := []string{";", "|", "&", "`", "$", ">", "<", "&&", "||", "(", ")", "{", "}"}

	for i, arg := range args {
		// 检查参数长度
		if len(arg) > MaxArgumentLength {
			return fmt.Errorf("参数 %d 过长", i)
		}

		// 检查危险字符
		for _, dangerous := range dangerousChars {
			if strings.Contains(arg, dangerous) {
				return fmt.Errorf("参数 %d 包含危险字符: %s", i, dangerous)
			}
		}
	}

	return nil
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[int]*ManagedProcess),
	}
}

func (pm *ProcessManager) StartProcess(name string, args []string, config ProcessConfig) (*ManagedProcess, error) {
	// G204安全修复：在执行命令前验证输入
	if err := validateExecutablePath(name); err != nil {
		return nil, fmt.Errorf("可执行文件路径验证失败: %v", err)
	}

	if err := validateProcessArgs(args); err != nil {
		return nil, fmt.Errorf("命令参数验证失败: %v", err)
	}

	cmd := exec.Command(name, args...)

	// 配置进程属性
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	if config.Environment != nil {
		cmd.Env = append(os.Environ(), config.Environment...)
	}

	// 配置标准输入输出
	if config.Stdin != nil {
		cmd.Stdin = config.Stdin
	}
	if config.Stdout != nil {
		cmd.Stdout = config.Stdout
	}
	if config.Stderr != nil {
		cmd.Stderr = config.Stderr
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %v", err)
	}

	managedProc := &ManagedProcess{
		Process:     cmd.Process,
		Command:     cmd,
		StartTime:   time.Now(),
		MaxRestarts: config.MaxRestarts,
		AutoRestart: config.AutoRestart,
		HealthCheck: config.HealthCheck,
		Status: ProcessStatus{
			State:     ProcessRunning,
			Timestamp: time.Now(),
		},
	}

	pm.mutex.Lock()
	pm.processes[cmd.Process.Pid] = managedProc
	pm.mutex.Unlock()

	// 启动监控
	go pm.monitorProcess(managedProc)

	return managedProc, nil
}

// ProcessConfig 进程配置
type ProcessConfig struct {
	WorkingDir  string
	Environment []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	MaxRestarts int
	AutoRestart bool
	HealthCheck func() bool
	Timeout     time.Duration
	KillTimeout time.Duration
}

func (pm *ProcessManager) monitorProcess(proc *ManagedProcess) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查进程状态
			if !pm.isProcessAlive(proc.Pid) {
				proc.Status.State = ProcessStopped
				proc.Status.Timestamp = time.Now()

				if proc.AutoRestart && proc.Restarts < proc.MaxRestarts {
					pm.restartProcess(proc)
				}
				return
			}

			// 健康检查
			if proc.HealthCheck != nil && !proc.HealthCheck() {
				log.Printf("Process %d failed health check", proc.Pid)
				if proc.AutoRestart {
					pm.restartProcess(proc)
				}
			}
		}
	}
}

func (pm *ProcessManager) isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送信号0检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (pm *ProcessManager) restartProcess(proc *ManagedProcess) {
	log.Printf("Restarting process %d (restart count: %d)", proc.Pid, proc.Restarts)

	// 杀死旧进程
	proc.Kill()

	// 重新启动
	newCmd := exec.Command(proc.Command.Path, proc.Command.Args[1:]...)
	newCmd.Dir = proc.Command.Dir
	newCmd.Env = proc.Command.Env
	newCmd.Stdin = proc.Command.Stdin
	newCmd.Stdout = proc.Command.Stdout
	newCmd.Stderr = proc.Command.Stderr

	if err := newCmd.Start(); err != nil {
		log.Printf("Failed to restart process: %v", err)
		return
	}

	pm.mutex.Lock()
	delete(pm.processes, proc.Pid)

	newProc := &ManagedProcess{
		Process:     newCmd.Process,
		Command:     newCmd,
		StartTime:   time.Now(),
		MaxRestarts: proc.MaxRestarts,
		AutoRestart: proc.AutoRestart,
		HealthCheck: proc.HealthCheck,
		Restarts:    proc.Restarts + 1,
		Status: ProcessStatus{
			State:     ProcessRunning,
			Timestamp: time.Now(),
		},
	}

	pm.processes[newCmd.Process.Pid] = newProc
	pm.mutex.Unlock()

	go pm.monitorProcess(newProc)
}

func (pm *ProcessManager) StopProcess(pid int, graceful bool) error {
	pm.mutex.RLock()
	proc, exists := pm.processes[pid]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("process %d not found", pid)
	}

	if graceful {
		// 优雅停止：先发送SIGTERM，等待一段时间后强制杀死
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			return err
		}

		// 等待进程自然退出
		done := make(chan error, 1)
		go func() {
			_, err := proc.Wait()
			done <- err
		}()

		select {
		case err := <-done:
			return err
		case <-time.After(10 * time.Second):
			// 超时，强制杀死
			return proc.Kill()
		}
	} else {
		return proc.Kill()
	}
}

func (pm *ProcessManager) GetProcesses() map[int]*ManagedProcess {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[int]*ManagedProcess)
	for pid, proc := range pm.processes {
		result[pid] = proc
	}

	return result
}

// ==================
// 3. 信号处理系统
// ==================

// SignalHandler 信号处理器
type SignalHandler struct {
	handlers map[os.Signal][]SignalCallback
	mutex    sync.RWMutex
	running  bool
	stopCh   chan struct{}
}

// SignalCallback 信号回调函数
type SignalCallback func(signal os.Signal, context SignalContext)

// SignalContext 信号上下文
type SignalContext struct {
	Signal    os.Signal
	Timestamp time.Time
	PID       int
	Info      map[string]interface{}
}

func NewSignalHandler() *SignalHandler {
	return &SignalHandler{
		handlers: make(map[os.Signal][]SignalCallback),
		stopCh:   make(chan struct{}),
	}
}

func (sh *SignalHandler) RegisterHandler(signal os.Signal, callback SignalCallback) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.handlers[signal] = append(sh.handlers[signal], callback)
}

func (sh *SignalHandler) Start() {
	sh.mutex.Lock()
	if sh.running {
		sh.mutex.Unlock()
		return
	}
	sh.running = true
	sh.mutex.Unlock()

	// 收集所有需要监听的信号
	var signals []os.Signal
	sh.mutex.RLock()
	for sig := range sh.handlers {
		signals = append(signals, sig)
	}
	sh.mutex.RUnlock()

	// 创建信号通道
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, signals...)

	go func() {
		for {
			select {
			case sig := <-signalCh:
				sh.handleSignal(sig)
			case <-sh.stopCh:
				signal.Stop(signalCh)
				return
			}
		}
	}()
}

func (sh *SignalHandler) handleSignal(sig os.Signal) {
	context := SignalContext{
		Signal:    sig,
		Timestamp: time.Now(),
		PID:       os.Getpid(),
		Info:      make(map[string]interface{}),
	}

	sh.mutex.RLock()
	callbacks := sh.handlers[sig]
	sh.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(sig, context)
	}
}

func (sh *SignalHandler) Stop() {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if !sh.running {
		return
	}

	sh.running = false
	close(sh.stopCh)
}

// ==================
// 4. 内存映射管理
// ==================

// MemoryMapper 内存映射管理器
type MemoryMapper struct {
	mappings map[uintptr]*MemoryMapping
	mutex    sync.RWMutex
}

// MemoryMapping 内存映射
type MemoryMapping struct {
	Address    uintptr
	Size       uintptr
	Protection int
	Flags      int
	File       *os.File
	Offset     int64
	CreateTime time.Time
	AccessTime time.Time
	PageFaults int64
	data       []byte // Added for Windows compatibility
}

func NewMemoryMapper() *MemoryMapper {
	return &MemoryMapper{
		mappings: make(map[uintptr]*MemoryMapping),
	}
}

func (mm *MemoryMapper) MapFile(file *os.File, offset int64, size uintptr, protection, flags int) (*MemoryMapping, error) {
	// 模拟内存映射实现（跨平台兼容）
	// 实际实现中会使用平台特定的内存映射API
	buffer := make([]byte, size)

	// 读取文件内容到缓冲区
	_, err := file.ReadAt(buffer, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for mapping: %v", err)
	}

	// 模拟内存地址（实际应该是真实的内存地址）
	addr := uintptr(unsafe.Pointer(&buffer[0]))

	mapping := &MemoryMapping{
		Address:    addr,
		Size:       size,
		Protection: protection,
		Flags:      flags,
		File:       file,
		Offset:     offset,
		CreateTime: time.Now(),
		AccessTime: time.Now(),
	}

	mm.mutex.Lock()
	mm.mappings[addr] = mapping
	mm.mutex.Unlock()

	return mapping, nil
}

func (mm *MemoryMapper) MapAnonymous(size uintptr, protection, flags int) (*MemoryMapping, error) {
	// Windows compatible implementation using memory allocation
	slice := make([]byte, size)
	addr := uintptr(unsafe.Pointer(&slice[0]))

	mapping := &MemoryMapping{
		Address:    addr,
		Size:       size,
		Protection: protection,
		Flags:      flags,
		Offset:     0,
		data:       slice, // Keep reference to prevent GC
	}

	mm.mappings[addr] = mapping
	return mapping, nil
}

func (mm *MemoryMapper) Unmap(addr uintptr) error {
	mm.mutex.Lock()
	mapping, exists := mm.mappings[addr]
	if !exists {
		mm.mutex.Unlock()
		return fmt.Errorf("mapping not found at address %x", addr)
	}

	delete(mm.mappings, addr)
	mm.mutex.Unlock()

	// Log the unmapped memory information
	_ = mapping // Use the mapping variable to avoid unused variable error

	// Windows compatible implementation - memory is automatically freed by GC
	// when the slice reference is removed
	return nil
}

func (mm *MemoryMapper) ProtectMemory(addr uintptr, size uintptr, protection int) error {
	// Windows compatible implementation - memory protection simulation
	mm.mutex.Lock()
	if mapping, exists := mm.mappings[addr]; exists {
		mapping.Protection = protection
	}
	mm.mutex.Unlock()

	return nil
}

func (mm *MemoryMapper) GetMappings() map[uintptr]*MemoryMapping {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make(map[uintptr]*MemoryMapping)
	for addr, mapping := range mm.mappings {
		result[addr] = mapping
	}

	return result
}

// ==================
// 5. 进程间通信(IPC)
// ==================

// IPCManager 进程间通信管理器
type IPCManager struct {
	pipes         map[string]*NamedPipe
	sharedMem     map[string]*SharedMemory
	semaphores    map[string]*Semaphore
	messageQueues map[string]*MessageQueue
	mutex         sync.RWMutex
}

// NamedPipe 命名管道
type NamedPipe struct {
	Name    string
	Path    string
	Mode    os.FileMode
	ReadFD  *os.File
	WriteFD *os.File
	Created time.Time
}

// SharedMemory 共享内存
type SharedMemory struct {
	Name    string
	Key     int
	Size    uintptr
	Address uintptr
	Mode    int
	Created time.Time
	data    []byte // Added for Windows compatibility
}

// Semaphore 信号量
type Semaphore struct {
	Name     string
	Key      int
	Value    int32
	MaxValue int32
	Waiters  int32
	Created  time.Time
}

// MessageQueue 消息队列
type MessageQueue struct {
	Name     string
	Key      int
	MaxSize  int
	Messages []Message
	Readers  int
	Writers  int
	Created  time.Time
	mutex    sync.RWMutex
}

// Message 消息
type Message struct {
	Type      int64
	Data      []byte
	Sender    int
	Timestamp time.Time
}

func NewIPCManager() *IPCManager {
	return &IPCManager{
		pipes:         make(map[string]*NamedPipe),
		sharedMem:     make(map[string]*SharedMemory),
		semaphores:    make(map[string]*Semaphore),
		messageQueues: make(map[string]*MessageQueue),
	}
}

func (ipc *IPCManager) CreateNamedPipe(name, path string, mode os.FileMode) (*NamedPipe, error) {
	// Windows compatible implementation - named pipes not directly supported
	// Create a regular file as placeholder
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe file: %v", err)
	}
	file.Close()

	pipe := &NamedPipe{
		Name:    name,
		Path:    path,
		Mode:    mode,
		Created: time.Now(),
	}

	ipc.mutex.Lock()
	ipc.pipes[name] = pipe
	ipc.mutex.Unlock()

	return pipe, nil
}

func (ipc *IPCManager) OpenPipe(name string, forWriting bool) error {
	ipc.mutex.RLock()
	pipe, exists := ipc.pipes[name]
	ipc.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pipe %s not found", name)
	}

	var err error
	if forWriting {
		pipe.WriteFD, err = os.OpenFile(pipe.Path, os.O_WRONLY, pipe.Mode)
	} else {
		pipe.ReadFD, err = os.OpenFile(pipe.Path, os.O_RDONLY, pipe.Mode)
	}

	return err
}

func (ipc *IPCManager) CreateSharedMemory(name string, size uintptr, mode int) (*SharedMemory, error) {
	// Windows compatible implementation using regular memory allocation
	key := ipc.generateKey(name)

	// Create memory segment using Go slice
	slice := make([]byte, size)
	addr := uintptr(unsafe.Pointer(&slice[0]))

	shm := &SharedMemory{
		Name:    name,
		Key:     key,
		Size:    size,
		Address: addr,
		Mode:    mode,
		Created: time.Now(),
		data:    slice, // Keep reference to prevent GC
	}

	ipc.mutex.Lock()
	ipc.sharedMem[name] = shm
	ipc.mutex.Unlock()

	return shm, nil
}

func (ipc *IPCManager) CreateSemaphore(name string, initialValue, maxValue int32) (*Semaphore, error) {
	key := ipc.generateKey(name)

	sem := &Semaphore{
		Name:     name,
		Key:      key,
		Value:    initialValue,
		MaxValue: maxValue,
		Created:  time.Now(),
	}

	ipc.mutex.Lock()
	ipc.semaphores[name] = sem
	ipc.mutex.Unlock()

	return sem, nil
}

func (ipc *IPCManager) Wait(semName string) error {
	ipc.mutex.RLock()
	sem, exists := ipc.semaphores[semName]
	ipc.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("semaphore %s not found", semName)
	}

	// 原子减操作
	for {
		oldValue := atomic.LoadInt32(&sem.Value)
		if oldValue > 0 {
			if atomic.CompareAndSwapInt32(&sem.Value, oldValue, oldValue-1) {
				return nil
			}
		} else {
			// 需要等待
			atomic.AddInt32(&sem.Waiters, 1)
			time.Sleep(time.Millisecond) // 简化的等待逻辑
			atomic.AddInt32(&sem.Waiters, -1)
		}
	}
}

func (ipc *IPCManager) Signal(semName string) error {
	ipc.mutex.RLock()
	sem, exists := ipc.semaphores[semName]
	ipc.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("semaphore %s not found", semName)
	}

	// 原子增操作
	for {
		oldValue := atomic.LoadInt32(&sem.Value)
		if oldValue < sem.MaxValue {
			if atomic.CompareAndSwapInt32(&sem.Value, oldValue, oldValue+1) {
				return nil
			}
		} else {
			return fmt.Errorf("semaphore %s is at maximum value", semName)
		}
	}
}

func (ipc *IPCManager) CreateMessageQueue(name string, maxSize int) (*MessageQueue, error) {
	key := ipc.generateKey(name)

	mq := &MessageQueue{
		Name:     name,
		Key:      key,
		MaxSize:  maxSize,
		Messages: make([]Message, 0),
		Created:  time.Now(),
	}

	ipc.mutex.Lock()
	ipc.messageQueues[name] = mq
	ipc.mutex.Unlock()

	return mq, nil
}

func (ipc *IPCManager) SendMessage(queueName string, msgType int64, data []byte) error {
	ipc.mutex.RLock()
	mq, exists := ipc.messageQueues[queueName]
	ipc.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("message queue %s not found", queueName)
	}

	message := Message{
		Type:      msgType,
		Data:      data,
		Sender:    os.Getpid(),
		Timestamp: time.Now(),
	}

	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.Messages) >= mq.MaxSize {
		return fmt.Errorf("message queue %s is full", queueName)
	}

	mq.Messages = append(mq.Messages, message)
	return nil
}

func (ipc *IPCManager) ReceiveMessage(queueName string, msgType int64) (*Message, error) {
	ipc.mutex.RLock()
	mq, exists := ipc.messageQueues[queueName]
	ipc.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("message queue %s not found", queueName)
	}

	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	for i, msg := range mq.Messages {
		if msgType == 0 || msg.Type == msgType {
			// 删除消息并返回
			mq.Messages = append(mq.Messages[:i], mq.Messages[i+1:]...)
			return &msg, nil
		}
	}

	return nil, fmt.Errorf("no message of type %d found", msgType)
}

func (ipc *IPCManager) generateKey(name string) int {
	// 简单的键生成算法
	hash := 0
	for _, c := range name {
		hash = hash*HashMultiplier + int(c)
	}
	return hash & 0x7FFFFFFF // 确保是正数
}

// ==================
// 6. 系统资源监控
// ==================

func NewSystemResourceMonitor() *SystemResourceMonitor {
	return &SystemResourceMonitor{
		processes:   make(map[int]*ProcessInfo),
		fileHandles: make(map[int]*FileHandle),
		stopCh:      make(chan struct{}),
	}
}

func (srm *SystemResourceMonitor) Start() {
	srm.mutex.Lock()
	if srm.running {
		srm.mutex.Unlock()
		return
	}
	srm.running = true
	srm.mutex.Unlock()

	go srm.monitorLoop()
}

func (srm *SystemResourceMonitor) Stop() {
	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	if !srm.running {
		return
	}

	srm.running = false
	close(srm.stopCh)
}

func (srm *SystemResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			srm.updateSystemInfo()
		case <-srm.stopCh:
			return
		}
	}
}

func (srm *SystemResourceMonitor) updateSystemInfo() {
	// 更新内存信息
	srm.updateMemoryInfo()

	// 更新进程信息
	srm.updateProcessInfo()

	// 更新网络信息
	srm.updateNetworkInfo()

	// 更新文件句柄信息
	srm.updateFileHandleInfo()
}

func (srm *SystemResourceMonitor) updateMemoryInfo() {
	// 读取 /proc/meminfo
	if runtime.GOOS == "linux" {
		srm.updateLinuxMemoryInfo()
	} else {
		srm.updateGenericMemoryInfo()
	}
}

func (srm *SystemResourceMonitor) updateLinuxMemoryInfo() {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	memInfo := &srm.memory

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// 转换为字节 (从kB)
		if len(fields) > 2 && fields[2] == "kB" {
			value *= KBToBytes
		}

		switch name {
		case "MemTotal":
			memInfo.TotalPhysical = value
		case "MemAvailable":
			memInfo.AvailablePhysical = value
		case "SwapTotal":
			memInfo.SwapTotal = value
		case "SwapFree":
			memInfo.SwapUsed = memInfo.SwapTotal - value
		}
	}

	memInfo.UsedPhysical = memInfo.TotalPhysical - memInfo.AvailablePhysical
}

func (srm *SystemResourceMonitor) updateGenericMemoryInfo() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	srm.memory = MemoryInfo{
		TotalPhysical:     m.Sys,
		AvailablePhysical: m.Sys - m.Alloc,
		UsedPhysical:      m.Alloc,
		PageSize:          DefaultPageSize, // 默认页面大小
	}
}

func (srm *SystemResourceMonitor) updateProcessInfo() {
	if runtime.GOOS == "linux" {
		srm.updateLinuxProcessInfo()
	} else {
		srm.updateGenericProcessInfo()
	}
}

func (srm *SystemResourceMonitor) updateLinuxProcessInfo() {
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return
	}

	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		processInfo := srm.readLinuxProcessInfo(pid)
		if processInfo != nil {
			srm.processes[pid] = processInfo
		}
	}
}

func (srm *SystemResourceMonitor) readLinuxProcessInfo(pid int) *ProcessInfo {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statContent, err := os.ReadFile(statPath)
	if err != nil {
		return nil
	}

	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineContent, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return nil
	}

	fields := strings.Fields(string(statContent))
	if len(fields) < 20 {
		return nil
	}

	ppid, _ := strconv.Atoi(fields[3])

	processInfo := &ProcessInfo{
		PID:       pid,
		PPID:      ppid,
		Name:      strings.Trim(fields[1], "()"),
		Command:   string(cmdlineContent),
		Arguments: strings.Split(string(cmdlineContent), "\x00"),
	}

	// 解析状态
	switch fields[2] {
	case "R":
		processInfo.State = ProcessRunning
	case "S":
		processInfo.State = ProcessSleeping
	case "Z":
		processInfo.State = ProcessZombie
	case "T":
		processInfo.State = ProcessStopped
	default:
		processInfo.State = ProcessWaiting
	}

	return processInfo
}

func (srm *SystemResourceMonitor) updateGenericProcessInfo() {
	// 通用进程信息更新（简化版本）
	currentPID := os.Getpid()

	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	srm.processes[currentPID] = &ProcessInfo{
		PID:       currentPID,
		PPID:      os.Getppid(),
		Name:      filepath.Base(os.Args[0]),
		State:     ProcessRunning,
		StartTime: time.Now(),
		Command:   strings.Join(os.Args, " "),
		Arguments: os.Args,
	}
}

func (srm *SystemResourceMonitor) updateNetworkInfo() {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	networkInterfaces := make([]NetworkInterface, 0, len(interfaces))

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		netInterface := NetworkInterface{
			Name:         iface.Name,
			Index:        iface.Index,
			MTU:          iface.MTU,
			Flags:        iface.Flags,
			HardwareAddr: iface.HardwareAddr,
			Addresses:    addrs,
		}

		networkInterfaces = append(networkInterfaces, netInterface)
	}

	srm.network.Interfaces = networkInterfaces
}

func (srm *SystemResourceMonitor) updateFileHandleInfo() {
	// 这是一个简化的实现
	// 实际应用中需要更复杂的文件句柄监控

	currentPID := os.Getpid()

	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	// 模拟一些文件句柄
	srm.fileHandles[0] = &FileHandle{
		FD:      0,
		Path:    "/dev/stdin",
		Process: currentPID,
		Type:    FileTypeDevice,
	}

	srm.fileHandles[1] = &FileHandle{
		FD:      1,
		Path:    "/dev/stdout",
		Process: currentPID,
		Type:    FileTypeDevice,
	}

	srm.fileHandles[2] = &FileHandle{
		FD:      2,
		Path:    "/dev/stderr",
		Process: currentPID,
		Type:    FileTypeDevice,
	}
}

func (srm *SystemResourceMonitor) GetSystemSnapshot() SystemSnapshot {
	srm.mutex.RLock()
	defer srm.mutex.RUnlock()

	// 复制当前状态
	processes := make(map[int]*ProcessInfo)
	for pid, proc := range srm.processes {
		processes[pid] = proc
	}

	fileHandles := make(map[int]*FileHandle)
	for fd, handle := range srm.fileHandles {
		fileHandles[fd] = handle
	}

	return SystemSnapshot{
		Timestamp:   time.Now(),
		Memory:      srm.memory,
		Processes:   processes,
		FileHandles: fileHandles,
		Network:     srm.network,
	}
}

// SystemSnapshot 系统快照
type SystemSnapshot struct {
	Timestamp   time.Time
	Memory      MemoryInfo
	Processes   map[int]*ProcessInfo
	FileHandles map[int]*FileHandle
	Network     NetworkInfo
}

// ==================
// 7. 主演示函数
// ==================

// demonstrateSystemCallWrapper 演示系统调用包装器
func demonstrateSystemCallWrapper() {
	fmt.Println("1. 系统调用包装器演示")
	wrapper := NewSystemCallWrapper()

	// 模拟一些系统调用
	for i := 0; i < TestLoopCount; i++ {
		wrapper.WrapSyscall("open", func() (uintptr, uintptr, error) {
			time.Sleep(time.Microsecond * time.Duration(MicrosecondsBase+i*MicrosecondsIncrement))
			if i%5 == 0 {
				return 0, 0, fmt.Errorf("permission denied")
			}
			return uintptr(i + 3), 0, nil
		})

		wrapper.WrapSyscall("read", func() (uintptr, uintptr, error) {
			time.Sleep(time.Microsecond * time.Duration(ReadMicroseconds+i*ReadIncrement))
			return uintptr(KBToBytes), 0, nil
		})
	}

	stats := wrapper.GetStatistics()
	fmt.Println("系统调用统计:")
	for name, stat := range stats {
		fmt.Printf("  %s: 调用%d次, 平均耗时%v, 错误率%.2f%%\n",
			name, stat.CallCount, stat.AverageTime, stat.ErrorRate*100)
	}
}

// demonstrateProcessManagement 演示进程管理
func demonstrateProcessManagement() {
	fmt.Println("\n2. 进程管理演示")
	procManager := NewProcessManager()

	// 启动一个简单的进程
	config := ProcessConfig{
		MaxRestarts: 3,
		AutoRestart: true,
		HealthCheck: func() bool {
			return true // 简单的健康检查
		},
		Timeout: time.Second * ProcessTimeoutSec,
	}

	// 创建一个简单的echo进程
	proc, err := procManager.StartProcess("echo", []string{"Hello from managed process"}, config)
	if err != nil {
		fmt.Printf("启动进程失败: %v\n", err)
	} else {
		fmt.Printf("成功启动进程 PID: %d\n", proc.Pid)

		// 等待进程完成
		go func() {
			proc.Wait()
			fmt.Printf("进程 %d 已退出\n", proc.Pid)
		}()
	}

	// 显示管理的进程
	processes := procManager.GetProcesses()
	fmt.Printf("管理的进程数量: %d\n", len(processes))
}

// demonstrateSignalHandling 演示信号处理
func demonstrateSignalHandling() (*SignalHandler, func()) {
	fmt.Println("\n3. 信号处理演示")
	signalHandler := NewSignalHandler()

	signalHandler.RegisterHandler(SIGUSR1, func(signal os.Signal, context SignalContext) {
		fmt.Printf("收到信号 %v 在 %v\n", signal, context.Timestamp)
	})

	signalHandler.RegisterHandler(syscall.SIGTERM, func(signal os.Signal, context SignalContext) {
		fmt.Printf("收到终止信号，准备优雅关闭\n")
	})

	signalHandler.Start()
	return signalHandler, func() { signalHandler.Stop() }
}

// demonstrateMemoryMapping 演示内存映射
func demonstrateMemoryMapping() {
	fmt.Println("\n4. 内存映射演示")
	mapper := NewMemoryMapper()

	// 创建一个临时文件进行映射
	tmpFile, err := os.CreateTemp("", "mmap_test")
	if err == nil {
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// 写入一些数据
		testData := []byte("Hello, Memory Mapped File!")
		tmpFile.Write(testData)
		tmpFile.Sync()

		// 映射文件到内存
		mapping, err := mapper.MapFile(tmpFile, 0, uintptr(len(testData)),
			PROT_READ|PROT_WRITE, MAP_SHARED)
		if err != nil {
			fmt.Printf("内存映射失败: %v\n", err)
		} else {
			fmt.Printf("成功映射文件到地址: 0x%x, 大小: %d\n", mapping.Address, mapping.Size)

			// 清理映射
			mapper.Unmap(mapping.Address)
		}
	}

	// 创建匿名内存映射
	anonMapping, err := mapper.MapAnonymous(DefaultPageSize,
		PROT_READ|PROT_WRITE, MAP_PRIVATE)
	if err != nil {
		fmt.Printf("匿名内存映射失败: %v\n", err)
	} else {
		fmt.Printf("成功创建匿名映射到地址: 0x%x, 大小: %d\n",
			anonMapping.Address, anonMapping.Size)

		// 清理映射
		mapper.Unmap(anonMapping.Address)
	}

	mappings := mapper.GetMappings()
	fmt.Printf("当前活跃的内存映射数量: %d\n", len(mappings))
}

// demonstrateIPC 演示进程间通信
func demonstrateIPC() {
	fmt.Println("\n5. 进程间通信(IPC)演示")
	ipcManager := NewIPCManager()

	// 创建命名管道
	pipeName := "test_pipe"
	// G306安全修复：避免在/tmp目录创建文件，使用更安全的位置和权限
	// 创建安全的临时目录
	tempDir, err := os.MkdirTemp("", "ipc_test")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	pipePath := filepath.Join(tempDir, "test_fifo")
	pipe, err := ipcManager.CreateNamedPipe(pipeName, pipePath, DefaultPermission) // 使用更安全的权限
	if err != nil {
		fmt.Printf("创建命名管道失败: %v\n", err)
	} else {
		fmt.Printf("成功创建命名管道: %s -> %s\n", pipeName, pipePath)
		fmt.Printf("管道路径: %s\n", pipe.Path)
		defer os.Remove(pipePath)
	}

	// 创建信号量
	sem, err := ipcManager.CreateSemaphore("test_sem", 1, MaxSemaphoreValue)
	if err != nil {
		fmt.Printf("创建信号量失败: %v\n", err)
	} else {
		fmt.Printf("成功创建信号量: %s, 初值: %d\n", sem.Name, sem.Value)

		// 测试信号量操作
		fmt.Printf("信号量当前值: %d\n", atomic.LoadInt32(&sem.Value))
		ipcManager.Wait("test_sem")
		fmt.Printf("Wait后信号量值: %d\n", atomic.LoadInt32(&sem.Value))
		ipcManager.Signal("test_sem")
		fmt.Printf("Signal后信号量值: %d\n", atomic.LoadInt32(&sem.Value))
	}

	// 创建消息队列
	mq, err := ipcManager.CreateMessageQueue("test_mq", MaxMessageQueueSize)
	if err != nil {
		fmt.Printf("创建消息队列失败: %v\n", err)
	} else {
		fmt.Printf("成功创建消息队列: %s, 最大大小: %d\n", mq.Name, mq.MaxSize)

		// 测试消息发送和接收
		err := ipcManager.SendMessage("test_mq", 1, []byte("Hello, IPC!"))
		if err != nil {
			fmt.Printf("发送消息失败: %v\n", err)
		} else {
			fmt.Println("成功发送消息")

			msg, err := ipcManager.ReceiveMessage("test_mq", 1)
			if err != nil {
				fmt.Printf("接收消息失败: %v\n", err)
			} else {
				fmt.Printf("接收到消息: %s (类型: %d, 发送者: %d)\n",
					string(msg.Data), msg.Type, msg.Sender)
			}
		}
	}
}

// demonstrateResourceMonitoring 演示系统资源监控
func demonstrateResourceMonitoring() (*SystemResourceMonitor, func()) {
	fmt.Println("\n6. 系统资源监控演示")
	monitor := NewSystemResourceMonitor()
	monitor.Start()

	// 等待一些监控数据
	time.Sleep(time.Second * 2)

	snapshot := monitor.GetSystemSnapshot()

	fmt.Printf("系统快照 (时间: %v):\n", snapshot.Timestamp.Format("15:04:05"))
	fmt.Printf("  内存信息:\n")
	fmt.Printf("    总物理内存: %d MB\n", snapshot.Memory.TotalPhysical/KBToBytes/KBToBytes)
	fmt.Printf("    可用物理内存: %d MB\n", snapshot.Memory.AvailablePhysical/KBToBytes/KBToBytes)
	fmt.Printf("    已用物理内存: %d MB\n", snapshot.Memory.UsedPhysical/KBToBytes/KBToBytes)
	fmt.Printf("    页面大小: %d bytes\n", snapshot.Memory.PageSize)

	fmt.Printf("  进程信息:\n")
	fmt.Printf("    监控的进程数量: %d\n", len(snapshot.Processes))

	// 显示前5个进程信息
	var pids []int
	for pid := range snapshot.Processes {
		pids = append(pids, pid)
	}
	sort.Ints(pids)

	count := 0
	for _, pid := range pids {
		if count >= MaxProcessDisplay {
			break
		}
		proc := snapshot.Processes[pid]
		fmt.Printf("    PID %d: %s (状态: %s, PPID: %d)\n",
			proc.PID, proc.Name, proc.State, proc.PPID)
		count++
	}

	fmt.Printf("  网络接口:\n")
	for _, iface := range snapshot.Network.Interfaces {
		fmt.Printf("    %s: MTU=%d, 地址数量=%d\n",
			iface.Name, iface.MTU, len(iface.Addresses))
	}

	fmt.Printf("  文件句柄:\n")
	fmt.Printf("    监控的文件句柄数量: %d\n", len(snapshot.FileHandles))

	return monitor, func() { monitor.Stop() }
}

// demonstrateUserAndSystemInfo 演示用户和系统信息
func demonstrateUserAndSystemInfo() {
	// 7. 用户和组信息
	fmt.Println("\n7. 用户和组信息")
	currentUser, err := user.Current()
	if err != nil {
		fmt.Printf("获取当前用户失败: %v\n", err)
	} else {
		fmt.Printf("当前用户: %s (UID: %s, GID: %s)\n",
			currentUser.Username, currentUser.Uid, currentUser.Gid)
		fmt.Printf("用户主目录: %s\n", currentUser.HomeDir)
	}

	// 获取当前进程的UID和GID
	fmt.Printf("进程 UID: %d\n", os.Getuid())
	fmt.Printf("进程 GID: %d\n", os.Getgid())
	fmt.Printf("进程 EUID: %d\n", os.Geteuid())
	fmt.Printf("进程 EGID: %d\n", os.Getegid())

	// 8. 系统信息汇总
	fmt.Println("\n8. 系统信息汇总")
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("CPU核心数: %d\n", runtime.NumCPU())
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("Go协程数: %d\n", runtime.NumGoroutine())

	// 获取系统负载（Linux）
	if runtime.GOOS == "linux" {
		loadavg, err := os.ReadFile("/proc/loadavg")
		if err == nil {
			fmt.Printf("系统负载: %s", string(loadavg))
		}
	}

	// 获取系统运行时间（Linux）
	if runtime.GOOS == "linux" {
		uptime, err := os.ReadFile("/proc/uptime")
		if err == nil {
			fields := strings.Fields(string(uptime))
			if len(fields) >= 1 {
				seconds, err := strconv.ParseFloat(fields[0], 64)
				if err == nil {
					fmt.Printf("系统运行时间: %.2f 小时\n", seconds/SecondsToHours)
				}
			}
		}
	}
}

func demonstrateSystemProgramming() {
	fmt.Println("=== Go系统编程：操作系统接口掌控演示 ===")

	// 1. 系统调用包装器演示
	demonstrateSystemCallWrapper()

	// 2. 进程管理演示
	demonstrateProcessManagement()

	// 3. 信号处理演示
	_, stopSignalHandler := demonstrateSignalHandling()
	defer stopSignalHandler()

	// 4. 内存映射演示
	demonstrateMemoryMapping()

	// 5. 进程间通信演示
	demonstrateIPC()

	// 6. 系统资源监控演示
	_, stopMonitor := demonstrateResourceMonitoring()
	defer stopMonitor()

	// 7. 用户和系统信息演示
	demonstrateUserAndSystemInfo()
}

func main() {
	demonstrateSystemProgramming()

	fmt.Println("\n=== Go系统编程：操作系统接口掌控演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 系统调用封装：统一的系统调用接口和性能监控")
	fmt.Println("2. 进程管理：进程生命周期控制和自动重启机制")
	fmt.Println("3. 信号处理：异步信号处理和优雅关闭")
	fmt.Println("4. 内存映射：高效的文件和内存操作")
	fmt.Println("5. IPC机制：进程间通信的多种实现方式")
	fmt.Println("6. 资源监控：系统资源的实时监控和统计")
	fmt.Println("7. 跨平台抽象：统一的系统编程接口")

	fmt.Println("\n高级系统编程特性:")
	fmt.Println("- 零拷贝I/O和高性能网络编程")
	fmt.Println("- 系统调用性能优化和批处理")
	fmt.Println("- 内存管理和虚拟内存操作")
	fmt.Println("- 实时系统监控和性能调优")
	fmt.Println("- 容器和虚拟化技术集成")
	fmt.Println("- 系统安全和权限管理")
}

/*
=== 练习题 ===

1. 系统调用优化：
   - 实现系统调用批处理机制
   - 添加系统调用缓存策略
   - 创建异步系统调用接口
   - 分析系统调用开销和优化

2. 进程管理增强：
   - 实现进程组管理
   - 添加进程资源限制
   - 创建进程调度策略
   - 实现进程热迁移

3. 高级IPC：
   - 实现零拷贝消息传递
   - 添加分布式IPC机制
   - 创建高性能共享内存
   - 实现消息序列化优化

4. 系统监控扩展：
   - 实现实时性能分析
   - 添加系统瓶颈检测
   - 创建资源使用预测
   - 实现自适应监控频率

重要概念：
- System Call: 系统调用和内核接口
- Process Management: 进程生命周期管理
- Signal Handling: 信号处理和异步通信
- Memory Mapping: 内存映射和虚拟内存
- IPC: 进程间通信机制
- Resource Monitor: 系统资源监控
- Performance Tuning: 系统性能调优
*/
