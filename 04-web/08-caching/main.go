package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

/*
缓存策略和性能优化练习

本练习涵盖Go语言中的缓存策略和性能优化，包括：
1. 内存缓存（TTL和LRU策略）
2. Redis分布式缓存
3. HTTP缓存头设置
4. 数据库查询缓存
5. 缓存预热和失效策略
6. 分布式缓存一致性
7. 性能监控和指标
8. 缓存模式（透写、回写、旁路）

主要概念：
- 缓存命中率优化
- 缓存穿透和雪崩防护
- 缓存更新策略
- 性能监控指标
- 内存管理优化
*/

// === 缓存接口定义 ===

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Stats() CacheStats
}

type CacheStats struct {
	Hits       int64   `json:"hits"`
	Misses     int64   `json:"misses"`
	HitRate    float64 `json:"hit_rate"`
	Size       int     `json:"size"`
	MaxSize    int     `json:"max_size"`
	MemoryUsed int64   `json:"memory_used"`
}

// === LRU内存缓存实现 ===

type LRUCache struct {
	maxSize    int
	items      map[string]*CacheItem
	head       *CacheItem
	tail       *CacheItem
	mutex      sync.RWMutex
	hits       int64
	misses     int64
	memoryUsed int64
}

type CacheItem struct {
	Key        string
	Value      interface{}
	ExpireTime time.Time
	Size       int64
	Prev       *CacheItem
	Next       *CacheItem
}

func NewLRUCache(maxSize int) *LRUCache {
	cache := &LRUCache{
		maxSize: maxSize,
		items:   make(map[string]*CacheItem),
	}

	// 创建虚拟头尾节点
	cache.head = &CacheItem{}
	cache.tail = &CacheItem{}
	cache.head.Next = cache.tail
	cache.tail.Prev = cache.head

	// 启动清理协程
	go cache.cleanupExpired()

	return cache
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// 检查是否过期
	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		c.removeItem(item)
		c.misses++
		return nil, false
	}

	// 移动到链表头部（最近使用）
	c.moveToHead(item)
	c.hits++
	return item.Value, true
}

func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 计算值大小（简单估算）
	size := c.estimateSize(value)

	// 如果键已存在，更新值
	if existingItem, exists := c.items[key]; exists {
		c.memoryUsed -= existingItem.Size
		existingItem.Value = value
		existingItem.Size = size
		if ttl > 0 {
			existingItem.ExpireTime = time.Now().Add(ttl)
		} else {
			existingItem.ExpireTime = time.Time{}
		}
		c.memoryUsed += size
		c.moveToHead(existingItem)
		return nil
	}

	// 创建新项目
	var expireTime time.Time
	if ttl > 0 {
		expireTime = time.Now().Add(ttl)
	}

	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpireTime: expireTime,
		Size:       size,
	}

	// 检查是否需要淘汰
	for len(c.items) >= c.maxSize {
		c.removeTail()
	}

	// 添加到缓存
	c.items[key] = item
	c.addToHead(item)
	c.memoryUsed += size

	return nil
}

func (c *LRUCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
	}

	return nil
}

func (c *LRUCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	c.head.Next = c.tail
	c.tail.Prev = c.head
	c.memoryUsed = 0

	return nil
}

func (c *LRUCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var hitRate float64
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:       c.hits,
		Misses:     c.misses,
		HitRate:    hitRate,
		Size:       len(c.items),
		MaxSize:    c.maxSize,
		MemoryUsed: c.memoryUsed,
	}
}

// LRU链表操作
func (c *LRUCache) addToHead(item *CacheItem) {
	item.Prev = c.head
	item.Next = c.head.Next
	c.head.Next.Prev = item
	c.head.Next = item
}

func (c *LRUCache) removeItem(item *CacheItem) {
	item.Prev.Next = item.Next
	item.Next.Prev = item.Prev
	delete(c.items, item.Key)
	c.memoryUsed -= item.Size
}

func (c *LRUCache) moveToHead(item *CacheItem) {
	c.removeFromList(item)
	c.addToHead(item)
}

func (c *LRUCache) removeFromList(item *CacheItem) {
	item.Prev.Next = item.Next
	item.Next.Prev = item.Prev
}

func (c *LRUCache) removeTail() {
	lastItem := c.tail.Prev
	if lastItem != c.head {
		c.removeItem(lastItem)
	}
}

func (c *LRUCache) estimateSize(value interface{}) int64 {
	// 简单的大小估算
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, float32, float64:
		return 8
	default:
		// 对于复杂对象，使用JSON序列化来估算大小
		if data, err := json.Marshal(v); err == nil {
			return int64(len(data))
		}
		return 64 // 默认大小
	}
}

func (c *LRUCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		var expiredKeys []string

		for key, item := range c.items {
			if !item.ExpireTime.IsZero() && now.After(item.ExpireTime) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			if item, exists := c.items[key]; exists {
				c.removeItem(item)
			}
		}
		c.mutex.Unlock()
	}
}

// === Redis缓存实现 ===

type RedisCache struct {
	client *redis.Client
	hits   int64
	misses int64
	mutex  sync.RWMutex
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: rdb,
	}
}

func (r *RedisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()

	r.mutex.Lock()
	if err == redis.Nil {
		r.misses++
		r.mutex.Unlock()
		return nil, false
	} else if err != nil {
		r.misses++
		r.mutex.Unlock()
		return nil, false
	}

	r.hits++
	r.mutex.Unlock()

	// 尝试解析JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err == nil {
		return result, true
	}

	return val, true
}

func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	ctx := context.Background()

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Clear() error {
	ctx := context.Background()
	return r.client.FlushDB(ctx).Err()
}

func (r *RedisCache) Stats() CacheStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var hitRate float64
	total := r.hits + r.misses
	if total > 0 {
		hitRate = float64(r.hits) / float64(total)
	}

	ctx := context.Background()
	// Note: Redis info not used in basic stats
	_, _ = r.client.Info(ctx, "memory").Result()

	return CacheStats{
		Hits:    r.hits,
		Misses:  r.misses,
		HitRate: hitRate,
		Size:    -1, // Redis不直接提供键数量
	}
}

// === 多级缓存 ===

type MultiLevelCache struct {
	l1Cache Cache // 内存缓存
	l2Cache Cache // Redis缓存
}

func NewMultiLevelCache(l1Cache, l2Cache Cache) *MultiLevelCache {
	return &MultiLevelCache{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
	}
}

func (m *MultiLevelCache) Get(key string) (interface{}, bool) {
	// 首先尝试L1缓存
	if value, ok := m.l1Cache.Get(key); ok {
		return value, true
	}

	// 然后尝试L2缓存
	if value, ok := m.l2Cache.Get(key); ok {
		// 回填到L1缓存
		m.l1Cache.Set(key, value, 5*time.Minute)
		return value, true
	}

	return nil, false
}

func (m *MultiLevelCache) Set(key string, value interface{}, ttl time.Duration) error {
	// 同时设置L1和L2缓存
	m.l1Cache.Set(key, value, ttl)
	return m.l2Cache.Set(key, value, ttl)
}

func (m *MultiLevelCache) Delete(key string) error {
	m.l1Cache.Delete(key)
	return m.l2Cache.Delete(key)
}

func (m *MultiLevelCache) Clear() error {
	m.l1Cache.Clear()
	return m.l2Cache.Clear()
}

func (m *MultiLevelCache) Stats() CacheStats {
	l1Stats := m.l1Cache.Stats()
	l2Stats := m.l2Cache.Stats()

	return CacheStats{
		Hits:    l1Stats.Hits + l2Stats.Hits,
		Misses:  l1Stats.Misses + l2Stats.Misses,
		HitRate: (float64(l1Stats.Hits+l2Stats.Hits) / float64(l1Stats.Hits+l1Stats.Misses+l2Stats.Hits+l2Stats.Misses)),
	}
}

// === 数据模型 ===

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Created  time.Time `json:"created"`
}

type Post struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	Author  User      `json:"author"`
	Views   int       `json:"views"`
	Created time.Time `json:"created"`
}

// === 应用服务 ===

type AppService struct {
	cache Cache
	users []User
	posts []Post
}

func NewAppService(cache Cache) *AppService {
	service := &AppService{
		cache: cache,
	}

	// 创建示例数据
	service.createSampleData()

	return service
}

func (s *AppService) createSampleData() {
	s.users = []User{
		{ID: 1, Username: "alice", Email: "alice@example.com", Role: "admin", Created: time.Now().AddDate(0, -6, 0)},
		{ID: 2, Username: "bob", Email: "bob@example.com", Role: "user", Created: time.Now().AddDate(0, -3, 0)},
		{ID: 3, Username: "charlie", Email: "charlie@example.com", Role: "user", Created: time.Now().AddDate(0, -1, 0)},
	}

	s.posts = []Post{
		{ID: 1, Title: "Go缓存策略", Content: "如何在Go中实现高效的缓存...", Author: s.users[0], Views: 150, Created: time.Now().AddDate(0, 0, -7)},
		{ID: 2, Title: "性能优化技巧", Content: "Web应用性能优化的最佳实践...", Author: s.users[1], Views: 89, Created: time.Now().AddDate(0, 0, -3)},
		{ID: 3, Title: "Redis应用指南", Content: "Redis在分布式系统中的应用...", Author: s.users[0], Views: 234, Created: time.Now().AddDate(0, 0, -1)},
	}
}

// 获取用户（带缓存）
func (s *AppService) GetUser(id int) (*User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)

	// 尝试从缓存获取
	if cached, ok := s.cache.Get(cacheKey); ok {
		if user, ok := cached.(User); ok {
			return &user, nil
		}
	}

	// 模拟数据库查询
	time.Sleep(50 * time.Millisecond) // 模拟DB延迟

	for _, user := range s.users {
		if user.ID == id {
			// 缓存结果
			s.cache.Set(cacheKey, user, 10*time.Minute)
			return &user, nil
		}
	}

	return nil, fmt.Errorf("用户不存在")
}

// 获取文章列表（带缓存）
func (s *AppService) GetPosts(page, limit int) ([]Post, error) {
	cacheKey := fmt.Sprintf("posts:page:%d:limit:%d", page, limit)

	// 尝试从缓存获取
	if cached, ok := s.cache.Get(cacheKey); ok {
		if posts, ok := cached.([]Post); ok {
			return posts, nil
		}
	}

	// 模拟数据库查询
	time.Sleep(100 * time.Millisecond) // 模拟DB延迟

	start := (page - 1) * limit
	end := start + limit

	if start >= len(s.posts) {
		return []Post{}, nil
	}

	if end > len(s.posts) {
		end = len(s.posts)
	}

	result := s.posts[start:end]

	// 缓存结果
	s.cache.Set(cacheKey, result, 5*time.Minute)

	return result, nil
}

// 获取热门文章（带缓存）
func (s *AppService) GetPopularPosts(limit int) ([]Post, error) {
	cacheKey := fmt.Sprintf("popular_posts:%d", limit)

	// 尝试从缓存获取
	if cached, ok := s.cache.Get(cacheKey); ok {
		if posts, ok := cached.([]Post); ok {
			return posts, nil
		}
	}

	// 模拟复杂查询
	time.Sleep(200 * time.Millisecond) // 模拟复杂查询延迟

	// 简单排序（实际应用中会在数据库层面处理）
	posts := make([]Post, len(s.posts))
	copy(posts, s.posts)

	// 按浏览量排序
	for i := 0; i < len(posts)-1; i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[i].Views < posts[j].Views {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	if limit > len(posts) {
		limit = len(posts)
	}

	result := posts[:limit]

	// 缓存结果（较长时间）
	s.cache.Set(cacheKey, result, 30*time.Minute)

	return result, nil
}

// === HTTP处理器 ===

type HTTPHandler struct {
	service *AppService
	cache   Cache
}

func NewHTTPHandler(service *AppService, cache Cache) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		cache:   cache,
	}
}

// 获取用户API
func (h *HTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	start := time.Now()
	user, err := h.service.GetUser(id)
	duration := time.Since(start)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 设置缓存头
	w.Header().Set("Cache-Control", "public, max-age=300") // 5分钟
	w.Header().Set("X-Response-Time", duration.String())
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user)
}

// 获取文章列表API
func (h *HTTPHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	start := time.Now()
	posts, err := h.service.GetPosts(page, limit)
	duration := time.Since(start)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置缓存头
	w.Header().Set("Cache-Control", "public, max-age=180") // 3分钟
	w.Header().Set("X-Response-Time", duration.String())
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"posts": posts,
		"page":  page,
		"limit": limit,
		"total": len(posts),
	}

	json.NewEncoder(w).Encode(response)
}

// 获取热门文章API
func (h *HTTPHandler) GetPopularPosts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 5
	}

	start := time.Now()
	posts, err := h.service.GetPopularPosts(limit)
	duration := time.Since(start)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置缓存头（较长时间）
	w.Header().Set("Cache-Control", "public, max-age=1800") // 30分钟
	w.Header().Set("X-Response-Time", duration.String())
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
		"limit": limit,
	})
}

// 缓存统计API
func (h *HTTPHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.cache.Stats()

	// 添加系统内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemStats := map[string]interface{}{
		"alloc":       bToMb(m.Alloc),
		"total_alloc": bToMb(m.TotalAlloc),
		"sys":         bToMb(m.Sys),
		"num_gc":      m.NumGC,
		"goroutines":  runtime.NumGoroutine(),
	}

	response := map[string]interface{}{
		"cache":  stats,
		"system": systemStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 缓存管理API
func (h *HTTPHandler) ManageCache(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")

	switch action {
	case "clear":
		h.cache.Clear()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "缓存已清空"})

	case "warmup":
		// 预热缓存
		go h.warmupCache()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "缓存预热已开始"})

	default:
		http.Error(w, "无效的操作", http.StatusBadRequest)
	}
}

// 缓存预热
func (h *HTTPHandler) warmupCache() {
	log.Println("开始缓存预热...")

	// 预热用户数据
	for _, user := range []int{1, 2, 3} {
		h.service.GetUser(user)
	}

	// 预热文章列表
	for page := 1; page <= 3; page++ {
		h.service.GetPosts(page, 10)
	}

	// 预热热门文章
	h.service.GetPopularPosts(5)

	log.Println("缓存预热完成")
}

// === 性能监控中间件 ===

type PerformanceMiddleware struct {
	requestCount  int64
	totalDuration time.Duration
	slowRequests  int64
	mutex         sync.RWMutex
	slowThreshold time.Duration
}

func NewPerformanceMiddleware(slowThreshold time.Duration) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		slowThreshold: slowThreshold,
	}
}

func (pm *PerformanceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		pm.mutex.Lock()
		pm.requestCount++
		pm.totalDuration += duration
		if duration > pm.slowThreshold {
			pm.slowRequests++
		}
		pm.mutex.Unlock()

		// 记录慢请求
		if duration > pm.slowThreshold {
			log.Printf("慢请求: %s %s - %v", r.Method, r.URL.Path, duration)
		}
	})
}

func (pm *PerformanceMiddleware) GetStats() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var avgDuration time.Duration
	if pm.requestCount > 0 {
		avgDuration = pm.totalDuration / time.Duration(pm.requestCount)
	}

	return map[string]interface{}{
		"total_requests":  pm.requestCount,
		"slow_requests":   pm.slowRequests,
		"avg_duration":    avgDuration.String(),
		"slow_percentage": float64(pm.slowRequests) / float64(pm.requestCount) * 100,
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// === 缓存键生成器 ===

func generateCacheKey(prefix string, params ...interface{}) string {
	key := prefix
	for _, param := range params {
		key += fmt.Sprintf(":%v", param)
	}

	// 为了防止键过长，可以选择对键进行哈希
	if len(key) > 100 {
		hash := md5.Sum([]byte(key))
		return prefix + ":" + hex.EncodeToString(hash[:])
	}

	return key
}

// === 辅助函数 ===

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// === 缓存模式演示 ===

func demonstrateCachePatterns() {
	fmt.Println("=== 缓存模式演示 ===")

	fmt.Println("1. Cache-Aside (旁路缓存):")
	fmt.Println("   - 应用程序直接管理缓存")
	fmt.Println("   - 读取时先查缓存，未命中则查数据库并更新缓存")
	fmt.Println("   - 写入时直接写数据库，然后删除缓存")

	fmt.Println("2. Read-Through (透读缓存):")
	fmt.Println("   - 缓存层负责加载数据")
	fmt.Println("   - 应用程序只与缓存层交互")
	fmt.Println("   - 缓存未命中时自动从数据源加载")

	fmt.Println("3. Write-Through (透写缓存):")
	fmt.Println("   - 写入时同时更新缓存和数据库")
	fmt.Println("   - 保证数据一致性")
	fmt.Println("   - 写入延迟较高")

	fmt.Println("4. Write-Behind (回写缓存):")
	fmt.Println("   - 写入时只更新缓存")
	fmt.Println("   - 异步批量写入数据库")
	fmt.Println("   - 写入性能高，但有数据丢失风险")
}

func main() {
	// 创建LRU缓存
	lruCache := NewLRUCache(1000)

	// 创建Redis缓存（如果有Redis服务器）
	// redisCache := NewRedisCache("localhost:6379", "", 0)

	// 使用LRU缓存作为主缓存
	cache := lruCache

	// 创建多级缓存（如果有Redis）
	// cache = NewMultiLevelCache(lruCache, redisCache)

	// 创建应用服务
	service := NewAppService(cache)

	// 创建HTTP处理器
	handler := NewHTTPHandler(service, cache)

	// 创建性能监控中间件
	perfMiddleware := NewPerformanceMiddleware(100 * time.Millisecond)

	// 创建路由器
	router := mux.NewRouter()

	// 添加性能监控中间件
	router.Use(perfMiddleware.Middleware)

	// API路由
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	api.HandleFunc("/posts", handler.GetPosts).Methods("GET")
	api.HandleFunc("/posts/popular", handler.GetPopularPosts).Methods("GET")
	api.HandleFunc("/cache/stats", handler.GetCacheStats).Methods("GET")
	api.HandleFunc("/cache/manage", handler.ManageCache).Methods("POST")

	// 性能统计API
	router.HandleFunc("/api/performance", func(w http.ResponseWriter, r *http.Request) {
		stats := perfMiddleware.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET")

	// 演示缓存模式
	demonstrateCachePatterns()

	fmt.Println("=== 缓存策略服务器启动 ===")
	fmt.Println("API端点:")
	fmt.Println("  GET  /api/users/{id}     - 获取用户信息")
	fmt.Println("  GET  /api/posts          - 获取文章列表")
	fmt.Println("  GET  /api/posts/popular  - 获取热门文章")
	fmt.Println("  GET  /api/cache/stats    - 缓存统计")
	fmt.Println("  POST /api/cache/manage   - 缓存管理")
	fmt.Println("  GET  /api/performance    - 性能统计")
	fmt.Println()
	fmt.Println("缓存管理:")
	fmt.Println("  POST /api/cache/manage?action=clear  - 清空缓存")
	fmt.Println("  POST /api/cache/manage?action=warmup - 预热缓存")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  curl http://localhost:8080/api/users/1")
	fmt.Println("  curl http://localhost:8080/api/posts?page=1&limit=5")
	fmt.Println("  curl http://localhost:8080/api/cache/stats")
	fmt.Println()
	fmt.Println("服务器运行在 http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
练习任务：

1. 基础练习：
   - 实现TTL过期策略
   - 添加缓存命中率监控
   - 实现缓存大小限制
   - 添加缓存键命名空间

2. 中级练习：
   - 实现分布式缓存一致性
   - 添加缓存预热策略
   - 实现缓存降级机制
   - 添加缓存穿透防护

3. 高级练习：
   - 实现Redis Cluster集成
   - 添加缓存雪崩防护
   - 实现缓存更新策略
   - 集成Prometheus监控

4. 性能优化：
   - 实现并发缓存操作
   - 添加内存池优化
   - 实现缓存压缩
   - 优化序列化性能

5. 安全练习：
   - 实现缓存访问控制
   - 添加缓存数据加密
   - 实现缓存审计日志
   - 防止缓存投毒攻击

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/go-redis/redis/v8

2. 可选：启动Redis服务器
   docker run -d -p 6379:6379 redis:alpine

3. 运行程序：go run main.go

测试缓存效果：
1. 第一次请求（缓存未命中）：
   curl http://localhost:8080/api/users/1

2. 第二次请求（缓存命中）：
   curl http://localhost:8080/api/users/1

3. 查看缓存统计：
   curl http://localhost:8080/api/cache/stats

扩展建议：
- 集成监控系统（Grafana、Prometheus）
- 实现缓存数据同步机制
- 添加缓存A/B测试功能
- 实现智能缓存策略选择
*/
