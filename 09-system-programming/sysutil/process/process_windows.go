//go:build windows
// +build windows

/*
Windows 平台的进程管理实现

本文件实现了 Windows 系统上的进程信息获取功能：
  - 使用 Windows API 获取进程列表
  - 通过 CreateToolhelp32Snapshot 枚举进程
  - 支持进程状态、内存、CPU 等信息的获取

注意事项：
  - 某些操作需要管理员权限
  - Windows 的进程状态模型与 Unix 不同
*/
package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Windows API 常量
const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	TH32CS_SNAPPROCESS        = 0x00000002
	MAX_PATH                  = 260
)

// Windows API 结构体
type PROCESSENTRY32 struct {
	Size              uint32
	Usage             uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	Threads           uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

type PROCESS_MEMORY_COUNTERS struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

type FILETIME struct {
	LowDateTime  uint32
	HighDateTime uint32
}

// Windows API 函数
var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	psapi                          = syscall.NewLazyDLL("psapi.dll")
	procCreateToolhelp32Snapshot   = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First             = kernel32.NewProc("Process32FirstW")
	procProcess32Next              = kernel32.NewProc("Process32NextW")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
	procGetProcessMemoryInfo       = psapi.NewProc("GetProcessMemoryInfo")
	procGetProcessTimes            = kernel32.NewProc("GetProcessTimes")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
)

// listProcesses 获取所有进程列表（Windows实现）
func listProcesses() ([]*Info, error) {
	// 创建进程快照
	handle, _, err := procCreateToolhelp32Snapshot.Call(
		uintptr(TH32CS_SNAPPROCESS),
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %v", err)
	}
	defer procCloseHandle.Call(handle)

	var procs []*Info
	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))

	// 获取第一个进程
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	for {
		info := &Info{
			PID:        int(entry.ProcessID),
			PPID:       int(entry.ParentProcessID),
			Name:       syscall.UTF16ToString(entry.ExeFile[:]),
			NumThreads: int(entry.Threads),
			State:      StateRunning, // Windows 进程默认为运行状态
		}

		// 获取更多详细信息
		if detailed, err := getProcessDetails(int(entry.ProcessID)); err == nil {
			info.Executable = detailed.Executable
			info.Username = detailed.Username
			info.MemoryInfo = detailed.MemoryInfo
			info.CPUPercent = detailed.CPUPercent
			info.CreateTime = detailed.CreateTime
		}

		procs = append(procs, info)

		// 获取下一个进程
		ret, _, _ = procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return procs, nil
}

// getProcessInfo 获取指定进程的详细信息（Windows实现）
func getProcessInfo(pid int) (*Info, error) {
	// 首先从进程列表中获取基本信息
	procs, err := listProcesses()
	if err != nil {
		return nil, err
	}

	for _, p := range procs {
		if p.PID == pid {
			return p, nil
		}
	}

	return nil, ErrProcessNotFound
}

// getProcessDetails 获取进程详细信息
func getProcessDetails(pid int) (*Info, error) {
	info := &Info{PID: pid}

	// 打开进程句柄
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		0,
		uintptr(pid),
	)
	if handle == 0 {
		// 尝试使用较少的权限
		handle, _, err = procOpenProcess.Call(
			uintptr(PROCESS_QUERY_INFORMATION),
			0,
			uintptr(pid),
		)
		if handle == 0 {
			return nil, fmt.Errorf("OpenProcess failed: %v", err)
		}
	}
	defer procCloseHandle.Call(handle)

	// 获取可执行文件路径
	var exePath [MAX_PATH * 2]uint16
	size := uint32(MAX_PATH * 2)
	ret, _, _ := procQueryFullProcessImageNameW.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&exePath[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		info.Executable = syscall.UTF16ToString(exePath[:size])
	}

	// 获取内存信息
	var memCounters PROCESS_MEMORY_COUNTERS
	memCounters.cb = uint32(unsafe.Sizeof(memCounters))
	ret, _, _ = procGetProcessMemoryInfo.Call(
		handle,
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.cb),
	)
	if ret != 0 {
		info.MemoryInfo.RSS = uint64(memCounters.WorkingSetSize)
		info.MemoryInfo.VMS = uint64(memCounters.PagefileUsage)
	}

	// 获取进程时间
	var creationTime, exitTime, kernelTime, userTime FILETIME
	ret, _, _ = procGetProcessTimes.Call(
		handle,
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if ret != 0 {
		info.CreateTime = filetimeToTime(creationTime)
		info.CPUPercent = calculateWindowsCPUPercent(kernelTime, userTime, info.CreateTime)
	}

	// 获取用户名（使用 WMI 或其他方法）
	info.Username = getProcessUsername(pid)

	return info, nil
}

// filetimeToTime 将 Windows FILETIME 转换为 Go time.Time
func filetimeToTime(ft FILETIME) time.Time {
	// FILETIME 是从 1601-01-01 开始的 100 纳秒间隔数
	nsec := int64(ft.HighDateTime)<<32 + int64(ft.LowDateTime)
	// 转换为 Unix 时间戳（从 1970-01-01 开始）
	// 1601 到 1970 之间的 100 纳秒间隔数
	const epochDiff = 116444736000000000
	nsec -= epochDiff
	return time.Unix(0, nsec*100)
}

// calculateWindowsCPUPercent 计算 Windows 进程的 CPU 使用率
func calculateWindowsCPUPercent(kernelTime, userTime FILETIME, createTime time.Time) float64 {
	if createTime.IsZero() {
		return 0
	}

	// 计算总 CPU 时间（100 纳秒单位）
	kernel := int64(kernelTime.HighDateTime)<<32 + int64(kernelTime.LowDateTime)
	user := int64(userTime.HighDateTime)<<32 + int64(userTime.LowDateTime)
	totalCPU := float64(kernel+user) / 10000000 // 转换为秒

	// 进程运行时间
	elapsed := time.Since(createTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	return (totalCPU / elapsed) * 100
}

// getProcessUsername 获取进程的用户名
func getProcessUsername(pid int) string {
	// 使用 wmic 命令获取用户名
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "Owner", "/value")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Owner=") {
			return strings.TrimPrefix(line, "Owner=")
		}
	}

	return ""
}

// getCommandLine 获取进程命令行（Windows）
func getCommandLine(pid int) string {
	// 使用 wmic 命令获取命令行
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "CommandLine", "/value")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CommandLine=") {
			return strings.TrimPrefix(line, "CommandLine=")
		}
	}

	return ""
}

// getTotalMemory 获取系统总内存（Windows）
func getTotalMemory() uint64 {
	// 使用 wmic 获取总内存
	cmd := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory", "/value")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalPhysicalMemory=") {
			value := strings.TrimPrefix(line, "TotalPhysicalMemory=")
			mem, _ := strconv.ParseUint(value, 10, 64)
			return mem
		}
	}

	return 0
}

// lookupUsername 根据 SID 查找用户名（Windows）
func lookupUsername(uid int) string {
	// Windows 不使用数字 UID，这个函数在 Windows 上不适用
	return ""
}
