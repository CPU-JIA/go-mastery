// Package main 演示领域特定语言（DSL）实现
// 本模块涵盖 DSL 设计的核心技术：
// - 内部 DSL vs 外部 DSL
// - 流式接口设计
// - 配置 DSL
// - 查询 DSL
package main

import (
	"fmt"
	"strings"
)

// ============================================================================
// 流式接口 DSL - SQL 查询构建器
// ============================================================================

// QueryBuilder SQL 查询构建器
type QueryBuilder struct {
	table      string
	columns    []string
	conditions []string
	orderBy    []string
	limit      int
	offset     int
	joins      []string
	groupBy    []string
	having     string
}

// NewQuery 创建新查询
func NewQuery() *QueryBuilder {
	return &QueryBuilder{
		columns: []string{"*"},
	}
}

// Select 选择列
func (q *QueryBuilder) Select(columns ...string) *QueryBuilder {
	q.columns = columns
	return q
}

// From 指定表
func (q *QueryBuilder) From(table string) *QueryBuilder {
	q.table = table
	return q
}

// Where 添加条件
func (q *QueryBuilder) Where(condition string) *QueryBuilder {
	q.conditions = append(q.conditions, condition)
	return q
}

// And 添加 AND 条件
func (q *QueryBuilder) And(condition string) *QueryBuilder {
	return q.Where(condition)
}

// Or 添加 OR 条件
func (q *QueryBuilder) Or(condition string) *QueryBuilder {
	if len(q.conditions) > 0 {
		last := q.conditions[len(q.conditions)-1]
		q.conditions[len(q.conditions)-1] = fmt.Sprintf("(%s OR %s)", last, condition)
	}
	return q
}

// Join 添加 JOIN
func (q *QueryBuilder) Join(table, condition string) *QueryBuilder {
	q.joins = append(q.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return q
}

// LeftJoin 添加 LEFT JOIN
func (q *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	q.joins = append(q.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return q
}

// OrderBy 排序
func (q *QueryBuilder) OrderBy(column string) *QueryBuilder {
	q.orderBy = append(q.orderBy, column)
	return q
}

// OrderByDesc 降序排序
func (q *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	q.orderBy = append(q.orderBy, column+" DESC")
	return q
}

// GroupBy 分组
func (q *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	q.groupBy = columns
	return q
}

// Having HAVING 子句
func (q *QueryBuilder) Having(condition string) *QueryBuilder {
	q.having = condition
	return q
}

// Limit 限制数量
func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.limit = n
	return q
}

// Offset 偏移量
func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.offset = n
	return q
}

// Build 构建 SQL
func (q *QueryBuilder) Build() string {
	var parts []string

	// SELECT
	parts = append(parts, "SELECT "+strings.Join(q.columns, ", "))

	// FROM
	if q.table != "" {
		parts = append(parts, "FROM "+q.table)
	}

	// JOIN
	for _, join := range q.joins {
		parts = append(parts, join)
	}

	// WHERE
	if len(q.conditions) > 0 {
		parts = append(parts, "WHERE "+strings.Join(q.conditions, " AND "))
	}

	// GROUP BY
	if len(q.groupBy) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(q.groupBy, ", "))
	}

	// HAVING
	if q.having != "" {
		parts = append(parts, "HAVING "+q.having)
	}

	// ORDER BY
	if len(q.orderBy) > 0 {
		parts = append(parts, "ORDER BY "+strings.Join(q.orderBy, ", "))
	}

	// LIMIT
	if q.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", q.limit))
	}

	// OFFSET
	if q.offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", q.offset))
	}

	return strings.Join(parts, " ")
}

// ============================================================================
// 配置 DSL - 服务器配置构建器
// ============================================================================

// ServerConfig 服务器配置
type ServerConfig struct {
	host       string
	port       int
	timeout    int
	maxConns   int
	tls        *TLSConfig
	middleware []string
	routes     []RouteConfig
}

// TLSConfig TLS 配置
type TLSConfig struct {
	certFile string
	keyFile  string
	minVer   string
}

// RouteConfig 路由配置
type RouteConfig struct {
	method  string
	path    string
	handler string
}

// ServerBuilder 服务器配置构建器
type ServerBuilder struct {
	config *ServerConfig
}

// NewServer 创建服务器配置
func NewServer() *ServerBuilder {
	return &ServerBuilder{
		config: &ServerConfig{
			host:     "localhost",
			port:     8080,
			timeout:  30,
			maxConns: 100,
		},
	}
}

// Host 设置主机
func (b *ServerBuilder) Host(host string) *ServerBuilder {
	b.config.host = host
	return b
}

// Port 设置端口
func (b *ServerBuilder) Port(port int) *ServerBuilder {
	b.config.port = port
	return b
}

// Timeout 设置超时
func (b *ServerBuilder) Timeout(seconds int) *ServerBuilder {
	b.config.timeout = seconds
	return b
}

// MaxConnections 设置最大连接数
func (b *ServerBuilder) MaxConnections(n int) *ServerBuilder {
	b.config.maxConns = n
	return b
}

// WithTLS 启用 TLS
func (b *ServerBuilder) WithTLS(certFile, keyFile string) *ServerBuilder {
	b.config.tls = &TLSConfig{
		certFile: certFile,
		keyFile:  keyFile,
		minVer:   "1.2",
	}
	return b
}

// UseMiddleware 添加中间件
func (b *ServerBuilder) UseMiddleware(middleware ...string) *ServerBuilder {
	b.config.middleware = append(b.config.middleware, middleware...)
	return b
}

// Route 添加路由
func (b *ServerBuilder) Route(method, path, handler string) *ServerBuilder {
	b.config.routes = append(b.config.routes, RouteConfig{
		method:  method,
		path:    path,
		handler: handler,
	})
	return b
}

// Get 添加 GET 路由
func (b *ServerBuilder) Get(path, handler string) *ServerBuilder {
	return b.Route("GET", path, handler)
}

// Post 添加 POST 路由
func (b *ServerBuilder) Post(path, handler string) *ServerBuilder {
	return b.Route("POST", path, handler)
}

// Build 构建配置
func (b *ServerBuilder) Build() *ServerConfig {
	return b.config
}

// String 配置字符串表示
func (c *ServerConfig) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Server Configuration:\n"))
	sb.WriteString(fmt.Sprintf("  Host: %s\n", c.host))
	sb.WriteString(fmt.Sprintf("  Port: %d\n", c.port))
	sb.WriteString(fmt.Sprintf("  Timeout: %ds\n", c.timeout))
	sb.WriteString(fmt.Sprintf("  Max Connections: %d\n", c.maxConns))

	if c.tls != nil {
		sb.WriteString(fmt.Sprintf("  TLS: enabled (cert: %s)\n", c.tls.certFile))
	}

	if len(c.middleware) > 0 {
		sb.WriteString(fmt.Sprintf("  Middleware: %v\n", c.middleware))
	}

	if len(c.routes) > 0 {
		sb.WriteString("  Routes:\n")
		for _, r := range c.routes {
			sb.WriteString(fmt.Sprintf("    %s %s -> %s\n", r.method, r.path, r.handler))
		}
	}

	return sb.String()
}

// ============================================================================
// 规则引擎 DSL
// ============================================================================

// Rule 规则
type Rule struct {
	name       string
	conditions []Condition
	actions    []Action
	priority   int
}

// Condition 条件
type Condition struct {
	field    string
	operator string
	value    interface{}
}

// Action 动作
type Action struct {
	actionType string
	params     map[string]interface{}
}

// RuleBuilder 规则构建器
type RuleBuilder struct {
	rule *Rule
}

// NewRule 创建规则
func NewRule(name string) *RuleBuilder {
	return &RuleBuilder{
		rule: &Rule{
			name:     name,
			priority: 0,
		},
	}
}

// When 添加条件
func (b *RuleBuilder) When(field, operator string, value interface{}) *RuleBuilder {
	b.rule.conditions = append(b.rule.conditions, Condition{
		field:    field,
		operator: operator,
		value:    value,
	})
	return b
}

// And 添加 AND 条件
func (b *RuleBuilder) And(field, operator string, value interface{}) *RuleBuilder {
	return b.When(field, operator, value)
}

// Then 添加动作
func (b *RuleBuilder) Then(actionType string, params map[string]interface{}) *RuleBuilder {
	b.rule.actions = append(b.rule.actions, Action{
		actionType: actionType,
		params:     params,
	})
	return b
}

// Priority 设置优先级
func (b *RuleBuilder) Priority(p int) *RuleBuilder {
	b.rule.priority = p
	return b
}

// Build 构建规则
func (b *RuleBuilder) Build() *Rule {
	return b.rule
}

// String 规则字符串表示
func (r *Rule) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Rule: %s (priority: %d)\n", r.name, r.priority))

	sb.WriteString("  WHEN:\n")
	for _, c := range r.conditions {
		sb.WriteString(fmt.Sprintf("    %s %s %v\n", c.field, c.operator, c.value))
	}

	sb.WriteString("  THEN:\n")
	for _, a := range r.actions {
		sb.WriteString(fmt.Sprintf("    %s: %v\n", a.actionType, a.params))
	}

	return sb.String()
}

// Evaluate 评估规则
func (r *Rule) Evaluate(data map[string]interface{}) bool {
	for _, cond := range r.conditions {
		value, exists := data[cond.field]
		if !exists {
			return false
		}

		if !evaluateCondition(value, cond.operator, cond.value) {
			return false
		}
	}
	return true
}

func evaluateCondition(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	case ">":
		return compareNumbers(actual, expected) > 0
	case ">=":
		return compareNumbers(actual, expected) >= 0
	case "<":
		return compareNumbers(actual, expected) < 0
	case "<=":
		return compareNumbers(actual, expected) <= 0
	case "contains":
		if s, ok := actual.(string); ok {
			if sub, ok := expected.(string); ok {
				return strings.Contains(s, sub)
			}
		}
	}
	return false
}

func compareNumbers(a, b interface{}) int {
	var aFloat, bFloat float64

	switch v := a.(type) {
	case int:
		aFloat = float64(v)
	case float64:
		aFloat = v
	}

	switch v := b.(type) {
	case int:
		bFloat = float64(v)
	case float64:
		bFloat = v
	}

	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// ============================================================================
// HTML 模板 DSL
// ============================================================================

// HTMLElement HTML 元素
type HTMLElement struct {
	tag        string
	attributes map[string]string
	children   []interface{} // string 或 *HTMLElement
	selfClose  bool
}

// Tag 创建标签
func Tag(name string) *HTMLElement {
	return &HTMLElement{
		tag:        name,
		attributes: make(map[string]string),
	}
}

// Attr 添加属性
func (e *HTMLElement) Attr(name, value string) *HTMLElement {
	e.attributes[name] = value
	return e
}

// ID 设置 ID
func (e *HTMLElement) ID(id string) *HTMLElement {
	return e.Attr("id", id)
}

// Class 设置 class
func (e *HTMLElement) Class(class string) *HTMLElement {
	return e.Attr("class", class)
}

// Text 添加文本内容
func (e *HTMLElement) Text(text string) *HTMLElement {
	e.children = append(e.children, text)
	return e
}

// Child 添加子元素
func (e *HTMLElement) Child(child *HTMLElement) *HTMLElement {
	e.children = append(e.children, child)
	return e
}

// Children 添加多个子元素
func (e *HTMLElement) Children(children ...*HTMLElement) *HTMLElement {
	for _, child := range children {
		e.children = append(e.children, child)
	}
	return e
}

// SelfClose 设置为自闭合标签
func (e *HTMLElement) SelfClose() *HTMLElement {
	e.selfClose = true
	return e
}

// Render 渲染 HTML
func (e *HTMLElement) Render() string {
	return e.renderWithIndent(0)
}

func (e *HTMLElement) renderWithIndent(indent int) string {
	var sb strings.Builder
	indentStr := strings.Repeat("  ", indent)

	// 开始标签
	sb.WriteString(indentStr)
	sb.WriteString("<")
	sb.WriteString(e.tag)

	// 属性
	for name, value := range e.attributes {
		sb.WriteString(fmt.Sprintf(` %s="%s"`, name, value))
	}

	if e.selfClose {
		sb.WriteString(" />")
		return sb.String()
	}

	sb.WriteString(">")

	// 子元素
	hasElementChildren := false
	for _, child := range e.children {
		if _, ok := child.(*HTMLElement); ok {
			hasElementChildren = true
			break
		}
	}

	if hasElementChildren {
		sb.WriteString("\n")
		for _, child := range e.children {
			switch c := child.(type) {
			case string:
				sb.WriteString(strings.Repeat("  ", indent+1))
				sb.WriteString(c)
				sb.WriteString("\n")
			case *HTMLElement:
				sb.WriteString(c.renderWithIndent(indent + 1))
				sb.WriteString("\n")
			}
		}
		sb.WriteString(indentStr)
	} else {
		for _, child := range e.children {
			if text, ok := child.(string); ok {
				sb.WriteString(text)
			}
		}
	}

	// 结束标签
	sb.WriteString("</")
	sb.WriteString(e.tag)
	sb.WriteString(">")

	return sb.String()
}

// 便捷函数
func Div() *HTMLElement  { return Tag("div") }
func Span() *HTMLElement { return Tag("span") }
func P() *HTMLElement    { return Tag("p") }
func H1() *HTMLElement   { return Tag("h1") }
func H2() *HTMLElement   { return Tag("h2") }
func A() *HTMLElement    { return Tag("a") }
func Ul() *HTMLElement   { return Tag("ul") }
func Li() *HTMLElement   { return Tag("li") }
func Img() *HTMLElement  { return Tag("img").SelfClose() }
func Br() *HTMLElement   { return Tag("br").SelfClose() }

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateQueryDSL() {
	fmt.Println("\n=== SQL 查询 DSL 演示 ===")

	// 简单查询
	query1 := NewQuery().
		Select("id", "name", "email").
		From("users").
		Where("status = 'active'").
		OrderBy("created_at").
		Limit(10).
		Build()

	fmt.Println("简单查询:")
	fmt.Printf("  %s\n", query1)

	// 复杂查询
	query2 := NewQuery().
		Select("u.id", "u.name", "COUNT(o.id) as order_count").
		From("users u").
		LeftJoin("orders o", "u.id = o.user_id").
		Where("u.status = 'active'").
		And("o.created_at > '2024-01-01'").
		GroupBy("u.id", "u.name").
		Having("COUNT(o.id) > 5").
		OrderByDesc("order_count").
		Limit(20).
		Build()

	fmt.Println("\n复杂查询:")
	fmt.Printf("  %s\n", query2)
}

func demonstrateServerDSL() {
	fmt.Println("\n=== 服务器配置 DSL 演示 ===")

	config := NewServer().
		Host("0.0.0.0").
		Port(8443).
		Timeout(60).
		MaxConnections(1000).
		WithTLS("cert.pem", "key.pem").
		UseMiddleware("logging", "auth", "cors").
		Get("/api/users", "UserHandler.List").
		Get("/api/users/:id", "UserHandler.Get").
		Post("/api/users", "UserHandler.Create").
		Build()

	fmt.Println(config.String())
}

func demonstrateRuleDSL() {
	fmt.Println("\n=== 规则引擎 DSL 演示 ===")

	// 定义规则
	discountRule := NewRule("VIP折扣规则").
		When("user_type", "==", "vip").
		And("order_amount", ">=", 100).
		Then("apply_discount", map[string]interface{}{"rate": 0.2}).
		Priority(10).
		Build()

	fmt.Println(discountRule.String())

	// 测试规则
	testData := map[string]interface{}{
		"user_type":    "vip",
		"order_amount": 150,
	}

	fmt.Printf("测试数据: %v\n", testData)
	fmt.Printf("规则匹配: %v\n", discountRule.Evaluate(testData))

	// 不匹配的数据
	testData2 := map[string]interface{}{
		"user_type":    "normal",
		"order_amount": 150,
	}

	fmt.Printf("\n测试数据: %v\n", testData2)
	fmt.Printf("规则匹配: %v\n", discountRule.Evaluate(testData2))
}

func demonstrateHTMLDSL() {
	fmt.Println("\n=== HTML 模板 DSL 演示 ===")

	html := Div().ID("container").Class("main").Children(
		H1().Text("欢迎使用 DSL"),
		P().Class("intro").Text("这是一个使用 Go 构建的 HTML DSL 示例。"),
		Ul().Class("features").Children(
			Li().Text("流式接口"),
			Li().Text("类型安全"),
			Li().Text("易于扩展"),
		),
		Div().Class("footer").Child(
			A().Attr("href", "https://golang.org").Text("了解更多"),
		),
	)

	fmt.Println("生成的 HTML:")
	fmt.Println(html.Render())
}

func main() {
	fmt.Println("=== 领域特定语言（DSL）实现 ===")
	fmt.Println()
	fmt.Println("本模块演示 DSL 设计的核心技术:")
	fmt.Println("1. SQL 查询构建器")
	fmt.Println("2. 服务器配置 DSL")
	fmt.Println("3. 规则引擎 DSL")
	fmt.Println("4. HTML 模板 DSL")

	demonstrateQueryDSL()
	demonstrateServerDSL()
	demonstrateRuleDSL()
	demonstrateHTMLDSL()

	fmt.Println("\n=== DSL 演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 流式接口通过返回 self 实现方法链")
	fmt.Println("- 内部 DSL 利用宿主语言的语法")
	fmt.Println("- Builder 模式是 DSL 的常见实现方式")
	fmt.Println("- 好的 DSL 应该易读、易写、难以误用")
}
