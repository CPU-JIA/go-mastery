# System Programming Utilities (sysutil)

跨平台系统编程工具包，提供生产级的系统操作工具。

## 功能模块

### 1. process - 进程管理

提供跨平台的进程管理功能：

```go
import "go-mastery/09-system-programming/sysutil/process"

// 获取进程列表
mgr := process.NewManager()
procs, err := mgr.List()

// 获取当前进程信息
self, err := process.Self()

// 按名称查找进程
procs, err := mgr.FindByName("nginx")

// 获取进程树
tree, err := mgr.GetTree(1)
fmt.Println(tree.Print())

// 监控进程
monitor := process.NewMonitor(pid, 100*time.Millisecond)
events := monitor.Start()
for event := range events {
    fmt.Printf("Event: %v\n", event.Type)
}

// 获取 CPU/内存 使用最高的进程
topCPU, _ := process.TopByCPU(10)
topMem, _ := process.TopByMemory(10)
```

### 2. network - 网络工具

提供网络诊断和连接管理功能：

```go
import "go-mastery/09-system-programming/sysutil/network"

// 获取网络接口信息
ifaces, err := network.GetInterfaces()
localIP, err := network.GetLocalIP()

// 端口扫描
scanner := network.NewPortScanner()
result := scanner.ScanPort("127.0.0.1", 80)
results := scanner.ScanCommonPorts("example.com")

// TCP Ping
result := network.TCPPing("google.com", 443, 5*time.Second)
stats, results := network.TCPPingN("google.com", 443, 10, 5*time.Second)

// DNS 解析
dns := network.ResolveDNS("example.com")
names, _ := network.ReverseDNS("8.8.8.8")

// 连接池
pool, err := network.NewConnPool(network.PoolConfig{
    Address:     "localhost:6379",
    MaxSize:     10,
    MinSize:     2,
    MaxIdleTime: 5 * time.Minute,
})
conn, err := pool.Get(ctx)
defer conn.Close()

// 工具函数
port, _ := network.GetFreePort()
open := network.IsPortOpen("localhost", 8080, time.Second)
isPrivate := network.IsPrivateIP(net.ParseIP("192.168.1.1"))
```

### 3. resource - 资源管理

提供系统资源监控和管理功能：

```go
import "go-mastery/09-system-programming/sysutil/resource"

// 内存信息
mem, err := resource.GetMemoryInfo()
fmt.Printf("Total: %s, Used: %.1f%%\n",
    resource.FormatBytes(mem.Total), mem.UsedPercent)

// CPU 信息和使用率
cpu, err := resource.GetCPUInfo()
usage, err := resource.GetCPUUsage()

// 磁盘信息
disks, err := resource.GetDiskInfo()

// 文件描述符限制
limits, err := resource.GetFDLimits()

// Go 运行时信息
runtime := resource.GetRuntimeInfo()
procMem := resource.GetProcessMemoryInfo()

// 文件描述符追踪器（检测泄漏）
tracker := resource.NewFDTracker()
tracker.Track(fd, "/path/to/file", "file")
leaks := tracker.FindLeaks(1 * time.Hour)

// 资源监控器
monitor := resource.NewResourceMonitor(time.Second)
monitor.OnMemory(func(info resource.MemoryInfo) {
    fmt.Printf("Memory: %.1f%%\n", info.UsedPercent)
})
monitor.OnCPU(func(usage resource.CPUUsage) {
    fmt.Printf("CPU: %.1f%%\n", usage.Total)
})
monitor.Start()
defer monitor.Stop()

// 工具函数
formatted := resource.FormatBytes(1073741824) // "1.00 GB"
bytes, _ := resource.ParseBytes("1.5 GB")
```

## 跨平台支持

| 功能       | Linux             | macOS             | Windows                 |
| ---------- | ----------------- | ----------------- | ----------------------- |
| 进程列表   | ✅ /proc          | ✅ ps             | ✅ Windows API          |
| 进程详情   | ✅ /proc          | ✅ ps             | ✅ Windows API          |
| 内存信息   | ✅ /proc/meminfo  | ✅ vm_stat        | ✅ GlobalMemoryStatusEx |
| CPU 信息   | ✅ /proc/cpuinfo  | ✅ sysctl         | ✅ WMIC                 |
| CPU 使用率 | ✅ /proc/stat     | ✅ top            | ✅ WMIC                 |
| 磁盘信息   | ✅ /proc/mounts   | ✅ df             | ✅ GetDiskFreeSpaceEx   |
| FD 限制    | ✅ getrlimit      | ✅ getrlimit      | ⚠️ 有限支持             |
| 网络接口   | ✅ net.Interfaces | ✅ net.Interfaces | ✅ net.Interfaces       |

## 安装

```bash
go get go-mastery/09-system-programming/sysutil
```

## 测试

```bash
go test ./... -v
```

## 注意事项

1. **权限要求**：某些操作可能需要管理员/root 权限
2. **平台差异**：部分功能在不同平台上的行为可能略有不同
3. **性能考虑**：进程列表获取在 Windows 上可能较慢
4. **资源清理**：使用连接池和监控器后记得调用 Close/Stop

## 许可证

MIT License
