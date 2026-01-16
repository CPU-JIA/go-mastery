package main

import (
	"sync"
	"testing"
	"time"
)

// =============================================================================
// 管道模式函数测试
// =============================================================================

// TestGenerateNumbers 测试数字生成器
func TestGenerateNumbers(t *testing.T) {
	tests := []struct {
		name     string
		max      int
		expected []int
	}{
		{
			name:     "生成1到5的数字",
			max:      5,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "生成1到1的数字",
			max:      1,
			expected: []int{1},
		},
		{
			name:     "生成1到10的数字",
			max:      10,
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := generateNumbers(tt.max)

			var results []int
			for num := range ch {
				results = append(results, num)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("期望长度 %d, 实际长度 %d", len(tt.expected), len(results))
				return
			}

			for i, v := range results {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %d, 实际 %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestSquareNumbers 测试平方计算
func TestSquareNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "计算1到5的平方",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 4, 9, 16, 25},
		},
		{
			name:     "计算单个数字的平方",
			input:    []int{7},
			expected: []int{49},
		},
		{
			name:     "空输入",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建输入通道
			inputCh := make(chan int, len(tt.input))
			for _, v := range tt.input {
				inputCh <- v
			}
			close(inputCh)

			// 执行平方计算
			outputCh := squareNumbers(inputCh)

			var results []int
			for num := range outputCh {
				results = append(results, num)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("期望长度 %d, 实际长度 %d", len(tt.expected), len(results))
				return
			}

			for i, v := range results {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %d, 实际 %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestFilterEven 测试偶数过滤
func TestFilterEven(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "过滤1到10中的偶数",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expected: []int{2, 4, 6, 8, 10},
		},
		{
			name:     "全部是奇数",
			input:    []int{1, 3, 5, 7, 9},
			expected: []int{},
		},
		{
			name:     "全部是偶数",
			input:    []int{2, 4, 6},
			expected: []int{2, 4, 6},
		},
		{
			name:     "空输入",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建输入通道
			inputCh := make(chan int, len(tt.input))
			for _, v := range tt.input {
				inputCh <- v
			}
			close(inputCh)

			// 执行过滤
			outputCh := filterEven(inputCh)

			var results []int
			for num := range outputCh {
				results = append(results, num)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("期望长度 %d, 实际长度 %d", len(tt.expected), len(results))
				return
			}

			for i, v := range results {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %d, 实际 %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestAddPrefix 测试添加前缀
func TestAddPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		prefix   string
		expected []string
	}{
		{
			name:     "添加数字前缀",
			input:    []int{1, 2, 3},
			prefix:   "num:",
			expected: []string{"num:1", "num:2", "num:3"},
		},
		{
			name:     "添加空前缀",
			input:    []int{10, 20},
			prefix:   "",
			expected: []string{"10", "20"},
		},
		{
			name:     "空输入",
			input:    []int{},
			prefix:   "test:",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建输入通道
			inputCh := make(chan int, len(tt.input))
			for _, v := range tt.input {
				inputCh <- v
			}
			close(inputCh)

			// 执行添加前缀
			outputCh := addPrefix(inputCh, tt.prefix)

			var results []string
			for str := range outputCh {
				results = append(results, str)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("期望长度 %d, 实际长度 %d", len(tt.expected), len(results))
				return
			}

			for i, v := range results {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %s, 实际 %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestPipelineIntegration 测试管道集成
func TestPipelineIntegration(t *testing.T) {
	// 构建完整管道：生成数字 -> 计算平方 -> 过滤偶数
	numbers := generateNumbers(5)     // 1, 2, 3, 4, 5
	squares := squareNumbers(numbers) // 1, 4, 9, 16, 25
	evens := filterEven(squares)      // 4, 16

	expected := []int{4, 16}
	var results []int

	for num := range evens {
		results = append(results, num)
	}

	if len(results) != len(expected) {
		t.Errorf("期望长度 %d, 实际长度 %d", len(expected), len(results))
		return
	}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("索引 %d: 期望 %d, 实际 %d", i, expected[i], v)
		}
	}
}

// =============================================================================
// 处理器函数测试
// =============================================================================

// TestProcessor 测试处理器函数
func TestProcessor(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int // 平方值
	}{
		{
			name:     "处理多个数字",
			input:    []int{2, 3, 4},
			expected: []int{4, 9, 16},
		},
		{
			name:     "处理单个数字",
			input:    []int{5},
			expected: []int{25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCh := make(chan int, len(tt.input))
			outputCh := make(chan int, len(tt.input))

			// 启动处理器
			go processor(inputCh, outputCh)

			// 发送输入
			for _, v := range tt.input {
				inputCh <- v
			}
			close(inputCh)

			// 收集结果
			var results []int
			for num := range outputCh {
				results = append(results, num)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("期望长度 %d, 实际长度 %d", len(tt.expected), len(results))
				return
			}

			for i, v := range results {
				if v != tt.expected[i] {
					t.Errorf("索引 %d: 期望 %d, 实际 %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// =============================================================================
// TaskQueue 测试
// =============================================================================

// TestNewTaskQueue 测试创建任务队列
func TestNewTaskQueue(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		workers    int
	}{
		{
			name:       "标准配置",
			bufferSize: 10,
			workers:    3,
		},
		{
			name:       "单工作者",
			bufferSize: 5,
			workers:    1,
		},
		{
			name:       "大缓冲区",
			bufferSize: 100,
			workers:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewTaskQueue(tt.bufferSize, tt.workers)

			if queue == nil {
				t.Fatal("NewTaskQueue 返回 nil")
			}

			if queue.workers != tt.workers {
				t.Errorf("期望工作者数量 %d, 实际 %d", tt.workers, queue.workers)
			}

			if cap(queue.tasks) != tt.bufferSize {
				t.Errorf("期望任务通道容量 %d, 实际 %d", tt.bufferSize, cap(queue.tasks))
			}

			if cap(queue.results) != tt.bufferSize {
				t.Errorf("期望结果通道容量 %d, 实际 %d", tt.bufferSize, cap(queue.results))
			}
		})
	}
}

// TestTaskQueueSubmitAndProcess 测试任务提交和处理
func TestTaskQueueSubmitAndProcess(t *testing.T) {
	queue := NewTaskQueue(10, 2)
	queue.Start()

	// 提交任务
	tasks := []Task{
		{ID: 1, Data: "任务1", Priority: 1},
		{ID: 2, Data: "任务2", Priority: 2},
		{ID: 3, Data: "任务3", Priority: 1},
	}

	for _, task := range tasks {
		queue.Submit(task)
	}

	// 收集结果
	results := make(map[int]TaskResult)
	for i := 0; i < len(tasks); i++ {
		result := queue.GetResult()
		results[result.TaskID] = result
	}

	// 验证所有任务都被处理
	for _, task := range tasks {
		if _, exists := results[task.ID]; !exists {
			t.Errorf("任务 %d 未被处理", task.ID)
		}
	}

	queue.Stop()
}

// TestTaskQueueConcurrency 测试任务队列并发安全性
func TestTaskQueueConcurrency(t *testing.T) {
	queue := NewTaskQueue(100, 5)
	queue.Start()

	const numTasks = 50
	var wg sync.WaitGroup

	// 并发提交任务
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := Task{
				ID:       id,
				Data:     "并发任务",
				Priority: id % 3,
			}
			queue.Submit(task)
		}(i)
	}

	wg.Wait()

	// 收集所有结果
	receivedTasks := make(map[int]bool)
	for i := 0; i < numTasks; i++ {
		result := queue.GetResult()
		receivedTasks[result.TaskID] = true
	}

	// 验证所有任务都被处理
	if len(receivedTasks) != numTasks {
		t.Errorf("期望处理 %d 个任务, 实际处理 %d 个", numTasks, len(receivedTasks))
	}

	queue.Stop()
}

// =============================================================================
// 扇入扇出测试
// =============================================================================

// TestFanIn 测试扇入功能
func TestFanIn(t *testing.T) {
	input1 := make(chan int, 5)
	input2 := make(chan int, 5)
	output := make(chan int, 10)

	// 发送数据到两个输入通道
	go func() {
		for i := 1; i <= 5; i++ {
			input1 <- i
		}
		close(input1)
	}()

	go func() {
		for i := 6; i <= 10; i++ {
			input2 <- i
		}
		close(input2)
	}()

	// 启动扇入
	go fanIn(input1, input2, output)

	// 收集结果
	var results []int
	for num := range output {
		results = append(results, num)
	}

	// 验证收到所有数据
	if len(results) != 10 {
		t.Errorf("期望收到 10 个数据, 实际收到 %d 个", len(results))
	}

	// 验证所有数字都存在（顺序可能不同）
	received := make(map[int]bool)
	for _, v := range results {
		received[v] = true
	}

	for i := 1; i <= 10; i++ {
		if !received[i] {
			t.Errorf("缺少数字 %d", i)
		}
	}
}

// =============================================================================
// 安全随机数测试
// =============================================================================

// TestSecureRandomInt 测试安全随机数生成
func TestSecureRandomInt(t *testing.T) {
	tests := []struct {
		name string
		max  int
	}{
		{name: "小范围", max: 10},
		{name: "中等范围", max: 100},
		{name: "大范围", max: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				result := secureRandomInt(tt.max)
				if result < 0 || result >= tt.max {
					t.Errorf("随机数 %d 超出范围 [0, %d)", result, tt.max)
				}
			}
		})
	}
}

// TestSecureRandomFloat32 测试安全随机浮点数生成
func TestSecureRandomFloat32(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := secureRandomFloat32()
		if result < 0.0 || result >= 1.0 {
			t.Errorf("随机浮点数 %f 超出范围 [0.0, 1.0)", result)
		}
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkGenerateNumbers 基准测试数字生成器
func BenchmarkGenerateNumbers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ch := generateNumbers(100)
		for range ch {
			// 消费所有数据
		}
	}
}

// BenchmarkSquareNumbers 基准测试平方计算
func BenchmarkSquareNumbers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		inputCh := make(chan int, 100)
		go func() {
			for j := 1; j <= 100; j++ {
				inputCh <- j
			}
			close(inputCh)
		}()

		outputCh := squareNumbers(inputCh)
		for range outputCh {
			// 消费所有数据
		}
	}
}

// BenchmarkPipeline 基准测试完整管道
func BenchmarkPipeline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numbers := generateNumbers(100)
		squares := squareNumbers(numbers)
		evens := filterEven(squares)

		for range evens {
			// 消费所有数据
		}
	}
}

// BenchmarkTaskQueue 基准测试任务队列
func BenchmarkTaskQueue(b *testing.B) {
	queue := NewTaskQueue(100, 4)
	queue.Start()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := Task{
			ID:       i,
			Data:     "benchmark",
			Priority: 1,
		}
		queue.Submit(task)
		queue.GetResult()
	}

	b.StopTimer()
	queue.Stop()
}

// =============================================================================
// 竞态条件测试
// =============================================================================

// TestPipelineConcurrentAccess 测试管道并发访问
func TestPipelineConcurrentAccess(t *testing.T) {
	const numGoroutines = 10

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			numbers := generateNumbers(10)
			squares := squareNumbers(numbers)
			evens := filterEven(squares)

			count := 0
			for range evens {
				count++
			}

			// 1-10的平方中偶数有: 4, 16, 36, 64, 100 = 5个
			if count != 5 {
				t.Errorf("期望 5 个偶数平方, 实际 %d 个", count)
			}
		}()
	}

	wg.Wait()
}

// TestTaskQueueStress 压力测试任务队列
func TestTaskQueueStress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	queue := NewTaskQueue(1000, 10)
	queue.Start()

	const numTasks = 500
	var wg sync.WaitGroup

	// 并发提交
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numTasks; i++ {
			task := Task{
				ID:       i,
				Data:     "压力测试",
				Priority: i % 5,
			}
			queue.Submit(task)
		}
	}()

	// 并发收集结果
	results := make(chan TaskResult, numTasks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numTasks; i++ {
			result := queue.GetResult()
			results <- result
		}
		close(results)
	}()

	wg.Wait()

	// 统计结果
	count := 0
	for range results {
		count++
	}

	if count != numTasks {
		t.Errorf("期望处理 %d 个任务, 实际处理 %d 个", numTasks, count)
	}

	queue.Stop()
}

// =============================================================================
// 超时测试
// =============================================================================

// TestPipelineTimeout 测试管道超时处理
func TestPipelineTimeout(t *testing.T) {
	done := make(chan bool)

	go func() {
		numbers := generateNumbers(1000)
		squares := squareNumbers(numbers)
		evens := filterEven(squares)

		for range evens {
			// 消费数据
		}
		done <- true
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(5 * time.Second):
		t.Error("管道处理超时")
	}
}
