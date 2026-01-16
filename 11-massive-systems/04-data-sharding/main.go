// Package main 演示数据分片策略
// 本模块涵盖数据库分片的核心概念：
// - 哈希分片（Hash Sharding）
// - 范围分片（Range Sharding）
// - 目录分片（Directory Sharding）
// - 分片再平衡（Rebalancing）
package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// 分片基础定义
// ============================================================================

// Shard 分片
type Shard struct {
	ID        string
	Name      string
	Endpoint  string
	Status    ShardStatus
	DataCount int64
	DataSize  int64 // bytes
	mu        sync.RWMutex
}

// ShardStatus 分片状态
type ShardStatus int

const (
	ShardStatusActive ShardStatus = iota
	ShardStatusReadOnly
	ShardStatusMigrating
	ShardStatusOffline
)

func (s ShardStatus) String() string {
	switch s {
	case ShardStatusActive:
		return "Active"
	case ShardStatusReadOnly:
		return "ReadOnly"
	case ShardStatusMigrating:
		return "Migrating"
	case ShardStatusOffline:
		return "Offline"
	default:
		return "Unknown"
	}
}

// Record 数据记录
type Record struct {
	Key       string
	Value     interface{}
	ShardKey  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ShardingStrategy 分片策略接口
type ShardingStrategy interface {
	Name() string
	GetShard(key string) *Shard
	GetAllShards() []*Shard
	AddShard(shard *Shard) error
	RemoveShard(shardID string) error
	Rebalance() error
}

// ============================================================================
// 哈希分片
// ============================================================================

// HashSharding 哈希分片策略
type HashSharding struct {
	shards     []*Shard
	shardCount int
	mu         sync.RWMutex
}

// NewHashSharding 创建哈希分片策略
func NewHashSharding() *HashSharding {
	return &HashSharding{
		shards: make([]*Shard, 0),
	}
}

func (h *HashSharding) Name() string { return "Hash Sharding" }

// hash 计算哈希值
func (h *HashSharding) hash(key string) uint32 {
	hash := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

func (h *HashSharding) GetShard(key string) *Shard {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.shards) == 0 {
		return nil
	}

	// 简单取模分片
	hashValue := h.hash(key)
	index := int(hashValue) % len(h.shards)

	// 确保选择活跃的分片
	for i := 0; i < len(h.shards); i++ {
		shard := h.shards[(index+i)%len(h.shards)]
		if shard.Status == ShardStatusActive {
			return shard
		}
	}

	return nil
}

func (h *HashSharding) GetAllShards() []*Shard {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]*Shard{}, h.shards...)
}

func (h *HashSharding) AddShard(shard *Shard) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.shards = append(h.shards, shard)
	h.shardCount = len(h.shards)
	fmt.Printf("  [HashSharding] 添加分片: %s (总数: %d)\n", shard.ID, h.shardCount)
	return nil
}

func (h *HashSharding) RemoveShard(shardID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, shard := range h.shards {
		if shard.ID == shardID {
			h.shards = append(h.shards[:i], h.shards[i+1:]...)
			h.shardCount = len(h.shards)
			fmt.Printf("  [HashSharding] 移除分片: %s (总数: %d)\n", shardID, h.shardCount)
			return nil
		}
	}
	return fmt.Errorf("分片不存在: %s", shardID)
}

func (h *HashSharding) Rebalance() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Println("  [HashSharding] 哈希分片需要数据迁移来重新平衡")
	fmt.Println("  [HashSharding] 这通常涉及一致性哈希或虚拟节点")
	return nil
}

// ============================================================================
// 范围分片
// ============================================================================

// RangeSharding 范围分片策略
type RangeSharding struct {
	shards []*RangeShard
	mu     sync.RWMutex
}

// RangeShard 范围分片
type RangeShard struct {
	*Shard
	StartKey string
	EndKey   string
}

// NewRangeSharding 创建范围分片策略
func NewRangeSharding() *RangeSharding {
	return &RangeSharding{
		shards: make([]*RangeShard, 0),
	}
}

func (r *RangeSharding) Name() string { return "Range Sharding" }

func (r *RangeSharding) GetShard(key string) *Shard {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, shard := range r.shards {
		if key >= shard.StartKey && (shard.EndKey == "" || key < shard.EndKey) {
			if shard.Status == ShardStatusActive {
				return shard.Shard
			}
		}
	}
	return nil
}

func (r *RangeSharding) GetAllShards() []*Shard {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Shard, len(r.shards))
	for i, rs := range r.shards {
		result[i] = rs.Shard
	}
	return result
}

// AddRangeShard 添加范围分片
func (r *RangeSharding) AddRangeShard(shard *Shard, startKey, endKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rangeShard := &RangeShard{
		Shard:    shard,
		StartKey: startKey,
		EndKey:   endKey,
	}

	r.shards = append(r.shards, rangeShard)

	// 按起始键排序
	sort.Slice(r.shards, func(i, j int) bool {
		return r.shards[i].StartKey < r.shards[j].StartKey
	})

	fmt.Printf("  [RangeSharding] 添加分片: %s [%s, %s)\n", shard.ID, startKey, endKey)
	return nil
}

func (r *RangeSharding) AddShard(shard *Shard) error {
	return fmt.Errorf("请使用 AddRangeShard 方法添加范围分片")
}

func (r *RangeSharding) RemoveShard(shardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, shard := range r.shards {
		if shard.ID == shardID {
			r.shards = append(r.shards[:i], r.shards[i+1:]...)
			fmt.Printf("  [RangeSharding] 移除分片: %s\n", shardID)
			return nil
		}
	}
	return fmt.Errorf("分片不存在: %s", shardID)
}

func (r *RangeSharding) Rebalance() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Println("  [RangeSharding] 范围分片可以通过分裂热点分片来重新平衡")
	return nil
}

// SplitShard 分裂分片
func (r *RangeSharding) SplitShard(shardID string, splitKey string, newShard *Shard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, shard := range r.shards {
		if shard.ID == shardID {
			// 创建新分片
			newRangeShard := &RangeShard{
				Shard:    newShard,
				StartKey: splitKey,
				EndKey:   shard.EndKey,
			}

			// 更新原分片的结束键
			r.shards[i].EndKey = splitKey

			// 插入新分片
			r.shards = append(r.shards, newRangeShard)

			// 重新排序
			sort.Slice(r.shards, func(i, j int) bool {
				return r.shards[i].StartKey < r.shards[j].StartKey
			})

			fmt.Printf("  [RangeSharding] 分裂分片: %s -> %s (分裂点: %s)\n",
				shardID, newShard.ID, splitKey)
			return nil
		}
	}
	return fmt.Errorf("分片不存在: %s", shardID)
}

// GetRangeInfo 获取范围信息
func (r *RangeSharding) GetRangeInfo() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fmt.Println("\n  范围分片分布:")
	for _, shard := range r.shards {
		endKey := shard.EndKey
		if endKey == "" {
			endKey = "+inf"
		}
		fmt.Printf("    %s: [%s, %s) - %s\n",
			shard.ID, shard.StartKey, endKey, shard.Status)
	}
}

// ============================================================================
// 目录分片
// ============================================================================

// DirectorySharding 目录分片策略
type DirectorySharding struct {
	directory map[string]*Shard // key -> shard 映射
	shards    map[string]*Shard // shardID -> shard
	mu        sync.RWMutex
}

// NewDirectorySharding 创建目录分片策略
func NewDirectorySharding() *DirectorySharding {
	return &DirectorySharding{
		directory: make(map[string]*Shard),
		shards:    make(map[string]*Shard),
	}
}

func (d *DirectorySharding) Name() string { return "Directory Sharding" }

func (d *DirectorySharding) GetShard(key string) *Shard {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if shard, exists := d.directory[key]; exists {
		return shard
	}
	return nil
}

func (d *DirectorySharding) GetAllShards() []*Shard {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*Shard, 0, len(d.shards))
	for _, shard := range d.shards {
		result = append(result, shard)
	}
	return result
}

func (d *DirectorySharding) AddShard(shard *Shard) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.shards[shard.ID] = shard
	fmt.Printf("  [DirectorySharding] 添加分片: %s\n", shard.ID)
	return nil
}

func (d *DirectorySharding) RemoveShard(shardID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.shards[shardID]; !exists {
		return fmt.Errorf("分片不存在: %s", shardID)
	}

	// 移除目录中指向该分片的所有映射
	for key, shard := range d.directory {
		if shard.ID == shardID {
			delete(d.directory, key)
		}
	}

	delete(d.shards, shardID)
	fmt.Printf("  [DirectorySharding] 移除分片: %s\n", shardID)
	return nil
}

// MapKey 映射键到分片
func (d *DirectorySharding) MapKey(key string, shardID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	shard, exists := d.shards[shardID]
	if !exists {
		return fmt.Errorf("分片不存在: %s", shardID)
	}

	d.directory[key] = shard
	return nil
}

// MigrateKey 迁移键到新分片
func (d *DirectorySharding) MigrateKey(key string, newShardID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	newShard, exists := d.shards[newShardID]
	if !exists {
		return fmt.Errorf("目标分片不存在: %s", newShardID)
	}

	oldShard := d.directory[key]
	d.directory[key] = newShard

	if oldShard != nil {
		fmt.Printf("  [DirectorySharding] 迁移键 %s: %s -> %s\n",
			key, oldShard.ID, newShardID)
	} else {
		fmt.Printf("  [DirectorySharding] 映射键 %s -> %s\n", key, newShardID)
	}
	return nil
}

func (d *DirectorySharding) Rebalance() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Println("  [DirectorySharding] 目录分片可以灵活地重新分配键")
	return nil
}

// GetDirectoryStats 获取目录统计
func (d *DirectorySharding) GetDirectoryStats() map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make(map[string]int)
	for _, shard := range d.directory {
		stats[shard.ID]++
	}
	return stats
}

// ============================================================================
// 分片管理器
// ============================================================================

// ShardManager 分片管理器
type ShardManager struct {
	strategy ShardingStrategy
	data     map[string]*Record
	mu       sync.RWMutex
}

// NewShardManager 创建分片管理器
func NewShardManager(strategy ShardingStrategy) *ShardManager {
	return &ShardManager{
		strategy: strategy,
		data:     make(map[string]*Record),
	}
}

// Put 写入数据
func (m *ShardManager) Put(key string, value interface{}) error {
	shard := m.strategy.GetShard(key)
	if shard == nil {
		return fmt.Errorf("无法找到分片: key=%s", key)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if existing, exists := m.data[key]; exists {
		existing.Value = value
		existing.UpdatedAt = now
	} else {
		m.data[key] = &Record{
			Key:       key,
			Value:     value,
			ShardKey:  shard.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	shard.mu.Lock()
	shard.DataCount++
	shard.mu.Unlock()

	return nil
}

// Get 读取数据
func (m *ShardManager) Get(key string) (*Record, error) {
	shard := m.strategy.GetShard(key)
	if shard == nil {
		return nil, fmt.Errorf("无法找到分片: key=%s", key)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if record, exists := m.data[key]; exists {
		return record, nil
	}
	return nil, fmt.Errorf("数据不存在: key=%s", key)
}

// GetShardStats 获取分片统计
func (m *ShardManager) GetShardStats() {
	fmt.Printf("\n  分片统计 (%s):\n", m.strategy.Name())
	for _, shard := range m.strategy.GetAllShards() {
		shard.mu.RLock()
		fmt.Printf("    %s: %d 条记录, 状态=%s\n",
			shard.ID, shard.DataCount, shard.Status)
		shard.mu.RUnlock()
	}
}

// ============================================================================
// 分片再平衡器
// ============================================================================

// Rebalancer 分片再平衡器
type Rebalancer struct {
	manager   *ShardManager
	threshold float64 // 不平衡阈值
}

// NewRebalancer 创建再平衡器
func NewRebalancer(manager *ShardManager, threshold float64) *Rebalancer {
	return &Rebalancer{
		manager:   manager,
		threshold: threshold,
	}
}

// CheckBalance 检查平衡状态
func (r *Rebalancer) CheckBalance() (bool, map[string]int64) {
	shards := r.manager.strategy.GetAllShards()
	if len(shards) == 0 {
		return true, nil
	}

	distribution := make(map[string]int64)
	var total int64
	var min, max int64 = -1, 0

	for _, shard := range shards {
		shard.mu.RLock()
		count := shard.DataCount
		shard.mu.RUnlock()

		distribution[shard.ID] = count
		total += count

		if min < 0 || count < min {
			min = count
		}
		if count > max {
			max = count
		}
	}

	if total == 0 {
		return true, distribution
	}

	avg := total / int64(len(shards))
	imbalance := float64(max-min) / float64(avg)

	fmt.Printf("\n  平衡检查:\n")
	fmt.Printf("    总数据量: %d\n", total)
	fmt.Printf("    平均值: %d\n", avg)
	fmt.Printf("    最小值: %d, 最大值: %d\n", min, max)
	fmt.Printf("    不平衡度: %.2f%% (阈值: %.2f%%)\n", imbalance*100, r.threshold*100)

	return imbalance <= r.threshold, distribution
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateHashSharding() {
	fmt.Println("\n=== 哈希分片演示 ===")
	fmt.Println("特点: 数据均匀分布，但范围查询效率低")

	strategy := NewHashSharding()

	// 添加分片
	for i := 1; i <= 3; i++ {
		shard := &Shard{
			ID:       fmt.Sprintf("shard-%d", i),
			Name:     fmt.Sprintf("分片%d", i),
			Endpoint: fmt.Sprintf("db%d.example.com:5432", i),
			Status:   ShardStatusActive,
		}
		strategy.AddShard(shard)
	}

	manager := NewShardManager(strategy)

	// 写入测试数据
	fmt.Println("\n写入测试数据:")
	testKeys := []string{
		"user:1001", "user:1002", "user:1003",
		"order:2001", "order:2002", "order:2003",
		"product:3001", "product:3002", "product:3003",
	}

	for _, key := range testKeys {
		shard := strategy.GetShard(key)
		fmt.Printf("  %s -> %s\n", key, shard.ID)
		manager.Put(key, map[string]string{"key": key})
	}

	manager.GetShardStats()

	// 演示一致性
	fmt.Println("\n验证路由一致性:")
	for _, key := range testKeys[:3] {
		shard1 := strategy.GetShard(key)
		shard2 := strategy.GetShard(key)
		match := "一致"
		if shard1.ID != shard2.ID {
			match = "不一致"
		}
		fmt.Printf("  %s: %s [%s]\n", key, shard1.ID, match)
	}
}

func demonstrateRangeSharding() {
	fmt.Println("\n=== 范围分片演示 ===")
	fmt.Println("特点: 支持范围查询，但可能产生热点")

	strategy := NewRangeSharding()

	// 添加范围分片
	strategy.AddRangeShard(&Shard{
		ID:       "shard-a",
		Name:     "分片A",
		Endpoint: "db-a.example.com:5432",
		Status:   ShardStatusActive,
	}, "a", "m")

	strategy.AddRangeShard(&Shard{
		ID:       "shard-m",
		Name:     "分片M",
		Endpoint: "db-m.example.com:5432",
		Status:   ShardStatusActive,
	}, "m", "")

	strategy.GetRangeInfo()

	manager := NewShardManager(strategy)

	// 写入测试数据
	fmt.Println("\n写入测试数据:")
	testKeys := []string{
		"apple", "banana", "cherry",
		"mango", "orange", "peach",
	}

	for _, key := range testKeys {
		shard := strategy.GetShard(key)
		if shard != nil {
			fmt.Printf("  %s -> %s\n", key, shard.ID)
			manager.Put(key, map[string]string{"fruit": key})
		}
	}

	manager.GetShardStats()

	// 演示分片分裂
	fmt.Println("\n模拟热点分片分裂:")
	newShard := &Shard{
		ID:       "shard-a2",
		Name:     "分片A2",
		Endpoint: "db-a2.example.com:5432",
		Status:   ShardStatusActive,
	}
	strategy.SplitShard("shard-a", "f", newShard)
	strategy.GetRangeInfo()

	// 验证分裂后的路由
	fmt.Println("\n分裂后的路由:")
	for _, key := range testKeys[:3] {
		shard := strategy.GetShard(key)
		if shard != nil {
			fmt.Printf("  %s -> %s\n", key, shard.ID)
		}
	}
}

func demonstrateDirectorySharding() {
	fmt.Println("\n=== 目录分片演示 ===")
	fmt.Println("特点: 完全灵活的映射，但需要维护目录")

	strategy := NewDirectorySharding()

	// 添加分片
	for i := 1; i <= 3; i++ {
		shard := &Shard{
			ID:       fmt.Sprintf("shard-%d", i),
			Name:     fmt.Sprintf("分片%d", i),
			Endpoint: fmt.Sprintf("db%d.example.com:5432", i),
			Status:   ShardStatusActive,
		}
		strategy.AddShard(shard)
	}

	// 手动映射键
	fmt.Println("\n配置键映射:")
	mappings := map[string]string{
		"vip:user:1":    "shard-1", // VIP用户放在专用分片
		"vip:user:2":    "shard-1",
		"normal:user:1": "shard-2",
		"normal:user:2": "shard-2",
		"normal:user:3": "shard-3",
	}

	for key, shardID := range mappings {
		strategy.MigrateKey(key, shardID)
	}

	// 显示分布
	fmt.Println("\n键分布统计:")
	stats := strategy.GetDirectoryStats()
	for shardID, count := range stats {
		fmt.Printf("  %s: %d 个键\n", shardID, count)
	}

	// 演示键迁移
	fmt.Println("\n模拟键迁移 (VIP用户迁移到新分片):")
	strategy.MigrateKey("vip:user:1", "shard-3")

	fmt.Println("\n迁移后的分布:")
	stats = strategy.GetDirectoryStats()
	for shardID, count := range stats {
		fmt.Printf("  %s: %d 个键\n", shardID, count)
	}
}

func demonstrateRebalancing() {
	fmt.Println("\n=== 分片再平衡演示 ===")
	fmt.Println("场景: 检测和处理数据倾斜")

	strategy := NewHashSharding()

	// 添加分片
	shards := make([]*Shard, 3)
	for i := 0; i < 3; i++ {
		shards[i] = &Shard{
			ID:       fmt.Sprintf("shard-%d", i+1),
			Name:     fmt.Sprintf("分片%d", i+1),
			Endpoint: fmt.Sprintf("db%d.example.com:5432", i+1),
			Status:   ShardStatusActive,
		}
		strategy.AddShard(shards[i])
	}

	manager := NewShardManager(strategy)

	// 模拟不均匀的数据分布
	fmt.Println("\n模拟数据写入:")
	shards[0].DataCount = 1000
	shards[1].DataCount = 500
	shards[2].DataCount = 100

	manager.GetShardStats()

	// 创建再平衡器
	rebalancer := NewRebalancer(manager, 0.5) // 50% 不平衡阈值

	// 检查平衡状态
	balanced, distribution := rebalancer.CheckBalance()

	if !balanced {
		fmt.Println("\n  检测到数据倾斜，需要再平衡")
		fmt.Println("  再平衡策略:")
		fmt.Println("    1. 识别热点分片")
		fmt.Println("    2. 计算迁移计划")
		fmt.Println("    3. 执行数据迁移")
		fmt.Println("    4. 更新路由表")

		// 计算理想分布
		var total int64
		for _, count := range distribution {
			total += count
		}
		ideal := total / int64(len(distribution))

		fmt.Printf("\n  迁移计划 (目标: 每分片 %d 条):\n", ideal)
		for shardID, count := range distribution {
			diff := count - ideal
			if diff > 0 {
				fmt.Printf("    %s: 迁出 %d 条\n", shardID, diff)
			} else if diff < 0 {
				fmt.Printf("    %s: 迁入 %d 条\n", shardID, -diff)
			} else {
				fmt.Printf("    %s: 无需迁移\n", shardID)
			}
		}
	} else {
		fmt.Println("\n  数据分布均衡，无需再平衡")
	}
}

func main() {
	fmt.Println("=== 数据分片策略 ===")
	fmt.Println()
	fmt.Println("本模块演示三种常见的数据分片策略:")
	fmt.Println("1. 哈希分片 - 均匀分布，适合点查询")
	fmt.Println("2. 范围分片 - 支持范围查询，需处理热点")
	fmt.Println("3. 目录分片 - 完全灵活，适合特殊需求")
	fmt.Println("4. 分片再平衡 - 处理数据倾斜")

	demonstrateHashSharding()
	demonstrateRangeSharding()
	demonstrateDirectorySharding()
	demonstrateRebalancing()

	fmt.Println("\n=== 数据分片演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 哈希分片: 简单高效，但扩容需要数据迁移")
	fmt.Println("- 范围分片: 支持范围查询，可通过分裂处理热点")
	fmt.Println("- 目录分片: 最灵活，但需要维护映射表")
	fmt.Println("- 再平衡: 监控数据分布，及时处理倾斜")
	fmt.Println()
	fmt.Println("选择建议:")
	fmt.Println("- 随机访问为主 -> 哈希分片")
	fmt.Println("- 范围查询为主 -> 范围分片")
	fmt.Println("- 特殊路由需求 -> 目录分片")
	fmt.Println("- 大规模系统 -> 组合使用 + 一致性哈希")
}
