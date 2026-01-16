//go:build linux || darwin || freebsd || openbsd || netbsd
// +build linux darwin freebsd openbsd netbsd

/*
Unix/Linux 平台的资源管理实现

本文件实现了 Unix 系统上的资源信息获取功能：
  - 通过 /proc 文件系统读取系统信息（Linux）
  - 通过 sysctl 获取系统信息（BSD/macOS）
  - 支持内存、CPU、磁盘等资源的监控
*/
package resource

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// GetMemoryInfo 获取系统内存信息（Unix实现）
func GetMemoryInfo() (*MemoryInfo, error) {
	if runtime.GOOS == "linux" {
		return getMemoryInfoLinux()
	}
	return getMemoryInfoBSD()
}

// getMemoryInfoLinux 通过 /proc/meminfo 获取内存信息
func getMemoryInfoLinux() (*MemoryInfo, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	info := &MemoryInfo{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := parseMemValue(strings.TrimSpace(parts[1]))

		switch key {
		case "MemTotal":
			info.Total = value
		case "MemFree":
			info.Free = value
		case "MemAvailable":
			info.Available = value
		case "Buffers":
			info.Buffers = value
		case "Cached":
			info.Cached = value
		case "SwapTotal":
			info.SwapTotal = value
		case "SwapFree":
			info.SwapFree = value
		}
	}

	// 计算已用内存
	info.Used = info.Total - info.Available
	info.SwapUsed = info.SwapTotal - info.SwapFree

	// 计算使用百分比
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// getMemoryInfoBSD 获取 BSD/macOS 系统的内存信息
func getMemoryInfoBSD() (*MemoryInfo, error) {
	info := &MemoryInfo{}

	// 使用 sysctl 获取内存信息
	// macOS: hw.memsize
	// FreeBSD: hw.physmem

	var pageSize uint64 = 4096 // 默认页面大小

	// 获取页面大小
	if ps := syscall.Getpagesize(); ps > 0 {
		pageSize = uint64(ps)
	}

	// 使用 vm_stat 命令获取内存统计（macOS）
	if runtime.GOOS == "darwin" {
		output, err := runCommand("vm_stat")
		if err == nil {
			info = parseVMStat(output, pageSize)
		}

		// 获取总内存
		totalOutput, err := runCommand("sysctl", "-n", "hw.memsize")
		if err == nil {
			info.Total, _ = strconv.ParseUint(strings.TrimSpace(totalOutput), 10, 64)
		}
	}

	// 计算使用百分比
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// parseVMStat 解析 vm_stat 输出
func parseVMStat(output string, pageSize uint64) *MemoryInfo {
	info := &MemoryInfo{}

	var freePages, activePages, inactivePages, wiredPages, compressedPages uint64

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "."))
		value, _ := strconv.ParseUint(valueStr, 10, 64)

		switch key {
		case "Pages free":
			freePages = value
		case "Pages active":
			activePages = value
		case "Pages inactive":
			inactivePages = value
		case "Pages wired down":
			wiredPages = value
		case "Pages occupied by compressor":
			compressedPages = value
		}
	}

	info.Free = freePages * pageSize
	info.Available = (freePages + inactivePages) * pageSize
	info.Used = (activePages + wiredPages + compressedPages) * pageSize
	info.Cached = inactivePages * pageSize

	return info
}

// GetCPUInfo 获取CPU信息（Unix实现）
func GetCPUInfo() (*CPUInfo, error) {
	info := &CPUInfo{
		NumCPU:        runtime.NumCPU(),
		NumLogicalCPU: runtime.NumCPU(),
	}

	if runtime.GOOS == "linux" {
		// 读取 /proc/cpuinfo
		data, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			parseCPUInfo(string(data), info)
		}
	} else if runtime.GOOS == "darwin" {
		// macOS 使用 sysctl
		if output, err := runCommand("sysctl", "-n", "machdep.cpu.brand_string"); err == nil {
			info.ModelName = strings.TrimSpace(output)
		}
		if output, err := runCommand("sysctl", "-n", "machdep.cpu.vendor"); err == nil {
			info.Vendor = strings.TrimSpace(output)
		}
	}

	return info, nil
}

// parseCPUInfo 解析 /proc/cpuinfo
func parseCPUInfo(data string, info *CPUInfo) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if info.ModelName == "" {
				info.ModelName = value
			}
		case "vendor_id":
			if info.Vendor == "" {
				info.Vendor = value
			}
		case "cpu family":
			if info.Family == "" {
				info.Family = value
			}
		case "model":
			if info.Model == "" {
				info.Model = value
			}
		case "cpu MHz":
			if info.MHz == 0 {
				info.MHz, _ = strconv.ParseFloat(value, 64)
			}
		case "cache size":
			if info.CacheSize == 0 {
				// 格式: "6144 KB"
				fields := strings.Fields(value)
				if len(fields) > 0 {
					info.CacheSize, _ = strconv.Atoi(fields[0])
				}
			}
		}
	}
}

// CPU 时间记录（用于计算使用率）
var lastCPUTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
	timestamp                                              time.Time
}

// GetCPUUsage 获取CPU使用率（Unix实现）
func GetCPUUsage() (*CPUUsage, error) {
	if runtime.GOOS == "linux" {
		return getCPUUsageLinux()
	}
	return getCPUUsageBSD()
}

// getCPUUsageLinux 通过 /proc/stat 获取CPU使用率
func getCPUUsageLinux() (*CPUUsage, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/stat: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			user, _ := strconv.ParseUint(fields[1], 10, 64)
			nice, _ := strconv.ParseUint(fields[2], 10, 64)
			system, _ := strconv.ParseUint(fields[3], 10, 64)
			idle, _ := strconv.ParseUint(fields[4], 10, 64)
			iowait, _ := strconv.ParseUint(fields[5], 10, 64)
			irq, _ := strconv.ParseUint(fields[6], 10, 64)
			softirq, _ := strconv.ParseUint(fields[7], 10, 64)

			var steal uint64
			if len(fields) > 8 {
				steal, _ = strconv.ParseUint(fields[8], 10, 64)
			}

			// 计算差值
			now := time.Now()
			if !lastCPUTimes.timestamp.IsZero() {
				userDiff := user - lastCPUTimes.user
				niceDiff := nice - lastCPUTimes.nice
				systemDiff := system - lastCPUTimes.system
				idleDiff := idle - lastCPUTimes.idle
				iowaitDiff := iowait - lastCPUTimes.iowait
				irqDiff := irq - lastCPUTimes.irq
				softirqDiff := softirq - lastCPUTimes.softirq
				stealDiff := steal - lastCPUTimes.steal

				total := userDiff + niceDiff + systemDiff + idleDiff + iowaitDiff + irqDiff + softirqDiff + stealDiff

				if total > 0 {
					usage := &CPUUsage{
						User:    float64(userDiff+niceDiff) / float64(total) * 100,
						System:  float64(systemDiff) / float64(total) * 100,
						Idle:    float64(idleDiff) / float64(total) * 100,
						IOWait:  float64(iowaitDiff) / float64(total) * 100,
						IRQ:     float64(irqDiff) / float64(total) * 100,
						SoftIRQ: float64(softirqDiff) / float64(total) * 100,
						Steal:   float64(stealDiff) / float64(total) * 100,
					}
					usage.Total = 100 - usage.Idle

					// 更新记录
					lastCPUTimes.user = user
					lastCPUTimes.nice = nice
					lastCPUTimes.system = system
					lastCPUTimes.idle = idle
					lastCPUTimes.iowait = iowait
					lastCPUTimes.irq = irq
					lastCPUTimes.softirq = softirq
					lastCPUTimes.steal = steal
					lastCPUTimes.timestamp = now

					return usage, nil
				}
			}

			// 首次调用，保存当前值
			lastCPUTimes.user = user
			lastCPUTimes.nice = nice
			lastCPUTimes.system = system
			lastCPUTimes.idle = idle
			lastCPUTimes.iowait = iowait
			lastCPUTimes.irq = irq
			lastCPUTimes.softirq = softirq
			lastCPUTimes.steal = steal
			lastCPUTimes.timestamp = now

			// 返回估计值
			total := user + nice + system + idle + iowait + irq + softirq + steal
			if total > 0 {
				return &CPUUsage{
					User:   float64(user+nice) / float64(total) * 100,
					System: float64(system) / float64(total) * 100,
					Idle:   float64(idle) / float64(total) * 100,
					Total:  float64(total-idle) / float64(total) * 100,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse CPU stats")
}

// getCPUUsageBSD 获取 BSD/macOS 系统的CPU使用率
func getCPUUsageBSD() (*CPUUsage, error) {
	// 使用 top 命令获取 CPU 使用率
	output, err := runCommand("top", "-l", "1", "-n", "0")
	if err != nil {
		return nil, err
	}

	usage := &CPUUsage{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "CPU usage:") {
			// 格式: CPU usage: 5.26% user, 10.52% sys, 84.21% idle
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "user") {
					usage.User = parsePercent(part)
				} else if strings.Contains(part, "sys") {
					usage.System = parsePercent(part)
				} else if strings.Contains(part, "idle") {
					usage.Idle = parsePercent(part)
				}
			}
			usage.Total = 100 - usage.Idle
			break
		}
	}

	return usage, nil
}

// GetDiskInfo 获取磁盘信息（Unix实现）
func GetDiskInfo() ([]DiskInfo, error) {
	var disks []DiskInfo

	if runtime.GOOS == "linux" {
		// 读取 /proc/mounts
		data, err := os.ReadFile("/proc/mounts")
		if err != nil {
			return nil, fmt.Errorf("failed to read /proc/mounts: %w", err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}

			device := fields[0]
			mountPoint := fields[1]
			fsType := fields[2]

			// 跳过虚拟文件系统
			if strings.HasPrefix(fsType, "proc") ||
				strings.HasPrefix(fsType, "sys") ||
				strings.HasPrefix(fsType, "devpts") ||
				strings.HasPrefix(fsType, "tmpfs") ||
				strings.HasPrefix(fsType, "cgroup") {
				continue
			}

			// 获取磁盘使用情况
			var stat syscall.Statfs_t
			if err := syscall.Statfs(mountPoint, &stat); err != nil {
				continue
			}

			blockSize := uint64(stat.Bsize)
			disk := DiskInfo{
				Path:        mountPoint,
				Device:      device,
				FSType:      fsType,
				Total:       stat.Blocks * blockSize,
				Free:        stat.Bavail * blockSize,
				InodesTotal: stat.Files,
				InodesFree:  stat.Ffree,
			}
			disk.Used = disk.Total - disk.Free
			disk.InodesUsed = disk.InodesTotal - disk.InodesFree

			if disk.Total > 0 {
				disk.UsedPercent = float64(disk.Used) / float64(disk.Total) * 100
			}

			disks = append(disks, disk)
		}
	} else {
		// macOS/BSD 使用 df 命令
		output, err := runCommand("df", "-k")
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(strings.NewReader(output))
		first := true
		for scanner.Scan() {
			if first {
				first = false
				continue // 跳过标题行
			}

			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 6 {
				continue
			}

			total, _ := strconv.ParseUint(fields[1], 10, 64)
			used, _ := strconv.ParseUint(fields[2], 10, 64)
			free, _ := strconv.ParseUint(fields[3], 10, 64)

			disk := DiskInfo{
				Device: fields[0],
				Total:  total * 1024,
				Used:   used * 1024,
				Free:   free * 1024,
				Path:   fields[len(fields)-1],
			}

			if disk.Total > 0 {
				disk.UsedPercent = float64(disk.Used) / float64(disk.Total) * 100
			}

			disks = append(disks, disk)
		}
	}

	return disks, nil
}

// GetFDLimits 获取文件描述符限制（Unix实现）
func GetFDLimits() (*FDLimits, error) {
	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return nil, fmt.Errorf("failed to get rlimit: %w", err)
	}

	limits := &FDLimits{
		SoftLimit: rlimit.Cur,
		HardLimit: rlimit.Max,
	}

	// 获取当前使用的文件描述符数
	if runtime.GOOS == "linux" {
		fdPath := fmt.Sprintf("/proc/%d/fd", os.Getpid())
		entries, err := os.ReadDir(fdPath)
		if err == nil {
			limits.Current = uint64(len(entries))
		}
	}

	return limits, nil
}

// SetFDLimits 设置文件描述符限制（Unix实现）
func SetFDLimits(soft, hard uint64) error {
	rlimit := syscall.Rlimit{
		Cur: soft,
		Max: hard,
	}

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return fmt.Errorf("failed to set rlimit: %w", err)
	}

	return nil
}

// GetProcessFDs 获取当前进程的文件描述符列表（Unix实现）
func GetProcessFDs() ([]*FDInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("not supported on %s", runtime.GOOS)
	}

	fdPath := fmt.Sprintf("/proc/%d/fd", os.Getpid())
	entries, err := os.ReadDir(fdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fd directory: %w", err)
	}

	var fds []*FDInfo

	for _, entry := range entries {
		fd, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		info := &FDInfo{FD: fd}

		// 读取符号链接获取文件路径
		linkPath := filepath.Join(fdPath, entry.Name())
		if target, err := os.Readlink(linkPath); err == nil {
			info.Path = target

			// 判断类型
			switch {
			case strings.HasPrefix(target, "socket:"):
				info.Type = "socket"
			case strings.HasPrefix(target, "pipe:"):
				info.Type = "pipe"
			case strings.HasPrefix(target, "anon_inode:"):
				info.Type = "anon_inode"
			case strings.HasPrefix(target, "/dev/"):
				info.Type = "device"
			default:
				info.Type = "file"
			}
		}

		fds = append(fds, info)
	}

	return fds, nil
}

// GetResourceLimits 获取资源限制（Unix实现）
func GetResourceLimits() (*ResourceLimits, error) {
	limits := &ResourceLimits{}

	// 文件描述符限制
	var nofile syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &nofile); err == nil {
		limits.MaxOpenFiles = nofile.Cur
	}

	// 进程数限制（仅Linux）
	if runtime.GOOS == "linux" {
		var nproc syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_NPROC, &nproc); err == nil {
			limits.MaxProcesses = nproc.Cur
		}
	}

	// 内存限制
	var as syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_AS, &as); err == nil {
		limits.MaxMemory = as.Cur
	}

	// CPU时间限制
	var cpu syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_CPU, &cpu); err == nil {
		limits.MaxCPU = cpu.Cur
	}

	// 文件大小限制
	var fsize syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &fsize); err == nil {
		limits.MaxFileSize = fsize.Cur
	}

	// 栈大小限制
	var stack syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_STACK, &stack); err == nil {
		limits.MaxStack = stack.Cur
	}

	return limits, nil
}

// ===================
// 辅助函数
// ===================

// parseMemValue 解析内存值（如 "1234 kB"）
func parseMemValue(value string) uint64 {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0
	}

	val, _ := strconv.ParseUint(fields[0], 10, 64)

	// 检查单位
	if len(fields) > 1 {
		unit := strings.ToLower(fields[1])
		switch unit {
		case "kb":
			val *= 1024
		case "mb":
			val *= 1024 * 1024
		case "gb":
			val *= 1024 * 1024 * 1024
		}
	}

	return val
}

// parsePercent 解析百分比值
func parsePercent(s string) float64 {
	// 查找数字
	var numStr string
	for _, c := range s {
		if (c >= '0' && c <= '9') || c == '.' {
			numStr += string(c)
		} else if len(numStr) > 0 {
			break
		}
	}

	val, _ := strconv.ParseFloat(numStr, 64)
	return val
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

import "os/exec"
