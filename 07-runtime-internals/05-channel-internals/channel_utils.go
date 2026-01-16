/*
=== Channel 工具函数库 ===

提供可复用的 Channel 监控、分析和优化工具函数。
这些函数可以在生产环境中使用，帮助诊断和优化 Channel 性能。

主要功能：
1. Channel 统计收集器
2. Channel 泄漏检测器
3. Channel 性能分析器
4. 扇入扇出模式实现
5. Channel 池管理
*/

package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 1. Channel 统计收集器
// ==================

// ChannelStatsCollector Channel 统计收集器
// 用于收集和分析 Channel 操作统计信息
type ChannelStatsCollector struct {
	// 发送统计
	sendCount    atomic.Int64
	sendBlocked  atomic.Int64
	sendDuration atomic.Int64 // 纳秒

	// 接收统计
	recvCount    atomic.Int64
	recvBlocked  atomic.Int64
	recvDuration atomic.Int64 // 纳秒

	// Select 统计
	selectCount   atomic.Int64
	selectDefault atomic.Int64

	// 开始时间
	startTime time.Time
}

// NewChannelStatsCollector 创建新的 Channel 统计收集器
func NewChannelStatsCollector() *ChannelStatsCollector {
	return &ChannelStatsCollector{
		startTime: time.Now(),
	}
}

// RecordSend 记录发送操作
func (c *ChannelStatsCollector) RecordSend(duration time.Duration, blocked bool) {
	c.sendCount.Add(1)
	c.sendDuration.Add(int64(duration))
	if blocked {
		c.sendBlocked.Add(1)
	}
}

// RecordRecv 记录接收操作
func (c *ChannelStatsCollector) RecordRecv(duration time.Duration, blocked bool) {
	c.recvCount.Add(1)
	c.recvDuration.Add(int64(duration))
	if blocked {
		c.recvBlocked.Add(1)
	}
}

// RecordSelect 记录 Select 操作
func (c *ChannelStatsCollector) RecordSelect(usedDefault bool) {
	c.selectCount.Add(1)
	if usedDefault {
		c.selectDefault.Add(1)
	}
}

// GetStats 获取统计信息
func (c *ChannelStatsCollector) GetStats() ChannelOperationStats {
	sendCount := c.sendCount.Load()
	recvCount := c.recvCount.Load()

	var avgSendDuration, avgRecvDuration time.Duration
	if sendCount > 0 {
		avgSendDuration = time.Duration(c.sendDuration.Load() / sendCount)
	}
	if recvCount > 0 {
		avgRecvDuration = time.Duration(c.recvDuration.Load() / recvCount)
	}

	return ChannelOperationStats{
		SendCount:       sendCount,
		SendBlocked:     c.sendBlocked.Load(),
		AvgSendDuration: avgSendDuration,
		RecvCount:       recvCount,
		RecvBlocked:     c.recvBlocked.Load(),
		AvgRecvDuration: avgRecvDuration,
		SelectCount:     c.selectCount.Load(),
		SelectDefault:   c.selectDefault.Load(),
		Uptime:          time.Since(c.startTime),
	}
}

// ChannelOperationStats Channel 操作统计
type ChannelOperationStats struct {
	SendCount       int64
	SendBlocked     int64
	AvgSendDuration time.Duration
	RecvCount       int64
	RecvBlocked     int64
	AvgRecvDuration time.Duration
	SelectCount     int64
	SelectDefault   int64
	Uptime          time.Duration
}

// String 格式化输出
func (s ChannelOperationStats) String() string {
	sendBlockRate := float64(0)
	if s.SendCount > 0 {
		sendBlockRate = float64(s.SendBlocked) / float64(s.SendCount) * 100
	}

	recvBlockRate := float64(0)
	if s.RecvCount > 0 {
		recvBlockRate = float64(s.RecvBlocked) / float64(s.RecvCount) * 100
	}

	return fmt.Sprintf(`
=== Channel 操作统计 ===
运行时间: %v

发送操作:
  总数: %d
  阻塞次数: %d (%.2f%%)
  平均耗时: %v

接收操作:
  总数: %d
  阻塞次数: %d (%.2f%%)
  平均耗时: %v

Select 操作:
  总数: %d
  使用 Default: %d
`,
		s.Uptime,
		s.SendCount, s.SendBlocked, sendBlockRate, s.AvgSendDuration,
		s.RecvCount, s.RecvBlocked, recvBlockRate, s.AvgRecvDuration,
		s.SelectCount, s.SelectDefault,
	)
}

// ==================
// 2. Channel 泄漏检测器
// ==================

// ChannelLeakDetector Channel 泄漏检测器
// 用于检测可能导致 Goroutine 泄漏的 Channel 问题
type ChannelLeakDetector struct {
	// 基线 Goroutine 数量
	baseline int
	// 检测阈值
	threshold int
	// 检测间隔
	interval time.Duration
	// 泄漏回调
	onLeak func(current, baseline int, stackTrace string)
	// 运行状态
	running atomic.Bool
	// 停止信号
	stopCh chan struct{}
}

// NewChannelLeakDetector 创建新的 Channel 泄漏检测器
func NewChannelLeakDetector(threshold int, interval time.Duration) *ChannelLeakDetector {
	return &ChannelLeakDetector{
		baseline:  runtime.NumGoroutine(),
		threshold: threshold,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// SetLeakCallback 设置泄漏回调
func (d *ChannelLeakDetector) SetLeakCallback(callback func(current, baseline int, stackTrace string)) {
	d.onLeak = callback
}

// ResetBaseline 重置基线
func (d *ChannelLeakDetector) ResetBaseline() {
	d.baseline = runtime.NumGoroutine()
}

// Start 启动检测
func (d *ChannelLeakDetector) Start() {
	if !d.running.CompareAndSwap(false, true) {
		return
	}

	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				d.check()
			case <-d.stopCh:
				return
			}
		}
	}()
}

// Stop 停止检测
func (d *ChannelLeakDetector) Stop() {
	if d.running.CompareAndSwap(true, false) {
		close(d.stopCh)
	}
}

func (d *ChannelLeakDetector) check() {
	current := runtime.NumGoroutine()
	if current-d.baseline > d.threshold && d.onLeak != nil {
		// 获取 Goroutine 堆栈
		buf := make([]byte, 1<<16)
		n := runtime.Stack(buf, true)
		d.onLeak(current, d.baseline, string(buf[:n]))
	}
}

// CheckNow 立即检测
func (d *ChannelLeakDetector) CheckNow() (leaked bool, current, baseline int) {
	current = runtime.NumGoroutine()
	baseline = d.baseline
	leaked = current-baseline > d.threshold
	return
}

// ==================
// 3. 带监控的 Channel 包装器
// ==================

// MonitoredChannel 带监控的 Channel 包装器
type MonitoredChannel[T any] struct {
	ch        chan T
	name      string
	collector *ChannelStatsCollector
	closed    atomic.Bool
}

// NewMonitoredChannel 创建新的带监控 Channel
func NewMonitoredChannel[T any](name string, size int, collector *ChannelStatsCollector) *MonitoredChannel[T] {
	return &MonitoredChannel[T]{
		ch:        make(chan T, size),
		name:      name,
		collector: collector,
	}
}

// Send 发送数据（带监控）
func (mc *MonitoredChannel[T]) Send(ctx context.Context, value T) error {
	start := time.Now()
	blocked := false

	select {
	case mc.ch <- value:
		// 立即发送成功
	default:
		// 需要阻塞
		blocked = true
		select {
		case mc.ch <- value:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if mc.collector != nil {
		mc.collector.RecordSend(time.Since(start), blocked)
	}
	return nil
}

// Recv 接收数据（带监控）
func (mc *MonitoredChannel[T]) Recv(ctx context.Context) (T, bool, error) {
	start := time.Now()
	blocked := false
	var zero T

	select {
	case value, ok := <-mc.ch:
		if mc.collector != nil {
			mc.collector.RecordRecv(time.Since(start), blocked)
		}
		return value, ok, nil
	default:
		// 需要阻塞
		blocked = true
		select {
		case value, ok := <-mc.ch:
			if mc.collector != nil {
				mc.collector.RecordRecv(time.Since(start), blocked)
			}
			return value, ok, nil
		case <-ctx.Done():
			return zero, false, ctx.Err()
		}
	}
}

// TrySend 尝试发送（非阻塞）
func (mc *MonitoredChannel[T]) TrySend(value T) bool {
	select {
	case mc.ch <- value:
		if mc.collector != nil {
			mc.collector.RecordSend(0, false)
		}
		return true
	default:
		return false
	}
}

// TryRecv 尝试接收（非阻塞）
func (mc *MonitoredChannel[T]) TryRecv() (T, bool) {
	select {
	case value, ok := <-mc.ch:
		if mc.collector != nil {
			mc.collector.RecordRecv(0, false)
		}
		return value, ok
	default:
		var zero T
		return zero, false
	}
}

// Close 关闭 Channel
func (mc *MonitoredChannel[T]) Close() {
	if mc.closed.CompareAndSwap(false, true) {
		close(mc.ch)
	}
}

// Len 获取当前长度
func (mc *MonitoredChannel[T]) Len() int {
	return len(mc.ch)
}

// Cap 获取容量
func (mc *MonitoredChannel[T]) Cap() int {
	return cap(mc.ch)
}

// ==================
// 4. 扇入扇出模式
// ==================

// FanOut 扇出：将一个 Channel 的数据分发到多个 Channel
func FanOut[T any](ctx context.Context, input <-chan T, numWorkers int) []<-chan T {
	outputs := make([]chan T, numWorkers)
	for i := range outputs {
		outputs[i] = make(chan T)
	}

	go func() {
		defer func() {
			for _, ch := range outputs {
				close(ch)
			}
		}()

		i := 0
		for {
			select {
			case value, ok := <-input:
				if !ok {
					return
				}
				select {
				case outputs[i] <- value:
					i = (i + 1) % numWorkers
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	result := make([]<-chan T, numWorkers)
	for i, ch := range outputs {
		result[i] = ch
	}
	return result
}

// FanIn 扇入：将多个 Channel 的数据合并到一个 Channel
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	output := make(chan T)

	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, input := range inputs {
		go func(ch <-chan T) {
			defer wg.Done()
			for {
				select {
				case value, ok := <-ch:
					if !ok {
						return
					}
					select {
					case output <- value:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// ==================
// 5. Channel 池
// ==================

// ChannelPool Channel 池
// 用于复用 Channel 减少分配
type ChannelPool[T any] struct {
	pool     sync.Pool
	size     int
	borrowed atomic.Int64
	returned atomic.Int64
}

// NewChannelPool 创建新的 Channel 池
func NewChannelPool[T any](channelSize int) *ChannelPool[T] {
	return &ChannelPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return make(chan T, channelSize)
			},
		},
		size: channelSize,
	}
}

// Get 获取 Channel
func (p *ChannelPool[T]) Get() chan T {
	p.borrowed.Add(1)
	return p.pool.Get().(chan T)
}

// Put 归还 Channel
func (p *ChannelPool[T]) Put(ch chan T) {
	// 清空 Channel
	for {
		select {
		case <-ch:
		default:
			p.returned.Add(1)
			p.pool.Put(ch)
			return
		}
	}
}

// GetStats 获取池统计
func (p *ChannelPool[T]) GetStats() ChannelPoolStats {
	return ChannelPoolStats{
		Borrowed: p.borrowed.Load(),
		Returned: p.returned.Load(),
		Size:     p.size,
	}
}

// ChannelPoolStats Channel 池统计
type ChannelPoolStats struct {
	Borrowed int64
	Returned int64
	Size     int
}

// ==================
// 6. 超时 Channel 操作
// ==================

// SendWithTimeout 带超时的发送
func SendWithTimeout[T any](ch chan<- T, value T, timeout time.Duration) bool {
	select {
	case ch <- value:
		return true
	case <-time.After(timeout):
		return false
	}
}

// RecvWithTimeout 带超时的接收
func RecvWithTimeout[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	select {
	case value, ok := <-ch:
		return value, ok
	case <-time.After(timeout):
		var zero T
		return zero, false
	}
}

// ==================
// 7. 批量 Channel 操作
// ==================

// BatchSend 批量发送
func BatchSend[T any](ctx context.Context, ch chan<- T, values []T) error {
	for _, v := range values {
		select {
		case ch <- v:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// BatchRecv 批量接收
func BatchRecv[T any](ctx context.Context, ch <-chan T, count int) ([]T, error) {
	result := make([]T, 0, count)
	for i := 0; i < count; i++ {
		select {
		case value, ok := <-ch:
			if !ok {
				return result, nil
			}
			result = append(result, value)
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}
	return result, nil
}

// BatchRecvWithTimeout 带超时的批量接收
func BatchRecvWithTimeout[T any](ch <-chan T, count int, timeout time.Duration) []T {
	result := make([]T, 0, count)
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for i := 0; i < count; i++ {
		select {
		case value, ok := <-ch:
			if !ok {
				return result
			}
			result = append(result, value)
		case <-timer.C:
			return result
		}
	}
	return result
}

// ==================
// 8. Channel 管道模式
// ==================

// Pipeline 管道处理器
type Pipeline[T any] struct {
	stages []func(context.Context, <-chan T) <-chan T
}

// NewPipeline 创建新的管道
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		stages: make([]func(context.Context, <-chan T) <-chan T, 0),
	}
}

// AddStage 添加处理阶段
func (p *Pipeline[T]) AddStage(stage func(context.Context, <-chan T) <-chan T) *Pipeline[T] {
	p.stages = append(p.stages, stage)
	return p
}

// Run 运行管道
func (p *Pipeline[T]) Run(ctx context.Context, input <-chan T) <-chan T {
	current := input
	for _, stage := range p.stages {
		current = stage(ctx, current)
	}
	return current
}

// ==================
// 9. 便捷函数
// ==================

// Drain 排空 Channel
func Drain[T any](ch <-chan T) {
	for range ch {
		// 丢弃所有值
	}
}

// DrainWithCallback 排空 Channel（带回调）
func DrainWithCallback[T any](ch <-chan T, callback func(T)) {
	for v := range ch {
		callback(v)
	}
}

// Merge 合并多个 Channel（简化版 FanIn）
func Merge[T any](channels ...<-chan T) <-chan T {
	return FanIn(context.Background(), channels...)
}

// Broadcast 广播：将一个值发送到多个 Channel
func Broadcast[T any](value T, channels ...chan<- T) {
	for _, ch := range channels {
		select {
		case ch <- value:
		default:
			// Channel 已满，跳过
		}
	}
}

// OrDone 包装 Channel 以支持 context 取消
func OrDone[T any](ctx context.Context, ch <-chan T) <-chan T {
	output := make(chan T)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				select {
				case output <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output
}

// Tee 分流：将一个 Channel 的数据复制到两个 Channel
func Tee[T any](ctx context.Context, input <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for v := range OrDone(ctx, input) {
			// 使用局部变量避免数据竞争
			v := v
			select {
			case <-ctx.Done():
				return
			case out1 <- v:
			}
			select {
			case <-ctx.Done():
				return
			case out2 <- v:
			}
		}
	}()

	return out1, out2
}

// Take 从 Channel 获取指定数量的值
func Take[T any](ctx context.Context, input <-chan T, n int) <-chan T {
	output := make(chan T)
	go func() {
		defer close(output)
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case output <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output
}

// Skip 跳过 Channel 中的前 n 个值
func Skip[T any](ctx context.Context, input <-chan T, n int) <-chan T {
	output := make(chan T)
	go func() {
		defer close(output)

		// 跳过前 n 个
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-input:
				if !ok {
					return
				}
			}
		}

		// 传递剩余的
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case output <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output
}

// Filter 过滤 Channel 中的值
func Filter[T any](ctx context.Context, input <-chan T, predicate func(T) bool) <-chan T {
	output := make(chan T)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				if predicate(v) {
					select {
					case output <- v:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return output
}

// Map 转换 Channel 中的值
func Map[T, U any](ctx context.Context, input <-chan T, transform func(T) U) <-chan U {
	output := make(chan U)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case output <- transform(v):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output
}
