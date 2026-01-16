/*
系统编程工具演示程序

本程序演示 sysutil 包的各种功能：
  - 进程管理和监控
  - 网络诊断工具
  - 系统资源监控

运行方式：

	go run example.go
*/
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-mastery/09-system-programming/sysutil/network"
	"go-mastery/09-system-programming/sysutil/process"
	"go-mastery/09-system-programming/sysutil/resource"
)

func main() {
	fmt.Println("=== 系统编程工具演示 ===")
	fmt.Println()

	// 1. 进程管理演示
	demonstrateProcessManagement()

	// 2. 网络工具演示
	demonstrateNetworkTools()

	// 3. 资源监控演示
	demonstrateResourceMonitoring()

	fmt.Println("\n=== 演示完成 ===")
}

// demonstrateProcessManagement 演示进程管理功能
func demonstrateProcessManagement() {
	fmt.Println("【1. 进程管理】")
	fmt.Println(strings("-", 50))

	// 获取当前进程信息
	self, err := process.Self()
	if err != nil {
		fmt.Printf("获取当前进程信息失败: %v\n", err)
	} else {
		fmt.Printf("当前进程:\n")
		fmt.Printf("  PID: %d\n", self.PID)
		fmt.Printf("  名称: %s\n", self.Name)
		fmt.Printf("  状态: %s\n", self.State)
		fmt.Printf("  内存(RSS): %s\n", resource.FormatBytes(self.MemoryInfo.RSS))
		fmt.Printf("  CPU使用率: %.2f%%\n", self.CPUPercent)
	}

	// 获取父进程信息
	parent, err := process.Parent()
	if err != nil {
		fmt.Printf("获取父进程信息失败: %v\n", err)
	} else {
		fmt.Printf("\n父进程:\n")
		fmt.Printf("  PID: %d\n", parent.PID)
		fmt.Printf("  名称: %s\n", parent.Name)
	}

	// 获取进程总数
	count, err := process.Count()
	if err != nil {
		fmt.Printf("获取进程总数失败: %v\n", err)
	} else {
		fmt.Printf("\n系统进程总数: %d\n", count)
	}

	// 获取 CPU 使用率最高的进程
	fmt.Println("\nCPU 使用率 Top 5:")
	topCPU, err := process.TopByCPU(5)
	if err != nil {
		fmt.Printf("获取 Top CPU 进程失败: %v\n", err)
	} else {
		for i, p := range topCPU {
			fmt.Printf("  %d. [%d] %s - %.2f%%\n", i+1, p.PID, p.Name, p.CPUPercent)
		}
	}

	// 获取内存使用最高的进程
	fmt.Println("\n内存使用 Top 5:")
	topMem, err := process.TopByMemory(5)
	if err != nil {
		fmt.Printf("获取 Top Memory 进程失败: %v\n", err)
	} else {
		for i, p := range topMem {
			fmt.Printf("  %d. [%d] %s - %s\n", i+1, p.PID, p.Name,
				resource.FormatBytes(p.MemoryInfo.RSS))
		}
	}

	fmt.Println()
}

// demonstrateNetworkTools 演示网络工具功能
func demonstrateNetworkTools() {
	fmt.Println("【2. 网络工具】")
	fmt.Println(strings("-", 50))

	// 获取网络接口信息
	fmt.Println("网络接口:")
	ifaces, err := network.GetInterfaces()
	if err != nil {
		fmt.Printf("获取网络接口失败: %v\n", err)
	} else {
		for _, iface := range ifaces {
			if !iface.IsUp {
				continue // 只显示启用的接口
			}
			status := "UP"
			if iface.IsLoopback {
				status += " (回环)"
			}
			fmt.Printf("  %s [%s]\n", iface.Name, status)
			for _, ip := range iface.IPv4Addresses {
				fmt.Printf("    IPv4: %s\n", ip)
			}
		}
	}

	// 获取本机 IP
	localIP, err := network.GetLocalIP()
	if err != nil {
		fmt.Printf("\n获取本机 IP 失败: %v\n", err)
	} else {
		fmt.Printf("\n本机 IP: %s\n", localIP)
	}

	// 获取空闲端口
	freePort, err := network.GetFreePort()
	if err != nil {
		fmt.Printf("获取空闲端口失败: %v\n", err)
	} else {
		fmt.Printf("可用端口: %d\n", freePort)
	}

	// TCP Ping 测试
	fmt.Println("\nTCP Ping 测试 (localhost:80):")
	result := network.TCPPing("127.0.0.1", 80, 1*time.Second)
	if result.Success {
		fmt.Printf("  成功 - RTT: %v\n", result.RTT)
	} else {
		fmt.Printf("  失败 - %v\n", result.Error)
	}

	// DNS 解析测试
	fmt.Println("\nDNS 解析 (localhost):")
	dns := network.ResolveDNS("localhost")
	if dns.Error != nil {
		fmt.Printf("  失败: %v\n", dns.Error)
	} else {
		fmt.Printf("  解析结果: %v\n", dns.IPs)
		fmt.Printf("  解析耗时: %v\n", dns.Latency)
	}

	// 端口扫描演示
	fmt.Println("\n常用端口检测 (localhost):")
	commonPorts := []int{22, 80, 443, 3306, 5432, 6379, 8080}
	scanner := network.NewPortScanner()
	scanner.Timeout = 500 * time.Millisecond
	for _, port := range commonPorts {
		open := network.IsPortOpen("127.0.0.1", port, 200*time.Millisecond)
		status := "关闭"
		if open {
			status = "开放"
		}
		service := getServiceName(port)
		fmt.Printf("  端口 %d (%s): %s\n", port, service, status)
	}

	fmt.Println()
}

// demonstrateResourceMonitoring 演示资源监控功能
func demonstrateResourceMonitoring() {
	fmt.Println("【3. 资源监控】")
	fmt.Println(strings("-", 50))

	// 内存信息
	fmt.Println("内存信息:")
	mem, err := resource.GetMemoryInfo()
	if err != nil {
		fmt.Printf("  获取内存信息失败: %v\n", err)
	} else {
		fmt.Printf("  总内存: %s\n", resource.FormatBytes(mem.Total))
		fmt.Printf("  可用内存: %s\n", resource.FormatBytes(mem.Available))
		fmt.Printf("  已用内存: %s (%.1f%%)\n",
			resource.FormatBytes(mem.Used), mem.UsedPercent)
		if mem.SwapTotal > 0 {
			fmt.Printf("  交换空间: %s / %s\n",
				resource.FormatBytes(mem.SwapUsed),
				resource.FormatBytes(mem.SwapTotal))
		}
	}

	// CPU 信息
	fmt.Println("\nCPU 信息:")
	cpu, err := resource.GetCPUInfo()
	if err != nil {
		fmt.Printf("  获取 CPU 信息失败: %v\n", err)
	} else {
		fmt.Printf("  型号: %s\n", cpu.ModelName)
		fmt.Printf("  核心数: %d\n", cpu.NumCPU)
		if cpu.MHz > 0 {
			fmt.Printf("  频率: %.0f MHz\n", cpu.MHz)
		}
	}

	// CPU 使用率
	fmt.Println("\nCPU 使用率:")
	// 第一次调用初始化
	resource.GetCPUUsage()
	time.Sleep(200 * time.Millisecond)
	usage, err := resource.GetCPUUsage()
	if err != nil {
		fmt.Printf("  获取 CPU 使用率失败: %v\n", err)
	} else {
		fmt.Printf("  总使用率: %.1f%%\n", usage.Total)
		fmt.Printf("  用户态: %.1f%%\n", usage.User)
		fmt.Printf("  内核态: %.1f%%\n", usage.System)
		fmt.Printf("  空闲: %.1f%%\n", usage.Idle)
	}

	// 磁盘信息
	fmt.Println("\n磁盘信息:")
	disks, err := resource.GetDiskInfo()
	if err != nil {
		fmt.Printf("  获取磁盘信息失败: %v\n", err)
	} else {
		for _, disk := range disks {
			if disk.Total == 0 {
				continue
			}
			fmt.Printf("  %s (%s)\n", disk.Path, disk.FSType)
			fmt.Printf("    总空间: %s\n", resource.FormatBytes(disk.Total))
			fmt.Printf("    可用: %s (%.1f%% 已用)\n",
				resource.FormatBytes(disk.Free), disk.UsedPercent)
		}
	}

	// 文件描述符限制
	fmt.Println("\n文件描述符限制:")
	fdLimits, err := resource.GetFDLimits()
	if err != nil {
		fmt.Printf("  获取 FD 限制失败: %v\n", err)
	} else {
		fmt.Printf("  软限制: %d\n", fdLimits.SoftLimit)
		fmt.Printf("  硬限制: %d\n", fdLimits.HardLimit)
		if fdLimits.Current > 0 {
			fmt.Printf("  当前使用: %d\n", fdLimits.Current)
		}
	}

	// Go 运行时信息
	fmt.Println("\nGo 运行时信息:")
	rt := resource.GetRuntimeInfo()
	fmt.Printf("  Go 版本: %s\n", rt.Version)
	fmt.Printf("  CPU 数量: %d\n", rt.NumCPU)
	fmt.Printf("  Goroutine 数量: %d\n", rt.NumGoroutine)
	fmt.Printf("  GOMAXPROCS: %d\n", rt.GOMAXPROCS)

	// 进程内存信息
	fmt.Println("\n当前进程内存:")
	procMem := resource.GetProcessMemoryInfo()
	fmt.Printf("  堆分配: %s\n", resource.FormatBytes(procMem.HeapAlloc))
	fmt.Printf("  堆系统: %s\n", resource.FormatBytes(procMem.HeapSys))
	fmt.Printf("  栈使用: %s\n", resource.FormatBytes(procMem.StackInuse))
	fmt.Printf("  GC 次数: %d\n", procMem.NumGC)

	// 资源监控器演示
	fmt.Println("\n资源监控器 (监控 3 秒):")
	monitor := resource.NewResourceMonitor(500 * time.Millisecond)

	sampleCount := 0
	monitor.OnMemory(func(info resource.MemoryInfo) {
		sampleCount++
		fmt.Printf("  [%d] 内存: %.1f%% | ", sampleCount, info.UsedPercent)
	})

	monitor.OnCPU(func(usage resource.CPUUsage) {
		fmt.Printf("CPU: %.1f%%\n", usage.Total)
	})

	monitor.Start()
	time.Sleep(3 * time.Second)
	monitor.Stop()

	fmt.Println()
}

// strings 生成重复字符串
func strings(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

// getServiceName 获取服务名称
func getServiceName(port int) string {
	services := map[int]string{
		22:   "SSH",
		80:   "HTTP",
		443:  "HTTPS",
		3306: "MySQL",
		5432: "PostgreSQL",
		6379: "Redis",
		8080: "HTTP-Alt",
	}
	if name, ok := services[port]; ok {
		return name
	}
	return "Unknown"
}

// 连接池使用示例（注释掉，需要实际服务器）
func demonstrateConnectionPool() {
	fmt.Println("【连接池示例】")

	// 创建连接池
	pool, err := network.NewConnPool(network.PoolConfig{
		Address:     "localhost:6379", // Redis 服务器
		MaxSize:     10,
		MinSize:     2,
		MaxIdleTime: 5 * time.Minute,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("创建连接池失败: %v\n", err)
		return
	}
	defer pool.Close()

	// 获取连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pool.Get(ctx)
	if err != nil {
		fmt.Printf("获取连接失败: %v\n", err)
		return
	}

	// 使用连接...
	fmt.Printf("获取连接成功，池大小: %d\n", pool.Size())

	// 归还连接
	conn.Close()

	// 查看统计信息
	stats := pool.Stats()
	fmt.Printf("连接池统计:\n")
	fmt.Printf("  总连接数: %d\n", stats.TotalConnections)
	fmt.Printf("  活跃连接: %d\n", stats.ActiveConnections)
	fmt.Printf("  命中次数: %d\n", stats.HitCount)
	fmt.Printf("  未命中次数: %d\n", stats.MissCount)
}

// 进程监控示例
func demonstrateProcessMonitor() {
	fmt.Println("【进程监控示例】")

	// 监控当前进程
	pid := os.Getpid()
	monitor := process.NewMonitor(pid, 500*time.Millisecond)

	events := monitor.Start()

	// 监控 5 秒
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case event, ok := <-events:
			if !ok {
				break loop
			}
			switch event.Type {
			case process.EventStarted:
				fmt.Printf("进程启动: PID=%d\n", event.PID)
			case process.EventExited:
				fmt.Printf("进程退出: PID=%d\n", event.PID)
			case process.EventCPUHigh:
				fmt.Printf("CPU 使用率过高: %.1f%%\n", event.Info.CPUPercent)
			case process.EventMemoryHigh:
				fmt.Printf("内存使用率过高: %.1f%%\n", event.Info.MemoryInfo.Percent)
			}
		case <-timeout:
			break loop
		}
	}

	monitor.Stop()
}

// 文件描述符追踪示例
func demonstrateFDTracker() {
	fmt.Println("【文件描述符追踪示例】")

	tracker := resource.NewFDTracker()

	// 设置警告阈值
	tracker.SetWarningThreshold(80.0)

	// 设置警告回调
	tracker.SetWarningCallback(func(current, limit uint64) {
		fmt.Printf("警告: 文件描述符使用率过高 (%d/%d)\n", current, limit)
	})

	// 模拟追踪文件描述符
	tracker.Track(10, "/var/log/app.log", "file")
	tracker.Track(11, "socket:[12345]", "socket")
	tracker.Track(12, "pipe:[67890]", "pipe")

	// 获取统计信息
	stats := tracker.GetStats()
	fmt.Printf("FD 统计:\n")
	fmt.Printf("  当前打开: %d\n", stats.CurrentOpen)
	fmt.Printf("  总打开数: %d\n", stats.TotalOpened)
	fmt.Printf("  最大同时打开: %d\n", stats.MaxOpen)

	// 查找泄漏
	leaks := tracker.FindLeaks(1 * time.Hour)
	if len(leaks) > 0 {
		fmt.Printf("发现 %d 个可疑泄漏:\n", len(leaks))
		for _, fd := range leaks {
			fmt.Printf("  FD %d: %s (%s)\n", fd.FD, fd.Path, fd.Type)
		}
	}

	// 清理
	tracker.Untrack(10)
	tracker.Untrack(11)
	tracker.Untrack(12)
}
