document.addEventListener('DOMContentLoaded', () => {
    const registerForm = document.getElementById('registerForm');
    const loginForm = document.getElementById('loginForm');
    const messageDiv = document.getElementById('message');
    const registerFormContainer = document.getElementById('register-form-container');
    const loginFormContainer = document.getElementById('login-form-container');
    const registerLink = document.getElementById('register-link');
    const loginLink = document.getElementById('login-link');

    // Токен CSRF
    let csrfToken = '';

    // Получение CSRF токена
    const getCsrfToken = async () => {
        try {
            const response = await fetch('/api/v1/csrf-token', {
                credentials: 'include'
            });
            const data = await response.json();
            csrfToken = data.token;
        } catch (error) {
            console.error('Failed to get CSRF token:', error);
        }
    };

    // Инициализация
    getCsrfToken();

    const showMessage = (text, isSuccess) => {
        messageDiv.textContent = text;
        messageDiv.className = isSuccess ? 'success' : 'error';
        messageDiv.style.display = 'block';

        if (isSuccess) {
            setTimeout(() => {
                messageDiv.style.display = 'none';
            }, 3000);
        }
    };

    const clearForms = () => {
        registerForm.reset();
        loginForm.reset();
    };

    // Переключение форм
    registerLink.addEventListener('click', (e) => {
        e.preventDefault();
        registerFormContainer.style.display = 'block';
        loginFormContainer.style.display = 'none';
        clearForms();
    });

    loginLink.addEventListener('click', (e) => {
        e.preventDefault();
        registerFormContainer.style.display = 'none';
        loginFormContainer.style.display = 'block';
        clearForms();
    });

    // Обработка регистрации
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = {
            username: document.getElementById('username').value.trim(),
            email: document.getElementById('email').value.trim(),
            password: document.getElementById('password').value,
            confirmPassword: document.getElementById('confirmPassword').value
        };

        // Валидация
        if (formData.password !== formData.confirmPassword) {
            showMessage("Passwords don't match", false);
            return;
        }

        if (formData.password.length < 8) {
            showMessage("Password must be at least 8 characters", false);
            return;
        }

        try {
            const response = await fetch('/api/v1/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                credentials: 'include',
                body: JSON.stringify({
                    username: formData.username,
                    email: formData.email,
                    password: formData.password
                })
            });

            const data = await response.json();

            if (response.ok) {
                showMessage("Registration successful! Redirecting...", true);
                localStorage.setItem('token', data.access_token);
                localStorage.setItem('username', data.username);
                setTimeout(() => window.location.href = '/', 1500);
            } else {
                showMessage(data.error || "Registration failed", false);
            }
        } catch (error) {
            showMessage("Network error. Please try again.", false);
            console.error('Registration error:', error);
        }
    });

    // Обработка входа
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = {
            username: document.getElementById('loginUsername').value.trim(),
            password: document.getElementById('loginPassword').value
        };

        try {
            const response = await fetch('/api/v1/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                credentials: 'include',
                body: JSON.stringify(formData)
            });

            const data = await response.json();

            if (response.ok) {
                showMessage("Login successful! Redirecting...", true);
                localStorage.setItem('token', data.access_token);
                localStorage.setItem('username', data.username);
                setTimeout(() => window.location.href = '/', 1500);
            } else {
                showMessage(data.error || "Login failed", false);
            }
        } catch (error) {
            showMessage("Network error. Please try again.", false);
            console.error('Login error:', error);
        }
    });
});