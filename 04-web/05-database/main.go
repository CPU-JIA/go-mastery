package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
数据库集成练习

本练习涵盖Go语言中的数据库操作，包括：
1. 原生SQL操作（database/sql）
2. ORM框架使用（GORM）
3. 连接池管理
4. 数据库迁移
5. 事务处理
6. 安全防护（SQL注入防护）
7. 查询优化
8. 批量操作

主要概念：
- 数据库驱动
- 连接池配置
- 预处理语句
- 事务管理
- 对象关系映射（ORM）
*/

// === 数据模型定义 ===

// 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // JSON序列化时隐藏密码
	Profile   Profile   `json:"profile" gorm:"foreignKey:UserID"`
	Posts     []Post    `json:"posts" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 用户资料模型
type Profile struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	UserID   uint   `json:"user_id" gorm:"uniqueIndex"`
	FullName string `json:"full_name"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
	Age      int    `json:"age"`
}

// 文章模型
type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text"`
	Tags      []Tag     `json:"tags" gorm:"many2many:post_tags;"`
	Comments  []Comment `json:"comments" gorm:"foreignKey:PostID"`
	Published bool      `json:"published" gorm:"default:false"`
	Views     int       `json:"views" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 标签模型
type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"uniqueIndex;not null"`
	Color string `json:"color" gorm:"default:'#007bff'"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags;"`
}

// 评论模型
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PostID    uint      `json:"post_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// === 数据库管理器 ===

type DatabaseManager struct {
	db         *gorm.DB
	rawDB      *sql.DB
	migrations []Migration
}

// 数据库迁移定义
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// 创建数据库管理器
func NewDatabaseManager(dsn string, driver string) (*DatabaseManager, error) {
	var db *gorm.DB
	var err error

	// 根据驱动类型连接数据库
	switch driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取原生数据库连接
	rawDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取原生数据库连接失败: %w", err)
	}

	// 配置连接池
	rawDB.SetMaxOpenConns(25)                  // 最大打开连接数
	rawDB.SetMaxIdleConns(5)                   // 最大空闲连接数
	rawDB.SetConnMaxLifetime(time.Hour)        // 连接最大生存时间
	rawDB.SetConnMaxIdleTime(10 * time.Minute) // 连接最大空闲时间

	manager := &DatabaseManager{
		db:    db,
		rawDB: rawDB,
	}

	// 定义迁移
	manager.defineMigrations()

	return manager, nil
}

// 定义数据库迁移
func (dm *DatabaseManager) defineMigrations() {
	dm.migrations = []Migration{
		{
			Version:     "001",
			Description: "创建用户表",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&User{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&User{})
			},
		},
		{
			Version:     "002",
			Description: "创建用户资料表",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&Profile{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&Profile{})
			},
		},
		{
			Version:     "003",
			Description: "创建文章和标签表",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&Post{}, &Tag{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&Post{}, &Tag{})
			},
		},
		{
			Version:     "004",
			Description: "创建评论表",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&Comment{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&Comment{})
			},
		},
	}
}

// 执行迁移
func (dm *DatabaseManager) Migrate() error {
	for _, migration := range dm.migrations {
		log.Printf("执行迁移: %s - %s", migration.Version, migration.Description)
		if err := migration.Up(dm.db); err != nil {
			return fmt.Errorf("迁移 %s 失败: %w", migration.Version, err)
		}
	}
	return nil
}

// 创建示例数据
func (dm *DatabaseManager) SeedData() error {
	// 创建用户
	users := []User{
		{
			Username: "alice",
			Email:    "alice@example.com",
			Password: "hashedpassword1",
			Profile: Profile{
				FullName: "Alice Johnson",
				Bio:      "软件工程师，喜欢Go语言",
				Age:      28,
			},
		},
		{
			Username: "bob",
			Email:    "bob@example.com",
			Password: "hashedpassword2",
			Profile: Profile{
				FullName: "Bob Smith",
				Bio:      "全栈开发者",
				Age:      32,
			},
		},
	}

	for _, user := range users {
		result := dm.db.Where("username = ?", user.Username).FirstOrCreate(&user)
		if result.Error != nil {
			return result.Error
		}
	}

	// 创建标签
	tags := []Tag{
		{Name: "Go语言", Color: "#00ADD8"},
		{Name: "Web开发", Color: "#007bff"},
		{Name: "数据库", Color: "#28a745"},
		{Name: "微服务", Color: "#ffc107"},
	}

	for _, tag := range tags {
		dm.db.Where("name = ?", tag.Name).FirstOrCreate(&tag)
	}

	return nil
}

// === 数据访问层（DAO） ===

// 用户DAO
type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// 创建用户
func (dao *UserDAO) Create(user *User) error {
	return dao.db.Create(user).Error
}

// 根据ID获取用户
func (dao *UserDAO) GetByID(id uint) (*User, error) {
	var user User
	err := dao.db.Preload("Profile").Preload("Posts").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*User, error) {
	var user User
	err := dao.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 更新用户
func (dao *UserDAO) Update(user *User) error {
	return dao.db.Save(user).Error
}

// 删除用户
func (dao *UserDAO) Delete(id uint) error {
	return dao.db.Delete(&User{}, id).Error
}

// 获取所有用户（分页）
func (dao *UserDAO) GetAll(offset, limit int) ([]User, error) {
	var users []User
	err := dao.db.Preload("Profile").Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// 文章DAO
type PostDAO struct {
	db *gorm.DB
}

func NewPostDAO(db *gorm.DB) *PostDAO {
	return &PostDAO{db: db}
}

// 创建文章
func (dao *PostDAO) Create(post *Post) error {
	return dao.db.Create(post).Error
}

// 获取文章（包含关联数据）
func (dao *PostDAO) GetByID(id uint) (*Post, error) {
	var post Post
	err := dao.db.Preload("Tags").Preload("Comments").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// 获取用户的所有文章
func (dao *PostDAO) GetByUserID(userID uint, offset, limit int) ([]Post, error) {
	var posts []Post
	err := dao.db.Where("user_id = ?", userID).
		Preload("Tags").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// 搜索文章
func (dao *PostDAO) Search(keyword string, offset, limit int) ([]Post, error) {
	var posts []Post
	err := dao.db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Preload("Tags").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// 增加文章浏览量
func (dao *PostDAO) IncrementViews(id uint) error {
	return dao.db.Model(&Post{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// === 原生SQL操作示例 ===

// 原生SQL查询器
type RawSQLQuerier struct {
	db *sql.DB
}

func NewRawSQLQuerier(db *sql.DB) *RawSQLQuerier {
	return &RawSQLQuerier{db: db}
}

// 获取用户统计信息（原生SQL）
func (q *RawSQLQuerier) GetUserStats(userID int) (map[string]interface{}, error) {
	query := `
		SELECT
			u.username,
			COUNT(DISTINCT p.id) as post_count,
			COUNT(DISTINCT c.id) as comment_count,
			COALESCE(SUM(p.views), 0) as total_views
		FROM users u
		LEFT JOIN posts p ON u.id = p.user_id
		LEFT JOIN comments c ON u.id = c.user_id
		WHERE u.id = $1
		GROUP BY u.id, u.username
	`

	row := q.db.QueryRow(query, userID)

	var username string
	var postCount, commentCount, totalViews int

	err := row.Scan(&username, &postCount, &commentCount, &totalViews)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"username":      username,
		"post_count":    postCount,
		"comment_count": commentCount,
		"total_views":   totalViews,
	}, nil
}

// 批量插入操作（原生SQL）
func (q *RawSQLQuerier) BatchInsertComments(comments []Comment) error {
	tx, err := q.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 预处理语句
	stmt, err := tx.Prepare("INSERT INTO comments (post_id, user_id, content, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 批量执行
	for _, comment := range comments {
		_, err := stmt.Exec(comment.PostID, comment.UserID, comment.Content, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// === 事务管理示例 ===

// 服务层事务管理
type PostService struct {
	db      *gorm.DB
	userDAO *UserDAO
	postDAO *PostDAO
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{
		db:      db,
		userDAO: NewUserDAO(db),
		postDAO: NewPostDAO(db),
	}
}

// 创建文章（带事务）
func (s *PostService) CreatePostWithTags(userID uint, title, content string, tagNames []string) (*Post, error) {
	var post *Post
	var err error

	// 使用事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 验证用户存在
		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("用户不存在: %w", err)
		}

		// 创建文章
		post = &Post{
			UserID:  userID,
			Title:   title,
			Content: content,
		}

		if err := tx.Create(post).Error; err != nil {
			return fmt.Errorf("创建文章失败: %w", err)
		}

		// 处理标签
		for _, tagName := range tagNames {
			var tag Tag
			// 查找或创建标签
			err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, Tag{Name: tagName}).Error
			if err != nil {
				return fmt.Errorf("处理标签失败: %w", err)
			}

			// 关联标签到文章
			if err := tx.Model(post).Association("Tags").Append(&tag); err != nil {
				return fmt.Errorf("关联标签失败: %w", err)
			}
		}

		return nil
	})

	return post, err
}

// === HTTP API处理器 ===

type APIHandler struct {
	db          *DatabaseManager
	userDAO     *UserDAO
	postDAO     *PostDAO
	postService *PostService
	rawQuerier  *RawSQLQuerier
}

func NewAPIHandler(db *DatabaseManager) *APIHandler {
	return &APIHandler{
		db:          db,
		userDAO:     NewUserDAO(db.db),
		postDAO:     NewPostDAO(db.db),
		postService: NewPostService(db.db),
		rawQuerier:  NewRawSQLQuerier(db.rawDB),
	}
}

// 获取用户列表
func (h *APIHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// 解析分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	users, err := h.userDAO.GetAll(offset, limit)
	if err != nil {
		http.Error(w, "获取用户列表失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"page":  page,
		"limit": limit,
	})
}

// 获取用户详情
func (h *APIHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	user, err := h.userDAO.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "用户不存在", http.StatusNotFound)
		} else {
			http.Error(w, "获取用户失败", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// 创建文章
func (h *APIHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   uint     `json:"user_id"`
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		TagNames []string `json:"tag_names"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	post, err := h.postService.CreatePostWithTags(req.UserID, req.Title, req.Content, req.TagNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// 获取用户统计
func (h *APIHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	stats, err := h.rawQuerier.GetUserStats(id)
	if err != nil {
		http.Error(w, "获取统计信息失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// === 数据库性能优化示例 ===

// 查询优化器
type QueryOptimizer struct {
	db *gorm.DB
}

func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// 使用索引优化查询
func (qo *QueryOptimizer) GetPopularPosts(limit int) ([]Post, error) {
	var posts []Post
	err := qo.db.
		Select("id, title, views, created_at"). // 只选择需要的字段
		Where("published = ?", true).           // 使用索引字段
		Order("views DESC").                    // 按浏览量排序
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// 预加载关联数据
func (qo *QueryOptimizer) GetPostsWithRelations(userID uint) ([]Post, error) {
	var posts []Post
	err := qo.db.
		Preload("Tags").     // 预加载标签
		Preload("Comments"). // 预加载评论
		Where("user_id = ?", userID).
		Find(&posts).Error
	return posts, err
}

// 批量操作
func (qo *QueryOptimizer) BatchUpdateViews(postIDs []uint) error {
	return qo.db.Model(&Post{}).
		Where("id IN ?", postIDs).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// === 数据库安全示例 ===

func demonstrateDatabaseSecurity() {
	fmt.Println("=== 数据库安全最佳实践 ===")

	fmt.Println("1. 防止SQL注入:")
	fmt.Println("   ✓ 使用参数化查询/预处理语句")
	fmt.Println("   ✓ 输入验证和转义")
	fmt.Println("   ✓ 使用ORM框架")
	fmt.Println("   ✗ 避免字符串拼接SQL")

	fmt.Println("2. 连接安全:")
	fmt.Println("   ✓ 使用SSL/TLS连接")
	fmt.Println("   ✓ 限制数据库访问权限")
	fmt.Println("   ✓ 使用专用数据库用户")
	fmt.Println("   ✓ 定期更新数据库密码")

	fmt.Println("3. 数据保护:")
	fmt.Println("   ✓ 敏感数据加密存储")
	fmt.Println("   ✓ 定期备份数据")
	fmt.Println("   ✓ 审计日志记录")
	fmt.Println("   ✓ 数据脱敏处理")
}

func main() {
	// 初始化数据库（使用SQLite进行演示）
	dbManager, err := NewDatabaseManager("demo.db", "sqlite")
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 执行数据库迁移
	if err := dbManager.Migrate(); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建示例数据
	if err := dbManager.SeedData(); err != nil {
		log.Fatal("创建示例数据失败:", err)
	}

	// 创建API处理器
	apiHandler := NewAPIHandler(dbManager)

	// 设置路由
	router := mux.NewRouter()

	// 用户相关路由
	router.HandleFunc("/users", apiHandler.GetUsers).Methods("GET")
	router.HandleFunc("/users/{id}", apiHandler.GetUser).Methods("GET")
	router.HandleFunc("/users/{id}/stats", apiHandler.GetUserStats).Methods("GET")

	// 文章相关路由
	router.HandleFunc("/posts", apiHandler.CreatePost).Methods("POST")

	// 演示数据库安全最佳实践
	demonstrateDatabaseSecurity()

	fmt.Println("=== 数据库集成服务器启动 ===")
	fmt.Println("API端点:")
	fmt.Println("  GET  /users           - 获取用户列表")
	fmt.Println("  GET  /users/{id}      - 获取用户详情")
	fmt.Println("  GET  /users/{id}/stats - 获取用户统计")
	fmt.Println("  POST /posts           - 创建文章")
	fmt.Println()
	fmt.Println("服务器运行在 http://localhost:8080")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

/*
练习任务：

1. 基础练习：
   - 实现完整的CRUD操作（用户、文章、评论）
   - 添加数据验证和错误处理
   - 实现软删除功能
   - 添加时间戳自动更新

2. 中级练习：
   - 实现复杂查询（联表查询、聚合函数）
   - 添加数据库索引优化
   - 实现全文搜索功能
   - 添加数据缓存层（Redis）

3. 高级练习：
   - 实现读写分离
   - 添加数据库分片
   - 实现数据同步机制
   - 集成消息队列处理异步操作

4. 安全练习：
   - 实现数据加密存储
   - 添加SQL注入防护测试
   - 实现访问控制和权限管理
   - 添加审计日志功能

5. 性能优化：
   - 实现连接池监控
   - 添加慢查询日志
   - 优化复杂查询性能
   - 实现数据库监控指标

数据库配置示例：

PostgreSQL:
dsn := "host=localhost user=username password=password dbname=mydb port=5432 sslmode=disable"

MySQL:
dsn := "username:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"

SQLite:
dsn := "demo.db"

运行前准备：
1. 安装依赖：
   go get gorm.io/gorm
   go get gorm.io/driver/sqlite
   go get gorm.io/driver/postgres
   go get github.com/lib/pq
   go get github.com/gorilla/mux

2. 运行程序：go run main.go

3. 测试API：
   curl http://localhost:8080/users
   curl http://localhost:8080/users/1
   curl -X POST http://localhost:8080/posts -d '{"user_id":1,"title":"测试文章","content":"内容","tag_names":["Go","Web"]}'

扩展建议：
- 集成数据库迁移工具（如golang-migrate）
- 使用数据库连接池监控工具
- 实现多数据库支持切换
- 集成ETL数据处理流程
*/
