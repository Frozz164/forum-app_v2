import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../../services/api';

export default function Login({ setToken, setUser }) {
    const [formData, setFormData] = useState({
        username: '',
        password: ''
    });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const navigate = useNavigate();

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!formData.username || !formData.password) {
            setError('Please fill all fields');
            return;
        }

        setIsLoading(true);
        setError('');

        try {
            const { data } = await loginUser(formData);

            if (!data.access_token) {
                throw new Error('No token received from server');
            }

            console.log('Token received:', data.access_token);

            localStorage.setItem('token', data.access_token);
            localStorage.setItem('user', JSON.stringify({
                id: data.user_id,
                username: data.username
            }));

            setToken(data.access_token);
            setUser({
                id: data.user_id,
                username: data.username
            });

            navigate('/');
        } catch (err) {
            console.error('Login error:', err);
            setError(err.response?.data?.error || err.message || 'Login failed. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="auth-form">
            <h2>Login</h2>
            {error && <div className="error-message">{error}</div>}

            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    name="username"
                    placeholder="Username"
                    value={formData.username}
                    onChange={handleChange}
                    required
                    minLength={3}
                />
                <input
                    type="password"
                    name="password"
                    placeholder="Password"
                    value={formData.password}
                    onChange={handleChange}
                    required
                    minLength={6}
                />
                <button
                    type="submit"
                    disabled={isLoading}
                >
                    {isLoading ? 'Logging in...' : 'Login'}
                </button>
            </form>
        </div>
    );
}