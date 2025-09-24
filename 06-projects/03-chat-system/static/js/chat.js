class ChatClient {
    constructor() {
        this.ws = null;
        this.currentRoom = null;
        this.currentUser = 'user_' + Math.random().toString(36).substr(2, 9);
        this.isConnected = false;
        this.typingTimer = null;

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connect();
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?user=${this.currentUser}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('Connected to chat server');
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
            console.log('Disconnected from chat server');
            this.isConnected = false;
            // 尝试重连
            setTimeout(() => this.connect(), 3000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    setupEventListeners() {
        // 发送消息
        const sendBtn = document.getElementById('sendBtn');
        const messageInput = document.getElementById('messageInput');

        sendBtn.addEventListener('click', () => this.sendMessage());
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendMessage();
            } else {
                this.handleTyping();
            }
        });

        // 创建房间
        const createRoomBtn = document.getElementById('createRoomBtn');
        createRoomBtn.addEventListener('click', () => this.showCreateRoomModal());

        // 创建房间表单
        const createRoomForm = document.getElementById('createRoomForm');
        createRoomForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.createRoom();
        });

        // 关闭模态框
        const closeBtn = document.querySelector('.close');
        closeBtn.addEventListener('click', () => this.closeModal());

        window.addEventListener('click', (e) => {
            const modal = document.getElementById('createRoomModal');
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    handleMessage(message) {
        switch (message.type) {
            case 'initial_data':
                this.handleInitialData(message.data);
                break;
            case 'message':
                this.displayMessage(message.data);
                break;
            case 'user_joined':
                this.handleUserJoined(message.data);
                break;
            case 'user_left':
                this.handleUserLeft(message.data);
                break;
            case 'room_created':
                this.handleRoomCreated(message.data);
                break;
            case 'typing':
                this.handleTyping(message.data);
                break;
            case 'user_status':
                this.handleUserStatus(message.data);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    handleInitialData(data) {
        this.updateRoomList(data.rooms || []);
        this.updateUserList(data.users || []);
        if (data.messages) {
            this.displayMessages(data.messages);
        }
    }

    sendMessage() {
        const messageInput = document.getElementById('messageInput');
        const content = messageInput.value.trim();

        if (!content || !this.currentRoom || !this.isConnected) {
            return;
        }

        const message = {
            type: 'message',
            data: {
                room_id: this.currentRoom,
                content: content,
                type: 'text'
            }
        };

        this.ws.send(JSON.stringify(message));
        messageInput.value = '';
    }

    joinRoom(roomId) {
        if (!this.isConnected || this.currentRoom === roomId) {
            return;
        }

        const message = {
            type: 'join_room',
            data: { room_id: roomId }
        };

        this.ws.send(JSON.stringify(message));
        this.currentRoom = roomId;

        // 更新 UI
        document.querySelectorAll('.room-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-room-id="${roomId}"]`).classList.add('active');

        // 清空消息容器
        document.getElementById('messagesContainer').innerHTML = '';
    }

    createRoom() {
        const roomName = document.getElementById('roomName').value.trim();
        const roomDescription = document.getElementById('roomDescription').value.trim();

        if (!roomName) {
            alert('请输入房间名称');
            return;
        }

        const message = {
            type: 'create_room',
            data: {
                name: roomName,
                description: roomDescription
            }
        };

        this.ws.send(JSON.stringify(message));
        this.closeModal();
    }

    handleTyping() {
        if (!this.currentRoom || !this.isConnected) {
            return;
        }

        // 清除之前的定时器
        if (this.typingTimer) {
            clearTimeout(this.typingTimer);
        }

        // 发送正在输入的消息
        const message = {
            type: 'typing',
            data: {
                room_id: this.currentRoom,
                is_typing: true
            }
        };
        this.ws.send(JSON.stringify(message));

        // 设置停止输入的定时器
        this.typingTimer = setTimeout(() => {
            const stopMessage = {
                type: 'typing',
                data: {
                    room_id: this.currentRoom,
                    is_typing: false
                }
            };
            this.ws.send(JSON.stringify(stopMessage));
        }, 2000);
    }

    displayMessage(message) {
        const messagesContainer = document.getElementById('messagesContainer');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${message.user_id === this.currentUser ? 'own' : ''} ${message.type === 'system' ? 'system' : ''}`;

        const time = new Date(message.timestamp).toLocaleTimeString();

        messageDiv.innerHTML = `
            <div class="message-header">
                <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=${message.username}" alt="头像" class="message-avatar">
                <span class="message-username">${message.username}</span>
                <span class="message-time">${time}</span>
            </div>
            <div class="message-content">${message.content}</div>
        `;

        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    displayMessages(messages) {
        const messagesContainer = document.getElementById('messagesContainer');
        messagesContainer.innerHTML = '';
        messages.forEach(message => this.displayMessage(message));
    }

    updateRoomList(rooms) {
        const roomList = document.getElementById('roomList');
        roomList.innerHTML = '';

        rooms.forEach(room => {
            const roomItem = document.createElement('li');
            roomItem.className = 'room-item';
            roomItem.setAttribute('data-room-id', room.id);
            roomItem.innerHTML = `
                <span>🏠</span>
                <div>
                    <div>${room.name}</div>
                    <div style="font-size: 12px; color: #bdc3c7;">${room.description || '无描述'}</div>
                </div>
            `;
            roomItem.addEventListener('click', () => {
                this.joinRoom(room.id);
                document.getElementById('currentRoomName').textContent = room.name;
            });
            roomList.appendChild(roomItem);
        });
    }

    updateUserList(users) {
        const userList = document.getElementById('userList');
        userList.innerHTML = '';

        users.forEach(user => {
            const userItem = document.createElement('li');
            userItem.className = 'user-item';
            const statusColor = user.status === 'online' ? '#2ecc71' : user.status === 'away' ? '#f39c12' : '#e74c3c';
            userItem.innerHTML = `
                <div style="width: 10px; height: 10px; background: ${statusColor}; border-radius: 50%;"></div>
                <span>${user.username}</span>
            `;
            userList.appendChild(userItem);
        });
    }

    handleUserJoined(data) {
        // 可以显示系统消息
        this.displayMessage({
            type: 'system',
            content: `${data.username} 加入了房间`,
            username: '系统',
            timestamp: new Date().toISOString()
        });
    }

    handleUserLeft(data) {
        this.displayMessage({
            type: 'system',
            content: `${data.username} 离开了房间`,
            username: '系统',
            timestamp: new Date().toISOString()
        });
    }

    handleRoomCreated(data) {
        // 刷新房间列表或添加新房间
        location.reload(); // 简单的实现，实际中应该动态添加
    }

    handleTyping(data) {
        const typingIndicator = document.getElementById('typingIndicator');
        if (data.is_typing && data.user_id !== this.currentUser) {
            typingIndicator.textContent = `${data.username} 正在输入...`;
        } else {
            typingIndicator.textContent = '';
        }
    }

    handleUserStatus(data) {
        // 更新用户状态，这里可以重新加载用户列表或动态更新
        this.loadOnlineUsers();
    }

    async loadOnlineUsers() {
        try {
            const response = await fetch('/api/users');
            const users = await response.json();
            this.updateUserList(users.data || []);
        } catch (error) {
            console.error('Failed to load users:', error);
        }
    }

    updateCurrentUser() {
        document.getElementById('currentUser').textContent = this.currentUser;
    }

    showCreateRoomModal() {
        document.getElementById('createRoomModal').style.display = 'block';
    }

    closeModal() {
        document.getElementById('createRoomModal').style.display = 'none';
        document.getElementById('roomName').value = '';
        document.getElementById('roomDescription').value = '';
    }
}

// 初始化聊天客户端
document.addEventListener('DOMContentLoaded', () => {
    new ChatClient();
});