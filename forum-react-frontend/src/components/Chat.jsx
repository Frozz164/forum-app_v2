import { useState, useEffect } from 'react';

export default function Chat({ token, user }) {
    const [messages, setMessages] = useState([]);
    const [message, setMessage] = useState('');
    const [ws, setWs] = useState(null);

    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`ws://localhost:8081/ws?token=${token}`);

        socket.onopen = () => {
            console.log('WebSocket connected');
            setWs(socket);
        };

        socket.onmessage = (e) => {
            const newMessage = JSON.parse(e.data);
            setMessages(prev => [...prev, newMessage]);
        };

        return () => {
            socket.close();
        };
    }, [token]);

    const sendMessage = () => {
        if (ws && message.trim()) {
            ws.send(JSON.stringify({
                Content: message,
                Type: 1,
                Sender: user.username,
                UserID: user.id
            }));
            setMessage('');
        }
    };

    return (
        <div>
            <h2>Chat</h2>
            <div style={{ height: '300px', overflowY: 'scroll', border: '1px solid #ddd', padding: '10px' }}>
                {messages.map((msg, i) => (
                    <div key={i} style={{ marginBottom: '10px' }}>
                        <strong>{msg.Sender}:</strong> {msg.Content}
                    </div>
                ))}
            </div>
            {token && (
                <div style={{ display: 'flex', marginTop: '10px' }}>
                    <input
                        type="text"
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                        style={{ flex: 1 }}
                    />
                    <button onClick={sendMessage}>Send</button>
                </div>
            )}
        </div>
    );
}