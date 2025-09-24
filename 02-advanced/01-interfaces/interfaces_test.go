package main

import (
	"math"
	"reflect"
	"testing"
)

// ====================
// 1. 基本接口实现测试
// ====================

func TestRectangleShape(t *testing.T) {
	rect := Rectangle{Width: 10, Height: 5}

	// 测试面积计算
	expectedArea := 50.0
	if area := rect.Area(); area != expectedArea {
		t.Errorf("Rectangle.Area() = %f, want %f", area, expectedArea)
	}

	// 测试周长计算
	expectedPerimeter := 30.0
	if perimeter := rect.Perimeter(); perimeter != expectedPerimeter {
		t.Errorf("Rectangle.Perimeter() = %f, want %f", perimeter, expectedPerimeter)
	}

	// 测试接口类型
	var shape Shape = rect
	if shape.Area() != expectedArea {
		t.Errorf("Rectangle as Shape interface: Area() = %f, want %f", shape.Area(), expectedArea)
	}
}

func TestCircleShape(t *testing.T) {
	circle := Circle{Radius: 3}

	// 测试面积计算
	expectedArea := math.Pi * 9
	if area := circle.Area(); math.Abs(area-expectedArea) > 0.001 {
		t.Errorf("Circle.Area() = %f, want %f", area, expectedArea)
	}

	// 测试周长计算
	expectedPerimeter := 2 * math.Pi * 3
	if perimeter := circle.Perimeter(); math.Abs(perimeter-expectedPerimeter) > 0.001 {
		t.Errorf("Circle.Perimeter() = %f, want %f", perimeter, expectedPerimeter)
	}

	// 测试接口类型
	var shape Shape = circle
	if math.Abs(shape.Area()-expectedArea) > 0.001 {
		t.Errorf("Circle as Shape interface: Area() = %f, want %f", shape.Area(), expectedArea)
	}
}

func TestTriangleShape(t *testing.T) {
	triangle := Triangle{Base: 6, Height: 4}

	// 测试面积计算
	expectedArea := 12.0
	if area := triangle.Area(); area != expectedArea {
		t.Errorf("Triangle.Area() = %f, want %f", area, expectedArea)
	}

	// 测试周长计算 (等腰三角形)
	side := math.Sqrt(9 + 16) // sqrt((3)^2 + (4)^2)
	expectedPerimeter := 6 + 2*side
	if perimeter := triangle.Perimeter(); math.Abs(perimeter-expectedPerimeter) > 0.001 {
		t.Errorf("Triangle.Perimeter() = %f, want %f", perimeter, expectedPerimeter)
	}
}

// ====================
// 2. 接口多态测试
// ====================

func TestShapePolymorphism(t *testing.T) {
	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 6, Height: 4},
	}

	expectedAreas := []float64{12.0, math.Pi * 4, 12.0}

	for i, shape := range shapes {
		area := shape.Area()
		if math.Abs(area-expectedAreas[i]) > 0.001 {
			t.Errorf("Shape[%d].Area() = %f, want %f", i, area, expectedAreas[i])
		}
	}
}

func TestShapeInterfaceComparison(t *testing.T) {
	var s1, s2 Shape
	s1 = Rectangle{Width: 2, Height: 3}
	s2 = Rectangle{Width: 2, Height: 3}

	// 测试相同值的接口比较
	if s1 != s2 {
		t.Error("Equal Rectangle shapes should be equal as interfaces")
	}

	s2 = Rectangle{Width: 3, Height: 2}
	if s1 == s2 {
		t.Error("Different Rectangle shapes should not be equal as interfaces")
	}
}

// ====================
// 3. 接口组合测试
// ====================

func TestDrawableInterface(t *testing.T) {
	rect := Rectangle{Width: 5, Height: 3}

	// 测试 Drawable 接口
	var drawable Drawable = rect
	// Draw() 方法测试 - 这里我们只能测试它不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Rectangle.Draw() panicked: %v", r)
		}
	}()
	drawable.Draw()
}

func TestResizableInterface(t *testing.T) {
	rect := &Rectangle{Width: 10, Height: 5}

	// 测试 Resizable 接口
	var resizable Resizable = rect

	originalWidth := rect.Width
	originalHeight := rect.Height
	factor := 1.5

	resizable.Resize(factor)

	expectedWidth := originalWidth * factor
	expectedHeight := originalHeight * factor

	if rect.Width != expectedWidth {
		t.Errorf("After Resize(%f), Width = %f, want %f", factor, rect.Width, expectedWidth)
	}
	if rect.Height != expectedHeight {
		t.Errorf("After Resize(%f), Height = %f, want %f", factor, rect.Height, expectedHeight)
	}
}

func TestDrawableShapeComposition(t *testing.T) {
	rect := &Rectangle{Width: 8, Height: 4}

	// 测试组合接口
	var drawableShape DrawableShape = rect

	// 测试 Shape 方法
	expectedArea := 32.0
	if area := drawableShape.Area(); area != expectedArea {
		t.Errorf("DrawableShape.Area() = %f, want %f", area, expectedArea)
	}

	// 测试 Drawable 方法（不会panic）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DrawableShape.Draw() panicked: %v", r)
		}
	}()
	drawableShape.Draw()
}

func TestResizableDrawableShapeComposition(t *testing.T) {
	rect := &Rectangle{Width: 6, Height: 3}

	var resizableShape ResizableDrawableShape = rect

	// 测试所有接口方法
	originalArea := resizableShape.Area()
	resizableShape.Resize(2.0)
	newArea := resizableShape.Area()

	// 面积应该是原来的4倍 (2^2)
	expectedNewArea := originalArea * 4
	if math.Abs(newArea-expectedNewArea) > 0.001 {
		t.Errorf("After resize by 2.0, area = %f, want %f", newArea, expectedNewArea)
	}

	// 测试 Draw 方法
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ResizableDrawableShape.Draw() panicked: %v", r)
		}
	}()
	resizableShape.Draw()
}

// ====================
// 4. 接口参数和返回值测试
// ====================

func TestCalculateTotalArea(t *testing.T) {
	tests := []struct {
		name     string
		shapes   []Shape
		expected float64
	}{
		{
			name:     "empty slice",
			shapes:   []Shape{},
			expected: 0.0,
		},
		{
			name: "single rectangle",
			shapes: []Shape{
				Rectangle{Width: 4, Height: 3},
			},
			expected: 12.0,
		},
		{
			name: "multiple shapes",
			shapes: []Shape{
				Rectangle{Width: 4, Height: 3},
				Circle{Radius: 2},
				Triangle{Base: 6, Height: 4},
			},
			expected: 12.0 + math.Pi*4 + 12.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := calculateTotalArea(tt.shapes)
			if math.Abs(total-tt.expected) > 0.001 {
				t.Errorf("calculateTotalArea() = %f, want %f", total, tt.expected)
			}
		})
	}
}

func TestFindLargestShape(t *testing.T) {
	tests := []struct {
		name         string
		shapes       []Shape
		expectedArea float64
		expectedNil  bool
	}{
		{
			name:        "empty slice",
			shapes:      []Shape{},
			expectedNil: true,
		},
		{
			name: "single shape",
			shapes: []Shape{
				Rectangle{Width: 4, Height: 3},
			},
			expectedArea: 12.0,
		},
		{
			name: "rectangle is largest",
			shapes: []Shape{
				Rectangle{Width: 5, Height: 4}, // 20
				Circle{Radius: 2},              // ~12.57
				Triangle{Base: 6, Height: 4},   // 12
			},
			expectedArea: 20.0,
		},
		{
			name: "circle is largest",
			shapes: []Shape{
				Rectangle{Width: 2, Height: 3}, // 6
				Circle{Radius: 3},              // ~28.27
				Triangle{Base: 4, Height: 3},   // 6
			},
			expectedArea: math.Pi * 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			largest := findLargestShape(tt.shapes)

			if tt.expectedNil {
				if largest != nil {
					t.Errorf("findLargestShape() = %v, want nil", largest)
				}
			} else {
				if largest == nil {
					t.Fatal("findLargestShape() = nil, want non-nil")
				}
				area := largest.Area()
				if math.Abs(area-tt.expectedArea) > 0.001 {
					t.Errorf("findLargestShape().Area() = %f, want %f", area, tt.expectedArea)
				}
			}
		})
	}
}

// ====================
// 5. 工厂函数测试
// ====================

func TestGetShapeFactory(t *testing.T) {
	tests := []struct {
		name         string
		shapeType    string
		expectedType string
		expectedArea float64
	}{
		{
			name:         "rectangle factory",
			shapeType:    "rectangle",
			expectedType: "*main.Rectangle",
			expectedArea: 12.0, // 4 * 3
		},
		{
			name:         "circle factory",
			shapeType:    "circle",
			expectedType: "*main.Circle",
			expectedArea: math.Pi * 4, // π * 2^2
		},
		{
			name:         "unknown type defaults to rectangle",
			shapeType:    "unknown",
			expectedType: "*main.Rectangle",
			expectedArea: 1.0, // 1 * 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := getShapeFactory(tt.shapeType)
			shape := factory()

			// 检查类型
			actualType := reflect.TypeOf(shape).String()
			if actualType != tt.expectedType {
				t.Errorf("getShapeFactory(%q) created type %s, want %s", tt.shapeType, actualType, tt.expectedType)
			}

			// 检查面积
			area := shape.Area()
			if math.Abs(area-tt.expectedArea) > 0.001 {
				t.Errorf("factory created shape area = %f, want %f", area, tt.expectedArea)
			}
		})
	}
}

// ====================
// 6. 函数式接口测试
// ====================

func TestGetShapeProcessor(t *testing.T) {
	shapes := []Shape{
		Rectangle{Width: 2, Height: 2}, // Area: 4
		Circle{Radius: 2},              // Area: ~12.57
		Triangle{Base: 4, Height: 6},   // Area: 12
		Rectangle{Width: 1, Height: 5}, // Area: 5
	}

	processor := getShapeProcessor()

	// 过滤面积大于10的图形
	filtered := processor(shapes, func(s Shape) bool {
		return s.Area() > 10
	})

	expectedCount := 2 // Circle 和 Triangle
	if len(filtered) != expectedCount {
		t.Errorf("processor filtered %d shapes, want %d", len(filtered), expectedCount)
	}

	// 验证过滤结果
	for _, shape := range filtered {
		if shape.Area() <= 10 {
			t.Errorf("filtered shape has area %f, should be > 10", shape.Area())
		}
	}

	// 过滤面积小于6的图形
	filtered = processor(shapes, func(s Shape) bool {
		return s.Area() < 6
	})

	expectedCount = 2 // 两个 Rectangle: 4, 5
	if len(filtered) != expectedCount {
		t.Errorf("processor filtered %d shapes, want %d", len(filtered), expectedCount)
	}
}

// ====================
// 7. 空接口测试
// ====================

func TestEmptyInterface(t *testing.T) {
	var empty interface{}

	// 测试存储不同类型
	testValues := []interface{}{
		42,
		"hello",
		3.14,
		[]int{1, 2, 3},
		map[string]int{"key": 1},
		Rectangle{Width: 2, Height: 3},
	}

	for i, value := range testValues {
		empty = value
		if !reflect.DeepEqual(empty, value) {
			t.Errorf("empty interface test %d: stored %v, got %v", i, value, empty)
		}
	}
}

// ====================
// 8. 类型断言测试（模拟）
// ====================

func TestTypeAssertions(t *testing.T) {
	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 6, Height: 4},
	}

	// 测试类型检查
	rectangleCount := 0
	circleCount := 0
	triangleCount := 0

	for _, shape := range shapes {
		switch shape.(type) {
		case Rectangle:
			rectangleCount++
		case Circle:
			circleCount++
		case Triangle:
			triangleCount++
		}
	}

	if rectangleCount != 1 {
		t.Errorf("Expected 1 rectangle, got %d", rectangleCount)
	}
	if circleCount != 1 {
		t.Errorf("Expected 1 circle, got %d", circleCount)
	}
	if triangleCount != 1 {
		t.Errorf("Expected 1 triangle, got %d", triangleCount)
	}
}

// ====================
// 9. 接口性能基准测试
// ====================

func BenchmarkShapeInterface(b *testing.B) {
	shape := Rectangle{Width: 10, Height: 5}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var s Shape = shape
		_ = s.Area()
	}
}

func BenchmarkDirectMethodCall(b *testing.B) {
	shape := Rectangle{Width: 10, Height: 5}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = shape.Area()
	}
}

func BenchmarkCalculateTotalArea(b *testing.B) {
	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 6, Height: 4},
		Rectangle{Width: 2, Height: 5},
		Circle{Radius: 1.5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTotalArea(shapes)
	}
}

// ====================
// 10. 边界情况测试
// ====================

func TestEdgeCases(t *testing.T) {
	// 测试零值
	t.Run("zero value shapes", func(t *testing.T) {
		rect := Rectangle{} // Width: 0, Height: 0
		if area := rect.Area(); area != 0 {
			t.Errorf("Zero rectangle area = %f, want 0", area)
		}

		circle := Circle{} // Radius: 0
		if area := circle.Area(); area != 0 {
			t.Errorf("Zero circle area = %f, want 0", area)
		}

		triangle := Triangle{} // Base: 0, Height: 0
		if area := triangle.Area(); area != 0 {
			t.Errorf("Zero triangle area = %f, want 0", area)
		}
	})

	// 测试负值
	t.Run("negative values", func(t *testing.T) {
		rect := Rectangle{Width: -2, Height: 3}
		area := rect.Area()
		if area != -6 {
			t.Errorf("Negative width rectangle area = %f, want -6", area)
		}

		circle := Circle{Radius: -2}
		area = circle.Area()
		expected := math.Pi * 4 // radius^2 is positive
		if math.Abs(area-expected) > 0.001 {
			t.Errorf("Negative radius circle area = %f, want %f", area, expected)
		}
	})

	// 测试非常大的值
	t.Run("large values", func(t *testing.T) {
		rect := Rectangle{Width: 1e6, Height: 1e6}
		area := rect.Area()
		expected := 1e12
		if area != expected {
			t.Errorf("Large rectangle area = %f, want %f", area, expected)
		}
	})
}

// ====================
// 11. 接口实现验证测试
// ====================

func TestInterfaceImplementations(t *testing.T) {
	// 验证所有类型都实现了预期的接口

	// 检查 Rectangle 实现的接口
	var _ Shape = Rectangle{}
	var _ Drawable = Rectangle{}
	var _ Resizable = &Rectangle{}
	var _ DrawableShape = Rectangle{}
	var _ ResizableDrawableShape = &Rectangle{}

	// 检查 Circle 实现的接口
	var _ Shape = Circle{}
	var _ Drawable = Circle{}
	var _ Resizable = &Circle{}
	var _ DrawableShape = Circle{}
	var _ ResizableDrawableShape = &Circle{}

	// 检查 Triangle 只实现 Shape 接口
	var _ Shape = Triangle{}
	// Triangle 不实现 Drawable 接口
	// var _ Drawable = Triangle{} // 这会编译错误

	t.Log("All interface implementations verified")
}
