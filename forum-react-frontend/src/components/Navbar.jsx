import { useNavigate } from 'react-router-dom';

export default function Navbar({ token, onLogout }) {
    const navigate = useNavigate();

    return (
        <div className="navbar">
            <h2>Forum App</h2>
            <div className="nav-buttons">
                {token ? (
                    <button onClick={onLogout}>Logout</button>
                ) : (
                    <>
                        <button onClick={() => navigate('/login')}>Login</button>
                        <button onClick={() => navigate('/register')}>Register</button>
                    </>
                )}
            </div>
        </div>
    );
}