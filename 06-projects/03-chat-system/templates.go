package main

import (
	"html/template"
	"net/http"
)

// TemplateHandler handles HTML template rendering
type TemplateHandler struct {
	templates map[string]*template.Template
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler() *TemplateHandler {
	th := &TemplateHandler{
		templates: make(map[string]*template.Template),
	}

	// Load templates
	th.loadTemplates()

	return th
}

// loadTemplates loads all HTML templates
func (th *TemplateHandler) loadTemplates() {
	// Define template files
	templateFiles := map[string]string{
		"chat": "templates/chat.html",
	}

	// Parse each template
	for name, file := range templateFiles {
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			// If template file doesn't exist, create a simple fallback
			tmpl = template.Must(template.New(name).Parse(getDefaultChatTemplate()))
		}
		th.templates[name] = tmpl
	}
}

// RenderTemplate renders a template with given data
func (th *TemplateHandler) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, exists := th.templates[name]
	if !exists {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

// ServeStaticFiles handles static file serving
func ServeStaticFiles() {
	// Serve CSS files
	http.Handle("/static/css/", http.StripPrefix("/static/css/",
		http.FileServer(http.Dir("static/css/"))))

	// Serve JS files
	http.Handle("/static/js/", http.StripPrefix("/static/js/",
		http.FileServer(http.Dir("static/js/"))))
}

// getDefaultChatTemplate returns a fallback template if files don't exist
func getDefaultChatTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>💬 实时聊天系统</title>
    <style>` + getDefaultCSS() + `</style>
</head>
<body>` + getDefaultHTML() + getDefaultJS() + `</body>
</html>`
}

// getDefaultCSS returns default CSS styles
func getDefaultCSS() string {
	return `
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
.chat-container { display: flex; height: 100vh; }
.sidebar { width: 300px; background: #2c3e50; color: white; display: flex; flex-direction: column; }
.sidebar-header { padding: 20px; border-bottom: 1px solid #34495e; }
.user-info { display: flex; align-items: center; gap: 10px; margin-bottom: 20px; }
.user-avatar { width: 40px; height: 40px; border-radius: 50%; }
.user-name { font-weight: bold; }
.status-indicator { width: 10px; height: 10px; background: #2ecc71; border-radius: 50%; }
.sidebar-section { padding: 20px; }
.sidebar-section h3 { margin-bottom: 15px; font-size: 14px; color: #bdc3c7; text-transform: uppercase; }
.room-list, .user-list { list-style: none; }
.room-item, .user-item { padding: 10px; cursor: pointer; border-radius: 5px; margin-bottom: 5px; display: flex; align-items: center; gap: 10px; }
.room-item:hover, .user-item:hover { background: #34495e; }
.room-item.active { background: #3498db; }
.main-chat { flex: 1; display: flex; flex-direction: column; background: #ecf0f1; }
.chat-header { padding: 20px; background: white; border-bottom: 1px solid #bdc3c7; }
.chat-header h2 { margin: 0; color: #2c3e50; }
.messages-container { flex: 1; overflow-y: auto; padding: 20px; }
.message { margin-bottom: 15px; max-width: 70%; }
.message.own { margin-left: auto; }
.message-header { display: flex; align-items: center; gap: 10px; margin-bottom: 5px; }
.message-avatar { width: 32px; height: 32px; border-radius: 50%; }
.message-username { font-weight: bold; font-size: 14px; color: #2c3e50; }
.message-time { font-size: 12px; color: #7f8c8d; }
.message-content { background: white; padding: 12px 16px; border-radius: 18px; box-shadow: 0 1px 2px rgba(0,0,0,0.1); }
.message.own .message-content { background: #3498db; color: white; }
.message.system .message-content { background: #f39c12; color: white; text-align: center; font-style: italic; }
.message-input-container { padding: 20px; background: white; border-top: 1px solid #bdc3c7; }
.input-group { display: flex; gap: 10px; }
.message-input { flex: 1; padding: 12px 16px; border: 1px solid #bdc3c7; border-radius: 25px; font-size: 14px; outline: none; }
.message-input:focus { border-color: #3498db; }
.btn { padding: 12px 20px; border: none; border-radius: 25px; cursor: pointer; font-size: 14px; font-weight: bold; transition: background-color 0.2s; }
.btn-primary { background: #3498db; color: white; }
.btn-primary:hover { background: #2980b9; }
.btn-secondary { background: #95a5a6; color: white; }
.btn-secondary:hover { background: #7f8c8d; }
.modal { display: none; position: fixed; z-index: 1000; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.5); }
.modal-content { background-color: white; margin: 10% auto; padding: 0; border-radius: 10px; width: 500px; max-width: 90%; box-shadow: 0 4px 20px rgba(0,0,0,0.3); }
.modal-header { padding: 20px; border-bottom: 1px solid #ecf0f1; display: flex; justify-content: space-between; align-items: center; }
.modal-header h3 { margin: 0; color: #2c3e50; }
.close { color: #aaa; font-size: 28px; font-weight: bold; cursor: pointer; }
.close:hover { color: #2c3e50; }
.modal-body { padding: 20px; }
.form-group { margin-bottom: 20px; }
.form-group label { display: block; margin-bottom: 5px; font-weight: bold; color: #2c3e50; }
.form-group input, .form-group textarea { width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px; }
.form-group textarea { height: 100px; resize: vertical; }
.form-actions { display: flex; gap: 10px; justify-content: flex-end; }
.typing-indicator { font-size: 12px; color: #7f8c8d; margin-top: 5px; font-style: italic; }
`
}

// getDefaultHTML returns default HTML structure
func getDefaultHTML() string {
	return `
    <div class="chat-container">
        <!-- 侧边栏 -->
        <div class="sidebar">
            <div class="sidebar-header">
                <div class="user-info">
                    <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=user" alt="用户头像" class="user-avatar">
                    <div>
                        <div class="user-name" id="currentUser">匿名用户</div>
                        <div class="status-indicator"></div>
                    </div>
                </div>

                <div class="actions">
                    <button id="createRoomBtn" class="btn btn-primary">创建房间</button>
                </div>
            </div>

            <div class="sidebar-section">
                <h3>房间列表</h3>
                <ul class="room-list" id="roomList"></ul>
            </div>

            <div class="sidebar-section">
                <h3>在线用户</h3>
                <ul class="user-list" id="userList"></ul>
            </div>
        </div>

        <!-- 主聊天区域 -->
        <div class="main-chat">
            <div class="chat-header">
                <h2 id="currentRoomName">选择一个房间开始聊天</h2>
                <div class="typing-indicator" id="typingIndicator"></div>
            </div>

            <div class="messages-container" id="messagesContainer"></div>

            <div class="message-input-container">
                <div class="input-group">
                    <input type="text" id="messageInput" placeholder="输入消息..." class="message-input">
                    <button id="sendBtn" class="btn btn-primary">发送</button>
                </div>
            </div>
        </div>
    </div>

    <!-- 创建房间模态框 -->
    <div id="createRoomModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3>创建新房间</h3>
                <span class="close">&times;</span>
            </div>
            <div class="modal-body">
                <form id="createRoomForm">
                    <div class="form-group">
                        <label for="roomName">房间名称</label>
                        <input type="text" id="roomName" required>
                    </div>
                    <div class="form-group">
                        <label for="roomDescription">房间描述</label>
                        <textarea id="roomDescription"></textarea>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">创建</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                    </div>
                </form>
            </div>
        </div>
    </div>`
}

// getDefaultJS returns default JavaScript
func getDefaultJS() string {
	return `
class ChatClient {
    constructor() {
        this.ws = null;
        this.currentRoom = null;
        this.currentUser = 'user_' + Math.random().toString(36).substr(2, 9);
        this.isConnected = false;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connect();
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = protocol + '//' + window.location.host + '/ws?user=' + this.currentUser;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.isConnected = true;
            this.updateCurrentUser();
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Error parsing message:', error);
            }
        };

        this.ws.onclose = () => {
            this.isConnected = false;
            setTimeout(() => this.connect(), 3000);
        };
    }

    setupEventListeners() {
        const sendBtn = document.getElementById('sendBtn');
        const messageInput = document.getElementById('messageInput');
        const createRoomBtn = document.getElementById('createRoomBtn');

        sendBtn.addEventListener('click', () => this.sendMessage());
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.sendMessage();
        });
        createRoomBtn.addEventListener('click', () => this.showCreateRoomModal());
    }

    handleMessage(message) {
        switch (message.type) {
            case 'initial_data':
                this.handleInitialData(message.data);
                break;
            case 'message':
                this.displayMessage(message.data);
                break;
        }
    }

    handleInitialData(data) {
        this.updateRoomList(data.rooms || []);
        this.updateUserList(data.users || []);
    }

    sendMessage() {
        const messageInput = document.getElementById('messageInput');
        const content = messageInput.value.trim();
        if (!content || !this.currentRoom) return;

        this.ws.send(JSON.stringify({
            type: 'message',
            data: { room_id: this.currentRoom, content: content, type: 'text' }
        }));
        messageInput.value = '';
    }

    displayMessage(message) {
        const container = document.getElementById('messagesContainer');
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message ' + (message.user_id === this.currentUser ? 'own' : '');
        messageDiv.innerHTML = '<div class="message-content">' + message.content + '</div>';
        container.appendChild(messageDiv);
        container.scrollTop = container.scrollHeight;
    }

    updateRoomList(rooms) {
        const roomList = document.getElementById('roomList');
        roomList.innerHTML = '';
        rooms.forEach(room => {
            const li = document.createElement('li');
            li.className = 'room-item';
            li.textContent = room.name;
            li.onclick = () => this.joinRoom(room.id);
            roomList.appendChild(li);
        });
    }

    updateUserList(users) {
        const userList = document.getElementById('userList');
        userList.innerHTML = '';
        users.forEach(user => {
            const li = document.createElement('li');
            li.className = 'user-item';
            li.textContent = user.username;
            userList.appendChild(li);
        });
    }

    joinRoom(roomId) {
        this.currentRoom = roomId;
        this.ws.send(JSON.stringify({
            type: 'join_room',
            data: { room_id: roomId }
        }));
    }

    updateCurrentUser() {
        document.getElementById('currentUser').textContent = this.currentUser;
    }

    showCreateRoomModal() {
        document.getElementById('createRoomModal').style.display = 'block';
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new ChatClient();
});`
}
