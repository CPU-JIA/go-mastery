package main

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"
)

// ====================
// 1. 数据模型测试
// ====================

func TestUser(t *testing.T) {
	now := time.Now()
	user := User{
		ID:       "user123",
		Username: "testuser",
		Avatar:   "https://example.com/avatar.jpg",
		Status:   "online",
		LastSeen: now,
		JoinedAt: now,
	}

	// 测试用户结构体字段
	if user.ID != "user123" {
		t.Errorf("User.ID = %q, want %q", user.ID, "user123")
	}
	if user.Username != "testuser" {
		t.Errorf("User.Username = %q, want %q", user.Username, "testuser")
	}
	if user.Status != "online" {
		t.Errorf("User.Status = %q, want %q", user.Status, "online")
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal User to JSON: %v", err)
	}

	var unmarshaled User
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal User from JSON: %v", err)
	}

	if !reflect.DeepEqual(user, unmarshaled) {
		t.Errorf("JSON marshal/unmarshal failed: got %+v, want %+v", unmarshaled, user)
	}
}

func TestRoom(t *testing.T) {
	now := time.Now()
	room := Room{
		ID:          "room123",
		Name:        "General Chat",
		Description: "General discussion room",
		Type:        "public",
		CreatedBy:   "user123",
		CreatedAt:   now,
		Members:     []string{"user123", "user456"},
		IsActive:    true,
	}

	// 测试房间结构体字段
	if room.ID != "room123" {
		t.Errorf("Room.ID = %q, want %q", room.ID, "room123")
	}
	if room.Type != "public" {
		t.Errorf("Room.Type = %q, want %q", room.Type, "public")
	}
	if len(room.Members) != 2 {
		t.Errorf("Room.Members length = %d, want %d", len(room.Members), 2)
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("Failed to marshal Room to JSON: %v", err)
	}

	var unmarshaled Room
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Room from JSON: %v", err)
	}

	if !reflect.DeepEqual(room, unmarshaled) {
		t.Errorf("JSON marshal/unmarshal failed: got %+v, want %+v", unmarshaled, room)
	}
}

func TestMessage(t *testing.T) {
	now := time.Now()
	editedTime := now.Add(5 * time.Minute)

	message := Message{
		ID:        "msg123",
		RoomID:    "room123",
		UserID:    "user123",
		Username:  "testuser",
		Content:   "Hello, world!",
		Type:      "text",
		Metadata:  map[string]interface{}{"version": "1.0"},
		Timestamp: now,
		EditedAt:  &editedTime,
		ReplyTo:   "msg122",
	}

	// 测试消息结构体字段
	if message.ID != "msg123" {
		t.Errorf("Message.ID = %q, want %q", message.ID, "msg123")
	}
	if message.Content != "Hello, world!" {
		t.Errorf("Message.Content = %q, want %q", message.Content, "Hello, world!")
	}
	if message.Type != "text" {
		t.Errorf("Message.Type = %q, want %q", message.Type, "text")
	}
	if message.EditedAt == nil {
		t.Error("Message.EditedAt should not be nil")
	}
	if message.ReplyTo != "msg122" {
		t.Errorf("Message.ReplyTo = %q, want %q", message.ReplyTo, "msg122")
	}

	// 测试元数据
	if version, ok := message.Metadata["version"].(string); !ok || version != "1.0" {
		t.Errorf("Message.Metadata[version] = %v, want %q", message.Metadata["version"], "1.0")
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal Message to JSON: %v", err)
	}

	var unmarshaled Message
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message from JSON: %v", err)
	}

	// 注意：时间比较需要特殊处理
	if unmarshaled.ID != message.ID {
		t.Errorf("Unmarshaled Message.ID = %q, want %q", unmarshaled.ID, message.ID)
	}
	if unmarshaled.Content != message.Content {
		t.Errorf("Unmarshaled Message.Content = %q, want %q", unmarshaled.Content, message.Content)
	}
}

// ====================
// 2. Hub测试 (需要模拟)
// ====================

// 创建模拟的Hub结构（简化版本）
type MockHub struct {
	clients     map[*MockClient]bool
	userClients map[string]*MockClient
	roomClients map[string]map[*MockClient]bool
	register    chan *MockClient
	unregister  chan *MockClient
	broadcast   chan *MockBroadcastMessage
	mutex       sync.RWMutex
}

type MockClient struct {
	ID    string
	User  *User
	Rooms map[string]bool
	Send  chan []byte
	hub   *MockHub
}

type MockBroadcastMessage struct {
	Type    string
	RoomID  string
	Message interface{}
}

func NewMockHub() *MockHub {
	return &MockHub{
		clients:     make(map[*MockClient]bool),
		userClients: make(map[string]*MockClient),
		roomClients: make(map[string]map[*MockClient]bool),
		register:    make(chan *MockClient, 10),
		unregister:  make(chan *MockClient, 10),
		broadcast:   make(chan *MockBroadcastMessage, 100),
	}
}

func (h *MockHub) RegisterClient(client *MockClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	h.userClients[client.User.ID] = client

	// 将客户端添加到其房间
	for roomID := range client.Rooms {
		if h.roomClients[roomID] == nil {
			h.roomClients[roomID] = make(map[*MockClient]bool)
		}
		h.roomClients[roomID][client] = true
	}
}

func (h *MockHub) UnregisterClient(client *MockClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.userClients, client.User.ID)

		// 从房间中移除客户端
		for roomID := range client.Rooms {
			if clients, ok := h.roomClients[roomID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.roomClients, roomID)
				}
			}
		}

		close(client.Send)
	}
}

func (h *MockHub) BroadcastToRoom(roomID string, message interface{}) {
	h.mutex.RLock()
	clients := h.roomClients[roomID]
	h.mutex.RUnlock()

	if clients == nil {
		return
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return
	}

	for client := range clients {
		select {
		case client.Send <- messageData:
		default:
			// 客户端缓冲区满，跳过
		}
	}
}

func (h *MockHub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

func (h *MockHub) GetRoomClientCount(roomID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if clients, ok := h.roomClients[roomID]; ok {
		return len(clients)
	}
	return 0
}

func TestMockHub(t *testing.T) {
	hub := NewMockHub()

	// 测试空Hub
	if count := hub.GetClientCount(); count != 0 {
		t.Errorf("Empty hub client count = %d, want 0", count)
	}

	// 创建测试用户和客户端
	user1 := &User{
		ID:       "user1",
		Username: "testuser1",
		Status:   "online",
	}

	user2 := &User{
		ID:       "user2",
		Username: "testuser2",
		Status:   "online",
	}

	client1 := &MockClient{
		ID:   "client1",
		User: user1,
		Rooms: map[string]bool{
			"room1": true,
			"room2": true,
		},
		Send: make(chan []byte, 10),
		hub:  hub,
	}

	client2 := &MockClient{
		ID:   "client2",
		User: user2,
		Rooms: map[string]bool{
			"room1": true,
		},
		Send: make(chan []byte, 10),
		hub:  hub,
	}

	// 测试客户端注册
	hub.RegisterClient(client1)
	if count := hub.GetClientCount(); count != 1 {
		t.Errorf("After registering client1, count = %d, want 1", count)
	}

	hub.RegisterClient(client2)
	if count := hub.GetClientCount(); count != 2 {
		t.Errorf("After registering client2, count = %d, want 2", count)
	}

	// 测试房间客户端计数
	if count := hub.GetRoomClientCount("room1"); count != 2 {
		t.Errorf("Room1 client count = %d, want 2", count)
	}

	if count := hub.GetRoomClientCount("room2"); count != 1 {
		t.Errorf("Room2 client count = %d, want 1", count)
	}

	if count := hub.GetRoomClientCount("nonexistent"); count != 0 {
		t.Errorf("Nonexistent room client count = %d, want 0", count)
	}

	// 测试广播到房间
	testMessage := map[string]interface{}{
		"type":    "message",
		"content": "Hello, room1!",
		"user":    "testuser1",
	}

	hub.BroadcastToRoom("room1", testMessage)

	// 检查两个客户端都收到消息
	select {
	case msg := <-client1.Send:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		if received["content"] != "Hello, room1!" {
			t.Errorf("Client1 received wrong message content: %v", received["content"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive message")
	}

	select {
	case msg := <-client2.Send:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		if received["content"] != "Hello, room1!" {
			t.Errorf("Client2 received wrong message content: %v", received["content"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 did not receive message")
	}

	// 测试广播到room2 (只有client1应该收到)
	hub.BroadcastToRoom("room2", map[string]interface{}{
		"type":    "message",
		"content": "Hello, room2!",
	})

	select {
	case msg := <-client1.Send:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		if received["content"] != "Hello, room2!" {
			t.Errorf("Client1 received wrong message content: %v", received["content"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive room2 message")
	}

	// Client2不应该收到room2的消息
	select {
	case <-client2.Send:
		t.Error("Client2 should not receive room2 message")
	case <-time.After(50 * time.Millisecond):
		// 正确，没有收到消息
	}

	// 测试客户端注销
	hub.UnregisterClient(client1)
	if count := hub.GetClientCount(); count != 1 {
		t.Errorf("After unregistering client1, count = %d, want 1", count)
	}

	if count := hub.GetRoomClientCount("room1"); count != 1 {
		t.Errorf("After unregistering client1, room1 count = %d, want 1", count)
	}

	if count := hub.GetRoomClientCount("room2"); count != 0 {
		t.Errorf("After unregistering client1, room2 count = %d, want 0", count)
	}

	hub.UnregisterClient(client2)
	if count := hub.GetClientCount(); count != 0 {
		t.Errorf("After unregistering all clients, count = %d, want 0", count)
	}
}

// ====================
// 3. 消息验证测试
// ====================

func validateMessage(msg *Message) []string {
	var errors []string

	if msg.ID == "" {
		errors = append(errors, "message ID is required")
	}

	if msg.RoomID == "" {
		errors = append(errors, "room ID is required")
	}

	if msg.UserID == "" {
		errors = append(errors, "user ID is required")
	}

	if msg.Username == "" {
		errors = append(errors, "username is required")
	}

	if msg.Content == "" && msg.Type == "text" {
		errors = append(errors, "text message content cannot be empty")
	}

	if msg.Type == "" {
		errors = append(errors, "message type is required")
	}

	validTypes := map[string]bool{
		"text":   true,
		"image":  true,
		"file":   true,
		"system": true,
	}

	if !validTypes[msg.Type] {
		errors = append(errors, "invalid message type: "+msg.Type)
	}

	if msg.Timestamp.IsZero() {
		errors = append(errors, "timestamp is required")
	}

	return errors
}

func TestMessageValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		message        *Message
		expectedErrors []string
	}{
		{
			name: "valid text message",
			message: &Message{
				ID:        "msg123",
				RoomID:    "room123",
				UserID:    "user123",
				Username:  "testuser",
				Content:   "Hello, world!",
				Type:      "text",
				Timestamp: now,
			},
			expectedErrors: nil,
		},
		{
			name: "missing ID",
			message: &Message{
				RoomID:    "room123",
				UserID:    "user123",
				Username:  "testuser",
				Content:   "Hello, world!",
				Type:      "text",
				Timestamp: now,
			},
			expectedErrors: []string{"message ID is required"},
		},
		{
			name: "empty text message content",
			message: &Message{
				ID:        "msg123",
				RoomID:    "room123",
				UserID:    "user123",
				Username:  "testuser",
				Content:   "",
				Type:      "text",
				Timestamp: now,
			},
			expectedErrors: []string{"text message content cannot be empty"},
		},
		{
			name: "invalid message type",
			message: &Message{
				ID:        "msg123",
				RoomID:    "room123",
				UserID:    "user123",
				Username:  "testuser",
				Content:   "Hello",
				Type:      "invalid",
				Timestamp: now,
			},
			expectedErrors: []string{"invalid message type: invalid"},
		},
		{
			name: "multiple validation errors",
			message: &Message{
				ID:       "msg123",
				Content:  "Hello",
				Type:     "invalid",
				Username: "testuser",
				// Missing RoomID, UserID, Timestamp
			},
			expectedErrors: []string{
				"room ID is required",
				"user ID is required",
				"invalid message type: invalid",
				"timestamp is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateMessage(tt.message)

			if len(tt.expectedErrors) == 0 {
				if len(errors) > 0 {
					t.Errorf("validateMessage() expected no errors, got %v", errors)
				}
			} else {
				if len(errors) != len(tt.expectedErrors) {
					t.Errorf("validateMessage() error count = %d, want %d", len(errors), len(tt.expectedErrors))
					t.Errorf("Got errors: %v", errors)
					t.Errorf("Expected errors: %v", tt.expectedErrors)
				} else {
					for i, expectedErr := range tt.expectedErrors {
						if errors[i] != expectedErr {
							t.Errorf("validateMessage() error[%d] = %q, want %q", i, errors[i], expectedErr)
						}
					}
				}
			}
		})
	}
}

// ====================
// 4. 房间管理测试
// ====================

type MockRoomManager struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

func NewMockRoomManager() *MockRoomManager {
	return &MockRoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *MockRoomManager) CreateRoom(room *Room) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if room.ID == "" {
		return &ValidationError{Field: "id", Message: "room ID is required"}
	}

	if _, exists := rm.rooms[room.ID]; exists {
		return &ValidationError{Field: "id", Message: "room already exists"}
	}

	rm.rooms[room.ID] = room
	return nil
}

func (rm *MockRoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if room, exists := rm.rooms[roomID]; exists {
		return room, nil
	}

	return nil, &ValidationError{Field: "id", Message: "room not found"}
}

func (rm *MockRoomManager) AddMemberToRoom(roomID, userID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return &ValidationError{Field: "roomID", Message: "room not found"}
	}

	// 检查用户是否已经在房间中
	for _, member := range room.Members {
		if member == userID {
			return nil // 用户已在房间中
		}
	}

	room.Members = append(room.Members, userID)
	return nil
}

func (rm *MockRoomManager) RemoveMemberFromRoom(roomID, userID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return &ValidationError{Field: "roomID", Message: "room not found"}
	}

	for i, member := range room.Members {
		if member == userID {
			// 移除成员
			room.Members = append(room.Members[:i], room.Members[i+1:]...)
			return nil
		}
	}

	return &ValidationError{Field: "userID", Message: "user not found in room"}
}

func (rm *MockRoomManager) GetRoomCount() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return len(rm.rooms)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func TestRoomManager(t *testing.T) {
	rm := NewMockRoomManager()
	now := time.Now()

	// 测试创建房间
	room1 := &Room{
		ID:          "room1",
		Name:        "General",
		Description: "General discussion",
		Type:        "public",
		CreatedBy:   "user1",
		CreatedAt:   now,
		Members:     []string{"user1"},
		IsActive:    true,
	}

	err := rm.CreateRoom(room1)
	if err != nil {
		t.Errorf("CreateRoom() failed: %v", err)
	}

	if count := rm.GetRoomCount(); count != 1 {
		t.Errorf("Room count after creation = %d, want 1", count)
	}

	// 测试获取房间
	retrieved, err := rm.GetRoom("room1")
	if err != nil {
		t.Errorf("GetRoom() failed: %v", err)
	}

	if !reflect.DeepEqual(retrieved, room1) {
		t.Errorf("Retrieved room doesn't match original")
	}

	// 测试获取不存在的房间
	_, err = rm.GetRoom("nonexistent")
	if err == nil {
		t.Error("GetRoom() should fail for nonexistent room")
	}

	// 测试重复创建房间
	err = rm.CreateRoom(room1)
	if err == nil {
		t.Error("CreateRoom() should fail for duplicate room")
	}

	// 测试添加成员到房间
	err = rm.AddMemberToRoom("room1", "user2")
	if err != nil {
		t.Errorf("AddMemberToRoom() failed: %v", err)
	}

	// 验证成员已添加
	updated, _ := rm.GetRoom("room1")
	if len(updated.Members) != 2 {
		t.Errorf("Room members count = %d, want 2", len(updated.Members))
	}

	found := false
	for _, member := range updated.Members {
		if member == "user2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("User2 not found in room members")
	}

	// 测试添加已存在的成员（应该不重复）
	err = rm.AddMemberToRoom("room1", "user2")
	if err != nil {
		t.Errorf("AddMemberToRoom() for existing member failed: %v", err)
	}

	updated, _ = rm.GetRoom("room1")
	if len(updated.Members) != 2 {
		t.Errorf("Room members count after duplicate add = %d, want 2", len(updated.Members))
	}

	// 测试从房间移除成员
	err = rm.RemoveMemberFromRoom("room1", "user2")
	if err != nil {
		t.Errorf("RemoveMemberFromRoom() failed: %v", err)
	}

	updated, _ = rm.GetRoom("room1")
	if len(updated.Members) != 1 {
		t.Errorf("Room members count after removal = %d, want 1", len(updated.Members))
	}

	// 测试移除不存在的成员
	err = rm.RemoveMemberFromRoom("room1", "nonexistent")
	if err == nil {
		t.Error("RemoveMemberFromRoom() should fail for nonexistent user")
	}

	// 测试从不存在的房间操作
	err = rm.AddMemberToRoom("nonexistent", "user1")
	if err == nil {
		t.Error("AddMemberToRoom() should fail for nonexistent room")
	}

	err = rm.RemoveMemberFromRoom("nonexistent", "user1")
	if err == nil {
		t.Error("RemoveMemberFromRoom() should fail for nonexistent room")
	}
}

// ====================
// 5. 并发安全测试
// ====================

func TestConcurrentHubOperations(t *testing.T) {
	hub := NewMockHub()

	// 并发添加客户端
	const numClients = 100
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()

			user := &User{
				ID:       fmt.Sprintf("user%d", id),
				Username: fmt.Sprintf("testuser%d", id),
				Status:   "online",
			}

			client := &MockClient{
				ID:   fmt.Sprintf("client%d", id),
				User: user,
				Rooms: map[string]bool{
					"room1": true,
				},
				Send: make(chan []byte, 10),
				hub:  hub,
			}

			hub.RegisterClient(client)
		}(i)
	}

	wg.Wait()

	if count := hub.GetClientCount(); count != numClients {
		t.Errorf("Concurrent client registration: count = %d, want %d", count, numClients)
	}

	if count := hub.GetRoomClientCount("room1"); count != numClients {
		t.Errorf("Concurrent room1 client count = %d, want %d", count, numClients)
	}
}

func TestConcurrentRoomOperations(t *testing.T) {
	rm := NewMockRoomManager()

	// 并发创建房间
	const numRooms = 50
	var wg sync.WaitGroup
	wg.Add(numRooms)

	for i := 0; i < numRooms; i++ {
		go func(id int) {
			defer wg.Done()

			room := &Room{
				ID:          fmt.Sprintf("room%d", id),
				Name:        fmt.Sprintf("Room %d", id),
				Description: fmt.Sprintf("Test room %d", id),
				Type:        "public",
				CreatedBy:   "admin",
				CreatedAt:   time.Now(),
				Members:     []string{"admin"},
				IsActive:    true,
			}

			err := rm.CreateRoom(room)
			if err != nil {
				t.Errorf("Concurrent CreateRoom() failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if count := rm.GetRoomCount(); count != numRooms {
		t.Errorf("Concurrent room creation: count = %d, want %d", count, numRooms)
	}
}

// ====================
// 6. 基准测试
// ====================

func BenchmarkHubRegisterClient(b *testing.B) {
	hub := NewMockHub()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &User{
			ID:       fmt.Sprintf("user%d", i),
			Username: fmt.Sprintf("testuser%d", i),
		}

		client := &MockClient{
			ID:    fmt.Sprintf("client%d", i),
			User:  user,
			Rooms: map[string]bool{"room1": true},
			Send:  make(chan []byte, 10),
		}

		hub.RegisterClient(client)
	}
}

func BenchmarkHubBroadcast(b *testing.B) {
	hub := NewMockHub()

	// 预先注册一些客户端
	for i := 0; i < 100; i++ {
		user := &User{
			ID:       fmt.Sprintf("user%d", i),
			Username: fmt.Sprintf("testuser%d", i),
		}

		client := &MockClient{
			ID:    fmt.Sprintf("client%d", i),
			User:  user,
			Rooms: map[string]bool{"room1": true},
			Send:  make(chan []byte, 100),
		}

		hub.RegisterClient(client)
	}

	message := map[string]interface{}{
		"type":    "message",
		"content": "Benchmark message",
		"user":    "testuser",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastToRoom("room1", message)
	}
}

func BenchmarkMessageValidation(b *testing.B) {
	message := &Message{
		ID:        "msg123",
		RoomID:    "room123",
		UserID:    "user123",
		Username:  "testuser",
		Content:   "Hello, world!",
		Type:      "text",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateMessage(message)
	}
}

func BenchmarkJSONMarshalMessage(b *testing.B) {
	message := Message{
		ID:        "msg123",
		RoomID:    "room123",
		UserID:    "user123",
		Username:  "testuser",
		Content:   "Hello, world!",
		Type:      "text",
		Metadata:  map[string]interface{}{"version": "1.0"},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(message)
		if err != nil {
			b.Fatal(err)
		}
	}
}
