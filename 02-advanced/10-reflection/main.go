package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// 1. 反射基础概念
// =============================================================================

/*
反射（Reflection）是程序在运行时检查、修改和操作对象的能力。

Go 语言反射的核心：
1. reflect.Type: 表示类型信息
2. reflect.Value: 表示值信息
3. reflect.TypeOf(): 获取类型信息
4. reflect.ValueOf(): 获取值信息

反射的主要用途：
1. 类型检查和类型断言
2. 动态调用方法
3. 结构体字段的检查和修改
4. 实现通用的序列化/反序列化
5. 实现ORM映射
6. 依赖注入
7. 配置解析

反射的缺点：
1. 性能开销较大
2. 类型安全性降低
3. 代码可读性变差
4. 编译时错误变为运行时错误
*/

// =============================================================================
// 2. 基本类型反射
// =============================================================================

func demonstrateBasicReflection() {
	fmt.Println("=== 1. 基本类型反射 ===")

	// 基本数据类型反射
	var x int = 42
	var y string = "Hello, Go!"
	var z float64 = 3.14
	var b bool = true

	values := []interface{}{x, y, z, b}
	names := []string{"整数", "字符串", "浮点数", "布尔值"}

	for i, value := range values {
		t := reflect.TypeOf(value)
		v := reflect.ValueOf(value)

		fmt.Printf("%s - 类型: %v, 种类: %v, 值: %v\n",
			names[i], t, t.Kind(), v)
	}

	// 指针反射
	var ptr *int = &x
	ptrType := reflect.TypeOf(ptr)
	ptrValue := reflect.ValueOf(ptr)

	fmt.Printf("指针 - 类型: %v, 种类: %v, 元素类型: %v\n",
		ptrType, ptrType.Kind(), ptrType.Elem())
	fmt.Printf("指针值: %v, 指向的值: %v\n",
		ptrValue, ptrValue.Elem())

	fmt.Println()
}

// =============================================================================
// 3. 结构体反射
// =============================================================================

// Student 学生结构体（用于反射演示）
type Student struct {
	ID       int    `json:"id" db:"student_id" validate:"required"`
	Name     string `json:"name" db:"student_name" validate:"required,min=2"`
	Age      int    `json:"age" db:"age" validate:"min=0,max=150"`
	Email    string `json:"email" db:"email" validate:"email"`
	Address  string `json:"address" db:"address"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

// GetInfo 获取学生信息
func (s Student) GetInfo() string {
	return fmt.Sprintf("学生：%s，年龄：%d，邮箱：%s", s.Name, s.Age, s.Email)
}

// UpdateAge 更新年龄（需要指针接收者才能修改）
func (s *Student) UpdateAge(newAge int) {
	s.Age = newAge
}

// GetFullInfo 获取完整信息（私有方法）
func (s Student) getFullInfo() string {
	return fmt.Sprintf("ID: %d, Name: %s, Age: %d, Email: %s, Address: %s, Active: %v",
		s.ID, s.Name, s.Age, s.Email, s.Address, s.IsActive)
}

func demonstrateStructReflection() {
	fmt.Println("=== 2. 结构体反射 ===")

	student := Student{
		ID:       1,
		Name:     "张三",
		Age:      20,
		Email:    "zhangsan@example.com",
		Address:  "北京市",
		IsActive: true,
	}

	t := reflect.TypeOf(student)
	v := reflect.ValueOf(student)

	fmt.Printf("结构体类型: %v, 种类: %v\n", t, t.Kind())
	fmt.Printf("字段数量: %d\n", t.NumField())

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		fmt.Printf("字段 %d: %s, 类型: %v, 值: %v\n",
			i, field.Name, field.Type, value)

		// 解析结构体标签
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fmt.Printf("  JSON标签: %s\n", jsonTag)
		}
		if dbTag := field.Tag.Get("db"); dbTag != "" {
			fmt.Printf("  数据库标签: %s\n", dbTag)
		}
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			fmt.Printf("  验证标签: %s\n", validateTag)
		}
	}

	fmt.Println()
}

// =============================================================================
// 4. 方法反射和动态调用
// =============================================================================

func demonstrateMethodReflection() {
	fmt.Println("=== 3. 方法反射和动态调用 ===")

	student := Student{
		ID:    1,
		Name:  "李四",
		Age:   22,
		Email: "lisi@example.com",
	}

	// 值反射（值接收者方法）
	v := reflect.ValueOf(student)
	t := reflect.TypeOf(student)

	fmt.Printf("方法数量: %d\n", t.NumMethod())

	// 遍历所有方法
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		fmt.Printf("方法 %d: %s, 类型: %v\n", i, method.Name, method.Type)

		// 动态调用无参数方法
		if method.Name == "GetInfo" {
			results := v.MethodByName("GetInfo").Call(nil)
			fmt.Printf("调用 GetInfo() 结果: %s\n", results[0])
		}
	}

	// 指针反射（指针接收者方法）
	studentPtr := &student
	ptrV := reflect.ValueOf(studentPtr)
	ptrT := reflect.TypeOf(studentPtr)

	fmt.Printf("指针方法数量: %d\n", ptrT.NumMethod())

	// 动态调用带参数的方法
	updateAgeMethod := ptrV.MethodByName("UpdateAge")
	if updateAgeMethod.IsValid() {
		args := []reflect.Value{reflect.ValueOf(25)}
		updateAgeMethod.Call(args)
		fmt.Printf("更新年龄后: %s\n", studentPtr.GetInfo())
	}

	fmt.Println()
}

// =============================================================================
// 5. 切片和映射反射
// =============================================================================

func demonstrateSliceAndMapReflection() {
	fmt.Println("=== 4. 切片和映射反射 ===")

	// 切片反射
	numbers := []int{1, 2, 3, 4, 5}
	sliceV := reflect.ValueOf(numbers)
	sliceT := reflect.TypeOf(numbers)

	fmt.Printf("切片类型: %v, 种类: %v, 元素类型: %v\n",
		sliceT, sliceT.Kind(), sliceT.Elem())
	fmt.Printf("切片长度: %d, 容量: %d\n", sliceV.Len(), sliceV.Cap())

	// 遍历切片元素
	fmt.Print("切片元素: ")
	for i := 0; i < sliceV.Len(); i++ {
		fmt.Printf("%v ", sliceV.Index(i))
	}
	fmt.Println()

	// 动态创建切片
	newSliceV := reflect.MakeSlice(sliceT, 3, 5)
	for i := 0; i < newSliceV.Len(); i++ {
		newSliceV.Index(i).Set(reflect.ValueOf(i * 10))
	}
	fmt.Printf("动态创建的切片: %v\n", newSliceV.Interface())

	// 映射反射
	userMap := map[string]int{
		"Alice": 25,
		"Bob":   30,
		"Carol": 28,
	}

	mapV := reflect.ValueOf(userMap)
	mapT := reflect.TypeOf(userMap)

	fmt.Printf("映射类型: %v, 种类: %v\n", mapT, mapT.Kind())
	fmt.Printf("键类型: %v, 值类型: %v\n", mapT.Key(), mapT.Elem())

	// 遍历映射
	fmt.Println("映射内容:")
	for _, key := range mapV.MapKeys() {
		value := mapV.MapIndex(key)
		fmt.Printf("  %v: %v\n", key, value)
	}

	// 动态创建映射
	newMapV := reflect.MakeMap(mapT)
	newMapV.SetMapIndex(reflect.ValueOf("David"), reflect.ValueOf(35))
	newMapV.SetMapIndex(reflect.ValueOf("Eve"), reflect.ValueOf(27))
	fmt.Printf("动态创建的映射: %v\n", newMapV.Interface())

	fmt.Println()
}

// =============================================================================
// 6. 接口反射
// =============================================================================

// Shape 接口
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle 矩形
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Circle 圆形
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * 3.14159 * c.Radius
}

func demonstrateInterfaceReflection() {
	fmt.Println("=== 5. 接口反射 ===")

	shapes := []Shape{
		Rectangle{Width: 10, Height: 5},
		Circle{Radius: 3},
	}

	for i, shape := range shapes {
		v := reflect.ValueOf(shape)
		t := reflect.TypeOf(shape)

		fmt.Printf("形状 %d:\n", i+1)
		fmt.Printf("  类型: %v, 种类: %v\n", t, t.Kind())

		// 检查是否实现了接口
		shapeInterfaceType := reflect.TypeOf((*Shape)(nil)).Elem()
		if t.Implements(shapeInterfaceType) {
			fmt.Printf("  实现了 Shape 接口\n")
		}

		// 动态调用接口方法
		areaMethod := v.MethodByName("Area")
		if areaMethod.IsValid() {
			results := areaMethod.Call(nil)
			fmt.Printf("  面积: %.2f\n", results[0].Float())
		}

		perimeterMethod := v.MethodByName("Perimeter")
		if perimeterMethod.IsValid() {
			results := perimeterMethod.Call(nil)
			fmt.Printf("  周长: %.2f\n", results[0].Float())
		}

		// 获取底层具体类型
		if t == reflect.TypeOf(Rectangle{}) {
			rect := shape.(Rectangle)
			fmt.Printf("  矩形详情: 宽=%.1f, 高=%.1f\n", rect.Width, rect.Height)
		} else if t == reflect.TypeOf(Circle{}) {
			circle := shape.(Circle)
			fmt.Printf("  圆形详情: 半径=%.1f\n", circle.Radius)
		}
	}

	fmt.Println()
}

// =============================================================================
// 7. 修改值（可寻址性）
// =============================================================================

func demonstrateValueModification() {
	fmt.Println("=== 6. 值的修改（可寻址性） ===")

	// 基本类型修改
	x := 42
	fmt.Printf("原始值: %d\n", x)

	// 直接值是不可寻址的
	v := reflect.ValueOf(x)
	fmt.Printf("是否可寻址: %v, 是否可设置: %v\n", v.CanAddr(), v.CanSet())

	// 指针值是可寻址的
	ptrV := reflect.ValueOf(&x)
	elemV := ptrV.Elem()
	fmt.Printf("指针元素是否可寻址: %v, 是否可设置: %v\n", elemV.CanAddr(), elemV.CanSet())

	// 修改值
	elemV.SetInt(100)
	fmt.Printf("修改后的值: %d\n", x)

	// 结构体字段修改
	student := Student{
		ID:   1,
		Name: "原始姓名",
		Age:  20,
	}

	fmt.Printf("原始学生: %+v\n", student)

	studentPtrV := reflect.ValueOf(&student)
	studentV := studentPtrV.Elem()

	// 修改字段
	nameField := studentV.FieldByName("Name")
	if nameField.CanSet() {
		nameField.SetString("新姓名")
	}

	ageField := studentV.FieldByName("Age")
	if ageField.CanSet() {
		ageField.SetInt(25)
	}

	fmt.Printf("修改后学生: %+v\n", student)

	fmt.Println()
}

// =============================================================================
// 8. 实际应用：JSON 序列化器
// =============================================================================

// SimpleJSONSerializer 简单的JSON序列化器
type SimpleJSONSerializer struct{}

// Marshal 序列化结构体为JSON字符串
func (s *SimpleJSONSerializer) Marshal(v interface{}) (string, error) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// 只处理结构体
	if typ.Kind() != reflect.Struct {
		return "", fmt.Errorf("只支持结构体类型")
	}

	var result strings.Builder
	result.WriteString("{")

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 跳过未导出的字段
		if !fieldValue.CanInterface() {
			continue
		}

		if i > 0 {
			result.WriteString(",")
		}

		// 获取JSON标签名
		jsonName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}

		result.WriteString(`"` + jsonName + `":`)

		// 根据字段类型序列化值
		switch fieldValue.Kind() {
		case reflect.String:
			result.WriteString(`"` + fieldValue.String() + `"`)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result.WriteString(strconv.FormatInt(fieldValue.Int(), 10))
		case reflect.Bool:
			result.WriteString(strconv.FormatBool(fieldValue.Bool()))
		case reflect.Float32, reflect.Float64:
			result.WriteString(strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64))
		default:
			result.WriteString(`"` + fmt.Sprintf("%v", fieldValue.Interface()) + `"`)
		}
	}

	result.WriteString("}")
	return result.String(), nil
}

func demonstrateJSONSerializer() {
	fmt.Println("=== 7. 实际应用：JSON序列化器 ===")

	serializer := &SimpleJSONSerializer{}

	student := Student{
		ID:       1,
		Name:     "张三",
		Age:      20,
		Email:    "zhangsan@example.com",
		Address:  "北京市",
		IsActive: true,
	}

	jsonStr, err := serializer.Marshal(student)
	if err != nil {
		fmt.Printf("序列化错误: %v\n", err)
	} else {
		fmt.Printf("序列化结果: %s\n", jsonStr)
	}

	fmt.Println()
}

// =============================================================================
// 9. 实际应用：结构体验证器
// =============================================================================

// Validator 结构体验证器
type Validator struct{}

// Validate 验证结构体
func (validator *Validator) Validate(v interface{}) []string {
	var errors []string

	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// 如果是指针，获取其元素
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return []string{"只支持结构体类型"}
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 获取验证标签
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// 解析验证规则
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			switch {
			case rule == "required":
				if validator.isEmpty(fieldValue) {
					errors = append(errors, fmt.Sprintf("字段 %s 是必需的", field.Name))
				}

			case strings.HasPrefix(rule, "min="):
				minStr := strings.TrimPrefix(rule, "min=")
				if min, err := strconv.Atoi(minStr); err == nil {
					if err := validator.checkMin(fieldValue, min, field.Name); err != "" {
						errors = append(errors, err)
					}
				}

			case strings.HasPrefix(rule, "max="):
				maxStr := strings.TrimPrefix(rule, "max=")
				if max, err := strconv.Atoi(maxStr); err == nil {
					if err := validator.checkMax(fieldValue, max, field.Name); err != "" {
						errors = append(errors, err)
					}
				}

			case rule == "email":
				if fieldValue.Kind() == reflect.String {
					if !strings.Contains(fieldValue.String(), "@") {
						errors = append(errors, fmt.Sprintf("字段 %s 必须是有效的邮箱地址", field.Name))
					}
				}
			}
		}
	}

	return errors
}

// isEmpty 检查字段是否为空
func (validator *Validator) isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		return false
	}
}

// checkMin 检查最小值
func (validator *Validator) checkMin(v reflect.Value, min int, fieldName string) string {
	switch v.Kind() {
	case reflect.String:
		if len(v.String()) < min {
			return fmt.Sprintf("字段 %s 长度不能少于 %d", fieldName, min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < int64(min) {
			return fmt.Sprintf("字段 %s 不能小于 %d", fieldName, min)
		}
	}
	return ""
}

// checkMax 检查最大值
func (validator *Validator) checkMax(v reflect.Value, max int, fieldName string) string {
	switch v.Kind() {
	case reflect.String:
		if len(v.String()) > max {
			return fmt.Sprintf("字段 %s 长度不能超过 %d", fieldName, max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > int64(max) {
			return fmt.Sprintf("字段 %s 不能大于 %d", fieldName, max)
		}
	}
	return ""
}

func demonstrateValidator() {
	fmt.Println("=== 8. 实际应用：结构体验证器 ===")

	validator := &Validator{}

	// 有效的学生数据
	validStudent := Student{
		ID:    1,
		Name:  "张三",
		Age:   20,
		Email: "zhangsan@example.com",
	}

	errors := validator.Validate(validStudent)
	if len(errors) == 0 {
		fmt.Println("有效学生数据验证通过")
	} else {
		fmt.Printf("验证错误: %v\n", errors)
	}

	// 无效的学生数据
	invalidStudent := Student{
		ID:    0,               // 违反 required
		Name:  "A",             // 违反 min=2
		Age:   200,             // 违反 max=150
		Email: "invalid-email", // 违反 email 格式
	}

	errors = validator.Validate(invalidStudent)
	if len(errors) > 0 {
		fmt.Println("无效学生数据验证错误:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Println()
}

// =============================================================================
// 10. 实际应用：ORM 映射器
// =============================================================================

// SimpleORM 简单的ORM映射器
type SimpleORM struct{}

// GenerateInsertSQL 生成插入SQL
func (orm *SimpleORM) GenerateInsertSQL(tableName string, v interface{}) (string, []interface{}) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// 如果是指针，获取其元素
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 获取数据库列名
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, "?")
		values = append(values, fieldValue.Interface())
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, values
}

// GenerateSelectSQL 生成查询SQL
func (orm *SimpleORM) GenerateSelectSQL(tableName string, v interface{}) string {
	typ := reflect.TypeOf(v)

	// 如果是指针，获取其元素
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var columns []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 获取数据库列名
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		columns = append(columns, dbTag)
	}

	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)
}

func demonstrateORM() {
	fmt.Println("=== 9. 实际应用：ORM映射器 ===")

	orm := &SimpleORM{}

	student := Student{
		ID:       1,
		Name:     "李四",
		Age:      22,
		Email:    "lisi@example.com",
		Address:  "上海市",
		IsActive: true,
	}

	// 生成插入SQL
	insertSQL, values := orm.GenerateInsertSQL("students", student)
	fmt.Printf("插入SQL: %s\n", insertSQL)
	fmt.Printf("参数值: %v\n", values)

	// 生成查询SQL
	selectSQL := orm.GenerateSelectSQL("students", student)
	fmt.Printf("查询SQL: %s\n", selectSQL)

	fmt.Println()
}

// =============================================================================
// 11. 性能考虑和最佳实践
// =============================================================================

func demonstrateReflectionPerformance() {
	fmt.Println("=== 10. 反射性能考虑 ===")

	student := Student{ID: 1, Name: "性能测试", Age: 25}

	// 直接访问
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		_ = student.Name
	}
	directTime := time.Since(start)

	// 反射访问
	v := reflect.ValueOf(student)
	start = time.Now()
	for i := 0; i < 1000000; i++ {
		_ = v.FieldByName("Name").String()
	}
	reflectTime := time.Since(start)

	fmt.Printf("直接访问耗时: %v\n", directTime)
	fmt.Printf("反射访问耗时: %v\n", reflectTime)
	fmt.Printf("反射性能损失: %.2fx\n", float64(reflectTime)/float64(directTime))

	fmt.Println("\n反射最佳实践:")
	fmt.Println("1. 缓存 reflect.Type 和 reflect.Value")
	fmt.Println("2. 避免在热路径中使用反射")
	fmt.Println("3. 使用类型开关替代反射（如果可能）")
	fmt.Println("4. 预先验证反射操作的有效性")
	fmt.Println("5. 考虑使用代码生成替代运行时反射")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 反射机制 - 完整示例")
	fmt.Println("========================")

	demonstrateBasicReflection()
	demonstrateStructReflection()
	demonstrateMethodReflection()
	demonstrateSliceAndMapReflection()
	demonstrateInterfaceReflection()
	demonstrateValueModification()
	demonstrateJSONSerializer()
	demonstrateValidator()
	demonstrateORM()
	demonstrateReflectionPerformance()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个通用的深拷贝函数")
	fmt.Println("2. 创建一个配置文件解析器，支持结构体标签")
	fmt.Println("3. 实现一个简单的依赖注入容器")
	fmt.Println("4. 编写一个通用的对象比较器")
	fmt.Println("5. 创建一个结构体字段映射工具")
	fmt.Println("6. 实现一个简单的表单验证框架")
	fmt.Println("\n在此文件中实现这些练习，深入理解Go反射机制！")
}
