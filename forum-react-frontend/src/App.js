import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import Posts from './components/Posts';
import Chat from './components/Chat';
import Navbar from './components/Navbar';
import './styles/chat.css';

function App() {
    const [token, setToken] = useState(() => {
        const storedToken = localStorage.getItem('token');
        return storedToken || null;
    });

    const [user, setUser] = useState(() => {
        const userData = localStorage.getItem('user');
        return userData ? JSON.parse(userData) : null;
    });

    const logout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setToken(null);
        setUser(null);
        window.location.href = '/login';
    };

    return (
        <BrowserRouter>
            <Navbar token={token} onLogout={logout} />
            <div className="app-content">
                <Routes>
                    <Route path="/login" element={
                        token ? <Navigate to="/" /> : <Login setToken={setToken} setUser={setUser} />
                    } />
                    <Route path="/register" element={
                        token ? <Navigate to="/" /> : <Register setToken={setToken} setUser={setUser} />
                    } />
                    <Route path="/" element={
                        <div className="main-content">
                            <Posts token={token} user={user} />
                            <Chat token={token} user={user} />
                        </div>
                    } />
                </Routes>
            </div>
        </BrowserRouter>
    );
}

export default App;