/*
=== Go 1.24特性模块测试文件 ===

这个测试文件展示了现代Go测试的最佳实践，包括：
1. 单元测试 - 测试个别函数和方法
2. 基准测试 - 性能测试和比较
3. 示例测试 - 文档化的代码示例
4. 表格驱动测试 - 数据驱动的测试方法
5. 模拟测试 - 依赖注入和接口测试
6. 集成测试 - 组件间集成测试
7. 并发测试 - 并发安全性测试
8. 性能回归测试 - 防止性能退化

测试覆盖目标: >90%
*/

package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ==================
// 1. 泛型类型别名测试
// ==================

func TestGenericTypeAliases(t *testing.T) {
	t.Run("GenericMap", func(t *testing.T) {
		// 测试泛型映射别名
		userMap := make(GenericMap[string, int])
		userMap["alice"] = 25
		userMap["bob"] = 30

		if len(userMap) != 2 {
			t.Errorf("期望映射长度为2，实际为%d", len(userMap))
		}

		if age, ok := userMap["alice"]; !ok || age != 25 {
			t.Errorf("期望alice的年龄为25，实际为%d", age)
		}
	})

	t.Run("GenericSlice", func(t *testing.T) {
		// 测试泛型切片别名
		numbers := GenericSlice[int]{1, 2, 3, 4, 5}

		if len(numbers) != 5 {
			t.Errorf("期望切片长度为5，实际为%d", len(numbers))
		}

		for i, num := range numbers {
			if num != i+1 {
				t.Errorf("期望numbers[%d]为%d，实际为%d", i, i+1, num)
			}
		}
	})

	t.Run("GenericChannel", func(t *testing.T) {
		// 测试泛型通道别名
		ch := make(GenericChannel[string], 2)

		go func() {
			ch <- "hello"
			ch <- "world"
			close(ch)
		}()

		messages := make([]string, 0)
		for msg := range ch {
			messages = append(messages, msg)
		}

		expected := []string{"hello", "world"}
		if !reflect.DeepEqual(messages, expected) {
			t.Errorf("期望消息为%v，实际为%v", expected, messages)
		}
	})

	t.Run("KeyValuePair", func(t *testing.T) {
		// 测试键值对泛型结构
		pair := KeyValuePair[string, int]{
			Key:   "age",
			Value: 25,
		}

		if pair.Key != "age" {
			t.Errorf("期望键为'age'，实际为'%s'", pair.Key)
		}

		if pair.Value != 25 {
			t.Errorf("期望值为25，实际为%d", pair.Value)
		}
	})
}

func TestGenericResult(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		err     error
		wantErr bool
	}{
		{
			name:    "成功情况",
			data:    "success",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "错误情况",
			data:    "",
			err:     fmt.Errorf("test error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericResult[string, error]{
				Data:  tt.data,
				Error: tt.err,
			}

			if (result.Error != nil) != tt.wantErr {
				t.Errorf("GenericResult.Error = %v, wantErr %v", result.Error, tt.wantErr)
			}

			if result.Data != tt.data {
				t.Errorf("GenericResult.Data = %v, want %v", result.Data, tt.data)
			}
		})
	}
}

func TestGenericOptional(t *testing.T) {
	t.Run("有效值", func(t *testing.T) {
		optional := GenericOptional[string]{
			Value: "test",
			Valid: true,
		}

		if !optional.Valid {
			t.Error("期望Valid为true")
		}

		if optional.Value != "test" {
			t.Errorf("期望Value为'test'，实际为'%s'", optional.Value)
		}
	})

	t.Run("无效值", func(t *testing.T) {
		optional := GenericOptional[string]{
			Value: "",
			Valid: false,
		}

		if optional.Valid {
			t.Error("期望Valid为false")
		}
	})
}

func TestGenericFunctionTypes(t *testing.T) {
	t.Run("GenericMapper", func(t *testing.T) {
		var mapper GenericMapper[int, string] = func(i int) string {
			return fmt.Sprintf("数字_%d", i)
		}

		result := mapper(42)
		expected := "数字_42"

		if result != expected {
			t.Errorf("期望映射结果为'%s'，实际为'%s'", expected, result)
		}
	})

	t.Run("GenericPredicate", func(t *testing.T) {
		var isEven GenericPredicate[int] = func(i int) bool {
			return i%2 == 0
		}

		if !isEven(4) {
			t.Error("期望4为偶数")
		}

		if isEven(3) {
			t.Error("期望3不为偶数")
		}
	})

	t.Run("GenericComparator", func(t *testing.T) {
		var intComparator GenericComparator[int] = func(a, b int) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		}

		if intComparator(1, 2) != -1 {
			t.Error("期望1小于2")
		}

		if intComparator(2, 1) != 1 {
			t.Error("期望2大于1")
		}

		if intComparator(2, 2) != 0 {
			t.Error("期望2等于2")
		}
	})
}

// ==================
// 2. 工具依赖管理测试
// ==================

func TestToolDependencyManager(t *testing.T) {
	t.Run("创建管理器", func(t *testing.T) {
		manager := NewToolDependencyManager("test.mod")

		if manager == nil {
			t.Error("期望创建工具依赖管理器成功")
		}

		if manager.goModPath != "test.mod" {
			t.Errorf("期望goModPath为'test.mod'，实际为'%s'", manager.goModPath)
		}
	})

	t.Run("添加工具", func(t *testing.T) {
		manager := NewToolDependencyManager("test.mod")

		manager.AddTool("golangci-lint", "v1.55.2",
			"github.com/golangci/golangci-lint", "Go代码静态分析工具")

		if len(manager.tools) != 1 {
			t.Errorf("期望工具数量为1，实际为%d", len(manager.tools))
		}

		tool, exists := manager.tools["golangci-lint"]
		if !exists {
			t.Error("期望找到golangci-lint工具")
		}

		if tool.Version != "v1.55.2" {
			t.Errorf("期望工具版本为'v1.55.2'，实际为'%s'", tool.Version)
		}
	})

	t.Run("使用工具", func(t *testing.T) {
		manager := NewToolDependencyManager("test.mod")

		manager.AddTool("test-tool", "v1.0.0", "test/path", "测试工具")

		// 第一次使用
		err := manager.UseTool("test-tool")
		if err != nil {
			t.Errorf("期望使用工具成功，实际错误: %v", err)
		}

		tool := manager.tools["test-tool"]
		if tool.UsageCount != 1 {
			t.Errorf("期望使用次数为1，实际为%d", tool.UsageCount)
		}

		// 第二次使用
		err = manager.UseTool("test-tool")
		if err != nil {
			t.Errorf("期望使用工具成功，实际错误: %v", err)
		}

		tool = manager.tools["test-tool"]
		if tool.UsageCount != 2 {
			t.Errorf("期望使用次数为2，实际为%d", tool.UsageCount)
		}
	})

	t.Run("使用不存在的工具", func(t *testing.T) {
		manager := NewToolDependencyManager("test.mod")

		err := manager.UseTool("nonexistent")
		if err == nil {
			t.Error("期望使用不存在的工具返回错误")
		}

		expectedError := "工具 'nonexistent' 未找到"
		if err.Error() != expectedError {
			t.Errorf("期望错误消息为'%s'，实际为'%s'", expectedError, err.Error())
		}
	})
}

func TestGenerateModernGoMod(t *testing.T) {
	goModContent := generateModernGoMod()

	// 检查是否包含必要的部分
	requiredSections := []string{
		"module github.com/example/modern-go-project",
		"go 1.24",
		"require (",
		"tool (",
		"github.com/golangci/golangci-lint v1.55.2",
	}

	for _, section := range requiredSections {
		if !containsString(goModContent, section) {
			t.Errorf("go.mod内容应该包含: %s", section)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================
// 3. 瑞士表映射测试
// ==================

func TestSwissTableAnalyzer(t *testing.T) {
	t.Run("创建分析器", func(t *testing.T) {
		analyzer := NewSwissTableAnalyzer()

		if analyzer == nil {
			t.Error("期望创建瑞士表分析器成功")
		}

		if len(analyzer.benchmarks) != 0 {
			t.Errorf("期望初始基准测试结果为空，实际长度为%d", len(analyzer.benchmarks))
		}
	})

	t.Run("基准测试执行", func(t *testing.T) {
		analyzer := NewSwissTableAnalyzer()

		// 模拟基准测试执行
		analyzer.benchmarkInsertOperations()
		analyzer.benchmarkLookupOperations()
		analyzer.benchmarkDeleteOperations()

		expectedBenchmarks := 3
		if len(analyzer.benchmarks) != expectedBenchmarks {
			t.Errorf("期望基准测试结果数量为%d，实际为%d",
				expectedBenchmarks, len(analyzer.benchmarks))
		}

		// 检查操作指标
		if analyzer.metrics.InsertOps == 0 {
			t.Error("期望插入操作数量大于0")
		}

		if analyzer.metrics.LookupOps == 0 {
			t.Error("期望查找操作数量大于0")
		}

		if analyzer.metrics.DeleteOps == 0 {
			t.Error("期望删除操作数量大于0")
		}
	})

	t.Run("性能指标验证", func(t *testing.T) {
		analyzer := NewSwissTableAnalyzer()
		analyzer.benchmarkInsertOperations()

		// 验证性能指标
		for _, result := range analyzer.benchmarks {
			if result.Throughput <= 0 {
				t.Errorf("期望吞吐量大于0，实际为%.2f", result.Throughput)
			}

			if result.Duration <= 0 {
				t.Errorf("期望执行时间大于0，实际为%v", result.Duration)
			}
		}
	})
}

// ==================
// 4. 运行时性能测试
// ==================

func TestPerformanceProfiler(t *testing.T) {
	t.Run("创建性能分析器", func(t *testing.T) {
		profiler := NewPerformanceProfiler("test")

		if profiler == nil {
			t.Error("期望创建性能分析器成功")
		}

		if profiler.name != "test" {
			t.Errorf("期望分析器名称为'test'，实际为'%s'", profiler.name)
		}
	})

	t.Run("内存分配测试", func(t *testing.T) {
		profiler := NewPerformanceProfiler("memory_test")

		measurement := profiler.StartMeasurement("test_allocation")

		// 分配更多内存以确保测量时间 > 0
		data := make([][]byte, 10000)
		for i := 0; i < 10000; i++ {
			data[i] = make([]byte, 2048)
		}

		profiler.EndMeasurement(measurement)

		if measurement.Duration <= 0 {
			t.Error("期望测量时间大于0")
		}

		if measurement.MemoryAfter <= measurement.MemoryBefore {
			t.Error("期望内存使用量增加")
		}

		// 保持data的引用以防止优化
		_ = data[0]
	})

	t.Run("并发性能测试", func(t *testing.T) {
		profiler := NewPerformanceProfiler("concurrency_test")

		measurement := profiler.StartMeasurement("concurrent_work")

		var wg sync.WaitGroup
		const numGoroutines = 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// 模拟一些工作
				sum := 0
				for j := 0; j < 1000; j++ {
					sum += j
				}
				_ = sum
			}()
		}

		wg.Wait()
		profiler.EndMeasurement(measurement)

		if measurement.GoroutinesAfter < measurement.GoroutinesBefore {
			t.Error("期望goroutine数量在执行期间增加")
		}
	})
}

// ==================
// 5. 现代工具链测试
// ==================

func TestModernToolchain(t *testing.T) {
	t.Run("创建工具链", func(t *testing.T) {
		toolchain := NewModernToolchain()

		if toolchain == nil {
			t.Error("期望创建现代工具链成功")
		}

		if len(toolchain.tools) == 0 {
			t.Error("期望默认工具不为空")
		}
	})

	t.Run("创建工作流", func(t *testing.T) {
		toolchain := NewModernToolchain()
		toolchain.CreateModernWorkflow()

		if len(toolchain.workflows) == 0 {
			t.Error("期望创建工作流成功")
		}

		workflow := toolchain.workflows[0]
		if workflow.Name != "modern-ci" {
			t.Errorf("期望工作流名称为'modern-ci'，实际为'%s'", workflow.Name)
		}

		if len(workflow.Steps) == 0 {
			t.Error("期望工作流包含步骤")
		}
	})

	t.Run("执行工作流", func(t *testing.T) {
		toolchain := NewModernToolchain()
		toolchain.CreateModernWorkflow()

		err := toolchain.ExecuteWorkflow("modern-ci")
		if err != nil {
			t.Errorf("期望执行工作流成功，实际错误: %v", err)
		}

		if toolchain.statistics.WorkflowsRun != 1 {
			t.Errorf("期望工作流运行次数为1，实际为%d", toolchain.statistics.WorkflowsRun)
		}
	})

	t.Run("执行不存在的工作流", func(t *testing.T) {
		toolchain := NewModernToolchain()

		err := toolchain.ExecuteWorkflow("nonexistent")
		if err == nil {
			t.Error("期望执行不存在的工作流返回错误")
		}
	})
}

// ==================
// 6. 面向未来编程模式测试
// ==================

func TestGenericOrderedSet(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		set := NewGenericOrderedSet[string]()

		// 测试添加
		set.Add("apple")
		set.Add("banana")
		set.Add("apple") // 重复元素

		if set.Size() != 2 {
			t.Errorf("期望集合大小为2，实际为%d", set.Size())
		}

		// 测试包含
		if !set.Contains("apple") {
			t.Error("期望集合包含'apple'")
		}

		if !set.Contains("banana") {
			t.Error("期望集合包含'banana'")
		}

		if set.Contains("orange") {
			t.Error("期望集合不包含'orange'")
		}

		// 测试删除
		if !set.Remove("apple") {
			t.Error("期望删除'apple'成功")
		}

		if set.Contains("apple") {
			t.Error("期望删除后不包含'apple'")
		}

		if set.Size() != 1 {
			t.Errorf("期望删除后集合大小为1，实际为%d", set.Size())
		}

		// 测试删除不存在的元素
		if set.Remove("orange") {
			t.Error("期望删除不存在元素返回false")
		}

		// 测试清空
		set.Clear()
		if set.Size() != 0 {
			t.Errorf("期望清空后集合大小为0，实际为%d", set.Size())
		}
	})

	t.Run("并发安全", func(t *testing.T) {
		set := NewGenericOrderedSet[int]()
		const numGoroutines = 10
		const numOperations = 100

		var wg sync.WaitGroup

		// 并发添加
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(start int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					set.Add(start*numOperations + j)
				}
			}(i)
		}

		wg.Wait()

		expectedSize := numGoroutines * numOperations
		if set.Size() != expectedSize {
			t.Errorf("期望并发添加后集合大小为%d，实际为%d", expectedSize, set.Size())
		}
	})
}

func TestConcurrentPipeline(t *testing.T) {
	t.Run("基本功能", func(t *testing.T) {
		pipeline := NewConcurrentPipeline(2, func(x int) int {
			return x * x
		})

		pipeline.Start()

		// 发送数据
		go func() {
			for i := 1; i <= 5; i++ {
				pipeline.Input() <- i
			}
			pipeline.Close()
		}()

		// 收集结果
		results := make([]int, 0)
		for result := range pipeline.Output() {
			results = append(results, result)
		}

		if len(results) != 5 {
			t.Errorf("期望结果数量为5，实际为%d", len(results))
		}

		// 检查结果是否正确（顺序可能不同）
		expectedSum := 1 + 4 + 9 + 16 + 25 // 1² + 2² + 3² + 4² + 5²
		actualSum := 0
		for _, result := range results {
			actualSum += result
		}

		if actualSum != expectedSum {
			t.Errorf("期望结果总和为%d，实际为%d", expectedSum, actualSum)
		}
	})

	t.Run("取消操作", func(t *testing.T) {
		pipeline := NewConcurrentPipeline(2, func(x int) int {
			time.Sleep(100 * time.Millisecond) // 模拟慢操作
			return x * x
		})

		pipeline.Start()

		// 快速关闭
		go func() {
			pipeline.Input() <- 1
			pipeline.Close()
		}()

		// 应该能够正常完成而不会阻塞
		results := make([]int, 0)
		for result := range pipeline.Output() {
			results = append(results, result)
		}

		if len(results) > 1 {
			t.Errorf("期望结果数量不超过1，实际为%d", len(results))
		}
	})
}

func TestContextualService(t *testing.T) {
	t.Run("成功执行", func(t *testing.T) {
		service := NewContextualService("test-service",
			func(ctx context.Context, req int) (int, error) {
				return req * 2, nil
			},
			1*time.Second,
		)

		ctx := context.Background()
		result, err := service.Execute(ctx, 21)

		if err != nil {
			t.Errorf("期望执行成功，实际错误: %v", err)
		}

		if result != 42 {
			t.Errorf("期望结果为42，实际为%d", result)
		}

		metrics := service.GetMetrics()
		if metrics.RequestCount != 1 {
			t.Errorf("期望请求数量为1，实际为%d", metrics.RequestCount)
		}

		if metrics.SuccessCount != 1 {
			t.Errorf("期望成功数量为1，实际为%d", metrics.SuccessCount)
		}
	})

	t.Run("超时处理", func(t *testing.T) {
		service := NewContextualService("slow-service",
			func(ctx context.Context, req int) (int, error) {
				select {
				case <-time.After(2 * time.Second): // 比超时时间长
					return req * 2, nil
				case <-ctx.Done():
					return 0, ctx.Err()
				}
			},
			100*time.Millisecond,
		)

		ctx := context.Background()
		_, err := service.Execute(ctx, 21)

		if err == nil {
			t.Error("期望超时错误")
		}

		metrics := service.GetMetrics()
		if metrics.TimeoutCount == 0 {
			t.Error("期望超时计数大于0")
		}
	})

	t.Run("重试机制", func(t *testing.T) {
		attemptCount := int64(0)

		service := NewContextualService("retry-service",
			func(ctx context.Context, req int) (int, error) {
				count := atomic.AddInt64(&attemptCount, 1)
				if count < 3 {
					return 0, fmt.Errorf("temporary error")
				}
				return req * 2, nil
			},
			1*time.Second,
		)

		ctx := context.Background()
		result, err := service.Execute(ctx, 21)

		if err != nil {
			t.Errorf("期望重试后成功，实际错误: %v", err)
		}

		if result != 42 {
			t.Errorf("期望结果为42，实际为%d", result)
		}

		if attemptCount != 3 {
			t.Errorf("期望尝试次数为3，实际为%d", attemptCount)
		}
	})
}

// ==================
// 7. 基准测试
// ==================

func BenchmarkGenericMap(b *testing.B) {
	userMap := make(GenericMap[string, int])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("user_%d", i%1000)
		userMap[key] = i
	}
}

func BenchmarkGenericSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := make(GenericSlice[int], 0, 100)
		for j := 0; j < 100; j++ {
			slice = append(slice, j)
		}
	}
}

func BenchmarkGenericOrderedSetAdd(b *testing.B) {
	set := NewGenericOrderedSet[int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Add(i)
	}
}

func BenchmarkGenericOrderedSetContains(b *testing.B) {
	set := NewGenericOrderedSet[int]()

	// 预填充数据
	for i := 0; i < 10000; i++ {
		set.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Contains(i % 10000)
	}
}

func BenchmarkConcurrentPipeline(b *testing.B) {
	pipeline := NewConcurrentPipeline(runtime.NumCPU(), func(x int) int {
		return x * x
	})

	pipeline.Start()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			select {
			case pipeline.Input() <- i:
				i++
			default:
			}
		}
	})

	pipeline.Close()

	// 消费所有结果
	for range pipeline.Output() {
	}
}

func BenchmarkContextualService(b *testing.B) {
	service := NewContextualService("bench-service",
		func(ctx context.Context, req int) (int, error) {
			return req * 2, nil
		},
		1*time.Second,
	)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.Execute(ctx, 42)
			if err != nil {
				b.Errorf("服务执行失败: %v", err)
			}
		}
	})
}

// 内存分配基准测试
func BenchmarkGenericMapMemory(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userMap := make(GenericMap[string, int], 100)
		for j := 0; j < 100; j++ {
			key := fmt.Sprintf("user_%d", j)
			userMap[key] = j
		}
	}
}

func BenchmarkGenericSliceMemory(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := make(GenericSlice[int], 0, 100)
		for j := 0; j < 100; j++ {
			slice = append(slice, j)
		}
	}
}

// ==================
// 8. 示例测试
// ==================

func ExampleGenericMap() {
	// 使用泛型映射别名
	userMap := make(GenericMap[string, int])
	userMap["alice"] = 25
	userMap["bob"] = 30

	fmt.Printf("用户数量: %d\n", len(userMap))
	fmt.Printf("Alice的年龄: %d\n", userMap["alice"])

	// Output:
	// 用户数量: 2
	// Alice的年龄: 25
}

func ExampleGenericSlice() {
	// 使用泛型切片别名
	numbers := GenericSlice[int]{1, 2, 3, 4, 5}

	fmt.Printf("数字数量: %d\n", len(numbers))
	fmt.Printf("第一个数字: %d\n", numbers[0])
	fmt.Printf("最后一个数字: %d\n", numbers[len(numbers)-1])

	// Output:
	// 数字数量: 5
	// 第一个数字: 1
	// 最后一个数字: 5
}

func ExampleGenericOrderedSet() {
	// 创建字符串集合
	set := NewGenericOrderedSet[string]()

	// 添加元素
	set.Add("apple")
	set.Add("banana")
	set.Add("apple") // 重复元素不会被添加

	fmt.Printf("集合大小: %d\n", set.Size())
	fmt.Printf("包含apple: %t\n", set.Contains("apple"))
	fmt.Printf("包含orange: %t\n", set.Contains("orange"))

	// Output:
	// 集合大小: 2
	// 包含apple: true
	// 包含orange: false
}

func ExampleConcurrentPipeline() {
	// 创建计算平方的并发管道
	pipeline := NewConcurrentPipeline(2, func(x int) int {
		return x * x
	})

	pipeline.Start()

	// 发送数据
	go func() {
		for i := 1; i <= 3; i++ {
			pipeline.Input() <- i
		}
		pipeline.Close()
	}()

	// 给足够时间处理数据
	time.Sleep(10 * time.Millisecond)

	// 收集结果
	results := make([]int, 0)
	for result := range pipeline.Output() {
		results = append(results, result)
	}

	fmt.Printf("处理了 %d 个结果\n", len(results))

	// Output:
	// 处理了 3 个结果
}

func ExampleContextualService() {
	// 创建数学服务
	mathService := NewContextualService("math-service",
		func(ctx context.Context, req int) (int, error) {
			return req * 2, nil
		},
		1*time.Second,
	)

	ctx := context.Background()
	result, err := mathService.Execute(ctx, 21)

	if err != nil {
		fmt.Printf("服务错误: %v\n", err)
		return
	}

	fmt.Printf("计算结果: %d\n", result)

	metrics := mathService.GetMetrics()
	fmt.Printf("服务请求次数: %d\n", metrics.RequestCount)

	// Output:
	// 计算结果: 42
	// 服务请求次数: 1
}

// ==================
// 9. 表格驱动测试示例
// ==================

func TestGenericMapOperations(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() GenericMap[string, int]
		key      string
		value    int
		expected int
		exists   bool
	}{
		{
			name: "添加新键值对",
			setup: func() GenericMap[string, int] {
				return make(GenericMap[string, int])
			},
			key:      "new_key",
			value:    100,
			expected: 100,
			exists:   true,
		},
		{
			name: "覆盖存在的键",
			setup: func() GenericMap[string, int] {
				m := make(GenericMap[string, int])
				m["existing"] = 50
				return m
			},
			key:      "existing",
			value:    200,
			expected: 200,
			exists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			m[tt.key] = tt.value

			if value, exists := m[tt.key]; exists != tt.exists || value != tt.expected {
				t.Errorf("映射操作失败: 期望值=%d, 存在=%t, 实际值=%d, 存在=%t",
					tt.expected, tt.exists, value, exists)
			}
		})
	}
}

// ==================
// 10. 测试辅助函数
// ==================

// 测试设置和清理
func TestMain(m *testing.M) {
	// 测试前设置
	fmt.Println("开始Go 1.24特性测试...")

	// 运行测试
	code := m.Run()

	// 测试后清理
	fmt.Println("Go 1.24特性测试完成")

	// 退出
	os.Exit(code)
}

// 性能测试辅助函数
func setupBenchmarkData(size int) []int {
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	return data
}

// 并发测试辅助函数
func runConcurrentTest(t *testing.T, workers int, work func()) {
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			work()
		}()
	}

	wg.Wait()
}

// 超时测试辅助函数
func runWithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	done := make(chan bool)

	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(timeout):
		t.Errorf("测试超时（%v）", timeout)
	}
}

/*
测试运行命令：
go test -v                           # 运行所有测试
go test -v -run TestGeneric          # 运行泛型相关测试
go test -v -bench=.                  # 运行所有基准测试
go test -v -bench=BenchmarkGeneric   # 运行泛型基准测试
go test -v -cover                    # 运行测试并显示覆盖率
go test -v -race                     # 运行竞态检测
go test -v -cpuprofile=cpu.prof      # 生成CPU性能分析
go test -v -memprofile=mem.prof      # 生成内存性能分析

覆盖率目标：>90%
性能基准：确保新特性不会导致性能回退
并发安全：所有并发组件都通过竞态检测
*/
