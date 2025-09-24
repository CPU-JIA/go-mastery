package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

/*
WebSocket实时通信练习

WebSocket是一种在单个TCP连接上进行全双工通信的协议。
与HTTP不同，WebSocket允许服务器主动向客户端推送数据。

主要概念：
1. 连接升级：从HTTP升级到WebSocket协议
2. 双向通信：客户端和服务器都可以发送数据
3. 实时性：低延迟的消息传递
4. 连接管理：维护多个WebSocket连接
5. 广播机制：向多个客户端发送消息
*/

// WebSocket升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域连接（生产环境需要更严格的检查）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息类型定义
type MessageType string

const (
	TypeJoin      MessageType = "join"
	TypeLeave     MessageType = "leave"
	TypeMessage   MessageType = "message"
	TypeBroadcast MessageType = "broadcast"
	TypePing      MessageType = "ping"
	TypePong      MessageType = "pong"
)

// 消息结构体
type Message struct {
	Type      MessageType `json:"type"`
	UserID    string      `json:"user_id"`
	Username  string      `json:"username"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	RoomID    string      `json:"room_id,omitempty"`
}

// 客户端连接结构体
type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan Message
	RoomID   string
}

// 连接中心（Hub）管理所有WebSocket连接
type Hub struct {
	// 按房间分组的客户端
	Rooms map[string]map[*Client]bool

	// 注册新客户端
	Register chan *Client

	// 注销客户端
	Unregister chan *Client

	// 广播消息给指定房间
	Broadcast chan Message

	// 保护并发访问
	mutex sync.RWMutex
}

// 创建新的Hub
func newHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

// Hub运行主循环
func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// 注册客户端到指定房间
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.Rooms[client.RoomID] == nil {
		h.Rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.Rooms[client.RoomID][client] = true

	log.Printf("客户端 %s (%s) 加入房间 %s", client.ID, client.Username, client.RoomID)

	// 通知房间内其他用户
	joinMessage := Message{
		Type:      TypeJoin,
		UserID:    client.ID,
		Username:  client.Username,
		Content:   fmt.Sprintf("%s 加入了聊天室", client.Username),
		Timestamp: time.Now(),
		RoomID:    client.RoomID,
	}

	// 向房间内所有客户端广播加入消息
	for roomClient := range h.Rooms[client.RoomID] {
		if roomClient != client {
			select {
			case roomClient.Send <- joinMessage:
			default:
				close(roomClient.Send)
				delete(h.Rooms[client.RoomID], roomClient)
			}
		}
	}
}

// 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.Rooms[client.RoomID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)

			log.Printf("客户端 %s (%s) 离开房间 %s", client.ID, client.Username, client.RoomID)

			// 通知房间内其他用户
			leaveMessage := Message{
				Type:      TypeLeave,
				UserID:    client.ID,
				Username:  client.Username,
				Content:   fmt.Sprintf("%s 离开了聊天室", client.Username),
				Timestamp: time.Now(),
				RoomID:    client.RoomID,
			}

			// 向房间内剩余客户端广播离开消息
			for roomClient := range clients {
				select {
				case roomClient.Send <- leaveMessage:
				default:
					close(roomClient.Send)
					delete(clients, roomClient)
				}
			}

			// 如果房间为空，删除房间
			if len(clients) == 0 {
				delete(h.Rooms, client.RoomID)
			}
		}
	}
}

// 广播消息到指定房间
func (h *Hub) broadcastMessage(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.Rooms[message.RoomID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

// 获取房间在线用户列表
func (h *Hub) getRoomUsers(roomID string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var users []string
	if clients, ok := h.Rooms[roomID]; ok {
		for client := range clients {
			users = append(users, client.Username)
		}
	}
	return users
}

// 客户端读消息协程
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// 设置读取超时
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		// 设置消息元数据
		message.UserID = c.ID
		message.Username = c.Username
		message.Timestamp = time.Now()
		message.RoomID = c.RoomID

		// 处理不同类型的消息
		switch message.Type {
		case TypeMessage:
			// 广播普通消息
			c.Hub.Broadcast <- message

		case TypePing:
			// 响应ping消息
			pongMessage := Message{
				Type:      TypePong,
				UserID:    "system",
				Username:  "System",
				Content:   "pong",
				Timestamp: time.Now(),
				RoomID:    c.RoomID,
			}
			select {
			case c.Send <- pongMessage:
			default:
				close(c.Send)
				return
			}
		}
	}
}

// 客户端写消息协程
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

			// 发送消息
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("写入消息错误: %v", err)
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

// 全局Hub实例
var hub = newHub()

// WebSocket连接处理器
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接到WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 获取用户信息
	userID := r.URL.Query().Get("user_id")
	username := r.URL.Query().Get("username")
	roomID := r.URL.Query().Get("room_id")

	if userID == "" || username == "" || roomID == "" {
		conn.Close()
		return
	}

	// 创建客户端
	client := &Client{
		ID:       userID,
		Username: username,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan Message, 256),
		RoomID:   roomID,
	}

	// 注册客户端
	hub.Register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// 获取房间信息API
func handleRoomInfo(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "缺少room_id参数", http.StatusBadRequest)
		return
	}

	users := hub.getRoomUsers(roomID)

	response := map[string]interface{}{
		"room_id":    roomID,
		"user_count": len(users),
		"users":      users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 聊天室HTML模板
const chatHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>WebSocket聊天室</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        #messages { height: 400px; overflow-y: scroll; border: 1px solid #ccc; padding: 10px; margin-bottom: 10px; }
        .message { margin: 5px 0; }
        .system { color: #888; font-style: italic; }
        .user { color: #007bff; font-weight: bold; }
        .timestamp { font-size: 0.8em; color: #666; }
        #messageInput { width: 70%; padding: 5px; }
        #sendButton { padding: 5px 10px; }
        #userInfo { background: #f8f9fa; padding: 10px; margin-bottom: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>WebSocket聊天室</h1>

    <div id="userInfo">
        <strong>用户名:</strong> <span id="username"></span> |
        <strong>房间:</strong> <span id="roomId"></span> |
        <strong>连接状态:</strong> <span id="status">连接中...</span>
    </div>

    <div id="messages"></div>

    <input type="text" id="messageInput" placeholder="输入消息..." />
    <button id="sendButton">发送</button>
    <button id="pingButton">Ping</button>

    <script>
        // 获取URL参数
        const urlParams = new URLSearchParams(window.location.search);
        const userId = urlParams.get('user_id') || 'user_' + Math.random().toString(36).substr(2, 9);
        const username = urlParams.get('username') || 'Anonymous';
        const roomId = urlParams.get('room_id') || 'general';

        document.getElementById('username').textContent = username;
        document.getElementById('roomId').textContent = roomId;

        // 建立WebSocket连接
        const ws = new WebSocket('ws://localhost:8080/ws?user_id=' + userId + '&username=' + encodeURIComponent(username) + '&room_id=' + roomId);

        const messagesDiv = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const sendButton = document.getElementById('sendButton');
        const pingButton = document.getElementById('pingButton');
        const statusSpan = document.getElementById('status');

        ws.onopen = function(event) {
            statusSpan.textContent = '已连接';
            statusSpan.style.color = 'green';
            addMessage('系统', '连接成功', 'system');
        };

        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            addMessage(message.username, message.content, message.type, message.timestamp);
        };

        ws.onclose = function(event) {
            statusSpan.textContent = '已断开';
            statusSpan.style.color = 'red';
            addMessage('系统', '连接已断开', 'system');
        };

        ws.onerror = function(error) {
            statusSpan.textContent = '连接错误';
            statusSpan.style.color = 'red';
            addMessage('系统', '连接错误: ' + error, 'system');
        };

        function addMessage(username, content, type, timestamp) {
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message';

            const time = timestamp ? new Date(timestamp).toLocaleTimeString() : new Date().toLocaleTimeString();

            if (type === 'system' || type === 'join' || type === 'leave') {
                messageDiv.innerHTML = '<span class="system">[' + time + '] ' + content + '</span>';
            } else {
                messageDiv.innerHTML = '<span class="user">' + username + '</span>: ' + content + ' <span class="timestamp">[' + time + ']</span>';
            }

            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }

        function sendMessage() {
            const content = messageInput.value.trim();
            if (content && ws.readyState === WebSocket.OPEN) {
                const message = {
                    type: 'message',
                    content: content
                };
                ws.send(JSON.stringify(message));
                messageInput.value = '';
            }
        }

        function sendPing() {
            if (ws.readyState === WebSocket.OPEN) {
                const message = {
                    type: 'ping',
                    content: 'ping'
                };
                ws.send(JSON.stringify(message));
            }
        }

        sendButton.addEventListener('click', sendMessage);
        pingButton.addEventListener('click', sendPing);

        messageInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>
`

// 聊天室页面处理器
func handleChat(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("chat").Parse(chatHTML))
	tmpl.Execute(w, nil)
}

// 示例：简单的WebSocket回声服务器
func simpleEchoWebSocket() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		// 升级连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("升级失败: %v", err)
			return
		}
		defer conn.Close()

		log.Println("新的WebSocket连接建立")

		// 简单的回声循环
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取消息错误: %v", err)
				break
			}

			log.Printf("收到消息: %s", message)

			// 回声消息
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				log.Printf("写入消息错误: %v", err)
				break
			}
		}
	})
}

// 示例：WebSocket连接池管理
type ConnectionPool struct {
	connections map[string]*websocket.Conn
	mutex       sync.RWMutex
}

func newConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*websocket.Conn),
	}
}

func (pool *ConnectionPool) addConnection(id string, conn *websocket.Conn) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.connections[id] = conn
}

func (pool *ConnectionPool) removeConnection(id string) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	if conn, exists := pool.connections[id]; exists {
		conn.Close()
		delete(pool.connections, id)
	}
}

func (pool *ConnectionPool) broadcast(message []byte) {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	for id, conn := range pool.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("向连接 %s 发送消息失败: %v", id, err)
			// 移除失败的连接
			go pool.removeConnection(id)
		}
	}
}

// 演示不同的WebSocket应用场景
func demonstrateWebSocketUseCases() {
	fmt.Println("=== WebSocket应用场景演示 ===")

	// 1. 实时聊天
	fmt.Println("1. 实时聊天系统")
	fmt.Println("   - 多用户聊天室")
	fmt.Println("   - 私聊功能")
	fmt.Println("   - 消息广播")

	// 2. 实时通知
	fmt.Println("2. 实时通知系统")
	fmt.Println("   - 推送通知")
	fmt.Println("   - 状态更新")
	fmt.Println("   - 系统告警")

	// 3. 实时数据更新
	fmt.Println("3. 实时数据更新")
	fmt.Println("   - 股票价格")
	fmt.Println("   - 游戏状态")
	fmt.Println("   - 监控数据")

	// 4. 协作应用
	fmt.Println("4. 协作应用")
	fmt.Println("   - 在线文档编辑")
	fmt.Println("   - 白板协作")
	fmt.Println("   - 代码协同编程")
}

// WebSocket安全最佳实践
func webSocketSecurityBestPractices() {
	fmt.Println("=== WebSocket安全最佳实践 ===")

	fmt.Println("1. 身份认证和授权")
	fmt.Println("   - 在握手阶段验证用户身份")
	fmt.Println("   - 使用JWT或会话令牌")
	fmt.Println("   - 实施细粒度权限控制")

	fmt.Println("2. 输入验证和消毒")
	fmt.Println("   - 验证所有传入消息")
	fmt.Println("   - 防止XSS攻击")
	fmt.Println("   - 限制消息大小和频率")

	fmt.Println("3. 连接管理")
	fmt.Println("   - 设置连接超时")
	fmt.Println("   - 监控连接数量")
	fmt.Println("   - 实施速率限制")

	fmt.Println("4. 传输安全")
	fmt.Println("   - 使用WSS（WebSocket Secure）")
	fmt.Println("   - 验证证书")
	fmt.Println("   - 防止中间人攻击")
}

func main() {
	// 启动Hub
	go hub.run()

	// 设置路由
	http.HandleFunc("/", handleChat)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/room/info", handleRoomInfo)

	// 设置简单回声服务器
	simpleEchoWebSocket()

	// 演示WebSocket概念
	demonstrateWebSocketUseCases()
	webSocketSecurityBestPractices()

	fmt.Println("=== WebSocket服务器启动 ===")
	fmt.Println("访问 http://localhost:8080 打开聊天室")
	fmt.Println("示例URL:")
	fmt.Println("  http://localhost:8080?username=Alice&room_id=general")
	fmt.Println("  http://localhost:8080?username=Bob&room_id=tech")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  WebSocket: ws://localhost:8080/ws")
	fmt.Println("  回声服务: ws://localhost:8080/echo")
	fmt.Println("  房间信息: http://localhost:8080/room/info?room_id=general")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
练习任务：

1. 基础练习：
   - 修改HTML模板，添加用户列表显示功能
   - 实现私聊功能（点对点消息）
   - 添加消息历史记录功能

2. 中级练习：
   - 实现文件传输功能
   - 添加表情符号支持
   - 实现消息撤回功能
   - 添加用户状态显示（在线/离线/忙碌）

3. 高级练习：
   - 实现分布式聊天室（多服务器）
   - 添加消息持久化（数据库存储）
   - 实现负载均衡
   - 添加消息推送服务集成
   - 实现WebRTC视频聊天集成

4. 安全练习：
   - 实现JWT身份验证
   - 添加速率限制和防洪攻击
   - 实现消息加密
   - 添加管理员权限控制

5. 性能优化：
   - 实现连接池优化
   - 添加消息压缩
   - 实现心跳检测优化
   - 添加内存使用监控

运行说明：
1. 安装依赖：go get github.com/gorilla/websocket
2. 运行程序：go run main.go
3. 打开浏览器访问 http://localhost:8080
4. 在多个浏览器标签页中测试多用户聊天

扩展建议：
- 结合Redis实现分布式会话管理
- 使用消息队列处理高并发场景
- 集成第三方推送服务（如APNs、FCM）
- 实现聊天机器人集成
*/
