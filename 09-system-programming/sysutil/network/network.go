/*
Package network 提供跨平台的网络诊断和工具。

本包支持以下功能：
  - 网络接口信息获取
  - 端口扫描和检测
  - 连接池管理
  - 网络延迟测量
  - DNS 解析工具
  - TCP/UDP 工具函数

跨平台支持：
  - Windows: 使用 Windows Socket API
  - Linux/Unix: 使用标准 socket 和 netlink
*/
package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ===================
// 错误定义
// ===================

var (
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("connection failed")
	// ErrTimeout 操作超时
	ErrTimeout = errors.New("operation timed out")
	// ErrPortClosed 端口关闭
	ErrPortClosed = errors.New("port is closed")
	// ErrInvalidAddress 无效地址
	ErrInvalidAddress = errors.New("invalid address")
	// ErrPoolExhausted 连接池耗尽
	ErrPoolExhausted = errors.New("connection pool exhausted")
	// ErrPoolClosed 连接池已关闭
	ErrPoolClosed = errors.New("connection pool is closed")
)

// ===================
// 网络接口信息
// ===================

// InterfaceInfo 网络接口信息
type InterfaceInfo struct {
	// Name 接口名称
	Name string
	// Index 接口索引
	Index int
	// MTU 最大传输单元
	MTU int
	// HardwareAddr MAC地址
	HardwareAddr net.HardwareAddr
	// Flags 接口标志
	Flags net.Flags
	// Addresses IP地址列表
	Addresses []net.Addr
	// IPv4Addresses IPv4地址列表
	IPv4Addresses []string
	// IPv6Addresses IPv6地址列表
	IPv6Addresses []string
	// IsUp 接口是否启用
	IsUp bool
	// IsLoopback 是否是回环接口
	IsLoopback bool
	// IsPointToPoint 是否是点对点接口
	IsPointToPoint bool
	// IsMulticast 是否支持多播
	IsMulticast bool
	// IsBroadcast 是否支持广播
	IsBroadcast bool
}

// GetInterfaces 获取所有网络接口信息
func GetInterfaces() ([]*InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	var result []*InterfaceInfo

	for _, iface := range ifaces {
		info := &InterfaceInfo{
			Name:           iface.Name,
			Index:          iface.Index,
			MTU:            iface.MTU,
			HardwareAddr:   iface.HardwareAddr,
			Flags:          iface.Flags,
			IsUp:           iface.Flags&net.FlagUp != 0,
			IsLoopback:     iface.Flags&net.FlagLoopback != 0,
			IsPointToPoint: iface.Flags&net.FlagPointToPoint != 0,
			IsMulticast:    iface.Flags&net.FlagMulticast != 0,
			IsBroadcast:    iface.Flags&net.FlagBroadcast != 0,
		}

		// 获取地址
		addrs, err := iface.Addrs()
		if err == nil {
			info.Addresses = addrs
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}

				ip := ipNet.IP
				if ip4 := ip.To4(); ip4 != nil {
					info.IPv4Addresses = append(info.IPv4Addresses, ip4.String())
				} else {
					info.IPv6Addresses = append(info.IPv6Addresses, ip.String())
				}
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// GetDefaultInterface 获取默认网络接口（有默认路由的接口）
func GetDefaultInterface() (*InterfaceInfo, error) {
	ifaces, err := GetInterfaces()
	if err != nil {
		return nil, err
	}

	// 查找第一个启用的非回环接口
	for _, iface := range ifaces {
		if iface.IsUp && !iface.IsLoopback && len(iface.IPv4Addresses) > 0 {
			return iface, nil
		}
	}

	return nil, fmt.Errorf("no default interface found")
}

// GetLocalIP 获取本机IP地址
func GetLocalIP() (string, error) {
	iface, err := GetDefaultInterface()
	if err != nil {
		return "", err
	}

	if len(iface.IPv4Addresses) > 0 {
		return iface.IPv4Addresses[0], nil
	}

	return "", fmt.Errorf("no IPv4 address found")
}

// ===================
// 端口扫描
// ===================

// PortScanner 端口扫描器
type PortScanner struct {
	// Timeout 单个端口扫描超时时间
	Timeout time.Duration
	// Concurrency 并发数
	Concurrency int
	// Protocol 协议（tcp/udp）
	Protocol string
}

// PortResult 端口扫描结果
type PortResult struct {
	// Port 端口号
	Port int
	// Open 是否开放
	Open bool
	// Service 服务名称（如果已知）
	Service string
	// Banner 服务横幅（如果获取到）
	Banner string
	// Latency 连接延迟
	Latency time.Duration
	// Error 错误信息
	Error error
}

// NewPortScanner 创建端口扫描器
func NewPortScanner() *PortScanner {
	return &PortScanner{
		Timeout:     2 * time.Second,
		Concurrency: 100,
		Protocol:    "tcp",
	}
}

// ScanPort 扫描单个端口
func (ps *PortScanner) ScanPort(host string, port int) *PortResult {
	result := &PortResult{
		Port: port,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout(ps.Protocol, address, ps.Timeout)
	result.Latency = time.Since(start)

	if err != nil {
		result.Open = false
		result.Error = err
		return result
	}
	defer conn.Close()

	result.Open = true
	result.Service = getServiceName(port)

	// 尝试获取 banner
	if ps.Protocol == "tcp" {
		result.Banner = ps.grabBanner(conn)
	}

	return result
}

// ScanPorts 扫描多个端口
func (ps *PortScanner) ScanPorts(host string, ports []int) []*PortResult {
	results := make([]*PortResult, len(ports))
	var wg sync.WaitGroup

	// 使用信号量控制并发
	sem := make(chan struct{}, ps.Concurrency)

	for i, port := range ports {
		wg.Add(1)
		go func(idx, p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = ps.ScanPort(host, p)
		}(i, port)
	}

	wg.Wait()
	return results
}

// ScanRange 扫描端口范围
func (ps *PortScanner) ScanRange(host string, startPort, endPort int) []*PortResult {
	var ports []int
	for p := startPort; p <= endPort; p++ {
		ports = append(ports, p)
	}
	return ps.ScanPorts(host, ports)
}

// ScanCommonPorts 扫描常用端口
func (ps *PortScanner) ScanCommonPorts(host string) []*PortResult {
	commonPorts := []int{
		21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445,
		993, 995, 1723, 3306, 3389, 5432, 5900, 8080, 8443,
	}
	return ps.ScanPorts(host, commonPorts)
}

// grabBanner 获取服务横幅
func (ps *PortScanner) grabBanner(conn net.Conn) string {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(buffer[:n]))
}

// getServiceName 根据端口号获取服务名称
func getServiceName(port int) string {
	services := map[int]string{
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		111:   "RPC",
		135:   "MSRPC",
		139:   "NetBIOS",
		143:   "IMAP",
		443:   "HTTPS",
		445:   "SMB",
		993:   "IMAPS",
		995:   "POP3S",
		1433:  "MSSQL",
		1521:  "Oracle",
		1723:  "PPTP",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		5900:  "VNC",
		6379:  "Redis",
		8080:  "HTTP-Proxy",
		8443:  "HTTPS-Alt",
		27017: "MongoDB",
	}

	if name, ok := services[port]; ok {
		return name
	}
	return ""
}

// ===================
// 连接池
// ===================

// ConnPool TCP连接池
type ConnPool struct {
	// 配置
	address     string
	maxSize     int
	minSize     int
	maxIdleTime time.Duration
	dialTimeout time.Duration

	// 状态
	pool   chan net.Conn
	size   int32
	closed int32
	mu     sync.Mutex

	// 统计
	stats PoolStats
}

// PoolStats 连接池统计
type PoolStats struct {
	// TotalConnections 总连接数
	TotalConnections int64
	// ActiveConnections 活跃连接数
	ActiveConnections int64
	// IdleConnections 空闲连接数
	IdleConnections int64
	// WaitCount 等待获取连接的次数
	WaitCount int64
	// WaitDuration 等待获取连接的总时间
	WaitDuration time.Duration
	// HitCount 命中次数（从池中获取）
	HitCount int64
	// MissCount 未命中次数（新建连接）
	MissCount int64
	// TimeoutCount 超时次数
	TimeoutCount int64
}

// PoolConfig 连接池配置
type PoolConfig struct {
	// Address 服务器地址
	Address string
	// MaxSize 最大连接数
	MaxSize int
	// MinSize 最小连接数
	MinSize int
	// MaxIdleTime 最大空闲时间
	MaxIdleTime time.Duration
	// DialTimeout 连接超时时间
	DialTimeout time.Duration
}

// NewConnPool 创建连接池
func NewConnPool(config PoolConfig) (*ConnPool, error) {
	if config.MaxSize <= 0 {
		config.MaxSize = 10
	}
	if config.MinSize < 0 {
		config.MinSize = 0
	}
	if config.MinSize > config.MaxSize {
		config.MinSize = config.MaxSize
	}
	if config.MaxIdleTime <= 0 {
		config.MaxIdleTime = 5 * time.Minute
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second
	}

	pool := &ConnPool{
		address:     config.Address,
		maxSize:     config.MaxSize,
		minSize:     config.MinSize,
		maxIdleTime: config.MaxIdleTime,
		dialTimeout: config.DialTimeout,
		pool:        make(chan net.Conn, config.MaxSize),
	}

	// 预创建最小连接数
	for i := 0; i < config.MinSize; i++ {
		conn, err := pool.dial()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create initial connections: %w", err)
		}
		pool.pool <- conn
		atomic.AddInt32(&pool.size, 1)
	}

	// 启动清理协程
	go pool.cleaner()

	return pool, nil
}

// Get 从池中获取连接
func (p *ConnPool) Get(ctx context.Context) (net.Conn, error) {
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrPoolClosed
	}

	atomic.AddInt64(&p.stats.WaitCount, 1)
	start := time.Now()

	// 尝试从池中获取
	select {
	case conn := <-p.pool:
		atomic.AddInt64(&p.stats.HitCount, 1)
		atomic.AddInt64(&p.stats.ActiveConnections, 1)
		atomic.AddInt64(&p.stats.IdleConnections, -1)
		p.stats.WaitDuration += time.Since(start)
		return &pooledConn{Conn: conn, pool: p}, nil
	default:
	}

	// 池中没有可用连接，尝试创建新连接
	if atomic.LoadInt32(&p.size) < int32(p.maxSize) {
		conn, err := p.dial()
		if err != nil {
			atomic.AddInt64(&p.stats.MissCount, 1)
			return nil, err
		}
		atomic.AddInt32(&p.size, 1)
		atomic.AddInt64(&p.stats.TotalConnections, 1)
		atomic.AddInt64(&p.stats.ActiveConnections, 1)
		atomic.AddInt64(&p.stats.MissCount, 1)
		p.stats.WaitDuration += time.Since(start)
		return &pooledConn{Conn: conn, pool: p}, nil
	}

	// 等待可用连接
	select {
	case conn := <-p.pool:
		atomic.AddInt64(&p.stats.HitCount, 1)
		atomic.AddInt64(&p.stats.ActiveConnections, 1)
		atomic.AddInt64(&p.stats.IdleConnections, -1)
		p.stats.WaitDuration += time.Since(start)
		return &pooledConn{Conn: conn, pool: p}, nil
	case <-ctx.Done():
		atomic.AddInt64(&p.stats.TimeoutCount, 1)
		return nil, ErrTimeout
	}
}

// Put 将连接放回池中
func (p *ConnPool) Put(conn net.Conn) {
	if atomic.LoadInt32(&p.closed) == 1 {
		conn.Close()
		return
	}

	atomic.AddInt64(&p.stats.ActiveConnections, -1)

	select {
	case p.pool <- conn:
		atomic.AddInt64(&p.stats.IdleConnections, 1)
	default:
		// 池已满，关闭连接
		conn.Close()
		atomic.AddInt32(&p.size, -1)
	}
}

// Close 关闭连接池
func (p *ConnPool) Close() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return nil
	}

	close(p.pool)

	// 关闭所有连接
	for conn := range p.pool {
		conn.Close()
	}

	return nil
}

// Stats 获取连接池统计信息
func (p *ConnPool) Stats() PoolStats {
	return p.stats
}

// Size 获取当前连接数
func (p *ConnPool) Size() int {
	return int(atomic.LoadInt32(&p.size))
}

// dial 创建新连接
func (p *ConnPool) dial() (net.Conn, error) {
	return net.DialTimeout("tcp", p.address, p.dialTimeout)
}

// cleaner 清理空闲连接
func (p *ConnPool) cleaner() {
	ticker := time.NewTicker(p.maxIdleTime / 2)
	defer ticker.Stop()

	for range ticker.C {
		if atomic.LoadInt32(&p.closed) == 1 {
			return
		}

		// 清理超过最小连接数的空闲连接
		for atomic.LoadInt32(&p.size) > int32(p.minSize) {
			select {
			case conn := <-p.pool:
				conn.Close()
				atomic.AddInt32(&p.size, -1)
				atomic.AddInt64(&p.stats.IdleConnections, -1)
			default:
				return
			}
		}
	}
}

// pooledConn 池化连接包装器
type pooledConn struct {
	net.Conn
	pool   *ConnPool
	closed bool
	mu     sync.Mutex
}

// Close 关闭连接（实际上是放回池中）
func (c *pooledConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	c.pool.Put(c.Conn)
	return nil
}

// ===================
// 网络延迟测量
// ===================

// PingResult Ping结果
type PingResult struct {
	// Address 目标地址
	Address string
	// RTT 往返时间
	RTT time.Duration
	// Success 是否成功
	Success bool
	// Error 错误信息
	Error error
	// Timestamp 测量时间
	Timestamp time.Time
}

// PingStats Ping统计
type PingStats struct {
	// PacketsSent 发送的包数
	PacketsSent int
	// PacketsRecv 接收的包数
	PacketsRecv int
	// PacketLoss 丢包率
	PacketLoss float64
	// MinRTT 最小RTT
	MinRTT time.Duration
	// MaxRTT 最大RTT
	MaxRTT time.Duration
	// AvgRTT 平均RTT
	AvgRTT time.Duration
	// StdDevRTT RTT标准差
	StdDevRTT time.Duration
}

// TCPPing 使用TCP进行延迟测量
// 这是一个跨平台的替代方案，不需要 root 权限
func TCPPing(address string, port int, timeout time.Duration) *PingResult {
	result := &PingResult{
		Address:   address,
		Timestamp: time.Now(),
	}

	target := fmt.Sprintf("%s:%d", address, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", target, timeout)
	result.RTT = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}
	defer conn.Close()

	result.Success = true
	return result
}

// TCPPingN 执行多次TCP Ping并返回统计信息
func TCPPingN(address string, port int, count int, timeout time.Duration) (*PingStats, []*PingResult) {
	results := make([]*PingResult, count)
	var rtts []time.Duration

	for i := 0; i < count; i++ {
		results[i] = TCPPing(address, port, timeout)
		if results[i].Success {
			rtts = append(rtts, results[i].RTT)
		}
		if i < count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	stats := &PingStats{
		PacketsSent: count,
		PacketsRecv: len(rtts),
	}

	if count > 0 {
		stats.PacketLoss = float64(count-len(rtts)) / float64(count) * 100
	}

	if len(rtts) > 0 {
		// 计算统计信息
		sort.Slice(rtts, func(i, j int) bool {
			return rtts[i] < rtts[j]
		})

		stats.MinRTT = rtts[0]
		stats.MaxRTT = rtts[len(rtts)-1]

		var total time.Duration
		for _, rtt := range rtts {
			total += rtt
		}
		stats.AvgRTT = total / time.Duration(len(rtts))

		// 计算标准差
		var variance float64
		avgNs := float64(stats.AvgRTT.Nanoseconds())
		for _, rtt := range rtts {
			diff := float64(rtt.Nanoseconds()) - avgNs
			variance += diff * diff
		}
		variance /= float64(len(rtts))
		stats.StdDevRTT = time.Duration(int64(variance)) // 简化计算
	}

	return stats, results
}

// ===================
// DNS 工具
// ===================

// DNSResult DNS解析结果
type DNSResult struct {
	// Host 主机名
	Host string
	// IPs IP地址列表
	IPs []net.IP
	// IPv4 IPv4地址列表
	IPv4 []string
	// IPv6 IPv6地址列表
	IPv6 []string
	// CNAME CNAME记录
	CNAME string
	// MX MX记录
	MX []*net.MX
	// NS NS记录
	NS []*net.NS
	// TXT TXT记录
	TXT []string
	// Latency 解析延迟
	Latency time.Duration
	// Error 错误信息
	Error error
}

// ResolveDNS 解析DNS
func ResolveDNS(host string) *DNSResult {
	result := &DNSResult{
		Host: host,
	}

	start := time.Now()

	// 解析IP地址
	ips, err := net.LookupIP(host)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	result.IPs = ips
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			result.IPv4 = append(result.IPv4, ip4.String())
		} else {
			result.IPv6 = append(result.IPv6, ip.String())
		}
	}

	// 解析CNAME
	cname, _ := net.LookupCNAME(host)
	result.CNAME = cname

	// 解析MX记录
	mx, _ := net.LookupMX(host)
	result.MX = mx

	// 解析NS记录
	ns, _ := net.LookupNS(host)
	result.NS = ns

	// 解析TXT记录
	txt, _ := net.LookupTXT(host)
	result.TXT = txt

	return result
}

// ReverseDNS 反向DNS解析
func ReverseDNS(ip string) ([]string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return nil, fmt.Errorf("reverse DNS lookup failed: %w", err)
	}
	return names, nil
}

// ===================
// 工具函数
// ===================

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetFreePort 获取一个可用的端口
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// GetFreePorts 获取多个可用端口
func GetFreePorts(count int) ([]int, error) {
	ports := make([]int, count)
	listeners := make([]net.Listener, count)

	for i := 0; i < count; i++ {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			// 关闭已创建的监听器
			for j := 0; j < i; j++ {
				listeners[j].Close()
			}
			return nil, err
		}
		listeners[i] = listener
		ports[i] = listener.Addr().(*net.TCPAddr).Port
	}

	// 关闭所有监听器
	for _, l := range listeners {
		l.Close()
	}

	return ports, nil
}

// ParseAddress 解析地址字符串
func ParseAddress(address string) (host string, port int, err error) {
	h, p, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, ErrInvalidAddress
	}

	port, err = net.LookupPort("tcp", p)
	if err != nil {
		return "", 0, ErrInvalidAddress
	}

	return h, port, nil
}

// IsPrivateIP 检查是否是私有IP
func IsPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
	}
	return false
}

// IsLoopback 检查是否是回环地址
func IsLoopback(ip net.IP) bool {
	return ip.IsLoopback()
}
