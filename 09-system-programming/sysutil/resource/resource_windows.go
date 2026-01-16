//go:build windows
// +build windows

/*
Windows 平台的资源管理实现

本文件实现了 Windows 系统上的资源信息获取功能：
  - 使用 Windows API 获取系统信息
  - 支持内存、CPU、磁盘等资源的监控

注意事项：
  - 某些操作需要管理员权限
  - Windows 的资源模型与 Unix 不同
*/
package resource

import (
	"fmt"
	"os/exec"
	"runtime"
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
)

// Windows API 结构体
type MEMORYSTATUSEX struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type SYSTEM_INFO struct {
	ProcessorArchitecture     uint16
	Reserved                  uint16
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uintptr
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

// Windows API 函数
var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	psapi                    = syscall.NewLazyDLL("psapi.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procGetDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetLogicalDrives     = kernel32.NewProc("GetLogicalDrives")
	procGetDriveTypeW        = kernel32.NewProc("GetDriveTypeW")
)

// GetMemoryInfo 获取系统内存信息（Windows实现）
func GetMemoryInfo() (*MemoryInfo, error) {
	var memStatus MEMORYSTATUSEX
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))

	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return nil, fmt.Errorf("GlobalMemoryStatusEx failed: %v", err)
	}

	info := &MemoryInfo{
		Total:       memStatus.TotalPhys,
		Available:   memStatus.AvailPhys,
		Used:        memStatus.TotalPhys - memStatus.AvailPhys,
		Free:        memStatus.AvailPhys,
		UsedPercent: float64(memStatus.MemoryLoad),
		SwapTotal:   memStatus.TotalPageFile - memStatus.TotalPhys,
		SwapFree:    memStatus.AvailPageFile - memStatus.AvailPhys,
	}

	if info.SwapTotal > 0 {
		info.SwapUsed = info.SwapTotal - info.SwapFree
	}

	return info, nil
}

// GetCPUInfo 获取CPU信息（Windows实现）
func GetCPUInfo() (*CPUInfo, error) {
	info := &CPUInfo{
		NumCPU:        runtime.NumCPU(),
		NumLogicalCPU: runtime.NumCPU(),
	}

	// 使用 WMIC 获取 CPU 信息
	output, err := runWMIC("cpu", "get", "Name,Manufacturer,NumberOfCores,NumberOfLogicalProcessors,MaxClockSpeed")
	if err == nil {
		lines := strings.Split(output, "\n")
		if len(lines) >= 2 {
			// 解析第二行（数据行）
			fields := strings.Fields(lines[1])
			if len(fields) >= 1 {
				info.ModelName = strings.Join(fields[:len(fields)-4], " ")
			}
		}
	}

	// 获取更详细的信息
	if nameOutput, err := runWMIC("cpu", "get", "Name", "/value"); err == nil {
		if name := parseWMICValue(nameOutput, "Name"); name != "" {
			info.ModelName = name
		}
	}

	if vendorOutput, err := runWMIC("cpu", "get", "Manufacturer", "/value"); err == nil {
		if vendor := parseWMICValue(vendorOutput, "Manufacturer"); vendor != "" {
			info.Vendor = vendor
		}
	}

	if mhzOutput, err := runWMIC("cpu", "get", "MaxClockSpeed", "/value"); err == nil {
		if mhz := parseWMICValue(mhzOutput, "MaxClockSpeed"); mhz != "" {
			info.MHz, _ = strconv.ParseFloat(mhz, 64)
		}
	}

	return info, nil
}

// CPU 时间记录（用于计算使用率）
var lastCPUTimes struct {
	idle, kernel, user uint64
	timestamp          time.Time
}

// GetCPUUsage 获取CPU使用率（Windows实现）
func GetCPUUsage() (*CPUUsage, error) {
	// 使用 WMIC 获取 CPU 使用率
	output, err := runWMIC("cpu", "get", "LoadPercentage", "/value")
	if err != nil {
		return nil, err
	}

	usage := &CPUUsage{}

	if load := parseWMICValue(output, "LoadPercentage"); load != "" {
		usage.Total, _ = strconv.ParseFloat(load, 64)
		usage.Idle = 100 - usage.Total
		// Windows 不容易区分 user 和 system，这里做估算
		usage.User = usage.Total * 0.7
		usage.System = usage.Total * 0.3
	}

	return usage, nil
}

// GetDiskInfo 获取磁盘信息（Windows实现）
func GetDiskInfo() ([]DiskInfo, error) {
	var disks []DiskInfo

	// 获取逻辑驱动器
	drives, _, _ := procGetLogicalDrives.Call()
	if drives == 0 {
		return nil, fmt.Errorf("GetLogicalDrives failed")
	}

	for i := 0; i < 26; i++ {
		if drives&(1<<uint(i)) == 0 {
			continue
		}

		driveLetter := string(rune('A'+i)) + ":\\"
		driveLetterW, _ := syscall.UTF16PtrFromString(driveLetter)

		// 检查驱动器类型
		driveType, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(driveLetterW)))
		if driveType != 3 { // DRIVE_FIXED = 3
			continue // 只处理固定磁盘
		}

		// 获取磁盘空间信息
		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		ret, _, _ := procGetDiskFreeSpaceExW.Call(
			uintptr(unsafe.Pointer(driveLetterW)),
			uintptr(unsafe.Pointer(&freeBytesAvailable)),
			uintptr(unsafe.Pointer(&totalBytes)),
			uintptr(unsafe.Pointer(&totalFreeBytes)),
		)

		if ret != 0 {
			disk := DiskInfo{
				Path:   driveLetter,
				Device: driveLetter,
				Total:  totalBytes,
				Free:   freeBytesAvailable,
				Used:   totalBytes - freeBytesAvailable,
			}

			if disk.Total > 0 {
				disk.UsedPercent = float64(disk.Used) / float64(disk.Total) * 100
			}

			// 获取文件系统类型
			disk.FSType = getDriveFileSystem(driveLetter)

			disks = append(disks, disk)
		}
	}

	return disks, nil
}

// getDriveFileSystem 获取驱动器文件系统类型
func getDriveFileSystem(drive string) string {
	output, err := runWMIC("logicaldisk", "where", fmt.Sprintf("DeviceID='%s'", drive[:2]), "get", "FileSystem", "/value")
	if err != nil {
		return ""
	}
	return parseWMICValue(output, "FileSystem")
}

// GetFDLimits 获取文件描述符限制（Windows实现）
// Windows 没有传统的文件描述符限制概念
func GetFDLimits() (*FDLimits, error) {
	// Windows 使用句柄而不是文件描述符
	// 默认限制通常很高（约 16 million）
	return &FDLimits{
		SoftLimit: 16777216, // Windows 默认句柄限制
		HardLimit: 16777216,
		Current:   0, // 难以准确获取
	}, nil
}

// SetFDLimits 设置文件描述符限制（Windows实现）
// Windows 不支持此操作
func SetFDLimits(soft, hard uint64) error {
	return fmt.Errorf("setting file descriptor limits is not supported on Windows")
}

// GetProcessFDs 获取当前进程的文件描述符列表（Windows实现）
// Windows 使用句柄而不是文件描述符
func GetProcessFDs() ([]*FDInfo, error) {
	// Windows 需要使用 NtQuerySystemInformation 或类似 API
	// 这里返回空列表
	return nil, fmt.Errorf("not implemented on Windows")
}

// GetResourceLimits 获取资源限制（Windows实现）
func GetResourceLimits() (*ResourceLimits, error) {
	limits := &ResourceLimits{
		MaxOpenFiles: 16777216, // Windows 默认句柄限制
	}

	// 获取内存信息作为参考
	memInfo, err := GetMemoryInfo()
	if err == nil {
		limits.MaxMemory = memInfo.Total
	}

	return limits, nil
}

// ===================
// 辅助函数
// ===================

// runWMIC 运行 WMIC 命令
func runWMIC(args ...string) (string, error) {
	cmd := exec.Command("wmic", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// parseWMICValue 解析 WMIC 输出中的值
func parseWMICValue(output, key string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+"=") {
			return strings.TrimPrefix(line, key+"=")
		}
	}
	return ""
}

// runCommand 运行命令并返回输出
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
