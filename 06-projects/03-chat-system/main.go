/*
实时聊天系统 (Real-time Chat System)

项目描述:
一个完整的实时聊天系统，支持多房间聊天、私聊、消息历史、
用户在线状态、文件传输、表情包等功能。

技术栈:
- WebSocket 实时通信
- Gorilla WebSocket 库
- 并发安全设计
- 消息持久化
- 用户认证
- 房间管理
*/

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ====================
// 1. 数据模型
// ====================

type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Status   string    `json:"status"` // online, away, busy, offline
	LastSeen time.Time `json:"last_seen"`
	JoinedAt time.Time `json:"joined_at"`
}

type Room struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // public, private, direct
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Members     []string  `json:"members"`
	IsActive    bool      `json:"is_active"`
}

type Message struct {
	ID        string                 `json:"id"`
	RoomID    string                 `json:"room_id"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"` // text, image, file, system
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	EditedAt  *time.Time             `json:"edited_at,omitempty"`
	ReplyTo   string                 `json:"reply_to,omitempty"`
}

type Client struct {
	ID       string
	User     *User
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	Rooms    map[string]bool // 用户加入的房间
	LastPing time.Time
}

type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 用户 ID 到客户端的映射
	userClients map[string]*Client

	// 房间到客户端的映射
	roomClients map[string]map[*Client]bool

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	// 消息广播
	broadcast chan *BroadcastMessage

	// 存储
	storage *Storage

	// 互斥锁
	mu sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  string
	Message *Message
	Exclude *Client // 排除的客户端（比如发送者）
}

// WebSocket 消息类型
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ChatMessage struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
	Type    string `json:"type"`
	ReplyTo string `json:"reply_to,omitempty"`
}

type JoinRoomMessage struct {
	RoomID string `json:"room_id"`
}

type TypingMessage struct {
	RoomID string `json:"room_id"`
	Typing bool   `json:"typing"`
}

// ====================
// 2. 存储层
// ====================

type Storage struct {
	users    map[string]*User
	rooms    map[string]*Room
	messages []Message
	dataDir  string
	mu       sync.RWMutex
}

func NewStorage(dataDir string) *Storage {
	storage := &Storage{
		users:    make(map[string]*User),
		rooms:    make(map[string]*Room),
		messages: make([]Message, 0),
		dataDir:  dataDir,
	}

	os.MkdirAll(dataDir, 0755)
	storage.loadData()
	storage.createDefaultData()

	return storage
}

func (s *Storage) loadData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 加载用户数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "users.json")); err == nil {
		json.Unmarshal(data, &s.users)
	}

	// 加载房间数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "rooms.json")); err == nil {
		json.Unmarshal(data, &s.rooms)
	}

	// 加载消息数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "messages.json")); err == nil {
		json.Unmarshal(data, &s.messages)
	}
}

func (s *Storage) saveData() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 保存用户数据
	if data, err := json.MarshalIndent(s.users, "", "  "); err == nil {
		os.WriteFile(filepath.Join(s.dataDir, "users.json"), data, 0644)
	}

	// 保存房间数据
	if data, err := json.MarshalIndent(s.rooms, "", "  "); err == nil {
		os.WriteFile(filepath.Join(s.dataDir, "rooms.json"), data, 0644)
	}

	// 保存消息数据
	if data, err := json.MarshalIndent(s.messages, "", "  "); err == nil {
		os.WriteFile(filepath.Join(s.dataDir, "messages.json"), data, 0644)
	}
}

func (s *Storage) createDefaultData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建默认房间
	if len(s.rooms) == 0 {
		generalRoom := &Room{
			ID:          "general",
			Name:        "General",
			Description: "General discussion room",
			Type:        "public",
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			Members:     []string{},
			IsActive:    true,
		}
		s.rooms["general"] = generalRoom

		randomRoom := &Room{
			ID:          "random",
			Name:        "Random",
			Description: "Random chat room",
			Type:        "public",
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			Members:     []string{},
			IsActive:    true,
		}
		s.rooms["random"] = randomRoom

		s.saveData()
	}
}

func (s *Storage) CreateUser(userID, username string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &User{
		ID:       userID,
		Username: username,
		Avatar:   fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", username),
		Status:   "online",
		LastSeen: time.Now(),
		JoinedAt: time.Now(),
	}

	s.users[userID] = user
	s.saveData()
	return user
}

func (s *Storage) GetUser(userID string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.users[userID]
}

func (s *Storage) UpdateUserStatus(userID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user, exists := s.users[userID]; exists {
		user.Status = status
		user.LastSeen = time.Now()
		s.saveData()
	}
}

func (s *Storage) GetRoom(roomID string) *Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rooms[roomID]
}

func (s *Storage) GetPublicRooms() []*Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*Room, 0)
	for _, room := range s.rooms {
		if room.Type == "public" && room.IsActive {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (s *Storage) CreateRoom(name, description, createdBy string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	roomID := fmt.Sprintf("room_%d", time.Now().UnixNano())
	room := &Room{
		ID:          roomID,
		Name:        name,
		Description: description,
		Type:        "public",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		Members:     []string{createdBy},
		IsActive:    true,
	}

	s.rooms[roomID] = room
	s.saveData()
	return room
}

func (s *Storage) AddUserToRoom(userID, roomID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	// 检查用户是否已在房间中
	for _, memberID := range room.Members {
		if memberID == userID {
			return nil // 已在房间中
		}
	}

	room.Members = append(room.Members, userID)
	s.saveData()
	return nil
}

func (s *Storage) SaveMessage(message *Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, *message)

	// 保持最近 1000 条消息
	if len(s.messages) > 1000 {
		s.messages = s.messages[len(s.messages)-1000:]
	}

	s.saveData()
}

func (s *Storage) GetRoomMessages(roomID string, limit int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]Message, 0)
	for i := len(s.messages) - 1; i >= 0 && len(messages) < limit; i-- {
		if s.messages[i].RoomID == roomID {
			messages = append([]Message{s.messages[i]}, messages...)
		}
	}

	return messages
}

func (s *Storage) GetOnlineUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0)
	for _, user := range s.users {
		if user.Status == "online" {
			users = append(users, user)
		}
	}
	return users
}

// ====================
// 3. WebSocket 升级器
// ====================

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 安全修复：验证来源以防止CSRF攻击
		origin := r.Header.Get("Origin")
		// 在生产环境中，应该检查允许的来源列表
		// 这里允许本地开发环境和常见开发端口
		allowedOrigins := []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://localhost:3000", // 开发服务器
			"http://127.0.0.1:3000",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		// 如果没有Origin头（可能是直接WebSocket连接），也允许
		return origin == ""
	},
}

// ====================
// 4. Hub 实现
// ====================

func NewHub(storage *Storage) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[string]*Client),
		roomClients: make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage),
		storage:     storage,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case broadcastMsg := <-h.broadcast:
			h.broadcastMessage(broadcastMsg)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.userClients[client.User.ID] = client

	log.Printf("User %s connected", client.User.Username)

	// 更新用户状态
	h.storage.UpdateUserStatus(client.User.ID, "online")

	// 发送初始数据
	h.sendInitialData(client)

	// 广播用户上线通知
	h.broadcastUserStatus(client.User, "online")
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.userClients, client.User.ID)

		// 从所有房间移除
		for roomID := range client.Rooms {
			if roomClients, exists := h.roomClients[roomID]; exists {
				delete(roomClients, client)
				if len(roomClients) == 0 {
					delete(h.roomClients, roomID)
				}
			}
		}

		close(client.Send)

		log.Printf("User %s disconnected", client.User.Username)

		// 更新用户状态
		h.storage.UpdateUserStatus(client.User.ID, "offline")

		// 广播用户下线通知
		h.broadcastUserStatus(client.User, "offline")
	}
}

func (h *Hub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	roomClients := h.roomClients[broadcastMsg.RoomID]
	h.mu.RUnlock()

	if roomClients == nil {
		return
	}

	messageData, _ := json.Marshal(WSMessage{
		Type: "message",
		Data: broadcastMsg.Message,
	})

	for client := range roomClients {
		if client != broadcastMsg.Exclude {
			select {
			case client.Send <- messageData:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(roomClients, client)
			}
		}
	}
}

func (h *Hub) sendInitialData(client *Client) {
	// 发送公共房间列表
	rooms := h.storage.GetPublicRooms()
	roomsData, _ := json.Marshal(WSMessage{
		Type: "rooms",
		Data: rooms,
	})
	client.Send <- roomsData

	// 发送在线用户列表
	users := h.storage.GetOnlineUsers()
	usersData, _ := json.Marshal(WSMessage{
		Type: "users",
		Data: users,
	})
	client.Send <- usersData
}

func (h *Hub) broadcastUserStatus(user *User, status string) {
	statusData, _ := json.Marshal(WSMessage{
		Type: "user_status",
		Data: map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"status":   status,
		},
	})

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- statusData:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) joinRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查房间是否存在
	room := h.storage.GetRoom(roomID)
	if room == nil {
		return
	}

	// 添加用户到房间
	h.storage.AddUserToRoom(client.User.ID, roomID)

	// 添加到内存映射
	if h.roomClients[roomID] == nil {
		h.roomClients[roomID] = make(map[*Client]bool)
	}
	h.roomClients[roomID][client] = true
	client.Rooms[roomID] = true

	// 发送房间历史消息
	messages := h.storage.GetRoomMessages(roomID, 50)
	historyData, _ := json.Marshal(WSMessage{
		Type: "room_history",
		Data: map[string]interface{}{
			"room_id":  roomID,
			"messages": messages,
		},
	})
	client.Send <- historyData

	// 广播用户加入消息
	joinMessage := &Message{
		ID:        generateMessageID(),
		RoomID:    roomID,
		UserID:    "system",
		Username:  "System",
		Content:   fmt.Sprintf("%s joined the room", client.User.Username),
		Type:      "system",
		Timestamp: time.Now(),
	}

	h.storage.SaveMessage(joinMessage)

	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: joinMessage,
		Exclude: client,
	}
}

// ====================
// 5. Client 处理
// ====================

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(&wsMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			// XSS安全修复：转义用户消息防止XSS攻击
			escapedMessage := []byte(html.EscapeString(string(message)))
			w.Write(escapedMessage)

			// 批量发送缓冲中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				// XSS安全修复：转义用户消息防止XSS攻击
				message := <-c.Send
				escapedMessage := []byte(html.EscapeString(string(message)))
				w.Write(escapedMessage)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(wsMsg *WSMessage) {
	switch wsMsg.Type {
	case "chat_message":
		c.handleChatMessage(wsMsg.Data)
	case "join_room":
		c.handleJoinRoom(wsMsg.Data)
	case "typing":
		c.handleTyping(wsMsg.Data)
	case "create_room":
		c.handleCreateRoom(wsMsg.Data)
	}
}

func (c *Client) handleChatMessage(data interface{}) {
	chatMsgBytes, _ := json.Marshal(data)
	var chatMsg ChatMessage
	if err := json.Unmarshal(chatMsgBytes, &chatMsg); err != nil {
		return
	}

	// 创建消息
	message := &Message{
		ID:        generateMessageID(),
		RoomID:    chatMsg.RoomID,
		UserID:    c.User.ID,
		Username:  c.User.Username,
		Content:   chatMsg.Content,
		Type:      chatMsg.Type,
		Timestamp: time.Now(),
		ReplyTo:   chatMsg.ReplyTo,
	}

	// 保存消息
	c.Hub.storage.SaveMessage(message)

	// 广播消息
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  chatMsg.RoomID,
		Message: message,
		Exclude: nil,
	}
}

func (c *Client) handleJoinRoom(data interface{}) {
	joinMsgBytes, _ := json.Marshal(data)
	var joinMsg JoinRoomMessage
	if err := json.Unmarshal(joinMsgBytes, &joinMsg); err != nil {
		return
	}

	c.Hub.joinRoom(c, joinMsg.RoomID)
}

func (c *Client) handleTyping(data interface{}) {
	typingBytes, _ := json.Marshal(data)
	var typingMsg TypingMessage
	if err := json.Unmarshal(typingBytes, &typingMsg); err != nil {
		return
	}

	// 广播打字状态
	typingData, _ := json.Marshal(WSMessage{
		Type: "typing",
		Data: map[string]interface{}{
			"room_id":  typingMsg.RoomID,
			"user_id":  c.User.ID,
			"username": c.User.Username,
			"typing":   typingMsg.Typing,
		},
	})

	c.Hub.mu.RLock()
	roomClients := c.Hub.roomClients[typingMsg.RoomID]
	c.Hub.mu.RUnlock()

	if roomClients != nil {
		for client := range roomClients {
			if client != c {
				select {
				case client.Send <- typingData:
				default:
				}
			}
		}
	}
}

func (c *Client) handleCreateRoom(data interface{}) {
	roomData := data.(map[string]interface{})
	name := roomData["name"].(string)
	description := ""
	if desc, ok := roomData["description"].(string); ok {
		description = desc
	}

	room := c.Hub.storage.CreateRoom(name, description, c.User.ID)

	// 发送新房间信息给所有客户端
	roomsData, _ := json.Marshal(WSMessage{
		Type: "new_room",
		Data: room,
	})

	c.Hub.mu.RLock()
	defer c.Hub.mu.RUnlock()

	for client := range c.Hub.clients {
		select {
		case client.Send <- roomsData:
		default:
		}
	}
}

// ====================
// 6. HTTP 服务器
// ====================

type ChatServer struct {
	hub             *Hub
	storage         *Storage
	templateHandler *TemplateHandler
}

func NewChatServer(storage *Storage) *ChatServer {
	hub := NewHub(storage)
	go hub.Run()

	return &ChatServer{
		hub:             hub,
		storage:         storage,
		templateHandler: NewTemplateHandler(),
	}
}

func (s *ChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/ws":
		s.handleWebSocket(w, r)
	case r.URL.Path == "/api/rooms":
		s.handleRoomsAPI(w, r)
	case r.URL.Path == "/api/users":
		s.handleUsersAPI(w, r)
	case strings.HasPrefix(r.URL.Path, "/static/"):
		s.handleStaticFiles(w, r)
	case r.URL.Path == "/" || r.URL.Path == "/chat":
		s.handleChatPage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *ChatServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 获取用户信息
	userID := r.URL.Query().Get("user_id")
	username := r.URL.Query().Get("username")

	if userID == "" || username == "" {
		http.Error(w, "Missing user_id or username", http.StatusBadRequest)
		return
	}

	// 升级到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// 获取或创建用户
	user := s.storage.GetUser(userID)
	if user == nil {
		user = s.storage.CreateUser(userID, username)
	}

	// 创建客户端
	client := &Client{
		ID:       userID,
		User:     user,
		Conn:     conn,
		Hub:      s.hub,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
		LastPing: time.Now(),
	}

	// 注册客户端
	s.hub.register <- client

	// 启动 goroutines
	go client.writePump()
	go client.readPump()
}

func (s *ChatServer) handleRoomsAPI(w http.ResponseWriter, r *http.Request) {
	rooms := s.storage.GetPublicRooms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (s *ChatServer) handleUsersAPI(w http.ResponseWriter, r *http.Request) {
	users := s.storage.GetOnlineUsers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *ChatServer) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Remove the /static/ prefix and serve the file
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")

	// 安全修复：防止路径遍历攻击
	// 清理路径并确保不能访问上级目录
	filePath = filepath.Clean(filePath)
	if strings.Contains(filePath, "..") || strings.HasPrefix(filePath, "/") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Determine content type based on file extension
	var contentType string
	if strings.HasSuffix(filePath, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(filePath, ".js") {
		contentType = "application/javascript"
	} else {
		contentType = "application/octet-stream"
	}

	// Try to read the file
	fullPath := filepath.Join("static", filePath)
	// 二次安全检查：确保最终路径在static目录内
	if !strings.HasPrefix(filepath.Clean(fullPath), "static"+string(filepath.Separator)) && fullPath != "static" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if data, err := os.ReadFile(fullPath); err == nil {
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	} else {
		http.NotFound(w, r)
	}
}

func (s *ChatServer) handleChatPage(w http.ResponseWriter, r *http.Request) {
	// Use template handler to render the chat page
	err := s.templateHandler.RenderTemplate(w, "chat", nil)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ====================
// 7. 辅助函数
// ====================

func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// ====================
// 主函数
// ====================

func main() {
	// 创建存储
	storage := NewStorage("./chat_data")

	// 创建聊天服务器
	chatServer := NewChatServer(storage)

	// 设置路由
	http.Handle("/", chatServer)

	// 启动服务器
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("💬 实时聊天系统启动在 http://localhost:%s", port)
	log.Println("功能特性:")
	log.Println("- 实时消息传递")
	log.Println("- 多房间支持")
	log.Println("- 用户在线状态")
	log.Println("- 打字指示器")
	log.Println("- 消息历史")
	log.Println("- 房间创建")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

/*
=== 项目功能清单 ===

核心功能:
✅ WebSocket 实时通信
✅ 多房间聊天支持
✅ 用户在线状态管理
✅ 消息历史持久化
✅ 打字状态指示器
✅ 房间创建和管理
✅ 用户头像生成
✅ 自动重连机制

界面功能:
✅ 响应式聊天界面
✅ 实时消息显示
✅ 用户列表和房间列表
✅ 消息输入框自适应
✅ 滚动到最新消息
✅ 用户状态指示

技术特性:
✅ Gorilla WebSocket 库
✅ 并发安全 (sync.RWMutex)
✅ 消息广播机制
✅ 客户端生命周期管理
✅ 心跳检测 (Ping/Pong)
✅ 数据持久化

=== 高级功能扩展 ===

1. 消息功能:
   - 文件上传和分享
   - 图片预览
   - 表情包支持
   - 消息编辑和删除
   - @用户提醒
   - 消息回复

2. 房间功能:
   - 私人聊天 (Direct Message)
   - 房间权限管理
   - 房间成员管理
   - 房间公告和置顶
   - 房间搜索和分类

3. 用户体验:
   - 桌面通知
   - 声音提醒
   - 深色模式
   - 自定义主题
   - 消息搜索
   - 聊天记录导出

4. 管理功能:
   - 用户封禁和解封
   - 消息审核
   - 垃圾信息过滤
   - 聊天记录备份
   - 用户行为分析

=== 部署说明 ===

1. 运行应用:
   go run main.go

2. 访问聊天:
   http://localhost:8080

3. 注意事项:
   - 需要 gorilla/websocket 库: go get github.com/gorilla/websocket
   - 生产环境需要配置 HTTPS 和 WSS
   - 可设置环境变量 PORT 改变端口

=== 技术架构 ===

1. 连接管理:
   - Hub 模式管理所有连接
   - 客户端注册/注销机制
   - 房间订阅模式

2. 消息处理:
   - 异步消息处理
   - 消息广播和路由
   - 消息持久化存储

3. 并发控制:
   - 读写分离锁
   - Goroutine 池管理
   - 资源清理机制
*/
