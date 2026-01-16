package examples

import (
	"sync"
	"testing"
)

// ==================
// String Concatenation Benchmarks
// ==================

func BenchmarkBeforeStringConcat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BeforeStringConcat(100)
	}
}

func BenchmarkAfterStringConcat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterStringConcat(100)
	}
}

func BenchmarkBeforeStringConcat_Large(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BeforeStringConcat(1000)
	}
}

func BenchmarkAfterStringConcat_Large(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterStringConcat(1000)
	}
}

// ==================
// Slice Preallocation Benchmarks
// ==================

func BenchmarkBeforeSliceAppend(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BeforeSliceAppend(1000)
	}
}

func BenchmarkAfterSliceAppend(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterSliceAppend(1000)
	}
}

func BenchmarkAfterSliceDirect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterSliceDirect(1000)
	}
}

// ==================
// Object Pool Benchmarks
// ==================

func BenchmarkBeforeObjectPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		BeforeObjectPool(10)
	}
}

func BenchmarkAfterObjectPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AfterObjectPool(10)
	}
}

// ==================
// Map vs Slice Lookup Benchmarks
// ==================

func BenchmarkBeforeMapLookup_Small(b *testing.B) {
	keys := []string{"a", "b", "c", "d", "e"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BeforeMapLookup(keys, "c")
	}
}

func BenchmarkAfterSliceLookup_Small(b *testing.B) {
	keys := []string{"a", "b", "c", "d", "e"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterSliceLookup(keys, "c")
	}
}

func BenchmarkBeforeMapLookup_Large(b *testing.B) {
	keys := make([]string, 100)
	for i := range keys {
		keys[i] = string(rune('a' + i%26))
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BeforeMapLookup(keys, "m")
	}
}

func BenchmarkAfterSliceLookup_Large(b *testing.B) {
	keys := make([]string, 100)
	for i := range keys {
		keys[i] = string(rune('a' + i%26))
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = AfterSliceLookup(keys, "m")
	}
}

// ==================
// Mutex vs RWMutex Benchmarks
// ==================

func BenchmarkBeforeMutexReadHeavy(b *testing.B) {
	m := &BeforeMutexReadHeavy{value: 42}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Read()
	}
}

func BenchmarkAfterRWMutexReadHeavy(b *testing.B) {
	m := &AfterRWMutexReadHeavy{value: 42}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Read()
	}
}

func BenchmarkBeforeMutexReadHeavy_Parallel(b *testing.B) {
	m := &BeforeMutexReadHeavy{value: 42}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Read()
		}
	})
}

func BenchmarkAfterRWMutexReadHeavy_Parallel(b *testing.B) {
	m := &AfterRWMutexReadHeavy{value: 42}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Read()
		}
	})
}

// ==================
// Atomic vs Mutex Counter Benchmarks
// ==================

func BenchmarkBeforeMutexCounter(b *testing.B) {
	c := &BeforeMutexCounter{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkAfterAtomicCounter(b *testing.B) {
	c := &AfterAtomicCounter{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkBeforeMutexCounter_Parallel(b *testing.B) {
	c := &BeforeMutexCounter{}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

func BenchmarkAfterAtomicCounter_Parallel(b *testing.B) {
	c := &AfterAtomicCounter{}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

// ==================
// Struct Alignment Benchmarks
// ==================

func BenchmarkBeforeBadAlignment(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := BeforeBadAlignment{
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

func BenchmarkAfterGoodAlignment(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := AfterGoodAlignment{
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

func BenchmarkBeforeBadAlignment_Slice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]BeforeBadAlignment, 1000)
		for j := range s {
			s[j].b = int64(j)
		}
	}
}

func BenchmarkAfterGoodAlignment_Slice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]AfterGoodAlignment, 1000)
		for j := range s {
			s[j].b = int64(j)
		}
	}
}

// ==================
// Channel Buffer Benchmarks
// ==================

func BenchmarkBeforeUnbufferedChannel(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		BeforeUnbufferedChannel(100)
	}
}

func BenchmarkAfterBufferedChannel(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AfterBufferedChannel(100)
	}
}

// ==================
// Bytes Buffer Benchmarks
// ==================

func BenchmarkBeforeBytesConcat(b *testing.B) {
	parts := make([][]byte, 100)
	for i := range parts {
		parts[i] = []byte("hello world ")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BeforeBytesConcat(parts)
	}
}

func BenchmarkAfterBytesBuffer(b *testing.B) {
	parts := make([][]byte, 100)
	for i := range parts {
		parts[i] = []byte("hello world ")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AfterBytesBuffer(parts)
	}
}

// ==================
// Interface vs Concrete Type Benchmarks
// ==================

func BenchmarkBeforeInterfaceSlice(b *testing.B) {
	processors := make([]Processor, 10)
	for i := range processors {
		processors[i] = &ConcreteProcessor{multiplier: i + 1}
	}
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BeforeInterfaceSlice(processors, data)
	}
}

func BenchmarkAfterConcreteSlice(b *testing.B) {
	processors := make([]*ConcreteProcessor, 10)
	for i := range processors {
		processors[i] = &ConcreteProcessor{multiplier: i + 1}
	}
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AfterConcreteSlice(processors, data)
	}
}

// ==================
// Sort Benchmarks
// ==================

func BenchmarkBeforeSortInterface(b *testing.B) {
	original := make([]int, 1000)
	for i := range original {
		original[i] = 1000 - i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := make([]int, len(original))
		copy(data, original)
		BeforeSortInterface(data)
	}
}

func BenchmarkAfterSortSlice(b *testing.B) {
	original := make([]int, 1000)
	for i := range original {
		original[i] = 1000 - i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := make([]int, len(original))
		copy(data, original)
		AfterSortSlice(data)
	}
}

// ==================
// Unit Tests
// ==================

func TestStringConcat(t *testing.T) {
	n := 100
	before := BeforeStringConcat(n)
	after := AfterStringConcat(n)

	if len(before) != n {
		t.Errorf("BeforeStringConcat: expected length %d, got %d", n, len(before))
	}
	if len(after) != n {
		t.Errorf("AfterStringConcat: expected length %d, got %d", n, len(after))
	}
	if before != after {
		t.Error("Results should be equal")
	}
}

func TestSliceAppend(t *testing.T) {
	n := 100
	before := BeforeSliceAppend(n)
	afterAppend := AfterSliceAppend(n)
	afterDirect := AfterSliceDirect(n)

	if len(before) != n {
		t.Errorf("BeforeSliceAppend: expected length %d, got %d", n, len(before))
	}
	if len(afterAppend) != n {
		t.Errorf("AfterSliceAppend: expected length %d, got %d", n, len(afterAppend))
	}
	if len(afterDirect) != n {
		t.Errorf("AfterSliceDirect: expected length %d, got %d", n, len(afterDirect))
	}

	for i := 0; i < n; i++ {
		if before[i] != i || afterAppend[i] != i || afterDirect[i] != i {
			t.Errorf("Values at index %d don't match", i)
		}
	}
}

func TestMapVsSliceLookup(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}

	// Test found
	if !BeforeMapLookup(keys, "c") {
		t.Error("BeforeMapLookup should find 'c'")
	}
	if !AfterSliceLookup(keys, "c") {
		t.Error("AfterSliceLookup should find 'c'")
	}

	// Test not found
	if BeforeMapLookup(keys, "z") {
		t.Error("BeforeMapLookup should not find 'z'")
	}
	if AfterSliceLookup(keys, "z") {
		t.Error("AfterSliceLookup should not find 'z'")
	}
}

func TestMutexVsRWMutex(t *testing.T) {
	before := &BeforeMutexReadHeavy{}
	after := &AfterRWMutexReadHeavy{}

	before.Write(42)
	after.Write(42)

	if before.Read() != 42 {
		t.Error("BeforeMutexReadHeavy: expected 42")
	}
	if after.Read() != 42 {
		t.Error("AfterRWMutexReadHeavy: expected 42")
	}
}

func TestAtomicVsMutexCounter(t *testing.T) {
	before := &BeforeMutexCounter{}
	after := &AfterAtomicCounter{}

	for i := 0; i < 100; i++ {
		before.Inc()
		after.Inc()
	}

	if before.Value() != 100 {
		t.Errorf("BeforeMutexCounter: expected 100, got %d", before.Value())
	}
	if after.Value() != 100 {
		t.Errorf("AfterAtomicCounter: expected 100, got %d", after.Value())
	}
}

func TestAtomicVsMutexCounter_Concurrent(t *testing.T) {
	before := &BeforeMutexCounter{}
	after := &AfterAtomicCounter{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			before.Inc()
		}()
		go func() {
			defer wg.Done()
			after.Inc()
		}()
	}
	wg.Wait()

	if before.Value() != 100 {
		t.Errorf("BeforeMutexCounter: expected 100, got %d", before.Value())
	}
	if after.Value() != 100 {
		t.Errorf("AfterAtomicCounter: expected 100, got %d", after.Value())
	}
}

func TestBytesConcat(t *testing.T) {
	parts := [][]byte{[]byte("hello"), []byte(" "), []byte("world")}

	before := BeforeBytesConcat(parts)
	after := AfterBytesBuffer(parts)

	expected := "hello world"
	if string(before) != expected {
		t.Errorf("BeforeBytesConcat: expected %q, got %q", expected, string(before))
	}
	if string(after) != expected {
		t.Errorf("AfterBytesBuffer: expected %q, got %q", expected, string(after))
	}
}

func TestProcessor(t *testing.T) {
	processors := []Processor{&ConcreteProcessor{multiplier: 2}}
	concreteProcessors := []*ConcreteProcessor{{multiplier: 2}}
	data := []int{1, 2, 3, 4, 5}

	before := BeforeInterfaceSlice(processors, data)
	after := AfterConcreteSlice(concreteProcessors, data)

	for i := range data {
		expected := data[i] * 2
		if before[i] != expected {
			t.Errorf("BeforeInterfaceSlice[%d]: expected %d, got %d", i, expected, before[i])
		}
		if after[i] != expected {
			t.Errorf("AfterConcreteSlice[%d]: expected %d, got %d", i, expected, after[i])
		}
	}
}

func TestSort(t *testing.T) {
	data1 := []int{5, 3, 1, 4, 2}
	data2 := []int{5, 3, 1, 4, 2}

	BeforeSortInterface(data1)
	AfterSortSlice(data2)

	for i := 0; i < len(data1); i++ {
		if data1[i] != i+1 {
			t.Errorf("BeforeSortInterface[%d]: expected %d, got %d", i, i+1, data1[i])
		}
		if data2[i] != i+1 {
			t.Errorf("AfterSortSlice[%d]: expected %d, got %d", i, i+1, data2[i])
		}
	}
}
