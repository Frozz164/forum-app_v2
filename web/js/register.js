document.addEventListener('DOMContentLoaded', () => {
    const registerForm = document.getElementById('registerForm');
    const loginForm = document.getElementById('loginForm');
    const messageDiv = document.getElementById('message');
    const registerFormContainer = document.getElementById('register-form-container');
    const loginFormContainer = document.getElementById('login-form-container');
    const registerLink = document.getElementById('register-link');
    const loginLink = document.getElementById('login-link');

    // Toggle between forms
    const showRegisterForm = () => {
        registerFormContainer.style.display = 'block';
        loginFormContainer.style.display = 'none';
        clearMessage();
    };

    const showLoginForm = () => {
        registerFormContainer.style.display = 'none';
        loginFormContainer.style.display = 'block';
        clearMessage();
    };

    const clearMessage = () => {
        messageDiv.textContent = '';
        messageDiv.className = '';
    };

    const showMessage = (text, isSuccess) => {
        messageDiv.textContent = text;
        messageDiv.className = isSuccess ? 'success' : 'error';
    };

    // Form toggle handlers
    registerLink.addEventListener('click', (e) => {
        e.preventDefault();
        showRegisterForm();
    });

    loginLink.addEventListener('click', (e) => {
        e.preventDefault();
        showLoginForm();
    });

    // Registration handler
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const username = document.getElementById('username').value;
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        // Validation
        if (username.length < 3) {
            showMessage("Username must be at least 3 characters", false);
            return;
        }

        if (password !== confirmPassword) {
            showMessage("Passwords don't match", false);
            return;
        }

        if (password.length < 8) {
            showMessage("Password must be at least 8 characters", false);
            return;
        }

        try {
            const response = await fetch('/api/v1/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, email, password })
            });

            const result = await response.json();

            if (response.ok) {
                showMessage("Registration successful! Redirecting...", true);
                setTimeout(() => window.location.href = '/login.html', 1500);
            } else {
                showMessage(result.message || "Registration failed", false);
            }
        } catch (error) {
            showMessage("Network error. Please try again.", false);
            console.error('Registration error:', error);
        }
    });

    // Login handler
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const usernameOrEmail = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        try {
            const response = await fetch('/api/v1/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: usernameOrEmail.includes('@') ? undefined : usernameOrEmail,
                    email: usernameOrEmail.includes('@') ? usernameOrEmail : undefined,
                    password
                })
            });

            const result = await response.json();

            if (response.ok) {
                showMessage("Login successful! Redirecting...", true);
                localStorage.setItem('token', result.token);
                localStorage.setItem('username', result.username);
                setTimeout(() => window.location.href = '/profile.html', 1500);
            } else {
                showMessage(result.message || "Login failed", false);
            }
        } catch (error) {
            showMessage("Network error. Please try again.", false);
            console.error('Login error:', error);
        }
    });
});