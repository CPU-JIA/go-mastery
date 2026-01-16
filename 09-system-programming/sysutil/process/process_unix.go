//go:build linux || darwin || freebsd || openbsd || netbsd
// +build linux darwin freebsd openbsd netbsd

/*
Unix/Linux 平台的进程管理实现

本文件实现了 Unix 系统上的进程信息获取功能：
  - 通过 /proc 文件系统读取进程信息（Linux）
  - 通过 sysctl 获取进程信息（BSD/macOS）
  - 支持进程状态、内存、CPU 等信息的获取
*/
package process

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

// listProcesses 获取所有进程列表（Unix实现）
func listProcesses() ([]*Info, error) {
	if runtime.GOOS == "linux" {
		return listProcessesLinux()
	}
	// macOS/BSD 使用 ps 命令作为后备方案
	return listProcessesPS()
}

// listProcessesLinux 通过 /proc 文件系统获取进程列表
func listProcessesLinux() ([]*Info, error) {
	// 读取 /proc 目录
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc: %w", err)
	}

	var procs []*Info

	for _, entry := range entries {
		// 只处理数字目录（进程ID）
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // 不是进程目录
		}

		info, err := getProcessInfoLinux(pid)
		if err != nil {
			continue // 进程可能已退出
		}

		procs = append(procs, info)
	}

	return procs, nil
}

// listProcessesPS 使用 ps 命令获取进程列表（macOS/BSD）
func listProcessesPS() ([]*Info, error) {
	output, err := Run("ps", []string{"ax", "-o", "pid,ppid,user,state,comm"}, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to run ps: %w", err)
	}

	var procs []*Info
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])

		info := &Info{
			PID:      pid,
			PPID:     ppid,
			Username: fields[2],
			State:    parseStatePS(fields[3]),
			Name:     fields[4],
		}

		// 获取更多详细信息
		if detailed, err := getProcessInfo(pid); err == nil {
			info.Executable = detailed.Executable
			info.CommandLine = detailed.CommandLine
			info.MemoryInfo = detailed.MemoryInfo
			info.CPUPercent = detailed.CPUPercent
		}

		procs = append(procs, info)
	}

	return procs, nil
}

// getProcessInfo 获取指定进程的详细信息（Unix实现）
func getProcessInfo(pid int) (*Info, error) {
	if runtime.GOOS == "linux" {
		return getProcessInfoLinux(pid)
	}
	return getProcessInfoBSD(pid)
}

// getProcessInfoLinux 通过 /proc 获取进程信息
func getProcessInfoLinux(pid int) (*Info, error) {
	procPath := fmt.Sprintf("/proc/%d", pid)

	// 检查进程是否存在
	if _, err := os.Stat(procPath); os.IsNotExist(err) {
		return nil, ErrProcessNotFound
	}

	info := &Info{
		PID: pid,
	}

	// 读取 /proc/[pid]/stat
	if err := readProcStat(pid, info); err != nil {
		return nil, err
	}

	// 读取 /proc/[pid]/status 获取更多信息
	if err := readProcStatus(pid, info); err != nil {
		// 非致命错误，继续
	}

	// 读取 /proc/[pid]/cmdline
	if cmdline, err := readProcFile(pid, "cmdline"); err == nil {
		// cmdline 使用 null 字符分隔参数
		info.CommandLine = strings.ReplaceAll(string(cmdline), "\x00", " ")
		info.CommandLine = strings.TrimSpace(info.CommandLine)
	}

	// 读取 /proc/[pid]/exe 获取可执行文件路径
	exePath := filepath.Join(procPath, "exe")
	if target, err := os.Readlink(exePath); err == nil {
		info.Executable = target
	}

	// 读取 /proc/[pid]/cwd 获取工作目录
	cwdPath := filepath.Join(procPath, "cwd")
	if target, err := os.Readlink(cwdPath); err == nil {
		info.WorkingDir = target
	}

	// 读取 /proc/[pid]/statm 获取内存信息
	if err := readProcStatm(pid, info); err != nil {
		// 非致命错误
	}

	// 读取 /proc/[pid]/io 获取I/O信息
	if err := readProcIO(pid, info); err != nil {
		// 非致命错误，可能需要 root 权限
	}

	// 计算文件描述符数量
	fdPath := filepath.Join(procPath, "fd")
	if entries, err := os.ReadDir(fdPath); err == nil {
		info.NumFDs = len(entries)
	}

	return info, nil
}

// readProcStat 读取 /proc/[pid]/stat 文件
func readProcStat(pid int, info *Info) error {
	data, err := readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	// stat 文件格式复杂，进程名可能包含空格和括号
	// 格式: pid (comm) state ppid pgrp session tty_nr tpgid flags ...
	content := string(data)

	// 找到进程名的开始和结束位置
	start := strings.Index(content, "(")
	end := strings.LastIndex(content, ")")
	if start == -1 || end == -1 {
		return fmt.Errorf("invalid stat format")
	}

	info.Name = content[start+1 : end]

	// 解析括号后的字段
	fields := strings.Fields(content[end+2:])
	if len(fields) < 20 {
		return fmt.Errorf("insufficient stat fields")
	}

	// 字段索引（从括号后开始）
	// 0: state, 1: ppid, 2: pgrp, 3: session, ...
	info.State = parseStateProcStat(fields[0])
	info.PPID, _ = strconv.Atoi(fields[1])
	info.Nice, _ = strconv.Atoi(fields[16])
	info.NumThreads, _ = strconv.Atoi(fields[17])

	// 计算进程启动时间
	// starttime 是从系统启动开始的时钟滴答数
	starttime, _ := strconv.ParseUint(fields[19], 10, 64)
	info.CreateTime = calculateStartTime(starttime)

	// 计算 CPU 使用率
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)
	info.CPUPercent = calculateCPUPercent(utime, stime, info.CreateTime)

	return nil
}

// readProcStatus 读取 /proc/[pid]/status 文件
func readProcStatus(pid int, info *Info) error {
	data, err := readProcFile(pid, "status")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Uid":
			// 获取有效用户ID
			fields := strings.Fields(value)
			if len(fields) >= 1 {
				uid, _ := strconv.Atoi(fields[0])
				info.Username = lookupUsername(uid)
			}
		case "VmRSS":
			// 常驻内存大小（KB）
			info.MemoryInfo.RSS = parseMemoryValue(value)
		case "VmSize":
			// 虚拟内存大小（KB）
			info.MemoryInfo.VMS = parseMemoryValue(value)
		case "VmData":
			// 数据段大小
			info.MemoryInfo.Data = parseMemoryValue(value)
		case "VmStk":
			// 栈大小
			info.MemoryInfo.Stack = parseMemoryValue(value)
		case "RssAnon", "RssShmem":
			// 共享内存
			info.MemoryInfo.Shared += parseMemoryValue(value)
		}
	}

	// 计算内存使用百分比
	if totalMem := getTotalMemory(); totalMem > 0 {
		info.MemoryInfo.Percent = float64(info.MemoryInfo.RSS) / float64(totalMem) * 100
	}

	return nil
}

// readProcStatm 读取 /proc/[pid]/statm 文件
func readProcStatm(pid int, info *Info) error {
	data, err := readProcFile(pid, "statm")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 7 {
		return fmt.Errorf("insufficient statm fields")
	}

	pageSize := uint64(syscall.Getpagesize())

	// statm 字段: size resident shared text lib data dt
	size, _ := strconv.ParseUint(fields[0], 10, 64)
	resident, _ := strconv.ParseUint(fields[1], 10, 64)
	shared, _ := strconv.ParseUint(fields[2], 10, 64)

	info.MemoryInfo.VMS = size * pageSize
	info.MemoryInfo.RSS = resident * pageSize
	info.MemoryInfo.Shared = shared * pageSize

	return nil
}

// readProcIO 读取 /proc/[pid]/io 文件
func readProcIO(pid int, info *Info) error {
	data, err := readProcFile(pid, "io")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "syscr":
			info.IOCounters.ReadCount, _ = strconv.ParseUint(value, 10, 64)
		case "syscw":
			info.IOCounters.WriteCount, _ = strconv.ParseUint(value, 10, 64)
		case "read_bytes":
			info.IOCounters.ReadBytes, _ = strconv.ParseUint(value, 10, 64)
		case "write_bytes":
			info.IOCounters.WriteBytes, _ = strconv.ParseUint(value, 10, 64)
		}
	}

	return nil
}

// readProcFile 读取 /proc/[pid]/[name] 文件
func readProcFile(pid int, name string) ([]byte, error) {
	path := fmt.Sprintf("/proc/%d/%s", pid, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProcessNotFound
		}
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	return data, nil
}

// getProcessInfoBSD 获取 BSD/macOS 系统的进程信息
func getProcessInfoBSD(pid int) (*Info, error) {
	// 使用 ps 命令获取基本信息
	output, err := Run("ps", []string{"-p", strconv.Itoa(pid), "-o", "pid,ppid,user,state,comm,rss,vsz"}, 5*time.Second)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return nil, ErrProcessNotFound
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid ps output")
	}

	ppid, _ := strconv.Atoi(fields[1])
	rss, _ := strconv.ParseUint(fields[5], 10, 64)
	vms, _ := strconv.ParseUint(fields[6], 10, 64)

	info := &Info{
		PID:      pid,
		PPID:     ppid,
		Username: fields[2],
		State:    parseStatePS(fields[3]),
		Name:     fields[4],
		MemoryInfo: MemoryInfo{
			RSS: rss * 1024, // ps 输出是 KB
			VMS: vms * 1024,
		},
	}

	return info, nil
}

// parseStateProcStat 解析 /proc/[pid]/stat 中的状态字符
func parseStateProcStat(state string) ProcessState {
	if len(state) == 0 {
		return StateUnknown
	}

	switch state[0] {
	case 'R':
		return StateRunning
	case 'S':
		return StateSleeping
	case 'D':
		return StateWaiting
	case 'Z':
		return StateZombie
	case 'T', 't':
		return StateStopped
	case 'I':
		return StateIdle
	default:
		return StateUnknown
	}
}

// parseStatePS 解析 ps 命令输出的状态
func parseStatePS(state string) ProcessState {
	if len(state) == 0 {
		return StateUnknown
	}

	switch state[0] {
	case 'R':
		return StateRunning
	case 'S':
		return StateSleeping
	case 'D', 'U':
		return StateWaiting
	case 'Z':
		return StateZombie
	case 'T':
		return StateStopped
	case 'I':
		return StateIdle
	default:
		return StateUnknown
	}
}

// parseMemoryValue 解析内存值（如 "1234 kB"）
func parseMemoryValue(value string) uint64 {
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

// calculateStartTime 计算进程启动时间
func calculateStartTime(starttime uint64) time.Time {
	// 获取系统启动时间
	bootTime := getBootTime()
	if bootTime.IsZero() {
		return time.Time{}
	}

	// starttime 是从系统启动开始的时钟滴答数
	// 需要转换为秒
	ticksPerSecond := uint64(100) // 通常是 100 Hz
	seconds := starttime / ticksPerSecond

	return bootTime.Add(time.Duration(seconds) * time.Second)
}

// getBootTime 获取系统启动时间
func getBootTime() time.Time {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				btime, _ := strconv.ParseInt(fields[1], 10, 64)
				return time.Unix(btime, 0)
			}
		}
	}

	return time.Time{}
}

// calculateCPUPercent 计算 CPU 使用率
func calculateCPUPercent(utime, stime uint64, startTime time.Time) float64 {
	if startTime.IsZero() {
		return 0
	}

	// 总 CPU 时间（时钟滴答）
	totalTime := utime + stime

	// 转换为秒
	ticksPerSecond := uint64(100)
	cpuSeconds := float64(totalTime) / float64(ticksPerSecond)

	// 进程运行时间
	elapsed := time.Since(startTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	// CPU 使用率
	return (cpuSeconds / elapsed) * 100
}

// getTotalMemory 获取系统总内存
func getTotalMemory() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			return parseMemoryValue(strings.TrimPrefix(line, "MemTotal:"))
		}
	}

	return 0
}

// lookupUsername 根据 UID 查找用户名
func lookupUsername(uid int) string {
	// 读取 /etc/passwd
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 3 {
			if u, _ := strconv.Atoi(fields[2]); u == uid {
				return fields[0]
			}
		}
	}

	return strconv.Itoa(uid)
}
