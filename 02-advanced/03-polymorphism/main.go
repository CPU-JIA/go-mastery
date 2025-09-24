package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

/*
=== Go语言进阶特性第三课：多态(Polymorphism) ===

学习目标：
1. 理解多态的概念和实现
2. 掌握接口多态的使用
3. 学会类型断言和类型选择
4. 了解多态的设计模式
5. 掌握面向对象编程思想

Go多态特点：
- 通过接口实现多态
- 隐式实现，无需显式声明
- 运行时动态分发
- 支持组合而非继承
- 面向接口编程的核心
*/

func main() {
	fmt.Println("=== Go语言多态学习 ===")

	// 1. 基本多态概念
	demonstrateBasicPolymorphism()

	// 2. 接口多态
	demonstrateInterfacePolymorphism()

	// 3. 多态中的类型断言
	demonstratePolymorphicTypeAssertion()

	// 4. 多态设计模式
	demonstratePolymorphicPatterns()

	// 5. 组合vs继承
	demonstrateCompositionVsInheritance()

	// 6. 多态的动态特性
	demonstrateDynamicPolymorphism()

	// 7. 多态的最佳实践
	demonstrateBestPractices()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本接口定义
type Shape interface {
	Area() float64
	Perimeter() float64
	String() string
}

type Drawable interface {
	Draw()
}

type Transformable interface {
	Scale(factor float64)
	Move(dx, dy float64)
}

// 形状类型实现
type Rectangle struct {
	Width, Height float64
	X, Y          float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f×%.1f) at (%.1f,%.1f)",
		r.Width, r.Height, r.X, r.Y)
}

func (r Rectangle) Draw() {
	fmt.Printf("  绘制矩形: %s\n", r.String())
}

func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func (r *Rectangle) Move(dx, dy float64) {
	r.X += dx
	r.Y += dy
}

type Circle struct {
	Radius float64
	X, Y   float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.1f) at (%.1f,%.1f)",
		c.Radius, c.X, c.Y)
}

func (c Circle) Draw() {
	fmt.Printf("  绘制圆形: %s\n", c.String())
}

func (c *Circle) Scale(factor float64) {
	c.Radius *= factor
}

func (c *Circle) Move(dx, dy float64) {
	c.X += dx
	c.Y += dy
}

type Triangle struct {
	Base, Height float64
	X, Y         float64
}

func (t Triangle) Area() float64 {
	return 0.5 * t.Base * t.Height
}

func (t Triangle) Perimeter() float64 {
	// 简化计算，假设是等腰三角形
	side := math.Sqrt((t.Base/2)*(t.Base/2) + t.Height*t.Height)
	return t.Base + 2*side
}

func (t Triangle) String() string {
	return fmt.Sprintf("Triangle(%.1f×%.1f) at (%.1f,%.1f)",
		t.Base, t.Height, t.X, t.Y)
}

func (t Triangle) Draw() {
	fmt.Printf("  绘制三角形: %s\n", t.String())
}

func (t *Triangle) Scale(factor float64) {
	t.Base *= factor
	t.Height *= factor
}

func (t *Triangle) Move(dx, dy float64) {
	t.X += dx
	t.Y += dy
}

// 基本多态概念
func demonstrateBasicPolymorphism() {
	fmt.Println("1. 基本多态概念:")

	// 创建不同类型的形状
	shapes := []Shape{
		Rectangle{Width: 10, Height: 5, X: 0, Y: 0},
		Circle{Radius: 3, X: 10, Y: 10},
		Triangle{Base: 6, Height: 4, X: 20, Y: 20},
	}

	fmt.Println("多态调用相同方法:")
	for i, shape := range shapes {
		fmt.Printf("形状%d: %s\n", i+1, shape)
		fmt.Printf("  面积: %.2f\n", shape.Area())
		fmt.Printf("  周长: %.2f\n", shape.Perimeter())
		fmt.Println()
	}

	// 统计信息
	totalArea := 0.0
	totalPerimeter := 0.0
	for _, shape := range shapes {
		totalArea += shape.Area()
		totalPerimeter += shape.Perimeter()
	}

	fmt.Printf("总面积: %.2f\n", totalArea)
	fmt.Printf("总周长: %.2f\n", totalPerimeter)

	// 查找最大形状
	maxShape := findLargestShape(shapes)
	fmt.Printf("最大形状: %s (面积: %.2f)\n", maxShape, maxShape.Area())

	fmt.Println()
}

func findLargestShape(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}

	largest := shapes[0]
	for _, shape := range shapes[1:] {
		if shape.Area() > largest.Area() {
			largest = shape
		}
	}
	return largest
}

// 接口多态
func demonstrateInterfacePolymorphism() {
	fmt.Println("2. 接口多态:")

	// 可绘制的形状
	drawableShapes := []Drawable{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 5, Height: 4},
	}

	fmt.Println("绘制所有形状:")
	drawAll(drawableShapes)

	// 可变换的形状
	fmt.Println("\n变换形状:")
	transformableShapes := []Transformable{
		&Rectangle{Width: 4, Height: 3, X: 0, Y: 0},
		&Circle{Radius: 2, X: 5, Y: 5},
		&Triangle{Base: 3, Height: 4, X: 10, Y: 10},
	}

	fmt.Println("原始位置:")
	for i, shape := range transformableShapes {
		if s, ok := shape.(Shape); ok {
			fmt.Printf("  形状%d: %s\n", i+1, s)
		}
	}

	// 移动所有形状
	moveAll(transformableShapes, 5, 3)

	fmt.Println("移动后位置:")
	for i, shape := range transformableShapes {
		if s, ok := shape.(Shape); ok {
			fmt.Printf("  形状%d: %s\n", i+1, s)
		}
	}

	// 缩放所有形状
	scaleAll(transformableShapes, 1.5)

	fmt.Println("缩放后:")
	for i, shape := range transformableShapes {
		if s, ok := shape.(Shape); ok {
			fmt.Printf("  形状%d: %s (面积: %.2f)\n", i+1, s, s.Area())
		}
	}

	fmt.Println()
}

func drawAll(shapes []Drawable) {
	for _, shape := range shapes {
		shape.Draw()
	}
}

func moveAll(shapes []Transformable, dx, dy float64) {
	for _, shape := range shapes {
		shape.Move(dx, dy)
	}
}

func scaleAll(shapes []Transformable, factor float64) {
	for _, shape := range shapes {
		shape.Scale(factor)
	}
}

// 多态中的类型断言
func demonstratePolymorphicTypeAssertion() {
	fmt.Println("3. 多态中的类型断言:")

	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 6, Height: 4},
	}

	fmt.Println("类型特定的操作:")
	for i, shape := range shapes {
		fmt.Printf("形状%d (%T): ", i+1, shape)

		switch s := shape.(type) {
		case Rectangle:
			fmt.Printf("矩形 - 对角线长度: %.2f\n",
				math.Sqrt(s.Width*s.Width+s.Height*s.Height))
		case Circle:
			fmt.Printf("圆形 - 直径: %.2f\n", 2*s.Radius)
		case Triangle:
			fmt.Printf("三角形 - 底边与高的比例: %.2f\n", s.Base/s.Height)
		default:
			fmt.Printf("未知形状类型\n")
		}
	}

	// 安全的类型断言
	fmt.Println("\n安全的类型断言:")
	for i, shape := range shapes {
		if rect, ok := shape.(Rectangle); ok {
			fmt.Printf("形状%d是矩形: 宽高比 %.2f\n", i+1, rect.Width/rect.Height)
		} else if circle, ok := shape.(Circle); ok {
			fmt.Printf("形状%d是圆形: 面积密度 %.2f\n", i+1,
				circle.Area()/(4*circle.Radius*circle.Radius))
		}
	}

	// 接口查询
	fmt.Println("\n接口查询:")
	queryInterfaces(shapes)

	fmt.Println()
}

func queryInterfaces(shapes []Shape) {
	for i, shape := range shapes {
		fmt.Printf("形状%d (%T) 实现的接口: ", i+1, shape)

		interfaces := []string{}

		if _, ok := shape.(Shape); ok {
			interfaces = append(interfaces, "Shape")
		}
		if _, ok := shape.(Drawable); ok {
			interfaces = append(interfaces, "Drawable")
		}
		if _, ok := shape.(fmt.Stringer); ok {
			interfaces = append(interfaces, "Stringer")
		}

		fmt.Println(strings.Join(interfaces, ", "))
	}
}

// 多态设计模式
func demonstratePolymorphicPatterns() {
	fmt.Println("4. 多态设计模式:")

	// 1. 策略模式
	fmt.Println("策略模式:")
	calculator := &AreaCalculator{}

	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
	}

	// 使用不同的计算策略
	calculator.SetStrategy(&SimpleAreaStrategy{})
	fmt.Printf("简单策略总面积: %.2f\n", calculator.Calculate(shapes))

	calculator.SetStrategy(&WeightedAreaStrategy{})
	fmt.Printf("加权策略总面积: %.2f\n", calculator.Calculate(shapes))

	// 2. 观察者模式
	fmt.Println("\n观察者模式:")
	subject := &ShapeSubject{}

	// 添加观察者
	subject.AddObserver(&AreaObserver{})
	subject.AddObserver(&PerimeterObserver{})
	subject.AddObserver(&DrawingObserver{})

	// 通知观察者
	subject.NotifyAdd(Rectangle{Width: 5, Height: 4})
	subject.NotifyAdd(Circle{Radius: 3})

	// 3. 访问者模式
	fmt.Println("\n访问者模式:")
	visitors := []ShapeVisitor{
		&AreaVisitor{},
		&InfoVisitor{},
		&ValidationVisitor{},
	}

	testShapes := []Shape{
		Rectangle{Width: 6, Height: 4},
		Circle{Radius: 2.5},
		Triangle{Base: 4, Height: 3},
	}

	for _, visitor := range visitors {
		fmt.Printf("%T 访问结果:\n", visitor)
		for _, shape := range testShapes {
			acceptVisitor(shape, visitor)
		}
		fmt.Println()
	}

	// 4. 工厂模式
	fmt.Println("工厂模式:")
	factory := &ShapeFactory{}

	createdShapes := []Shape{
		factory.CreateShape("rectangle", 5, 3),
		factory.CreateShape("circle", 2.5, 0),
		factory.CreateShape("triangle", 4, 6),
	}

	for i, shape := range createdShapes {
		fmt.Printf("工厂形状%d: %s (面积: %.2f)\n",
			i+1, shape, shape.Area())
	}

	fmt.Println()
}

// 策略模式
type AreaCalculationStrategy interface {
	Calculate(shapes []Shape) float64
}

type AreaCalculator struct {
	strategy AreaCalculationStrategy
}

func (ac *AreaCalculator) SetStrategy(strategy AreaCalculationStrategy) {
	ac.strategy = strategy
}

func (ac *AreaCalculator) Calculate(shapes []Shape) float64 {
	if ac.strategy == nil {
		return 0
	}
	return ac.strategy.Calculate(shapes)
}

type SimpleAreaStrategy struct{}

func (sas *SimpleAreaStrategy) Calculate(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		total += shape.Area()
	}
	return total
}

type WeightedAreaStrategy struct{}

func (was *WeightedAreaStrategy) Calculate(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		weight := 1.0
		switch shape.(type) {
		case Circle:
			weight = 1.5 // 圆形权重更高
		case Rectangle:
			weight = 1.2
		}
		total += shape.Area() * weight
	}
	return total
}

// 观察者模式
type ShapeObserver interface {
	OnShapeAdded(shape Shape)
}

type ShapeSubject struct {
	observers []ShapeObserver
}

func (ss *ShapeSubject) AddObserver(observer ShapeObserver) {
	ss.observers = append(ss.observers, observer)
}

func (ss *ShapeSubject) NotifyAdd(shape Shape) {
	for _, observer := range ss.observers {
		observer.OnShapeAdded(shape)
	}
}

type AreaObserver struct{}

func (ao *AreaObserver) OnShapeAdded(shape Shape) {
	fmt.Printf("  [面积观察者] 新增形状面积: %.2f\n", shape.Area())
}

type PerimeterObserver struct{}

func (po *PerimeterObserver) OnShapeAdded(shape Shape) {
	fmt.Printf("  [周长观察者] 新增形状周长: %.2f\n", shape.Perimeter())
}

type DrawingObserver struct{}

func (do *DrawingObserver) OnShapeAdded(shape Shape) {
	fmt.Printf("  [绘制观察者] 新增形状: %s\n", shape)
}

// 访问者模式
type ShapeVisitor interface {
	VisitRectangle(r Rectangle)
	VisitCircle(c Circle)
	VisitTriangle(t Triangle)
}

func acceptVisitor(shape Shape, visitor ShapeVisitor) {
	switch s := shape.(type) {
	case Rectangle:
		visitor.VisitRectangle(s)
	case Circle:
		visitor.VisitCircle(s)
	case Triangle:
		visitor.VisitTriangle(s)
	}
}

type AreaVisitor struct{}

func (av *AreaVisitor) VisitRectangle(r Rectangle) {
	fmt.Printf("  矩形面积: %.2f\n", r.Area())
}

func (av *AreaVisitor) VisitCircle(c Circle) {
	fmt.Printf("  圆形面积: %.2f\n", c.Area())
}

func (av *AreaVisitor) VisitTriangle(t Triangle) {
	fmt.Printf("  三角形面积: %.2f\n", t.Area())
}

type InfoVisitor struct{}

func (iv *InfoVisitor) VisitRectangle(r Rectangle) {
	fmt.Printf("  矩形信息: %s\n", r.String())
}

func (iv *InfoVisitor) VisitCircle(c Circle) {
	fmt.Printf("  圆形信息: %s\n", c.String())
}

func (iv *InfoVisitor) VisitTriangle(t Triangle) {
	fmt.Printf("  三角形信息: %s\n", t.String())
}

type ValidationVisitor struct{}

func (vv *ValidationVisitor) VisitRectangle(r Rectangle) {
	valid := r.Width > 0 && r.Height > 0
	fmt.Printf("  矩形有效性: %t\n", valid)
}

func (vv *ValidationVisitor) VisitCircle(c Circle) {
	valid := c.Radius > 0
	fmt.Printf("  圆形有效性: %t\n", valid)
}

func (vv *ValidationVisitor) VisitTriangle(t Triangle) {
	valid := t.Base > 0 && t.Height > 0
	fmt.Printf("  三角形有效性: %t\n", valid)
}

// 工厂模式
type ShapeFactory struct{}

func (sf *ShapeFactory) CreateShape(shapeType string, param1, param2 float64) Shape {
	switch shapeType {
	case "rectangle":
		return Rectangle{Width: param1, Height: param2}
	case "circle":
		return Circle{Radius: param1}
	case "triangle":
		return Triangle{Base: param1, Height: param2}
	default:
		return Rectangle{Width: 1, Height: 1} // 默认形状
	}
}

// 组合vs继承
func demonstrateCompositionVsInheritance() {
	fmt.Println("5. 组合vs继承:")

	// Go通过组合实现类似继承的效果
	fmt.Println("通过嵌入实现组合:")

	// 基础图形
	colored := ColoredShape{
		Shape: Rectangle{Width: 4, Height: 3},
		Color: "红色",
	}

	border := BorderedShape{
		Shape:       Circle{Radius: 2},
		BorderWidth: 2.0,
		BorderColor: "黑色",
	}

	fmt.Printf("彩色形状: %s\n", colored)
	fmt.Printf("边框形状: %s\n", border)

	// 多重组合
	coloredBordered := ColoredBorderedShape{
		ColoredShape: ColoredShape{
			Shape: Triangle{Base: 5, Height: 4},
			Color: "蓝色",
		},
		BorderWidth: 1.5,
		BorderColor: "绿色",
	}

	fmt.Printf("彩色边框形状: %s\n", coloredBordered)
	fmt.Printf("总面积(含边框): %.2f\n", coloredBordered.TotalArea())

	// 组合的灵活性
	fmt.Println("\n组合的灵活性:")
	shapes := []Shape{
		Rectangle{Width: 3, Height: 2},
		ColoredShape{Shape: Circle{Radius: 1.5}, Color: "黄色"},
		BorderedShape{Shape: Triangle{Base: 4, Height: 3}, BorderWidth: 1},
	}

	for i, shape := range shapes {
		fmt.Printf("组合形状%d: %s (面积: %.2f)\n",
			i+1, shape, shape.Area())
	}

	fmt.Println()
}

// 组合类型
type ColoredShape struct {
	Shape
	Color string
}

func (cs ColoredShape) String() string {
	return fmt.Sprintf("%s的%s", cs.Color, cs.Shape.String())
}

type BorderedShape struct {
	Shape
	BorderWidth float64
	BorderColor string
}

func (bs BorderedShape) String() string {
	return fmt.Sprintf("%s (边框: %.1f %s)",
		bs.Shape.String(), bs.BorderWidth, bs.BorderColor)
}

func (bs BorderedShape) TotalArea() float64 {
	return bs.Shape.Area() + bs.BorderWidth*bs.Perimeter()
}

type ColoredBorderedShape struct {
	ColoredShape
	BorderWidth float64
	BorderColor string
}

func (cbs ColoredBorderedShape) String() string {
	return fmt.Sprintf("%s (边框: %.1f %s)",
		cbs.ColoredShape.String(), cbs.BorderWidth, cbs.BorderColor)
}

func (cbs ColoredBorderedShape) TotalArea() float64 {
	return cbs.Shape.Area() + cbs.BorderWidth*cbs.Perimeter()
}

// 多态的动态特性
func demonstrateDynamicPolymorphism() {
	fmt.Println("6. 多态的动态特性:")

	// 运行时多态
	fmt.Println("运行时形状创建:")
	shapeTypes := []string{"rectangle", "circle", "triangle"}
	factory := &ShapeFactory{}

	var shapes []Shape
	for i, shapeType := range shapeTypes {
		shape := factory.CreateShape(shapeType, float64(i+2), float64(i+3))
		shapes = append(shapes, shape)
		fmt.Printf("创建 %s: %s\n", shapeType, shape)
	}

	// 动态接口查询
	fmt.Println("\n动态接口查询:")
	for i, shape := range shapes {
		fmt.Printf("形状%d 接口支持:\n", i+1)

		if drawable, ok := shape.(Drawable); ok {
			fmt.Println("  支持绘制")
			drawable.Draw()
		}

		if stringer, ok := shape.(fmt.Stringer); ok {
			fmt.Printf("  支持字符串化: %s\n", stringer.String())
		}
	}

	// 动态方法调用
	fmt.Println("\n动态方法调用:")
	processor := &ShapeProcessor{}
	results := processor.ProcessShapes(shapes, []string{"area", "perimeter", "info"})

	for operation, values := range results {
		fmt.Printf("%s 结果: %v\n", operation, values)
	}

	fmt.Println()
}

type ShapeProcessor struct{}

func (sp *ShapeProcessor) ProcessShapes(shapes []Shape, operations []string) map[string][]float64 {
	results := make(map[string][]float64)

	for _, operation := range operations {
		var values []float64
		for _, shape := range shapes {
			switch operation {
			case "area":
				values = append(values, shape.Area())
			case "perimeter":
				values = append(values, shape.Perimeter())
			case "info":
				values = append(values, float64(len(shape.String())))
			}
		}
		results[operation] = values
	}

	return results
}

// 多态的最佳实践
func demonstrateBestPractices() {
	fmt.Println("7. 多态的最佳实践:")

	// 1. 接受接口，返回具体类型
	fmt.Println("接受接口，返回具体类型:")
	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
	}

	transformed := transformShapes(shapes, 2.0)
	for i, shape := range transformed {
		fmt.Printf("变换后形状%d: %s\n", i+1, shape)
	}

	// 2. 小接口原则
	fmt.Println("\n小接口原则:")
	sizers := []Sizer{
		Rectangle{Width: 3, Height: 4},
		Circle{Radius: 2},
	}

	totalSize := calculateTotalSize(sizers)
	fmt.Printf("总大小: %.2f\n", totalSize)

	// 3. 零值接口的处理
	fmt.Println("\n零值接口处理:")
	var nilShape Shape
	safeProcess(nilShape)
	safeProcess(Rectangle{Width: 2, Height: 3})

	// 4. 接口组合
	fmt.Println("\n接口组合:")
	drawable := DrawableTransformable(&Rectangle{Width: 3, Height: 4})
	processDrawableTransformable(drawable)

	fmt.Println()
}

type Sizer interface {
	Area() float64
}

func transformShapes(shapes []Shape, factor float64) []Rectangle {
	var results []Rectangle
	for _, shape := range shapes {
		// 转换为标准矩形
		area := shape.Area()
		side := math.Sqrt(area * factor)
		results = append(results, Rectangle{Width: side, Height: side})
	}
	return results
}

func calculateTotalSize(sizers []Sizer) float64 {
	total := 0.0
	for _, sizer := range sizers {
		total += sizer.Area()
	}
	return total
}

func safeProcess(shape Shape) {
	if shape == nil {
		fmt.Println("  处理nil形状: 跳过")
		return
	}
	fmt.Printf("  处理形状: %s (面积: %.2f)\n", shape, shape.Area())
}

type DrawableTransformable interface {
	Drawable
	Transformable
}

func processDrawableTransformable(dt DrawableTransformable) {
	dt.Draw()
	dt.Scale(1.5)
	dt.Move(5, 5)
	if s, ok := dt.(Shape); ok {
		fmt.Printf("  处理后: %s\n", s)
	}
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 图形编辑器
	fmt.Println("图形编辑器:")
	editor := NewGraphicsEditor()

	editor.AddShape(Rectangle{Width: 10, Height: 5})
	editor.AddShape(Circle{Radius: 3})
	editor.AddShape(Triangle{Base: 6, Height: 4})

	fmt.Printf("画布总面积: %.2f\n", editor.GetTotalArea())
	fmt.Println("渲染画布:")
	editor.Render()

	// 2. 动画系统
	fmt.Println("\n动画系统:")
	animator := NewAnimator()

	animatable1 := &AnimatableRectangle{Rectangle: Rectangle{Width: 4, Height: 3}}
	animatable2 := &AnimatableCircle{Circle: Circle{Radius: 2}}

	animator.AddAnimatable(animatable1)
	animator.AddAnimatable(animatable2)

	fmt.Println("播放动画:")
	animator.PlayAnimation("move", 3.0)

	// 3. 序列化系统
	fmt.Println("\n序列化系统:")
	serializer := NewShapeSerializer()

	shapes := []Shape{
		Rectangle{Width: 5, Height: 3},
		Circle{Radius: 2.5},
	}

	for i, shape := range shapes {
		data := serializer.Serialize(shape)
		fmt.Printf("形状%d序列化: %s\n", i+1, data)

		deserialized := serializer.Deserialize(data)
		fmt.Printf("反序列化: %s\n", deserialized)
	}

	// 4. 插件系统
	fmt.Println("\n插件系统:")
	pluginManager := NewShapePluginManager()

	pluginManager.RegisterPlugin("area_calculator", &AreaCalculatorPlugin{})
	pluginManager.RegisterPlugin("validator", &ValidationPlugin{})
	pluginManager.RegisterPlugin("transformer", &TransformPlugin{})

	testShape := Rectangle{Width: 4, Height: 6}
	pluginManager.ProcessShape("area_calculator", testShape)
	pluginManager.ProcessShape("validator", testShape)
	pluginManager.ProcessShape("transformer", testShape)

	fmt.Println()
}

// 图形编辑器
type GraphicsEditor struct {
	shapes []Shape
}

func NewGraphicsEditor() *GraphicsEditor {
	return &GraphicsEditor{shapes: make([]Shape, 0)}
}

func (ge *GraphicsEditor) AddShape(shape Shape) {
	ge.shapes = append(ge.shapes, shape)
}

func (ge *GraphicsEditor) GetTotalArea() float64 {
	total := 0.0
	for _, shape := range ge.shapes {
		total += shape.Area()
	}
	return total
}

func (ge *GraphicsEditor) Render() {
	for i, shape := range ge.shapes {
		fmt.Printf("  [%d] 渲染: %s\n", i+1, shape)
		if drawable, ok := shape.(Drawable); ok {
			drawable.Draw()
		}
	}
}

// 动画系统
type Animatable interface {
	Animate(animation string, duration float64)
}

type Animator struct {
	animatables []Animatable
}

func NewAnimator() *Animator {
	return &Animator{animatables: make([]Animatable, 0)}
}

func (a *Animator) AddAnimatable(animatable Animatable) {
	a.animatables = append(a.animatables, animatable)
}

func (a *Animator) PlayAnimation(animation string, duration float64) {
	for i, animatable := range a.animatables {
		fmt.Printf("  [%d] ", i+1)
		animatable.Animate(animation, duration)
	}
}

type AnimatableRectangle struct {
	Rectangle
}

func (ar *AnimatableRectangle) Animate(animation string, duration float64) {
	fmt.Printf("矩形动画 '%s' (%.1f秒): %s\n", animation, duration, ar.String())
}

type AnimatableCircle struct {
	Circle
}

func (ac *AnimatableCircle) Animate(animation string, duration float64) {
	fmt.Printf("圆形动画 '%s' (%.1f秒): %s\n", animation, duration, ac.String())
}

// 序列化系统
type ShapeSerializer struct{}

func NewShapeSerializer() *ShapeSerializer {
	return &ShapeSerializer{}
}

func (ss *ShapeSerializer) Serialize(shape Shape) string {
	switch s := shape.(type) {
	case Rectangle:
		return fmt.Sprintf("rect:%.1f:%.1f", s.Width, s.Height)
	case Circle:
		return fmt.Sprintf("circle:%.1f", s.Radius)
	case Triangle:
		return fmt.Sprintf("triangle:%.1f:%.1f", s.Base, s.Height)
	default:
		return "unknown"
	}
}

func (ss *ShapeSerializer) Deserialize(data string) Shape {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return Rectangle{Width: 1, Height: 1}
	}

	switch parts[0] {
	case "rect":
		if len(parts) >= 3 {
			width, _ := strconv.ParseFloat(parts[1], 64)
			height, _ := strconv.ParseFloat(parts[2], 64)
			return Rectangle{Width: width, Height: height}
		}
	case "circle":
		radius, _ := strconv.ParseFloat(parts[1], 64)
		return Circle{Radius: radius}
	case "triangle":
		if len(parts) >= 3 {
			base, _ := strconv.ParseFloat(parts[1], 64)
			height, _ := strconv.ParseFloat(parts[2], 64)
			return Triangle{Base: base, Height: height}
		}
	}

	return Rectangle{Width: 1, Height: 1}
}

// 插件系统
type ShapePlugin interface {
	Process(shape Shape)
	Name() string
}

type ShapePluginManager struct {
	plugins map[string]ShapePlugin
}

func NewShapePluginManager() *ShapePluginManager {
	return &ShapePluginManager{plugins: make(map[string]ShapePlugin)}
}

func (spm *ShapePluginManager) RegisterPlugin(name string, plugin ShapePlugin) {
	spm.plugins[name] = plugin
}

func (spm *ShapePluginManager) ProcessShape(pluginName string, shape Shape) {
	if plugin, exists := spm.plugins[pluginName]; exists {
		fmt.Printf("  使用插件 '%s':", plugin.Name())
		plugin.Process(shape)
	} else {
		fmt.Printf("  插件 '%s' 未找到\n", pluginName)
	}
}

type AreaCalculatorPlugin struct{}

func (acp *AreaCalculatorPlugin) Name() string {
	return "面积计算器"
}

func (acp *AreaCalculatorPlugin) Process(shape Shape) {
	fmt.Printf(" 面积=%.2f\n", shape.Area())
}

type ValidationPlugin struct{}

func (vp *ValidationPlugin) Name() string {
	return "形状验证器"
}

func (vp *ValidationPlugin) Process(shape Shape) {
	valid := shape.Area() > 0 && shape.Perimeter() > 0
	fmt.Printf(" 有效性=%t\n", valid)
}

type TransformPlugin struct{}

func (tp *TransformPlugin) Name() string {
	return "形状变换器"
}

func (tp *TransformPlugin) Process(shape Shape) {
	fmt.Printf(" 变换信息=%s\n", shape.String())
}

/*
=== 练习题 ===

1. 实现一个媒体播放器，支持不同格式的音频和视频

2. 创建一个绘图程序，支持多种绘图工具和形状

3. 设计一个游戏引擎，支持不同类型的游戏对象

4. 实现一个数据处理管道，支持多种数据源和处理器

5. 创建一个通知系统，支持多种通知方式

6. 设计一个文件系统抽象，支持本地和远程文件操作

7. 实现一个支付系统，支持多种支付方式

运行命令：
go run main.go

高级练习：
1. 实现一个可扩展的Web框架
2. 创建一个插件化的数据库ORM
3. 设计一个分布式系统的服务抽象
4. 实现一个通用的事件处理系统
5. 创建一个多协议的网络客户端

重要概念：
- 多态通过接口实现
- 运行时动态分发
- 组合优于继承
- 接口隔离原则
- 开放封闭原则
- 依赖倒置原则
*/

// （已将 strconv 导入移动到顶部的 import 块）

/*
=== 练习题 ===

1. 实现一个媒体播放器，支持不同格式的音频和视频

2. 创建一个绘图程序，支持多种绘图工具和形状

3. 设计一个游戏引擎，支持不同类型的游戏对象

4. 实现一个数据处理管道，支持多种数据源和处理器

5. 创建一个通知系统，支持多种通知方式

6. 设计一个文件系统抽象，支持本地和远程文件操作

7. 实现一个支付系统，支持多种支付方式

运行命令：
go run main.go

高级练习：
1. 实现一个可扩展的Web框架
2. 创建一个插件化的数据库ORM
3. 设计一个分布式系统的服务抽象
4. 实现一个通用的事件处理系统
5. 创建一个多协议的网络客户端

重要概念：
- 多态通过接口实现
- 运行时动态分发
- 组合优于继承
- 接口隔离原则
- 开放封闭原则
- 依赖倒置原则
*/
