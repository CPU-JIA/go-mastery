package main

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
)

// ==================
// Sorting Algorithm Benchmarks
// ==================

// generateTestData creates random test data for benchmarks
func generateTestData(size int) []int {
	data := make([]int, size)
	for i := range data {
		data[i] = rand.Intn(size * 10)
	}
	return data
}

// copyData creates a copy of the data slice
func copyData(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	return result
}

// BenchmarkStdlibSort benchmarks the standard library sort
func BenchmarkStdlibSort(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		data := generateTestData(size)
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testData := copyData(data)
				sort.Ints(testData)
			}
		})
	}
}

// BenchmarkQuickSort benchmarks quicksort implementation
func BenchmarkQuickSort(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		data := generateTestData(size)
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testData := copyData(data)
				quickSort(testData, 0, len(testData)-1)
			}
		})
	}
}

// BenchmarkMergeSort benchmarks mergesort implementation
func BenchmarkMergeSort(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		data := generateTestData(size)
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testData := copyData(data)
				mergeSort(testData)
			}
		})
	}
}

// BenchmarkRadixSort benchmarks radix sort implementation
func BenchmarkRadixSort(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		data := generateTestData(size)
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testData := copyData(data)
				radixSort(testData)
			}
		})
	}
}

// BenchmarkParallelQuickSort benchmarks parallel quicksort
func BenchmarkParallelQuickSort(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	numCPU := runtime.NumCPU()

	for _, size := range sizes {
		data := generateTestData(size)
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testData := copyData(data)
				parallelQuickSort(testData, 0, len(testData)-1, numCPU)
			}
		})
	}
}

func formatSize(size int) string {
	switch {
	case size >= 1000000:
		return string(rune('0'+size/1000000)) + "M"
	case size >= 1000:
		return string(rune('0'+size/1000)) + "K"
	default:
		return string(rune('0' + size))
	}
}

// ==================
// Memory Optimization Benchmarks
// ==================

// BenchmarkWithoutPool benchmarks allocation without object pool
func BenchmarkWithoutPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj := make([]byte, 1024)
		_ = obj
	}
}

// BenchmarkWithSyncPool benchmarks allocation with sync.Pool
func BenchmarkWithSyncPool(b *testing.B) {
	pool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj := pool.Get().([]byte)
		pool.Put(obj)
	}
}

// BenchmarkWithoutPool_Parallel benchmarks parallel allocation without pool
func BenchmarkWithoutPool_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := make([]byte, 1024)
			_ = obj
		}
	})
}

// BenchmarkWithSyncPool_Parallel benchmarks parallel allocation with sync.Pool
func BenchmarkWithSyncPool_Parallel(b *testing.B) {
	pool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get().([]byte)
			pool.Put(obj)
		}
	})
}

// ==================
// Slice Optimization Benchmarks
// ==================

// BenchmarkSliceAppend benchmarks slice append without preallocation
func BenchmarkSliceAppend(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s []int
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
	}
}

// BenchmarkSlicePrealloc benchmarks slice with preallocation
func BenchmarkSlicePrealloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 1000)
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
	}
}

// BenchmarkSliceDirect benchmarks direct slice assignment
func BenchmarkSliceDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]int, 1000)
		for j := 0; j < 1000; j++ {
			s[j] = j
		}
	}
}

// ==================
// String Concatenation Benchmarks
// ==================

// BenchmarkStringConcat benchmarks string concatenation with +
func BenchmarkStringConcat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 100; j++ {
			s += "a"
		}
		_ = s
	}
}

// BenchmarkStringBuilder benchmarks strings.Builder
func BenchmarkStringBuilder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sb stringBuilder
		for j := 0; j < 100; j++ {
			sb.WriteString("a")
		}
		_ = sb.String()
	}
}

// BenchmarkByteSliceConcat benchmarks byte slice concatenation
func BenchmarkByteSliceConcat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 100)
		for j := 0; j < 100; j++ {
			buf = append(buf, 'a')
		}
		_ = string(buf)
	}
}

// stringBuilder is a simple string builder for benchmarking
type stringBuilder struct {
	buf []byte
}

func (sb *stringBuilder) WriteString(s string) {
	sb.buf = append(sb.buf, s...)
}

func (sb *stringBuilder) String() string {
	return string(sb.buf)
}

// ==================
// Concurrency Optimization Benchmarks
// ==================

// BenchmarkMutex benchmarks sync.Mutex
func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	var counter int

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		counter++
		mu.Unlock()
	}
}

// BenchmarkRWMutex_Read benchmarks sync.RWMutex for reads
func BenchmarkRWMutex_Read(b *testing.B) {
	var mu sync.RWMutex
	counter := 0

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = counter
		mu.RUnlock()
	}
}

// BenchmarkRWMutex_Write benchmarks sync.RWMutex for writes
func BenchmarkRWMutex_Write(b *testing.B) {
	var mu sync.RWMutex
	counter := 0

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		counter++
		mu.Unlock()
	}
}

// BenchmarkAtomic benchmarks atomic operations
func BenchmarkAtomic(b *testing.B) {
	var counter int64

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&counter, 1)
	}
}

// BenchmarkMutex_Parallel benchmarks parallel mutex access
func BenchmarkMutex_Parallel(b *testing.B) {
	var mu sync.Mutex
	var counter int

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	})
}

// BenchmarkRWMutex_Parallel_Read benchmarks parallel RWMutex reads
func BenchmarkRWMutex_Parallel_Read(b *testing.B) {
	var mu sync.RWMutex
	counter := 0

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			_ = counter
			mu.RUnlock()
		}
	})
}

// BenchmarkAtomic_Parallel benchmarks parallel atomic operations
func BenchmarkAtomic_Parallel(b *testing.B) {
	var counter int64

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&counter, 1)
		}
	})
}

// ==================
// Channel Benchmarks
// ==================

// BenchmarkUnbufferedChannel benchmarks unbuffered channel
func BenchmarkUnbufferedChannel(b *testing.B) {
	ch := make(chan int)

	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

// BenchmarkBufferedChannel benchmarks buffered channel
func BenchmarkBufferedChannel(b *testing.B) {
	ch := make(chan int, 100)

	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

// BenchmarkBufferedChannel_Large benchmarks large buffered channel
func BenchmarkBufferedChannel_Large(b *testing.B) {
	ch := make(chan int, 1000)

	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

// ==================
// Map Benchmarks
// ==================

// BenchmarkMapRead benchmarks map read operations
func BenchmarkMapRead(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < 1000; i++ {
		m[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%1000]
	}
}

// BenchmarkMapWrite benchmarks map write operations
func BenchmarkMapWrite(b *testing.B) {
	m := make(map[int]int)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m[i%1000] = i
	}
}

// BenchmarkSyncMapRead benchmarks sync.Map read operations
func BenchmarkSyncMapRead(b *testing.B) {
	var m sync.Map
	for i := 0; i < 1000; i++ {
		m.Store(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Load(i % 1000)
	}
}

// BenchmarkSyncMapWrite benchmarks sync.Map write operations
func BenchmarkSyncMapWrite(b *testing.B) {
	var m sync.Map

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.Store(i%1000, i)
	}
}

// BenchmarkMapRead_Parallel benchmarks parallel map reads with mutex
func BenchmarkMapRead_Parallel(b *testing.B) {
	m := make(map[int]int)
	var mu sync.RWMutex
	for i := 0; i < 1000; i++ {
		m[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.RLock()
			_ = m[i%1000]
			mu.RUnlock()
			i++
		}
	})
}

// BenchmarkSyncMapRead_Parallel benchmarks parallel sync.Map reads
func BenchmarkSyncMapRead_Parallel(b *testing.B) {
	var m sync.Map
	for i := 0; i < 1000; i++ {
		m.Store(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % 1000)
			i++
		}
	})
}

// ==================
// Interface vs Concrete Type Benchmarks
// ==================

type adder interface {
	Add(a, b int) int
}

type concreteAdder struct{}

func (c concreteAdder) Add(a, b int) int {
	return a + b
}

// BenchmarkInterfaceCall benchmarks interface method calls
func BenchmarkInterfaceCall(b *testing.B) {
	var a adder = concreteAdder{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = a.Add(1, 2)
	}
}

// BenchmarkConcreteCall benchmarks concrete type method calls
func BenchmarkConcreteCall(b *testing.B) {
	a := concreteAdder{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = a.Add(1, 2)
	}
}

// BenchmarkFunctionCall benchmarks direct function calls
func BenchmarkFunctionCall(b *testing.B) {
	add := func(a, b int) int {
		return a + b
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = add(1, 2)
	}
}

// ==================
// Struct Layout Benchmarks
// ==================

// BadLayout has poor memory alignment
type BadLayout struct {
	a bool  // 1 byte
	b int64 // 8 bytes (7 bytes padding before)
	c bool  // 1 byte
	d int64 // 8 bytes (7 bytes padding before)
	e bool  // 1 byte
	f int64 // 8 bytes (7 bytes padding before)
}

// GoodLayout has optimal memory alignment
type GoodLayout struct {
	b int64 // 8 bytes
	d int64 // 8 bytes
	f int64 // 8 bytes
	a bool  // 1 byte
	c bool  // 1 byte
	e bool  // 1 byte (5 bytes padding at end)
}

// BenchmarkBadLayout benchmarks struct with poor alignment
func BenchmarkBadLayout(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := BadLayout{
			a: true,
			b: int64(i),
			c: true,
			d: int64(i),
			e: true,
			f: int64(i),
		}
		_ = s
	}
}

// BenchmarkGoodLayout benchmarks struct with optimal alignment
func BenchmarkGoodLayout(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := GoodLayout{
			a: true,
			b: int64(i),
			c: true,
			d: int64(i),
			e: true,
			f: int64(i),
		}
		_ = s
	}
}

// BenchmarkBadLayoutSlice benchmarks slice of poorly aligned structs
func BenchmarkBadLayoutSlice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]BadLayout, 1000)
		for j := range s {
			s[j].b = int64(j)
		}
	}
}

// BenchmarkGoodLayoutSlice benchmarks slice of well-aligned structs
func BenchmarkGoodLayoutSlice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]GoodLayout, 1000)
		for j := range s {
			s[j].b = int64(j)
		}
	}
}

// ==================
// Loop Optimization Benchmarks
// ==================

// BenchmarkRangeLoop benchmarks range loop
func BenchmarkRangeLoop(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range data {
			sum += v
		}
		_ = sum
	}
}

// BenchmarkIndexLoop benchmarks index-based loop
func BenchmarkIndexLoop(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < len(data); j++ {
			sum += data[j]
		}
		_ = sum
	}
}

// BenchmarkIndexLoopCachedLen benchmarks index loop with cached length
func BenchmarkIndexLoopCachedLen(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		n := len(data)
		for j := 0; j < n; j++ {
			sum += data[j]
		}
		_ = sum
	}
}

// ==================
// Defer Benchmarks
// ==================

// BenchmarkWithDefer benchmarks function with defer
func BenchmarkWithDefer(b *testing.B) {
	var mu sync.Mutex

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() {
			mu.Lock()
			defer mu.Unlock()
			_ = i
		}()
	}
}

// BenchmarkWithoutDefer benchmarks function without defer
func BenchmarkWithoutDefer(b *testing.B) {
	var mu sync.Mutex

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() {
			mu.Lock()
			_ = i
			mu.Unlock()
		}()
	}
}

// ==================
// Goroutine Pool Benchmarks
// ==================

// BenchmarkGoroutineCreation benchmarks creating new goroutines
func BenchmarkGoroutineCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		<-done
	}
}

// WorkerPool is a simple worker pool for benchmarking
type WorkerPool struct {
	tasks   chan func()
	workers int
}

func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		tasks:   make(chan func(), workers*2),
		workers: workers,
	}
	for i := 0; i < workers; i++ {
		go wp.worker()
	}
	return wp
}

func (wp *WorkerPool) worker() {
	for task := range wp.tasks {
		task()
	}
}

func (wp *WorkerPool) Submit(task func()) {
	wp.tasks <- task
}

func (wp *WorkerPool) Close() {
	close(wp.tasks)
}

// BenchmarkWorkerPool benchmarks using a worker pool
func BenchmarkWorkerPool(b *testing.B) {
	pool := NewWorkerPool(runtime.NumCPU())
	defer pool.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		done := make(chan struct{})
		pool.Submit(func() {
			close(done)
		})
		<-done
	}
}
