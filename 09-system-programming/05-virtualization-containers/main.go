/*
=== Go系统编程：虚拟化与容器大师 ===

本模块专注于Go语言虚拟化与容器技术的深度掌握，探索：
1. 容器运行时的底层实现
2. Linux命名空间的深度控制
3. 资源隔离与限制技术
4. 容器网络的实现原理
5. 存储驱动与文件系统
6. 安全沙箱与权限控制
7. 容器编排引擎设计
8. 虚拟机监控器实现
9. 云原生基础设施
10. 高性能容器优化

学习目标：
- 掌握容器技术的核心原理
- 理解虚拟化的底层机制
- 学会容器运行时开发
- 掌握云原生基础设施设计
*/

package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go-mastery/common/security"
)

// 安全随机数生成函数
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(max)
		if fallback > int64(^uint(0)>>1) {
			fallback = fallback % int64(^uint(0)>>1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result = result % int64(max)
	}
	return int(result)
}

func secureRandomInt63() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		// 安全fallback：使用时间戳
		return time.Now().UnixNano()
	}
	return n.Int64()
}

// G204安全修复：输入验证函数防止命令注入
func validateNetworkName(name string) error {
	// 网络名称只允许字母数字字符、连字符和下划线，长度限制1-63字符
	if len(name) == 0 || len(name) > 63 {
		return fmt.Errorf("网络名称长度必须在1-63字符之间")
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if !matched {
		return fmt.Errorf("网络名称只能包含字母、数字、连字符和下划线")
	}
	return nil
}

func validateIPAddress(ip string) error {
	// 验证IP地址格式
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("无效的IP地址格式: %s", ip)
	}
	return nil
}

func validateExecutablePath(path string) error {
	// 验证可执行文件路径，只允许白名单中的命令
	allowedCommands := map[string]bool{
		"sh":      true,
		"bash":    true,
		"python":  true,
		"python3": true,
		"node":    true,
		"java":    true,
		"go":      true,
		"php":     true,
		"ruby":    true,
		"perl":    true,
	}

	// 提取命令名称
	cmdName := filepath.Base(path)
	if !allowedCommands[cmdName] {
		return fmt.Errorf("不允许的命令: %s", cmdName)
	}

	// 检查路径是否包含危险字符
	if strings.Contains(path, "..") || strings.Contains(path, ";") ||
		strings.Contains(path, "&") || strings.Contains(path, "|") ||
		strings.Contains(path, "$") || strings.Contains(path, "`") {
		return fmt.Errorf("命令路径包含危险字符")
	}

	return nil
}

// Windows compatible clone constants (placeholders)
const (
	CLONE_NEWNS  = 0x00020000
	CLONE_NEWPID = 0x20000000
	CLONE_NEWNET = 0x40000000
	CLONE_NEWIPC = 0x08000000
	CLONE_NEWUTS = 0x04000000
)

// Windows compatible syscall extensions
var (
	syscallCLONE_NEWNS  = CLONE_NEWNS
	syscallCLONE_NEWPID = CLONE_NEWPID
	syscallCLONE_NEWNET = CLONE_NEWNET
	syscallCLONE_NEWIPC = CLONE_NEWIPC
	syscallCLONE_NEWUTS = CLONE_NEWUTS
)

// Windows compatible syscall functions
func windowsMount(source, target, fstype string, flags uintptr, data string) error {
	// Placeholder implementation for Windows
	return fmt.Errorf("mount not supported on Windows")
}

func windowsUnmount(target string, flags int) error {
	// Placeholder implementation for Windows
	return fmt.Errorf("unmount not supported on Windows")
}

func setns(fd uintptr, nstype int) error {
	// Placeholder implementation for Windows
	return fmt.Errorf("setns not supported on Windows")
}

// Define syscall constants for Windows compatibility
const (
	SYS_SETNS = 308 // Placeholder value
)

// Windows compatible SysProcAttr wrapper
type WindowsSysProcAttr struct {
	*syscall.SysProcAttr
	Cloneflags   uintptr
	Unshareflags uintptr
}

// ==================
// 1. 容器运行时核心
// ==================

// ContainerRuntime 容器运行时
type ContainerRuntime struct {
	containers map[string]*Container
	images     map[string]*ContainerImage
	networks   map[string]*ContainerNetwork
	volumes    map[string]*ContainerVolume
	namespaces *NamespaceManager
	cgroups    *CgroupManager
	seccomp    *SeccompManager
	apparmor   *ApparmorManager
	storage    *StorageManager
	network    *NetworkManager
	config     RuntimeConfig
	statistics RuntimeStatistics
	eventBus   *ContainerEventBus
	monitor    *ContainerMonitor
	mutex      sync.RWMutex
	running    bool
	stopCh     chan struct{}
}

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	RootDirectory      string
	StateDirectory     string
	LogLevel           string
	MaxContainers      int
	DefaultRuntime     string
	EnableSelinux      bool
	EnableApparmor     bool
	EnableSeccomp      bool
	DefaultNetworkMode string
	StorageDriver      string
	CgroupVersion      int
	OOMKillDisable     bool
	PidsLimit          int64
	ShmSize            int64
}

// Container 容器实例
type Container struct {
	ID              string
	Name            string
	Image           *ContainerImage
	Config          *ContainerConfig
	State           *ContainerState
	Process         *ContainerProcess
	Namespaces      map[string]*Namespace
	Cgroups         map[string]*Cgroup
	Mounts          []*Mount
	Networks        []*NetworkInterface
	Volumes         []*Volume
	SecurityContext *SecurityContext
	Resources       *ResourceConstraints
	Statistics      *ContainerStatistics
	CreatedAt       time.Time
	StartedAt       time.Time
	FinishedAt      time.Time
	ExitCode        int
	mutex           sync.RWMutex
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	ExposedPorts    map[string]struct{}
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	Healthcheck     *HealthConfig
	ArgsEscaped     bool
	Image           string
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	MacAddress      string
	OnBuild         []string
	Labels          map[string]string
	StopSignal      string
	StopTimeout     *int
	Shell           []string
}

// ContainerState 容器状态
type ContainerState struct {
	Status     ContainerStatus
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	ExitCode   int
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
	Health     *Health
}

// ContainerStatus 容器状态枚举
type ContainerStatus int

const (
	StatusCreated ContainerStatus = iota
	StatusRunning
	StatusPaused
	StatusRestarting
	StatusRemoving
	StatusExited
	StatusDead
)

func (cs ContainerStatus) String() string {
	statuses := []string{"created", "running", "paused", "restarting", "removing", "exited", "dead"}
	if int(cs) < len(statuses) {
		return statuses[cs]
	}
	return "unknown"
}

// ContainerProcess 容器进程
type ContainerProcess struct {
	Pid      int
	Args     []string
	Env      []string
	Cwd      string
	Stdin    io.WriteCloser
	Stdout   io.ReadCloser
	Stderr   io.ReadCloser
	Wait     chan error
	Started  time.Time
	ExitCode int
}

func NewContainerRuntime(config RuntimeConfig) *ContainerRuntime {
	return &ContainerRuntime{
		containers: make(map[string]*Container),
		images:     make(map[string]*ContainerImage),
		networks:   make(map[string]*ContainerNetwork),
		volumes:    make(map[string]*ContainerVolume),
		namespaces: NewNamespaceManager(),
		cgroups:    NewCgroupManager(),
		seccomp:    NewSeccompManager(),
		apparmor:   NewApparmorManager(),
		storage:    NewStorageManager(),
		network:    NewNetworkManager(),
		config:     config,
		eventBus:   NewContainerEventBus(),
		monitor:    NewContainerMonitor(),
		stopCh:     make(chan struct{}),
	}
}

func (cr *ContainerRuntime) Start() error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if cr.running {
		return fmt.Errorf("container runtime already running")
	}

	// 初始化存储驱动
	if err := cr.storage.Initialize(cr.config.StorageDriver); err != nil {
		return fmt.Errorf("failed to initialize storage: %v", err)
	}

	// 初始化网络
	if err := cr.network.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize network: %v", err)
	}

	// 启动监控服务
	go cr.monitorLoop()
	go cr.eventLoop()
	go cr.cleanupLoop()

	cr.running = true
	fmt.Println("容器运行时已启动")
	return nil
}

func (cr *ContainerRuntime) CreateContainer(config *ContainerConfig) (*Container, error) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	// 生成容器ID
	containerID := generateContainerID()

	// 查找镜像
	image, exists := cr.images[config.Image]
	if !exists {
		return nil, fmt.Errorf("image not found: %s", config.Image)
	}

	// 创建容器实例
	container := &Container{
		ID:         containerID,
		Name:       generateContainerName(),
		Image:      image,
		Config:     config,
		State:      &ContainerState{Status: StatusCreated},
		Namespaces: make(map[string]*Namespace),
		Cgroups:    make(map[string]*Cgroup),
		Mounts:     make([]*Mount, 0),
		Networks:   make([]*NetworkInterface, 0),
		Volumes:    make([]*Volume, 0),
		CreatedAt:  time.Now(),
	}

	// 创建命名空间
	if err := cr.createNamespaces(container); err != nil {
		return nil, fmt.Errorf("failed to create namespaces: %v", err)
	}

	// 创建Cgroups
	if err := cr.createCgroups(container); err != nil {
		return nil, fmt.Errorf("failed to create cgroups: %v", err)
	}

	// 准备文件系统
	if err := cr.prepareFilesystem(container); err != nil {
		return nil, fmt.Errorf("failed to prepare filesystem: %v", err)
	}

	cr.containers[containerID] = container
	fmt.Printf("创建容器: %s (镜像: %s)\n", containerID[:12], config.Image)

	// 发送事件
	cr.eventBus.Publish(&ContainerEvent{
		Type:      EventContainerCreate,
		Container: container,
		Timestamp: time.Now(),
	})

	return container, nil
}

func (cr *ContainerRuntime) StartContainer(containerID string) error {
	cr.mutex.RLock()
	container, exists := cr.containers[containerID]
	cr.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	container.mutex.Lock()
	defer container.mutex.Unlock()

	if container.State.Status != StatusCreated {
		return fmt.Errorf("container not in created state: %s", container.State.Status)
	}

	// 启动容器进程
	process, err := cr.startContainerProcess(container)
	if err != nil {
		return fmt.Errorf("failed to start container process: %v", err)
	}

	container.Process = process
	container.State.Status = StatusRunning
	container.State.Running = true
	container.State.Pid = process.Pid
	container.StartedAt = time.Now()

	fmt.Printf("启动容器: %s (PID: %d)\n", containerID[:12], process.Pid)

	// 发送事件
	cr.eventBus.Publish(&ContainerEvent{
		Type:      EventContainerStart,
		Container: container,
		Timestamp: time.Now(),
	})

	// 异步等待进程结束
	go cr.waitForProcess(container)

	return nil
}

func (cr *ContainerRuntime) StopContainer(containerID string, timeout time.Duration) error {
	cr.mutex.RLock()
	container, exists := cr.containers[containerID]
	cr.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	container.mutex.Lock()
	defer container.mutex.Unlock()

	if !container.State.Running {
		return fmt.Errorf("container not running: %s", containerID)
	}

	// 发送终止信号
	if container.Process != nil && container.Process.Pid > 0 {
		process, err := os.FindProcess(container.Process.Pid)
		if err == nil {
			// 先发送SIGTERM
			if err := process.Signal(syscall.SIGTERM); err != nil {
				log.Printf("Warning: failed to send SIGTERM to process: %v", err)
			}

			// 等待超时或进程结束
			done := make(chan bool, 1)
			go func() {
				select {
				case <-container.Process.Wait:
					done <- true
				case <-time.After(timeout):
					// 超时后发送SIGKILL
					if err := process.Signal(syscall.SIGKILL); err != nil {
						log.Printf("Warning: failed to send SIGKILL to process: %v", err)
					}
					done <- true
				}
			}()

			<-done
		}
	}

	container.State.Status = StatusExited
	container.State.Running = false
	container.FinishedAt = time.Now()

	fmt.Printf("停止容器: %s\n", containerID[:12])

	// 发送事件
	cr.eventBus.Publish(&ContainerEvent{
		Type:      EventContainerStop,
		Container: container,
		Timestamp: time.Now(),
	})

	return nil
}

func (cr *ContainerRuntime) RemoveContainer(containerID string, force bool) error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	container, exists := cr.containers[containerID]
	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	if container.State.Running && !force {
		return fmt.Errorf("cannot remove running container without force")
	}

	// 强制停止运行中的容器
	if container.State.Running && force {
		if err := cr.StopContainer(containerID, 5*time.Second); err != nil {
			log.Printf("Warning: failed to stop container: %v", err)
		}
	}

	// 清理资源
	cr.cleanupContainer(container)

	delete(cr.containers, containerID)
	fmt.Printf("删除容器: %s\n", containerID[:12])

	// 发送事件
	cr.eventBus.Publish(&ContainerEvent{
		Type:      EventContainerRemove,
		Container: container,
		Timestamp: time.Now(),
	})

	return nil
}

func (cr *ContainerRuntime) createNamespaces(container *Container) error {
	// 创建各种命名空间
	namespaces := []string{"pid", "net", "ipc", "uts", "mnt", "user"}

	for _, nsType := range namespaces {
		ns, err := cr.namespaces.CreateNamespace(nsType, container.ID)
		if err != nil {
			return fmt.Errorf("failed to create %s namespace: %v", nsType, err)
		}
		container.Namespaces[nsType] = ns
	}

	return nil
}

func (cr *ContainerRuntime) createCgroups(container *Container) error {
	// 创建cgroup层次结构
	subsystems := []string{"memory", "cpu", "cpuset", "blkio", "net_cls", "freezer"}

	for _, subsystem := range subsystems {
		cgroup, err := cr.cgroups.CreateCgroup(subsystem, container.ID)
		if err != nil {
			return fmt.Errorf("failed to create %s cgroup: %v", subsystem, err)
		}
		container.Cgroups[subsystem] = cgroup
	}

	return nil
}

func (cr *ContainerRuntime) prepareFilesystem(container *Container) error {
	// 创建容器根目录
	containerRoot := filepath.Join(cr.config.RootDirectory, "containers", container.ID)
	// #nosec G301 -- Linux容器标准目录权限0755，需要可执行位支持目录访问
	if err := os.MkdirAll(containerRoot, 0755); err != nil {
		return err
	}

	// 准备镜像层
	layerPath := filepath.Join(containerRoot, "layer")
	if err := cr.storage.PrepareLayer(container.Image, layerPath); err != nil {
		return err
	}

	// 创建读写层
	rwLayer := filepath.Join(containerRoot, "rw")
	// #nosec G301 -- Linux容器文件系统层，需要0755权限支持overlay文件系统挂载
	if err := os.MkdirAll(rwLayer, 0755); err != nil {
		return err
	}

	// 创建合并挂载点
	mergedPath := filepath.Join(containerRoot, "merged")
	// #nosec G301 -- overlay文件系统挂载点，需要0755权限支持容器根文件系统访问
	if err := os.MkdirAll(mergedPath, 0755); err != nil {
		return err
	}

	// 挂载文件系统
	mount := &Mount{
		Source:      layerPath,
		Target:      mergedPath,
		Type:        "overlay",
		Options:     fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s/work", layerPath, rwLayer, containerRoot),
		Propagation: "private",
	}

	container.Mounts = append(container.Mounts, mount)
	return nil
}

func (cr *ContainerRuntime) startContainerProcess(container *Container) (*ContainerProcess, error) {
	// 构建命令
	var cmd *exec.Cmd
	if len(container.Config.Entrypoint) > 0 {
		// G204安全修复：验证可执行文件路径
		if err := validateExecutablePath(container.Config.Entrypoint[0]); err != nil {
			return nil, fmt.Errorf("无效的容器入口点: %v", err)
		}
		args := append(container.Config.Entrypoint, container.Config.Cmd...)
		// #nosec G204 - 命令已通过validateExecutablePath白名单验证
		cmd = exec.Command(args[0], args[1:]...)
	} else if len(container.Config.Cmd) > 0 {
		// G204安全修复：验证可执行文件路径
		if err := validateExecutablePath(container.Config.Cmd[0]); err != nil {
			return nil, fmt.Errorf("无效的容器命令: %v", err)
		}
		// #nosec G204 - 命令已通过validateExecutablePath白名单验证
		cmd = exec.Command(container.Config.Cmd[0], container.Config.Cmd[1:]...)
	} else {
		return nil, fmt.Errorf("no command specified")
	}

	// 设置环境变量
	cmd.Env = container.Config.Env

	// 设置工作目录
	if container.Config.WorkingDir != "" {
		cmd.Dir = container.Config.WorkingDir
	}

	// 配置命名空间 (Windows下禁用)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	// Cloneflags: syscallCLONE_NEWNS | syscallCLONE_NEWPID | syscallCLONE_NEWNET |
	// 	syscallCLONE_NEWIPC | syscallCLONE_NEWUTS,
	// Unshareflags: syscallCLONE_NEWNS,

	// 设置标准输入输出
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	process := &ContainerProcess{
		Pid:     cmd.Process.Pid,
		Args:    cmd.Args,
		Env:     cmd.Env,
		Cwd:     cmd.Dir,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Wait:    make(chan error, 1),
		Started: time.Now(),
	}

	// 异步等待进程结束
	go func() {
		err := cmd.Wait()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				process.ExitCode = exitError.ExitCode()
			}
		}
		process.Wait <- err
	}()

	return process, nil
}

func (cr *ContainerRuntime) waitForProcess(container *Container) {
	err := <-container.Process.Wait

	container.mutex.Lock()
	defer container.mutex.Unlock()

	container.State.Running = false
	container.State.Status = StatusExited
	container.FinishedAt = time.Now()

	if err != nil {
		container.State.Error = err.Error()
		container.ExitCode = container.Process.ExitCode
	}

	fmt.Printf("容器进程结束: %s (退出码: %d)\n", container.ID[:12], container.ExitCode)

	// 发送事件
	cr.eventBus.Publish(&ContainerEvent{
		Type:      EventContainerDie,
		Container: container,
		Timestamp: time.Now(),
	})
}

func (cr *ContainerRuntime) cleanupContainer(container *Container) {
	// 清理命名空间
	for _, ns := range container.Namespaces {
		if err := cr.namespaces.DestroyNamespace(ns); err != nil {
			log.Printf("Warning: failed to destroy namespace: %v", err)
		}
	}

	// 清理Cgroups
	for _, cgroup := range container.Cgroups {
		if err := cr.cgroups.DestroyCgroup(cgroup); err != nil {
			log.Printf("Warning: failed to destroy cgroup: %v", err)
		}
	}

	// 清理挂载点
	for _, mount := range container.Mounts {
		if err := windowsUnmount(mount.Target, 0); err != nil {
			log.Printf("Warning: failed to unmount %s: %v", mount.Target, err)
		}
	}

	// 清理文件系统
	containerRoot := filepath.Join(cr.config.RootDirectory, "containers", container.ID)
	if err := os.RemoveAll(containerRoot); err != nil {
		log.Printf("Warning: failed to remove container root directory: %v", err)
	}
}

// monitorLoop 监控循环
func (cr *ContainerRuntime) monitorLoop() {
	for {
		select {
		case <-cr.stopCh:
			return
		default:
			// 监控容器状态
			cr.mutex.RLock()
			for _, container := range cr.containers {
				if container.State.Status == StatusRunning {
					// 检查容器健康状态
				}
			}
			cr.mutex.RUnlock()
			time.Sleep(5 * time.Second)
		}
	}
}

// eventLoop 事件循环
func (cr *ContainerRuntime) eventLoop() {
	for {
		select {
		case <-cr.stopCh:
			return
		default:
			// 处理容器事件
			time.Sleep(1 * time.Second)
		}
	}
}

// cleanupLoop 清理循环
func (cr *ContainerRuntime) cleanupLoop() {
	for {
		select {
		case <-cr.stopCh:
			return
		default:
			// 清理停止的容器
			cr.mutex.RLock()
			for _, container := range cr.containers {
				if container.State.Status == StatusExited {
					// 执行清理操作
				}
			}
			cr.mutex.RUnlock()
			time.Sleep(30 * time.Second)
		}
	}
}

// ==================
// 2. 命名空间管理
// ==================

// NamespaceManager 命名空间管理器
type NamespaceManager struct {
	namespaces map[string]*Namespace
	mutex      sync.RWMutex
}

// Namespace 命名空间
type Namespace struct {
	Type      string
	ID        string
	Path      string
	Pid       int
	CreatedAt time.Time
	RefCount  int32
}

func NewNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		namespaces: make(map[string]*Namespace),
	}
}

func (nm *NamespaceManager) CreateNamespace(nsType, containerID string) (*Namespace, error) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nsID := fmt.Sprintf("%s-%s", nsType, containerID)

	ns := &Namespace{
		Type:      nsType,
		ID:        nsID,
		Path:      fmt.Sprintf("/proc/self/ns/%s", nsType),
		CreatedAt: time.Now(),
		RefCount:  1,
	}

	nm.namespaces[nsID] = ns
	fmt.Printf("创建命名空间: %s (类型: %s)\n", nsID, nsType)

	return ns, nil
}

func (nm *NamespaceManager) DestroyNamespace(ns *Namespace) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if atomic.AddInt32(&ns.RefCount, -1) <= 0 {
		delete(nm.namespaces, ns.ID)
		fmt.Printf("销毁命名空间: %s\n", ns.ID)
	}

	return nil
}

func (nm *NamespaceManager) EnterNamespace(ns *Namespace) error {
	// 进入指定命名空间
	fd, err := syscall.Open(ns.Path, syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := syscall.Close(fd); err != nil {
			log.Printf("Warning: failed to close file descriptor: %v", err)
		}
	}()

	return setns(uintptr(fd), 0)
}

// ==================
// 3. Cgroup资源管理
// ==================

// CgroupManager Cgroup管理器
type CgroupManager struct {
	cgroups     map[string]*Cgroup
	controllers map[string]*CgroupController
	version     int
	mountPoint  string
	mutex       sync.RWMutex
}

// Cgroup 控制组
type Cgroup struct {
	Subsystem   string
	Path        string
	Controllers []string
	Processes   []int
	Limits      map[string]interface{}
	Stats       map[string]interface{}
	CreatedAt   time.Time
}

// CgroupController Cgroup控制器
type CgroupController struct {
	Name       string
	MountPoint string
	Hierarchy  int
	NumCgroups int
	Enabled    bool
	Subsystems []string
}

func NewCgroupManager() *CgroupManager {
	return &CgroupManager{
		cgroups:     make(map[string]*Cgroup),
		controllers: make(map[string]*CgroupController),
		version:     2, // 默认使用cgroup v2
		mountPoint:  "/sys/fs/cgroup",
	}
}

func (cm *CgroupManager) CreateCgroup(subsystem, containerID string) (*Cgroup, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cgroupPath := filepath.Join(cm.mountPoint, subsystem, "docker", containerID)

	// 创建cgroup目录
	// #nosec G301 -- Linux cgroup系统目录，需要0755权限支持内核cgroup子系统访问
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return nil, err
	}

	cgroup := &Cgroup{
		Subsystem: subsystem,
		Path:      cgroupPath,
		Processes: make([]int, 0),
		Limits:    make(map[string]interface{}),
		Stats:     make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	cgroupID := fmt.Sprintf("%s-%s", subsystem, containerID)
	cm.cgroups[cgroupID] = cgroup

	fmt.Printf("创建Cgroup: %s (子系统: %s)\n", cgroupID, subsystem)
	return cgroup, nil
}

func (cm *CgroupManager) DestroyCgroup(cgroup *Cgroup) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 移除所有进程
	if err := cm.removeAllProcesses(cgroup); err != nil {
		return err
	}

	// 删除cgroup目录
	if err := os.Remove(cgroup.Path); err != nil {
		return err
	}

	fmt.Printf("销毁Cgroup: %s\n", cgroup.Path)
	return nil
}

func (cm *CgroupManager) AddProcess(cgroup *Cgroup, pid int) error {
	cgroup.Processes = append(cgroup.Processes, pid)

	// 将进程ID写入cgroup.procs文件（cgroup系统接口）
	procsFile := filepath.Join(cgroup.Path, "cgroup.procs")
	return security.SecureWriteFile(procsFile, []byte(strconv.Itoa(pid)), &security.SecureFileOptions{
		Mode:      security.DefaultFileMode,
		CreateDir: false,
	})
}

func (cm *CgroupManager) SetMemoryLimit(cgroup *Cgroup, limit int64) error {
	cgroup.Limits["memory"] = limit

	limitFile := filepath.Join(cgroup.Path, "memory.max")
	return security.SecureWriteFile(limitFile, []byte(strconv.FormatInt(limit, 10)), &security.SecureFileOptions{
		Mode:      security.DefaultFileMode,
		CreateDir: false,
	})
}

func (cm *CgroupManager) SetCPUQuota(cgroup *Cgroup, quota int64, period int64) error {
	cgroup.Limits["cpu_quota"] = quota
	cgroup.Limits["cpu_period"] = period

	quotaFile := filepath.Join(cgroup.Path, "cpu.max")
	quotaValue := fmt.Sprintf("%d %d", quota, period)
	return security.SecureWriteFile(quotaFile, []byte(quotaValue), &security.SecureFileOptions{
		Mode:      security.DefaultFileMode,
		CreateDir: false,
	})
}

func (cm *CgroupManager) GetStats(cgroup *Cgroup) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 读取内存统计
	memStatFile := filepath.Join(cgroup.Path, "memory.stat")
	// #nosec G304 -- cgroup.Path由CgroupManager管理，memory.stat是Linux内核标准cgroup文件，系统编程操作安全
	if data, err := os.ReadFile(memStatFile); err == nil {
		memStats := cm.parseMemoryStats(string(data))
		stats["memory"] = memStats
	}

	// 读取CPU统计
	cpuStatFile := filepath.Join(cgroup.Path, "cpu.stat")
	// #nosec G304 -- cgroup.Path由CgroupManager管理，cpu.stat是Linux内核标准cgroup文件，系统编程操作安全
	if data, err := os.ReadFile(cpuStatFile); err == nil {
		cpuStats := cm.parseCPUStats(string(data))
		stats["cpu"] = cpuStats
	}

	cgroup.Stats = stats
	return stats, nil
}

func (cm *CgroupManager) parseMemoryStats(data string) map[string]int64 {
	stats := make(map[string]int64)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			if value, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				stats[parts[0]] = value
			}
		}
	}

	return stats
}

func (cm *CgroupManager) parseCPUStats(data string) map[string]int64 {
	stats := make(map[string]int64)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			if value, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				stats[parts[0]] = value
			}
		}
	}

	return stats
}

func (cm *CgroupManager) removeAllProcesses(cgroup *Cgroup) error {
	// 将所有进程移动到根cgroup
	for _, pid := range cgroup.Processes {
		if err := cm.moveProcessToRoot(cgroup.Subsystem, pid); err != nil {
			fmt.Printf("警告: 无法移动进程 %d: %v\n", pid, err)
		}
	}
	cgroup.Processes = cgroup.Processes[:0]
	return nil
}

func (cm *CgroupManager) moveProcessToRoot(subsystem string, pid int) error {
	rootProcsFile := filepath.Join(cm.mountPoint, subsystem, "cgroup.procs")
	return security.SecureWriteFile(rootProcsFile, []byte(strconv.Itoa(pid)), &security.SecureFileOptions{
		Mode:      security.DefaultFileMode,
		CreateDir: false,
	})
}

// ==================
// 4. 存储管理系统
// ==================

// StorageManager 存储管理器
type StorageManager struct {
	drivers      map[string]StorageDriver
	activeDriver StorageDriver
	graphRoot    string
	runRoot      string
	layers       map[string]*Layer
	images       map[string]*ContainerImage
	mutex        sync.RWMutex
}

// StorageDriver 存储驱动接口
type StorageDriver interface {
	Name() string
	Initialize(root string) error
	CreateLayer(id string, parent string) (*Layer, error)
	RemoveLayer(id string) error
	GetLayer(id string) (*Layer, error)
	MountLayer(id string, mountPoint string) error
	UnmountLayer(id string) error
	GetLayerSize(id string) (int64, error)
	Cleanup() error
}

// Layer 镜像层
type Layer struct {
	ID         string
	Parent     string
	Size       int64
	CreatedAt  time.Time
	MountPoint string
	Mounted    bool
	Metadata   map[string]interface{}
}

// ContainerImage 容器镜像
type ContainerImage struct {
	ID           string
	RepoTags     []string
	RepoDigests  []string
	Parent       string
	Comment      string
	Created      time.Time
	Config       *ImageConfig
	Architecture string
	Os           string
	Size         int64
	VirtualSize  int64
	Labels       map[string]string
	Layers       []string
}

// ImageConfig 镜像配置
type ImageConfig struct {
	Hostname        string
	Domainname      string
	User            string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	ExposedPorts    map[string]struct{}
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	Healthcheck     *HealthConfig
	ArgsEscaped     bool
	Image           string
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	MacAddress      string
	OnBuild         []string
	Labels          map[string]string
	StopSignal      string
	StopTimeout     *int
	Shell           []string
}

func NewStorageManager() *StorageManager {
	sm := &StorageManager{
		drivers: make(map[string]StorageDriver),
		layers:  make(map[string]*Layer),
		images:  make(map[string]*ContainerImage),
	}

	// 注册存储驱动
	sm.RegisterDriver(&OverlayFSDriver{})
	sm.RegisterDriver(&AufsDriver{})
	sm.RegisterDriver(&DeviceMapperDriver{})

	return sm
}

func (sm *StorageManager) RegisterDriver(driver StorageDriver) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.drivers[driver.Name()] = driver
	fmt.Printf("注册存储驱动: %s\n", driver.Name())
}

func (sm *StorageManager) Initialize(driverName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	driver, exists := sm.drivers[driverName]
	if !exists {
		return fmt.Errorf("storage driver not found: %s", driverName)
	}

	if err := driver.Initialize(sm.graphRoot); err != nil {
		return err
	}

	sm.activeDriver = driver
	fmt.Printf("初始化存储驱动: %s\n", driverName)
	return nil
}

func (sm *StorageManager) PrepareLayer(image *ContainerImage, mountPoint string) error {
	if sm.activeDriver == nil {
		return fmt.Errorf("no active storage driver")
	}

	// 为镜像的每一层创建layer
	var parentID string
	for _, layerID := range image.Layers {
		layer, err := sm.activeDriver.CreateLayer(layerID, parentID)
		if err != nil {
			return err
		}
		sm.layers[layerID] = layer
		parentID = layerID
	}

	// 挂载顶层
	if len(image.Layers) > 0 {
		topLayerID := image.Layers[len(image.Layers)-1]
		return sm.activeDriver.MountLayer(topLayerID, mountPoint)
	}

	return nil
}

// ==================
// 4.1 OverlayFS驱动实现
// ==================

// OverlayFSDriver OverlayFS存储驱动
type OverlayFSDriver struct {
	root      string
	layersDir string
	diffsDir  string
}

func (od *OverlayFSDriver) Name() string {
	return "overlay2"
}

func (od *OverlayFSDriver) Initialize(root string) error {
	od.root = root
	od.layersDir = filepath.Join(root, "overlay2")
	od.diffsDir = filepath.Join(od.layersDir, "l")

	// 创建目录结构
	dirs := []string{od.layersDir, od.diffsDir}
	for _, dir := range dirs {
		// #nosec G301 -- OverlayFS驱动系统目录，需要0755权限支持Docker镜像层管理
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	fmt.Printf("初始化OverlayFS驱动: %s\n", root)
	return nil
}

func (od *OverlayFSDriver) CreateLayer(id string, parent string) (*Layer, error) {
	layerDir := filepath.Join(od.layersDir, id)
	diffDir := filepath.Join(layerDir, "diff")
	workDir := filepath.Join(layerDir, "work")
	mergedDir := filepath.Join(layerDir, "merged")

	// 创建目录
	dirs := []string{layerDir, diffDir, workDir, mergedDir}
	for _, dir := range dirs {
		// #nosec G301 -- OverlayFS镜像层目录（diff/work/merged），需要0755支持容器文件系统操作
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	layer := &Layer{
		ID:        id,
		Parent:    parent,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 创建链接文件
	linkFile := filepath.Join(layerDir, "link")
	linkName := generateShortID()
	if err := security.SecureWriteFile(linkFile, []byte(linkName), &security.SecureFileOptions{
		Mode:      security.DefaultFileMode,
		CreateDir: false,
	}); err != nil {
		return nil, err
	}

	// 在l目录下创建符号链接
	linkPath := filepath.Join(od.diffsDir, linkName)
	if err := os.Symlink(diffDir, linkPath); err != nil {
		return nil, err
	}

	fmt.Printf("创建OverlayFS层: %s (父层: %s)\n", id, parent)
	return layer, nil
}

func (od *OverlayFSDriver) MountLayer(id string, mountPoint string) error {
	layerDir := filepath.Join(od.layersDir, id)
	diffDir := filepath.Join(layerDir, "diff")
	workDir := filepath.Join(layerDir, "work")

	// 构建OverlayFS挂载选项
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", diffDir, diffDir, workDir)

	// 执行挂载
	err := windowsMount("overlay", mountPoint, "overlay", 0, options)
	if err != nil {
		return fmt.Errorf("failed to mount overlay: %v", err)
	}

	fmt.Printf("挂载OverlayFS层: %s -> %s\n", id, mountPoint)
	return nil
}

func (od *OverlayFSDriver) UnmountLayer(id string) error {
	layerDir := filepath.Join(od.layersDir, id)
	mergedDir := filepath.Join(layerDir, "merged")

	err := windowsUnmount(mergedDir, 0)
	if err != nil {
		return fmt.Errorf("failed to unmount layer: %v", err)
	}

	fmt.Printf("卸载OverlayFS层: %s\n", id)
	return nil
}

func (od *OverlayFSDriver) GetLayer(id string) (*Layer, error) {
	layerDir := filepath.Join(od.layersDir, id)
	if _, err := os.Stat(layerDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("layer not found: %s", id)
	}

	return &Layer{
		ID:       id,
		Metadata: make(map[string]interface{}),
	}, nil
}

func (od *OverlayFSDriver) GetLayerSize(id string) (int64, error) {
	layerDir := filepath.Join(od.layersDir, id, "diff")
	return calculateDirectorySize(layerDir)
}

func (od *OverlayFSDriver) RemoveLayer(id string) error {
	layerDir := filepath.Join(od.layersDir, id)
	return os.RemoveAll(layerDir)
}

func (od *OverlayFSDriver) Cleanup() error {
	fmt.Println("清理OverlayFS驱动")
	return nil
}

// ==================
// 5. 网络管理系统
// ==================

// NetworkManager 网络管理器
type NetworkManager struct {
	networks   map[string]*ContainerNetwork
	bridges    map[string]*NetworkBridge
	interfaces map[string]*NetworkInterface
	ipam       *IPAddressManager
	drivers    map[string]NetworkDriver
	config     NetworkConfig
	mutex      sync.RWMutex
}

// ContainerNetwork 容器网络
type ContainerNetwork struct {
	ID         string
	Name       string
	Driver     string
	Scope      string
	Internal   bool
	Attachable bool
	Ingress    bool
	IPAM       *NetworkIPAM
	ConfigFrom *NetworkConfigReference
	ConfigOnly bool
	Containers map[string]*EndpointConfig
	Options    map[string]string
	Labels     map[string]string
	Created    time.Time
}

// NetworkIPAM IP地址管理
type NetworkIPAM struct {
	Driver  string
	Options map[string]string
	Config  []IPAMConfig
}

// IPAMConfig IPAM配置
type IPAMConfig struct {
	Subnet     string
	IPRange    string
	Gateway    string
	AuxAddress map[string]string
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name         string
	Type         string
	HardwareAddr string
	MTU          int
	IPAddresses  []string
	Gateway      string
	Bridge       string
	VethPeer     string
	Namespace    string
	Created      time.Time
}

// NetworkBridge 网络桥接
type NetworkBridge struct {
	Name       string
	Interface  string
	IPAddress  string
	Subnet     string
	Gateway    string
	MTU        int
	Interfaces []string
	Created    time.Time
}

// NetworkDriver 网络驱动接口
type NetworkDriver interface {
	Name() string
	CreateNetwork(config *NetworkConfig) (*ContainerNetwork, error)
	DeleteNetwork(networkID string) error
	CreateEndpoint(networkID, containerID string) (*EndpointConfig, error)
	DeleteEndpoint(networkID, containerID string) error
	Join(networkID, containerID string) error
	Leave(networkID, containerID string) error
}

func NewNetworkManager() *NetworkManager {
	nm := &NetworkManager{
		networks:   make(map[string]*ContainerNetwork),
		bridges:    make(map[string]*NetworkBridge),
		interfaces: make(map[string]*NetworkInterface),
		ipam:       NewIPAddressManager(),
		drivers:    make(map[string]NetworkDriver),
	}

	// 注册网络驱动
	nm.RegisterDriver(&BridgeDriver{})
	nm.RegisterDriver(&HostDriver{})
	nm.RegisterDriver(&OverlayDriver{})

	return nm
}

func (nm *NetworkManager) RegisterDriver(driver NetworkDriver) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.drivers[driver.Name()] = driver
	fmt.Printf("注册网络驱动: %s\n", driver.Name())
}

func (nm *NetworkManager) Initialize() error {
	// 创建默认网络
	defaultConfig := &NetworkConfig{
		Name:   "bridge",
		Driver: "bridge",
		IPAM: &NetworkIPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{
					Subnet:  "172.17.0.0/16",
					Gateway: "172.17.0.1",
				},
			},
		},
	}

	_, err := nm.CreateNetwork(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to create default network: %v", err)
	}

	fmt.Println("网络管理器初始化完成")
	return nil
}

func (nm *NetworkManager) CreateNetwork(config *NetworkConfig) (*ContainerNetwork, error) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	driver, exists := nm.drivers[config.Driver]
	if !exists {
		return nil, fmt.Errorf("network driver not found: %s", config.Driver)
	}

	network, err := driver.CreateNetwork(config)
	if err != nil {
		return nil, err
	}

	nm.networks[network.ID] = network
	fmt.Printf("创建网络: %s (驱动: %s)\n", network.Name, config.Driver)

	return network, nil
}

// ==================
// 5.1 Bridge网络驱动
// ==================

// BridgeDriver 桥接网络驱动
type BridgeDriver struct {
	bridges map[string]*NetworkBridge
	mutex   sync.RWMutex
}

func (bd *BridgeDriver) Name() string {
	return "bridge"
}

func (bd *BridgeDriver) CreateNetwork(config *NetworkConfig) (*ContainerNetwork, error) {
	if bd.bridges == nil {
		bd.bridges = make(map[string]*NetworkBridge)
	}

	networkID := generateNetworkID()
	bridgeName := fmt.Sprintf("br-%s", networkID[:12])

	// 创建网桥
	bridge := &NetworkBridge{
		Name:      bridgeName,
		IPAddress: config.IPAM.Config[0].Gateway,
		Subnet:    config.IPAM.Config[0].Subnet,
		Gateway:   config.IPAM.Config[0].Gateway,
		MTU:       1500,
		Created:   time.Now(),
	}

	// 执行系统命令创建网桥
	if err := bd.createBridge(bridge); err != nil {
		return nil, err
	}

	bd.mutex.Lock()
	bd.bridges[networkID] = bridge
	bd.mutex.Unlock()

	network := &ContainerNetwork{
		ID:         networkID,
		Name:       config.Name,
		Driver:     "bridge",
		IPAM:       config.IPAM,
		Containers: make(map[string]*EndpointConfig),
		Options:    make(map[string]string),
		Labels:     make(map[string]string),
		Created:    time.Now(),
	}

	fmt.Printf("创建桥接网络: %s (网桥: %s)\n", config.Name, bridgeName)
	return network, nil
}

func (bd *BridgeDriver) createBridge(bridge *NetworkBridge) error {
	// G204安全修复：验证网络名称
	if err := validateNetworkName(bridge.Name); err != nil {
		return fmt.Errorf("无效的网桥名称: %v", err)
	}

	// G204安全修复：验证IP地址
	if err := validateIPAddress(bridge.IPAddress); err != nil {
		return fmt.Errorf("无效的网桥IP地址: %v", err)
	}

	// 创建网桥接口
	// #nosec G204 - bridge.Name已通过validateNetworkName验证，固定命令用于网络配置
	cmd := exec.Command("ip", "link", "add", bridge.Name, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge: %v", err)
	}

	// 设置网桥IP地址
	// #nosec G204 - bridge.IPAddress已通过validateIPAddress验证，固定命令用于网络配置
	cmd = exec.Command("ip", "addr", "add", bridge.IPAddress+"/24", "dev", bridge.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set bridge IP: %v", err)
	}

	// 启用网桥接口
	// #nosec G204 - bridge.Name已验证，固定命令用于网络配置
	cmd = exec.Command("ip", "link", "set", bridge.Name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable bridge: %v", err)
	}

	return nil
}

func (bd *BridgeDriver) CreateEndpoint(networkID, containerID string) (*EndpointConfig, error) {
	bd.mutex.RLock()
	bridge, exists := bd.bridges[networkID]
	bd.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("network not found: %s", networkID)
	}

	// 创建veth对
	vethHost := fmt.Sprintf("veth%s", containerID[:7])
	vethContainer := fmt.Sprintf("eth0")

	// 创建veth pair
	// #nosec G204 - vethHost和vethContainer是内部生成的安全标识符，固定命令用于网络配置
	cmd := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create veth pair: %v", err)
	}

	// 将host端连接到网桥
	// #nosec G204 - vethHost和bridge.Name都是内部生成的安全值，固定命令用于网络配置
	cmd = exec.Command("ip", "link", "set", vethHost, "master", bridge.Name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to attach veth to bridge: %v", err)
	}

	// 启用host端接口
	// #nosec G204 - vethHost是内部生成的安全标识符，固定命令用于网络配置
	cmd = exec.Command("ip", "link", "set", vethHost, "up")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to enable veth host: %v", err)
	}

	endpoint := &EndpointConfig{
		NetworkID:   networkID,
		ContainerID: containerID,
		Interface:   vethContainer,
		IPAddress:   "", // 将由IPAM分配
		Gateway:     bridge.Gateway,
	}

	fmt.Printf("创建网络端点: %s -> %s\n", containerID[:12], networkID[:12])
	return endpoint, nil
}

func (bd *BridgeDriver) Join(networkID, containerID string) error {
	// 将容器网络接口移动到容器命名空间
	vethContainer := "eth0"

	// 这里需要实际的容器PID来设置网络命名空间
	// 简化实现，实际需要从容器管理器获取PID
	fmt.Printf("加入网络: 容器 %s 加入网络 %s (接口: %s)\n", containerID[:12], networkID[:12], vethContainer)

	// 实际操作需要：
	// 1. 获取容器进程PID
	// 2. 将veth接口移动到容器网络命名空间
	// 3. 在容器内配置IP地址和路由

	return nil
}

func (bd *BridgeDriver) Leave(networkID, containerID string) error {
	fmt.Printf("离开网络: 容器 %s 离开网络 %s\n", containerID[:12], networkID[:12])
	return nil
}

func (bd *BridgeDriver) DeleteEndpoint(networkID, containerID string) error {
	vethHost := fmt.Sprintf("veth%s", containerID[:7])

	// 删除veth接口
	// #nosec G204 - vethHost是内部生成的安全标识符，固定命令用于网络清理
	cmd := exec.Command("ip", "link", "delete", vethHost)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete veth: %v", err)
	}

	fmt.Printf("删除网络端点: %s\n", containerID[:12])
	return nil
}

func (bd *BridgeDriver) DeleteNetwork(networkID string) error {
	bd.mutex.Lock()
	defer bd.mutex.Unlock()

	bridge, exists := bd.bridges[networkID]
	if !exists {
		return fmt.Errorf("network not found: %s", networkID)
	}

	// 删除网桥
	// #nosec G204 - bridge.Name是内部管理的网桥名称，固定命令用于网络清理
	cmd := exec.Command("ip", "link", "delete", bridge.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete bridge: %v", err)
	}

	delete(bd.bridges, networkID)
	fmt.Printf("删除桥接网络: %s\n", bridge.Name)
	return nil
}

// ==================
// 6. 安全管理系统
// ==================

// SecurityContext 安全上下文
type SecurityContext struct {
	RunAsUser                *int64
	RunAsGroup               *int64
	RunAsNonRoot             *bool
	ReadOnlyRootFilesystem   *bool
	AllowPrivilegeEscalation *bool
	Privileged               *bool
	Capabilities             *Capabilities
	SelinuxOptions           *SelinuxOptions
	WindowsOptions           *WindowsSecurityContextOptions
	FsGroup                  *int64
	SupplementalGroups       []int64
	SeccompProfile           *SeccompProfile
	AppArmorProfile          *AppArmorProfile
}

// Capabilities 权限能力
type Capabilities struct {
	Add  []string
	Drop []string
}

// SeccompManager Seccomp管理器
type SeccompManager struct {
	profiles map[string]*SeccompProfile
	mutex    sync.RWMutex
}

// SeccompProfile Seccomp配置文件
type SeccompProfile struct {
	Type             string
	LocalhostProfile *string
	DefaultAction    string
	Architectures    []string
	Syscalls         []SyscallRule
}

// SyscallRule 系统调用规则
type SyscallRule struct {
	Names  []string
	Action string
	Args   []SyscallArg
}

// SyscallArg 系统调用参数
type SyscallArg struct {
	Index    uint
	Value    uint64
	ValueTwo uint64
	Op       string
}

func NewSeccompManager() *SeccompManager {
	return &SeccompManager{
		profiles: make(map[string]*SeccompProfile),
	}
}

func (sm *SeccompManager) LoadProfile(name string, profile *SeccompProfile) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.profiles[name] = profile
	fmt.Printf("加载Seccomp配置文件: %s\n", name)
	return nil
}

func (sm *SeccompManager) ApplyProfile(containerID string, profileName string) error {
	sm.mutex.RLock()
	profile, exists := sm.profiles[profileName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("seccomp profile not found: %s", profileName)
	}

	// 使用profile防止未使用错误
	if profile == nil {
		return fmt.Errorf("seccomp profile is nil: %s", profileName)
	}

	// 应用Seccomp规则到容器
	fmt.Printf("应用Seccomp配置: 容器 %s 使用配置 %s\n", containerID[:12], profileName)

	// 实际实现需要：
	// 1. 将seccomp规则转换为BPF程序
	// 2. 通过prctl系统调用应用到进程
	// 3. 验证规则是否正确应用

	return nil
}

// ApparmorManager AppArmor管理器
type ApparmorManager struct {
	profiles map[string]*AppArmorProfile
	mutex    sync.RWMutex
}

// AppArmorProfile AppArmor配置文件
type AppArmorProfile struct {
	Type             string
	LocalhostProfile *string
	Rules            []string
}

func NewApparmorManager() *ApparmorManager {
	return &ApparmorManager{
		profiles: make(map[string]*AppArmorProfile),
	}
}

func (am *ApparmorManager) LoadProfile(name string, profile *AppArmorProfile) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.profiles[name] = profile
	fmt.Printf("加载AppArmor配置文件: %s\n", name)
	return nil
}

func (am *ApparmorManager) ApplyProfile(containerID string, profileName string) error {
	am.mutex.RLock()
	profile, exists := am.profiles[profileName]
	am.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("apparmor profile not found: %s", profileName)
	}

	// 使用profile防止未使用错误
	if profile == nil {
		return fmt.Errorf("apparmor profile is nil: %s", profileName)
	}

	fmt.Printf("应用AppArmor配置: 容器 %s 使用配置 %s\n", containerID[:12], profileName)
	return nil
}

// ==================
// 7. 容器编排引擎
// ==================

// ContainerOrchestrator 容器编排器
type ContainerOrchestrator struct {
	runtime     *ContainerRuntime
	scheduler   *ContainerScheduler
	serviceMgr  *ServiceManager
	deployments map[string]*Deployment
	services    map[string]*Service
	pods        map[string]*Pod
	nodes       map[string]*Node
	config      OrchestratorConfig
	eventBus    *ContainerEventBus
	monitor     *ClusterMonitor
	mutex       sync.RWMutex
	running     bool
}

// Pod 容器组
type Pod struct {
	ID             string
	Name           string
	Namespace      string
	Labels         map[string]string
	Annotations    map[string]string
	Containers     []*Container
	InitContainers []*Container
	Volumes        []*Volume
	RestartPolicy  RestartPolicy
	DNSPolicy      DNSPolicy
	NodeName       string
	Status         PodStatus
	CreatedAt      time.Time
	StartedAt      time.Time
}

// Service 服务
type Service struct {
	ID           string
	Name         string
	Namespace    string
	Type         ServiceType
	Selector     map[string]string
	Ports        []ServicePort
	ClusterIP    string
	ExternalIPs  []string
	LoadBalancer *LoadBalancerStatus
	CreatedAt    time.Time
}

// Deployment 部署
type Deployment struct {
	ID        string
	Name      string
	Namespace string
	Replicas  int32
	Selector  map[string]string
	Template  *PodTemplate
	Strategy  DeploymentStrategy
	Status    DeploymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Node 节点
type Node struct {
	ID          string
	Name        string
	Address     string
	Status      NodeStatus
	Capacity    ResourceList
	Allocatable ResourceList
	Conditions  []NodeCondition
	Info        NodeSystemInfo
	CreatedAt   time.Time
}

// ContainerScheduler 容器调度器
type ContainerScheduler struct {
	algorithms map[string]SchedulingAlgorithm
	policies   []SchedulingPolicy
	queue      *SchedulingQueue
	cache      *SchedulerCache
	mutex      sync.RWMutex
}

// SchedulingAlgorithm 调度算法接口
type SchedulingAlgorithm interface {
	Name() string
	Schedule(pod *Pod, nodes []*Node) (*Node, error)
	Preempt(pod *Pod, nodes []*Node) ([]*Pod, *Node, error)
}

func NewContainerOrchestrator(runtime *ContainerRuntime) *ContainerOrchestrator {
	return &ContainerOrchestrator{
		runtime:     runtime,
		scheduler:   NewContainerScheduler(),
		serviceMgr:  NewServiceManager(),
		deployments: make(map[string]*Deployment),
		services:    make(map[string]*Service),
		pods:        make(map[string]*Pod),
		nodes:       make(map[string]*Node),
		eventBus:    NewContainerEventBus(),
		monitor:     NewClusterMonitor(),
	}
}

func (co *ContainerOrchestrator) Start() error {
	co.mutex.Lock()
	defer co.mutex.Unlock()

	if co.running {
		return fmt.Errorf("orchestrator already running")
	}

	// 启动调度器
	go co.schedulingLoop()

	// 启动服务管理
	go co.serviceLoop()

	// 启动监控
	go co.monitorLoop()

	co.running = true
	fmt.Println("容器编排器已启动")
	return nil
}

func (co *ContainerOrchestrator) CreatePod(podSpec *PodSpec) (*Pod, error) {
	co.mutex.Lock()
	defer co.mutex.Unlock()

	pod := &Pod{
		ID:         generatePodID(),
		Name:       podSpec.Name,
		Namespace:  podSpec.Namespace,
		Labels:     podSpec.Labels,
		Containers: make([]*Container, 0),
		Status:     PodPending,
		CreatedAt:  time.Now(),
	}

	// 创建Pod中的容器
	for _, containerSpec := range podSpec.Containers {
		container, err := co.runtime.CreateContainer(&ContainerConfig{
			Image:      containerSpec.Image,
			Cmd:        containerSpec.Command,
			Env:        containerSpec.Env,
			WorkingDir: containerSpec.WorkingDir,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %v", err)
		}
		pod.Containers = append(pod.Containers, container)
	}

	co.pods[pod.ID] = pod
	fmt.Printf("创建Pod: %s (容器数: %d)\n", pod.Name, len(pod.Containers))

	// 提交给调度器
	go co.schedulePod(pod)

	return pod, nil
}

func (co *ContainerOrchestrator) schedulePod(pod *Pod) {
	// 获取可用节点
	nodes := co.getAvailableNodes()
	if len(nodes) == 0 {
		fmt.Printf("没有可用节点调度Pod: %s\n", pod.Name)
		return
	}

	// 选择调度算法
	algorithm := co.scheduler.getDefaultAlgorithm()

	// 执行调度
	selectedNode, err := algorithm.Schedule(pod, nodes)
	if err != nil {
		fmt.Printf("Pod调度失败: %s - %v\n", pod.Name, err)
		return
	}

	// 绑定到节点
	pod.NodeName = selectedNode.Name
	pod.Status = PodScheduled

	fmt.Printf("Pod调度成功: %s -> 节点 %s\n", pod.Name, selectedNode.Name)

	// 启动Pod中的容器
	go co.startPodContainers(pod)
}

func (co *ContainerOrchestrator) startPodContainers(pod *Pod) {
	pod.Status = PodRunning
	pod.StartedAt = time.Now()

	for _, container := range pod.Containers {
		if err := co.runtime.StartContainer(container.ID); err != nil {
			fmt.Printf("启动容器失败: %s - %v\n", container.ID[:12], err)
			pod.Status = PodFailed
			return
		}
	}

	fmt.Printf("Pod运行中: %s (节点: %s)\n", pod.Name, pod.NodeName)
}

func (co *ContainerOrchestrator) CreateDeployment(deploySpec *DeploymentSpec) (*Deployment, error) {
	co.mutex.Lock()
	defer co.mutex.Unlock()

	deployment := &Deployment{
		ID:        generateDeploymentID(),
		Name:      deploySpec.Name,
		Namespace: deploySpec.Namespace,
		Replicas:  deploySpec.Replicas,
		Selector:  deploySpec.Selector,
		Template:  deploySpec.Template,
		Status:    DeploymentProgressing,
		CreatedAt: time.Now(),
	}

	co.deployments[deployment.ID] = deployment
	fmt.Printf("创建Deployment: %s (副本数: %d)\n", deployment.Name, deployment.Replicas)

	// 创建副本Pod
	go co.createReplicaPods(deployment)

	return deployment, nil
}

func (co *ContainerOrchestrator) createReplicaPods(deployment *Deployment) {
	var createdPods int32

	for i := int32(0); i < deployment.Replicas; i++ {
		podSpec := &PodSpec{
			Name:       fmt.Sprintf("%s-%d", deployment.Name, i),
			Namespace:  deployment.Namespace,
			Labels:     deployment.Selector,
			Containers: deployment.Template.Spec.Containers,
		}

		pod, err := co.CreatePod(podSpec)
		if err != nil {
			fmt.Printf("创建副本Pod失败: %v\n", err)
			continue
		}

		createdPods++
		fmt.Printf("创建副本Pod: %s (%d/%d)\n", pod.Name, createdPods, deployment.Replicas)
	}

	deployment.Status = DeploymentAvailable
	deployment.UpdatedAt = time.Now()
}

func (co *ContainerOrchestrator) getAvailableNodes() []*Node {
	co.mutex.RLock()
	defer co.mutex.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range co.nodes {
		if node.Status == NodeReady {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func (co *ContainerOrchestrator) schedulingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for co.running {
		select {
		case <-ticker.C:
			co.reconcileState()
		}
	}
}

func (co *ContainerOrchestrator) serviceLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for co.running {
		select {
		case <-ticker.C:
			co.updateServices()
		}
	}
}

func (co *ContainerOrchestrator) monitorLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for co.running {
		select {
		case <-ticker.C:
			co.monitor.CollectMetrics()
		}
	}
}

func (co *ContainerOrchestrator) reconcileState() {
	// 确保期望状态与实际状态一致
	for _, deployment := range co.deployments {
		co.reconcileDeployment(deployment)
	}
}

func (co *ContainerOrchestrator) reconcileDeployment(deployment *Deployment) {
	// 统计当前运行的Pod数量
	runningPods := co.countRunningPodsForDeployment(deployment)

	if runningPods < deployment.Replicas {
		// 需要创建更多Pod
		needed := deployment.Replicas - runningPods
		fmt.Printf("Deployment %s 需要创建 %d 个Pod\n", deployment.Name, needed)
	} else if runningPods > deployment.Replicas {
		// 需要删除多余的Pod
		excess := runningPods - deployment.Replicas
		fmt.Printf("Deployment %s 需要删除 %d 个Pod\n", deployment.Name, excess)
	}
}

func (co *ContainerOrchestrator) countRunningPodsForDeployment(deployment *Deployment) int32 {
	var count int32
	for _, pod := range co.pods {
		if pod.Namespace == deployment.Namespace {
			// 检查标签选择器匹配
			if co.labelsMatch(pod.Labels, deployment.Selector) && pod.Status == PodRunning {
				count++
			}
		}
	}
	return count
}

func (co *ContainerOrchestrator) labelsMatch(podLabels, selector map[string]string) bool {
	for key, value := range selector {
		if podLabels[key] != value {
			return false
		}
	}
	return true
}

func (co *ContainerOrchestrator) updateServices() {
	// 更新服务端点
	for _, service := range co.services {
		endpoints := co.getServiceEndpoints(service)
		fmt.Printf("服务 %s 有 %d 个端点\n", service.Name, len(endpoints))
	}
}

func (co *ContainerOrchestrator) getServiceEndpoints(service *Service) []string {
	endpoints := make([]string, 0)

	for _, pod := range co.pods {
		if pod.Namespace == service.Namespace && pod.Status == PodRunning {
			if co.labelsMatch(pod.Labels, service.Selector) {
				// 获取Pod IP地址
				endpoints = append(endpoints, fmt.Sprintf("pod-%s", pod.ID[:12]))
			}
		}
	}

	return endpoints
}

// ==================
// 8. 辅助结构和函数
// ==================

// 事件系统
type ContainerEventBus struct {
	subscribers map[EventType][]EventHandler
	mutex       sync.RWMutex
}

type EventType int

const (
	EventContainerCreate EventType = iota
	EventContainerStart
	EventContainerStop
	EventContainerRemove
	EventContainerDie
	EventPodCreate
	EventPodSchedule
	EventPodStart
	EventPodStop
)

type ContainerEvent struct {
	Type      EventType
	Container *Container
	Pod       *Pod
	Message   string
	Timestamp time.Time
}

type EventHandler func(*ContainerEvent)

func NewContainerEventBus() *ContainerEventBus {
	return &ContainerEventBus{
		subscribers: make(map[EventType][]EventHandler),
	}
}

func (ceb *ContainerEventBus) Subscribe(eventType EventType, handler EventHandler) {
	ceb.mutex.Lock()
	defer ceb.mutex.Unlock()

	ceb.subscribers[eventType] = append(ceb.subscribers[eventType], handler)
}

func (ceb *ContainerEventBus) Publish(event *ContainerEvent) {
	ceb.mutex.RLock()
	handlers := ceb.subscribers[event.Type]
	ceb.mutex.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

// 监控组件
type ContainerMonitor struct {
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

type ClusterMonitor struct {
	nodeMetrics map[string]*NodeMetrics
	podMetrics  map[string]*PodMetrics
	mutex       sync.RWMutex
}

type NodeMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	NetworkIO   NetworkIOMetrics
	Timestamp   time.Time
}

type PodMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
	NetworkIO   NetworkIOMetrics
	Timestamp   time.Time
}

type NetworkIOMetrics struct {
	BytesReceived   int64
	BytesSent       int64
	PacketsReceived int64
	PacketsSent     int64
}

func NewContainerMonitor() *ContainerMonitor {
	return &ContainerMonitor{
		metrics: make(map[string]interface{}),
	}
}

func NewClusterMonitor() *ClusterMonitor {
	return &ClusterMonitor{
		nodeMetrics: make(map[string]*NodeMetrics),
		podMetrics:  make(map[string]*PodMetrics),
	}
}

func (cm *ClusterMonitor) CollectMetrics() {
	// 收集节点指标
	fmt.Println("收集集群指标...")

	// 模拟指标收集
	timestamp := time.Now()

	// 更新节点指标
	cm.mutex.Lock()
	cm.nodeMetrics["node-1"] = &NodeMetrics{
		CPUUsage:    65.5,
		MemoryUsage: 78.2,
		DiskUsage:   45.3,
		Timestamp:   timestamp,
	}
	cm.mutex.Unlock()
}

// 调度器组件
func NewContainerScheduler() *ContainerScheduler {
	cs := &ContainerScheduler{
		algorithms: make(map[string]SchedulingAlgorithm),
		policies:   make([]SchedulingPolicy, 0),
	}

	// 注册调度算法
	cs.algorithms["default"] = &DefaultSchedulingAlgorithm{}
	cs.algorithms["least-allocated"] = &LeastAllocatedAlgorithm{}

	return cs
}

func (cs *ContainerScheduler) getDefaultAlgorithm() SchedulingAlgorithm {
	return cs.algorithms["default"]
}

// 默认调度算法
type DefaultSchedulingAlgorithm struct{}

func (dsa *DefaultSchedulingAlgorithm) Name() string {
	return "default"
}

func (dsa *DefaultSchedulingAlgorithm) Schedule(pod *Pod, nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// 简单的轮询调度
	selectedNode := nodes[0]
	fmt.Printf("调度算法选择节点: %s\n", selectedNode.Name)

	return selectedNode, nil
}

func (dsa *DefaultSchedulingAlgorithm) Preempt(pod *Pod, nodes []*Node) ([]*Pod, *Node, error) {
	return nil, nil, fmt.Errorf("preemption not implemented")
}

// 最少分配调度算法
type LeastAllocatedAlgorithm struct{}

func (laa *LeastAllocatedAlgorithm) Name() string {
	return "least-allocated"
}

func (laa *LeastAllocatedAlgorithm) Schedule(pod *Pod, nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// 选择资源使用率最低的节点
	var bestNode *Node
	var lowestScore float64 = 100.0

	for _, node := range nodes {
		score := laa.calculateNodeScore(node)
		if score < lowestScore {
			lowestScore = score
			bestNode = node
		}
	}

	if bestNode == nil {
		return nodes[0], nil
	}

	fmt.Printf("最少分配算法选择节点: %s (得分: %.2f)\n", bestNode.Name, lowestScore)
	return bestNode, nil
}

func (laa *LeastAllocatedAlgorithm) calculateNodeScore(node *Node) float64 {
	// 简化的评分计算
	// 实际应该基于CPU和内存使用率
	return 50.0 // 模拟评分
}

func (laa *LeastAllocatedAlgorithm) Preempt(pod *Pod, nodes []*Node) ([]*Pod, *Node, error) {
	return nil, nil, fmt.Errorf("preemption not implemented")
}

// 各种枚举和结构定义
type RestartPolicy string
type DNSPolicy string
type ServiceType string
type PodStatus string
type DeploymentStatus string
type NodeStatus string

const (
	RestartPolicyAlways    RestartPolicy = "Always"
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	RestartPolicyNever     RestartPolicy = "Never"

	DNSClusterFirst DNSPolicy = "ClusterFirst"
	DNSDefault      DNSPolicy = "Default"

	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"

	PodPending   PodStatus = "Pending"
	PodScheduled PodStatus = "Scheduled"
	PodRunning   PodStatus = "Running"
	PodSucceeded PodStatus = "Succeeded"
	PodFailed    PodStatus = "Failed"

	DeploymentProgressing    DeploymentStatus = "Progressing"
	DeploymentAvailable      DeploymentStatus = "Available"
	DeploymentReplicaFailure DeploymentStatus = "ReplicaFailure"

	NodeReady    NodeStatus = "Ready"
	NodeNotReady NodeStatus = "NotReady"
)

// 各种规格和配置结构
type PodSpec struct {
	Name       string
	Namespace  string
	Labels     map[string]string
	Containers []ContainerSpec
}

type ContainerSpec struct {
	Name       string
	Image      string
	Command    []string
	Args       []string
	Env        []string
	WorkingDir string
}

type DeploymentSpec struct {
	Name      string
	Namespace string
	Replicas  int32
	Selector  map[string]string
	Template  *PodTemplate
}

type PodTemplate struct {
	Spec PodTemplateSpec
}

type PodTemplateSpec struct {
	Containers []ContainerSpec
}

type DeploymentStrategy struct {
	Type string
}

// 各种资源和配置
type ResourceConstraints struct {
	Memory string
	CPU    string
}

type ContainerStatistics struct {
	CPUUsage    float64
	MemoryUsage int64
	NetworkIO   NetworkIOMetrics
	BlockIO     BlockIOMetrics
}

type BlockIOMetrics struct {
	BytesRead    int64
	BytesWritten int64
	ReadsCount   int64
	WritesCount  int64
}

type Volume struct {
	Name     string
	Type     string
	Source   string
	Target   string
	ReadOnly bool
}

type Mount struct {
	Source      string
	Target      string
	Type        string
	Options     string
	Propagation string
}

type EndpointConfig struct {
	NetworkID   string
	ContainerID string
	Interface   string
	IPAddress   string
	Gateway     string
}

type NetworkConfig struct {
	Name   string
	Driver string
	IPAM   *NetworkIPAM
}

type ContainerVolume struct {
	Name      string
	Driver    string
	MountPath string
	Options   map[string]string
}

type HealthConfig struct {
	Test        []string
	Interval    time.Duration
	Timeout     time.Duration
	Retries     int
	StartPeriod time.Duration
}

type Health struct {
	Status        string
	FailingStreak int
	Log           []HealthcheckResult
}

type HealthcheckResult struct {
	Start    time.Time
	End      time.Time
	ExitCode int
	Output   string
}

// 各种Placeholder类型
type (
	RuntimeStatistics struct {
		ContainersCreated int64
		ContainersRunning int64
		ContainersStopped int64
		ImagesPulled      int64
		NetworksCreated   int64
		VolumesCreated    int64
	}
	OrchestratorConfig struct {
		ClusterName       string
		MaxNodes          int
		SchedulerPolicy   string
		MonitoringEnabled bool
	}
	NetworkConfigReference struct {
		Network string
	}
	SelinuxOptions struct {
		User  string
		Role  string
		Type  string
		Level string
	}
	WindowsSecurityContextOptions struct {
		GMSACredentialSpecName string
		GMSACredentialSpec     string
		RunAsUserName          string
	}
	ServicePort struct {
		Name       string
		Protocol   string
		Port       int32
		TargetPort int32
		NodePort   int32
	}
	LoadBalancerStatus struct {
		Ingress []LoadBalancerIngress
	}
	LoadBalancerIngress struct {
		IP       string
		Hostname string
	}
	NodeCondition struct {
		Type               string
		Status             string
		LastHeartbeatTime  time.Time
		LastTransitionTime time.Time
		Reason             string
		Message            string
	}
	NodeSystemInfo struct {
		MachineID     string
		SystemUUID    string
		BootID        string
		KernelVersion string
		OSImage       string
		Architecture  string
	}
	ResourceList     map[string]string
	SchedulingPolicy struct {
		Name    string
		Weight  int
		Enabled bool
	}
	SchedulingQueue struct {
		pods  []*Pod
		mutex sync.Mutex
	}
	SchedulerCache struct {
		nodes map[string]*Node
		pods  map[string]*Pod
		mutex sync.RWMutex
	}
	ServiceManager struct {
		services map[string]*Service
		mutex    sync.RWMutex
	}
	IPAddressManager struct {
		pools map[string]*IPPool
		mutex sync.RWMutex
	}
	IPPool struct {
		Subnet    string
		Gateway   string
		Allocated map[string]bool
		Available []string
	}
)

// 构造函数
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]*Service),
	}
}
func NewIPAddressManager() *IPAddressManager {
	return &IPAddressManager{
		pools: make(map[string]*IPPool),
	}
}

// ==================
// Host网络驱动实现
// ==================

// HostDriver 主机网络驱动
type HostDriver struct{}

func (hd *HostDriver) Name() string {
	return "host"
}

func (hd *HostDriver) CreateNetwork(config *NetworkConfig) (*ContainerNetwork, error) {
	network := &ContainerNetwork{
		ID:         generateNetworkID(),
		Name:       config.Name,
		Driver:     "host",
		Containers: make(map[string]*EndpointConfig),
		Options:    make(map[string]string),
		Labels:     make(map[string]string),
		Created:    time.Now(),
	}
	fmt.Printf("创建主机网络: %s\n", config.Name)
	return network, nil
}

func (hd *HostDriver) DeleteNetwork(networkID string) error {
	fmt.Printf("删除主机网络: %s\n", networkID)
	return nil
}

func (hd *HostDriver) CreateEndpoint(networkID, containerID string) (*EndpointConfig, error) {
	endpoint := &EndpointConfig{
		NetworkID:   networkID,
		ContainerID: containerID,
		Interface:   "host",
	}
	fmt.Printf("创建主机网络端点: %s\n", containerID[:12])
	return endpoint, nil
}

func (hd *HostDriver) DeleteEndpoint(networkID, containerID string) error {
	fmt.Printf("删除主机网络端点: %s\n", containerID[:12])
	return nil
}

func (hd *HostDriver) Join(networkID, containerID string) error {
	fmt.Printf("加入主机网络: %s\n", containerID[:12])
	return nil
}

func (hd *HostDriver) Leave(networkID, containerID string) error {
	fmt.Printf("离开主机网络: %s\n", containerID[:12])
	return nil
}

// ==================
// Overlay网络驱动实现
// ==================

// OverlayDriver 覆盖网络驱动
type OverlayDriver struct {
	networks map[string]*ContainerNetwork
	mutex    sync.RWMutex
}

func (od *OverlayDriver) Name() string {
	return "overlay"
}

func (od *OverlayDriver) CreateNetwork(config *NetworkConfig) (*ContainerNetwork, error) {
	if od.networks == nil {
		od.networks = make(map[string]*ContainerNetwork)
	}

	networkID := generateNetworkID()
	network := &ContainerNetwork{
		ID:         networkID,
		Name:       config.Name,
		Driver:     "overlay",
		IPAM:       config.IPAM,
		Containers: make(map[string]*EndpointConfig),
		Options:    make(map[string]string),
		Labels:     make(map[string]string),
		Created:    time.Now(),
	}

	od.mutex.Lock()
	od.networks[networkID] = network
	od.mutex.Unlock()

	fmt.Printf("创建覆盖网络: %s\n", config.Name)
	return network, nil
}

func (od *OverlayDriver) DeleteNetwork(networkID string) error {
	od.mutex.Lock()
	defer od.mutex.Unlock()

	if _, exists := od.networks[networkID]; !exists {
		return fmt.Errorf("overlay network not found: %s", networkID)
	}

	delete(od.networks, networkID)
	fmt.Printf("删除覆盖网络: %s\n", networkID)
	return nil
}

func (od *OverlayDriver) CreateEndpoint(networkID, containerID string) (*EndpointConfig, error) {
	endpoint := &EndpointConfig{
		NetworkID:   networkID,
		ContainerID: containerID,
		Interface:   "eth0",
	}
	fmt.Printf("创建覆盖网络端点: %s\n", containerID[:12])
	return endpoint, nil
}

func (od *OverlayDriver) DeleteEndpoint(networkID, containerID string) error {
	fmt.Printf("删除覆盖网络端点: %s\n", containerID[:12])
	return nil
}

func (od *OverlayDriver) Join(networkID, containerID string) error {
	fmt.Printf("加入覆盖网络: %s\n", containerID[:12])
	return nil
}

func (od *OverlayDriver) Leave(networkID, containerID string) error {
	fmt.Printf("离开覆盖网络: %s\n", containerID[:12])
	return nil
}

// ==================
// AUFS存储驱动实现
// ==================

// AufsDriver AUFS存储驱动
type AufsDriver struct {
	root      string
	layersDir string
	diffsDir  string
}

func (ad *AufsDriver) Name() string {
	return "aufs"
}

func (ad *AufsDriver) Initialize(root string) error {
	ad.root = root
	ad.layersDir = filepath.Join(root, "aufs")
	ad.diffsDir = filepath.Join(ad.layersDir, "diff")

	dirs := []string{ad.layersDir, ad.diffsDir}
	for _, dir := range dirs {
		// #nosec G301 -- AUFS驱动系统目录，需要0755权限支持Docker镜像层管理
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	fmt.Printf("初始化AUFS驱动: %s\n", root)
	return nil
}

func (ad *AufsDriver) CreateLayer(id string, parent string) (*Layer, error) {
	layerDir := filepath.Join(ad.layersDir, id)
	diffDir := filepath.Join(ad.diffsDir, id)

	dirs := []string{layerDir, diffDir}
	for _, dir := range dirs {
		// #nosec G301 -- AUFS镜像层目录，需要0755权限支持容器文件系统操作
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	layer := &Layer{
		ID:        id,
		Parent:    parent,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	fmt.Printf("创建AUFS层: %s\n", id)
	return layer, nil
}

func (ad *AufsDriver) MountLayer(id string, mountPoint string) error {
	fmt.Printf("挂载AUFS层: %s -> %s\n", id, mountPoint)
	return nil
}

func (ad *AufsDriver) UnmountLayer(id string) error {
	fmt.Printf("卸载AUFS层: %s\n", id)
	return nil
}

func (ad *AufsDriver) GetLayer(id string) (*Layer, error) {
	layerDir := filepath.Join(ad.layersDir, id)
	if _, err := os.Stat(layerDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("layer not found: %s", id)
	}
	return &Layer{
		ID:       id,
		Metadata: make(map[string]interface{}),
	}, nil
}

func (ad *AufsDriver) GetLayerSize(id string) (int64, error) {
	layerDir := filepath.Join(ad.diffsDir, id)
	return calculateDirectorySize(layerDir)
}

func (ad *AufsDriver) RemoveLayer(id string) error {
	layerDir := filepath.Join(ad.layersDir, id)
	return os.RemoveAll(layerDir)
}

func (ad *AufsDriver) Cleanup() error {
	fmt.Println("清理AUFS驱动")
	return nil
}

// ==================
// DeviceMapper存储驱动实现
// ==================

// DeviceMapperDriver DeviceMapper存储驱动
type DeviceMapperDriver struct {
	root       string
	deviceRoot string
	poolName   string
}

func (dmd *DeviceMapperDriver) Name() string {
	return "devicemapper"
}

func (dmd *DeviceMapperDriver) Initialize(root string) error {
	dmd.root = root
	dmd.deviceRoot = filepath.Join(root, "devicemapper")
	dmd.poolName = "docker-pool"

	// #nosec G301 -- DeviceMapper驱动系统目录，需要0755权限支持块设备管理
	if err := os.MkdirAll(dmd.deviceRoot, 0755); err != nil {
		return err
	}

	fmt.Printf("初始化DeviceMapper驱动: %s\n", root)
	return nil
}

func (dmd *DeviceMapperDriver) CreateLayer(id string, parent string) (*Layer, error) {
	layer := &Layer{
		ID:        id,
		Parent:    parent,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	fmt.Printf("创建DeviceMapper层: %s\n", id)
	return layer, nil
}

func (dmd *DeviceMapperDriver) MountLayer(id string, mountPoint string) error {
	fmt.Printf("挂载DeviceMapper层: %s -> %s\n", id, mountPoint)
	return nil
}

func (dmd *DeviceMapperDriver) UnmountLayer(id string) error {
	fmt.Printf("卸载DeviceMapper层: %s\n", id)
	return nil
}

func (dmd *DeviceMapperDriver) GetLayer(id string) (*Layer, error) {
	return &Layer{
		ID:       id,
		Metadata: make(map[string]interface{}),
	}, nil
}

func (dmd *DeviceMapperDriver) GetLayerSize(id string) (int64, error) {
	return 0, nil
}

func (dmd *DeviceMapperDriver) RemoveLayer(id string) error {
	fmt.Printf("删除DeviceMapper层: %s\n", id)
	return nil
}

func (dmd *DeviceMapperDriver) Cleanup() error {
	fmt.Println("清理DeviceMapper驱动")
	return nil
}

// 辅助函数
func generateContainerID() string {
	return fmt.Sprintf("container_%d_%d", time.Now().UnixNano(), secureRandomInt63())
}

func generateContainerName() string {
	adjectives := []string{"happy", "clever", "brave", "gentle", "bright"}
	nouns := []string{"tiger", "eagle", "dolphin", "phoenix", "dragon"}

	adj := adjectives[secureRandomInt(len(adjectives))]
	noun := nouns[secureRandomInt(len(nouns))]

	return fmt.Sprintf("%s_%s", adj, noun)
}

func generateNetworkID() string {
	return fmt.Sprintf("network_%d_%d", time.Now().UnixNano(), secureRandomInt63())
}

func generatePodID() string {
	return fmt.Sprintf("pod_%d_%d", time.Now().UnixNano(), secureRandomInt63())
}

func generateDeploymentID() string {
	return fmt.Sprintf("deployment_%d_%d", time.Now().UnixNano(), secureRandomInt63())
}

func generateShortID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based ID if random generation fails
		return fmt.Sprintf("%x", time.Now().UnixNano())[:12]
	}
	return fmt.Sprintf("%x", bytes)[:12]
}

func calculateDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// ==================
// 9. 主演示函数
// ==================

func demonstrateVirtualizationContainers() {
	fmt.Println("=== Go虚拟化与容器大师演示 ===")

	// 1. 容器运行时演示
	fmt.Println("\n1. 容器运行时初始化")
	config := RuntimeConfig{
		RootDirectory:      "/var/lib/container-runtime",
		StateDirectory:     "/var/run/container-runtime",
		LogLevel:           "info",
		MaxContainers:      100,
		DefaultRuntime:     "runc",
		EnableSelinux:      false,
		EnableApparmor:     true,
		EnableSeccomp:      true,
		DefaultNetworkMode: "bridge",
		StorageDriver:      "overlay2",
		CgroupVersion:      2,
		PidsLimit:          1024,
		ShmSize:            64 * 1024 * 1024,
	}

	runtime := NewContainerRuntime(config)
	if err := runtime.Start(); err != nil {
		fmt.Printf("启动运行时失败: %v\n", err)
		return
	}

	// 2. 镜像管理演示
	fmt.Println("\n2. 容器镜像管理")

	// 创建示例镜像
	image := &ContainerImage{
		ID:       "image_123456",
		RepoTags: []string{"demo:latest"},
		Created:  time.Now(),
		Size:     100 * 1024 * 1024, // 100MB
		Layers:   []string{"layer_001", "layer_002", "layer_003"},
		Config: &ImageConfig{
			Cmd:        []string{"/bin/sh"},
			Env:        []string{"PATH=/usr/bin:/bin"},
			WorkingDir: "/",
		},
	}

	runtime.images[image.ID] = image
	fmt.Printf("加载镜像: %s (大小: %d MB)\n", image.RepoTags[0], image.Size/1024/1024)

	// 3. 容器生命周期演示
	fmt.Println("\n3. 容器生命周期管理")

	containerConfig := &ContainerConfig{
		Image:      image.ID,
		Cmd:        []string{"/bin/sleep", "60"},
		Env:        []string{"HOME=/root", "USER=root"},
		WorkingDir: "/",
		Hostname:   "demo-container",
	}

	// 创建容器
	container, err := runtime.CreateContainer(containerConfig)
	if err != nil {
		fmt.Printf("创建容器失败: %v\n", err)
		return
	}

	// 启动容器
	if err := runtime.StartContainer(container.ID); err != nil {
		fmt.Printf("启动容器失败: %v\n", err)
		return
	}

	fmt.Printf("容器状态: %s (PID: %d)\n", container.State.Status, container.State.Pid)

	// 4. 网络管理演示
	fmt.Println("\n4. 容器网络管理")

	networkConfig := &NetworkConfig{
		Name:   "demo-network",
		Driver: "bridge",
		IPAM: &NetworkIPAM{
			Driver: "default",
			Config: []IPAMConfig{
				{
					Subnet:  "172.20.0.0/16",
					Gateway: "172.20.0.1",
				},
			},
		},
	}

	network, err := runtime.network.CreateNetwork(networkConfig)
	if err != nil {
		fmt.Printf("创建网络失败: %v\n", err)
	} else {
		fmt.Printf("创建网络: %s (子网: %s)\n", network.Name, networkConfig.IPAM.Config[0].Subnet)
	}

	// 5. 资源限制演示
	fmt.Println("\n5. 资源限制和Cgroup管理")

	// 设置内存限制
	if memoryCgroup, exists := container.Cgroups["memory"]; exists {
		if err := runtime.cgroups.SetMemoryLimit(memoryCgroup, 128*1024*1024); err != nil { // 128MB
			log.Printf("Warning: failed to set memory limit: %v", err)
		} else {
			fmt.Printf("设置内存限制: 128MB\n")
		}
	}

	// 设置CPU限制
	if cpuCgroup, exists := container.Cgroups["cpu"]; exists {
		if err := runtime.cgroups.SetCPUQuota(cpuCgroup, 50000, 100000); err != nil { // 50%
			log.Printf("Warning: failed to set CPU quota: %v", err)
		} else {
			fmt.Printf("设置CPU限制: 50%%\n")
		}
	}

	// 6. 安全管理演示
	fmt.Println("\n6. 安全管理和隔离")

	// Seccomp配置
	seccompProfile := &SeccompProfile{
		Type:          "RuntimeDefault",
		DefaultAction: "SCMP_ACT_ERRNO",
		Syscalls: []SyscallRule{
			{
				Names:  []string{"read", "write", "open", "close"},
				Action: "SCMP_ACT_ALLOW",
			},
		},
	}

	if err := runtime.seccomp.LoadProfile("demo-profile", seccompProfile); err != nil {
		log.Printf("Warning: failed to load seccomp profile: %v", err)
	}
	if err := runtime.seccomp.ApplyProfile(container.ID, "demo-profile"); err != nil {
		log.Printf("Warning: failed to apply seccomp profile: %v", err)
	}

	// AppArmor配置
	apparmorProfile := &AppArmorProfile{
		Type: "RuntimeDefault",
		Rules: []string{
			"deny network raw",
			"deny mount",
		},
	}

	if err := runtime.apparmor.LoadProfile("demo-apparmor", apparmorProfile); err != nil {
		log.Printf("Warning: failed to load apparmor profile: %v", err)
	}
	if err := runtime.apparmor.ApplyProfile(container.ID, "demo-apparmor"); err != nil {
		log.Printf("Warning: failed to apply apparmor profile: %v", err)
	}

	// 7. 容器编排演示
	fmt.Println("\n7. 容器编排和调度")

	orchestrator := NewContainerOrchestrator(runtime)
	if err := orchestrator.Start(); err != nil {
		fmt.Printf("启动编排器失败: %v\n", err)
		return
	}

	// 添加节点
	node := &Node{
		ID:      "node-1",
		Name:    "worker-node-1",
		Address: "192.168.1.100",
		Status:  NodeReady,
		Capacity: ResourceList{
			"cpu":    "4",
			"memory": "8Gi",
		},
		CreatedAt: time.Now(),
	}

	orchestrator.nodes[node.ID] = node
	fmt.Printf("添加节点: %s (CPU: %s, 内存: %s)\n", node.Name, node.Capacity["cpu"], node.Capacity["memory"])

	// 创建Pod
	podSpec := &PodSpec{
		Name:      "demo-pod",
		Namespace: "default",
		Labels:    map[string]string{"app": "demo"},
		Containers: []ContainerSpec{
			{
				Name:    "web-server",
				Image:   "nginx:latest",
				Command: []string{"nginx", "-g", "daemon off;"},
			},
		},
	}

	pod, err := orchestrator.CreatePod(podSpec)
	if err != nil {
		fmt.Printf("创建Pod失败: %v\n", err)
	} else {
		fmt.Printf("创建Pod: %s (容器数: %d)\n", pod.Name, len(pod.Containers))
	}

	// 创建Deployment
	deploymentSpec := &DeploymentSpec{
		Name:      "web-deployment",
		Namespace: "default",
		Replicas:  3,
		Selector:  map[string]string{"app": "web"},
		Template: &PodTemplate{
			Spec: PodTemplateSpec{
				Containers: []ContainerSpec{
					{
						Name:  "web",
						Image: "nginx:latest",
					},
				},
			},
		},
	}

	deployment, err := orchestrator.CreateDeployment(deploymentSpec)
	if err != nil {
		fmt.Printf("创建Deployment失败: %v\n", err)
	} else {
		fmt.Printf("创建Deployment: %s (副本数: %d)\n", deployment.Name, deployment.Replicas)
	}

	// 8. 监控和指标演示
	fmt.Println("\n8. 容器监控和指标收集")

	// 收集容器统计信息
	if memoryCgroup, exists := container.Cgroups["memory"]; exists {
		stats, err := runtime.cgroups.GetStats(memoryCgroup)
		if err == nil {
			fmt.Printf("容器资源使用情况:\n")
			if memStats, ok := stats["memory"].(map[string]int64); ok {
				for key, value := range memStats {
					if key == "anon" || key == "file" {
						fmt.Printf("  %s: %d KB\n", key, value/1024)
					}
				}
			}
		}
	}

	// 9. 存储卷演示
	fmt.Println("\n9. 存储卷和持久化")

	volume := &ContainerVolume{
		Name:      "data-volume",
		Driver:    "local",
		MountPath: "/data",
		Options: map[string]string{
			"type":   "bind",
			"source": "/host/data",
		},
	}

	runtime.volumes[volume.Name] = volume
	fmt.Printf("创建存储卷: %s -> %s\n", volume.Name, volume.MountPath)

	// 10. 事件和日志演示
	fmt.Println("\n10. 事件系统和日志管理")

	// 订阅容器事件
	runtime.eventBus.Subscribe(EventContainerStart, func(event *ContainerEvent) {
		fmt.Printf("📢 事件通知: 容器 %s 已启动\n", event.Container.ID[:12])
	})

	runtime.eventBus.Subscribe(EventContainerStop, func(event *ContainerEvent) {
		fmt.Printf("📢 事件通知: 容器 %s 已停止\n", event.Container.ID[:12])
	})

	// 让系统运行一段时间
	fmt.Println("\n监控运行状态...")
	time.Sleep(5 * time.Second)

	// 11. 清理演示
	fmt.Println("\n11. 资源清理")

	// 停止容器
	if err := runtime.StopContainer(container.ID, 10*time.Second); err != nil {
		log.Printf("Warning: failed to stop container: %v", err)
	}

	// 删除容器
	if err := runtime.RemoveContainer(container.ID, false); err != nil {
		log.Printf("Warning: failed to remove container: %v", err)
	}

	fmt.Println("\n=== 虚拟化与容器演示完成 ===")
}

func main() {
	demonstrateVirtualizationContainers()

	fmt.Println("\n=== Go虚拟化与容器大师演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 容器运行时：生命周期管理、进程隔离、资源控制")
	fmt.Println("2. Linux命名空间：PID、网络、文件系统、用户隔离")
	fmt.Println("3. Cgroups资源管理：CPU、内存、I/O限制和统计")
	fmt.Println("4. 存储驱动：OverlayFS、AUFS、设备映射器")
	fmt.Println("5. 网络管理：桥接、主机、覆盖网络驱动")
	fmt.Println("6. 安全隔离：Seccomp、AppArmor、权限控制")
	fmt.Println("7. 容器编排：Pod调度、服务发现、负载均衡")
	fmt.Println("8. 集群管理：节点管理、资源调度、故障恢复")

	fmt.Println("\n高级虚拟化特性:")
	fmt.Println("- 微服务架构和服务网格")
	fmt.Println("- 无服务器容器和FaaS")
	fmt.Println("- 容器镜像优化和安全扫描")
	fmt.Println("- 多租户隔离和资源配额")
	fmt.Println("- 实时迁移和零停机部署")
	fmt.Println("- 混合云和多云容器管理")
	fmt.Println("- AI/ML工作负载容器化")
}

/*
=== 练习题 ===

1. 容器运行时增强：
   - 实现OCI兼容的运行时接口
   - 添加Windows容器支持
   - 创建GPU资源管理
   - 实现容器镜像构建功能

2. 高级网络功能：
   - 实现CNI插件接口
   - 添加服务网格支持
   - 创建网络策略引擎
   - 实现多租户网络隔离

3. 存储系统扩展：
   - 实现CSI存储接口
   - 添加分布式存储支持
   - 创建快照和备份机制
   - 实现存储QoS控制

4. 安全强化：
   - 实现零信任安全模型
   - 添加镜像签名验证
   - 创建运行时安全监控
   - 实现合规性检查

5. 编排优化：
   - 实现多集群管理
   - 添加自动伸缩功能
   - 创建灾难恢复机制
   - 实现成本优化调度

重要概念：
- Container Runtime: 容器运行时和OCI标准
- Linux Namespaces: 进程隔离和资源视图
- Control Groups: 资源限制和统计
- Container Networking: 容器网络模型
- Image Management: 镜像分层和存储
- Orchestration: 容器编排和调度
- Security: 容器安全和隔离
*/
