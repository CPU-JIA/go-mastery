package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"

	"go-mastery/common/security"
)

/*
微服务架构 - 配置管理和分布式配置练习

本练习涵盖微服务架构中的配置管理，包括：
1. 配置中心（Configuration Center）
2. 环境特定配置（Environment-specific Config）
3. 动态配置更新（Dynamic Configuration）
4. 配置验证和类型安全
5. 密钥管理和加密
6. 配置版本控制
7. 配置热重载
8. 分布式配置一致性

主要概念：
- 配置外部化
- 配置分层和继承
- 配置监听和变更通知
- 配置安全和加密
- 配置审计和版本控制
*/

// === 配置模型定义 ===

// ConfigItem 配置项
type ConfigItem struct {
	Key         string      `json:"key" yaml:"key"`
	Value       interface{} `json:"value" yaml:"value"`
	Type        string      `json:"type" yaml:"type"` // string, int, bool, float, json
	Environment string      `json:"environment" yaml:"environment"`
	Application string      `json:"application" yaml:"application"`
	Group       string      `json:"group" yaml:"group"`
	Description string      `json:"description" yaml:"description"`
	IsSecret    bool        `json:"is_secret" yaml:"is_secret"`
	Version     int64       `json:"version" yaml:"version"`
	CreatedAt   time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" yaml:"updated_at"`
	CreatedBy   string      `json:"created_by" yaml:"created_by"`
	UpdatedBy   string      `json:"updated_by" yaml:"updated_by"`
}

// Configuration 配置集合
type Configuration struct {
	Application string                 `json:"application" yaml:"application"`
	Environment string                 `json:"environment" yaml:"environment"`
	Version     int64                  `json:"version" yaml:"version"`
	Items       map[string]*ConfigItem `json:"items" yaml:"items"`
	Metadata    map[string]string      `json:"metadata" yaml:"metadata"`
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
}

// ConfigurationSource 配置源接口
type ConfigurationSource interface {
	Load(app, env string) (*Configuration, error)
	Save(config *Configuration) error
	Watch(app, env string) (<-chan *Configuration, error)
	GetHistory(app, env string, limit int) ([]*Configuration, error)
}

// === 内存配置源实现 ===

type MemoryConfigurationSource struct {
	configs   map[string]*Configuration // key: app:env
	watchers  map[string][]chan *Configuration
	mutex     sync.RWMutex
	encryptor *ConfigEncryptor
}

func NewMemoryConfigurationSource(encryptionKey string) *MemoryConfigurationSource {
	source := &MemoryConfigurationSource{
		configs:   make(map[string]*Configuration),
		watchers:  make(map[string][]chan *Configuration),
		encryptor: NewConfigEncryptor(encryptionKey),
	}

	// 初始化示例配置
	source.initSampleConfigs()

	return source
}

func (m *MemoryConfigurationSource) initSampleConfigs() {
	configs := []*Configuration{
		{
			Application: "user-service",
			Environment: "development",
			Version:     1,
			Items: map[string]*ConfigItem{
				"database.host": {
					Key:         "database.host",
					Value:       "localhost",
					Type:        "string",
					Environment: "development",
					Application: "user-service",
					Group:       "database",
					Description: "数据库主机地址",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"database.port": {
					Key:         "database.port",
					Value:       5432,
					Type:        "int",
					Environment: "development",
					Application: "user-service",
					Group:       "database",
					Description: "数据库端口",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"database.password": {
					Key:         "database.password",
					Value:       "encrypted:secret123",
					Type:        "string",
					Environment: "development",
					Application: "user-service",
					Group:       "database",
					Description: "数据库密码",
					IsSecret:    true,
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"cache.enabled": {
					Key:         "cache.enabled",
					Value:       true,
					Type:        "bool",
					Environment: "development",
					Application: "user-service",
					Group:       "cache",
					Description: "是否启用缓存",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"server.port": {
					Key:         "server.port",
					Value:       8080,
					Type:        "int",
					Environment: "development",
					Application: "user-service",
					Group:       "server",
					Description: "服务器端口",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			UpdatedAt: time.Now(),
		},
		{
			Application: "user-service",
			Environment: "production",
			Version:     1,
			Items: map[string]*ConfigItem{
				"database.host": {
					Key:         "database.host",
					Value:       "prod-db.example.com",
					Type:        "string",
					Environment: "production",
					Application: "user-service",
					Group:       "database",
					Description: "数据库主机地址",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"database.port": {
					Key:         "database.port",
					Value:       5432,
					Type:        "int",
					Environment: "production",
					Application: "user-service",
					Group:       "database",
					Description: "数据库端口",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"cache.enabled": {
					Key:         "cache.enabled",
					Value:       true,
					Type:        "bool",
					Environment: "production",
					Application: "user-service",
					Group:       "cache",
					Description: "是否启用缓存",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				"server.port": {
					Key:         "server.port",
					Value:       80,
					Type:        "int",
					Environment: "production",
					Application: "user-service",
					Group:       "server",
					Description: "服务器端口",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			UpdatedAt: time.Now(),
		},
	}

	for _, config := range configs {
		key := fmt.Sprintf("%s:%s", config.Application, config.Environment)
		m.configs[key] = config
	}
}

func (m *MemoryConfigurationSource) Load(app, env string) (*Configuration, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := fmt.Sprintf("%s:%s", app, env)
	config, exists := m.configs[key]
	if !exists {
		return nil, fmt.Errorf("配置不存在: %s", key)
	}

	// 解密敏感配置
	decryptedConfig := m.decryptConfig(config)

	return decryptedConfig, nil
}

func (m *MemoryConfigurationSource) Save(config *Configuration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 加密敏感配置
	encryptedConfig := m.encryptConfig(config)

	// 更新版本和时间戳
	encryptedConfig.Version++
	encryptedConfig.UpdatedAt = time.Now()

	key := fmt.Sprintf("%s:%s", config.Application, config.Environment)
	m.configs[key] = encryptedConfig

	// 通知监听者
	m.notifyWatchers(key, encryptedConfig)

	return nil
}

func (m *MemoryConfigurationSource) Watch(app, env string) (<-chan *Configuration, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", app, env)
	watcher := make(chan *Configuration, 10)

	if m.watchers[key] == nil {
		m.watchers[key] = make([]chan *Configuration, 0)
	}
	m.watchers[key] = append(m.watchers[key], watcher)

	return watcher, nil
}

func (m *MemoryConfigurationSource) GetHistory(app, env string, limit int) ([]*Configuration, error) {
	// 简化实现，实际应该存储历史版本
	config, err := m.Load(app, env)
	if err != nil {
		return nil, err
	}

	return []*Configuration{config}, nil
}

func (m *MemoryConfigurationSource) notifyWatchers(key string, config *Configuration) {
	watchers, exists := m.watchers[key]
	if !exists {
		return
	}

	decryptedConfig := m.decryptConfig(config)

	for _, watcher := range watchers {
		select {
		case watcher <- decryptedConfig:
		default:
			// 如果channel满了，跳过
		}
	}
}

func (m *MemoryConfigurationSource) encryptConfig(config *Configuration) *Configuration {
	encrypted := &Configuration{
		Application: config.Application,
		Environment: config.Environment,
		Version:     config.Version,
		Items:       make(map[string]*ConfigItem),
		Metadata:    config.Metadata,
		UpdatedAt:   config.UpdatedAt,
	}

	for key, item := range config.Items {
		newItem := *item
		if item.IsSecret {
			encryptedValue, _ := m.encryptor.Encrypt(fmt.Sprintf("%v", item.Value))
			newItem.Value = "encrypted:" + encryptedValue
		}
		encrypted.Items[key] = &newItem
	}

	return encrypted
}

func (m *MemoryConfigurationSource) decryptConfig(config *Configuration) *Configuration {
	decrypted := &Configuration{
		Application: config.Application,
		Environment: config.Environment,
		Version:     config.Version,
		Items:       make(map[string]*ConfigItem),
		Metadata:    config.Metadata,
		UpdatedAt:   config.UpdatedAt,
	}

	for key, item := range config.Items {
		newItem := *item
		if item.IsSecret {
			valueStr := fmt.Sprintf("%v", item.Value)
			if strings.HasPrefix(valueStr, "encrypted:") {
				encryptedData := strings.TrimPrefix(valueStr, "encrypted:")
				if decryptedValue, err := m.encryptor.Decrypt(encryptedData); err == nil {
					newItem.Value = decryptedValue
				}
			}
		}
		decrypted.Items[key] = &newItem
	}

	return decrypted
}

// === 配置加密器 ===

type ConfigEncryptor struct {
	key []byte
}

func NewConfigEncryptor(keyString string) *ConfigEncryptor {
	hash := sha256.Sum256([]byte(keyString))
	return &ConfigEncryptor{key: hash[:]}
}

func (e *ConfigEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *ConfigEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// === 配置客户端 ===

type ConfigClient struct {
	source      ConfigurationSource
	application string
	environment string
	config      *Configuration
	watchers    []ConfigWatcher
	mutex       sync.RWMutex
}

type ConfigWatcher interface {
	OnConfigChanged(config *Configuration)
}

func NewConfigClient(source ConfigurationSource, app, env string) *ConfigClient {
	return &ConfigClient{
		source:      source,
		application: app,
		environment: env,
		watchers:    make([]ConfigWatcher, 0),
	}
}

func (c *ConfigClient) Load() error {
	config, err := c.source.Load(c.application, c.environment)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.config = config
	c.mutex.Unlock()

	return nil
}

func (c *ConfigClient) StartWatching() error {
	watcher, err := c.source.Watch(c.application, c.environment)
	if err != nil {
		return err
	}

	go func() {
		for config := range watcher {
			c.mutex.Lock()
			c.config = config
			c.mutex.Unlock()

			// 通知所有监听者
			for _, w := range c.watchers {
				w.OnConfigChanged(config)
			}

			log.Printf("配置已更新: %s:%s version %d",
				c.application, c.environment, config.Version)
		}
	}()

	return nil
}

func (c *ConfigClient) AddWatcher(watcher ConfigWatcher) {
	c.watchers = append(c.watchers, watcher)
}

func (c *ConfigClient) GetString(key string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.config == nil {
		return ""
	}

	item, exists := c.config.Items[key]
	if !exists {
		return ""
	}

	return fmt.Sprintf("%v", item.Value)
}

func (c *ConfigClient) GetInt(key string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.config == nil {
		return 0
	}

	item, exists := c.config.Items[key]
	if !exists {
		return 0
	}

	switch v := item.Value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return 0
}

func (c *ConfigClient) GetBool(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.config == nil {
		return false
	}

	item, exists := c.config.Items[key]
	if !exists {
		return false
	}

	switch v := item.Value.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	}

	return false
}

func (c *ConfigClient) GetFloat(key string) float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.config == nil {
		return 0.0
	}

	item, exists := c.config.Items[key]
	if !exists {
		return 0.0
	}

	switch v := item.Value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return 0.0
}

func (c *ConfigClient) GetJSON(key string, target interface{}) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.config == nil {
		return fmt.Errorf("配置未加载")
	}

	item, exists := c.config.Items[key]
	if !exists {
		return fmt.Errorf("配置项不存在: %s", key)
	}

	jsonStr := fmt.Sprintf("%v", item.Value)
	return json.Unmarshal([]byte(jsonStr), target)
}

// === 配置管理服务 ===

type ConfigServer struct {
	source   ConfigurationSource
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

func NewConfigServer(source ConfigurationSource) *ConfigServer {
	return &ConfigServer{
		source: source,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// 获取配置
func (s *ConfigServer) GetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	app := vars["app"]
	env := vars["env"]

	config, err := s.source.Load(app, env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// 更新配置
func (s *ConfigServer) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "无效的配置数据", http.StatusBadRequest)
		return
	}

	if err := s.source.Save(&config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 广播配置更新
	s.broadcastConfigUpdate(&config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// 更新单个配置项
func (s *ConfigServer) UpdateConfigItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	app := vars["app"]
	env := vars["env"]
	key := vars["key"]

	var item ConfigItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "无效的配置项数据", http.StatusBadRequest)
		return
	}

	// 加载现有配置
	config, err := s.source.Load(app, env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 更新配置项
	item.Key = key
	item.Application = app
	item.Environment = env
	item.UpdatedAt = time.Now()
	config.Items[key] = &item

	// 保存配置
	if err := s.source.Save(config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// WebSocket连接处理
func (s *ConfigServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	s.mutex.Lock()
	s.clients[conn] = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.clients, conn)
		s.mutex.Unlock()
	}()

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *ConfigServer) broadcastConfigUpdate(config *Configuration) {
	message, _ := json.Marshal(map[string]interface{}{
		"type":   "config_update",
		"config": config,
	})

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for conn := range s.clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// 获取配置历史
func (s *ConfigServer) GetConfigHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	app := vars["app"]
	env := vars["env"]

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	history, err := s.source.GetHistory(app, env, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// === 配置验证器 ===

type ConfigValidator struct {
	schemas map[string]ConfigSchema
}

type ConfigSchema struct {
	Application string                    `json:"application"`
	Environment string                    `json:"environment"`
	Fields      map[string]FieldValidator `json:"fields"`
}

type FieldValidator struct {
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	MinValue     *float64    `json:"min_value"`
	MaxValue     *float64    `json:"max_value"`
	Pattern      string      `json:"pattern"`
	Options      []string    `json:"options"`
}

func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		schemas: make(map[string]ConfigSchema),
	}

	// 初始化示例配置模式
	validator.initSchemas()

	return validator
}

func (v *ConfigValidator) initSchemas() {
	userServiceSchema := ConfigSchema{
		Application: "user-service",
		Environment: "*",
		Fields: map[string]FieldValidator{
			"database.host": {
				Type:     "string",
				Required: true,
			},
			"database.port": {
				Type:     "int",
				Required: true,
				MinValue: func(f float64) *float64 { return &f }(1),
				MaxValue: func(f float64) *float64 { return &f }(65535),
			},
			"server.port": {
				Type:     "int",
				Required: true,
				MinValue: func(f float64) *float64 { return &f }(1),
				MaxValue: func(f float64) *float64 { return &f }(65535),
			},
			"cache.enabled": {
				Type:         "bool",
				Required:     false,
				DefaultValue: false,
			},
		},
	}

	v.schemas["user-service:*"] = userServiceSchema
}

func (v *ConfigValidator) Validate(config *Configuration) []string {
	var errors []string

	// 查找匹配的模式
	key := fmt.Sprintf("%s:%s", config.Application, config.Environment)
	schema, exists := v.schemas[key]
	if !exists {
		// 尝试通用模式
		key = fmt.Sprintf("%s:*", config.Application)
		schema, exists = v.schemas[key]
		if !exists {
			return errors
		}
	}

	// 验证必需字段
	for fieldName, validator := range schema.Fields {
		item, exists := config.Items[fieldName]
		if !exists {
			if validator.Required {
				errors = append(errors, fmt.Sprintf("必需字段缺失: %s", fieldName))
			}
			continue
		}

		// 类型验证
		if err := v.validateType(item.Value, validator.Type); err != nil {
			errors = append(errors, fmt.Sprintf("字段 %s 类型错误: %s", fieldName, err.Error()))
		}

		// 数值范围验证
		if validator.MinValue != nil || validator.MaxValue != nil {
			if err := v.validateRange(item.Value, validator.MinValue, validator.MaxValue); err != nil {
				errors = append(errors, fmt.Sprintf("字段 %s 数值范围错误: %s", fieldName, err.Error()))
			}
		}

		// 选项验证
		if len(validator.Options) > 0 {
			if err := v.validateOptions(item.Value, validator.Options); err != nil {
				errors = append(errors, fmt.Sprintf("字段 %s 选项错误: %s", fieldName, err.Error()))
			}
		}
	}

	return errors
}

func (v *ConfigValidator) validateType(value interface{}, expectedType string) error {
	valueType := reflect.TypeOf(value).Kind()

	switch expectedType {
	case "string":
		if valueType != reflect.String {
			return fmt.Errorf("期望string类型，实际为%s", valueType)
		}
	case "int":
		if valueType != reflect.Int && valueType != reflect.Float64 {
			return fmt.Errorf("期望int类型，实际为%s", valueType)
		}
	case "bool":
		if valueType != reflect.Bool {
			return fmt.Errorf("期望bool类型，实际为%s", valueType)
		}
	case "float":
		if valueType != reflect.Float64 && valueType != reflect.Int {
			return fmt.Errorf("期望float类型，实际为%s", valueType)
		}
	}

	return nil
}

func (v *ConfigValidator) validateRange(value interface{}, min, max *float64) error {
	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		return fmt.Errorf("无法进行数值范围验证")
	}

	if min != nil && numValue < *min {
		return fmt.Errorf("值 %v 小于最小值 %v", numValue, *min)
	}

	if max != nil && numValue > *max {
		return fmt.Errorf("值 %v 大于最大值 %v", numValue, *max)
	}

	return nil
}

func (v *ConfigValidator) validateOptions(value interface{}, options []string) error {
	valueStr := fmt.Sprintf("%v", value)
	for _, option := range options {
		if valueStr == option {
			return nil
		}
	}

	return fmt.Errorf("值 %s 不在允许的选项中: %v", valueStr, options)
}

// === 文件配置源实现 ===

type FileConfigurationSource struct {
	basePath  string
	encryptor *ConfigEncryptor
}

func NewFileConfigurationSource(basePath, encryptionKey string) *FileConfigurationSource {
	// #nosec G301 -- 教学示例代码，配置服务数据目录需要0755权限支持文件读写
	os.MkdirAll(basePath, 0755)
	return &FileConfigurationSource{
		basePath:  basePath,
		encryptor: NewConfigEncryptor(encryptionKey),
	}
}

func (f *FileConfigurationSource) Load(app, env string) (*Configuration, error) {
	filename := filepath.Join(f.basePath, fmt.Sprintf("%s-%s.yaml", app, env))

	// #nosec G304 -- 配置服务内部操作，路径由basePath和受控的app/env参数构建，basePath在初始化时设定
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Configuration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (f *FileConfigurationSource) Save(config *Configuration) error {
	filename := filepath.Join(f.basePath, fmt.Sprintf("%s-%s.yaml", config.Application, config.Environment))

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return security.SecureWriteFile(filename, data, &security.SecureFileOptions{
		Mode:      security.GetRecommendedMode("configuration"),
		CreateDir: true,
	})
}

func (f *FileConfigurationSource) Watch(app, env string) (<-chan *Configuration, error) {
	// 简化实现，实际应该使用文件系统监听
	watcher := make(chan *Configuration)
	return watcher, nil
}

func (f *FileConfigurationSource) GetHistory(app, env string, limit int) ([]*Configuration, error) {
	// 简化实现
	config, err := f.Load(app, env)
	if err != nil {
		return nil, err
	}

	return []*Configuration{config}, nil
}

// === 示例应用配置监听器 ===

type AppConfigWatcher struct {
	name string
}

func (w *AppConfigWatcher) OnConfigChanged(config *Configuration) {
	log.Printf("[%s] 配置已更新，版本: %d", w.name, config.Version)

	// 这里可以实现配置变更的具体处理逻辑
	for key, item := range config.Items {
		log.Printf("  %s = %v", key, item.Value)
	}
}

func main() {
	// 创建配置源
	source := NewMemoryConfigurationSource("config-encryption-key-123")

	// 创建配置服务器
	server := NewConfigServer(source)

	// 创建配置验证器
	validator := NewConfigValidator()

	// 创建路由器
	router := mux.NewRouter()

	// 配置管理API
	api := router.PathPrefix("/api/config").Subrouter()
	api.HandleFunc("/{app}/{env}", server.GetConfig).Methods("GET")
	api.HandleFunc("/{app}/{env}", server.UpdateConfig).Methods("PUT")
	api.HandleFunc("/{app}/{env}/{key}", server.UpdateConfigItem).Methods("PUT")
	api.HandleFunc("/{app}/{env}/history", server.GetConfigHistory).Methods("GET")

	// WebSocket端点
	router.HandleFunc("/ws", server.HandleWebSocket)

	// 配置验证API
	router.HandleFunc("/api/config/{app}/{env}/validate", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		app := vars["app"]
		env := vars["env"]

		config, err := source.Load(app, env)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		errors := validator.Validate(config)

		response := map[string]interface{}{
			"valid":  len(errors) == 0,
			"errors": errors,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// 演示配置客户端
	go func() {
		time.Sleep(2 * time.Second)

		client := NewConfigClient(source, "user-service", "development")
		if err := client.Load(); err != nil {
			log.Printf("加载配置失败: %v", err)
			return
		}

		// 添加配置监听器
		watcher := &AppConfigWatcher{name: "user-service-dev"}
		client.AddWatcher(watcher)

		// 启动配置监听
		if err := client.StartWatching(); err != nil {
			log.Printf("启动配置监听失败: %v", err)
			return
		}

		// 演示配置读取
		log.Printf("数据库主机: %s", client.GetString("database.host"))
		log.Printf("数据库端口: %d", client.GetInt("database.port"))
		log.Printf("缓存启用: %v", client.GetBool("cache.enabled"))
		log.Printf("服务器端口: %d", client.GetInt("server.port"))

		// 演示配置更新
		go func() {
			time.Sleep(5 * time.Second)

			log.Println("更新配置...")
			config, _ := source.Load("user-service", "development")
			config.Items["server.port"].Value = 8888
			source.Save(config)
		}()
	}()

	fmt.Println("=== 配置管理服务器启动 ===")
	fmt.Println("服务端点:")
	fmt.Println("  配置服务:   http://localhost:8080")
	fmt.Println("  WebSocket:  ws://localhost:8080/ws")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  GET    /api/config/{app}/{env}        - 获取配置")
	fmt.Println("  PUT    /api/config/{app}/{env}        - 更新配置")
	fmt.Println("  PUT    /api/config/{app}/{env}/{key}  - 更新配置项")
	fmt.Println("  GET    /api/config/{app}/{env}/history - 获取配置历史")
	fmt.Println("  GET    /api/config/{app}/{env}/validate - 验证配置")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 获取配置")
	fmt.Println("  curl http://localhost:8080/api/config/user-service/development")
	fmt.Println()
	fmt.Println("  # 更新单个配置项")
	fmt.Println("  curl -X PUT http://localhost:8080/api/config/user-service/development/server.port \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"value\": 9999, \"type\": \"int\", \"description\": \"服务器端口\"}'")
	fmt.Println()
	fmt.Println("  # 验证配置")
	fmt.Println("  curl http://localhost:8080/api/config/user-service/development/validate")

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())
}

/*
练习任务：

1. 基础练习：
   - 实现配置模板和继承
   - 添加配置导入导出功能
   - 实现配置备份和恢复
   - 添加配置比较功能

2. 中级练习：
   - 集成Consul/Etcd作为配置存储
   - 实现配置权限管理
   - 添加配置审计日志
   - 实现配置回滚功能

3. 高级练习：
   - 实现分布式配置一致性
   - 添加配置变更影响分析
   - 实现智能配置推荐
   - 集成K8s ConfigMap/Secret

4. 安全和治理：
   - 实现配置敏感数据脱敏
   - 添加配置访问控制
   - 实现配置合规性检查
   - 添加配置变更审批流程

5. 监控和运维：
   - 实现配置变更监控
   - 添加配置性能分析
   - 实现配置健康检查
   - 添加配置变更告警

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/gorilla/websocket
   go get gopkg.in/yaml.v3

2. 运行程序：go run main.go

配置架构：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  配置客户端  │────│  配置服务器  │────│  配置存储   │
└─────────────┘    └─────────────┘    └─────────────┘
        │                  │                   │
   ┌─────────┐      ┌─────────────┐    ┌─────────────┐
   │配置监听器│      │  配置验证器  │    │  配置加密器  │
   └─────────┘      └─────────────┘    └─────────────┘
        │                  │                   │
   ┌─────────────────────────────────────────────────┐
   │            配置变更通知系统                       │
   └─────────────────────────────────────────────────┘

配置层次结构：
Global Config (全局配置)
├── Application Config (应用配置)
│   ├── Environment Config (环境配置)
│   │   ├── Development
│   │   ├── Testing
│   │   ├── Staging
│   │   └── Production
│   └── Feature Flags (功能开关)
└── Runtime Config (运行时配置)

扩展建议：
- 实现配置中心集群部署
- 集成配置变更流水线
- 添加配置性能优化
- 实现配置多数据中心同步
*/
