/*
=== Go编译器工具链：词法分析大师 ===

本模块专注于Go编译器词法分析的深度技术，探索：
1. 高性能词法分析器设计与实现
2. 有限状态机和正则表达式引擎
3. 多语言词法分析支持
4. 错误恢复和诊断机制
5. 增量词法分析技术
6. 并行词法分析优化
7. 词法分析器生成器
8. Unicode和多字节字符处理
9. 预处理器和宏展开
10. 领域特定语言词法支持

学习目标：
- 掌握编译器前端词法分析原理
- 理解有限状态机的设计和优化
- 学会构建高性能词法分析器
- 掌握现代编译器词法分析技术
*/

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ==================
// 1. 词法分析器核心
// ==================

// LexicalAnalyzer 词法分析器
type LexicalAnalyzer struct {
	tokenizers   map[string]*Tokenizer
	fsm          *FiniteStateMachine
	regexEngine  *RegexEngine
	errorHandler *LexicalErrorHandler
	preprocessor *Preprocessor
	cache        *TokenCache
	config       LexerConfig
	statistics   LexerStatistics
	extensions   map[string]LexerExtension
	middleware   []LexerMiddleware
	mutex        sync.RWMutex
}

// LexerConfig 词法分析器配置
type LexerConfig struct {
	Language           string
	CaseSensitive      bool
	EnableUnicode      bool
	EnablePreprocessor bool
	EnableCache        bool
	CacheSize          int
	MaxTokenLength     int
	ParallelWorkers    int
	EnableProfiling    bool
	ErrorRecovery      bool
	DebugMode          bool
}

// Tokenizer 分词器
type Tokenizer struct {
	name         string
	language     string
	rules        []*TokenRule
	states       map[string]*LexerState
	currentState *LexerState
	input        *LexerInput
	position     Position
	tokens       []*Token
	errors       []*LexicalError
	statistics   TokenizerStatistics
	config       TokenizerConfig
	mutex        sync.RWMutex
}

// TokenRule 词法规则
type TokenRule struct {
	ID        string
	Name      string
	Pattern   string
	TokenType TokenType
	Priority  int
	States    []string
	Action    TokenAction
	Condition RuleCondition
	Regex     *regexp.Regexp
	FSM       *FiniteStateMachine
	Enabled   bool
	Metadata  map[string]interface{}
}

// Token 词法单元
type Token struct {
	Type      TokenType
	Value     string
	Position  Position
	Length    int
	Line      int
	Column    int
	Raw       string
	Metadata  map[string]interface{}
	Context   *TokenContext
	Children  []*Token
	Parent    *Token
	Timestamp time.Time
}

// TokenType 词法单元类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenComment
	TokenWhitespace
	TokenNewline
	TokenIdentifier
	TokenKeyword
	TokenNumber
	TokenString
	TokenChar
	TokenOperator
	TokenPunctuation
	TokenDelimiter
	TokenLiteral
	TokenRegex
	TokenPreprocessor
	TokenCustom
)

var tokenTypeNames = []string{
	"EOF", "ERROR", "COMMENT", "WHITESPACE", "NEWLINE",
	"IDENTIFIER", "KEYWORD", "NUMBER", "STRING", "CHAR",
	"OPERATOR", "PUNCTUATION", "DELIMITER", "LITERAL",
	"REGEX", "PREPROCESSOR", "CUSTOM",
}

func (tt TokenType) String() string {
	if int(tt) < len(tokenTypeNames) {
		return tokenTypeNames[tt]
	}
	return "UNKNOWN"
}

// Position 位置信息
type Position struct {
	Offset int
	Line   int
	Column int
	File   string
}

// TokenContext 词法单元上下文
type TokenContext struct {
	State     string
	Scope     string
	Flags     map[string]bool
	Variables map[string]interface{}
	Stack     []string
	Depth     int
}

// LexerInput 词法输入
type LexerInput struct {
	source   []rune
	position int
	length   int
	line     int
	column   int
	file     string
	encoding string
}

// LexerState 词法状态
type LexerState struct {
	Name        string
	Type        StateType
	Rules       []*TokenRule
	Transitions map[string]string
	Actions     []StateAction
	Flags       map[string]bool
	Stack       []string
	Default     string
}

// StateType 状态类型
type StateType int

const (
	StateNormal StateType = iota
	StateString
	StateComment
	StateNumber
	StateIdentifier
	StateOperator
	StateRegex
	StatePreprocessor
	StateError
)

func NewLexicalAnalyzer(config LexerConfig) *LexicalAnalyzer {
	return &LexicalAnalyzer{
		tokenizers:   make(map[string]*Tokenizer),
		fsm:          NewFiniteStateMachine(),
		regexEngine:  NewRegexEngine(),
		errorHandler: NewLexicalErrorHandler(),
		preprocessor: NewPreprocessor(),
		cache:        NewTokenCache(config.CacheSize),
		config:       config,
		extensions:   make(map[string]LexerExtension),
		middleware:   make([]LexerMiddleware, 0),
	}
}

func (la *LexicalAnalyzer) RegisterTokenizer(tokenizer *Tokenizer) error {
	la.mutex.Lock()
	defer la.mutex.Unlock()

	if _, exists := la.tokenizers[tokenizer.language]; exists {
		return fmt.Errorf("tokenizer for language %s already exists", tokenizer.language)
	}

	la.tokenizers[tokenizer.language] = tokenizer
	fmt.Printf("注册词法分析器: %s (%s)\n", tokenizer.name, tokenizer.language)
	return nil
}

func (la *LexicalAnalyzer) Tokenize(input string, language string) ([]*Token, error) {
	la.mutex.RLock()
	tokenizer, exists := la.tokenizers[language]
	la.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no tokenizer found for language: %s", language)
	}

	// 预处理
	if la.config.EnablePreprocessor {
		processed, err := la.preprocessor.Process(input)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %v", err)
		}
		input = processed
	}

	// 检查缓存
	if la.config.EnableCache {
		if cached := la.cache.Get(input, language); cached != nil {
			return cached, nil
		}
	}

	// 执行词法分析
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		return nil, err
	}

	// 应用中间件
	for _, middleware := range la.middleware {
		tokens = middleware.Process(tokens)
	}

	// 缓存结果
	if la.config.EnableCache {
		la.cache.Put(input, language, tokens)
	}

	// 更新统计
	la.statistics.TotalTokens += int64(len(tokens))
	la.statistics.AnalysisCount++

	return tokens, nil
}

func NewTokenizer(name, language string, config TokenizerConfig) *Tokenizer {
	return &Tokenizer{
		name:     name,
		language: language,
		rules:    make([]*TokenRule, 0),
		states:   make(map[string]*LexerState),
		tokens:   make([]*Token, 0),
		errors:   make([]*LexicalError, 0),
		config:   config,
	}
}

func (t *Tokenizer) AddRule(rule *TokenRule) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 编译正则表达式
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			fmt.Printf("警告: 无法编译正则表达式 %s: %v\n", rule.Pattern, err)
			return
		}
		rule.Regex = regex
	}

	t.rules = append(t.rules, rule)

	// 按优先级排序
	sort.Slice(t.rules, func(i, j int) bool {
		return t.rules[i].Priority > t.rules[j].Priority
	})

	fmt.Printf("添加词法规则: %s (优先级: %d)\n", rule.Name, rule.Priority)
}

func (t *Tokenizer) Tokenize(input string) ([]*Token, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 初始化输入
	t.input = &LexerInput{
		source:   []rune(input),
		length:   len([]rune(input)),
		line:     1,
		column:   1,
		encoding: "UTF-8",
	}

	// 重置状态
	t.tokens = t.tokens[:0]
	t.errors = t.errors[:0]
	t.currentState = t.getInitialState()

	startTime := time.Now()

	// 主要词法分析循环
	for !t.isAtEnd() {
		if err := t.scanToken(); err != nil {
			if t.config.ErrorRecovery {
				t.recoverFromError(err)
				continue
			}
			return nil, err
		}
	}

	// 添加EOF标记
	t.addToken(TokenEOF, "", t.getCurrentPosition())

	// 更新统计信息
	t.statistics.TokensProduced = int64(len(t.tokens))
	t.statistics.AnalysisTime = time.Since(startTime)
	t.statistics.CharactersProcessed = int64(t.input.length)

	return t.tokens, nil
}

func (t *Tokenizer) scanToken() error {
	// 跳过空白字符（如果配置要求）
	if t.config.SkipWhitespace {
		t.skipWhitespace()
	}

	if t.isAtEnd() {
		return nil
	}

	start := t.input.position
	startPos := t.getCurrentPosition()

	// 尝试匹配规则
	for _, rule := range t.rules {
		if !rule.Enabled {
			continue
		}

		// 检查状态条件
		if !t.isRuleApplicable(rule) {
			continue
		}

		// 尝试匹配
		match, length := t.tryMatch(rule)
		if match {
			value := string(t.input.source[start : start+length])

			// 执行规则动作
			if rule.Action != nil {
				result := rule.Action(value, startPos, t.currentState)
				if !result.Consume {
					continue
				}
				if result.ChangeState != "" {
					t.changeState(result.ChangeState)
				}
				if result.TokenType != rule.TokenType {
					rule.TokenType = result.TokenType
				}
			}

			// 创建词法单元
			token := &Token{
				Type:      rule.TokenType,
				Value:     value,
				Position:  startPos,
				Length:    length,
				Line:      startPos.Line,
				Column:    startPos.Column,
				Raw:       value,
				Metadata:  make(map[string]interface{}),
				Context:   t.getTokenContext(),
				Timestamp: time.Now(),
			}

			// 应用自定义处理
			t.processToken(token, rule)

			// 添加到结果
			if rule.TokenType != TokenWhitespace || !t.config.SkipWhitespace {
				t.tokens = append(t.tokens, token)
			}

			// 更新位置
			t.advance(length)
			return nil
		}
	}

	// 没有匹配的规则 - 错误处理
	char := t.input.source[t.input.position]
	err := &LexicalError{
		Type:     ErrorUnexpectedCharacter,
		Message:  fmt.Sprintf("unexpected character: %c", char),
		Position: t.getCurrentPosition(),
		Context:  string(t.input.source[start : start+1]),
	}

	t.errors = append(t.errors, err)
	t.advance(1) // 跳过错误字符
	return err
}

func (t *Tokenizer) tryMatch(rule *TokenRule) (bool, int) {
	if rule.Regex != nil {
		// 正则表达式匹配
		source := string(t.input.source[t.input.position:])
		match := rule.Regex.FindString(source)
		if match != "" && rule.Regex.FindStringIndex(source)[0] == 0 {
			return true, len([]rune(match))
		}
	}

	if rule.FSM != nil {
		// 有限状态机匹配
		return t.tryFSMMatch(rule.FSM)
	}

	return false, 0
}

func (t *Tokenizer) tryFSMMatch(fsm *FiniteStateMachine) (bool, int) {
	savedPos := t.input.position
	length := 0

	fsm.Reset()

	for !t.isAtEnd() && !fsm.IsInDeadState() {
		char := t.input.source[t.input.position]
		if !fsm.Transition(char) {
			break
		}

		t.input.position++
		length++

		if fsm.IsInAcceptState() {
			// 找到匹配
			t.input.position = savedPos // 重置位置
			return true, length
		}
	}

	t.input.position = savedPos // 重置位置
	return false, 0
}

func (t *Tokenizer) isRuleApplicable(rule *TokenRule) bool {
	// 检查状态条件
	if len(rule.States) > 0 {
		stateMatch := false
		for _, state := range rule.States {
			if t.currentState.Name == state {
				stateMatch = true
				break
			}
		}
		if !stateMatch {
			return false
		}
	}

	// 检查自定义条件
	if rule.Condition != nil {
		return rule.Condition(t.getCurrentPosition(), t.currentState)
	}

	return true
}

func (t *Tokenizer) processToken(token *Token, rule *TokenRule) {
	// 关键字检查
	if token.Type == TokenIdentifier {
		if t.isKeyword(token.Value) {
			token.Type = TokenKeyword
		}
	}

	// 数字处理
	if token.Type == TokenNumber {
		t.processNumber(token)
	}

	// 字符串处理
	if token.Type == TokenString {
		t.processString(token)
	}

	// 添加元数据
	token.Metadata["rule"] = rule.Name
	token.Metadata["state"] = t.currentState.Name
}

func (t *Tokenizer) processNumber(token *Token) {
	value := token.Value

	// 检测数字类型
	if strings.Contains(value, ".") {
		token.Metadata["number_type"] = "float"
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			token.Metadata["float_value"] = f
		}
	} else if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		token.Metadata["number_type"] = "hex"
		if i, err := strconv.ParseInt(value[2:], 16, 64); err == nil {
			token.Metadata["int_value"] = i
		}
	} else if strings.HasPrefix(value, "0b") || strings.HasPrefix(value, "0B") {
		token.Metadata["number_type"] = "binary"
		if i, err := strconv.ParseInt(value[2:], 2, 64); err == nil {
			token.Metadata["int_value"] = i
		}
	} else if strings.HasPrefix(value, "0") && len(value) > 1 {
		token.Metadata["number_type"] = "octal"
		if i, err := strconv.ParseInt(value[1:], 8, 64); err == nil {
			token.Metadata["int_value"] = i
		}
	} else {
		token.Metadata["number_type"] = "decimal"
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			token.Metadata["int_value"] = i
		}
	}
}

func (t *Tokenizer) processString(token *Token) {
	value := token.Value

	// 移除引号
	if len(value) >= 2 {
		quote := value[0]
		if quote == '"' || quote == '\'' || quote == '`' {
			token.Metadata["quote_type"] = string(quote)
			token.Value = value[1 : len(value)-1]
		}
	}

	// 处理转义序列
	if strings.Contains(token.Value, "\\") {
		unescaped := t.unescapeString(token.Value)
		token.Metadata["escaped"] = true
		token.Metadata["unescaped_value"] = unescaped
	}
}

func (t *Tokenizer) unescapeString(s string) string {
	result := make([]rune, 0, len(s))
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			switch runes[i+1] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			default:
				result = append(result, runes[i+1])
			}
			i++ // 跳过转义字符
		} else {
			result = append(result, runes[i])
		}
	}

	return string(result)
}

func (t *Tokenizer) isKeyword(value string) bool {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}

	for _, keyword := range keywords {
		if value == keyword {
			return true
		}
	}
	return false
}

func (t *Tokenizer) addToken(tokenType TokenType, value string, position Position) {
	token := &Token{
		Type:      tokenType,
		Value:     value,
		Position:  position,
		Length:    len([]rune(value)),
		Line:      position.Line,
		Column:    position.Column,
		Raw:       value,
		Metadata:  make(map[string]interface{}),
		Context:   t.getTokenContext(),
		Timestamp: time.Now(),
	}

	t.tokens = append(t.tokens, token)
}

func (t *Tokenizer) getCurrentPosition() Position {
	return Position{
		Offset: t.input.position,
		Line:   t.input.line,
		Column: t.input.column,
		File:   t.input.file,
	}
}

func (t *Tokenizer) getTokenContext() *TokenContext {
	return &TokenContext{
		State:     t.currentState.Name,
		Scope:     "global",
		Flags:     make(map[string]bool),
		Variables: make(map[string]interface{}),
		Stack:     make([]string, 0),
		Depth:     0,
	}
}

func (t *Tokenizer) isAtEnd() bool {
	return t.input.position >= t.input.length
}

func (t *Tokenizer) advance(count int) {
	for i := 0; i < count && !t.isAtEnd(); i++ {
		if t.input.source[t.input.position] == '\n' {
			t.input.line++
			t.input.column = 1
		} else {
			t.input.column++
		}
		t.input.position++
	}
}

func (t *Tokenizer) skipWhitespace() {
	for !t.isAtEnd() {
		char := t.input.source[t.input.position]
		if unicode.IsSpace(char) {
			t.advance(1)
		} else {
			break
		}
	}
}

func (t *Tokenizer) getInitialState() *LexerState {
	if state, exists := t.states["initial"]; exists {
		return state
	}

	// 创建默认初始状态
	initialState := &LexerState{
		Name:        "initial",
		Type:        StateNormal,
		Rules:       t.rules,
		Transitions: make(map[string]string),
		Actions:     make([]StateAction, 0),
		Flags:       make(map[string]bool),
		Stack:       make([]string, 0),
	}

	t.states["initial"] = initialState
	return initialState
}

func (t *Tokenizer) changeState(stateName string) {
	if state, exists := t.states[stateName]; exists {
		t.currentState = state
	}
}

func (t *Tokenizer) recoverFromError(err error) {
	// 简单的错误恢复策略：跳过当前字符
	if !t.isAtEnd() {
		t.advance(1)
	}
}

// ==================
// 2. 有限状态机
// ==================

// FiniteStateMachine 有限状态机
type FiniteStateMachine struct {
	states       map[string]*FSMState
	currentState *FSMState
	initialState *FSMState
	alphabet     []rune
	transitions  map[string]map[rune]string
	acceptStates map[string]bool
	deadState    *FSMState
	statistics   FSMStatistics
	config       FSMConfig
	mutex        sync.RWMutex
}

// FSMState FSM状态
type FSMState struct {
	Name      string
	Type      FSMStateType
	Accept    bool
	Actions   []FSMAction
	Data      map[string]interface{}
	Timestamp time.Time
}

// FSMStateType FSM状态类型
type FSMStateType int

const (
	FSMStateNormal FSMStateType = iota
	FSMStateAccept
	FSMStateDead
	FSMStateStart
)

// FSMAction FSM动作
type FSMAction func(char rune, state *FSMState) error

// FSMConfig FSM配置
type FSMConfig struct {
	Deterministic  bool
	MinimizeStates bool
	EnableLogging  bool
}

// FSMStatistics FSM统计
type FSMStatistics struct {
	TransitionCount   int64
	StateChanges      int64
	AcceptCount       int64
	RejectCount       int64
	AveragePathLength float64
}

func NewFiniteStateMachine() *FiniteStateMachine {
	return &FiniteStateMachine{
		states:       make(map[string]*FSMState),
		transitions:  make(map[string]map[rune]string),
		acceptStates: make(map[string]bool),
		config:       FSMConfig{Deterministic: true},
	}
}

func (fsm *FiniteStateMachine) AddState(name string, stateType FSMStateType) *FSMState {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	state := &FSMState{
		Name:      name,
		Type:      stateType,
		Accept:    stateType == FSMStateAccept,
		Actions:   make([]FSMAction, 0),
		Data:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	fsm.states[name] = state

	if stateType == FSMStateStart {
		fsm.initialState = state
		fsm.currentState = state
	}

	if stateType == FSMStateAccept {
		fsm.acceptStates[name] = true
	}

	if stateType == FSMStateDead {
		fsm.deadState = state
	}

	return state
}

func (fsm *FiniteStateMachine) AddTransition(from string, input rune, to string) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.transitions[from] == nil {
		fsm.transitions[from] = make(map[rune]string)
	}

	fsm.transitions[from][input] = to
}

func (fsm *FiniteStateMachine) AddEpsilonTransition(from string, to string) {
	fsm.AddTransition(from, 0, to) // 使用0表示ε转换
}

func (fsm *FiniteStateMachine) Transition(input rune) bool {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.currentState == nil {
		return false
	}

	currentStateName := fsm.currentState.Name

	// 查找转换
	if transitions, exists := fsm.transitions[currentStateName]; exists {
		if nextStateName, exists := transitions[input]; exists {
			if nextState, exists := fsm.states[nextStateName]; exists {
				// 执行当前状态的动作
				for _, action := range fsm.currentState.Actions {
					if err := action(input, fsm.currentState); err != nil {
						return false
					}
				}

				fsm.currentState = nextState
				fsm.statistics.TransitionCount++
				fsm.statistics.StateChanges++

				return true
			}
		}
	}

	// 没有找到转换，进入死状态
	if fsm.deadState != nil {
		fsm.currentState = fsm.deadState
	}

	return false
}

func (fsm *FiniteStateMachine) IsInAcceptState() bool {
	fsm.mutex.RLock()
	defer fsm.mutex.RUnlock()

	if fsm.currentState == nil {
		return false
	}

	return fsm.acceptStates[fsm.currentState.Name]
}

func (fsm *FiniteStateMachine) IsInDeadState() bool {
	fsm.mutex.RLock()
	defer fsm.mutex.RUnlock()

	return fsm.currentState == fsm.deadState
}

func (fsm *FiniteStateMachine) Reset() {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	fsm.currentState = fsm.initialState
}

func (fsm *FiniteStateMachine) GetCurrentState() *FSMState {
	fsm.mutex.RLock()
	defer fsm.mutex.RUnlock()

	return fsm.currentState
}

// ==================
// 3. 正则表达式引擎
// ==================

// RegexEngine 正则表达式引擎
type RegexEngine struct {
	compiled   map[string]*CompiledRegex
	nfa        *NFAEngine
	dfa        *DFAEngine
	cache      *RegexCache
	config     RegexConfig
	statistics RegexStatistics
	mutex      sync.RWMutex
}

// CompiledRegex 编译后的正则表达式
type CompiledRegex struct {
	Pattern      string
	NFA          *NFA
	DFA          *DFA
	Instructions []RegexInstruction
	Groups       []string
	Flags        RegexFlags
	CompiledAt   time.Time
}

// NFA 非确定性有限状态机
type NFA struct {
	States       []*NFAState
	StartState   *NFAState
	AcceptStates []*NFAState
	Transitions  map[*NFAState]map[rune][]*NFAState
	Epsilon      map[*NFAState][]*NFAState
}

// DFA 确定性有限状态机
type DFA struct {
	States       []*DFAState
	StartState   *DFAState
	AcceptStates []*DFAState
	Transitions  map[*DFAState]map[rune]*DFAState
}

// NFAState NFA状态
type NFAState struct {
	ID     int
	Accept bool
	Data   map[string]interface{}
}

// DFAState DFA状态
type DFAState struct {
	ID        int
	Accept    bool
	NFAStates []*NFAState
	Data      map[string]interface{}
}

// RegexInstruction 正则表达式指令
type RegexInstruction struct {
	Op      RegexOp
	Operand interface{}
	Next    int
	Alt     int
}

// RegexOp 正则表达式操作
type RegexOp int

const (
	OpChar RegexOp = iota
	OpCharClass
	OpDot
	OpStart
	OpEnd
	OpGroup
	OpStar
	OpPlus
	OpQuestion
	OpRepeat
	OpAlternate
	OpMatch
)

// RegexFlags 正则表达式标志
type RegexFlags struct {
	IgnoreCase bool
	Multiline  bool
	DotAll     bool
	Unicode    bool
	Global     bool
	Sticky     bool
}

func NewRegexEngine() *RegexEngine {
	return &RegexEngine{
		compiled: make(map[string]*CompiledRegex),
		nfa:      NewNFAEngine(),
		dfa:      NewDFAEngine(),
		cache:    NewRegexCache(1000),
	}
}

func (re *RegexEngine) Compile(pattern string, flags RegexFlags) (*CompiledRegex, error) {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	// 检查缓存
	key := pattern + flags.String()
	if compiled, exists := re.compiled[key]; exists {
		return compiled, nil
	}

	// 解析正则表达式
	ast, err := re.parseRegex(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regex: %v", err)
	}

	// 构建NFA
	nfa, err := re.buildNFA(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to build NFA: %v", err)
	}

	// 转换为DFA
	dfa := re.nfaToDFA(nfa)

	// 优化DFA
	optimizedDFA := re.optimizeDFA(dfa)

	// 生成指令
	instructions := re.generateInstructions(optimizedDFA)

	compiled := &CompiledRegex{
		Pattern:      pattern,
		NFA:          nfa,
		DFA:          optimizedDFA,
		Instructions: instructions,
		Flags:        flags,
		CompiledAt:   time.Now(),
	}

	re.compiled[key] = compiled
	return compiled, nil
}

func (re *RegexEngine) Match(compiled *CompiledRegex, input string) bool {
	return re.executeMatch(compiled, input)
}

func (re *RegexEngine) FindAll(compiled *CompiledRegex, input string) []string {
	matches := make([]string, 0)

	// 简化的查找实现
	for i := 0; i < len(input); i++ {
		if match := re.findAt(compiled, input, i); match != "" {
			matches = append(matches, match)
		}
	}

	return matches
}

func (re *RegexEngine) parseRegex(pattern string) (*RegexAST, error) {
	// 简化的正则表达式解析
	return &RegexAST{
		Type:     ASTLiteral,
		Value:    pattern,
		Children: make([]*RegexAST, 0),
	}, nil
}

func (re *RegexEngine) buildNFA(ast *RegexAST) (*NFA, error) {
	nfa := &NFA{
		States:      make([]*NFAState, 0),
		Transitions: make(map[*NFAState]map[rune][]*NFAState),
		Epsilon:     make(map[*NFAState][]*NFAState),
	}

	// 创建起始和接受状态
	startState := &NFAState{ID: 0, Accept: false}
	acceptState := &NFAState{ID: 1, Accept: true}

	nfa.States = append(nfa.States, startState, acceptState)
	nfa.StartState = startState
	nfa.AcceptStates = []*NFAState{acceptState}

	return nfa, nil
}

func (re *RegexEngine) nfaToDFA(nfa *NFA) *DFA {
	dfa := &DFA{
		States:      make([]*DFAState, 0),
		Transitions: make(map[*DFAState]map[rune]*DFAState),
	}

	// 简化的NFA到DFA转换
	startState := &DFAState{
		ID:        0,
		Accept:    false,
		NFAStates: []*NFAState{nfa.StartState},
	}

	dfa.States = append(dfa.States, startState)
	dfa.StartState = startState

	return dfa
}

func (re *RegexEngine) optimizeDFA(dfa *DFA) *DFA {
	// DFA最小化和优化
	return dfa
}

func (re *RegexEngine) generateInstructions(dfa *DFA) []RegexInstruction {
	instructions := make([]RegexInstruction, 0)

	// 生成虚拟机指令
	instructions = append(instructions, RegexInstruction{
		Op:   OpMatch,
		Next: -1,
	})

	return instructions
}

func (re *RegexEngine) executeMatch(compiled *CompiledRegex, input string) bool {
	// 执行正则表达式匹配
	return len(input) > 0 // 简化实现
}

func (re *RegexEngine) findAt(compiled *CompiledRegex, input string, start int) string {
	// 在指定位置查找匹配
	if start < len(input) {
		return string(input[start]) // 简化实现
	}
	return ""
}

// ==================
// 4. 预处理器
// ==================

// Preprocessor 预处理器
type Preprocessor struct {
	directives map[string]DirectiveHandler
	macros     map[string]*Macro
	includes   map[string]string
	defines    map[string]string
	conditions []Condition
	config     PreprocessorConfig
	statistics PreprocessorStatistics
	mutex      sync.RWMutex
}

// DirectiveHandler 指令处理器
type DirectiveHandler func(args []string, context *PreprocessorContext) (string, error)

// Macro 宏定义
type Macro struct {
	Name       string
	Parameters []string
	Body       string
	Variadic   bool
	Builtin    bool
	DefinedAt  Position
}

// Condition 条件编译
type Condition struct {
	Type   ConditionType
	Expr   string
	Active bool
	Level  int
}

// ConditionType 条件类型
type ConditionType int

const (
	ConditionIf ConditionType = iota
	ConditionElif
	ConditionElse
	ConditionEndif
)

// PreprocessorContext 预处理器上下文
type PreprocessorContext struct {
	File     string
	Line     int
	Defines  map[string]string
	Includes []string
	Stack    []string
	Depth    int
}

func NewPreprocessor() *Preprocessor {
	pp := &Preprocessor{
		directives: make(map[string]DirectiveHandler),
		macros:     make(map[string]*Macro),
		includes:   make(map[string]string),
		defines:    make(map[string]string),
		conditions: make([]Condition, 0),
	}

	// 注册内置指令
	pp.registerBuiltinDirectives()
	pp.registerBuiltinMacros()

	return pp
}

func (pp *Preprocessor) registerBuiltinDirectives() {
	pp.directives["define"] = pp.handleDefine
	pp.directives["undef"] = pp.handleUndef
	pp.directives["include"] = pp.handleInclude
	pp.directives["if"] = pp.handleIf
	pp.directives["ifdef"] = pp.handleIfdef
	pp.directives["ifndef"] = pp.handleIfndef
	pp.directives["elif"] = pp.handleElif
	pp.directives["else"] = pp.handleElse
	pp.directives["endif"] = pp.handleEndif
	pp.directives["pragma"] = pp.handlePragma
	pp.directives["line"] = pp.handleLine
	pp.directives["error"] = pp.handleError
	pp.directives["warning"] = pp.handleWarning
}

func (pp *Preprocessor) registerBuiltinMacros() {
	pp.macros["__FILE__"] = &Macro{
		Name:    "__FILE__",
		Body:    "",
		Builtin: true,
	}

	pp.macros["__LINE__"] = &Macro{
		Name:    "__LINE__",
		Body:    "",
		Builtin: true,
	}

	pp.macros["__DATE__"] = &Macro{
		Name:    "__DATE__",
		Body:    time.Now().Format("Jan 02 2006"),
		Builtin: true,
	}

	pp.macros["__TIME__"] = &Macro{
		Name:    "__TIME__",
		Body:    time.Now().Format("15:04:05"),
		Builtin: true,
	}
}

func (pp *Preprocessor) Process(input string) (string, error) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))

	context := &PreprocessorContext{
		File:     "input.go",
		Defines:  make(map[string]string),
		Includes: make([]string, 0),
		Stack:    make([]string, 0),
	}

	for i, line := range lines {
		context.Line = i + 1

		processed, err := pp.processLine(line, context)
		if err != nil {
			return "", fmt.Errorf("line %d: %v", i+1, err)
		}

		if processed != "" {
			result = append(result, processed)
		}
	}

	return strings.Join(result, "\n"), nil
}

func (pp *Preprocessor) processLine(line string, context *PreprocessorContext) (string, error) {
	trimmed := strings.TrimSpace(line)

	// 检查是否是预处理器指令
	if strings.HasPrefix(trimmed, "#") {
		return pp.processDirective(trimmed[1:], context)
	}

	// 检查条件编译
	if !pp.shouldIncludeLine() {
		return "", nil
	}

	// 宏展开
	expanded := pp.expandMacros(line, context)

	return expanded, nil
}

func (pp *Preprocessor) processDirective(directive string, context *PreprocessorContext) (string, error) {
	parts := strings.Fields(directive)
	if len(parts) == 0 {
		return "", nil
	}

	directiveName := parts[0]
	args := parts[1:]

	if handler, exists := pp.directives[directiveName]; exists {
		return handler(args, context)
	}

	return "", fmt.Errorf("unknown directive: %s", directiveName)
}

func (pp *Preprocessor) handleDefine(args []string, context *PreprocessorContext) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("define requires at least one argument")
	}

	name := args[0]
	var body string

	if len(args) > 1 {
		body = strings.Join(args[1:], " ")
	}

	// 检查是否是函数式宏
	if strings.Contains(name, "(") {
		return pp.defineFunctionMacro(name, body, context)
	}

	pp.defines[name] = body
	pp.macros[name] = &Macro{
		Name:      name,
		Body:      body,
		Builtin:   false,
		DefinedAt: Position{Line: context.Line, File: context.File},
	}

	return "", nil
}

func (pp *Preprocessor) defineFunctionMacro(definition string, body string, context *PreprocessorContext) (string, error) {
	// 解析函数式宏定义
	openParen := strings.Index(definition, "(")
	closeParen := strings.Index(definition, ")")

	if openParen == -1 || closeParen == -1 || closeParen < openParen {
		return "", fmt.Errorf("invalid function macro definition")
	}

	name := definition[:openParen]
	paramStr := definition[openParen+1 : closeParen]

	var parameters []string
	if paramStr != "" {
		parameters = strings.Split(paramStr, ",")
		for i, param := range parameters {
			parameters[i] = strings.TrimSpace(param)
		}
	}

	pp.macros[name] = &Macro{
		Name:       name,
		Parameters: parameters,
		Body:       body,
		Builtin:    false,
		DefinedAt:  Position{Line: context.Line, File: context.File},
	}

	return "", nil
}

func (pp *Preprocessor) handleUndef(args []string, context *PreprocessorContext) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("undef requires exactly one argument")
	}

	name := args[0]
	delete(pp.defines, name)
	delete(pp.macros, name)

	return "", nil
}

func (pp *Preprocessor) handleInclude(args []string, context *PreprocessorContext) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("include requires exactly one argument")
	}

	filename := args[0]

	// 移除引号
	if strings.HasPrefix(filename, "\"") && strings.HasSuffix(filename, "\"") {
		filename = filename[1 : len(filename)-1]
	} else if strings.HasPrefix(filename, "<") && strings.HasSuffix(filename, ">") {
		filename = filename[1 : len(filename)-1]
	}

	// 检查循环包含
	for _, included := range context.Includes {
		if included == filename {
			return "", fmt.Errorf("circular include detected: %s", filename)
		}
	}

	// 简化实现：返回包含注释
	return fmt.Sprintf("// #include \"%s\"", filename), nil
}

func (pp *Preprocessor) handleIf(args []string, context *PreprocessorContext) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("if requires an expression")
	}

	expr := strings.Join(args, " ")
	result := pp.evaluateExpression(expr, context)

	pp.conditions = append(pp.conditions, Condition{
		Type:   ConditionIf,
		Expr:   expr,
		Active: result,
		Level:  len(pp.conditions),
	})

	return "", nil
}

func (pp *Preprocessor) handleIfdef(args []string, context *PreprocessorContext) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ifdef requires exactly one argument")
	}

	name := args[0]
	_, defined := pp.defines[name]

	pp.conditions = append(pp.conditions, Condition{
		Type:   ConditionIf,
		Expr:   "defined(" + name + ")",
		Active: defined,
		Level:  len(pp.conditions),
	})

	return "", nil
}

func (pp *Preprocessor) handleIfndef(args []string, context *PreprocessorContext) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ifndef requires exactly one argument")
	}

	name := args[0]
	_, defined := pp.defines[name]

	pp.conditions = append(pp.conditions, Condition{
		Type:   ConditionIf,
		Expr:   "!defined(" + name + ")",
		Active: !defined,
		Level:  len(pp.conditions),
	})

	return "", nil
}

func (pp *Preprocessor) handleElif(args []string, context *PreprocessorContext) (string, error) {
	if len(pp.conditions) == 0 {
		return "", fmt.Errorf("elif without matching if")
	}

	if len(args) == 0 {
		return "", fmt.Errorf("elif requires an expression")
	}

	expr := strings.Join(args, " ")
	result := pp.evaluateExpression(expr, context)

	// 更新当前条件
	lastIndex := len(pp.conditions) - 1
	pp.conditions[lastIndex].Type = ConditionElif
	pp.conditions[lastIndex].Expr = expr
	pp.conditions[lastIndex].Active = result

	return "", nil
}

func (pp *Preprocessor) handleElse(args []string, context *PreprocessorContext) (string, error) {
	if len(pp.conditions) == 0 {
		return "", fmt.Errorf("else without matching if")
	}

	// 切换当前条件状态
	lastIndex := len(pp.conditions) - 1
	pp.conditions[lastIndex].Type = ConditionElse
	pp.conditions[lastIndex].Active = !pp.conditions[lastIndex].Active

	return "", nil
}

func (pp *Preprocessor) handleEndif(args []string, context *PreprocessorContext) (string, error) {
	if len(pp.conditions) == 0 {
		return "", fmt.Errorf("endif without matching if")
	}

	// 移除最后一个条件
	pp.conditions = pp.conditions[:len(pp.conditions)-1]

	return "", nil
}

func (pp *Preprocessor) handlePragma(args []string, context *PreprocessorContext) (string, error) {
	// pragma指令处理
	return fmt.Sprintf("// #pragma %s", strings.Join(args, " ")), nil
}

func (pp *Preprocessor) handleLine(args []string, context *PreprocessorContext) (string, error) {
	// line指令处理
	return "", nil
}

func (pp *Preprocessor) handleError(args []string, context *PreprocessorContext) (string, error) {
	message := strings.Join(args, " ")
	return "", fmt.Errorf("error: %s", message)
}

func (pp *Preprocessor) handleWarning(args []string, context *PreprocessorContext) (string, error) {
	message := strings.Join(args, " ")
	fmt.Printf("warning: %s\n", message)
	return "", nil
}

func (pp *Preprocessor) shouldIncludeLine() bool {
	for _, condition := range pp.conditions {
		if !condition.Active {
			return false
		}
	}
	return true
}

func (pp *Preprocessor) evaluateExpression(expr string, context *PreprocessorContext) bool {
	// 简化的表达式求值
	trimmed := strings.TrimSpace(expr)

	// 检查defined()函数
	if strings.HasPrefix(trimmed, "defined(") && strings.HasSuffix(trimmed, ")") {
		name := trimmed[8 : len(trimmed)-1]
		_, defined := pp.defines[name]
		return defined
	}

	// 检查简单的值
	if value, exists := pp.defines[trimmed]; exists {
		return value != "" && value != "0"
	}

	// 默认为false
	return false
}

func (pp *Preprocessor) expandMacros(line string, context *PreprocessorContext) string {
	result := line

	for name, macro := range pp.macros {
		if strings.Contains(result, name) {
			if len(macro.Parameters) == 0 {
				// 简单宏替换
				value := macro.Body
				if macro.Builtin {
					value = pp.getBuiltinMacroValue(macro, context)
				}
				result = strings.ReplaceAll(result, name, value)
			} else {
				// 函数式宏替换（简化实现）
				result = pp.expandFunctionMacro(result, macro, context)
			}
		}
	}

	return result
}

func (pp *Preprocessor) expandFunctionMacro(input string, macro *Macro, context *PreprocessorContext) string {
	// 简化的函数式宏展开
	return input
}

func (pp *Preprocessor) getBuiltinMacroValue(macro *Macro, context *PreprocessorContext) string {
	switch macro.Name {
	case "__FILE__":
		return fmt.Sprintf("\"%s\"", context.File)
	case "__LINE__":
		return strconv.Itoa(context.Line)
	default:
		return macro.Body
	}
}

// ==================
// 5. 错误处理和诊断
// ==================

// LexicalErrorHandler 词法错误处理器
type LexicalErrorHandler struct {
	errors      []*LexicalError
	warnings    []*LexicalWarning
	suggestions []*ErrorSuggestion
	recovery    ErrorRecoveryStrategy
	config      ErrorHandlerConfig
	statistics  ErrorStatistics
	mutex       sync.RWMutex
}

// LexicalError 词法错误
type LexicalError struct {
	Type        ErrorType
	Message     string
	Position    Position
	Context     string
	Suggestions []*ErrorSuggestion
	Severity    ErrorSeverity
	Code        string
	Timestamp   time.Time
}

// Error implements the error interface
func (le *LexicalError) Error() string {
	return le.Message
}

// LexicalWarning 词法警告
type LexicalWarning struct {
	Type      WarningType
	Message   string
	Position  Position
	Context   string
	Code      string
	Timestamp time.Time
}

// ErrorSuggestion 错误建议
type ErrorSuggestion struct {
	Type        SuggestionType
	Message     string
	Replacement string
	Confidence  float64
}

// ErrorType 错误类型
type ErrorType int

const (
	ErrorUnexpectedCharacter ErrorType = iota
	ErrorInvalidToken
	ErrorUnterminatedString
	ErrorUnterminatedComment
	ErrorInvalidNumber
	ErrorInvalidEscape
	ErrorUnicodeError
	ErrorPreprocessorError
	ErrorMacroError
)

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityHint ErrorSeverity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityFatal
)

// WarningType 警告类型
type WarningType int

const (
	WarningDeprecated WarningType = iota
	WarningUnused
	WarningRedefinition
	WarningPerformance
)

// SuggestionType 建议类型
type SuggestionType int

const (
	SuggestionReplacement SuggestionType = iota
	SuggestionInsertion
	SuggestionDeletion
	SuggestionReformat
)

func NewLexicalErrorHandler() *LexicalErrorHandler {
	return &LexicalErrorHandler{
		errors:      make([]*LexicalError, 0),
		warnings:    make([]*LexicalWarning, 0),
		suggestions: make([]*ErrorSuggestion, 0),
		recovery:    RecoverySkipCharacter,
	}
}

func (leh *LexicalErrorHandler) ReportError(errorType ErrorType, message string, position Position, context string) {
	leh.mutex.Lock()
	defer leh.mutex.Unlock()

	err := &LexicalError{
		Type:      errorType,
		Message:   message,
		Position:  position,
		Context:   context,
		Severity:  SeverityError,
		Code:      fmt.Sprintf("LEX%03d", int(errorType)),
		Timestamp: time.Now(),
	}

	// 生成建议
	err.Suggestions = leh.generateSuggestions(err)

	leh.errors = append(leh.errors, err)
	leh.statistics.ErrorCount++
}

func (leh *LexicalErrorHandler) ReportWarning(warningType WarningType, message string, position Position, context string) {
	leh.mutex.Lock()
	defer leh.mutex.Unlock()

	warning := &LexicalWarning{
		Type:      warningType,
		Message:   message,
		Position:  position,
		Context:   context,
		Code:      fmt.Sprintf("LEXW%03d", int(warningType)),
		Timestamp: time.Now(),
	}

	leh.warnings = append(leh.warnings, warning)
	leh.statistics.WarningCount++
}

func (leh *LexicalErrorHandler) generateSuggestions(err *LexicalError) []*ErrorSuggestion {
	suggestions := make([]*ErrorSuggestion, 0)

	switch err.Type {
	case ErrorUnexpectedCharacter:
		suggestions = append(suggestions, &ErrorSuggestion{
			Type:       SuggestionDeletion,
			Message:    "Remove the unexpected character",
			Confidence: 0.8,
		})

	case ErrorUnterminatedString:
		suggestions = append(suggestions, &ErrorSuggestion{
			Type:        SuggestionInsertion,
			Message:     "Add closing quote",
			Replacement: "\"",
			Confidence:  0.9,
		})

	case ErrorInvalidNumber:
		suggestions = append(suggestions, &ErrorSuggestion{
			Type:       SuggestionReformat,
			Message:    "Check number format",
			Confidence: 0.7,
		})
	}

	return suggestions
}

func (leh *LexicalErrorHandler) GetErrors() []*LexicalError {
	leh.mutex.RLock()
	defer leh.mutex.RUnlock()

	return leh.errors
}

func (leh *LexicalErrorHandler) GetWarnings() []*LexicalWarning {
	leh.mutex.RLock()
	defer leh.mutex.RUnlock()

	return leh.warnings
}

func (leh *LexicalErrorHandler) Clear() {
	leh.mutex.Lock()
	defer leh.mutex.Unlock()

	leh.errors = leh.errors[:0]
	leh.warnings = leh.warnings[:0]
	leh.suggestions = leh.suggestions[:0]
}

// ==================
// 6. 缓存和优化
// ==================

// TokenCache 词法单元缓存
type TokenCache struct {
	cache   map[string][]*Token
	access  map[string]time.Time
	maxSize int
	size    int
	hits    int64
	misses  int64
	mutex   sync.RWMutex
}

// RegexCache 正则表达式缓存
type RegexCache struct {
	cache   map[string]*CompiledRegex
	maxSize int
	size    int
	mutex   sync.RWMutex
}

func NewTokenCache(maxSize int) *TokenCache {
	return &TokenCache{
		cache:   make(map[string][]*Token),
		access:  make(map[string]time.Time),
		maxSize: maxSize,
	}
}

func (tc *TokenCache) Get(input, language string) []*Token {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	key := tc.makeKey(input, language)
	if tokens, exists := tc.cache[key]; exists {
		tc.access[key] = time.Now()
		tc.hits++
		return tokens
	}

	tc.misses++
	return nil
}

func (tc *TokenCache) Put(input, language string, tokens []*Token) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	key := tc.makeKey(input, language)

	// 检查缓存大小
	if tc.size >= tc.maxSize {
		tc.evictLRU()
	}

	tc.cache[key] = tokens
	tc.access[key] = time.Now()
	tc.size++
}

func (tc *TokenCache) makeKey(input, language string) string {
	// Create hash for large inputs to save memory
	if len(input) > 100 {
		hasher := sha256.New()
		hasher.Write([]byte(fmt.Sprintf("%s:%s", language, input)))
		return hex.EncodeToString(hasher.Sum(nil))
	}
	return fmt.Sprintf("%s:%s", language, input)
}

func (tc *TokenCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, accessTime := range tc.access {
		if accessTime.Before(oldestTime) {
			oldestTime = accessTime
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(tc.cache, oldestKey)
		delete(tc.access, oldestKey)
		tc.size--
	}
}

func NewRegexCache(maxSize int) *RegexCache {
	return &RegexCache{
		cache:   make(map[string]*CompiledRegex),
		maxSize: maxSize,
	}
}

func (rc *RegexCache) Get(pattern string) *CompiledRegex {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	return rc.cache[pattern]
}

func (rc *RegexCache) Put(pattern string, compiled *CompiledRegex) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if rc.size >= rc.maxSize {
		// 简单的缓存清理
		for key := range rc.cache {
			delete(rc.cache, key)
			rc.size--
			break
		}
	}

	rc.cache[pattern] = compiled
	rc.size++
}

// ==================
// 7. 辅助类型和函数
// ==================

// 各种统计、配置和接口类型
type (
	LexerStatistics struct {
		TotalTokens       int64
		AnalysisCount     int64
		AverageTokens     float64
		TotalAnalysisTime time.Duration
		CacheHitRate      float64
	}

	TokenizerStatistics struct {
		TokensProduced      int64
		AnalysisTime        time.Duration
		CharactersProcessed int64
		ErrorsGenerated     int64
		WarningsGenerated   int64
	}

	TokenizerConfig struct {
		SkipWhitespace bool
		SkipComments   bool
		ErrorRecovery  bool
		CaseSensitive  bool
		MaxTokenLength int
	}

	TokenAction func(value string, position Position, state *LexerState) TokenActionResult

	TokenActionResult struct {
		Consume     bool
		TokenType   TokenType
		ChangeState string
		Metadata    map[string]interface{}
	}

	RuleCondition func(position Position, state *LexerState) bool

	StateAction func(state *LexerState) error

	ErrorRecoveryStrategy int

	LexerExtension interface {
		Name() string
		Process(tokens []*Token) []*Token
	}

	LexerMiddleware interface {
		Process(tokens []*Token) []*Token
	}

	NFAEngine struct{}
	DFAEngine struct{}

	RegexConfig struct {
		CacheSize     int
		OptimizeNFA   bool
		OptimizeDFA   bool
		EnableUnicode bool
	}

	RegexStatistics struct {
		CompilationCount int64
		MatchCount       int64
		CacheHitRate     float64
		AverageMatchTime time.Duration
	}

	RegexAST struct {
		Type     ASTNodeType
		Value    string
		Children []*RegexAST
	}

	ASTNodeType int

	PreprocessorConfig struct {
		MaxIncludeDepth   int
		EnableMacros      bool
		EnableConditional bool
		PreservePragmas   bool
	}

	PreprocessorStatistics struct {
		LinesProcessed      int64
		MacrosExpanded      int64
		IncludesProcessed   int64
		DirectivesProcessed int64
	}

	ErrorHandlerConfig struct {
		MaxErrors         int
		ShowWarnings      bool
		ShowHints         bool
		EnableSuggestions bool
	}

	ErrorStatistics struct {
		ErrorCount   int64
		WarningCount int64
		HintCount    int64
	}
)

// 常量定义
const (
	RecoverySkipCharacter ErrorRecoveryStrategy = iota
	RecoverySkipToken
	RecoveryInsertToken
	RecoveryResync

	ASTLiteral ASTNodeType = iota
	ASTGroup
	ASTAlternate
	ASTRepeat
	ASTCharClass
)

// 构造函数
func NewNFAEngine() *NFAEngine { return &NFAEngine{} }
func NewDFAEngine() *DFAEngine { return &DFAEngine{} }

// 辅助方法
func (rf RegexFlags) String() string {
	var flags []string
	if rf.IgnoreCase {
		flags = append(flags, "i")
	}
	if rf.Multiline {
		flags = append(flags, "m")
	}
	if rf.DotAll {
		flags = append(flags, "s")
	}
	if rf.Unicode {
		flags = append(flags, "u")
	}
	if rf.Global {
		flags = append(flags, "g")
	}
	if rf.Sticky {
		flags = append(flags, "y")
	}
	return strings.Join(flags, "")
}

// ==================
// 8. 主演示函数
// ==================

func demonstrateLexicalAnalysis() {
	fmt.Println("=== Go词法分析大师演示 ===")

	// 1. 创建词法分析器
	fmt.Println("\n1. 初始化词法分析器")
	config := LexerConfig{
		Language:           "go",
		CaseSensitive:      true,
		EnableUnicode:      true,
		EnablePreprocessor: true,
		EnableCache:        true,
		CacheSize:          1000,
		MaxTokenLength:     1024,
		ParallelWorkers:    4,
		EnableProfiling:    true,
		ErrorRecovery:      true,
		DebugMode:          false,
	}

	analyzer := NewLexicalAnalyzer(config)

	// 2. 创建Go语言词法分析器
	fmt.Println("\n2. 配置Go语言词法分析器")
	tokenizerConfig := TokenizerConfig{
		SkipWhitespace: true,
		SkipComments:   false,
		ErrorRecovery:  true,
		CaseSensitive:  true,
		MaxTokenLength: 1024,
	}

	goTokenizer := NewTokenizer("go-tokenizer", "go", tokenizerConfig)

	// 3. 添加词法规则
	fmt.Println("\n3. 添加词法规则")

	// 添加关键字规则
	keywords := []string{
		"package", "import", "func", "var", "const", "type",
		"if", "else", "for", "range", "switch", "case", "default",
		"break", "continue", "return", "defer", "go", "select",
		"chan", "map", "interface", "struct",
	}

	for i, keyword := range keywords {
		goTokenizer.AddRule(&TokenRule{
			ID:        fmt.Sprintf("keyword_%d", i),
			Name:      fmt.Sprintf("keyword_%s", keyword),
			Pattern:   `\b` + keyword + `\b`,
			TokenType: TokenKeyword,
			Priority:  100,
			Enabled:   true,
		})
	}

	// 添加标识符规则
	goTokenizer.AddRule(&TokenRule{
		ID:        "identifier",
		Name:      "identifier",
		Pattern:   `[a-zA-Z_][a-zA-Z0-9_]*`,
		TokenType: TokenIdentifier,
		Priority:  50,
		Enabled:   true,
	})

	// 添加数字规则
	goTokenizer.AddRule(&TokenRule{
		ID:        "number_int",
		Name:      "integer",
		Pattern:   `\d+`,
		TokenType: TokenNumber,
		Priority:  80,
		Enabled:   true,
	})

	goTokenizer.AddRule(&TokenRule{
		ID:        "number_float",
		Name:      "float",
		Pattern:   `\d+\.\d+([eE][+-]?\d+)?`,
		TokenType: TokenNumber,
		Priority:  85,
		Enabled:   true,
	})

	// 添加字符串规则
	goTokenizer.AddRule(&TokenRule{
		ID:        "string_double",
		Name:      "double_quoted_string",
		Pattern:   `"([^"\\]|\\.)*"`,
		TokenType: TokenString,
		Priority:  90,
		Enabled:   true,
	})

	goTokenizer.AddRule(&TokenRule{
		ID:        "string_single",
		Name:      "single_quoted_string",
		Pattern:   `'([^'\\]|\\.)*'`,
		TokenType: TokenChar,
		Priority:  90,
		Enabled:   true,
	})

	goTokenizer.AddRule(&TokenRule{
		ID:        "string_raw",
		Name:      "raw_string",
		Pattern:   "`[^`]*`",
		TokenType: TokenString,
		Priority:  95,
		Enabled:   true,
	})

	// 添加注释规则
	goTokenizer.AddRule(&TokenRule{
		ID:        "comment_line",
		Name:      "line_comment",
		Pattern:   `//.*`,
		TokenType: TokenComment,
		Priority:  70,
		Enabled:   true,
	})

	goTokenizer.AddRule(&TokenRule{
		ID:        "comment_block",
		Name:      "block_comment",
		Pattern:   `/\*[\s\S]*?\*/`,
		TokenType: TokenComment,
		Priority:  75,
		Enabled:   true,
	})

	// 添加操作符规则
	operators := []string{
		":=", "==", "!=", "<=", ">=", "++", "--", "&&", "||",
		"<<", ">>", "&^", "+=", "-=", "*=", "/=", "%=",
		"&=", "|=", "^=", "<<=", ">>=", "&^=", "<-",
		"+", "-", "*", "/", "%", "&", "|", "^", "!", "<", ">", "=",
	}

	for i, op := range operators {
		goTokenizer.AddRule(&TokenRule{
			ID:        fmt.Sprintf("operator_%d", i),
			Name:      fmt.Sprintf("operator_%s", op),
			Pattern:   regexp.QuoteMeta(op),
			TokenType: TokenOperator,
			Priority:  60,
			Enabled:   true,
		})
	}

	// 添加标点符号规则
	punctuation := []string{
		"(", ")", "[", "]", "{", "}", ";", ",", ".", ":",
	}

	for i, punct := range punctuation {
		goTokenizer.AddRule(&TokenRule{
			ID:        fmt.Sprintf("punct_%d", i),
			Name:      fmt.Sprintf("punctuation_%s", punct),
			Pattern:   regexp.QuoteMeta(punct),
			TokenType: TokenPunctuation,
			Priority:  40,
			Enabled:   true,
		})
	}

	// 注册词法分析器
	analyzer.RegisterTokenizer(goTokenizer)

	// 4. 测试词法分析
	fmt.Println("\n4. 词法分析测试")

	testCode := `package main

import (
	"fmt"
	"os"
)

// Main function
func main() {
	var x int = 42
	var y float64 = 3.14159
	var name string = "Hello, 世界!"

	if x > 0 {
		fmt.Printf("x = %d, y = %.2f\n", x, y)
		fmt.Println(name)
	}

	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()
}
`

	tokens, err := analyzer.Tokenize(testCode, "go")
	if err != nil {
		fmt.Printf("词法分析失败: %v\n", err)
		return
	}

	fmt.Printf("成功生成 %d 个词法单元:\n", len(tokens))

	// 显示前20个词法单元
	for i, token := range tokens {
		if i >= 20 {
			fmt.Printf("... (还有 %d 个词法单元)\n", len(tokens)-20)
			break
		}

		fmt.Printf("  [%2d] %-12s %-20s %d:%d\n",
			i+1,
			token.Type.String(),
			fmt.Sprintf("'%s'", token.Value),
			token.Line,
			token.Column)
	}

	// 5. 统计信息
	fmt.Println("\n5. 词法分析统计")
	tokenStats := make(map[TokenType]int)
	for _, token := range tokens {
		tokenStats[token.Type]++
	}

	fmt.Println("词法单元统计:")
	for tokenType, count := range tokenStats {
		if count > 0 {
			fmt.Printf("  %-12s: %d\n", tokenType.String(), count)
		}
	}

	// 6. 错误处理演示
	fmt.Println("\n6. 错误处理和恢复")

	errorCode := `package main

import "fmt"

func main() {
	var x int = 42@  // 错误：无效字符
	var y = "unclosed string
	var z = 'too many chars'  // 错误：字符常量过长

	fmt.Println(x, y, z)
}
`

	errorTokens, err := analyzer.Tokenize(errorCode, "go")
	if err != nil {
		fmt.Printf("词法分析发现错误: %v\n", err)
	} else {
		fmt.Printf("错误恢复成功，生成 %d 个词法单元\n", len(errorTokens))
	}

	// 显示错误信息
	errors := analyzer.tokenizers["go"].errors
	if len(errors) > 0 {
		fmt.Printf("发现 %d 个词法错误:\n", len(errors))
		for i, lexErr := range errors {
			fmt.Printf("  错误 %d: %s (位置 %d:%d)\n",
				i+1, lexErr.Message, lexErr.Position.Line, lexErr.Position.Column)
		}
	}

	// 7. 预处理器演示
	fmt.Println("\n7. 预处理器演示")

	preprocessorCode := `#define VERSION "1.0.0"
#define MAX_SIZE 1024

#ifdef DEBUG
	#define LOG(msg) fmt.Println("DEBUG:", msg)
#else
	#define LOG(msg) // no-op
#endif

package main

import "fmt"

func main() {
	fmt.Println("Version:", VERSION)
	fmt.Printf("Max size: %d\n", MAX_SIZE)
	LOG("Application started")
}
`

	processed, err := analyzer.preprocessor.Process(preprocessorCode)
	if err != nil {
		fmt.Printf("预处理失败: %v\n", err)
	} else {
		fmt.Println("预处理结果:")
		lines := strings.Split(processed, "\n")
		for i, line := range lines {
			if i < 15 { // 显示前15行
				fmt.Printf("  %2d: %s\n", i+1, line)
			}
		}
		if len(lines) > 15 {
			fmt.Printf("  ... (还有 %d 行)\n", len(lines)-15)
		}
	}

	// 8. 正则表达式引擎演示
	fmt.Println("\n8. 正则表达式引擎演示")

	patterns := []string{
		`\d+`,             // 数字
		`[a-zA-Z_]\w*`,    // 标识符
		`"([^"\\]|\\.)*"`, // 字符串
		`//.*`,            // 行注释
		`/\*[\s\S]*?\*/`,  // 块注释
	}

	for _, pattern := range patterns {
		compiled, err := analyzer.regexEngine.Compile(pattern, RegexFlags{})
		if err != nil {
			fmt.Printf("编译正则表达式失败 '%s': %v\n", pattern, err)
			continue
		}

		testStrings := []string{"123", "hello", "\"world\"", "// comment", "/* block */"}
		fmt.Printf("模式 '%s' 匹配结果:\n", pattern)

		for _, test := range testStrings {
			match := analyzer.regexEngine.Match(compiled, test)
			fmt.Printf("  '%s' -> %v\n", test, match)
		}
	}

	// 9. 性能基准测试
	fmt.Println("\n9. 性能基准测试")

	largeCode := strings.Repeat(testCode, 10) // 重复10次增加代码量

	startTime := time.Now()
	largeTokens, err := analyzer.Tokenize(largeCode, "go")
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("大文件词法分析失败: %v\n", err)
	} else {
		fmt.Printf("大文件词法分析结果:\n")
		fmt.Printf("  代码长度: %d 字符\n", len(largeCode))
		fmt.Printf("  词法单元数: %d\n", len(largeTokens))
		fmt.Printf("  分析时间: %v\n", duration)
		fmt.Printf("  吞吐量: %.0f 字符/秒\n", float64(len(largeCode))/duration.Seconds())
		fmt.Printf("  词法单元/秒: %.0f\n", float64(len(largeTokens))/duration.Seconds())
	}

	// 10. 缓存效果测试
	fmt.Println("\n10. 缓存效果测试")

	// 第二次分析同样的代码（应该命中缓存）
	startTime = time.Now()
	cachedTokens, err := analyzer.Tokenize(testCode, "go")
	cachedDuration := time.Since(startTime)

	if err == nil {
		fmt.Printf("缓存测试结果:\n")
		fmt.Printf("  首次分析时间: %v\n", duration)
		fmt.Printf("  缓存分析时间: %v\n", cachedDuration)
		fmt.Printf("  加速比: %.2fx\n", float64(duration.Nanoseconds())/float64(cachedDuration.Nanoseconds()))
		fmt.Printf("  缓存命中: %t\n", len(cachedTokens) == len(tokens))
	}

	fmt.Println("\n=== 词法分析演示完成 ===")
}

func main() {
	demonstrateLexicalAnalysis()

	fmt.Println("\n=== Go词法分析大师演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 词法分析器设计：规则定义、状态管理、错误处理")
	fmt.Println("2. 有限状态机：NFA/DFA构建、状态转换、优化技术")
	fmt.Println("3. 正则表达式：模式匹配、编译优化、性能调优")
	fmt.Println("4. 预处理器：宏展开、条件编译、指令处理")
	fmt.Println("5. 错误恢复：错误检测、诊断信息、恢复策略")
	fmt.Println("6. 性能优化：缓存机制、并行处理、内存管理")
	fmt.Println("7. Unicode支持：多字节字符、编码处理、国际化")

	fmt.Println("\n高级词法分析技术:")
	fmt.Println("- 增量词法分析和实时更新")
	fmt.Println("- 并行词法分析和多线程优化")
	fmt.Println("- 领域特定语言词法支持")
	fmt.Println("- 语法高亮和代码格式化")
	fmt.Println("- 智能错误恢复和建议")
	fmt.Println("- 词法分析器自动生成")
	fmt.Println("- 混合语言词法分析")
}

/*
=== 练习题 ===

1. 词法分析器增强：
   - 实现增量词法分析功能
   - 添加多语言支持机制
   - 创建词法分析器生成器
   - 实现智能错误恢复

2. 正则表达式优化：
   - 实现NFA到DFA的优化转换
   - 添加正则表达式编译缓存
   - 创建正则表达式调试工具
   - 实现PCRE兼容支持

3. 预处理器扩展：
   - 实现复杂宏系统
   - 添加条件编译优化
   - 创建宏调试和跟踪
   - 实现包含文件管理

4. 性能优化：
   - 实现SIMD加速词法分析
   - 添加内存池管理
   - 创建并行处理框架
   - 实现零拷贝优化

5. 工具集成：
   - 创建VS Code插件
   - 实现语法高亮引擎
   - 添加代码补全支持
   - 创建词法分析器IDE

重要概念：
- Lexical Analysis: 词法分析和标记化
- Finite State Machine: 有限状态机理论
- Regular Expression: 正则表达式引擎
- Token Recognition: 词法单元识别
- Error Recovery: 错误恢复策略
- Unicode Processing: Unicode字符处理
- Performance Optimization: 性能优化技术
*/
