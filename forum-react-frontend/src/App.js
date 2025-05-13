
import { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import Posts from './components/Posts';
import Chat from './components/Chat';

function App() {
    const [token, setToken] = useState(localStorage.getItem('token'));
    const [user, setUser] = useState(JSON.parse(localStorage.getItem('user')));

    return (
        <BrowserRouter>
            <div className="app-container" style={{ display: 'flex', minHeight: '100vh' }}>
                <Routes>
                    <Route path="/login" element={
                        token ? <Navigate to="/" /> :
                            <Login setToken={setToken} setUser={setUser} />
                    } />
                    <Route path="/register" element={<Register />} />
                    <Route path="/" element={
                        token ? (
                            <div style={{ display: 'flex', flex: 1 }}>
                                <div style={{ width: '70%', padding: '20px' }}>
                                    <Posts token={token} user={user} />
                                </div>
                                <div style={{ width: '30%', padding: '20px', borderLeft: '1px solid #ddd' }}>
                                    <Chat token={token} user={user} />
                                </div>
                            </div>
                        ) : (
                            <Navigate to="/login" />
                        )
                    } />
                </Routes>
            </div>
        </BrowserRouter>
    );
}

export default App;