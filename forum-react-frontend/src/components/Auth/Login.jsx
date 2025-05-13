import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../../services/api';

export default function Login({ setToken, setUser }) {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const response = await loginUser({ username, password });
            localStorage.setItem('token', response.data.access_token);
            localStorage.setItem('user', JSON.stringify({
                id: response.data.user_id,
                username: response.data.username
            }));
            setToken(response.data.access_token);
            setUser({
                id: response.data.user_id,
                username: response.data.username
            });
            navigate('/');
        } catch (err) {
            setError('Invalid credentials');
        }
    };

    return (
        <div style={{ maxWidth: '400px', margin: '0 auto' }}>
            <h2>Login</h2>
            {error && <p style={{ color: 'red' }}>{error}</p>}
            <form onSubmit={handleSubmit}>
                <div>
                    <label>Username:</label>
                    <input
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                    />
                </div>
                <div>
                    <label>Password:</label>
                    <input
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                    />
                </div>
                <button type="submit">Login</button>
            </form>
        </div>
    );
}