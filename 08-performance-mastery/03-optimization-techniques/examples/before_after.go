/*
Package examples provides before/after optimization examples with measurable results.

This file demonstrates common Go performance optimizations with clear comparisons
showing the impact of each optimization technique.
*/
package examples

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// Example 1: String Concatenation
// ==================

// BeforeStringConcat demonstrates inefficient string concatenation
// Problem: Each += creates a new string, causing O(n^2) allocations
func BeforeStringConcat(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

// AfterStringConcat demonstrates efficient string building
// Solution: Use strings.Builder for O(n) performance
func AfterStringConcat(n int) string {
	var sb strings.Builder
	sb.Grow(n) // Pre-allocate capacity
	for i := 0; i < n; i++ {
		sb.WriteString("a")
	}
	return sb.String()
}

// ==================
// Example 2: Slice Preallocation
// ==================

// BeforeSliceAppend demonstrates slice growth without preallocation
// Problem: Multiple reallocations as slice grows
func BeforeSliceAppend(n int) []int {
	var result []int
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// AfterSliceAppend demonstrates slice with preallocation
// Solution: Preallocate with make([]T, 0, capacity)
func AfterSliceAppend(n int) []int {
	result := make([]int, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// AfterSliceDirect demonstrates direct assignment (fastest)
// Solution: Use direct indexing when size is known
func AfterSliceDirect(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

// ==================
// Example 3: Object Pool
// ==================

// Buffer represents a reusable buffer
type Buffer struct {
	data []byte
}

// BeforeObjectPool demonstrates allocation without pooling
// Problem: Creates new allocation for each request
func BeforeObjectPool(iterations int) {
	for i := 0; i < iterations; i++ {
		buf := &Buffer{data: make([]byte, 4096)}
		// Simulate work
		for j := range buf.data {
			buf.data[j] = byte(j % 256)
		}
		// Buffer is discarded, causing GC pressure
	}
}

// AfterObjectPool demonstrates allocation with sync.Pool
// Solution: Reuse objects via sync.Pool
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{data: make([]byte, 4096)}
	},
}

func AfterObjectPool(iterations int) {
	for i := 0; i < iterations; i++ {
		buf := bufferPool.Get().(*Buffer)
		// Simulate work
		for j := range buf.data {
			buf.data[j] = byte(j % 256)
		}
		// Clear and return to pool
		for j := range buf.data {
			buf.data[j] = 0
		}
		bufferPool.Put(buf)
	}
}

// ==================
// Example 4: Map vs Slice for Small Collections
// ==================

// BeforeMapLookup uses map for small collection lookup
// Problem: Map has overhead for small collections
func BeforeMapLookup(keys []string, target string) bool {
	m := make(map[string]bool)
	for _, k := range keys {
		m[k] = true
	}
	return m[target]
}

// AfterSliceLookup uses slice for small collection lookup
// Solution: Linear search is faster for small collections (< ~10 items)
func AfterSliceLookup(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}

// ==================
// Example 5: Mutex vs RWMutex
// ==================

// BeforeMutexReadHeavy uses Mutex for read-heavy workload
// Problem: Mutex blocks all readers
type BeforeMutexReadHeavy struct {
	mu    sync.Mutex
	value int
}

func (b *BeforeMutexReadHeavy) Read() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.value
}

func (b *BeforeMutexReadHeavy) Write(v int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.value = v
}

// AfterRWMutexReadHeavy uses RWMutex for read-heavy workload
// Solution: RWMutex allows concurrent reads
type AfterRWMutexReadHeavy struct {
	mu    sync.RWMutex
	value int
}

func (a *AfterRWMutexReadHeavy) Read() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.value
}

func (a *AfterRWMutexReadHeavy) Write(v int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.value = v
}

// ==================
// Example 6: Atomic vs Mutex for Counters
// ==================

// BeforeMutexCounter uses Mutex for counter
// Problem: Mutex has higher overhead for simple operations
type BeforeMutexCounter struct {
	mu    sync.Mutex
	count int64
}

func (b *BeforeMutexCounter) Inc() {
	b.mu.Lock()
	b.count++
	b.mu.Unlock()
}

func (b *BeforeMutexCounter) Value() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}

// AfterAtomicCounter uses atomic operations
// Solution: Atomic operations are lock-free and faster
type AfterAtomicCounter struct {
	count int64
}

func (a *AfterAtomicCounter) Inc() {
	atomic.AddInt64(&a.count, 1)
}

func (a *AfterAtomicCounter) Value() int64 {
	return atomic.LoadInt64(&a.count)
}

// ==================
// Example 7: Struct Field Ordering
// ==================

// BeforeBadAlignment has poor memory alignment
// Problem: Padding wastes memory and hurts cache performance
// Size: 48 bytes (with padding)
type BeforeBadAlignment struct {
	a bool  // 1 byte + 7 padding
	b int64 // 8 bytes
	c bool  // 1 byte + 7 padding
	d int64 // 8 bytes
	e bool  // 1 byte + 7 padding
	f int64 // 8 bytes
}

// AfterGoodAlignment has optimal memory alignment
// Solution: Order fields from largest to smallest
// Size: 32 bytes (minimal padding)
type AfterGoodAlignment struct {
	b int64 // 8 bytes
	d int64 // 8 bytes
	f int64 // 8 bytes
	a bool  // 1 byte
	c bool  // 1 byte
	e bool  // 1 byte + 5 padding
}

// ==================
// Example 8: Channel Buffer Size
// ==================

// BeforeUnbufferedChannel uses unbuffered channel
// Problem: Sender blocks until receiver is ready
func BeforeUnbufferedChannel(n int) {
	ch := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			<-ch
		}
	}()

	for i := 0; i < n; i++ {
		ch <- i
	}
	wg.Wait()
}

// AfterBufferedChannel uses appropriately buffered channel
// Solution: Buffer reduces synchronization overhead
func AfterBufferedChannel(n int) {
	ch := make(chan int, 100) // Buffer size based on workload
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			<-ch
		}
	}()

	for i := 0; i < n; i++ {
		ch <- i
	}
	wg.Wait()
}

// ==================
// Example 9: Bytes Buffer vs String Concatenation
// ==================

// BeforeBytesConcat demonstrates inefficient byte concatenation
func BeforeBytesConcat(parts [][]byte) []byte {
	var result []byte
	for _, part := range parts {
		result = append(result, part...)
	}
	return result
}

// AfterBytesBuffer demonstrates efficient byte concatenation
// Solution: Use bytes.Buffer with preallocation
func AfterBytesBuffer(parts [][]byte) []byte {
	// Calculate total size
	totalSize := 0
	for _, part := range parts {
		totalSize += len(part)
	}

	// Preallocate buffer
	var buf bytes.Buffer
	buf.Grow(totalSize)

	for _, part := range parts {
		buf.Write(part)
	}
	return buf.Bytes()
}

// ==================
// Example 10: Interface vs Concrete Type
// ==================

// Processor interface for demonstration
type Processor interface {
	Process(data int) int
}

// ConcreteProcessor implements Processor
type ConcreteProcessor struct {
	multiplier int
}

func (c *ConcreteProcessor) Process(data int) int {
	return data * c.multiplier
}

// BeforeInterfaceSlice uses interface slice
// Problem: Interface indirection has overhead
func BeforeInterfaceSlice(processors []Processor, data []int) []int {
	result := make([]int, len(data))
	for i, d := range data {
		result[i] = processors[i%len(processors)].Process(d)
	}
	return result
}

// AfterConcreteSlice uses concrete type slice
// Solution: Use concrete types when possible
func AfterConcreteSlice(processors []*ConcreteProcessor, data []int) []int {
	result := make([]int, len(data))
	for i, d := range data {
		result[i] = processors[i%len(processors)].Process(d)
	}
	return result
}

// ==================
// Example 11: Sort Optimization
// ==================

// BeforeSortInterface uses sort.Interface
// Problem: Interface method calls have overhead
type IntSlice []int

func (s IntSlice) Len() int           { return len(s) }
func (s IntSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s IntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func BeforeSortInterface(data []int) {
	sort.Sort(IntSlice(data))
}

// AfterSortSlice uses sort.Ints (optimized)
// Solution: Use type-specific sort functions
func AfterSortSlice(data []int) {
	sort.Ints(data)
}

// ==================
// Measurement Utilities
// ==================

// MeasureResult holds measurement results
type MeasureResult struct {
	Name        string
	Duration    time.Duration
	Allocations uint64
	BytesAlloc  uint64
}

// Measure measures the performance of a function
func Measure(name string, iterations int, fn func()) MeasureResult {
	// Warmup
	for i := 0; i < 10; i++ {
		fn()
	}

	// Force GC before measurement
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		fn()
	}
	duration := time.Since(start)

	runtime.ReadMemStats(&m2)

	return MeasureResult{
		Name:        name,
		Duration:    duration / time.Duration(iterations),
		Allocations: (m2.Mallocs - m1.Mallocs) / uint64(iterations),
		BytesAlloc:  (m2.TotalAlloc - m1.TotalAlloc) / uint64(iterations),
	}
}

// CompareResults compares before and after results
func CompareResults(before, after MeasureResult) {
	speedup := float64(before.Duration) / float64(after.Duration)
	memSaving := float64(before.BytesAlloc) / float64(after.BytesAlloc)

	fmt.Printf("\n=== %s vs %s ===\n", before.Name, after.Name)
	fmt.Printf("Before: %v/op, %d allocs, %d bytes\n",
		before.Duration, before.Allocations, before.BytesAlloc)
	fmt.Printf("After:  %v/op, %d allocs, %d bytes\n",
		after.Duration, after.Allocations, after.BytesAlloc)
	fmt.Printf("Speedup: %.2fx faster\n", speedup)
	if after.BytesAlloc > 0 {
		fmt.Printf("Memory:  %.2fx less allocations\n", memSaving)
	}
}

// RunAllComparisons runs all before/after comparisons
func RunAllComparisons() {
	iterations := 10000

	// Example 1: String Concatenation
	n := 1000
	before := Measure("String +=", iterations, func() {
		_ = BeforeStringConcat(n)
	})
	after := Measure("strings.Builder", iterations, func() {
		_ = AfterStringConcat(n)
	})
	CompareResults(before, after)

	// Example 2: Slice Preallocation
	before = Measure("Slice append", iterations, func() {
		_ = BeforeSliceAppend(n)
	})
	after = Measure("Slice preallocated", iterations, func() {
		_ = AfterSliceAppend(n)
	})
	CompareResults(before, after)

	afterDirect := Measure("Slice direct", iterations, func() {
		_ = AfterSliceDirect(n)
	})
	CompareResults(before, afterDirect)

	// Example 3: Object Pool
	poolIterations := 100
	before = Measure("No pool", iterations/10, func() {
		BeforeObjectPool(poolIterations)
	})
	after = Measure("sync.Pool", iterations/10, func() {
		AfterObjectPool(poolIterations)
	})
	CompareResults(before, after)

	// Example 4: Atomic vs Mutex Counter
	counterIterations := 10000
	mutexCounter := &BeforeMutexCounter{}
	atomicCounter := &AfterAtomicCounter{}

	before = Measure("Mutex counter", iterations, func() {
		for i := 0; i < counterIterations/iterations; i++ {
			mutexCounter.Inc()
		}
	})
	after = Measure("Atomic counter", iterations, func() {
		for i := 0; i < counterIterations/iterations; i++ {
			atomicCounter.Inc()
		}
	})
	CompareResults(before, after)

	// Example 5: Channel Buffer
	channelN := 1000
	before = Measure("Unbuffered channel", iterations/100, func() {
		BeforeUnbufferedChannel(channelN)
	})
	after = Measure("Buffered channel", iterations/100, func() {
		AfterBufferedChannel(channelN)
	})
	CompareResults(before, after)

	fmt.Println("\n=== All comparisons complete ===")
}
