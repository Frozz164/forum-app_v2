import { useState, useEffect, useRef } from 'react';

export default function Chat({ token, user }) {
    const [messages, setMessages] = useState([]);
    const [inputMessage, setInputMessage] = useState('');
    const wsRef = useRef(null);

    useEffect(() => {
        const socket = new WebSocket(`ws://localhost:8081/ws`);
        wsRef.current = socket;

        socket.onopen = () => {
            if (token) {
                socket.send(JSON.stringify({
                    type: 'auth',
                    token: token
                }));
            }
            socket.send(JSON.stringify({
                type: 'get_history',
                limit: 100
            }));
        };

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (Array.isArray(data)) {
                    setMessages(data);
                } else {
                    setMessages(prev => [...prev, data]);
                }
            } catch (err) {
                console.error('Error parsing message:', err);
            }
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        return () => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.close();
            }
        };
    }, [token]);

    const sendMessage = () => {
        if (!inputMessage.trim() || !token) return;

        const message = {
            type: 1,
            content: inputMessage,
            sender: user.username,
            userId: user.id,
            timestamp: Date.now()
        };

        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(message));
            setInputMessage('');
        } else {
            console.error('WebSocket is not connected');
        }
    };

    return (
        <div className="chat-container">
            <div className="chat-header">
                <h2>Community Chat</h2>
            </div>

            <div className="messages-container">
                {messages.map((msg) => (
                    <div key={msg.id || msg.timestamp} className={`message ${msg.userId === user?.id ? 'own-message' : ''}`}>
                        <div className="message-header">
                            <span className="message-sender">{msg.sender || 'Anonymous'}</span>
                            <span className="message-time">
                                {new Date(msg.timestamp).toLocaleString()}
                            </span>
                        </div>
                        <div className="message-content">{msg.content}</div>
                    </div>
                ))}
            </div>

            {token ? (
                <div className="message-input-container">
                    <input
                        type="text"
                        value={inputMessage}
                        onChange={(e) => setInputMessage(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                        placeholder="Type your message..."
                    />
                    <button onClick={sendMessage}>Send</button>
                </div>
            ) : (
                <div className="login-notice">
                    Please login to send messages
                </div>
            )}
        </div>
    );
}