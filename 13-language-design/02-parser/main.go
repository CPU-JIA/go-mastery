// Package main 演示语法解析器实现
// 本模块涵盖编译器前端的核心技术：
// - 词法分析（Lexer）
// - 语法分析（Parser）
// - 抽象语法树（AST）
// - 递归下降解析
package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ============================================================================
// 词法分析器（Lexer）
// ============================================================================

// TokenType 词法单元类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenInt
	TokenFloat
	TokenString
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenSemicolon
	TokenAssign
	TokenEq
	TokenNe
	TokenLt
	TokenGt
	TokenLe
	TokenGe
	TokenAnd
	TokenOr
	TokenNot
	TokenKeywordFunc
	TokenKeywordVar
	TokenKeywordIf
	TokenKeywordElse
	TokenKeywordFor
	TokenKeywordReturn
)

func (t TokenType) String() string {
	names := map[TokenType]string{
		TokenEOF:           "EOF",
		TokenIdent:         "IDENT",
		TokenInt:           "INT",
		TokenFloat:         "FLOAT",
		TokenString:        "STRING",
		TokenPlus:          "+",
		TokenMinus:         "-",
		TokenStar:          "*",
		TokenSlash:         "/",
		TokenLParen:        "(",
		TokenRParen:        ")",
		TokenLBrace:        "{",
		TokenRBrace:        "}",
		TokenComma:         ",",
		TokenSemicolon:     ";",
		TokenAssign:        "=",
		TokenEq:            "==",
		TokenNe:            "!=",
		TokenLt:            "<",
		TokenGt:            ">",
		TokenLe:            "<=",
		TokenGe:            ">=",
		TokenAnd:           "&&",
		TokenOr:            "||",
		TokenNot:           "!",
		TokenKeywordFunc:   "func",
		TokenKeywordVar:    "var",
		TokenKeywordIf:     "if",
		TokenKeywordElse:   "else",
		TokenKeywordFor:    "for",
		TokenKeywordReturn: "return",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Token 词法单元
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Type, t.Literal, t.Line, t.Column)
}

// Lexer 词法分析器
type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
	column  int
}

// NewLexer 创建词法分析器
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// NextToken 获取下一个词法单元
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := Token{Line: l.line, Column: l.column}

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	case '+':
		tok.Type = TokenPlus
		tok.Literal = "+"
	case '-':
		tok.Type = TokenMinus
		tok.Literal = "-"
	case '*':
		tok.Type = TokenStar
		tok.Literal = "*"
	case '/':
		tok.Type = TokenSlash
		tok.Literal = "/"
	case '(':
		tok.Type = TokenLParen
		tok.Literal = "("
	case ')':
		tok.Type = TokenRParen
		tok.Literal = ")"
	case '{':
		tok.Type = TokenLBrace
		tok.Literal = "{"
	case '}':
		tok.Type = TokenRBrace
		tok.Literal = "}"
	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
	case ';':
		tok.Type = TokenSemicolon
		tok.Literal = ";"
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenEq
			tok.Literal = "=="
		} else {
			tok.Type = TokenAssign
			tok.Literal = "="
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenNe
			tok.Literal = "!="
		} else {
			tok.Type = TokenNot
			tok.Literal = "!"
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenLe
			tok.Literal = "<="
		} else {
			tok.Type = TokenLt
			tok.Literal = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenGe
			tok.Literal = ">="
		} else {
			tok.Type = TokenGt
			tok.Literal = ">"
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok.Type = TokenAnd
			tok.Literal = "&&"
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok.Type = TokenOr
			tok.Literal = "||"
		}
	case '"':
		tok.Type = TokenString
		tok.Literal = l.readString()
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupKeyword(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			return tok
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() (string, TokenType) {
	pos := l.pos
	tokenType := TokenInt

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = TokenFloat
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[pos:l.pos], tokenType
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening quote
	pos := l.pos
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	str := l.input[pos:l.pos]
	return str
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

var keywords = map[string]TokenType{
	"func":   TokenKeywordFunc,
	"var":    TokenKeywordVar,
	"if":     TokenKeywordIf,
	"else":   TokenKeywordElse,
	"for":    TokenKeywordFor,
	"return": TokenKeywordReturn,
}

func lookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}

// ============================================================================
// 抽象语法树（AST）
// ============================================================================

// Node AST 节点接口
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement 语句接口
type Statement interface {
	Node
	statementNode()
}

// Expression 表达式接口
type Expression interface {
	Node
	expressionNode()
}

// Program 程序（根节点）
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// VarStatement 变量声明语句
type VarStatement struct {
	Token Token
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }
func (vs *VarStatement) String() string {
	return fmt.Sprintf("var %s = %s;", vs.Name.String(), vs.Value.String())
}

// ReturnStatement 返回语句
type ReturnStatement struct {
	Token Token
	Value Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	if rs.Value != nil {
		return fmt.Sprintf("return %s;", rs.Value.String())
	}
	return "return;"
}

// ExpressionStatement 表达式语句
type ExpressionStatement struct {
	Token      Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Identifier 标识符
type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral 整数字面量
type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral 浮点数字面量
type FloatLiteral struct {
	Token Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral 字符串字面量
type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// PrefixExpression 前缀表达式
type PrefixExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression 中缀表达式
type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// CallExpression 函数调用表达式
type CallExpression struct {
	Token     Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// ============================================================================
// 语法分析器（Parser）
// ============================================================================

// 运算符优先级
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[TokenType]int{
	TokenEq:     EQUALS,
	TokenNe:     EQUALS,
	TokenLt:     LESSGREATER,
	TokenGt:     LESSGREATER,
	TokenLe:     LESSGREATER,
	TokenGe:     LESSGREATER,
	TokenPlus:   SUM,
	TokenMinus:  SUM,
	TokenSlash:  PRODUCT,
	TokenStar:   PRODUCT,
	TokenLParen: CALL,
}

// Parser 语法分析器
type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []string

	prefixParseFns map[TokenType]func() Expression
	infixParseFns  map[TokenType]func(Expression) Expression
}

// NewParser 创建语法分析器
func NewParser(l *Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[TokenType]func() Expression)
	p.registerPrefix(TokenIdent, p.parseIdentifier)
	p.registerPrefix(TokenInt, p.parseIntegerLiteral)
	p.registerPrefix(TokenFloat, p.parseFloatLiteral)
	p.registerPrefix(TokenString, p.parseStringLiteral)
	p.registerPrefix(TokenMinus, p.parsePrefixExpression)
	p.registerPrefix(TokenNot, p.parsePrefixExpression)
	p.registerPrefix(TokenLParen, p.parseGroupedExpression)

	p.infixParseFns = make(map[TokenType]func(Expression) Expression)
	p.registerInfix(TokenPlus, p.parseInfixExpression)
	p.registerInfix(TokenMinus, p.parseInfixExpression)
	p.registerInfix(TokenSlash, p.parseInfixExpression)
	p.registerInfix(TokenStar, p.parseInfixExpression)
	p.registerInfix(TokenEq, p.parseInfixExpression)
	p.registerInfix(TokenNe, p.parseInfixExpression)
	p.registerInfix(TokenLt, p.parseInfixExpression)
	p.registerInfix(TokenGt, p.parseInfixExpression)
	p.registerInfix(TokenLe, p.parseInfixExpression)
	p.registerInfix(TokenGe, p.parseInfixExpression)
	p.registerInfix(TokenLParen, p.parseCallExpression)

	// 读取两个 token，初始化 curToken 和 peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType TokenType, fn func() Expression) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType TokenType, fn func(Expression) Expression) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("期望下一个 token 是 %s，实际是 %s", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// Errors 获取解析错误
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram 解析程序
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for !p.curTokenIs(TokenEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case TokenKeywordVar:
		return p.parseVarStatement()
	case TokenKeywordReturn:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVarStatement() *VarStatement {
	stmt := &VarStatement{Token: p.curToken}

	if !p.expectPeek(TokenIdent) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(TokenAssign) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(TokenSemicolon) {
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("没有找到 %s 的前缀解析函数", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(TokenSemicolon) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("无法解析 %q 为整数", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("无法解析 %q 为浮点数", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(TokenRParen) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}

	if p.peekTokenIs(TokenRParen) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(TokenComma) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(TokenRParen) {
		return nil
	}

	return args
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateLexer() {
	fmt.Println("=== 词法分析器演示 ===")

	input := `var x = 10;
var y = 20;
x + y * 2`

	fmt.Println("输入代码:")
	fmt.Println(input)
	fmt.Println("词法单元:")

	lexer := NewLexer(input)
	for {
		tok := lexer.NextToken()
		fmt.Printf("  %s\n", tok)
		if tok.Type == TokenEOF {
			break
		}
	}
}

func demonstrateParser() {
	fmt.Println("=== 语法分析器演示 ===")

	input := `var x = 10;
var y = x + 5;
return x * y;`

	fmt.Println("输入代码:")
	fmt.Println(input)

	lexer := NewLexer(input)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		fmt.Println("解析错误:")
		for _, err := range parser.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	fmt.Println("抽象语法树:")
	for i, stmt := range program.Statements {
		fmt.Printf("  语句 %d: %s\n", i+1, stmt.String())
	}
}

func demonstrateExpressionParsing() {
	fmt.Println("=== 表达式解析演示 ===")

	expressions := []string{
		"1 + 2",
		"1 + 2 * 3",
		"(1 + 2) * 3",
		"-5 + 10",
		"a + b * c",
		"add(1, 2 * 3)",
	}

	for _, expr := range expressions {
		lexer := NewLexer(expr)
		parser := NewParser(lexer)
		program := parser.ParseProgram()

		if len(parser.Errors()) > 0 {
			fmt.Printf("  %s => 错误: %v\n", expr, parser.Errors())
			continue
		}

		if len(program.Statements) > 0 {
			fmt.Printf("  %s => %s\n", expr, program.Statements[0].String())
		}
	}
}

func demonstrateAST() {
	fmt.Println("=== AST 结构演示 ===")
	fmt.Println(`抽象语法树节点类型:

Program (程序)
├── Statement (语句)
│   ├── VarStatement (变量声明)
│   ├── ReturnStatement (返回语句)
│   └── ExpressionStatement (表达式语句)
└── Expression (表达式)
    ├── Identifier (标识符)
    ├── IntegerLiteral (整数字面量)
    ├── FloatLiteral (浮点数字面量)
    ├── StringLiteral (字符串字面量)
    ├── PrefixExpression (前缀表达式)
    ├── InfixExpression (中缀表达式)
    └── CallExpression (函数调用)

示例 AST:
  var x = 1 + 2 * 3;

  VarStatement
  ├── Name: Identifier("x")
  └── Value: InfixExpression
      ├── Left: IntegerLiteral(1)
      ├── Operator: "+"
      └── Right: InfixExpression
          ├── Left: IntegerLiteral(2)
          ├── Operator: "*"
          └── Right: IntegerLiteral(3)`)
}

func main() {
	fmt.Println("=== 语法解析器实现 ===")
	fmt.Println()
	fmt.Println("本模块演示编译器前端的核心技术:")
	fmt.Println("1. 词法分析（Lexer）")
	fmt.Println("2. 语法分析（Parser）")
	fmt.Println("3. 抽象语法树（AST）")
	fmt.Println("4. 递归下降解析")

	demonstrateLexer()
	demonstrateAST()
	demonstrateParser()
	demonstrateExpressionParsing()

	fmt.Println("=== 语法解析器演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 词法分析将源代码转换为 Token 流")
	fmt.Println("- 语法分析将 Token 流转换为 AST")
	fmt.Println("- Pratt 解析器优雅处理运算符优先级")
	fmt.Println("- AST 是后续编译阶段的基础")
}
